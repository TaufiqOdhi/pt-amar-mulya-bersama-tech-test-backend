DROP INDEX IF EXISTS idx_tasks_deleted_at;
ALTER TABLE tasks DROP COLUMN IF EXISTS deleted_at;
