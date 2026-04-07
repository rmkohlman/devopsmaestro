// Package cmd — TDD Phase 2 (RED) tests for issue #183
//
// Bug: Two singular get commands lack proper YAML output:
//
//  1. dvm get credential <name> -o yaml (#183): The command in cmd/get_credential.go
//     RunE only calls render.Plainf() and never checks getOutputFormat — so -o yaml
//     is silently ignored and plain text is printed instead.
//
//  2. dvm get gitrepo <name> -o yaml (#183): runGetGitRepo uses gitRepoToMap() which
//     produces a flat map (name, url, slug, ...) — not the standard apiVersion/kind/
//     metadata/spec structure required by dvm apply.
//
// Both should route through their respective resource handlers for YAML output,
// matching the pattern used by dvm get app <name> -o yaml (fixed in #176).
//
// These tests MUST FAIL until the fix is implemented (Phase 3).
package cmd

import (
	"bytes"
	"context"
	"testing"
	"time"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/credentialbridge"
	"devopsmaestro/pkg/resource/handlers"
	"github.com/rmkohlman/MaestroSDK/render"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// =============================================================================
// Test Helpers
// =============================================================================

// newSingularGetCredentialTestCmd creates a minimal cobra.Command with the
// given DataStore injected into its context and the credential scope flags
// wired, mirroring the flag set registered in get_credential.go init().
func newSingularGetCredentialTestCmd(t *testing.T, ds interface{}) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "test"}
	ctx := context.WithValue(context.Background(), "dataStore", ds)
	cmd.SetContext(ctx)
	cmd.Flags().StringP("ecosystem", "e", "", "Ecosystem scope")
	cmd.Flags().StringP("domain", "d", "", "Domain scope")
	cmd.Flags().StringP("app", "a", "", "App scope")
	cmd.Flags().StringP("workspace", "w", "", "Workspace scope")
	return cmd
}

// newSingularGetGitRepoTestCmd creates a minimal cobra.Command for testing
// runGetGitRepo, with the DataStore injected and the --output flag wired.
func newSingularGetGitRepoTestCmd(t *testing.T, ds interface{}) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "test"}
	ctx := context.WithValue(context.Background(), "dataStore", ds)
	cmd.SetContext(ctx)
	// runGetGitRepo reads "output" directly via cmd.Flags().GetString("output")
	cmd.Flags().StringP("output", "o", "", "Output format")
	return cmd
}

// captureCredentialYAML calls getCredentialCmd.RunE with -o yaml set,
// capturing render writer output and returning raw bytes.
// It exercises the actual RunE of getCredentialCmd by setting getOutputFormat.
func captureCredentialYAML(t *testing.T, cmd *cobra.Command, credName string, scopeFlag, scopeVal string) []byte {
	t.Helper()

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "yaml"

	// Set the scope flag
	require.NoError(t, cmd.Flags().Set(scopeFlag, scopeVal))

	// Wire the RunE from getCredentialCmd and execute it
	cmd.RunE = getCredentialCmd.RunE
	err := cmd.RunE(cmd, []string{credName})
	require.NoError(t, err, "getCredentialCmd.RunE must not error for credential %q", credName)

	return buf.Bytes()
}

// captureGitRepoYAML calls runGetGitRepo with -o yaml, capturing render writer
// output and returning raw bytes.
func captureGitRepoYAML(t *testing.T, cmd *cobra.Command, repoName string) []byte {
	t.Helper()

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	require.NoError(t, cmd.Flags().Set("output", "yaml"))

	err := runGetGitRepo(cmd, []string{repoName})
	require.NoError(t, err, "runGetGitRepo must not error for repo %q", repoName)

	return buf.Bytes()
}

// seedEcosystemCredential seeds an ecosystem-scoped credential and sets the
// active ecosystem context so getCredentialCmd can resolve by scope.
func seedEcosystemCredential(t *testing.T, ds db.DataStore) (*models.Ecosystem, *models.CredentialDB) {
	t.Helper()

	eco := &models.Ecosystem{Name: "singular-yaml-cred-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))
	require.NoError(t, ds.SetActiveEcosystem(&eco.ID))

	envVar := "MY_API_KEY"
	cred := &models.CredentialDB{
		Name:      "singular-yaml-cred",
		ScopeType: models.CredentialScopeEcosystem,
		ScopeID:   int64(eco.ID),
		Source:    "env",
		EnvVar:    &envVar,
	}
	require.NoError(t, ds.CreateCredential(cred))

	return eco, cred
}

// seedGitRepo seeds a simple git repository in the datastore.
func seedGitRepo(t *testing.T, ds db.DataStore) *models.GitRepoDB {
	t.Helper()

	repo := &models.GitRepoDB{
		Name:                "singular-yaml-repo",
		URL:                 "https://github.com/org/repo.git",
		Slug:                "org-repo",
		DefaultRef:          "main",
		AuthType:            "none",
		AutoSync:            false,
		SyncIntervalMinutes: 0,
		SyncStatus:          "pending",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	require.NoError(t, ds.CreateGitRepo(repo))

	return repo
}

// =============================================================================
// TestGetCredentialSingular_YAMLOutput  [Issue #183 — RED]
//
// dvm get credential <name> -o yaml must output valid YAML with
// apiVersion/kind/metadata/spec structure. Currently FAILS because
// cmd/get_credential.go RunE only calls render.Plainf() and never checks
// the output format flag — plain text is unconditionally printed.
// =============================================================================

func TestGetCredentialSingular_YAMLOutput(t *testing.T) {
	handlers.RegisterAll()

	ds := createFullTestDataStore(t)
	defer ds.Close()

	_, _ = seedEcosystemCredential(t, ds)

	cmd := newSingularGetCredentialTestCmd(t, ds)
	out := captureCredentialYAML(t, cmd, "singular-yaml-cred", "ecosystem", "singular-yaml-cred-eco")

	require.NotEmpty(t, out,
		"BUG #183: singular get credential -o yaml produced no output; "+
			"cmd/get_credential.go RunE only calls render.Plainf() and never checks format")

	// Must be valid YAML
	var result map[string]interface{}
	err := yaml.Unmarshal(out, &result)
	require.NoError(t, err,
		"BUG #183: singular get credential -o yaml output is not valid YAML; got:\n%s", string(out))

	// --- KEY ASSERTIONS (BUG #183) ---
	assert.Equal(t, "devopsmaestro.io/v1", result["apiVersion"],
		"BUG #183: singular credential YAML must have apiVersion: devopsmaestro.io/v1; got:\n%s", string(out))

	assert.Equal(t, "Credential", result["kind"],
		"BUG #183: singular credential YAML must have kind: Credential; got:\n%s", string(out))

	_, hasMetadata := result["metadata"]
	assert.True(t, hasMetadata,
		"BUG #183: singular credential YAML must have a 'metadata' key; got:\n%s", string(out))

	_, hasSpec := result["spec"]
	assert.True(t, hasSpec,
		"BUG #183: singular credential YAML must have a 'spec' key; got:\n%s", string(out))
}

// =============================================================================
// TestGetGitRepoSingular_YAMLOutput  [Issue #183 — RED]
//
// dvm get gitrepo <name> -o yaml must output valid YAML with
// apiVersion/kind/metadata/spec structure. Currently FAILS because
// cmd/gitrepo.go runGetGitRepo calls gitRepoToMap() which produces a flat map
// (name, url, slug, authType, ...) instead of the standard resource format.
// =============================================================================

func TestGetGitRepoSingular_YAMLOutput(t *testing.T) {
	handlers.RegisterAll()

	ds := createFullTestDataStore(t)
	defer ds.Close()

	_ = seedGitRepo(t, ds)

	cmd := newSingularGetGitRepoTestCmd(t, ds)
	out := captureGitRepoYAML(t, cmd, "singular-yaml-repo")

	require.NotEmpty(t, out,
		"singular get gitrepo -o yaml produced no output")

	// Must be valid YAML
	var result map[string]interface{}
	err := yaml.Unmarshal(out, &result)
	require.NoError(t, err,
		"singular get gitrepo -o yaml output is not valid YAML; got:\n%s", string(out))

	// --- KEY ASSERTIONS (BUG #183) ---
	assert.Equal(t, "devopsmaestro.io/v1", result["apiVersion"],
		"BUG #183: singular gitrepo YAML must have apiVersion: devopsmaestro.io/v1; "+
			"currently gitRepoToMap() returns a flat map without apiVersion; got:\n%s", string(out))

	assert.Equal(t, "GitRepo", result["kind"],
		"BUG #183: singular gitrepo YAML must have kind: GitRepo; "+
			"currently gitRepoToMap() returns a flat map without kind; got:\n%s", string(out))

	_, hasMetadata := result["metadata"]
	assert.True(t, hasMetadata,
		"BUG #183: singular gitrepo YAML must have a 'metadata' key; "+
			"currently gitRepoToMap() returns flat keys (name, url, slug, ...); got:\n%s", string(out))

	_, hasSpec := result["spec"]
	assert.True(t, hasSpec,
		"BUG #183: singular gitrepo YAML must have a 'spec' key; "+
			"currently gitRepoToMap() returns flat keys instead of spec; got:\n%s", string(out))

	// Negative: flat-map keys must NOT appear at the top level
	_, hasTopLevelSlug := result["slug"]
	assert.False(t, hasTopLevelSlug,
		"BUG #183: singular gitrepo YAML must NOT have a top-level 'slug' key; "+
			"that indicates gitRepoToMap() flat output is still being used; got:\n%s", string(out))
}

// =============================================================================
// TestGetCredentialSingular_YAMLApplyCompatible  [Issue #183 — RED]
//
// The YAML produced by singular get credential must be parseable as a
// CredentialYAML (the same struct used by dvm apply). This round-trip
// confirms the output is apply-compatible.
// =============================================================================

func TestGetCredentialSingular_YAMLApplyCompatible(t *testing.T) {
	handlers.RegisterAll()

	ds := createFullTestDataStore(t)
	defer ds.Close()

	eco, _ := seedEcosystemCredential(t, ds)

	cmd := newSingularGetCredentialTestCmd(t, ds)
	out := captureCredentialYAML(t, cmd, "singular-yaml-cred", "ecosystem", "singular-yaml-cred-eco")

	require.NotEmpty(t, out,
		"BUG #183: singular get credential -o yaml produced no output")

	// Parse as CredentialYAML — this is what dvm apply does
	var credYAML models.CredentialYAML
	err := yaml.Unmarshal(out, &credYAML)
	require.NoError(t, err,
		"BUG #183: singular credential YAML must be parseable as CredentialYAML; got:\n%s", string(out))

	// Validate standard resource fields
	assert.Equal(t, "devopsmaestro.io/v1", credYAML.APIVersion,
		"BUG #183: credential YAML apiVersion must be devopsmaestro.io/v1")
	assert.Equal(t, "Credential", credYAML.Kind,
		"BUG #183: credential YAML kind must be Credential")

	// Metadata must include name
	assert.Equal(t, "singular-yaml-cred", credYAML.Metadata.Name,
		"BUG #183: credential YAML metadata.name must match the credential name")

	// Scope must be populated (ecosystem in this case) for apply to work
	assert.Equal(t, eco.Name, credYAML.Metadata.Ecosystem,
		"BUG #183: credential YAML must include metadata.ecosystem for apply to work; "+
			"plain-text output from render.Plainf() is not apply-compatible")

	// Spec must have the source
	assert.Equal(t, "env", credYAML.Spec.Source,
		"BUG #183: credential YAML spec.source must be 'env'")

	// Validate the YAML passes the model validator (same check apply uses)
	err = credentialbridge.ValidateCredentialYAML(credYAML)
	assert.NoError(t, err,
		"BUG #183: singular credential YAML must pass ValidateCredentialYAML; "+
			"plain-text output will fail validation")
}

// =============================================================================
// TestGetGitRepoSingular_YAMLApplyCompatible  [Issue #183 — RED]
//
// The YAML produced by singular get gitrepo must be parseable as a GitRepoYAML
// and must contain the correct apiVersion/kind/metadata/spec structure required
// by dvm apply. The current gitRepoToMap() output does NOT meet this bar.
// =============================================================================

func TestGetGitRepoSingular_YAMLApplyCompatible(t *testing.T) {
	handlers.RegisterAll()

	ds := createFullTestDataStore(t)
	defer ds.Close()

	repo := seedGitRepo(t, ds)

	cmd := newSingularGetGitRepoTestCmd(t, ds)
	out := captureGitRepoYAML(t, cmd, "singular-yaml-repo")

	require.NotEmpty(t, out,
		"singular get gitrepo -o yaml produced no output")

	// Parse as GitRepoYAML — this is what dvm apply does
	var gitRepoYAML models.GitRepoYAML
	err := yaml.Unmarshal(out, &gitRepoYAML)
	require.NoError(t, err,
		"BUG #183: singular gitrepo YAML must be parseable as GitRepoYAML; got:\n%s", string(out))

	// Validate standard resource fields
	assert.Equal(t, "devopsmaestro.io/v1", gitRepoYAML.APIVersion,
		"BUG #183: gitrepo YAML apiVersion must be devopsmaestro.io/v1; "+
			"gitRepoToMap() returns a flat map that YAML-decodes with empty APIVersion")

	assert.Equal(t, "GitRepo", gitRepoYAML.Kind,
		"BUG #183: gitrepo YAML kind must be GitRepo; "+
			"gitRepoToMap() returns a flat map that YAML-decodes with empty Kind")

	// Metadata must include the name
	assert.Equal(t, repo.Name, gitRepoYAML.Metadata.Name,
		"BUG #183: gitrepo YAML metadata.name must match the repo name; "+
			"gitRepoToMap() puts the name at a top-level 'name' key, not under metadata")

	// Spec must include the URL
	assert.Equal(t, repo.URL, gitRepoYAML.Spec.URL,
		"BUG #183: gitrepo YAML spec.url must match the repo URL; "+
			"gitRepoToMap() puts url at a top-level key, not under spec")
}
