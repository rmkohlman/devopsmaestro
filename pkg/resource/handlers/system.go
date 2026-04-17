package handlers

import (
	"database/sql"
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/resource"

	"gopkg.in/yaml.v3"
)

const KindSystem = "System"

// SystemHandler handles System resources.
type SystemHandler struct{}

// NewSystemHandler creates a new System handler.
func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

func (h *SystemHandler) Kind() string {
	return KindSystem
}

// Apply creates or updates a system from YAML data.
func (h *SystemHandler) Apply(ctx resource.Context, data []byte) (resource.Resource, error) {
	var systemYAML models.SystemYAML
	if err := yaml.Unmarshal(data, &systemYAML); err != nil {
		return nil, fmt.Errorf("failed to parse system YAML: %w", err)
	}

	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return nil, err
	}

	// Resolve domain from YAML metadata (optional for systems)
	domainName := systemYAML.Metadata.Domain
	var domainID sql.NullInt64
	if domainName != "" {
		ecosystemName := systemYAML.Metadata.Ecosystem
		if ecosystemName == "" {
			return nil, fmt.Errorf("system YAML with domain must also specify metadata.ecosystem")
		}
		ecosystem, err := ds.GetEcosystemByName(ecosystemName)
		if err != nil {
			return nil, fmt.Errorf("ecosystem '%s' not found: %w", ecosystemName, err)
		}
		domain, err := ds.GetDomainByName(sql.NullInt64{Int64: int64(ecosystem.ID), Valid: true}, domainName)
		if err != nil {
			return nil, fmt.Errorf("domain '%s' not found in ecosystem '%s': %w", domainName, ecosystemName, err)
		}
		domainID = sql.NullInt64{Int64: int64(domain.ID), Valid: true}
	}

	system := &models.System{
		DomainID: domainID,
	}
	system.FromYAML(systemYAML)

	// Check if system exists
	existing, _ := ds.GetSystemByName(domainID, system.Name)
	if existing != nil {
		system.ID = existing.ID
		if err := ds.UpdateSystem(system); err != nil {
			return nil, fmt.Errorf("failed to update system: %w", err)
		}
	} else {
		if err := ds.CreateSystem(system); err != nil {
			return nil, fmt.Errorf("failed to create system: %w", err)
		}
		system, err = ds.GetSystemByName(domainID, system.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve created system: %w", err)
		}
	}

	return NewSystemResource(system, domainName, systemYAML.Metadata.Ecosystem), nil
}

// Get retrieves a system by name.
func (h *SystemHandler) Get(ctx resource.Context, name string) (resource.Resource, error) {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return nil, err
	}

	// Get active domain from context (optional for systems)
	dbCtx, err := ds.GetContext()
	if err != nil {
		return nil, fmt.Errorf("failed to get context: %w", err)
	}

	var domainID sql.NullInt64
	if dbCtx.ActiveDomainID != nil {
		domainID = sql.NullInt64{Int64: int64(*dbCtx.ActiveDomainID), Valid: true}
	}

	system, err := ds.GetSystemByName(domainID, name)
	if err != nil {
		return nil, err
	}

	domainName := ""
	ecosystemName := ""
	if system.DomainID.Valid {
		if domain, e := ds.GetDomainByID(int(system.DomainID.Int64)); e == nil {
			domainName = domain.Name
			if domain.EcosystemID.Valid {
				if eco, e2 := ds.GetEcosystemByID(int(domain.EcosystemID.Int64)); e2 == nil {
					ecosystemName = eco.Name
				}
			}
		}
	}

	return NewSystemResource(system, domainName, ecosystemName), nil
}

// List returns all systems.
func (h *SystemHandler) List(ctx resource.Context) ([]resource.Resource, error) {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return nil, err
	}

	dbCtx, err := ds.GetContext()
	if err != nil {
		return nil, fmt.Errorf("failed to get context: %w", err)
	}

	var systems []*models.System
	if dbCtx.ActiveDomainID != nil {
		systems, err = ds.ListSystemsByDomain(*dbCtx.ActiveDomainID)
	} else {
		systems, err = ds.ListSystems()
	}
	if err != nil {
		return nil, err
	}

	result := make([]resource.Resource, len(systems))
	for i, s := range systems {
		domainName := ""
		ecosystemName := ""
		if s.DomainID.Valid {
			if domain, e := ds.GetDomainByID(int(s.DomainID.Int64)); e == nil {
				domainName = domain.Name
				if domain.EcosystemID.Valid {
					if eco, e2 := ds.GetEcosystemByID(int(domain.EcosystemID.Int64)); e2 == nil {
						ecosystemName = eco.Name
					}
				}
			}
		}
		result[i] = NewSystemResource(s, domainName, ecosystemName)
	}
	return result, nil
}

// Delete removes a system by name.
func (h *SystemHandler) Delete(ctx resource.Context, name string) error {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return err
	}

	dbCtx, err := ds.GetContext()
	if err != nil {
		return fmt.Errorf("failed to get context: %w", err)
	}

	var domainID sql.NullInt64
	if dbCtx.ActiveDomainID != nil {
		domainID = sql.NullInt64{Int64: int64(*dbCtx.ActiveDomainID), Valid: true}
	}

	system, err := ds.GetSystemByName(domainID, name)
	if err != nil {
		return err
	}

	return ds.DeleteSystem(system.ID)
}

// ToYAML serializes a system to YAML.
func (h *SystemHandler) ToYAML(res resource.Resource) ([]byte, error) {
	sr, ok := res.(*SystemResource)
	if !ok {
		return nil, fmt.Errorf("expected SystemResource, got %T", res)
	}

	yamlDoc := sr.system.ToYAML(sr.domainName, sr.ecosystemName, nil)
	return yaml.Marshal(yamlDoc)
}

// SystemResource wraps a models.System to implement resource.Resource.
type SystemResource struct {
	system        *models.System
	domainName    string
	ecosystemName string
}

func (r *SystemResource) GetKind() string {
	return KindSystem
}

func (r *SystemResource) GetName() string {
	return r.system.Name
}

func (r *SystemResource) Validate() error {
	if r.system.Name == "" {
		return fmt.Errorf("system name is required")
	}
	return nil
}

// System returns the underlying models.System.
func (r *SystemResource) System() *models.System {
	return r.system
}

// DomainName returns the domain name for this system.
func (r *SystemResource) DomainName() string {
	return r.domainName
}

// EcosystemName returns the ecosystem name for this system.
func (r *SystemResource) EcosystemName() string {
	return r.ecosystemName
}

// NewSystemResource creates a new SystemResource from a model.
func NewSystemResource(system *models.System, domainName, ecosystemName string) *SystemResource {
	return &SystemResource{system: system, domainName: domainName, ecosystemName: ecosystemName}
}

// NewSystemFromModel creates a System model from parameters.
func NewSystemFromModel(name string, domainID sql.NullInt64, ecosystemID sql.NullInt64, description string) *models.System {
	return &models.System{
		Name:        name,
		DomainID:    domainID,
		EcosystemID: ecosystemID,
		Description: sql.NullString{
			String: description,
			Valid:  description != "",
		},
	}
}
