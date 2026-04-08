-- SQLite does not support DROP COLUMN before 3.35.0.
-- Recreate the table without the inherits column.
CREATE TABLE nvim_themes_backup AS SELECT
    id, name, description, author, category, plugin_repo, plugin_branch,
    plugin_tag, style, transparent, colors, options, custom_highlights,
    is_active, created_at, updated_at
FROM nvim_themes;

DROP TABLE nvim_themes;

CREATE TABLE nvim_themes (
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
    custom_highlights TEXT,
    is_active BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO nvim_themes (id, name, description, author, category, plugin_repo,
    plugin_branch, plugin_tag, style, transparent, colors, options,
    custom_highlights, is_active, created_at, updated_at)
SELECT id, name, description, author, category, plugin_repo,
    plugin_branch, plugin_tag, style, transparent, colors, options,
    custom_highlights, is_active, created_at, updated_at
FROM nvim_themes_backup;

DROP TABLE nvim_themes_backup;

CREATE INDEX IF NOT EXISTS idx_nvim_themes_category ON nvim_themes(category);
CREATE INDEX IF NOT EXISTS idx_nvim_themes_active ON nvim_themes(is_active) WHERE is_active = TRUE;
