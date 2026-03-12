-- Migration 011: Add label and keychain_type columns for label-based keychain lookup
-- The label column stores the keychain entry display name (used with -l flag)
-- The keychain_type column stores 'generic' or 'internet' (maps to find-generic-password vs find-internet-password)

ALTER TABLE credentials ADD COLUMN label TEXT;
ALTER TABLE credentials ADD COLUMN keychain_type TEXT DEFAULT 'generic' CHECK(keychain_type IN ('generic', 'internet'));

-- Backfill: copy service values to label for existing keychain credentials
-- This preserves backward compatibility since for generic passwords created by dvm,
-- the service name was also used as the label (they default to the same value)
UPDATE credentials SET label = service WHERE source = 'keychain' AND service IS NOT NULL;
