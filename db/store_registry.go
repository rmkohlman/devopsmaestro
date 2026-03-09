package db

import (
	"database/sql"
	"devopsmaestro/models"
	"fmt"
	"strings"
)

// =============================================================================
// Registry Operations
// =============================================================================

// CreateRegistry inserts a new registry.
func (ds *SQLDataStore) CreateRegistry(registry *models.Registry) error {
	// Validate before creation
	if err := registry.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check for port conflicts if port is specified
	if registry.Port > 0 {
		existing, err := ds.GetRegistryByPort(registry.Port)
		if err == nil && existing != nil {
			return fmt.Errorf("port %d is already in use by registry '%s'", registry.Port, existing.Name)
		}
	}

	// Apply defaults
	if registry.Status == "" {
		registry.Status = "stopped"
	}
	if registry.Lifecycle == "" {
		registry.Lifecycle = "manual"
	}

	query := fmt.Sprintf(`INSERT INTO registries (name, type, enabled, port, lifecycle, storage, idle_timeout, description, config, status, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)`, ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, registry.Name, registry.Type, registry.Enabled, registry.Port, registry.Lifecycle, registry.Storage, registry.IdleTimeout, registry.Description, registry.Config, registry.Status)
	if err != nil {
		// Check for unique constraint violations
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "unique") {
			if strings.Contains(err.Error(), "name") {
				return fmt.Errorf("registry with name '%s' already exists", registry.Name)
			}
			if strings.Contains(err.Error(), "port") {
				return fmt.Errorf("port %d is already in use", registry.Port)
			}
		}
		return fmt.Errorf("failed to create registry: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}
	registry.ID = int(id)

	return nil
}

// GetRegistryByName retrieves a registry by name.
func (ds *SQLDataStore) GetRegistryByName(name string) (*models.Registry, error) {
	query := `SELECT id, name, type, enabled, port, lifecycle, storage, idle_timeout, description, config, status, created_at, updated_at 
		FROM registries WHERE name = ?`

	row := ds.driver.QueryRow(query, name)

	registry := &models.Registry{}
	err := row.Scan(&registry.ID, &registry.Name, &registry.Type, &registry.Enabled, &registry.Port, &registry.Lifecycle,
		&registry.Storage, &registry.IdleTimeout, &registry.Description, &registry.Config, &registry.Status, &registry.CreatedAt, &registry.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("registry '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to get registry: %w", err)
	}

	return registry, nil
}

// GetRegistryByID retrieves a registry by ID.
func (ds *SQLDataStore) GetRegistryByID(id int) (*models.Registry, error) {
	query := `SELECT id, name, type, enabled, port, lifecycle, storage, idle_timeout, description, config, status, created_at, updated_at 
		FROM registries WHERE id = ?`

	row := ds.driver.QueryRow(query, id)

	registry := &models.Registry{}
	err := row.Scan(&registry.ID, &registry.Name, &registry.Type, &registry.Enabled, &registry.Port, &registry.Lifecycle,
		&registry.Storage, &registry.IdleTimeout, &registry.Description, &registry.Config, &registry.Status, &registry.CreatedAt, &registry.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("registry with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get registry: %w", err)
	}

	return registry, nil
}

// GetRegistryByPort retrieves a registry by port (for conflict detection).
func (ds *SQLDataStore) GetRegistryByPort(port int) (*models.Registry, error) {
	query := `SELECT id, name, type, enabled, port, lifecycle, storage, idle_timeout, description, config, status, created_at, updated_at 
		FROM registries WHERE port = ?`

	row := ds.driver.QueryRow(query, port)

	registry := &models.Registry{}
	err := row.Scan(&registry.ID, &registry.Name, &registry.Type, &registry.Enabled, &registry.Port, &registry.Lifecycle,
		&registry.Storage, &registry.IdleTimeout, &registry.Description, &registry.Config, &registry.Status, &registry.CreatedAt, &registry.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no registry found on port %d", port)
		}
		return nil, fmt.Errorf("failed to get registry: %w", err)
	}

	return registry, nil
}

// UpdateRegistry updates an existing registry.
func (ds *SQLDataStore) UpdateRegistry(registry *models.Registry) error {
	// Validate before update
	if err := registry.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check for port conflicts (excluding this registry)
	if registry.Port > 0 {
		existing, err := ds.GetRegistryByPort(registry.Port)
		if err == nil && existing != nil && existing.ID != registry.ID {
			return fmt.Errorf("port %d is already in use by registry '%s'", registry.Port, existing.Name)
		}
	}

	query := fmt.Sprintf(`UPDATE registries 
		SET type = ?, enabled = ?, port = ?, lifecycle = ?, storage = ?, idle_timeout = ?, description = ?, config = ?, status = ?, updated_at = %s 
		WHERE id = ?`, ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, registry.Type, registry.Enabled, registry.Port, registry.Lifecycle,
		registry.Storage, registry.IdleTimeout, registry.Description, registry.Config, registry.Status, registry.ID)
	if err != nil {
		return fmt.Errorf("failed to update registry: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("registry with ID %d not found", registry.ID)
	}

	return nil
}

// DeleteRegistry removes a registry by name.
func (ds *SQLDataStore) DeleteRegistry(name string) error {
	// Check if registry exists
	_, err := ds.GetRegistryByName(name)
	if err != nil {
		return err
	}

	query := `DELETE FROM registries WHERE name = ?`

	result, err := ds.driver.Execute(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete registry: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("registry '%s' not found", name)
	}

	return nil
}

// ListRegistries retrieves all registries.
func (ds *SQLDataStore) ListRegistries() ([]*models.Registry, error) {
	query := `SELECT id, name, type, enabled, port, lifecycle, storage, idle_timeout, description, config, status, created_at, updated_at 
		FROM registries ORDER BY name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list registries: %w", err)
	}
	defer rows.Close()

	var registries []*models.Registry
	for rows.Next() {
		registry := &models.Registry{}
		err := rows.Scan(&registry.ID, &registry.Name, &registry.Type, &registry.Enabled, &registry.Port, &registry.Lifecycle,
			&registry.Storage, &registry.IdleTimeout, &registry.Description, &registry.Config, &registry.Status, &registry.CreatedAt, &registry.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan registry: %w", err)
		}
		registries = append(registries, registry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating registries: %w", err)
	}

	return registries, nil
}

// ListRegistriesByType retrieves registries by type.
func (ds *SQLDataStore) ListRegistriesByType(registryType string) ([]*models.Registry, error) {
	query := `SELECT id, name, type, enabled, port, lifecycle, storage, idle_timeout, description, config, status, created_at, updated_at 
		FROM registries WHERE type = ? ORDER BY name`

	rows, err := ds.driver.Query(query, registryType)
	if err != nil {
		return nil, fmt.Errorf("failed to list registries: %w", err)
	}
	defer rows.Close()

	var registries []*models.Registry
	for rows.Next() {
		registry := &models.Registry{}
		err := rows.Scan(&registry.ID, &registry.Name, &registry.Type, &registry.Enabled, &registry.Port, &registry.Lifecycle,
			&registry.Storage, &registry.IdleTimeout, &registry.Description, &registry.Config, &registry.Status, &registry.CreatedAt, &registry.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan registry: %w", err)
		}
		registries = append(registries, registry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating registries: %w", err)
	}

	return registries, nil
}

// ListRegistriesByStatus retrieves registries by status.
func (ds *SQLDataStore) ListRegistriesByStatus(status string) ([]*models.Registry, error) {
	query := `SELECT id, name, type, enabled, port, lifecycle, storage, idle_timeout, description, config, status, created_at, updated_at 
		FROM registries WHERE status = ? ORDER BY name`

	rows, err := ds.driver.Query(query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to list registries: %w", err)
	}
	defer rows.Close()

	var registries []*models.Registry
	for rows.Next() {
		registry := &models.Registry{}
		err := rows.Scan(&registry.ID, &registry.Name, &registry.Type, &registry.Enabled, &registry.Port, &registry.Lifecycle,
			&registry.Storage, &registry.IdleTimeout, &registry.Description, &registry.Config, &registry.Status, &registry.CreatedAt, &registry.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan registry: %w", err)
		}
		registries = append(registries, registry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating registries: %w", err)
	}

	return registries, nil
}
