// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"errors"
	"strings"

	"github.com/hugginsio/taskkit/client"
	cmdformat "github.com/hugginsio/taskkit/cmd/taskkit/format"
	"github.com/hugginsio/taskkit/filter"
	"github.com/spf13/cobra"
)

var modifyCmd = &cobra.Command{
	Use:     "modify <id> [description]",
	Aliases: []string{"mod", "m"},
	Short:   "Modify a task",
	Args:    cobra.MinimumNArgs(1),
	GroupID: CommandGroupTask,
	RunE: func(cmd *cobra.Command, args []string) error {
		task, err := resolveTask(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		mods, err := modificationsFromFlags(cmd)
		if err != nil {
			return err
		}

		if len(args) > 1 {
			mods = append(mods, client.SetDescription(strings.Join(args[1:], " ")))
		}

		if len(mods) <= 0 {
			return errors.New("modify: no modifications to apply")
		} else {
			if err := c.Modify(cmd.Context(), task.TaskID.String(), mods...); err != nil {
				return err
			}
		}

		// --blocking: this task blocks other tasks, so each target is blocked by this task.
		if cmd.Flags().Changed("blocking") {
			ids, _ := cmd.Flags().GetStringSlice("blocking")
			for _, id := range ids {
				dep, err := resolveTask(cmd.Context(), id)
				if err != nil {
					return err
				}

				if err := c.Modify(cmd.Context(), dep.TaskID.String(), client.AddBlockedBy(task.TaskID.String())); err != nil {
					return err
				}
			}
		}

		tasks, err := c.Get(cmd.Context(), filter.ID(task.TaskID.String()))
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
	registerDateFlags(modifyCmd)
	registerDependencyFlags(modifyCmd)
	registerOutputFlag(modifyCmd)
	registerRemoveFlags(modifyCmd)
	registerTaskFlags(modifyCmd)

	rootCmd.AddCommand(modifyCmd)
}
