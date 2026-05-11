package git

import (
	"fmt"
	"os/exec"
)

func CheckGitInstalled() error {
	_, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git is not installed")
	}
	return nil
}

func CheckGitRepository() error {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("not a git repository - run inside a repo")
	}
	return nil
}

func CheckRepositoryHasCommits() error {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("no commits to undo")
	}
	return nil
}
