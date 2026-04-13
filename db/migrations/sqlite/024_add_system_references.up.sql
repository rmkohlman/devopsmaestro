-- 024_add_system_references.up.sql
-- Add system_id FK to apps table and active_system_id to context table.
-- Part of the System entity integration: Ecosystem -> Domain -> System -> App -> Workspace

-- Add system_id to apps (nullable — existing apps have NULL system_id)
ALTER TABLE apps ADD COLUMN system_id INTEGER REFERENCES systems(id) ON DELETE CASCADE;

-- Add active_system_id to context (nullable — SET NULL on system deletion)
ALTER TABLE context ADD COLUMN active_system_id INTEGER REFERENCES systems(id) ON DELETE SET NULL;

-- Index for efficient lookups by system
CREATE INDEX IF NOT EXISTS idx_apps_system_id ON apps(system_id);
