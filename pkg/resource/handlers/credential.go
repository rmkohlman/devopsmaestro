package handlers

import (
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/resource"

	"gopkg.in/yaml.v3"
)

const KindCredential = "Credential"

// CredentialHandler handles Credential resources.
type CredentialHandler struct{}

// NewCredentialHandler creates a new Credential handler.
func NewCredentialHandler() *CredentialHandler {
	return &CredentialHandler{}
}

func (h *CredentialHandler) Kind() string {
	return KindCredential
}

// Apply creates or updates a credential from YAML data.
func (h *CredentialHandler) Apply(ctx resource.Context, data []byte) (resource.Resource, error) {
	// Parse the YAML
	var credYAML models.CredentialYAML
	if err := yaml.Unmarshal(data, &credYAML); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate the YAML
	if err := models.ValidateCredentialYAML(credYAML); err != nil {
		return nil, err
	}

	// Get the full DataStore (needs scope resolution + credential CRUD)
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return nil, err
	}

	// Extract scope type and scope name from metadata
	scopeType, scopeName := credYAML.Metadata.ScopeInfo()
	if scopeType == "" {
		return nil, fmt.Errorf("exactly one scope (ecosystem, domain, app, or workspace) must be specified")
	}

	// Resolve scope name → scope ID
	scopeID, err := resolveCredentialScope(ds, scopeType, scopeName)
	if err != nil {
		return nil, err
	}

	// Convert YAML to CredentialDB model
	cred := &models.CredentialDB{}
	cred.FromYAML(credYAML)
	cred.ScopeType = scopeType
	cred.ScopeID = scopeID

	// Upsert: check if credential already exists
	existing, err := ds.GetCredential(scopeType, scopeID, cred.Name)
	if err == nil && existing != nil {
		// Update existing
		cred.ID = existing.ID
		if err := ds.UpdateCredential(cred); err != nil {
			return nil, fmt.Errorf("failed to update credential: %w", err)
		}
	} else {
		// Create new
		if err := ds.CreateCredential(cred); err != nil {
			return nil, fmt.Errorf("failed to create credential: %w", err)
		}
	}

	return &CredentialResource{credential: cred, scopeName: scopeName}, nil
}

// resolveCredentialScope resolves a scope name to its numeric ID.
// For domain/app/workspace scopes, it first attempts to use the active context,
// falling back to a global search if no context is set.
func resolveCredentialScope(ds db.DataStore, scopeType models.CredentialScopeType, scopeName string) (int64, error) {
	switch scopeType {
	case models.CredentialScopeEcosystem:
		eco, err := ds.GetEcosystemByName(scopeName)
		if err != nil {
			return 0, fmt.Errorf("ecosystem '%s' not found: %w", scopeName, err)
		}
		return int64(eco.ID), nil

	case models.CredentialScopeDomain:
		// Try active context first
		dbCtx, err := ds.GetContext()
		if err == nil && dbCtx.ActiveEcosystemID != nil {
			domain, err := ds.GetDomainByName(*dbCtx.ActiveEcosystemID, scopeName)
			if err == nil {
				return int64(domain.ID), nil
			}
		}
		// Fall back to global search
		domains, err := ds.ListAllDomains()
		if err != nil {
			return 0, fmt.Errorf("failed to list domains: %w", err)
		}
		for _, d := range domains {
			if d.Name == scopeName {
				return int64(d.ID), nil
			}
		}
		return 0, fmt.Errorf("domain '%s' not found", scopeName)

	case models.CredentialScopeApp:
		// Try active context first
		dbCtx, err := ds.GetContext()
		if err == nil && dbCtx.ActiveDomainID != nil {
			app, err := ds.GetAppByName(*dbCtx.ActiveDomainID, scopeName)
			if err == nil {
				return int64(app.ID), nil
			}
		}
		// Fall back to global search
		app, err := ds.GetAppByNameGlobal(scopeName)
		if err != nil {
			return 0, fmt.Errorf("app '%s' not found: %w", scopeName, err)
		}
		return int64(app.ID), nil

	case models.CredentialScopeWorkspace:
		// Try active context first
		dbCtx, err := ds.GetContext()
		if err == nil && dbCtx.ActiveAppID != nil {
			ws, err := ds.GetWorkspaceByName(*dbCtx.ActiveAppID, scopeName)
			if err == nil {
				return int64(ws.ID), nil
			}
		}
		// Fall back to global search
		workspaces, err := ds.ListAllWorkspaces()
		if err != nil {
			return 0, fmt.Errorf("failed to list workspaces: %w", err)
		}
		for _, ws := range workspaces {
			if ws.Name == scopeName {
				return int64(ws.ID), nil
			}
		}
		return 0, fmt.Errorf("workspace '%s' not found", scopeName)

	default:
		return 0, fmt.Errorf("unknown credential scope type: %s", scopeType)
	}
}

// findCredentialByContext searches for a credential by name using the active
// context cascade (workspace → app → domain → ecosystem). It returns the
// credential, the reverse-resolved scope name (for ToYAML), and any error.
func findCredentialByContext(ds db.DataStore, name string) (*models.CredentialDB, string, error) {
	dbCtx, err := ds.GetContext()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get context: %w", err)
	}

	// Waterfall: most specific scope first
	type scopeCheck struct {
		activeID  *int
		scopeType models.CredentialScopeType
	}
	checks := []scopeCheck{
		{dbCtx.ActiveWorkspaceID, models.CredentialScopeWorkspace},
		{dbCtx.ActiveAppID, models.CredentialScopeApp},
		{dbCtx.ActiveDomainID, models.CredentialScopeDomain},
		{dbCtx.ActiveEcosystemID, models.CredentialScopeEcosystem},
	}

	for _, sc := range checks {
		if sc.activeID == nil {
			continue
		}
		cred, err := ds.GetCredential(sc.scopeType, int64(*sc.activeID), name)
		if err != nil {
			continue
		}
		// Reverse-resolve the scopeID to a human-readable scope name
		scopeName, err := resolveCredentialScopeName(ds, cred.ScopeType, cred.ScopeID)
		if err != nil {
			return nil, "", fmt.Errorf("failed to resolve scope name: %w", err)
		}
		return cred, scopeName, nil
	}

	return nil, "", fmt.Errorf("credential '%s' not found in any active scope", name)
}

// resolveCredentialScopeName reverse-resolves a scope ID back to its human-readable name.
func resolveCredentialScopeName(ds db.DataStore, scopeType models.CredentialScopeType, scopeID int64) (string, error) {
	switch scopeType {
	case models.CredentialScopeEcosystem:
		eco, err := ds.GetEcosystemByID(int(scopeID))
		if err != nil {
			return "", fmt.Errorf("ecosystem ID %d not found: %w", scopeID, err)
		}
		return eco.Name, nil
	case models.CredentialScopeDomain:
		domain, err := ds.GetDomainByID(int(scopeID))
		if err != nil {
			return "", fmt.Errorf("domain ID %d not found: %w", scopeID, err)
		}
		return domain.Name, nil
	case models.CredentialScopeApp:
		app, err := ds.GetAppByID(int(scopeID))
		if err != nil {
			return "", fmt.Errorf("app ID %d not found: %w", scopeID, err)
		}
		return app.Name, nil
	case models.CredentialScopeWorkspace:
		ws, err := ds.GetWorkspaceByID(int(scopeID))
		if err != nil {
			return "", fmt.Errorf("workspace ID %d not found: %w", scopeID, err)
		}
		return ws.Name, nil
	default:
		return "", fmt.Errorf("unknown credential scope type: %s", scopeType)
	}
}

// Get retrieves a credential by name.
// Uses active context to determine scope, searching from most specific
// (workspace) to least specific (ecosystem).
func (h *CredentialHandler) Get(ctx resource.Context, name string) (resource.Resource, error) {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return nil, err
	}

	cred, scopeName, err := findCredentialByContext(ds, name)
	if err != nil {
		return nil, err
	}

	return &CredentialResource{credential: cred, scopeName: scopeName}, nil
}

// List retrieves all credentials across all scopes.
func (h *CredentialHandler) List(ctx resource.Context) ([]resource.Resource, error) {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return nil, err
	}

	creds, err := ds.ListAllCredentials()
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, len(creds))
	for i, cred := range creds {
		resources[i] = &CredentialResource{credential: cred}
	}
	return resources, nil
}

// Delete removes a credential by name.
// Uses active context to determine scope, searching from most specific
// (workspace) to least specific (ecosystem).
func (h *CredentialHandler) Delete(ctx resource.Context, name string) error {
	ds, err := resource.DataStoreAs[db.DataStore](ctx)
	if err != nil {
		return err
	}

	cred, _, err := findCredentialByContext(ds, name)
	if err != nil {
		return err
	}

	return ds.DeleteCredential(cred.ScopeType, cred.ScopeID, name)
}

// ToYAML converts a credential resource to YAML.
func (h *CredentialHandler) ToYAML(res resource.Resource) ([]byte, error) {
	credRes, ok := res.(*CredentialResource)
	if !ok {
		return nil, fmt.Errorf("resource is not a Credential")
	}

	credYAML := credRes.credential.ToYAML(credRes.scopeName)
	return yaml.Marshal(credYAML)
}

// CredentialResource wraps a models.CredentialDB as a resource.Resource.
type CredentialResource struct {
	credential *models.CredentialDB
	scopeName  string
}

func (r *CredentialResource) GetKind() string {
	return KindCredential
}

func (r *CredentialResource) GetName() string {
	return r.credential.Name
}

func (r *CredentialResource) Validate() error {
	y := r.credential.ToYAML(r.scopeName)
	return models.ValidateCredentialYAML(y)
}

// Credential returns the underlying CredentialDB model.
func (r *CredentialResource) Credential() *models.CredentialDB {
	return r.credential
}

// ScopeName returns the resolved scope name for this credential.
func (r *CredentialResource) ScopeName() string {
	return r.scopeName
}

// NewCredentialResource creates a new CredentialResource from a model.
func NewCredentialResource(cred *models.CredentialDB, scopeName string) *CredentialResource {
	return &CredentialResource{credential: cred, scopeName: scopeName}
}
