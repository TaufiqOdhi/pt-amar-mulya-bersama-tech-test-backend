CREATE INDEX IF NOT EXISTS idx_tasks_user_deleted_created ON tasks(user_id, deleted_at, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tasks_user_status ON tasks(user_id, status);
