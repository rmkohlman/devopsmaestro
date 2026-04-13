// Package cmd — TDD Phase 2 (RED) tests for issue #176
//
// Bug: dvm get app <name> -o yaml calls app.ToYAML() directly (model method),
// bypassing the resource handler's ToYAML() which enriches the YAML with
// metadata.ecosystem. The singular app export is missing the ecosystem field,
// making it non-restorable without an active ecosystem context.
//
// These tests MUST FAIL until the fix is implemented (Phase 3).
// Root cause: cmd/app.go ~line 498 calls app.ToYAML(domain.Name, wsNames)
// directly instead of routing through handlers.AppHandler.ToYAML(res).
package cmd

import (
	"bytes"
	"context"
	"database/sql"
	"testing"

	"devopsmaestro/models"
	"devopsmaestro/pkg/resource/handlers"
	"github.com/rmkohlman/MaestroSDK/render"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newSingularGetAppTestCmd creates a minimal cobra.Command with the given
// DataStore injected into its context and the --domain flag wired, mirroring
// the flag set registered in app.go init().
func newSingularGetAppTestCmd(t *testing.T, ds interface{}) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "test"}
	ctx := context.WithValue(context.Background(), CtxKeyDataStore, ds)
	cmd.SetContext(ctx)
	cmd.Flags().StringP("domain", "d", "", "Domain name (defaults to active domain)")
	cmd.Flags().Bool("show-theme", false, "")
	return cmd
}

// captureAppYAML calls getApp with the given cmd and name, capturing YAML
// output from the render writer and returning it as raw bytes.
func captureAppYAML(t *testing.T, cmd *cobra.Command, appName string) []byte {
	t.Helper()

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "yaml"

	err := getApp(cmd, appName)
	require.NoError(t, err, "getApp must not error for app %q", appName)

	return buf.Bytes()
}

// seedAppHierarchy seeds a full ecosystem → domain → app hierarchy in ds and
// sets the active domain context so getApp can resolve the app by name.
// Returns the ecosystem, domain, and app models.
func seedAppHierarchy(t *testing.T, ds interface {
	CreateEcosystem(*models.Ecosystem) error
	CreateDomain(*models.Domain) error
	CreateApp(*models.App) error
	SetActiveEcosystem(*int) error
	SetActiveDomain(*int) error
}) (eco *models.Ecosystem, dom *models.Domain, app *models.App) {
	t.Helper()

	eco = &models.Ecosystem{Name: "singular-yaml-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))

	dom = &models.Domain{Name: "singular-yaml-dom", EcosystemID: sql.NullInt64{Int64: int64(eco.ID), Valid: true}}
	require.NoError(t, ds.CreateDomain(dom))

	app = &models.App{
		Name:     "singular-yaml-app",
		Path:     "/srv/singular",
		DomainID: sql.NullInt64{Int64: int64(dom.ID), Valid: true},
	}
	require.NoError(t, ds.CreateApp(app))

	// Set active context so getApp can resolve by name without --domain flag
	require.NoError(t, ds.SetActiveEcosystem(&eco.ID))
	require.NoError(t, ds.SetActiveDomain(&dom.ID))

	return eco, dom, app
}

// ---------------------------------------------------------------------------
// TestSingularGetApp_YAML_IncludesEcosystem  [Issue #176 — RED]
//
// dvm get app <name> -o yaml must include metadata.ecosystem in its output.
// Currently FAILS because cmd/app.go ~line 498 calls app.ToYAML(domain.Name, wsNames)
// directly, which never sets metadata.ecosystem. The handler's ToYAML() sets it
// correctly (pkg/resource/handlers/app.go line 214-216), but is bypassed.
// ---------------------------------------------------------------------------

func TestSingularGetApp_YAML_IncludesEcosystem(t *testing.T) {
	// Ensure handlers are registered (required for resource.Get)
	handlers.RegisterAll()

	ds := createFullTestDataStore(t)
	defer ds.Close()

	eco, _, _ := seedAppHierarchy(t, ds)

	cmd := newSingularGetAppTestCmd(t, ds)

	out := captureAppYAML(t, cmd, "singular-yaml-app")

	require.NotEmpty(t, out, "YAML output must not be empty")

	// Parse the YAML
	var result map[string]interface{}
	err := yaml.Unmarshal(out, &result)
	require.NoError(t, err, "singular app YAML output must be valid YAML; got:\n%s", string(out))

	// Verify basic structure
	assert.Equal(t, "App", result["kind"],
		"singular app YAML must have kind: App; got:\n%s", string(out))
	assert.Equal(t, "devopsmaestro.io/v1", result["apiVersion"],
		"singular app YAML must have apiVersion: devopsmaestro.io/v1; got:\n%s", string(out))

	// --- THE KEY ASSERTION (BUG #176): metadata.ecosystem must be present ---
	metadata, ok := result["metadata"].(map[string]interface{})
	require.True(t, ok, "singular app YAML must have a 'metadata' map; got:\n%s", string(out))

	ecosystemVal, hasEcosystem := metadata["ecosystem"]
	assert.True(t, hasEcosystem,
		"BUG #176: singular app YAML is missing metadata.ecosystem — "+
			"cmd/app.go calls app.ToYAML() directly, bypassing the handler's "+
			"ToYAML() which sets ecosystemName; got:\n%s", string(out))

	if hasEcosystem {
		assert.Equal(t, eco.Name, ecosystemVal,
			"metadata.ecosystem must equal the seeded ecosystem name %q; got:\n%s",
			eco.Name, string(out))
	}
}

// ---------------------------------------------------------------------------
// TestSingularGetApp_YAML_Restorable  [Issue #176 — RED]
//
// The YAML produced by dvm get app <name> -o yaml must be self-contained
// enough that the app handler's Apply() can parse it back without needing
// an active ecosystem context. This requires metadata.ecosystem to be present.
//
// Currently FAILS because the ecosystem field is absent from the output,
// so Apply() would fall through to the active-context branch — meaning the
// output is NOT independently restorable.
// ---------------------------------------------------------------------------

func TestSingularGetApp_YAML_Restorable(t *testing.T) {
	// Ensure handlers are registered
	handlers.RegisterAll()

	ds := createFullTestDataStore(t)
	defer ds.Close()

	eco, _, _ := seedAppHierarchy(t, ds)

	cmd := newSingularGetAppTestCmd(t, ds)

	out := captureAppYAML(t, cmd, "singular-yaml-app")

	require.NotEmpty(t, out, "YAML output must not be empty")

	// Parse the output YAML into AppYAML struct (mirrors what Apply() does)
	var appYAML models.AppYAML
	err := yaml.Unmarshal(out, &appYAML)
	require.NoError(t, err, "singular app YAML must be parseable as AppYAML; got:\n%s", string(out))

	// --- THE KEY ASSERTION (BUG #176): ecosystem must be populated so Apply()
	// can resolve the domain without needing active context ---
	assert.NotEmpty(t, appYAML.Metadata.Ecosystem,
		"BUG #176: singular app YAML has empty metadata.ecosystem — "+
			"Apply() would fall through to active-context branch, making this "+
			"YAML non-restorable without 'dvm use ecosystem' first; got YAML:\n%s",
		string(out))

	// If ecosystem is populated, verify it matches the correct ecosystem
	if appYAML.Metadata.Ecosystem != "" {
		assert.Equal(t, eco.Name, appYAML.Metadata.Ecosystem,
			"metadata.ecosystem must match the app's actual ecosystem %q", eco.Name)
	}

	// Also verify domain is present (required by Apply())
	assert.NotEmpty(t, appYAML.Metadata.Domain,
		"singular app YAML must include metadata.domain (required by Apply())")

	// Verify name is present
	assert.Equal(t, "singular-yaml-app", appYAML.Metadata.Name,
		"singular app YAML must include the correct metadata.name")
}

// ---------------------------------------------------------------------------
// TestSingularGetApp_YAML_EcosystemMatchesHandlerOutput  [Issue #176 — RED]
//
// The YAML produced by cmd/app.go (singular get) must match the YAML produced
// by handlers.AppHandler.ToYAML() (the correct path used by get all).
// This test uses the handler directly as a reference and compares ecosystem
// metadata. Currently FAILS because the two paths diverge at ecosystem.
// ---------------------------------------------------------------------------

func TestSingularGetApp_YAML_EcosystemMatchesHandlerOutput(t *testing.T) {
	// Ensure handlers are registered
	handlers.RegisterAll()

	ds := createFullTestDataStore(t)
	defer ds.Close()

	eco, dom, app := seedAppHierarchy(t, ds)

	// --- Reference path: produce YAML via the handler (correct path) ---
	handlerRef := handlers.NewAppHandler()
	appRes := handlers.NewAppResource(app, dom.Name, eco.Name)
	handlerYAMLBytes, err := handlerRef.ToYAML(appRes)
	require.NoError(t, err, "handler.ToYAML must not error")

	var handlerYAML models.AppYAML
	require.NoError(t, yaml.Unmarshal(handlerYAMLBytes, &handlerYAML))

	// Confirm the handler output includes ecosystem (the correct behavior)
	require.Equal(t, eco.Name, handlerYAML.Metadata.Ecosystem,
		"handler ToYAML must include ecosystem (reference path sanity check)")

	// --- Bug path: produce YAML via cmd getApp (the broken path) ---
	cmd := newSingularGetAppTestCmd(t, ds)
	cmdOut := captureAppYAML(t, cmd, "singular-yaml-app")

	var cmdYAML models.AppYAML
	require.NoError(t, yaml.Unmarshal(cmdOut, &cmdYAML),
		"cmd YAML output must be parseable; got:\n%s", string(cmdOut))

	// --- THE KEY ASSERTION (BUG #176): both paths must agree on ecosystem ---
	assert.Equal(t, handlerYAML.Metadata.Ecosystem, cmdYAML.Metadata.Ecosystem,
		"BUG #176: cmd singular get produces ecosystem=%q but handler produces "+
			"ecosystem=%q — the two paths are divergent; cmd/app.go must route "+
			"through the handler's ToYAML() instead of calling app.ToYAML() directly",
		cmdYAML.Metadata.Ecosystem, handlerYAML.Metadata.Ecosystem)
}
