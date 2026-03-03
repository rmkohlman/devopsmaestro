package cmd

import (
	"context"
	"database/sql"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/nvimops/theme"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==============================================================================
// TDD Phase 2 - RED: Failing Tests for Build Theme Resolution
// ==============================================================================
// These tests expose the bug where `dvm build` uses a global theme file
// instead of resolving theme from the workspace hierarchy.
//
// GitHub Issue: Theme not resolved from hierarchy during build
//
// Expected behavior:
// 1. resolveWorkspaceTheme() should use HierarchyThemeResolver
// 2. generateNvimConfig() should use hierarchy-resolved theme, not global
// 3. generateShellConfig() should use hierarchy-resolved theme for palette
// 4. prepareStagingDirectory() should accept workspace parameter
// ==============================================================================

// MockDataStoreForBuild is a minimal mock for build tests
type MockDataStoreForBuild struct {
	db.DataStore
	ecosystems map[int]*models.Ecosystem
	domains    map[int]*models.Domain
	apps       map[int]*models.App
	workspaces map[int]*models.Workspace
	defaults   map[string]string
}

func NewMockDataStoreForBuild() *MockDataStoreForBuild {
	return &MockDataStoreForBuild{
		ecosystems: make(map[int]*models.Ecosystem),
		domains:    make(map[int]*models.Domain),
		apps:       make(map[int]*models.App),
		workspaces: make(map[int]*models.Workspace),
		defaults:   make(map[string]string),
	}
}

func (m *MockDataStoreForBuild) GetEcosystemByID(id int) (*models.Ecosystem, error) {
	if e, ok := m.ecosystems[id]; ok {
		return e, nil
	}
	return nil, sql.ErrNoRows
}

func (m *MockDataStoreForBuild) GetDomainByID(id int) (*models.Domain, error) {
	if d, ok := m.domains[id]; ok {
		return d, nil
	}
	return nil, sql.ErrNoRows
}

func (m *MockDataStoreForBuild) GetAppByID(id int) (*models.App, error) {
	if a, ok := m.apps[id]; ok {
		return a, nil
	}
	return nil, sql.ErrNoRows
}

func (m *MockDataStoreForBuild) GetWorkspaceByID(id int) (*models.Workspace, error) {
	if w, ok := m.workspaces[id]; ok {
		return w, nil
	}
	return nil, sql.ErrNoRows
}

func (m *MockDataStoreForBuild) GetDefault(key string) (string, error) {
	if val, ok := m.defaults[key]; ok {
		return val, nil
	}
	return "", sql.ErrNoRows
}

// MockThemeStoreForBuild is a minimal mock for theme storage
type MockThemeStoreForBuild struct {
	themes map[string]*theme.Theme
}

func NewMockThemeStoreForBuild() *MockThemeStoreForBuild {
	return &MockThemeStoreForBuild{
		themes: make(map[string]*theme.Theme),
	}
}

func (m *MockThemeStoreForBuild) Get(name string) (*theme.Theme, error) {
	if t, ok := m.themes[name]; ok {
		return t, nil
	}
	return nil, sql.ErrNoRows
}

func (m *MockThemeStoreForBuild) List() ([]*theme.Theme, error)    { return nil, nil }
func (m *MockThemeStoreForBuild) Save(*theme.Theme) error          { return nil }
func (m *MockThemeStoreForBuild) Delete(string) error              { return nil }
func (m *MockThemeStoreForBuild) GetActive() (*theme.Theme, error) { return nil, nil }
func (m *MockThemeStoreForBuild) SetActive(name string) error      { return nil }
func (m *MockThemeStoreForBuild) Path() string                     { return "" }

// ==============================================================================
// Test 1: resolveWorkspaceTheme() helper function
// ==============================================================================
// This function should use HierarchyThemeResolver to resolve themes from the
// workspace hierarchy (workspace → app → domain → ecosystem → global).

func TestResolveWorkspaceTheme_WorkspaceLevel(t *testing.T) {
	// Setup: Workspace has theme, domain has different theme
	ds := NewMockDataStoreForBuild()
	themeStore := NewMockThemeStoreForBuild()

	// Create hierarchy
	ds.ecosystems[1] = &models.Ecosystem{
		ID:   1,
		Name: "sandbox",
		Theme: sql.NullString{
			String: "ecosystem-theme",
			Valid:  true,
		},
	}
	ds.domains[1] = &models.Domain{
		ID:          1,
		EcosystemID: 1,
		Name:        "backend",
		Theme: sql.NullString{
			String: "domain-theme",
			Valid:  true,
		},
	}
	ds.apps[1] = &models.App{
		ID:       1,
		DomainID: 1,
		Name:     "api-service",
		Theme: sql.NullString{
			Valid: false, // No app theme
		},
	}
	ds.workspaces[1] = &models.Workspace{
		ID:    1,
		AppID: 1,
		Name:  "dev",
		Theme: sql.NullString{
			String: "workspace-theme",
			Valid:  true,
		},
	}

	// Add themes to store
	themeStore.themes["workspace-theme"] = &theme.Theme{
		Name:     "workspace-theme",
		Category: "dark",
	}
	themeStore.themes["domain-theme"] = &theme.Theme{
		Name:     "domain-theme",
		Category: "dark",
	}
	themeStore.themes["ecosystem-theme"] = &theme.Theme{
		Name:     "ecosystem-theme",
		Category: "dark",
	}

	ctx := context.Background()
	workspace := ds.workspaces[1]

	// THIS FUNCTION DOESN'T EXIST YET - Test should fail
	// Expected signature: resolveWorkspaceTheme(ctx context.Context, ds db.DataStore, themeStore theme.Store, workspace *models.Workspace) (*theme.Theme, error)
	resolvedTheme, err := resolveWorkspaceTheme(ctx, ds, themeStore, workspace)

	require.NoError(t, err)
	require.NotNil(t, resolvedTheme)
	assert.Equal(t, "workspace-theme", resolvedTheme.Name,
		"Should resolve workspace-level theme when set")
}

func TestResolveWorkspaceTheme_DomainLevelInheritance(t *testing.T) {
	// Setup: Workspace has NO theme, domain has theme
	ds := NewMockDataStoreForBuild()
	themeStore := NewMockThemeStoreForBuild()

	// Create hierarchy
	ds.ecosystems[1] = &models.Ecosystem{
		ID:   1,
		Name: "sandbox",
	}
	ds.domains[1] = &models.Domain{
		ID:          1,
		EcosystemID: 1,
		Name:        "backend",
		Theme: sql.NullString{
			String: "domain-theme",
			Valid:  true,
		},
	}
	ds.apps[1] = &models.App{
		ID:       1,
		DomainID: 1,
		Name:     "api-service",
	}
	ds.workspaces[1] = &models.Workspace{
		ID:    1,
		AppID: 1,
		Name:  "dev",
		Theme: sql.NullString{
			Valid: false, // No workspace theme
		},
	}

	// Add theme to store
	themeStore.themes["domain-theme"] = &theme.Theme{
		Name:     "domain-theme",
		Category: "dark",
	}

	ctx := context.Background()
	workspace := ds.workspaces[1]

	// Should inherit from domain
	resolvedTheme, err := resolveWorkspaceTheme(ctx, ds, themeStore, workspace)

	require.NoError(t, err)
	require.NotNil(t, resolvedTheme)
	assert.Equal(t, "domain-theme", resolvedTheme.Name,
		"Should inherit theme from domain when workspace has none")
}

func TestResolveWorkspaceTheme_GlobalDefault(t *testing.T) {
	// Setup: No themes in hierarchy, falls back to global default
	ds := NewMockDataStoreForBuild()
	themeStore := NewMockThemeStoreForBuild()

	// Create hierarchy with NO themes set
	ds.ecosystems[1] = &models.Ecosystem{
		ID:   1,
		Name: "sandbox",
	}
	ds.domains[1] = &models.Domain{
		ID:          1,
		EcosystemID: 1,
		Name:        "backend",
	}
	ds.apps[1] = &models.App{
		ID:       1,
		DomainID: 1,
		Name:     "api-service",
	}
	ds.workspaces[1] = &models.Workspace{
		ID:    1,
		AppID: 1,
		Name:  "dev",
	}

	// Set global default in database
	ds.defaults["theme"] = "global-default"
	themeStore.themes["global-default"] = &theme.Theme{
		Name:     "global-default",
		Category: "dark",
	}

	ctx := context.Background()
	workspace := ds.workspaces[1]

	// Should fall back to global default
	resolvedTheme, err := resolveWorkspaceTheme(ctx, ds, themeStore, workspace)

	require.NoError(t, err)
	require.NotNil(t, resolvedTheme)
	assert.Equal(t, "global-default", resolvedTheme.Name,
		"Should use global default when no theme in hierarchy")
}

func TestResolveWorkspaceTheme_ErrorHandling(t *testing.T) {
	// Setup: Theme name exists but fails to load
	ds := NewMockDataStoreForBuild()
	themeStore := NewMockThemeStoreForBuild()

	// Create hierarchy
	ds.ecosystems[1] = &models.Ecosystem{
		ID:   1,
		Name: "sandbox",
	}
	ds.domains[1] = &models.Domain{
		ID:          1,
		EcosystemID: 1,
		Name:        "backend",
		Theme: sql.NullString{
			String: "nonexistent-theme", // Theme doesn't exist in store
			Valid:  true,
		},
	}
	ds.apps[1] = &models.App{
		ID:       1,
		DomainID: 1,
		Name:     "api-service",
	}
	ds.workspaces[1] = &models.Workspace{
		ID:    1,
		AppID: 1,
		Name:  "dev",
	}

	// Set a fallback default
	ds.defaults["theme"] = "fallback-default"
	themeStore.themes["fallback-default"] = &theme.Theme{
		Name:     "fallback-default",
		Category: "dark",
	}

	ctx := context.Background()
	workspace := ds.workspaces[1]

	// Should gracefully fall back to default when theme loading fails
	resolvedTheme, err := resolveWorkspaceTheme(ctx, ds, themeStore, workspace)

	require.NoError(t, err, "Should not error, but fall back gracefully")
	require.NotNil(t, resolvedTheme)
	assert.Equal(t, "fallback-default", resolvedTheme.Name,
		"Should fall back to default when configured theme fails to load")
}

func TestResolveWorkspaceTheme_NilWorkspace(t *testing.T) {
	// Setup: Handle nil workspace gracefully
	ds := NewMockDataStoreForBuild()
	themeStore := NewMockThemeStoreForBuild()

	// Set global default (must match resolver.DefaultTheme constant = "coolnight-ocean")
	themeStore.themes["coolnight-ocean"] = &theme.Theme{
		Name:     "coolnight-ocean",
		Category: "dark",
	}

	ctx := context.Background()

	// Should handle nil workspace and return hardcoded default (resolver.DefaultTheme)
	resolvedTheme, err := resolveWorkspaceTheme(ctx, ds, themeStore, nil)

	require.NoError(t, err)
	require.NotNil(t, resolvedTheme)
	assert.Equal(t, "coolnight-ocean", resolvedTheme.Name,
		"Should return resolver.DefaultTheme (coolnight-ocean) when workspace is nil")
}

// ==============================================================================
// Test 2: Integration with prepareStagingDirectory
// ==============================================================================
// prepareStagingDirectory should accept workspace parameter and pass it to
// generateShellConfig for theme resolution.

func TestPrepareStagingDirectory_AcceptsWorkspace(t *testing.T) {
	t.Skip("Integration test - requires file system operations and database mock")

	// This test verifies the signature change:
	// OLD: prepareStagingDirectory(stagingDir, appPath, appName, workspaceName string, ds db.DataStore)
	// NEW: prepareStagingDirectory(stagingDir, appPath, appName, workspaceName string, workspace *models.Workspace, ds db.DataStore)

	// Expected behavior:
	// 1. Function accepts workspace parameter
	// 2. Passes workspace to generateShellConfig
	// 3. generateShellConfig uses workspace for theme resolution
}

// ==============================================================================
// Test 3: Integration with generateShellConfig
// ==============================================================================
// generateShellConfig should use resolveWorkspaceTheme to get the correct
// palette instead of using hardcoded default.

func TestGenerateShellConfig_UsesHierarchyResolvedTheme(t *testing.T) {
	t.Skip("Integration test - requires file system operations and database mock")

	// This test verifies:
	// 1. generateShellConfig calls resolveWorkspaceTheme
	// 2. Uses the resolved theme's palette for starship.toml generation
	// 3. Doesn't use hardcoded createDefaultPalette() when workspace has theme

	// Expected signature change:
	// OLD: generateShellConfig(stagingDir, appName, workspaceName string, ds db.DataStore)
	// NEW: generateShellConfig(stagingDir, appName, workspaceName string, workspace *models.Workspace, ds db.DataStore)
}

// ==============================================================================
// Test 4: Integration with generateNvimConfig
// ==============================================================================
// generateNvimConfig should use resolveWorkspaceTheme instead of themeStore.GetActive()

func TestGenerateNvimConfig_UsesHierarchyResolvedTheme(t *testing.T) {
	t.Skip("Integration test - requires file system operations and database mock")

	// This test verifies:
	// 1. generateNvimConfig calls resolveWorkspaceTheme
	// 2. Uses the hierarchy-resolved theme instead of global active theme
	// 3. Workspace-level theme overrides domain/ecosystem themes

	// Expected signature change:
	// OLD: generateNvimConfig(..., ds db.DataStore, app *models.App, workspace *models.Workspace, ...)
	// NEW: generateNvimConfig should use workspace parameter to resolve theme via hierarchy
}

// ==============================================================================
// Test 5: End-to-End Hierarchy Resolution
// ==============================================================================
// Comprehensive test that verifies the full workflow uses hierarchy correctly.

func TestBuildWorkflow_UsesHierarchyThemeResolution(t *testing.T) {
	t.Skip("End-to-end integration test - requires full build mock setup")

	// This test verifies the complete build workflow:
	// 1. buildWorkspace() retrieves workspace
	// 2. prepareStagingDirectory() passes workspace to generateShellConfig
	// 3. generateShellConfig() resolves theme via hierarchy
	// 4. generateNvimConfig() resolves theme via hierarchy
	// 5. Both use the SAME theme from workspace hierarchy, not global theme
}

// ==============================================================================
// NOTE: Helper function implementation moved to build.go (TDD Phase 3 - GREEN)
// ==============================================================================
//
// The resolveWorkspaceTheme() function has been implemented in build.go
// and is now used by generateShellConfig() to resolve themes from the
// workspace hierarchy instead of using a hardcoded default palette.
//
// See build.go for the implementation.
