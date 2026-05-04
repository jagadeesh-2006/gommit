package cmd

import (
	"fmt"
	"github.com/jagadeesh-2006/gommit/internals/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure Gommit settings",
	Run: func(cmd *cobra.Command, args []string) {
		if !config.Exists(){
			fmt.Println("No existing configuration found. Please run 'gommit init' to set up your configuration.")
			return
		}	
		cfg, err := config.Load()
		if err != nil {
			fmt.Println("Error loading configuration:", err)
			return
		}
		fmt.Println("Current Configuration:")
		fmt.Printf("Provider:     %s\n", cfg.Provider)
		fmt.Printf("Model: %s\n", cfg.Model)
		fmt.Printf("API Key: %s\n", cfg.APIKey[:8]+"********") 
		fmt.Printf("Commit Style: %s\n", cfg.CommitStyle)
		prompt := cfg.CustomPrompt
		if prompt == "" {
			prompt = "none"
		}
		fmt.Printf("Custom Prompt: %s\n", prompt)
	},
}