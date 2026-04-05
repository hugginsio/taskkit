// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hugginsio/taskkit"
	"github.com/hugginsio/taskkit/filter"
)

// resolveTask looks up a task by display ID (integer string) or ULID.
// Returns ErrNotFound (from client) if no matching task exists.
func resolveTask(ctx context.Context, id string) (*taskkit.Task, error) {
	var filters []filter.Filter

	if n, err := strconv.Atoi(id); err == nil {
		filters = append(filters, filter.DisplayID(n))
	} else {
		filters = append(filters, filter.ID(id))
	}

	tasks, err := c.Get(ctx, filters...)
	if err != nil {
		return nil, fmt.Errorf("resolve %q: %w", id, err)
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("resolve %q: task not found", id)
	}

	return tasks[0], nil
}
