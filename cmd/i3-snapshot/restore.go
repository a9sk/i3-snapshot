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
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		fmt.Printf("restoring snapshot: %s\n", name)

		if err := snapshot.Restore(name); err != nil {
			return fmt.Errorf("error restoring snapshot: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}
