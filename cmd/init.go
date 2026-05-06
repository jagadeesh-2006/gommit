package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/jagadeesh-2006/gommit/internals/ai"
	"github.com/jagadeesh-2006/gommit/internals/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:     "init",
	Aliases: []string{"i"},
	Short:   "Setup gommit for the first time",
	Run: func(cmd *cobra.Command, args []string) {
		color.Cyan("Setting up gommit for the first time...")

		color.White("Enter your AI provider (anthropic, groq, openai): ")
		var provider string
		fmt.Scanln(&provider)
		if provider == "" {
			color.Red("Provider cannot be empty")
			return
		}

		color.White("Enter your API key: ")
		var apiKey string
		fmt.Scanln(&apiKey)
		if apiKey == "" {
			color.Red("API key cannot be empty")
			return
		}
		p := ai.GetProvider(provider, apiKey, "", "", "")
		if p == nil {
			color.Red("Invalid provider. Choose: anthropic, groq, openai")
			return
		}

		color.Cyan("Fetching models...")
		models, err := p.FetchModels()
		if err != nil {
			color.Red("Error fetching models: %s", err)
			return
		}

		color.Blue("Available models:")
		for i, m := range models {
			color.Blue("%d. %s", i+1, m)
		}

		color.White("Select a model number: ")
		var choice int
		fmt.Scanln(&choice)

		if choice < 1 || choice > len(models) {
			color.Red("Invalid choice")
			return
		}

		selected := models[choice-1]

		color.White("Commit style (conventional, simple, emoji, any): ")
		var style string
		fmt.Scanln(&style)
		if style == "" {
			style = "conventional"
		}

		cfg := &config.Config{
			Version:      "1",
			Provider:     provider,
			Model:        selected,
			APIKey:       apiKey,
			CommitStyle:  style,
			CustomPrompt: "",
		}

		if err := config.Save(cfg); err != nil {
			color.Red("Error saving config: %s", err)
			return
		}

		color.Green("Config saved! Run `gommit run` in any repo.")
	},
}
