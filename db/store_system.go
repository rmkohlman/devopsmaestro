package db

import (
	"database/sql"
	"errors"
	"fmt"

	"devopsmaestro/models"
)

// =============================================================================
// System Operations
// =============================================================================

// CreateSystem inserts a new system into the database.
func (ds *SQLDataStore) CreateSystem(system *models.System) error {
	query := fmt.Sprintf(`INSERT INTO systems (ecosystem_id, domain_id, name, description, theme, nvim_package, terminal_package, build_args, ca_certs, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)`, ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, system.EcosystemID, system.DomainID, system.Name, system.Description, system.Theme, system.NvimPackage, system.TerminalPackage, system.BuildArgs, system.CACerts)
	if err != nil {
		return fmt.Errorf("failed to create system: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		system.ID = int(id)
	}

	return nil
}

// GetSystemByID retrieves a system by its ID.
func (ds *SQLDataStore) GetSystemByID(id int) (*models.System, error) {
	system := &models.System{}
	query := `SELECT id, ecosystem_id, domain_id, name, description, theme, nvim_package, terminal_package, build_args, ca_certs, created_at, updated_at FROM systems WHERE id = ?`

	row := ds.driver.QueryRow(query, id)
	if err := row.Scan(&system.ID, &system.EcosystemID, &system.DomainID, &system.Name, &system.Description, &system.Theme, &system.NvimPackage, &system.TerminalPackage, &system.BuildArgs, &system.CACerts, &system.CreatedAt, &system.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewErrNotFound("system", id)
		}
		return nil, fmt.Errorf("failed to scan system: %w", err)
	}

	return system, nil
}

// GetSystemByName retrieves a system by domain ID and name.
// domainID is nullable because systems can exist without a parent domain.
func (ds *SQLDataStore) GetSystemByName(domainID sql.NullInt64, name string) (*models.System, error) {
	system := &models.System{}
	var query string
	var row Row

	if domainID.Valid {
		query = `SELECT id, ecosystem_id, domain_id, name, description, theme, nvim_package, terminal_package, build_args, ca_certs, created_at, updated_at FROM systems WHERE domain_id = ? AND name = ?`
		row = ds.driver.QueryRow(query, domainID.Int64, name)
	} else {
		query = `SELECT id, ecosystem_id, domain_id, name, description, theme, nvim_package, terminal_package, build_args, ca_certs, created_at, updated_at FROM systems WHERE domain_id IS NULL AND name = ?`
		row = ds.driver.QueryRow(query, name)
	}

	if err := row.Scan(&system.ID, &system.EcosystemID, &system.DomainID, &system.Name, &system.Description, &system.Theme, &system.NvimPackage, &system.TerminalPackage, &system.BuildArgs, &system.CACerts, &system.CreatedAt, &system.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewErrNotFound("system", name)
		}
		return nil, fmt.Errorf("failed to scan system: %w", err)
	}

	return system, nil
}

// UpdateSystem updates an existing system.
func (ds *SQLDataStore) UpdateSystem(system *models.System) error {
	query := fmt.Sprintf(`UPDATE systems SET ecosystem_id = ?, domain_id = ?, name = ?, description = ?, theme = ?, nvim_package = ?, terminal_package = ?, build_args = ?, ca_certs = ?, updated_at = %s WHERE id = ?`,
		ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query, system.EcosystemID, system.DomainID, system.Name, system.Description, system.Theme, system.NvimPackage, system.TerminalPackage, system.BuildArgs, system.CACerts, system.ID)
	if err != nil {
		return fmt.Errorf("failed to update system: %w", err)
	}
	return nil
}

// DeleteSystem removes a system by ID.
// Also cleans up orphaned credentials scoped to this system
// (polymorphic scope_type/scope_id has no FK constraint).
// Child app/workspace credentials are handled by ON DELETE CASCADE + app delete logic.
func (ds *SQLDataStore) DeleteSystem(id int) error {
	tx, err := ds.driver.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	// Clean up orphaned credentials scoped to this system
	if _, err := tx.Execute(`DELETE FROM credentials WHERE scope_type = 'system' AND scope_id = ?`, id); err != nil {
		return fmt.Errorf("failed to delete system credentials: %w", err)
	}

	query := `DELETE FROM systems WHERE id = ?`
	result, err := tx.Execute(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete system: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return NewErrNotFound("system", id)
	}

	return tx.Commit()
}

// ListSystems retrieves all systems across all domains.
func (ds *SQLDataStore) ListSystems() ([]*models.System, error) {
	query := `SELECT id, ecosystem_id, domain_id, name, description, theme, nvim_package, terminal_package, build_args, ca_certs, created_at, updated_at FROM systems ORDER BY domain_id, name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all systems: %w", err)
	}
	defer rows.Close()

	var systems []*models.System
	for rows.Next() {
		system := &models.System{}
		if err := rows.Scan(&system.ID, &system.EcosystemID, &system.DomainID, &system.Name, &system.Description, &system.Theme, &system.NvimPackage, &system.TerminalPackage, &system.BuildArgs, &system.CACerts, &system.CreatedAt, &system.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan system: %w", err)
		}
		systems = append(systems, system)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over systems: %w", err)
	}

	return systems, nil
}

// ListSystemsByDomain retrieves all systems for a domain.
func (ds *SQLDataStore) ListSystemsByDomain(domainID int) ([]*models.System, error) {
	query := `SELECT id, ecosystem_id, domain_id, name, description, theme, nvim_package, terminal_package, build_args, ca_certs, created_at, updated_at FROM systems WHERE domain_id = ? ORDER BY name`

	rows, err := ds.driver.Query(query, domainID)
	if err != nil {
		return nil, fmt.Errorf("failed to list systems: %w", err)
	}
	defer rows.Close()

	var systems []*models.System
	for rows.Next() {
		system := &models.System{}
		if err := rows.Scan(&system.ID, &system.EcosystemID, &system.DomainID, &system.Name, &system.Description, &system.Theme, &system.NvimPackage, &system.TerminalPackage, &system.BuildArgs, &system.CACerts, &system.CreatedAt, &system.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan system: %w", err)
		}
		systems = append(systems, system)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over systems: %w", err)
	}

	return systems, nil
}

// FindSystemsByName retrieves all systems with the given name,
// including their parent domain and ecosystem via LEFT JOINs.
// Returns an empty slice (not an error) if no systems match.
func (ds *SQLDataStore) FindSystemsByName(name string) ([]*models.SystemWithHierarchy, error) {
	query := `SELECT 
		s.id, s.ecosystem_id, s.domain_id, s.name, s.description, s.theme, s.nvim_package, s.terminal_package, s.build_args, s.ca_certs, s.created_at, s.updated_at,
		d.id, d.ecosystem_id, d.name, d.description, d.theme, d.nvim_package, d.terminal_package, d.build_args, d.ca_certs, d.created_at, d.updated_at,
		e.id, e.name, e.description, e.theme, e.nvim_package, e.terminal_package, e.build_args, e.ca_certs, e.created_at, e.updated_at
	FROM systems s
	LEFT JOIN domains d ON s.domain_id = d.id
	LEFT JOIN ecosystems e ON s.ecosystem_id = e.id
	WHERE s.name = ?
	ORDER BY e.name, d.name`

	rows, err := ds.driver.Query(query, name)
	if err != nil {
		return nil, fmt.Errorf("failed to find systems by name: %w", err)
	}
	defer rows.Close()

	var results []*models.SystemWithHierarchy
	for rows.Next() {
		system := &models.System{}
		var domainID, domainEcosystemID sql.NullInt64
		var domainName, domainDesc, domainTheme sql.NullString
		var domainNvimPkg, domainTermPkg, domainBuildArgs, domainCACerts sql.NullString
		var domainCreatedAt, domainUpdatedAt sql.NullTime
		var ecoID sql.NullInt64
		var ecoName, ecoDesc, ecoTheme sql.NullString
		var ecoNvimPkg, ecoTermPkg, ecoBuildArgs, ecoCACerts sql.NullString
		var ecoCreatedAt, ecoUpdatedAt sql.NullTime

		if err := rows.Scan(
			// System fields
			&system.ID, &system.EcosystemID, &system.DomainID,
			&system.Name, &system.Description, &system.Theme,
			&system.NvimPackage, &system.TerminalPackage,
			&system.BuildArgs, &system.CACerts,
			&system.CreatedAt, &system.UpdatedAt,
			// Domain fields (nullable via LEFT JOIN)
			&domainID, &domainEcosystemID, &domainName,
			&domainDesc, &domainTheme, &domainNvimPkg,
			&domainTermPkg, &domainBuildArgs, &domainCACerts,
			&domainCreatedAt, &domainUpdatedAt,
			// Ecosystem fields (nullable via LEFT JOIN)
			&ecoID, &ecoName, &ecoDesc, &ecoTheme,
			&ecoNvimPkg, &ecoTermPkg, &ecoBuildArgs,
			&ecoCACerts, &ecoCreatedAt, &ecoUpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan system with hierarchy: %w", err)
		}

		result := &models.SystemWithHierarchy{System: system}

		if domainID.Valid {
			result.Domain = &models.Domain{
				ID:              int(domainID.Int64),
				EcosystemID:     int(domainEcosystemID.Int64),
				Name:            domainName.String,
				Description:     domainDesc,
				Theme:           domainTheme,
				NvimPackage:     domainNvimPkg,
				TerminalPackage: domainTermPkg,
				BuildArgs:       domainBuildArgs,
				CACerts:         domainCACerts,
			}
			if domainCreatedAt.Valid {
				result.Domain.CreatedAt = domainCreatedAt.Time
			}
			if domainUpdatedAt.Valid {
				result.Domain.UpdatedAt = domainUpdatedAt.Time
			}
		}

		if ecoID.Valid {
			result.Ecosystem = &models.Ecosystem{
				ID:              int(ecoID.Int64),
				Name:            ecoName.String,
				Description:     ecoDesc,
				Theme:           ecoTheme,
				NvimPackage:     ecoNvimPkg,
				TerminalPackage: ecoTermPkg,
				BuildArgs:       ecoBuildArgs,
				CACerts:         ecoCACerts,
			}
			if ecoCreatedAt.Valid {
				result.Ecosystem.CreatedAt = ecoCreatedAt.Time
			}
			if ecoUpdatedAt.Valid {
				result.Ecosystem.UpdatedAt = ecoUpdatedAt.Time
			}
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over systems: %w", err)
	}

	return results, nil
}

// CountSystems returns the total number of systems.
func (ds *SQLDataStore) CountSystems() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM systems`

	row := ds.driver.QueryRow(query)
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count systems: %w", err)
	}

	return count, nil
}
