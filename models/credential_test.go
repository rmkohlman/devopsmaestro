package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yamlv3 "gopkg.in/yaml.v3"
)

// =============================================================================
// Credential YAML Struct Tests
// =============================================================================

func TestCredentialYAML_StructFields(t *testing.T) {
	yaml := CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata: CredentialMetadata{
			Name:      "GITHUB_TOKEN",
			Ecosystem: "testlab",
		},
		Spec: CredentialSpec{
			Source:      "vault",
			VaultSecret: "dvm-github-token",
			Description: "GitHub PAT for private repo access",
		},
	}

	assert.Equal(t, "devopsmaestro.io/v1", yaml.APIVersion)
	assert.Equal(t, "Credential", yaml.Kind)
	assert.Equal(t, "GITHUB_TOKEN", yaml.Metadata.Name)
	assert.Equal(t, "testlab", yaml.Metadata.Ecosystem)
	assert.Equal(t, "vault", yaml.Spec.Source)
	assert.Equal(t, "dvm-github-token", yaml.Spec.VaultSecret)
	assert.Equal(t, "GitHub PAT for private repo access", yaml.Spec.Description)
}

func TestCredentialMetadata_ScopeFields(t *testing.T) {
	tests := []struct {
		name      string
		meta      CredentialMetadata
		wantScope string
	}{
		{
			name:      "ecosystem scope",
			meta:      CredentialMetadata{Name: "TOKEN", Ecosystem: "testlab"},
			wantScope: "testlab",
		},
		{
			name:      "domain scope",
			meta:      CredentialMetadata{Name: "TOKEN", Domain: "backend"},
			wantScope: "backend",
		},
		{
			name:      "app scope",
			meta:      CredentialMetadata{Name: "TOKEN", App: "rust-service"},
			wantScope: "rust-service",
		},
		{
			name:      "workspace scope",
			meta:      CredentialMetadata{Name: "TOKEN", Workspace: "main"},
			wantScope: "main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch {
			case tt.meta.Ecosystem != "":
				assert.Equal(t, tt.wantScope, tt.meta.Ecosystem)
			case tt.meta.Domain != "":
				assert.Equal(t, tt.wantScope, tt.meta.Domain)
			case tt.meta.App != "":
				assert.Equal(t, tt.wantScope, tt.meta.App)
			case tt.meta.Workspace != "":
				assert.Equal(t, tt.wantScope, tt.meta.Workspace)
			}
		})
	}
}

func TestCredentialSpec_EnvSource(t *testing.T) {
	yaml := CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata: CredentialMetadata{
			Name: "NPM_TOKEN",
			App:  "rust-service",
		},
		Spec: CredentialSpec{
			Source:      "env",
			EnvVar:      "MY_NPM",
			Description: "npm publish token",
		},
	}

	assert.Equal(t, "env", yaml.Spec.Source)
	assert.Equal(t, "MY_NPM", yaml.Spec.EnvVar)
	assert.Empty(t, yaml.Spec.VaultSecret)
}

// =============================================================================
// Credential ToYAML Tests
// =============================================================================

func TestCredentialDB_ToYAML_Vault(t *testing.T) {
	vaultSecret := "dvm-github-token"
	desc := "GitHub PAT"
	cred := &CredentialDB{
		Name:        "GITHUB_TOKEN",
		ScopeType:   CredentialScopeEcosystem,
		Source:      "vault",
		VaultSecret: &vaultSecret,
		Description: &desc,
	}

	yaml := cred.ToYAML("testlab")

	assert.Equal(t, "devopsmaestro.io/v1", yaml.APIVersion)
	assert.Equal(t, "Credential", yaml.Kind)
	assert.Equal(t, "GITHUB_TOKEN", yaml.Metadata.Name)
	assert.Equal(t, "testlab", yaml.Metadata.Ecosystem)
	assert.Empty(t, yaml.Metadata.Domain)
	assert.Empty(t, yaml.Metadata.App)
	assert.Empty(t, yaml.Metadata.Workspace)
	assert.Equal(t, "vault", yaml.Spec.Source)
	assert.Equal(t, "dvm-github-token", yaml.Spec.VaultSecret)
	assert.Empty(t, yaml.Spec.EnvVar)
	assert.Equal(t, "GitHub PAT", yaml.Spec.Description)
}

func TestCredentialDB_ToYAML_Env(t *testing.T) {
	envVar := "MY_NPM"
	cred := &CredentialDB{
		Name:      "NPM_TOKEN",
		ScopeType: CredentialScopeApp,
		Source:    "env",
		EnvVar:    &envVar,
	}

	yaml := cred.ToYAML("rust-service")

	assert.Equal(t, "devopsmaestro.io/v1", yaml.APIVersion)
	assert.Equal(t, "Credential", yaml.Kind)
	assert.Equal(t, "NPM_TOKEN", yaml.Metadata.Name)
	assert.Equal(t, "rust-service", yaml.Metadata.App)
	assert.Empty(t, yaml.Metadata.Ecosystem)
	assert.Equal(t, "env", yaml.Spec.Source)
	assert.Equal(t, "MY_NPM", yaml.Spec.EnvVar)
	assert.Empty(t, yaml.Spec.VaultSecret)
}

func TestCredentialDB_ToYAML_AllScopes(t *testing.T) {
	tests := []struct {
		name          string
		scopeType     CredentialScopeType
		scopeName     string
		wantEcosystem string
		wantDomain    string
		wantApp       string
		wantWorkspace string
	}{
		{
			name:          "ecosystem scope",
			scopeType:     CredentialScopeEcosystem,
			scopeName:     "testlab",
			wantEcosystem: "testlab",
		},
		{
			name:       "domain scope",
			scopeType:  CredentialScopeDomain,
			scopeName:  "backend",
			wantDomain: "backend",
		},
		{
			name:      "app scope",
			scopeType: CredentialScopeApp,
			scopeName: "rust-service",
			wantApp:   "rust-service",
		},
		{
			name:          "workspace scope",
			scopeType:     CredentialScopeWorkspace,
			scopeName:     "main",
			wantWorkspace: "main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vaultSecret := "dvm-test"
			cred := &CredentialDB{
				Name:        "TEST_CRED",
				ScopeType:   tt.scopeType,
				Source:      "vault",
				VaultSecret: &vaultSecret,
			}

			yaml := cred.ToYAML(tt.scopeName)

			assert.Equal(t, tt.wantEcosystem, yaml.Metadata.Ecosystem)
			assert.Equal(t, tt.wantDomain, yaml.Metadata.Domain)
			assert.Equal(t, tt.wantApp, yaml.Metadata.App)
			assert.Equal(t, tt.wantWorkspace, yaml.Metadata.Workspace)
		})
	}
}

func TestCredentialDB_ToYAML_NilOptionals(t *testing.T) {
	cred := &CredentialDB{
		Name:        "BARE_CRED",
		ScopeType:   CredentialScopeEcosystem,
		Source:      "vault",
		VaultSecret: nil,
		EnvVar:      nil,
		Description: nil,
	}

	yaml := cred.ToYAML("testlab")

	assert.Empty(t, yaml.Spec.VaultSecret)
	assert.Empty(t, yaml.Spec.EnvVar)
	assert.Empty(t, yaml.Spec.Description)
}

// =============================================================================
// Credential FromYAML Tests
// =============================================================================

func TestCredentialDB_FromYAML_Vault(t *testing.T) {
	yaml := CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata: CredentialMetadata{
			Name:      "GITHUB_TOKEN",
			Ecosystem: "testlab",
		},
		Spec: CredentialSpec{
			Source:      "vault",
			VaultSecret: "dvm-github-token",
			Description: "GitHub PAT",
		},
	}

	cred := &CredentialDB{}
	cred.FromYAML(yaml)

	assert.Equal(t, "GITHUB_TOKEN", cred.Name)
	assert.Equal(t, "vault", cred.Source)
	require.NotNil(t, cred.VaultSecret)
	assert.Equal(t, "dvm-github-token", *cred.VaultSecret)
	assert.Nil(t, cred.EnvVar)
	require.NotNil(t, cred.Description)
	assert.Equal(t, "GitHub PAT", *cred.Description)
}

func TestCredentialDB_FromYAML_Env(t *testing.T) {
	yaml := CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata: CredentialMetadata{
			Name: "NPM_TOKEN",
			App:  "rust-service",
		},
		Spec: CredentialSpec{
			Source: "env",
			EnvVar: "MY_NPM",
		},
	}

	cred := &CredentialDB{}
	cred.FromYAML(yaml)

	assert.Equal(t, "NPM_TOKEN", cred.Name)
	assert.Equal(t, "env", cred.Source)
	assert.Nil(t, cred.VaultSecret)
	require.NotNil(t, cred.EnvVar)
	assert.Equal(t, "MY_NPM", *cred.EnvVar)
	assert.Nil(t, cred.Description)
}

func TestCredentialDB_FromYAML_EmptyOptionals(t *testing.T) {
	yaml := CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata: CredentialMetadata{
			Name:      "BARE_CRED",
			Ecosystem: "testlab",
		},
		Spec: CredentialSpec{
			Source:      "vault",
			VaultSecret: "bare-secret",
		},
	}

	cred := &CredentialDB{}
	cred.FromYAML(yaml)

	assert.Equal(t, "BARE_CRED", cred.Name)
	assert.Equal(t, "vault", cred.Source)
	require.NotNil(t, cred.VaultSecret)
	assert.Equal(t, "bare-secret", *cred.VaultSecret)
	assert.Nil(t, cred.EnvVar)
	assert.Nil(t, cred.Description)
}

// =============================================================================
// Credential RoundTrip Tests
// =============================================================================

func TestCredentialDB_RoundTrip_Vault(t *testing.T) {
	vaultSecret := "dvm-gh"
	desc := "GitHub token"
	original := &CredentialDB{
		Name:        "GITHUB_TOKEN",
		ScopeType:   CredentialScopeEcosystem,
		Source:      "vault",
		VaultSecret: &vaultSecret,
		Description: &desc,
	}

	yaml := original.ToYAML("testlab")
	restored := &CredentialDB{}
	restored.FromYAML(yaml)

	assert.Equal(t, original.Name, restored.Name)
	assert.Equal(t, original.Source, restored.Source)
	require.NotNil(t, restored.VaultSecret)
	assert.Equal(t, *original.VaultSecret, *restored.VaultSecret)
	require.NotNil(t, restored.Description)
	assert.Equal(t, *original.Description, *restored.Description)
}

func TestCredentialDB_RoundTrip_Env(t *testing.T) {
	envVar := "MY_NPM"
	original := &CredentialDB{
		Name:      "NPM_TOKEN",
		ScopeType: CredentialScopeApp,
		Source:    "env",
		EnvVar:    &envVar,
	}

	yaml := original.ToYAML("rust-service")
	restored := &CredentialDB{}
	restored.FromYAML(yaml)

	assert.Equal(t, original.Name, restored.Name)
	assert.Equal(t, original.Source, restored.Source)
	require.NotNil(t, restored.EnvVar)
	assert.Equal(t, *original.EnvVar, *restored.EnvVar)
	assert.Nil(t, restored.VaultSecret)
}

// =============================================================================
// Credential ScopeTypeFromYAML Tests
// =============================================================================

func TestCredentialMetadata_ScopeType(t *testing.T) {
	tests := []struct {
		name     string
		meta     CredentialMetadata
		wantType CredentialScopeType
		wantName string
	}{
		{
			name:     "ecosystem",
			meta:     CredentialMetadata{Name: "T", Ecosystem: "testlab"},
			wantType: CredentialScopeEcosystem,
			wantName: "testlab",
		},
		{
			name:     "domain",
			meta:     CredentialMetadata{Name: "T", Domain: "backend"},
			wantType: CredentialScopeDomain,
			wantName: "backend",
		},
		{
			name:     "app",
			meta:     CredentialMetadata{Name: "T", App: "rust-svc"},
			wantType: CredentialScopeApp,
			wantName: "rust-svc",
		},
		{
			name:     "workspace",
			meta:     CredentialMetadata{Name: "T", Workspace: "main"},
			wantType: CredentialScopeWorkspace,
			wantName: "main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scopeType, scopeName := tt.meta.ScopeInfo()
			assert.Equal(t, tt.wantType, scopeType)
			assert.Equal(t, tt.wantName, scopeName)
		})
	}
}

// === Dual-Field Credential Tests (v0.37.1 / v0.40.0 vault) ===

// =============================================================================
// Dual-Field Model Tests
// =============================================================================

func TestCredentialDB_IsDualField(t *testing.T) {
	tests := []struct {
		name        string
		usernameVar *string
		passwordVar *string
		want        bool
	}{
		{
			name:        "legacy single-field (no vars)",
			usernameVar: nil,
			passwordVar: nil,
			want:        false,
		},
		{
			name:        "password-only",
			usernameVar: nil,
			passwordVar: strPtr("GITHUB_PAT"),
			want:        true,
		},
		{
			name:        "username-only",
			usernameVar: strPtr("GITHUB_USERNAME"),
			passwordVar: nil,
			want:        true,
		},
		{
			name:        "both fields",
			usernameVar: strPtr("GITHUB_USERNAME"),
			passwordVar: strPtr("GITHUB_PAT"),
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred := &CredentialDB{
				Name:        "github-creds",
				Source:      "vault",
				UsernameVar: tt.usernameVar,
				PasswordVar: tt.passwordVar,
			}
			assert.Equal(t, tt.want, cred.IsDualField())
		})
	}
}

// =============================================================================
// Dual-Field YAML Tests
// =============================================================================

func TestCredentialDB_ToYAML_DualField(t *testing.T) {
	vaultSecret := "github-pat"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"
	cred := &CredentialDB{
		Name:        "github-creds",
		ScopeType:   CredentialScopeEcosystem,
		Source:      "vault",
		VaultSecret: &vaultSecret,
		UsernameVar: &usernameVar,
		PasswordVar: &passwordVar,
	}

	y := cred.ToYAML("testlab")

	assert.Equal(t, "GITHUB_USERNAME", y.Spec.UsernameVar)
	assert.Equal(t, "GITHUB_PAT", y.Spec.PasswordVar)
}

func TestCredentialDB_FromYAML_DualField(t *testing.T) {
	y := CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata: CredentialMetadata{
			Name:      "github-creds",
			Ecosystem: "testlab",
		},
		Spec: CredentialSpec{
			Source:      "vault",
			VaultSecret: "github-pat",
			UsernameVar: "GH_USER",
			PasswordVar: "GH_PAT",
		},
	}

	cred := &CredentialDB{}
	cred.FromYAML(y)

	require.NotNil(t, cred.UsernameVar, "UsernameVar should not be nil after FromYAML")
	assert.Equal(t, "GH_USER", *cred.UsernameVar)
	require.NotNil(t, cred.PasswordVar, "PasswordVar should not be nil after FromYAML")
	assert.Equal(t, "GH_PAT", *cred.PasswordVar)
}

func TestCredentialDB_RoundTrip_DualField(t *testing.T) {
	vaultSecret := "github-pat"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"
	original := &CredentialDB{
		Name:        "github-creds",
		ScopeType:   CredentialScopeEcosystem,
		Source:      "vault",
		VaultSecret: &vaultSecret,
		UsernameVar: &usernameVar,
		PasswordVar: &passwordVar,
	}

	y := original.ToYAML("testlab")
	restored := &CredentialDB{}
	restored.FromYAML(y)

	require.NotNil(t, restored.UsernameVar, "UsernameVar should survive round trip")
	assert.Equal(t, *original.UsernameVar, *restored.UsernameVar)
	require.NotNil(t, restored.PasswordVar, "PasswordVar should survive round trip")
	assert.Equal(t, *original.PasswordVar, *restored.PasswordVar)
}

// =============================================================================
// Test Helpers
// =============================================================================

func strPtr(s string) *string { return &s }

// =============================================================================
// TDD Phase 2 (RED): MaestroVault Credential Tests (v0.40.0)
// =============================================================================

func TestCredentialDB_VaultFields(t *testing.T) {
	vaultSecret := "github-pat"
	vaultEnv := "production"
	vaultUsernameSecret := "github-username"

	cred := &CredentialDB{
		Name:                "github-creds",
		Source:              "vault",
		VaultSecret:         &vaultSecret,
		VaultEnv:            &vaultEnv,
		VaultUsernameSecret: &vaultUsernameSecret,
	}

	require.NotNil(t, cred.VaultSecret, "VaultSecret must be non-nil when set")
	assert.Equal(t, "github-pat", *cred.VaultSecret)

	require.NotNil(t, cred.VaultEnv, "VaultEnv must be non-nil when set")
	assert.Equal(t, "production", *cred.VaultEnv)

	require.NotNil(t, cred.VaultUsernameSecret, "VaultUsernameSecret must be non-nil when set")
	assert.Equal(t, "github-username", *cred.VaultUsernameSecret)
}

func TestCredentialDB_VaultFields_NilDefaults(t *testing.T) {
	cred := &CredentialDB{
		Name:   "non-vault-cred",
		Source: "env",
	}
	assert.Nil(t, cred.VaultSecret)
	assert.Nil(t, cred.VaultEnv)
	assert.Nil(t, cred.VaultUsernameSecret)
}

// =============================================================================
// CredentialSpec Vault Field Tests
// =============================================================================

func TestCredentialSpec_VaultFields(t *testing.T) {
	spec := CredentialSpec{
		Source:              "vault",
		VaultSecret:         "my-secret",
		VaultEnvironment:    "production",
		VaultUsernameSecret: "my-user-secret",
	}

	assert.Equal(t, "my-secret", spec.VaultSecret)
	assert.Equal(t, "production", spec.VaultEnvironment)
	assert.Equal(t, "my-user-secret", spec.VaultUsernameSecret)
}

func TestCredentialSpec_VaultFields_ZeroValues(t *testing.T) {
	spec := CredentialSpec{
		Source: "env",
		EnvVar: "SOME_VAR",
	}
	assert.Empty(t, spec.VaultSecret)
	assert.Empty(t, spec.VaultEnvironment)
	assert.Empty(t, spec.VaultUsernameSecret)
}

// =============================================================================
// CredentialDB.ToYAML Vault Tests
// =============================================================================

func TestCredentialDB_ToYAML_VaultFull(t *testing.T) {
	vaultSecret := "my-github-pat"
	vaultEnv := "production"

	cred := &CredentialDB{
		Name:        "GITHUB_TOKEN",
		ScopeType:   CredentialScopeEcosystem,
		Source:      "vault",
		VaultSecret: &vaultSecret,
		VaultEnv:    &vaultEnv,
	}
	y := cred.ToYAML("testlab")

	assert.Equal(t, "devopsmaestro.io/v1", y.APIVersion)
	assert.Equal(t, "Credential", y.Kind)
	assert.Equal(t, "GITHUB_TOKEN", y.Metadata.Name)
	assert.Equal(t, "testlab", y.Metadata.Ecosystem)
	assert.Equal(t, "vault", y.Spec.Source)
	assert.Equal(t, "my-github-pat", y.Spec.VaultSecret)
	assert.Equal(t, "production", y.Spec.VaultEnvironment)
}

func TestCredentialDB_ToYAML_Vault_WithUsernameSecret(t *testing.T) {
	vaultSecret := "github-pat"
	vaultUsernameSecret := "github-username"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"

	cred := &CredentialDB{
		Name:                "github-creds",
		ScopeType:           CredentialScopeEcosystem,
		Source:              "vault",
		VaultSecret:         &vaultSecret,
		VaultUsernameSecret: &vaultUsernameSecret,
		UsernameVar:         &usernameVar,
		PasswordVar:         &passwordVar,
	}
	y := cred.ToYAML("testlab")

	assert.Equal(t, "vault", y.Spec.Source)
	assert.Equal(t, "github-pat", y.Spec.VaultSecret)
	assert.Equal(t, "github-username", y.Spec.VaultUsernameSecret)
	assert.Equal(t, "GITHUB_USERNAME", y.Spec.UsernameVar)
	assert.Equal(t, "GITHUB_PAT", y.Spec.PasswordVar)
}

func TestCredentialDB_ToYAML_Vault_NilVaultEnv(t *testing.T) {
	vaultSecret := "my-secret"

	cred := &CredentialDB{
		Name:        "MY_SECRET",
		ScopeType:   CredentialScopeApp,
		Source:      "vault",
		VaultSecret: &vaultSecret,
		VaultEnv:    nil,
	}
	y := cred.ToYAML("test-app")

	assert.Equal(t, "vault", y.Spec.Source)
	assert.Equal(t, "my-secret", y.Spec.VaultSecret)
	assert.Empty(t, y.Spec.VaultEnvironment)
}

// =============================================================================
// CredentialDB.FromYAML Vault Tests
// =============================================================================

func TestCredentialDB_FromYAML_VaultFull(t *testing.T) {
	y := CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata:   CredentialMetadata{Name: "GITHUB_TOKEN", Ecosystem: "testlab"},
		Spec: CredentialSpec{
			Source:           "vault",
			VaultSecret:      "my-github-pat",
			VaultEnvironment: "production",
		},
	}
	cred := &CredentialDB{}
	cred.FromYAML(y)

	assert.Equal(t, "GITHUB_TOKEN", cred.Name)
	assert.Equal(t, "vault", cred.Source)
	require.NotNil(t, cred.VaultSecret)
	assert.Equal(t, "my-github-pat", *cred.VaultSecret)
	require.NotNil(t, cred.VaultEnv)
	assert.Equal(t, "production", *cred.VaultEnv)
	assert.Nil(t, cred.VaultUsernameSecret)
}

func TestCredentialDB_FromYAML_Vault_WithUsernameSecret(t *testing.T) {
	y := CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata:   CredentialMetadata{Name: "github-creds", Ecosystem: "testlab"},
		Spec: CredentialSpec{
			Source:              "vault",
			VaultSecret:         "github-pat",
			VaultEnvironment:    "staging",
			VaultUsernameSecret: "github-username",
			UsernameVar:         "GITHUB_USERNAME",
			PasswordVar:         "GITHUB_PAT",
		},
	}
	cred := &CredentialDB{}
	cred.FromYAML(y)

	require.NotNil(t, cred.VaultSecret)
	assert.Equal(t, "github-pat", *cred.VaultSecret)
	require.NotNil(t, cred.VaultEnv)
	assert.Equal(t, "staging", *cred.VaultEnv)
	require.NotNil(t, cred.VaultUsernameSecret)
	assert.Equal(t, "github-username", *cred.VaultUsernameSecret)
	require.NotNil(t, cred.UsernameVar)
	assert.Equal(t, "GITHUB_USERNAME", *cred.UsernameVar)
	require.NotNil(t, cred.PasswordVar)
	assert.Equal(t, "GITHUB_PAT", *cred.PasswordVar)
}

func TestCredentialDB_FromYAML_Vault_EmptyVaultEnvironment(t *testing.T) {
	y := CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata:   CredentialMetadata{Name: "MY_SECRET", App: "test-app"},
		Spec: CredentialSpec{
			Source:      "vault",
			VaultSecret: "my-secret",
		},
	}
	cred := &CredentialDB{}
	cred.FromYAML(y)

	require.NotNil(t, cred.VaultSecret)
	assert.Equal(t, "my-secret", *cred.VaultSecret)
	assert.Nil(t, cred.VaultEnv)
}

// =============================================================================
// CredentialDB Vault RoundTrip Tests
// =============================================================================

func TestCredentialDB_RoundTrip_VaultFull(t *testing.T) {
	vaultSecret := "my-vault-secret"
	vaultEnv := "production"

	original := &CredentialDB{
		Name:        "MY_TOKEN",
		ScopeType:   CredentialScopeEcosystem,
		Source:      "vault",
		VaultSecret: &vaultSecret,
		VaultEnv:    &vaultEnv,
	}
	y := original.ToYAML("testlab")
	restored := &CredentialDB{}
	restored.FromYAML(y)

	assert.Equal(t, original.Name, restored.Name)
	assert.Equal(t, original.Source, restored.Source)
	require.NotNil(t, restored.VaultSecret)
	assert.Equal(t, *original.VaultSecret, *restored.VaultSecret)
	require.NotNil(t, restored.VaultEnv)
	assert.Equal(t, *original.VaultEnv, *restored.VaultEnv)
}

func TestCredentialDB_RoundTrip_Vault_WithUsernameSecret(t *testing.T) {
	vaultSecret := "github-pat-secret"
	vaultEnv := "staging"
	vaultUsernameSecret := "github-username-secret"
	usernameVar := "GH_USER"
	passwordVar := "GH_PAT"

	original := &CredentialDB{
		Name:                "github-creds",
		ScopeType:           CredentialScopeEcosystem,
		Source:              "vault",
		VaultSecret:         &vaultSecret,
		VaultEnv:            &vaultEnv,
		VaultUsernameSecret: &vaultUsernameSecret,
		UsernameVar:         &usernameVar,
		PasswordVar:         &passwordVar,
	}
	y := original.ToYAML("testlab")
	restored := &CredentialDB{}
	restored.FromYAML(y)

	require.NotNil(t, restored.VaultSecret)
	assert.Equal(t, *original.VaultSecret, *restored.VaultSecret)
	require.NotNil(t, restored.VaultEnv)
	assert.Equal(t, *original.VaultEnv, *restored.VaultEnv)
	require.NotNil(t, restored.VaultUsernameSecret)
	assert.Equal(t, *original.VaultUsernameSecret, *restored.VaultUsernameSecret)
	require.NotNil(t, restored.UsernameVar)
	assert.Equal(t, *original.UsernameVar, *restored.UsernameVar)
	require.NotNil(t, restored.PasswordVar)
	assert.Equal(t, *original.PasswordVar, *restored.PasswordVar)
}

// =============================================================================
// VaultFields Field Tests
// =============================================================================

func TestCredentialDB_VaultFields_FieldExists(t *testing.T) {
	raw := `{"GITHUB_TOKEN":"token","GITHUB_USER":"username"}`

	cred := &CredentialDB{
		Name:        "github-creds",
		Source:      "vault",
		VaultSecret: strPtr("github/creds"),
		VaultFields: strPtr(raw),
	}

	require.NotNil(t, cred.VaultFields, "VaultFields must not be nil when set")
	assert.Equal(t, raw, *cred.VaultFields, "VaultFields must hold the JSON blob")
}

func TestCredentialDB_VaultFields_NilByDefault(t *testing.T) {
	cred := &CredentialDB{
		Name:        "env-cred",
		Source:      "env",
		EnvVar:      strPtr("MY_TOKEN"),
		VaultFields: nil,
	}

	assert.Nil(t, cred.VaultFields, "VaultFields must be nil when not set")
}

// =============================================================================
// HasVaultFields Tests
// =============================================================================

func TestCredentialDB_HasVaultFields_True(t *testing.T) {
	raw := `{"API_TOKEN":"token"}`

	cred := &CredentialDB{
		VaultFields: strPtr(raw),
	}
	result := cred.HasVaultFields()

	assert.True(t, result, "HasVaultFields must return true when VaultFields is set")
}

func TestCredentialDB_HasVaultFields_FalseWhenNil(t *testing.T) {
	cred := &CredentialDB{VaultFields: nil}
	result := cred.HasVaultFields()

	assert.False(t, result, "HasVaultFields must return false when VaultFields is nil")
}

func TestCredentialDB_HasVaultFields_FalseWhenEmptyJSON(t *testing.T) {
	cred := &CredentialDB{VaultFields: strPtr("{}")}
	result := cred.HasVaultFields()

	assert.False(t, result, "HasVaultFields must return false when VaultFields is empty JSON")
}

// =============================================================================
// GetVaultFieldsMap Tests
// =============================================================================

func TestCredentialDB_GetVaultFieldsMap_ParsesJSON(t *testing.T) {
	raw := `{"GITHUB_TOKEN":"token","GITHUB_USER":"username"}`

	cred := &CredentialDB{
		VaultFields: strPtr(raw),
	}
	result, err := cred.GetVaultFieldsMap()

	require.NoError(t, err, "GetVaultFieldsMap must not return error for valid JSON")
	require.Len(t, result, 2, "map must contain 2 entries")
	assert.Equal(t, "token", result["GITHUB_TOKEN"])
	assert.Equal(t, "username", result["GITHUB_USER"])
}

func TestCredentialDB_GetVaultFieldsMap_ReturnsEmptyMapWhenNil(t *testing.T) {
	cred := &CredentialDB{VaultFields: nil}
	result, err := cred.GetVaultFieldsMap()

	require.NoError(t, err, "GetVaultFieldsMap must not error when VaultFields is nil")
	assert.Nil(t, result, "GetVaultFieldsMap must return nil when VaultFields is nil")
	assert.Empty(t, result, "returned map must be empty when VaultFields is nil")
}

func TestCredentialDB_GetVaultFieldsMap_ErrorOnInvalidJSON(t *testing.T) {
	cred := &CredentialDB{VaultFields: strPtr("not-valid-json")}
	_, err := cred.GetVaultFieldsMap()

	assert.Error(t, err, "GetVaultFieldsMap must return error for invalid JSON in VaultFields")
}

// =============================================================================
// CredentialSpec VaultFields Tests
// =============================================================================

func TestCredentialSpec_VaultFields_FieldExists(t *testing.T) {
	spec := CredentialSpec{
		Source:      "vault",
		VaultSecret: "github/creds",
		VaultFields: map[string]string{
			"GITHUB_TOKEN": "token",
			"GITHUB_USER":  "username",
		},
	}

	assert.Len(t, spec.VaultFields, 2)
	assert.Equal(t, "token", spec.VaultFields["GITHUB_TOKEN"])
	assert.Equal(t, "username", spec.VaultFields["GITHUB_USER"])
}

func TestCredentialSpec_VaultFields_NilByDefault(t *testing.T) {
	spec := CredentialSpec{
		Source: "env",
		EnvVar: "MY_TOKEN",
	}

	assert.Nil(t, spec.VaultFields, "VaultFields must be nil when not set")
}

// =============================================================================
// CredentialYAML VaultFields Round-Trip Tests
// =============================================================================

func TestCredentialYAML_VaultFields_RoundTrip(t *testing.T) {
	original := CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata:   CredentialMetadata{Name: "github-creds", Ecosystem: "mylab"},
		Spec: CredentialSpec{
			Source:      "vault",
			VaultSecret: "github/creds",
			VaultFields: map[string]string{
				"GITHUB_TOKEN": "token",
				"GITHUB_USER":  "username",
			},
		},
	}

	data, err := yamlv3.Marshal(original)
	require.NoError(t, err, "yaml.Marshal must not error")

	var restored CredentialYAML
	err = yamlv3.Unmarshal(data, &restored)
	require.NoError(t, err, "yaml.Unmarshal must not error")

	assert.Equal(t, original.Spec.VaultFields, restored.Spec.VaultFields,
		"VaultFields must survive a marshal/unmarshal round-trip")
}

// =============================================================================
// Validate Tests
// =============================================================================

func TestCredentialDB_Validate(t *testing.T) {
	// VaultFields and dual-field are mutually exclusive
	cred := &CredentialDB{
		VaultFields: strPtr(`{"A":"a"}`),
		UsernameVar: strPtr("X"),
	}
	assert.Error(t, cred.Validate(), "vault fields and dual-field should be mutually exclusive")

	// VaultFields and VaultUsernameSecret are mutually exclusive
	vus := "user-secret"
	cred2 := &CredentialDB{
		VaultFields:         strPtr(`{"A":"a"}`),
		VaultUsernameSecret: &vus,
	}
	assert.Error(t, cred2.Validate(), "vault fields and vault username secret should be mutually exclusive")

	// Valid: no conflicts
	cred3 := &CredentialDB{
		Name:   "test",
		Source: "vault",
	}
	assert.NoError(t, cred3.Validate())
}
