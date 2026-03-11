package db

import (
	"database/sql"
	"devopsmaestro/models"
	"fmt"
)

// =============================================================================
// Credential Operations
// =============================================================================

// CreateCredential inserts a new credential configuration.
func (ds *SQLDataStore) CreateCredential(credential *models.CredentialDB) error {
	// Validate source - only keychain and env allowed
	if credential.Source != "keychain" && credential.Source != "env" {
		return fmt.Errorf("plaintext credentials not allowed: source must be 'keychain' or 'env', got '%s'", credential.Source)
	}

	// Validate required fields based on source
	if credential.Source == "keychain" && credential.Service == nil {
		return fmt.Errorf("service required for keychain credentials")
	}
	if credential.Source == "env" && credential.EnvVar == nil {
		return fmt.Errorf("env_var required for env credentials")
	}

	query := fmt.Sprintf(`INSERT INTO credentials (scope_type, scope_id, name, source, service, env_var, description, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, %s, %s)`, ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query,
		credential.ScopeType,
		credential.ScopeID,
		credential.Name,
		credential.Source,
		credential.Service,
		credential.EnvVar,
		credential.Description,
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
	credential := &models.CredentialDB{}
	query := `SELECT id, scope_type, scope_id, name, source, service, env_var, description, created_at, updated_at 
		FROM credentials WHERE scope_type = ? AND scope_id = ? AND name = ?`

	row := ds.driver.QueryRow(query, scopeType, scopeID, name)
	if err := row.Scan(
		&credential.ID,
		&credential.ScopeType,
		&credential.ScopeID,
		&credential.Name,
		&credential.Source,
		&credential.Service,
		&credential.EnvVar,
		&credential.Description,
		&credential.CreatedAt,
		&credential.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, NewErrNotFound("credential", fmt.Sprintf("%s (scope: %s, id: %d)", name, scopeType, scopeID))
		}
		return nil, fmt.Errorf("failed to scan credential: %w", err)
	}

	return credential, nil
}

// UpdateCredential updates an existing credential.
func (ds *SQLDataStore) UpdateCredential(credential *models.CredentialDB) error {
	// Validate source - only keychain and env allowed
	if credential.Source != "keychain" && credential.Source != "env" {
		return fmt.Errorf("plaintext credentials not allowed: source must be 'keychain' or 'env', got '%s'", credential.Source)
	}

	// Validate required fields based on source
	if credential.Source == "keychain" && credential.Service == nil {
		return fmt.Errorf("service required for keychain credentials")
	}
	if credential.Source == "env" && credential.EnvVar == nil {
		return fmt.Errorf("env_var required for env credentials")
	}

	query := fmt.Sprintf(`UPDATE credentials SET source = ?, service = ?, env_var = ?, description = ?, updated_at = %s 
		WHERE scope_type = ? AND scope_id = ? AND name = ?`, ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query,
		credential.Source,
		credential.Service,
		credential.EnvVar,
		credential.Description,
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
	query := `SELECT id, scope_type, scope_id, name, source, service, env_var, description, created_at, updated_at 
		FROM credentials WHERE scope_type = ? AND scope_id = ? ORDER BY name`

	rows, err := ds.driver.Query(query, scopeType, scopeID)
	if err != nil {
		return nil, fmt.Errorf("failed to list credentials: %w", err)
	}
	defer rows.Close()

	var credentials []*models.CredentialDB
	for rows.Next() {
		credential := &models.CredentialDB{}
		if err := rows.Scan(
			&credential.ID,
			&credential.ScopeType,
			&credential.ScopeID,
			&credential.Name,
			&credential.Source,
			&credential.Service,
			&credential.EnvVar,
			&credential.Description,
			&credential.CreatedAt,
			&credential.UpdatedAt,
		); err != nil {
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
	query := `SELECT id, scope_type, scope_id, name, source, service, env_var, description, created_at, updated_at 
		FROM credentials ORDER BY scope_type, scope_id, name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all credentials: %w", err)
	}
	defer rows.Close()

	var credentials []*models.CredentialDB
	for rows.Next() {
		credential := &models.CredentialDB{}
		if err := rows.Scan(
			&credential.ID,
			&credential.ScopeType,
			&credential.ScopeID,
			&credential.Name,
			&credential.Source,
			&credential.Service,
			&credential.EnvVar,
			&credential.Description,
			&credential.CreatedAt,
			&credential.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan credential: %w", err)
		}
		credentials = append(credentials, credential)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating credentials: %w", err)
	}

	return credentials, nil
}
