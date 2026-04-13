-- Migration 025: Add 'system' to credentials scope_type CHECK constraint
-- SQLite does not support ALTER TABLE to modify constraints, so we must
-- rebuild the table using the standard create-copy-drop-rename pattern.
--
-- New constraint: scope_type IN ('ecosystem','domain','system','app','workspace')

-- Step 1: Create new table with updated CHECK constraint
CREATE TABLE credentials_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scope_type TEXT NOT NULL CHECK(scope_type IN ('ecosystem', 'domain', 'system', 'app', 'workspace')),
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
    vault_fields TEXT,
    expires_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(scope_type, scope_id, name)
);

-- Step 2: Copy all existing data
INSERT INTO credentials_new (id, scope_type, scope_id, name, source, env_var, description, username_var, password_var, vault_secret, vault_env, vault_username_secret, vault_fields, expires_at, created_at, updated_at)
SELECT id, scope_type, scope_id, name, source, env_var, description, username_var, password_var, vault_secret, vault_env, vault_username_secret, vault_fields, expires_at, created_at, updated_at
FROM credentials;

-- Step 3: Drop old table
DROP TABLE credentials;

-- Step 4: Rename new table
ALTER TABLE credentials_new RENAME TO credentials;

-- Step 5: Recreate indexes
CREATE INDEX IF NOT EXISTS idx_credentials_scope ON credentials(scope_type, scope_id);
CREATE INDEX IF NOT EXISTS idx_credentials_name ON credentials(name);
