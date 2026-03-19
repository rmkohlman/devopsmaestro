package models

// =============================================================================
// Issue #113 — TDD Phase 2: ToolsConfig on WorkspaceSpec
//
// RED: These tests WILL NOT COMPILE until ToolsConfig is implemented:
//
//   - models.ToolsConfig struct does not exist yet
//   - WorkspaceSpec.Tools field does not exist yet
//   - ToolsConfig.IsZero() method does not exist yet
//
// Once the production code is added, all tests in this file MUST pass.
// =============================================================================

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestToolsConfig_StructExists verifies that ToolsConfig struct exists with the
// expected Opencode bool field. This is the minimal struct definition required.
//
// RED: WILL NOT COMPILE — ToolsConfig struct does not exist yet.
func TestToolsConfig_StructExists(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// models.ToolsConfig does not exist until #113 is implemented.
	cfg := ToolsConfig{}
	assert.False(t, cfg.Opencode, "ToolsConfig.Opencode should default to false")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestToolsConfig_OpencodeField verifies that ToolsConfig has an Opencode bool
// field that can be set to true.
//
// RED: WILL NOT COMPILE — ToolsConfig struct does not exist yet.
func TestToolsConfig_OpencodeField(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cfg := ToolsConfig{
		Opencode: true,
	}
	assert.True(t, cfg.Opencode, "ToolsConfig.Opencode should be settable to true")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestToolsConfig_IsZero_ReturnsTrueWhenAllFalse verifies that IsZero() returns
// true when all fields are at their zero value (no tools enabled).
// This is required for yaml:"omitempty" to correctly omit the tools section.
//
// RED: WILL NOT COMPILE — ToolsConfig.IsZero() does not exist yet.
func TestToolsConfig_IsZero_ReturnsTrueWhenAllFalse(t *testing.T) {
	tests := []struct {
		name     string
		cfg      ToolsConfig
		wantZero bool
	}{
		{
			name:     "default zero value — all false",
			cfg:      ToolsConfig{},
			wantZero: true,
		},
		{
			name:     "opencode explicitly false",
			cfg:      ToolsConfig{Opencode: false},
			wantZero: true,
		},
		{
			name:     "opencode true — not zero",
			cfg:      ToolsConfig{Opencode: true},
			wantZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ── COMPILE ERROR EXPECTED BELOW ──────────────────────────────────
			// ToolsConfig.IsZero() does not exist until #113 is implemented.
			got := tt.cfg.IsZero()
			assert.Equal(t, tt.wantZero, got,
				"ToolsConfig.IsZero() mismatch for case %q", tt.name)
			// ──────────────────────────────────────────────────────────────────
		})
	}
}

// TestWorkspaceSpec_HasToolsField verifies that WorkspaceSpec has a Tools field
// of type ToolsConfig.
//
// RED: WILL NOT COMPILE — WorkspaceSpec.Tools field does not exist yet.
func TestWorkspaceSpec_HasToolsField(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// WorkspaceSpec.Tools does not exist until #113 is implemented.
	spec := WorkspaceSpec{
		Tools: ToolsConfig{
			Opencode: true,
		},
	}
	assert.True(t, spec.Tools.Opencode,
		"WorkspaceSpec.Tools.Opencode should be accessible and set to true")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestWorkspaceSpec_Tools_DefaultsToZero verifies that a zero-value WorkspaceSpec
// has a zero-value ToolsConfig (no tools enabled by default).
//
// RED: WILL NOT COMPILE — WorkspaceSpec.Tools field does not exist yet.
func TestWorkspaceSpec_Tools_DefaultsToZero(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	spec := WorkspaceSpec{}
	assert.False(t, spec.Tools.Opencode,
		"WorkspaceSpec.Tools.Opencode should default to false (opt-in, not opt-out)")
	assert.True(t, spec.Tools.IsZero(),
		"WorkspaceSpec.Tools should be zero-value when not configured")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestToolsConfig_YAMLUnmarshal_OpencodeTrue verifies that YAML with
// `tools:\n  opencode: true` correctly populates ToolsConfig.Opencode.
//
// RED: WILL NOT COMPILE — WorkspaceSpec.Tools field does not exist yet.
func TestToolsConfig_YAMLUnmarshal_OpencodeTrue(t *testing.T) {
	yamlContent := `
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: dev
  app: my-api
spec:
  image:
    name: python:3.11
  tools:
    opencode: true
  nvim:
    structure: lazyvim
  shell:
    type: zsh
  container:
    user: dev
    uid: 1000
    gid: 1000
    workingDir: /workspace
`

	var wsYAML WorkspaceYAML
	err := yaml.Unmarshal([]byte(yamlContent), &wsYAML)
	require.NoError(t, err, "should unmarshal workspace YAML with tools section")

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// WorkspaceSpec.Tools does not exist until #113 is implemented.
	assert.True(t, wsYAML.Spec.Tools.Opencode,
		"RED: WorkspaceSpec.Tools.Opencode should be true when 'tools: opencode: true' is set in YAML — "+
			"FAILS until WorkspaceSpec.Tools field (type ToolsConfig) is added")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestToolsConfig_YAMLUnmarshal_OpencodeFalse verifies that YAML with
// `tools:\n  opencode: false` (or no tools section) leaves Opencode as false.
//
// RED: WILL NOT COMPILE — WorkspaceSpec.Tools field does not exist yet.
func TestToolsConfig_YAMLUnmarshal_OpencodeFalse(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
	}{
		{
			name: "explicit opencode false",
			yamlContent: `
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: dev
  app: my-api
spec:
  image:
    name: python:3.11
  tools:
    opencode: false
`,
		},
		{
			name: "no tools section at all",
			yamlContent: `
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: dev
  app: my-api
spec:
  image:
    name: python:3.11
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wsYAML WorkspaceYAML
			err := yaml.Unmarshal([]byte(tt.yamlContent), &wsYAML)
			require.NoError(t, err, "should unmarshal workspace YAML")

			// ── COMPILE ERROR EXPECTED BELOW ──────────────────────────────────
			// WorkspaceSpec.Tools does not exist until #113 is implemented.
			assert.False(t, wsYAML.Spec.Tools.Opencode,
				"RED: WorkspaceSpec.Tools.Opencode should be false when not set — "+
					"FAILS until WorkspaceSpec.Tools field is added")
			// ──────────────────────────────────────────────────────────────────
		})
	}
}

// TestToolsConfig_YAMLOmitempty_ToolsSectionOmittedWhenZero verifies that when
// no tools are enabled, the marshaled YAML does NOT include a "tools:" section.
// This requires ToolsConfig.IsZero() to return true for the zero value.
//
// RED: WILL NOT COMPILE — WorkspaceSpec.Tools field does not exist yet.
func TestToolsConfig_YAMLOmitempty_ToolsSectionOmittedWhenZero(t *testing.T) {
	wsYAML := WorkspaceYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Workspace",
		Metadata: WorkspaceMetadata{
			Name: "no-tools",
			App:  "testapp",
		},
		Spec: WorkspaceSpec{
			Image: ImageConfig{Name: "python:3.11"},
			// Tools not set — zero value
		},
	}

	data, err := yaml.Marshal(wsYAML)
	require.NoError(t, err, "should marshal WorkspaceYAML without error")

	yamlStr := string(data)
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// WorkspaceSpec.Tools does not exist until #113 is implemented.
	// Even after it exists, this test verifies omitempty works via IsZero().
	assert.False(t, strings.Contains(yamlStr, "tools:"),
		"RED: 'tools:' should NOT appear in YAML when ToolsConfig is zero — "+
			"FAILS until WorkspaceSpec.Tools with yaml:\"tools,omitempty\" is added AND "+
			"ToolsConfig.IsZero() returns true for the zero value")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestToolsConfig_YAMLRoundTrip verifies that a WorkspaceYAML with
// tools.opencode=true round-trips through marshal/unmarshal correctly.
//
// RED: WILL NOT COMPILE — WorkspaceSpec.Tools field does not exist yet.
func TestToolsConfig_YAMLRoundTrip(t *testing.T) {
	original := WorkspaceYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Workspace",
		Metadata: WorkspaceMetadata{
			Name: "tools-roundtrip",
			App:  "testapp",
		},
		Spec: WorkspaceSpec{
			Image: ImageConfig{Name: "python:3.11"},
			// ── COMPILE ERROR EXPECTED BELOW ──────────────────────────────
			// WorkspaceSpec.Tools does not exist until #113 is implemented.
			Tools: ToolsConfig{
				Opencode: true,
			},
			// ──────────────────────────────────────────────────────────────
		},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(original)
	require.NoError(t, err, "should marshal WorkspaceYAML with tools.opencode=true")

	// Verify "tools:" section is present
	yamlStr := string(data)
	assert.Contains(t, yamlStr, "tools:",
		"RED: serialized YAML should include 'tools:' section when opencode=true")
	assert.Contains(t, yamlStr, "opencode: true",
		"RED: serialized YAML should include 'opencode: true'")

	// Unmarshal back and verify round-trip
	var decoded WorkspaceYAML
	err = yaml.Unmarshal(data, &decoded)
	require.NoError(t, err, "should unmarshal WorkspaceYAML with tools section")

	assert.True(t, decoded.Spec.Tools.Opencode,
		"RED: Tools.Opencode should survive marshal/unmarshal round-trip — "+
			"FAILS until WorkspaceSpec.Tools field is added")
}
