package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var saveCmd = &cobra.Command{
	Use:   "save [name]",
	Short: "Save the current workspace layout",
	Args:  cobra.ExactArgs(1), // force user to provide exactly 1 argument (the name)
	Run: func(cmd *cobra.Command, args []string) {
		snapshotName := args[0]
		fmt.Printf("saving snapshot: %s\n", snapshotName)

		// TODO: call internal/snapshot.Save(snapshotName) here
	},
}

func init() {
	rootCmd.AddCommand(saveCmd)
}
