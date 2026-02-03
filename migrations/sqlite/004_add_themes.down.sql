-- 004_add_themes.down.sql
-- Removes nvim_themes table and related objects

DROP INDEX IF EXISTS idx_workspace_themes_workspace;
DROP TABLE IF EXISTS workspace_themes;
DROP INDEX IF EXISTS idx_nvim_themes_active;
DROP INDEX IF EXISTS idx_nvim_themes_category;
DROP TABLE IF EXISTS nvim_themes;
