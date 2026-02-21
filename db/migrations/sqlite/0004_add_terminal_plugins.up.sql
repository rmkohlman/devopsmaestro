-- Add terminal_plugins table
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
    dependencies TEXT NOT NULL DEFAULT '[]',  -- Always valid JSON
    env_vars TEXT NOT NULL DEFAULT '{}',      -- Always valid JSON
    labels TEXT NOT NULL DEFAULT '{}',        -- Always valid JSON  
    enabled BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Performance indexes
CREATE INDEX IF NOT EXISTS idx_terminal_plugins_name ON terminal_plugins(name);
CREATE INDEX IF NOT EXISTS idx_terminal_plugins_category ON terminal_plugins(category);
CREATE INDEX IF NOT EXISTS idx_terminal_plugins_shell ON terminal_plugins(shell);
CREATE INDEX IF NOT EXISTS idx_terminal_plugins_manager ON terminal_plugins(manager);
CREATE INDEX IF NOT EXISTS idx_terminal_plugins_enabled ON terminal_plugins(enabled);