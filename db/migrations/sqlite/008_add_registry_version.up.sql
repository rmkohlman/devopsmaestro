-- 008_add_registry_version.up.sql
-- Add version column to registries table for declarative version management

ALTER TABLE registries ADD COLUMN version TEXT NOT NULL DEFAULT '';

-- Backfill existing zot registries with the current default version
UPDATE registries SET version = '2.1.15' WHERE type = 'zot' AND version = '';
