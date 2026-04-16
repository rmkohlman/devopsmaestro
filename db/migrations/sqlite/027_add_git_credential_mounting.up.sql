-- Add git_credential_mounting boolean column to workspaces table
-- Defaults to false (0) for existing rows

ALTER TABLE workspaces ADD COLUMN git_credential_mounting BOOLEAN NOT NULL DEFAULT 0;
