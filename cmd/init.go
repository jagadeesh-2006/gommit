package cmd

import (
	"fmt"
	"bufio"
	"os"
	"strings"
	"github.com/jagadeesh-2006/gommit/internals/ai"
	"github.com/jagadeesh-2006/gommit/internals/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Aliases: []string{"i"},
	Short: "Setup gommit for the first time",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Setting up gommit for the first time...")

		fmt.Print("Enter your AI provider (anthropic, groq, openai): ")
		var provider string
		fmt.Scanln(&provider)

		fmt.Print("Enter your API key: ")
		var apiKey string
		fmt.Scanln(&apiKey)

		p := ai.GetProvider(provider, apiKey, "", "","")
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

		selected := models[choice-1]

		fmt.Print("Commit style (conventional, simple, emoji, any): ")
		var style string
		fmt.Scanln(&style)
		if style == "" {
			style = "conventional"
		}

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Custom prompt (optional, press Enter to skip): ")
		customPrompt, _ := reader.ReadString('\n')
		customPrompt = strings.TrimSpace(customPrompt)

		cfg := &config.Config{
			Version:     "1",
			Provider:    provider,
			Model:       selected,
			APIKey:      apiKey,
			CommitStyle: style,
			CustomPrompt: customPrompt,
		}

		if err := config.Save(cfg); err != nil {
			fmt.Println("Error saving config:", err)
			return
		}

		fmt.Println("Config saved! Run `gommit run` in any repo.")
	},
}
