-- 013_add_defaults.up.sql
-- Adds defaults table for storing global defaults (plugins, theme, prompt)

CREATE TABLE IF NOT EXISTS defaults (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Index for faster lookups (though key is primary key, explicit index helps with queries)
CREATE INDEX IF NOT EXISTS idx_defaults_updated_at ON defaults(updated_at);