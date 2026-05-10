package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"time"
	"strings"
	"github.com/fatih/color"
	"github.com/jagadeesh-2006/gommit/internals/ai"
	"github.com/jagadeesh-2006/gommit/internals/git"
	"github.com/jagadeesh-2006/gommit/internals/grouping"
)

type SmartRunner struct {
	provider ai.Provider
	reader   *bufio.Reader
}

// RunSmart executes the smart grouping workflow
func RunSmart(provider ai.Provider) error {
	sr := &SmartRunner{
		provider: provider,
		reader:   bufio.NewReader(os.Stdin),
	}

	// Step 1: Get staged files
	files, err := grouping.GetStagedFiles()
	if err != nil {
		return err
	}

	if len(files) == 0 {
		color.Yellow("No staged files found")
		return nil
	}

	// Step 2: Group files
	grouped := grouping.GroupFiles(files)

	// Step 3: Show summary
	sr.showGroupSummary(grouped)

	// Step 4: Process each group
	totalCommits := 0
	for _, group := range []grouping.FileGroup{
		grouping.GroupCode,
		grouping.GroupConfig,
		grouping.GroupDocs,
		grouping.GroupTest,
		grouping.GroupCI,
	} {
		if groupFiles, exists := grouped.Files[group]; exists && len(groupFiles) > 0 {
			commits, err := sr.processGroup(group, groupFiles)
			if err != nil {
				color.Red("Error processing %s: %v", group, err)
				continue
			}
			totalCommits += commits
		}
	}

	// Step 5: Auto-skip lock files silently
	if skipFiles, exists := grouped.Files[grouping.GroupSkip]; exists && len(skipFiles) > 0 {
		color.White("\n🚫 Skipped %d dependency lock files (no value in commits)", len(skipFiles))
	}

	// Step 6: Summary
	color.Green("\n✅ Done. %d commits created.\n", totalCommits)
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
		if files, exists := grouped.Files[g]; exists && len(files) > 0 {
			lines := 0
			for _, f := range files {
				lines += f.Lines
			}
			color.White("  %s (%d file(s), %d line(s) changed)", g, len(files), lines)
		}
	}
	fmt.Println()
}

func (sr *SmartRunner) processGroup(group grouping.FileGroup, files []*grouping.FileInfo) (int, error) {
	// For code files, sub-group by directory if multiple directories
	var subGroups map[string][]*grouping.FileInfo
	if group == grouping.GroupCode {
		subGroups = grouping.SubGroupCodeByDirectory(files)
	} else {
		subGroups = make(map[string][]*grouping.FileInfo)
		subGroups[""] = files
	}

	commits := 0
	for dir, subGroupFiles := range subGroups {
		commit, err := sr.processSubGroup(group, dir, subGroupFiles)
		if err != nil {
			return commits, err
		}
		commits += commit
	}

	return commits, nil
}

func (sr *SmartRunner) processSubGroup(group grouping.FileGroup, dir string, files []*grouping.FileInfo) (int, error) {
    filePaths := make([]string, len(files))
    for i, f := range files {
        filePaths[i] = f.Path
    }

    // get stat for context
    stat, _ := grouping.GetStatForGroup(filePaths)

    // get right diff for this group type
    diff, err := grouping.GetDiffForGroup(group, filePaths)
    if err != nil {
        return 0, err
    }

    // build prompt with all context
    prompt := grouping.BuildPromptForGroup(group, files, stat, diff)

    // show header
    fmt.Println()
    if dir != "" && dir != "." {
        color.Cyan("─── %s / %s ───", group, dir)
    } else {
        color.Cyan("─── %s ───", group)
    }
    for _, f := range files {
        color.White("   %s (%d lines)", f.Path, f.Lines)
    }
    fmt.Println()

    // retry logic
    var messages []string
    for attempt := 1; attempt <= 3; attempt++ {
        messages, err = sr.provider.GenerateCommitMessage(diff, "", prompt)
        if err == nil {
            break
        }
        if attempt < 3 {
            color.Yellow("⚠ Retrying (%d/3)...", attempt)
            time.Sleep(2 * time.Second)
        }
    }
    if err != nil {
        color.Red("❌ Failed after 3 attempts: %v", err)
        color.White("Skip this group? (y/n): ")
        var skip string
        fmt.Scanln(&skip)
        if skip == "y" {
            return 0, nil
        }
        return 0, err
    }

    // show messages
    for i, msg := range messages {
        color.Cyan("  %d. %s", i+1, msg)
    }
    fmt.Println()

    color.White("Select (1/2/3), [e]dit, [s]kip, [r]egenerate: ")
    var choice string
	fmt.Scanln(&choice)

    switch choice {
    case "1", "2", "3":
        idx, _ := strconv.Atoi(choice)
        selected := messages[idx-1]
        // commit ONLY this group's files
        if err := git.CommitGroupFiles(filePaths, selected); err != nil {
            color.Red("❌ Commit failed: %v", err)
            return 0, err
        }
        color.Green("✅ %s", selected)
        return 1, nil

    case "e":
        return sr.handleEdit(messages, filePaths)

    case "s":
        color.Yellow("⏭  Skipped")
        return 0, nil

    case "r":
        return sr.handleRegenerate(group, files, filePaths, messages, stat)

    default:
        color.Yellow("⏭  Skipped")
        return 0, nil
    }
}

func (sr *SmartRunner) handleEdit(messages []string, filePaths []string) (int, error) {
    color.White("Which message? (1/2/3): ")
    var choice string
    fmt.Scanln(&choice)
    idx, _ := strconv.Atoi(choice)
    if idx < 1 || idx > len(messages) {
        color.Red("❌ Invalid choice")
        return 0, nil
    }

    selected := messages[idx-1]
    color.White("Current: ")
    color.Cyan("  %s", selected)
    color.White("New message: ")
    var edited string
    fmt.Scanln(&edited)

    if edited == "" {
        edited = selected // use original if empty
    }

    if err := git.CommitGroupFiles(filePaths, edited); err != nil {
        color.Red("❌ Commit failed: %v", err)
        return 0, nil
    }
    color.Green("✅ %s", edited)
    return 1, nil
}

func (sr *SmartRunner) handleRegenerate(
    group grouping.FileGroup,
    files []*grouping.FileInfo,
    filePaths []string,
    previousMessages []string,
    stat string,
) (int, error) {
    // get fresh diff
    diff, err := grouping.GetDiffForGroup(group, filePaths)
    if err != nil {
        return 0, err
    }

    // build prompt — tell AI to avoid previous messages
    avoid := strings.Join(previousMessages, " | ")
    prompt := grouping.BuildPromptForGroup(group, files, stat, diff)
    prompt += fmt.Sprintf("\n\nDo NOT repeat or rephrase any of these: %s", avoid)

    // retry logic same as processSubGroup
    var messages []string
    for attempt := 1; attempt <= 3; attempt++ {
        messages, err = sr.provider.GenerateCommitMessage(diff, "", prompt)
        if err == nil {
            break
        }
        if attempt < 3 {
            color.Yellow("⚠ Retrying (%d/3)...", attempt)
            time.Sleep(2 * time.Second)
        }
    }
    if err != nil {
        color.Red("❌ Failed: %v", err)
        return 0, nil
    }

    // show new messages
    for i, msg := range messages {
        color.Cyan("  %d. %s", i+1, msg)
    }
    fmt.Println()

    color.White("Select (1/2/3), [e]dit, [s]kip: ")
    var choice string
    fmt.Scanln(&choice)

    switch choice {
    case "1", "2", "3":
        idx, _ := strconv.Atoi(choice)
        selected := messages[idx-1]
        if err := git.CommitGroupFiles(filePaths, selected); err != nil {
            color.Red("❌ Commit failed: %v", err)
            return 0, nil
        }
        color.Green("✅ %s", selected)
        return 1, nil

    case "e":
        return sr.handleEdit(messages, filePaths)

    default:
        color.Yellow("⏭  Skipped")
        return 0, nil
    }
}