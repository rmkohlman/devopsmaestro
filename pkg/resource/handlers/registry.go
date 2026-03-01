package handlers

import (
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/resource"

	"gopkg.in/yaml.v3"
)

const KindRegistry = "Registry"

// RegistryHandler handles Registry resources.
type RegistryHandler struct{}

// NewRegistryHandler creates a new Registry handler.
func NewRegistryHandler() *RegistryHandler {
	return &RegistryHandler{}
}

func (h *RegistryHandler) Kind() string {
	return KindRegistry
}

// getDataStore extracts and type-asserts the DataStore from the context.
func (h *RegistryHandler) getDataStore(ctx resource.Context) (db.DataStore, error) {
	if ctx.DataStore == nil {
		return nil, fmt.Errorf("DataStore not provided in context")
	}

	ds, ok := ctx.DataStore.(db.DataStore)
	if !ok {
		return nil, fmt.Errorf("invalid DataStore type: %T", ctx.DataStore)
	}

	return ds, nil
}

// Apply creates or updates a registry from YAML data.
func (h *RegistryHandler) Apply(ctx resource.Context, data []byte) (resource.Resource, error) {
	var regYAML models.RegistryYAML
	if err := yaml.Unmarshal(data, &regYAML); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Get the datastore
	ds, err := h.getDataStore(ctx)
	if err != nil {
		return nil, err
	}

	// Create Registry model from YAML
	reg := &models.Registry{}
	reg.FromYAML(regYAML)

	// Validate
	if err := reg.Validate(); err != nil {
		return nil, err
	}

	// Check if registry exists
	existing, err := ds.GetRegistryByName(reg.Name)
	if err == nil && existing != nil {
		// Update existing
		reg.ID = existing.ID
		if err := ds.UpdateRegistry(reg); err != nil {
			return nil, fmt.Errorf("failed to update registry: %w", err)
		}
	} else {
		// Create new
		if err := ds.CreateRegistry(reg); err != nil {
			return nil, fmt.Errorf("failed to create registry: %w", err)
		}
	}

	return &RegistryResource{registry: reg}, nil
}

// Get retrieves a registry by name.
func (h *RegistryHandler) Get(ctx resource.Context, name string) (resource.Resource, error) {
	ds, err := h.getDataStore(ctx)
	if err != nil {
		return nil, err
	}

	reg, err := ds.GetRegistryByName(name)
	if err != nil {
		return nil, err
	}

	return &RegistryResource{registry: reg}, nil
}

// List retrieves all registries.
func (h *RegistryHandler) List(ctx resource.Context) ([]resource.Resource, error) {
	ds, err := h.getDataStore(ctx)
	if err != nil {
		return nil, err
	}

	registries, err := ds.ListRegistries()
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, len(registries))
	for i, reg := range registries {
		resources[i] = &RegistryResource{registry: reg}
	}

	return resources, nil
}

// Delete removes a registry by name.
func (h *RegistryHandler) Delete(ctx resource.Context, name string) error {
	ds, err := h.getDataStore(ctx)
	if err != nil {
		return err
	}

	return ds.DeleteRegistry(name)
}

// ToYAML converts a registry resource to YAML.
func (h *RegistryHandler) ToYAML(res resource.Resource) ([]byte, error) {
	regRes, ok := res.(*RegistryResource)
	if !ok {
		return nil, fmt.Errorf("resource is not a Registry")
	}

	regYAML := regRes.registry.ToYAML()

	return yaml.Marshal(regYAML)
}

// RegistryResource wraps a Registry model as a resource.Resource.
type RegistryResource struct {
	registry *models.Registry
}

func (r *RegistryResource) GetKind() string {
	return KindRegistry
}

func (r *RegistryResource) GetName() string {
	return r.registry.Name
}

func (r *RegistryResource) Validate() error {
	return r.registry.Validate()
}

// Registry returns the underlying Registry model.
func (r *RegistryResource) Registry() *models.Registry {
	return r.registry
}
