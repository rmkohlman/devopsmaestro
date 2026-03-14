-- Migration 013: Add MaestroVault columns and migrate from keychain to vault
-- Part 1 of the keychain→vault migration (v0.40.0)

-- Add new vault columns
ALTER TABLE credentials ADD COLUMN vault_secret TEXT;
ALTER TABLE credentials ADD COLUMN vault_env TEXT;
ALTER TABLE credentials ADD COLUMN vault_username_secret TEXT;

-- Backfill vault_secret from existing keychain credentials:
-- Use label if available, fall back to service
UPDATE credentials 
SET vault_secret = COALESCE(label, service) 
WHERE source = 'keychain';

-- Change source from 'keychain' to 'vault' for all keychain credentials
UPDATE credentials SET source = 'vault' WHERE source = 'keychain';
