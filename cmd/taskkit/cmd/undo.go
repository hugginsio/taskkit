// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"errors"
	"fmt"

	"github.com/hugginsio/taskkit/client"
	"github.com/spf13/cobra"
)

var undoCmd = &cobra.Command{
	Use:     "undo",
	Short:   "Reverse the most recent operation",
	Args:    cobra.NoArgs,
	GroupID: CommandGroupTask,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := c.Undo(cmd.Context()); err != nil {
			if errors.Is(err, client.ErrNothingToUndo) {
				fmt.Fprintln(cmd.OutOrStdout(), "Nothing to undo.")
				return nil
			}

			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Undone.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(undoCmd)
}
