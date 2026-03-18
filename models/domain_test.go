package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// =============================================================================
// v0.55.0 Phase 2 RED Tests: WI-3 — Add Build Args to Domain
//
// These tests verify that:
//   1. DomainSpec gains a Build BuildArgsConfig field (YAML: spec.build.args)
//   2. Domain model gains a BuildArgs sql.NullString field for DB persistence
//   3. FromYAML() persists build args into Domain.BuildArgs
//   4. ToYAML() restores build args from Domain.BuildArgs into Spec.Build.Args
//   5. Empty build args are omitted from YAML output (omitempty)
//
// RED: ALL tests in this file WILL NOT COMPILE until WI-3 is implemented:
//
//   - Domain.BuildArgs sql.NullString field does not exist on Domain struct
//   - BuildArgsConfig struct does not exist (shared with Ecosystem)
//   - DomainSpec.Build field of type BuildArgsConfig does not exist
//   - FromYAML() does not read or store spec.build.args
//   - ToYAML() does not emit spec.build.args from BuildArgs
//
// =============================================================================

// TestDomain_BuildArgs_RoundTrip verifies that spec.build.args set in
// DomainYAML is persisted to Domain.BuildArgs via FromYAML and then
// restored by ToYAML into Spec.Build.Args.
//
// RED: WILL NOT COMPILE — Domain.BuildArgs, BuildArgsConfig, and
// DomainSpec.Build do not exist yet (WI-3).
func TestDomain_BuildArgs_RoundTrip(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	// DomainSpec.Build of type BuildArgsConfig does not exist yet.
	// BuildArgsConfig struct does not exist yet.
	domYAML := DomainYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Domain",
		Metadata: DomainMetadata{
			Name:      "data-science",
			Ecosystem: "corporate-platform",
		},
		Spec: DomainSpec{
			Build: BuildArgsConfig{
				Args: map[string]string{
					"EXTRA_PIP_PACKAGES": "numpy pandas",
					"CUDA_VERSION":       "12.0",
				},
			},
		},
	}

	domain := &Domain{}
	domain.FromYAML(domYAML)

	// Verify that FromYAML persisted the build args.
	require.True(t, domain.BuildArgs.Valid,
		"RED: Domain.BuildArgs.Valid should be true after FromYAML with spec.build.args — FAILS until WI-3")

	// Verify round-trip: ToYAML should restore the args from BuildArgs.
	result := domain.ToYAML("corporate-platform", nil)
	require.NotNil(t, result.Spec.Build.Args,
		"RED: ToYAML should populate Spec.Build.Args from BuildArgs — FAILS until WI-3")
	assert.Equal(t, "numpy pandas", result.Spec.Build.Args["EXTRA_PIP_PACKAGES"],
		"EXTRA_PIP_PACKAGES should survive FromYAML → ToYAML round-trip")
	assert.Equal(t, "12.0", result.Spec.Build.Args["CUDA_VERSION"],
		"CUDA_VERSION should survive FromYAML → ToYAML round-trip")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestDomain_BuildArgs_MultipleKeys verifies that multiple build args at
// domain level all round-trip correctly.
//
// RED: WILL NOT COMPILE — same missing fields as TestDomain_BuildArgs_RoundTrip.
func TestDomain_BuildArgs_MultipleKeys(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	domYAML := DomainYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Domain",
		Metadata: DomainMetadata{
			Name:      "backend",
			Ecosystem: "my-eco",
		},
		Spec: DomainSpec{
			Build: BuildArgsConfig{
				Args: map[string]string{
					"DOMAIN_ARG_1": "domain-val-1",
					"DOMAIN_ARG_2": "domain-val-2",
					"DOMAIN_ARG_3": "domain-val-3",
				},
			},
		},
	}

	domain := &Domain{}
	domain.FromYAML(domYAML)

	require.True(t, domain.BuildArgs.Valid,
		"RED: Domain.BuildArgs.Valid should be true — FAILS until WI-3")

	result := domain.ToYAML("my-eco", nil)
	require.Len(t, result.Spec.Build.Args, 3,
		"RED: All 3 build args should survive round-trip — FAILS until WI-3")
	assert.Equal(t, "domain-val-1", result.Spec.Build.Args["DOMAIN_ARG_1"])
	assert.Equal(t, "domain-val-2", result.Spec.Build.Args["DOMAIN_ARG_2"])
	assert.Equal(t, "domain-val-3", result.Spec.Build.Args["DOMAIN_ARG_3"])
	// ─────────────────────────────────────────────────────────────────────────
}

// TestDomain_BuildArgs_Empty_OmittedFromYAML verifies that when no build args
// are set on a Domain, the "build:" section is entirely absent from the
// serialized YAML output (omitempty behavior).
//
// RED: WILL NOT COMPILE — Domain.BuildArgs and DomainSpec.Build do not
// exist yet (WI-3).
func TestDomain_BuildArgs_Empty_OmittedFromYAML(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// Domain.BuildArgs does not exist yet.
	domain := &Domain{
		Name: "no-build-args-domain",
	}
	// BuildArgs is not set — should remain zero-value (not valid)
	assert.False(t, domain.BuildArgs.Valid,
		"RED: Domain.BuildArgs.Valid should be false when not set — FAILS until WI-3 (field doesn't exist)")

	result := domain.ToYAML("my-eco", nil)
	data, err := yaml.Marshal(result)
	require.NoError(t, err, "should marshal domain YAML without error")

	yamlStr := string(data)
	assert.False(t, strings.Contains(yamlStr, "build:"),
		"RED: 'build:' should NOT appear in YAML when BuildArgs is empty — FAILS until WI-3 (DomainSpec.Build needs omitempty)")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestDomain_BuildArgs_FromYAMLString verifies that parsing a DomainYAML
// document string with spec.build.args correctly populates the Domain model.
//
// RED: WILL NOT COMPILE — DomainSpec.Build and BuildArgsConfig do not exist
// yet (WI-3). YAML unmarshal into DomainSpec will silently discard the build
// field until the struct field is added.
func TestDomain_BuildArgs_FromYAMLString(t *testing.T) {
	yamlContent := `
apiVersion: devopsmaestro.io/v1
kind: Domain
metadata:
  name: data-science
  ecosystem: corporate-platform
spec:
  build:
    args:
      EXTRA_PIP_PACKAGES: "numpy pandas"
      CUDA_VERSION: "12.0"
`

	var domYAML DomainYAML
	err := yaml.Unmarshal([]byte(yamlContent), &domYAML)
	require.NoError(t, err, "should parse domain YAML with build.args")

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// DomainSpec.Build does not exist yet.
	require.NotNil(t, domYAML.Spec.Build.Args,
		"RED: DomainSpec.Build.Args should be populated after YAML parse — FAILS until WI-3")
	assert.Equal(t, "numpy pandas", domYAML.Spec.Build.Args["EXTRA_PIP_PACKAGES"],
		"EXTRA_PIP_PACKAGES should be parsed from YAML build.args")
	assert.Equal(t, "12.0", domYAML.Spec.Build.Args["CUDA_VERSION"],
		"CUDA_VERSION should be parsed from YAML build.args")

	// Existing fields should still work
	assert.Equal(t, "data-science", domYAML.Metadata.Name)
	assert.Equal(t, "corporate-platform", domYAML.Metadata.Ecosystem)
	// ─────────────────────────────────────────────────────────────────────────
}

// TestDomain_BuildArgs_PreservesExistingFields verifies that adding build args
// to a domain YAML does not break existing field parsing (theme, apps).
//
// RED: WILL NOT COMPILE — DomainSpec.Build does not exist yet (WI-3).
func TestDomain_BuildArgs_PreservesExistingFields(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	domYAML := DomainYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Domain",
		Metadata: DomainMetadata{
			Name:      "full-domain",
			Ecosystem: "corporate-platform",
			Annotations: map[string]string{
				"description": "Full featured domain",
			},
		},
		Spec: DomainSpec{
			Theme: "catppuccin-mocha",
			Apps:  []string{"ml-api", "data-pipeline"},
			Build: BuildArgsConfig{
				Args: map[string]string{
					"DATA_BACKEND": "spark",
				},
			},
		},
	}

	domain := &Domain{}
	domain.FromYAML(domYAML)

	// Existing fields should still be populated correctly
	assert.Equal(t, "full-domain", domain.Name, "Name should be set correctly")
	assert.True(t, domain.Theme.Valid, "Theme should be valid")
	assert.Equal(t, "catppuccin-mocha", domain.Theme.String, "Theme value should be correct")
	assert.True(t, domain.Description.Valid, "Description should be valid")
	assert.Equal(t, "Full featured domain", domain.Description.String)

	// Build args should also be persisted
	require.True(t, domain.BuildArgs.Valid,
		"RED: Domain.BuildArgs.Valid should be true — FAILS until WI-3")

	// Full round-trip
	result := domain.ToYAML("corporate-platform", []string{"ml-api", "data-pipeline"})
	assert.Equal(t, "full-domain", result.Metadata.Name)
	assert.Equal(t, "catppuccin-mocha", result.Spec.Theme)
	assert.Equal(t, "spark", result.Spec.Build.Args["DATA_BACKEND"],
		"Build arg should survive round-trip alongside existing fields")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestDomain_BuildArgs_JSONEncoding verifies that Domain.BuildArgs stores build
// args as a valid JSON blob (same pattern as App.BuildConfig and Workspace.Env).
//
// RED: WILL NOT COMPILE — Domain.BuildArgs field does not exist yet (WI-3).
func TestDomain_BuildArgs_JSONEncoding(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	domYAML := DomainYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Domain",
		Metadata:   DomainMetadata{Name: "json-test-domain", Ecosystem: "my-eco"},
		Spec: DomainSpec{
			Build: BuildArgsConfig{
				Args: map[string]string{
					"MY_KEY": "my_value",
				},
			},
		},
	}

	domain := &Domain{}
	domain.FromYAML(domYAML)

	require.True(t, domain.BuildArgs.Valid,
		"RED: Domain.BuildArgs.Valid should be true — FAILS until WI-3")

	// The stored value should be a valid JSON string (for database storage)
	assert.True(t, strings.HasPrefix(domain.BuildArgs.String, "{") ||
		strings.HasPrefix(domain.BuildArgs.String, "["),
		"RED: Domain.BuildArgs.String should be valid JSON — FAILS until WI-3")
	// ─────────────────────────────────────────────────────────────────────────
}
