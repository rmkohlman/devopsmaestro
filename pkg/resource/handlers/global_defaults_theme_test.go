package handlers

// =============================================================================
// TDD Phase 2 (RED): Bug #174 — GlobalDefaults YAML export missing theme
//
// Root cause:
//   - globalDefaultsSpec has no Theme field
//   - loadGlobalDefaults() does not read the "theme" key from the defaults table
//   - ToYAML() does not include theme in the spec output
//   - Apply() does not restore theme from the YAML spec
//
// Fix required (per issue #174):
//   1. Add `Theme string` field to globalDefaultsSpec
//   2. Update loadGlobalDefaults() to read "theme" from the defaults table
//   3. Update ToYAML() to include theme in the serialized spec
//   4. Update Apply() to restore theme via ds.SetDefault("theme", ...)
//
// ALL tests in this file MUST FAIL until the fix is implemented.
// =============================================================================

import (
	"strings"
	"testing"

	"devopsmaestro/db"
	"github.com/rmkohlman/MaestroSDK/resource"
)

// =============================================================================
// Test 1: ToYAML must include the theme when the defaults store has one set
//
// RED: GlobalDefaultsResource has no theme field. loadGlobalDefaults() never
// reads "theme" from the store. ToYAML() therefore never emits a "theme:" key
// in the spec. The assertion below will fail.
// =============================================================================

func TestGlobalDefaultsHandler_ToYAML_IncludesTheme(t *testing.T) {
	h := NewGlobalDefaultsHandler()
	store := db.NewMockDataStore()

	// Seed the defaults store with a global theme
	if err := store.SetDefault("theme", "catppuccin"); err != nil {
		t.Fatalf("SetDefault(theme): %v", err)
	}

	ctx := resource.Context{DataStore: store}

	// Get() calls loadGlobalDefaults() internally and returns a
	// GlobalDefaultsResource. After the fix, that resource will carry the theme.
	res, err := h.Get(ctx, "global-defaults")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Serialize to YAML
	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	yamlStr := string(yamlBytes)

	// RED: loadGlobalDefaults() never reads "theme" → GlobalDefaultsResource
	// carries no theme → ToYAML() never emits "theme:" → this assertion fails.
	if !strings.Contains(yamlStr, "theme:") {
		t.Errorf(
			"ToYAML() output must contain 'theme:' field when a global theme is set\n"+
				"Bug: globalDefaultsSpec has no Theme field and loadGlobalDefaults() "+
				"does not read 'theme' from the defaults table\ngot YAML:\n%s",
			yamlStr,
		)
	}

	if !strings.Contains(yamlStr, "catppuccin") {
		t.Errorf(
			"ToYAML() output must contain the theme value 'catppuccin'\n"+
				"got YAML:\n%s",
			yamlStr,
		)
	}
}

// =============================================================================
// Test 2: Apply must restore the theme from the YAML spec into the defaults store
//
// RED: globalDefaultsSpec has no Theme field. yaml.Unmarshal silently drops the
// "theme: catppuccin" key. Apply() never calls ds.SetDefault("theme", ...).
// After Apply(), the store's "theme" key remains empty.
// =============================================================================

func TestGlobalDefaultsHandler_Apply_RestoresTheme(t *testing.T) {
	h := NewGlobalDefaultsHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	// A GlobalDefaults YAML document that includes a theme in the spec.
	// After the fix, Apply() must write this theme to the defaults store.
	yamlInput := []byte(`apiVersion: devopsmaestro.io/v1
kind: GlobalDefaults
metadata:
  name: global-defaults
spec:
  theme: catppuccin
`)

	_, err := h.Apply(ctx, yamlInput)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	// Check that the theme was written to the defaults store
	got, err := store.GetDefault("theme")
	if err != nil {
		t.Fatalf("GetDefault(theme) after Apply() error = %v", err)
	}

	// RED: globalDefaultsSpec has no Theme field → yaml.Unmarshal drops the key
	// → Apply() never calls SetDefault("theme", ...) → got == ""
	if got != "catppuccin" {
		t.Errorf(
			"Apply() must restore theme to defaults store\n"+
				"Bug: globalDefaultsSpec missing Theme field and Apply() does not "+
				"call SetDefault(\"theme\", ...)\n"+
				"got defaults[\"theme\"] = %q, want %q",
			got, "catppuccin",
		)
	}
}

// =============================================================================
// Test 3: Round-trip — export theme via ToYAML, then restore it via Apply
//
// RED: Both ToYAML() and Apply() are broken for the theme field, so the
// round-trip loses the theme at two points:
//   (a) ToYAML() omits the theme (no Theme field in GlobalDefaultsResource)
//   (b) Apply() ignores the theme key even if it were present in YAML
//
// This test documents the complete round-trip failure.
// =============================================================================

func TestGlobalDefaultsHandler_RoundTrip_Theme(t *testing.T) {
	h := NewGlobalDefaultsHandler()

	// ── SOURCE store: has a theme set ────────────────────────────────────────
	srcStore := db.NewMockDataStore()
	if err := srcStore.SetDefault("theme", "gruvbox"); err != nil {
		t.Fatalf("SetDefault(theme) on source store: %v", err)
	}
	// Also set a build-arg so List() does not return nil (currently List()
	// requires at least one of buildArgs or caCerts to be non-empty).
	// After the fix, List() should also return a resource when only theme is set.
	if err := srcStore.SetDefault("build-args", `{"GOENV":"production"}`); err != nil {
		t.Fatalf("SetDefault(build-args) on source store: %v", err)
	}

	srcCtx := resource.Context{DataStore: srcStore}

	// Export: Get the GlobalDefaults resource, then serialize to YAML
	res, err := h.Get(srcCtx, "global-defaults")
	if err != nil {
		t.Fatalf("Get() from source store error = %v", err)
	}

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	yamlStr := string(yamlBytes)

	// RED (point a): ToYAML() never emits "theme:" because GlobalDefaultsResource
	// has no theme field and loadGlobalDefaults() never reads it.
	if !strings.Contains(yamlStr, "theme:") {
		t.Errorf(
			"ToYAML() round-trip: exported YAML must contain 'theme:' field\n"+
				"Bug: theme not read by loadGlobalDefaults() and not stored in "+
				"GlobalDefaultsResource\ngot YAML:\n%s",
			yamlStr,
		)
	}

	if !strings.Contains(yamlStr, "gruvbox") {
		t.Errorf(
			"ToYAML() round-trip: exported YAML must contain theme value 'gruvbox'\n"+
				"got YAML:\n%s",
			yamlStr,
		)
	}

	// ── DESTINATION store: fresh, no theme set ───────────────────────────────
	dstStore := db.NewMockDataStore()
	dstCtx := resource.Context{DataStore: dstStore}

	// Apply the exported YAML to the destination store
	_, err = h.Apply(dstCtx, yamlBytes)
	if err != nil {
		t.Fatalf("Apply() to destination store error = %v", err)
	}

	// Verify the theme was restored
	restoredTheme, err := dstStore.GetDefault("theme")
	if err != nil {
		t.Fatalf("GetDefault(theme) from destination store error = %v", err)
	}

	// RED (point b): Even if YAML somehow contained "theme: gruvbox", Apply()
	// would still ignore it because globalDefaultsSpec has no Theme field.
	if restoredTheme != "gruvbox" {
		t.Errorf(
			"Round-trip Apply() must restore theme in destination store\n"+
				"Bug: Apply() does not handle the theme field from the YAML spec\n"+
				"got defaults[\"theme\"] = %q, want %q",
			restoredTheme, "gruvbox",
		)
	}
}

// =============================================================================
// Test 4: List must include GlobalDefaults resource when only a theme is set
//
// RED: List() currently returns nil when buildArgs and caCerts are both empty.
// After the fix, a theme-only config should still produce a resource so the
// theme is included in dvm get all -o yaml output.
// =============================================================================

func TestGlobalDefaultsHandler_List_IncludesResourceWhenOnlyThemeSet(t *testing.T) {
	h := NewGlobalDefaultsHandler()
	store := db.NewMockDataStore()

	// Only a theme is set — no build-args, no CA-certs
	if err := store.SetDefault("theme", "tokyonight"); err != nil {
		t.Fatalf("SetDefault(theme): %v", err)
	}

	ctx := resource.Context{DataStore: store}

	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// RED: List() currently checks `len(buildArgs) == 0 && len(caCerts) == 0`
	// and returns nil — it never checks whether a theme is set. So resources
	// will be nil/empty here, causing the assertion below to fail.
	if len(resources) == 0 {
		t.Errorf(
			"List() must return a non-empty slice when only a global theme is set\n" +
				"Bug: List() only checks buildArgs and caCerts, ignoring the theme\n" +
				"got len(resources) = 0",
		)
	}
}
