-- Remove credentials table
DROP INDEX IF EXISTS idx_credentials_name;
DROP INDEX IF EXISTS idx_credentials_scope;
DROP TABLE IF EXISTS credentials;
