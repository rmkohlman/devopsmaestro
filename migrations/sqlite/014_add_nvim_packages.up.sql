-- 014_add_nvim_packages.up.sql
-- Adds nvim_packages table for storing NvimPackage collections

CREATE TABLE IF NOT EXISTS nvim_packages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    category TEXT,
    labels TEXT, -- JSON object for labels (e.g., {"source": "lazyvim", "version": "15.0"})
    plugins TEXT NOT NULL, -- JSON array of plugin names (e.g., ["telescope", "treesitter"])
    extends TEXT, -- optional parent package name for inheritance
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_nvim_packages_name ON nvim_packages(name);
CREATE INDEX IF NOT EXISTS idx_nvim_packages_category ON nvim_packages(category);
CREATE INDEX IF NOT EXISTS idx_nvim_packages_extends ON nvim_packages(extends);
CREATE INDEX IF NOT EXISTS idx_nvim_packages_created_at ON nvim_packages(created_at);
CREATE INDEX IF NOT EXISTS idx_nvim_packages_updated_at ON nvim_packages(updated_at);