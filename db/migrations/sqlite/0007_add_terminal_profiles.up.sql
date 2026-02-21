-- 0007_add_terminal_profiles.up.sql
-- Adds terminal_profiles table for storing complete terminal configuration profiles

CREATE TABLE IF NOT EXISTS terminal_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    category TEXT,
    prompt_ref TEXT,          -- Reference to prompt by name
    plugin_refs TEXT NOT NULL DEFAULT '[]',  -- JSON array of plugin names
    shell_ref TEXT,           -- Reference to shell config
    theme_ref TEXT,           -- Reference to theme
    tags TEXT NOT NULL DEFAULT '[]',  -- JSON array
    labels TEXT NOT NULL DEFAULT '{}', -- JSON object
    enabled BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_terminal_profiles_name ON terminal_profiles(name);
CREATE INDEX IF NOT EXISTS idx_terminal_profiles_category ON terminal_profiles(category);
CREATE INDEX IF NOT EXISTS idx_terminal_profiles_enabled ON terminal_profiles(enabled);
