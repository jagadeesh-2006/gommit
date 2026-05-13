package grouping

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type LineStats struct {
	Added   int
	Removed int
}

// Total returns the combined number of changed lines.
func (ls LineStats) Total() int { return ls.Added + ls.Removed }

// chunkSummary is the extracted, signal-filtered content of one diff chunk.
type chunkSummary struct {
	FunctionContext string   // extracted from "@@ ... @@ funcName" header
	Added           []string // high/medium-signal added lines
	Removed         []string // high/medium-signal removed lines
}

// FileSummary is the fully structured extraction result for one file.
// It is what gets formatted into the LLM prompt — never raw diff text.
type FileSummary struct {
	Filename   string
	IsNew      bool
	IsDeleted  bool
	Stats      LineStats
	Signatures []string      // changed fn/class/type declarations, deduped across chunks
	Chunks      []chunkSummary // per-chunk extracted content, capped at maxchunksPerFile
}

const (
	maxchunksPerFile = 8   // cap chunk count per file to avoid runaway output
	maxLineLen      = 120 // characters — longer lines get truncated with "…"
	maxContextLen   = 100 // characters — function context from @@ header
)


// rechunkHeader parses the trailing function context out of a @@ line.
// e.g. "@@ -45,6 +45,12 @@ func createSession(..." → captures "func createSession(..."
var rechunkHeader = regexp.MustCompile(`@@[^@]+@@\s*(.*)`)

// sigPatterns: per-extension patterns that match function/class/type declarations.
// Only applied to changed (+/-) lines. The goal is the signature, not the body.
var sigPatterns = map[string]*regexp.Regexp{
	".go": regexp.MustCompile(
		`^[+-]\s*(func\s+(\([^)]+\)\s+)?\w+` +
			`|type\s+\w+\s+(struct|interface|func)` +
			`|var\s+\w+|const\s+\w+)`),
	".ts": regexp.MustCompile(
		`^[+-]\s*((export\s+)?(default\s+)?(async\s+)?function\s+\w+` +
			`|(export\s+)?(abstract\s+|declare\s+)?(class|interface|type|enum)\s+\w+` +
			`|(export\s+)?(const|let)\s+\w+\s*[=:])`),
	".tsx": regexp.MustCompile(
		`^[+-]\s*((export\s+)?(default\s+)?(async\s+)?function\s+\w+` +
			`|(export\s+)?(const|let)\s+\w+\s*[=:]` +
			`|(export\s+)?(class|interface|type)\s+\w+)`),
	".js": regexp.MustCompile(
		`^[+-]\s*((export\s+)?(async\s+)?function\s+\w+` +
			`|(module\.exports|exports\.\w+)\s*=` +
			`|(const|let|var)\s+\w+\s*=\s*(async\s+)?(function|\())`),
	".jsx": regexp.MustCompile(
		`^[+-]\s*((export\s+)?(default\s+)?(async\s+)?function\s+\w+` +
			`|(export\s+)?(const|let)\s+\w+\s*=)`),
	".py": regexp.MustCompile(
		`^[+-]\s*((async\s+)?def\s+\w+|class\s+\w+)`),
	".rs": regexp.MustCompile(
		`^[+-]\s*((pub(\(crate\))?\s+)?(async\s+)?fn\s+\w+` +
			`|(pub(\(crate\))?\s+)?(struct|enum|trait|impl|type)\s+\w+)`),
	".java": regexp.MustCompile(
		`^[+-]\s*((public|private|protected|static|final|abstract|\s+){1,5}[\w<>\[\]]+\s+\w+\s*\(` +
			`|(public|private|protected)?\s*(abstract\s+)?(class|interface|enum)\s+\w+)`),
	".cs": regexp.MustCompile(
		`^[+-]\s*(public|private|protected|internal|static|virtual|override|abstract|\s){1,5}` +
			`[\w<>\[\]]+\s+\w+\s*\(`),
	".rb": regexp.MustCompile(
		`^[+-]\s*(def\s+(self\.)?\w+|(class|module)\s+\w+)`),
	".php": regexp.MustCompile(
		`^[+-]\s*((public|private|protected|static|\s){0,3}function\s+\w+` +
			`|(abstract\s+)?(class|interface|trait)\s+\w+)`),
	".kt": regexp.MustCompile(
		`^[+-]\s*(fun|class|data\s+class|sealed\s+class|interface|object|enum\s+class)\s+\w+`),
	".swift": regexp.MustCompile(
		`^[+-]\s*(public|private|internal|open|fileprivate|\s)*` +
			`(override\s+)?(func|class|struct|enum|protocol|extension)\s+\w+`),
	".vue": regexp.MustCompile(
		`^[+-]\s*((export\s+)?(default\s+)?(async\s+)?function\s+\w+` +
			`|(export\s+)?(const|let)\s+\w+\s*[=:]` +
			`|defineComponent|defineProps|defineEmits)`),

	".svelte": regexp.MustCompile(
		`^[+-]\s*((export\s+)?(async\s+)?function\s+\w+` +
			`|(export\s+)?(const|let)\s+\w+\s*=)`),

	".dart": regexp.MustCompile(
		`^[+-]\s*((class|mixin|extension|enum)\s+\w+` +
			`|(Future|void|String|int|bool|Widget|List|Map)\s+\w+\s*\(` +
			`|\w+\s+\w+\s*\()`),

	".c": regexp.MustCompile(
		`^[+-]\s*[\w\s\*]+\s+\w+\s*\([^)]*\)\s*\{?`),

	".cpp": regexp.MustCompile(
		`^[+-]\s*(class|struct|template|namespace|[\w:<>\s\*]+\s+\w+\s*\([^)]*\))`),

	".scala": regexp.MustCompile(
		`^[+-]\s*((def|val|var|class|object|trait|case\s+class|sealed)\s+\w+)`),

	".ex": regexp.MustCompile(
		`^[+-]\s*(def\s+\w+|defp\s+\w+|defmodule\s+\w+|defmacro\s+\w+)`),

	".exs": regexp.MustCompile(
		`^[+-]\s*(def\s+\w+|defp\s+\w+|defmodule\s+\w+)`),
}

// reHighSignal: lines that carry actual logic and intent — always keep.
var reHighSignal = regexp.MustCompile(
    `(?i)(\bif\b|\belse\b|\bswitch\b|\bcase\b|` +
        `\breturn\b|\bthrow\b|\bpanic\(|\bfatal\b|` +          
        `\berr\b\s*!=?\s*nil|\berrors?\.(New|Wrap|Is|As)\b|fmt\.Errorf|` +
        `\b(make|append|new)\s*\(|` +
        `context\.|http\.|time\.|os\.|sql\.|json\.|grpc\.|` +      
        `\bawait\b|\basync\b|\.then\(|\.catch\(|\.finally\(|` +   // async control flow
        `\bchan\b|\bgo\b\s+func|` +
        `\bdefer\b|` +                                            //  defer —  in Go
        `\byield\b|\byield\*\b|` +                               //  yield — Python/JS generators
        `\braise\b|` +                                           //  raise — Python exceptions
        `\bassert\b|` +                                          //  assert — testing/validation
        `\b(SELECT|INSERT|UPDATE|DELETE|WHERE|JOIN|FROM)\b|` +   // already have most
        `\btry\b|\bcatch\b|\bfinally\b|` +                      //  try/catch
        `useState\(|useEffect\(|useCallback\(|useMemo\(|` +     // React hooks
        `\.subscribe\(|\.pipe\(|\.map\(|\.filter\(|` +          // RxJS/streams
        `redis\.|kafka\.|rabbit\.|nats\.)`,                      // message queues/cache
)

// reMediumSignal: declarations and imports — useful context, keep when there's room.
var reMediumSignal = regexp.MustCompile(
    `(?i)(\bimport\b|\brequire\(|\bfrom\b\s+['"]|` +
        `\bconst\b|\blet\b|\bvar\b|:=|` +      
        `\btype\b\s+\w+|\binterface\b|\bstruct\b\s*\{|` +
        `\bfunc\b|\bdef\b|\bclass\b|\bfn\b|` +
        `\benum\b|` +                        //  enum
        `\bprotocol\b|` +                    //  Swift protocol
        `\btrait\b|` +                       //  Rust/PHP trait
        `\bimpl\b|` +                        //  Rust impl
        `\bextension\b|` +                   //  Swift/Dart extension
        `\bdecorator\b|@\w+\()`,             //  decorators — Python/TS
)

// reNoise: lines with zero semantic value — always drop.
var reNoise = regexp.MustCompile(
    `^\s*$` +
        `|^\s*[{}()\[\]]\s*$` +         // lone braces and brackets 
        `|^\s*[{}()\[\],;]\s*(\/\/.*|#.*)?$` +       
        `|^\s*\/\/\s*$` +       // empty comment line
        `|^\s*#\s*$` +          // empty Python/Ruby comment
        `|^\s*\/\*\s*$` +       // opening block comment
        `|^\s*\*\/\s*$` +       // closing block comment
        `|^\s*\*\s*$`,          // middle of block comment
)

// ---------------------------------------------------------------------------
// Core extraction
// ---------------------------------------------------------------------------

// ExtractFileSummary is the primary entry point for structured diff extraction.
// It takes the raw -U0 diff for a single file and returns a FileSummary
// with function contexts, signatures, and signal-filtered chunk content.
//
// Parameters:
//   - filename:  the file path (used for extension-based sig pattern lookup)
//   - rawDiff:   the raw output of `git diff --cached -U0 -- <filename>`
//   - isNew:     true if this is a newly created file (status "A")
//   - isDeleted: true if this file was deleted (status "D")
//   - stats:     added/removed line counts from --numstat
func ExtractFileSummary(filename, rawDiff string, isNew, isDeleted bool, stats LineStats) FileSummary {
	summary := FileSummary{
		Filename:  filename,
		IsNew:     isNew,
		IsDeleted: isDeleted,
		Stats:     stats,
	}

	if strings.TrimSpace(rawDiff) == "" {
		return summary
	}

	ext := strings.ToLower(filepath.Ext(filename))
	pattern := sigPatterns[ext]
	seenSigs := make(map[string]bool)

	blocks := splitIntochunks(rawDiff)
	for i, block := range blocks {
		if i >= maxchunksPerFile {
			break
		}

		// extract signatures from this chunk and deduplicate globally across chunks
		for _, sig := range extractSignatures(block, pattern, seenSigs) {
			summary.Signatures = append(summary.Signatures, sig)
		}

		chunk := processchunk(block)
		// only append chunks that have actual content
		if len(chunk.Added) > 0 || len(chunk.Removed) > 0 || chunk.FunctionContext != "" {
			summary.Chunks = append(summary.Chunks, chunk)
		}
	}

	return summary
}

// splitIntochunks breaks a raw diff into individual chunk blocks.
// Each block starts at an @@ line and includes all lines until the next @@ or EOF.
func splitIntochunks(rawDiff string) []string {
	var blocks []string
	var current strings.Builder

	for _, line := range strings.Split(rawDiff, "\n") {
		if strings.HasPrefix(line, "@@") && current.Len() > 0 {
			blocks = append(blocks, current.String())
			current.Reset()
		}
		current.WriteString(line)
		current.WriteByte('\n')
	}
	if current.Len() > 0 {
		blocks = append(blocks, current.String())
	}
	return blocks
}

// processchunk extracts the function context and signal lines from one chunk block.
func processchunk(block string) chunkSummary {
	chunk := chunkSummary{}

	for _, line := range strings.Split(block, "\n") {
		// @@ line: extract function context from the trailing section
		if strings.HasPrefix(line, "@@") {
			if m := rechunkHeader.FindStringSubmatch(line); len(m) > 1 {
				ctx := strings.TrimSpace(m[1])
				// trim at opening brace — we only want the signature, not the body
				if idx := strings.Index(ctx, " {"); idx > 0 {
					ctx = ctx[:idx]
				}
				ctx = strings.TrimSpace(ctx)
				if len(ctx) > maxContextLen {
					ctx = ctx[:maxContextLen] + "…"
				}
				chunk.FunctionContext = ctx
			}
			continue
		}

		if len(line) == 0 {
			continue
		}

		isAdded := line[0] == '+'
		isRemoved := line[0] == '-'
		if !isAdded && !isRemoved {
			continue // context line — skip (U0 removes most anyway)
		}

		content := strings.TrimSpace(line[1:])

		// drop empty and pure noise
		if content == "" || reNoise.MatchString(content) {
			continue
		}

		// very short non-signal lines (lone operators, single chars) are noise
		if len(content) < 4 && !reHighSignal.MatchString(content) {
			continue
		}

		// keep high signal always; medium signal if it has substance
		if !reHighSignal.MatchString(content) && !reMediumSignal.MatchString(content) {
			// fallthrough: still include the line — it might be an
			// assignment or expression not matched by the patterns
			// but skip very short unrecognised lines
			if len(content) < 8 {
				continue
			}
		}

		if len(content) > maxLineLen {
			content = content[:maxLineLen] + "…"
		}

		if isAdded {
			chunk.Added = append(chunk.Added, content)
		} else {
			chunk.Removed = append(chunk.Removed, content)
		}
	}

	return chunk
}

// extractSignatures pulls function/class/type declaration lines from a chunk block.
// seen is shared across all chunks for the same file to avoid duplicates.
func extractSignatures(block string, pattern *regexp.Regexp, seen map[string]bool) []string {
	if pattern == nil {
		return nil
	}

	var sigs []string
	for _, line := range strings.Split(block, "\n") {
		if len(line) == 0 || (line[0] != '+' && line[0] != '-') {
			continue
		}
		if !pattern.MatchString(line) {
			continue
		}
		clean := strings.TrimSpace(line[1:])
		if clean == "" {
			continue
		}
		if len(clean) > maxLineLen {
			clean = clean[:maxLineLen] + "…"
		}
		if !seen[clean] {
			seen[clean] = true
			sigs = append(sigs, clean)
		}
	}
	return sigs
}

// FormatFileSummary renders a FileSummary into the compact structured string
// that is embedded in the LLM prompt. It is intentionally terse —
// signatures first, then per-chunk logic lines with function context.
func FormatFileSummary(s FileSummary) string {
	var sb strings.Builder

	// file header line
	switch {
	case s.IsNew:
		fmt.Fprintf(&sb, "### %s [NEW +%d lines]\n", s.Filename, s.Stats.Added)
	case s.IsDeleted:
		fmt.Fprintf(&sb, "### %s [DELETED -%d lines]\n", s.Filename, s.Stats.Removed)
	default:
		fmt.Fprintf(&sb, "### %s [+%d -%d]\n", s.Filename, s.Stats.Added, s.Stats.Removed)
	}

	// signatures: structural overview of what changed at the declaration level
	if len(s.Signatures) > 0 {
		sb.WriteString("Changed:\n")
		for _, sig := range s.Signatures {
			fmt.Fprintf(&sb, "  %s\n", sig)
		}
	}

	// per-chunk logic lines, grouped by function context
	for _, chunk := range s.Chunks {
		if len(chunk.Added) == 0 && len(chunk.Removed) == 0 {
			continue
		}
		if chunk.FunctionContext != "" {
			fmt.Fprintf(&sb, "in `%s`:\n", chunk.FunctionContext)
		}
		for _, line := range chunk.Removed {
			fmt.Fprintf(&sb, "  - %s\n", line)
		}
		for _, line := range chunk.Added {
			fmt.Fprintf(&sb, "  + %s\n", line)
		}
	}

	sb.WriteByte('\n')
	return sb.String()
}

// ScoreFile assigns an importance score to a file for token budget ordering.
// Higher score = included first. Called by BuildCompressedContext.
func ScoreFile(s FileSummary) float64 {
	score := 0.0
	lower := strings.ToLower(s.Filename)

	// new and deleted files signal the most important changes
	if s.IsNew {
		score += 4.0
	}
	if s.IsDeleted {
		score += 2.0
	}

	// files with extracted signatures carry clear structural signal
	if len(s.Signatures) > 0 {
		score += 2.0
	}

	// entrypoints and primary handlers matter more for the commit message
	entrypoints := []string{
		"main.go", "main.py", "index.ts", "index.js", "index.jsx", "index.tsx",
		"app.go", "app.py", "server.go", "server.ts",
		"cmd/", "handler", "controller", "router", "routes", "middleware",
		 "service", "repository", "store", "api/", "core/", "domain/", 
	}
	for _, ep := range entrypoints {
		if strings.Contains(lower, ep) {
			score += 2.0
			break
		}
	}

	// small focused changes are signal-dense
	switch total := s.Stats.Total(); {
	case total <= 15:
		score += 2.0
	case total <= 30:
		score += 1.0
	case total <= 60:
		score += 0.5
	}

	// test files contribute less to the commit message intent
	testIndicators := []string{"_test.", ".test.", ".spec.", "/test/", "/tests/", "__tests__"}
	for _, t := range testIndicators {
		if strings.Contains(lower, t) {
			score -= 1.0
			break
		}
	}

	return score
}

func BuildCompressedContext(summaries []FileSummary, maxChars int) string {
	// sort highest score first so the most important files always fit
	sort.Slice(summaries, func(i, j int) bool {
		return ScoreFile(summaries[i]) > ScoreFile(summaries[j])
	})

	var sb strings.Builder
	used := 0

	for _, s := range summaries {
		block := FormatFileSummary(s)
		if used+len(block) > maxChars {
			// over budget: one-liner so the AI knows the file exists
			oneliner := fmt.Sprintf("// %s [+%d -%d — detail omitted: budget]\n",
				s.Filename, s.Stats.Added, s.Stats.Removed)
			sb.WriteString(oneliner)
			used += len(oneliner)
			continue
		}
		sb.WriteString(block)
		used += len(block)
	}

	return sb.String()
}

// isNoiseLine reports whether a content string (already stripped of +/- prefix)
func isNoiseLine(content string) bool {
	return content == "" || reNoise.MatchString(content)
}

// ExtractSignalLines is the original flat-line extractor, kept for fallback use.
// Prefer ExtractFileSummary + FormatFileSummary for new code.
func ExtractSignalLines(diff string, maxChars int) string {
	var kept []string
	charCount := 0

	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "diff --git") ||
			strings.HasPrefix(line, "---") ||
			strings.HasPrefix(line, "+++") ||
			strings.HasPrefix(line, "@@") {
			kept = append(kept, line)
			charCount += len(line)
			continue
		}

		if !strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "-") {
			continue
		}

		content := strings.TrimSpace(line[1:])
		if isNoiseLine(content) {
			continue
		}

		if charCount+len(line) > maxChars {
			kept = append(kept, "... (truncated)")
			break
		}

		kept = append(kept, line)
		charCount += len(line)
	}

	return strings.Join(kept, "\n")
}