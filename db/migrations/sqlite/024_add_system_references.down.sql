-- 024_add_system_references.down.sql
-- Remove system_id from apps and active_system_id from context.
-- SQLite cannot DROP COLUMN, so we rebuild both tables.

-- =========================================================================
-- Drop the index first
-- =========================================================================
DROP INDEX IF EXISTS idx_apps_system_id;

-- =========================================================================
-- apps: remove system_id column via table rebuild
-- =========================================================================
CREATE TABLE apps_backup AS SELECT
    id, domain_id, name, path, description, language, build_config,
    theme, git_repo_id, nvim_package, terminal_package,
    created_at, updated_at
FROM apps;

DROP TABLE apps;

CREATE TABLE apps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    domain_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    path TEXT NOT NULL,
    description TEXT,
    language TEXT,
    build_config TEXT,
    theme TEXT,
    git_repo_id INTEGER REFERENCES git_repos(id) ON DELETE SET NULL,
    nvim_package TEXT,
    terminal_package TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (domain_id) REFERENCES domains(id) ON DELETE CASCADE,
    UNIQUE(domain_id, name)
);

INSERT INTO apps (id, domain_id, name, path, description, language,
    build_config, theme, git_repo_id, nvim_package, terminal_package,
    created_at, updated_at)
SELECT id, domain_id, name, path, description, language,
    build_config, theme, git_repo_id, nvim_package, terminal_package,
    created_at, updated_at
FROM apps_backup;

DROP TABLE apps_backup;

CREATE INDEX IF NOT EXISTS idx_apps_domain ON apps(domain_id);
CREATE INDEX IF NOT EXISTS idx_apps_name ON apps(name);
CREATE INDEX IF NOT EXISTS idx_apps_path ON apps(path);

-- =========================================================================
-- context: remove active_system_id column via table rebuild
-- =========================================================================
CREATE TABLE context_backup AS SELECT
    id, active_workspace_id, active_ecosystem_id,
    active_domain_id, active_app_id, updated_at
FROM context;

DROP TABLE context;

CREATE TABLE context (
    id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    active_workspace_id INTEGER,
    active_ecosystem_id INTEGER REFERENCES ecosystems(id) ON DELETE SET NULL,
    active_domain_id INTEGER REFERENCES domains(id) ON DELETE SET NULL,
    active_app_id INTEGER REFERENCES apps(id) ON DELETE SET NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (active_workspace_id) REFERENCES workspaces(id) ON DELETE SET NULL
);

INSERT INTO context (id, active_workspace_id, active_ecosystem_id,
    active_domain_id, active_app_id, updated_at)
SELECT id, active_workspace_id, active_ecosystem_id,
    active_domain_id, active_app_id, updated_at
FROM context_backup;

DROP TABLE context_backup;
