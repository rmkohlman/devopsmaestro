package db

import (
	"database/sql"
	"testing"

	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Terminal Fields Persistence Tests (Issue #8)
// =============================================================================
// These tests verify that terminal_prompt, terminal_plugins, and terminal_package
// fields are properly persisted via UpdateWorkspace, ListWorkspacesByApp, and
// ListAllWorkspaces.
//
// EXPECTED TO FAIL with current code - UpdateWorkspace is missing these columns.
// =============================================================================

// TestWorkspace_TerminalPromptPersistence verifies terminal_prompt persists via UpdateWorkspace
func TestWorkspace_TerminalPromptPersistence(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create ecosystem, domain, app
	ecosystem := &models.Ecosystem{Name: "test-eco"}
	require.NoError(t, ds.CreateEcosystem(ecosystem))

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "test-domain",
	}
	require.NoError(t, ds.CreateDomain(domain))

	app := &models.App{
		DomainID: domain.ID,
		Name:     "test-app",
		Path:     "/test",
	}
	require.NoError(t, ds.CreateApp(app))

	// Create workspace
	workspace := &models.Workspace{
		AppID:     app.ID,
		Name:      "test-workspace",
		ImageName: "test-image",
		Slug:      "test-app-test-workspace",
	}
	require.NoError(t, ds.CreateWorkspace(workspace))

	// Set terminal_prompt and update
	workspace.TerminalPrompt = sql.NullString{String: "starship", Valid: true}
	err := ds.UpdateWorkspace(workspace)
	require.NoError(t, err, "UpdateWorkspace should succeed")

	// Retrieve workspace and verify terminal_prompt persisted
	retrieved, err := ds.GetWorkspaceByID(workspace.ID)
	require.NoError(t, err)
	assert.True(t, retrieved.TerminalPrompt.Valid, "terminal_prompt should be valid")
	assert.Equal(t, "starship", retrieved.TerminalPrompt.String, "terminal_prompt should match")
}

// TestWorkspace_TerminalPluginsPersistence verifies terminal_plugins persists via UpdateWorkspace
func TestWorkspace_TerminalPluginsPersistence(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create ecosystem, domain, app
	ecosystem := &models.Ecosystem{Name: "test-eco"}
	require.NoError(t, ds.CreateEcosystem(ecosystem))

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "test-domain",
	}
	require.NoError(t, ds.CreateDomain(domain))

	app := &models.App{
		DomainID: domain.ID,
		Name:     "test-app",
		Path:     "/test",
	}
	require.NoError(t, ds.CreateApp(app))

	// Create workspace
	workspace := &models.Workspace{
		AppID:     app.ID,
		Name:      "test-workspace",
		ImageName: "test-image",
		Slug:      "test-app-test-workspace",
	}
	require.NoError(t, ds.CreateWorkspace(workspace))

	// Set terminal_plugins using the SetTerminalPlugins helper
	plugins := []string{"zoxide", "fzf", "eza"}
	workspace.SetTerminalPlugins(plugins)
	err := ds.UpdateWorkspace(workspace)
	require.NoError(t, err, "UpdateWorkspace should succeed")

	// Retrieve workspace and verify terminal_plugins persisted
	retrieved, err := ds.GetWorkspaceByID(workspace.ID)
	require.NoError(t, err)
	assert.True(t, retrieved.TerminalPlugins.Valid, "terminal_plugins should be valid")

	// Verify plugins using GetTerminalPlugins helper
	retrievedPlugins := retrieved.GetTerminalPlugins()
	assert.Equal(t, plugins, retrievedPlugins, "terminal_plugins should match")
}

// TestWorkspace_TerminalPackagePersistence verifies terminal_package persists via UpdateWorkspace
func TestWorkspace_TerminalPackagePersistence(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create ecosystem, domain, app
	ecosystem := &models.Ecosystem{Name: "test-eco"}
	require.NoError(t, ds.CreateEcosystem(ecosystem))

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "test-domain",
	}
	require.NoError(t, ds.CreateDomain(domain))

	app := &models.App{
		DomainID: domain.ID,
		Name:     "test-app",
		Path:     "/test",
	}
	require.NoError(t, ds.CreateApp(app))

	// Create workspace
	workspace := &models.Workspace{
		AppID:     app.ID,
		Name:      "test-workspace",
		ImageName: "test-image",
		Slug:      "test-app-test-workspace",
	}
	require.NoError(t, ds.CreateWorkspace(workspace))

	// Set terminal_package and update
	workspace.TerminalPackage = sql.NullString{String: "maestro", Valid: true}
	err := ds.UpdateWorkspace(workspace)
	require.NoError(t, err, "UpdateWorkspace should succeed")

	// Retrieve workspace and verify terminal_package persisted
	retrieved, err := ds.GetWorkspaceByID(workspace.ID)
	require.NoError(t, err)
	assert.True(t, retrieved.TerminalPackage.Valid, "terminal_package should be valid")
	assert.Equal(t, "maestro", retrieved.TerminalPackage.String, "terminal_package should match")
}

// TestWorkspace_AllTerminalFieldsPersistence verifies all terminal fields persist together
func TestWorkspace_AllTerminalFieldsPersistence(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create ecosystem, domain, app
	ecosystem := &models.Ecosystem{Name: "test-eco"}
	require.NoError(t, ds.CreateEcosystem(ecosystem))

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "test-domain",
	}
	require.NoError(t, ds.CreateDomain(domain))

	app := &models.App{
		DomainID: domain.ID,
		Name:     "test-app",
		Path:     "/test",
	}
	require.NoError(t, ds.CreateApp(app))

	// Create workspace
	workspace := &models.Workspace{
		AppID:     app.ID,
		Name:      "test-workspace",
		ImageName: "test-image",
		Slug:      "test-app-test-workspace",
	}
	require.NoError(t, ds.CreateWorkspace(workspace))

	// Set all terminal fields
	workspace.TerminalPrompt = sql.NullString{String: "starship", Valid: true}
	workspace.SetTerminalPlugins([]string{"zoxide", "fzf"})
	workspace.TerminalPackage = sql.NullString{String: "developer", Valid: true}

	err := ds.UpdateWorkspace(workspace)
	require.NoError(t, err, "UpdateWorkspace should succeed")

	// Retrieve workspace and verify all fields persisted
	retrieved, err := ds.GetWorkspaceByID(workspace.ID)
	require.NoError(t, err)

	assert.True(t, retrieved.TerminalPrompt.Valid, "terminal_prompt should be valid")
	assert.Equal(t, "starship", retrieved.TerminalPrompt.String, "terminal_prompt should match")

	assert.True(t, retrieved.TerminalPlugins.Valid, "terminal_plugins should be valid")
	assert.Equal(t, []string{"zoxide", "fzf"}, retrieved.GetTerminalPlugins(), "terminal_plugins should match")

	assert.True(t, retrieved.TerminalPackage.Valid, "terminal_package should be valid")
	assert.Equal(t, "developer", retrieved.TerminalPackage.String, "terminal_package should match")
}

// TestListWorkspacesByApp_IncludesTerminalFields verifies terminal fields are returned
func TestListWorkspacesByApp_IncludesTerminalFields(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create ecosystem, domain, app
	ecosystem := &models.Ecosystem{Name: "test-eco"}
	require.NoError(t, ds.CreateEcosystem(ecosystem))

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "test-domain",
	}
	require.NoError(t, ds.CreateDomain(domain))

	app := &models.App{
		DomainID: domain.ID,
		Name:     "test-app",
		Path:     "/test",
	}
	require.NoError(t, ds.CreateApp(app))

	// Create workspace with terminal fields
	workspace := &models.Workspace{
		AppID:           app.ID,
		Name:            "test-workspace",
		ImageName:       "test-image",
		Slug:            "test-app-test-workspace",
		TerminalPrompt:  sql.NullString{String: "starship", Valid: true},
		TerminalPackage: sql.NullString{String: "developer", Valid: true},
	}
	workspace.SetTerminalPlugins([]string{"zoxide"})
	require.NoError(t, ds.CreateWorkspace(workspace))

	// List workspaces by app
	workspaces, err := ds.ListWorkspacesByApp(app.ID)
	require.NoError(t, err)
	require.Len(t, workspaces, 1)

	// Verify terminal fields are included
	ws := workspaces[0]
	assert.True(t, ws.TerminalPrompt.Valid, "terminal_prompt should be valid")
	assert.Equal(t, "starship", ws.TerminalPrompt.String, "terminal_prompt should match")

	assert.True(t, ws.TerminalPlugins.Valid, "terminal_plugins should be valid")
	assert.Equal(t, []string{"zoxide"}, ws.GetTerminalPlugins(), "terminal_plugins should match")

	assert.True(t, ws.TerminalPackage.Valid, "terminal_package should be valid")
	assert.Equal(t, "developer", ws.TerminalPackage.String, "terminal_package should match")
}

// TestListAllWorkspaces_IncludesTerminalFields verifies terminal fields are returned
func TestListAllWorkspaces_IncludesTerminalFields(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create ecosystem, domain, app
	ecosystem := &models.Ecosystem{Name: "test-eco"}
	require.NoError(t, ds.CreateEcosystem(ecosystem))

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "test-domain",
	}
	require.NoError(t, ds.CreateDomain(domain))

	app := &models.App{
		DomainID: domain.ID,
		Name:     "test-app",
		Path:     "/test",
	}
	require.NoError(t, ds.CreateApp(app))

	// Create workspace with terminal fields
	workspace := &models.Workspace{
		AppID:           app.ID,
		Name:            "test-workspace",
		ImageName:       "test-image",
		Slug:            "test-app-test-workspace",
		TerminalPrompt:  sql.NullString{String: "powerlevel10k", Valid: true},
		TerminalPackage: sql.NullString{String: "maestro", Valid: true},
	}
	workspace.SetTerminalPlugins([]string{"eza", "bat"})
	require.NoError(t, ds.CreateWorkspace(workspace))

	// List all workspaces
	workspaces, err := ds.ListAllWorkspaces()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(workspaces), 1, "should have at least one workspace")

	// Find our workspace
	var ws *models.Workspace
	for _, w := range workspaces {
		if w.ID == workspace.ID {
			ws = w
			break
		}
	}
	require.NotNil(t, ws, "should find our workspace")

	// Verify terminal fields are included
	assert.True(t, ws.TerminalPrompt.Valid, "terminal_prompt should be valid")
	assert.Equal(t, "powerlevel10k", ws.TerminalPrompt.String, "terminal_prompt should match")

	assert.True(t, ws.TerminalPlugins.Valid, "terminal_plugins should be valid")
	assert.Equal(t, []string{"eza", "bat"}, ws.GetTerminalPlugins(), "terminal_plugins should match")

	assert.True(t, ws.TerminalPackage.Valid, "terminal_package should be valid")
	assert.Equal(t, "maestro", ws.TerminalPackage.String, "terminal_package should match")
}
