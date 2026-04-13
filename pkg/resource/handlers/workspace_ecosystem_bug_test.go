package handlers

// =============================================================================
// TDD Phase 2 (RED): Bug #153 — Workspace apply picks wrong domain when name
// exists across ecosystems
//
// Root cause (from debug audit on epic #151):
//   1. WorkspaceMetadata in models/workspace.go:48-54 has NO "ecosystem" field —
//      there is no way to encode which ecosystem a workspace belongs to in YAML.
//   2. workspace.go:53 calls ds.ListAllDomains() and takes the FIRST domain
//      whose name matches — ignoring ecosystem entirely.
//
// Fix required (per issue #153 acceptance criteria):
//   1. Add `Ecosystem string` field to models.WorkspaceMetadata
//   2. Update models.Workspace.ToYAML() to populate Metadata.Ecosystem
//   3. Update WorkspaceHandler.ToYAML() to include ecosystem in metadata output
//   4. Update WorkspaceHandler.Apply(): when metadata.ecosystem is present,
//      use ds.ListDomainsByEcosystem() to scope domain lookup to that ecosystem.
//
// ALL tests in this file MUST FAIL until the fix is implemented.
// =============================================================================

import (
	"database/sql"
	"strings"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/resource"
	"gopkg.in/yaml.v3"
)

// =============================================================================
// Helper: build a cross-ecosystem collision scenario
//
// Creates two ecosystems (A and B) each with a domain named "library" and an
// app named "api". Workspace "dev-env" belongs to ecosystemA/library/api.
// =============================================================================

type collisionScenario struct {
	store      *db.MockDataStore
	ecosystemA *models.Ecosystem
	ecosystemB *models.Ecosystem
	domainA    *models.Domain // name="library", ecosystem=A
	domainB    *models.Domain // name="library", ecosystem=B
	appA       *models.App    // name="api", domain=A
	appB       *models.App    // name="api", domain=B
	workspaceA *models.Workspace
}

func setupCollisionScenario(t *testing.T) *collisionScenario {
	t.Helper()
	store := db.NewMockDataStore()

	// Ecosystem A
	ecoA := &models.Ecosystem{Name: "eco-alpha"}
	if err := store.CreateEcosystem(ecoA); err != nil {
		t.Fatalf("CreateEcosystem(A): %v", err)
	}

	// Ecosystem B
	ecoB := &models.Ecosystem{Name: "eco-beta"}
	if err := store.CreateEcosystem(ecoB); err != nil {
		t.Fatalf("CreateEcosystem(B): %v", err)
	}

	// Domain "library" in A
	domA := &models.Domain{Name: "library", EcosystemID: sql.NullInt64{Int64: int64(ecoA.ID), Valid: true}}
	if err := store.CreateDomain(domA); err != nil {
		t.Fatalf("CreateDomain(A/library): %v", err)
	}

	// Domain "library" in B  (same name — this triggers the bug)
	domB := &models.Domain{Name: "library", EcosystemID: sql.NullInt64{Int64: int64(ecoB.ID), Valid: true}}
	if err := store.CreateDomain(domB); err != nil {
		t.Fatalf("CreateDomain(B/library): %v", err)
	}

	// App "api" under domainA
	appA := &models.App{Name: "api", DomainID: sql.NullInt64{Int64: int64(domA.ID), Valid: true}, Path: "/alpha/api"}
	if err := store.CreateApp(appA); err != nil {
		t.Fatalf("CreateApp(A/api): %v", err)
	}

	// App "api" under domainB
	appB := &models.App{Name: "api", DomainID: sql.NullInt64{Int64: int64(domB.ID), Valid: true}, Path: "/beta/api"}
	if err := store.CreateApp(appB); err != nil {
		t.Fatalf("CreateApp(B/api): %v", err)
	}

	// Workspace belonging to eco-alpha / library / api
	ws := &models.Workspace{
		Name:      "dev-env",
		AppID:     appA.ID,
		ImageName: "golang:1.22",
		Status:    "stopped",
	}
	if err := store.CreateWorkspace(ws); err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	return &collisionScenario{
		store:      store,
		ecosystemA: ecoA,
		ecosystemB: ecoB,
		domainA:    domA,
		domainB:    domB,
		appA:       appA,
		appB:       appB,
		workspaceA: ws,
	}
}

// =============================================================================
// Test 1: WorkspaceYAML must include metadata.ecosystem on export
//
// RED: models.WorkspaceMetadata has no Ecosystem field, so ToYAML() can never
// produce "ecosystem: eco-alpha". This test will fail because the exported YAML
// will not contain any ecosystem field.
// =============================================================================

func TestWorkspaceToYAML_IncludesEcosystem(t *testing.T) {
	h := NewWorkspaceHandler()

	sc := setupCollisionScenario(t)

	// Simulate what get_all.go does: build a WorkspaceResource with all names
	// resolved (appName, domainName, gitRepoName are pre-computed).
	// After fix, NewWorkspaceResource (or a new variant) must also accept
	// ecosystemName so ToYAML() can include metadata.ecosystem.
	res := NewWorkspaceResource(sc.workspaceA, "api", "library", "", "eco-alpha")

	// Also need to set ecosystem name — this requires either a new field in
	// WorkspaceResource or an updated NewWorkspaceResource signature.
	// For now we call ToYAML; the test asserts the output includes ecosystem.

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}
	yamlStr := string(yamlBytes)

	// RED: WorkspaceMetadata has no Ecosystem field — this will never appear
	if !strings.Contains(yamlStr, "ecosystem: eco-alpha") {
		t.Errorf("ToYAML() YAML must contain 'ecosystem: eco-alpha' in metadata\ngot:\n%s", yamlStr)
	}
}

// =============================================================================
// Test 2: Apply with metadata.ecosystem must resolve domain scoped to that ecosystem
//
// RED: WorkspaceMetadata has no Ecosystem field, so yaml.Unmarshal silently
// drops the "ecosystem: eco-alpha" key. The handler then calls ListAllDomains()
// with no scoping — the ecosystem hint is lost.
//
// This test verifies the struct can capture the ecosystem field at all,
// and that Apply() uses it for scoped domain resolution.
// =============================================================================

func TestWorkspaceApply_ResolvesCorrectEcosystem(t *testing.T) {
	// Step 1: verify the raw YAML key "ecosystem:" survives round-trip through
	// WorkspaceYAML struct. We marshal a map that has the key and check whether
	// unmarshaling it into WorkspaceYAML round-trips the value.
	//
	// Use yaml.Node / map[string]interface{} approach so we can check without
	// a compile dependency on the (missing) Ecosystem field.
	yamlWithEcosystem := []byte(`
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: dev-env
  app: api
  domain: library
  ecosystem: eco-alpha
spec:
  image:
    name: golang:1.22
`)

	// Unmarshal into a generic map to confirm the raw key is present
	var rawDoc map[string]interface{}
	if err := yaml.Unmarshal(yamlWithEcosystem, &rawDoc); err != nil {
		t.Fatalf("raw yaml.Unmarshal: %v", err)
	}
	rawMeta, _ := rawDoc["metadata"].(map[string]interface{})
	rawEcosystem, _ := rawMeta["ecosystem"].(string)
	if rawEcosystem != "eco-alpha" {
		t.Fatalf("test setup error: raw YAML 'metadata.ecosystem' = %q, want 'eco-alpha'", rawEcosystem)
	}

	// Unmarshal into WorkspaceYAML — the ecosystem key must survive into
	// Metadata.Ecosystem. If the field is absent it will silently disappear.
	// We then re-marshal to YAML and check the key is present.
	var wsYAML models.WorkspaceYAML
	if err := yaml.Unmarshal(yamlWithEcosystem, &wsYAML); err != nil {
		t.Fatalf("yaml.Unmarshal into WorkspaceYAML: %v", err)
	}

	// Re-marshal and check ecosystem is preserved
	remarshaled, err := yaml.Marshal(wsYAML)
	if err != nil {
		t.Fatalf("yaml.Marshal WorkspaceYAML: %v", err)
	}
	// RED: Ecosystem field missing from struct → key disappears on unmarshal→marshal
	if !strings.Contains(string(remarshaled), "ecosystem: eco-alpha") {
		t.Errorf(
			"WorkspaceYAML.Metadata lost 'ecosystem: eco-alpha' during unmarshal/marshal round-trip\n"+
				"Bug: models.WorkspaceMetadata has no Ecosystem field — "+
				"the key is silently dropped\nremarshaled:\n%s",
			string(remarshaled),
		)
	}

	// Step 2: Apply() must use the ecosystem to scope domain resolution.
	// With two "library" domains (one per ecosystem), Apply with ecosystem=eco-alpha
	// must pick domainA (eco-alpha's domain), not domainB.
	h := NewWorkspaceHandler()
	sc := setupCollisionScenario(t)
	ctx := resource.Context{DataStore: sc.store}

	res, err := h.Apply(ctx, yamlWithEcosystem)
	if err != nil {
		t.Fatalf("Apply() error = %v (should succeed with valid metadata.ecosystem)", err)
	}

	wr, ok := res.(*WorkspaceResource)
	if !ok {
		t.Fatalf("result is not *WorkspaceResource")
	}

	// The workspace must be linked to eco-alpha's app, not eco-beta's app.
	ws := wr.Workspace()
	if ws.AppID != sc.appA.ID {
		t.Errorf(
			"Apply() resolved AppID=%d (eco-beta's app), want AppID=%d (eco-alpha's app)\n"+
				"Bug: ecosystem scoping not implemented — first-match from ListAllDomains() used",
			ws.AppID, sc.appA.ID,
		)
	}
}

// =============================================================================
// Test 3: Cross-ecosystem collision — apply WITHOUT ecosystem hint picks wrong app
//
// This documents the current (broken) behaviour: without metadata.ecosystem the
// handler is non-deterministic. This test is a negative-path assertion to
// confirm the collision exists today. It deliberately has no ecosystem in YAML.
//
// After the fix, the handler should either:
//   (a) require metadata.ecosystem when a domain name is ambiguous, OR
//   (b) document that metadata.ecosystem is the tie-breaker.
//
// RED: Currently the handler silently picks whichever domain comes first in the
// all-domain scan — no error is raised and no disambiguation is performed.
// =============================================================================

func TestWorkspaceApply_CollisionWithoutEcosystem_ReturnsError(t *testing.T) {
	h := NewWorkspaceHandler()

	sc := setupCollisionScenario(t)
	ctx := resource.Context{DataStore: sc.store}

	// YAML with domain name that exists in BOTH ecosystems — no ecosystem hint
	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: dev-env
  app: api
  domain: library
spec:
  image:
    name: golang:1.22
`)

	_, err := h.Apply(ctx, yamlData)
	// RED: The handler currently does NOT return an error when a domain name
	// is ambiguous — it silently picks the wrong one. This test asserts that
	// after the fix an error (or at least a deterministic resolution) is
	// produced. For now it FAILS because no error is returned.
	if err == nil {
		t.Error(
			"Apply() with ambiguous domain name across ecosystems should return an error " +
				"(or require metadata.ecosystem); got nil — silent wrong-ecosystem pick",
		)
	}
	if err != nil && !strings.Contains(err.Error(), "ecosystem") {
		t.Errorf(
			"Apply() ambiguity error should mention 'ecosystem'; got: %q",
			err.Error(),
		)
	}
}

// =============================================================================
// Test 4: Export → Apply round-trip preserves correct ecosystem
//
// Full round-trip: export workspace from eco-alpha → apply on a fresh store
// that has both ecosystems → verify the workspace ends up in eco-alpha.
//
// RED: The ToYAML() call below cannot include ecosystem in metadata (no field),
// so the subsequent Apply() will either fail or pick the wrong ecosystem.
// =============================================================================

func TestWorkspaceRoundTrip_EcosystemPreserved(t *testing.T) {
	h := NewWorkspaceHandler()

	// ── SOURCE store: full hierarchy ─────────────────────────────────────────
	srcStore := db.NewMockDataStore()

	srcEcoA := &models.Ecosystem{Name: "eco-alpha"}
	if err := srcStore.CreateEcosystem(srcEcoA); err != nil {
		t.Fatalf("src CreateEcosystem(A): %v", err)
	}
	srcEcoB := &models.Ecosystem{Name: "eco-beta"}
	if err := srcStore.CreateEcosystem(srcEcoB); err != nil {
		t.Fatalf("src CreateEcosystem(B): %v", err)
	}
	srcDomA := &models.Domain{Name: "library", EcosystemID: sql.NullInt64{Int64: int64(srcEcoA.ID), Valid: true}}
	if err := srcStore.CreateDomain(srcDomA); err != nil {
		t.Fatalf("src CreateDomain(A/library): %v", err)
	}
	srcDomB := &models.Domain{Name: "library", EcosystemID: sql.NullInt64{Int64: int64(srcEcoB.ID), Valid: true}}
	if err := srcStore.CreateDomain(srcDomB); err != nil {
		t.Fatalf("src CreateDomain(B/library): %v", err)
	}
	srcAppA := &models.App{Name: "api", DomainID: sql.NullInt64{Int64: int64(srcDomA.ID), Valid: true}, Path: "/alpha/api"}
	if err := srcStore.CreateApp(srcAppA); err != nil {
		t.Fatalf("src CreateApp(A/api): %v", err)
	}
	srcAppB := &models.App{Name: "api", DomainID: sql.NullInt64{Int64: int64(srcDomB.ID), Valid: true}, Path: "/beta/api"}
	if err := srcStore.CreateApp(srcAppB); err != nil {
		t.Fatalf("src CreateApp(B/api): %v", err)
	}
	srcWS := &models.Workspace{
		Name:      "dev-env",
		AppID:     srcAppA.ID,
		ImageName: "golang:1.22",
		Status:    "stopped",
	}
	if err := srcStore.CreateWorkspace(srcWS); err != nil {
		t.Fatalf("src CreateWorkspace: %v", err)
	}

	// Export: simulate get_all.go path.
	// After fix, NewWorkspaceResource must accept (and store) ecosystemName.
	exportRes := NewWorkspaceResource(srcWS, "api", "library", "", "eco-alpha")
	// GREEN: ecosystemName ("eco-alpha") is now stored via 5th parameter.

	yamlBytes, err := h.ToYAML(exportRes)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}
	yamlStr := string(yamlBytes)

	// Assert the exported YAML contains ecosystem — RED: will not be present
	if !strings.Contains(yamlStr, "ecosystem: eco-alpha") {
		t.Errorf(
			"exported YAML must contain 'ecosystem: eco-alpha' in metadata\ngot:\n%s",
			yamlStr,
		)
	}

	// ── DESTINATION store: fresh, both ecosystems present ───────────────────
	dstStore := db.NewMockDataStore()

	dstEcoA := &models.Ecosystem{Name: "eco-alpha"}
	if err := dstStore.CreateEcosystem(dstEcoA); err != nil {
		t.Fatalf("dst CreateEcosystem(A): %v", err)
	}
	dstEcoB := &models.Ecosystem{Name: "eco-beta"}
	if err := dstStore.CreateEcosystem(dstEcoB); err != nil {
		t.Fatalf("dst CreateEcosystem(B): %v", err)
	}
	dstDomA := &models.Domain{Name: "library", EcosystemID: sql.NullInt64{Int64: int64(dstEcoA.ID), Valid: true}}
	if err := dstStore.CreateDomain(dstDomA); err != nil {
		t.Fatalf("dst CreateDomain(A/library): %v", err)
	}
	dstDomB := &models.Domain{Name: "library", EcosystemID: sql.NullInt64{Int64: int64(dstEcoB.ID), Valid: true}}
	if err := dstStore.CreateDomain(dstDomB); err != nil {
		t.Fatalf("dst CreateDomain(B/library): %v", err)
	}
	dstAppA := &models.App{Name: "api", DomainID: sql.NullInt64{Int64: int64(dstDomA.ID), Valid: true}, Path: "/alpha/api"}
	if err := dstStore.CreateApp(dstAppA); err != nil {
		t.Fatalf("dst CreateApp(A/api): %v", err)
	}
	dstAppB := &models.App{Name: "api", DomainID: sql.NullInt64{Int64: int64(dstDomB.ID), Valid: true}, Path: "/beta/api"}
	if err := dstStore.CreateApp(dstAppB); err != nil {
		t.Fatalf("dst CreateApp(B/api): %v", err)
	}

	// Apply: no active context — must rely purely on YAML metadata
	dstCtx := resource.Context{DataStore: dstStore}
	res, err := h.Apply(dstCtx, yamlBytes)
	if err != nil {
		t.Fatalf("Apply() on destination store error = %v", err)
	}

	wr, ok := res.(*WorkspaceResource)
	if !ok {
		t.Fatalf("result is not *WorkspaceResource")
	}

	// Workspace must be linked to eco-alpha's app (dstAppA), not dstAppB
	ws := wr.Workspace()
	if ws.AppID != dstAppA.ID {
		t.Errorf(
			"round-trip Apply() resolved AppID=%d (eco-beta's app), want AppID=%d (eco-alpha's app)\n"+
				"Bug: ecosystem was not preserved in the YAML → wrong domain picked on apply",
			ws.AppID, dstAppA.ID,
		)
	}
}

// =============================================================================
// Test 5: WorkspaceMetadata yaml round-trip must preserve ecosystem field
//
// Uses gopkg.in/yaml.v3 unmarshal: if WorkspaceMetadata has no Ecosystem
// field (or has the wrong yaml tag), the "ecosystem: eco-alpha" key in the
// YAML snippet will be silently ignored and meta.Ecosystem stays "".
//
// RED: models.WorkspaceMetadata has no Ecosystem field → unmarshal silently
// drops the key → assertion below fails.
// =============================================================================

func TestWorkspaceMetadata_EcosystemYAMLRoundTrip(t *testing.T) {
	// Build a WorkspaceYAML manually and marshal it, then check the raw bytes
	// include "ecosystem:" — this will fail because WorkspaceMetadata has no
	// Ecosystem field, so yaml.Marshal will never emit that key.
	doc := models.WorkspaceYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Workspace",
		Metadata: models.WorkspaceMetadata{
			Name:   "dev-env",
			App:    "api",
			Domain: "library",
			// Ecosystem field intentionally absent from struct — this is the bug.
			// After fix, set: Ecosystem: "eco-alpha"
		},
	}

	out, err := yaml.Marshal(doc)
	if err != nil {
		t.Fatalf("yaml.Marshal WorkspaceYAML: %v", err)
	}
	outStr := string(out)

	// RED: no Ecosystem field → "ecosystem:" never appears in output
	if !strings.Contains(outStr, "ecosystem:") {
		t.Errorf(
			"marshaled WorkspaceYAML metadata must contain 'ecosystem:' key\n"+
				"Bug: models.WorkspaceMetadata is missing the Ecosystem field\ngot:\n%s",
			outStr,
		)
	}
}
