-- Remove terminal_plugins table
DROP INDEX IF EXISTS idx_terminal_plugins_enabled;
DROP INDEX IF EXISTS idx_terminal_plugins_manager;
DROP INDEX IF EXISTS idx_terminal_plugins_shell;
DROP INDEX IF EXISTS idx_terminal_plugins_category;
DROP INDEX IF EXISTS idx_terminal_plugins_name;
DROP TABLE IF EXISTS terminal_plugins;