// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package taskkit

import (
	"time"

	"github.com/oklog/ulid/v2"
)

// Status represents the lifecycle state of a Task.
type Status uint

const (
	StatusPending   Status = iota // Task is active and visible.
	StatusWaiting                 // Task is hidden until its Wait date passes.
	StatusCompleted               // Task has been completed.
	StatusRemoved                 // Task has been soft-deleted.
)

// Priority represents the user-assigned importance of a Task.
type Priority uint

const (
	PriorityNone   Priority = iota // No priority (default).
	PriorityLow                    // Low priority.
	PriorityMedium                 // Medium priority.
	PriorityHigh                   // High priority.
)

type Task struct {
	TaskID      ulid.ULID  `json:"id"`                  // The unique task identifier.
	DisplayID   uint       `json:"display_id"`          // User-friendly incrementing integer; unique amongst pending tasks and recycled when tasks are completed or deleted.
	Project     string     `json:"project,omitempty"`   // The unique Project identifier.
	Description string     `json:"description"`         // The user-facing Task text.
	Tags        []string   `json:"tags,omitempty"`      // Arbitrary labels for filtering and grouping.
	Status      Status     `json:"status"`              // The current lifecycle state of the Task.
	Priority    Priority   `json:"priority"`            // The importance of the Task.
	Deadline    *time.Time `json:"deadline,omitempty"`  // When the Task is supposed to be completed.
	Scheduled   *time.Time `json:"scheduled,omitempty"` // When work on the Task is supposed to begin.
	Wait        *time.Time `json:"wait,omitempty"`      // When the Task becomes visible; the Task has StatusWaiting until this time passes.
	Modified    time.Time  `json:"modified"`            // When the Task was last modified.
}
