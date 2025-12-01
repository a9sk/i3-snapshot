package main

import (
	"fmt"
	"os"

	"github.com/a9sk/i3-snapshot/internal/i3"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "i3-snapshot",
	Short: "A layout and session manager for i3wm",
	// power users can use this, hope i remember to document it somewhere
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
	Long: `i3-snapshot allows you to save the current state of your i3 workspace
(including window layouts and running commands) and restore them later.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	i3.Connect()
}
