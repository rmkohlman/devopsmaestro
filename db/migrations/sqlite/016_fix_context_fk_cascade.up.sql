-- Migration 016: Fix context table foreign keys to use ON DELETE SET NULL
-- Without this, deleting an active workspace/app/domain/ecosystem fails
-- with "FOREIGN KEY constraint failed" because the context table still
-- references the entity being deleted.
--
-- SQLite does not support ALTER TABLE to modify constraints, so we must
-- rebuild the table using the standard create-copy-drop-rename pattern.

CREATE TABLE context_new (
    id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    active_workspace_id INTEGER,
    active_ecosystem_id INTEGER REFERENCES ecosystems(id) ON DELETE SET NULL,
    active_domain_id INTEGER REFERENCES domains(id) ON DELETE SET NULL,
    active_app_id INTEGER REFERENCES apps(id) ON DELETE SET NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (active_workspace_id) REFERENCES workspaces(id) ON DELETE SET NULL
);

INSERT INTO context_new SELECT * FROM context;
DROP TABLE context;
ALTER TABLE context_new RENAME TO context;
