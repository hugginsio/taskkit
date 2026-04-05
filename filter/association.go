// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package filter

// ID matches the task with the given ULID string.
func ID(id string) Filter {
	return func() Clause {
		return Clause{SQL: "task_id = ?", Params: []any{id}}
	}
}

// DisplayID matches the active task with the given display ID.
func DisplayID(id int) Filter {
	return func() Clause {
		return Clause{SQL: "display_id = ?", Params: []any{int64(id)}}
	}
}

// BlockedBy matches tasks that are blocked by the given task ID.
func BlockedBy(dependsOnID string) Filter {
	return func() Clause {
		return Clause{
			SQL:    "task_id IN (SELECT task_id FROM task_dependencies WHERE depends_on = ?)",
			Params: []any{dependsOnID},
		}
	}
}

// Blocking matches tasks that are blocking the given task ID.
func Blocking(taskID string) Filter {
	return func() Clause {
		return Clause{
			SQL:    "task_id IN (SELECT depends_on FROM task_dependencies WHERE task_id = ?)",
			Params: []any{taskID},
		}
	}
}
