package main

import (
	"fmt"

	"github.com/a9sk/i3-snapshot/internal/snapshot"
	"github.com/spf13/cobra"
)

var saveCmd = &cobra.Command{
	Use:   "save [name]",
	Short: "Save the current workspace layout",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		snapshotName := args[0]
		fmt.Printf("saving snapshot: %s\n", snapshotName)

		if err := snapshot.Save(snapshotName); err != nil {
			return fmt.Errorf("error saving snapshot: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(saveCmd)
}
