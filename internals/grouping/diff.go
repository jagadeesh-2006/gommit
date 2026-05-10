package grouping

import (
    "fmt"
    "os/exec"
    "strings"
)

// Stage 1 — cheapest, just file names
func GetStagedFileNames() ([]string, error) {
    out, err := exec.Command("git", "diff", "--cached", "--name-only").Output()
    if err != nil {
        return nil, fmt.Errorf("not in a git repo or nothing staged")
    }
    if len(strings.TrimSpace(string(out))) == 0 {
        return nil, fmt.Errorf("no staged changes found")
    }
    var files []string
    for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
        if line = strings.TrimSpace(line); line != "" {
            files = append(files, line)
        }
    }
    return files, nil
}

// Stage 2 — overall header, always include in prompt
func GetShortStat() (string, error) {
    out, err := exec.Command("git", "diff", "--cached", "--shortstat").Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(out)), nil
}

// Stage 3 — per file stats for a group
func GetStatForGroup(filePaths []string) (string, error) {
    args := append([]string{"diff", "--cached", "--stat", "--"}, filePaths...)
    out, err := exec.Command("git", args...).Output()
    if err != nil {
        return "", err
    }
    // remove last summary line "N files changed..."
    lines := strings.Split(strings.TrimSpace(string(out)), "\n")
    if len(lines) > 1 {
        lines = lines[:len(lines)-1]
    }
    return strings.Join(lines, "\n"), nil
}

// Stage 4 — actual diff, U0 removes surrounding context lines
// biggest token saving — removes 3 unchanged lines around each change
func GetDiffU0(filePaths []string) (string, error) {
    args := append([]string{"diff", "--cached", "-U0", "--"}, filePaths...)
    out, err := exec.Command("git", args...).Output()
    if err != nil {
        return "", err
    }
    return string(out), nil
}

// Stage 5 — word diff for config files only
// shows: timeout: [-30s-]{+60s+} instead of full line diff
func GetWordDiff(filePaths []string) (string, error) {
    args := append([]string{"diff", "--cached", "--word-diff=plain", "-U0", "--"}, filePaths...)
    out, err := exec.Command("git", args...).Output()
    if err != nil {
        return "", err
    }
    return string(out), nil
}

// GetDiffForGroup — picks right diff command per group
func GetDiffForGroup(group FileGroup, filePaths []string) (string, error) {
    switch group {
    case GroupConfig:
        // word diff for config — much more compact
        return GetWordDiff(filePaths)
    case GroupSkip:
        return "", nil
    default:
        // U0 for everything else — no surrounding context
        return GetDiffU0(filePaths)
    }
}