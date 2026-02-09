package handlers

import (
	"database/sql"
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/resource"

	"gopkg.in/yaml.v3"
)

const KindDomain = "Domain"

// DomainHandler handles Domain resources.
type DomainHandler struct{}

// NewDomainHandler creates a new Domain handler.
func NewDomainHandler() *DomainHandler {
	return &DomainHandler{}
}

func (h *DomainHandler) Kind() string {
	return KindDomain
}

// Apply creates or updates a domain from YAML data.
func (h *DomainHandler) Apply(ctx resource.Context, data []byte) (resource.Resource, error) {
	// Parse the YAML
	var domainYAML models.DomainYAML
	if err := yaml.Unmarshal(data, &domainYAML); err != nil {
		return nil, fmt.Errorf("failed to parse domain YAML: %w", err)
	}

	// Get the datastore
	ds, err := h.getDataStore(ctx)
	if err != nil {
		return nil, err
	}

	// Resolve ecosystem from YAML metadata
	ecosystemName := domainYAML.Metadata.Ecosystem
	if ecosystemName == "" {
		return nil, fmt.Errorf("domain YAML must specify metadata.ecosystem")
	}

	ecosystem, err := ds.GetEcosystemByName(ecosystemName)
	if err != nil {
		return nil, fmt.Errorf("ecosystem '%s' not found: %w", ecosystemName, err)
	}

	// Convert to model
	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
	}
	domain.FromYAML(domainYAML)

	// Check if domain exists
	existing, _ := ds.GetDomainByName(ecosystem.ID, domain.Name)
	if existing != nil {
		// Update existing
		domain.ID = existing.ID
		if err := ds.UpdateDomain(domain); err != nil {
			return nil, fmt.Errorf("failed to update domain: %w", err)
		}
	} else {
		// Create new
		if err := ds.CreateDomain(domain); err != nil {
			return nil, fmt.Errorf("failed to create domain: %w", err)
		}
		// Fetch to get the ID
		domain, err = ds.GetDomainByName(ecosystem.ID, domain.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve created domain: %w", err)
		}
	}

	return &DomainResource{domain: domain, ecosystemName: ecosystemName}, nil
}

// Get retrieves a domain by name.
// Note: This requires an active ecosystem context to resolve the domain.
func (h *DomainHandler) Get(ctx resource.Context, name string) (resource.Resource, error) {
	ds, err := h.getDataStore(ctx)
	if err != nil {
		return nil, err
	}

	// Get active ecosystem from context
	dbCtx, err := ds.GetContext()
	if err != nil {
		return nil, fmt.Errorf("failed to get context: %w", err)
	}

	if dbCtx.ActiveEcosystemID == nil {
		return nil, fmt.Errorf("no active ecosystem set; use 'dvm use ecosystem <name>' first")
	}

	domain, err := ds.GetDomainByName(*dbCtx.ActiveEcosystemID, name)
	if err != nil {
		return nil, err
	}

	ecosystem, _ := ds.GetEcosystemByID(domain.EcosystemID)
	ecosystemName := ""
	if ecosystem != nil {
		ecosystemName = ecosystem.Name
	}

	return &DomainResource{domain: domain, ecosystemName: ecosystemName}, nil
}

// List returns all domains in the active ecosystem.
func (h *DomainHandler) List(ctx resource.Context) ([]resource.Resource, error) {
	ds, err := h.getDataStore(ctx)
	if err != nil {
		return nil, err
	}

	// Get active ecosystem from context
	dbCtx, err := ds.GetContext()
	if err != nil {
		return nil, fmt.Errorf("failed to get context: %w", err)
	}

	var domains []*models.Domain
	if dbCtx.ActiveEcosystemID != nil {
		domains, err = ds.ListDomainsByEcosystem(*dbCtx.ActiveEcosystemID)
	} else {
		// If no active ecosystem, list all domains
		domains, err = ds.ListAllDomains()
	}
	if err != nil {
		return nil, err
	}

	result := make([]resource.Resource, len(domains))
	for i, d := range domains {
		ecosystem, _ := ds.GetEcosystemByID(d.EcosystemID)
		ecosystemName := ""
		if ecosystem != nil {
			ecosystemName = ecosystem.Name
		}
		result[i] = &DomainResource{domain: d, ecosystemName: ecosystemName}
	}
	return result, nil
}

// Delete removes a domain by name.
func (h *DomainHandler) Delete(ctx resource.Context, name string) error {
	ds, err := h.getDataStore(ctx)
	if err != nil {
		return err
	}

	// Get active ecosystem from context
	dbCtx, err := ds.GetContext()
	if err != nil {
		return fmt.Errorf("failed to get context: %w", err)
	}

	if dbCtx.ActiveEcosystemID == nil {
		return fmt.Errorf("no active ecosystem set; use 'dvm use ecosystem <name>' first")
	}

	domain, err := ds.GetDomainByName(*dbCtx.ActiveEcosystemID, name)
	if err != nil {
		return err
	}

	return ds.DeleteDomain(domain.ID)
}

// ToYAML serializes a domain to YAML.
func (h *DomainHandler) ToYAML(res resource.Resource) ([]byte, error) {
	dr, ok := res.(*DomainResource)
	if !ok {
		return nil, fmt.Errorf("expected DomainResource, got %T", res)
	}

	yamlDoc := dr.domain.ToYAML(dr.ecosystemName)
	return yaml.Marshal(yamlDoc)
}

// getDataStore returns the DataStore from the context.
func (h *DomainHandler) getDataStore(ctx resource.Context) (db.DataStore, error) {
	if ctx.DataStore == nil {
		return nil, fmt.Errorf("DataStore not provided in context")
	}

	ds, ok := ctx.DataStore.(db.DataStore)
	if !ok {
		return nil, fmt.Errorf("invalid DataStore type: %T", ctx.DataStore)
	}

	return ds, nil
}

// DomainResource wraps a models.Domain to implement resource.Resource.
type DomainResource struct {
	domain        *models.Domain
	ecosystemName string
}

func (r *DomainResource) GetKind() string {
	return KindDomain
}

func (r *DomainResource) GetName() string {
	return r.domain.Name
}

func (r *DomainResource) Validate() error {
	if r.domain.Name == "" {
		return fmt.Errorf("domain name is required")
	}
	if r.domain.EcosystemID == 0 {
		return fmt.Errorf("domain ecosystem_id is required")
	}
	return nil
}

// Domain returns the underlying models.Domain.
func (r *DomainResource) Domain() *models.Domain {
	return r.domain
}

// EcosystemName returns the ecosystem name for this domain.
func (r *DomainResource) EcosystemName() string {
	return r.ecosystemName
}

// NewDomainResource creates a new DomainResource from a model.
func NewDomainResource(domain *models.Domain, ecosystemName string) *DomainResource {
	return &DomainResource{domain: domain, ecosystemName: ecosystemName}
}

// NewDomainFromModel creates a Domain model from parameters.
func NewDomainFromModel(name string, ecosystemID int, description string) *models.Domain {
	return &models.Domain{
		Name:        name,
		EcosystemID: ecosystemID,
		Description: sql.NullString{
			String: description,
			Valid:  description != "",
		},
	}
}
