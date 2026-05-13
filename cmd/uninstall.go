package cmd

import (
    "fmt"
    "github.com/fatih/color"
    "github.com/jagadeesh-2006/gommit/internals/config"
    "github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
    Use:   "uninstall",
	Aliases: []string{"remove"},
    Short: "Remove gommit configuration",
    Run: func(cmd *cobra.Command, args []string) {
        color.White("Are you sure you want to remove gommit config? (y/n): ")
        var confirm string
        fmt.Scanln(&confirm)
        if confirm != "y" {
            color.Yellow("Uninstall cancelled.")
            return
        }
        err := config.Remove()
        if err != nil {
            color.Red(" Error removing config: %s", err)
            return
        }
        color.Green(" gommit config removed successfully.")
        color.White("Binary is still installed. To fully remove run:")
        color.White("  Windows: del %USERPROFILE%\\go\\bin\\gommit.exe")
        color.White("  Mac/Linux: rm /usr/local/bin/gommit")
    },
}