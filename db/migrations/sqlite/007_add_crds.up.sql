-- 007_add_crds.up.sql
-- Add CRD (Custom Resource Definition) tables for extensibility

-- Table 1: custom_resource_definitions
-- Stores the definitions of custom resource types
CREATE TABLE IF NOT EXISTS custom_resource_definitions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    kind TEXT NOT NULL UNIQUE,
    "group" TEXT NOT NULL,
    singular TEXT NOT NULL,
    plural TEXT NOT NULL,
    short_names TEXT,
    scope TEXT NOT NULL CHECK(scope IN ('Global', 'Workspace', 'App', 'Domain', 'Ecosystem')),
    versions TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for custom_resource_definitions
CREATE INDEX IF NOT EXISTS idx_crd_kind ON custom_resource_definitions(kind);
CREATE INDEX IF NOT EXISTS idx_crd_group ON custom_resource_definitions("group");
CREATE INDEX IF NOT EXISTS idx_crd_scope ON custom_resource_definitions(scope);

-- Table 2: custom_resources
-- Stores instances of custom resources
CREATE TABLE IF NOT EXISTS custom_resources (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    kind TEXT NOT NULL,
    name TEXT NOT NULL,
    namespace TEXT,
    spec TEXT,
    status TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(kind, name, namespace)
);

-- Indexes for custom_resources
CREATE INDEX IF NOT EXISTS idx_cr_kind ON custom_resources(kind);
CREATE INDEX IF NOT EXISTS idx_cr_name ON custom_resources(name);
CREATE INDEX IF NOT EXISTS idx_cr_namespace ON custom_resources(namespace);
CREATE INDEX IF NOT EXISTS idx_cr_kind_name ON custom_resources(kind, name);
