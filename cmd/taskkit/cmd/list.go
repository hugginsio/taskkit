// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	cmdformat "github.com/hugginsio/taskkit/cmd/taskkit/format"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		filters, err := filtersFromFlags(cmd.Context(), cmd)
		if err != nil {
			return err
		}

		tasks, err := c.Get(cmd.Context(), filters...)
		if err != nil {
			return err
		}

		sortBy, _ := cmd.Flags().GetString("sort")
		sortTasks(tasks, sortBy)

		formatter, err := cmdformat.NewFormatterFromCmd(cmd)
		if err != nil {
			return err
		}

		return formatter.Tasks(cmd.OutOrStdout(), tasks)
	},
}

func init() {
	registerFilterFlags(listCmd)
	registerOutputFlag(listCmd)
	registerSortFlag(listCmd)

	rootCmd.AddCommand(listCmd)
}
