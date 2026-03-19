package db

import (
	"database/sql"
	"devopsmaestro/models"
	"fmt"
)

// =============================================================================
// Package Operations
// =============================================================================

// CreatePackage inserts a new nvim package into the database.
func (ds *SQLDataStore) CreatePackage(pkg *models.NvimPackageDB) error {
	query := fmt.Sprintf(`INSERT INTO nvim_packages (name, description, category, labels, plugins, extends, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, %s, %s)`, ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, pkg.Name, pkg.Description, pkg.Category, pkg.Labels, pkg.Plugins, pkg.Extends)
	if err != nil {
		return fmt.Errorf("failed to create package: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		pkg.ID = int(id)
	}

	return nil
}

// UpdatePackage updates an existing nvim package.
func (ds *SQLDataStore) UpdatePackage(pkg *models.NvimPackageDB) error {
	query := fmt.Sprintf(`UPDATE nvim_packages 
		SET description = ?, category = ?, labels = ?, plugins = ?, extends = ?, updated_at = %s 
		WHERE name = ?`, ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, pkg.Description, pkg.Category, pkg.Labels, pkg.Plugins, pkg.Extends, pkg.Name)
	if err != nil {
		return fmt.Errorf("failed to update package: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected == 0 {
		return NewErrNotFound("package", pkg.Name)
	}

	return nil
}

// UpsertPackage creates or updates an nvim package (by name) atomically using ON CONFLICT.
func (ds *SQLDataStore) UpsertPackage(pkg *models.NvimPackageDB) error {
	query := fmt.Sprintf(`INSERT INTO nvim_packages (name, description, category, labels, plugins, extends, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, %s, %s)
		%s, updated_at = %s`,
		ds.queryBuilder.Now(), ds.queryBuilder.Now(),
		ds.queryBuilder.UpsertSuffix([]string{"name"}, []string{
			"description", "category", "labels", "plugins", "extends",
		}),
		ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, pkg.Name, pkg.Description, pkg.Category, pkg.Labels, pkg.Plugins, pkg.Extends)
	if err != nil {
		return fmt.Errorf("failed to upsert package: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		pkg.ID = int(id)
	}

	return nil
}

// DeletePackage removes a package by name.
func (ds *SQLDataStore) DeletePackage(name string) error {
	return ds.deleteByName("nvim_packages", "package", name)
}

// GetPackage retrieves a package by its name.
func (ds *SQLDataStore) GetPackage(name string) (*models.NvimPackageDB, error) {
	pkg := &models.NvimPackageDB{}
	query := `SELECT id, name, description, category, labels, plugins, extends, created_at, updated_at FROM nvim_packages WHERE name = ?`

	row := ds.driver.QueryRow(query, name)
	if err := row.Scan(&pkg.ID, &pkg.Name, &pkg.Description, &pkg.Category, &pkg.Labels, &pkg.Plugins, &pkg.Extends, &pkg.CreatedAt, &pkg.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, NewErrNotFound("package", name)
		}
		return nil, fmt.Errorf("failed to scan package: %w", err)
	}

	return pkg, nil
}

// ListPackages retrieves all packages.
func (ds *SQLDataStore) ListPackages() ([]*models.NvimPackageDB, error) {
	query := `SELECT id, name, description, category, labels, plugins, extends, created_at, updated_at FROM nvim_packages ORDER BY name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}
	defer rows.Close()

	var packages []*models.NvimPackageDB
	for rows.Next() {
		pkg := &models.NvimPackageDB{}
		if err := rows.Scan(&pkg.ID, &pkg.Name, &pkg.Description, &pkg.Category, &pkg.Labels, &pkg.Plugins, &pkg.Extends, &pkg.CreatedAt, &pkg.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan package: %w", err)
		}
		packages = append(packages, pkg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating packages: %w", err)
	}

	return packages, nil
}

// ListPackagesByLabel retrieves packages that have a specific label key-value pair.
func (ds *SQLDataStore) ListPackagesByLabel(key, value string) ([]*models.NvimPackageDB, error) {
	// Validate key to prevent SQL injection via json_extract path
	if err := validateLabelKey(key); err != nil {
		return nil, fmt.Errorf("invalid label key: %w", err)
	}

	// Use JSON_EXTRACT if available (SQLite 3.45+) or simple string matching as fallback
	// key is validated above to contain only safe characters (alphanumeric, hyphens, underscores, dots)
	query := `SELECT id, name, description, category, labels, plugins, extends, created_at, updated_at 
		FROM nvim_packages 
		WHERE labels IS NOT NULL 
		AND (json_extract(labels, '$.' || ?) = ? OR labels LIKE '%"' || ? || '":"' || ? || '"%')
		ORDER BY name`

	rows, err := ds.driver.Query(query, key, value, key, value)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages by label: %w", err)
	}
	defer rows.Close()

	var packages []*models.NvimPackageDB
	for rows.Next() {
		pkg := &models.NvimPackageDB{}
		if err := rows.Scan(&pkg.ID, &pkg.Name, &pkg.Description, &pkg.Category, &pkg.Labels, &pkg.Plugins, &pkg.Extends, &pkg.CreatedAt, &pkg.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan package: %w", err)
		}

		// Double-check the label match in Go for accuracy
		labels := pkg.GetLabels()
		if labels[key] == value {
			packages = append(packages, pkg)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating packages: %w", err)
	}

	return packages, nil
}
