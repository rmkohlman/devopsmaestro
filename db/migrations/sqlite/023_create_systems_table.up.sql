-- 023_create_systems_table.up.sql
-- Create systems table for the System entity in the hierarchy:
-- Ecosystem -> Domain -> System -> App -> Workspace

CREATE TABLE IF NOT EXISTS systems (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ecosystem_id INTEGER,
    domain_id INTEGER,
    name TEXT NOT NULL,
    description TEXT,
    theme TEXT,
    nvim_package TEXT,
    terminal_package TEXT,
    build_args TEXT,
    ca_certs TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (ecosystem_id) REFERENCES ecosystems(id) ON DELETE SET NULL,
    FOREIGN KEY (domain_id) REFERENCES domains(id) ON DELETE CASCADE,
    UNIQUE(domain_id, name)
);

CREATE INDEX IF NOT EXISTS idx_systems_domain_id ON systems(domain_id);
CREATE INDEX IF NOT EXISTS idx_systems_ecosystem_id ON systems(ecosystem_id);
