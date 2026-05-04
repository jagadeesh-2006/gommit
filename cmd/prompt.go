package cmd 

import (
    "bufio"
    "os"
    "strings"
	"fmt"
	"github.com/jagadeesh-2006/gommit/internals/config"
	"github.com/spf13/cobra"
)


var promptCmd =& cobra.Command{
	Use: "prompt",
	Aliases: []string{"p"},
	Short: "Test your custom prompt with the current configuration",
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
		p := cfg.CustomPrompt
		if p == "" {
			p = "none"
		}
		fmt.Printf("Current prompt:\n%s\n", p)
		fmt.Print("Enter a test prompt (or press Enter to use the current one): ")
		
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			p = input
			cfg.CustomPrompt = input
		}
		fmt.Printf("Test prompt:\n%s\n", p)
		err = config.Save(cfg)
		if err != nil {
			fmt.Println("Error saving configuration:", err)
			return
		}
		fmt.Println("Prompt updated successfully!")
	},
}