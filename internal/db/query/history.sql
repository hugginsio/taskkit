-- name: InsertHistory :exec
INSERT INTO history (operation_id, task_id, action, field, old_value, new_value)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetHistoryByOperationID :many
SELECT * FROM history WHERE operation_id = ? ORDER BY id;

-- name: GetLatestOperationID :one
-- Returns the operation_id of the most recent operation, for use by undo.
SELECT operation_id FROM history ORDER BY id DESC LIMIT 1;

-- name: GetHistoryByTaskID :many
SELECT * FROM history WHERE task_id = ? ORDER BY id;

-- name: GetLatestHistoryForTask :one
-- Returns the most recent history row for a task; used to derive modified time.
SELECT * FROM history WHERE task_id = ? ORDER BY id DESC LIMIT 1;
