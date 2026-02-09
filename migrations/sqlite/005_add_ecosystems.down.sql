-- 005_add_ecosystems.down.sql
-- Removes ecosystems table

DROP INDEX IF EXISTS idx_ecosystems_name;
DROP TABLE IF EXISTS ecosystems;
