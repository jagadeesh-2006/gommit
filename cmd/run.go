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
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:     "run",
	Aliases: []string{"r"},
	Short:   "Generate and commit messages",
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("Running gommit...")
		// load config
		if !config.Exists() {
			color.Red(" Config not found — run `gommit init` first")
			return
		}

		cfg, err := config.Load()
		if err != nil {
			color.Red("Error loading config: %s", err)
			return
		}

		//  get diff
		diff, err := git.Diff()
		if err != nil {
			color.Red("Error getting diff: %s", err)
			return
		}
		// sensitive data check
		if containsSensitiveData(diff) {
			color.Yellow("Sensitive data detected in diff (passwords, keys, tokens)")
			color.Yellow("  This diff will be sent to: %s", cfg.Provider)
			fmt.Print("  Continue anyway? (y/n): ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "y" {
				color.Yellow("Commit cancelled.")
				return
			}
		}
		// fmt.Println("Got diff ")
		// step 3 - get context
		contextInput, _ := cmd.Flags().GetString("context")
		if contextInput == "" {
			reader := bufio.NewReader(os.Stdin)
			color.White("Why did you make this change? (optional, press Enter to skip): ")
			contextInput, _ = reader.ReadString('\n')
			contextInput = strings.TrimSpace(contextInput)
		}

		// get provider
		provider := ai.GetProvider(cfg.Provider, cfg.APIKey, cfg.Model, cfg.CommitStyle, cfg.CustomPrompt)
		if provider == nil {
			color.Red("Error getting AI provider:")
			return
		}
		// generate commit messages
		messages, err := provider.GenerateCommitMessage(diff, contextInput, "")
		if err != nil {
			color.Red("Error generating commit messages: %s", err)
			return
		}

		// show 3 messages
		fmt.Println()
		for i, msg := range messages {
			color.Cyan("  %d. %s", i+1, msg)
		}
		fmt.Println()

		color.White("Select (1/2/3), [r] regenerate, [e] edit, [n] cancel: ")
		var choice string
		fmt.Scanln(&choice)

		switch choice {
		case "1", "2", "3":
			idx, _ := strconv.Atoi(choice)
			selected := messages[idx-1]
			err := git.Commit(selected)
			if err != nil {
				color.Red("Error committing changes: %s", err)
				return
			}
			color.Green("✅ Committed: %s", selected)

		case "r":
			color.Cyan("Regenerating commit messages...")
			newMessages, err := provider.GenerateCommitMessage(diff, contextInput, messages[0])
			if err != nil {
				color.Red("Error generating commit messages: %s", err)
				return
			}

			// show 3 new messages
			fmt.Println()
			for i, msg := range newMessages {
				color.Cyan("  %d. %s", i+1, msg)
			}
			fmt.Println()

			color.White("Select (1/2/3), [e] edit, [n] cancel: ")
			var newChoice string
			fmt.Scanln(&newChoice)

			switch newChoice {
			case "1", "2", "3":
				idx, _ := strconv.Atoi(newChoice)
				selected := newMessages[idx-1]
				err := git.Commit(selected)
				if err != nil {
					color.Red("Error committing changes: %s", err)
					return
				}
				color.Green("✅ Committed: %s", selected)
			case "e":
				color.White("Which message to edit? (1/2/3): ")
				var editChoice string
				fmt.Scanln(&editChoice)
				idx, _ := strconv.Atoi(editChoice)
				if idx < 1 || idx > 3 {
					color.Red("Invalid choice.")
					return
				}
				selected := newMessages[idx-1]

				reader := bufio.NewReader(os.Stdin)
				color.White("Edit message: ")
				fmt.Print(selected)
				edited, _ := reader.ReadString('\n')
				edited = strings.TrimSpace(edited)

				if edited != "" {
					err := git.Commit(edited)
					if err != nil {
						color.Red("Error committing changes: %s", err)
						return
					}
					color.Green("✅ Committed: %s", edited)
				} else {
					color.Yellow("Commit aborted.")
				}
			case "n":
				color.Yellow("Commit cancelled.")
			default:
				color.Red("Invalid choice. Commit aborted.")
			}

		case "e":
    color.White("Which message to edit? (1/2/3): ")
    var editChoice string
    fmt.Scanln(&editChoice)
    idx, _ := strconv.Atoi(editChoice)
    if idx < 1 || idx > len(messages) {
        color.Red("❌ Invalid choice.")
        return
    }
    selected := messages[idx-1]

    rl, err := readline.New("> ")
    if err != nil {
        color.Red("Error: %s", err)
        return
    }
    defer rl.Close()

    // prefill with selected message
    rl.WriteStdin([]byte(selected))
    color.White("Edit message:")
    edited, _ := rl.Readline()
    edited = strings.TrimSpace(edited)

    if edited == "" {
        edited = selected
    }

    err = git.Commit(edited)
    if err != nil {
        color.Red("Error committing: %s", err)
        return
    }
    color.Green("✅ Committed: %s", edited)
		case "n":
			color.Yellow("Commit cancelled.")
		default:
			color.Red("Invalid choice. Commit aborted.")
		}
	},
}

func containsSensitiveData(diff string) bool {
	// only scan added lines
	addedLines := []string{}
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			addedLines = append(addedLines, line)
		}
	}
	addedContent := strings.Join(addedLines, "\n")

	patterns := []string{
		// assignment patterns with actual values
		`(?i)(password|passwd|secret|api_key|apikey|access_key)\s*[:=]\s*["']?[a-zA-Z0-9+/]{8,}`,
		// known API key formats
		`sk-[a-zA-Z0-9]{20,}`,
		`gsk_[a-zA-Z0-9]{20,}`,
		`AIza[a-zA-Z0-9]{20,}`,
		`sk-ant-[a-zA-Z0-9]{20,}`,
		// bearer tokens
		`(?i)Bearer\s+[a-zA-Z0-9\-._~+/]{20,}`,
		// private keys
		`-----BEGIN (RSA |EC )?PRIVATE KEY-----`,
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, addedContent)
		if matched {
			return true
		}
	}
	return false
}

func init() {
	runCmd.Flags().StringP("context", "c", "", "Context for why changes were made")
}
