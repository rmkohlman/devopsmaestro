-- 021_add_package_hierarchy.down.sql
-- Remove nvim_package and terminal_package columns from entity tables.
-- SQLite < 3.35.0 does not support DROP COLUMN; recreate tables.

-- =========================================================================
-- ecosystems: drop nvim_package, terminal_package
-- =========================================================================
CREATE TABLE ecosystems_backup AS SELECT
    id, name, description, theme, build_args, ca_certs, created_at, updated_at
FROM ecosystems;

DROP TABLE ecosystems;

CREATE TABLE ecosystems (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    theme TEXT,
    build_args TEXT,
    ca_certs TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO ecosystems (id, name, description, theme, build_args, ca_certs, created_at, updated_at)
SELECT id, name, description, theme, build_args, ca_certs, created_at, updated_at
FROM ecosystems_backup;

DROP TABLE ecosystems_backup;

CREATE INDEX IF NOT EXISTS idx_ecosystems_name ON ecosystems(name);

-- =========================================================================
-- domains: drop nvim_package, terminal_package
-- =========================================================================
CREATE TABLE domains_backup AS SELECT
    id, ecosystem_id, name, description, theme, build_args, ca_certs, created_at, updated_at
FROM domains;

DROP TABLE domains;

CREATE TABLE domains (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ecosystem_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    theme TEXT,
    build_args TEXT,
    ca_certs TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (ecosystem_id) REFERENCES ecosystems(id) ON DELETE CASCADE,
    UNIQUE(ecosystem_id, name)
);

INSERT INTO domains (id, ecosystem_id, name, description, theme, build_args, ca_certs, created_at, updated_at)
SELECT id, ecosystem_id, name, description, theme, build_args, ca_certs, created_at, updated_at
FROM domains_backup;

DROP TABLE domains_backup;

CREATE INDEX IF NOT EXISTS idx_domains_ecosystem ON domains(ecosystem_id);
CREATE INDEX IF NOT EXISTS idx_domains_name ON domains(name);

-- =========================================================================
-- apps: drop nvim_package, terminal_package
-- =========================================================================
CREATE TABLE apps_backup AS SELECT
    id, domain_id, name, path, description, theme, language, build_config, git_repo_id, created_at, updated_at
FROM apps;

DROP TABLE apps;

CREATE TABLE apps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    domain_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    path TEXT NOT NULL,
    description TEXT,
    theme TEXT,
    language TEXT,
    build_config TEXT,
    git_repo_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (domain_id) REFERENCES domains(id) ON DELETE CASCADE,
    UNIQUE(domain_id, name)
);

INSERT INTO apps (id, domain_id, name, path, description, theme, language, build_config, git_repo_id, created_at, updated_at)
SELECT id, domain_id, name, path, description, theme, language, build_config, git_repo_id, created_at, updated_at
FROM apps_backup;

DROP TABLE apps_backup;

CREATE INDEX IF NOT EXISTS idx_apps_domain ON apps(domain_id);
CREATE INDEX IF NOT EXISTS idx_apps_name ON apps(name);
CREATE INDEX IF NOT EXISTS idx_apps_path ON apps(path);

-- =========================================================================
-- workspaces: drop nvim_package only (terminal_package stays — it's from migration 004)
-- =========================================================================
CREATE TABLE workspaces_backup AS SELECT
    id, app_id, name, slug, description, image_name, container_id, status,
    ssh_agent_forwarding, nvim_structure, nvim_plugins, theme,
    terminal_prompt, terminal_plugins, terminal_package,
    git_repo_id, env, build_config, created_at, updated_at
FROM workspaces;

DROP TABLE workspaces;

CREATE TABLE workspaces (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id INTEGER NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    description TEXT,
    image_name TEXT NOT NULL,
    container_id TEXT,
    status TEXT NOT NULL DEFAULT 'stopped',
    ssh_agent_forwarding INTEGER DEFAULT 0,
    nvim_structure TEXT,
    nvim_plugins TEXT,
    theme TEXT,
    terminal_prompt TEXT,
    terminal_plugins TEXT,
    terminal_package TEXT,
    git_repo_id INTEGER,
    env TEXT NOT NULL DEFAULT '{}',
    build_config TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(app_id, name)
);

INSERT INTO workspaces (id, app_id, name, slug, description, image_name, container_id, status,
    ssh_agent_forwarding, nvim_structure, nvim_plugins, theme,
    terminal_prompt, terminal_plugins, terminal_package,
    git_repo_id, env, build_config, created_at, updated_at)
SELECT id, app_id, name, slug, description, image_name, container_id, status,
    ssh_agent_forwarding, nvim_structure, nvim_plugins, theme,
    terminal_prompt, terminal_plugins, terminal_package,
    git_repo_id, env, build_config, created_at, updated_at
FROM workspaces_backup;

DROP TABLE workspaces_backup;

CREATE INDEX IF NOT EXISTS idx_workspaces_slug ON workspaces(slug);
CREATE INDEX IF NOT EXISTS idx_workspaces_app ON workspaces(app_id);
