package db

import (
	"database/sql"
	"devopsmaestro/models"
	"fmt"
)

// =============================================================================
// CRD Operations (Custom Resource Definitions)
// =============================================================================

// CreateCRD inserts a new custom resource definition.
func (ds *SQLDataStore) CreateCRD(crd *models.CustomResourceDefinition) error {
	query := fmt.Sprintf(`INSERT INTO custom_resource_definitions 
		(kind, "group", singular, plural, short_names, scope, versions, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, %s, %s)`, ds.queryBuilder.Now(), ds.queryBuilder.Now())

	shortNames := sql.NullString{}
	if crd.ShortNames.Valid {
		shortNames = crd.ShortNames
	}

	versions := sql.NullString{}
	if crd.Versions.Valid {
		versions = crd.Versions
	}

	result, err := ds.driver.Execute(query, crd.Kind, crd.Group, crd.Singular, crd.Plural, shortNames, crd.Scope, versions)
	if err != nil {
		return fmt.Errorf("failed to create CRD: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		crd.ID = int(id)
	}

	return nil
}

// GetCRDByKind retrieves a CRD by its kind name.
func (ds *SQLDataStore) GetCRDByKind(kind string) (*models.CustomResourceDefinition, error) {
	crd := &models.CustomResourceDefinition{}
	query := `SELECT id, kind, "group", singular, plural, short_names, scope, versions, created_at, updated_at 
		FROM custom_resource_definitions WHERE kind = ?`

	row := ds.driver.QueryRow(query, kind)
	if err := row.Scan(&crd.ID, &crd.Kind, &crd.Group, &crd.Singular, &crd.Plural, &crd.ShortNames, &crd.Scope, &crd.Versions, &crd.CreatedAt, &crd.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("CRD not found: %s", kind)
		}
		return nil, fmt.Errorf("failed to scan CRD: %w", err)
	}

	return crd, nil
}

// UpdateCRD updates an existing CRD.
func (ds *SQLDataStore) UpdateCRD(crd *models.CustomResourceDefinition) error {
	query := fmt.Sprintf(`UPDATE custom_resource_definitions SET 
		kind = ?, "group" = ?, singular = ?, plural = ?, short_names = ?, scope = ?, versions = ?, updated_at = %s 
		WHERE id = ?`, ds.queryBuilder.Now())

	shortNames := sql.NullString{}
	if crd.ShortNames.Valid {
		shortNames = crd.ShortNames
	}

	versions := sql.NullString{}
	if crd.Versions.Valid {
		versions = crd.Versions
	}

	_, err := ds.driver.Execute(query, crd.Kind, crd.Group, crd.Singular, crd.Plural, shortNames, crd.Scope, versions, crd.ID)
	if err != nil {
		return fmt.Errorf("failed to update CRD: %w", err)
	}
	return nil
}

// DeleteCRD removes a CRD by kind.
func (ds *SQLDataStore) DeleteCRD(kind string) error {
	query := `DELETE FROM custom_resource_definitions WHERE kind = ?`
	result, err := ds.driver.Execute(query, kind)
	if err != nil {
		return fmt.Errorf("failed to delete CRD: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return NewErrNotFound("CRD", kind)
	}
	return nil
}

// ListCRDs retrieves all CRDs.
func (ds *SQLDataStore) ListCRDs() ([]*models.CustomResourceDefinition, error) {
	query := `SELECT id, kind, "group", singular, plural, short_names, scope, versions, created_at, updated_at 
		FROM custom_resource_definitions ORDER BY kind`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list CRDs: %w", err)
	}
	defer rows.Close()

	var crds []*models.CustomResourceDefinition
	for rows.Next() {
		crd := &models.CustomResourceDefinition{}
		if err := rows.Scan(&crd.ID, &crd.Kind, &crd.Group, &crd.Singular, &crd.Plural, &crd.ShortNames, &crd.Scope, &crd.Versions, &crd.CreatedAt, &crd.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan CRD: %w", err)
		}
		crds = append(crds, crd)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over CRDs: %w", err)
	}

	return crds, nil
}

// =============================================================================
// Custom Resource Operations (instances of CRDs)
// =============================================================================

// CreateCustomResource inserts a new custom resource instance.
func (ds *SQLDataStore) CreateCustomResource(resource *models.CustomResource) error {
	query := fmt.Sprintf(`INSERT INTO custom_resources 
		(kind, name, namespace, spec, status, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, %s, %s)`, ds.queryBuilder.Now(), ds.queryBuilder.Now())

	namespace := sql.NullString{}
	if resource.Namespace.Valid {
		namespace = resource.Namespace
	}

	spec := sql.NullString{}
	if resource.Spec.Valid {
		spec = resource.Spec
	}

	status := sql.NullString{}
	if resource.Status.Valid {
		status = resource.Status
	}

	result, err := ds.driver.Execute(query, resource.Kind, resource.Name, namespace, spec, status)
	if err != nil {
		return fmt.Errorf("failed to create custom resource: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		resource.ID = int(id)
	}

	return nil
}

// GetCustomResource retrieves a custom resource by kind, name, and namespace.
func (ds *SQLDataStore) GetCustomResource(kind, name, namespace string) (*models.CustomResource, error) {
	resource := &models.CustomResource{}
	query := `SELECT id, kind, name, namespace, spec, status, created_at, updated_at 
		FROM custom_resources WHERE kind = ? AND name = ? AND (namespace = ? OR (namespace IS NULL AND ? = ''))`

	row := ds.driver.QueryRow(query, kind, name, namespace, namespace)
	if err := row.Scan(&resource.ID, &resource.Kind, &resource.Name, &resource.Namespace, &resource.Spec, &resource.Status, &resource.CreatedAt, &resource.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("custom resource not found: %s/%s/%s", kind, namespace, name)
		}
		return nil, fmt.Errorf("failed to scan custom resource: %w", err)
	}

	return resource, nil
}

// UpdateCustomResource updates an existing custom resource.
func (ds *SQLDataStore) UpdateCustomResource(resource *models.CustomResource) error {
	query := fmt.Sprintf(`UPDATE custom_resources SET 
		kind = ?, name = ?, namespace = ?, spec = ?, status = ?, updated_at = %s 
		WHERE id = ?`, ds.queryBuilder.Now())

	namespace := sql.NullString{}
	if resource.Namespace.Valid {
		namespace = resource.Namespace
	}

	spec := sql.NullString{}
	if resource.Spec.Valid {
		spec = resource.Spec
	}

	status := sql.NullString{}
	if resource.Status.Valid {
		status = resource.Status
	}

	_, err := ds.driver.Execute(query, resource.Kind, resource.Name, namespace, spec, status, resource.ID)
	if err != nil {
		return fmt.Errorf("failed to update custom resource: %w", err)
	}
	return nil
}

// DeleteCustomResource removes a custom resource by kind, name, and namespace.
func (ds *SQLDataStore) DeleteCustomResource(kind, name, namespace string) error {
	query := `DELETE FROM custom_resources WHERE kind = ? AND name = ? AND (namespace = ? OR (namespace IS NULL AND ? = ''))`
	result, err := ds.driver.Execute(query, kind, name, namespace, namespace)
	if err != nil {
		return fmt.Errorf("failed to delete custom resource: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return NewErrNotFound("custom resource", fmt.Sprintf("%s/%s/%s", kind, namespace, name))
	}
	return nil
}

// ListCustomResources retrieves all custom resources of a given kind.
func (ds *SQLDataStore) ListCustomResources(kind string) ([]*models.CustomResource, error) {
	query := `SELECT id, kind, name, namespace, spec, status, created_at, updated_at 
		FROM custom_resources WHERE kind = ? ORDER BY name`

	rows, err := ds.driver.Query(query, kind)
	if err != nil {
		return nil, fmt.Errorf("failed to list custom resources: %w", err)
	}
	defer rows.Close()

	var resources []*models.CustomResource
	for rows.Next() {
		resource := &models.CustomResource{}
		if err := rows.Scan(&resource.ID, &resource.Kind, &resource.Name, &resource.Namespace, &resource.Spec, &resource.Status, &resource.CreatedAt, &resource.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan custom resource: %w", err)
		}
		resources = append(resources, resource)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over custom resources: %w", err)
	}

	return resources, nil
}
