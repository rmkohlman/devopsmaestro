-- 014_add_nvim_packages.down.sql
-- Removes nvim_packages table

DROP INDEX IF EXISTS idx_nvim_packages_updated_at;
DROP INDEX IF EXISTS idx_nvim_packages_created_at;
DROP INDEX IF EXISTS idx_nvim_packages_extends;
DROP INDEX IF EXISTS idx_nvim_packages_category;
DROP INDEX IF EXISTS idx_nvim_packages_name;

DROP TABLE IF EXISTS nvim_packages;