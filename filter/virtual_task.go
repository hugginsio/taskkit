// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package filter

import (
	"time"

	"github.com/hugginsio/taskkit"
)

func init() {
	// Task has a scheduled date set.
	register("SCHEDULED", virtualTag{
		positive: func(_ time.Time) Clause {
			return Clause{SQL: "scheduled IS NOT NULL"}
		},
		negative: func(_ time.Time) Clause {
			return Clause{SQL: "scheduled IS NULL"}
		},
		applies: func(t *taskkit.Task, _ time.Time) bool { return t.Scheduled != nil },
	})

	// Task has a wait date set.
	register("WAITING", virtualTag{
		positive: func(_ time.Time) Clause {
			return Clause{SQL: "wait IS NOT NULL"}
		},
		negative: func(_ time.Time) Clause {
			return Clause{SQL: "wait IS NULL"}
		},
		applies: func(t *taskkit.Task, _ time.Time) bool { return t.Wait != nil },
	})

	// Task has at least one user-defined tag.
	register("TAGGED", virtualTag{
		positive: func(_ time.Time) Clause {
			return Clause{SQL: "task_id IN (SELECT task_id FROM task_tags)"}
		},
		negative: func(_ time.Time) Clause {
			return Clause{SQL: "task_id NOT IN (SELECT task_id FROM task_tags)"}
		},
		applies: func(t *taskkit.Task, _ time.Time) bool { return len(t.Tags) > 0 },
	})

	// Task belongs to a project.
	register("PROJECT", virtualTag{
		positive: func(_ time.Time) Clause {
			return Clause{SQL: "project IS NOT NULL AND project != ''"}
		},
		negative: func(_ time.Time) Clause {
			return Clause{SQL: "(project IS NULL OR project = '')"}
		},
		applies: func(t *taskkit.Task, _ time.Time) bool { return t.Project != "" },
	})
}
