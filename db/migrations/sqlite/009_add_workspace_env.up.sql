-- Add env column to workspaces for environment variable configuration (WI-5)
ALTER TABLE workspaces ADD COLUMN env TEXT NOT NULL DEFAULT '{}';
