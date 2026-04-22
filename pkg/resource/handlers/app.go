package handlers

import (
	"database/sql"
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/resource"

	"gopkg.in/yaml.v3"
)

const KindApp = "App"

// AppHandler handles App resources.
type AppHandler struct{}

// NewAppHandler creates a new App handler.
func NewAppHandler() *AppHandler {
	return &AppHandler{}
}

func (h *AppHandler) Kind() string {
	return KindApp
}

// Apply creates or updates an app from YAML data.
func (h *AppHandler) Apply(ctx resource.Context, data []byte) (resource.Resource, error) {
	// Parse the YAML
	var appYAML models.AppYAML
	if err := yaml.Unmarshal(data, &appYAML); err != nil {
		return nil, fmt.Errorf("failed to parse app YAML: %w", err)
	}

	// Get the datastore
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return nil, err
	}

	// Resolve domain from YAML metadata
	domainName := appYAML.Metadata.Domain
	if domainName == "" {
		return nil, fmt.Errorf("app YAML must specify metadata.domain")
	}

	// Resolve ecosystem: try metadata.ecosystem first, then fall back to active context
	var ecosystemID int
	if appYAML.Metadata.Ecosystem != "" {
		// Context-free path: resolve ecosystem by name from YAML metadata
		eco, err := ds.GetEcosystemByName(appYAML.Metadata.Ecosystem)
		if err != nil {
			return nil, fmt.Errorf("ecosystem '%s' not found: %w", appYAML.Metadata.Ecosystem, err)
		}
		ecosystemID = eco.ID
	} else {
		// Fall back to active ecosystem context (existing behavior)
		dbCtx, err := ds.GetContext()
		if err != nil {
			return nil, fmt.Errorf("failed to get context: %w", err)
		}
		if dbCtx.ActiveEcosystemID == nil {
			return nil, fmt.Errorf("no active ecosystem set and no metadata.ecosystem specified; use 'dvm use ecosystem <name>' or add metadata.ecosystem to YAML")
		}
		ecosystemID = *dbCtx.ActiveEcosystemID
	}

	domain, err := ds.GetDomainByName(sql.NullInt64{Int64: int64(ecosystemID), Valid: true}, domainName)
	if err != nil {
		return nil, fmt.Errorf("domain '%s' not found: %w", domainName, err)
	}

	// Convert to model
	domainNullID := sql.NullInt64{Int64: int64(domain.ID), Valid: true}
	app := &models.App{
		DomainID: domainNullID,
	}
	app.FromYAML(appYAML)

	// Resolve GitRepo if specified in YAML
	if appYAML.Spec.GitRepo != "" {
		gitRepo, err := ds.GetGitRepoByName(appYAML.Spec.GitRepo)
		if err != nil {
			return nil, fmt.Errorf("gitrepo '%s' not found: %w", appYAML.Spec.GitRepo, err)
		}
		app.GitRepoID = sql.NullInt64{Int64: int64(gitRepo.ID), Valid: true}
	}

	// Resolve System if specified in YAML metadata
	if appYAML.Metadata.System != "" {
		// Global system lookup: find system across all domains (#287)
		matches, findErr := ds.FindSystemsByName(appYAML.Metadata.System)
		if findErr != nil {
			return nil, fmt.Errorf("failed to find system '%s': %w", appYAML.Metadata.System, findErr)
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("system '%s' not found", appYAML.Metadata.System)
		}
		var chosen *models.SystemWithHierarchy
		if len(matches) == 1 {
			chosen = matches[0]
		} else {
			// Multiple matches — use the YAML domain as tie-breaker
			for _, m := range matches {
				if m.Domain != nil && m.Domain.ID == domain.ID {
					chosen = m
					break
				}
			}
			if chosen == nil {
				return nil, fmt.Errorf("ambiguous system name '%s' — found in multiple domains; specify metadata.domain to disambiguate", appYAML.Metadata.System)
			}
		}
		app.SystemID = sql.NullInt64{Int64: int64(chosen.System.ID), Valid: true}
		// If system's domain differs from the YAML domain, use system's domain
		if chosen.Domain != nil && chosen.Domain.ID != domain.ID {
			domain = chosen.Domain
			domainName = domain.Name
			domainNullID = sql.NullInt64{Int64: int64(domain.ID), Valid: true}
			app.DomainID = domainNullID
		}
	}

	// === Reparent detection (#397) ===
	// Look up the app globally first; if it already exists in this ecosystem
	// under a different (domain, system) pair than the YAML specifies,
	// delegate to MoveApp so denormalized FKs stay consistent — instead of
	// silently creating a duplicate row.
	globalMatches, _ := ds.FindAppsByName(app.Name)
	var existingGlobal *models.App
	for _, m := range globalMatches {
		if m.Ecosystem != nil && m.Ecosystem.ID == ecosystemID {
			existingGlobal = m.App
			break
		}
	}
	if existingGlobal != nil {
		parentChanged := existingGlobal.DomainID != domainNullID || existingGlobal.SystemID != app.SystemID
		if parentChanged {
			if err := ds.MoveApp(existingGlobal.ID, domainNullID, app.SystemID); err != nil {
				return nil, fmt.Errorf("failed to reparent app '%s': %w", app.Name, err)
			}
			app.ID = existingGlobal.ID
			if err := ds.UpdateApp(app); err != nil {
				return nil, fmt.Errorf("failed to update app after move: %w", err)
			}
			moved, err := ds.GetAppByID(existingGlobal.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to retrieve moved app: %w", err)
			}
			return &AppResource{app: moved, domainName: domainName, ecosystemName: appYAML.Metadata.Ecosystem, systemName: appYAML.Metadata.System, gitRepoName: appYAML.Spec.GitRepo}, nil
		}
	}

	// Check if app exists
	existing, _ := ds.GetAppByName(domainNullID, app.Name)
	if existing != nil {
		// Update existing
		app.ID = existing.ID
		if err := ds.UpdateApp(app); err != nil {
			return nil, fmt.Errorf("failed to update app: %w", err)
		}
	} else {
		// Create new
		if err := ds.CreateApp(app); err != nil {
			return nil, fmt.Errorf("failed to create app: %w", err)
		}
		// Fetch to get the ID
		app, err = ds.GetAppByName(domainNullID, app.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve created app: %w", err)
		}
	}

	return &AppResource{app: app, domainName: domainName, ecosystemName: appYAML.Metadata.Ecosystem, systemName: appYAML.Metadata.System, gitRepoName: appYAML.Spec.GitRepo}, nil
}

// Get retrieves an app by name.
// Note: This requires an active domain context to resolve the app.
func (h *AppHandler) Get(ctx resource.Context, name string) (resource.Resource, error) {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return nil, err
	}

	// Get active domain from context
	dbCtx, err := ds.GetContext()
	if err != nil {
		return nil, fmt.Errorf("failed to get context: %w", err)
	}

	if dbCtx.ActiveDomainID == nil {
		return nil, fmt.Errorf("no active domain set; use 'dvm use domain <name>' first")
	}

	app, err := ds.GetAppByName(sql.NullInt64{Int64: int64(*dbCtx.ActiveDomainID), Valid: true}, name)
	if err != nil {
		return nil, err
	}

	domainName := ""
	ecosystemName := ""
	if app.DomainID.Valid {
		domain, _ := ds.GetDomainByID(int(app.DomainID.Int64))
		if domain != nil {
			domainName = domain.Name
			// Resolve ecosystem name for round-trip fidelity
			if domain.EcosystemID.Valid {
				if eco, ecoErr := ds.GetEcosystemByID(int(domain.EcosystemID.Int64)); ecoErr == nil {
					ecosystemName = eco.Name
				}
			}
		}
	}

	// Resolve GitRepo name if GitRepoID is set
	gitRepoName := ""
	if app.GitRepoID.Valid {
		gitRepo, grErr := ds.GetGitRepoByID(app.GitRepoID.Int64)
		if grErr == nil && gitRepo != nil {
			gitRepoName = gitRepo.Name
		}
	}

	// Resolve System name if SystemID is set
	systemName := ""
	if app.SystemID.Valid {
		system, sErr := ds.GetSystemByID(int(app.SystemID.Int64))
		if sErr == nil && system != nil {
			systemName = system.Name
		}
	}

	return &AppResource{app: app, domainName: domainName, ecosystemName: ecosystemName, systemName: systemName, gitRepoName: gitRepoName}, nil
}

// List returns all apps in the active domain.
func (h *AppHandler) List(ctx resource.Context) ([]resource.Resource, error) {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return nil, err
	}

	// Get active domain from context
	dbCtx, err := ds.GetContext()
	if err != nil {
		return nil, fmt.Errorf("failed to get context: %w", err)
	}

	var apps []*models.App
	if dbCtx.ActiveDomainID != nil {
		apps, err = ds.ListAppsByDomain(*dbCtx.ActiveDomainID)
	} else {
		// If no active domain, list all apps
		apps, err = ds.ListAllApps()
	}
	if err != nil {
		return nil, err
	}

	result := make([]resource.Resource, len(apps))
	for i, a := range apps {
		domainName := ""
		ecosystemName := ""
		if a.DomainID.Valid {
			domain, _ := ds.GetDomainByID(int(a.DomainID.Int64))
			if domain != nil {
				domainName = domain.Name
				if domain.EcosystemID.Valid {
					if eco, ecoErr := ds.GetEcosystemByID(int(domain.EcosystemID.Int64)); ecoErr == nil {
						ecosystemName = eco.Name
					}
				}
			}
		}
		// Resolve GitRepo name if GitRepoID is set
		gitRepoName := ""
		if a.GitRepoID.Valid {
			gitRepo, grErr := ds.GetGitRepoByID(a.GitRepoID.Int64)
			if grErr == nil && gitRepo != nil {
				gitRepoName = gitRepo.Name
			}
		}
		// Resolve System name if SystemID is set
		systemName := ""
		if a.SystemID.Valid {
			system, sErr := ds.GetSystemByID(int(a.SystemID.Int64))
			if sErr == nil && system != nil {
				systemName = system.Name
			}
		}
		result[i] = &AppResource{app: a, domainName: domainName, ecosystemName: ecosystemName, systemName: systemName, gitRepoName: gitRepoName}
	}
	return result, nil
}

// Delete removes an app by name.
func (h *AppHandler) Delete(ctx resource.Context, name string) error {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return err
	}

	// Get active domain from context
	dbCtx, err := ds.GetContext()
	if err != nil {
		return fmt.Errorf("failed to get context: %w", err)
	}

	if dbCtx.ActiveDomainID == nil {
		return fmt.Errorf("no active domain set; use 'dvm use domain <name>' first")
	}

	app, err := ds.GetAppByName(sql.NullInt64{Int64: int64(*dbCtx.ActiveDomainID), Valid: true}, name)
	if err != nil {
		return err
	}

	return ds.DeleteApp(app.ID)
}

// ToYAML serializes an app to YAML.
func (h *AppHandler) ToYAML(res resource.Resource) ([]byte, error) {
	ar, ok := res.(*AppResource)
	if !ok {
		return nil, fmt.Errorf("expected AppResource, got %T", res)
	}

	yamlDoc := ar.app.ToYAML(ar.domainName, nil, ar.gitRepoName, ar.systemName)
	// Include ecosystem name in metadata for context-free round-trip
	if ar.ecosystemName != "" {
		yamlDoc.Metadata.Ecosystem = ar.ecosystemName
	}
	return yaml.Marshal(yamlDoc)
}

// AppResource wraps a models.App to implement resource.Resource.
type AppResource struct {
	app           *models.App
	domainName    string
	ecosystemName string
	systemName    string // Name of the parent System, if any
	gitRepoName   string // Name of the GitRepo, if any
}

func (r *AppResource) GetKind() string {
	return KindApp
}

func (r *AppResource) GetName() string {
	return r.app.Name
}

func (r *AppResource) Validate() error {
	if r.app.Name == "" {
		return fmt.Errorf("app name is required")
	}
	if r.app.Path == "" {
		return fmt.Errorf("app path is required")
	}
	return nil
}

// App returns the underlying models.App.
func (r *AppResource) App() *models.App {
	return r.app
}

// DomainName returns the domain name for this app.
func (r *AppResource) DomainName() string {
	return r.domainName
}

// NewAppResource creates a new AppResource from a model.
// ecosystemName is needed for context-free YAML round-trip. Pass "" if unknown.
// extra[0] = gitRepoName (optional), extra[1] = systemName (optional).
func NewAppResource(app *models.App, domainName, ecosystemName string, extra ...string) *AppResource {
	gitRepoName := ""
	if len(extra) > 0 {
		gitRepoName = extra[0]
	}
	sysName := ""
	if len(extra) > 1 {
		sysName = extra[1]
	}
	return &AppResource{app: app, domainName: domainName, ecosystemName: ecosystemName, gitRepoName: gitRepoName, systemName: sysName}
}

// NewAppFromModel creates an App model from parameters.
func NewAppFromModel(name string, domainID int, path, description string) *models.App {
	return &models.App{
		Name:     name,
		DomainID: sql.NullInt64{Int64: int64(domainID), Valid: domainID != 0},
		Path:     path,
		Description: sql.NullString{
			String: description,
			Valid:  description != "",
		},
	}
}

// Move reparents an app to a new System (and that System's Domain), or to a
// new Domain with no System (issue #397).
//
// target.SystemName: required for system-scoped moves; if empty and
// target.DomainName is set, the app is moved to the Domain with no System.
//
// target.EcosystemName / target.DomainName act as disambiguators when the
// system name is not unique.
//
// Idempotent: if the app is already at the target (System,Domain), returns
// NoOp=true with no DB writes.
func (h *AppHandler) Move(ctx resource.Context, name string, target MoveTarget) (*MoveResult, error) {
	ds, err := dataStoreFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	// Resolve the app being moved (global lookup, ecosystem hint if given).
	matches, err := ds.FindAppsByName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to find app '%s': %w", name, err)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("app '%s' not found", name)
	}
	var src *models.AppWithHierarchy
	if target.EcosystemName != "" {
		for _, m := range matches {
			if m.Ecosystem != nil && m.Ecosystem.Name == target.EcosystemName {
				src = m
				break
			}
		}
		if src == nil {
			return nil, fmt.Errorf("app '%s' not found in ecosystem '%s'", name, target.EcosystemName)
		}
	} else {
		if len(matches) > 1 {
			return nil, fmt.Errorf("app '%s' exists in multiple ecosystems; specify -e <ecosystem> to disambiguate", name)
		}
		src = matches[0]
	}

	// Build "fromParent" string for result.
	fromParent := "(unparented)"
	switch {
	case src.App.SystemID.Valid:
		if sys, e := ds.GetSystemByID(int(src.App.SystemID.Int64)); e == nil && sys != nil {
			fromParent = "system/" + sys.Name
		} else {
			fromParent = "system/(unknown)"
		}
	case src.Domain != nil:
		fromParent = "domain/" + src.Domain.Name
	}

	// Resolve target: prefer System; fall back to Domain-only move.
	var newDomainID, newSystemID sql.NullInt64
	var toParent string

	if target.SystemName != "" {
		targetSystem, targetDomain, err := resolveSystemTarget(ds, target.EcosystemName, target.DomainName, target.SystemName)
		if err != nil {
			return nil, err
		}
		// Cross-ecosystem guard.
		if src.Ecosystem != nil && targetDomain.EcosystemID.Valid &&
			int64(src.Ecosystem.ID) != targetDomain.EcosystemID.Int64 {
			return nil, fmt.Errorf("cannot move app across ecosystems")
		}
		newDomainID = sql.NullInt64{Int64: int64(targetDomain.ID), Valid: true}
		newSystemID = sql.NullInt64{Int64: int64(targetSystem.ID), Valid: true}
		toParent = "system/" + targetSystem.Name
	} else if target.DomainName != "" {
		targetDomain, _, err := resolveDomainTarget(ds, target.EcosystemName, target.DomainName)
		if err != nil {
			return nil, err
		}
		if src.Ecosystem != nil && targetDomain.EcosystemID.Valid &&
			int64(src.Ecosystem.ID) != targetDomain.EcosystemID.Int64 {
			return nil, fmt.Errorf("cannot move app across ecosystems")
		}
		newDomainID = sql.NullInt64{Int64: int64(targetDomain.ID), Valid: true}
		newSystemID = sql.NullInt64{} // no system
		toParent = "domain/" + targetDomain.Name
	} else {
		return nil, fmt.Errorf("move target requires --to-system or --to-domain")
	}

	// Idempotency.
	if src.App.DomainID == newDomainID && src.App.SystemID == newSystemID {
		return &MoveResult{
			Kind: "app", Name: name,
			FromParent: fromParent, ToParent: toParent, NoOp: true,
		}, nil
	}

	if err := ds.MoveApp(src.App.ID, newDomainID, newSystemID); err != nil {
		return nil, fmt.Errorf("failed to move app '%s': %w", name, err)
	}
	return &MoveResult{
		Kind: "app", Name: name,
		FromParent: fromParent, ToParent: toParent,
	}, nil
}

// Detach fully detaches an App from its System AND Domain — the App ends up
// at ecosystem level with both DomainID and SystemID set to NULL (issue #397
// use case 3: "Remove App from System but keep in Ecosystem").
//
// Returns an error if the App is not currently attached to a System.
// (An app already at ecosystem level is not a "detach from system" target.)
func (h *AppHandler) Detach(ctx resource.Context, name string, ecosystemHint string) (*MoveResult, error) {
	ds, err := dataStoreFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	matches, err := ds.FindAppsByName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to find app '%s': %w", name, err)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("app '%s' not found", name)
	}
	var src *models.AppWithHierarchy
	if ecosystemHint != "" {
		for _, m := range matches {
			if m.Ecosystem != nil && m.Ecosystem.Name == ecosystemHint {
				src = m
				break
			}
		}
		if src == nil {
			return nil, fmt.Errorf("app '%s' not found in ecosystem '%s'", name, ecosystemHint)
		}
	} else {
		if len(matches) > 1 {
			return nil, fmt.Errorf("app '%s' exists in multiple ecosystems; specify -e <ecosystem> to disambiguate", name)
		}
		src = matches[0]
	}

	if !src.App.SystemID.Valid {
		return nil, fmt.Errorf("app '%s' is not attached to a system; nothing to detach", name)
	}

	fromParent := "system/(unknown)"
	if sys, e := ds.GetSystemByID(int(src.App.SystemID.Int64)); e == nil && sys != nil {
		fromParent = "system/" + sys.Name
	}

	if err := ds.MoveApp(src.App.ID, sql.NullInt64{}, sql.NullInt64{}); err != nil {
		return nil, fmt.Errorf("failed to detach app '%s': %w", name, err)
	}

	return &MoveResult{
		Kind: "app", Name: name,
		FromParent: fromParent,
		ToParent:   "(ecosystem level — no system, no domain)",
	}, nil
}
