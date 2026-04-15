-- Best-effort reversal: restore "waiting" status for pending tasks that have
-- a future wait date. Tasks that were previously "waiting" but had a past wait
-- date cannot be distinguished from regular pending tasks and are left as-is.
UPDATE tasks SET status = 'waiting' WHERE status = 'pending' AND wait IS NOT NULL AND wait > datetime('now');
