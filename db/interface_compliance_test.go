package db

import (
	"database/sql"
	"testing"

	"devopsmaestro/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Compile-Time Interface Compliance Tests
// =============================================================================

// TestDataStoreInterfaceCompliance verifies that all implementations
// satisfy the DataStore interface at compile time.
func TestDataStoreInterfaceCompliance(t *testing.T) {
	// Compile-time checks — composed DataStore interface
	var _ DataStore = (*SQLDataStore)(nil)
	var _ DataStore = (*MockDataStore)(nil)
}

// TestSubInterfaceCompliance_SQLDataStore verifies that SQLDataStore satisfies
// each of the 19 domain sub-interfaces individually.
func TestSubInterfaceCompliance_SQLDataStore(t *testing.T) {
	var _ EcosystemStore = (*SQLDataStore)(nil)
	var _ DomainStore = (*SQLDataStore)(nil)
	var _ AppStore = (*SQLDataStore)(nil)
	var _ WorkspaceStore = (*SQLDataStore)(nil)
	var _ ContextStore = (*SQLDataStore)(nil)
	var _ PluginStore = (*SQLDataStore)(nil)
	var _ ThemeStore = (*SQLDataStore)(nil)
	var _ TerminalPromptStore = (*SQLDataStore)(nil)
	var _ TerminalProfileStore = (*SQLDataStore)(nil)
	var _ TerminalPluginStore = (*SQLDataStore)(nil)
	var _ TerminalEmulatorStore = (*SQLDataStore)(nil)
	var _ CredentialStore = (*SQLDataStore)(nil)
	var _ GitRepoStore = (*SQLDataStore)(nil)
	var _ DefaultsStore = (*SQLDataStore)(nil)
	var _ NvimPackageStore = (*SQLDataStore)(nil)
	var _ TerminalPackageStore = (*SQLDataStore)(nil)
	var _ RegistryStore = (*SQLDataStore)(nil)
	var _ RegistryHistoryStore = (*SQLDataStore)(nil)
	var _ CustomResourceStore = (*SQLDataStore)(nil)
}

// TestSubInterfaceCompliance_MockDataStore verifies that MockDataStore satisfies
// each of the 19 domain sub-interfaces individually.
func TestSubInterfaceCompliance_MockDataStore(t *testing.T) {
	var _ EcosystemStore = (*MockDataStore)(nil)
	var _ DomainStore = (*MockDataStore)(nil)
	var _ AppStore = (*MockDataStore)(nil)
	var _ WorkspaceStore = (*MockDataStore)(nil)
	var _ ContextStore = (*MockDataStore)(nil)
	var _ PluginStore = (*MockDataStore)(nil)
	var _ ThemeStore = (*MockDataStore)(nil)
	var _ TerminalPromptStore = (*MockDataStore)(nil)
	var _ TerminalProfileStore = (*MockDataStore)(nil)
	var _ TerminalPluginStore = (*MockDataStore)(nil)
	var _ TerminalEmulatorStore = (*MockDataStore)(nil)
	var _ CredentialStore = (*MockDataStore)(nil)
	var _ GitRepoStore = (*MockDataStore)(nil)
	var _ DefaultsStore = (*MockDataStore)(nil)
	var _ NvimPackageStore = (*MockDataStore)(nil)
	var _ TerminalPackageStore = (*MockDataStore)(nil)
	var _ RegistryStore = (*MockDataStore)(nil)
	var _ RegistryHistoryStore = (*MockDataStore)(nil)
	var _ CustomResourceStore = (*MockDataStore)(nil)
}

// TestDriverInterfaceCompliance verifies that all driver implementations
// satisfy the Driver interface at compile time.
func TestDriverInterfaceCompliance(t *testing.T) {
	// Compile-time checks
	var _ Driver = (*SQLiteDriver)(nil)
	var _ Driver = (*MockDriver)(nil)
}

// =============================================================================
// Runtime Interface Compliance Tests for MockDataStore
// =============================================================================

// TestMockDataStore_AllEcosystemMethods verifies MockDataStore implements
// all Ecosystem CRUD operations correctly.
func TestMockDataStore_AllEcosystemMethods(t *testing.T) {
	store := NewMockDataStore()

	// CreateEcosystem
	ecosystem := &models.Ecosystem{
		Name:        "test-eco",
		Description: sql.NullString{String: "Test ecosystem", Valid: true},
	}
	err := store.CreateEcosystem(ecosystem)
	require.NoError(t, err)
	assert.NotZero(t, ecosystem.ID, "ID should be assigned")

	// GetEcosystemByName
	retrieved, err := store.GetEcosystemByName("test-eco")
	require.NoError(t, err)
	assert.Equal(t, "test-eco", retrieved.Name)

	// GetEcosystemByID
	retrieved, err = store.GetEcosystemByID(ecosystem.ID)
	require.NoError(t, err)
	assert.Equal(t, ecosystem.ID, retrieved.ID)

	// UpdateEcosystem
	ecosystem.Description = sql.NullString{String: "Updated description", Valid: true}
	err = store.UpdateEcosystem(ecosystem)
	require.NoError(t, err)

	// ListEcosystems
	ecosystems, err := store.ListEcosystems()
	require.NoError(t, err)
	assert.Len(t, ecosystems, 1)

	// DeleteEcosystem
	err = store.DeleteEcosystem("test-eco")
	require.NoError(t, err)

	// Verify deletion
	_, err = store.GetEcosystemByName("test-eco")
	assert.Error(t, err)
}

// TestMockDataStore_AllDomainMethods verifies MockDataStore implements
// all Domain CRUD operations correctly.
func TestMockDataStore_AllDomainMethods(t *testing.T) {
	store := NewMockDataStore()

	// Setup: Create ecosystem first
	ecosystem := &models.Ecosystem{Name: "test-eco"}
	store.CreateEcosystem(ecosystem)

	// CreateDomain
	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "test-domain",
		Description: sql.NullString{String: "Test domain", Valid: true},
	}
	err := store.CreateDomain(domain)
	require.NoError(t, err)
	assert.NotZero(t, domain.ID)

	// GetDomainByName
	retrieved, err := store.GetDomainByName(ecosystem.ID, "test-domain")
	require.NoError(t, err)
	assert.Equal(t, "test-domain", retrieved.Name)

	// GetDomainByID
	retrieved, err = store.GetDomainByID(domain.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ID, retrieved.ID)

	// UpdateDomain
	domain.Description = sql.NullString{String: "Updated description", Valid: true}
	err = store.UpdateDomain(domain)
	require.NoError(t, err)

	// ListDomainsByEcosystem
	domains, err := store.ListDomainsByEcosystem(ecosystem.ID)
	require.NoError(t, err)
	assert.Len(t, domains, 1)

	// ListAllDomains
	allDomains, err := store.ListAllDomains()
	require.NoError(t, err)
	assert.Len(t, allDomains, 1)

	// DeleteDomain
	err = store.DeleteDomain(domain.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = store.GetDomainByID(domain.ID)
	assert.Error(t, err)
}

// TestMockDataStore_AllAppMethods verifies MockDataStore implements
// all App CRUD operations correctly.
func TestMockDataStore_AllAppMethods(t *testing.T) {
	store := NewMockDataStore()

	// Setup: Create ecosystem and domain
	ecosystem := &models.Ecosystem{Name: "test-eco"}
	store.CreateEcosystem(ecosystem)
	domain := &models.Domain{EcosystemID: ecosystem.ID, Name: "test-domain"}
	store.CreateDomain(domain)

	// CreateApp
	app := &models.App{
		DomainID:    domain.ID,
		Name:        "test-app",
		Description: sql.NullString{String: "Test application", Valid: true},
		Path:        "/test/app",
	}
	err := store.CreateApp(app)
	require.NoError(t, err)
	assert.NotZero(t, app.ID)

	// GetAppByName
	retrieved, err := store.GetAppByName(domain.ID, "test-app")
	require.NoError(t, err)
	assert.Equal(t, "test-app", retrieved.Name)

	// GetAppByNameGlobal
	retrieved, err = store.GetAppByNameGlobal("test-app")
	require.NoError(t, err)
	assert.Equal(t, "test-app", retrieved.Name)

	// GetAppByID
	retrieved, err = store.GetAppByID(app.ID)
	require.NoError(t, err)
	assert.Equal(t, app.ID, retrieved.ID)

	// UpdateApp
	app.Description = sql.NullString{String: "Updated description", Valid: true}
	err = store.UpdateApp(app)
	require.NoError(t, err)

	// ListAppsByDomain
	apps, err := store.ListAppsByDomain(domain.ID)
	require.NoError(t, err)
	assert.Len(t, apps, 1)

	// ListAllApps
	allApps, err := store.ListAllApps()
	require.NoError(t, err)
	assert.Len(t, allApps, 1)

	// DeleteApp
	err = store.DeleteApp(app.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = store.GetAppByID(app.ID)
	assert.Error(t, err)
}

// TestMockDataStore_AllWorkspaceMethods verifies MockDataStore implements
// all Workspace CRUD operations correctly.
func TestMockDataStore_AllWorkspaceMethods(t *testing.T) {
	store := NewMockDataStore()

	// Setup: Create full hierarchy
	ecosystem := &models.Ecosystem{Name: "test-eco"}
	store.CreateEcosystem(ecosystem)
	domain := &models.Domain{EcosystemID: ecosystem.ID, Name: "test-domain"}
	store.CreateDomain(domain)
	app := &models.App{DomainID: domain.ID, Name: "test-app", Path: "/test"}
	store.CreateApp(app)

	// CreateWorkspace
	workspace := &models.Workspace{
		AppID:  app.ID,
		Name:   "main",
		Slug:   "test-eco-test-domain-test-app-main",
		Status: "stopped",
	}
	err := store.CreateWorkspace(workspace)
	require.NoError(t, err)
	assert.NotZero(t, workspace.ID)

	// GetWorkspaceByName
	retrieved, err := store.GetWorkspaceByName(app.ID, "main")
	require.NoError(t, err)
	assert.Equal(t, "main", retrieved.Name)

	// GetWorkspaceByID
	retrieved, err = store.GetWorkspaceByID(workspace.ID)
	require.NoError(t, err)
	assert.Equal(t, workspace.ID, retrieved.ID)

	// GetWorkspaceBySlug
	retrieved, err = store.GetWorkspaceBySlug("test-eco-test-domain-test-app-main")
	require.NoError(t, err)
	assert.Equal(t, workspace.ID, retrieved.ID)

	// UpdateWorkspace
	workspace.Status = "running"
	err = store.UpdateWorkspace(workspace)
	require.NoError(t, err)

	// ListWorkspacesByApp
	workspaces, err := store.ListWorkspacesByApp(app.ID)
	require.NoError(t, err)
	assert.Len(t, workspaces, 1)

	// ListAllWorkspaces
	allWorkspaces, err := store.ListAllWorkspaces()
	require.NoError(t, err)
	assert.Len(t, allWorkspaces, 1)

	// FindWorkspaces
	filter := models.WorkspaceFilter{
		EcosystemName: "test-eco",
		WorkspaceName: "main",
	}
	found, err := store.FindWorkspaces(filter)
	require.NoError(t, err)
	assert.Len(t, found, 1)

	// GetWorkspaceSlug
	slug, err := store.GetWorkspaceSlug(workspace.ID)
	require.NoError(t, err)
	assert.Equal(t, workspace.Slug, slug)

	// DeleteWorkspace
	err = store.DeleteWorkspace(workspace.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = store.GetWorkspaceByID(workspace.ID)
	assert.Error(t, err)
}

// TestMockDataStore_AllContextMethods verifies MockDataStore implements
// all Context operations correctly.
func TestMockDataStore_AllContextMethods(t *testing.T) {
	store := NewMockDataStore()

	// GetContext
	ctx, err := store.GetContext()
	require.NoError(t, err)
	assert.Equal(t, 1, ctx.ID)

	// SetActiveEcosystem
	ecosystemID := 42
	err = store.SetActiveEcosystem(&ecosystemID)
	require.NoError(t, err)

	ctx, _ = store.GetContext()
	require.NotNil(t, ctx.ActiveEcosystemID)
	assert.Equal(t, 42, *ctx.ActiveEcosystemID)

	// SetActiveDomain
	domainID := 7
	err = store.SetActiveDomain(&domainID)
	require.NoError(t, err)

	// SetActiveApp
	appID := 99
	err = store.SetActiveApp(&appID)
	require.NoError(t, err)

	// SetActiveWorkspace
	workspaceID := 123
	err = store.SetActiveWorkspace(&workspaceID)
	require.NoError(t, err)

	// Verify all are set
	ctx, _ = store.GetContext()
	require.NotNil(t, ctx.ActiveWorkspaceID)
	assert.Equal(t, 123, *ctx.ActiveWorkspaceID)
}

// TestMockDataStore_AllPluginMethods verifies MockDataStore implements
// all Plugin CRUD operations correctly.
func TestMockDataStore_AllPluginMethods(t *testing.T) {
	store := NewMockDataStore()

	// CreatePlugin
	plugin := &models.NvimPluginDB{
		Name:        "telescope.nvim",
		Description: sql.NullString{String: "Fuzzy finder", Valid: true},
		Repo:        "nvim-telescope/telescope.nvim",
		Lazy:        false,
	}
	err := store.CreatePlugin(plugin)
	require.NoError(t, err)
	assert.NotZero(t, plugin.ID)

	// GetPluginByName
	retrieved, err := store.GetPluginByName("telescope.nvim")
	require.NoError(t, err)
	assert.Equal(t, "telescope.nvim", retrieved.Name)

	// GetPluginByID
	retrieved, err = store.GetPluginByID(plugin.ID)
	require.NoError(t, err)
	assert.Equal(t, plugin.ID, retrieved.ID)

	// UpdatePlugin
	plugin.Description = sql.NullString{String: "Updated description", Valid: true}
	err = store.UpdatePlugin(plugin)
	require.NoError(t, err)

	// UpsertPlugin (update case)
	plugin.Description = sql.NullString{String: "Upserted description", Valid: true}
	err = store.UpsertPlugin(plugin)
	require.NoError(t, err)

	// UpsertPlugin (insert case)
	newPlugin := &models.NvimPluginDB{Name: "new-plugin.nvim", Repo: "user/new-plugin"}
	err = store.UpsertPlugin(newPlugin)
	require.NoError(t, err)
	assert.NotZero(t, newPlugin.ID)

	// ListPlugins
	plugins, err := store.ListPlugins()
	require.NoError(t, err)
	assert.Len(t, plugins, 2)

	// DeletePlugin
	err = store.DeletePlugin("telescope.nvim")
	require.NoError(t, err)

	// Verify deletion
	_, err = store.GetPluginByName("telescope.nvim")
	assert.Error(t, err)
}

// TestMockDataStore_AllThemeMethods verifies MockDataStore implements
// all Theme CRUD operations correctly.
func TestMockDataStore_AllThemeMethods(t *testing.T) {
	store := NewMockDataStore()

	// CreateTheme
	theme := &models.NvimThemeDB{
		Name:        "catppuccin",
		Description: sql.NullString{String: "Soothing pastel theme", Valid: true},
		PluginRepo:  "catppuccin/nvim",
		IsActive:    false,
	}
	err := store.CreateTheme(theme)
	require.NoError(t, err)
	assert.NotZero(t, theme.ID)

	// GetThemeByName
	retrieved, err := store.GetThemeByName("catppuccin")
	require.NoError(t, err)
	assert.Equal(t, "catppuccin", retrieved.Name)

	// GetThemeByID
	retrieved, err = store.GetThemeByID(theme.ID)
	require.NoError(t, err)
	assert.Equal(t, theme.ID, retrieved.ID)

	// UpdateTheme
	theme.Description = sql.NullString{String: "Updated description", Valid: true}
	err = store.UpdateTheme(theme)
	require.NoError(t, err)

	// ListThemes
	themes, err := store.ListThemes()
	require.NoError(t, err)
	assert.Len(t, themes, 1)

	// SetActiveTheme
	err = store.SetActiveTheme("catppuccin")
	require.NoError(t, err)

	// GetActiveTheme
	activeTheme, err := store.GetActiveTheme()
	require.NoError(t, err)
	assert.Equal(t, "catppuccin", activeTheme.Name)

	// ClearActiveTheme
	err = store.ClearActiveTheme()
	require.NoError(t, err)

	activeTheme, err = store.GetActiveTheme()
	assert.Nil(t, activeTheme)

	// DeleteTheme
	err = store.DeleteTheme("catppuccin")
	require.NoError(t, err)
}

// TestMockDataStore_AllCredentialMethods verifies MockDataStore implements
// all Credential CRUD operations correctly.
func TestMockDataStore_AllCredentialMethods(t *testing.T) {
	store := NewMockDataStore()

	// Setup: Create ecosystem for scoping
	ecosystem := &models.Ecosystem{Name: "test-eco"}
	store.CreateEcosystem(ecosystem)

	// CreateCredential
	service := "test-service"
	cred := &models.CredentialDB{
		Name:        "test-cred",
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     int64(ecosystem.ID),
		Source:      "vault",
		VaultSecret: &service,
	}
	err := store.CreateCredential(cred)
	require.NoError(t, err)

	// GetCredential
	retrieved, err := store.GetCredential(models.CredentialScopeEcosystem, int64(ecosystem.ID), "test-cred")
	require.NoError(t, err)
	assert.Equal(t, "test-cred", retrieved.Name)

	// UpdateCredential
	newService := "updated-service"
	cred.VaultSecret = &newService
	err = store.UpdateCredential(cred)
	require.NoError(t, err)

	// ListCredentialsByScope
	creds, err := store.ListCredentialsByScope(models.CredentialScopeEcosystem, int64(ecosystem.ID))
	require.NoError(t, err)
	assert.Len(t, creds, 1)

	// ListAllCredentials
	allCreds, err := store.ListAllCredentials()
	require.NoError(t, err)
	assert.Len(t, allCreds, 1)

	// DeleteCredential
	err = store.DeleteCredential(models.CredentialScopeEcosystem, int64(ecosystem.ID), "test-cred")
	require.NoError(t, err)

	// Verify deletion
	_, err = store.GetCredential(models.CredentialScopeEcosystem, int64(ecosystem.ID), "test-cred")
	assert.Error(t, err)
}

// TestMockDataStore_AllDefaultMethods verifies MockDataStore implements
// all Default operations correctly.
func TestMockDataStore_AllDefaultMethods(t *testing.T) {
	store := NewMockDataStore()

	// SetDefault
	err := store.SetDefault("test-key", "test-value")
	require.NoError(t, err)

	// GetDefault
	value, err := store.GetDefault("test-key")
	require.NoError(t, err)
	assert.Equal(t, "test-value", value)

	// ListDefaults
	defaults, err := store.ListDefaults()
	require.NoError(t, err)
	assert.Len(t, defaults, 1)
	assert.Equal(t, "test-value", defaults["test-key"])

	// DeleteDefault
	err = store.DeleteDefault("test-key")
	require.NoError(t, err)

	// Verify deletion
	value, err = store.GetDefault("test-key")
	require.NoError(t, err)
	assert.Empty(t, value)
}

// TestMockDataStore_AllGitRepoMethods verifies MockDataStore implements
// all GitRepo CRUD operations correctly.
func TestMockDataStore_AllGitRepoMethods(t *testing.T) {
	store := NewMockDataStore()

	// CreateGitRepo
	repo := &models.GitRepoDB{
		Name: "test-repo",
		Slug: "test-slug",
		URL:  "https://github.com/user/repo.git",
	}
	err := store.CreateGitRepo(repo)
	require.NoError(t, err)
	assert.NotZero(t, repo.ID)

	// GetGitRepoByName
	retrieved, err := store.GetGitRepoByName("test-repo")
	require.NoError(t, err)
	assert.Equal(t, "test-repo", retrieved.Name)

	// GetGitRepoByID
	retrieved, err = store.GetGitRepoByID(int64(repo.ID))
	require.NoError(t, err)
	assert.Equal(t, repo.ID, retrieved.ID)

	// GetGitRepoBySlug
	retrieved, err = store.GetGitRepoBySlug("test-slug")
	require.NoError(t, err)
	assert.Equal(t, "test-slug", retrieved.Slug)

	// UpdateGitRepo
	repo.URL = "https://github.com/user/updated-repo.git"
	err = store.UpdateGitRepo(repo)
	require.NoError(t, err)

	// ListGitRepos
	repos, err := store.ListGitRepos()
	require.NoError(t, err)
	assert.Len(t, repos, 1)

	// DeleteGitRepo
	err = store.DeleteGitRepo("test-repo")
	require.NoError(t, err)

	// Verify deletion
	_, err = store.GetGitRepoByName("test-repo")
	assert.Error(t, err)
}

// TestMockDataStore_ErrorInjectionForAllMethods verifies that error
// injection works for all methods.
func TestMockDataStore_ErrorInjectionForAllMethods(t *testing.T) {
	store := NewMockDataStore()
	expectedErr := assert.AnError

	tests := []struct {
		name   string
		setErr func()
		testFn func() error
	}{
		{"CreateEcosystem", func() { store.CreateEcosystemErr = expectedErr }, func() error {
			return store.CreateEcosystem(&models.Ecosystem{})
		}},
		{"GetEcosystemByName", func() { store.GetEcosystemByNameErr = expectedErr }, func() error {
			_, err := store.GetEcosystemByName("test")
			return err
		}},
		{"CreateWorkspace", func() { store.CreateWorkspaceErr = expectedErr }, func() error {
			return store.CreateWorkspace(&models.Workspace{})
		}},
		{"CreatePlugin", func() { store.CreatePluginErr = expectedErr }, func() error {
			return store.CreatePlugin(&models.NvimPluginDB{})
		}},
		{"Close", func() { store.CloseErr = expectedErr }, func() error {
			return store.Close()
		}},
		{"Ping", func() { store.PingErr = expectedErr }, func() error {
			return store.Ping()
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset before each test
			store.Reset()

			// Set error
			tt.setErr()

			// Test function should return expected error
			err := tt.testFn()
			assert.Equal(t, expectedErr, err)
		})
	}
}
