// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package filter

import (
	"time"

	"github.com/oklog/ulid/v2"
)

const iso8601 = time.RFC3339

// HasDeadline matches tasks with a deadline set.
func HasDeadline() Filter {
	return func() Clause { return Clause{SQL: "deadline IS NOT NULL"} }
}

// NoDeadline matches tasks with no deadline set.
func NoDeadline() Filter {
	return func() Clause { return Clause{SQL: "deadline IS NULL"} }
}

// DeadlineBefore matches tasks whose deadline is before t.
func DeadlineBefore(t time.Time) Filter {
	return func() Clause {
		return Clause{SQL: "deadline < ?", Params: []any{t.UTC().Format(iso8601)}}
	}
}

// DeadlineAfter matches tasks whose deadline is after t.
func DeadlineAfter(t time.Time) Filter {
	return func() Clause {
		return Clause{SQL: "deadline > ?", Params: []any{t.UTC().Format(iso8601)}}
	}
}

// HasScheduled matches tasks with a scheduled date set.
func HasScheduled() Filter {
	return func() Clause { return Clause{SQL: "scheduled IS NOT NULL"} }
}

// NoScheduled matches tasks with no scheduled date set.
func NoScheduled() Filter {
	return func() Clause { return Clause{SQL: "scheduled IS NULL"} }
}

// ScheduledBefore matches tasks whose scheduled date is before t.
func ScheduledBefore(t time.Time) Filter {
	return func() Clause {
		return Clause{SQL: "scheduled < ?", Params: []any{t.UTC().Format(iso8601)}}
	}
}

// ScheduledAfter matches tasks whose scheduled date is after t.
func ScheduledAfter(t time.Time) Filter {
	return func() Clause {
		return Clause{SQL: "scheduled > ?", Params: []any{t.UTC().Format(iso8601)}}
	}
}

// HasWait matches tasks with a wait date set.
func HasWait() Filter {
	return func() Clause { return Clause{SQL: "wait IS NOT NULL"} }
}

// NoWait matches tasks with no wait date set.
func NoWait() Filter {
	return func() Clause { return Clause{SQL: "wait IS NULL"} }
}

// WaitBefore matches tasks whose wait date is before t.
func WaitBefore(t time.Time) Filter {
	return func() Clause {
		return Clause{SQL: "wait < ?", Params: []any{t.UTC().Format(iso8601)}}
	}
}

// WaitAfter matches tasks whose wait date is after t.
func WaitAfter(t time.Time) Filter {
	return func() Clause {
		return Clause{SQL: "wait > ?", Params: []any{t.UTC().Format(iso8601)}}
	}
}

// ulidSentinel returns a ULID string whose embedded timestamp equals t.
// Used for lexicographic time comparisons against ULID columns.
func ulidSentinel(t time.Time) string {
	return ulid.MustNew(ulid.Timestamp(t), ulid.DefaultEntropy()).String()
}

// CreatedBefore matches tasks created before t. Since task_id is a ULID,
// lexicographic comparison is equivalent to time comparison.
func CreatedBefore(t time.Time) Filter {
	return func() Clause {
		return Clause{SQL: "task_id < ?", Params: []any{ulidSentinel(t)}}
	}
}

// CreatedAfter matches tasks created after t.
func CreatedAfter(t time.Time) Filter {
	return func() Clause {
		return Clause{SQL: "task_id > ?", Params: []any{ulidSentinel(t)}}
	}
}

// ModifiedBefore matches tasks whose most recent history entry is before t.
func ModifiedBefore(t time.Time) Filter {
	return func() Clause {
		return Clause{
			SQL:    "(SELECT MAX(operation_id) FROM history WHERE task_id = tasks.task_id) < ?",
			Params: []any{ulidSentinel(t)},
		}
	}
}

// ModifiedAfter matches tasks whose most recent history entry is after t.
func ModifiedAfter(t time.Time) Filter {
	return func() Clause {
		return Clause{
			SQL:    "(SELECT MAX(operation_id) FROM history WHERE task_id = tasks.task_id) > ?",
			Params: []any{ulidSentinel(t)},
		}
	}
}
