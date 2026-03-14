-- Migration 014: Drop keychain columns and update CHECK constraint (v0.40.0)
-- SQLite table rebuild to remove deprecated keychain columns

-- Step 1: Create new table without keychain columns
CREATE TABLE credentials_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scope_type TEXT NOT NULL CHECK(scope_type IN ('ecosystem', 'domain', 'app', 'workspace')),
    scope_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    source TEXT NOT NULL CHECK(source IN ('vault', 'env')),
    env_var TEXT,
    description TEXT,
    username_var TEXT,
    password_var TEXT,
    vault_secret TEXT,
    vault_env TEXT,
    vault_username_secret TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(scope_type, scope_id, name)
);

-- Step 2: Copy data (excluding dropped columns)
INSERT INTO credentials_new (id, scope_type, scope_id, name, source, env_var, description, username_var, password_var, vault_secret, vault_env, vault_username_secret, created_at, updated_at)
SELECT id, scope_type, scope_id, name, source, env_var, description, username_var, password_var, vault_secret, vault_env, vault_username_secret, created_at, updated_at
FROM credentials;

-- Step 3: Drop old table
DROP TABLE credentials;

-- Step 4: Rename new table
ALTER TABLE credentials_new RENAME TO credentials;

-- Step 5: Recreate indexes
CREATE INDEX IF NOT EXISTS idx_credentials_scope ON credentials(scope_type, scope_id);
CREATE INDEX IF NOT EXISTS idx_credentials_name ON credentials(name);
