package git

import (
	"fmt"
	"os/exec"
)


// run the git diff command and return the output as a string
func Diff() (string , error) {
	// run git diff using os.exec
	cmd := exec.Command("git","diff","--staged")
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