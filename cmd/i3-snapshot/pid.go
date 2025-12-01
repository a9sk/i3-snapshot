package main

import (
	"fmt"
	"strconv"

	"github.com/a9sk/i3-snapshot/internal/proc"
	"github.com/spf13/cobra"
)

var pidCmd = &cobra.Command{
	Use:   "pid [pid]",
	Short: "Show the command used to start a given PID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid pid %q: %w", args[0], err)
		}

		command, err := proc.GetCommandFromPID(pid)
		if err != nil {
			return err
		}

		fmt.Println(command)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pidCmd)
}
