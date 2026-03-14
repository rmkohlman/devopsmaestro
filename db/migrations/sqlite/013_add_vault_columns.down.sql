-- Revert Migration 013: Restore keychain source and remove vault columns
UPDATE credentials SET source = 'keychain' WHERE source = 'vault';

-- SQLite doesn't support DROP COLUMN before 3.35.0, but we target modern SQLite
ALTER TABLE credentials DROP COLUMN vault_secret;
ALTER TABLE credentials DROP COLUMN vault_env;
ALTER TABLE credentials DROP COLUMN vault_username_secret;
