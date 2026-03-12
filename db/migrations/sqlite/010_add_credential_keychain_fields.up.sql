-- Migration 010: Add dual-field keychain support
-- Adds username_var and password_var columns for credentials that extract
-- both account and password from a single keychain entry.
ALTER TABLE credentials ADD COLUMN username_var TEXT;
ALTER TABLE credentials ADD COLUMN password_var TEXT;
