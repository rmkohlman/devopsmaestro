-- Reverse migration: remove expires_at column from credentials

ALTER TABLE credentials DROP COLUMN expires_at;
