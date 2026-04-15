-- Migrate any tasks stored with the legacy "waiting" status to "pending".
-- The wait date is already present on these rows; waiting state is now
-- determined solely by the presence of a wait date, not by a stored status.
UPDATE tasks SET status = 'pending' WHERE status = 'waiting';
