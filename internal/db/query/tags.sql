-- name: InsertTag :exec
INSERT INTO task_tags (task_id, tag) VALUES (?, ?);

-- name: DeleteTag :exec
DELETE FROM task_tags WHERE task_id = ? AND tag = ?;

-- name: DeleteTagsByTaskID :exec
DELETE FROM task_tags WHERE task_id = ?;

-- name: GetTagsByTaskID :many
SELECT tag FROM task_tags WHERE task_id = ? ORDER BY tag;
