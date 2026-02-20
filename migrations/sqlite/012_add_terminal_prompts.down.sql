-- 012_add_terminal_prompts.down.sql
-- Removes terminal_prompts table

DROP INDEX IF EXISTS idx_terminal_prompts_enabled;
DROP INDEX IF EXISTS idx_terminal_prompts_category;
DROP INDEX IF EXISTS idx_terminal_prompts_type;
DROP INDEX IF EXISTS idx_terminal_prompts_name;

DROP TABLE IF EXISTS terminal_prompts;