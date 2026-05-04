package cmd

import (
	"fmt"
	"bufio"
	"os"
	"strings"
	"github.com/jagadeesh-2006/gommit/internals/config"
	"github.com/jagadeesh-2006/gommit/internals/ai"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use: "update",
	Aliases: []string{"u"},
	Short: "Update existing Gommit configuration",
	Run: func(cmd *cobra.Command, args []string) {
		if !config.Exists() {
			fmt.Println("No existing configuration found. Please run 'gommit init' to set up your configuration.")
			return
		}
		cfg, err := config.Load()
		if err != nil {
			fmt.Println("Error loading configuration:", err)
			return
		}
		var choice int
		fmt.Println("Select the setting you want to update:")
		fmt.Println("1. Provider + model + API key")
		fmt.Println("2. Model")
		fmt.Println("3. API Key")
		fmt.Println("4. Commit Style")
		fmt.Println("5. Custom Prompt")
		fmt.Scanln(&choice)
		
		switch choice {

			case 1:
				fmt.Print("Enter new provider (e.g., groq, openai): ")
				fmt.Scanln(&cfg.Provider)

				fmt.Print("Enter new API key: ")
				fmt.Scanln(&cfg.APIKey)

				p := ai.GetProvider(cfg.Provider, cfg.APIKey, "", cfg.CommitStyle,cfg.CustomPrompt)
				if p == nil {
					fmt.Println("Invalid provider. Choose: anthropic, groq, openai")
					return
				}

				fmt.Println("Fetching models...")
				models, err := p.FetchModels()
				if err != nil {
					fmt.Println("Error fetching models:", err)
					return
				}

				fmt.Println("Available models:")
				for i, m := range models {
					fmt.Printf("%d. %s\n", i+1, m)
				}

				fmt.Print("Select a model number: ")
				var choice int
				fmt.Scanln(&choice)

				if choice < 1 || choice > len(models) {
					fmt.Println("Invalid choice")
					return
				}

				cfg.Model = models[choice-1]


			case 2:
				// fetch models for the current provider
				p := ai.GetProvider(cfg.Provider, cfg.APIKey, "", cfg.CommitStyle, cfg.CustomPrompt)

				// fmt.Println("Fetching models...")
				models, err := p.FetchModels()
				if err != nil {
					fmt.Println("Error fetching models:", err)
					return
				}

				fmt.Println("Available models:")
				for i, m := range models {
					fmt.Printf("%d. %s\n", i+1, m)
				}

				fmt.Print("Select a model number: ")
				var choice int
				fmt.Scanln(&choice)

				if choice < 1 || choice > len(models) {
					fmt.Println("Invalid choice")
					return
				}
				cfg.Model = models[choice-1]


			case 3:
				fmt.Print("Enter new API key: ")
				fmt.Scanln(&cfg.APIKey)


			case 4:
				fmt.Print("Enter new commit style (e.g., conventional, gitmoji , simple , any): ")
				fmt.Scanln(&cfg.CommitStyle)


			case 5:
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("Enter new custom prompt (or leave blank for default): ")
				input, _ := reader.ReadString('\n')
				cfg.CustomPrompt = strings.TrimSpace(input)
				
			default:
				fmt.Println("Invalid choice")
				return
		}

		err = config.Save(cfg)
		if err != nil {
			fmt.Println("Error saving configuration:", err)
			return
		}

		fmt.Println("Configuration updated successfully!")
	},
}