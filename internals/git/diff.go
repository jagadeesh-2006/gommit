package git

import (
	"fmt"
	"os/exec"
)


// run the git diff command and return the output as a string
func Diff() (string , error) {
	// check if git is installed
	_, err := exec.LookPath("git")
	if err != nil {
		return "", fmt.Errorf("git is not installed")
	}
	// check if we are in a git repository
	cmd := exec.Command("git","rev-parse","--is-inside-work-tree")
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("not a git repository - run inside a repo")
	}	
	// run git diff using os.exec
	cmd = exec.Command("git","diff","--staged")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	// if output is empty 
	if len(output) == 0 {
		return "", fmt.Errorf("no staged changes found")
	}
	return string(output), nil
}