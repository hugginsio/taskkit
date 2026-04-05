// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package engine

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/hugginsio/taskkit"
	"github.com/hugginsio/taskkit/internal/db"
)

// ErrCyclicDependency is returned when adding a dependency would create a cycle.
var ErrCyclicDependency = errors.New("engine: cyclic dependency")

// Mutation is a single change applied within a Mutate call. All mutations
// in a call share one operation_id so they undo as a group.
type Mutation func(ctx context.Context, q *db.Queries, taskID, operationID string) error

// SetDescription returns a Mutation that updates the task description.
func SetDescription(current, next string) Mutation {
	return func(ctx context.Context, q *db.Queries, taskID, opID string) error {
		if err := q.UpdateTaskDescription(ctx, db.UpdateTaskDescriptionParams{
			Description: next,
			TaskID:      taskID,
		}); err != nil {
			return fmt.Errorf("set description: %w", err)
		}

		return q.InsertHistory(ctx, db.InsertHistoryParams{
			OperationID: opID,
			TaskID:      taskID,
			Action:      "update",
			Field:       sql.NullString{String: "description", Valid: true},
			OldValue:    sql.NullString{String: current, Valid: true},
			NewValue:    sql.NullString{String: next, Valid: true},
		})
	}
}

// SetStatus returns a Mutation that updates the task status.
func SetStatus(current, next taskkit.Status) Mutation {
	return func(ctx context.Context, q *db.Queries, taskID, opID string) error {
		if err := q.UpdateTaskStatus(ctx, db.UpdateTaskStatusParams{
			Status: string(next),
			TaskID: taskID,
		}); err != nil {
			return fmt.Errorf("set status: %w", err)
		}

		return q.InsertHistory(ctx, db.InsertHistoryParams{
			OperationID: opID,
			TaskID:      taskID,
			Action:      "update",
			Field:       sql.NullString{String: "status", Valid: true},
			OldValue:    sql.NullString{String: string(current), Valid: true},
			NewValue:    sql.NullString{String: string(next), Valid: true},
		})
	}
}

// SetProject returns a Mutation that updates or clears the task project.
// An empty next string clears the field.
func SetProject(current, next string) Mutation {
	return func(ctx context.Context, q *db.Queries, taskID, opID string) error {
		if err := q.UpdateTaskProject(ctx, db.UpdateTaskProjectParams{
			Project: toNullString(next),
			TaskID:  taskID,
		}); err != nil {
			return fmt.Errorf("set project: %w", err)
		}

		return q.InsertHistory(ctx, db.InsertHistoryParams{
			OperationID: opID,
			TaskID:      taskID,
			Action:      "update",
			Field:       sql.NullString{String: "project", Valid: true},
			OldValue:    sql.NullString{String: current, Valid: current != ""},
			NewValue:    sql.NullString{String: next, Valid: next != ""},
		})
	}
}

// SetDeadline returns a Mutation that updates or clears the task deadline.
func SetDeadline(current, next *time.Time) Mutation {
	return func(ctx context.Context, q *db.Queries, taskID, opID string) error {
		if err := q.UpdateTaskDeadline(ctx, db.UpdateTaskDeadlineParams{
			Deadline: toNullTime(next),
			TaskID:   taskID,
		}); err != nil {
			return fmt.Errorf("set deadline: %w", err)
		}

		return q.InsertHistory(ctx, db.InsertHistoryParams{
			OperationID: opID,
			TaskID:      taskID,
			Action:      "update",
			Field:       sql.NullString{String: "deadline", Valid: true},
			OldValue:    toNullTime(current),
			NewValue:    toNullTime(next),
		})
	}
}

// SetScheduled returns a Mutation that updates or clears the task scheduled date.
func SetScheduled(current, next *time.Time) Mutation {
	return func(ctx context.Context, q *db.Queries, taskID, opID string) error {
		if err := q.UpdateTaskScheduled(ctx, db.UpdateTaskScheduledParams{
			Scheduled: toNullTime(next),
			TaskID:    taskID,
		}); err != nil {
			return fmt.Errorf("set scheduled: %w", err)
		}

		return q.InsertHistory(ctx, db.InsertHistoryParams{
			OperationID: opID,
			TaskID:      taskID,
			Action:      "update",
			Field:       sql.NullString{String: "scheduled", Valid: true},
			OldValue:    toNullTime(current),
			NewValue:    toNullTime(next),
		})
	}
}

// SetWait returns a Mutation that updates or clears the task wait date.
func SetWait(current, next *time.Time) Mutation {
	return func(ctx context.Context, q *db.Queries, taskID, opID string) error {
		if err := q.UpdateTaskWait(ctx, db.UpdateTaskWaitParams{
			Wait:   toNullTime(next),
			TaskID: taskID,
		}); err != nil {
			return fmt.Errorf("set wait: %w", err)
		}

		return q.InsertHistory(ctx, db.InsertHistoryParams{
			OperationID: opID,
			TaskID:      taskID,
			Action:      "update",
			Field:       sql.NullString{String: "wait", Valid: true},
			OldValue:    toNullTime(current),
			NewValue:    toNullTime(next),
		})
	}
}

// SetDisplayID returns a Mutation that updates the task display ID.
func SetDisplayID(current, next int64) Mutation {
	return func(ctx context.Context, q *db.Queries, taskID, opID string) error {
		if err := q.UpdateTaskDisplayID(ctx, db.UpdateTaskDisplayIDParams{
			DisplayID: next,
			TaskID:    taskID,
		}); err != nil {
			return fmt.Errorf("set display_id: %w", err)
		}

		return q.InsertHistory(ctx, db.InsertHistoryParams{
			OperationID: opID,
			TaskID:      taskID,
			Action:      "update",
			Field:       sql.NullString{String: "display_id", Valid: true},
			OldValue:    sql.NullString{String: fmt.Sprintf("%d", current), Valid: true},
			NewValue:    sql.NullString{String: fmt.Sprintf("%d", next), Valid: true},
		})
	}
}

// AddTag returns a Mutation that adds a tag to the task.
func AddTag(tag string) Mutation {
	return func(ctx context.Context, q *db.Queries, taskID, opID string) error {
		if err := q.InsertTag(ctx, db.InsertTagParams{TaskID: taskID, Tag: tag}); err != nil {
			return fmt.Errorf("add tag: %w", err)
		}

		return q.InsertHistory(ctx, db.InsertHistoryParams{
			OperationID: opID,
			TaskID:      taskID,
			Action:      "update",
			Field:       sql.NullString{String: "tag.add", Valid: true},
			NewValue:    sql.NullString{String: tag, Valid: true},
		})
	}
}

// RemoveTag returns a Mutation that removes a tag from the task.
func RemoveTag(tag string) Mutation {
	return func(ctx context.Context, q *db.Queries, taskID, opID string) error {
		if err := q.DeleteTag(ctx, db.DeleteTagParams{TaskID: taskID, Tag: tag}); err != nil {
			return fmt.Errorf("remove tag: %w", err)
		}

		return q.InsertHistory(ctx, db.InsertHistoryParams{
			OperationID: opID,
			TaskID:      taskID,
			Action:      "update",
			Field:       sql.NullString{String: "tag.remove", Valid: true},
			OldValue:    sql.NullString{String: tag, Valid: true},
		})
	}
}

// AddDependency returns a Mutation that records that taskID is blocked by dependsOnID.
// Returns ErrCyclicDependency if the new edge would create a cycle.
func AddDependency(dependsOnID string) Mutation {
	return func(ctx context.Context, q *db.Queries, taskID, opID string) error {
		// Cycle check: a cycle exists if there is already a path from dependsOnID to taskID.
		visited := make(map[string]bool)
		queue := []string{dependsOnID}
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			if visited[current] {
				continue
			}

			visited[current] = true
			deps, err := q.GetBlockedByTaskID(ctx, current)
			if err != nil {
				return fmt.Errorf("add dependency: cycle check: %w", err)
			}

			for _, dep := range deps {
				if dep == taskID {
					return ErrCyclicDependency
				}

				if !visited[dep] {
					queue = append(queue, dep)
				}
			}
		}

		if err := q.InsertDependency(ctx, db.InsertDependencyParams{
			TaskID:    taskID,
			DependsOn: dependsOnID,
		}); err != nil {
			return fmt.Errorf("add dependency: %w", err)
		}

		return q.InsertHistory(ctx, db.InsertHistoryParams{
			OperationID: opID,
			TaskID:      taskID,
			Action:      "update",
			Field:       sql.NullString{String: "dependency.add", Valid: true},
			NewValue:    sql.NullString{String: dependsOnID, Valid: true},
		})
	}
}

// RemoveDependency returns a Mutation that removes the dependency where taskID is blocked by dependsOnID.
func RemoveDependency(dependsOnID string) Mutation {
	return func(ctx context.Context, q *db.Queries, taskID, opID string) error {
		if err := q.DeleteDependency(ctx, db.DeleteDependencyParams{
			TaskID:    taskID,
			DependsOn: dependsOnID,
		}); err != nil {
			return fmt.Errorf("remove dependency: %w", err)
		}

		return q.InsertHistory(ctx, db.InsertHistoryParams{
			OperationID: opID,
			TaskID:      taskID,
			Action:      "update",
			Field:       sql.NullString{String: "dependency.remove", Valid: true},
			OldValue:    sql.NullString{String: dependsOnID, Valid: true},
		})
	}
}
