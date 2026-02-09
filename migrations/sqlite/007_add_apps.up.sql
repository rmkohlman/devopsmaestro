-- 007_add_apps.up.sql
-- Adds apps table for codebase/application within a domain
-- Part of the new object hierarchy: Ecosystem -> Domain -> App -> Workspace

CREATE TABLE IF NOT EXISTS apps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    domain_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    path TEXT NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (domain_id) REFERENCES domains(id) ON DELETE CASCADE,
    UNIQUE(domain_id, name)
);

-- Index for domain lookups
CREATE INDEX IF NOT EXISTS idx_apps_domain ON apps(domain_id);

-- Index for name lookups within domain
CREATE INDEX IF NOT EXISTS idx_apps_name ON apps(name);

-- Index for path lookups
CREATE INDEX IF NOT EXISTS idx_apps_path ON apps(path);
