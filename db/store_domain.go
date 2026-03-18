package db

import (
	"database/sql"
	"devopsmaestro/models"
	"fmt"
)

// =============================================================================
// Domain Operations
// =============================================================================

// CreateDomain inserts a new domain into the database.
func (ds *SQLDataStore) CreateDomain(domain *models.Domain) error {
	query := fmt.Sprintf(`INSERT INTO domains (ecosystem_id, name, description, theme, build_args, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, %s, %s)`, ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, domain.EcosystemID, domain.Name, domain.Description, domain.Theme, domain.BuildArgs)
	if err != nil {
		return fmt.Errorf("failed to create domain: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		domain.ID = int(id)
	}

	return nil
}

// GetDomainByName retrieves a domain by ecosystem ID and name.
func (ds *SQLDataStore) GetDomainByName(ecosystemID int, name string) (*models.Domain, error) {
	domain := &models.Domain{}
	query := `SELECT id, ecosystem_id, name, description, theme, build_args, created_at, updated_at FROM domains WHERE ecosystem_id = ? AND name = ?`

	row := ds.driver.QueryRow(query, ecosystemID, name)
	if err := row.Scan(&domain.ID, &domain.EcosystemID, &domain.Name, &domain.Description, &domain.Theme, &domain.BuildArgs, &domain.CreatedAt, &domain.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, NewErrNotFound("domain", name)
		}
		return nil, fmt.Errorf("failed to scan domain: %w", err)
	}

	return domain, nil
}

// GetDomainByID retrieves a domain by its ID.
func (ds *SQLDataStore) GetDomainByID(id int) (*models.Domain, error) {
	domain := &models.Domain{}
	query := `SELECT id, ecosystem_id, name, description, theme, build_args, created_at, updated_at FROM domains WHERE id = ?`

	row := ds.driver.QueryRow(query, id)
	if err := row.Scan(&domain.ID, &domain.EcosystemID, &domain.Name, &domain.Description, &domain.Theme, &domain.BuildArgs, &domain.CreatedAt, &domain.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, NewErrNotFound("domain", id)
		}
		return nil, fmt.Errorf("failed to scan domain: %w", err)
	}

	return domain, nil
}

// UpdateDomain updates an existing domain.
func (ds *SQLDataStore) UpdateDomain(domain *models.Domain) error {
	query := fmt.Sprintf(`UPDATE domains SET ecosystem_id = ?, name = ?, description = ?, theme = ?, build_args = ?, updated_at = %s WHERE id = ?`,
		ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query, domain.EcosystemID, domain.Name, domain.Description, domain.Theme, domain.BuildArgs, domain.ID)
	if err != nil {
		return fmt.Errorf("failed to update domain: %w", err)
	}
	return nil
}

// DeleteDomain removes a domain by ID.
// Also cleans up orphaned credentials scoped to this domain and its child apps/workspaces
// (polymorphic scope_type/scope_id has no FK constraint).
func (ds *SQLDataStore) DeleteDomain(id int) error {
	// Clean up credentials scoped to workspaces under apps in this domain
	if _, err := ds.driver.Execute(`DELETE FROM credentials WHERE scope_type = 'workspace' AND scope_id IN (SELECT w.id FROM workspaces w JOIN apps a ON w.app_id = a.id WHERE a.domain_id = ?)`, id); err != nil {
		return fmt.Errorf("failed to delete workspace credentials for domain: %w", err)
	}
	// Clean up credentials scoped to apps in this domain
	if _, err := ds.driver.Execute(`DELETE FROM credentials WHERE scope_type = 'app' AND scope_id IN (SELECT id FROM apps WHERE domain_id = ?)`, id); err != nil {
		return fmt.Errorf("failed to delete app credentials for domain: %w", err)
	}
	// Clean up credentials scoped to this domain
	if _, err := ds.driver.Execute(`DELETE FROM credentials WHERE scope_type = 'domain' AND scope_id = ?`, id); err != nil {
		return fmt.Errorf("failed to delete domain credentials: %w", err)
	}

	query := `DELETE FROM domains WHERE id = ?`
	result, err := ds.driver.Execute(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete domain: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return NewErrNotFound("domain", id)
	}
	return nil
}

// ListDomainsByEcosystem retrieves all domains for an ecosystem.
func (ds *SQLDataStore) ListDomainsByEcosystem(ecosystemID int) ([]*models.Domain, error) {
	query := `SELECT id, ecosystem_id, name, description, theme, build_args, created_at, updated_at FROM domains WHERE ecosystem_id = ? ORDER BY name`

	rows, err := ds.driver.Query(query, ecosystemID)
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}
	defer rows.Close()

	var domains []*models.Domain
	for rows.Next() {
		domain := &models.Domain{}
		if err := rows.Scan(&domain.ID, &domain.EcosystemID, &domain.Name, &domain.Description, &domain.Theme, &domain.BuildArgs, &domain.CreatedAt, &domain.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan domain: %w", err)
		}
		domains = append(domains, domain)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over domains: %w", err)
	}

	return domains, nil
}

// ListAllDomains retrieves all domains across all ecosystems.
func (ds *SQLDataStore) ListAllDomains() ([]*models.Domain, error) {
	query := `SELECT id, ecosystem_id, name, description, theme, build_args, created_at, updated_at FROM domains ORDER BY ecosystem_id, name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all domains: %w", err)
	}
	defer rows.Close()

	var domains []*models.Domain
	for rows.Next() {
		domain := &models.Domain{}
		if err := rows.Scan(&domain.ID, &domain.EcosystemID, &domain.Name, &domain.Description, &domain.Theme, &domain.BuildArgs, &domain.CreatedAt, &domain.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan domain: %w", err)
		}
		domains = append(domains, domain)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over domains: %w", err)
	}

	return domains, nil
}
