-- 002_add_workspace_theme.up.sql
-- Add theme column to workspaces table for consistency with other hierarchy levels

ALTER TABLE workspaces ADD COLUMN theme TEXT;
