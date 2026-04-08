-- Add expires_at column to credentials table for rotation reminders (v0.58.0)
-- Nullable so existing credentials are unaffected.

ALTER TABLE credentials ADD COLUMN expires_at DATETIME;
