// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/hugginsio/taskkit"
	"github.com/hugginsio/taskkit/client"
	"github.com/hugginsio/taskkit/config"
	"github.com/hugginsio/taskkit/filter"
	"github.com/spf13/cobra"
)

const dateLayout = "2006-01-02"

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
	cmd.Flags().String("output", "pretty", "output format (pretty, json)")
	cmd.Flags().StringSlice("columns", nil, "e.g. description, deadline (pretty only)")
}

// registerSortFlag attaches the --sort flag to cmd.
func registerSortFlag(cmd *cobra.Command) {
	cmd.Flags().String("sort", "urgency", "sort order: urgency, id, modified, wait")
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
	cmd.Flags().String("project", "", "project name")

	cmd.Flags().String("status", "", "pending, completed, or removed")
	cmd.RegisterFlagCompletionFunc("status", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return []cobra.Completion{""}, cobra.ShellCompDirectiveNoFileComp
	})

	cmd.Flags().StringSlice("tags", nil, "comma-separated list of tags")
}

// registerDateFlags attaches date mutation flags to cmd.
func registerDateFlags(cmd *cobra.Command) {
	cmd.Flags().String("deadline", "", "date the task should be completed")
	cmd.Flags().String("scheduled", "", "date that work should begin")
	cmd.Flags().String("wait", "", "date the task becomes visible in reports")
}

// registerDependencyFlags attaches dependency mutation flags to cmd.
func registerDependencyFlags(cmd *cobra.Command) {
	cmd.Flags().StringSlice("blocked", nil, "tasks that this task blocks")
	cmd.Flags().StringSlice("blocking", nil, "tasks that block this task")
	cmd.Flags().StringSlice("unblocked", nil, "remove blocked-by dependencies")
}

// registerRemoveFlags attaches removal flags to cmd (for modify only).
func registerRemoveFlags(cmd *cobra.Command) {
	cmd.Flags().StringSlice("untag", nil, "tags to remove")
}

// registerFilterFlags attaches query filter flags to cmd.
func registerFilterFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("status-all", false, "include tasks of any status")
	cmd.Flags().String("blocked-by", "", "tasks blocked by the given task")
	cmd.Flags().String("blocking", "", "tasks that are blocking the given task")
	cmd.Flags().String("deadline-after", "", "tasks with a deadline after date")
	cmd.Flags().String("deadline-before", "", "tasks with a deadline before date")
	cmd.Flags().String("modified-after", "", "tasks last modified after date")
	cmd.Flags().String("modified-before", "", "tasks last modified before date")
	cmd.Flags().String("project", "", "filter by project name")
	cmd.Flags().String("scheduled-after", "", "tasks scheduled to start after date")
	cmd.Flags().String("scheduled-before", "", "tasks scheduled to start before date")
	cmd.Flags().String("wait-after", "", "tasks with a wait date after date")
	cmd.Flags().String("wait-before", "", "tasks with a wait date before date")
	cmd.Flags().StringSlice("status", nil, "filter by statuses (pending, waiting, completed, removed)")
	cmd.Flags().StringSlice("tag", nil, "filter by tags (all must match)")
	cmd.Flags().StringSlice("tags-any", nil, "filter by tags (any may match)")
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
			case taskkit.StatusPending, taskkit.StatusWaiting, taskkit.StatusCompleted, taskkit.StatusRemoved:
			default:
				return nil, fmt.Errorf("unknown status %q: must be pending, waiting, completed, or removed", raw)
			}

			ss = append(ss, s)
		}

		filters = append(filters, filter.StatusAny(ss...))
	default:
		filters = append(filters, filter.StatusAny(taskkit.StatusPending, taskkit.StatusWaiting))
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
		case taskkit.StatusPending, taskkit.StatusWaiting, taskkit.StatusCompleted, taskkit.StatusRemoved:
		default:
			return nil, fmt.Errorf("unknown status %q: must be pending, waiting, completed, or removed", raw)
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
	return parseDateString(flag, raw)
}

// parseDateString parses a date string for a named flag.
// Accepts "today", "tomorrow", or "YYYY-MM-DD".
func parseDateString(flag, raw string) (*time.Time, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	switch raw {
	case "now":
		return &now, nil
	case "today":
		return &today, nil
	case "tomorrow":
		t := today.AddDate(0, 0, 1)
		return &t, nil
	}

	t, err := time.ParseInLocation(dateLayout, raw, time.Local)
	if err != nil {
		return nil, fmt.Errorf("--%s: expected YYYY-MM-DD, now, today, or tomorrow; got %q", flag, raw)
	}

	return &t, nil
}
