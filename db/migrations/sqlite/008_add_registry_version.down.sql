-- 008_add_registry_version.down.sql
-- Remove version column from registries table

-- SQLite doesn't support DROP COLUMN, so we recreate the table
CREATE TABLE registries_backup AS SELECT
    id, name, type, enabled, lifecycle, port, storage, idle_timeout, config, description, status, created_at, updated_at
FROM registries;

DROP TABLE registries;

CREATE TABLE registries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL CHECK(type IN ('zot', 'athens', 'devpi', 'verdaccio', 'squid')),
    enabled BOOLEAN NOT NULL DEFAULT 1,
    lifecycle TEXT NOT NULL DEFAULT 'manual' CHECK(lifecycle IN ('persistent', 'on-demand', 'manual')),
    port INTEGER NOT NULL UNIQUE,
    storage TEXT NOT NULL,
    idle_timeout INTEGER DEFAULT 1800,
    config TEXT,
    description TEXT,
    status TEXT DEFAULT 'stopped' CHECK(status IN ('running', 'stopped', 'starting', 'error')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO registries SELECT * FROM registries_backup;
DROP TABLE registries_backup;

CREATE INDEX IF NOT EXISTS idx_registries_name ON registries(name);
CREATE INDEX IF NOT EXISTS idx_registries_type ON registries(type);
CREATE INDEX IF NOT EXISTS idx_registries_port ON registries(port);
CREATE INDEX IF NOT EXISTS idx_registries_status ON registries(status);
CREATE INDEX IF NOT EXISTS idx_registries_lifecycle ON registries(lifecycle);
