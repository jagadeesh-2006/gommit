package cmd

import (
	"github.com/fatih/color"
	"github.com/jagadeesh-2006/gommit/internals/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Aliases: []string{"cfg"},
	Short: "Configure Gommit settings",
	Run: func(cmd *cobra.Command, args []string) {
		if !config.Exists(){
			color.Red("No existing configuration found. Please run 'gommit init' to set up your configuration.")
			return
		}	
		cfg, err := config.Load()
		if err != nil {
			color.Red("Error loading configuration: %s", err)
			return
		}
		color.White("Current Configuration:")
		color.White("Provider:     %s", cfg.Provider)
		color.White("Model: %s", cfg.Model)
		color.White("API Key: %s", cfg.APIKey[:8]+"********") 
		color.White("Commit Style: %s", cfg.CommitStyle)
		prompt := cfg.CustomPrompt
		if prompt == "" {
			prompt = "none"
		}
		color.White("Custom Prompt: %s", prompt)
	},
}