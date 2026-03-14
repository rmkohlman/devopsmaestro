-- Revert Migration 014: Restore keychain columns
-- This is a table rebuild to add back the dropped columns

CREATE TABLE credentials_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scope_type TEXT NOT NULL CHECK(scope_type IN ('ecosystem', 'domain', 'app', 'workspace')),
    scope_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    source TEXT NOT NULL CHECK(source IN ('keychain', 'env')),
    service TEXT,
    env_var TEXT,
    description TEXT,
    username_var TEXT,
    password_var TEXT,
    label TEXT,
    keychain_type TEXT DEFAULT 'internet' CHECK(keychain_type IN ('generic', 'internet')),
    vault_secret TEXT,
    vault_env TEXT,
    vault_username_secret TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(scope_type, scope_id, name)
);

INSERT INTO credentials_old (id, scope_type, scope_id, name, source, env_var, description, username_var, password_var, vault_secret, vault_env, vault_username_secret, created_at, updated_at)
SELECT id, scope_type, scope_id, name, 
    CASE WHEN source = 'vault' THEN 'keychain' ELSE source END,
    env_var, description, username_var, password_var, vault_secret, vault_env, vault_username_secret, created_at, updated_at
FROM credentials;

-- Backfill service/label from vault_secret
UPDATE credentials_old SET service = vault_secret, label = vault_secret WHERE source = 'keychain';

DROP TABLE credentials;
ALTER TABLE credentials_old RENAME TO credentials;

CREATE INDEX IF NOT EXISTS idx_credentials_scope ON credentials(scope_type, scope_id);
CREATE INDEX IF NOT EXISTS idx_credentials_name ON credentials(name);
