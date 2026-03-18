package db

// =============================================================================
// Build Args / Build Config Round-Trip Tests (v0.55.0)
// =============================================================================
// Migration 017 adds:
//   - ecosystems.build_args TEXT (nullable)
//   - domains.build_args TEXT (nullable)
//   - workspaces.build_config TEXT (nullable)
//
// These tests are RED until:
//   1. createTestSchema() includes the new columns, AND
//   2. store_ecosystem.go, store_domain.go, store_workspace.go SQL queries
//      include build_args / build_config in INSERT, UPDATE, and SELECT.
//
// Tests use the same createTestDataStore() helper used throughout db/ tests.
// =============================================================================

import (
	"database/sql"
	"fmt"
	"testing"

	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Helpers
// =============================================================================

// createTestHierarchy creates the full Ecosystem → Domain → App chain and
// returns all three objects. Name is used as a unique suffix to avoid collisions.
func createTestHierarchy(t *testing.T, ds *SQLDataStore, name string) (*models.Ecosystem, *models.Domain, *models.App) {
	t.Helper()

	ecosystem := &models.Ecosystem{Name: "ba-eco-" + name}
	require.NoError(t, ds.CreateEcosystem(ecosystem), "setup: CreateEcosystem")

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "ba-dom-" + name,
	}
	require.NoError(t, ds.CreateDomain(domain), "setup: CreateDomain")

	app := &models.App{
		DomainID: domain.ID,
		Name:     "ba-app-" + name,
		Path:     "/ba/" + name,
	}
	require.NoError(t, ds.CreateApp(app), "setup: CreateApp")

	return ecosystem, domain, app
}

// =============================================================================
// 1. TestEcosystem_BuildArgs_RoundTrip
// =============================================================================

func TestEcosystem_BuildArgs_RoundTrip(t *testing.T) {
	tests := []struct {
		name            string
		initialArgs     sql.NullString
		wantValid       bool
		wantString      string
		updateArgs      sql.NullString
		wantValidAfter  bool
		wantStringAfter string
	}{
		{
			name: "ecosystem with build args round-trips",
			initialArgs: sql.NullString{
				String: `{"PIP_INDEX_URL":"https://pypi.corp.com/root/prod"}`,
				Valid:  true,
			},
			wantValid:  true,
			wantString: `{"PIP_INDEX_URL":"https://pypi.corp.com/root/prod"}`,
			updateArgs: sql.NullString{
				String: `{"PIP_INDEX_URL":"https://pypi.corp.com/root/staging","PIP_TRUSTED_HOST":"pypi.corp.com"}`,
				Valid:  true,
			},
			wantValidAfter:  true,
			wantStringAfter: `{"PIP_INDEX_URL":"https://pypi.corp.com/root/staging","PIP_TRUSTED_HOST":"pypi.corp.com"}`,
		},
		{
			name:            "ecosystem without build args has Valid=false",
			initialArgs:     sql.NullString{},
			wantValid:       false,
			wantString:      "",
			updateArgs:      sql.NullString{String: `{"NEW_ARG":"value"}`, Valid: true},
			wantValidAfter:  true,
			wantStringAfter: `{"NEW_ARG":"value"}`,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := createTestDataStore(t)
			defer ds.Close()

			// --- Create ---
			eco := &models.Ecosystem{
				Name:      fmt.Sprintf("ba-eco-rt-%d", i),
				BuildArgs: tt.initialArgs,
			}
			require.NoError(t, ds.CreateEcosystem(eco))
			require.NotZero(t, eco.ID, "CreateEcosystem should set ID")

			// --- GetByName round-trip ---
			byName, err := ds.GetEcosystemByName(eco.Name)
			require.NoError(t, err)
			assert.Equal(t, tt.wantValid, byName.BuildArgs.Valid,
				"GetEcosystemByName: BuildArgs.Valid mismatch")
			if tt.wantValid {
				assert.Equal(t, tt.wantString, byName.BuildArgs.String,
					"GetEcosystemByName: BuildArgs.String mismatch")
			}

			// --- GetByID round-trip ---
			byID, err := ds.GetEcosystemByID(eco.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantValid, byID.BuildArgs.Valid,
				"GetEcosystemByID: BuildArgs.Valid mismatch")
			if tt.wantValid {
				assert.Equal(t, tt.wantString, byID.BuildArgs.String,
					"GetEcosystemByID: BuildArgs.String mismatch")
			}

			// --- Update ---
			eco.BuildArgs = tt.updateArgs
			require.NoError(t, ds.UpdateEcosystem(eco))

			// --- GetByID after update ---
			afterUpdate, err := ds.GetEcosystemByID(eco.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantValidAfter, afterUpdate.BuildArgs.Valid,
				"after UpdateEcosystem: BuildArgs.Valid mismatch")
			if tt.wantValidAfter {
				assert.Equal(t, tt.wantStringAfter, afterUpdate.BuildArgs.String,
					"after UpdateEcosystem: BuildArgs.String mismatch")
			}
		})
	}
}

// =============================================================================
// 2. TestDomain_BuildArgs_RoundTrip
// =============================================================================

func TestDomain_BuildArgs_RoundTrip(t *testing.T) {
	tests := []struct {
		name            string
		initialArgs     sql.NullString
		wantValid       bool
		wantString      string
		updateArgs      sql.NullString
		wantValidAfter  bool
		wantStringAfter string
	}{
		{
			name: "domain with build args round-trips",
			initialArgs: sql.NullString{
				String: `{"NPM_REGISTRY":"https://npm.corp.com"}`,
				Valid:  true,
			},
			wantValid:  true,
			wantString: `{"NPM_REGISTRY":"https://npm.corp.com"}`,
			updateArgs: sql.NullString{
				String: `{"NPM_REGISTRY":"https://npm.corp.com","NPM_STRICT_SSL":"false"}`,
				Valid:  true,
			},
			wantValidAfter:  true,
			wantStringAfter: `{"NPM_REGISTRY":"https://npm.corp.com","NPM_STRICT_SSL":"false"}`,
		},
		{
			name:            "domain without build args has Valid=false",
			initialArgs:     sql.NullString{},
			wantValid:       false,
			wantString:      "",
			updateArgs:      sql.NullString{String: `{"MAVEN_REPO":"https://nexus.corp.com"}`, Valid: true},
			wantValidAfter:  true,
			wantStringAfter: `{"MAVEN_REPO":"https://nexus.corp.com"}`,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := createTestDataStore(t)
			defer ds.Close()

			eco := &models.Ecosystem{Name: fmt.Sprintf("ba-eco-dom-rt-%d", i)}
			require.NoError(t, ds.CreateEcosystem(eco))

			dom := &models.Domain{
				EcosystemID: eco.ID,
				Name:        fmt.Sprintf("ba-dom-rt-%d", i),
				BuildArgs:   tt.initialArgs,
			}
			require.NoError(t, ds.CreateDomain(dom))
			require.NotZero(t, dom.ID, "CreateDomain should set ID")

			// --- GetByName round-trip ---
			byName, err := ds.GetDomainByName(eco.ID, dom.Name)
			require.NoError(t, err)
			assert.Equal(t, tt.wantValid, byName.BuildArgs.Valid,
				"GetDomainByName: BuildArgs.Valid mismatch")
			if tt.wantValid {
				assert.Equal(t, tt.wantString, byName.BuildArgs.String,
					"GetDomainByName: BuildArgs.String mismatch")
			}

			// --- GetByID round-trip ---
			byID, err := ds.GetDomainByID(dom.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantValid, byID.BuildArgs.Valid,
				"GetDomainByID: BuildArgs.Valid mismatch")
			if tt.wantValid {
				assert.Equal(t, tt.wantString, byID.BuildArgs.String,
					"GetDomainByID: BuildArgs.String mismatch")
			}

			// --- Update ---
			dom.BuildArgs = tt.updateArgs
			require.NoError(t, ds.UpdateDomain(dom))

			// --- GetByID after update ---
			afterUpdate, err := ds.GetDomainByID(dom.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantValidAfter, afterUpdate.BuildArgs.Valid,
				"after UpdateDomain: BuildArgs.Valid mismatch")
			if tt.wantValidAfter {
				assert.Equal(t, tt.wantStringAfter, afterUpdate.BuildArgs.String,
					"after UpdateDomain: BuildArgs.String mismatch")
			}
		})
	}
}

// =============================================================================
// 3. TestWorkspace_BuildConfig_RoundTrip
// =============================================================================

func TestWorkspace_BuildConfig_RoundTrip(t *testing.T) {
	devBuildJSON := `{"args":{"GOPROXY":"https://goproxy.corp.com","GOPRIVATE":"*.corp.com"},"caCerts":[{"name":"corp-ca","vaultSecret":"corp-root-ca"}]}`
	updatedBuildJSON := `{"args":{"GOPROXY":"direct","GONOSUMCHECK":"*.corp.com"}}`

	tests := []struct {
		name            string
		initialConfig   sql.NullString
		wantValid       bool
		wantString      string
		updateConfig    sql.NullString
		wantValidAfter  bool
		wantStringAfter string
	}{
		{
			name:            "workspace with build config round-trips",
			initialConfig:   sql.NullString{String: devBuildJSON, Valid: true},
			wantValid:       true,
			wantString:      devBuildJSON,
			updateConfig:    sql.NullString{String: updatedBuildJSON, Valid: true},
			wantValidAfter:  true,
			wantStringAfter: updatedBuildJSON,
		},
		{
			name:            "workspace without build config has Valid=false",
			initialConfig:   sql.NullString{},
			wantValid:       false,
			wantString:      "",
			updateConfig:    sql.NullString{String: devBuildJSON, Valid: true},
			wantValidAfter:  true,
			wantStringAfter: devBuildJSON,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := createTestDataStore(t)
			defer ds.Close()

			_, _, app := createTestHierarchy(t, ds, fmt.Sprintf("ws-bc-rt-%d", i))

			ws := &models.Workspace{
				AppID:       app.ID,
				Name:        "ws-build",
				Slug:        fmt.Sprintf("ba-eco-ws-bc-rt-%d-ba-dom-ws-bc-rt-%d-ba-app-ws-bc-rt-%d-ws-build", i, i, i),
				ImageName:   "golang:1.22",
				Status:      "stopped",
				BuildConfig: tt.initialConfig,
			}
			require.NoError(t, ds.CreateWorkspace(ws))
			require.NotZero(t, ws.ID, "CreateWorkspace should set ID")

			// --- GetByName round-trip ---
			byName, err := ds.GetWorkspaceByName(app.ID, "ws-build")
			require.NoError(t, err)
			assert.Equal(t, tt.wantValid, byName.BuildConfig.Valid,
				"GetWorkspaceByName: BuildConfig.Valid mismatch")
			if tt.wantValid {
				assert.Equal(t, tt.wantString, byName.BuildConfig.String,
					"GetWorkspaceByName: BuildConfig.String mismatch")
			}

			// --- GetByID round-trip ---
			byID, err := ds.GetWorkspaceByID(ws.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantValid, byID.BuildConfig.Valid,
				"GetWorkspaceByID: BuildConfig.Valid mismatch")
			if tt.wantValid {
				assert.Equal(t, tt.wantString, byID.BuildConfig.String,
					"GetWorkspaceByID: BuildConfig.String mismatch")
			}

			// --- GetBySlug round-trip ---
			bySlug, err := ds.GetWorkspaceBySlug(ws.Slug)
			require.NoError(t, err)
			assert.Equal(t, tt.wantValid, bySlug.BuildConfig.Valid,
				"GetWorkspaceBySlug: BuildConfig.Valid mismatch")
			if tt.wantValid {
				assert.Equal(t, tt.wantString, bySlug.BuildConfig.String,
					"GetWorkspaceBySlug: BuildConfig.String mismatch")
			}

			// --- Update ---
			ws.BuildConfig = tt.updateConfig
			require.NoError(t, ds.UpdateWorkspace(ws))

			// --- GetByID after update ---
			afterUpdate, err := ds.GetWorkspaceByID(ws.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantValidAfter, afterUpdate.BuildConfig.Valid,
				"after UpdateWorkspace: BuildConfig.Valid mismatch")
			if tt.wantValidAfter {
				assert.Equal(t, tt.wantStringAfter, afterUpdate.BuildConfig.String,
					"after UpdateWorkspace: BuildConfig.String mismatch")
			}
		})
	}
}

// =============================================================================
// 4. TestListEcosystems_IncludesBuildArgs
// =============================================================================

func TestListEcosystems_IncludesBuildArgs(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	buildArgsJSON := `{"PIP_INDEX_URL":"https://pypi.corp.com/root/prod"}`

	// Create ecosystem WITH build args
	ecoWith := &models.Ecosystem{
		Name:      "list-eco-with-args",
		BuildArgs: sql.NullString{String: buildArgsJSON, Valid: true},
	}
	require.NoError(t, ds.CreateEcosystem(ecoWith))

	// Create ecosystem WITHOUT build args
	ecoWithout := &models.Ecosystem{
		Name:      "list-eco-without-args",
		BuildArgs: sql.NullString{},
	}
	require.NoError(t, ds.CreateEcosystem(ecoWithout))

	// List all ecosystems
	ecosystems, err := ds.ListEcosystems()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(ecosystems), 2, "should return at least 2 ecosystems")

	// Find each by ID and verify BuildArgs
	var foundWith, foundWithout *models.Ecosystem
	for _, e := range ecosystems {
		switch e.ID {
		case ecoWith.ID:
			foundWith = e
		case ecoWithout.ID:
			foundWithout = e
		}
	}

	require.NotNil(t, foundWith, "ListEcosystems should include ecosystem with build args")
	assert.True(t, foundWith.BuildArgs.Valid,
		"ListEcosystems: ecosystem with build args should have BuildArgs.Valid=true")
	assert.Equal(t, buildArgsJSON, foundWith.BuildArgs.String,
		"ListEcosystems: ecosystem build args string should match")

	require.NotNil(t, foundWithout, "ListEcosystems should include ecosystem without build args")
	assert.False(t, foundWithout.BuildArgs.Valid,
		"ListEcosystems: ecosystem without build args should have BuildArgs.Valid=false")
}

// =============================================================================
// 5. TestListDomains_IncludesBuildArgs
// =============================================================================

func TestListDomains_IncludesBuildArgs(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "list-dom-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))

	buildArgsJSON := `{"NPM_REGISTRY":"https://npm.corp.com"}`

	domWith := &models.Domain{
		EcosystemID: eco.ID,
		Name:        "domain-with-args",
		BuildArgs:   sql.NullString{String: buildArgsJSON, Valid: true},
	}
	require.NoError(t, ds.CreateDomain(domWith))

	domWithout := &models.Domain{
		EcosystemID: eco.ID,
		Name:        "domain-without-args",
		BuildArgs:   sql.NullString{},
	}
	require.NoError(t, ds.CreateDomain(domWithout))

	t.Run("ListDomainsByEcosystem includes BuildArgs", func(t *testing.T) {
		domains, err := ds.ListDomainsByEcosystem(eco.ID)
		require.NoError(t, err)
		require.Len(t, domains, 2, "should return 2 domains")

		var foundWith, foundWithout *models.Domain
		for _, d := range domains {
			switch d.ID {
			case domWith.ID:
				foundWith = d
			case domWithout.ID:
				foundWithout = d
			}
		}

		require.NotNil(t, foundWith, "ListDomainsByEcosystem should include domain with build args")
		assert.True(t, foundWith.BuildArgs.Valid,
			"ListDomainsByEcosystem: domain with build args should have BuildArgs.Valid=true")
		assert.Equal(t, buildArgsJSON, foundWith.BuildArgs.String,
			"ListDomainsByEcosystem: domain build args string should match")

		require.NotNil(t, foundWithout, "ListDomainsByEcosystem should include domain without build args")
		assert.False(t, foundWithout.BuildArgs.Valid,
			"ListDomainsByEcosystem: domain without build args should have BuildArgs.Valid=false")
	})

	t.Run("ListAllDomains includes BuildArgs", func(t *testing.T) {
		domains, err := ds.ListAllDomains()
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(domains), 2)

		var foundWith, foundWithout *models.Domain
		for _, d := range domains {
			switch d.ID {
			case domWith.ID:
				foundWith = d
			case domWithout.ID:
				foundWithout = d
			}
		}

		require.NotNil(t, foundWith, "ListAllDomains should include domain with build args")
		assert.True(t, foundWith.BuildArgs.Valid,
			"ListAllDomains: domain with build args should have BuildArgs.Valid=true")
		assert.Equal(t, buildArgsJSON, foundWith.BuildArgs.String,
			"ListAllDomains: domain build args string should match")

		require.NotNil(t, foundWithout, "ListAllDomains should include domain without build args")
		assert.False(t, foundWithout.BuildArgs.Valid,
			"ListAllDomains: domain without build args should have BuildArgs.Valid=false")
	})
}

// =============================================================================
// 6. TestListWorkspaces_IncludesBuildConfig
// =============================================================================

func TestListWorkspaces_IncludesBuildConfig(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	_, _, app := createTestHierarchy(t, ds, "listws")

	buildConfigJSON := `{"args":{"GOPROXY":"https://goproxy.corp.com"}}`

	wsWithConfig := &models.Workspace{
		AppID:       app.ID,
		Name:        "ws-with-config",
		Slug:        "ba-eco-listws-ba-dom-listws-ba-app-listws-ws-with-config",
		ImageName:   "golang:1.22",
		Status:      "stopped",
		BuildConfig: sql.NullString{String: buildConfigJSON, Valid: true},
	}
	require.NoError(t, ds.CreateWorkspace(wsWithConfig))

	wsWithoutConfig := &models.Workspace{
		AppID:     app.ID,
		Name:      "ws-without-config",
		Slug:      "ba-eco-listws-ba-dom-listws-ba-app-listws-ws-without-config",
		ImageName: "golang:1.22",
		Status:    "stopped",
	}
	require.NoError(t, ds.CreateWorkspace(wsWithoutConfig))

	t.Run("ListWorkspacesByApp includes BuildConfig", func(t *testing.T) {
		workspaces, err := ds.ListWorkspacesByApp(app.ID)
		require.NoError(t, err)
		require.Len(t, workspaces, 2, "should return 2 workspaces")

		var foundWith, foundWithout *models.Workspace
		for _, ws := range workspaces {
			switch ws.ID {
			case wsWithConfig.ID:
				foundWith = ws
			case wsWithoutConfig.ID:
				foundWithout = ws
			}
		}

		require.NotNil(t, foundWith, "ListWorkspacesByApp should include workspace with build config")
		assert.True(t, foundWith.BuildConfig.Valid,
			"ListWorkspacesByApp: workspace with build config should have BuildConfig.Valid=true")
		assert.Equal(t, buildConfigJSON, foundWith.BuildConfig.String,
			"ListWorkspacesByApp: workspace build config string should match")

		require.NotNil(t, foundWithout, "ListWorkspacesByApp should include workspace without build config")
		assert.False(t, foundWithout.BuildConfig.Valid,
			"ListWorkspacesByApp: workspace without build config should have BuildConfig.Valid=false")
	})

	t.Run("ListAllWorkspaces includes BuildConfig", func(t *testing.T) {
		workspaces, err := ds.ListAllWorkspaces()
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(workspaces), 2)

		var foundWith, foundWithout *models.Workspace
		for _, ws := range workspaces {
			switch ws.ID {
			case wsWithConfig.ID:
				foundWith = ws
			case wsWithoutConfig.ID:
				foundWithout = ws
			}
		}

		require.NotNil(t, foundWith, "ListAllWorkspaces should include workspace with build config")
		assert.True(t, foundWith.BuildConfig.Valid,
			"ListAllWorkspaces: workspace with build config should have BuildConfig.Valid=true")
		assert.Equal(t, buildConfigJSON, foundWith.BuildConfig.String,
			"ListAllWorkspaces: workspace build config string should match")

		require.NotNil(t, foundWithout, "ListAllWorkspaces should include workspace without build config")
		assert.False(t, foundWithout.BuildConfig.Valid,
			"ListAllWorkspaces: workspace without build config should have BuildConfig.Valid=false")
	})
}

// =============================================================================
// 7. TestFindWorkspaces_IncludesBuildConfig
// =============================================================================

func TestFindWorkspaces_IncludesBuildConfig(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	eco, _, app := createTestHierarchy(t, ds, "findws")

	buildConfigJSON := `{"args":{"GOPROXY":"https://goproxy.corp.com","GOPRIVATE":"*.corp.com"}}`

	ws := &models.Workspace{
		AppID:       app.ID,
		Name:        "find-ws-dev",
		Slug:        "ba-eco-findws-ba-dom-findws-ba-app-findws-find-ws-dev",
		ImageName:   "golang:1.22",
		Status:      "stopped",
		BuildConfig: sql.NullString{String: buildConfigJSON, Valid: true},
	}
	require.NoError(t, ds.CreateWorkspace(ws))

	// FindWorkspaces with ecosystem filter
	filter := models.WorkspaceFilter{
		EcosystemName: eco.Name,
		WorkspaceName: "find-ws-dev",
	}
	results, err := ds.FindWorkspaces(filter)
	require.NoError(t, err)
	require.Len(t, results, 1, "FindWorkspaces should return exactly 1 result")

	found := results[0]
	require.NotNil(t, found.Workspace, "FindWorkspaces result should have a Workspace")

	assert.True(t, found.Workspace.BuildConfig.Valid,
		"FindWorkspaces: workspace BuildConfig.Valid should be true")
	assert.Equal(t, buildConfigJSON, found.Workspace.BuildConfig.String,
		"FindWorkspaces: workspace BuildConfig.String should match")

	// Also verify hierarchy objects are present
	assert.Equal(t, ws.ID, found.Workspace.ID, "Workspace ID should match")
	require.NotNil(t, found.App, "FindWorkspaces result should have an App")
	require.NotNil(t, found.Domain, "FindWorkspaces result should have a Domain")
	require.NotNil(t, found.Ecosystem, "FindWorkspaces result should have an Ecosystem")
	assert.Equal(t, eco.Name, found.Ecosystem.Name, "Ecosystem name should match")
}
