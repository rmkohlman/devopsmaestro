package db

// =============================================================================
// Regression Tests: Delete Active Context Entity + Credential Cleanup
// =============================================================================
//
// These tests verify the fix for the critical bug where deleting an ecosystem,
// domain, app, or workspace that was currently set as the active context failed
// with "FOREIGN KEY constraint failed".
//
// Fix involved:
//   1. Migration 016: context table rebuilt with ON DELETE SET NULL on all 4
//      FK references (active_workspace_id, active_ecosystem_id, active_domain_id,
//      active_app_id).
//   2. Credential cleanup: each Delete* method now cleans up orphaned credentials
//      before deleting the entity.

import (
	"devopsmaestro/models"
	"fmt"
	"testing"
)

// =============================================================================
// Helpers
// =============================================================================

// createCascadeTestDataStore creates an in-memory SQLite DB with ON DELETE CASCADE
// on apps.domain_id and workspaces.app_id (matching the production schema) so
// that DeleteDomain / DeleteApp can be called even when child rows exist.
// The context table is created with ON DELETE SET NULL to match migration 016.
func createCascadeTestDataStore(t *testing.T) *SQLDataStore {
	t.Helper()

	cfg := DriverConfig{Type: DriverMemory}
	driver, err := NewMemorySQLiteDriver(cfg)
	if err != nil {
		t.Fatalf("Failed to create test driver: %v", err)
	}
	if err := driver.Connect(); err != nil {
		t.Fatalf("Failed to connect test driver: %v", err)
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS ecosystems (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			theme TEXT,
			build_args TEXT,
			ca_certs TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// domains cascade-delete when the parent ecosystem is deleted
		`CREATE TABLE IF NOT EXISTS domains (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ecosystem_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			theme TEXT,
			build_args TEXT,
			ca_certs TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (ecosystem_id) REFERENCES ecosystems(id) ON DELETE CASCADE,
			UNIQUE(ecosystem_id, name)
		)`,
		// apps cascade-delete when the parent domain is deleted (matches production)
		`CREATE TABLE IF NOT EXISTS apps (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			domain_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			path TEXT NOT NULL,
			description TEXT,
			theme TEXT,
			language TEXT,
			build_config TEXT,
			git_repo_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (domain_id) REFERENCES domains(id) ON DELETE CASCADE,
			UNIQUE(domain_id, name)
		)`,
		// workspaces cascade-delete when the parent app is deleted (matches production)
		`CREATE TABLE IF NOT EXISTS workspaces (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			app_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			image_name TEXT,
			container_id TEXT,
			status TEXT DEFAULT 'stopped',
			nvim_structure TEXT,
			nvim_plugins TEXT,
			theme TEXT,
			terminal_prompt TEXT,
			terminal_plugins TEXT,
			terminal_package TEXT,
			slug TEXT NOT NULL UNIQUE,
			ssh_agent_forwarding INTEGER DEFAULT 0,
			git_repo_id INTEGER,
			env TEXT NOT NULL DEFAULT '{}',
			build_config TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE,
			UNIQUE(app_id, name)
		)`,
		// context with ON DELETE SET NULL (migration 016 fix)
		`CREATE TABLE IF NOT EXISTS context (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			active_ecosystem_id INTEGER,
			active_domain_id INTEGER,
			active_app_id INTEGER,
			active_workspace_id INTEGER,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (active_ecosystem_id) REFERENCES ecosystems(id) ON DELETE SET NULL,
			FOREIGN KEY (active_domain_id) REFERENCES domains(id) ON DELETE SET NULL,
			FOREIGN KEY (active_app_id) REFERENCES apps(id) ON DELETE SET NULL,
			FOREIGN KEY (active_workspace_id) REFERENCES workspaces(id) ON DELETE SET NULL
		)`,
		`CREATE TABLE IF NOT EXISTS credentials (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			scope_type TEXT NOT NULL CHECK(scope_type IN ('ecosystem', 'domain', 'app', 'workspace')),
			scope_id INTEGER,
			name TEXT NOT NULL,
			source TEXT NOT NULL CHECK(source IN ('vault', 'env')),
			vault_secret TEXT,
			vault_env TEXT,
			vault_username_secret TEXT,
			vault_fields TEXT,
			env_var TEXT,
			description TEXT,
			username_var TEXT,
			password_var TEXT,
			expires_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(scope_type, scope_id, name)
		)`,
		// git_repos stub — required for apps/workspaces git_repo_id FK
		`CREATE TABLE IF NOT EXISTS git_repos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			url TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			default_ref TEXT NOT NULL DEFAULT 'main',
			auth_type TEXT NOT NULL CHECK(auth_type IN ('none', 'ssh', 'token')),
			credential_id INTEGER,
			auto_sync BOOLEAN NOT NULL DEFAULT 0,
			sync_interval_minutes INTEGER NOT NULL DEFAULT 0,
			last_synced_at DATETIME,
			sync_status TEXT NOT NULL DEFAULT 'pending',
			sync_error TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`INSERT OR IGNORE INTO context (id) VALUES (1)`,
	}

	for _, q := range queries {
		if _, err := driver.Execute(q); err != nil {
			driver.Close()
			t.Fatalf("Failed to create cascade test schema: %v\nQuery: %s", err, q)
		}
	}

	return NewSQLDataStore(driver, nil)
}

// createEcosystemCred creates a minimal env-sourced credential scoped to the
// given ecosystem ID and returns it (with ID populated).
func createEcosystemCred(t *testing.T, ds *SQLDataStore, ecoID int, name string) *models.CredentialDB {
	t.Helper()
	envVal := "ECO_ENV_VAR_" + name
	cred := &models.CredentialDB{
		ScopeType: models.CredentialScopeEcosystem,
		ScopeID:   int64(ecoID),
		Name:      name,
		Source:    "env",
		EnvVar:    strPtr(envVal),
	}
	if err := ds.CreateCredential(cred); err != nil {
		t.Fatalf("setup: CreateCredential(%q) error = %v", name, err)
	}
	return cred
}

// createDomainCred creates a minimal env-sourced credential scoped to a domain.
func createDomainCred(t *testing.T, ds *SQLDataStore, domainID int, name string) *models.CredentialDB {
	t.Helper()
	envVal := "DOMAIN_ENV_VAR_" + name
	cred := &models.CredentialDB{
		ScopeType: models.CredentialScopeDomain,
		ScopeID:   int64(domainID),
		Name:      name,
		Source:    "env",
		EnvVar:    strPtr(envVal),
	}
	if err := ds.CreateCredential(cred); err != nil {
		t.Fatalf("setup: CreateCredential(%q) error = %v", name, err)
	}
	return cred
}

// createAppCred creates a minimal env-sourced credential scoped to an app.
func createAppCred(t *testing.T, ds *SQLDataStore, appID int, name string) *models.CredentialDB {
	t.Helper()
	envVal := "APP_ENV_VAR_" + name
	cred := &models.CredentialDB{
		ScopeType: models.CredentialScopeApp,
		ScopeID:   int64(appID),
		Name:      name,
		Source:    "env",
		EnvVar:    strPtr(envVal),
	}
	if err := ds.CreateCredential(cred); err != nil {
		t.Fatalf("setup: CreateCredential(%q) error = %v", name, err)
	}
	return cred
}

// createWorkspaceCred creates a minimal env-sourced credential scoped to a workspace.
func createWorkspaceCred(t *testing.T, ds *SQLDataStore, wsID int, name string) *models.CredentialDB {
	t.Helper()
	envVal := "WS_ENV_VAR_" + name
	cred := &models.CredentialDB{
		ScopeType: models.CredentialScopeWorkspace,
		ScopeID:   int64(wsID),
		Name:      name,
		Source:    "env",
		EnvVar:    strPtr(envVal),
	}
	if err := ds.CreateCredential(cred); err != nil {
		t.Fatalf("setup: CreateCredential(%q) error = %v", name, err)
	}
	return cred
}

// createTestWorkspace builds a workspace under the given app and returns it.
func createTestWorkspace(t *testing.T, ds *SQLDataStore, appID int, suffix string) *models.Workspace {
	t.Helper()
	ws := &models.Workspace{
		AppID:     appID,
		Name:      "ws-" + suffix,
		Slug:      fmt.Sprintf("test-eco-test-domain-test-app-%s-ws-%s", suffix, suffix),
		ImageName: "test:latest",
		Status:    "stopped",
	}
	if err := ds.CreateWorkspace(ws); err != nil {
		t.Fatalf("setup: CreateWorkspace(%q) error = %v", ws.Name, err)
	}
	return ws
}

// =============================================================================
// Context null-out tests — regression for "FOREIGN KEY constraint failed" bug
// =============================================================================

// TestDeleteEcosystem_WhenActiveContext verifies that deleting an ecosystem
// that is currently set as the active context:
//   - does NOT fail with a FOREIGN KEY constraint error
//   - sets context.active_ecosystem_id to NULL
//   - removes the ecosystem from the database
func TestDeleteEcosystem_WhenActiveContext(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Arrange: create ecosystem and set it as the active context
	eco := &models.Ecosystem{Name: "active-eco-delete"}
	if err := ds.CreateEcosystem(eco); err != nil {
		t.Fatalf("setup: CreateEcosystem() error = %v", err)
	}
	if err := ds.SetActiveEcosystem(&eco.ID); err != nil {
		t.Fatalf("setup: SetActiveEcosystem() error = %v", err)
	}

	// Confirm it is the active ecosystem before deleting
	ctx, err := ds.GetContext()
	if err != nil {
		t.Fatalf("setup: GetContext() error = %v", err)
	}
	if ctx.ActiveEcosystemID == nil || *ctx.ActiveEcosystemID != eco.ID {
		t.Fatalf("setup: expected ActiveEcosystemID=%d, got %v", eco.ID, ctx.ActiveEcosystemID)
	}

	// Act: delete the active ecosystem
	if err := ds.DeleteEcosystem("active-eco-delete"); err != nil {
		t.Fatalf("DeleteEcosystem() error = %v (expected no FK constraint error)", err)
	}

	// Assert: context's active_ecosystem_id is now NULL
	ctx, err = ds.GetContext()
	if err != nil {
		t.Fatalf("GetContext() after delete error = %v", err)
	}
	if ctx.ActiveEcosystemID != nil {
		t.Errorf("GetContext().ActiveEcosystemID = %v, want nil (ON DELETE SET NULL)", ctx.ActiveEcosystemID)
	}

	// Assert: ecosystem is gone from the DB
	_, err = ds.GetEcosystemByName("active-eco-delete")
	if err == nil {
		t.Errorf("GetEcosystemByName() should return error after delete, got nil")
	}
}

// TestDeleteDomain_WhenActiveContext verifies that deleting a domain that is
// currently set as the active context succeeds and nulls out the context field.
func TestDeleteDomain_WhenActiveContext(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Arrange
	eco := &models.Ecosystem{Name: "active-domain-eco"}
	if err := ds.CreateEcosystem(eco); err != nil {
		t.Fatalf("setup: CreateEcosystem() error = %v", err)
	}
	domain := &models.Domain{EcosystemID: eco.ID, Name: "active-domain-delete"}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("setup: CreateDomain() error = %v", err)
	}
	if err := ds.SetActiveDomain(&domain.ID); err != nil {
		t.Fatalf("setup: SetActiveDomain() error = %v", err)
	}

	// Confirm active context
	ctx, err := ds.GetContext()
	if err != nil {
		t.Fatalf("setup: GetContext() error = %v", err)
	}
	if ctx.ActiveDomainID == nil || *ctx.ActiveDomainID != domain.ID {
		t.Fatalf("setup: expected ActiveDomainID=%d, got %v", domain.ID, ctx.ActiveDomainID)
	}

	// Act: delete the active domain
	if err := ds.DeleteDomain(domain.ID); err != nil {
		t.Fatalf("DeleteDomain() error = %v (expected no FK constraint error)", err)
	}

	// Assert: context's active_domain_id is now NULL
	ctx, err = ds.GetContext()
	if err != nil {
		t.Fatalf("GetContext() after delete error = %v", err)
	}
	if ctx.ActiveDomainID != nil {
		t.Errorf("GetContext().ActiveDomainID = %v, want nil (ON DELETE SET NULL)", ctx.ActiveDomainID)
	}

	// Assert: domain is gone from the DB
	_, err = ds.GetDomainByID(domain.ID)
	if err == nil {
		t.Errorf("GetDomainByID() should return error after delete, got nil")
	}
}

// TestDeleteApp_WhenActiveContext verifies that deleting an app that is
// currently set as the active context succeeds and nulls out the context field.
func TestDeleteApp_WhenActiveContext(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Arrange: build full hierarchy
	eco := &models.Ecosystem{Name: "active-app-eco"}
	if err := ds.CreateEcosystem(eco); err != nil {
		t.Fatalf("setup: CreateEcosystem() error = %v", err)
	}
	domain := &models.Domain{EcosystemID: eco.ID, Name: "active-app-domain"}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("setup: CreateDomain() error = %v", err)
	}
	app := &models.App{DomainID: domain.ID, Name: "active-app-delete", Path: "/tmp/active-app"}
	if err := ds.CreateApp(app); err != nil {
		t.Fatalf("setup: CreateApp() error = %v", err)
	}
	if err := ds.SetActiveApp(&app.ID); err != nil {
		t.Fatalf("setup: SetActiveApp() error = %v", err)
	}

	// Confirm active context
	ctx, err := ds.GetContext()
	if err != nil {
		t.Fatalf("setup: GetContext() error = %v", err)
	}
	if ctx.ActiveAppID == nil || *ctx.ActiveAppID != app.ID {
		t.Fatalf("setup: expected ActiveAppID=%d, got %v", app.ID, ctx.ActiveAppID)
	}

	// Act: delete the active app
	if err := ds.DeleteApp(app.ID); err != nil {
		t.Fatalf("DeleteApp() error = %v (expected no FK constraint error)", err)
	}

	// Assert: context's active_app_id is now NULL
	ctx, err = ds.GetContext()
	if err != nil {
		t.Fatalf("GetContext() after delete error = %v", err)
	}
	if ctx.ActiveAppID != nil {
		t.Errorf("GetContext().ActiveAppID = %v, want nil (ON DELETE SET NULL)", ctx.ActiveAppID)
	}

	// Assert: app is gone from the DB
	_, err = ds.GetAppByID(app.ID)
	if err == nil {
		t.Errorf("GetAppByID() should return error after delete, got nil")
	}
}

// TestDeleteWorkspace_WhenActiveContext verifies that deleting a workspace that
// is currently set as the active context succeeds and nulls out the context field.
func TestDeleteWorkspace_WhenActiveContext(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Arrange: build full hierarchy
	eco := &models.Ecosystem{Name: "active-ws-eco"}
	if err := ds.CreateEcosystem(eco); err != nil {
		t.Fatalf("setup: CreateEcosystem() error = %v", err)
	}
	domain := &models.Domain{EcosystemID: eco.ID, Name: "active-ws-domain"}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("setup: CreateDomain() error = %v", err)
	}
	app := &models.App{DomainID: domain.ID, Name: "active-ws-app", Path: "/tmp/active-ws-app"}
	if err := ds.CreateApp(app); err != nil {
		t.Fatalf("setup: CreateApp() error = %v", err)
	}
	ws := &models.Workspace{
		AppID:     app.ID,
		Name:      "active-ws-delete",
		Slug:      "active-ws-eco-active-ws-domain-active-ws-app-active-ws-delete",
		ImageName: "test:latest",
		Status:    "stopped",
	}
	if err := ds.CreateWorkspace(ws); err != nil {
		t.Fatalf("setup: CreateWorkspace() error = %v", err)
	}
	if err := ds.SetActiveWorkspace(&ws.ID); err != nil {
		t.Fatalf("setup: SetActiveWorkspace() error = %v", err)
	}

	// Confirm active context
	ctx, err := ds.GetContext()
	if err != nil {
		t.Fatalf("setup: GetContext() error = %v", err)
	}
	if ctx.ActiveWorkspaceID == nil || *ctx.ActiveWorkspaceID != ws.ID {
		t.Fatalf("setup: expected ActiveWorkspaceID=%d, got %v", ws.ID, ctx.ActiveWorkspaceID)
	}

	// Act: delete the active workspace
	if err := ds.DeleteWorkspace(ws.ID); err != nil {
		t.Fatalf("DeleteWorkspace() error = %v (expected no FK constraint error)", err)
	}

	// Assert: context's active_workspace_id is now NULL
	ctx, err = ds.GetContext()
	if err != nil {
		t.Fatalf("GetContext() after delete error = %v", err)
	}
	if ctx.ActiveWorkspaceID != nil {
		t.Errorf("GetContext().ActiveWorkspaceID = %v, want nil (ON DELETE SET NULL)", ctx.ActiveWorkspaceID)
	}

	// Assert: workspace is gone from the DB
	_, err = ds.GetWorkspaceByID(ws.ID)
	if err == nil {
		t.Errorf("GetWorkspaceByID() should return error after delete, got nil")
	}
}

// =============================================================================
// Credential cleanup tests
// =============================================================================

// TestDeleteEcosystem_CleansUpCredentials verifies that deleting an ecosystem
// removes all credentials whose scope_type='ecosystem' and scope_id matches,
// preventing orphaned credential rows.
func TestDeleteEcosystem_CleansUpCredentials(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Arrange: create ecosystem with two ecosystem-scoped credentials
	eco := &models.Ecosystem{Name: "cred-cleanup-eco"}
	if err := ds.CreateEcosystem(eco); err != nil {
		t.Fatalf("setup: CreateEcosystem() error = %v", err)
	}
	createEcosystemCred(t, ds, eco.ID, "cred-one")
	createEcosystemCred(t, ds, eco.ID, "cred-two")

	// Confirm 2 credentials exist
	creds, err := ds.ListCredentialsByScope(models.CredentialScopeEcosystem, int64(eco.ID))
	if err != nil {
		t.Fatalf("setup: ListCredentialsByScope() error = %v", err)
	}
	if len(creds) != 2 {
		t.Fatalf("setup: expected 2 credentials, got %d", len(creds))
	}

	// Act: delete the ecosystem
	if err := ds.DeleteEcosystem("cred-cleanup-eco"); err != nil {
		t.Fatalf("DeleteEcosystem() error = %v", err)
	}

	// Assert: credentials are gone — querying by the now-deleted scope should
	// return an empty list (no error, just nothing left).
	remaining, err := ds.ListCredentialsByScope(models.CredentialScopeEcosystem, int64(eco.ID))
	if err != nil {
		t.Fatalf("ListCredentialsByScope() after delete error = %v", err)
	}
	if len(remaining) != 0 {
		t.Errorf("expected 0 credentials after ecosystem delete, got %d", len(remaining))
	}
}

// TestDeleteApp_CleansUpWorkspaceAndAppCredentials verifies that deleting an app
// removes both app-scoped credentials AND workspace-scoped credentials for
// workspaces that belonged to that app.
//
// Uses createCascadeTestDataStore (ON DELETE CASCADE on workspaces.app_id) so
// that the workspace row is automatically removed when the app is deleted,
// matching the production schema behaviour.
func TestDeleteApp_CleansUpWorkspaceAndAppCredentials(t *testing.T) {
	ds := createCascadeTestDataStore(t)
	defer ds.Close()

	// Arrange: build hierarchy with credentials at app and workspace level
	eco := &models.Ecosystem{Name: "app-cred-cleanup-eco"}
	if err := ds.CreateEcosystem(eco); err != nil {
		t.Fatalf("setup: CreateEcosystem() error = %v", err)
	}
	domain := &models.Domain{EcosystemID: eco.ID, Name: "app-cred-cleanup-domain"}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("setup: CreateDomain() error = %v", err)
	}
	app := &models.App{DomainID: domain.ID, Name: "app-cred-cleanup-app", Path: "/tmp/app-cred-cleanup"}
	if err := ds.CreateApp(app); err != nil {
		t.Fatalf("setup: CreateApp() error = %v", err)
	}

	// Create a workspace under the app and add credentials at both scopes
	ws := createTestWorkspace(t, ds, app.ID, "cred-cleanup")
	createAppCred(t, ds, app.ID, "app-cred")
	createWorkspaceCred(t, ds, ws.ID, "ws-cred")

	// Confirm credentials exist before deletion
	appCreds, err := ds.ListCredentialsByScope(models.CredentialScopeApp, int64(app.ID))
	if err != nil {
		t.Fatalf("setup: ListCredentialsByScope(app) error = %v", err)
	}
	if len(appCreds) != 1 {
		t.Fatalf("setup: expected 1 app credential, got %d", len(appCreds))
	}
	wsCreds, err := ds.ListCredentialsByScope(models.CredentialScopeWorkspace, int64(ws.ID))
	if err != nil {
		t.Fatalf("setup: ListCredentialsByScope(workspace) error = %v", err)
	}
	if len(wsCreds) != 1 {
		t.Fatalf("setup: expected 1 workspace credential, got %d", len(wsCreds))
	}

	// Act: delete the app — the credential cleanup SQL runs first, then the
	// app row is deleted; the workspace row is cascade-deleted by SQLite.
	if err := ds.DeleteApp(app.ID); err != nil {
		t.Fatalf("DeleteApp() error = %v", err)
	}

	// Assert: all credentials scoped to this app or its former workspace are gone
	allCreds, err := ds.ListAllCredentials()
	if err != nil {
		t.Fatalf("ListAllCredentials() after delete error = %v", err)
	}
	for _, c := range allCreds {
		if (c.ScopeType == models.CredentialScopeApp && c.ScopeID == int64(app.ID)) ||
			(c.ScopeType == models.CredentialScopeWorkspace && c.ScopeID == int64(ws.ID)) {
			t.Errorf("found unexpected orphaned credential: scope_type=%s scope_id=%d name=%q",
				c.ScopeType, c.ScopeID, c.Name)
		}
	}
}

// TestDeleteDomain_CleansUpHierarchyCredentials verifies that deleting a domain
// removes credentials scoped to:
//   - the domain itself
//   - apps within the domain
//   - workspaces within those apps
//
// Uses createCascadeTestDataStore (ON DELETE CASCADE on apps.domain_id and
// workspaces.app_id) so that the child rows are automatically removed when the
// domain is deleted, matching the production schema behaviour.
func TestDeleteDomain_CleansUpHierarchyCredentials(t *testing.T) {
	ds := createCascadeTestDataStore(t)
	defer ds.Close()

	// Arrange: build full hierarchy with credentials at every level
	eco := &models.Ecosystem{Name: "domain-cred-cleanup-eco"}
	if err := ds.CreateEcosystem(eco); err != nil {
		t.Fatalf("setup: CreateEcosystem() error = %v", err)
	}
	domain := &models.Domain{EcosystemID: eco.ID, Name: "domain-cred-cleanup"}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("setup: CreateDomain() error = %v", err)
	}

	// Two apps, each with a workspace
	app1 := &models.App{DomainID: domain.ID, Name: "domain-cred-app1", Path: "/tmp/app1"}
	if err := ds.CreateApp(app1); err != nil {
		t.Fatalf("setup: CreateApp(app1) error = %v", err)
	}
	app2 := &models.App{DomainID: domain.ID, Name: "domain-cred-app2", Path: "/tmp/app2"}
	if err := ds.CreateApp(app2); err != nil {
		t.Fatalf("setup: CreateApp(app2) error = %v", err)
	}

	ws1 := createTestWorkspace(t, ds, app1.ID, "domain-cleanup-1")
	ws2 := createTestWorkspace(t, ds, app2.ID, "domain-cleanup-2")

	// Create credentials at every scope level
	createDomainCred(t, ds, domain.ID, "domain-level-cred")
	createAppCred(t, ds, app1.ID, "app1-cred")
	createAppCred(t, ds, app2.ID, "app2-cred")
	createWorkspaceCred(t, ds, ws1.ID, "ws1-cred")
	createWorkspaceCred(t, ds, ws2.ID, "ws2-cred")

	// Sanity check: 5 total credentials before deletion
	allBefore, err := ds.ListAllCredentials()
	if err != nil {
		t.Fatalf("setup: ListAllCredentials() error = %v", err)
	}
	if len(allBefore) != 5 {
		t.Fatalf("setup: expected 5 credentials, got %d", len(allBefore))
	}

	// Act: delete the domain — credential cleanup runs first, then the domain
	// row is deleted; apps and workspaces are cascade-deleted by SQLite.
	if err := ds.DeleteDomain(domain.ID); err != nil {
		t.Fatalf("DeleteDomain() error = %v", err)
	}

	// Assert: all 5 credentials are gone
	allAfter, err := ds.ListAllCredentials()
	if err != nil {
		t.Fatalf("ListAllCredentials() after delete error = %v", err)
	}
	if len(allAfter) != 0 {
		t.Errorf("expected 0 credentials after domain delete, got %d:", len(allAfter))
		for _, c := range allAfter {
			t.Logf("  orphaned: scope_type=%s scope_id=%d name=%q", c.ScopeType, c.ScopeID, c.Name)
		}
	}
}

// TestDeleteWorkspace_CleansUpCredentials verifies that deleting a workspace
// removes all credentials whose scope_type='workspace' and scope_id matches.
func TestDeleteWorkspace_CleansUpCredentials(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Arrange: build minimal hierarchy for the workspace
	eco := &models.Ecosystem{Name: "ws-cred-cleanup-eco"}
	if err := ds.CreateEcosystem(eco); err != nil {
		t.Fatalf("setup: CreateEcosystem() error = %v", err)
	}
	domain := &models.Domain{EcosystemID: eco.ID, Name: "ws-cred-cleanup-domain"}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("setup: CreateDomain() error = %v", err)
	}
	app := &models.App{DomainID: domain.ID, Name: "ws-cred-cleanup-app", Path: "/tmp/ws-cred"}
	if err := ds.CreateApp(app); err != nil {
		t.Fatalf("setup: CreateApp() error = %v", err)
	}

	ws := &models.Workspace{
		AppID:     app.ID,
		Name:      "ws-cred-cleanup",
		Slug:      "ws-cred-cleanup-eco-domain-app-ws",
		ImageName: "test:latest",
		Status:    "stopped",
	}
	if err := ds.CreateWorkspace(ws); err != nil {
		t.Fatalf("setup: CreateWorkspace() error = %v", err)
	}

	// Add two workspace-scoped credentials
	createWorkspaceCred(t, ds, ws.ID, "ws-cred-a")
	createWorkspaceCred(t, ds, ws.ID, "ws-cred-b")

	// Confirm 2 credentials exist
	creds, err := ds.ListCredentialsByScope(models.CredentialScopeWorkspace, int64(ws.ID))
	if err != nil {
		t.Fatalf("setup: ListCredentialsByScope() error = %v", err)
	}
	if len(creds) != 2 {
		t.Fatalf("setup: expected 2 credentials, got %d", len(creds))
	}

	// Act: delete the workspace
	if err := ds.DeleteWorkspace(ws.ID); err != nil {
		t.Fatalf("DeleteWorkspace() error = %v", err)
	}

	// Assert: credentials are gone
	remaining, err := ds.ListCredentialsByScope(models.CredentialScopeWorkspace, int64(ws.ID))
	if err != nil {
		t.Fatalf("ListCredentialsByScope() after delete error = %v", err)
	}
	if len(remaining) != 0 {
		t.Errorf("expected 0 credentials after workspace delete, got %d", len(remaining))
	}
}
