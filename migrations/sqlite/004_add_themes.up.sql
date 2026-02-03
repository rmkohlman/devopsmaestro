-- 004_add_themes.up.sql
-- Adds nvim_themes table for storing Neovim colorscheme configurations

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
    colors TEXT,        -- JSON object of color definitions
    options TEXT,       -- JSON object of theme options
    is_active BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Index for category queries
CREATE INDEX IF NOT EXISTS idx_nvim_themes_category ON nvim_themes(category);

-- Index for active theme lookup
CREATE INDEX IF NOT EXISTS idx_nvim_themes_active ON nvim_themes(is_active) WHERE is_active = TRUE;

-- Workspace theme associations (for per-workspace theme support)
CREATE TABLE IF NOT EXISTS workspace_themes (
    workspace_id INTEGER NOT NULL,
    theme_id INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (workspace_id, theme_id),
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
    FOREIGN KEY (theme_id) REFERENCES nvim_themes(id) ON DELETE CASCADE
);

-- Index for workspace theme lookups
CREATE INDEX IF NOT EXISTS idx_workspace_themes_workspace ON workspace_themes(workspace_id);
