// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/hugginsio/dateparse"
	"github.com/hugginsio/taskkit"
	"github.com/hugginsio/taskkit/client"
	"github.com/hugginsio/taskkit/config"
	"github.com/hugginsio/taskkit/filter"
	"github.com/spf13/cobra"
)

// ensureClient initializes the global client if it is not already open.
// Completion funcs that need database access call this instead of relying on
// PersistentPreRunE, which is skipped for __complete commands.
func ensureClient(ctx context.Context) error {
	if c != nil {
		return nil
	}

	if cfg == nil {
		loaded, err := config.Load()
		if err != nil {
			return err
		}
		cfg = loaded
	}

	cl, err := client.NewClient(ctx, cfg)
	if err != nil {
		return err
	}

	c = cl
	return nil
}

// registerOutputFlag attaches the --output and --columns flags to cmd.
func registerOutputFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("output", "o", "pretty", "output format (pretty, json)")
	cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return []cobra.Completion{
			"pretty",
			"json",
		}, cobra.ShellCompDirectiveNoFileComp
	})

	cmd.Flags().StringSliceP("columns", "c", nil, "e.g. description, deadline (pretty only)")
}

// registerSortFlag attaches the --sort flag to cmd.
func registerSortFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("sort", "z", "urgency", "sort order: urgency, id, modified, wait")
	cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return []cobra.Completion{
			"urgency",
			"id",
			"modified",
			"wait",
		}, cobra.ShellCompDirectiveNoFileComp
	})
}

// sortTasks sorts tasks in-place according to sortBy.
// Recognised values: "urgency" (default, descending), "id" (ascending),
// "modified" (most-recent first), "wait" (earliest wait first).
func sortTasks(tasks []*taskkit.Task, sortBy string) {
	switch sortBy {
	case "id":
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].DisplayID < tasks[j].DisplayID
		})
	case "modified":
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Modified.After(tasks[j].Modified)
		})
	case "wait":
		sort.Slice(tasks, func(i, j int) bool {
			if tasks[i].Wait == nil {
				return false
			}

			if tasks[j].Wait == nil {
				return true
			}

			return tasks[i].Wait.Before(*tasks[j].Wait)
		})
	default: // urgency
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Urgency > tasks[j].Urgency
		})
	}
}

// registerTaskFlags attaches the shared task modifier flags to cmd.
func registerTaskFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("project", "p", "", "project name")

	cmd.Flags().StringP("status", "T", "", "pending, completed, or removed")
	cmd.RegisterFlagCompletionFunc("status", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return []cobra.Completion{
			cobra.Completion(taskkit.StatusPending),
			cobra.Completion(taskkit.StatusCompleted),
			cobra.Completion(taskkit.StatusRemoved),
		}, cobra.ShellCompDirectiveNoFileComp
	})

	cmd.Flags().StringSliceP("tags", "t", nil, "comma separated list of tags")
}

// registerDateFlags attaches date mutation flags to cmd.
func registerDateFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("deadline", "d", "", "date the task should be completed")
	cmd.Flags().StringP("scheduled", "s", "", "date that work should begin")
	cmd.Flags().StringP("wait", "w", "", "date the task becomes visible in reports")
}

// registerDependencyFlags attaches dependency mutation flags to cmd.
func registerDependencyFlags(cmd *cobra.Command) {
	cmd.Flags().StringSliceP("blocked", "B", nil, "tasks blocked by the given task")
	cmd.Flags().StringSliceP("blocking", "b", nil, "tasks blocking the given task")
	cmd.Flags().StringSliceP("unblocked", "u", nil, "remove blocked-by dependencies")
}

// registerRemoveFlags attaches removal flags to cmd (for modify only).
func registerRemoveFlags(cmd *cobra.Command) {
	cmd.Flags().StringSliceP("untag", "U", nil, "tags to remove")
}

// registerFilterFlags attaches query filter flags to cmd.
func registerFilterFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP("status-all", "A", false, "include tasks of any status")
	cmd.Flags().StringP("blocked", "B", "", "tasks blocked by the given task")
	cmd.Flags().StringP("blocking", "b", "", "tasks that are blocking the given task")
	cmd.Flags().StringP("deadline-after", "d", "", "tasks with a deadline after date")
	cmd.Flags().StringP("deadline-before", "D", "", "tasks with a deadline before date")
	cmd.Flags().StringP("modified-after", "m", "", "tasks last modified after date")
	cmd.Flags().StringP("modified-before", "M", "", "tasks last modified before date")
	cmd.Flags().StringP("project", "p", "", "filter by project name")
	cmd.Flags().StringP("scheduled-after", "s", "", "tasks scheduled to start after date")
	cmd.Flags().StringP("scheduled-before", "S", "", "tasks scheduled to start before date")
	cmd.Flags().StringP("wait-after", "w", "", "tasks with a wait date after date")
	cmd.Flags().StringP("wait-before", "W", "", "tasks with a wait date before date")
	cmd.Flags().StringSliceP("status", "T", nil, "filter by statuses (pending, completed, removed)")
	cmd.Flags().StringSliceP("tag", "t", nil, "filter by tags (all must match)")
	cmd.Flags().StringSliceP("tags-any", "a", nil, "filter by tags (any may match)")
}

// filtersFromFlags returns a Filter for each filter flag that was explicitly set.
// If neither --status-any nor --status-all is set, defaults to pending and waiting.
func filtersFromFlags(ctx context.Context, cmd *cobra.Command) ([]filter.Filter, error) {
	var filters []filter.Filter

	statusAll, _ := cmd.Flags().GetBool("status-all")
	switch {
	case statusAll:
		// No status filter — include everything.
	case cmd.Flags().Changed("status"):
		statuses, _ := cmd.Flags().GetStringSlice("status")
		var ss []taskkit.Status
		for _, raw := range statuses {
			s := taskkit.Status(raw)
			switch s {
			case taskkit.StatusPending, taskkit.StatusCompleted, taskkit.StatusRemoved:
			default:
				return nil, fmt.Errorf("unknown status %q: must be pending, completed, or removed", raw)
			}

			ss = append(ss, s)
		}

		filters = append(filters, filter.StatusAny(ss...))
	default:
		filters = append(filters, filter.Status(taskkit.StatusPending))
	}

	if cmd.Flags().Changed("project") {
		v, _ := cmd.Flags().GetString("project")
		filters = append(filters, filter.Project(v))
	}

	if cmd.Flags().Changed("tag") {
		tags, _ := cmd.Flags().GetStringSlice("tag")
		for _, tag := range tags {
			filters = append(filters, filter.HasTag(tag))
		}
	}

	if cmd.Flags().Changed("tags-any") {
		tags, _ := cmd.Flags().GetStringSlice("tags-any")
		tagFilters := make([]filter.Filter, len(tags))
		for i, tag := range tags {
			tagFilters[i] = filter.HasTag(tag)
		}

		filters = append(filters, filter.Or(tagFilters...))
	}

	dateFilters, err := filtersFromDateFlags(cmd)
	if err != nil {
		return nil, err
	}

	filters = append(filters, dateFilters...)

	if cmd.Flags().Changed("blocked-by") {
		raw, _ := cmd.Flags().GetString("blocked-by")
		dep, err := resolveTask(ctx, raw)
		if err != nil {
			return nil, fmt.Errorf("--blocked-by %s: %w", raw, err)
		}

		filters = append(filters, filter.BlockedBy(dep.TaskID.String()))
	}

	if cmd.Flags().Changed("blocking") {
		raw, _ := cmd.Flags().GetString("blocking")
		dep, err := resolveTask(ctx, raw)
		if err != nil {
			return nil, fmt.Errorf("--blocking %s: %w", raw, err)
		}

		filters = append(filters, filter.Blocking(dep.TaskID.String()))
	}

	return filters, nil
}

// filtersFromDateFlags returns a Filter for each date filter flag that was explicitly set.
func filtersFromDateFlags(cmd *cobra.Command) ([]filter.Filter, error) {
	type spec struct {
		flag string
		fn   func(time.Time) filter.Filter
	}

	specs := []spec{
		{"deadline-before", filter.DeadlineBefore},
		{"deadline-after", filter.DeadlineAfter},
		{"scheduled-before", filter.ScheduledBefore},
		{"scheduled-after", filter.ScheduledAfter},
		{"wait-before", filter.WaitBefore},
		{"wait-after", filter.WaitAfter},
		{"modified-before", filter.ModifiedBefore},
		{"modified-after", filter.ModifiedAfter},
	}

	var filters []filter.Filter
	for _, s := range specs {
		if !cmd.Flags().Changed(s.flag) {
			continue
		}

		t, err := parseDate(cmd, s.flag)
		if err != nil {
			return nil, err
		}

		filters = append(filters, s.fn(*t))
	}

	return filters, nil
}

// modificationsFromFlags returns a Modification for each task flag that was
// explicitly set. Flags that were not passed are left untouched.
func modificationsFromFlags(cmd *cobra.Command) ([]client.Modification, error) {
	var mods []client.Modification

	if cmd.Flags().Changed("project") {
		v, _ := cmd.Flags().GetString("project")
		mods = append(mods, client.SetProject(v))
	}

	if cmd.Flags().Changed("tags") {
		tags, _ := cmd.Flags().GetStringSlice("tags")
		for _, tag := range tags {
			mods = append(mods, client.AddTag(tag))
		}
	}

	if cmd.Flags().Changed("status") {
		raw, _ := cmd.Flags().GetString("status")
		s := taskkit.Status(raw)
		switch s {
		case taskkit.StatusPending, taskkit.StatusCompleted, taskkit.StatusRemoved:
		default:
			return nil, fmt.Errorf("unknown status %q: must be pending, completed, or removed", raw)
		}

		mods = append(mods, client.SetStatus(s))
	}

	if cmd.Flags().Changed("deadline") {
		t, err := parseDate(cmd, "deadline")
		if err != nil {
			return nil, err
		}

		mods = append(mods, client.SetDeadline(t))
	}

	if cmd.Flags().Changed("scheduled") {
		t, err := parseDate(cmd, "scheduled")
		if err != nil {
			return nil, err
		}

		mods = append(mods, client.SetScheduled(t))
	}

	if cmd.Flags().Changed("wait") {
		t, err := parseDate(cmd, "wait")
		if err != nil {
			return nil, err
		}

		mods = append(mods, client.SetWait(t))
	}

	if cmd.Flags().Changed("blocked") {
		ids, _ := cmd.Flags().GetStringSlice("blocked")
		for _, id := range ids {
			dep, err := resolveTask(cmd.Context(), id)
			if err != nil {
				return nil, err
			}

			mods = append(mods, client.AddBlockedBy(dep.TaskID.String()))
		}
	}

	if cmd.Flags().Changed("unblocked") {
		ids, _ := cmd.Flags().GetStringSlice("unblocked")
		for _, id := range ids {
			dep, err := resolveTask(cmd.Context(), id)
			if err != nil {
				return nil, err
			}

			mods = append(mods, client.RemoveBlockedBy(dep.TaskID.String()))
		}
	}

	if cmd.Flags().Changed("untag") {
		tags, _ := cmd.Flags().GetStringSlice("untag")
		for _, tag := range tags {
			mods = append(mods, client.RemoveTag(tag))
		}
	}

	return mods, nil
}

func parseDate(cmd *cobra.Command, flag string) (*time.Time, error) {
	raw, _ := cmd.Flags().GetString(flag)
	dp := dateparse.New()
	t, err := dp.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("--%s: %w", flag, err)
	}

	return &t, nil
}
