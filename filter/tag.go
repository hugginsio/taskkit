// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package filter

import "time"

// HasTag matches tasks that have the given tag, or whose state matches the
// named virtual tag if the name is registered (e.g. "BLOCKED", "OVERDUE").
// Virtual tag names are uppercase by convention; "blocked" still matches the
// stored tag.
func HasTag(tag string) Filter {
	return func() Clause {
		if vt, ok := virtualTags[tag]; ok {
			return vt.positive(time.Now())
		}

		return Clause{
			SQL:    "task_id IN (SELECT task_id FROM task_tags WHERE tag = ?)",
			Params: []any{tag},
		}
	}
}

// LacksTag matches tasks that do not have the given tag, or whose state does
// not match the named virtual tag if the name is registered.
func LacksTag(tag string) Filter {
	return func() Clause {
		if vt, ok := virtualTags[tag]; ok {
			return vt.negative(time.Now())
		}

		return Clause{
			SQL:    "task_id NOT IN (SELECT task_id FROM task_tags WHERE tag = ?)",
			Params: []any{tag},
		}
	}
}
