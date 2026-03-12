-- Note: SQLite does not support DROP COLUMN in older versions.
-- For rollback, we recreate the table without the new columns.
CREATE TABLE credentials_backup AS SELECT
    id, scope_type, scope_id, name, source, service, env_var, description, username_var, password_var, created_at, updated_at
FROM credentials;

DROP TABLE credentials;

CREATE TABLE credentials (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    scope_type  TEXT NOT NULL CHECK(scope_type IN ('ecosystem','domain','app','workspace')),
    scope_id    INTEGER NOT NULL,
    name        TEXT NOT NULL,
    source      TEXT NOT NULL CHECK(source IN ('keychain','env')),
    service     TEXT,
    env_var     TEXT,
    description TEXT,
    username_var TEXT,
    password_var TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(scope_type, scope_id, name)
);

INSERT INTO credentials SELECT * FROM credentials_backup;
DROP TABLE credentials_backup;
