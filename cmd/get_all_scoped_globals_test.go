package cmd

// =============================================================================
// TDD Phase 2 (RED): Bug #131 — Scoped get all drops global resources silently
//
// get_all.go lines 175-226: GitRepos, Registries, NvimPlugins, NvimThemes,
// NvimPackages, TerminalPrompts, TerminalPackages are all gated by:
//   if scope.ShowAll { ... }
// When ShowAll=false (scoped to ecosystem/domain/app), these resources are
// entirely omitted with no warning.
//
// Fix design:
//   - Emit a warning when `-o yaml` or `-o json` is used without `-A`
//     (Decision 9: exclude globals from scoped exports, but warn user)
//   - Table output: global resources always show (existing behavior OK)
//
// These tests verify:
//   1. ShowAll=true includes global resources in YAML/JSON output
//   2. ShowAll=false OMITS global resources (this is the intended behavior
//      per Decision 9, but the lack of warning is the bug)
//   3. A warning is emitted when -o yaml/json is used without -A
//
// Tests #1 and #2 document the CURRENT behavior (they pass today).
// Test #3 WILL FAIL until the warning is implemented.
// =============================================================================

import (
	"bytes"
	"strings"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetAll_ShowAllTrue_IncludesGlobalResources verifies that when ShowAll=true
// (-A flag), global resources (Registries, GitRepos, NvimPlugins, etc.) appear
// in the YAML/JSON output.
//
// GREEN: This already passes — documenting correct ShowAll=true behavior.
func TestGetAll_ShowAllTrue_IncludesGlobalResources(t *testing.T) {
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	// Seed a registry (global resource)
	reg := &models.Registry{
		Name:    "show-all-reg",
		Type:    "zot",
		Port:    5400,
		Storage: "/tmp/reg",
		Version: "2.1.0",
		Status:  "stopped",
	}
	require.NoError(t, dataStore.CreateRegistry(reg))

	// Seed a git repo (global resource)
	gitRepo := &models.GitRepoDB{
		Name:       "show-all-repo",
		URL:        "https://github.com/org/repo.git",
		Slug:       "github.com_org_repo",
		DefaultRef: "main",
		AuthType:   "none",
		SyncStatus: "pending",
	}
	require.NoError(t, dataStore.CreateGitRepo(gitRepo))

	cmd := newScopedGetAllCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "yaml"

	// -A flag → ShowAll=true → global resources included
	cmd.SetArgs([]string{"-A"})
	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "show-all-reg",
		"global registry must appear in YAML output when -A is used")
	assert.Contains(t, output, "show-all-repo",
		"global git repo must appear in YAML output when -A is used")
}

// TestGetAll_ShowAllFalse_OmitsGlobalResources verifies that when scoped
// (ShowAll=false), global resources are NOT in the YAML/JSON output.
// This is Decision 9 behavior — excluded from scoped exports.
//
// GREEN: This already passes — the omission is the current (intentional)
// behavior. The BUG is the missing warning (see TestGetAll_ScopedYAML_EmitsWarning).
func TestGetAll_ShowAllFalse_OmitsGlobalResources(t *testing.T) {
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	// Seed an ecosystem so we can scope to it
	eco := &models.Ecosystem{Name: "scope-only-eco"}
	require.NoError(t, dataStore.CreateEcosystem(eco))

	// Seed global resources that should be omitted when scoped
	reg := &models.Registry{
		Name:    "scope-omit-reg",
		Type:    "zot",
		Port:    5500,
		Storage: "/tmp/reg",
		Version: "2.1.0",
		Status:  "stopped",
	}
	require.NoError(t, dataStore.CreateRegistry(reg))

	gitRepo := &models.GitRepoDB{
		Name:       "scope-omit-repo",
		URL:        "https://github.com/org/omit.git",
		Slug:       "github.com_org_omit",
		DefaultRef: "main",
		AuthType:   "none",
		SyncStatus: "pending",
	}
	require.NoError(t, dataStore.CreateGitRepo(gitRepo))

	cmd := newScopedGetAllCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "yaml"

	// Scoped to ecosystem — ShowAll=false → global resources excluded
	cmd.SetArgs([]string{"--ecosystem", "scope-only-eco"})
	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.NotContains(t, output, "scope-omit-reg",
		"global registry should NOT appear in scoped YAML export (Decision 9)")
	assert.NotContains(t, output, "scope-omit-repo",
		"global git repo should NOT appear in scoped YAML export (Decision 9)")
}

// TestGetAll_ScopedYAML_FewerResourcesThanShowAll verifies that the count of
// resources in a scoped YAML export is strictly less than in an unscoped export
// when global resources exist.
//
// RED: This test documents the BUG — when global resources exist, a scoped
// export silently drops them with no count feedback or warning to the user.
// After fix, a warning MUST be emitted (see TestGetAll_ScopedYAML_EmitsWarning).
func TestGetAll_ScopedYAML_FewerResourcesThanShowAll(t *testing.T) {
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	// Seed ecosystem (hierarchical — appears in both scoped and unscoped)
	eco := &models.Ecosystem{Name: "count-eco"}
	require.NoError(t, dataStore.CreateEcosystem(eco))

	// Seed global resources (only appear with ShowAll=true)
	reg := &models.Registry{
		Name:    "count-reg",
		Type:    "zot",
		Port:    5600,
		Storage: "/tmp/reg",
		Version: "2.1.0",
		Status:  "stopped",
	}
	require.NoError(t, dataStore.CreateRegistry(reg))

	var scopedCount, unscopedCount int

	// --- Scoped run (ecosystem only, no globals) ---
	{
		cmd := newScopedGetAllCmd(t, dataStore)

		var buf bytes.Buffer
		render.SetWriter(&buf)

		origFormat := getOutputFormat
		defer func() { getOutputFormat = origFormat }()
		getOutputFormat = "yaml"

		cmd.SetArgs([]string{"--ecosystem", "count-eco"})
		require.NoError(t, cmd.Execute())

		output := buf.String()
		scopedCount = strings.Count(output, "kind:")
		render.SetWriter(render.GetWriter()) // restore
	}

	// --- Unscoped run (all resources including globals) ---
	{
		cmd := newScopedGetAllCmd(t, dataStore)

		var buf bytes.Buffer
		render.SetWriter(&buf)

		origFormat := getOutputFormat
		getOutputFormat = "yaml"
		defer func() { getOutputFormat = origFormat }()

		cmd.SetArgs([]string{"-A"})
		require.NoError(t, cmd.Execute())

		output := buf.String()
		unscopedCount = strings.Count(output, "kind:")
		render.SetWriter(render.GetWriter()) // restore
	}

	// The scoped export must have fewer items than the full export
	assert.Less(t, scopedCount, unscopedCount,
		"scoped YAML export should have fewer resources than ShowAll=true export "+
			"(global resources only appear with -A)")
}

// TestGetAll_ScopedYAML_EmitsWarning verifies that a warning is printed when
// `-o yaml` (or `-o json`) is used without `-A`, informing the user that
// global resources are excluded from the export.
//
// RED: Will fail until the warning is implemented in getAll().
// The warning message should tell users to add -A to include global resources.
func TestGetAll_ScopedYAML_EmitsWarning(t *testing.T) {
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	// Seed an ecosystem so we can scope to it
	eco := &models.Ecosystem{Name: "warn-eco"}
	require.NoError(t, dataStore.CreateEcosystem(eco))

	// Seed a global resource to make the omission meaningful
	reg := &models.Registry{
		Name:    "warn-reg",
		Type:    "zot",
		Port:    5700,
		Storage: "/tmp/reg",
		Version: "2.1.0",
		Status:  "stopped",
	}
	require.NoError(t, dataStore.CreateRegistry(reg))

	cmd := newScopedGetAllCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "yaml"

	// Scoped YAML export without -A
	cmd.SetArgs([]string{"--ecosystem", "warn-eco"})
	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()

	// RED: A warning must be emitted telling the user that global resources
	// are excluded and -A should be used to include them.
	// The exact wording is flexible — just needs to mention global resources and -A.
	hasWarning := strings.Contains(output, "-A") ||
		strings.Contains(output, "global") ||
		strings.Contains(output, "warning") ||
		strings.Contains(output, "Warning") ||
		strings.Contains(strings.ToLower(output), "excluded")

	assert.True(t, hasWarning,
		"scoped YAML/JSON export without -A should emit a warning about excluded global resources; got output:\n%s", output)
}

// TestGetAll_ScopedJSON_EmitsWarning verifies the same warning for -o json.
//
// RED: Will fail until the warning is implemented.
func TestGetAll_ScopedJSON_EmitsWarning(t *testing.T) {
	dataStore := createFullTestDataStore(t)
	defer dataStore.Close()

	eco := &models.Ecosystem{Name: "warn-json-eco"}
	require.NoError(t, dataStore.CreateEcosystem(eco))

	reg := &models.Registry{
		Name:    "warn-json-reg",
		Type:    "zot",
		Port:    5800,
		Storage: "/tmp/reg",
		Version: "2.1.0",
		Status:  "stopped",
	}
	require.NoError(t, dataStore.CreateRegistry(reg))

	cmd := newScopedGetAllCmd(t, dataStore)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "json"

	cmd.SetArgs([]string{"--ecosystem", "warn-json-eco"})
	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()

	// For JSON output, the warning should be on stderr or a warning header.
	// The JSON payload itself must remain valid — warning should go to stderr
	// or be a separate non-JSON line before the JSON.
	// We check that something indicating global resource exclusion was communicated.
	hasWarning := strings.Contains(output, "-A") ||
		strings.Contains(output, "global") ||
		strings.Contains(output, "warning") ||
		strings.Contains(output, "Warning") ||
		strings.Contains(strings.ToLower(output), "excluded")

	assert.True(t, hasWarning,
		"scoped JSON export without -A should emit a warning about excluded global resources; got output:\n%s", output)
}

// TestGetAll_MockStore_GlobalResourcesGatedByShowAll verifies via unit test
// (using MockDataStore) that the getAll() code path correctly gates global
// resources via scope.ShowAll — this is the internal mechanism of Bug #131.
//
// This test uses a MockDataStore with a Registry seeded but scoped to an
// ecosystem (ShowAll=false). It verifies the Registry is absent from YAML output.
// The companion test (TestGetAll_ScopedYAML_EmitsWarning) verifies the warning.
//
// PARTIAL GREEN: The scoping already works correctly — globals are excluded.
// The missing piece is the warning (covered by TestGetAll_ScopedYAML_EmitsWarning).
func TestGetAll_MockStore_GlobalResourcesGatedByShowAll(t *testing.T) {
	mock := db.NewMockDataStore()

	eco := &models.Ecosystem{ID: 1, Name: "mock-eco"}
	mock.Ecosystems["mock-eco"] = eco

	// Seed a registry in the mock (keyed by name, as MockDataStore.Registries is map[string])
	mock.Registries["mock-reg"] = &models.Registry{
		Name:    "mock-reg",
		Type:    "zot",
		Port:    5900,
		Storage: "/tmp/reg",
		Version: "2.1.0",
		Status:  "stopped",
	}

	// Set active ecosystem (triggers ShowAll=false path)
	ecoID := 1
	mock.Context = &models.Context{ID: 1, ActiveEcosystemID: &ecoID}

	cmd := newGetAllTestCmd(t, mock)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "yaml"

	err := getAll(cmd)
	require.NoError(t, err)

	output := buf.String()

	// Global registry should NOT be in scoped YAML (Decision 9)
	assert.NotContains(t, output, "mock-reg",
		"global registry should be excluded from scoped YAML export")
}
