package git

import (
	"fmt"
	"strings"
)

// Diff runs the git diff command and returns the output as a string
func Diff() (string, error) {
	// Validate environment
	if err := CheckGitInstalled(); err != nil {
		return "", err
	}
	if err := CheckGitRepository(); err != nil {
		return "", err
	}

	// Get default diff for all staged files
	out, err := execCommand("git", "diff", "--staged")
	if err != nil {
		return "", err
	}
	if len(out) == 0 {
		return "", fmt.Errorf("no staged changes found")
	}
	return out, nil
}

// GetStagedFilesNameOnly returns only the names of staged files
func GetStagedFilesNameOnly() ([]string, error) {
	out, err := execCommand("git", "diff", "--cached", "--name-only")
	if err != nil {
		return nil, fmt.Errorf("not in a git repo or nothing staged")
	}
	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return nil, fmt.Errorf("no staged changes found")
	}
	var files []string
	for _, line := range strings.Split(trimmed, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

// GetStagedFilesNameStatus returns staged files with their git status (A/M/D/R)
// Output format: "status\tfilepath"
func GetStagedFilesNameStatus() (string, error) {
	out, err := execCommand("git", "diff", "--cached", "--name-status")
	if err != nil {
		return "", fmt.Errorf("failed to get staged files: %w", err)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return "", fmt.Errorf("no staged changes found")
	}
	return trimmed, nil
}

// GetStagedShortStat returns a single-line summary of all staged changes
// Example: "5 files changed, 47 insertions(+), 12 deletions(-)"
func GetStagedShortStat() (string, error) {
	out, err := execCommand("git", "diff", "--cached", "--shortstat")
	if err != nil {
		return "", fmt.Errorf("git diff --shortstat: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// GetStagedNumStat returns line counts for all staged files
// Output format per line: "added\tremoved\tfilename"
func GetStagedNumStat() (string, error) {
	out, err := execCommand("git", "diff", "--cached", "--numstat")
	if err != nil {
		return "", fmt.Errorf("git diff --numstat: %w", err)
	}
	return strings.TrimSpace(out), nil
}


// GetFileDiffU0 returns the -U0 diff for a single file
// -U0 removes the 3 surrounding unchanged context lines, showing only changed lines
func GetFileDiffU0(filePath string) (string, error) {
	out, err := execCommand("git", "diff", "--cached", "-U0", "--", filePath)
	if err != nil {
		return "", fmt.Errorf("git diff -U0 %s: %w", filePath, err)
	}
	return out, nil
}

// GetDiffU0 returns the -U0 diff for multiple files in one call
func GetDiffU0(filePaths []string) (string, error) {
	if len(filePaths) == 0 {
		return "", nil
	}
	args := append([]string{"diff", "--cached", "-U0", "--"}, filePaths...)
	out, err := execCommand("git", args...)
	if err != nil {
		return "", fmt.Errorf("git diff -U0: %w", err)
	}
	return out, nil
}

// GetWordDiff returns an inline word-level diff for a set of files
// Output format: timeout: [-30s-]{+60s+} — shows only changed values
func GetWordDiff(filePaths []string) (string, error) {
	if len(filePaths) == 0 {
		return "", nil
	}
	args := append([]string{"diff", "--cached", "--word-diff=plain", "-U0", "--"}, filePaths...)
	out, err := execCommand("git", args...)
	if err != nil {
		return "", fmt.Errorf("git diff --word-diff: %w", err)
	}
	return out, nil
}

// GetDiffStat returns the --stat block for a specific set of files
func GetDiffStat(filePaths []string) (string, error) {
	if len(filePaths) == 0 {
		return "", nil
	}
	args := append([]string{"diff", "--cached", "--stat", "--"}, filePaths...)
	out, err := execCommand("git", args...)
	if err != nil {
		return "", fmt.Errorf("git diff --stat: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) > 1 {
		lines = lines[:len(lines)-1] // drop "N files changed…" summary line
	}
	return strings.Join(lines, "\n"), nil
}

// GetDiffDefault returns the default diff (with context) for files
func GetDiffDefault(filePaths []string) (string, error) {
	if len(filePaths) == 0 {
		return "", nil
	}
	args := append([]string{"diff", "--cached", "--"}, filePaths...)
	out, err := execCommand("git", args...)
	if err != nil {
		return "", fmt.Errorf("failed to get diff: %w", err)
	}
	return out, nil
}

// Returns a map of filepath → (added, removed)
func GetNumStatBatch(filePaths []string) (map[string]LineCountPair, error) {
	if len(filePaths) == 0 {
		return make(map[string]LineCountPair), nil
	}

	args := append([]string{"diff", "--cached", "--numstat", "--"}, filePaths...)
	out, err := execCommand("git", args...)
	if err != nil {
		return nil, fmt.Errorf("git diff --numstat: %w", err)
	}

	result := make(map[string]LineCountPair, len(filePaths))
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		result[parts[2]] = LineCountPair{
			Added:   parseNumStatField(parts[0]),
			Removed: parseNumStatField(parts[1]),
		}
	}
	return result, nil
}
