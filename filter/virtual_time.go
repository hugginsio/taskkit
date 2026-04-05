// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package filter

import (
	"time"

	"github.com/hugginsio/taskkit"
)

// startOfDay returns midnight in local time for the given time.
func startOfDay(t time.Time) time.Time {
	y, m, d := t.Local().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

// startOfWeek returns midnight local time of the Sunday that begins t's week.
func startOfWeek(t time.Time) time.Time {
	local := t.Local()
	y, m, d := local.Date()
	return time.Date(y, m, d-int(local.Weekday()), 0, 0, 0, 0, time.Local)
}

// startOfMonth returns the first moment of the month containing t (local time).
func startOfMonth(t time.Time) time.Time {
	y, m, _ := t.Local().Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, time.Local)
}

// startOfQuarter returns the first moment of the calendar quarter containing t (local time).
// Quarters begin in January, April, July, and October.
func startOfQuarter(t time.Time) time.Time {
	y, m, _ := t.Local().Date()
	startMonth := time.Month(((int(m)-1)/3)*3 + 1)
	return time.Date(y, startMonth, 1, 0, 0, 0, 0, time.Local)
}

// startOfYear returns the first moment of the year containing t (local time).
func startOfYear(t time.Time) time.Time {
	return time.Date(t.Local().Year(), 1, 1, 0, 0, 0, 0, time.Local)
}

// deadlineInWindow returns a positive/negative Clause pair for tasks whose
// deadline falls within [start, end).
func deadlineInWindow(start, end time.Time) (positive, negative Clause) {
	s := start.UTC().Format(iso8601)
	e := end.UTC().Format(iso8601)
	positive = Clause{
		SQL:    "deadline IS NOT NULL AND deadline >= ? AND deadline < ?",
		Params: []any{s, e},
	}
	negative = Clause{
		SQL:    "(deadline IS NULL OR deadline < ? OR deadline >= ?)",
		Params: []any{s, e},
	}
	return
}

func init() {
	// Task has a deadline that has already passed.
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

	// Task deadline falls sometime today (local time).
	register("TODAY", virtualTag{
		positive: func(now time.Time) Clause {
			p, _ := deadlineInWindow(startOfDay(now), startOfDay(now).AddDate(0, 0, 1))
			return p
		},
		negative: func(now time.Time) Clause {
			_, n := deadlineInWindow(startOfDay(now), startOfDay(now).AddDate(0, 0, 1))
			return n
		},
		applies: func(t *taskkit.Task, now time.Time) bool {
			if t.Deadline == nil {
				return false
			}

			start := startOfDay(now)
			return !t.Deadline.Before(start) && t.Deadline.Before(start.AddDate(0, 0, 1))
		},
	})

	// Task deadline fell yesterday (local time).
	register("YESTERDAY", virtualTag{
		positive: func(now time.Time) Clause {
			p, _ := deadlineInWindow(startOfDay(now).AddDate(0, 0, -1), startOfDay(now))
			return p
		},
		negative: func(now time.Time) Clause {
			_, n := deadlineInWindow(startOfDay(now).AddDate(0, 0, -1), startOfDay(now))
			return n
		},
		applies: func(t *taskkit.Task, now time.Time) bool {
			if t.Deadline == nil {
				return false
			}

			start := startOfDay(now).AddDate(0, 0, -1)
			return !t.Deadline.Before(start) && t.Deadline.Before(start.AddDate(0, 0, 1))
		},
	})

	// Task deadline falls tomorrow (local time).
	register("TOMORROW", virtualTag{
		positive: func(now time.Time) Clause {
			p, _ := deadlineInWindow(startOfDay(now).AddDate(0, 0, 1), startOfDay(now).AddDate(0, 0, 2))
			return p
		},
		negative: func(now time.Time) Clause {
			_, n := deadlineInWindow(startOfDay(now).AddDate(0, 0, 1), startOfDay(now).AddDate(0, 0, 2))
			return n
		},
		applies: func(t *taskkit.Task, now time.Time) bool {
			if t.Deadline == nil {
				return false
			}

			start := startOfDay(now).AddDate(0, 0, 1)
			return !t.Deadline.Before(start) && t.Deadline.Before(start.AddDate(0, 0, 1))
		},
	})

	// Task deadline falls within the current calendar week (Mon–Sun, local time).
	register("WEEK", virtualTag{
		positive: func(now time.Time) Clause {
			p, _ := deadlineInWindow(startOfWeek(now), startOfWeek(now).AddDate(0, 0, 7))
			return p
		},
		negative: func(now time.Time) Clause {
			_, n := deadlineInWindow(startOfWeek(now), startOfWeek(now).AddDate(0, 0, 7))
			return n
		},
		applies: func(t *taskkit.Task, now time.Time) bool {
			if t.Deadline == nil {
				return false
			}

			start := startOfWeek(now)
			return !t.Deadline.Before(start) && t.Deadline.Before(start.AddDate(0, 0, 7))
		},
	})

	// Task deadline falls within the current calendar month (local time).
	register("MONTH", virtualTag{
		positive: func(now time.Time) Clause {
			p, _ := deadlineInWindow(startOfMonth(now), startOfMonth(now).AddDate(0, 1, 0))
			return p
		},
		negative: func(now time.Time) Clause {
			_, n := deadlineInWindow(startOfMonth(now), startOfMonth(now).AddDate(0, 1, 0))
			return n
		},
		applies: func(t *taskkit.Task, now time.Time) bool {
			if t.Deadline == nil {
				return false
			}

			start := startOfMonth(now)
			return !t.Deadline.Before(start) && t.Deadline.Before(start.AddDate(0, 1, 0))
		},
	})

	// Task deadline falls within the current calendar quarter (local time).
	register("QUARTER", virtualTag{
		positive: func(now time.Time) Clause {
			p, _ := deadlineInWindow(startOfQuarter(now), startOfQuarter(now).AddDate(0, 3, 0))
			return p
		},
		negative: func(now time.Time) Clause {
			_, n := deadlineInWindow(startOfQuarter(now), startOfQuarter(now).AddDate(0, 3, 0))
			return n
		},
		applies: func(t *taskkit.Task, now time.Time) bool {
			if t.Deadline == nil {
				return false
			}

			start := startOfQuarter(now)
			return !t.Deadline.Before(start) && t.Deadline.Before(start.AddDate(0, 3, 0))
		},
	})

	// Task deadline falls within the current calendar year (local time).
	register("YEAR", virtualTag{
		positive: func(now time.Time) Clause {
			p, _ := deadlineInWindow(startOfYear(now), startOfYear(now).AddDate(1, 0, 0))
			return p
		},
		negative: func(now time.Time) Clause {
			_, n := deadlineInWindow(startOfYear(now), startOfYear(now).AddDate(1, 0, 0))
			return n
		},
		applies: func(t *taskkit.Task, now time.Time) bool {
			if t.Deadline == nil {
				return false
			}

			start := startOfYear(now)
			return !t.Deadline.Before(start) && t.Deadline.Before(start.AddDate(1, 0, 0))
		},
	})
}
