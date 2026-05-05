//1. check config exists → if not, error
// 2. load config
// 3. get staged diff
// 4. get provider using config
// 5. generate commit message
// 6. show options y/e/r/n
// 7. if y → run git commit -m "message"

package cmd

import (
	"fmt"
	"bufio"
	"os"
	"strings"
	"github.com/jagadeesh-2006/gommit/internals/config"
	"github.com/jagadeesh-2006/gommit/internals/git"
	"github.com/jagadeesh-2006/gommit/internals/ai"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Aliases: []string{"r"},
	Short: "Generate and commit messages",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running gommit...")	
		// step 1 - load config
		if !config.Exists() {
			// fmt.Println("Config not found.")
			fmt.Println("Please run `gommit init` first to set up your configuration.")
			return
		}
		
		// fmt.Println("Config exists ✓")
		cfg, err := config.Load()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		// step 2 - get diff
		diff, err := git.Diff()
		if err != nil {
			fmt.Println("Error getting diff:", err)
			return
		}
		// fmt.Println("Got diff ")
		// step 3 - get provider
		provider:= ai.GetProvider(cfg.Provider, cfg.APIKey, cfg.Model, cfg.CommitStyle , cfg.CustomPrompt)
		if provider == nil {
			fmt.Println("Error getting AI provider:")
			return
		}
		// fmt.Println("Got AI provider ")
		// step 4 - generate commit message
		message, err := provider.GenerateCommitMessage(diff)
		if err != nil {
			fmt.Println("Error generating commit message:", err)
			return
		}
		fmt.Println("suggested commit:", message)
		var choice string
		fmt.Print("Do you want to use this commit message? (y/e/r/n): ")
		fmt.Scanln(&choice)

		switch choice {
			case "y":
				err := git.Commit(message)
				if err != nil {
					fmt.Println("Error committing changes:", err)
					return
				}
				fmt.Println("Changes committed with message:", message)
			case "e":
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("Enter your commit message: ")
				customMessage, _ := reader.ReadString('\n')
				customMessage = strings.TrimSpace(customMessage)
				err := git.Commit(customMessage)
				if err != nil {
					fmt.Println("Error committing changes:", err)
					return
				}
			case "r":
				fmt.Println("Regenerating commit message...")
				newMessage, err := provider.GenerateCommitMessage(diff)
				if err != nil {
					fmt.Println("Error generating commit message:", err)
					return
				}
				fmt.Println("Regenerated commit message:", newMessage)
				fmt.Print("Do you want to use this commit message? (y/n): ")
				var newChoice string
				fmt.Scanln(&newChoice)
				if newChoice == "y" {
					err := git.Commit(newMessage)
					if err != nil {
						fmt.Println("Error committing changes:", err)
						return
					}
					fmt.Println("Changes committed with message:", newMessage)
				} else {
					fmt.Println("Commit aborted.")
				}
			case "n":
				fmt.Println("Commit aborted.")
			default:
				fmt.Println("Invalid choice. Commit aborted.")
		}
	},
}
