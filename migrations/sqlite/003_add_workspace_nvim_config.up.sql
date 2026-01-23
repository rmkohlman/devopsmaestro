-- Add nvim configuration columns to workspaces table
ALTER TABLE workspaces ADD COLUMN nvim_structure TEXT DEFAULT '';
ALTER TABLE workspaces ADD COLUMN nvim_plugins TEXT DEFAULT '';
