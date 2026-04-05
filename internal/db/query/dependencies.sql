-- name: InsertDependency :exec
INSERT INTO task_dependencies (task_id, depends_on) VALUES (?, ?);

-- name: DeleteDependency :exec
DELETE FROM task_dependencies WHERE task_id = ? AND depends_on = ?;

-- name: DeleteDependenciesByTaskID :exec
DELETE FROM task_dependencies WHERE task_id = ? OR depends_on = ?;

-- name: GetBlockedByTaskID :many
-- Returns all task IDs that block the given task (regardless of status).
SELECT depends_on FROM task_dependencies WHERE task_id = ? ORDER BY depends_on;

-- name: GetBlockingTaskID :many
-- Returns all task IDs that the given task blocks (regardless of status).
SELECT task_id FROM task_dependencies WHERE depends_on = ? ORDER BY task_id;

-- name: GetActiveBlockedByTaskID :many
-- Returns task IDs of blockers that are still active (pending or waiting).
SELECT td.depends_on
FROM task_dependencies td
JOIN tasks t ON t.task_id = td.depends_on
WHERE td.task_id = ?
  AND t.status IN ('pending', 'waiting')
ORDER BY td.depends_on;

-- name: GetActiveBlockingTaskID :many
-- Returns task IDs of dependents that are still active (pending or waiting).
SELECT td.task_id
FROM task_dependencies td
JOIN tasks t ON t.task_id = td.task_id
WHERE td.depends_on = ?
  AND t.status IN ('pending', 'waiting')
ORDER BY td.task_id;
