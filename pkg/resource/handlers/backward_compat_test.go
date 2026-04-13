package handlers

// =============================================================================
// Backward Compatibility Tests — System Hierarchy Layer (Issue #270, E.2)
//
// These tests verify that existing YAML files that predate the System layer
// still work correctly after the new `ecosystem → domain → system → app →
// workspace` hierarchy was introduced. All intermediate levels are optional,
// and these tests confirm that omitting system/domain/ecosystem fields does
// not break existing workflows.
// =============================================================================

import (
	"database/sql"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/resource"
)

// =============================================================================
// Test: App YAML without system field → null SystemID (legacy behavior)
// =============================================================================

func TestBackwardCompat_AppWithoutSystem_NullSystemID(t *testing.T) {
	store, _, _, _ := setupBackCompatStore(t)
	ctx := resource.Context{DataStore: store}
	h := NewAppHandler()

	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: legacy-app
  domain: compat-domain
spec:
  path: /legacy/app
`)

	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() error = %v (expected nil for app without system)", err)
	}
	ar, ok := res.(*AppResource)
	if !ok {
		t.Fatalf("result is not *AppResource")
	}
	if ar.App().SystemID.Valid {
		t.Errorf("SystemID.Valid = true, want false for app without system field")
	}
}

// =============================================================================
// Test: Workspace YAML with only app (no system/domain/ecosystem) → works
// =============================================================================

func TestBackwardCompat_WorkspaceWithOnlyApp(t *testing.T) {
	store, _, _, _ := setupBackCompatStore(t)
	ctx := resource.Context{DataStore: store}
	h := NewWorkspaceHandler()

	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: legacy-workspace
  app: compat-app
spec:
  image:
    name: ubuntu:22.04
`)

	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() error = %v (expected nil for workspace with only app)", err)
	}
	wr, ok := res.(*WorkspaceResource)
	if !ok {
		t.Fatalf("result is not *WorkspaceResource")
	}
	if wr.AppName() != "compat-app" {
		t.Errorf("AppName() = %q, want %q", wr.AppName(), "compat-app")
	}
}

// =============================================================================
// Test: Legacy 4-level YAML (eco+domain+app+workspace, no system) → works
// =============================================================================

func TestBackwardCompat_Legacy4LevelHierarchy(t *testing.T) {
	store, _, _, _ := setupBackCompatStore(t)
	ctx := resource.Context{DataStore: store}

	// Apply workspace against pre-existing legacy hierarchy
	h := NewWorkspaceHandler()
	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: legacy-ws-4level
  app: compat-app
spec:
  image:
    name: debian:bookworm
`)

	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() error = %v (expected nil for legacy 4-level hierarchy)", err)
	}
	if res.GetName() != "legacy-ws-4level" {
		t.Errorf("GetName() = %q, want %q", res.GetName(), "legacy-ws-4level")
	}
}

// =============================================================================
// Test: Full 5-level hierarchy (eco+domain+system+app+workspace) → works
// =============================================================================

func TestBackwardCompat_Full5LevelHierarchy(t *testing.T) {
	store, _, domainID, _ := setupBackCompatStore(t)
	ctx := resource.Context{DataStore: store}

	// Create a system within the existing domain
	sys := &models.System{
		Name:     "compat-system",
		DomainID: sql.NullInt64{Int64: int64(domainID), Valid: true},
	}
	if err := store.CreateSystem(sys); err != nil {
		t.Fatalf("CreateSystem: %v", err)
	}

	// Apply an app that references the system
	hApp := NewAppHandler()
	appYAML := []byte(`
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: new-sys-app
  domain: compat-domain
  system: compat-system
spec:
  path: /new/sys/app
`)
	appRes, err := hApp.Apply(ctx, appYAML)
	if err != nil {
		t.Fatalf("App Apply() error = %v", err)
	}
	ar := appRes.(*AppResource)
	if !ar.App().SystemID.Valid {
		t.Errorf("SystemID.Valid = false, want true when system is specified")
	}

	// Apply workspace against the new app
	sysAppID := ar.App().ID
	store.Context.ActiveAppID = &sysAppID

	hWs := NewWorkspaceHandler()
	wsYAML := []byte(`
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: full-5level-ws
  app: new-sys-app
spec:
  image:
    name: ubuntu:24.04
`)
	wsRes, err := hWs.Apply(ctx, wsYAML)
	if err != nil {
		t.Fatalf("Workspace Apply() error = %v", err)
	}
	if wsRes.GetName() != "full-5level-ws" {
		t.Errorf("workspace name = %q, want %q", wsRes.GetName(), "full-5level-ws")
	}
}

// =============================================================================
// Test: System YAML without domain → system with null DomainID
// =============================================================================

func TestBackwardCompat_SystemWithoutDomain_NullDomainID(t *testing.T) {
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}
	h := NewSystemHandler()

	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: System
metadata:
  name: orphan-system
spec: {}
`)

	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() error = %v (expected nil for system without domain)", err)
	}
	sr, ok := res.(*SystemResource)
	if !ok {
		t.Fatalf("result is not *SystemResource")
	}
	if sr.System().DomainID.Valid {
		t.Errorf("DomainID.Valid = true, want false for system without domain field")
	}
}

// setupBackCompatStore sets up a store with a full legacy hierarchy:
// ecosystem → domain → app (no system). Active context is set for all.
func setupBackCompatStore(t *testing.T) (*db.MockDataStore, int, int, int) {
	t.Helper()
	store := db.NewMockDataStore()

	eco := &models.Ecosystem{Name: "compat-eco"}
	if err := store.CreateEcosystem(eco); err != nil {
		t.Fatalf("CreateEcosystem: %v", err)
	}

	domain := &models.Domain{
		Name:        "compat-domain",
		EcosystemID: sql.NullInt64{Int64: int64(eco.ID), Valid: true},
	}
	if err := store.CreateDomain(domain); err != nil {
		t.Fatalf("CreateDomain: %v", err)
	}

	app := &models.App{
		Name:     "compat-app",
		Path:     "/compat/app",
		DomainID: sql.NullInt64{Int64: int64(domain.ID), Valid: true},
	}
	if err := store.CreateApp(app); err != nil {
		t.Fatalf("CreateApp: %v", err)
	}

	ecoID := eco.ID
	domainID := domain.ID
	appID := app.ID
	store.Context.ActiveEcosystemID = &ecoID
	store.Context.ActiveDomainID = &domainID
	store.Context.ActiveAppID = &appID

	return store, ecoID, domainID, appID
}
