-- 006_add_domains.down.sql
-- Removes domains table

DROP INDEX IF EXISTS idx_domains_name;
DROP INDEX IF EXISTS idx_domains_ecosystem;
DROP TABLE IF EXISTS domains;
