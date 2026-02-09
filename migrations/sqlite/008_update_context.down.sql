-- 008_update_context.down.sql
-- Removes new hierarchy columns from context table
-- Note: SQLite doesn't support DROP COLUMN before version 3.35.0
-- This creates a new table without the columns and migrates data

-- Create temporary table with original schema
CREATE TABLE context_backup (
    id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    active_project_id INTEGER,
    active_workspace_id INTEGER,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (active_project_id) REFERENCES projects(id),
    FOREIGN KEY (active_workspace_id) REFERENCES workspaces(id)
);

-- Copy existing data
INSERT INTO context_backup (id, active_project_id, active_workspace_id, updated_at)
SELECT id, active_project_id, active_workspace_id, updated_at FROM context;

-- Drop original table
DROP TABLE context;

-- Rename backup to original
ALTER TABLE context_backup RENAME TO context;
