CREATE TABLE IF NOT EXISTS registry_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    registry_id INTEGER NOT NULL,
    revision INTEGER NOT NULL,
    config TEXT NOT NULL,
    enabled BOOLEAN NOT NULL,
    lifecycle TEXT NOT NULL,
    port INTEGER NOT NULL,
    storage TEXT NOT NULL,
    idle_timeout INTEGER,
    action TEXT NOT NULL,
    status TEXT NOT NULL,
    user TEXT,
    error_message TEXT,
    previous_revision INTEGER,
    registry_version TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    FOREIGN KEY (registry_id) REFERENCES registries(id) ON DELETE CASCADE,
    UNIQUE(registry_id, revision)
);

CREATE INDEX IF NOT EXISTS idx_registry_history_registry ON registry_history(registry_id);
CREATE INDEX IF NOT EXISTS idx_registry_history_status ON registry_history(status);
