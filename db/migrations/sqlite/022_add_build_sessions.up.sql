-- 022_add_build_sessions.up.sql
-- Add build session persistence tables for tracking workspace build history.

CREATE TABLE IF NOT EXISTS build_sessions (
    id TEXT PRIMARY KEY,
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    status TEXT NOT NULL DEFAULT 'running',
    total_workspaces INTEGER NOT NULL DEFAULT 0,
    succeeded INTEGER DEFAULT 0,
    failed INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS build_session_workspaces (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    workspace_id INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'queued',
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    duration_seconds INTEGER,
    image_tag TEXT,
    error_message TEXT,
    FOREIGN KEY (session_id) REFERENCES build_sessions(id) ON DELETE CASCADE,
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_build_sessions_started ON build_sessions(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_build_session_workspaces_session ON build_session_workspaces(session_id);
