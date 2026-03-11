package db

import (
	"database/sql"
	"devopsmaestro/models"
	"fmt"
)

// =============================================================================
// Terminal Package Operations
// =============================================================================

// CreateTerminalPackage inserts a new terminal package into the database.
func (ds *SQLDataStore) CreateTerminalPackage(pkg *models.TerminalPackageDB) error {
	query := fmt.Sprintf(`INSERT INTO terminal_packages (name, description, category, labels, plugins, prompts, profiles, wezterm, extends, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)`, ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, pkg.Name, pkg.Description, pkg.Category, pkg.Labels, pkg.Plugins, pkg.Prompts, pkg.Profiles, pkg.WezTerm, pkg.Extends)
	if err != nil {
		return fmt.Errorf("failed to create terminal package: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		pkg.ID = int(id)
	}

	return nil
}

// UpdateTerminalPackage updates an existing terminal package.
func (ds *SQLDataStore) UpdateTerminalPackage(pkg *models.TerminalPackageDB) error {
	query := fmt.Sprintf(`UPDATE terminal_packages 
		SET description = ?, category = ?, labels = ?, plugins = ?, prompts = ?, profiles = ?, wezterm = ?, extends = ?, updated_at = %s 
		WHERE name = ?`, ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, pkg.Description, pkg.Category, pkg.Labels, pkg.Plugins, pkg.Prompts, pkg.Profiles, pkg.WezTerm, pkg.Extends, pkg.Name)
	if err != nil {
		return fmt.Errorf("failed to update terminal package: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected == 0 {
		return NewErrNotFound("terminal package", pkg.Name)
	}

	return nil
}

// UpsertTerminalPackage creates or updates a terminal package (by name).
func (ds *SQLDataStore) UpsertTerminalPackage(pkg *models.TerminalPackageDB) error {
	// Try to get existing package first
	existing, err := ds.GetTerminalPackage(pkg.Name)
	if err != nil {
		// Package doesn't exist, create it
		return ds.CreateTerminalPackage(pkg)
	}

	// Package exists, update it (preserve the ID)
	pkg.ID = existing.ID
	return ds.UpdateTerminalPackage(pkg)
}

// DeleteTerminalPackage removes a terminal package by name.
func (ds *SQLDataStore) DeleteTerminalPackage(name string) error {
	return ds.deleteByName("terminal_packages", "terminal package", name)
}

// GetTerminalPackage retrieves a terminal package by its name.
func (ds *SQLDataStore) GetTerminalPackage(name string) (*models.TerminalPackageDB, error) {
	pkg := &models.TerminalPackageDB{}
	query := `SELECT id, name, description, category, labels, plugins, prompts, profiles, wezterm, extends, created_at, updated_at FROM terminal_packages WHERE name = ?`

	row := ds.driver.QueryRow(query, name)
	if err := row.Scan(&pkg.ID, &pkg.Name, &pkg.Description, &pkg.Category, &pkg.Labels, &pkg.Plugins, &pkg.Prompts, &pkg.Profiles, &pkg.WezTerm, &pkg.Extends, &pkg.CreatedAt, &pkg.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, NewErrNotFound("terminal package", name)
		}
		return nil, fmt.Errorf("failed to scan terminal package: %w", err)
	}

	return pkg, nil
}

// ListTerminalPackages retrieves all terminal packages.
func (ds *SQLDataStore) ListTerminalPackages() ([]*models.TerminalPackageDB, error) {
	query := `SELECT id, name, description, category, labels, plugins, prompts, profiles, wezterm, extends, created_at, updated_at FROM terminal_packages ORDER BY name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list terminal packages: %w", err)
	}
	defer rows.Close()

	var packages []*models.TerminalPackageDB
	for rows.Next() {
		pkg := &models.TerminalPackageDB{}
		if err := rows.Scan(&pkg.ID, &pkg.Name, &pkg.Description, &pkg.Category, &pkg.Labels, &pkg.Plugins, &pkg.Prompts, &pkg.Profiles, &pkg.WezTerm, &pkg.Extends, &pkg.CreatedAt, &pkg.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan terminal package: %w", err)
		}
		packages = append(packages, pkg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating terminal packages: %w", err)
	}

	return packages, nil
}

// ListTerminalPackagesByLabel retrieves terminal packages that have a specific label key-value pair.
func (ds *SQLDataStore) ListTerminalPackagesByLabel(key, value string) ([]*models.TerminalPackageDB, error) {
	// Use JSON_EXTRACT if available (SQLite 3.45+) or simple string matching as fallback
	query := `SELECT id, name, description, category, labels, plugins, prompts, profiles, wezterm, extends, created_at, updated_at 
		FROM terminal_packages 
		WHERE labels IS NOT NULL 
		AND (json_extract(labels, '$.` + key + `') = ? OR labels LIKE '%"' || ? || '":"' || ? || '"%')
		ORDER BY name`

	rows, err := ds.driver.Query(query, value, key, value)
	if err != nil {
		return nil, fmt.Errorf("failed to list terminal packages by label: %w", err)
	}
	defer rows.Close()

	var packages []*models.TerminalPackageDB
	for rows.Next() {
		pkg := &models.TerminalPackageDB{}
		if err := rows.Scan(&pkg.ID, &pkg.Name, &pkg.Description, &pkg.Category, &pkg.Labels, &pkg.Plugins, &pkg.Prompts, &pkg.Profiles, &pkg.WezTerm, &pkg.Extends, &pkg.CreatedAt, &pkg.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan terminal package: %w", err)
		}

		// Double-check the label match in Go for accuracy
		labels := pkg.GetLabels()
		if labels[key] == value {
			packages = append(packages, pkg)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating terminal packages: %w", err)
	}

	return packages, nil
}
