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
	query := fmt.Sprintf(`INSERT INTO ecosystems (name, description, theme, created_at, updated_at) 
		VALUES (?, ?, ?, %s, %s)`, ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, ecosystem.Name, ecosystem.Description, ecosystem.Theme)
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
	query := `SELECT id, name, description, theme, created_at, updated_at FROM ecosystems WHERE name = ?`

	row := ds.driver.QueryRow(query, name)
	if err := row.Scan(&ecosystem.ID, &ecosystem.Name, &ecosystem.Description, &ecosystem.Theme, &ecosystem.CreatedAt, &ecosystem.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ecosystem not found: %s", name)
		}
		return nil, fmt.Errorf("failed to scan ecosystem: %w", err)
	}

	return ecosystem, nil
}

// GetEcosystemByID retrieves an ecosystem by its ID.
func (ds *SQLDataStore) GetEcosystemByID(id int) (*models.Ecosystem, error) {
	ecosystem := &models.Ecosystem{}
	query := `SELECT id, name, description, theme, created_at, updated_at FROM ecosystems WHERE id = ?`

	row := ds.driver.QueryRow(query, id)
	if err := row.Scan(&ecosystem.ID, &ecosystem.Name, &ecosystem.Description, &ecosystem.Theme, &ecosystem.CreatedAt, &ecosystem.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ecosystem not found: %d", id)
		}
		return nil, fmt.Errorf("failed to scan ecosystem: %w", err)
	}

	return ecosystem, nil
}

// UpdateEcosystem updates an existing ecosystem.
func (ds *SQLDataStore) UpdateEcosystem(ecosystem *models.Ecosystem) error {
	query := fmt.Sprintf(`UPDATE ecosystems SET name = ?, description = ?, theme = ?, updated_at = %s WHERE id = ?`,
		ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query, ecosystem.Name, ecosystem.Description, ecosystem.Theme, ecosystem.ID)
	if err != nil {
		return fmt.Errorf("failed to update ecosystem: %w", err)
	}
	return nil
}

// DeleteEcosystem removes an ecosystem by name.
func (ds *SQLDataStore) DeleteEcosystem(name string) error {
	return ds.deleteByName("ecosystems", "ecosystem", name)
}

// ListEcosystems retrieves all ecosystems.
func (ds *SQLDataStore) ListEcosystems() ([]*models.Ecosystem, error) {
	query := `SELECT id, name, description, theme, created_at, updated_at FROM ecosystems ORDER BY name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list ecosystems: %w", err)
	}
	defer rows.Close()

	var ecosystems []*models.Ecosystem
	for rows.Next() {
		ecosystem := &models.Ecosystem{}
		if err := rows.Scan(&ecosystem.ID, &ecosystem.Name, &ecosystem.Description, &ecosystem.Theme, &ecosystem.CreatedAt, &ecosystem.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan ecosystem: %w", err)
		}
		ecosystems = append(ecosystems, ecosystem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over ecosystems: %w", err)
	}

	return ecosystems, nil
}
