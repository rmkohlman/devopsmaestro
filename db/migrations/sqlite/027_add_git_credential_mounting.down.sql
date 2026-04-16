-- Remove git_credential_mounting column from workspaces table

ALTER TABLE workspaces DROP COLUMN git_credential_mounting;
