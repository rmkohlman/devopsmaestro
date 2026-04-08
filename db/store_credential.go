package db

import (
	"database/sql"
	"errors"
	"fmt"

	"devopsmaestro/models"
)

// =============================================================================
// Credential Operations
// =============================================================================

// credentialColumns is the canonical SELECT column list for credentials.
// Every SELECT and every Scan must use this exact order.
const credentialColumns = `id, scope_type, scope_id, name, source, env_var, description, username_var, password_var, vault_secret, vault_env, vault_username_secret, vault_fields, expires_at, created_at, updated_at`

// scanCredential scans a database row into a CredentialDB model.
// The scanner interface matches both *sql.Row and *sql.Rows.
func scanCredential(scanner interface{ Scan(...any) error }) (*models.CredentialDB, error) {
	c := &models.CredentialDB{}
	err := scanner.Scan(
		&c.ID, &c.ScopeType, &c.ScopeID, &c.Name, &c.Source,
		&c.EnvVar, &c.Description, &c.UsernameVar, &c.PasswordVar,
		&c.VaultSecret, &c.VaultEnv, &c.VaultUsernameSecret, &c.VaultFields,
		&c.ExpiresAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// CreateCredential inserts a new credential configuration.
func (ds *SQLDataStore) CreateCredential(credential *models.CredentialDB) error {
	if err := credential.Validate(); err != nil {
		return err
	}

	// Validate source - only vault and env allowed
	if credential.Source != "vault" && credential.Source != "env" {
		return fmt.Errorf("plaintext credentials not allowed: source must be 'vault' or 'env', got '%s'", credential.Source)
	}

	// Validate required fields based on source
	if credential.Source == "vault" && credential.VaultSecret == nil {
		return fmt.Errorf("vault_secret required for vault credentials")
	}
	if credential.Source == "env" && credential.EnvVar == nil {
		return fmt.Errorf("env_var required for env credentials")
	}

	query := fmt.Sprintf(`INSERT INTO credentials (scope_type, scope_id, name, source, env_var, description, username_var, password_var, vault_secret, vault_env, vault_username_secret, vault_fields, expires_at, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)`, ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query,
		credential.ScopeType,
		credential.ScopeID,
		credential.Name,
		credential.Source,
		credential.EnvVar,
		credential.Description,
		credential.UsernameVar,
		credential.PasswordVar,
		credential.VaultSecret,
		credential.VaultEnv,
		credential.VaultUsernameSecret,
		credential.VaultFields,
		credential.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create credential: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		credential.ID = id
	}

	return nil
}

// GetCredential retrieves a credential by scope and name.
func (ds *SQLDataStore) GetCredential(scopeType models.CredentialScopeType, scopeID int64, name string) (*models.CredentialDB, error) {
	query := `SELECT ` + credentialColumns + ` 
		FROM credentials WHERE scope_type = ? AND scope_id = ? AND name = ?`

	row := ds.driver.QueryRow(query, scopeType, scopeID, name)
	credential, err := scanCredential(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewErrNotFound("credential", fmt.Sprintf("%s (scope: %s, id: %d)", name, scopeType, scopeID))
		}
		return nil, fmt.Errorf("failed to scan credential: %w", err)
	}

	return credential, nil
}

// GetCredentialByName retrieves a credential by name across all scopes.
// Returns the first match if multiple credentials have the same name in different scopes.
func (ds *SQLDataStore) GetCredentialByName(name string) (*models.CredentialDB, error) {
	query := `SELECT ` + credentialColumns + ` 
		FROM credentials WHERE name = ? LIMIT 1`

	row := ds.driver.QueryRow(query, name)
	credential, err := scanCredential(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewErrNotFound("credential", name)
		}
		return nil, fmt.Errorf("failed to scan credential: %w", err)
	}

	return credential, nil
}

// UpdateCredential updates an existing credential.
func (ds *SQLDataStore) UpdateCredential(credential *models.CredentialDB) error {
	if err := credential.Validate(); err != nil {
		return err
	}

	// Validate source - only vault and env allowed
	if credential.Source != "vault" && credential.Source != "env" {
		return fmt.Errorf("plaintext credentials not allowed: source must be 'vault' or 'env', got '%s'", credential.Source)
	}

	// Validate required fields based on source
	if credential.Source == "vault" && credential.VaultSecret == nil {
		return fmt.Errorf("vault_secret required for vault credentials")
	}
	if credential.Source == "env" && credential.EnvVar == nil {
		return fmt.Errorf("env_var required for env credentials")
	}

	query := fmt.Sprintf(`UPDATE credentials SET source = ?, env_var = ?, description = ?, username_var = ?, password_var = ?, vault_secret = ?, vault_env = ?, vault_username_secret = ?, vault_fields = ?, expires_at = ?, updated_at = %s 
		WHERE scope_type = ? AND scope_id = ? AND name = ?`, ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query,
		credential.Source,
		credential.EnvVar,
		credential.Description,
		credential.UsernameVar,
		credential.PasswordVar,
		credential.VaultSecret,
		credential.VaultEnv,
		credential.VaultUsernameSecret,
		credential.VaultFields,
		credential.ExpiresAt,
		credential.ScopeType,
		credential.ScopeID,
		credential.Name,
	)
	if err != nil {
		return fmt.Errorf("failed to update credential: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return NewErrNotFound("credential", fmt.Sprintf("%s (scope: %s, id: %d)", credential.Name, credential.ScopeType, credential.ScopeID))
	}

	return nil
}

// DeleteCredential removes a credential by scope and name.
func (ds *SQLDataStore) DeleteCredential(scopeType models.CredentialScopeType, scopeID int64, name string) error {
	query := `DELETE FROM credentials WHERE scope_type = ? AND scope_id = ? AND name = ?`

	result, err := ds.driver.Execute(query, scopeType, scopeID, name)
	if err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return NewErrNotFound("credential", fmt.Sprintf("%s (scope: %s, id: %d)", name, scopeType, scopeID))
	}

	return nil
}

// ListCredentialsByScope retrieves all credentials for a specific scope.
func (ds *SQLDataStore) ListCredentialsByScope(scopeType models.CredentialScopeType, scopeID int64) ([]*models.CredentialDB, error) {
	query := `SELECT ` + credentialColumns + ` 
		FROM credentials WHERE scope_type = ? AND scope_id = ? ORDER BY name`

	rows, err := ds.driver.Query(query, scopeType, scopeID)
	if err != nil {
		return nil, fmt.Errorf("failed to list credentials: %w", err)
	}
	defer rows.Close()

	var credentials []*models.CredentialDB
	for rows.Next() {
		credential, err := scanCredential(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan credential: %w", err)
		}
		credentials = append(credentials, credential)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating credentials: %w", err)
	}

	return credentials, nil
}

// ListAllCredentials retrieves all credentials across all scopes.
func (ds *SQLDataStore) ListAllCredentials() ([]*models.CredentialDB, error) {
	query := `SELECT ` + credentialColumns + ` 
		FROM credentials ORDER BY scope_type, scope_id, name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all credentials: %w", err)
	}
	defer rows.Close()

	var credentials []*models.CredentialDB
	for rows.Next() {
		credential, err := scanCredential(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan credential: %w", err)
		}
		credentials = append(credentials, credential)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating credentials: %w", err)
	}

	return credentials, nil
}
