-- SQLite doesn't support DROP COLUMN in older versions; rebuild table
CREATE TABLE credentials_backup AS SELECT
    id, scope_type, scope_id, name, source, vault_secret, vault_env,
    vault_username_secret, env_var, description, username_var, password_var,
    created_at, updated_at
FROM credentials;

DROP TABLE credentials;

CREATE TABLE credentials (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scope_type TEXT NOT NULL CHECK(scope_type IN ('ecosystem', 'domain', 'app', 'workspace')),
    scope_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    source TEXT NOT NULL CHECK(source IN ('vault', 'env')),
    vault_secret TEXT,
    vault_env TEXT,
    vault_username_secret TEXT,
    env_var TEXT,
    description TEXT,
    username_var TEXT,
    password_var TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(scope_type, scope_id, name)
);

INSERT INTO credentials SELECT * FROM credentials_backup;
DROP TABLE credentials_backup;
