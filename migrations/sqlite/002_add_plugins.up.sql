-- Add plugins table to store reusable nvim plugin definitions
CREATE TABLE IF NOT EXISTS nvim_plugins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE, -- e.g., "telescope", "mason", "copilot"
    description TEXT,
    repo TEXT NOT NULL, -- e.g., "nvim-telescope/telescope.nvim"
    branch TEXT, -- e.g., "0.1.x"
    version TEXT, -- e.g., "v2.*"
    priority INTEGER, -- Plugin load priority (for colorschemes, etc.)
    lazy BOOLEAN DEFAULT TRUE, -- Whether plugin should be lazy-loaded
    
    -- Event triggers for lazy loading
    event TEXT, -- JSON array: ["BufReadPre", "BufNewFile"]
    ft TEXT, -- JSON array: ["python", "go"]
    keys TEXT, -- JSON array of key mappings
    cmd TEXT, -- JSON array: ["Telescope", "LazyGit"]
    
    -- Dependencies
    dependencies TEXT, -- JSON array of dependency specs
    
    -- Build commands
    build TEXT, -- e.g., "make", ":TSUpdate"
    
    -- Configuration
    config TEXT, -- Lua code for config function
    init TEXT, -- Lua code for init function
    opts TEXT, -- JSON object for opts
    
    -- Key mappings defined in plugin definition
    keymaps TEXT, -- JSON array of keymap definitions
    
    -- Metadata
    category TEXT, -- e.g., "colorscheme", "lsp", "git", "completion"
    tags TEXT, -- JSON array: ["fuzzy-finder", "telescope"]
    enabled BOOLEAN DEFAULT TRUE,
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Index for faster lookups
CREATE INDEX IF NOT EXISTS idx_nvim_plugins_name ON nvim_plugins(name);
CREATE INDEX IF NOT EXISTS idx_nvim_plugins_category ON nvim_plugins(category);
CREATE INDEX IF NOT EXISTS idx_nvim_plugins_enabled ON nvim_plugins(enabled);

-- Add workspace_plugins junction table to track which plugins each workspace uses
CREATE TABLE IF NOT EXISTS workspace_plugins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    workspace_id INTEGER NOT NULL,
    plugin_id INTEGER NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    custom_config TEXT, -- Optional: workspace-specific config overrides
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
    FOREIGN KEY (plugin_id) REFERENCES nvim_plugins(id) ON DELETE CASCADE,
    UNIQUE(workspace_id, plugin_id)
);

CREATE INDEX IF NOT EXISTS idx_workspace_plugins_workspace ON workspace_plugins(workspace_id);
CREATE INDEX IF NOT EXISTS idx_workspace_plugins_plugin ON workspace_plugins(plugin_id);
