-- 002_add_workspace_theme.up.sql
-- Add theme column to workspaces table

ALTER TABLE workspaces ADD COLUMN theme TEXT;