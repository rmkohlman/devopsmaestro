-- Add terminal_packages table
CREATE TABLE IF NOT EXISTS terminal_packages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    category TEXT,
    labels TEXT, -- JSON object
    plugins TEXT NOT NULL DEFAULT '[]', -- JSON array
    prompts TEXT NOT NULL DEFAULT '[]', -- JSON array
    profiles TEXT NOT NULL DEFAULT '[]', -- JSON array
    wezterm TEXT, -- JSON object for WezTerm config
    extends TEXT, -- optional parent package
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_terminal_packages_name ON terminal_packages(name);
CREATE INDEX IF NOT EXISTS idx_terminal_packages_category ON terminal_packages(category);
CREATE INDEX IF NOT EXISTS idx_terminal_packages_extends ON terminal_packages(extends);