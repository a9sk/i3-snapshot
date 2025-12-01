package main

import (
	"github.com/a9sk/i3-snapshot/internal/i3"
	"github.com/spf13/cobra"
)

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Print the current i3 workspace tree",
	Run: func(cmd *cobra.Command, args []string) {
		i3.PrintTree()
	},
}

func init() {
	rootCmd.AddCommand(treeCmd)
}
