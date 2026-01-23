-- Remove nvim configuration columns from workspaces table
-- Note: SQLite doesn't support DROP COLUMN directly
-- This would require recreating the table in a real scenario
-- For now, we'll leave the columns (they won't hurt)
SELECT 'Migration down not fully supported in SQLite' as message;
