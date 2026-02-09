-- 006_add_domains.up.sql
-- Adds domains table for bounded contexts within an ecosystem
-- Part of the new object hierarchy: Ecosystem -> Domain -> App -> Workspace

CREATE TABLE IF NOT EXISTS domains (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ecosystem_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (ecosystem_id) REFERENCES ecosystems(id) ON DELETE CASCADE,
    UNIQUE(ecosystem_id, name)
);

-- Index for ecosystem lookups
CREATE INDEX IF NOT EXISTS idx_domains_ecosystem ON domains(ecosystem_id);

-- Index for name lookups within ecosystem
CREATE INDEX IF NOT EXISTS idx_domains_name ON domains(name);
