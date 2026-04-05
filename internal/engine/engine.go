// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package engine

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"time"

	migrate "github.com/golang-migrate/migrate/v4"
	migratesqlite "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/hugginsio/taskkit"
	"github.com/hugginsio/taskkit/internal/db"
	"github.com/oklog/ulid/v2"
	_ "modernc.org/sqlite"
)

const iso8601 = time.RFC3339

// ErrNotFound is returned when a requested task does not exist.
var ErrNotFound = errors.New("task not found")

// Engine is the core service layer. All reads and writes flow through it.
type Engine struct {
	db *sql.DB
}

// NewEngine opens the SQLite database at dbPath, applies any pending
// migrations, and returns an Engine ready for use.
func NewEngine(ctx context.Context, dbPath string, migrations fs.FS) (*Engine, error) {
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("engine: open: %w", err)
	}

	for _, pragma := range []string{
		"PRAGMA foreign_keys = ON;",
		"PRAGMA journal_mode = WAL;",
		"PRAGMA busy_timeout = 5000;",
	} {
		if _, err := sqlDB.ExecContext(ctx, pragma); err != nil {
			return nil, fmt.Errorf("engine: pragma: %w", err)
		}
	}

	// SQLite is not safe for concurrent writes and in-memory databases are
	// per-connection. A single connection avoids both problems.
	sqlDB.SetMaxOpenConns(1)

	src, err := iofs.New(migrations, ".")
	if err != nil {
		return nil, fmt.Errorf("engine: migration source: %w", err)
	}

	driver, err := migratesqlite.WithInstance(sqlDB, &migratesqlite.Config{})
	if err != nil {
		return nil, fmt.Errorf("engine: migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "sqlite", driver)
	if err != nil {
		return nil, fmt.Errorf("engine: migrate: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("engine: migrate up: %w", err)
	}

	return &Engine{db: sqlDB}, nil
}

// Close closes the underlying database connection.
func (e *Engine) Close() error {
	return e.db.Close()
}

// TaskByID returns the task with the given ULID string, fully hydrated.
func (e *Engine) TaskByID(ctx context.Context, id string) (*taskkit.Task, error) {
	q := db.New(e.db)
	row, err := q.GetTaskByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("engine: get task: %w", err)
	}

	return e.hydrate(ctx, q, row)
}

// QueryTasks executes a SQL query against the tasks table and returns matching
// tasks, fully hydrated. query and args are expected to be produced by
// filter.BuildTaskQuery.
func (e *Engine) QueryTasks(ctx context.Context, query string, args ...any) ([]*taskkit.Task, error) {
	rows, err := e.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("engine: query: %w", err)
	}

	var dbRows []db.Task
	for rows.Next() {
		var row db.Task
		if err := rows.Scan(
			&row.TaskID, &row.DisplayID, &row.Project, &row.Description,
			&row.Status, &row.Deadline, &row.Scheduled, &row.Wait,
		); err != nil {
			rows.Close()
			return nil, fmt.Errorf("engine: scan: %w", err)
		}

		dbRows = append(dbRows, row)
	}

	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("engine: rows close: %w", err)
	}

	q := db.New(e.db)
	tasks := make([]*taskkit.Task, 0, len(dbRows))
	for _, row := range dbRows {
		t, err := e.hydrate(ctx, q, row)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

// CreateTask inserts a new task, assigns it a display ID, and logs a create
// history entry. The task must have TaskID set.
func (e *Engine) CreateTask(ctx context.Context, task *taskkit.Task) error {
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("engine: begin tx: %w", err)
	}

	defer tx.Rollback()

	q := db.New(tx)
	opID := ulid.Make().String()

	displayID, err := nextDisplayID(ctx, q)
	if err != nil {
		return fmt.Errorf("engine: create task: %w", err)
	}

	if err := q.CreateTask(ctx, db.CreateTaskParams{
		TaskID:      task.TaskID.String(),
		DisplayID:   displayID,
		Project:     toNullString(task.Project),
		Description: task.Description,
		Status:      string(task.Status),
		Deadline:    toNullTime(task.Deadline),
		Scheduled:   toNullTime(task.Scheduled),
		Wait:        toNullTime(task.Wait),
	}); err != nil {
		return fmt.Errorf("engine: insert task: %w", err)
	}

	for _, tag := range task.Tags {
		if err := q.InsertTag(ctx, db.InsertTagParams{TaskID: task.TaskID.String(), Tag: tag}); err != nil {
			return fmt.Errorf("engine: insert tag: %w", err)
		}
	}

	if err := q.InsertHistory(ctx, db.InsertHistoryParams{
		OperationID: opID,
		TaskID:      task.TaskID.String(),
		Action:      "create",
	}); err != nil {
		return fmt.Errorf("engine: history: %w", err)
	}

	return tx.Commit()
}

// Mutate applies one or more mutations to a task within a single transaction,
// all sharing one operation_id for grouped undo.
func (e *Engine) Mutate(ctx context.Context, taskID string, muts ...Mutation) error {
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("engine: begin tx: %w", err)
	}

	defer tx.Rollback()

	q := db.New(tx)
	opID := ulid.Make().String()

	for _, mut := range muts {
		if err := mut(ctx, q, taskID, opID); err != nil {
			return fmt.Errorf("engine: mutation: %w", err)
		}
	}

	return tx.Commit()
}

// DeleteTask permanently removes a task and all associated records. Irreversible.
// ON DELETE CASCADE handles task_tags and task_dependencies automatically.
func (e *Engine) DeleteTask(ctx context.Context, id string) error {
	_, err := e.db.ExecContext(ctx, "DELETE FROM tasks WHERE task_id = ?", id)
	if err != nil {
		return fmt.Errorf("engine: delete task: %w", err)
	}

	return nil
}

// NextDisplayID returns the smallest positive integer not currently in use
// by an active (pending or waiting) task.
func (e *Engine) NextDisplayID(ctx context.Context) (int64, error) {
	return nextDisplayID(ctx, db.New(e.db))
}

// nextDisplayID computes the next available display ID within the provided
// Queries, so it can be called inside a transaction.
func nextDisplayID(ctx context.Context, q *db.Queries) (int64, error) {
	active, err := q.GetActiveDisplayIDs(ctx)
	if err != nil {
		return 0, fmt.Errorf("get display ids: %w", err)
	}

	inUse := make(map[int64]bool, len(active))
	for _, id := range active {
		inUse[id] = true
	}

	var i int64 = 1
	for inUse[i] {
		i++
	}

	return i, nil
}

// GetHistory returns the full audit history for the task with the given ULID,
// grouped by operation and ordered oldest-first.
func (e *Engine) GetHistory(ctx context.Context, taskID string) ([]taskkit.HistoryEntry, error) {
	rows, err := db.New(e.db).GetHistoryByTaskID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("engine: get history: %w", err)
	}

	// Group rows by operation_id, preserving insertion order.
	var order []string
	byOp := make(map[string][]taskkit.HistoryChange)
	opAction := make(map[string]string)
	opTime := make(map[string]time.Time)

	for _, row := range rows {
		opID := row.OperationID
		if _, seen := byOp[opID]; !seen {
			order = append(order, opID)
			opAction[opID] = row.Action
			if parsed, err := ulid.Parse(opID); err == nil {
				opTime[opID] = ulid.Time(parsed.Time()).UTC()
			}
		}

		if row.Field.Valid {
			byOp[opID] = append(byOp[opID], taskkit.HistoryChange{
				Field:    row.Field.String,
				OldValue: row.OldValue.String,
				NewValue: row.NewValue.String,
			})
		}
	}

	entries := make([]taskkit.HistoryEntry, len(order))
	for i, opID := range order {
		entries[i] = taskkit.HistoryEntry{
			OperationID: opID,
			Action:      opAction[opID],
			Time:        opTime[opID],
			Changes:     byOp[opID],
		}
	}

	return entries, nil
}

// HasDependencyPath reports whether there is a transitive path from fromID to
// toID following depends_on links. Used to detect cycles before inserting a
// new dependency: adding (A depends_on B) is safe iff !HasDependencyPath(B, A).
func (e *Engine) HasDependencyPath(ctx context.Context, fromID, toID string) (bool, error) {
	q := db.New(e.db)
	visited := make(map[string]bool)
	queue := []string{fromID}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}

		visited[current] = true

		deps, err := q.GetBlockedByTaskID(ctx, current)
		if err != nil {
			return false, fmt.Errorf("engine: dependency path: %w", err)
		}

		for _, dep := range deps {
			if dep == toID {
				return true, nil
			}

			if !visited[dep] {
				queue = append(queue, dep)
			}
		}
	}

	return false, nil
}

// hydrate fetches tags, dependencies, and timestamps for a db.Task row and
// maps it to a taskkit.Task. Dependency tasks are hydrated one level deep
// (their own BlockedBy/Blocking lists are not expanded) to prevent cycles.
func (e *Engine) hydrate(ctx context.Context, q *db.Queries, row db.Task) (*taskkit.Task, error) {
	tags, err := q.GetTagsByTaskID(ctx, row.TaskID)
	if err != nil {
		return nil, fmt.Errorf("engine: get tags: %w", err)
	}

	blockedByIDs, err := q.GetActiveBlockedByTaskID(ctx, row.TaskID)
	if err != nil {
		return nil, fmt.Errorf("engine: get blocked_by: %w", err)
	}

	blockedBy := make([]*taskkit.Task, 0, len(blockedByIDs))
	for _, depID := range blockedByIDs {
		depRow, err := q.GetTaskByID(ctx, depID)
		if err != nil {
			return nil, fmt.Errorf("engine: get blocked_by task: %w", err)
		}

		dep, err := e.hydrateShallow(ctx, q, depRow)
		if err != nil {
			return nil, err
		}

		blockedBy = append(blockedBy, dep)
	}

	var blocking []*taskkit.Task
	if isActive(taskkit.Status(row.Status)) {
		blockingIDs, err := q.GetActiveBlockingTaskID(ctx, row.TaskID)
		if err != nil {
			return nil, fmt.Errorf("engine: get blocking: %w", err)
		}

		blocking = make([]*taskkit.Task, 0, len(blockingIDs))
		for _, depID := range blockingIDs {
			depRow, err := q.GetTaskByID(ctx, depID)
			if err != nil {
				return nil, fmt.Errorf("engine: get blocking task: %w", err)
			}
			dep, err := e.hydrateShallow(ctx, q, depRow)
			if err != nil {
				return nil, err
			}
			blocking = append(blocking, dep)
		}
	}

	taskID, err := ulid.Parse(row.TaskID)
	if err != nil {
		return nil, fmt.Errorf("engine: parse task id: %w", err)
	}

	created := ulid.Time(taskID.Time()).UTC()
	modified := created

	latest, err := q.GetLatestHistoryForTask(ctx, row.TaskID)
	if err == nil {
		if opID, err := ulid.Parse(latest.OperationID); err == nil {
			modified = ulid.Time(opID.Time()).UTC()
		}
	}

	t := &taskkit.Task{
		TaskID:      taskID,
		DisplayID:   int(row.DisplayID),
		Description: row.Description,
		Status:      taskkit.Status(row.Status),
		Tags:        tags,
		BlockedBy:   blockedBy,
		Blocking:    blocking,
		Created:     created,
		Modified:    modified,
	}

	if row.Project.Valid {
		t.Project = row.Project.String
	}

	if row.Deadline.Valid {
		if parsed, err := time.Parse(iso8601, row.Deadline.String); err == nil {
			t.Deadline = &parsed
		}
	}

	if row.Scheduled.Valid {
		if parsed, err := time.Parse(iso8601, row.Scheduled.String); err == nil {
			t.Scheduled = &parsed
		}
	}

	if row.Wait.Valid {
		if parsed, err := time.Parse(iso8601, row.Wait.String); err == nil {
			t.Wait = &parsed
		}
	}

	return t, nil
}

// hydrateShallow maps a db.Task row to a taskkit.Task without expanding its
// BlockedBy or Blocking lists. Used for dependency nodes to prevent cycles.
func (e *Engine) hydrateShallow(ctx context.Context, q *db.Queries, row db.Task) (*taskkit.Task, error) {
	tags, err := q.GetTagsByTaskID(ctx, row.TaskID)
	if err != nil {
		return nil, fmt.Errorf("engine: get tags: %w", err)
	}

	taskID, err := ulid.Parse(row.TaskID)
	if err != nil {
		return nil, fmt.Errorf("engine: parse task id: %w", err)
	}

	created := ulid.Time(taskID.Time()).UTC()
	modified := created

	latest, err := q.GetLatestHistoryForTask(ctx, row.TaskID)
	if err == nil {
		if opID, err := ulid.Parse(latest.OperationID); err == nil {
			modified = ulid.Time(opID.Time()).UTC()
		}
	}

	t := &taskkit.Task{
		TaskID:      taskID,
		DisplayID:   int(row.DisplayID),
		Description: row.Description,
		Status:      taskkit.Status(row.Status),
		Tags:        tags,
		Created:     created,
		Modified:    modified,
	}

	if row.Project.Valid {
		t.Project = row.Project.String
	}

	if row.Deadline.Valid {
		if parsed, err := time.Parse(iso8601, row.Deadline.String); err == nil {
			t.Deadline = &parsed
		}
	}

	if row.Scheduled.Valid {
		if parsed, err := time.Parse(iso8601, row.Scheduled.String); err == nil {
			t.Scheduled = &parsed
		}
	}

	if row.Wait.Valid {
		if parsed, err := time.Parse(iso8601, row.Wait.String); err == nil {
			t.Wait = &parsed
		}
	}

	return t, nil
}

func isActive(s taskkit.Status) bool {
	return s == taskkit.StatusPending || s == taskkit.StatusWaiting
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}

	return sql.NullString{String: s, Valid: true}
}

func toNullTime(t *time.Time) sql.NullString {
	if t == nil {
		return sql.NullString{}
	}

	return sql.NullString{String: t.UTC().Format(iso8601), Valid: true}
}
