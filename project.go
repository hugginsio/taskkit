// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package taskkit

// ProjectSummary holds aggregated Task stats for a single Project.
type ProjectSummary struct {
	Name      string  `json:"name"`      // The unique Project name.
	Remaining int     `json:"remaining"` // The sum of all pending and waiting Tasks.
	Completed int     `json:"completed"` // The number of completed Tasks.
	Percent   int     `json:"percent"`   // The percentage of completed tasks.
	Tasks     []*Task `json:"tasks"`     // All Tasks associated with the Project.
}
