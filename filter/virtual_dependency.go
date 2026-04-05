// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package filter

import (
	"time"

	"github.com/hugginsio/taskkit"
)

// activeDepsSubquery returns the task_ids that are blocked by at least one
// active (pending or waiting) dependency.
const activeDepsSubquery = `SELECT td.task_id FROM task_dependencies td
	JOIN tasks t ON t.task_id = td.depends_on
	WHERE t.status IN ('pending', 'waiting')`

// blockingSubquery returns the task_ids that are blocking at least one
// active (pending or waiting) dependent task.
const blockingSubquery = `SELECT td.depends_on FROM task_dependencies td
	JOIN tasks t ON t.task_id = td.task_id
	WHERE t.status IN ('pending', 'waiting')`

func init() {
	// Task has at least one dependency that is not yet complete.
	register("BLOCKED", virtualTag{
		positive: func(_ time.Time) Clause {
			return Clause{SQL: "task_id IN (" + activeDepsSubquery + ")"}
		},
		negative: func(_ time.Time) Clause {
			return Clause{SQL: "task_id NOT IN (" + activeDepsSubquery + ")"}
		},
		applies: func(t *taskkit.Task, _ time.Time) bool { return len(t.BlockedBy) > 0 },
	})

	// Task has no active dependencies blocking it. Convenience inverse of BLOCKED.
	register("UNBLOCKED", virtualTag{
		positive: func(_ time.Time) Clause {
			return Clause{SQL: "task_id NOT IN (" + activeDepsSubquery + ")"}
		},
		negative: func(_ time.Time) Clause {
			return Clause{SQL: "task_id IN (" + activeDepsSubquery + ")"}
		},
		applies: func(t *taskkit.Task, _ time.Time) bool { return len(t.BlockedBy) == 0 },
	})

	// At least one Task depends on this one.
	register("BLOCKING", virtualTag{
		positive: func(_ time.Time) Clause {
			return Clause{SQL: "task_id IN (" + blockingSubquery + ")"}
		},
		negative: func(_ time.Time) Clause {
			return Clause{SQL: "task_id NOT IN (" + blockingSubquery + ")"}
		},
		applies: func(t *taskkit.Task, _ time.Time) bool { return len(t.Blocking) > 0 },
	})
}
