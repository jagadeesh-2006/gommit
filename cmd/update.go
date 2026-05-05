package cmd

import (
	"fmt"
	"bufio"
	"os"
	"strings"
	"github.com/jagadeesh-2006/gommit/internals/config"
	"github.com/jagadeesh-2006/gommit/internals/ai"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use: "update",
	Aliases: []string{"u"},
	Short: "Update existing Gommit configuration",
	Run: func(cmd *cobra.Command, args []string) {
		if !config.Exists() {
			color.Red(" Config not found — run `gommit init` first")
			return
		}
		cfg, err := config.Load()
		if err != nil {
			color.Red("Error loading configuration: %s", err)
			return
		}
		var choice int
		color.White("Select the setting you want to update:")
		color.White("1. Provider + model + API key")
		color.White("2. Model")
		color.White("3. API Key")
		color.White("4. Commit Style")
		color.White("5. Custom Prompt")
		fmt.Scanln(&choice)
		
		switch choice {

			case 1:
				color.White("Enter new provider (e.g., groq, openai): ")
				fmt.Scanln(&cfg.Provider)

				color.White("Enter new API key: ")
				fmt.Scanln(&cfg.APIKey)

				p := ai.GetProvider(cfg.Provider, cfg.APIKey, "", cfg.CommitStyle,cfg.CustomPrompt)
				if p == nil {
					color.Red("Invalid provider. Choose: anthropic, groq, openai")
					return
				}

				color.White("Fetching models...")
				models, err := p.FetchModels()
				if err != nil {
					color.Red("Error fetching models: %s", err)
					return
				}

				color.White("Available models:")
				for i, m := range models {
					color.White("%d. %s", i+1, m)
				}

				color.White("Select a model number: ")
				var choice int
				fmt.Scanln(&choice)

				if choice < 1 || choice > len(models) {
					color.Red("Invalid choice")
					return
				}

				cfg.Model = models[choice-1]


			case 2:
				// fetch models for the current provider
				p := ai.GetProvider(cfg.Provider, cfg.APIKey, "", cfg.CommitStyle, cfg.CustomPrompt)

				// fmt.Println("Fetching models...")
				models, err := p.FetchModels()
				if err != nil {
					color.Red("Error fetching models: %s", err)
					return
				}

				color.White("Available models:")
				for i, m := range models {
					color.White("%d. %s", i+1, m)
				}

				color.White("Select a model number: ")
				var choice int
				fmt.Scanln(&choice)

				if choice < 1 || choice > len(models) {
					color.Red("Invalid choice")
					return
				}
				cfg.Model = models[choice-1]


			case 3:
				color.White("Enter new API key: ")
				fmt.Scanln(&cfg.APIKey)


			case 4:
				color.White("Enter new commit style (e.g., conventional, gitmoji , simple , any): ")
				fmt.Scanln(&cfg.CommitStyle)


			case 5:
				reader := bufio.NewReader(os.Stdin)
				color.White("Enter new custom prompt (or leave blank for default): ")
				input, _ := reader.ReadString('\n')
				cfg.CustomPrompt = strings.TrimSpace(input)
				
			default:
				color.Red("Invalid choice")
				return
		}

		err = config.Save(cfg)
		if err != nil {
			color.Red("Error saving configuration: %s", err)
			return
		}

		color.Green("Configuration updated successfully!")
	},
}