// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	cmdformat "github.com/hugginsio/taskkit/cmd/taskkit/format"
	"github.com/spf13/cobra"
)

var explainCmd = &cobra.Command{
	Use:     "explain <id>",
	Aliases: []string{"ex"},
	Short:   "Show urgency calculation for a task",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		task, err := resolveTask(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		task.UrgencyBreakdown = c.UrgencyBreakdown(task)

		formatter, err := cmdformat.NewFormatterFromCmd(cmd)
		if err != nil {
			return err
		}

		return formatter.UrgencyBreakdown(cmd.OutOrStdout(), task)
	},
}

func init() {
	registerOutputFlag(explainCmd)
	rootCmd.AddCommand(explainCmd)
}
