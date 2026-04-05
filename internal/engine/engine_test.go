// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package engine_test

import (
	"context"
	"io/fs"
	"testing"

	"github.com/hugginsio/taskkit"
	"github.com/hugginsio/taskkit/internal/db"
	"github.com/hugginsio/taskkit/internal/engine"
	"github.com/oklog/ulid/v2"
)

// newTestEngine creates an in-memory Engine with migrations applied.
func newTestEngine(t *testing.T) *engine.Engine {
	t.Helper()

	migrations, err := fs.Sub(db.Migrations, "migration")
	if err != nil {
		t.Fatalf("migrations sub: %v", err)
	}

	eng, err := engine.NewEngine(context.Background(), ":memory:", migrations)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	t.Cleanup(func() { eng.Close() })

	return eng
}

// newTask returns a minimal pending task with a fresh ULID.
func newTask(description string) *taskkit.Task {
	return &taskkit.Task{
		TaskID:      ulid.Make(),
		Description: description,
		Status:      taskkit.StatusPending,
	}
}

func TestCreateTask(t *testing.T) {
	ctx := context.Background()
	eng := newTestEngine(t)
	task := newTask("buy milk")

	if err := eng.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	got, err := eng.TaskByID(ctx, task.TaskID.String())
	if err != nil {
		t.Fatalf("TaskByID: %v", err)
	}

	if got.Description != task.Description {
		t.Errorf("description: got %q, want %q", got.Description, task.Description)
	}

	if got.Status != taskkit.StatusPending {
		t.Errorf("status: got %q, want %q", got.Status, taskkit.StatusPending)
	}

	if got.DisplayID < 1 {
		t.Errorf("display_id: got %d, want >= 1", got.DisplayID)
	}
}

func TestCreateTask_AssignsDisplayID(t *testing.T) {}

func TestCreateTask_RecyclesDisplayID(t *testing.T) {}

func TestTaskByID(t *testing.T) {}

func TestTaskByID_NotFound(t *testing.T) {}

func TestQueryTasks(t *testing.T) {}

func TestDeleteTask(t *testing.T) {}

func TestNextDisplayID(t *testing.T) {}

func TestHasDependencyPath(t *testing.T) {}

func TestHasDependencyPath_NoCycle(t *testing.T) {}
