-- 023_create_systems_table.down.sql
-- Drop the systems table and its indexes.

DROP INDEX IF EXISTS idx_systems_ecosystem_id;
DROP INDEX IF EXISTS idx_systems_domain_id;
DROP TABLE IF EXISTS systems;
