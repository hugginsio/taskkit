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

var logCmd = &cobra.Command{
	Use:     "log <description>",
	Aliases: []string{"l"},
	Short:   "Record a task that is already completed",
	Args:    cobra.MinimumNArgs(1),
	GroupID: CommandGroupTask,
	RunE: func(cmd *cobra.Command, args []string) error {
		task := &taskkit.Task{
			Description: strings.Join(args, " "),
			Status:      taskkit.StatusCompleted,
		}

		mods, err := modificationsFromFlags(cmd)
		if err != nil {
			return err
		}

		added, err := c.Add(cmd.Context(), task)
		if err != nil {
			return err
		}

		if len(mods) > 0 {
			if err := c.Modify(cmd.Context(), added.TaskID.String(), mods...); err != nil {
				return err
			}
		}

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
	registerDateFlags(logCmd)
	registerDependencyFlags(logCmd)
	registerOutputFlag(logCmd)
	registerTaskFlags(logCmd)

	rootCmd.AddCommand(logCmd)
}
