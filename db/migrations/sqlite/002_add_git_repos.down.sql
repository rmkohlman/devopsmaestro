-- 002_add_git_repos.down.sql
-- Rollback git_repos table

DROP INDEX IF EXISTS idx_git_repos_auto_sync;
DROP INDEX IF EXISTS idx_git_repos_sync_status;
DROP INDEX IF EXISTS idx_git_repos_slug;
DROP INDEX IF EXISTS idx_git_repos_name;

DROP TABLE IF EXISTS git_repos;
