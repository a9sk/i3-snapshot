package main

import (
	"fmt"

	"github.com/a9sk/i3-snapshot/internal/snapshot"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore [name]",
	Short: "Restore a previously saved workspace layout",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		fmt.Printf("restoring snapshot: %s\n", name)

		if err := snapshot.Restore(name); err != nil {
			fmt.Printf("error restoring snapshot: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}
