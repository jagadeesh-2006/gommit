package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/jagadeesh-2006/gommit/internals/ai"
	"github.com/jagadeesh-2006/gommit/internals/config"
	"github.com/jagadeesh-2006/gommit/internals/git"
	"github.com/jagadeesh-2006/gommit/internals/grouping"
	"github.com/spf13/cobra"
)

const (
	smartMinFiles      = 5    // auto-activate if >= this many staged files
	smartMinTotalLines = 1000 // auto-activate if >= this many total changed lines
	smartMinAvgLines   = 200  // auto-activate if average changed lines per file >= this
)

// runCmd is the primary command: generate commit messages and commit.
var runCmd = &cobra.Command{
	Use:     "run",
	Aliases: []string{"r"},
	Short:   "Generate and commit messages",
	Run:     runCmdHandler,
}

func init() {
	runCmd.Flags().StringP("context", "c", "", "Why you made this change (optional)")
	runCmd.Flags().BoolP("smart", "s", false, "Force smart grouping mode")
	runCmd.Flags().Bool("no-smart", false, "Disable smart mode even on large diffs")
}

func shouldActivateSmart(files []*grouping.FileInfo) bool {
	if len(files) == 0 {
		return false
	}

	totalLines := 0
	nonSkipFiles := 0
	for _, f := range files {
		if f.Group != grouping.GroupSkip {
			nonSkipFiles++
			totalLines += f.Lines
		}
	}

	if nonSkipFiles == 0 {
		return false
	}

	avgLines := totalLines / nonSkipFiles

	if nonSkipFiles >= smartMinFiles && avgLines >= smartMinAvgLines {
		return true
	}

	if totalLines >= smartMinTotalLines {
		return true
	}

	groups := map[grouping.FileGroup]bool{}
	for _, f := range files {
		if f.Group != grouping.GroupSkip {
			groups[f.Group] = true
		}
	}
	if len(groups) >= 2 && nonSkipFiles >= 3 {
		return true
	}

	return false
}

func runCmdHandler(cmd *cobra.Command, _ []string) {
	// load and validate config
	if !config.Exists() {
		color.Red("Config not found — run `gommit init` first.")
		return
	}
	cfg, err := config.Load()
	if err != nil {
		color.Red("Error loading config: %v", err)
		return
	}

	provider := ai.GetProvider(cfg.Provider, cfg.APIKey, cfg.Model, cfg.CommitStyle, cfg.CustomPrompt)
	if provider == nil {
		color.Red("Error: could not initialise AI provider.")
		return
	}

	// determine mode: explicit flags first, then auto-detection
	smart, _ := cmd.Flags().GetBool("smart")
	noSmart, _ := cmd.Flags().GetBool("no-smart")

	if !noSmart && !smart {
		files, _ := grouping.GetStagedFiles()
		if shouldActivateSmart(files) {
			color.Yellow("ℹ️  Large diff detected (%d files). Activating smart grouping…\n", len(files))
			smart = true
		}
	}

	if smart {
		if err := RunSmart(provider); err != nil {
			color.Red("Error: %v", err)
		}
		return
	}

	runRegularMode(provider, cmd)
}

func runRegularMode(provider ai.Provider, cmd *cobra.Command) {
	reader := bufio.NewReader(os.Stdin)

	// get staged diff
	diff, err := git.Diff()
	if err != nil {
		color.Red("Error getting diff: %v", err)
		return
	}
	if strings.TrimSpace(diff) == "" {
		color.Yellow("Nothing staged. Use `git add` first.")
		return
	}

	// sensitive data check
	if containsSensitiveData(diff) {
		color.Yellow("⚠  Sensitive data detected in diff (passwords, keys, tokens).")
		color.Yellow("   This diff will be sent to: %s", provider)
		color.White("   Continue anyway? (y/n): ")
		confirm := readLineFrom(reader)
		if confirm != "y" {
			color.Yellow("Commit cancelled.")
			return
		}
	}

	// optional context
	contextInput, _ := cmd.Flags().GetString("context")
	if contextInput == "" {
		color.White("Why did you make this change? (optional, Enter to skip): ")
		contextInput = readLineFrom(reader)
	}

	// generate messages
	messages, err := provider.GenerateCommitMessage(diff, contextInput, "", "")
	if err != nil {
		color.Red("Error generating commit messages: %v", err)
		return
	}

	printNumberedMessages(messages)

	color.White("Select (1/2/3), [r] regenerate, [e] edit, [n] cancel: ")
	choice := readLineFrom(reader)

	switch choice {
	case "1", "2", "3":
		commitByIndex(messages, choice)

	case "r":
		handleRegularRegenerate(provider, diff, contextInput, messages, reader)

	case "e":
		handleRegularEdit(messages, reader)

	case "n":
		color.Yellow("Commit cancelled.")

	default:
		color.Red("Invalid choice. Commit aborted.")
	}
}

func handleRegularRegenerate(
	provider ai.Provider,
	diff, contextInput string,
	previousMessages []string,
	reader *bufio.Reader,
) {
	color.Cyan("Regenerating commit messages…")

	newMessages, err := provider.GenerateCommitMessage(diff, contextInput, previousMessages[0], "")
	if err != nil {
		color.Red("Error generating commit messages: %v", err)
		return
	}
	printNumberedMessages(newMessages)

	color.White("Select (1/2/3), [e] edit, [n] cancel: ")
	choice := readLineFrom(reader)

	switch choice {
	case "1", "2", "3":
		commitByIndex(newMessages, choice)
	case "e":
		handleRegularEdit(newMessages, reader)
	case "n":
		color.Yellow("Commit cancelled.")
	default:
		color.Red("Invalid choice.")
	}
}

func handleRegularEdit(messages []string, reader *bufio.Reader) {
	color.White("Which message to edit? (1/2/3): ")
	choice := readLineFrom(reader)

	idx, err := strconv.Atoi(choice)
	if err != nil || idx < 1 || idx > len(messages) {
		color.Red("❌ Invalid choice.")
		return
	}
	selected := messages[idx-1]

	// use readline for prefilled editing
	rl, err := readline.New("> ")
	if err != nil {
		color.Red("Error starting editor: %v", err)
		return
	}
	defer rl.Close()

	rl.WriteStdin([]byte(selected))
	color.White("Edit message:")
	edited, _ := rl.Readline()
	edited = strings.TrimSpace(edited)
	if edited == "" {
		edited = selected
	}

	if err := git.Commit(edited); err != nil {
		color.Red("Error committing: %v", err)
		return
	}
	color.Green("✅ Committed: %s", edited)
}

// commitByIndex commits the message at the 1-based index string.
func commitByIndex(messages []string, choice string) {
	idx, err := strconv.Atoi(choice)
	if err != nil || idx < 1 || idx > len(messages) {
		color.Red("❌ Invalid selection.")
		return
	}
	selected := messages[idx-1]
	if err := git.Commit(selected); err != nil {
		color.Red("Error committing: %v", err)
		return
	}
	color.Green("✅ Committed: %s", selected)
}

// printNumberedMessages prints the numbered list of suggestions.
func printNumberedMessages(messages []string) {
	fmt.Println()
	for i, msg := range messages {
		color.Cyan("  %d. %s", i+1, msg)
	}
	fmt.Println()
}

func readLineFrom(r *bufio.Reader) string {
	line, _ := r.ReadString('\n')
	return strings.TrimSpace(line)
}

// containsSensitiveData scans only the added (+) lines in a diff for
// patterns that look like secrets, keys, or tokens.
func containsSensitiveData(diff string) bool {
	var addedLines []string
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			addedLines = append(addedLines, line)
		}
	}
	addedContent := strings.Join(addedLines, "\n")

	patterns := []string{
		`(?i)(password|passwd|secret|api_key|apikey|access_key)\s*[:=]\s*["']?[a-zA-Z0-9+/]{8,}`,
		`sk-[a-zA-Z0-9]{20,}`,
		`gsk_[a-zA-Z0-9]{20,}`,
		`AIza[a-zA-Z0-9]{20,}`,
		`sk-ant-[a-zA-Z0-9]{20,}`,
		`(?i)Bearer\s+[a-zA-Z0-9\-._~+/]{20,}`,
		`-----BEGIN (RSA |EC )?PRIVATE KEY-----`,
	}
	for _, p := range patterns {
		if matched, _ := regexp.MatchString(p, addedContent); matched {
			return true
		}
	}
	return false
}