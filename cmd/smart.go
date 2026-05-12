package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/jagadeesh-2006/gommit/internals/ai"
	"github.com/jagadeesh-2006/gommit/internals/git"
	"github.com/jagadeesh-2006/gommit/internals/grouping"
)

type SmartRunner struct {
	provider ai.Provider
	reader   *bufio.Reader
}

func RunSmart(provider ai.Provider) error {
	sr := &SmartRunner{
		provider: provider,
		reader:   bufio.NewReader(os.Stdin),
	}

	// Step 1 — get staged files (batch line count, no N+1)
	files, err := grouping.GetStagedFiles()
	if err != nil {
		return fmt.Errorf("getting staged files: %w", err)
	}
	if len(files) == 0 {
		color.Yellow("No staged files found.")
		return nil
	}

	// Step 2 — group by file type
	grouped := grouping.GroupFiles(files)

	// Step 3 — show summary so the user can see what will happen
	sr.showGroupSummary(grouped)

	// Step 4 — process each active group in a deterministic order
	totalCommits := 0
	for _, group := range []grouping.FileGroup{
		grouping.GroupCode,
		grouping.GroupConfig,
		grouping.GroupDocs,
		grouping.GroupTest,
		grouping.GroupCI,
	} {
		groupFiles, exists := grouped.Files[group]
		if !exists || len(groupFiles) == 0 {
			continue
		}
		commits, err := sr.processGroup(group, groupFiles)
		if err != nil {
			color.Red("Error processing %s: %v", group, err)
			continue
		}
		totalCommits += commits
	}

	// Step 5 — silently report skipped files (lock files etc.)
	if skipFiles, exists := grouped.Files[grouping.GroupSkip]; exists && len(skipFiles) > 0 {
		color.White("\n🚫 Skipped %d lock/binary file(s) — no value in commits.", len(skipFiles))
	}

	// Step 6 — final summary
	fmt.Println()
	color.Green("✅ Done. %d commit(s) created.", totalCommits)
	fmt.Println()
	return nil
}

func (sr *SmartRunner) showGroupSummary(grouped *grouping.GroupedFiles) {
	fmt.Println()
	color.White("Auto-grouped staged files:\n")

	for _, g := range []grouping.FileGroup{
		grouping.GroupCode,
		grouping.GroupConfig,
		grouping.GroupDocs,
		grouping.GroupTest,
		grouping.GroupCI,
		grouping.GroupSkip,
	} {
		files, exists := grouped.Files[g]
		if !exists || len(files) == 0 {
			continue
		}
		totalLines := 0
		for _, f := range files {
			totalLines += f.Lines
		}
		color.White("  %s  %d file(s), %d line(s) changed", g, len(files), totalLines)
	}
	fmt.Println()
}

func (sr *SmartRunner) processGroup(group grouping.FileGroup, files []*grouping.FileInfo) (int, error) {
	if group != grouping.GroupCode {
		// non-code groups: commit as one unit
		return sr.processSubGroup(group, "", files)
	}
	subGroups := grouping.SubGroupCodeByDirectory(files)
	total := 0
	for dir, subFiles := range subGroups {
		n, err := sr.processSubGroup(group, dir, subFiles)
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}

// processSubGroup handles one atomic commit unit: one group (optionally
// further scoped to a directory) and its files.
//
// Flow:
//  1. Build structured prompt via BuildStructuredPrompt (structured extraction)
//  2. Call provider with retry
//  3. Present options to user
//  4. Commit or skip based on user choice
func (sr *SmartRunner) processSubGroup(
	group grouping.FileGroup,
	dir string,
	files []*grouping.FileInfo,
) (int, error) {
	// collect file paths for git operations
	filePaths := make([]string, len(files))
	for i, f := range files {
		filePaths[i] = f.Path
	}

	// --- structured diff extraction + prompt building ---
	// BuildStructuredPrompt:
	//   1. gets the right diff per group type (U0 / word-diff)
	//   2. runs per-file signature + hunk extraction
	//   3. ranks files by importance score
	//   4. packs within token budget
	//   5. returns the compressed diff and the full prompt
	compressedDiff, prompt, err := grouping.BuildStructuredPrompt(group, files)
	if err != nil {
		return 0, fmt.Errorf("building prompt: %w", err)
	}

	// --- display header ---
	fmt.Println()
	if dir != "" && dir != "." {
		color.Cyan("─── %s / %s ───", group, dir)
	} else {
		color.Cyan("─── %s ───", group)
	}
	for _, f := range files {
		status := statusLabel(f.Status)
		color.White("   %s%s (%d lines)", f.Path, status, f.Lines)
	}
	fmt.Println()

	// --- generate with retry ---
	var messages []string
	for attempt := 1; attempt <= 3; attempt++ {
		messages, err = sr.provider.GenerateCommitMessage(compressedDiff, "", "", prompt)
		if err == nil {
			break
		}
		if attempt < 3 {
			color.Yellow("⚠  Retrying (%d/3)...", attempt)
			time.Sleep(2 * time.Second)
		}
	}
	if err != nil {
		color.Red("❌ Failed after 3 attempts: %v", err)
		if sr.confirmSkip() {
			return 0, nil
		}
		return 0, err
	}

	// --- present messages ---
	sr.printMessages(messages)

	// --- user choice ---
	return sr.handleChoice(messages, filePaths, group, files)
}

// handleChoice presents the selection prompt and dispatches to the right handler.
func (sr *SmartRunner) handleChoice(
	messages []string,
	filePaths []string,
	group grouping.FileGroup,
	files []*grouping.FileInfo,
) (int, error) {
	color.White("Select (1/2/3), [e]dit, [s]kip, [r]egenerate: ")
	choice := sr.readLine()

	switch choice {
	case "1", "2", "3":
		return sr.commitSelected(messages, filePaths, choice)

	case "e":
		return sr.handleEdit(messages, filePaths)

	case "s":
		color.Yellow("⏭  Skipped.")
		return 0, nil

	case "r":
		return sr.handleRegenerate(group, files, filePaths, messages)

	default:
		color.Yellow("⏭  Invalid choice — skipped.")
		return 0, nil
	}
}

// commitSelected commits the message at the given 1-based index string.
func (sr *SmartRunner) commitSelected(messages []string, filePaths []string, choice string) (int, error) {
	idx, err := strconv.Atoi(choice)
	if err != nil || idx < 1 || idx > len(messages) {
		color.Red("❌ Invalid selection.")
		return 0, nil
	}
	selected := messages[idx-1]
	if err := git.CommitGroupFiles(filePaths, selected); err != nil {
		color.Red("❌ Commit failed: %v", err)
		return 0, err
	}
	color.Green("✅ %s", selected)
	return 1, nil
}

func (sr *SmartRunner) handleEdit(messages []string, filePaths []string) (int, error) {
	color.White("Which message to edit? (1/2/3): ")
	choice := sr.readLine()

	idx, err := strconv.Atoi(choice)
	if err != nil || idx < 1 || idx > len(messages) {
		color.Red("❌ Invalid choice.")
		return 0, nil
	}

	selected := messages[idx-1]
	color.White("Current:")
	color.Cyan("  %s", selected)
	color.White("New message (press Enter to keep current): ")
	edited := sr.readLine()

	if strings.TrimSpace(edited) == "" {
		edited = selected
	}

	if err := git.CommitGroupFiles(filePaths, edited); err != nil {
		color.Red("❌ Commit failed: %v", err)
		return 0, err
	}
	color.Green("✅ %s", edited)
	return 1, nil
}

func (sr *SmartRunner) handleRegenerate(group grouping.FileGroup, files []*grouping.FileInfo, filePaths []string, previousMessages []string) (int, error) {
	compressedDiff, prompt, err := grouping.BuildStructuredPrompt(group, files)
	if err != nil {
		return 0, fmt.Errorf("rebuilding prompt: %w", err)
	}
	avoid := strings.Join(previousMessages, " | ")
	prompt += fmt.Sprintf("\n\nDo NOT repeat or rephrase any of these previous messages: %s", avoid)

	// retry
	var messages []string
	for attempt := 1; attempt <= 3; attempt++ {
		messages, err = sr.provider.GenerateCommitMessage(compressedDiff,"", "", prompt)
		if err == nil {
			break
		}
		if attempt < 3 {
			color.Yellow("⚠  Retrying (%d/3)...", attempt)
			time.Sleep(2 * time.Second)
		}
	}
	if err != nil {
		color.Red("❌ Regeneration failed: %v", err)
		return 0, nil
	}

	sr.printMessages(messages)

	color.White("Select (1/2/3), [e]dit, [s]kip: ")
	choice := sr.readLine()

	switch choice {
	case "1", "2", "3":
		return sr.commitSelected(messages, filePaths, choice)
	case "e":
		return sr.handleEdit(messages, filePaths)
	default:
		color.Yellow("⏭  Skipped.")
		return 0, nil
	}
}

func (sr *SmartRunner) printMessages(messages []string) {
	fmt.Println()
	for i, msg := range messages {
		color.Cyan("  %d. %s", i+1, msg)
	}
	fmt.Println()
}

func (sr *SmartRunner) readLine() string {
	line, _ := sr.reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func (sr *SmartRunner) confirmSkip() bool {
	color.White("Skip this group? (y/n): ")
	return strings.TrimSpace(sr.readLine()) == "y"
}

func statusLabel(status string) string {
	switch status {
	case "A":
		return " [new]"
	case "D":
		return " [deleted]"
	case "R":
		return " [renamed]"
	default:
		return ""
	}
}
