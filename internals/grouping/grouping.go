package grouping

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type FileGroup int

const (
	GroupCode FileGroup = iota
	GroupConfig
	GroupDocs
	GroupTest
	GroupCI
	GroupSkip
)

func (g FileGroup) String() string {
	switch g {
	case GroupCode:
		return "📦 Code"
	case GroupConfig:
		return "⚙️  Config"
	case GroupDocs:
		return "📄 Docs"
	case GroupTest:
		return "🧪 Test"
	case GroupCI:
		return "🔧 CI"
	case GroupSkip:
		return "🚫 Skip"
	default:
		return "Unknown"
	}
}

type FileInfo struct {
	Path  string
	Group FileGroup
	Lines int // added + deleted lines
}

type GroupedFiles struct {
	Files map[FileGroup][]*FileInfo
}

// GetStagedFiles returns all staged files with their line counts
func GetStagedFiles() ([]*FileInfo, error) {
	// Get staged files with diff stats
	cmd := exec.Command("git", "diff", "--cached", "--name-status")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get staged files: %w", err)
	}

	if len(output) == 0 {
		return nil, fmt.Errorf("no staged changes found")
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var files []*FileInfo

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		// Skip deleted files for now
		if parts[0] == "D" {
			continue
		}

		path := parts[1]
		// Get line count for this file
		lineCount := getFileLineCount(path)

		files = append(files, &FileInfo{
			Path:  path,
			Group: ClassifyFile(path),
			Lines: lineCount,
		})
	}

	return files, nil
}

// ClassifyFile categorizes a file into a group
func ClassifyFile(filename string) FileGroup {
	filename = strings.ToLower(filename)

	// Skip patterns first — most important
	skipPatterns := []string{
		"package-lock.json", "yarn.lock", "pnpm-lock.yaml",
		"go.sum", "go.work.sum", "cargo.lock",
		".min.js", ".min.css",
	}
	for _, pattern := range skipPatterns {
		if strings.Contains(filename, pattern) {
			return GroupSkip
		}
	}

	ext := strings.ToLower(filepath.Ext(filename))
	base := strings.ToLower(filepath.Base(filename))

	// Docs
	docExts := []string{".md", ".txt", ".rst", ".adoc"}
	docFiles := []string{"readme", "changelog", "license", "contributing"}
	for _, d := range docExts {
		if ext == d {
			return GroupDocs
		}
	}
	for _, d := range docFiles {
		if strings.Contains(base, d) {
			return GroupDocs
		}
	}

	// Config
	configExts := []string{".yaml", ".yml", ".toml", ".json", ".env", ".ini", ".cfg"}
	configFiles := []string{"dockerfile", "makefile", ".gitignore", ".dockerignore"}
	for _, c := range configExts {
		if ext == c {
			return GroupConfig
		}
	}
	for _, c := range configFiles {
		if strings.Contains(base, c) {
			return GroupConfig
		}
	}

	// Test
	if strings.Contains(filename, "_test.") ||
		strings.Contains(filename, ".test.") ||
		strings.Contains(filename, "/test/") ||
		strings.Contains(filename, "/tests/") {
		return GroupTest
	}

	// CI
	if strings.Contains(filename, ".github/") ||
		strings.Contains(filename, ".gitlab-ci") ||
		strings.Contains(filename, "jenkinsfile") ||
		strings.Contains(filename, ".circleci") {
		return GroupCI
	}

	// Everything else is code
	return GroupCode
}

// GroupFiles organizes files into logical groups
func GroupFiles(files []*FileInfo) *GroupedFiles {
	grouped := &GroupedFiles{
		Files: make(map[FileGroup][]*FileInfo),
	}

	for _, file := range files {
		grouped.Files[file.Group] = append(grouped.Files[file.Group], file)
	}

	return grouped
}

// SubGroupCodeByDirectory splits large code groups by directory
func SubGroupCodeByDirectory(files []*FileInfo) map[string][]*FileInfo {
	groups := make(map[string][]*FileInfo)
	for _, file := range files {
		dir := filepath.Dir(file.Path)
		groups[dir] = append(groups[dir], file)
	}
	return groups
}

// GetDiffForFiles returns the diff for a list of files
func GetDiffForFiles(filePaths []string) (string, error) {
	args := []string{"diff", "--cached"}
	args = append(args, filePaths...)
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get diff: %w", err)
	}
	return string(output), nil
}

// Helper: get line count for a staged file
func getFileLineCount(path string) int {
	cmd := exec.Command("git", "diff", "--cached", "--numstat", path)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	parts := strings.Fields(string(output))
	if len(parts) < 2 {
		return 0
	}

	added := countLines(parts[0])
	deleted := countLines(parts[1])
	return added + deleted
}

func countLines(s string) int {
	if s == "-" {
		return 0
	}
	var count int
	fmt.Sscanf(s, "%d", &count)
	return count
}
func BuildPromptForGroup(group FileGroup, files []*FileInfo, stat string, diff string) string {
    // build file context
    fileList := ""
    for _, f := range files {
        fileList += fmt.Sprintf("  %s (+%d lines)\n", f.Path, f.Lines)
    }

    // extract only signal lines, cap at 4000 chars
    compressedDiff := ExtractSignalLines(diff, 4000)

    rules := `Rules:
- Return ONLY 3 messages in this exact format:
1. <message>
2. <message>
3. <message>
- Nothing else. No explanations.
- Each under 100 characters
- Each completely different angle`

    switch group {
    case GroupCode:
        return fmt.Sprintf(`You are a git commit message expert.

Summary: %s

Files changed:
%s
%s

Use conventional format: feat/fix/refactor/perf/style
Focus on WHAT changed functionally and WHY.

Diff:
===START===
%s
===END===`, stat, fileList, rules, compressedDiff)

    case GroupDocs:
        return fmt.Sprintf(`You are a git commit message expert.

Summary: %s

Files changed:
%s
%s

Always use "docs:" prefix.
Focus on what documentation changed.

Diff:
===START===
%s
===END===`, stat, fileList, rules, compressedDiff)

    case GroupConfig:
        return fmt.Sprintf(`You are a git commit message expert.

Summary: %s

Files changed:
%s
%s

Use "chore:" or "build:" prefix.
Word diff format: [-old value-]{+new value+}
Focus on what config value changed and impact.

Diff:
===START===
%s
===END===`, stat, fileList, rules, compressedDiff)

    case GroupTest:
        return fmt.Sprintf(`You are a git commit message expert.

Summary: %s

Files changed:
%s
%s

Always use "test:" prefix.
Focus on what was tested.

Diff:
===START===
%s
===END===`, stat, fileList, rules, compressedDiff)

    case GroupCI:
        return fmt.Sprintf(`You are a git commit message expert.

Summary: %s

Files changed:
%s
%s

Always use "ci:" prefix.
Focus on what pipeline changed.

Diff:
===START===
%s
===END===`, stat, fileList, rules, compressedDiff)

    default:
        return fmt.Sprintf(`Generate 3 commit messages.
%s

Diff:
===START===
%s
===END===`, rules, compressedDiff)
    }
}