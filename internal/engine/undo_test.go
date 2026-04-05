// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package engine_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hugginsio/taskkit/internal/engine"
)

func TestUndo_NothingToUndo(t *testing.T) {
	eng := newTestEngine(t)

	err := eng.Undo(context.Background())
	if !errors.Is(err, engine.ErrNothingToUndo) {
		t.Errorf("got %v, want ErrNothingToUndo", err)
	}
}

func TestUndo_Create(t *testing.T) {}

func TestUndo_SetDescription(t *testing.T) {}

func TestUndo_SetStatus(t *testing.T) {}

func TestUndo_SetProject(t *testing.T) {}

func TestUndo_SetDeadline(t *testing.T) {}

func TestUndo_SetScheduled(t *testing.T) {}

func TestUndo_SetWait(t *testing.T) {}

func TestUndo_SetDisplayID(t *testing.T) {}

func TestUndo_AddTag(t *testing.T) {}

func TestUndo_RemoveTag(t *testing.T) {}

func TestUndo_AddDependency(t *testing.T) {}

func TestUndo_RemoveDependency(t *testing.T) {}

func TestUndo_MultipleFields(t *testing.T) {}
