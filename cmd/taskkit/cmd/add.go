// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"strings"

	"github.com/hugginsio/taskkit"
	"github.com/hugginsio/taskkit/client"
	cmdformat "github.com/hugginsio/taskkit/cmd/taskkit/format"
	"github.com/hugginsio/taskkit/filter"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     "add <description>",
	Aliases: []string{"a"},
	Short:   "Add a new task",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		task := &taskkit.Task{
			Description: strings.Join(args, " "),
		}

		mods, err := modificationsFromFlags(cmd)
		if err != nil {
			return err
		}

		added, err := c.Add(cmd.Context(), task)
		if err != nil {
			return err
		}

		// Apply field mods and --blocked dependencies in one transaction.
		if len(mods) > 0 {
			if err := c.Modify(cmd.Context(), added.TaskID.String(), mods...); err != nil {
				return err
			}
		}

		// --blocking: the new task blocks other tasks, so each target is blocked by added.
		if cmd.Flags().Changed("blocking") {
			ids, _ := cmd.Flags().GetStringSlice("blocking")
			for _, id := range ids {
				dep, err := resolveTask(cmd.Context(), id)
				if err != nil {
					return err
				}

				if err := c.Modify(cmd.Context(), dep.TaskID.String(), client.AddBlockedBy(added.TaskID.String())); err != nil {
					return err
				}
			}
		}

		// Re-fetch to get the fully hydrated, final state.
		tasks, err := c.Get(cmd.Context(), filter.ID(added.TaskID.String()))
		if err != nil {
			return err
		}

		formatter, err := cmdformat.NewFormatterFromCmd(cmd)
		if err != nil {
			return err
		}

		return formatter.Task(cmd.OutOrStdout(), tasks[0])
	},
}

func init() {
	registerDateFlags(addCmd)
	registerDependencyFlags(addCmd)
	registerOutputFlag(addCmd)
	registerTaskFlags(addCmd)

	rootCmd.AddCommand(addCmd)
}
