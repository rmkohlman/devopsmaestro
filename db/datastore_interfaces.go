package db

import (
	"devopsmaestro/models"
)

// =============================================================================
// Domain Sub-Interfaces
// =============================================================================
//
// These narrow, domain-scoped interfaces follow the Interface Segregation
// Principle (ISP). Each sub-interface groups methods for a single entity type,
// matching the store_*.go implementation files.
//
// New consumers should depend on the narrowest sub-interface they need rather
// than the full DataStore interface. The composed DataStore interface provides
// backward compatibility for existing code.
//
// Example:
//
//	func listApps(store db.AppStore) ([]*models.App, error) {
//	    return store.ListAllApps()
//	}
//
// =============================================================================

// EcosystemStore defines operations for managing ecosystems (top-level grouping).
type EcosystemStore interface {
	// CreateEcosystem inserts a new ecosystem into the database.
	CreateEcosystem(ecosystem *models.Ecosystem) error

	// GetEcosystemByName retrieves an ecosystem by its name.
	GetEcosystemByName(name string) (*models.Ecosystem, error)

	// GetEcosystemByID retrieves an ecosystem by its ID.
	GetEcosystemByID(id int) (*models.Ecosystem, error)

	// UpdateEcosystem updates an existing ecosystem.
	UpdateEcosystem(ecosystem *models.Ecosystem) error

	// DeleteEcosystem removes an ecosystem by name.
	DeleteEcosystem(name string) error

	// ListEcosystems retrieves all ecosystems.
	ListEcosystems() ([]*models.Ecosystem, error)

	// CountEcosystems returns the total number of ecosystems.
	CountEcosystems() (int, error)
}

// DomainStore defines operations for managing domains (bounded context within an ecosystem).
type DomainStore interface {
	// CreateDomain inserts a new domain into the database.
	CreateDomain(domain *models.Domain) error

	// GetDomainByName retrieves a domain by ecosystem ID and name.
	GetDomainByName(ecosystemID int, name string) (*models.Domain, error)

	// GetDomainByID retrieves a domain by its ID.
	GetDomainByID(id int) (*models.Domain, error)

	// UpdateDomain updates an existing domain.
	UpdateDomain(domain *models.Domain) error

	// DeleteDomain removes a domain by ID.
	DeleteDomain(id int) error

	// ListDomainsByEcosystem retrieves all domains for an ecosystem.
	ListDomainsByEcosystem(ecosystemID int) ([]*models.Domain, error)

	// ListAllDomains retrieves all domains across all ecosystems.
	ListAllDomains() ([]*models.Domain, error)

	// FindDomainsByName retrieves all domains with the given name across all ecosystems,
	// including their parent ecosystem.
	// Returns an empty slice (not an error) if no domains match.
	FindDomainsByName(name string) ([]*models.DomainWithHierarchy, error)
}

// AppStore defines operations for managing apps (codebase/application within a domain).
type AppStore interface {
	// CreateApp inserts a new app into the database.
	CreateApp(app *models.App) error

	// GetAppByName retrieves an app by domain ID and name.
	GetAppByName(domainID int, name string) (*models.App, error)

	// GetAppByNameGlobal retrieves an app by name across all domains.
	// Returns the first match if multiple apps have the same name in different domains.
	// This is useful for CLI convenience when the user doesn't want to specify domain context.
	GetAppByNameGlobal(name string) (*models.App, error)

	// GetAppByID retrieves an app by its ID.
	GetAppByID(id int) (*models.App, error)

	// UpdateApp updates an existing app.
	UpdateApp(app *models.App) error

	// DeleteApp removes an app by ID.
	DeleteApp(id int) error

	// ListAppsByDomain retrieves all apps for a domain.
	ListAppsByDomain(domainID int) ([]*models.App, error)

	// ListAllApps retrieves all apps across all domains.
	ListAllApps() ([]*models.App, error)

	// FindAppsByName retrieves all apps with the given name across all domains,
	// including their full hierarchy (domain and ecosystem).
	// Returns an empty slice (not an error) if no apps match.
	FindAppsByName(name string) ([]*models.AppWithHierarchy, error)
}

// WorkspaceStore defines operations for managing workspaces.
type WorkspaceStore interface {
	// CreateWorkspace inserts a new workspace.
	// Callers must ensure defaults (nvim config, slug) are set via
	// workspace.PrepareDefaults() before calling this method.
	CreateWorkspace(workspace *models.Workspace) error

	// GetWorkspaceByName retrieves a workspace by app ID and name.
	GetWorkspaceByName(appID int, name string) (*models.Workspace, error)

	// GetWorkspaceByID retrieves a workspace by its ID.
	GetWorkspaceByID(id int) (*models.Workspace, error)

	// GetWorkspaceBySlug retrieves a workspace by its hierarchical slug.
	GetWorkspaceBySlug(slug string) (*models.Workspace, error)

	// UpdateWorkspace updates an existing workspace.
	UpdateWorkspace(workspace *models.Workspace) error

	// DeleteWorkspace removes a workspace by ID.
	DeleteWorkspace(id int) error

	// ListWorkspacesByApp retrieves all workspaces for an app.
	ListWorkspacesByApp(appID int) ([]*models.Workspace, error)

	// ListAllWorkspaces retrieves all workspaces across all apps.
	ListAllWorkspaces() ([]*models.Workspace, error)

	// FindWorkspaces searches for workspaces matching the given filter criteria.
	// Returns workspaces with their full hierarchy information (ecosystem, domain, app).
	// Use this for smart workspace resolution when the user provides partial criteria.
	FindWorkspaces(filter models.WorkspaceFilter) ([]*models.WorkspaceWithHierarchy, error)

	// GetWorkspaceSlug returns the slug for a workspace.
	GetWorkspaceSlug(workspaceID int) (string, error)
}

// ContextStore defines operations for active selection state tracking.
// The hierarchy is: Ecosystem -> Domain -> App -> Workspace
type ContextStore interface {
	// GetContext retrieves the current context.
	GetContext() (*models.Context, error)

	// SetActiveEcosystem sets the active ecosystem in the context.
	SetActiveEcosystem(ecosystemID *int) error

	// SetActiveDomain sets the active domain in the context.
	SetActiveDomain(domainID *int) error

	// SetActiveApp sets the active app in the context.
	SetActiveApp(appID *int) error

	// SetActiveWorkspace sets the active workspace in the context.
	SetActiveWorkspace(workspaceID *int) error
}

// PluginStore defines operations for managing nvim plugins and workspace plugin associations.
type PluginStore interface {
	// CreatePlugin inserts a new nvim plugin.
	CreatePlugin(plugin *models.NvimPluginDB) error

	// GetPluginByName retrieves a plugin by its name.
	GetPluginByName(name string) (*models.NvimPluginDB, error)

	// GetPluginByID retrieves a plugin by its ID.
	GetPluginByID(id int) (*models.NvimPluginDB, error)

	// UpdatePlugin updates an existing plugin.
	UpdatePlugin(plugin *models.NvimPluginDB) error

	// UpsertPlugin creates or updates a plugin (by name).
	UpsertPlugin(plugin *models.NvimPluginDB) error

	// DeletePlugin removes a plugin by name.
	DeletePlugin(name string) error

	// ListPlugins retrieves all plugins.
	ListPlugins() ([]*models.NvimPluginDB, error)

	// ListPluginsByCategory retrieves plugins filtered by category.
	ListPluginsByCategory(category string) ([]*models.NvimPluginDB, error)

	// ListPluginsByTags retrieves plugins that have any of the specified tags.
	ListPluginsByTags(tags []string) ([]*models.NvimPluginDB, error)

	// AddPluginToWorkspace associates a plugin with a workspace.
	AddPluginToWorkspace(workspaceID int, pluginID int) error

	// RemovePluginFromWorkspace removes a plugin association from a workspace.
	RemovePluginFromWorkspace(workspaceID int, pluginID int) error

	// GetWorkspacePlugins retrieves all plugins associated with a workspace.
	GetWorkspacePlugins(workspaceID int) ([]*models.NvimPluginDB, error)

	// SetWorkspacePluginEnabled enables or disables a plugin for a workspace.
	SetWorkspacePluginEnabled(workspaceID int, pluginID int, enabled bool) error
}

// ThemeStore defines operations for managing nvim themes.
type ThemeStore interface {
	// CreateTheme inserts a new nvim theme.
	CreateTheme(theme *models.NvimThemeDB) error

	// GetThemeByName retrieves a theme by its name.
	GetThemeByName(name string) (*models.NvimThemeDB, error)

	// GetThemeByID retrieves a theme by its ID.
	GetThemeByID(id int) (*models.NvimThemeDB, error)

	// UpdateTheme updates an existing theme.
	UpdateTheme(theme *models.NvimThemeDB) error

	// DeleteTheme removes a theme by name.
	DeleteTheme(name string) error

	// ListThemes retrieves all themes.
	ListThemes() ([]*models.NvimThemeDB, error)

	// ListThemesByCategory retrieves themes filtered by category.
	ListThemesByCategory(category string) ([]*models.NvimThemeDB, error)

	// GetActiveTheme retrieves the currently active theme.
	GetActiveTheme() (*models.NvimThemeDB, error)

	// SetActiveTheme sets the active theme by name (deactivates others).
	SetActiveTheme(name string) error

	// ClearActiveTheme deactivates all themes.
	ClearActiveTheme() error
}

// TerminalPromptStore defines operations for managing terminal prompts.
type TerminalPromptStore interface {
	// CreateTerminalPrompt inserts a new terminal prompt.
	CreateTerminalPrompt(prompt *models.TerminalPromptDB) error

	// GetTerminalPromptByName retrieves a terminal prompt by its name.
	GetTerminalPromptByName(name string) (*models.TerminalPromptDB, error)

	// UpdateTerminalPrompt updates an existing terminal prompt.
	UpdateTerminalPrompt(prompt *models.TerminalPromptDB) error

	// UpsertTerminalPrompt creates or updates a terminal prompt (by name).
	UpsertTerminalPrompt(prompt *models.TerminalPromptDB) error

	// DeleteTerminalPrompt removes a terminal prompt by name.
	DeleteTerminalPrompt(name string) error

	// ListTerminalPrompts retrieves all terminal prompts.
	ListTerminalPrompts() ([]*models.TerminalPromptDB, error)

	// ListTerminalPromptsByType retrieves terminal prompts filtered by type.
	ListTerminalPromptsByType(promptType string) ([]*models.TerminalPromptDB, error)

	// ListTerminalPromptsByCategory retrieves terminal prompts filtered by category.
	ListTerminalPromptsByCategory(category string) ([]*models.TerminalPromptDB, error)
}

// TerminalProfileStore defines operations for managing terminal profiles.
type TerminalProfileStore interface {
	// CreateTerminalProfile inserts a new terminal profile.
	CreateTerminalProfile(profile *models.TerminalProfileDB) error

	// GetTerminalProfileByName retrieves a terminal profile by its name.
	GetTerminalProfileByName(name string) (*models.TerminalProfileDB, error)

	// UpdateTerminalProfile updates an existing terminal profile.
	UpdateTerminalProfile(profile *models.TerminalProfileDB) error

	// UpsertTerminalProfile creates or updates a terminal profile (by name).
	UpsertTerminalProfile(profile *models.TerminalProfileDB) error

	// DeleteTerminalProfile removes a terminal profile by name.
	DeleteTerminalProfile(name string) error

	// ListTerminalProfiles retrieves all terminal profiles.
	ListTerminalProfiles() ([]*models.TerminalProfileDB, error)

	// ListTerminalProfilesByCategory retrieves terminal profiles filtered by category.
	ListTerminalProfilesByCategory(category string) ([]*models.TerminalProfileDB, error)
}

// TerminalPluginStore defines operations for managing terminal plugins.
type TerminalPluginStore interface {
	// CreateTerminalPlugin inserts a new terminal plugin.
	CreateTerminalPlugin(plugin *models.TerminalPluginDB) error

	// UpdateTerminalPlugin updates an existing terminal plugin.
	UpdateTerminalPlugin(plugin *models.TerminalPluginDB) error

	// UpsertTerminalPlugin creates or updates a terminal plugin (by name).
	UpsertTerminalPlugin(plugin *models.TerminalPluginDB) error

	// DeleteTerminalPlugin removes a terminal plugin by name.
	DeleteTerminalPlugin(name string) error

	// GetTerminalPlugin retrieves a terminal plugin by its name.
	GetTerminalPlugin(name string) (*models.TerminalPluginDB, error)

	// ListTerminalPlugins retrieves all terminal plugins.
	ListTerminalPlugins() ([]*models.TerminalPluginDB, error)

	// ListTerminalPluginsByCategory retrieves terminal plugins filtered by category.
	ListTerminalPluginsByCategory(category string) ([]*models.TerminalPluginDB, error)

	// ListTerminalPluginsByShell retrieves terminal plugins filtered by shell.
	ListTerminalPluginsByShell(shell string) ([]*models.TerminalPluginDB, error)

	// ListTerminalPluginsByManager retrieves terminal plugins filtered by manager.
	ListTerminalPluginsByManager(manager string) ([]*models.TerminalPluginDB, error)
}

// TerminalEmulatorStore defines operations for managing terminal emulator configs.
type TerminalEmulatorStore interface {
	// CreateTerminalEmulator inserts a new terminal emulator config.
	CreateTerminalEmulator(emulator *models.TerminalEmulatorDB) error

	// UpdateTerminalEmulator updates an existing terminal emulator config.
	UpdateTerminalEmulator(emulator *models.TerminalEmulatorDB) error

	// UpsertTerminalEmulator creates or updates a terminal emulator (by name).
	UpsertTerminalEmulator(emulator *models.TerminalEmulatorDB) error

	// DeleteTerminalEmulator removes a terminal emulator by name.
	DeleteTerminalEmulator(name string) error

	// GetTerminalEmulator retrieves a terminal emulator by its name.
	GetTerminalEmulator(name string) (*models.TerminalEmulatorDB, error)

	// ListTerminalEmulators retrieves all terminal emulators.
	ListTerminalEmulators() ([]*models.TerminalEmulatorDB, error)

	// ListTerminalEmulatorsByType retrieves terminal emulators filtered by type.
	ListTerminalEmulatorsByType(emulatorType string) ([]*models.TerminalEmulatorDB, error)

	// ListTerminalEmulatorsByWorkspace retrieves terminal emulators for a workspace.
	ListTerminalEmulatorsByWorkspace(workspace string) ([]*models.TerminalEmulatorDB, error)
}

// CredentialStore defines operations for managing credentials.
type CredentialStore interface {
	// CreateCredential inserts a new credential configuration.
	CreateCredential(credential *models.CredentialDB) error

	// GetCredential retrieves a credential by scope and name.
	GetCredential(scopeType models.CredentialScopeType, scopeID int64, name string) (*models.CredentialDB, error)

	// GetCredentialByName retrieves a credential by name across all scopes.
	// Returns the first match if multiple credentials have the same name in different scopes.
	// This is useful for CLI convenience (e.g., --credential flag on gitrepo create).
	GetCredentialByName(name string) (*models.CredentialDB, error)

	// UpdateCredential updates an existing credential.
	UpdateCredential(credential *models.CredentialDB) error

	// DeleteCredential removes a credential by scope and name.
	DeleteCredential(scopeType models.CredentialScopeType, scopeID int64, name string) error

	// ListCredentialsByScope retrieves all credentials for a specific scope.
	ListCredentialsByScope(scopeType models.CredentialScopeType, scopeID int64) ([]*models.CredentialDB, error)

	// ListAllCredentials retrieves all credentials across all scopes.
	ListAllCredentials() ([]*models.CredentialDB, error)
}

// GitRepoStore defines operations for managing git repository configurations.
type GitRepoStore interface {
	// CreateGitRepo inserts a new git repository configuration.
	CreateGitRepo(repo *models.GitRepoDB) error

	// GetGitRepoByName retrieves a git repository by its name.
	GetGitRepoByName(name string) (*models.GitRepoDB, error)

	// GetGitRepoByID retrieves a git repository by its ID.
	GetGitRepoByID(id int64) (*models.GitRepoDB, error)

	// GetGitRepoBySlug retrieves a git repository by its slug.
	GetGitRepoBySlug(slug string) (*models.GitRepoDB, error)

	// UpdateGitRepo updates an existing git repository configuration.
	UpdateGitRepo(repo *models.GitRepoDB) error

	// DeleteGitRepo removes a git repository by name.
	DeleteGitRepo(name string) error

	// ListGitRepos retrieves all git repositories.
	ListGitRepos() ([]models.GitRepoDB, error)
}

// DefaultsStore defines operations for managing default configuration values.
type DefaultsStore interface {
	// GetDefault retrieves a default value by key.
	// Returns empty string if key is not found (not an error).
	GetDefault(key string) (string, error)

	// SetDefault sets a default value for the given key.
	// Uses upsert behavior (INSERT OR REPLACE).
	SetDefault(key, value string) error

	// DeleteDefault removes a default value by key.
	// No error if key doesn't exist.
	DeleteDefault(key string) error

	// ListDefaults retrieves all default values as a key-value map.
	ListDefaults() (map[string]string, error)
}

// NvimPackageStore defines operations for managing nvim packages.
type NvimPackageStore interface {
	// CreatePackage inserts a new nvim package.
	CreatePackage(pkg *models.NvimPackageDB) error

	// UpdatePackage updates an existing nvim package.
	UpdatePackage(pkg *models.NvimPackageDB) error

	// UpsertPackage creates or updates an nvim package (by name).
	UpsertPackage(pkg *models.NvimPackageDB) error

	// DeletePackage removes a package by name.
	DeletePackage(name string) error

	// GetPackage retrieves a package by its name.
	GetPackage(name string) (*models.NvimPackageDB, error)

	// ListPackages retrieves all packages.
	ListPackages() ([]*models.NvimPackageDB, error)

	// ListPackagesByLabel retrieves packages that have a specific label key-value pair.
	ListPackagesByLabel(key, value string) ([]*models.NvimPackageDB, error)
}

// TerminalPackageStore defines operations for managing terminal packages.
type TerminalPackageStore interface {
	// CreateTerminalPackage inserts a new terminal package.
	CreateTerminalPackage(pkg *models.TerminalPackageDB) error

	// UpdateTerminalPackage updates an existing terminal package.
	UpdateTerminalPackage(pkg *models.TerminalPackageDB) error

	// UpsertTerminalPackage creates or updates a terminal package (by name).
	UpsertTerminalPackage(pkg *models.TerminalPackageDB) error

	// DeleteTerminalPackage removes a terminal package by name.
	DeleteTerminalPackage(name string) error

	// GetTerminalPackage retrieves a terminal package by its name.
	GetTerminalPackage(name string) (*models.TerminalPackageDB, error)

	// ListTerminalPackages retrieves all terminal packages.
	ListTerminalPackages() ([]*models.TerminalPackageDB, error)

	// ListTerminalPackagesByLabel retrieves terminal packages that have a specific label key-value pair.
	ListTerminalPackagesByLabel(key, value string) ([]*models.TerminalPackageDB, error)
}

// RegistryStore defines operations for managing package registries
// (zot, athens, devpi, verdaccio, squid).
type RegistryStore interface {
	// CreateRegistry inserts a new registry.
	CreateRegistry(registry *models.Registry) error

	// GetRegistryByName retrieves a registry by its name.
	GetRegistryByName(name string) (*models.Registry, error)

	// GetRegistryByID retrieves a registry by its ID.
	GetRegistryByID(id int) (*models.Registry, error)

	// GetRegistryByPort retrieves a registry by its port (for conflict detection).
	GetRegistryByPort(port int) (*models.Registry, error)

	// UpdateRegistry updates an existing registry.
	UpdateRegistry(registry *models.Registry) error

	// DeleteRegistry removes a registry by name.
	DeleteRegistry(name string) error

	// ListRegistries retrieves all registries.
	ListRegistries() ([]*models.Registry, error)

	// ListRegistriesByType retrieves registries filtered by type.
	ListRegistriesByType(registryType string) ([]*models.Registry, error)

	// ListRegistriesByStatus retrieves registries filtered by status.
	ListRegistriesByStatus(status string) ([]*models.Registry, error)
}

// RegistryHistoryStore defines operations for managing registry history entries.
type RegistryHistoryStore interface {
	// CreateRegistryHistory inserts a new registry history entry.
	CreateRegistryHistory(history *models.RegistryHistory) error

	// GetRegistryHistory retrieves a specific registry history entry by registryID and revision.
	GetRegistryHistory(registryID int, revision int) (*models.RegistryHistory, error)

	// GetLatestRegistryHistory retrieves the most recent history entry for a registry.
	GetLatestRegistryHistory(registryID int) (*models.RegistryHistory, error)

	// ListRegistryHistory retrieves all history entries for a registry, ordered by revision DESC.
	ListRegistryHistory(registryID int) ([]*models.RegistryHistory, error)

	// GetNextRevisionNumber returns the next available revision number for a registry.
	// Returns 1 if no history exists yet.
	GetNextRevisionNumber(registryID int) (int, error)
}

// CustomResourceStore defines operations for managing Custom Resource Definitions (CRDs)
// and their instances.
type CustomResourceStore interface {
	// CreateCRD inserts a new custom resource definition.
	CreateCRD(crd *models.CustomResourceDefinition) error

	// GetCRDByKind retrieves a CRD by its kind name.
	GetCRDByKind(kind string) (*models.CustomResourceDefinition, error)

	// UpdateCRD updates an existing CRD.
	UpdateCRD(crd *models.CustomResourceDefinition) error

	// DeleteCRD removes a CRD by kind.
	DeleteCRD(kind string) error

	// ListCRDs retrieves all CRDs.
	ListCRDs() ([]*models.CustomResourceDefinition, error)

	// CreateCustomResource inserts a new custom resource instance.
	CreateCustomResource(resource *models.CustomResource) error

	// GetCustomResource retrieves a custom resource by kind, name, and namespace.
	GetCustomResource(kind, name, namespace string) (*models.CustomResource, error)

	// UpdateCustomResource updates an existing custom resource.
	UpdateCustomResource(resource *models.CustomResource) error

	// DeleteCustomResource removes a custom resource by kind, name, and namespace.
	DeleteCustomResource(kind, name, namespace string) error

	// ListCustomResources retrieves all custom resources of a given kind.
	ListCustomResources(kind string) ([]*models.CustomResource, error)
}
