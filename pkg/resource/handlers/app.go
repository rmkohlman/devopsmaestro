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
		system, err := ds.GetSystemByName(sql.NullInt64{Int64: int64(domain.ID), Valid: true}, appYAML.Metadata.System)
		if err != nil {
			return nil, fmt.Errorf("system '%s' not found in domain '%s': %w", appYAML.Metadata.System, domainName, err)
		}
		app.SystemID = sql.NullInt64{Int64: int64(system.ID), Valid: true}
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
// gitRepoName is optional — pass as extra[0] if available.
func NewAppResource(app *models.App, domainName, ecosystemName string, extra ...string) *AppResource {
	gitRepoName := ""
	if len(extra) > 0 {
		gitRepoName = extra[0]
	}
	return &AppResource{app: app, domainName: domainName, ecosystemName: ecosystemName, gitRepoName: gitRepoName}
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
