package handlers

import (
	"database/sql"
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/resource"

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
	ds, err := h.getDataStore(ctx)
	if err != nil {
		return nil, err
	}

	// Resolve domain from YAML metadata
	domainName := appYAML.Metadata.Domain
	if domainName == "" {
		return nil, fmt.Errorf("app YAML must specify metadata.domain")
	}

	// Get active ecosystem from context to resolve domain
	dbCtx, err := ds.GetContext()
	if err != nil {
		return nil, fmt.Errorf("failed to get context: %w", err)
	}

	if dbCtx.ActiveEcosystemID == nil {
		return nil, fmt.Errorf("no active ecosystem set; use 'dvm use ecosystem <name>' first")
	}

	domain, err := ds.GetDomainByName(*dbCtx.ActiveEcosystemID, domainName)
	if err != nil {
		return nil, fmt.Errorf("domain '%s' not found: %w", domainName, err)
	}

	// Convert to model
	app := &models.App{
		DomainID: domain.ID,
	}
	app.FromYAML(appYAML)

	// Check if app exists
	existing, _ := ds.GetAppByName(domain.ID, app.Name)
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
		app, err = ds.GetAppByName(domain.ID, app.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve created app: %w", err)
		}
	}

	return &AppResource{app: app, domainName: domainName}, nil
}

// Get retrieves an app by name.
// Note: This requires an active domain context to resolve the app.
func (h *AppHandler) Get(ctx resource.Context, name string) (resource.Resource, error) {
	ds, err := h.getDataStore(ctx)
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

	app, err := ds.GetAppByName(*dbCtx.ActiveDomainID, name)
	if err != nil {
		return nil, err
	}

	domain, _ := ds.GetDomainByID(app.DomainID)
	domainName := ""
	if domain != nil {
		domainName = domain.Name
	}

	return &AppResource{app: app, domainName: domainName}, nil
}

// List returns all apps in the active domain.
func (h *AppHandler) List(ctx resource.Context) ([]resource.Resource, error) {
	ds, err := h.getDataStore(ctx)
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
		domain, _ := ds.GetDomainByID(a.DomainID)
		domainName := ""
		if domain != nil {
			domainName = domain.Name
		}
		result[i] = &AppResource{app: a, domainName: domainName}
	}
	return result, nil
}

// Delete removes an app by name.
func (h *AppHandler) Delete(ctx resource.Context, name string) error {
	ds, err := h.getDataStore(ctx)
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

	app, err := ds.GetAppByName(*dbCtx.ActiveDomainID, name)
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

	yamlDoc := ar.app.ToYAML(ar.domainName)
	return yaml.Marshal(yamlDoc)
}

// getDataStore returns the DataStore from the context.
func (h *AppHandler) getDataStore(ctx resource.Context) (db.DataStore, error) {
	if ctx.DataStore == nil {
		return nil, fmt.Errorf("DataStore not provided in context")
	}

	ds, ok := ctx.DataStore.(db.DataStore)
	if !ok {
		return nil, fmt.Errorf("invalid DataStore type: %T", ctx.DataStore)
	}

	return ds, nil
}

// AppResource wraps a models.App to implement resource.Resource.
type AppResource struct {
	app        *models.App
	domainName string
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
	if r.app.DomainID == 0 {
		return fmt.Errorf("app domain_id is required")
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
func NewAppResource(app *models.App, domainName string) *AppResource {
	return &AppResource{app: app, domainName: domainName}
}

// NewAppFromModel creates an App model from parameters.
func NewAppFromModel(name string, domainID int, path, description string) *models.App {
	return &models.App{
		Name:     name,
		DomainID: domainID,
		Path:     path,
		Description: sql.NullString{
			String: description,
			Valid:  description != "",
		},
	}
}
