package models

// =============================================================================
// v0.56.0 Phase 2 RED Tests: CA Certs Cascade — Model Extensions
//
// These tests verify that CA certificate support is extended beyond Workspace
// to Ecosystem, Domain, and AppBuildConfig levels.
//
// Specifically, this file tests:
//   1. Ecosystem.CACerts sql.NullString — JSON storage for CA cert configs
//   2. EcosystemSpec.CACerts []CACertConfig — YAML representation
//   3. Domain.CACerts sql.NullString — JSON storage for CA cert configs
//   4. DomainSpec.CACerts []CACertConfig — YAML representation
//   5. AppBuildConfig.CACerts []CACertConfig — CA certs within app build config
//   6. AppBuildConfig.IsEmpty() returns false when CACerts is present
//
// RED: ALL tests in this file WILL NOT COMPILE until v0.56.0 WI is implemented:
//
//   - Ecosystem.CACerts sql.NullString field does not exist yet
//   - EcosystemSpec.CACerts []CACertConfig field does not exist yet
//   - Domain.CACerts sql.NullString field does not exist yet
//   - DomainSpec.CACerts []CACertConfig field does not exist yet
//   - AppBuildConfig.CACerts []CACertConfig field does not exist yet
//     (AppBuildConfig already has Args, Dockerfile, Buildpack, Target, Context)
//
// NOTE: The separate CACertsConfig struct (not reusing BuildArgsConfig) is
// consistent with the SRP architecture decision from Phase 1.
// =============================================================================

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// =============================================================================
// Ecosystem CA Certs Tests
// =============================================================================

// TestEcosystem_CACerts_ToYAML_RoundTrip verifies that CA certs set in
// EcosystemYAML are persisted via FromYAML and restored by ToYAML.
//
// RED: WILL NOT COMPILE — Ecosystem.CACerts and EcosystemSpec.CACerts
// do not exist yet (v0.56.0 WI).
func TestEcosystem_CACerts_ToYAML_RoundTrip(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	// EcosystemSpec.CACerts field does not exist yet.
	ecoYAML := EcosystemYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Ecosystem",
		Metadata: EcosystemMetadata{
			Name: "corporate-platform",
		},
		Spec: EcosystemSpec{
			Description: "Corporate development platform",
			Theme:       "tokyonight-night",
			CACerts: []CACertConfig{
				{
					Name:        "corp-root-ca",
					VaultSecret: "corp-root-ca-pem",
				},
				{
					Name:             "corp-int-ca",
					VaultSecret:      "corp-int-ca-pem",
					VaultEnvironment: "prod",
					VaultField:       "certificate",
				},
			},
		},
	}

	eco := &Ecosystem{}
	eco.FromYAML(ecoYAML)

	// Verify FromYAML persisted the CA certs as JSON
	require.True(t, eco.CACerts.Valid,
		"RED: Ecosystem.CACerts.Valid should be true after FromYAML with spec.caCerts")

	// Verify round-trip: ToYAML restores certs from the stored JSON
	result := eco.ToYAML(nil)
	require.Len(t, result.Spec.CACerts, 2,
		"RED: ToYAML should restore 2 CA certs from Ecosystem.CACerts JSON")

	// Build a name→cert map for order-independent assertions
	certsByName := make(map[string]CACertConfig)
	for _, c := range result.Spec.CACerts {
		certsByName[c.Name] = c
	}

	require.Contains(t, certsByName, "corp-root-ca")
	assert.Equal(t, "corp-root-ca-pem", certsByName["corp-root-ca"].VaultSecret,
		"corp-root-ca vault secret should survive round-trip")

	require.Contains(t, certsByName, "corp-int-ca")
	assert.Equal(t, "corp-int-ca-pem", certsByName["corp-int-ca"].VaultSecret,
		"corp-int-ca vault secret should survive round-trip")
	assert.Equal(t, "prod", certsByName["corp-int-ca"].VaultEnvironment,
		"corp-int-ca vault environment should survive round-trip")
	assert.Equal(t, "certificate", certsByName["corp-int-ca"].VaultField,
		"corp-int-ca vault field should survive round-trip")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestEcosystem_CACerts_EmptyOmitted verifies that when no CA certs are set
// on an Ecosystem, the "caCerts:" field is absent from serialized YAML output.
//
// RED: WILL NOT COMPILE — Ecosystem.CACerts and EcosystemSpec.CACerts
// do not exist yet (v0.56.0 WI).
func TestEcosystem_CACerts_EmptyOmitted(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// Ecosystem.CACerts does not exist yet.
	eco := &Ecosystem{
		Name: "no-ca-certs-eco",
	}
	// CACerts is not set — should remain zero-value (not valid)
	assert.False(t, eco.CACerts.Valid,
		"RED: Ecosystem.CACerts.Valid should be false when not set")

	result := eco.ToYAML(nil)
	data, err := yaml.Marshal(result)
	require.NoError(t, err, "should marshal ecosystem YAML without error")

	yamlStr := string(data)
	assert.False(t, strings.Contains(yamlStr, "caCerts"),
		"RED: 'caCerts' should NOT appear in YAML when CACerts is empty — field needs omitempty")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestEcosystem_CACerts_FromYAMLString verifies that parsing an EcosystemYAML
// document string with spec.caCerts correctly populates the model.
//
// RED: WILL NOT COMPILE — EcosystemSpec.CACerts does not exist yet (v0.56.0 WI).
func TestEcosystem_CACerts_FromYAMLString(t *testing.T) {
	yamlContent := `
apiVersion: devopsmaestro.io/v1
kind: Ecosystem
metadata:
  name: corporate-platform
spec:
  description: Corporate development platform
  caCerts:
    - name: corp-root-ca
      vaultSecret: corp-root-ca-pem
    - name: corp-int-ca
      vaultSecret: corp-int-ca-pem
      vaultEnvironment: prod
`

	var ecoYAML EcosystemYAML
	err := yaml.Unmarshal([]byte(yamlContent), &ecoYAML)
	require.NoError(t, err, "should parse ecosystem YAML with caCerts")

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// EcosystemSpec.CACerts does not exist yet.
	require.Len(t, ecoYAML.Spec.CACerts, 2,
		"RED: EcosystemSpec.CACerts should have 2 entries after YAML parse")
	assert.Equal(t, "corp-root-ca", ecoYAML.Spec.CACerts[0].Name)
	assert.Equal(t, "corp-root-ca-pem", ecoYAML.Spec.CACerts[0].VaultSecret)
	assert.Equal(t, "corp-int-ca", ecoYAML.Spec.CACerts[1].Name)
	assert.Equal(t, "prod", ecoYAML.Spec.CACerts[1].VaultEnvironment)
	// ─────────────────────────────────────────────────────────────────────────
}

// TestEcosystem_CACerts_PreservesBuildArgs verifies that adding CA certs to
// an Ecosystem does not break the existing build args (SRP: separate fields).
//
// RED: WILL NOT COMPILE — EcosystemSpec.CACerts does not exist yet.
func TestEcosystem_CACerts_PreservesBuildArgs(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	ecoYAML := EcosystemYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Ecosystem",
		Metadata:   EcosystemMetadata{Name: "full-eco"},
		Spec: EcosystemSpec{
			Theme: "tokyonight-night",
			Build: BuildArgsConfig{
				Args: map[string]string{
					"PIP_INDEX_URL": "https://pypi.corp.com",
				},
			},
			CACerts: []CACertConfig{
				{Name: "corp-root-ca", VaultSecret: "corp-root-ca-pem"},
			},
		},
	}

	eco := &Ecosystem{}
	eco.FromYAML(ecoYAML)

	// Both BuildArgs and CACerts should be persisted separately
	require.True(t, eco.BuildArgs.Valid,
		"BuildArgs should still be valid after adding CACerts")
	require.True(t, eco.CACerts.Valid,
		"RED: CACerts should be valid after FromYAML with spec.caCerts")

	// Round-trip: both should survive
	result := eco.ToYAML(nil)
	assert.Equal(t, "https://pypi.corp.com", result.Spec.Build.Args["PIP_INDEX_URL"],
		"BuildArgs should survive round-trip alongside CACerts")
	require.Len(t, result.Spec.CACerts, 1,
		"CACerts should survive round-trip alongside BuildArgs")
	assert.Equal(t, "corp-root-ca", result.Spec.CACerts[0].Name,
		"CACerts name should survive round-trip")
	// ─────────────────────────────────────────────────────────────────────────
}

// =============================================================================
// Domain CA Certs Tests
// =============================================================================

// TestDomain_CACerts_ToYAML_RoundTrip verifies that CA certs set in
// DomainYAML are persisted via FromYAML and restored by ToYAML.
//
// RED: WILL NOT COMPILE — Domain.CACerts and DomainSpec.CACerts
// do not exist yet (v0.56.0 WI).
func TestDomain_CACerts_ToYAML_RoundTrip(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	// DomainSpec.CACerts field does not exist yet.
	domYAML := DomainYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Domain",
		Metadata: DomainMetadata{
			Name:      "data-science",
			Ecosystem: "corporate-platform",
		},
		Spec: DomainSpec{
			CACerts: []CACertConfig{
				{
					Name:        "domain-specific-ca",
					VaultSecret: "domain-ca-pem",
				},
			},
		},
	}

	domain := &Domain{}
	domain.FromYAML(domYAML)

	// Verify FromYAML persisted the CA certs
	require.True(t, domain.CACerts.Valid,
		"RED: Domain.CACerts.Valid should be true after FromYAML with spec.caCerts")

	// Verify round-trip
	result := domain.ToYAML("corporate-platform", nil)
	require.Len(t, result.Spec.CACerts, 1,
		"RED: ToYAML should restore 1 CA cert from Domain.CACerts JSON")
	assert.Equal(t, "domain-specific-ca", result.Spec.CACerts[0].Name,
		"domain-specific-ca name should survive round-trip")
	assert.Equal(t, "domain-ca-pem", result.Spec.CACerts[0].VaultSecret,
		"domain-specific-ca vault secret should survive round-trip")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestDomain_CACerts_EmptyOmitted verifies that when no CA certs are set
// on a Domain, the "caCerts:" field is absent from serialized YAML output.
//
// RED: WILL NOT COMPILE — Domain.CACerts and DomainSpec.CACerts
// do not exist yet (v0.56.0 WI).
func TestDomain_CACerts_EmptyOmitted(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// Domain.CACerts does not exist yet.
	domain := &Domain{
		Name: "no-ca-certs-domain",
	}
	assert.False(t, domain.CACerts.Valid,
		"RED: Domain.CACerts.Valid should be false when not set")

	result := domain.ToYAML("my-eco", nil)
	data, err := yaml.Marshal(result)
	require.NoError(t, err, "should marshal domain YAML without error")

	yamlStr := string(data)
	assert.False(t, strings.Contains(yamlStr, "caCerts"),
		"RED: 'caCerts' should NOT appear in YAML when CACerts is empty")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestDomain_CACerts_PreservesBuildArgs verifies that adding CA certs to
// a Domain does not break the existing build args.
//
// RED: WILL NOT COMPILE — DomainSpec.CACerts does not exist yet.
func TestDomain_CACerts_PreservesBuildArgs(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	domYAML := DomainYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Domain",
		Metadata: DomainMetadata{
			Name:      "backend",
			Ecosystem: "corp-eco",
		},
		Spec: DomainSpec{
			Build: BuildArgsConfig{
				Args: map[string]string{
					"MAVEN_REPO": "https://nexus.corp.com",
				},
			},
			CACerts: []CACertConfig{
				{Name: "domain-ca", VaultSecret: "domain-ca-pem"},
			},
		},
	}

	domain := &Domain{}
	domain.FromYAML(domYAML)

	// Both should be persisted
	require.True(t, domain.BuildArgs.Valid,
		"BuildArgs should still be valid after adding CACerts")
	require.True(t, domain.CACerts.Valid,
		"RED: CACerts should be valid after FromYAML with spec.caCerts")

	// Round-trip
	result := domain.ToYAML("corp-eco", nil)
	assert.Equal(t, "https://nexus.corp.com", result.Spec.Build.Args["MAVEN_REPO"],
		"BuildArgs should survive round-trip alongside CACerts")
	require.Len(t, result.Spec.CACerts, 1,
		"CACerts should survive round-trip alongside BuildArgs")
	// ─────────────────────────────────────────────────────────────────────────
}

// =============================================================================
// App Build Config CA Certs Tests
// =============================================================================

// TestApp_CACerts_ToYAML_RoundTrip verifies that CA certs in AppBuildConfig
// survive a FromYAML → GetBuildConfig round-trip.
//
// RED: WILL NOT COMPILE — AppBuildConfig.CACerts field does not exist yet.
func TestApp_CACerts_ToYAML_RoundTrip(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	// AppBuildConfig.CACerts does not exist yet.
	appYAML := AppYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "App",
		Metadata: AppMetadata{
			Name:   "ml-api",
			Domain: "data-science",
		},
		Spec: AppSpec{
			Path: "/code/ml-api",
			Build: AppBuildConfig{
				Dockerfile: "Dockerfile.dev",
				Args: map[string]string{
					"CGO_ENABLED": "0",
				},
				CACerts: []CACertConfig{
					{
						Name:        "app-specific-ca",
						VaultSecret: "app-ca-pem",
					},
				},
			},
		},
	}

	app := &App{}
	app.FromYAML(appYAML)

	// Verify build config was stored
	buildCfg := app.GetBuildConfig()
	require.NotNil(t, buildCfg,
		"RED: GetBuildConfig() should return non-nil after FromYAML with caCerts")

	// Verify existing fields survived
	assert.Equal(t, "Dockerfile.dev", buildCfg.Dockerfile,
		"Dockerfile should survive round-trip alongside CACerts")
	assert.Equal(t, "0", buildCfg.Args["CGO_ENABLED"],
		"Args should survive round-trip alongside CACerts")

	// Verify CA certs survived
	require.Len(t, buildCfg.CACerts, 1,
		"RED: AppBuildConfig.CACerts should have 1 cert after round-trip")
	assert.Equal(t, "app-specific-ca", buildCfg.CACerts[0].Name,
		"app-specific-ca name should survive FromYAML → GetBuildConfig round-trip")
	assert.Equal(t, "app-ca-pem", buildCfg.CACerts[0].VaultSecret,
		"app-specific-ca vault secret should survive round-trip")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestAppBuildConfig_IsEmpty_WithCACerts verifies that AppBuildConfig.IsEmpty()
// returns false when only CACerts is populated (no Dockerfile/Buildpack/Args).
//
// RED: WILL NOT COMPILE — AppBuildConfig.CACerts field does not exist yet.
// Also: AppBuildConfig.IsEmpty() already exists but must be updated to check CACerts.
func TestAppBuildConfig_IsEmpty_WithCACerts(t *testing.T) {
	tests := []struct {
		name      string
		cfg       AppBuildConfig
		wantEmpty bool
	}{
		{
			name:      "empty struct is empty",
			cfg:       AppBuildConfig{},
			wantEmpty: true,
		},
		{
			name: "caCerts only is not empty",
			cfg: AppBuildConfig{
				// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────
				// AppBuildConfig.CACerts does not exist yet.
				CACerts: []CACertConfig{
					{Name: "my-ca", VaultSecret: "my-ca-pem"},
				},
				// ─────────────────────────────────────────────────────────────
			},
			wantEmpty: false,
		},
		{
			name: "args only is not empty",
			cfg: AppBuildConfig{
				Args: map[string]string{"CGO_ENABLED": "0"},
			},
			wantEmpty: false,
		},
		{
			name: "dockerfile plus caCerts is not empty",
			cfg: AppBuildConfig{
				Dockerfile: "Dockerfile",
				// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────
				CACerts: []CACertConfig{
					{Name: "my-ca", VaultSecret: "my-ca-pem"},
				},
				// ─────────────────────────────────────────────────────────────
			},
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.IsEmpty()
			assert.Equal(t, tt.wantEmpty, got,
				"AppBuildConfig.IsEmpty() mismatch for case %q", tt.name)
		})
	}
}

// TestApp_CACerts_FromYAMLString verifies that parsing an AppYAML document
// string with spec.build.caCerts correctly populates AppBuildConfig.CACerts.
//
// RED: WILL NOT COMPILE — AppBuildConfig.CACerts does not exist yet.
func TestApp_CACerts_FromYAMLString(t *testing.T) {
	yamlContent := `
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: ml-api
  domain: data-science
spec:
  path: /code/ml-api
  build:
    dockerfile: Dockerfile.dev
    caCerts:
      - name: app-ca
        vaultSecret: app-ca-pem
        vaultEnvironment: prod
`

	var appYAML AppYAML
	err := yaml.Unmarshal([]byte(yamlContent), &appYAML)
	require.NoError(t, err, "should parse app YAML with build.caCerts")

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// AppBuildConfig.CACerts does not exist yet.
	require.Len(t, appYAML.Spec.Build.CACerts, 1,
		"RED: AppSpec.Build.CACerts should have 1 entry after YAML parse")
	assert.Equal(t, "app-ca", appYAML.Spec.Build.CACerts[0].Name)
	assert.Equal(t, "app-ca-pem", appYAML.Spec.Build.CACerts[0].VaultSecret)
	assert.Equal(t, "prod", appYAML.Spec.Build.CACerts[0].VaultEnvironment)
	// ─────────────────────────────────────────────────────────────────────────

	// Existing fields should still work
	assert.Equal(t, "ml-api", appYAML.Metadata.Name)
	assert.Equal(t, "Dockerfile.dev", appYAML.Spec.Build.Dockerfile)
}
