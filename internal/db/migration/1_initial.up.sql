-- tasks is the source of truth for all task data.
CREATE TABLE tasks (
    task_id     TEXT    NOT NULL PRIMARY KEY,
    display_id  INTEGER NOT NULL,
    project     TEXT,
    description TEXT    NOT NULL,
    status      TEXT    NOT NULL DEFAULT 'pending'
                        CHECK(status IN ('pending', 'waiting', 'completed', 'removed')),
    deadline    TEXT,  -- ISO8601
    scheduled   TEXT,  -- ISO8601
    wait        TEXT   -- ISO8601
) STRICT;

CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_project ON tasks(project);

-- task_tags stores tags in a normalized join table rather than as a JSON column.
CREATE TABLE task_tags (
    task_id TEXT NOT NULL REFERENCES tasks(task_id) ON DELETE CASCADE,
    tag     TEXT NOT NULL,
    PRIMARY KEY (task_id, tag)
) STRICT;

CREATE INDEX idx_task_tags_tag ON task_tags(tag);

-- task_dependencies stores blocking relationships between tasks.
-- task_id depends on depends_on (i.e., depends_on blocks task_id).
CREATE TABLE task_dependencies (
    task_id    TEXT NOT NULL REFERENCES tasks(task_id) ON DELETE CASCADE,
    depends_on TEXT NOT NULL REFERENCES tasks(task_id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, depends_on),
    CONSTRAINT no_self_dependency CHECK (task_id != depends_on)
) STRICT;

-- history is an append-only changelog for auditing and undo.
-- Each row records a single field change; rows from the same logical operation share an operation_id.
CREATE TABLE history (
    id           INTEGER NOT NULL PRIMARY KEY,
    operation_id TEXT    NOT NULL,
    task_id      TEXT    NOT NULL,
    action       TEXT    NOT NULL CHECK(action IN ('create', 'update', 'delete')),
    field        TEXT,   -- NULL for create/delete
    old_value    TEXT,   -- NULL for create
    new_value    TEXT    -- NULL for delete
) STRICT;

CREATE INDEX idx_history_operation_id ON history(operation_id);
CREATE INDEX idx_history_task_id ON history(task_id);
