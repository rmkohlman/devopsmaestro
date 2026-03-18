package db

// =============================================================================
// CA Certs Store Round-Trip Tests (v0.56.0)
// =============================================================================
// Migration 018 (planned) adds:
//   - ecosystems.ca_certs TEXT (nullable, JSON array of CACertConfig)
//   - domains.ca_certs TEXT (nullable, JSON array of CACertConfig)
//   - apps.build_config already has caCerts in JSON (no schema change needed)
//
// These tests are RED until:
//   1. createTestSchema() includes the new ca_certs columns on ecosystems/domains
//   2. store_ecosystem.go includes ca_certs in INSERT, UPDATE, and SELECT
//   3. store_domain.go includes ca_certs in INSERT, UPDATE, and SELECT
//   4. models.Ecosystem.CACerts and models.Domain.CACerts fields are added
//   5. models.AppBuildConfig.CACerts field is added
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

// createTestHierarchyForCACerts creates the full Ecosystem → Domain → App chain.
// Returns all three objects. Name suffix is used to avoid collisions.
func createTestHierarchyForCACerts(t *testing.T, ds *SQLDataStore, name string) (*models.Ecosystem, *models.Domain, *models.App) {
	t.Helper()

	ecosystem := &models.Ecosystem{Name: "ca-eco-" + name}
	require.NoError(t, ds.CreateEcosystem(ecosystem), "setup: CreateEcosystem")

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "ca-dom-" + name,
	}
	require.NoError(t, ds.CreateDomain(domain), "setup: CreateDomain")

	app := &models.App{
		DomainID: domain.ID,
		Name:     "ca-app-" + name,
		Path:     "/ca/" + name,
	}
	require.NoError(t, ds.CreateApp(app), "setup: CreateApp")

	return ecosystem, domain, app
}

// =============================================================================
// 1. TestEcosystem_CACerts_StoreRoundTrip
// =============================================================================

// TestEcosystem_CACerts_StoreRoundTrip verifies that an ecosystem created with
// CA certs (stored as JSON in the ca_certs column) can be retrieved with the
// CA certs intact via GetEcosystemByName and GetEcosystemByID.
//
// RED: WILL NOT COMPILE / FAIL until:
//   - Ecosystem.CACerts sql.NullString field exists in models
//   - ca_certs column exists in ecosystems table (createTestSchema)
//   - store_ecosystem.go queries include ca_certs in INSERT, UPDATE, SELECT
func TestEcosystem_CACerts_StoreRoundTrip(t *testing.T) {
	tests := []struct {
		name            string
		initialCACerts  sql.NullString
		wantValid       bool
		wantString      string
		updateCACerts   sql.NullString
		wantValidAfter  bool
		wantStringAfter string
	}{
		{
			name: "ecosystem with ca certs round-trips",
			initialCACerts: sql.NullString{
				String: `[{"name":"corp-root-ca","vaultSecret":"corp-root-ca-pem"}]`,
				Valid:  true,
			},
			wantValid:  true,
			wantString: `[{"name":"corp-root-ca","vaultSecret":"corp-root-ca-pem"}]`,
			updateCACerts: sql.NullString{
				String: `[{"name":"corp-root-ca","vaultSecret":"new-corp-root-ca-pem"},{"name":"corp-int-ca","vaultSecret":"corp-int-ca-pem"}]`,
				Valid:  true,
			},
			wantValidAfter:  true,
			wantStringAfter: `[{"name":"corp-root-ca","vaultSecret":"new-corp-root-ca-pem"},{"name":"corp-int-ca","vaultSecret":"corp-int-ca-pem"}]`,
		},
		{
			name:           "ecosystem without ca certs has Valid=false",
			initialCACerts: sql.NullString{},
			wantValid:      false,
			wantString:     "",
			updateCACerts: sql.NullString{
				String: `[{"name":"new-ca","vaultSecret":"new-ca-pem"}]`,
				Valid:  true,
			},
			wantValidAfter:  true,
			wantStringAfter: `[{"name":"new-ca","vaultSecret":"new-ca-pem"}]`,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := createTestDataStore(t)
			defer ds.Close()

			// --- Create ---
			// RED: Ecosystem.CACerts does not exist yet
			eco := &models.Ecosystem{
				Name:    fmt.Sprintf("ca-eco-rt-%d", i),
				CACerts: tt.initialCACerts,
			}
			require.NoError(t, ds.CreateEcosystem(eco))
			require.NotZero(t, eco.ID, "CreateEcosystem should set ID")

			// --- GetByName round-trip ---
			byName, err := ds.GetEcosystemByName(eco.Name)
			require.NoError(t, err)
			// RED: byName.CACerts does not exist yet
			assert.Equal(t, tt.wantValid, byName.CACerts.Valid,
				"GetEcosystemByName: CACerts.Valid mismatch")
			if tt.wantValid {
				assert.Equal(t, tt.wantString, byName.CACerts.String,
					"GetEcosystemByName: CACerts.String mismatch")
			}

			// --- GetByID round-trip ---
			byID, err := ds.GetEcosystemByID(eco.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantValid, byID.CACerts.Valid,
				"GetEcosystemByID: CACerts.Valid mismatch")
			if tt.wantValid {
				assert.Equal(t, tt.wantString, byID.CACerts.String,
					"GetEcosystemByID: CACerts.String mismatch")
			}

			// --- Update ---
			eco.CACerts = tt.updateCACerts
			require.NoError(t, ds.UpdateEcosystem(eco))

			// --- GetByID after update ---
			afterUpdate, err := ds.GetEcosystemByID(eco.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantValidAfter, afterUpdate.CACerts.Valid,
				"after UpdateEcosystem: CACerts.Valid mismatch")
			if tt.wantValidAfter {
				assert.Equal(t, tt.wantStringAfter, afterUpdate.CACerts.String,
					"after UpdateEcosystem: CACerts.String mismatch")
			}
		})
	}
}

// =============================================================================
// 2. TestEcosystem_CACerts_Update
// =============================================================================

// TestEcosystem_CACerts_Update verifies that CA certs can be updated on an
// existing ecosystem (from no certs to certs, and from certs to no certs).
//
// RED: WILL NOT COMPILE — Ecosystem.CACerts does not exist yet.
func TestEcosystem_CACerts_Update(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create ecosystem without CA certs
	eco := &models.Ecosystem{Name: "eco-ca-update-test"}
	require.NoError(t, ds.CreateEcosystem(eco))

	// Verify no CA certs initially
	retrieved, err := ds.GetEcosystemByID(eco.ID)
	require.NoError(t, err)
	// RED: retrieved.CACerts does not exist yet
	assert.False(t, retrieved.CACerts.Valid,
		"initially: CACerts should be invalid (not set)")

	// Update to add CA certs
	eco.CACerts = sql.NullString{
		String: `[{"name":"corp-root-ca","vaultSecret":"corp-root-ca-pem"}]`,
		Valid:  true,
	}
	require.NoError(t, ds.UpdateEcosystem(eco))

	// Verify CA certs were stored
	afterAdd, err := ds.GetEcosystemByID(eco.ID)
	require.NoError(t, err)
	assert.True(t, afterAdd.CACerts.Valid,
		"after update: CACerts should be valid")
	assert.Contains(t, afterAdd.CACerts.String, "corp-root-ca",
		"after update: stored JSON should contain cert name")

	// Update to remove CA certs
	eco.CACerts = sql.NullString{}
	require.NoError(t, ds.UpdateEcosystem(eco))

	// Verify CA certs were removed
	afterRemove, err := ds.GetEcosystemByID(eco.ID)
	require.NoError(t, err)
	assert.False(t, afterRemove.CACerts.Valid,
		"after remove: CACerts should be invalid again")
}

// =============================================================================
// 3. TestDomain_CACerts_StoreRoundTrip
// =============================================================================

// TestDomain_CACerts_StoreRoundTrip verifies that a domain created with CA certs
// can be retrieved with the CA certs intact.
//
// RED: WILL NOT COMPILE / FAIL until:
//   - Domain.CACerts sql.NullString field exists in models
//   - ca_certs column exists in domains table (createTestSchema)
//   - store_domain.go queries include ca_certs in INSERT, UPDATE, SELECT
func TestDomain_CACerts_StoreRoundTrip(t *testing.T) {
	tests := []struct {
		name            string
		initialCACerts  sql.NullString
		wantValid       bool
		wantString      string
		updateCACerts   sql.NullString
		wantValidAfter  bool
		wantStringAfter string
	}{
		{
			name: "domain with ca certs round-trips",
			initialCACerts: sql.NullString{
				String: `[{"name":"domain-ca","vaultSecret":"domain-ca-pem","vaultEnvironment":"prod"}]`,
				Valid:  true,
			},
			wantValid:  true,
			wantString: `[{"name":"domain-ca","vaultSecret":"domain-ca-pem","vaultEnvironment":"prod"}]`,
			updateCACerts: sql.NullString{
				String: `[{"name":"domain-ca","vaultSecret":"updated-domain-ca-pem"}]`,
				Valid:  true,
			},
			wantValidAfter:  true,
			wantStringAfter: `[{"name":"domain-ca","vaultSecret":"updated-domain-ca-pem"}]`,
		},
		{
			name:           "domain without ca certs has Valid=false",
			initialCACerts: sql.NullString{},
			wantValid:      false,
			wantString:     "",
			updateCACerts: sql.NullString{
				String: `[{"name":"new-domain-ca","vaultSecret":"new-domain-ca-pem"}]`,
				Valid:  true,
			},
			wantValidAfter:  true,
			wantStringAfter: `[{"name":"new-domain-ca","vaultSecret":"new-domain-ca-pem"}]`,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := createTestDataStore(t)
			defer ds.Close()

			eco := &models.Ecosystem{Name: fmt.Sprintf("ca-eco-dom-rt-%d", i)}
			require.NoError(t, ds.CreateEcosystem(eco))

			// RED: Domain.CACerts does not exist yet
			dom := &models.Domain{
				EcosystemID: eco.ID,
				Name:        fmt.Sprintf("ca-dom-rt-%d", i),
				CACerts:     tt.initialCACerts,
			}
			require.NoError(t, ds.CreateDomain(dom))
			require.NotZero(t, dom.ID, "CreateDomain should set ID")

			// --- GetByName round-trip ---
			byName, err := ds.GetDomainByName(eco.ID, dom.Name)
			require.NoError(t, err)
			// RED: byName.CACerts does not exist yet
			assert.Equal(t, tt.wantValid, byName.CACerts.Valid,
				"GetDomainByName: CACerts.Valid mismatch")
			if tt.wantValid {
				assert.Equal(t, tt.wantString, byName.CACerts.String,
					"GetDomainByName: CACerts.String mismatch")
			}

			// --- GetByID round-trip ---
			byID, err := ds.GetDomainByID(dom.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantValid, byID.CACerts.Valid,
				"GetDomainByID: CACerts.Valid mismatch")
			if tt.wantValid {
				assert.Equal(t, tt.wantString, byID.CACerts.String,
					"GetDomainByID: CACerts.String mismatch")
			}

			// --- Update ---
			dom.CACerts = tt.updateCACerts
			require.NoError(t, ds.UpdateDomain(dom))

			// --- GetByID after update ---
			afterUpdate, err := ds.GetDomainByID(dom.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantValidAfter, afterUpdate.CACerts.Valid,
				"after UpdateDomain: CACerts.Valid mismatch")
			if tt.wantValidAfter {
				assert.Equal(t, tt.wantStringAfter, afterUpdate.CACerts.String,
					"after UpdateDomain: CACerts.String mismatch")
			}
		})
	}
}

// =============================================================================
// 4. TestDomain_CACerts_Update
// =============================================================================

// TestDomain_CACerts_Update verifies that CA certs can be updated on an
// existing domain.
//
// RED: WILL NOT COMPILE — Domain.CACerts does not exist yet.
func TestDomain_CACerts_Update(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "eco-for-domain-ca-update"}
	require.NoError(t, ds.CreateEcosystem(eco))

	// Create domain without CA certs
	dom := &models.Domain{EcosystemID: eco.ID, Name: "dom-ca-update-test"}
	require.NoError(t, ds.CreateDomain(dom))

	// Verify no CA certs initially
	retrieved, err := ds.GetDomainByID(dom.ID)
	require.NoError(t, err)
	// RED: retrieved.CACerts does not exist yet
	assert.False(t, retrieved.CACerts.Valid,
		"initially: Domain.CACerts should be invalid (not set)")

	// Update to add CA certs
	dom.CACerts = sql.NullString{
		String: `[{"name":"domain-ca","vaultSecret":"domain-ca-pem"}]`,
		Valid:  true,
	}
	require.NoError(t, ds.UpdateDomain(dom))

	// Verify CA certs were stored
	afterAdd, err := ds.GetDomainByID(dom.ID)
	require.NoError(t, err)
	assert.True(t, afterAdd.CACerts.Valid,
		"after update: Domain.CACerts should be valid")
	assert.Contains(t, afterAdd.CACerts.String, "domain-ca",
		"after update: stored JSON should contain cert name")
}

// =============================================================================
// 5. TestApp_CACerts_StoreRoundTrip
// =============================================================================

// TestApp_CACerts_StoreRoundTrip verifies that an app created with CA certs
// inside its build_config JSON can be retrieved with the CA certs intact.
//
// Note: App CA certs live inside the build_config JSON column (no schema change
// needed for apps — the column already exists). Only AppBuildConfig.CACerts
// field needs to be added to the model.
//
// RED: WILL NOT COMPILE — AppBuildConfig.CACerts does not exist yet.
func TestApp_CACerts_StoreRoundTrip(t *testing.T) {
	// The expected JSON for build_config with caCerts
	buildConfigWithCACerts := `{"dockerfile":"Dockerfile.dev","args":{"CGO_ENABLED":"0"},"caCerts":[{"name":"app-ca","vaultSecret":"app-ca-pem"}]}`
	updatedBuildConfig := `{"dockerfile":"Dockerfile.dev","caCerts":[{"name":"app-ca","vaultSecret":"updated-app-ca-pem"},{"name":"app-int-ca","vaultSecret":"app-int-ca-pem"}]}`

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
			name:            "app with ca certs in build config round-trips",
			initialConfig:   sql.NullString{String: buildConfigWithCACerts, Valid: true},
			wantValid:       true,
			wantString:      buildConfigWithCACerts,
			updateConfig:    sql.NullString{String: updatedBuildConfig, Valid: true},
			wantValidAfter:  true,
			wantStringAfter: updatedBuildConfig,
		},
		{
			name:            "app without build config has Valid=false",
			initialConfig:   sql.NullString{},
			wantValid:       false,
			wantString:      "",
			updateConfig:    sql.NullString{String: buildConfigWithCACerts, Valid: true},
			wantValidAfter:  true,
			wantStringAfter: buildConfigWithCACerts,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := createTestDataStore(t)
			defer ds.Close()

			_, dom, _ := createTestHierarchyForCACerts(t, ds, fmt.Sprintf("app-ca-rt-%d", i))

			app := &models.App{
				DomainID:    dom.ID,
				Name:        "ca-test-app",
				Path:        "/code/ca-test-app",
				BuildConfig: tt.initialConfig,
			}
			require.NoError(t, ds.CreateApp(app))
			require.NotZero(t, app.ID, "CreateApp should set ID")

			// --- GetByName round-trip ---
			byName, err := ds.GetAppByName(dom.ID, "ca-test-app")
			require.NoError(t, err)
			assert.Equal(t, tt.wantValid, byName.BuildConfig.Valid,
				"GetAppByName: BuildConfig.Valid mismatch")
			if tt.wantValid {
				assert.Equal(t, tt.wantString, byName.BuildConfig.String,
					"GetAppByName: BuildConfig.String mismatch")
			}

			// --- GetByID round-trip ---
			byID, err := ds.GetAppByID(app.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantValid, byID.BuildConfig.Valid,
				"GetAppByID: BuildConfig.Valid mismatch")
			if tt.wantValid {
				assert.Equal(t, tt.wantString, byID.BuildConfig.String,
					"GetAppByID: BuildConfig.String mismatch")
			}

			// --- Update ---
			app.BuildConfig = tt.updateConfig
			require.NoError(t, ds.UpdateApp(app))

			// --- GetByID after update ---
			afterUpdate, err := ds.GetAppByID(app.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantValidAfter, afterUpdate.BuildConfig.Valid,
				"after UpdateApp: BuildConfig.Valid mismatch")
			if tt.wantValidAfter {
				assert.Equal(t, tt.wantStringAfter, afterUpdate.BuildConfig.String,
					"after UpdateApp: BuildConfig.String mismatch")
			}
		})
	}
}

// =============================================================================
// 6. TestWorkspace_CACerts_StoreRoundTrip (regression)
// =============================================================================

// TestWorkspace_CACerts_StoreRoundTrip verifies that the existing workspace
// CA certs support (already in v0.53.0) continues to work correctly.
// This is a regression test to ensure the v0.56.0 changes don't break workspace.
//
// This test should PASS (the workspace certs are already implemented).
// It's included here to document the full regression suite.
func TestWorkspace_CACerts_StoreRoundTrip(t *testing.T) {
	// workspace DevBuildConfig JSON with caCerts
	devBuildWithCACerts := `{"caCerts":[{"name":"corp-root-ca","vaultSecret":"corp-root-ca-pem"}]}`
	updatedDevBuild := `{"caCerts":[{"name":"corp-root-ca","vaultSecret":"updated-pem"},{"name":"corp-int-ca","vaultSecret":"corp-int-ca-pem"}]}`

	tests := []struct {
		name            string
		initialConfig   sql.NullString
		wantValid       bool
		wantContains    string
		updateConfig    sql.NullString
		wantValidAfter  bool
		wantContainsAft string
	}{
		{
			name:            "workspace with ca certs round-trips",
			initialConfig:   sql.NullString{String: devBuildWithCACerts, Valid: true},
			wantValid:       true,
			wantContains:    "corp-root-ca",
			updateConfig:    sql.NullString{String: updatedDevBuild, Valid: true},
			wantValidAfter:  true,
			wantContainsAft: "corp-int-ca",
		},
		{
			name:            "workspace without ca certs has Valid=false",
			initialConfig:   sql.NullString{},
			wantValid:       false,
			wantContains:    "",
			updateConfig:    sql.NullString{String: devBuildWithCACerts, Valid: true},
			wantValidAfter:  true,
			wantContainsAft: "corp-root-ca",
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := createTestDataStore(t)
			defer ds.Close()

			_, _, app := createTestHierarchyForCACerts(t, ds, fmt.Sprintf("ws-ca-rt-%d", i))

			ws := &models.Workspace{
				AppID:       app.ID,
				Name:        "ws-ca-test",
				Slug:        fmt.Sprintf("ca-eco-ws-ca-rt-%d-ca-dom-ws-ca-rt-%d-ca-app-ws-ca-rt-%d-ws-ca-test", i, i, i),
				ImageName:   "golang:1.22",
				Status:      "stopped",
				BuildConfig: tt.initialConfig,
			}
			require.NoError(t, ds.CreateWorkspace(ws))
			require.NotZero(t, ws.ID, "CreateWorkspace should set ID")

			// --- GetByID round-trip ---
			byID, err := ds.GetWorkspaceByID(ws.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantValid, byID.BuildConfig.Valid,
				"GetWorkspaceByID: BuildConfig.Valid mismatch")
			if tt.wantValid && tt.wantContains != "" {
				assert.Contains(t, byID.BuildConfig.String, tt.wantContains,
					"GetWorkspaceByID: BuildConfig.String should contain %q", tt.wantContains)
			}

			// --- Update ---
			ws.BuildConfig = tt.updateConfig
			require.NoError(t, ds.UpdateWorkspace(ws))

			// --- GetByID after update ---
			afterUpdate, err := ds.GetWorkspaceByID(ws.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantValidAfter, afterUpdate.BuildConfig.Valid,
				"after UpdateWorkspace: BuildConfig.Valid mismatch")
			if tt.wantValidAfter && tt.wantContainsAft != "" {
				assert.Contains(t, afterUpdate.BuildConfig.String, tt.wantContainsAft,
					"after UpdateWorkspace: BuildConfig.String should contain %q", tt.wantContainsAft)
			}
		})
	}
}

// =============================================================================
// 7. TestListEcosystems_IncludesCACerts
// =============================================================================

// TestListEcosystems_IncludesCACerts verifies that ListEcosystems returns
// the CACerts field for each ecosystem (not just GetByName/GetByID).
//
// RED: WILL NOT COMPILE / FAIL — Ecosystem.CACerts and ca_certs column
// do not exist yet. Also ListEcosystems SELECT must include ca_certs.
func TestListEcosystems_IncludesCACerts(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	caCertsJSON := `[{"name":"corp-root-ca","vaultSecret":"corp-root-ca-pem"}]`

	// Create ecosystem WITH CA certs
	// RED: Ecosystem.CACerts does not exist yet
	ecoWith := &models.Ecosystem{
		Name:    "list-eco-with-ca-certs",
		CACerts: sql.NullString{String: caCertsJSON, Valid: true},
	}
	require.NoError(t, ds.CreateEcosystem(ecoWith))

	// Create ecosystem WITHOUT CA certs
	ecoWithout := &models.Ecosystem{
		Name: "list-eco-without-ca-certs",
	}
	require.NoError(t, ds.CreateEcosystem(ecoWithout))

	// List all ecosystems
	ecosystems, err := ds.ListEcosystems()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(ecosystems), 2)

	var foundWith, foundWithout *models.Ecosystem
	for _, e := range ecosystems {
		switch e.ID {
		case ecoWith.ID:
			foundWith = e
		case ecoWithout.ID:
			foundWithout = e
		}
	}

	require.NotNil(t, foundWith, "ListEcosystems should include ecosystem with CA certs")
	assert.True(t, foundWith.CACerts.Valid,
		"ListEcosystems: ecosystem with CA certs should have CACerts.Valid=true")
	assert.Equal(t, caCertsJSON, foundWith.CACerts.String,
		"ListEcosystems: ecosystem CA certs string should match")

	require.NotNil(t, foundWithout, "ListEcosystems should include ecosystem without CA certs")
	assert.False(t, foundWithout.CACerts.Valid,
		"ListEcosystems: ecosystem without CA certs should have CACerts.Valid=false")
}

// =============================================================================
// 8. TestListDomains_IncludesCACerts
// =============================================================================

// TestListDomains_IncludesCACerts verifies that ListDomainsByEcosystem and
// ListAllDomains return the CACerts field for each domain.
//
// RED: WILL NOT COMPILE / FAIL — Domain.CACerts and ca_certs column
// do not exist yet.
func TestListDomains_IncludesCACerts(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "list-dom-ca-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))

	caCertsJSON := `[{"name":"domain-ca","vaultSecret":"domain-ca-pem"}]`

	// RED: Domain.CACerts does not exist yet
	domWith := &models.Domain{
		EcosystemID: eco.ID,
		Name:        "domain-with-ca-certs",
		CACerts:     sql.NullString{String: caCertsJSON, Valid: true},
	}
	require.NoError(t, ds.CreateDomain(domWith))

	domWithout := &models.Domain{
		EcosystemID: eco.ID,
		Name:        "domain-without-ca-certs",
	}
	require.NoError(t, ds.CreateDomain(domWithout))

	t.Run("ListDomainsByEcosystem includes CACerts", func(t *testing.T) {
		domains, err := ds.ListDomainsByEcosystem(eco.ID)
		require.NoError(t, err)
		require.Len(t, domains, 2)

		var foundWith, foundWithout *models.Domain
		for _, d := range domains {
			switch d.ID {
			case domWith.ID:
				foundWith = d
			case domWithout.ID:
				foundWithout = d
			}
		}

		require.NotNil(t, foundWith)
		assert.True(t, foundWith.CACerts.Valid,
			"ListDomainsByEcosystem: domain with CA certs should have CACerts.Valid=true")
		assert.Equal(t, caCertsJSON, foundWith.CACerts.String)

		require.NotNil(t, foundWithout)
		assert.False(t, foundWithout.CACerts.Valid,
			"ListDomainsByEcosystem: domain without CA certs should have CACerts.Valid=false")
	})

	t.Run("ListAllDomains includes CACerts", func(t *testing.T) {
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

		require.NotNil(t, foundWith)
		assert.True(t, foundWith.CACerts.Valid,
			"ListAllDomains: domain with CA certs should have CACerts.Valid=true")

		require.NotNil(t, foundWithout)
		assert.False(t, foundWithout.CACerts.Valid,
			"ListAllDomains: domain without CA certs should have CACerts.Valid=false")
	})
}
