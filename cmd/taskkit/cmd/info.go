// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	cmdformat "github.com/hugginsio/taskkit/cmd/taskkit/format"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:     "info <id>",
	Aliases: []string{"i"},
	Short:   "Show details for a task",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		task, err := resolveTask(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		task.History, err = c.GetHistory(cmd.Context(), task.TaskID.String())
		if err != nil {
			return err
		}

		formatter, err := cmdformat.NewFormatterFromCmd(cmd)
		if err != nil {
			return err
		}

		return formatter.Task(cmd.OutOrStdout(), task)
	},
}

func init() {
	registerOutputFlag(infoCmd)
	rootCmd.AddCommand(infoCmd)
}
