package cmd

import (
	"bufio"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/jagadeesh-2006/gommit/internals/config"
	"github.com/spf13/cobra"
)

var promptCmd = &cobra.Command{
	Use:     "prompt",
	Aliases: []string{"p"},
	Short:   "Test your custom prompt with the current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		if !config.Exists() {
			color.Red("No config exists. Please run 'gommit init' to set up your configuration.")
			return
		}
		cfg, err := config.Load()
		if err != nil {
			color.Red("Error loading configuration: %s", err)
			return
		}
		p := cfg.CustomPrompt
		if p == "" {
			p = "none"
		}
		color.Blue("Current prompt:")
		color.White("%s", p)
		color.White("Enter a test prompt (or press Enter to use the current one): ")

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			p = input
			cfg.CustomPrompt = input
		}
		color.Blue("Test prompt:")
		color.White("%s", p)
		err = config.Save(cfg)
		if err != nil {
			color.Red("Error saving configuration: %s", err)
			return
		}
		color.Green("Prompt updated successfully!")
	},
}
