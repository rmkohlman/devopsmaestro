// Package cmd — TDD Phase 2 (RED) tests for issue #154
//
// Bug: Plural get commands (dvm get ecosystems, dvm get workspaces, etc.) produce
// raw YAML arrays without a kind: List wrapper, making the output unparseable by
// "dvm apply -f".
//
// These tests MUST FAIL until the fix is implemented (Phase 3).
package cmd

import (
	"bytes"
	"context"
	"database/sql"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/render"
	sdkresource "github.com/rmkohlman/MaestroSDK/resource"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newPluralGetTestCmd creates a minimal cobra.Command with the given DataStore injected
// into its context, mirroring the pattern used by get_all_test.go.
func newPluralGetTestCmd(t *testing.T, ds db.DataStore) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "test"}
	ctx := context.WithValue(context.Background(), CtxKeyDataStore, ds)
	cmd.SetContext(ctx)
	// Wire the --all and hierarchy flags expected by getWorkspaces / getEcosystems
	cmd.Flags().Bool("all", false, "")
	return cmd
}

// captureYAML runs fn (which writes to the render writer) and returns the output
// as raw bytes with getOutputFormat set to "yaml".
func captureYAML(t *testing.T, fn func()) []byte {
	t.Helper()
	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "yaml"

	fn()
	return buf.Bytes()
}

// assertIsKindList asserts that the given YAML bytes parse to a map with
// kind == "List" and an "items" sequence — the format required by dvm apply.
func assertIsKindList(t *testing.T, yamlBytes []byte) {
	t.Helper()
	require.NotEmpty(t, yamlBytes, "YAML output must not be empty")

	var result map[string]interface{}
	err := yaml.Unmarshal(yamlBytes, &result)
	require.NoError(t, err, "plural YAML output must be valid YAML; got:\n%s", string(yamlBytes))

	assert.Equal(t, "List", result["kind"],
		"plural YAML output must have kind: List (required by dvm apply); got:\n%s", string(yamlBytes))
	assert.Equal(t, "devopsmaestro.io/v1", result["apiVersion"],
		"plural YAML output must have apiVersion: devopsmaestro.io/v1; got:\n%s", string(yamlBytes))
	_, hasItems := result["items"]
	assert.True(t, hasItems,
		"plural YAML output must have an 'items' key; got:\n%s", string(yamlBytes))
}

// assertIsSingleResource asserts that the given YAML bytes parse to a map with
// a non-empty "kind" field that is NOT "List" — correct for single-resource output.
func assertIsSingleResource(t *testing.T, yamlBytes []byte, expectedKind string) {
	t.Helper()
	require.NotEmpty(t, yamlBytes, "single-resource YAML output must not be empty")

	var result map[string]interface{}
	err := yaml.Unmarshal(yamlBytes, &result)
	require.NoError(t, err, "single-resource YAML output must be valid YAML; got:\n%s", string(yamlBytes))

	assert.Equal(t, expectedKind, result["kind"],
		"single-resource YAML output must have kind: %s; got:\n%s", expectedKind, string(yamlBytes))
}

// assertParsableByDetectKind verifies that the YAML bytes can be parsed by
// resource.DetectKind — this is the exact check the apply pipeline runs.
func assertParsableByDetectKind(t *testing.T, yamlBytes []byte) {
	t.Helper()
	kind, err := sdkresource.DetectKind(yamlBytes)
	assert.NoError(t, err,
		"plural YAML output must be parseable by resource.DetectKind (required by dvm apply); got:\n%s", string(yamlBytes))
	assert.NotEmpty(t, kind,
		"resource.DetectKind must return a non-empty kind; got:\n%s", string(yamlBytes))
}

// ---------------------------------------------------------------------------
// TestGetEcosystems_YAML_ProducesListWrapper  [Issue #154 — RED]
//
// dvm get ecosystems -o yaml must produce a kind: List document, not a raw array.
// Currently FAILS because getEcosystems calls render.OutputWith with a plain slice.
// ---------------------------------------------------------------------------

func TestGetEcosystems_YAML_ProducesListWrapper(t *testing.T) {
	ds := createFullTestDataStore(t)
	defer ds.Close()

	// Seed two ecosystems
	eco1 := &models.Ecosystem{
		Name:        "eco-list-a",
		Description: sql.NullString{String: "First ecosystem", Valid: true},
	}
	require.NoError(t, ds.CreateEcosystem(eco1))

	eco2 := &models.Ecosystem{
		Name:        "eco-list-b",
		Description: sql.NullString{String: "Second ecosystem", Valid: true},
	}
	require.NoError(t, ds.CreateEcosystem(eco2))

	cmd := newPluralGetTestCmd(t, ds)

	out := captureYAML(t, func() {
		err := getEcosystems(cmd)
		require.NoError(t, err, "getEcosystems must not error")
	})

	// Must be a kind: List document — currently produces a raw array (BUG)
	assertIsKindList(t, out)

	// Must be parseable by the apply pipeline
	assertParsableByDetectKind(t, out)

	// Items must contain both ecosystems
	var result map[string]interface{}
	require.NoError(t, yaml.Unmarshal(out, &result))
	items, _ := result["items"].([]interface{})
	assert.GreaterOrEqual(t, len(items), 2,
		"items must contain both seeded ecosystems")
}

// ---------------------------------------------------------------------------
// TestGetEcosystems_YAML_SingleEcosystem_StillProducesList  [Issue #154 — RED]
//
// Even with a single ecosystem, plural get must produce kind: List.
// ---------------------------------------------------------------------------

func TestGetEcosystems_YAML_SingleEcosystem_StillProducesList(t *testing.T) {
	ds := createFullTestDataStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "solo-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))

	cmd := newPluralGetTestCmd(t, ds)

	out := captureYAML(t, func() {
		err := getEcosystems(cmd)
		require.NoError(t, err, "getEcosystems must not error")
	})

	assertIsKindList(t, out)
	assertParsableByDetectKind(t, out)
}

// ---------------------------------------------------------------------------
// TestGetEcosystem_YAML_SingleResource_NotWrappedInList  [Issue #154 — RED]
//
// dvm get ecosystem <name> -o yaml must produce a single unwrapped resource,
// NOT a List. This verifies the fix doesn't over-wrap single resources.
// ---------------------------------------------------------------------------

func TestGetEcosystem_YAML_SingleResource_NotWrappedInList(t *testing.T) {
	ds := createFullTestDataStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "single-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))

	cmd := newPluralGetTestCmd(t, ds)

	out := captureYAML(t, func() {
		err := getEcosystem(cmd, "single-eco")
		require.NoError(t, err, "getEcosystem (singular) must not error")
	})

	// Single resource get must produce kind: Ecosystem, NOT kind: List
	assertIsSingleResource(t, out, "Ecosystem")
}

// ---------------------------------------------------------------------------
// TestGetWorkspaces_YAML_ProducesListWrapper  [Issue #154 — RED]
//
// dvm get workspaces -o yaml must produce a kind: List document.
// Currently FAILS because getWorkspaces calls render.OutputWith with a plain slice.
// ---------------------------------------------------------------------------

func TestGetWorkspaces_YAML_ProducesListWrapper(t *testing.T) {
	ds := createFullTestDataStore(t)
	defer ds.Close()

	// Build hierarchy needed for workspaces
	eco := &models.Ecosystem{Name: "ws-list-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))
	dom := &models.Domain{Name: "ws-list-dom", EcosystemID: sql.NullInt64{Int64: int64(eco.ID), Valid: true}}
	require.NoError(t, ds.CreateDomain(dom))
	app := &models.App{Name: "ws-list-app", Path: "/app", DomainID: sql.NullInt64{Int64: int64(dom.ID), Valid: true}}
	require.NoError(t, ds.CreateApp(app))

	ws1 := &models.Workspace{
		Name:      "ws-a",
		Slug:      "ws-list-eco/ws-list-dom/ws-list-app/ws-a",
		AppID:     app.ID,
		ImageName: "golang:1.22",
		Status:    "stopped",
	}
	require.NoError(t, ds.CreateWorkspace(ws1))

	ws2 := &models.Workspace{
		Name:      "ws-b",
		Slug:      "ws-list-eco/ws-list-dom/ws-list-app/ws-b",
		AppID:     app.ID,
		ImageName: "golang:1.22",
		Status:    "stopped",
	}
	require.NoError(t, ds.CreateWorkspace(ws2))

	// Set active app context so getWorkspaces knows which app to list from
	require.NoError(t, ds.SetActiveEcosystem(&eco.ID))
	require.NoError(t, ds.SetActiveDomain(&dom.ID))
	require.NoError(t, ds.SetActiveApp(&app.ID))

	cmd := newPluralGetTestCmd(t, ds)

	out := captureYAML(t, func() {
		err := getWorkspaces(cmd)
		require.NoError(t, err, "getWorkspaces must not error")
	})

	// Must be a kind: List document — currently produces a raw array (BUG)
	assertIsKindList(t, out)

	// Must be parseable by the apply pipeline
	assertParsableByDetectKind(t, out)
}

// ---------------------------------------------------------------------------
// TestGetEcosystems_YAML_ListDetectKind_ReturnsListKind  [Issue #154 — RED]
//
// Verifies the round-trip: output YAML from plural get → DetectKind → "List"
// This is the exact flow that dvm apply -f <(dvm get ecosystems -o yaml) must support.
// ---------------------------------------------------------------------------

func TestGetEcosystems_YAML_ListDetectKind_ReturnsListKind(t *testing.T) {
	ds := createFullTestDataStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "roundtrip-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))

	cmd := newPluralGetTestCmd(t, ds)

	out := captureYAML(t, func() {
		err := getEcosystems(cmd)
		require.NoError(t, err)
	})

	// The exact check that applyResource() runs on the YAML:
	kind, err := sdkresource.DetectKind(out)
	assert.NoError(t, err,
		"plural get YAML output must be parseable by resource.DetectKind")
	assert.Equal(t, "List", kind,
		"resource.DetectKind must return 'List' for plural get YAML; got %q", kind)
}

// ---------------------------------------------------------------------------
// TestGetEcosystems_YAML_RawArrayIsCurrentBrokenBehavior  [Issue #154 — RED]
//
// This test DOCUMENTS the current broken behavior (raw array) by verifying
// that DetectKind FAILS on the current output. When the bug is fixed, this
// test can be removed.
//
// NOTE: This test is expected to PASS right now (it asserts the bug exists)
// and will FAIL once the fix is in place — at that point it should be deleted.
// ---------------------------------------------------------------------------

func TestGetEcosystems_YAML_RawArrayIsCurrentBrokenBehavior(t *testing.T) {
	ds := createFullTestDataStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "broken-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))

	cmd := newPluralGetTestCmd(t, ds)

	out := captureYAML(t, func() {
		_ = getEcosystems(cmd)
	})

	// Document the current broken behavior: a raw YAML array has no "kind" field.
	// DetectKind should FAIL on the current output.
	_, err := sdkresource.DetectKind(out)
	if err == nil {
		// If this assertion doesn't error, the fix is already in place.
		// This test documents the pre-fix state; skip if already fixed.
		t.Skip("Bug #154 already fixed — raw array behavior no longer present")
	}
	assert.Error(t, err,
		"BUG #154: current plural YAML output is a raw array without 'kind' — "+
			"DetectKind should fail (this test documents the broken state)")
}

// ---------------------------------------------------------------------------
// TestGetEcosystems_YAML_EmptyDB_ProducesEmptyList  [Issue #154 — RED]
//
// Even with no ecosystems, dvm get ecosystems -o yaml must produce kind: List
// with an empty items array — not an empty array or null.
// ---------------------------------------------------------------------------

func TestGetEcosystems_YAML_EmptyDB_ProducesEmptyList(t *testing.T) {
	ds := createFullTestDataStore(t)
	defer ds.Close()

	cmd := newPluralGetTestCmd(t, ds)

	out := captureYAML(t, func() {
		// getEcosystems returns an "empty" render when there are no ecosystems.
		// After the fix, even the empty case should produce kind: List with items: [].
		_ = getEcosystems(cmd)
	})

	if len(out) == 0 {
		// Empty output means the empty-state path was hit (no render for YAML).
		// After the fix, even empty state should produce a List envelope.
		t.Log("INFO: empty YAML output for empty DB — fix should produce kind: List with items: []")
		t.Fail()
		return
	}

	var result map[string]interface{}
	err := yaml.Unmarshal(out, &result)
	require.NoError(t, err, "YAML output must be parseable even for empty DB")

	assert.Equal(t, "List", result["kind"],
		"empty plural get must still produce kind: List; got:\n%s", string(out))
}
