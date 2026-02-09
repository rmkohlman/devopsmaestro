-- Migration 009: Change workspace foreign key from project_id to app_id
-- This is a breaking change - existing workspaces will be orphaned

-- Create new workspaces table with app_id
CREATE TABLE workspaces_new (
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

-- Drop old table (no data migration since no users yet)
DROP TABLE IF EXISTS workspaces;

-- Rename new table
ALTER TABLE workspaces_new RENAME TO workspaces;

-- Also recreate workspace_plugins table to ensure foreign key constraint
CREATE TABLE IF NOT EXISTS workspace_plugins_new (
    workspace_id INTEGER NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    plugin_id INTEGER NOT NULL REFERENCES nvim_plugins(id) ON DELETE CASCADE,
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (workspace_id, plugin_id)
);

-- Copy any existing workspace_plugins data (if table exists)
INSERT OR IGNORE INTO workspace_plugins_new (workspace_id, plugin_id, enabled, created_at)
    SELECT workspace_id, plugin_id, enabled, created_at FROM workspace_plugins 
    WHERE workspace_id IN (SELECT id FROM workspaces);

-- Replace old table
DROP TABLE IF EXISTS workspace_plugins;
ALTER TABLE workspace_plugins_new RENAME TO workspace_plugins;
