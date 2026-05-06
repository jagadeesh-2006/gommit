//1. check config exists → if not, error
// 2. load config
// 3. get staged diff
// 4. get provider using config
// 5. generate commit message
// 6. show options y/e/r/n
// 7. if y → run git commit -m "message"

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/jagadeesh-2006/gommit/internals/ai"
	"github.com/jagadeesh-2006/gommit/internals/config"
	"github.com/jagadeesh-2006/gommit/internals/git"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:     "run",
	Aliases: []string{"r"},
	Short:   "Generate and commit messages",
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("Running gommit...")
		// step 1 - load config
		if !config.Exists() {
			color.Red(" Config not found — run `gommit init` first")
			return
		}

		// fmt.Println("Config exists ✓")
		cfg, err := config.Load()
		if err != nil {
			color.Red("Error loading config: %s", err)
			return
		}

		// step 2 - get diff
		diff, err := git.Diff()
		if err != nil {
			color.Red("Error getting diff: %s", err)
			return
		}
		// sensitive data check
		if containsSensitiveData(diff) {
			color.Yellow("Sensitive data detected in diff (passwords, keys, tokens)")
			color.Yellow("  This diff will be sent to: %s", cfg.Provider)
			fmt.Print("  Continue anyway? (y/n): ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "y" {
				color.Yellow("Commit cancelled.")
				return
			}
		}
		// fmt.Println("Got diff ")
		// step 3 - get provider
		provider := ai.GetProvider(cfg.Provider, cfg.APIKey, cfg.Model, cfg.CommitStyle, cfg.CustomPrompt)
		if provider == nil {
			color.Red("Error getting AI provider:")
			return
		}
		// fmt.Println("Got AI provider ")
		// step 4 - generate commit message
		message, err := provider.GenerateCommitMessage(diff)
		if err != nil {
			color.Red("Error generating commit message: %s", err)
			return
		}
		color.Cyan("Suggested commit message: %s", message)
		var choice string
		color.White("Do you want to use this commit message? (y/e/r/n): ")
		fmt.Scanln(&choice)

		switch choice {
		case "y":
			err := git.Commit(message)
			if err != nil {
				color.Red("Error committing changes: %s", err)
				return
			}
			color.Green("Changes committed with message: %s", message)
		case "e":
			reader := bufio.NewReader(os.Stdin)
			color.White("Enter your commit message: ")
			customMessage, _ := reader.ReadString('\n')
			customMessage = strings.TrimSpace(customMessage)
			color.White("Do you want to use this custom commit message? (y/n): ")
			var newChoice string
			fmt.Scanln(&newChoice)
			if newChoice == "y" {
				err := git.Commit(customMessage)
				if err != nil {
					color.Red("Error committing changes: %s", err)
					return
				}
				color.Green("Changes committed with message: %s", customMessage)
			} else {
				color.Yellow("Commit aborted.")
			}

		case "r":
			color.Cyan("Regenerating commit message...")
			newMessage, err := provider.GenerateCommitMessage(diff)
			if err != nil {
				color.Red("Error generating commit message: %s", err)
				return
			}
			color.Cyan("Regenerated commit message: %s", newMessage)
			color.White("Do you want to use this commit message? (y/n): ")
			var newChoice string
			fmt.Scanln(&newChoice)
			if newChoice == "y" {
				err := git.Commit(newMessage)
				if err != nil {
					color.Red("Error committing changes: %s", err)
					return
				}
				color.Green("Changes committed with message: %s", newMessage)
			} else {
				color.Yellow("Commit aborted.")
			}
		case "n":
			color.Yellow("Commit aborted.")
		default:
			color.Red("Invalid choice. Commit aborted.")
		}
	},
}

func containsSensitiveData(diff string) bool {
	patterns := []string{
		"password", "secret", "api_key", "token",
		"private_key", "access_key", "bearer",
		"credential", "auth", "passwd",
	}
	lowerDiff := strings.ToLower(diff)
	for _, pattern := range patterns {
		if strings.Contains(lowerDiff, pattern) {
			return true
		}
	}
	return false
}
