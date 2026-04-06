package handlers

// Tests for issue #181: NvimTheme/NvimPackage List() must NOT include embedded
// library items in the export output. Only user-configured (DB/store-backed)
// items should appear in List() results.
//
// These tests are written BEFORE the fix and are expected to FAIL, driving the
// implementation in the TDD workflow.

import (
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/resource"
	theme "github.com/rmkohlman/MaestroTheme"
)

// knownLibraryThemes are names that exist in the embedded library.
// These should NEVER appear in List() unless the user has explicitly configured them.
var knownLibraryThemes = []string{
	"coolnight-ocean",
	"catppuccin-mocha",
	"catppuccin-latte",
	"dracula",
	"nord",
	"tokyonight-night",
	"gruvbox-dark",
	"kanagawa",
	"onedark",
	"rose-pine",
}

// knownLibraryPackages are names that exist in the embedded NvimPackage library.
// These should NEVER appear in List() unless the user has explicitly configured them.
var knownLibraryPackages = []string{
	"core",
	"full",
	"maestro",
	"go-dev",
	"python-dev",
	"maestro-go",
}

// ---------------------------------------------------------------------------
// TestNvimThemeList_ExcludesLibraryItems
//
// Set up a MemoryStore with only user-configured themes (none overlap with
// library names). Call List(). Verify the result count equals ONLY the number
// of user themes — no library items smuggled in.
// ---------------------------------------------------------------------------
func TestNvimThemeList_ExcludesLibraryItems(t *testing.T) {
	h := NewNvimThemeHandler()
	memStore := theme.NewMemoryStore()

	// Add two user themes whose names do NOT exist in the embedded library.
	userThemes := []struct{ name, repo string }{
		{"my-custom-dark", "acme/my-custom-dark"},
		{"my-custom-light", "acme/my-custom-light"},
	}
	for _, ut := range userThemes {
		if err := memStore.Save(&theme.Theme{
			Name:   ut.name,
			Plugin: theme.ThemePlugin{Repo: ut.repo},
		}); err != nil {
			t.Fatalf("Save(%s) error = %v", ut.name, err)
		}
	}

	ctx := resource.Context{ThemeStore: memStore}

	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// The result must contain ONLY the two user themes — not the ~34 library themes.
	if len(resources) != len(userThemes) {
		t.Errorf("List() returned %d resources, want exactly %d (user themes only, no library items)",
			len(resources), len(userThemes))
	}

	// Double-check: no known library theme name should appear in the result.
	resultNames := make(map[string]bool, len(resources))
	for _, res := range resources {
		resultNames[res.GetName()] = true
	}

	for _, libName := range knownLibraryThemes {
		if resultNames[libName] {
			t.Errorf("List() included library theme %q — library items must be excluded from export", libName)
		}
	}

	// Verify user themes ARE present.
	for _, ut := range userThemes {
		if !resultNames[ut.name] {
			t.Errorf("List() missing user theme %q", ut.name)
		}
	}
}

// ---------------------------------------------------------------------------
// TestNvimPackageList_ExcludesLibraryItems
//
// Set up a MockDataStore with only user-configured packages. Call List().
// Verify the result count equals ONLY the number of DB packages — no library
// items appended.
// ---------------------------------------------------------------------------
func TestNvimPackageList_ExcludesLibraryItems(t *testing.T) {
	h := NewNvimPackageHandler()
	mockStore := db.NewMockDataStore()

	// Add two user packages whose names do NOT exist in the embedded library.
	userPkgNames := []string{"my-user-pkg-a", "my-user-pkg-b"}
	for i, name := range userPkgNames {
		dbPkg := &models.NvimPackageDB{
			ID:   i + 1,
			Name: name,
		}
		if err := dbPkg.SetPlugins([]string{"telescope"}); err != nil {
			t.Fatalf("SetPlugins(%s) error = %v", name, err)
		}
		mockStore.Packages[name] = dbPkg
	}

	ctx := resource.Context{DataStore: mockStore}

	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// The result must contain ONLY the two user packages — not the ~12 library packages.
	if len(resources) != len(userPkgNames) {
		t.Errorf("List() returned %d resources, want exactly %d (user packages only, no library items)",
			len(resources), len(userPkgNames))
	}

	// Double-check: no known library package name should appear in the result.
	resultNames := make(map[string]bool, len(resources))
	for _, res := range resources {
		resultNames[res.GetName()] = true
	}

	for _, libName := range knownLibraryPackages {
		if resultNames[libName] {
			t.Errorf("List() included library package %q — library items must be excluded from export", libName)
		}
	}

	// Verify user packages ARE present.
	for _, name := range userPkgNames {
		if !resultNames[name] {
			t.Errorf("List() missing user package %q", name)
		}
	}
}

// ---------------------------------------------------------------------------
// TestNvimThemeRoundTrip_NoLibraryPollution
//
// Round-trip: export via List() → apply each item back via Apply() → verify
// the store contains ONLY user items (library items must not be written to the
// store on restore).
// ---------------------------------------------------------------------------
func TestNvimThemeRoundTrip_NoLibraryPollution(t *testing.T) {
	h := NewNvimThemeHandler()

	// --- Export phase ---
	exportStore := theme.NewMemoryStore()

	// Add one user theme that does NOT exist in the library.
	userThemeName := "roundtrip-user-theme"
	if err := exportStore.Save(&theme.Theme{
		Name:        userThemeName,
		Description: "User configured round-trip theme",
		Plugin:      theme.ThemePlugin{Repo: "acme/roundtrip-theme"},
	}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	exportCtx := resource.Context{ThemeStore: exportStore}

	exported, err := h.List(exportCtx)
	if err != nil {
		t.Fatalf("List() (export) error = %v", err)
	}

	// Assertion 1: export must contain only the user theme — no library items.
	if len(exported) != 1 {
		t.Errorf("export List() returned %d items, want exactly 1 (user theme only, no library pollution)",
			len(exported))
	}

	// Check no known library themes leaked into the export.
	exportedNames := make(map[string]bool, len(exported))
	for _, res := range exported {
		exportedNames[res.GetName()] = true
	}
	for _, libName := range knownLibraryThemes {
		if exportedNames[libName] {
			t.Errorf("export contained library theme %q — must not appear in export", libName)
		}
	}

	// --- Restore phase ---
	// Serialize each exported item and apply it into a fresh store.
	restoreStore := theme.NewMemoryStore()
	restoreCtx := resource.Context{ThemeStore: restoreStore}

	for _, res := range exported {
		yamlBytes, err := h.ToYAML(res)
		if err != nil {
			t.Fatalf("ToYAML(%s) error = %v", res.GetName(), err)
		}
		if _, err := h.Apply(restoreCtx, yamlBytes); err != nil {
			t.Fatalf("Apply(%s) error = %v", res.GetName(), err)
		}
	}

	// Assertion 2: restore store must contain ONLY the user theme.
	restored, err := restoreStore.List()
	if err != nil {
		t.Fatalf("restoreStore.List() error = %v", err)
	}

	if len(restored) != 1 {
		t.Errorf("after restore, DB contains %d items, want exactly 1 — library items polluted the DB", len(restored))
	}

	// Verify the one item is the user theme, not a library item.
	restoredNames := make(map[string]bool, len(restored))
	for _, t := range restored {
		restoredNames[t.Name] = true
	}
	if !restoredNames[userThemeName] {
		t.Errorf("user theme %q missing from restored DB", userThemeName)
	}
	for _, libName := range knownLibraryThemes {
		if restoredNames[libName] {
			t.Errorf("library theme %q was written to DB on restore — this is DB pollution", libName)
		}
	}
}

// ---------------------------------------------------------------------------
// TestNvimPackageRoundTrip_NoLibraryPollution
//
// Round-trip: export via List() → apply each item back via Apply() → verify
// the DB (mock store) contains ONLY user items (library items must not be
// written to the store on restore).
// ---------------------------------------------------------------------------
func TestNvimPackageRoundTrip_NoLibraryPollution(t *testing.T) {
	h := NewNvimPackageHandler()

	// --- Export phase ---
	exportStore := db.NewMockDataStore()

	// Add one user package that does NOT exist in the library.
	userPkgName := "roundtrip-user-pkg"
	dbPkg := &models.NvimPackageDB{
		ID:   1,
		Name: userPkgName,
	}
	dbPkg.Description.String = "User-configured round-trip package"
	dbPkg.Description.Valid = true
	if err := dbPkg.SetPlugins([]string{"telescope", "treesitter"}); err != nil {
		t.Fatalf("SetPlugins() error = %v", err)
	}
	exportStore.Packages[userPkgName] = dbPkg

	exportCtx := resource.Context{DataStore: exportStore}

	exported, err := h.List(exportCtx)
	if err != nil {
		t.Fatalf("List() (export) error = %v", err)
	}

	// Assertion 1: export must contain only the user package — no library items.
	if len(exported) != 1 {
		t.Errorf("export List() returned %d items, want exactly 1 (user package only, no library pollution)",
			len(exported))
	}

	// Check no known library packages leaked into the export.
	exportedNames := make(map[string]bool, len(exported))
	for _, res := range exported {
		exportedNames[res.GetName()] = true
	}
	for _, libName := range knownLibraryPackages {
		if exportedNames[libName] {
			t.Errorf("export contained library package %q — must not appear in export", libName)
		}
	}

	// --- Restore phase ---
	// Serialize each exported item and apply it into a fresh store.
	restoreStore := db.NewMockDataStore()
	restoreStore.GetPackageErr = nil // packages not found by default returns nil from map lookup
	restoreCtx := resource.Context{DataStore: restoreStore}

	for _, res := range exported {
		yamlBytes, err := h.ToYAML(res)
		if err != nil {
			t.Fatalf("ToYAML(%s) error = %v", res.GetName(), err)
		}
		if _, err := h.Apply(restoreCtx, yamlBytes); err != nil {
			t.Fatalf("Apply(%s) error = %v", res.GetName(), err)
		}
	}

	// Assertion 2: restore store must contain ONLY the user package.
	restoredPkgs, err := restoreStore.ListPackages()
	if err != nil {
		t.Fatalf("restoreStore.ListPackages() error = %v", err)
	}

	if len(restoredPkgs) != 1 {
		t.Errorf("after restore, DB contains %d packages, want exactly 1 — library items polluted the DB",
			len(restoredPkgs))
	}

	// Verify the one item is the user package, not a library item.
	restoredNames := make(map[string]bool, len(restoredPkgs))
	for _, p := range restoredPkgs {
		restoredNames[p.Name] = true
	}
	if !restoredNames[userPkgName] {
		t.Errorf("user package %q missing from restored DB", userPkgName)
	}
	for _, libName := range knownLibraryPackages {
		if restoredNames[libName] {
			t.Errorf("library package %q was written to DB on restore — this is DB pollution", libName)
		}
	}
}
