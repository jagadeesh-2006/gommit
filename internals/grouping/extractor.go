// extractor.go in grouping package

package grouping

import (
    "strings"
)

// ExtractSignalLines filters diff to only high value lines
// drops: blank lines, lone braces, closing brackets
// keeps: logic, conditions, returns, assignments, calls
func ExtractSignalLines(diff string, maxChars int) string {
    var kept []string
    charCount := 0

    for _, line := range strings.Split(diff, "\n") {
        // always keep file headers — AI needs to know which file
        if strings.HasPrefix(line, "diff --git") ||
            strings.HasPrefix(line, "---") ||
            strings.HasPrefix(line, "+++") ||
            strings.HasPrefix(line, "@@") {
            kept = append(kept, line)
            charCount += len(line)
            continue
        }

        // only process added/removed lines
        if !strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "-") {
            continue // skip context lines (U0 removes most anyway)
        }

        content := strings.TrimSpace(line[1:])

        // drop pure noise
        if isNoiseLine(content) {
            continue
        }

        // budget check
        if charCount+len(line) > maxChars {
            kept = append(kept, "... (truncated)")
            break
        }

        kept = append(kept, line)
        charCount += len(line)
    }

    return strings.Join(kept, "\n")
}

func isNoiseLine(content string) bool {
    // blank lines
    if content == "" {
        return true
    }
    // lone braces/brackets
    if len(content) <= 2 && strings.ContainsAny(content, "{}[]()") {
        return true
    }
    // closing brace with optional comment
    trimmed := strings.TrimSpace(content)
    if trimmed == "}" || trimmed == "}," || trimmed == ");" || trimmed == ")," {
        return true
    }
    return false
}