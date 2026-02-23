-- 002_add_workspace_theme.down.sql
-- Remove theme column from workspaces table

-- SQLite doesn't support DROP COLUMN, so we need to recreate the table
-- Create temporary table without theme column
CREATE TABLE workspaces_backup (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id INTEGER NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    image_name TEXT NOT NULL,
    container_id TEXT,
    status TEXT NOT NULL DEFAULT 'stopped',
    nvim_structure TEXT,
    nvim_plugins TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(app_id, name)
);

-- Copy data (excluding theme column)
INSERT INTO workspaces_backup (id, app_id, name, description, image_name, container_id, status, nvim_structure, nvim_plugins, created_at, updated_at)
SELECT id, app_id, name, description, image_name, container_id, status, nvim_structure, nvim_plugins, created_at, updated_at
FROM workspaces;

-- Drop original table and rename backup
DROP TABLE workspaces;
ALTER TABLE workspaces_backup RENAME TO workspaces;

-- Recreate indexes if any existed
CREATE INDEX IF NOT EXISTS idx_workspaces_app ON workspaces(app_id);
CREATE INDEX IF NOT EXISTS idx_workspaces_name ON workspaces(name);