// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package engine_test

import (
	"context"
	"testing"

	"github.com/hugginsio/taskkit/internal/engine"
)

func TestSetDescription(t *testing.T) {
	ctx := context.Background()
	eng := newTestEngine(t)
	task := newTask("original")

	if err := eng.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	if err := eng.Mutate(ctx, task.TaskID.String(), engine.SetDescription("original", "updated")); err != nil {
		t.Fatalf("Mutate: %v", err)
	}

	got, err := eng.TaskByID(ctx, task.TaskID.String())
	if err != nil {
		t.Fatalf("TaskByID: %v", err)
	}

	if got.Description != "updated" {
		t.Errorf("description: got %q, want %q", got.Description, "updated")
	}
}

func TestSetStatus(t *testing.T) {}

func TestSetProject(t *testing.T) {}

func TestSetProject_Clear(t *testing.T) {}

func TestSetDeadline(t *testing.T) {}

func TestSetDeadline_Clear(t *testing.T) {}

func TestSetScheduled(t *testing.T) {}

func TestSetScheduled_Clear(t *testing.T) {}

func TestSetWait(t *testing.T) {}

func TestSetWait_Clear(t *testing.T) {}

func TestSetDisplayID(t *testing.T) {}

func TestAddTag(t *testing.T) {}

func TestRemoveTag(t *testing.T) {}

func TestAddDependency(t *testing.T) {}

func TestRemoveDependency(t *testing.T) {}
