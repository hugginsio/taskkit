-- name: CreateTask :exec
INSERT INTO tasks (task_id, display_id, project, description, status, deadline, scheduled, wait)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetTaskByID :one
SELECT * FROM tasks WHERE task_id = ?;

-- name: ListActiveTasks :many
SELECT * FROM tasks
WHERE status IN ('pending', 'waiting')
ORDER BY task_id;

-- name: GetActiveDisplayIDs :many
SELECT display_id FROM tasks
WHERE status IN ('pending', 'waiting');

-- name: UpdateTaskDisplayID :exec
UPDATE tasks SET display_id = ? WHERE task_id = ?;

-- name: UpdateTaskDescription :exec
UPDATE tasks SET description = ? WHERE task_id = ?;

-- name: UpdateTaskStatus :exec
UPDATE tasks SET status = ? WHERE task_id = ?;

-- name: UpdateTaskProject :exec
UPDATE tasks SET project = ? WHERE task_id = ?;

-- name: UpdateTaskDeadline :exec
UPDATE tasks SET deadline = ? WHERE task_id = ?;

-- name: UpdateTaskScheduled :exec
UPDATE tasks SET scheduled = ? WHERE task_id = ?;

-- name: UpdateTaskWait :exec
UPDATE tasks SET wait = ? WHERE task_id = ?;
