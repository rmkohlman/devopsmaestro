-- Down migration: Restore original context table without ON DELETE SET NULL
-- This reverts to the original 001_init schema for the context table.

CREATE TABLE context_old (
    id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    active_workspace_id INTEGER,
    active_ecosystem_id INTEGER REFERENCES ecosystems(id),
    active_domain_id INTEGER REFERENCES domains(id),
    active_app_id INTEGER REFERENCES apps(id),
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (active_workspace_id) REFERENCES workspaces(id)
);

INSERT INTO context_old SELECT * FROM context;
DROP TABLE context;
ALTER TABLE context_old RENAME TO context;
