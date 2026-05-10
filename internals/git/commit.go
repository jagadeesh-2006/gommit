package git

import (
	"os"
	"os/exec"
	"fmt"
)

func Commit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func CommitGroupFiles(filePaths []string, message string) error {
    // step 1 — unstage everything
    if err := exec.Command("git", "reset", "HEAD").Run(); err != nil {
        return fmt.Errorf("failed to unstage: %w", err)
    }

    // step 2 — stage only this group
    stageArgs := append([]string{"add", "--"}, filePaths...)
    if err := exec.Command("git", stageArgs...).Run(); err != nil {
        return fmt.Errorf("failed to stage group files: %w", err)
    }

    // step 3 — commit
    cmd := exec.Command("git", "commit", "-m", message)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}