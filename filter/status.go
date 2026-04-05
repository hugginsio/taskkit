// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package filter

import "github.com/hugginsio/taskkit"

// Status matches tasks with the given status.
func Status(s taskkit.Status) Filter {
	return func() Clause {
		return Clause{SQL: "status = ?", Params: []any{string(s)}}
	}
}

// StatusAny matches tasks whose status is one of the provided values.
func StatusAny(statuses ...taskkit.Status) Filter {
	return func() Clause {
		values := make([]any, len(statuses))
		for i, s := range statuses {
			values[i] = string(s)
		}

		return inClause("status", values)
	}
}
