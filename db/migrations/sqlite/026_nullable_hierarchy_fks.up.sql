-- 026_nullable_hierarchy_fks.up.sql
-- Make domains.ecosystem_id and apps.domain_id nullable, change ON DELETE CASCADE
-- to ON DELETE SET NULL. Also change apps.system_id from CASCADE to SET NULL.
-- This allows intermediate hierarchy levels (ecosystem, domain) to be optional.

-- =========================================================================
-- REBUILD domains: ecosystem_id nullable, ON DELETE SET NULL
-- =========================================================================

CREATE TABLE domains_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ecosystem_id INTEGER,
    name TEXT NOT NULL,
    description TEXT,
    theme TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    build_args TEXT,
    ca_certs TEXT,
    nvim_package TEXT,
    terminal_package TEXT,
    FOREIGN KEY (ecosystem_id) REFERENCES ecosystems(id) ON DELETE SET NULL,
    UNIQUE(ecosystem_id, name)
);

INSERT INTO domains_new (id, ecosystem_id, name, description, theme,
    created_at, updated_at, build_args, ca_certs, nvim_package, terminal_package)
SELECT id, ecosystem_id, name, description, theme,
    created_at, updated_at, build_args, ca_certs, nvim_package, terminal_package
FROM domains;

DROP TABLE domains;
ALTER TABLE domains_new RENAME TO domains;

CREATE INDEX IF NOT EXISTS idx_domains_ecosystem ON domains(ecosystem_id);
CREATE INDEX IF NOT EXISTS idx_domains_name ON domains(name);

-- =========================================================================
-- REBUILD apps: domain_id nullable + SET NULL, system_id SET NULL
-- =========================================================================

CREATE TABLE apps_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    domain_id INTEGER,
    name TEXT NOT NULL,
    path TEXT NOT NULL,
    description TEXT,
    language TEXT,
    build_config TEXT,
    theme TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    git_repo_id INTEGER REFERENCES git_repos(id) ON DELETE SET NULL,
    nvim_package TEXT,
    terminal_package TEXT,
    system_id INTEGER REFERENCES systems(id) ON DELETE SET NULL,
    FOREIGN KEY (domain_id) REFERENCES domains(id) ON DELETE SET NULL,
    UNIQUE(domain_id, name)
);

INSERT INTO apps_new (id, domain_id, name, path, description, language,
    build_config, theme, created_at, updated_at, git_repo_id,
    nvim_package, terminal_package, system_id)
SELECT id, domain_id, name, path, description, language,
    build_config, theme, created_at, updated_at, git_repo_id,
    nvim_package, terminal_package, system_id
FROM apps;

DROP TABLE apps;
ALTER TABLE apps_new RENAME TO apps;

CREATE INDEX IF NOT EXISTS idx_apps_domain ON apps(domain_id);
CREATE INDEX IF NOT EXISTS idx_apps_name ON apps(name);
CREATE INDEX IF NOT EXISTS idx_apps_path ON apps(path);
CREATE INDEX IF NOT EXISTS idx_apps_git_repo ON apps(git_repo_id);
CREATE INDEX IF NOT EXISTS idx_apps_system_id ON apps(system_id);


