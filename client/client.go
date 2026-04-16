// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

// Package client provides the public API for TaskKit.
package client

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/adrg/xdg"
	"github.com/hugginsio/taskkit"
	"github.com/hugginsio/taskkit/config"
	"github.com/hugginsio/taskkit/filter"
	"github.com/hugginsio/taskkit/internal/db"
	"github.com/hugginsio/taskkit/internal/engine"
	"github.com/hugginsio/taskkit/urgency"
	"github.com/oklog/ulid/v2"
)

// ErrNotFound is returned when a requested task does not exist.
var ErrNotFound = errors.New("task not found")

// ErrCyclicDependency is returned when adding a dependency would create a cycle.
var ErrCyclicDependency = engine.ErrCyclicDependency

// ErrNothingToUndo is returned when there is no operation to undo.
var ErrNothingToUndo = engine.ErrNothingToUndo

// Client is the public API for TaskKit.
type Client struct {
	engine *engine.Engine
	cfg    *config.Config
}

// NewClient initialises a Client using the provided configuration.
// Use config.Load or config.LoadFrom to obtain a Config.
func NewClient(ctx context.Context, cfg *config.Config) (*Client, error) {
	dbPath := cfg.Database
	if dbPath == "" {
		dbPath = filepath.Join(xdg.DataHome, "taskkit", "tasks.db")
	}

	if dbPath != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
			return nil, fmt.Errorf("client: mkdir: %w", err)
		}
	}

	migrations, err := fs.Sub(db.Migrations, "migration")
	if err != nil {
		return nil, fmt.Errorf("client: migrations: %w", err)
	}

	eng, err := engine.NewEngine(ctx, dbPath, migrations)
	if err != nil {
		return nil, fmt.Errorf("client: engine: %w", err)
	}

	return &Client{
		engine: eng,
		cfg:    cfg,
	}, nil
}

// Close releases the underlying database connection.
func (c *Client) Close() error {
	return c.engine.Close()
}

// Add creates a new pending task and returns it fully hydrated.
func (c *Client) Add(ctx context.Context, task *taskkit.Task) (*taskkit.Task, error) {
	task.TaskID = ulid.Make()
	if task.Status == "" {
		task.Status = taskkit.StatusPending
	}

	if err := validate(task); err != nil {
		return nil, err
	}

	if err := c.engine.CreateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("client: add: %w", err)
	}

	added, err := c.engine.TaskByID(ctx, task.TaskID.String())
	if err != nil {
		return nil, err
	}

	c.applyUrgency([]*taskkit.Task{added})
	c.applyVirtualTags([]*taskkit.Task{added})

	return added, nil
}

// Complete marks a task as completed and frees its display ID.
func (c *Client) Complete(ctx context.Context, id string) error {
	task, err := c.engine.TaskByID(ctx, id)
	if errors.Is(err, engine.ErrNotFound) {
		return ErrNotFound
	}

	if err != nil {
		return fmt.Errorf("client: get task: %w", err)
	}

	return c.engine.Mutate(ctx, id,
		engine.SetStatus(task.Status, taskkit.StatusCompleted),
		engine.SetDisplayID(int64(task.DisplayID), 0),
	)
}

// Modify applies one or more Modifications to a task in a single transaction.
func (c *Client) Modify(ctx context.Context, id string, mods ...Modification) error {
	task, err := c.engine.TaskByID(ctx, id)
	if errors.Is(err, engine.ErrNotFound) {
		return ErrNotFound
	}

	if err != nil {
		return fmt.Errorf("client: get task: %w", err)
	}

	var mutations []engine.Mutation
	for _, mod := range mods {
		muts, err := mod(task)
		if err != nil {
			return fmt.Errorf("client: modification: %w", err)
		}

		mutations = append(mutations, muts...)
	}

	if err := validate(task); err != nil {
		return err
	}

	return c.engine.Mutate(ctx, id, mutations...)
}

// Import restores tasks from a previous export, preserving all IDs and status.
// Validation is skipped since the data is assumed to be from a prior valid state.
// Each task is imported atomically; if one fails the others are unaffected.
func (c *Client) Import(ctx context.Context, tasks []*taskkit.Task) error {
	for _, task := range tasks {
		if err := c.engine.CreateTask(ctx, task); err != nil {
			return fmt.Errorf("client: import task %s: %w", task.TaskID, err)
		}
	}

	return nil
}

// Delete permanently removes a task and all associated records. Irreversible.
func (c *Client) Delete(ctx context.Context, id string) error {
	if _, err := c.engine.TaskByID(ctx, id); errors.Is(err, engine.ErrNotFound) {
		return ErrNotFound
	} else if err != nil {
		return fmt.Errorf("client: get task: %w", err)
	}

	return c.engine.DeleteTask(ctx, id)
}

// Undo reverses the last undoable operation.
func (c *Client) Undo(ctx context.Context) error {
	return c.engine.Undo(ctx)
}

// AddBlockedBy records that taskID is blocked by dependsOnID. Returns
// ErrCyclicDependency if this would create a cycle, ErrNotFound if taskID
// does not exist.
func (c *Client) AddBlockedBy(ctx context.Context, taskID, dependsOnID string) error {
	return c.Modify(ctx, taskID, AddBlockedBy(dependsOnID))
}

// RemoveBlockedBy removes the dependency where taskID is blocked by dependsOnID.
// Returns ErrNotFound if the task does not exist.
func (c *Client) RemoveBlockedBy(ctx context.Context, taskID, dependsOnID string) error {
	return c.Modify(ctx, taskID, RemoveBlockedBy(dependsOnID))
}

// GetHistory returns the full audit history for the task with the given ULID,
// grouped by operation and ordered oldest-first.
func (c *Client) GetHistory(ctx context.Context, taskID string) ([]taskkit.HistoryEntry, error) {
	return c.engine.GetHistory(ctx, taskID)
}

// Get returns tasks matching all provided filters. With no filters, all tasks
// are returned.
func (c *Client) Get(ctx context.Context, filters ...filter.Filter) ([]*taskkit.Task, error) {
	query, args := filter.BuildTaskQuery(filters...)
	tasks, err := c.engine.QueryTasks(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	c.applyUrgency(tasks)
	c.applyVirtualTags(tasks)
	return tasks, nil
}

// Projects returns a summary of all projects that have at least one task,
// sorted by project name.
func (c *Client) Projects(ctx context.Context) ([]*taskkit.ProjectSummary, error) {
	tasks, err := c.Get(ctx, filter.HasProject())
	if err != nil {
		return nil, fmt.Errorf("client: projects: %w", err)
	}

	byName := make(map[string]*taskkit.ProjectSummary)
	for _, t := range tasks {
		s, ok := byName[t.Project]
		if !ok {
			s = &taskkit.ProjectSummary{Name: t.Project}
			byName[t.Project] = s
		}

		switch t.Status {
		case taskkit.StatusPending:
			s.Remaining++
		case taskkit.StatusCompleted:
			s.Completed++
		}
	}

	summaries := make([]*taskkit.ProjectSummary, 0, len(byName))
	for _, s := range byName {
		total := s.Remaining + s.Completed
		if total > 0 {
			s.Percent = s.Completed * 100 / total
		}

		summaries = append(summaries, s)
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Name < summaries[j].Name
	})

	return summaries, nil
}

// applyUrgency populates the Urgency field on each task in place.
func (c *Client) applyUrgency(tasks []*taskkit.Task) {
	urgency.ScoreAll(tasks, c.cfg.Urgency, time.Now())
}

// UrgencyBreakdown returns the per-component urgency breakdown for a task.
func (c *Client) UrgencyBreakdown(task *taskkit.Task) []taskkit.UrgencyComponent {
	return urgency.Components(task, c.cfg.Urgency, time.Now())
}

// applyVirtualTags populates the VirtualTags field on each task in place.
func (c *Client) applyVirtualTags(tasks []*taskkit.Task) {
	now := time.Now()
	for _, t := range tasks {
		t.VirtualTags = filter.ComputeVirtualTags(t, now)
	}
}
