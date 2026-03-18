package models

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestWorkspace_FromYAML_AllFields(t *testing.T) {
	yamlContent := `
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: dev
  app: fastapi-test
  annotations:
    description: "Development workspace"
spec:
  image:
    name: golang:1.21
  nvim:
    structure: lazyvim
    theme: tokyonight-night
    pluginPackage: maestro
    plugins:
      - nvim-treesitter
      - telescope.nvim
  terminal:
    prompt: starship
    package: maestro
    plugins:
      - zsh-syntax-highlighting
      - zsh-autosuggestions
  gitrepo: fastapi-test
  container:
    user: dev
    uid: 1000
    gid: 1000
    workingDir: /workspace
`

	var wsYAML WorkspaceYAML
	err := yaml.Unmarshal([]byte(yamlContent), &wsYAML)
	require.NoError(t, err, "failed to parse YAML")

	// Convert to workspace model
	workspace := &Workspace{AppID: 1}
	workspace.FromYAML(wsYAML)

	// Verify all fields were mapped
	assert.Equal(t, "dev", workspace.Name)
	assert.Equal(t, "golang:1.21", workspace.ImageName)
	assert.Equal(t, "created", workspace.Status)

	// Description
	require.True(t, workspace.Description.Valid)
	assert.Equal(t, "Development workspace", workspace.Description.String)

	// Nvim fields
	require.True(t, workspace.Theme.Valid)
	assert.Equal(t, "tokyonight-night", workspace.Theme.String)

	require.True(t, workspace.NvimStructure.Valid)
	assert.Equal(t, "lazyvim", workspace.NvimStructure.String)

	require.True(t, workspace.NvimPlugins.Valid)
	assert.Equal(t, "nvim-treesitter,telescope.nvim", workspace.NvimPlugins.String)

	// Terminal fields
	require.True(t, workspace.TerminalPrompt.Valid)
	assert.Equal(t, "starship", workspace.TerminalPrompt.String)

	require.True(t, workspace.TerminalPackage.Valid)
	assert.Equal(t, "maestro", workspace.TerminalPackage.String)

	require.True(t, workspace.TerminalPlugins.Valid)
	termPlugins := workspace.GetTerminalPlugins()
	assert.Equal(t, []string{"zsh-syntax-highlighting", "zsh-autosuggestions"}, termPlugins)

	// Note: GitRepo resolution happens in the handler, not in FromYAML
	// So GitRepoID is not set here
}

func TestWorkspace_ToYAML_AllFields(t *testing.T) {
	workspace := &Workspace{
		ID:              1,
		AppID:           2,
		Name:            "dev",
		ImageName:       "golang:1.21",
		Status:          "running",
		Description:     sql.NullString{String: "Test workspace", Valid: true},
		Theme:           sql.NullString{String: "catppuccin-mocha", Valid: true},
		NvimStructure:   sql.NullString{String: "lazyvim", Valid: true},
		NvimPlugins:     sql.NullString{String: "plugin1,plugin2", Valid: true},
		TerminalPrompt:  sql.NullString{String: "starship", Valid: true},
		TerminalPackage: sql.NullString{String: "my-package", Valid: true},
	}
	workspace.SetTerminalPlugins([]string{"zsh-autosuggestions", "zsh-syntax-highlighting"})

	// Convert to YAML
	yamlDoc := workspace.ToYAML("fastapi-test", "my-gitrepo")

	// Verify structure
	assert.Equal(t, "devopsmaestro.io/v1", yamlDoc.APIVersion)
	assert.Equal(t, "Workspace", yamlDoc.Kind)
	assert.Equal(t, "dev", yamlDoc.Metadata.Name)
	assert.Equal(t, "fastapi-test", yamlDoc.Metadata.App)
	assert.Equal(t, "Test workspace", yamlDoc.Metadata.Annotations["description"])

	// Verify nvim config
	assert.Equal(t, "lazyvim", yamlDoc.Spec.Nvim.Structure)
	assert.Equal(t, "catppuccin-mocha", yamlDoc.Spec.Nvim.Theme)
	assert.Equal(t, []string{"plugin1", "plugin2"}, yamlDoc.Spec.Nvim.Plugins)

	// Verify terminal config
	assert.Equal(t, "starship", yamlDoc.Spec.Terminal.Prompt)
	assert.Equal(t, "my-package", yamlDoc.Spec.Terminal.Package)
	assert.Equal(t, []string{"zsh-autosuggestions", "zsh-syntax-highlighting"}, yamlDoc.Spec.Terminal.Plugins)

	// Verify gitrepo
	assert.Equal(t, "my-gitrepo", yamlDoc.Spec.GitRepo)
}

func TestWorkspace_YAML_RoundTrip(t *testing.T) {
	// Start with a YAML document
	originalYAML := `
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: roundtrip-test
  app: myapp
  annotations:
    description: "Round trip test"
spec:
  image:
    name: ubuntu:22.04
  nvim:
    structure: custom
    theme: gruvbox
    plugins:
      - plugin-a
      - plugin-b
  terminal:
    prompt: oh-my-posh
    plugins:
      - terminal-plugin-1
    package: my-terminal-package
  gitrepo: my-repo
  container:
    user: dev
    uid: 1000
    gid: 1000
    workingDir: /workspace
`

	// Parse YAML
	var wsYAML WorkspaceYAML
	err := yaml.Unmarshal([]byte(originalYAML), &wsYAML)
	require.NoError(t, err)

	// Convert to model
	workspace := &Workspace{AppID: 1}
	workspace.FromYAML(wsYAML)

	// Convert back to YAML
	resultYAML := workspace.ToYAML("myapp", "my-repo")

	// Verify all fields survived the round trip
	assert.Equal(t, "roundtrip-test", resultYAML.Metadata.Name)
	assert.Equal(t, "myapp", resultYAML.Metadata.App)
	assert.Equal(t, "Round trip test", resultYAML.Metadata.Annotations["description"])
	assert.Equal(t, "ubuntu:22.04", resultYAML.Spec.Image.Name)

	// Nvim
	assert.Equal(t, "custom", resultYAML.Spec.Nvim.Structure)
	assert.Equal(t, "gruvbox", resultYAML.Spec.Nvim.Theme)
	assert.Equal(t, []string{"plugin-a", "plugin-b"}, resultYAML.Spec.Nvim.Plugins)

	// Terminal
	assert.Equal(t, "oh-my-posh", resultYAML.Spec.Terminal.Prompt)
	assert.Equal(t, "my-terminal-package", resultYAML.Spec.Terminal.Package)
	assert.Equal(t, []string{"terminal-plugin-1"}, resultYAML.Spec.Terminal.Plugins)

	// GitRepo
	assert.Equal(t, "my-repo", resultYAML.Spec.GitRepo)
}

func TestWorkspace_FromYAML_PartialFields(t *testing.T) {
	// Test that FromYAML works even with minimal fields
	yamlContent := `
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: minimal
  app: myapp
spec:
  image:
    name: alpine:latest
  container:
    user: dev
`

	var wsYAML WorkspaceYAML
	err := yaml.Unmarshal([]byte(yamlContent), &wsYAML)
	require.NoError(t, err)

	workspace := &Workspace{AppID: 1}
	workspace.FromYAML(wsYAML)

	assert.Equal(t, "minimal", workspace.Name)
	assert.Equal(t, "alpine:latest", workspace.ImageName)
	assert.False(t, workspace.Description.Valid)
	assert.False(t, workspace.Theme.Valid)
	assert.False(t, workspace.NvimStructure.Valid)
	assert.False(t, workspace.NvimPlugins.Valid)
	assert.False(t, workspace.TerminalPrompt.Valid)
	assert.False(t, workspace.TerminalPackage.Valid)
	assert.False(t, workspace.TerminalPlugins.Valid)
}

func TestWorkspace_ToYAML_MarshalToBytes(t *testing.T) {
	workspace := &Workspace{
		Name:      "yaml-bytes-test",
		ImageName: "golang:1.21",
		Theme:     sql.NullString{String: "tokyonight", Valid: true},
	}

	yamlDoc := workspace.ToYAML("testapp", "")
	bytes, err := yaml.Marshal(yamlDoc)
	require.NoError(t, err)

	// Verify it's valid YAML
	yamlStr := string(bytes)
	assert.Contains(t, yamlStr, "apiVersion: devopsmaestro.io/v1")
	assert.Contains(t, yamlStr, "kind: Workspace")
	assert.Contains(t, yamlStr, "name: yaml-bytes-test")
	assert.Contains(t, yamlStr, "app: testapp")
	assert.Contains(t, yamlStr, "theme: tokyonight")

	// When gitRepoName is empty, gitrepo should not appear in YAML
	assert.False(t, strings.Contains(yamlStr, "gitrepo:"), "gitrepo should be omitted when empty")
}

// =============================================================================
// WI-5: Workspace env YAML Tests
// RED: These tests FAIL until WorkspaceSpec has Env and GetEnv/SetEnv are added.
// =============================================================================

// TestWorkspace_GetEnv_DefaultsToNil verifies that a workspace with no env
// returns an empty (non-nil) map from GetEnv.
func TestWorkspace_GetEnv_DefaultsToEmpty(t *testing.T) {
	ws := &Workspace{Name: "no-env"}
	got := ws.GetEnv()
	// Should return empty map, not nil
	if got == nil {
		t.Error("GetEnv() returned nil; expected empty map for workspace with no env set")
	}
	if len(got) != 0 {
		t.Errorf("GetEnv() = %v; expected empty map for workspace with no env set", got)
	}
}

// TestWorkspace_SetEnv_RoundTrip verifies that SetEnv encodes env and GetEnv
// decodes it back to the same map.
func TestWorkspace_SetEnv_RoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		envMap  map[string]string
		wantLen int
	}{
		{
			name:    "single key",
			envMap:  map[string]string{"MY_VAR": "hello"},
			wantLen: 1,
		},
		{
			name:    "multiple keys",
			envMap:  map[string]string{"KEY1": "val1", "KEY2": "val2", "KEY3": "val3"},
			wantLen: 3,
		},
		{
			name:    "empty map clears env",
			envMap:  map[string]string{},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &Workspace{}
			ws.SetEnv(tt.envMap)

			got := ws.GetEnv()
			require.NotNil(t, got, "GetEnv() returned nil after SetEnv")
			assert.Equal(t, tt.wantLen, len(got), "GetEnv() length mismatch")
			for k, v := range tt.envMap {
				assert.Equal(t, v, got[k], "GetEnv()[%s] mismatch", k)
			}
		})
	}
}

// TestWorkspace_ToYAML_IncludesEnv verifies that ToYAML includes the env map
// when it is populated on the workspace.
func TestWorkspace_ToYAML_IncludesEnv(t *testing.T) {
	ws := &Workspace{
		Name:      "env-ws",
		ImageName: "ubuntu:22.04",
	}
	ws.SetEnv(map[string]string{
		"BUILD_ENV": "production",
		"LOG_LEVEL": "info",
	})

	yamlDoc := ws.ToYAML("my-app", "")
	data, err := yaml.Marshal(yamlDoc)
	require.NoError(t, err)

	yamlStr := string(data)
	assert.Contains(t, yamlStr, "env:", "YAML should contain env section")
	assert.Contains(t, yamlStr, "BUILD_ENV", "YAML env should contain BUILD_ENV key")
	assert.Contains(t, yamlStr, "production", "YAML env should contain BUILD_ENV value")
}

// TestWorkspace_FromYAML_ParsesEnv verifies that FromYAML correctly parses
// the env map from a workspace YAML document.
func TestWorkspace_FromYAML_ParsesEnv(t *testing.T) {
	yamlContent := `
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: env-parse-test
  app: my-app
spec:
  image:
    name: python:3.11
  env:
    SECRET_KEY: my-secret
    API_URL: https://api.example.com
`

	var wsYAML WorkspaceYAML
	err := yaml.Unmarshal([]byte(yamlContent), &wsYAML)
	require.NoError(t, err)

	ws := &Workspace{AppID: 1}
	ws.FromYAML(wsYAML)

	gotEnv := ws.GetEnv()
	require.NotNil(t, gotEnv, "GetEnv() returned nil after FromYAML with env spec")
	assert.Equal(t, "my-secret", gotEnv["SECRET_KEY"], "env SECRET_KEY mismatch")
	assert.Equal(t, "https://api.example.com", gotEnv["API_URL"], "env API_URL mismatch")
}

// =============================================================================
// Sprint 4 Tests: WorkspaceMetadata.Domain field  [RED Phase]
// =============================================================================

// TestWorkspaceYAML_MetadataDomain_RoundTrip verifies that the WorkspaceMetadata
// struct has a Domain field that serializes to YAML as "domain:" and deserializes
// back correctly. This is needed for round-trip fidelity in "get all -o yaml".
//
// [RED: Will fail until WorkspaceMetadata.Domain field is added to the struct.]
func TestWorkspaceYAML_MetadataDomain_RoundTrip(t *testing.T) {
	// This YAML has the domain field in metadata — represents the exported format
	yamlWithDomain := `
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: my-workspace
  app: my-app
  domain: my-domain
spec:
  image:
    name: ubuntu:22.04
`

	var wsYAML WorkspaceYAML
	err := yaml.Unmarshal([]byte(yamlWithDomain), &wsYAML)
	require.NoError(t, err, "should parse YAML with metadata.domain field")

	// Verify known fields parsed correctly regardless of Domain support
	assert.Equal(t, "my-workspace", wsYAML.Metadata.Name,
		"Name should still be populated correctly")
	assert.Equal(t, "my-app", wsYAML.Metadata.App,
		"App should still be populated correctly")

	// Re-serialize to YAML — domain field should round-trip through the raw bytes
	data, err := yaml.Marshal(wsYAML)
	require.NoError(t, err, "should serialize WorkspaceYAML with Domain to YAML")

	yamlStr := string(data)
	// [RED] WorkspaceMetadata.Domain field does not exist yet, so "domain: my-domain"
	// will NOT appear in serialized output until the field is added to the struct.
	assert.Contains(t, yamlStr, "domain: my-domain",
		"serialized YAML should include 'domain: my-domain' — FAILS until WorkspaceMetadata.Domain field is added")

	// Verify round-trip: parse raw map and check domain key is preserved
	var rawDecoded map[string]interface{}
	err = yaml.Unmarshal(data, &rawDecoded)
	require.NoError(t, err, "round-trip: should parse serialized YAML without error")

	metadata, ok := rawDecoded["metadata"].(map[string]interface{})
	require.True(t, ok, "round-trip: metadata should be a map")

	domainVal, hasDomain := metadata["domain"]
	assert.True(t, hasDomain, "round-trip: metadata.domain key should be present — FAILS until WorkspaceMetadata.Domain field is added")
	assert.Equal(t, "my-domain", domainVal,
		"round-trip: domain value should be 'my-domain'")
}

// TestWorkspaceYAML_MetadataDomain_Optional verifies that the Domain field is
// optional — existing YAML without metadata.domain should still parse correctly
// without a domain key in the output.
func TestWorkspaceYAML_MetadataDomain_Optional(t *testing.T) {
	yamlWithoutDomain := `
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: legacy-workspace
  app: my-app
spec:
  image:
    name: ubuntu:22.04
`

	var wsYAML WorkspaceYAML
	err := yaml.Unmarshal([]byte(yamlWithoutDomain), &wsYAML)
	require.NoError(t, err, "should parse YAML without metadata.domain field")

	assert.Equal(t, "legacy-workspace", wsYAML.Metadata.Name,
		"Name should still parse correctly when domain is absent")
	assert.Equal(t, "my-app", wsYAML.Metadata.App,
		"App should still parse correctly when domain is absent")

	// Re-serialize and verify domain key is absent (correct default behavior)
	data, err := yaml.Marshal(wsYAML)
	require.NoError(t, err, "should serialize WorkspaceYAML without Domain")

	// Parse raw to check metadata keys
	var raw map[string]interface{}
	err = yaml.Unmarshal(data, &raw)
	require.NoError(t, err)

	metadata, ok := raw["metadata"].(map[string]interface{})
	require.True(t, ok, "metadata should be a map")

	_, hasDomain := metadata["domain"]
	assert.False(t, hasDomain,
		"serialized YAML should NOT include 'domain' key when Domain is empty")
}

// =============================================================================
// v0.54.0: CACertConfig YAML Round-Trip Tests  [RED Phase]
// =============================================================================

// TestCACertConfig_YAMLRoundTrip verifies that CACertConfig entries in
// spec.build.caCerts survive a YAML marshal/unmarshal round-trip with all
// fields preserved.
func TestCACertConfig_YAMLRoundTrip(t *testing.T) {
	original := WorkspaceYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Workspace",
		Metadata: WorkspaceMetadata{
			Name: "cacert-roundtrip",
			App:  "testapp",
		},
		Spec: WorkspaceSpec{
			Image: ImageConfig{Name: "ubuntu:22.04"},
			Build: DevBuildConfig{
				CACerts: []CACertConfig{
					{
						Name:             "corporate-ca",
						VaultSecret:      "corp-cert",
						VaultEnvironment: "prod",
					},
					{
						Name:             "intermediate-ca",
						VaultSecret:      "inter-cert",
						VaultEnvironment: "prod",
						VaultField:       "pem",
					},
				},
			},
		},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(original)
	require.NoError(t, err, "should marshal WorkspaceYAML with caCerts")

	// Unmarshal back
	var decoded WorkspaceYAML
	err = yaml.Unmarshal(data, &decoded)
	require.NoError(t, err, "should unmarshal WorkspaceYAML with caCerts")

	// Assert both certs survived
	require.Len(t, decoded.Spec.Build.CACerts, 2, "should have 2 CA certs after round-trip")

	// First cert
	assert.Equal(t, "corporate-ca", decoded.Spec.Build.CACerts[0].Name)
	assert.Equal(t, "corp-cert", decoded.Spec.Build.CACerts[0].VaultSecret)
	assert.Equal(t, "prod", decoded.Spec.Build.CACerts[0].VaultEnvironment)
	assert.Empty(t, decoded.Spec.Build.CACerts[0].VaultField, "VaultField should be empty for first cert")

	// Second cert
	assert.Equal(t, "intermediate-ca", decoded.Spec.Build.CACerts[1].Name)
	assert.Equal(t, "inter-cert", decoded.Spec.Build.CACerts[1].VaultSecret)
	assert.Equal(t, "prod", decoded.Spec.Build.CACerts[1].VaultEnvironment)
	assert.Equal(t, "pem", decoded.Spec.Build.CACerts[1].VaultField)
}

// TestCACertConfig_YAMLRoundTrip_Empty verifies that a WorkspaceYAML with no
// caCerts produces nil or empty slice after round-trip (omitempty behavior).
func TestCACertConfig_YAMLRoundTrip_Empty(t *testing.T) {
	original := WorkspaceYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Workspace",
		Metadata: WorkspaceMetadata{
			Name: "no-cacerts",
			App:  "testapp",
		},
		Spec: WorkspaceSpec{
			Image: ImageConfig{Name: "alpine:latest"},
		},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(original)
	require.NoError(t, err, "should marshal WorkspaceYAML without caCerts")

	// Unmarshal back
	var decoded WorkspaceYAML
	err = yaml.Unmarshal(data, &decoded)
	require.NoError(t, err, "should unmarshal WorkspaceYAML without caCerts")

	// CACerts should be nil or empty
	assert.Empty(t, decoded.Spec.Build.CACerts, "CACerts should be nil or empty when not specified")
}

// =============================================================================
// v0.55.0 Phase 2 RED Tests: WI-1 — Workspace BuildConfig Persistence
//
// These tests verify that spec.build (args, caCerts) is persisted to the
// database and round-trips correctly through FromYAML → DB → ToYAML.
//
// RED: All three tests WILL NOT COMPILE until WI-1 is implemented:
//
//	Workspace.BuildConfig sql.NullString field does not exist yet.
//	FromYAML() does not persist spec.build to BuildConfig.
//	ToYAML() does not read BuildConfig to populate Spec.Build.
//
// =============================================================================

// TestWorkspace_BuildConfig_RoundTrip_Args verifies that spec.build.args set in
// YAML is persisted to Workspace.BuildConfig via FromYAML and then restored by
// ToYAML, so build args survive a full model round-trip.
//
// RED: WILL NOT COMPILE — Workspace.BuildConfig field does not exist yet (WI-1).
func TestWorkspace_BuildConfig_RoundTrip_Args(t *testing.T) {
	wsYAML := WorkspaceYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Workspace",
		Metadata: WorkspaceMetadata{
			Name: "build-args-ws",
			App:  "ml-api",
		},
		Spec: WorkspaceSpec{
			Image: ImageConfig{Name: "python:3.11"},
			Build: DevBuildConfig{
				Args: map[string]string{
					"PIP_INDEX_URL": "https://pypi.example.com",
					"DEBUG_BUILD":   "true",
				},
			},
		},
	}

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// Workspace.BuildConfig does not exist until WI-1 is implemented.
	ws := &Workspace{AppID: 1}
	ws.FromYAML(wsYAML)

	// Verify that FromYAML persisted the build config into the BuildConfig column.
	require.True(t, ws.BuildConfig.Valid,
		"RED: Workspace.BuildConfig.Valid should be true after FromYAML with spec.build.args — FAILS until WI-1 (Workspace.BuildConfig field) is added")

	// Verify round-trip: ToYAML should restore the args from BuildConfig.
	result := ws.ToYAML("ml-api", "")
	require.NotNil(t, result.Spec.Build.Args,
		"RED: ToYAML should populate Spec.Build.Args from BuildConfig — FAILS until WI-1")
	assert.Equal(t, "https://pypi.example.com", result.Spec.Build.Args["PIP_INDEX_URL"],
		"PIP_INDEX_URL should survive FromYAML → ToYAML round-trip")
	assert.Equal(t, "true", result.Spec.Build.Args["DEBUG_BUILD"],
		"DEBUG_BUILD should survive FromYAML → ToYAML round-trip")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestWorkspace_BuildConfig_RoundTrip_CACerts verifies that spec.build.caCerts
// set in YAML is persisted to Workspace.BuildConfig via FromYAML and then
// restored by ToYAML.
//
// RED: WILL NOT COMPILE — Workspace.BuildConfig field does not exist yet (WI-1).
func TestWorkspace_BuildConfig_RoundTrip_CACerts(t *testing.T) {
	wsYAML := WorkspaceYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Workspace",
		Metadata: WorkspaceMetadata{
			Name: "cacerts-persist-ws",
			App:  "ml-api",
		},
		Spec: WorkspaceSpec{
			Image: ImageConfig{Name: "ubuntu:22.04"},
			Build: DevBuildConfig{
				CACerts: []CACertConfig{
					{
						Name:             "corporate-ca",
						VaultSecret:      "corp-cert",
						VaultEnvironment: "prod",
					},
				},
			},
		},
	}

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// Workspace.BuildConfig does not exist until WI-1 is implemented.
	ws := &Workspace{AppID: 1}
	ws.FromYAML(wsYAML)

	// Verify the build config was persisted with a valid JSON blob.
	require.True(t, ws.BuildConfig.Valid,
		"RED: Workspace.BuildConfig.Valid should be true after FromYAML with spec.build.caCerts — FAILS until WI-1")

	// Verify round-trip: ToYAML should restore the CA certs from BuildConfig.
	result := ws.ToYAML("ml-api", "")
	require.Len(t, result.Spec.Build.CACerts, 1,
		"RED: ToYAML should restore 1 CA cert from BuildConfig — FAILS until WI-1")
	assert.Equal(t, "corporate-ca", result.Spec.Build.CACerts[0].Name,
		"CA cert name should survive FromYAML → ToYAML round-trip")
	assert.Equal(t, "corp-cert", result.Spec.Build.CACerts[0].VaultSecret,
		"CA cert vaultSecret should survive FromYAML → ToYAML round-trip")
	assert.Equal(t, "prod", result.Spec.Build.CACerts[0].VaultEnvironment,
		"CA cert vaultEnvironment should survive FromYAML → ToYAML round-trip")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestWorkspace_BuildConfig_Empty_OmittedFromYAML verifies that when no
// spec.build is set in YAML, Workspace.BuildConfig is not valid (zero-value)
// and ToYAML does not emit a "build:" section.
//
// RED: WILL NOT COMPILE — Workspace.BuildConfig field does not exist yet (WI-1).
func TestWorkspace_BuildConfig_Empty_OmittedFromYAML(t *testing.T) {
	wsYAML := WorkspaceYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Workspace",
		Metadata: WorkspaceMetadata{
			Name: "no-build-config-ws",
			App:  "simple-app",
		},
		Spec: WorkspaceSpec{
			Image: ImageConfig{Name: "alpine:latest"},
			// No Build config
		},
	}

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// Workspace.BuildConfig does not exist until WI-1 is implemented.
	ws := &Workspace{AppID: 1}
	ws.FromYAML(wsYAML)

	// When no build config is present, BuildConfig should remain invalid (NULL in DB).
	assert.False(t, ws.BuildConfig.Valid,
		"RED: Workspace.BuildConfig.Valid should be false when no spec.build is present — FAILS until WI-1")

	// ToYAML should omit the build section entirely (omitempty).
	result := ws.ToYAML("simple-app", "")
	data, err := yaml.Marshal(result)
	require.NoError(t, err, "should marshal workspace YAML without error")

	yamlStr := string(data)
	assert.NotContains(t, yamlStr, "build:",
		"RED: 'build:' should NOT appear in YAML when BuildConfig is empty — FAILS until WI-1 (ToYAML must read BuildConfig and DevBuildConfig needs omitempty)")
	// ─────────────────────────────────────────────────────────────────────────
}
