-- Down migration 025: Remove 'system' from credentials scope_type CHECK constraint
-- Rebuild the table without 'system' in the CHECK constraint.
-- Any credentials with scope_type='system' will be deleted before rebuild.

-- Step 1: Delete any system-scoped credentials (they can't exist in old schema)
DELETE FROM credentials WHERE scope_type = 'system';

-- Step 2: Create table with original CHECK constraint
CREATE TABLE credentials_old (
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
    vault_fields TEXT,
    expires_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(scope_type, scope_id, name)
);

-- Step 3: Copy remaining data
INSERT INTO credentials_old (id, scope_type, scope_id, name, source, env_var, description, username_var, password_var, vault_secret, vault_env, vault_username_secret, vault_fields, expires_at, created_at, updated_at)
SELECT id, scope_type, scope_id, name, source, env_var, description, username_var, password_var, vault_secret, vault_env, vault_username_secret, vault_fields, expires_at, created_at, updated_at
FROM credentials;

-- Step 4: Drop table with system scope
DROP TABLE credentials;

-- Step 5: Rename back
ALTER TABLE credentials_old RENAME TO credentials;

-- Step 6: Recreate indexes
CREATE INDEX IF NOT EXISTS idx_credentials_scope ON credentials(scope_type, scope_id);
CREATE INDEX IF NOT EXISTS idx_credentials_name ON credentials(name);
