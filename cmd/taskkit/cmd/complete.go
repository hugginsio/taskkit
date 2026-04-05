// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"fmt"

	"github.com/hugginsio/taskkit/cmd/taskkit/format"
	"github.com/hugginsio/taskkit/filter"
	"github.com/spf13/cobra"
)

var doneCmd = &cobra.Command{
	Use:     "complete <id>",
	Aliases: []string{"c", "done", "d"},
	Short:   "Mark a task as completed",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		task, err := resolveTask(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		if err := c.Complete(cmd.Context(), task.TaskID.String()); err != nil {
			return fmt.Errorf("done: %w", err)
		}

		results, err := c.Get(cmd.Context(), filter.ID(task.TaskID.String()))
		if err != nil {
			return fmt.Errorf("done: refetch: %w", err)
		}

		if len(results) > 0 {
			task = results[0]
		}

		f, err := format.NewFormatterFromCmd(cmd)
		if err != nil {
			return err
		}

		return f.Task(cmd.OutOrStdout(), task)
	},
}

func init() {
	registerOutputFlag(doneCmd)

	rootCmd.AddCommand(doneCmd)
}
