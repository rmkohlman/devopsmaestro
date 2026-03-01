-- 004_add_terminal_fields.up.sql
-- Add terminal configuration fields to workspaces table

ALTER TABLE workspaces ADD COLUMN terminal_prompt TEXT;
ALTER TABLE workspaces ADD COLUMN terminal_plugins TEXT;
ALTER TABLE workspaces ADD COLUMN terminal_package TEXT;
