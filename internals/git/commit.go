package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ============================================================================
// MAIN COMMIT OPERATIONS
// ============================================================================

func Commit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// CommitGroupFiles commits only the specified files, leaving all other
// staged files untouched and still staged for subsequent group commits.
//
// Strategy (safe, no full reset):
//  1. Identify all currently staged files
//  2. Selectively UNSTAGE the files NOT in this group (git restore --staged)
//  3. Commit the group files (now the only ones staged)
//  4. Re-stage the temporarily unstaged files so the next group can commit them
//
// If the commit fails, the temporarily unstaged files are restored to
// the index before returning the error.
func CommitGroupFiles(filePaths []string, message string) error {
	if len(filePaths) == 0 {
		return fmt.Errorf("no file paths provided to CommitGroupFiles")
	}
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("commit message cannot be empty")
	}

	// Step 1 — find all currently staged files using centralized command
	allStagedOut, err := GetStagedFilesNameOnly()
	if err != nil {
		return fmt.Errorf("failed to list staged files: %w", err)
	}

	// build a lookup set for the files we want to commit
	targetSet := make(map[string]bool, len(filePaths))
	for _, p := range filePaths {
		targetSet[strings.TrimSpace(p)] = true
	}

	// Step 2 — identify staged files that are NOT in this group
	var toUnstage []string
	for _, line := range allStagedOut {
		if !targetSet[line] {
			toUnstage = append(toUnstage, line)
		}
	}

	// Step 3 — selectively unstage non-group files using centralized command
	// git restore --staged -- <files> is safe: it only touches the index,
	// not the working tree. The working-tree changes are preserved.
	if len(toUnstage) > 0 {
		if err := UnstageFiles(toUnstage); err != nil {
			return fmt.Errorf("failed to temporarily unstage non-group files: %w", err)
		}
	}

	// Step 4 — commit the group (now the only staged files)
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	commitErr := cmd.Run()

	// Step 5 — re-stage the files we temporarily unstaged,
	// regardless of whether the commit succeeded or failed
	if len(toUnstage) > 0 {
		if rerr := StageFiles(toUnstage); rerr != nil {
			// this is a critical failure — the user's staging area is now
			// partially modified; tell them explicitly
			return fmt.Errorf(
				"commit %s but failed to re-stage remaining files: %w\n"+
					"Run `git add .` to restore your staging area",
				map[bool]string{true: "succeeded", false: "failed"}[commitErr == nil],
				rerr,
			)
		}
	}

	return commitErr
}

// ============================================================================
// STAGING OPERATIONS
// ============================================================================

// StageFiles stages the given files (git add)
func StageFiles(filePaths []string) error {
	if len(filePaths) == 0 {
		return nil
	}
	args := append([]string{"add", "--"}, filePaths...)
	cmd := exec.Command("git", args...)
	return cmd.Run()
}

// UnstageFiles unstages the given files (git restore --staged)
func UnstageFiles(filePaths []string) error {
	if len(filePaths) == 0 {
		return nil
	}
	args := append([]string{"restore", "--staged", "--"}, filePaths...)
	cmd := exec.Command("git", args...)
	return cmd.Run()
}
