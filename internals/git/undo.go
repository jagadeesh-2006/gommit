package git

import (
	"os/exec"
)

func UndoLastCommit() error {
	// Validate git environment
	if err := CheckGitInstalled(); err != nil {
		return err
	}
	if err := CheckGitRepository(); err != nil {
		return err
	}
	if err := CheckRepositoryHasCommits(); err != nil {
		return err
	}

	// Reset to previous commit, keeping changes staged
	return ResetSoftLastCommit()
}

// ResetSoftLastCommit resets HEAD to the previous commit, keeping changes staged
func ResetSoftLastCommit() error {
	cmd := exec.Command("git", "reset", "--soft", "HEAD~1")
	return cmd.Run()
}
