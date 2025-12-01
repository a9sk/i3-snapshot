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
	Run: func(cmd *cobra.Command, args []string) {
		snapshotName := args[0]
		fmt.Printf("saving snapshot: %s\n", snapshotName)

		if err := snapshot.Save(snapshotName); err != nil {
			fmt.Printf("error saving snapshot: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(saveCmd)
}
