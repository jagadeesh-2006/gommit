package git

import (
	"fmt"
	"os/exec"
)

// execCommand is a helper that executes a git command and returns output as string
func execCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	return string(output), err
}

// LineCountPair represents added and removed line counts
type LineCountPair struct {
	Added   int
	Removed int
}

// Total returns the total changed lines
func (lc LineCountPair) Total() int {
	return lc.Added + lc.Removed
}

// parseNumStatField converts a numstat field to int
// Binary files show "-" instead of a number; those map to 0
func parseNumStatField(s string) int {
	if s == "-" {
		return 0
	}
	var n int
	fmt.Sscan(s, &n)
	return n
}
