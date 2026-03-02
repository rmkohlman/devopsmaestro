package crd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"devopsmaestro/db"
	"devopsmaestro/models"
)

// DataStoreAdapter adapts db.DataStore to CRDStore and CustomResourceStore interfaces.
// This allows the CRD system to work with the database through the standard DataStore interface.
type DataStoreAdapter struct {
	ds db.DataStore
}

// NewDataStoreAdapter creates a new adapter
func NewDataStoreAdapter(ds db.DataStore) *DataStoreAdapter {
	return &DataStoreAdapter{ds: ds}
}

// CRDStore interface implementation

// CreateCRD creates a new custom resource definition
func (a *DataStoreAdapter) CreateCRD(crd *CRDDefinition) error {
	modelCRD := a.toModelCRD(crd)
	return a.ds.CreateCRD(modelCRD)
}

// GetCRD retrieves a CRD by kind
func (a *DataStoreAdapter) GetCRD(kind string) (*CRDDefinition, error) {
	modelCRD, err := a.ds.GetCRDByKind(kind)
	if err != nil {
		return nil, err
	}
	return a.toCRDDefinition(modelCRD), nil
}

// ListCRDs returns all CRDs
func (a *DataStoreAdapter) ListCRDs() ([]*CRDDefinition, error) {
	modelCRDs, err := a.ds.ListCRDs()
	if err != nil {
		return nil, err
	}
	result := make([]*CRDDefinition, len(modelCRDs))
	for i, m := range modelCRDs {
		result[i] = a.toCRDDefinition(m)
	}
	return result, nil
}

// UpdateCRD updates an existing CRD
func (a *DataStoreAdapter) UpdateCRD(crd *CRDDefinition) error {
	modelCRD := a.toModelCRD(crd)
	return a.ds.UpdateCRD(modelCRD)
}

// DeleteCRD removes a CRD by kind
func (a *DataStoreAdapter) DeleteCRD(kind string) error {
	return a.ds.DeleteCRD(kind)
}

// CustomResourceStore interface implementation

// CreateResource creates a new custom resource instance
func (a *DataStoreAdapter) CreateResource(kind string, data map[string]interface{}) error {
	// Extract metadata
	metadata, ok := data["metadata"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing or invalid metadata")
	}

	name, _ := metadata["name"].(string)
	namespace, _ := metadata["namespace"].(string)

	// Extract spec and status
	spec, _ := data["spec"].(map[string]interface{})
	status, _ := data["status"].(map[string]interface{})

	specJSON, _ := json.Marshal(spec)
	statusJSON, _ := json.Marshal(status)

	cr := &models.CustomResource{
		Kind:      kind,
		Name:      name,
		Namespace: sql.NullString{String: namespace, Valid: namespace != ""},
		Spec:      sql.NullString{String: string(specJSON), Valid: len(specJSON) > 0},
		Status:    sql.NullString{String: string(statusJSON), Valid: len(statusJSON) > 0},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return a.ds.CreateCustomResource(cr)
}

// GetResource retrieves a custom resource by kind and name
func (a *DataStoreAdapter) GetResource(kind, name string) (map[string]interface{}, error) {
	// For simple get, use empty namespace
	cr, err := a.ds.GetCustomResource(kind, name, "")
	if err != nil {
		return nil, err
	}
	return a.toResourceMap(cr), nil
}

// ListResources returns all resources of a given kind
func (a *DataStoreAdapter) ListResources(kind string) ([]map[string]interface{}, error) {
	crs, err := a.ds.ListCustomResources(kind)
	if err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, len(crs))
	for i, cr := range crs {
		result[i] = a.toResourceMap(cr)
	}
	return result, nil
}

// UpdateResource updates an existing custom resource
func (a *DataStoreAdapter) UpdateResource(kind string, data map[string]interface{}) error {
	metadata, ok := data["metadata"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing or invalid metadata")
	}

	name, _ := metadata["name"].(string)
	namespace, _ := metadata["namespace"].(string)

	// Get existing to preserve ID
	existing, err := a.ds.GetCustomResource(kind, name, namespace)
	if err != nil {
		return err
	}

	// Update spec and status
	spec, _ := data["spec"].(map[string]interface{})
	status, _ := data["status"].(map[string]interface{})

	specJSON, _ := json.Marshal(spec)
	statusJSON, _ := json.Marshal(status)

	existing.Spec = sql.NullString{String: string(specJSON), Valid: len(specJSON) > 0}
	existing.Status = sql.NullString{String: string(statusJSON), Valid: len(statusJSON) > 0}
	existing.UpdatedAt = time.Now()

	return a.ds.UpdateCustomResource(existing)
}

// DeleteResource removes a custom resource
func (a *DataStoreAdapter) DeleteResource(kind, name string) error {
	return a.ds.DeleteCustomResource(kind, name, "")
}

// Conversion helpers

// toModelCRD converts CRDDefinition to models.CustomResourceDefinition
func (a *DataStoreAdapter) toModelCRD(crd *CRDDefinition) *models.CustomResourceDefinition {
	shortNamesJSON, _ := json.Marshal(crd.Names.ShortNames)

	// Convert CRDVersion to models.CRDVersion for JSON serialization
	modelVersions := make([]models.CRDVersion, len(crd.Versions))
	for i, v := range crd.Versions {
		modelVersions[i] = models.CRDVersion{
			Name:    v.Name,
			Served:  v.Served,
			Storage: v.Storage,
			Schema:  v.Schema.OpenAPIV3Schema,
		}
	}
	versionsJSON, _ := json.Marshal(modelVersions)

	return &models.CustomResourceDefinition{
		Kind:       crd.Names.Kind,
		Group:      crd.Group,
		Singular:   crd.Names.Singular,
		Plural:     crd.Names.Plural,
		ShortNames: sql.NullString{String: string(shortNamesJSON), Valid: len(crd.Names.ShortNames) > 0},
		Scope:      crd.Scope,
		Versions:   sql.NullString{String: string(versionsJSON), Valid: len(crd.Versions) > 0},
		CreatedAt:  crd.CreatedAt,
		UpdatedAt:  crd.UpdatedAt,
	}
}

// toCRDDefinition converts models.CustomResourceDefinition to CRDDefinition
func (a *DataStoreAdapter) toCRDDefinition(m *models.CustomResourceDefinition) *CRDDefinition {
	var shortNames []string
	if m.ShortNames.Valid && m.ShortNames.String != "" {
		json.Unmarshal([]byte(m.ShortNames.String), &shortNames)
	}

	var modelVersions []models.CRDVersion
	if m.Versions.Valid && m.Versions.String != "" {
		json.Unmarshal([]byte(m.Versions.String), &modelVersions)
	}

	// Convert models.CRDVersion to CRDVersion
	versions := make([]CRDVersion, len(modelVersions))
	for i, mv := range modelVersions {
		versions[i] = CRDVersion{
			Name:    mv.Name,
			Served:  mv.Served,
			Storage: mv.Storage,
			Schema: CRDSchema{
				OpenAPIV3Schema: mv.Schema,
			},
		}
	}

	return &CRDDefinition{
		Group: m.Group,
		Names: CRDNames{
			Kind:       m.Kind,
			Singular:   m.Singular,
			Plural:     m.Plural,
			ShortNames: shortNames,
		},
		Scope:     m.Scope,
		Versions:  versions,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// toResourceMap converts models.CustomResource to map[string]interface{}
func (a *DataStoreAdapter) toResourceMap(cr *models.CustomResource) map[string]interface{} {
	metadata := map[string]interface{}{
		"name": cr.Name,
	}
	if cr.Namespace.Valid && cr.Namespace.String != "" {
		metadata["namespace"] = cr.Namespace.String
	}

	var spec map[string]interface{}
	if cr.Spec.Valid && cr.Spec.String != "" {
		json.Unmarshal([]byte(cr.Spec.String), &spec)
	}

	var status map[string]interface{}
	if cr.Status.Valid && cr.Status.String != "" {
		json.Unmarshal([]byte(cr.Status.String), &status)
	}

	result := map[string]interface{}{
		"kind":     cr.Kind,
		"metadata": metadata,
	}
	if spec != nil {
		result["spec"] = spec
	}
	if status != nil {
		result["status"] = status
	}

	return result
}
