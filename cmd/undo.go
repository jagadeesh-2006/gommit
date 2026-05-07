package cmd

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/jagadeesh-2006/gommit/internals/git"
)

var undoCmd = &cobra.Command{
	Use:   "undo",
	Aliases: []string{"un"},
	Short: "Undo the last commit",
	Run: func(cmd *cobra.Command, args []string) {
		
		err := git.UndoLastCommit()
		if err != nil {
			color.Red("Error undoing last commit: %s", err)
			return
		}
		color.Green("Last commit undone successfully")
	},
}
