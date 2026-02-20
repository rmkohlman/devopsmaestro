-- 013_add_defaults.down.sql
-- Removes defaults table

DROP INDEX IF EXISTS idx_defaults_updated_at;

DROP TABLE IF EXISTS defaults;