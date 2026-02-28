-- 003_add_git_repo_fk.down.sql
-- Rollback git_repo_id foreign keys

-- SQLite doesn't support DROP COLUMN directly, but since this is a new column
-- that starts as NULL, we can keep the rollback simple
-- For a more complete rollback, we'd need to recreate the tables

DROP INDEX IF EXISTS idx_workspaces_git_repo;
DROP INDEX IF EXISTS idx_apps_git_repo;

-- Note: To fully remove the columns, you would need to:
-- 1. Create new tables without git_repo_id
-- 2. Copy data
-- 3. Drop old tables
-- 4. Rename new tables
-- For now, we leave the columns as they start NULL and don't break anything
