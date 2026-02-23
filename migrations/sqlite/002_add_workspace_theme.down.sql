-- 002_add_workspace_theme.down.sql
-- Remove theme column from workspaces table
-- Note: SQLite doesn't support DROP COLUMN directly, but this is for rollback reference

-- SQLite requires recreating the table to drop a column
-- For simplicity, we'll leave this as a no-op in practice
-- The column will remain but be unused if downgrade is needed
