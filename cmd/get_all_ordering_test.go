package cmd

// ---------------------------------------------------------------------------
// Ordering tests for get_all.go — Issue #184
//
// These are RED-phase TDD tests. They verify that the YAML export emits
// resources in dependency order so that ApplyList() can restore them without
// cross-reference failures. All tests in this file are expected to FAIL until
// the emit order in get_all.go is fixed.
//
// Correct dependency order:
//   GlobalDefaults → Ecosystem → Domain → Registry → Credential →
//   GitRepo → App → Workspace → NvimPlugin → NvimTheme → NvimPackage →
//   TerminalPrompt → TerminalPackage → TerminalPlugin → CRD
// ---------------------------------------------------------------------------

import (
	"bytes"
	"database/sql"
	"strings"
	"testing"

	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/render"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// kindIndexInYAML returns the byte-offset of the first occurrence of
// "kind: <k>" in the YAML output, or -1 if not found.
func kindIndexInYAML(output, k string) int {
	return strings.Index(output, "kind: "+k)
}

// ---------------------------------------------------------------------------
// TestGetAllOrdering_CredentialBeforeGitRepo
//
// Verifies that Credential items appear before GitRepo items in the YAML
// export. GitRepos may reference credentials via credentialID, so credentials
// must exist before git repos are applied during restore.
//
// FAILS until get_all.go emits Credential before GitRepo.
// ---------------------------------------------------------------------------

func TestGetAllOrdering_CredentialBeforeGitRepo(t *testing.T) {
	ds := createFullTestDataStore(t)
	defer ds.Close()

	// Seed one ecosystem so scope resolves to ShowAll (no active context)
	eco := &models.Ecosystem{Name: "ord-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))

	// Seed a credential scoped to the ecosystem
	envVar := "ORD_TOKEN"
	cred := &models.CredentialDB{
		Name:      "ord-cred",
		ScopeType: models.CredentialScopeEcosystem,
		ScopeID:   int64(eco.ID),
		Source:    "env",
		EnvVar:    &envVar,
	}
	require.NoError(t, ds.CreateCredential(cred))

	// Seed a git repo that references the credential
	repo := &models.GitRepoDB{
		Name:         "ord-repo",
		URL:          "https://github.com/org/ord.git",
		Slug:         "github.com_org_ord",
		DefaultRef:   "main",
		AuthType:     "ssh",
		SyncStatus:   "pending",
		CredentialID: sql.NullInt64{Int64: cred.ID, Valid: true},
	}
	require.NoError(t, ds.CreateGitRepo(repo))

	cmd := newGetAllTestCmd(t, ds)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "yaml"

	require.NoError(t, getAll(cmd), "getAll YAML should not error")

	output := buf.String()

	credIdx := kindIndexInYAML(output, "Credential")
	gitRepoIdx := kindIndexInYAML(output, "GitRepo")

	require.NotEqual(t, -1, credIdx, "YAML output must contain a Credential item")
	require.NotEqual(t, -1, gitRepoIdx, "YAML output must contain a GitRepo item")

	assert.Less(t, credIdx, gitRepoIdx,
		"Credential (idx=%d) must appear BEFORE GitRepo (idx=%d) in YAML export — FAILS until #184 is fixed",
		credIdx, gitRepoIdx)
}

// ---------------------------------------------------------------------------
// TestGetAllOrdering_DependencyChain
//
// Verifies the broader dependency chain:
//   Ecosystem → Domain → Registry → Credential → GitRepo → App → Workspace
//
// Each kind must appear strictly before the next in the YAML output.
// FAILS until get_all.go is reordered to match the dependency chain.
// ---------------------------------------------------------------------------

func TestGetAllOrdering_DependencyChain(t *testing.T) {
	ds := createFullTestDataStore(t)
	defer ds.Close()

	// Seed the full hierarchy
	eco := &models.Ecosystem{Name: "chain-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))

	dom := &models.Domain{Name: "chain-dom", EcosystemID: eco.ID}
	require.NoError(t, ds.CreateDomain(dom))

	reg := &models.Registry{
		Name:    "chain-reg",
		Type:    "zot",
		Port:    5200,
		Storage: "/tmp/chain-reg",
		Version: "2.1.0",
		Status:  "stopped",
	}
	require.NoError(t, ds.CreateRegistry(reg))

	envVar := "CHAIN_TOKEN"
	cred := &models.CredentialDB{
		Name:      "chain-cred",
		ScopeType: models.CredentialScopeEcosystem,
		ScopeID:   int64(eco.ID),
		Source:    "env",
		EnvVar:    &envVar,
	}
	require.NoError(t, ds.CreateCredential(cred))

	repo := &models.GitRepoDB{
		Name:       "chain-repo",
		URL:        "https://github.com/org/chain.git",
		Slug:       "github.com_org_chain",
		DefaultRef: "main",
		AuthType:   "none",
		SyncStatus: "pending",
	}
	require.NoError(t, ds.CreateGitRepo(repo))

	app := &models.App{
		Name:     "chain-app",
		Path:     "/chain",
		DomainID: dom.ID,
	}
	require.NoError(t, ds.CreateApp(app))

	ws := &models.Workspace{
		Name:      "chain-ws",
		AppID:     app.ID,
		ImageName: "ubuntu:22.04",
		Status:    "stopped",
	}
	require.NoError(t, ds.CreateWorkspace(ws))

	cmd := newGetAllTestCmd(t, ds)

	var buf bytes.Buffer
	origWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(origWriter)

	origFormat := getOutputFormat
	defer func() { getOutputFormat = origFormat }()
	getOutputFormat = "yaml"

	require.NoError(t, getAll(cmd), "getAll YAML should not error")

	output := buf.String()

	// Collect positions for each kind in the YAML
	positions := map[string]int{
		"Ecosystem":  kindIndexInYAML(output, "Ecosystem"),
		"Domain":     kindIndexInYAML(output, "Domain"),
		"Registry":   kindIndexInYAML(output, "Registry"),
		"Credential": kindIndexInYAML(output, "Credential"),
		"GitRepo":    kindIndexInYAML(output, "GitRepo"),
		"App":        kindIndexInYAML(output, "App"),
		"Workspace":  kindIndexInYAML(output, "Workspace"),
	}

	for kind, idx := range positions {
		require.NotEqual(t, -1, idx, "YAML output must contain kind %q", kind)
	}

	// Assert the dependency chain order
	chain := []string{"Ecosystem", "Domain", "Registry", "Credential", "GitRepo", "App", "Workspace"}
	for i := 0; i < len(chain)-1; i++ {
		a, b := chain[i], chain[i+1]
		assert.Less(t, positions[a], positions[b],
			"%s (idx=%d) must appear BEFORE %s (idx=%d) in YAML — FAILS until #184 is fixed",
			a, positions[a], b, positions[b])
	}
}
