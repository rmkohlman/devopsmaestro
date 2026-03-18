package handlers

import (
	"database/sql"
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/resource"
	ws "devopsmaestro/pkg/workspace"

	"gopkg.in/yaml.v3"
)

const KindWorkspace = "Workspace"

// WorkspaceHandler handles Workspace resources.
type WorkspaceHandler struct{}

// NewWorkspaceHandler creates a new Workspace handler.
func NewWorkspaceHandler() *WorkspaceHandler {
	return &WorkspaceHandler{}
}

func (h *WorkspaceHandler) Kind() string {
	return KindWorkspace
}

// Apply creates or updates a workspace from YAML data.
func (h *WorkspaceHandler) Apply(ctx resource.Context, data []byte) (resource.Resource, error) {
	// Parse the YAML
	var wsYAML models.WorkspaceYAML
	if err := yaml.Unmarshal(data, &wsYAML); err != nil {
		return nil, fmt.Errorf("failed to parse workspace YAML: %w", err)
	}

	// Get the datastore
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return nil, err
	}

	// Resolve app from YAML metadata
	appName := wsYAML.Metadata.App
	if appName == "" {
		return nil, fmt.Errorf("workspace YAML must specify metadata.app")
	}

	// Resolve domain: try metadata.domain first, then fall back to active context
	var domainID int
	if wsYAML.Metadata.Domain != "" {
		// Resolve domain by name across all ecosystems
		allDomains, err := ds.ListAllDomains()
		if err != nil {
			return nil, fmt.Errorf("failed to list domains: %w", err)
		}
		var found *models.Domain
		for _, d := range allDomains {
			if d.Name == wsYAML.Metadata.Domain {
				found = d
				break
			}
		}
		if found == nil {
			return nil, fmt.Errorf("domain '%s' not found", wsYAML.Metadata.Domain)
		}
		domainID = found.ID
	} else {
		// Fall back to active context (existing behavior)
		dbCtx, err := ds.GetContext()
		if err != nil {
			return nil, fmt.Errorf("failed to get context: %w", err)
		}
		if dbCtx.ActiveDomainID == nil {
			return nil, fmt.Errorf("no active domain set and no metadata.domain specified; use 'dvm use domain <name>' or add metadata.domain to YAML")
		}
		domainID = *dbCtx.ActiveDomainID
	}

	app, err := ds.GetAppByName(domainID, appName)
	if err != nil {
		return nil, fmt.Errorf("app '%s' not found: %w", appName, err)
	}

	// Convert to model
	workspace := &models.Workspace{
		AppID: app.ID,
	}
	workspace.FromYAML(wsYAML)

	// Resolve GitRepo if specified in YAML
	if wsYAML.Spec.GitRepo != "" {
		gitRepo, err := ds.GetGitRepoByName(wsYAML.Spec.GitRepo)
		if err != nil {
			return nil, fmt.Errorf("gitrepo '%s' not found: %w", wsYAML.Spec.GitRepo, err)
		}
		workspace.GitRepoID = sql.NullInt64{Int64: int64(gitRepo.ID), Valid: true}
	}

	// Check if workspace exists
	existing, _ := ds.GetWorkspaceByName(app.ID, workspace.Name)
	if existing != nil {
		// Update existing
		workspace.ID = existing.ID
		workspace.Slug = existing.Slug               // Preserve slug (has UNIQUE constraint)
		workspace.ContainerID = existing.ContainerID // Preserve container ID
		// Don't overwrite these fields if they weren't provided in YAML
		if !workspace.NvimStructure.Valid {
			workspace.NvimStructure = existing.NvimStructure
		}
		if !workspace.NvimPlugins.Valid {
			workspace.NvimPlugins = existing.NvimPlugins
		}
		if !workspace.TerminalPrompt.Valid {
			workspace.TerminalPrompt = existing.TerminalPrompt
		}
		if !workspace.TerminalPlugins.Valid {
			workspace.TerminalPlugins = existing.TerminalPlugins
		}
		if !workspace.TerminalPackage.Valid {
			workspace.TerminalPackage = existing.TerminalPackage
		}
		if !workspace.GitRepoID.Valid {
			workspace.GitRepoID = existing.GitRepoID
		}
		if err := ds.UpdateWorkspace(workspace); err != nil {
			return nil, fmt.Errorf("failed to update workspace: %w", err)
		}
	} else {
		// Create new — apply defaults before persisting
		if err := ws.PrepareDefaults(workspace, ds); err != nil {
			return nil, fmt.Errorf("failed to prepare workspace defaults: %w", err)
		}
		if err := ds.CreateWorkspace(workspace); err != nil {
			return nil, fmt.Errorf("failed to create workspace: %w", err)
		}
		// Fetch to get the ID
		workspace, err = ds.GetWorkspaceByName(app.ID, workspace.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve created workspace: %w", err)
		}
	}

	// Resolve domain name for the resource output
	domainName := wsYAML.Metadata.Domain
	if domainName == "" {
		// Look up domain name from ID for round-trip fidelity
		domain, domErr := ds.GetDomainByID(domainID)
		if domErr == nil {
			domainName = domain.Name
		}
	}

	return &WorkspaceResource{
		workspace:   workspace,
		appName:     appName,
		domainName:  domainName,
		gitRepoName: wsYAML.Spec.GitRepo, // Store gitrepo name from YAML
	}, nil
}

// Get retrieves a workspace by name.
// Note: This requires an active app context to resolve the workspace.
func (h *WorkspaceHandler) Get(ctx resource.Context, name string) (resource.Resource, error) {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return nil, err
	}

	// Get active app from context
	dbCtx, err := ds.GetContext()
	if err != nil {
		return nil, fmt.Errorf("failed to get context: %w", err)
	}

	if dbCtx.ActiveAppID == nil {
		return nil, fmt.Errorf("no active app set; use 'dvm use app <name>' first")
	}

	workspace, err := ds.GetWorkspaceByName(*dbCtx.ActiveAppID, name)
	if err != nil {
		return nil, err
	}

	app, _ := ds.GetAppByID(workspace.AppID)
	appName := ""
	domainName := ""
	if app != nil {
		appName = app.Name
		// Resolve domain name for round-trip fidelity
		domain, domErr := ds.GetDomainByID(app.DomainID)
		if domErr == nil {
			domainName = domain.Name
		}
	}

	// Resolve GitRepo name if GitRepoID is set
	gitRepoName := ""
	if workspace.GitRepoID.Valid {
		gitRepo, err := ds.GetGitRepoByID(workspace.GitRepoID.Int64)
		if err == nil && gitRepo != nil {
			gitRepoName = gitRepo.Name
		}
	}

	return &WorkspaceResource{
		workspace:   workspace,
		appName:     appName,
		domainName:  domainName,
		gitRepoName: gitRepoName,
	}, nil
}

// List returns all workspaces in the active app.
func (h *WorkspaceHandler) List(ctx resource.Context) ([]resource.Resource, error) {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return nil, err
	}

	// Get active app from context
	dbCtx, err := ds.GetContext()
	if err != nil {
		return nil, fmt.Errorf("failed to get context: %w", err)
	}

	var workspaces []*models.Workspace
	if dbCtx.ActiveAppID != nil {
		workspaces, err = ds.ListWorkspacesByApp(*dbCtx.ActiveAppID)
	} else {
		// If no active app, list all workspaces
		workspaces, err = ds.ListAllWorkspaces()
	}
	if err != nil {
		return nil, err
	}

	result := make([]resource.Resource, len(workspaces))
	for i, ws := range workspaces {
		app, _ := ds.GetAppByID(ws.AppID)
		appName := ""
		domainName := ""
		if app != nil {
			appName = app.Name
			// Resolve domain name for round-trip fidelity
			domain, domErr := ds.GetDomainByID(app.DomainID)
			if domErr == nil {
				domainName = domain.Name
			}
		}

		// Resolve GitRepo name if GitRepoID is set
		gitRepoName := ""
		if ws.GitRepoID.Valid {
			gitRepo, err := ds.GetGitRepoByID(ws.GitRepoID.Int64)
			if err == nil && gitRepo != nil {
				gitRepoName = gitRepo.Name
			}
		}

		result[i] = &WorkspaceResource{
			workspace:   ws,
			appName:     appName,
			domainName:  domainName,
			gitRepoName: gitRepoName,
		}
	}
	return result, nil
}

// Delete removes a workspace by name.
func (h *WorkspaceHandler) Delete(ctx resource.Context, name string) error {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return err
	}

	// Get active app from context
	dbCtx, err := ds.GetContext()
	if err != nil {
		return fmt.Errorf("failed to get context: %w", err)
	}

	if dbCtx.ActiveAppID == nil {
		return fmt.Errorf("no active app set; use 'dvm use app <name>' first")
	}

	workspace, err := ds.GetWorkspaceByName(*dbCtx.ActiveAppID, name)
	if err != nil {
		return err
	}

	return ds.DeleteWorkspace(workspace.ID)
}

// ToYAML serializes a workspace to YAML.
func (h *WorkspaceHandler) ToYAML(res resource.Resource) ([]byte, error) {
	wr, ok := res.(*WorkspaceResource)
	if !ok {
		return nil, fmt.Errorf("expected WorkspaceResource, got %T", res)
	}

	yamlDoc := wr.workspace.ToYAML(wr.appName, wr.gitRepoName)
	// Include domain name in metadata for context-free round-trip
	if wr.domainName != "" {
		yamlDoc.Metadata.Domain = wr.domainName
	}
	return yaml.Marshal(yamlDoc)
}

// WorkspaceResource wraps a models.Workspace to implement resource.Resource.
type WorkspaceResource struct {
	workspace   *models.Workspace
	appName     string
	domainName  string // Domain name for YAML output (context-free apply)
	gitRepoName string // Name of the GitRepo, if any
}

func (r *WorkspaceResource) GetKind() string {
	return KindWorkspace
}

func (r *WorkspaceResource) GetName() string {
	return r.workspace.Name
}

func (r *WorkspaceResource) Validate() error {
	if r.workspace.Name == "" {
		return fmt.Errorf("workspace name is required")
	}
	if r.workspace.AppID == 0 {
		return fmt.Errorf("workspace app_id is required")
	}
	if r.workspace.ImageName == "" {
		return fmt.Errorf("workspace image_name is required")
	}
	return nil
}

// Workspace returns the underlying models.Workspace.
func (r *WorkspaceResource) Workspace() *models.Workspace {
	return r.workspace
}

// AppName returns the app name for this workspace.
func (r *WorkspaceResource) AppName() string {
	return r.appName
}

// NewWorkspaceResource creates a new WorkspaceResource from a model.
func NewWorkspaceResource(workspace *models.Workspace, appName string) *WorkspaceResource {
	return &WorkspaceResource{
		workspace:   workspace,
		appName:     appName,
		gitRepoName: "", // GitRepo name will be resolved when needed
	}
}

// NewWorkspaceFromModel creates a Workspace model from parameters.
func NewWorkspaceFromModel(name string, appID int, imageName, description, status string) *models.Workspace {
	if status == "" {
		status = "stopped"
	}
	return &models.Workspace{
		Name:      name,
		AppID:     appID,
		ImageName: imageName,
		Description: sql.NullString{
			String: description,
			Valid:  description != "",
		},
		Status: status,
	}
}
