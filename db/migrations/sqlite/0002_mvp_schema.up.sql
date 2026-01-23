-- Add path column to projects table
ALTER TABLE projects ADD COLUMN path TEXT NOT NULL DEFAULT '';

-- Create workspaces table
CREATE TABLE IF NOT EXISTS workspaces (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    language TEXT,
    image_name TEXT,
    container_id TEXT,
    status TEXT NOT NULL DEFAULT 'idle',
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    UNIQUE(project_id, name)
);

-- Create context table (single row to store active context)
CREATE TABLE IF NOT EXISTS context (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    active_project_id INTEGER,
    active_workspace_id INTEGER,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (active_project_id) REFERENCES projects(id) ON DELETE SET NULL,
    FOREIGN KEY (active_workspace_id) REFERENCES workspaces(id) ON DELETE SET NULL
);

-- Insert initial context row
INSERT OR IGNORE INTO context (id, active_project_id, active_workspace_id) VALUES (1, NULL, NULL);
