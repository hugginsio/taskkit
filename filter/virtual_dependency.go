// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package filter

import (
	"time"

	"github.com/hugginsio/taskkit"
)

const activeDepsSubquery = `SELECT td.task_id FROM task_dependencies td
	JOIN tasks t ON t.task_id = td.depends_on
	WHERE t.status IN ('pending', 'waiting')`

func init() {
	register("BLOCKED", virtualTag{
		positive: func(_ time.Time) Clause {
			return Clause{SQL: "task_id IN (" + activeDepsSubquery + ")"}
		},
		negative: func(_ time.Time) Clause {
			return Clause{SQL: "task_id NOT IN (" + activeDepsSubquery + ")"}
		},
		applies: func(t *taskkit.Task, _ time.Time) bool { return len(t.BlockedBy) > 0 },
	})
}
