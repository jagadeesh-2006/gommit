package grouping

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jagadeesh-2006/gommit/internals/git"
)

type FileGroup int

const (
	GroupCode   FileGroup = iota // source code files
	GroupConfig                  // configuration, yaml, json, env, dockerfile
	GroupDocs                    // markdown, rst, txt, readme, changelog
	GroupTest                    // test files (*_test.go, *.test.ts, *.spec.js …)
	GroupCI                      // CI/CD pipeline files (.github/, .gitlab-ci, …)
	GroupSkip                    // lock files, binaries, generated files — skip entirely
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

// for one staged file
type FileInfo struct {
	Path   string    // relative path from repo root
	Group  FileGroup // classification result
	Lines  int       // total changed lines (added + removed)
	Status string    // git status: "A" added, "M" modified, "D" deleted, "R" renamed
}

type GroupedFiles struct {
	Files map[FileGroup][]*FileInfo
}

func GetStagedFiles() ([]*FileInfo, error) {
	// Step 1: get file paths and their git status in one call
	nameStatusOut, err := git.GetStagedFilesNameStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to get staged files: %w", err)
	}

	// Step 2: get all line counts in ONE subprocess call instead of N
	numStatOut, err := git.GetStagedNumStat()
	if err != nil {
		numStatOut = ""
	}
	lineCounts := parseNumStatOutput(numStatOut)

	var files []*FileInfo
	for _, line := range strings.Split(strings.TrimSpace(nameStatusOut), "\n") {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		status := parts[0]
		// renames show as "R100\told_path\tnew_path" — take the last field
		path := parts[len(parts)-1]

		ls := lineCounts[path]
		files = append(files, &FileInfo{
			Path:   path,
			Group:  ClassifyFile(path),
			Lines:  ls.Total(),
			Status: status,
		})
	}

	return files, nil
}

// parseNumStatOutput parses `git diff --numstat` output into a map.
// Called internally by GetStagedFiles.
func parseNumStatOutput(output string) map[string]LineStats {
	result := make(map[string]LineStats)
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		result[parts[2]] = LineStats{
			Added:   parseNumStatFieldLocal(parts[0]),
			Removed: parseNumStatFieldLocal(parts[1]),
		}
	}
	return result
}

// parseNumStatFieldLocal converts a numstat field to int.
// Binary files show "-" instead of a number; those map to 0.
func parseNumStatFieldLocal(s string) int {
	if s == "-" {
		return 0
	}
	var n int
	fmt.Sscan(s, &n)
	return n
}

func ClassifyFile(filename string) FileGroup {
	lower := strings.ToLower(filename)
	base := strings.ToLower(filepath.Base(filename))
	ext := filepath.Ext(lower)

	// Skip: lock files, binaries, minified files, generated files
	skipExact := []string{
		"package-lock.json", "yarn.lock", "pnpm-lock.yaml",
		"go.sum", "go.work.sum", "cargo.lock",
		"composer.lock", "poetry.lock", "gemfile.lock", "pubspec.lock",
	}
	for _, s := range skipExact {
		if base == s || strings.HasSuffix(lower, s) {
			return GroupSkip
		}
	}
	skipSuffix := []string{".min.js", ".min.css", ".pb.go", ".generated.go"}
	for _, s := range skipSuffix {
		if strings.HasSuffix(lower, s) {
			return GroupSkip
		}
	}
	skipDir := []string{"/vendor/", "/node_modules/", "/dist/", "/build/"}
	for _, d := range skipDir {
		if strings.Contains(lower, d) {
			return GroupSkip
		}
	}

	// Docs
	docExts := []string{".md", ".rst", ".txt", ".adoc", ".mdx"}
	for _, d := range docExts {
		if ext == d {
			return GroupDocs
		}
	}
	docBases := []string{"readme", "changelog", "license", "contributing", "authors", "notice"}
	for _, d := range docBases {
		if strings.Contains(base, d) {
			return GroupDocs
		}
	}

	// Config
	configExts := []string{".yaml", ".yml", ".toml", ".json", ".env", ".ini", ".cfg", ".conf", ".properties", ".xml"}
	for _, c := range configExts {
		if ext == c {
			return GroupConfig
		}
	}
	configBases := []string{"dockerfile", "makefile", ".gitignore", ".dockerignore", ".editorconfig", ".prettierrc", ".eslintrc"}
	for _, c := range configBases {
		if strings.Contains(base, c) {
			return GroupConfig
		}
	}

	// Test
	testPatterns := []string{"_test.", ".test.", ".spec.", "/test/", "/tests/", "__tests__", "testdata/"}
	for _, t := range testPatterns {
		if strings.Contains(lower, t) {
			return GroupTest
		}
	}

	// CI
	ciPatterns := []string{".github/", ".gitlab-ci", "jenkinsfile", ".circleci", ".travis.yml", "bitbucket-pipelines"}
	for _, c := range ciPatterns {
		if strings.Contains(lower, c) {
			return GroupCI
		}
	}

	// Everything else is code
	return GroupCode
}

// GroupFiles organises a flat list of FileInfo into a GroupedFiles map.
func GroupFiles(files []*FileInfo) *GroupedFiles {
	grouped := &GroupedFiles{
		Files: make(map[FileGroup][]*FileInfo),
	}
	for _, file := range files {
		grouped.Files[file.Group] = append(grouped.Files[file.Group], file)
	}
	return grouped
}

// SubGroupCodeByDirectory splits a code file group by their parent directory.
// Returns a map of directory path → files.
func SubGroupCodeByDirectory(files []*FileInfo) map[string][]*FileInfo {
	groups := make(map[string][]*FileInfo)
	for _, file := range files {
		dir := filepath.Dir(file.Path)
		groups[dir] = append(groups[dir], file)
	}
	return groups
}

func BuildStructuredPrompt(group FileGroup, files []*FileInfo) (compressedDiff string, prompt string, err error) {
	if group == GroupSkip || len(files) == 0 {
		return "", "", nil
	}

	filePaths := make([]string, len(files))
	for i, f := range files {
		filePaths[i] = f.Path
	}

	stat, _ := GetShortStat()

	// config files: word-diff gives inline value changes — no further extraction needed
	if group == GroupConfig {
		wordDiff, werr := GetWordDiff(filePaths)
		if werr != nil {
			return "", "", fmt.Errorf("getting word diff: %w", werr)
		}
		compressedDiff = wordDiff
		prompt = buildGroupPrompt(group, files, stat, compressedDiff)
		return compressedDiff, prompt, nil
	}

	// all code-type groups: per-file structured extraction
	lineStats, serr := GetNumStatBatch(filePaths)
	if serr != nil {
		return "", "", fmt.Errorf("getting line stats: %w", serr)
	}

	summaries := make([]FileSummary, 0, len(files))
	for _, f := range files {
		rawDiff, derr := GetFileDiffU0(f.Path)
		if derr != nil {
			ls := lineStats[f.Path]
			summaries = append(summaries, FileSummary{
				Filename:  f.Path,
				IsNew:     f.Status == "A",
				IsDeleted: f.Status == "D",
				Stats:     ls,
			})
			continue
		}

		ls := lineStats[f.Path]
		summary := ExtractFileSummary(
			f.Path,
			rawDiff,
			f.Status == "A",
			f.Status == "D",
			ls,
		)
		summaries = append(summaries, summary)
	}

	compressedDiff = BuildCompressedContext(summaries, 6000)
	prompt = buildGroupPrompt(group, files, stat, compressedDiff)
	return compressedDiff, prompt, nil
}

func buildGroupPrompt(group FileGroup, files []*FileInfo, stat, compressedDiff string) string {
	fileList := buildFileList(files)

	rules := `Rules:
- Return ONLY 3 messages in this exact format:
1. <message>
2. <message>
3. <message>
- Nothing else. No explanations. No preamble.
- Each under 100 characters
- Each from a completely different angle (what changed / why / impact)`

	diffSection := fmt.Sprintf("Diff:\n===START===\n%s\n===END===", compressedDiff)

	switch group {
	case GroupCode:
		return fmt.Sprintf(`You are a git commit message expert.

Summary: %s

Files changed:
%s
%s

Use conventional commit format: feat / fix / refactor / perf / style
The diff shows function signatures and logic changes. Focus on WHAT changed
functionally and WHY — not the file name.

%s`, stat, fileList, rules, diffSection)

	case GroupDocs:
		return fmt.Sprintf(`You are a git commit message expert.

Summary: %s

Files changed:
%s
%s

Always use "docs:" prefix.
Focus on what documentation was updated and why.

%s`, stat, fileList, rules, diffSection)

	case GroupConfig:
		return fmt.Sprintf(`You are a git commit message expert.

Summary: %s

Files changed:
%s
%s

Use "chore:" or "build:" prefix.
Word diff format: [-old value-]{+new value+} — focus on what config value
changed, what it affects, and why it was changed.

%s`, stat, fileList, rules, diffSection)

	case GroupTest:
		return fmt.Sprintf(`You are a git commit message expert.

Summary: %s

Files changed:
%s
%s

Always use "test:" prefix.
Focus on what behaviour was tested and what case was added or fixed.

%s`, stat, fileList, rules, diffSection)

	case GroupCI:
		return fmt.Sprintf(`You are a git commit message expert.

Summary: %s

Files changed:
%s
%s

Always use "ci:" prefix.
Focus on what pipeline step or workflow changed and why.

%s`, stat, fileList, rules, diffSection)

	default:
		return fmt.Sprintf(`Generate 3 conventional commit messages.

%s

%s`, rules, diffSection)
	}
}

// buildFileList formats the file list section of the prompt.
func buildFileList(files []*FileInfo) string {
	var sb strings.Builder
	for _, f := range files {
		status := ""
		switch f.Status {
		case "A":
			status = " [new]"
		case "D":
			status = " [deleted]"
		case "R":
			status = " [renamed]"
		}
		fmt.Fprintf(&sb, "  %s%s (%d lines)\n", f.Path, status, f.Lines)
	}
	return sb.String()
}

// GetDiffForFiles returns the full staged diff for a list of files.
// Prefer GetDiffForGroup or BuildStructuredPrompt for new code.
func GetDiffForFiles(filePaths []string) (string, error) {
	return git.GetDiffDefault(filePaths)
}
