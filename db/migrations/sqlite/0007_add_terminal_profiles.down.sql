-- 0007_add_terminal_profiles.down.sql
-- Removes terminal_profiles table

DROP INDEX IF EXISTS idx_terminal_profiles_enabled;
DROP INDEX IF EXISTS idx_terminal_profiles_category;
DROP INDEX IF EXISTS idx_terminal_profiles_name;

DROP TABLE IF EXISTS terminal_profiles;
