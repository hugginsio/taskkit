// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package engine

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hugginsio/taskkit"
	"github.com/hugginsio/taskkit/internal/db"
)

// ErrNothingToUndo is returned when there is no operation to undo.
var ErrNothingToUndo = errors.New("nothing to undo")

// Undo reverses the most recent operation by restoring old field values and
// logging the reversal as a new operation to preserve the full history.
func (e *Engine) Undo(ctx context.Context) error {
	q := db.New(e.db)

	opID, err := q.GetLatestOperationID(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNothingToUndo
	}

	if err != nil {
		return fmt.Errorf("undo: get latest operation: %w", err)
	}

	rows, err := q.GetHistoryByOperationID(ctx, opID)
	if err != nil {
		return fmt.Errorf("undo: get history: %w", err)
	}

	// Group reversal mutations by task_id (operations can span multiple tasks).
	byTask := make(map[string][]Mutation)
	for i := len(rows) - 1; i >= 0; i-- {
		row := rows[i]
		mut, err := reverseMutation(row)
		if err != nil {
			return fmt.Errorf("undo: reverse %q on %q: %w", row.Field.String, row.TaskID, err)
		}

		if mut != nil {
			byTask[row.TaskID] = append(byTask[row.TaskID], mut)
		}
	}

	for taskID, muts := range byTask {
		if err := e.Mutate(ctx, taskID, muts...); err != nil {
			return fmt.Errorf("undo: apply: %w", err)
		}
	}

	return nil
}

// reverseMutation returns the Mutation that reverses the given history row,
// or nil if no reversal is needed (e.g., hard deletes).
func reverseMutation(row db.History) (Mutation, error) {
	switch row.Action {
	case "create":
		// Undo a creation by marking the task removed.
		return SetStatus(taskkit.StatusPending, taskkit.StatusRemoved), nil

	case "update":
		if !row.Field.Valid {
			return nil, nil
		}

		switch row.Field.String {
		case "description":
			return SetDescription(row.NewValue.String, row.OldValue.String), nil

		case "status":
			return SetStatus(
				taskkit.Status(row.NewValue.String),
				taskkit.Status(row.OldValue.String),
			), nil

		case "project":
			return SetProject(row.NewValue.String, row.OldValue.String), nil

		case "deadline":
			cur, err := parseNullTime(row.NewValue)
			if err != nil {
				return nil, fmt.Errorf("parse deadline: %w", err)
			}

			prev, err := parseNullTime(row.OldValue)
			if err != nil {
				return nil, fmt.Errorf("parse deadline: %w", err)
			}

			return SetDeadline(cur, prev), nil

		case "scheduled":
			cur, err := parseNullTime(row.NewValue)
			if err != nil {
				return nil, fmt.Errorf("parse scheduled: %w", err)
			}

			prev, err := parseNullTime(row.OldValue)
			if err != nil {
				return nil, fmt.Errorf("parse scheduled: %w", err)
			}

			return SetScheduled(cur, prev), nil

		case "wait":
			cur, err := parseNullTime(row.NewValue)
			if err != nil {
				return nil, fmt.Errorf("parse wait: %w", err)
			}

			prev, err := parseNullTime(row.OldValue)
			if err != nil {
				return nil, fmt.Errorf("parse wait: %w", err)
			}

			return SetWait(cur, prev), nil

		case "display_id":
			cur, err := strconv.ParseInt(row.NewValue.String, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parse display_id: %w", err)
			}

			prev, err := strconv.ParseInt(row.OldValue.String, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parse display_id: %w", err)
			}

			return SetDisplayID(cur, prev), nil

		case "tag.remove":
			return AddTag(row.OldValue.String), nil

		case "tag.add":
			return RemoveTag(row.NewValue.String), nil

		case "dependency.remove":
			return AddDependency(row.OldValue.String), nil

		case "dependency.add":
			return RemoveDependency(row.NewValue.String), nil
		}
	}

	return nil, nil
}

func parseNullTime(ns sql.NullString) (*time.Time, error) {
	if !ns.Valid || ns.String == "" {
		return nil, nil
	}

	t, err := time.Parse(iso8601, ns.String)
	if err != nil {
		return nil, err
	}

	return &t, nil
}
