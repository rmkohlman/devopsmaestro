-- 007_add_apps.down.sql
-- Removes apps table

DROP INDEX IF EXISTS idx_apps_path;
DROP INDEX IF EXISTS idx_apps_name;
DROP INDEX IF EXISTS idx_apps_domain;
DROP TABLE IF EXISTS apps;
