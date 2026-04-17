package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// =============================================================================
// System Model Tests — Segment A.1 (#262)
//
// These tests verify:
//   1. SystemYAML struct with metadata (name, domain, ecosystem) and spec
//   2. ToYAML() serialization method
//   3. FromYAML() deserialization method
//   4. Build args round-trip (same pattern as Domain/Ecosystem)
//   5. CA certs round-trip
//   6. Empty build args omitted from YAML output (omitempty)
// =============================================================================

func TestSystem_BuildArgs_RoundTrip(t *testing.T) {
	sysYAML := SystemYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "System",
		Metadata: SystemMetadata{
			Name:   "payment-processing",
			Domain: "billing",
		},
		Spec: SystemSpec{
			Build: BuildArgsConfig{
				Args: map[string]string{
					"PAYMENT_GATEWAY": "stripe",
					"REGION":          "us-east-1",
				},
			},
		},
	}

	system := &System{}
	system.FromYAML(sysYAML)

	require.True(t, system.BuildArgs.Valid,
		"System.BuildArgs.Valid should be true after FromYAML with spec.build.args")

	result := system.ToYAML("billing", "", nil)
	require.NotNil(t, result.Spec.Build.Args,
		"ToYAML should populate Spec.Build.Args from BuildArgs")
	assert.Equal(t, "stripe", result.Spec.Build.Args["PAYMENT_GATEWAY"])
	assert.Equal(t, "us-east-1", result.Spec.Build.Args["REGION"])
}

func TestSystem_BuildArgs_MultipleKeys(t *testing.T) {
	sysYAML := SystemYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "System",
		Metadata:   SystemMetadata{Name: "multi-key-system", Domain: "my-domain"},
		Spec: SystemSpec{
			Build: BuildArgsConfig{
				Args: map[string]string{
					"SYS_ARG_1": "sys-val-1",
					"SYS_ARG_2": "sys-val-2",
					"SYS_ARG_3": "sys-val-3",
				},
			},
		},
	}

	system := &System{}
	system.FromYAML(sysYAML)

	require.True(t, system.BuildArgs.Valid)

	result := system.ToYAML("my-domain", "", nil)
	require.Len(t, result.Spec.Build.Args, 3)
	assert.Equal(t, "sys-val-1", result.Spec.Build.Args["SYS_ARG_1"])
	assert.Equal(t, "sys-val-2", result.Spec.Build.Args["SYS_ARG_2"])
	assert.Equal(t, "sys-val-3", result.Spec.Build.Args["SYS_ARG_3"])
}

func TestSystem_BuildArgs_Empty_OmittedFromYAML(t *testing.T) {
	system := &System{
		Name: "no-build-args-system",
	}
	assert.False(t, system.BuildArgs.Valid,
		"System.BuildArgs.Valid should be false when not set")

	result := system.ToYAML("my-domain", "", nil)
	data, err := yaml.Marshal(result)
	require.NoError(t, err, "should marshal system YAML without error")

	yamlStr := string(data)
	assert.False(t, strings.Contains(yamlStr, "build:"),
		"'build:' should NOT appear in YAML when BuildArgs is empty")
}

func TestSystem_BuildArgs_FromYAMLString(t *testing.T) {
	yamlContent := `
apiVersion: devopsmaestro.io/v1
kind: System
metadata:
  name: payment-processing
  domain: billing
  ecosystem: healthcare
spec:
  build:
    args:
      PAYMENT_GATEWAY: "stripe"
      REGION: "us-east-1"
`

	var sysYAML SystemYAML
	err := yaml.Unmarshal([]byte(yamlContent), &sysYAML)
	require.NoError(t, err, "should parse system YAML with build.args")

	require.NotNil(t, sysYAML.Spec.Build.Args)
	assert.Equal(t, "stripe", sysYAML.Spec.Build.Args["PAYMENT_GATEWAY"])
	assert.Equal(t, "us-east-1", sysYAML.Spec.Build.Args["REGION"])

	assert.Equal(t, "payment-processing", sysYAML.Metadata.Name)
	assert.Equal(t, "billing", sysYAML.Metadata.Domain)
	assert.Equal(t, "healthcare", sysYAML.Metadata.Ecosystem)
}

func TestSystem_BuildArgs_PreservesExistingFields(t *testing.T) {
	sysYAML := SystemYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "System",
		Metadata: SystemMetadata{
			Name:   "full-system",
			Domain: "backend",
			Annotations: map[string]string{
				"description": "Full featured system",
			},
		},
		Spec: SystemSpec{
			Theme: "catppuccin-mocha",
			Apps:  []string{"api-service", "worker"},
			Build: BuildArgsConfig{
				Args: map[string]string{
					"DATA_BACKEND": "spark",
				},
			},
		},
	}

	system := &System{}
	system.FromYAML(sysYAML)

	assert.Equal(t, "full-system", system.Name)
	assert.True(t, system.Theme.Valid)
	assert.Equal(t, "catppuccin-mocha", system.Theme.String)
	assert.True(t, system.Description.Valid)
	assert.Equal(t, "Full featured system", system.Description.String)

	require.True(t, system.BuildArgs.Valid)

	result := system.ToYAML("backend", "", []string{"api-service", "worker"})
	assert.Equal(t, "full-system", result.Metadata.Name)
	assert.Equal(t, "catppuccin-mocha", result.Spec.Theme)
	assert.Equal(t, "spark", result.Spec.Build.Args["DATA_BACKEND"])
	assert.Equal(t, []string{"api-service", "worker"}, result.Spec.Apps)
}

func TestSystem_BuildArgs_JSONEncoding(t *testing.T) {
	sysYAML := SystemYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "System",
		Metadata:   SystemMetadata{Name: "json-test-system", Domain: "my-domain"},
		Spec: SystemSpec{
			Build: BuildArgsConfig{
				Args: map[string]string{
					"MY_KEY": "my_value",
				},
			},
		},
	}

	system := &System{}
	system.FromYAML(sysYAML)

	require.True(t, system.BuildArgs.Valid)

	assert.True(t, strings.HasPrefix(system.BuildArgs.String, "{") ||
		strings.HasPrefix(system.BuildArgs.String, "["),
		"System.BuildArgs.String should be valid JSON")
}

func TestSystem_CACerts_RoundTrip(t *testing.T) {
	sysYAML := SystemYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "System",
		Metadata:   SystemMetadata{Name: "cert-system", Domain: "infra"},
		Spec: SystemSpec{
			CACerts: []CACertConfig{
				{Name: "internal-ca", VaultSecret: "pki/internal-ca"},
			},
		},
	}

	system := &System{}
	system.FromYAML(sysYAML)

	require.True(t, system.CACerts.Valid,
		"System.CACerts.Valid should be true after FromYAML with caCerts")

	result := system.ToYAML("infra", "", nil)
	require.Len(t, result.Spec.CACerts, 1)
	assert.Equal(t, "internal-ca", result.Spec.CACerts[0].Name)
	assert.Equal(t, "pki/internal-ca", result.Spec.CACerts[0].VaultSecret)
}

func TestSystem_ToYAML_Kind(t *testing.T) {
	system := &System{Name: "test-system"}
	result := system.ToYAML("my-domain", "", nil)

	assert.Equal(t, "devopsmaestro.io/v1", result.APIVersion)
	assert.Equal(t, "System", result.Kind)
	assert.Equal(t, "test-system", result.Metadata.Name)
	assert.Equal(t, "my-domain", result.Metadata.Domain)
}

func TestSystem_ToYAML_AllFields(t *testing.T) {
	sysYAML := SystemYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "System",
		Metadata: SystemMetadata{
			Name:   "complete-system",
			Domain: "platform",
			Annotations: map[string]string{
				"description": "A complete system",
			},
		},
		Spec: SystemSpec{
			Theme:           "tokyonight-night",
			NvimPackage:     "my-nvim-pkg",
			TerminalPackage: "my-terminal-pkg",
			Apps:            []string{"app-a", "app-b"},
		},
	}

	system := &System{}
	system.FromYAML(sysYAML)

	result := system.ToYAML("platform", "", []string{"app-a", "app-b"})

	assert.Equal(t, "complete-system", result.Metadata.Name)
	assert.Equal(t, "platform", result.Metadata.Domain)
	assert.Equal(t, "A complete system", result.Metadata.Annotations["description"])
	assert.Equal(t, "tokyonight-night", result.Spec.Theme)
	assert.Equal(t, "my-nvim-pkg", result.Spec.NvimPackage)
	assert.Equal(t, "my-terminal-pkg", result.Spec.TerminalPackage)
	assert.Equal(t, []string{"app-a", "app-b"}, result.Spec.Apps)
}
