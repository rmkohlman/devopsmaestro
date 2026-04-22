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
//
// Reparent semantics (issue #397): if a system with the given name already
// exists in the same ecosystem under a *different* domain than the YAML
// specifies, Apply detects the parent change and delegates to Move so that
// both denormalized FKs (DomainID, EcosystemID) and child Apps are updated
// atomically — instead of silently creating a duplicate.
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
	ecosystemName := systemYAML.Metadata.Ecosystem
	var domainID sql.NullInt64
	if domainName != "" {
		if ecosystemName != "" {
			// Ecosystem explicitly provided — look up directly
			ecosystem, err := ds.GetEcosystemByName(ecosystemName)
			if err != nil {
				return nil, fmt.Errorf("ecosystem '%s' not found: %w", ecosystemName, err)
			}
			domain, err := ds.GetDomainByName(sql.NullInt64{Int64: int64(ecosystem.ID), Valid: true}, domainName)
			if err != nil {
				return nil, fmt.Errorf("domain '%s' not found in ecosystem '%s': %w", domainName, ecosystemName, err)
			}
			domainID = sql.NullInt64{Int64: int64(domain.ID), Valid: true}
		} else {
			// Ecosystem not provided — infer from domain name
			matches, err := ds.FindDomainsByName(domainName)
			if err != nil {
				return nil, fmt.Errorf("failed to look up domain '%s': %w", domainName, err)
			}
			if len(matches) == 0 {
				return nil, fmt.Errorf("domain '%s' not found", domainName)
			}
			if len(matches) > 1 {
				return nil, fmt.Errorf("domain '%s' exists in multiple ecosystems; specify metadata.ecosystem to disambiguate", domainName)
			}
			domainID = sql.NullInt64{Int64: int64(matches[0].Domain.ID), Valid: true}
			if matches[0].Ecosystem != nil {
				ecosystemName = matches[0].Ecosystem.Name
			}
		}
	}

	system := &models.System{
		DomainID: domainID,
	}
	system.FromYAML(systemYAML)

	// === Reparent detection (#397) ===
	// First do a global lookup by name to catch systems that exist under a
	// different parent than the one specified in the YAML. The previous
	// implementation looked up by (newDomainID, name), which silently created
	// a duplicate row when the system already existed under a different domain.
	globalMatches, _ := ds.FindSystemsByName(system.Name)
	var existingGlobal *models.System
	if ecosystemName != "" {
		// Disambiguate by ecosystem hint when present.
		for _, m := range globalMatches {
			if m.Ecosystem != nil && m.Ecosystem.Name == ecosystemName {
				existingGlobal = m.System
				break
			}
		}
	} else if len(globalMatches) == 1 {
		existingGlobal = globalMatches[0].System
	}

	if existingGlobal != nil && domainID.Valid && existingGlobal.DomainID != domainID {
		// Parent changed — delegate to Move so denormalized FKs and child
		// apps are updated atomically.
		if err := ds.MoveSystem(existingGlobal.ID, domainID); err != nil {
			return nil, fmt.Errorf("failed to reparent system '%s' to domain id %d: %w", system.Name, domainID.Int64, err)
		}
		moved, err := ds.GetSystemByID(existingGlobal.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve moved system: %w", err)
		}
		// Apply any non-parent field updates from YAML on top.
		moved.Description = system.Description
		if err := ds.UpdateSystem(moved); err != nil {
			return nil, fmt.Errorf("failed to update system after move: %w", err)
		}
		return NewSystemResource(moved, domainName, ecosystemName), nil
	}

	// Check if system exists in the (possibly nil) target domain
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

	return NewSystemResource(system, domainName, ecosystemName), nil
}

// Move reparents a system to a new domain (issue #397).
//
// Resolves `name` globally (with optional ecosystem disambiguator), validates
// the target Domain exists in the same Ecosystem, and delegates the atomic
// FK rewrite + child-app cascade to ds.MoveSystem.
//
// Idempotent: if the system is already in the target domain, returns NoOp=true
// with no DB writes.
//
// Returns a *MoveResult wrapped as a resource.Resource via MoveResultResource
// so callers can render kubectl-style output.
func (h *SystemHandler) Move(ctx resource.Context, name string, target MoveTarget) (*MoveResult, error) {
	ds, err := dataStoreFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	// Resolve the system being moved (global lookup, ecosystem hint if given).
	matches, err := ds.FindSystemsByName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to find system '%s': %w", name, err)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("system '%s' not found", name)
	}
	var src *models.SystemWithHierarchy
	if target.EcosystemName != "" {
		for _, m := range matches {
			if m.Ecosystem != nil && m.Ecosystem.Name == target.EcosystemName {
				src = m
				break
			}
		}
		if src == nil {
			return nil, fmt.Errorf("system '%s' not found in ecosystem '%s'", name, target.EcosystemName)
		}
	} else {
		if len(matches) > 1 {
			return nil, fmt.Errorf("system '%s' exists in multiple ecosystems; specify -e <ecosystem> to disambiguate", name)
		}
		src = matches[0]
	}

	// Resolve target domain.
	targetDomain, targetEcoName, err := resolveDomainTarget(ds, target.EcosystemName, target.DomainName)
	if err != nil {
		return nil, err
	}

	// Cross-ecosystem move guard: target Domain's ecosystem must match the
	// system's current ecosystem (no implicit ecosystem hops).
	if src.Ecosystem != nil && targetDomain.EcosystemID.Valid {
		if int64(src.Ecosystem.ID) != targetDomain.EcosystemID.Int64 {
			return nil, fmt.Errorf("cannot move system across ecosystems (current: %s, target domain in: %s)",
				src.Ecosystem.Name, targetEcoName)
		}
	}

	fromParent := "(no domain)"
	if src.Domain != nil {
		fromParent = "domain/" + src.Domain.Name
	}
	toParent := "domain/" + targetDomain.Name

	// Idempotent: already in target domain.
	if src.System.DomainID.Valid && src.System.DomainID.Int64 == int64(targetDomain.ID) {
		return &MoveResult{
			Kind: "system", Name: name,
			FromParent: fromParent, ToParent: toParent, NoOp: true,
		}, nil
	}

	// Count cascade children for the result message.
	childApps, _ := ds.ListAppsByDomain(int(src.System.DomainID.Int64))
	cascade := 0
	for _, a := range childApps {
		if a.SystemID.Valid && a.SystemID.Int64 == int64(src.System.ID) {
			cascade++
		}
	}

	newDomainID := sql.NullInt64{Int64: int64(targetDomain.ID), Valid: true}
	if err := ds.MoveSystem(src.System.ID, newDomainID); err != nil {
		return nil, fmt.Errorf("failed to move system '%s': %w", name, err)
	}

	return &MoveResult{
		Kind: "system", Name: name,
		FromParent: fromParent, ToParent: toParent, CascadedApps: cascade,
	}, nil
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
