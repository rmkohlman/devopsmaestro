-- 005_add_ecosystems.up.sql
-- Adds ecosystems table for top-level platform grouping
-- Part of the new object hierarchy: Ecosystem -> Domain -> App -> Workspace

CREATE TABLE IF NOT EXISTS ecosystems (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Index for name lookups
CREATE INDEX IF NOT EXISTS idx_ecosystems_name ON ecosystems(name);
