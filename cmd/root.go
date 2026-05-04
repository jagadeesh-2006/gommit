package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
)
var rootCmd = &cobra.Command{
	Use:   "gommit",
	Short: "AI-powered commit message generator",
}
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func init()  {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(promptCmd)
}