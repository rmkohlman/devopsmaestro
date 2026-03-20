package handlers

// =============================================================================
// TDD Phase 2 (RED): Bug #129 — App Apply requires active ecosystem context
//
// app.go lines 48-56: Apply() requires dbCtx.ActiveEcosystemID — fails on
// fresh restore when no active context is set.
//
// Fix design: add `Ecosystem string` to AppMetadata, populate in ToYAML(),
// consume in Apply() — resolve ecosystem by name, then domain within it.
//
// These tests verify the round-trip requirement:
//   1. App YAML includes metadata.ecosystem when exported
//   2. Apply() can resolve an app from metadata.ecosystem without active context
//   3. Apply() still works with active context (backwards compatibility)
//
// ALL tests in this section WILL FAIL until:
//   - AppMetadata.Ecosystem field is added to models/app.go
//   - App.ToYAML() is updated to populate Ecosystem
//   - AppHandler.Apply() is updated to use metadata.ecosystem when no active context
// =============================================================================

import (
	"strings"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/resource"
	"gopkg.in/yaml.v3"
)

// TestAppHandler_Apply_WithMetadataEcosystem verifies that Apply() succeeds
// when metadata.ecosystem is specified in the YAML — without requiring an
// active ecosystem context to be set.
//
// RED: Fails because AppMetadata has no Ecosystem field and Apply() requires
// dbCtx.ActiveEcosystemID (line 54 of app.go).
func TestAppHandler_Apply_WithMetadataEcosystem(t *testing.T) {
	h := NewAppHandler()
	store := db.NewMockDataStore()

	// Create hierarchy but do NOT set active ecosystem context
	eco := &models.Ecosystem{Name: "restore-eco"}
	if err := store.CreateEcosystem(eco); err != nil {
		t.Fatalf("CreateEcosystem: %v", err)
	}

	domain := &models.Domain{Name: "restore-domain", EcosystemID: eco.ID}
	if err := store.CreateDomain(domain); err != nil {
		t.Fatalf("CreateDomain: %v", err)
	}

	// NO active ecosystem set — handler must use metadata.ecosystem
	ctx := resource.Context{DataStore: store}

	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: restore-app
  domain: restore-domain
  ecosystem: restore-eco
spec:
  path: /home/user/restore-app
`)

	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() with metadata.ecosystem should succeed without active context, got error: %v", err)
	}
	if res.GetName() != "restore-app" {
		t.Errorf("Apply() Name = %q, want %q", res.GetName(), "restore-app")
	}
}

// TestAppHandler_Apply_NoContextNoEcosystem verifies that when neither
// metadata.ecosystem nor an active ecosystem context is provided, Apply()
// returns a clear error.
func TestAppHandler_Apply_NoContextNoEcosystem(t *testing.T) {
	h := NewAppHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	// No active ecosystem context AND no metadata.ecosystem
	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: orphan-app
  domain: some-domain
spec:
  path: /some/path
`)

	_, err := h.Apply(ctx, yamlData)
	if err == nil {
		t.Error("Apply() should fail when no active ecosystem and no metadata.ecosystem, got nil")
	}
	// Error must mention 'ecosystem'
	if !strings.Contains(err.Error(), "ecosystem") {
		t.Errorf("Apply() error = %q, want it to mention 'ecosystem'", err.Error())
	}
}

// TestAppMetadata_EcosystemField verifies that AppMetadata has an Ecosystem
// field and that it is populated when YAML contains metadata.ecosystem.
//
// RED: Fails to COMPILE because AppMetadata has no Ecosystem field.
// Once models/app.go adds `Ecosystem string yaml:"ecosystem"`, this
// test will compile and then verify the field is populated correctly.
func TestAppMetadata_EcosystemField(t *testing.T) {
	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: eco-app
  domain: my-domain
  ecosystem: my-ecosystem
spec:
  path: /eco/app
`)

	var appYAML models.AppYAML
	if err := yaml.Unmarshal(yamlData, &appYAML); err != nil {
		t.Fatalf("yaml.Unmarshal error = %v", err)
	}

	// ── COMPILE ERROR EXPECTED BELOW — Ecosystem field doesn't exist yet ──────
	if appYAML.Metadata.Ecosystem == "" { // COMPILE ERROR: AppMetadata has no field 'Ecosystem'
		t.Error("AppYAML.Metadata.Ecosystem must be populated from YAML (field not yet added to AppMetadata)")
	}
	if appYAML.Metadata.Ecosystem != "my-ecosystem" {
		t.Errorf("AppYAML.Metadata.Ecosystem = %q, want %q", appYAML.Metadata.Ecosystem, "my-ecosystem")
	}
	// ──────────────────────────────────────────────────────────────────────────
}

// TestAppHandler_ToYAML_IncludesEcosystem verifies that ToYAML() includes
// metadata.ecosystem in the output when the app resource has ecosystem info.
//
// RED: Fails because (1) AppMetadata has no Ecosystem field, and (2)
// AppResource has no ecosystemName field. Both must be added.
func TestAppHandler_ToYAML_IncludesEcosystem(t *testing.T) {
	h := NewAppHandler()

	app := &models.App{
		ID:       1,
		DomainID: 1,
		Name:     "eco-app",
		Path:     "/eco/app",
	}

	res := NewAppResource(app, "my-domain", "my-ecosystem")

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}
	yamlStr := string(yamlBytes)

	// RED: After fix, exported YAML MUST include ecosystem field.
	// This assertion fails today (no ecosystem in output) — it's the bug.
	if !strings.Contains(yamlStr, "ecosystem:") {
		t.Errorf("ToYAML() must include 'ecosystem:' in metadata for context-free apply; got:\n%s", yamlStr)
	}
}

// TestAppHandler_Apply_FallbackToActiveContext verifies backwards compatibility:
// when metadata.ecosystem is NOT in the YAML, Apply() still uses the active
// ecosystem context (existing behavior preserved).
func TestAppHandler_Apply_FallbackToActiveContext(t *testing.T) {
	h := NewAppHandler()
	store, _, _ := setupAppTest(t)
	ctx := resource.Context{DataStore: store}

	// No metadata.ecosystem — should use active ecosystem from context (set by setupAppTest)
	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: fallback-app
  domain: app-domain
spec:
  path: /fallback/app
`)

	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() fallback to active context should succeed, got error: %v", err)
	}
	if res.GetName() != "fallback-app" {
		t.Errorf("Apply() Name = %q, want %q", res.GetName(), "fallback-app")
	}
}

// TestApp_RoundTrip_WithoutActiveContext verifies the full export→apply
// round-trip works without any active context: export an app (which must
// include metadata.ecosystem after fix), then apply the YAML into a fresh
// store with no active context.
//
// RED: Fails because (1) ToYAML doesn't emit metadata.ecosystem, and
// (2) Apply() requires active ecosystem context.
func TestApp_RoundTrip_WithoutActiveContext(t *testing.T) {
	h := NewAppHandler()

	// --- Source store: full hierarchy, no active context ---
	sourceStore := db.NewMockDataStore()

	eco := &models.Ecosystem{Name: "roundtrip-eco"}
	if err := sourceStore.CreateEcosystem(eco); err != nil {
		t.Fatalf("CreateEcosystem: %v", err)
	}

	domain := &models.Domain{Name: "roundtrip-domain", EcosystemID: eco.ID}
	if err := sourceStore.CreateDomain(domain); err != nil {
		t.Fatalf("CreateDomain: %v", err)
	}

	app := &models.App{
		Name:     "roundtrip-app",
		DomainID: domain.ID,
		Path:     "/rt/app",
	}
	if err := sourceStore.CreateApp(app); err != nil {
		t.Fatalf("CreateApp: %v", err)
	}

	// Export the app — NewAppResource now carries ecosystemName
	exportRes := NewAppResource(app, "roundtrip-domain", "roundtrip-eco")

	yamlBytes, err := h.ToYAML(exportRes)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	yamlStr := string(yamlBytes)

	// RED: The exported YAML must contain metadata.ecosystem after the fix
	if !strings.Contains(yamlStr, "ecosystem: roundtrip-eco") {
		t.Errorf("exported App YAML must contain 'ecosystem: roundtrip-eco' for context-free apply; got:\n%s", yamlStr)
	}

	// --- Destination store: fresh, no active context ---
	destStore := db.NewMockDataStore()

	dstEco := &models.Ecosystem{Name: "roundtrip-eco"}
	if err := destStore.CreateEcosystem(dstEco); err != nil {
		t.Fatalf("dest CreateEcosystem: %v", err)
	}
	dstDomain := &models.Domain{Name: "roundtrip-domain", EcosystemID: dstEco.ID}
	if err := destStore.CreateDomain(dstDomain); err != nil {
		t.Fatalf("dest CreateDomain: %v", err)
	}

	// No active context set on destination
	destCtx := resource.Context{DataStore: destStore}

	_, err = h.Apply(destCtx, yamlBytes)
	if err != nil {
		t.Fatalf("Apply() on fresh store should succeed with metadata.ecosystem in YAML, got: %v", err)
	}
}
