package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
	"github.com/fatih/color"
)
var rootCmd = &cobra.Command{
	Use:   "gommit",
	Short: "AI-powered commit message generator",
}
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		color.Red("Error: %s", err)
		os.Exit(1)
	}
}

func init()  {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(promptCmd)
	rootCmd.AddCommand(undoCmd)
	rootCmd.AddCommand(uninstallCmd)
	defaultHelp := rootCmd.HelpFunc()

    rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
        defaultHelp(cmd, args) 
        fmt.Println("\n⭐ Star us: https://github.com/jagadeesh-2006/gommit")
    })
}