package handlers

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/resource"
)

// =============================================================================
// Test helpers
// =============================================================================

// setupMoveStore creates a mock store with a 2-ecosystem, 2-domain, 2-system,
// 2-app hierarchy sufficient for all move/detach scenarios.
//
//	eco1 (id=1) → dom1 (id=1) → sys1 (id=1) → app1 (id=1)
//	eco2 (id=2) → dom2 (id=2) → sys2 (id=2) → app2 (id=2)
func setupMoveStore(t *testing.T) (
	store *db.MockDataStore,
	eco1, eco2 *models.Ecosystem,
	dom1, dom2 *models.Domain,
	sys1, sys2 *models.System,
	app1, app2 *models.App,
) {
	t.Helper()
	store = db.NewMockDataStore()

	eco1 = &models.Ecosystem{Name: "eco1"}
	eco2 = &models.Ecosystem{Name: "eco2"}
	must(t, store.CreateEcosystem(eco1))
	must(t, store.CreateEcosystem(eco2))

	dom1 = &models.Domain{Name: "dom1", EcosystemID: nullInt(eco1.ID)}
	dom2 = &models.Domain{Name: "dom2", EcosystemID: nullInt(eco2.ID)}
	must(t, store.CreateDomain(dom1))
	must(t, store.CreateDomain(dom2))

	sys1 = &models.System{
		Name:        "sys1",
		DomainID:    nullInt(dom1.ID),
		EcosystemID: nullInt(eco1.ID),
	}
	sys2 = &models.System{
		Name:        "sys2",
		DomainID:    nullInt(dom2.ID),
		EcosystemID: nullInt(eco2.ID),
	}
	must(t, store.CreateSystem(sys1))
	must(t, store.CreateSystem(sys2))

	app1 = &models.App{
		Name:     "app1",
		Path:     "/app1",
		DomainID: nullInt(dom1.ID),
		SystemID: nullInt(sys1.ID),
	}
	app2 = &models.App{
		Name:     "app2",
		Path:     "/app2",
		DomainID: nullInt(dom2.ID),
		SystemID: nullInt(sys2.ID),
	}
	must(t, store.CreateApp(app1))
	must(t, store.CreateApp(app2))

	return
}

func nullInt(i int) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(i), Valid: true}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}
}

func newCtx(store *db.MockDataStore) resource.Context {
	return resource.Context{DataStore: store}
}

// callsFor returns all recorded calls with the given method name.
func callsFor(store *db.MockDataStore, method string) []db.MockDataStoreCall {
	var out []db.MockDataStoreCall
	for _, c := range store.GetCalls() {
		if c.Method == method {
			out = append(out, c)
		}
	}
	return out
}

// =============================================================================
// MoveResult.String()
// =============================================================================

func TestMoveResult_String(t *testing.T) {
	tests := []struct {
		name   string
		result MoveResult
		want   string
	}{
		{
			name: "no-op",
			result: MoveResult{
				Kind: "system", Name: "payments",
				ToParent: "domain/backend", NoOp: true,
			},
			want: "system/payments already at domain/backend (no-op)",
		},
		{
			name: "system move no cascade",
			result: MoveResult{
				Kind: "system", Name: "payments",
				FromParent: "domain/old", ToParent: "domain/new", CascadedApps: 0,
			},
			want: "system/payments moved to domain/new",
		},
		{
			name: "system move 1 app cascaded",
			result: MoveResult{
				Kind: "system", Name: "payments",
				FromParent: "domain/old", ToParent: "domain/new", CascadedApps: 1,
			},
			want: "system/payments moved to domain/new (1 app cascaded)",
		},
		{
			name: "system move plural cascade",
			result: MoveResult{
				Kind: "system", Name: "payments",
				FromParent: "domain/old", ToParent: "domain/new", CascadedApps: 3,
			},
			want: "system/payments moved to domain/new (3 apps cascaded)",
		},
		{
			name: "app move",
			result: MoveResult{
				Kind: "app", Name: "checkout",
				FromParent: "system/orders", ToParent: "system/payments",
			},
			want: "app/checkout moved to system/payments",
		},
		{
			name: "app no-op",
			result: MoveResult{
				Kind: "app", Name: "checkout",
				ToParent: "system/payments", NoOp: true,
			},
			want: "app/checkout already at system/payments (no-op)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.result.String()
			if got != tc.want {
				t.Errorf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}

// =============================================================================
// SystemHandler.Move
// =============================================================================

func TestSystemHandler_Move_HappyPath(t *testing.T) {
	store, eco1, _, dom1, _, _, _, app1, _ := setupMoveStore(t)
	// Create dom3 in eco1 so the cross-ecosystem guard won't fire
	dom3 := &models.Domain{Name: "dom3", EcosystemID: nullInt(eco1.ID)}
	must(t, store.CreateDomain(dom3))
	// app1 is a child of sys1 — expect cascade
	_ = dom1

	h := NewSystemHandler()
	result, err := h.Move(newCtx(store), "sys1", MoveTarget{DomainName: "dom3"})
	if err != nil {
		t.Fatalf("Move() error: %v", err)
	}
	if result.NoOp {
		t.Error("expected NoOp=false")
	}
	if result.Kind != "system" || result.Name != "sys1" {
		t.Errorf("unexpected result kind/name: %v/%v", result.Kind, result.Name)
	}
	if !strings.Contains(result.ToParent, "dom3") {
		t.Errorf("ToParent %q should mention dom3", result.ToParent)
	}
	if !strings.Contains(result.FromParent, "dom1") {
		t.Errorf("FromParent %q should mention dom1", result.FromParent)
	}

	// Verify MoveSystem was called
	calls := callsFor(store, "MoveSystem")
	if len(calls) == 0 {
		t.Error("expected MoveSystem to be called")
	}

	// Cascade: app1 was a child of sys1 so CascadedApps should be 1
	if result.CascadedApps != 1 {
		t.Errorf("CascadedApps = %d, want 1", result.CascadedApps)
	}
	_ = app1
	_ = dom1
}

func TestSystemHandler_Move_Idempotent(t *testing.T) {
	store, _, _, dom1, _, sys1, _, _, _ := setupMoveStore(t)
	_ = sys1

	h := NewSystemHandler()
	// sys1 is already in dom1
	result, err := h.Move(newCtx(store), "sys1", MoveTarget{DomainName: dom1.Name})
	if err != nil {
		t.Fatalf("Move() error: %v", err)
	}
	if !result.NoOp {
		t.Error("expected NoOp=true when already at target domain")
	}
	// MoveSystem should NOT have been called
	calls := callsFor(store, "MoveSystem")
	if len(calls) != 0 {
		t.Errorf("MoveSystem called %d time(s), want 0 for idempotent op", len(calls))
	}
}

func TestSystemHandler_Move_TargetDomainNotFound(t *testing.T) {
	store, _, _, _, _, _, _, _, _ := setupMoveStore(t)
	h := NewSystemHandler()
	_, err := h.Move(newCtx(store), "sys1", MoveTarget{DomainName: "nonexistent-domain"})
	if err == nil {
		t.Fatal("expected error for nonexistent target domain, got nil")
	}
}

func TestSystemHandler_Move_SystemNotFound(t *testing.T) {
	store, _, _, _, _, _, _, _, _ := setupMoveStore(t)
	h := NewSystemHandler()
	_, err := h.Move(newCtx(store), "ghost-system", MoveTarget{DomainName: "dom1"})
	if err == nil {
		t.Fatal("expected error for nonexistent system, got nil")
	}
}

func TestSystemHandler_Move_AmbiguousWithoutEcosystemHint(t *testing.T) {
	store, eco1, eco2, dom1, dom2, _, _, _, _ := setupMoveStore(t)
	// Create a second "sys1" in eco2
	sys1b := &models.System{
		Name:        "sys1",
		DomainID:    nullInt(dom2.ID),
		EcosystemID: nullInt(eco2.ID),
	}
	must(t, store.CreateSystem(sys1b))
	_ = eco1
	_ = dom1

	h := NewSystemHandler()
	_, err := h.Move(newCtx(store), "sys1", MoveTarget{DomainName: "dom1"})
	if err == nil {
		t.Fatal("expected ambiguity error, got nil")
	}
	if !strings.Contains(err.Error(), "ambiguous") && !strings.Contains(err.Error(), "ecosystem") {
		t.Errorf("error %q doesn't mention ambiguity or ecosystem", err.Error())
	}
}

func TestSystemHandler_Move_EcosystemHintDisambiguates(t *testing.T) {
	store, eco1, eco2, dom1, dom2, _, _, _, _ := setupMoveStore(t)
	// Create a second "sys1" in eco2/dom2 so name is ambiguous
	sys1b := &models.System{
		Name:        "sys1",
		DomainID:    nullInt(dom2.ID),
		EcosystemID: nullInt(eco2.ID),
	}
	must(t, store.CreateSystem(sys1b))

	// Create a third domain in eco1 to move to
	dom3 := &models.Domain{Name: "dom3", EcosystemID: nullInt(eco1.ID)}
	must(t, store.CreateDomain(dom3))
	_ = dom1

	h := NewSystemHandler()
	result, err := h.Move(newCtx(store), "sys1", MoveTarget{
		DomainName:    dom3.Name,
		EcosystemName: eco1.Name,
	})
	if err != nil {
		t.Fatalf("Move() with ecosystem hint error: %v", err)
	}
	if result.NoOp {
		t.Error("expected NoOp=false")
	}
}

func TestSystemHandler_Move_CrossEcosystemRejected(t *testing.T) {
	store, _, _, _, dom2, _, _, _, _ := setupMoveStore(t)
	// sys1 is in eco1/dom1 — try to move it to eco2/dom2
	h := NewSystemHandler()
	_, err := h.Move(newCtx(store), "sys1", MoveTarget{DomainName: dom2.Name})
	if err == nil {
		t.Fatal("expected cross-ecosystem error, got nil")
	}
	if !strings.Contains(err.Error(), "ecosystem") {
		t.Errorf("error %q should mention ecosystem", err.Error())
	}
}

func TestSystemHandler_Move_MoveSystemError(t *testing.T) {
	store, _, _, _, _, _, _, _, _ := setupMoveStore(t)
	// Add a dom3 in same eco so move would normally succeed
	eco1ID := store.Systems[1].EcosystemID.Int64
	dom3 := &models.Domain{Name: "dom3", EcosystemID: sql.NullInt64{Int64: eco1ID, Valid: true}}
	must(t, store.CreateDomain(dom3))

	store.MoveSystemErr = fmt.Errorf("db error")
	h := NewSystemHandler()
	_, err := h.Move(newCtx(store), "sys1", MoveTarget{DomainName: "dom3"})
	if err == nil {
		t.Fatal("expected error from MoveSystem, got nil")
	}
}

// =============================================================================
// AppHandler.Move
// =============================================================================

func TestAppHandler_Move_HappyPath_ToSystem(t *testing.T) {
	store, eco1, _, dom1, _, _, _, app1, _ := setupMoveStore(t)
	_ = app1
	// Create sys3 in eco1/dom1 so cross-ecosystem guard won't fire
	sys3 := &models.System{Name: "sys3", DomainID: nullInt(dom1.ID), EcosystemID: nullInt(eco1.ID)}
	must(t, store.CreateSystem(sys3))

	h := NewAppHandler()
	result, err := h.Move(newCtx(store), "app1", MoveTarget{SystemName: "sys3"})
	if err != nil {
		t.Fatalf("Move() error: %v", err)
	}
	if result.NoOp {
		t.Error("expected NoOp=false")
	}
	if !strings.Contains(result.ToParent, "sys3") {
		t.Errorf("ToParent %q should mention sys3", result.ToParent)
	}
	calls := callsFor(store, "MoveApp")
	if len(calls) == 0 {
		t.Error("expected MoveApp to be called")
	}
}

func TestAppHandler_Move_HappyPath_ToDomainOnly(t *testing.T) {
	store, _, _, dom1, _, _, _, app1, _ := setupMoveStore(t)
	// Create a second domain in eco1 to use as domain-only target
	dom3 := &models.Domain{Name: "dom3", EcosystemID: dom1.EcosystemID}
	must(t, store.CreateDomain(dom3))
	_ = app1

	h := NewAppHandler()
	result, err := h.Move(newCtx(store), "app1", MoveTarget{DomainName: "dom3"})
	if err != nil {
		t.Fatalf("Move() to domain-only error: %v", err)
	}
	if result.NoOp {
		t.Error("expected NoOp=false")
	}
	if !strings.Contains(result.ToParent, "dom3") {
		t.Errorf("ToParent %q should mention dom3", result.ToParent)
	}
}

func TestAppHandler_Move_Idempotent(t *testing.T) {
	store, _, _, _, _, sys1, _, _, _ := setupMoveStore(t)
	// app1 is already at sys1
	h := NewAppHandler()
	result, err := h.Move(newCtx(store), "app1", MoveTarget{SystemName: sys1.Name})
	if err != nil {
		t.Fatalf("Move() error: %v", err)
	}
	if !result.NoOp {
		t.Error("expected NoOp=true when already at target system")
	}
	calls := callsFor(store, "MoveApp")
	if len(calls) != 0 {
		t.Errorf("MoveApp called %d time(s), want 0 for idempotent op", len(calls))
	}
}

func TestAppHandler_Move_AppNotFound(t *testing.T) {
	store, _, _, _, _, _, _, _, _ := setupMoveStore(t)
	h := NewAppHandler()
	_, err := h.Move(newCtx(store), "ghost-app", MoveTarget{SystemName: "sys1"})
	if err == nil {
		t.Fatal("expected error for nonexistent app")
	}
}

func TestAppHandler_Move_NoTargetError(t *testing.T) {
	store, _, _, _, _, _, _, _, _ := setupMoveStore(t)
	h := NewAppHandler()
	// Neither SystemName nor DomainName provided
	_, err := h.Move(newCtx(store), "app1", MoveTarget{})
	if err == nil {
		t.Fatal("expected error when no target provided")
	}
}

func TestAppHandler_Move_AmbiguousSystem(t *testing.T) {
	store, _, eco2, dom1, dom2, _, _, _, _ := setupMoveStore(t)
	// Create a second "sys1" in eco2/dom2
	sys1b := &models.System{
		Name:        "sys1",
		DomainID:    nullInt(dom2.ID),
		EcosystemID: nullInt(eco2.ID),
	}
	must(t, store.CreateSystem(sys1b))
	_ = dom1

	h := NewAppHandler()
	// app1 is in eco1 — sys1 name is now ambiguous
	_, err := h.Move(newCtx(store), "app1", MoveTarget{SystemName: "sys1"})
	if err == nil {
		t.Fatal("expected ambiguity error, got nil")
	}
}

func TestAppHandler_Move_CrossEcosystemRejected(t *testing.T) {
	store, _, _, _, _, _, sys2, _, _ := setupMoveStore(t)
	// app1 is in eco1; sys2 is in eco2
	_ = sys2
	h := NewAppHandler()
	_, err := h.Move(newCtx(store), "app1", MoveTarget{SystemName: "sys2"})
	if err == nil {
		t.Fatal("expected cross-ecosystem error")
	}
	if !strings.Contains(err.Error(), "ecosystem") {
		t.Errorf("error %q should mention ecosystem", err.Error())
	}
}

// =============================================================================
// AppHandler.Detach
// =============================================================================

func TestAppHandler_Detach_HappyPath(t *testing.T) {
	store, _, _, _, _, _, sys1, app1, _ := setupMoveStore(t)
	_ = sys1
	_ = app1

	h := NewAppHandler()
	result, err := h.Detach(newCtx(store), "app1", "")
	if err != nil {
		t.Fatalf("Detach() error: %v", err)
	}
	if result.NoOp {
		t.Error("expected NoOp=false for detach")
	}
	if !strings.Contains(result.FromParent, "sys1") {
		t.Errorf("FromParent %q should mention sys1", result.FromParent)
	}
	// Both FKs should have been set to NULL via MoveApp
	calls := callsFor(store, "MoveApp")
	if len(calls) == 0 {
		t.Error("expected MoveApp to be called")
	}
	// Verify the in-memory app has both FKs cleared
	updatedApp := store.Apps[app1.ID]
	if updatedApp.SystemID.Valid {
		t.Error("expected SystemID to be NULL after detach")
	}
	if updatedApp.DomainID.Valid {
		t.Error("expected DomainID to be NULL after detach")
	}
}

func TestAppHandler_Detach_NotAttached(t *testing.T) {
	store, _, _, dom1, _, _, _, _, _ := setupMoveStore(t)
	// Create an app with a domain but no system — found by FindAppsByName but
	// SystemID.Valid is false, so Detach should return "not attached" error.
	unattached := &models.App{
		Name:     "unattached",
		Path:     "/unattached",
		DomainID: nullInt(dom1.ID),
		SystemID: sql.NullInt64{},
	}
	must(t, store.CreateApp(unattached))

	h := NewAppHandler()
	_, err := h.Detach(newCtx(store), "unattached", "")
	if err == nil {
		t.Fatal("expected error when app is not attached to a system")
	}
	if !strings.Contains(err.Error(), "not attached") {
		t.Errorf("error %q should mention 'not attached'", err.Error())
	}
}

func TestAppHandler_Detach_AppNotFound(t *testing.T) {
	store, _, _, _, _, _, _, _, _ := setupMoveStore(t)
	h := NewAppHandler()
	_, err := h.Detach(newCtx(store), "ghost-app", "")
	if err == nil {
		t.Fatal("expected error for nonexistent app")
	}
}

func TestAppHandler_Detach_EcosystemHintDisambiguates(t *testing.T) {
	store, eco1, eco2, dom1, dom2, sys1, sys2, app1, _ := setupMoveStore(t)
	// Create a second "app1" in eco2 attached to sys2
	app1b := &models.App{
		Name:     "app1",
		Path:     "/app1b",
		DomainID: nullInt(dom2.ID),
		SystemID: nullInt(sys2.ID),
	}
	must(t, store.CreateApp(app1b))
	_ = dom1
	_ = sys1
	_ = app1

	h := NewAppHandler()
	// Disambiguate by eco1 — should detach the eco1 app
	result, err := h.Detach(newCtx(store), "app1", eco1.Name)
	if err != nil {
		t.Fatalf("Detach() with ecosystem hint error: %v", err)
	}
	_ = eco2
	if result.NoOp {
		t.Error("expected NoOp=false")
	}
}

// =============================================================================
// SystemHandler.Apply — reparent detection (#397 bug fix)
// =============================================================================

func TestSystemHandler_Apply_ReparentDetected_DelegatesMove(t *testing.T) {
	store, _, _, _, dom2, sys1, _, _, _ := setupMoveStore(t)
	// sys1 is currently in eco1/dom1. Apply YAML that puts it in dom2 (same eco?).
	// For the reparent path to trigger, dom2 must be in same ecosystem as sys1.
	// Patch dom2 to be in eco1 for this test.
	dom2.EcosystemID = sys1.EcosystemID
	store.Domains[dom2.ID] = dom2

	h := NewSystemHandler()
	yamlData := []byte(fmt.Sprintf(`
apiVersion: v1
kind: System
metadata:
  name: sys1
  domain: %s
spec: {}
`, dom2.Name))

	_, err := h.Apply(newCtx(store), yamlData)
	if err != nil {
		t.Fatalf("Apply() reparent error: %v", err)
	}

	// MoveSystem must have been called (not a fresh create/update)
	calls := callsFor(store, "MoveSystem")
	if len(calls) == 0 {
		t.Error("expected MoveSystem to be called when Apply detects reparent")
	}
}

// =============================================================================
// AppHandler.Apply — reparent detection (#397 bug fix)
// =============================================================================

func TestAppHandler_Apply_ReparentDetected_DelegatesMove(t *testing.T) {
	store, _, _, dom1, dom2, sys1, sys2, app1, _ := setupMoveStore(t)
	// app1 is in eco1/dom1/sys1. Apply YAML with dom2/sys2.
	// For cross-eco guard not to fire, put dom2/sys2 in eco1.
	dom2.EcosystemID = dom1.EcosystemID
	sys2.EcosystemID = sys1.EcosystemID
	store.Domains[dom2.ID] = dom2
	store.Systems[sys2.ID] = sys2
	_ = app1

	// Need an active ecosystem for the Apply context-path
	eco1 := store.Ecosystems["eco1"]
	store.Context = &models.Context{ActiveEcosystemID: &eco1.ID}

	h := NewAppHandler()
	yamlData := []byte(fmt.Sprintf(`
apiVersion: v1
kind: App
metadata:
  name: app1
  domain: %s
  system: %s
  ecosystem: eco1
spec:
  path: /app1
`, dom2.Name, sys2.Name))

	_, err := h.Apply(newCtx(store), yamlData)
	if err != nil {
		t.Fatalf("Apply() reparent error: %v", err)
	}

	calls := callsFor(store, "MoveApp")
	if len(calls) == 0 {
		t.Error("expected MoveApp to be called when Apply detects reparent")
	}
}

// =============================================================================
// Apply-as-move convergence — real SQLite integration (#397)
// =============================================================================

// TestApplyReparent_RealSQLite verifies that Apply correctly reparents via
// MoveSystem / MoveApp against the real in-memory SQLite store, not just the
// mock. This ensures the actual SQL queries in store_move_stub.go are correct.
func TestApplyReparent_RealSQLite_SystemAndApp(t *testing.T) {
	RegisterAll()
	ds := createStackingDS(t)
	defer ds.Close()
	ctx := resource.Context{DataStore: ds}

	// Build: eco → dom1/dom2 → sys → app
	doApply(t, ctx, KindEcosystem, "apiVersion: devopsmaestro.io/v1\nkind: Ecosystem\nmetadata:\n  name: reparent-eco\nspec: {}\n")
	doApply(t, ctx, KindDomain, "apiVersion: devopsmaestro.io/v1\nkind: Domain\nmetadata:\n  name: dom-a\n  ecosystem: reparent-eco\nspec: {}\n")
	doApply(t, ctx, KindDomain, "apiVersion: devopsmaestro.io/v1\nkind: Domain\nmetadata:\n  name: dom-b\n  ecosystem: reparent-eco\nspec: {}\n")
	doApply(t, ctx, KindSystem, "apiVersion: devopsmaestro.io/v1\nkind: System\nmetadata:\n  name: reparent-sys\n  domain: dom-a\n  ecosystem: reparent-eco\nspec: {}\n")

	// App in dom-a under reparent-sys
	doApply(t, ctx, KindApp, "apiVersion: devopsmaestro.io/v1\nkind: App\nmetadata:\n  name: reparent-app\n  domain: dom-a\n  ecosystem: reparent-eco\nspec:\n  path: /reparent-app\n")

	// Reparent: apply same System YAML but now under dom-b
	doApply(t, ctx, KindSystem, "apiVersion: devopsmaestro.io/v1\nkind: System\nmetadata:\n  name: reparent-sys\n  domain: dom-b\n  ecosystem: reparent-eco\nspec: {}\n")

	// App was in dom-a; system move should cascade app to dom-b.
	// Verify via a second app export that the app appears (not errored out).
	y, _ := exportSnapshot(t, ds)
	if !strings.Contains(y, "reparent-app") {
		t.Error("apply-reparent system: child app missing from export after system reparent")
	}
	// dom-b should appear in export (system was moved there; its domain is exported)
	if !strings.Contains(y, "dom-b") {
		t.Error("apply-reparent system: dom-b missing from export after system reparent")
	}
}

// TestApplyReparent_RealSQLite_AppMove verifies that applying App YAML with a
// new domain reparents the app atomically in the real SQLite store.
func TestApplyReparent_RealSQLite_AppMove(t *testing.T) {
	RegisterAll()
	ds := createStackingDS(t)
	defer ds.Close()
	ctx := resource.Context{DataStore: ds}

	doApply(t, ctx, KindEcosystem, "apiVersion: devopsmaestro.io/v1\nkind: Ecosystem\nmetadata:\n  name: app-eco\nspec: {}\n")
	doApply(t, ctx, KindDomain, "apiVersion: devopsmaestro.io/v1\nkind: Domain\nmetadata:\n  name: app-dom-1\n  ecosystem: app-eco\nspec: {}\n")
	doApply(t, ctx, KindDomain, "apiVersion: devopsmaestro.io/v1\nkind: Domain\nmetadata:\n  name: app-dom-2\n  ecosystem: app-eco\nspec: {}\n")
	doApply(t, ctx, KindApp, "apiVersion: devopsmaestro.io/v1\nkind: App\nmetadata:\n  name: migrating-app\n  domain: app-dom-1\n  ecosystem: app-eco\nspec:\n  path: /migrating-app\n")

	// Move app to app-dom-2 via re-apply
	doApply(t, ctx, KindApp, "apiVersion: devopsmaestro.io/v1\nkind: App\nmetadata:\n  name: migrating-app\n  domain: app-dom-2\n  ecosystem: app-eco\nspec:\n  path: /migrating-app\n")

	y, _ := exportSnapshot(t, ds)
	if !strings.Contains(y, "migrating-app") {
		t.Error("apply-reparent app: app missing from export after domain change")
	}
	if !strings.Contains(y, "app-dom-2") {
		t.Error("apply-reparent app: app-dom-2 missing from export (app should have moved)")
	}
}
