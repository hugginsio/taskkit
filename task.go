// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package taskkit

import (
	"time"

	"github.com/oklog/ulid/v2"
)

// Status represents the lifecycle state of a Task.
type Status string

const (
	StatusPending   Status = "pending"   // Task is active and visible.
	StatusWaiting   Status = "waiting"   // Task is hidden until its Wait date passes.
	StatusCompleted Status = "completed" // Task has been completed.
	StatusRemoved   Status = "removed"   // Task has been soft-deleted.
)

type Task struct {
	TaskID           ulid.ULID          `json:"id"`                          // The unique task identifier.
	DisplayID        int                `json:"display_id"`                  // User-friendly incrementing integer; unique amongst pending tasks and recycled when tasks are completed or deleted.
	Project          string             `json:"project,omitempty"`           // The unique Project identifier.
	Description      string             `json:"description"`                 // The user-facing Task text.
	Tags             []string           `json:"tags,omitempty"`              // Arbitrary labels for filtering and grouping.
	VirtualTags      []string           `json:"virtual_tags,omitempty"`      // Computed tags describing Task state; not stored.
	Status           Status             `json:"status"`                      // The current lifecycle state of the Task.
	BlockedBy        []*Task            `json:"blocked_by,omitempty"`        // Task that must be completed prior to this Task.
	Blocking         []*Task            `json:"blocking,omitempty"`          // Tasks that depend on this Task being completed.
	Deadline         *time.Time         `json:"deadline,omitempty"`          // When the Task is supposed to be completed.
	Scheduled        *time.Time         `json:"scheduled,omitempty"`         // When work on the Task is supposed to begin.
	Wait             *time.Time         `json:"wait,omitempty"`              // When the Task becomes visible; the Task has StatusWaiting until this time passes.
	Urgency          float64            `json:"urgency"`                     // Computed urgency of a Task; higher values are more urgent.
	UrgencyBreakdown []UrgencyComponent `json:"urgency_breakdown,omitempty"` // Per-component breakdown; populated only by info-style queries.
	Created          time.Time          `json:"created"`                     // When the Task was created.
	Modified         time.Time          `json:"modified"`                    // When the Task was last modified.
	History          []HistoryEntry     `json:"history,omitempty"`           // Audit log; populated only by info-style queries.
}

// UrgencyComponent holds one term of the urgency score breakdown.
type UrgencyComponent struct {
	Label       string  `json:"label"`
	Coefficient float64 `json:"coefficient"`
	Weight      float64 `json:"weight"`
}

// HistoryChange describes a single field change within an operation.
type HistoryChange struct {
	Field    string `json:"field"`
	OldValue string `json:"old_value,omitempty"`
	NewValue string `json:"new_value,omitempty"`
}

// HistoryEntry groups all changes that occurred in one atomic operation.
type HistoryEntry struct {
	OperationID string          `json:"operation_id"`
	Action      string          `json:"action"`
	Time        time.Time       `json:"time"`
	Changes     []HistoryChange `json:"changes,omitempty"`
}
