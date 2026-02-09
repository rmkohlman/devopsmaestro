-- Rollback migration 009: Restore workspace foreign key to project_id
-- Note: This will lose any workspace data as we cannot recover project_id

-- Create old-style workspaces table
CREATE TABLE workspaces_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    image_name TEXT NOT NULL,
    container_id TEXT,
    status TEXT NOT NULL DEFAULT 'stopped',
    nvim_structure TEXT,
    nvim_plugins TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, name)
);

-- Drop new table
DROP TABLE IF EXISTS workspaces;

-- Rename old table
ALTER TABLE workspaces_old RENAME TO workspaces;
