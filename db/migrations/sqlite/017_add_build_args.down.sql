-- Reverse migration: remove build args columns from workspaces, ecosystems, and domains
-- SQLite does not support DROP COLUMN in older versions; this creates replacement tables.
-- Note: Since SQLite 3.35.0+ supports DROP COLUMN, use that syntax here.

ALTER TABLE workspaces DROP COLUMN build_config;
ALTER TABLE ecosystems DROP COLUMN build_args;
ALTER TABLE domains DROP COLUMN build_args;
