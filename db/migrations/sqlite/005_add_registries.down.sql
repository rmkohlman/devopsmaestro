-- 005_add_registries.down.sql
-- Remove registries table

DROP INDEX IF EXISTS idx_registries_lifecycle;
DROP INDEX IF EXISTS idx_registries_status;
DROP INDEX IF EXISTS idx_registries_port;
DROP INDEX IF EXISTS idx_registries_type;
DROP INDEX IF EXISTS idx_registries_name;

DROP TABLE IF EXISTS registries;
