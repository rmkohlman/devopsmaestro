package handlers

import (
	"database/sql"
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/resource"

	"gopkg.in/yaml.v3"
)

const KindEcosystem = "Ecosystem"

// EcosystemHandler handles Ecosystem resources.
type EcosystemHandler struct{}

// NewEcosystemHandler creates a new Ecosystem handler.
func NewEcosystemHandler() *EcosystemHandler {
	return &EcosystemHandler{}
}

func (h *EcosystemHandler) Kind() string {
	return KindEcosystem
}

// Apply creates or updates an ecosystem from YAML data.
func (h *EcosystemHandler) Apply(ctx resource.Context, data []byte) (resource.Resource, error) {
	// Parse the YAML
	var ecosystemYAML models.EcosystemYAML
	if err := yaml.Unmarshal(data, &ecosystemYAML); err != nil {
		return nil, fmt.Errorf("failed to parse ecosystem YAML: %w", err)
	}

	// Convert to model
	ecosystem := &models.Ecosystem{}
	ecosystem.FromYAML(ecosystemYAML)

	// Get the datastore
	ds, err := h.getDataStore(ctx)
	if err != nil {
		return nil, err
	}

	// Check if ecosystem exists
	existing, _ := ds.GetEcosystemByName(ecosystem.Name)
	if existing != nil {
		// Update existing
		ecosystem.ID = existing.ID
		if err := ds.UpdateEcosystem(ecosystem); err != nil {
			return nil, fmt.Errorf("failed to update ecosystem: %w", err)
		}
	} else {
		// Create new
		if err := ds.CreateEcosystem(ecosystem); err != nil {
			return nil, fmt.Errorf("failed to create ecosystem: %w", err)
		}
		// Fetch to get the ID
		ecosystem, err = ds.GetEcosystemByName(ecosystem.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve created ecosystem: %w", err)
		}
	}

	return &EcosystemResource{ecosystem: ecosystem}, nil
}

// Get retrieves an ecosystem by name.
func (h *EcosystemHandler) Get(ctx resource.Context, name string) (resource.Resource, error) {
	ds, err := h.getDataStore(ctx)
	if err != nil {
		return nil, err
	}

	ecosystem, err := ds.GetEcosystemByName(name)
	if err != nil {
		return nil, err
	}

	return &EcosystemResource{ecosystem: ecosystem}, nil
}

// List returns all ecosystems.
func (h *EcosystemHandler) List(ctx resource.Context) ([]resource.Resource, error) {
	ds, err := h.getDataStore(ctx)
	if err != nil {
		return nil, err
	}

	ecosystems, err := ds.ListEcosystems()
	if err != nil {
		return nil, err
	}

	result := make([]resource.Resource, len(ecosystems))
	for i, e := range ecosystems {
		result[i] = &EcosystemResource{ecosystem: e}
	}
	return result, nil
}

// Delete removes an ecosystem by name.
func (h *EcosystemHandler) Delete(ctx resource.Context, name string) error {
	ds, err := h.getDataStore(ctx)
	if err != nil {
		return err
	}

	return ds.DeleteEcosystem(name)
}

// ToYAML serializes an ecosystem to YAML.
func (h *EcosystemHandler) ToYAML(res resource.Resource) ([]byte, error) {
	er, ok := res.(*EcosystemResource)
	if !ok {
		return nil, fmt.Errorf("expected EcosystemResource, got %T", res)
	}

	yamlDoc := er.ecosystem.ToYAML()
	return yaml.Marshal(yamlDoc)
}

// getDataStore returns the DataStore from the context.
func (h *EcosystemHandler) getDataStore(ctx resource.Context) (db.DataStore, error) {
	if ctx.DataStore == nil {
		return nil, fmt.Errorf("DataStore not provided in context")
	}

	ds, ok := ctx.DataStore.(db.DataStore)
	if !ok {
		return nil, fmt.Errorf("invalid DataStore type: %T", ctx.DataStore)
	}

	return ds, nil
}

// EcosystemResource wraps a models.Ecosystem to implement resource.Resource.
type EcosystemResource struct {
	ecosystem *models.Ecosystem
}

func (r *EcosystemResource) GetKind() string {
	return KindEcosystem
}

func (r *EcosystemResource) GetName() string {
	return r.ecosystem.Name
}

func (r *EcosystemResource) Validate() error {
	if r.ecosystem.Name == "" {
		return fmt.Errorf("ecosystem name is required")
	}
	return nil
}

// Ecosystem returns the underlying models.Ecosystem.
func (r *EcosystemResource) Ecosystem() *models.Ecosystem {
	return r.ecosystem
}

// NewEcosystemResource creates a new EcosystemResource from a model.
func NewEcosystemResource(ecosystem *models.Ecosystem) *EcosystemResource {
	return &EcosystemResource{ecosystem: ecosystem}
}

// NewEcosystemFromModel creates an Ecosystem model from parameters.
func NewEcosystemFromModel(name, description string) *models.Ecosystem {
	return &models.Ecosystem{
		Name: name,
		Description: sql.NullString{
			String: description,
			Valid:  description != "",
		},
	}
}
