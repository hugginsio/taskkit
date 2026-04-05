// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package client

import (
	"time"

	"github.com/hugginsio/taskkit"
	"github.com/hugginsio/taskkit/internal/engine"
)

// Modification is a function that, given the current task state, updates it
// in place and returns the engine mutations to apply. Multiple modifications
// can be passed to Modify and are applied atomically in a single transaction.
type Modification func(task *taskkit.Task) ([]engine.Mutation, error)

// SetDescription changes the task description.
func SetDescription(s string) Modification {
	return func(task *taskkit.Task) ([]engine.Mutation, error) {
		mut := engine.SetDescription(task.Description, s)
		task.Description = s
		return []engine.Mutation{mut}, nil
	}
}

// SetStatus changes the task status. If the new status closes the task
// (completed or removed), the display ID is also freed.
func SetStatus(s taskkit.Status) Modification {
	return func(task *taskkit.Task) ([]engine.Mutation, error) {
		muts := []engine.Mutation{engine.SetStatus(task.Status, s)}
		if s == taskkit.StatusCompleted || s == taskkit.StatusRemoved {
			muts = append(muts, engine.SetDisplayID(int64(task.DisplayID), 0))
		}

		task.Status = s
		return muts, nil
	}
}

// SetProject changes the task project.
func SetProject(project string) Modification {
	return func(task *taskkit.Task) ([]engine.Mutation, error) {
		mut := engine.SetProject(task.Project, project)
		task.Project = project
		return []engine.Mutation{mut}, nil
	}
}

// SetDeadline changes the task deadline. Pass nil to clear it.
func SetDeadline(t *time.Time) Modification {
	return func(task *taskkit.Task) ([]engine.Mutation, error) {
		mut := engine.SetDeadline(task.Deadline, t)
		task.Deadline = t
		return []engine.Mutation{mut}, nil
	}
}

// SetScheduled changes the task scheduled date. Pass nil to clear it.
func SetScheduled(t *time.Time) Modification {
	return func(task *taskkit.Task) ([]engine.Mutation, error) {
		mut := engine.SetScheduled(task.Scheduled, t)
		task.Scheduled = t
		return []engine.Mutation{mut}, nil
	}
}

// SetWait changes the task wait date. Pass nil to clear it.
func SetWait(t *time.Time) Modification {
	return func(task *taskkit.Task) ([]engine.Mutation, error) {
		mut := engine.SetWait(task.Wait, t)
		task.Wait = t
		return []engine.Mutation{mut}, nil
	}
}

// AddTag adds a tag to the task.
func AddTag(tag string) Modification {
	return func(task *taskkit.Task) ([]engine.Mutation, error) {
		task.Tags = append(task.Tags, tag)
		return []engine.Mutation{engine.AddTag(tag)}, nil
	}
}

// AddBlockedBy records that the task is blocked by dependsOnID.
// Returns ErrCyclicDependency if the dependency would create a cycle.
func AddBlockedBy(dependsOnID string) Modification {
	return func(task *taskkit.Task) ([]engine.Mutation, error) {
		return []engine.Mutation{engine.AddDependency(dependsOnID)}, nil
	}
}

// RemoveBlockedBy removes the dependency where the task is blocked by dependsOnID.
func RemoveBlockedBy(dependsOnID string) Modification {
	return func(task *taskkit.Task) ([]engine.Mutation, error) {
		return []engine.Mutation{engine.RemoveDependency(dependsOnID)}, nil
	}
}

// RemoveTag removes a tag from the task.
func RemoveTag(tag string) Modification {
	return func(task *taskkit.Task) ([]engine.Mutation, error) {
		updated := task.Tags[:0]
		for _, t := range task.Tags {
			if t != tag {
				updated = append(updated, t)
			}
		}

		task.Tags = updated
		return []engine.Mutation{engine.RemoveTag(tag)}, nil
	}
}
