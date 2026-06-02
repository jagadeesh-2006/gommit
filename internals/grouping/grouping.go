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

	//  1. SKIP
	skipExact := []string{
		"package-lock.json", "yarn.lock", "pnpm-lock.yaml", "bun.lockb",
		"go.sum", "go.work.sum", "cargo.lock",
		"composer.lock", "poetry.lock", "gemfile.lock", "pubspec.lock",
		"packages.lock.json", "paket.lock",
	}
	for _, s := range skipExact {
		if base == s || strings.HasSuffix(lower, s) {
			return GroupSkip
		}
	}

	skipSuffix := []string{
		".min.js", ".min.css",
		".pb.go", ".pb.ts", ".pb.js",
		".generated.go", ".generated.ts",
		"_generated.go", "_generated.ts",
		".g.dart",
		"_test.mocks.dart",
		".freezed.dart",
		".gr.dart",
		".pyc", ".pyo", ".pyd",
		".class",
		".o", ".a", ".so", ".dylib", ".dll",
		".exe", ".bin",
		".map",
	}
	for _, s := range skipSuffix {
		if strings.HasSuffix(lower, s) {
			return GroupSkip
		}
	}

	skipDir := []string{
		"/vendor/", "/node_modules/",
		"/dist/", "/build/", "/out/",
		"/.next/", "/.nuxt/", "/.svelte-kit/",
		"/__pycache__/", "/.pytest_cache/",
		"/.gradle/", "/target/",
		"/.dart_tool/",
		"/coverage/", "/.nyc_output/",
		"/storybook-static/",
	}
	for _, d := range skipDir {
		if strings.Contains(lower, d) {
			return GroupSkip
		}
	}

	skipBase := []string{
		".ds_store", "thumbs.db", "desktop.ini",
		".swp", ".swo",
		"*.orig",
	}
	for _, s := range skipBase {
		if base == s || strings.HasSuffix(base, s) {
			return GroupSkip
		}
	}

	//  2. CI — before Config (both use .yml/.yaml)
	ciPatterns := []string{
		".github/",
		".gitlab-ci",
		"jenkinsfile",
		".circleci/",
		".travis.yml",
		"bitbucket-pipelines",
		".drone.yml",
		"azure-pipelines.yml",
		".buildkite/",
		"cloudbuild.yaml",
		"appveyor.yml",
		".woodpecker.yml",
		"render.yaml",
		"fly.toml",
		"railway.toml",
		"vercel.json",
		"netlify.toml",
		".github/workflows/",
		".github/actions/",
	}
	for _, c := range ciPatterns {
		if strings.Contains(lower, c) {
			return GroupCI
		}
	}

	//  3. TEST
	testPatterns := []string{
		"_test.",
		".test.",
		".spec.",
		"/test/",
		"/tests/",
		"/__tests__/",
		"/e2e/",
		"/cypress/",
		"/playwright/",
		"cypress.config.",
		"playwright.config.",
		"jest.setup.",
		"vitest.setup.",
		".stories.",
	}
	for _, t := range testPatterns {
		if strings.Contains(lower, t) {
			return GroupTest
		}
	}

	//  4. DOCS
	docExts := []string{".md", ".mdx", ".rst", ".adoc", ".tex"}
	for _, d := range docExts {
		if ext == d {
			return GroupDocs
		}
	}
	docBases := []string{
		"readme", "changelog", "license", "contributing",
		"authors", "notice", "security", "codeowners",
		"code_of_conduct", "funding", "support",
		"history", "credits", "copying",
	}
	for _, d := range docBases {
		if strings.Contains(base, d) {
			return GroupDocs
		}
	}

	//  5. CONFIG
	configExts := []string{
		".yaml", ".yml", ".toml", ".ini",
		".cfg", ".conf", ".properties",
		".env",
	}
	for _, c := range configExts {
		if ext == c {
			return GroupConfig
		}
	}

	configBases := []string{
		"package.json", "composer.json", "pyproject.toml",
		"setup.cfg", "setup.py", "pipfile",
		"gemfile", "podfile", "pubspec.yaml",
		"dockerfile", "makefile", "rakefile", "gruntfile", "gulpfile",
		".dockerignore", "docker-compose",
		".gitignore", ".gitattributes",
		".editorconfig", ".prettierrc", ".prettierignore",
		".eslintrc", ".eslintignore",
		".stylelintrc", ".stylelintignore",
		".babelrc",
		"vite.config", "webpack.config", "rollup.config",
		"next.config", "nuxt.config", "svelte.config",
		"astro.config", "remix.config",
		"tailwind.config", "postcss.config",
		"jest.config", "vitest.config",
		"karma.config",
		"tsconfig", "jsconfig",
		"nx.json", "turbo.json", "lerna.json",
		"pnpm-workspace.yaml",
		".nvmrc", ".node-version",
		".python-version",
		".ruby-version",
		".tool-versions",
		"renovate.json",
		".releaserc",
		"sonar-project.properties",
		"codecov.yml",
		".goreleaser.yaml",
	}
	for _, c := range configBases {
		if strings.Contains(base, c) {
			return GroupConfig
		}
	}

	// 6. CODE — everything else
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

	// Config files: word-diff gives inline value changes — no further extraction needed.
	if group == GroupConfig {
		wordDiff, werr := GetWordDiff(filePaths)
		if werr != nil {
			return "", "", fmt.Errorf("getting word diff: %w", werr)
		}
		compressedDiff = wordDiff
		prompt = buildGroupPrompt(group, files, stat, compressedDiff)
		return compressedDiff, prompt, nil
	}

	// All code-type groups: per-file structured extraction.
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
	diffSection := fmt.Sprintf("Diff:\n===START===\n%s\n===END===", compressedDiff)

	// Shared rules block — the KEY change is requiring each message to cover
	// ALL changes holistically, not one change per message.
	rules := `Output rules (STRICT):
- Return EXACTLY 3 commit messages, numbered 1. 2. 3.
- EACH message must describe ALL of the changes together in a single line — not one change per message.
- Think of the 3 messages as 3 different ways to say the same thing: what changed overall.
- Vary the angle: message 1 = what was done, message 2 = why / intent, message 3 = user-facing impact or scope.
- Every message must be self-contained and make sense without reading the others.
- Length: 50 to 100 characters. No bullet points. No line breaks inside a message.
- No preamble, no explanation, no blank lines between messages.
- Format exactly:
1. <message>
2. <message>
3. <message>`

	switch group {
	case GroupCode:
		return fmt.Sprintf(`You are an expert at writing git commit messages.

Your job: read ALL the changes in the diff below and write 3 commit messages that each summarise the ENTIRE changeset.

Do NOT write one message per change or per file. Every message must cover all changes together.

Summary: %s

Files changed:
%s
Use conventional commit format (feat / fix / refactor / perf / style / chore).
Focus on the overall functional purpose of all the changes combined.
The type prefix should reflect the dominant intent (e.g. if most changes add behaviour, use feat:).

%s

%s`, stat, fileList, rules, diffSection)

	case GroupDocs:
		return fmt.Sprintf(`You are an expert at writing git commit messages.

Your job: read ALL the documentation changes below and write 3 commit messages summarising the ENTIRE changeset.

Do NOT write one message per file or per section updated. Every message covers all changes together.

Summary: %s

Files changed:
%s
Always use "docs:" prefix.
Focus on what documentation was updated overall and the reason for the update.

%s

%s`, stat, fileList, rules, diffSection)

	case GroupConfig:
		return fmt.Sprintf(`You are an expert at writing git commit messages.

Your job: read ALL the config changes below and write 3 commit messages summarising the ENTIRE changeset.

Do NOT write one message per config key or per file. Every message covers all config changes together.

Summary: %s

Files changed:
%s
Use "chore:" or "build:" prefix.
Word diff format: [-old value-]{+new value+} shows what changed.
Focus on the combined effect of all config changes (e.g. "increase timeouts and enable retry logic").

%s

%s`, stat, fileList, rules, diffSection)

	case GroupTest:
		return fmt.Sprintf(`You are an expert at writing git commit messages.

Your job: read ALL the test changes below and write 3 commit messages summarising the ENTIRE changeset.

Do NOT write one message per test case or per file. Every message covers all test changes together.

Summary: %s

Files changed:
%s
Always use "test:" prefix.
Focus on what behaviour is now covered across all the test changes combined.

%s

%s`, stat, fileList, rules, diffSection)

	case GroupCI:
		return fmt.Sprintf(`You are an expert at writing git commit messages.

Your job: read ALL the CI/CD changes below and write 3 commit messages summarising the ENTIRE changeset.

Do NOT write one message per pipeline step or per file. Every message covers all CI changes together.

Summary: %s

Files changed:
%s
Always use "ci:" prefix.
Focus on the combined effect of all pipeline/workflow changes.

%s

%s`, stat, fileList, rules, diffSection)

	default:
		return fmt.Sprintf(`You are an expert at writing git commit messages.

Write 3 commit messages that each cover ALL of the changes below as a single line cohesive summary.

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
func GetDiffForFiles(filePaths []string) (string, error) {
	return git.GetDiffDefault(filePaths)
}