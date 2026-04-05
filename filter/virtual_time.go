// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package filter

import (
	"time"

	"github.com/hugginsio/taskkit"
)

func init() {
	register("OVERDUE", virtualTag{
		positive: func(now time.Time) Clause {
			return Clause{
				SQL:    "deadline IS NOT NULL AND deadline < ?",
				Params: []any{now.UTC().Format(iso8601)},
			}
		},
		negative: func(now time.Time) Clause {
			return Clause{
				SQL:    "(deadline IS NULL OR deadline >= ?)",
				Params: []any{now.UTC().Format(iso8601)},
			}
		},
		applies: func(t *taskkit.Task, now time.Time) bool {
			return t.Deadline != nil && t.Deadline.Before(now)
		},
	})
}
