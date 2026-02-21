-- 0006_add_terminal_prompts.up.sql
-- Adds terminal_prompts table for storing terminal prompt configurations (Starship, Powerlevel10k, Oh My Posh)

CREATE TABLE IF NOT EXISTS terminal_prompts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    type TEXT NOT NULL, -- 'starship', 'powerlevel10k', 'oh-my-posh'
    add_newline BOOLEAN DEFAULT TRUE,
    palette TEXT,
    format TEXT,
    modules TEXT,       -- JSON object: map[string]ModuleConfig
    character TEXT,     -- JSON object: CharacterConfig
    palette_ref TEXT,
    colors TEXT,        -- JSON object: map[string]string
    raw_config TEXT,    -- Raw config for advanced users
    category TEXT,
    tags TEXT,          -- JSON array: []string
    enabled BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_terminal_prompts_name ON terminal_prompts(name);
CREATE INDEX IF NOT EXISTS idx_terminal_prompts_type ON terminal_prompts(type);
CREATE INDEX IF NOT EXISTS idx_terminal_prompts_category ON terminal_prompts(category);
CREATE INDEX IF NOT EXISTS idx_terminal_prompts_enabled ON terminal_prompts(enabled);
