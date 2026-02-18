-- Add credentials table for storing credential configurations
-- Credentials can be associated with ecosystems, domains, apps, or workspaces
-- This enables hierarchical credential inheritance

CREATE TABLE IF NOT EXISTS credentials (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    
    -- Scope: which level this credential belongs to
    scope_type TEXT NOT NULL CHECK(scope_type IN ('ecosystem', 'domain', 'app', 'workspace')),
    scope_id INTEGER NOT NULL,
    
    -- Credential definition
    name TEXT NOT NULL,           -- e.g., GITHUB_PAT, NPM_TOKEN, AWS_ACCESS_KEY_ID
    source TEXT NOT NULL CHECK(source IN ('keychain', 'env', 'value')),
    service TEXT,                 -- Keychain service name (when source='keychain')
    env_var TEXT,                 -- Environment variable name (when source='env')
    value TEXT,                   -- Plaintext value (when source='value', not recommended)
    
    -- Metadata
    description TEXT,             -- Optional description
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- Each credential name must be unique within a scope
    UNIQUE(scope_type, scope_id, name)
);

-- Index for fast lookups by scope
CREATE INDEX IF NOT EXISTS idx_credentials_scope ON credentials(scope_type, scope_id);

-- Index for credential name lookups
CREATE INDEX IF NOT EXISTS idx_credentials_name ON credentials(name);
