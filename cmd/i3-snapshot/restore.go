package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore [name]",
	Short: "Restore a saved layout",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		snapshotName := args[0]
		fmt.Printf("restoring snapshot: %s\n", snapshotName)

		// TODO: call internal/snapshot.Restore(snapshotName) here
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}
