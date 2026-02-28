-- 001_init.up.sql
-- Unified initial schema for DevOpsMaestro (dvm, nvp, dvt)
-- This creates all tables needed by the entire toolkit

-- =============================================================================
-- OBJECT HIERARCHY: Ecosystem -> Domain -> App -> Workspace
-- =============================================================================

CREATE TABLE IF NOT EXISTS ecosystems (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    theme TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ecosystems_name ON ecosystems(name);

CREATE TABLE IF NOT EXISTS domains (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ecosystem_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    theme TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (ecosystem_id) REFERENCES ecosystems(id) ON DELETE CASCADE,
    UNIQUE(ecosystem_id, name)
);

CREATE INDEX IF NOT EXISTS idx_domains_ecosystem ON domains(ecosystem_id);
CREATE INDEX IF NOT EXISTS idx_domains_name ON domains(name);

CREATE TABLE IF NOT EXISTS apps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    domain_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    path TEXT NOT NULL,
    description TEXT,
    language TEXT,
    build_config TEXT,
    theme TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (domain_id) REFERENCES domains(id) ON DELETE CASCADE,
    UNIQUE(domain_id, name)
);

CREATE INDEX IF NOT EXISTS idx_apps_domain ON apps(domain_id);
CREATE INDEX IF NOT EXISTS idx_apps_name ON apps(name);
CREATE INDEX IF NOT EXISTS idx_apps_path ON apps(path);

CREATE TABLE IF NOT EXISTS workspaces (
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
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(app_id, name)
);

CREATE INDEX IF NOT EXISTS idx_workspaces_slug ON workspaces(slug);
CREATE INDEX IF NOT EXISTS idx_workspaces_app ON workspaces(app_id);

-- =============================================================================
-- CONTEXT (Single row tracking active selections)
-- =============================================================================

CREATE TABLE IF NOT EXISTS context (
    id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    active_workspace_id INTEGER,
    active_ecosystem_id INTEGER REFERENCES ecosystems(id),
    active_domain_id INTEGER REFERENCES domains(id),
    active_app_id INTEGER REFERENCES apps(id),
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (active_workspace_id) REFERENCES workspaces(id)
);

INSERT OR IGNORE INTO context (id) VALUES (1);

-- =============================================================================
-- DEFAULTS (Global default settings)
-- =============================================================================

CREATE TABLE IF NOT EXISTS defaults (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_defaults_updated_at ON defaults(updated_at);

-- =============================================================================
-- CREDENTIALS (Hierarchical secret management)
-- =============================================================================

CREATE TABLE IF NOT EXISTS credentials (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scope_type TEXT NOT NULL CHECK(scope_type IN ('ecosystem', 'domain', 'app', 'workspace')),
    scope_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    source TEXT NOT NULL CHECK(source IN ('keychain', 'env')),
    service TEXT,
    env_var TEXT,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(scope_type, scope_id, name)
);

CREATE INDEX IF NOT EXISTS idx_credentials_scope ON credentials(scope_type, scope_id);
CREATE INDEX IF NOT EXISTS idx_credentials_name ON credentials(name);

-- =============================================================================
-- NVIM PLUGINS (NvimOps - nvp)
-- =============================================================================

CREATE TABLE IF NOT EXISTS nvim_plugins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    repo TEXT NOT NULL,
    branch TEXT,
    version TEXT,
    priority INTEGER,
    lazy BOOLEAN DEFAULT TRUE,
    event TEXT,
    ft TEXT,
    keys TEXT,
    cmd TEXT,
    dependencies TEXT,
    build TEXT,
    config TEXT,
    init TEXT,
    opts TEXT,
    keymaps TEXT,
    category TEXT,
    tags TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_nvim_plugins_name ON nvim_plugins(name);
CREATE INDEX IF NOT EXISTS idx_nvim_plugins_category ON nvim_plugins(category);
CREATE INDEX IF NOT EXISTS idx_nvim_plugins_enabled ON nvim_plugins(enabled);

CREATE TABLE IF NOT EXISTS workspace_plugins (
    workspace_id INTEGER NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    plugin_id INTEGER NOT NULL REFERENCES nvim_plugins(id) ON DELETE CASCADE,
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (workspace_id, plugin_id)
);

CREATE INDEX IF NOT EXISTS idx_workspace_plugins_workspace ON workspace_plugins(workspace_id);
CREATE INDEX IF NOT EXISTS idx_workspace_plugins_plugin ON workspace_plugins(plugin_id);

-- =============================================================================
-- NVIM THEMES (NvimOps - nvp)
-- =============================================================================

CREATE TABLE IF NOT EXISTS nvim_themes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    author TEXT,
    category TEXT,
    plugin_repo TEXT NOT NULL,
    plugin_branch TEXT,
    plugin_tag TEXT,
    style TEXT,
    transparent BOOLEAN DEFAULT FALSE,
    colors TEXT,
    options TEXT,
    is_active BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_nvim_themes_category ON nvim_themes(category);
CREATE INDEX IF NOT EXISTS idx_nvim_themes_active ON nvim_themes(is_active) WHERE is_active = TRUE;

CREATE TABLE IF NOT EXISTS workspace_themes (
    workspace_id INTEGER NOT NULL,
    theme_id INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (workspace_id, theme_id),
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
    FOREIGN KEY (theme_id) REFERENCES nvim_themes(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_workspace_themes_workspace ON workspace_themes(workspace_id);

-- =============================================================================
-- NVIM PACKAGES (NvimOps - nvp)
-- =============================================================================

CREATE TABLE IF NOT EXISTS nvim_packages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    category TEXT,
    labels TEXT,
    plugins TEXT NOT NULL,
    extends TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_nvim_packages_name ON nvim_packages(name);
CREATE INDEX IF NOT EXISTS idx_nvim_packages_category ON nvim_packages(category);
CREATE INDEX IF NOT EXISTS idx_nvim_packages_extends ON nvim_packages(extends);
CREATE INDEX IF NOT EXISTS idx_nvim_packages_created_at ON nvim_packages(created_at);
CREATE INDEX IF NOT EXISTS idx_nvim_packages_updated_at ON nvim_packages(updated_at);

-- =============================================================================
-- TERMINAL PACKAGES (TerminalOps - dvt)
-- =============================================================================

CREATE TABLE IF NOT EXISTS terminal_packages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    category TEXT,
    labels TEXT,
    plugins TEXT NOT NULL DEFAULT '[]',
    prompts TEXT NOT NULL DEFAULT '[]',
    profiles TEXT NOT NULL DEFAULT '[]',
    wezterm TEXT,
    extends TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_terminal_packages_name ON terminal_packages(name);
CREATE INDEX IF NOT EXISTS idx_terminal_packages_category ON terminal_packages(category);
CREATE INDEX IF NOT EXISTS idx_terminal_packages_extends ON terminal_packages(extends);

-- =============================================================================
-- TERMINAL PLUGINS (TerminalOps - dvt)
-- =============================================================================

CREATE TABLE IF NOT EXISTS terminal_plugins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    repo TEXT NOT NULL,
    category TEXT,
    shell TEXT NOT NULL DEFAULT 'zsh',
    manager TEXT NOT NULL DEFAULT 'manual',
    load_command TEXT,
    source_file TEXT,
    dependencies TEXT NOT NULL DEFAULT '[]',
    env_vars TEXT NOT NULL DEFAULT '{}',
    labels TEXT NOT NULL DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_terminal_plugins_name ON terminal_plugins(name);
CREATE INDEX IF NOT EXISTS idx_terminal_plugins_category ON terminal_plugins(category);
CREATE INDEX IF NOT EXISTS idx_terminal_plugins_shell ON terminal_plugins(shell);
CREATE INDEX IF NOT EXISTS idx_terminal_plugins_manager ON terminal_plugins(manager);
CREATE INDEX IF NOT EXISTS idx_terminal_plugins_enabled ON terminal_plugins(enabled);

-- =============================================================================
-- TERMINAL EMULATORS (TerminalOps - dvt)
-- =============================================================================

CREATE TABLE IF NOT EXISTS terminal_emulators (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    type TEXT NOT NULL,
    config TEXT NOT NULL DEFAULT '{}',
    theme_ref TEXT,
    category TEXT,
    labels TEXT NOT NULL DEFAULT '{}',
    workspace TEXT,
    enabled BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_terminal_emulators_name ON terminal_emulators(name);
CREATE INDEX IF NOT EXISTS idx_terminal_emulators_type ON terminal_emulators(type);
CREATE INDEX IF NOT EXISTS idx_terminal_emulators_category ON terminal_emulators(category);
CREATE INDEX IF NOT EXISTS idx_terminal_emulators_workspace ON terminal_emulators(workspace);
CREATE INDEX IF NOT EXISTS idx_terminal_emulators_enabled ON terminal_emulators(enabled);

-- =============================================================================
-- TERMINAL PROMPTS (TerminalOps - dvt)
-- =============================================================================

CREATE TABLE IF NOT EXISTS terminal_prompts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    type TEXT NOT NULL,
    add_newline BOOLEAN DEFAULT TRUE,
    palette TEXT,
    format TEXT,
    modules TEXT,
    character TEXT,
    palette_ref TEXT,
    colors TEXT,
    raw_config TEXT,
    category TEXT,
    tags TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_terminal_prompts_name ON terminal_prompts(name);
CREATE INDEX IF NOT EXISTS idx_terminal_prompts_type ON terminal_prompts(type);
CREATE INDEX IF NOT EXISTS idx_terminal_prompts_category ON terminal_prompts(category);
CREATE INDEX IF NOT EXISTS idx_terminal_prompts_enabled ON terminal_prompts(enabled);

-- =============================================================================
-- TERMINAL PROFILES (TerminalOps - dvt)
-- =============================================================================

CREATE TABLE IF NOT EXISTS terminal_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    category TEXT,
    prompt_ref TEXT,
    plugin_refs TEXT NOT NULL DEFAULT '[]',
    shell_ref TEXT,
    theme_ref TEXT,
    tags TEXT NOT NULL DEFAULT '[]',
    labels TEXT NOT NULL DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_terminal_profiles_name ON terminal_profiles(name);
CREATE INDEX IF NOT EXISTS idx_terminal_profiles_category ON terminal_profiles(category);
CREATE INDEX IF NOT EXISTS idx_terminal_profiles_enabled ON terminal_profiles(enabled);
