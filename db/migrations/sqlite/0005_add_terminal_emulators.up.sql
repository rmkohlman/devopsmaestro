-- Add terminal_emulators table
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

-- Performance indexes
CREATE INDEX IF NOT EXISTS idx_terminal_emulators_name ON terminal_emulators(name);
CREATE INDEX IF NOT EXISTS idx_terminal_emulators_type ON terminal_emulators(type);
CREATE INDEX IF NOT EXISTS idx_terminal_emulators_category ON terminal_emulators(category);
CREATE INDEX IF NOT EXISTS idx_terminal_emulators_workspace ON terminal_emulators(workspace);
CREATE INDEX IF NOT EXISTS idx_terminal_emulators_enabled ON terminal_emulators(enabled);