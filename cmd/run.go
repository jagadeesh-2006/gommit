package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"regexp"
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
    // only scan added lines
    addedLines := []string{}
    for _, line := range strings.Split(diff, "\n") {
        if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
            addedLines = append(addedLines, line)
        }
    }
    addedContent := strings.Join(addedLines, "\n")

    patterns := []string{
        // assignment patterns with actual values
        `(?i)(password|passwd|secret|api_key|apikey|access_key)\s*[:=]\s*["']?[a-zA-Z0-9+/]{8,}`,
        // known API key formats
        `sk-[a-zA-Z0-9]{20,}`,
        `gsk_[a-zA-Z0-9]{20,}`,
        `AIza[a-zA-Z0-9]{20,}`,
        `sk-ant-[a-zA-Z0-9]{20,}`,
        // bearer tokens
        `(?i)Bearer\s+[a-zA-Z0-9\-._~+/]{20,}`,
        // private keys
        `-----BEGIN (RSA |EC )?PRIVATE KEY-----`,
    }

    for _, pattern := range patterns {
        matched, _ := regexp.MatchString(pattern, addedContent)
        if matched {
            return true
        }
    }
    return false
}
