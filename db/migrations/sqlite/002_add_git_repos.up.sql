-- 002_add_git_repos.up.sql
-- Add git_repos table for v0.20.0 "Mirror" feature

CREATE TABLE IF NOT EXISTS git_repos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    url TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    default_ref TEXT DEFAULT 'main',
    auth_type TEXT DEFAULT 'none',
    credential_id INTEGER,
    auto_sync BOOLEAN DEFAULT true,
    sync_interval_minutes INTEGER DEFAULT 60,
    last_synced_at DATETIME,
    sync_status TEXT DEFAULT 'pending',
    sync_error TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (credential_id) REFERENCES credentials(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_git_repos_name ON git_repos(name);
CREATE INDEX IF NOT EXISTS idx_git_repos_slug ON git_repos(slug);
CREATE INDEX IF NOT EXISTS idx_git_repos_sync_status ON git_repos(sync_status);
CREATE INDEX IF NOT EXISTS idx_git_repos_auto_sync ON git_repos(auto_sync);
