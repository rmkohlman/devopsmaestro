-- 021_add_package_hierarchy.up.sql
-- Add nvim_package and terminal_package columns to entity tables
-- for hierarchical package resolution (workspace → app → domain → ecosystem → global).

ALTER TABLE ecosystems ADD COLUMN nvim_package TEXT;
ALTER TABLE ecosystems ADD COLUMN terminal_package TEXT;
ALTER TABLE domains ADD COLUMN nvim_package TEXT;
ALTER TABLE domains ADD COLUMN terminal_package TEXT;
ALTER TABLE apps ADD COLUMN nvim_package TEXT;
ALTER TABLE apps ADD COLUMN terminal_package TEXT;
-- workspaces already has terminal_package (from migration 004); add nvim_package only
ALTER TABLE workspaces ADD COLUMN nvim_package TEXT;
