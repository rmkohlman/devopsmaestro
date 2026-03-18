package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// =============================================================================
// v0.55.0 Phase 2 RED Tests: WI-3 — Add Build Args to Ecosystem
//
// These tests verify that:
//   1. EcosystemSpec gains a Build BuildArgsConfig field (YAML: spec.build.args)
//   2. Ecosystem model gains a BuildArgs sql.NullString field for DB persistence
//   3. FromYAML() persists build args into Ecosystem.BuildArgs
//   4. ToYAML() restores build args from Ecosystem.BuildArgs into Spec.Build.Args
//   5. Empty build args are omitted from YAML output (omitempty)
//
// RED: ALL tests in this file WILL NOT COMPILE until WI-3 is implemented:
//
//   - Ecosystem.BuildArgs sql.NullString field does not exist on Ecosystem struct
//   - BuildArgsConfig struct does not exist
//   - EcosystemSpec.Build field of type BuildArgsConfig does not exist
//   - FromYAML() does not read or store spec.build.args
//   - ToYAML() does not emit spec.build.args from BuildArgs
//
// =============================================================================

// TestEcosystem_BuildArgs_RoundTrip verifies that spec.build.args set in
// EcosystemYAML is persisted to Ecosystem.BuildArgs via FromYAML and then
// restored by ToYAML into Spec.Build.Args.
//
// RED: WILL NOT COMPILE — Ecosystem.BuildArgs, BuildArgsConfig, and
// EcosystemSpec.Build do not exist yet (WI-3).
func TestEcosystem_BuildArgs_RoundTrip(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	// EcosystemSpec.Build of type BuildArgsConfig does not exist yet.
	// BuildArgsConfig struct does not exist yet.
	ecoYAML := EcosystemYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Ecosystem",
		Metadata: EcosystemMetadata{
			Name: "corporate-platform",
		},
		Spec: EcosystemSpec{
			Description: "Corporate development platform",
			Theme:       "coolnight-ocean",
			Build: BuildArgsConfig{
				Args: map[string]string{
					"PIP_INDEX_URL":    "https://pypi.example.com",
					"PIP_TRUSTED_HOST": "pypi.example.com",
				},
			},
		},
	}

	eco := &Ecosystem{}
	eco.FromYAML(ecoYAML)

	// Verify that FromYAML persisted the build args.
	require.True(t, eco.BuildArgs.Valid,
		"RED: Ecosystem.BuildArgs.Valid should be true after FromYAML with spec.build.args — FAILS until WI-3")

	// Verify round-trip: ToYAML should restore the args from BuildArgs.
	result := eco.ToYAML(nil)
	require.NotNil(t, result.Spec.Build.Args,
		"RED: ToYAML should populate Spec.Build.Args from BuildArgs — FAILS until WI-3")
	assert.Equal(t, "https://pypi.example.com", result.Spec.Build.Args["PIP_INDEX_URL"],
		"PIP_INDEX_URL should survive FromYAML → ToYAML round-trip")
	assert.Equal(t, "pypi.example.com", result.Spec.Build.Args["PIP_TRUSTED_HOST"],
		"PIP_TRUSTED_HOST should survive FromYAML → ToYAML round-trip")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestEcosystem_BuildArgs_MultipleKeys verifies that multiple build args at
// ecosystem level all round-trip correctly.
//
// RED: WILL NOT COMPILE — same missing fields as TestEcosystem_BuildArgs_RoundTrip.
func TestEcosystem_BuildArgs_MultipleKeys(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	ecoYAML := EcosystemYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Ecosystem",
		Metadata:   EcosystemMetadata{Name: "test-eco"},
		Spec: EcosystemSpec{
			Build: BuildArgsConfig{
				Args: map[string]string{
					"KEY_ONE":   "value1",
					"KEY_TWO":   "value2",
					"KEY_THREE": "value3",
				},
			},
		},
	}

	eco := &Ecosystem{}
	eco.FromYAML(ecoYAML)

	require.True(t, eco.BuildArgs.Valid,
		"RED: Ecosystem.BuildArgs.Valid should be true — FAILS until WI-3")

	result := eco.ToYAML(nil)
	require.Len(t, result.Spec.Build.Args, 3,
		"RED: All 3 build args should survive round-trip — FAILS until WI-3")
	assert.Equal(t, "value1", result.Spec.Build.Args["KEY_ONE"])
	assert.Equal(t, "value2", result.Spec.Build.Args["KEY_TWO"])
	assert.Equal(t, "value3", result.Spec.Build.Args["KEY_THREE"])
	// ─────────────────────────────────────────────────────────────────────────
}

// TestEcosystem_BuildArgs_Empty_OmittedFromYAML verifies that when no build
// args are set on an Ecosystem, the "build:" section is entirely absent from
// the serialized YAML output (omitempty behavior).
//
// RED: WILL NOT COMPILE — Ecosystem.BuildArgs and EcosystemSpec.Build
// do not exist yet (WI-3). The omitempty behavior on BuildArgsConfig also
// requires the struct to have the yaml:"build,omitempty" tag.
func TestEcosystem_BuildArgs_Empty_OmittedFromYAML(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// Ecosystem.BuildArgs does not exist yet.
	eco := &Ecosystem{
		Name: "no-build-args-eco",
	}
	// BuildArgs is not set — should remain zero-value (not valid)
	assert.False(t, eco.BuildArgs.Valid,
		"RED: Ecosystem.BuildArgs.Valid should be false when not set — FAILS until WI-3 (field doesn't exist)")

	result := eco.ToYAML(nil)
	data, err := yaml.Marshal(result)
	require.NoError(t, err, "should marshal ecosystem YAML without error")

	yamlStr := string(data)
	assert.False(t, strings.Contains(yamlStr, "build:"),
		"RED: 'build:' should NOT appear in YAML when BuildArgs is empty — FAILS until WI-3 (EcosystemSpec.Build needs omitempty)")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestEcosystem_BuildArgs_FromYAMLString verifies that parsing an EcosystemYAML
// document string with spec.build.args correctly populates the Ecosystem.
//
// RED: WILL NOT COMPILE — EcosystemSpec.Build and BuildArgsConfig do not
// exist yet (WI-3). YAML unmarshal into EcosystemSpec will silently discard
// the build field until the struct field is added.
func TestEcosystem_BuildArgs_FromYAMLString(t *testing.T) {
	yamlContent := `
apiVersion: devopsmaestro.io/v1
kind: Ecosystem
metadata:
  name: corporate-platform
spec:
  description: Corporate development platform
  theme: coolnight-ocean
  build:
    args:
      PIP_INDEX_URL: "https://pypi.beans.sjrb.ad/root/prod"
      PIP_TRUSTED_HOST: "pypi.beans.sjrb.ad"
`

	var ecoYAML EcosystemYAML
	err := yaml.Unmarshal([]byte(yamlContent), &ecoYAML)
	require.NoError(t, err, "should parse ecosystem YAML with build.args")

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// EcosystemSpec.Build does not exist yet.
	require.NotNil(t, ecoYAML.Spec.Build.Args,
		"RED: EcosystemSpec.Build.Args should be populated after YAML parse — FAILS until WI-3")
	assert.Equal(t, "https://pypi.beans.sjrb.ad/root/prod", ecoYAML.Spec.Build.Args["PIP_INDEX_URL"],
		"PIP_INDEX_URL should be parsed from YAML build.args")
	assert.Equal(t, "pypi.beans.sjrb.ad", ecoYAML.Spec.Build.Args["PIP_TRUSTED_HOST"],
		"PIP_TRUSTED_HOST should be parsed from YAML build.args")

	// Existing fields should still work
	assert.Equal(t, "corporate-platform", ecoYAML.Metadata.Name)
	assert.Equal(t, "Corporate development platform", ecoYAML.Spec.Description)
	assert.Equal(t, "coolnight-ocean", ecoYAML.Spec.Theme)
	// ─────────────────────────────────────────────────────────────────────────
}

// TestEcosystem_BuildArgs_PreservesExistingFields verifies that adding build
// args to an ecosystem YAML does not break existing field parsing (description,
// theme, domains).
//
// RED: WILL NOT COMPILE — EcosystemSpec.Build does not exist yet (WI-3).
// The test also confirms existing fields still round-trip correctly after
// the WI-3 changes are applied.
func TestEcosystem_BuildArgs_PreservesExistingFields(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	ecoYAML := EcosystemYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Ecosystem",
		Metadata: EcosystemMetadata{
			Name: "full-ecosystem",
			Annotations: map[string]string{
				"description": "Full featured ecosystem",
			},
		},
		Spec: EcosystemSpec{
			Description: "Full featured ecosystem",
			Theme:       "tokyonight-night",
			Domains:     []string{"backend", "frontend"},
			Build: BuildArgsConfig{
				Args: map[string]string{
					"NPM_REGISTRY": "https://registry.example.com",
				},
			},
		},
	}

	eco := &Ecosystem{}
	eco.FromYAML(ecoYAML)

	// Existing fields should still be populated correctly
	assert.Equal(t, "full-ecosystem", eco.Name, "Name should be set correctly")
	assert.True(t, eco.Theme.Valid, "Theme should be valid")
	assert.Equal(t, "tokyonight-night", eco.Theme.String, "Theme value should be correct")
	assert.True(t, eco.Description.Valid, "Description should be valid")
	assert.Equal(t, "Full featured ecosystem", eco.Description.String)

	// Build args should also be persisted
	require.True(t, eco.BuildArgs.Valid,
		"RED: Ecosystem.BuildArgs.Valid should be true — FAILS until WI-3")

	// Full round-trip
	result := eco.ToYAML([]string{"backend", "frontend"})
	assert.Equal(t, "full-ecosystem", result.Metadata.Name)
	assert.Equal(t, "tokyonight-night", result.Spec.Theme)
	assert.Equal(t, "https://registry.example.com", result.Spec.Build.Args["NPM_REGISTRY"],
		"Build arg should survive round-trip alongside existing fields")
	// ─────────────────────────────────────────────────────────────────────────
}
