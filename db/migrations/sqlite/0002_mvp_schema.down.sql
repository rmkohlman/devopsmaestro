-- Drop context table
DROP TABLE IF EXISTS context;

-- Drop workspaces table
DROP TABLE IF EXISTS workspaces;

-- SQLite doesn't support DROP COLUMN, so we'd need to recreate the table to remove path
-- For MVP, we'll leave the path column (it's a non-breaking change)
