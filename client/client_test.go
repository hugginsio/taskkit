// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package client_test

import (
	"context"
	"testing"
	"time"

	"github.com/hugginsio/taskkit"
	"github.com/hugginsio/taskkit/client"
	"github.com/hugginsio/taskkit/config"
	"github.com/hugginsio/taskkit/filter"
)

// newTestClient creates a Client backed by the in-memory test configuration.
func newTestClient(t *testing.T) *client.Client {
	t.Helper()

	cfg, err := config.LoadFrom("testdata/config.yaml")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	c, err := client.NewClient(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	t.Cleanup(func() { c.Close() })
	return c
}

func mustAdd(t *testing.T, client *client.Client, task *taskkit.Task) *taskkit.Task {
	t.Helper()
	created, err := client.Add(context.Background(), task)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	return created
}

func mustGet(t *testing.T, client *client.Client, filters ...filter.Filter) []*taskkit.Task {
	tasks, err := client.Get(context.Background(), filters...)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	return tasks
}

func TestAdd(t *testing.T) {
	c := newTestClient(t)

	task := mustAdd(t, c, &taskkit.Task{Description: "buy milk"})

	if task.DisplayID != 1 {
		t.Errorf("display_id: got %d, want 1", task.DisplayID)
	}

	if task.Status != taskkit.StatusPending {
		t.Errorf("status: got %q, want %q", task.Status, taskkit.StatusPending)
	}

	tasks := mustGet(t, c, filter.ID(string(task.TaskID.String())))

	if len(tasks) != 1 || tasks[0].Description != "buy milk" {
		t.Errorf("Get after Add: got %v", tasks)
	}
}

func TestAdd_ValidationError(t *testing.T) {}

func TestAdd_DisplayIDsRecycle(t *testing.T) {}

func TestComplete(t *testing.T) {}

func TestComplete_NotFound(t *testing.T) {}

func TestModify(t *testing.T) {}

func TestModify_NotFound(t *testing.T) {}

func TestModify_ValidationError(t *testing.T) {}

func TestImport(t *testing.T) {}

func TestDelete(t *testing.T) {}

func TestDelete_NotFound(t *testing.T) {}

func TestUndo(t *testing.T) {}

func TestAddBlockedBy(t *testing.T) {}

func TestAddBlockedBy_Cycle(t *testing.T) {}

func TestAddBlockedBy_NotFound(t *testing.T) {}

func TestRemoveBlockedBy(t *testing.T) {}

func TestGet_NoFilters(t *testing.T) {}

func TestGet_ByDisplayID(t *testing.T) {
	c := newTestClient(t)

	task := mustAdd(t, c, &taskkit.Task{Description: "buy almonds"})
	tasks := mustGet(t, c, filter.ID(string(task.TaskID.String())))

	if len(tasks) != 1 {
		t.Errorf("Get after Add: got %v", tasks)
	}
}

func TestGet_ByStatus(t *testing.T) {
	c := newTestClient(t)
	waitDate := time.Now().Add(time.Hour)

	samples := []taskkit.Task{
		{Description: "avacados", Status: taskkit.StatusPending},
		{Description: "bananas", Status: taskkit.StatusPending, Wait: &waitDate},
		{Description: "chocolate", Status: taskkit.StatusCompleted},
		{Description: "donuts", Status: taskkit.StatusRemoved},
	}

	for _, task := range samples {
		mustAdd(t, c, &task)
	}

	pending := mustGet(t, c, filter.Status(taskkit.StatusPending))
	if len(pending) != 2 {
		t.Fatalf("Get pending: got %v", len(pending))
	}

	removed := mustGet(t, c, filter.Status(taskkit.StatusRemoved))
	if len(removed) != 1 {
		t.Fatalf("Get removed: got %v", len(removed))
	}

	any := mustGet(t, c, filter.StatusAny(taskkit.StatusPending, taskkit.StatusCompleted, taskkit.StatusRemoved))
	if len(any) != len(samples) {
		t.Fatalf("Get any: got %v", len(any))
	}
}

func TestGet_ByTag(t *testing.T) {
	c := newTestClient(t)

	samples := []taskkit.Task{
		taskkit.Task{Description: "avacados", Tags: []string{"fruit"}},
		taskkit.Task{Description: "bananas", Tags: []string{"fruit"}},
		taskkit.Task{Description: "chocolate", Tags: []string{"dessert", "fruit", "well-it's-derived-from-a-fruit"}},
		taskkit.Task{Description: "donuts", Tags: []string{"dessert", "not-a-fruit"}},
	}

	for _, task := range samples {
		mustAdd(t, c, &task)
	}

	fruit := mustGet(t, c, filter.HasTag("fruit"))
	if len(fruit) != 3 {
		t.Fatalf("Has tag fruit: got %v", len(fruit))
	}

	lacksFruit := mustGet(t, c, filter.LacksTag("fruit"))
	if len(lacksFruit) != 1 {
		t.Fatalf("Has tag lacksFruit: got %v", len(lacksFruit))
	}
}

func TestProjects(t *testing.T) {}

func TestProjectDetail(t *testing.T) {}

func TestProjectDetail_NotFound(t *testing.T) {}
