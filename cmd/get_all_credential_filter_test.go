package cmd

// =============================================================================
// TDD Phase 2 (RED): Bug #155 — filterCredentials() drops domain/app/workspace
//                    scoped credentials when filtering by ecosystem
//
// Root cause (cmd/get_all.go:754-784):
//   filterCredentials() only matches credentials whose ScopeID exactly matches
//   the scope pointer in scopeContext. When scoped to an ecosystem:
//     - sc.DomainID == nil  → domain-scoped credentials are DROPPED
//     - sc.AppID == nil     → app-scoped and workspace-scoped credentials are DROPPED
//
// Correct fix (mirrors filterApps/filterWorkspaces pattern):
//   Accept already-filtered domain/app/workspace slices, build allowed-ID maps,
//   and use those maps to include hierarchy-member credentials.
//
// Tests #1 (eco-scoped cred) already passes.
// Tests #2-#5 are RED until the fix is applied.
// Test #6 (unscoped) is a regression guard.
// =============================================================================

import (
	"bytes"
	"testing"

	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// TestFilterCredentials_HierarchyWalking
//
// All subtests share a single DataStore seeded with:
//   Ecosystem  "fc-eco"
//     Domain   "fc-dom"   (EcosystemID = fc-eco.ID)
//       App    "fc-app"   (DomainID    = fc-dom.ID)
//         Workspace "fc-ws" (AppID    = fc-app.ID)
//
// Credentials:
//   cred-eco-scope   → ScopeType=ecosystem, ScopeID=fc-eco.ID
//   cred-dom-scope   → ScopeType=domain,    ScopeID=fc-dom.ID
//   cred-app-scope   → ScopeType=app,       ScopeID=fc-app.ID
//   cred-ws-scope    → ScopeType=workspace,  ScopeID=fc-ws.ID
//
// For the "wrong ecosystem" exclusion test, a parallel hierarchy exists:
//   Ecosystem  "fc-other-eco"
//     Domain   "fc-other-dom"
//       App    "fc-other-app"
//         Workspace "fc-other-ws"
//   cred-other-eco   → ScopeType=ecosystem, ScopeID=fc-other-eco.ID
//   cred-other-dom   → ScopeType=domain,    ScopeID=fc-other-dom.ID
//   cred-other-app   → ScopeType=app,       ScopeID=fc-other-app.ID
//   cred-other-ws    → ScopeType=workspace,  ScopeID=fc-other-ws.ID
// ---------------------------------------------------------------------------

func TestFilterCredentials_HierarchyWalking(t *testing.T) {
	// ---- Arrange: shared in-memory DataStore with full hierarchy ----
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	// Primary ecosystem hierarchy
	eco := &models.Ecosystem{Name: "fc-eco"}
	require.NoError(t, dataStore.CreateEcosystem(eco))

	dom := &models.Domain{Name: "fc-dom", EcosystemID: eco.ID}
	require.NoError(t, dataStore.CreateDomain(dom))

	app := &models.App{Name: "fc-app", Path: "/fc", DomainID: dom.ID}
	require.NoError(t, dataStore.CreateApp(app))

	ws := &models.Workspace{
		Name:      "fc-ws",
		Slug:      "fc-eco/fc-dom/fc-app/fc-ws",
		AppID:     app.ID,
		ImageName: "ubuntu:22.04",
		Status:    "stopped",
	}
	require.NoError(t, dataStore.CreateWorkspace(ws))

	// Credentials scoped to each level of the primary hierarchy
	envEco := "FC_ECO_TOKEN"
	credEco := &models.CredentialDB{
		Name:      "cred-eco-scope",
		ScopeType: models.CredentialScopeEcosystem,
		ScopeID:   int64(eco.ID),
		Source:    "env",
		EnvVar:    &envEco,
	}
	require.NoError(t, dataStore.CreateCredential(credEco))

	envDom := "FC_DOM_TOKEN"
	credDom := &models.CredentialDB{
		Name:      "cred-dom-scope",
		ScopeType: models.CredentialScopeDomain,
		ScopeID:   int64(dom.ID),
		Source:    "env",
		EnvVar:    &envDom,
	}
	require.NoError(t, dataStore.CreateCredential(credDom))

	envApp := "FC_APP_TOKEN"
	credApp := &models.CredentialDB{
		Name:      "cred-app-scope",
		ScopeType: models.CredentialScopeApp,
		ScopeID:   int64(app.ID),
		Source:    "env",
		EnvVar:    &envApp,
	}
	require.NoError(t, dataStore.CreateCredential(credApp))

	envWs := "FC_WS_TOKEN"
	credWs := &models.CredentialDB{
		Name:      "cred-ws-scope",
		ScopeType: models.CredentialScopeWorkspace,
		ScopeID:   int64(ws.ID),
		Source:    "env",
		EnvVar:    &envWs,
	}
	require.NoError(t, dataStore.CreateCredential(credWs))

	// Out-of-scope parallel hierarchy
	otherEco := &models.Ecosystem{Name: "fc-other-eco"}
	require.NoError(t, dataStore.CreateEcosystem(otherEco))

	otherDom := &models.Domain{Name: "fc-other-dom", EcosystemID: otherEco.ID}
	require.NoError(t, dataStore.CreateDomain(otherDom))

	otherApp := &models.App{Name: "fc-other-app", Path: "/other", DomainID: otherDom.ID}
	require.NoError(t, dataStore.CreateApp(otherApp))

	otherWs := &models.Workspace{
		Name:      "fc-other-ws",
		Slug:      "fc-other-eco/fc-other-dom/fc-other-app/fc-other-ws",
		AppID:     otherApp.ID,
		ImageName: "ubuntu:22.04",
		Status:    "stopped",
	}
	require.NoError(t, dataStore.CreateWorkspace(otherWs))

	envOtherEco := "FC_OTHER_ECO_TOKEN"
	credOtherEco := &models.CredentialDB{
		Name:      "cred-other-eco",
		ScopeType: models.CredentialScopeEcosystem,
		ScopeID:   int64(otherEco.ID),
		Source:    "env",
		EnvVar:    &envOtherEco,
	}
	require.NoError(t, dataStore.CreateCredential(credOtherEco))

	envOtherDom := "FC_OTHER_DOM_TOKEN"
	credOtherDom := &models.CredentialDB{
		Name:      "cred-other-dom",
		ScopeType: models.CredentialScopeDomain,
		ScopeID:   int64(otherDom.ID),
		Source:    "env",
		EnvVar:    &envOtherDom,
	}
	require.NoError(t, dataStore.CreateCredential(credOtherDom))

	envOtherApp := "FC_OTHER_APP_TOKEN"
	credOtherApp := &models.CredentialDB{
		Name:      "cred-other-app",
		ScopeType: models.CredentialScopeApp,
		ScopeID:   int64(otherApp.ID),
		Source:    "env",
		EnvVar:    &envOtherApp,
	}
	require.NoError(t, dataStore.CreateCredential(credOtherApp))

	envOtherWs := "FC_OTHER_WS_TOKEN"
	credOtherWs := &models.CredentialDB{
		Name:      "cred-other-ws",
		ScopeType: models.CredentialScopeWorkspace,
		ScopeID:   int64(otherWs.ID),
		Source:    "env",
		EnvVar:    &envOtherWs,
	}
	require.NoError(t, dataStore.CreateCredential(credOtherWs))

	// ---- Helper: run getAll scoped to fc-eco and return output string ----
	runScoped := func(t *testing.T) string {
		t.Helper()
		cmd := newScopedGetAllCmd(t, dataStore)

		var buf bytes.Buffer
		origWriter := render.GetWriter()
		render.SetWriter(&buf)
		defer render.SetWriter(origWriter)

		origFormat := getOutputFormat
		defer func() { getOutputFormat = origFormat }()
		getOutputFormat = "" // human-readable table output

		cmd.SetArgs([]string{"--ecosystem", "fc-eco"})
		err := cmd.Execute()
		require.NoError(t, err)
		return buf.String()
	}

	// ---------------------------------------------------------------------------
	// Test 1: Ecosystem-scoped credential is included when filtering by ecosystem
	// GREEN — should already pass with the current implementation.
	// ---------------------------------------------------------------------------
	t.Run("ecosystem_scoped_credential_is_included", func(t *testing.T) {
		output := runScoped(t)
		assert.Contains(t, output, "cred-eco-scope",
			"Bug #155: ecosystem-scoped credential should appear when scoping to that ecosystem")
	})

	// ---------------------------------------------------------------------------
	// Test 2: Domain-scoped credential is included when the domain belongs to the
	//         scoped ecosystem.
	// RED — fails because filterCredentials() gates on sc.DomainID != nil, but
	//        when scoping by ecosystem only, sc.DomainID is nil.
	// ---------------------------------------------------------------------------
	t.Run("domain_scoped_credential_included_for_ecosystem_scope", func(t *testing.T) {
		output := runScoped(t)
		assert.Contains(t, output, "cred-dom-scope",
			"Bug #155: domain-scoped credential should appear when the domain belongs to the scoped ecosystem (filterCredentials must walk the hierarchy)")
	})

	// ---------------------------------------------------------------------------
	// Test 3: App-scoped credential is included when the app belongs to a domain
	//         in the scoped ecosystem.
	// RED — fails because filterCredentials() gates on sc.AppID != nil, but when
	//        scoping by ecosystem only, sc.AppID is nil.
	// ---------------------------------------------------------------------------
	t.Run("app_scoped_credential_included_for_ecosystem_scope", func(t *testing.T) {
		output := runScoped(t)
		assert.Contains(t, output, "cred-app-scope",
			"Bug #155: app-scoped credential should appear when the app belongs to a domain in the scoped ecosystem (filterCredentials must walk the hierarchy)")
	})

	// ---------------------------------------------------------------------------
	// Test 4: Workspace-scoped credential is included when the workspace belongs
	//         to an app in the scoped ecosystem.
	// RED — fails because filterCredentials() gates on sc.AppID != nil, and the
	//        workspace-scope branch has the same problem.
	// ---------------------------------------------------------------------------
	t.Run("workspace_scoped_credential_included_for_ecosystem_scope", func(t *testing.T) {
		output := runScoped(t)
		assert.Contains(t, output, "cred-ws-scope",
			"Bug #155: workspace-scoped credential should appear when the workspace belongs to the scoped ecosystem (filterCredentials must walk the hierarchy)")
	})

	// ---------------------------------------------------------------------------
	// Test 5: Credentials from a DIFFERENT ecosystem are excluded.
	// This verifies that the fix doesn't over-include: other-ecosystem credentials
	// must NOT appear.
	// RED — with the current broken implementation, other-ecosystem domain/app/ws
	//        credentials are also dropped (so this may incidentally "pass" today),
	//        but after the fix the exclusion must still hold.
	// ---------------------------------------------------------------------------
	t.Run("other_ecosystem_credentials_are_excluded", func(t *testing.T) {
		output := runScoped(t)
		assert.NotContains(t, output, "cred-other-eco",
			"Bug #155: ecosystem-scoped credential for OTHER ecosystem must NOT appear")
		assert.NotContains(t, output, "cred-other-dom",
			"Bug #155: domain-scoped credential in OTHER ecosystem must NOT appear")
		assert.NotContains(t, output, "cred-other-app",
			"Bug #155: app-scoped credential in OTHER ecosystem must NOT appear")
		assert.NotContains(t, output, "cred-other-ws",
			"Bug #155: workspace-scoped credential in OTHER ecosystem must NOT appear")
	})
}

// ---------------------------------------------------------------------------
// TestFilterCredentials_Unscoped_ShowsAllCredentials
//
// GREEN regression guard: when no scope is active (ShowAll=true), all
// credentials must be returned regardless of their scope type.
// ---------------------------------------------------------------------------
func TestFilterCredentials_Unscoped_ShowsAllCredentials(t *testing.T) {
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	eco := &models.Ecosystem{Name: "fc-unscoped-eco"}
	require.NoError(t, dataStore.CreateEcosystem(eco))

	dom := &models.Domain{Name: "fc-unscoped-dom", EcosystemID: eco.ID}
	require.NoError(t, dataStore.CreateDomain(dom))

	app := &models.App{Name: "fc-unscoped-app", Path: "/u", DomainID: dom.ID}
	require.NoError(t, dataStore.CreateApp(app))

	ws := &models.Workspace{
		Name:      "fc-unscoped-ws",
		Slug:      "fc-unscoped-eco/fc-unscoped-dom/fc-unscoped-app/fc-unscoped-ws",
		AppID:     app.ID,
		ImageName: "ubuntu:22.04",
		Status:    "stopped",
	}
	require.NoError(t, dataStore.CreateWorkspace(ws))

	envEco := "FC_UNSCOPED_ECO"
	require.NoError(t, dataStore.CreateCredential(&models.CredentialDB{
		Name:      "fc-unscoped-cred-eco",
		ScopeType: models.CredentialScopeEcosystem,
		ScopeID:   int64(eco.ID),
		Source:    "env",
		EnvVar:    &envEco,
	}))

	envDom := "FC_UNSCOPED_DOM"
	require.NoError(t, dataStore.CreateCredential(&models.CredentialDB{
		Name:      "fc-unscoped-cred-dom",
		ScopeType: models.CredentialScopeDomain,
		ScopeID:   int64(dom.ID),
		Source:    "env",
		EnvVar:    &envDom,
	}))

	envApp := "FC_UNSCOPED_APP"
	require.NoError(t, dataStore.CreateCredential(&models.CredentialDB{
		Name:      "fc-unscoped-cred-app",
		ScopeType: models.CredentialScopeApp,
		ScopeID:   int64(app.ID),
		Source:    "env",
		EnvVar:    &envApp,
	}))

	envWs := "FC_UNSCOPED_WS"
	require.NoError(t, dataStore.CreateCredential(&models.CredentialDB{
		Name:      "fc-unscoped-cred-ws",
		ScopeType: models.CredentialScopeWorkspace,
		ScopeID:   int64(ws.ID),
		Source:    "env",
		EnvVar:    &envWs,
	}))

	// Run with -A (ShowAll=true) — no scope filtering applied
	cmd := newScopedGetAllCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = ""

	cmd.SetArgs([]string{"-A"})
	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "fc-unscoped-cred-eco",
		"unscoped (-A) output must include ecosystem-scoped credentials")
	assert.Contains(t, output, "fc-unscoped-cred-dom",
		"unscoped (-A) output must include domain-scoped credentials")
	assert.Contains(t, output, "fc-unscoped-cred-app",
		"unscoped (-A) output must include app-scoped credentials")
	assert.Contains(t, output, "fc-unscoped-cred-ws",
		"unscoped (-A) output must include workspace-scoped credentials")
}

// ---------------------------------------------------------------------------
// TestFilterCredentials_DomainScoped_IncludesAllLevels
//
// When scoped to a specific domain (--ecosystem X --domain Y), credentials
// attached to apps and workspaces WITHIN that domain must still be included.
//
// RED — same root cause: sc.AppID is nil when only --domain is set, so
//
//	app-scoped and workspace-scoped credentials within the domain are dropped.
//
// ---------------------------------------------------------------------------
func TestFilterCredentials_DomainScoped_IncludesAllLevels(t *testing.T) {
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	eco := &models.Ecosystem{Name: "fc-domscope-eco"}
	require.NoError(t, dataStore.CreateEcosystem(eco))

	dom := &models.Domain{Name: "fc-domscope-dom", EcosystemID: eco.ID}
	require.NoError(t, dataStore.CreateDomain(dom))

	app := &models.App{Name: "fc-domscope-app", Path: "/ds", DomainID: dom.ID}
	require.NoError(t, dataStore.CreateApp(app))

	ws := &models.Workspace{
		Name:      "fc-domscope-ws",
		Slug:      "fc-domscope-eco/fc-domscope-dom/fc-domscope-app/fc-domscope-ws",
		AppID:     app.ID,
		ImageName: "ubuntu:22.04",
		Status:    "stopped",
	}
	require.NoError(t, dataStore.CreateWorkspace(ws))

	// Domain-scoped credential (attached directly to the domain)
	envDom := "FC_DS_DOM"
	require.NoError(t, dataStore.CreateCredential(&models.CredentialDB{
		Name:      "fc-ds-cred-dom",
		ScopeType: models.CredentialScopeDomain,
		ScopeID:   int64(dom.ID),
		Source:    "env",
		EnvVar:    &envDom,
	}))

	// App-scoped credential (attached to app within the scoped domain)
	envApp := "FC_DS_APP"
	require.NoError(t, dataStore.CreateCredential(&models.CredentialDB{
		Name:      "fc-ds-cred-app",
		ScopeType: models.CredentialScopeApp,
		ScopeID:   int64(app.ID),
		Source:    "env",
		EnvVar:    &envApp,
	}))

	// Workspace-scoped credential (attached to workspace within the scoped domain)
	envWs := "FC_DS_WS"
	require.NoError(t, dataStore.CreateCredential(&models.CredentialDB{
		Name:      "fc-ds-cred-ws",
		ScopeType: models.CredentialScopeWorkspace,
		ScopeID:   int64(ws.ID),
		Source:    "env",
		EnvVar:    &envWs,
	}))

	cmd := newScopedGetAllCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = ""

	// Scope to the specific domain
	cmd.SetArgs([]string{"--ecosystem", "fc-domscope-eco", "--domain", "fc-domscope-dom"})
	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "fc-ds-cred-dom",
		"Bug #155: domain-scoped credential should appear when scoping to that domain")
	assert.Contains(t, output, "fc-ds-cred-app",
		"Bug #155: app-scoped credential should appear when the app belongs to the scoped domain (filterCredentials must walk the hierarchy)")
	assert.Contains(t, output, "fc-ds-cred-ws",
		"Bug #155: workspace-scoped credential should appear when the workspace belongs to the scoped domain (filterCredentials must walk the hierarchy)")
}
