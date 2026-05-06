package git

import (
	"fmt"
	"os/exec"
)

func UndoLastCommit() error {
	// check if git is installed
	_, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git is not installed")
	}
	// check if we are in a git repository
	cmd := exec.Command("git","rev-parse","--is-inside-work-tree")
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("not a git repository - run inside a repo")
	}
	// check if HEAD exists (repo has at least one commit)
	cmd = exec.Command("git", "rev-parse", "HEAD")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("no commits to undo")
	}
	
	cmd = exec.Command("git","reset","--soft","HEAD~1")

	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}