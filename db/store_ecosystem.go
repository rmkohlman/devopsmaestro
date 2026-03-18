package db

import (
	"database/sql"
	"devopsmaestro/models"
	"fmt"
)

// =============================================================================
// Ecosystem Operations
// =============================================================================

// CreateEcosystem inserts a new ecosystem into the database.
func (ds *SQLDataStore) CreateEcosystem(ecosystem *models.Ecosystem) error {
	query := fmt.Sprintf(`INSERT INTO ecosystems (name, description, theme, build_args, ca_certs, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, %s, %s)`, ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, ecosystem.Name, ecosystem.Description, ecosystem.Theme, ecosystem.BuildArgs, ecosystem.CACerts)
	if err != nil {
		return fmt.Errorf("failed to create ecosystem: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		ecosystem.ID = int(id)
	}

	return nil
}

// GetEcosystemByName retrieves an ecosystem by its name.
func (ds *SQLDataStore) GetEcosystemByName(name string) (*models.Ecosystem, error) {
	ecosystem := &models.Ecosystem{}
	query := `SELECT id, name, description, theme, build_args, ca_certs, created_at, updated_at FROM ecosystems WHERE name = ?`

	row := ds.driver.QueryRow(query, name)
	if err := row.Scan(&ecosystem.ID, &ecosystem.Name, &ecosystem.Description, &ecosystem.Theme, &ecosystem.BuildArgs, &ecosystem.CACerts, &ecosystem.CreatedAt, &ecosystem.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, NewErrNotFound("ecosystem", name)
		}
		return nil, fmt.Errorf("failed to scan ecosystem: %w", err)
	}

	return ecosystem, nil
}

// GetEcosystemByID retrieves an ecosystem by its ID.
func (ds *SQLDataStore) GetEcosystemByID(id int) (*models.Ecosystem, error) {
	ecosystem := &models.Ecosystem{}
	query := `SELECT id, name, description, theme, build_args, ca_certs, created_at, updated_at FROM ecosystems WHERE id = ?`

	row := ds.driver.QueryRow(query, id)
	if err := row.Scan(&ecosystem.ID, &ecosystem.Name, &ecosystem.Description, &ecosystem.Theme, &ecosystem.BuildArgs, &ecosystem.CACerts, &ecosystem.CreatedAt, &ecosystem.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, NewErrNotFound("ecosystem", id)
		}
		return nil, fmt.Errorf("failed to scan ecosystem: %w", err)
	}

	return ecosystem, nil
}

// UpdateEcosystem updates an existing ecosystem.
func (ds *SQLDataStore) UpdateEcosystem(ecosystem *models.Ecosystem) error {
	query := fmt.Sprintf(`UPDATE ecosystems SET name = ?, description = ?, theme = ?, build_args = ?, ca_certs = ?, updated_at = %s WHERE id = ?`,
		ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query, ecosystem.Name, ecosystem.Description, ecosystem.Theme, ecosystem.BuildArgs, ecosystem.CACerts, ecosystem.ID)
	if err != nil {
		return fmt.Errorf("failed to update ecosystem: %w", err)
	}
	return nil
}

// DeleteEcosystem removes an ecosystem by name.
// Also cleans up orphaned credentials scoped to this ecosystem and all its
// child domains/apps/workspaces (polymorphic scope_type/scope_id has no FK constraint).
func (ds *SQLDataStore) DeleteEcosystem(name string) error {
	// Look up ecosystem ID for credential cleanup
	var ecoID int
	row := ds.driver.QueryRow(`SELECT id FROM ecosystems WHERE name = ?`, name)
	if err := row.Scan(&ecoID); err != nil {
		return NewErrNotFound("ecosystem", name)
	}

	// Clean up credentials scoped to workspaces under apps in domains in this ecosystem
	if _, err := ds.driver.Execute(`DELETE FROM credentials WHERE scope_type = 'workspace' AND scope_id IN (SELECT w.id FROM workspaces w JOIN apps a ON w.app_id = a.id JOIN domains d ON a.domain_id = d.id WHERE d.ecosystem_id = ?)`, ecoID); err != nil {
		return fmt.Errorf("failed to delete workspace credentials for ecosystem: %w", err)
	}
	// Clean up credentials scoped to apps in domains in this ecosystem
	if _, err := ds.driver.Execute(`DELETE FROM credentials WHERE scope_type = 'app' AND scope_id IN (SELECT a.id FROM apps a JOIN domains d ON a.domain_id = d.id WHERE d.ecosystem_id = ?)`, ecoID); err != nil {
		return fmt.Errorf("failed to delete app credentials for ecosystem: %w", err)
	}
	// Clean up credentials scoped to domains in this ecosystem
	if _, err := ds.driver.Execute(`DELETE FROM credentials WHERE scope_type = 'domain' AND scope_id IN (SELECT id FROM domains WHERE ecosystem_id = ?)`, ecoID); err != nil {
		return fmt.Errorf("failed to delete domain credentials for ecosystem: %w", err)
	}
	// Clean up credentials scoped to this ecosystem
	if _, err := ds.driver.Execute(`DELETE FROM credentials WHERE scope_type = 'ecosystem' AND scope_id = ?`, ecoID); err != nil {
		return fmt.Errorf("failed to delete ecosystem credentials: %w", err)
	}

	// Delete the ecosystem (cascades to domains -> apps -> workspaces via FK ON DELETE CASCADE)
	query := `DELETE FROM ecosystems WHERE name = ?`
	result, err := ds.driver.Execute(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete ecosystem: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return NewErrNotFound("ecosystem", name)
	}
	return nil
}

// ListEcosystems retrieves all ecosystems.
func (ds *SQLDataStore) ListEcosystems() ([]*models.Ecosystem, error) {
	query := `SELECT id, name, description, theme, build_args, ca_certs, created_at, updated_at FROM ecosystems ORDER BY name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list ecosystems: %w", err)
	}
	defer rows.Close()

	var ecosystems []*models.Ecosystem
	for rows.Next() {
		ecosystem := &models.Ecosystem{}
		if err := rows.Scan(&ecosystem.ID, &ecosystem.Name, &ecosystem.Description, &ecosystem.Theme, &ecosystem.BuildArgs, &ecosystem.CACerts, &ecosystem.CreatedAt, &ecosystem.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan ecosystem: %w", err)
		}
		ecosystems = append(ecosystems, ecosystem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over ecosystems: %w", err)
	}

	return ecosystems, nil
}
