-- 022_add_build_sessions.down.sql
-- Remove build session persistence tables.

DROP INDEX IF EXISTS idx_build_session_workspaces_session;
DROP INDEX IF EXISTS idx_build_sessions_started;
DROP TABLE IF EXISTS build_session_workspaces;
DROP TABLE IF EXISTS build_sessions;
