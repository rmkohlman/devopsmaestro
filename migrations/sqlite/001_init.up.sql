CREATE TABLE IF NOT EXISTS projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    path TEXT NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS workspaces (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    project_id INTEGER NOT NULL,
    description TEXT,
    image_name TEXT NOT NULL,
    container_id TEXT,
    status TEXT DEFAULT 'stopped',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id),
    UNIQUE(name, project_id)
);

CREATE TABLE IF NOT EXISTS context (
    id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    active_project_id INTEGER,
    active_workspace_id INTEGER,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (active_project_id) REFERENCES projects(id),
    FOREIGN KEY (active_workspace_id) REFERENCES workspaces(id)
);

-- Initialize context row
INSERT OR IGNORE INTO context (id) VALUES (1);
