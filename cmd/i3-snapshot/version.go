package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of i3-snapshot",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("i3-snapshot %s\ncommit %s\nbuilt at %s\n", version, commit, date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
