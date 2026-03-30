package models

import (
	"fmt"
	"testing"

	"devopsmaestro/config"
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
			// Verify the scope field is populated
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
// Credential Validation Tests
// =============================================================================

func TestValidateCredentialYAML(t *testing.T) {
	tests := []struct {
		name    string
		yaml    CredentialYAML
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid vault credential",
			yaml: CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   CredentialMetadata{Name: "GITHUB_TOKEN", Ecosystem: "testlab"},
				Spec:       CredentialSpec{Source: "vault", VaultSecret: "dvm-gh"},
			},
			wantErr: false,
		},
		{
			name: "valid env credential",
			yaml: CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   CredentialMetadata{Name: "NPM_TOKEN", App: "rust-svc"},
				Spec:       CredentialSpec{Source: "env", EnvVar: "MY_NPM"},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			yaml: CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   CredentialMetadata{Ecosystem: "testlab"},
				Spec:       CredentialSpec{Source: "vault", VaultSecret: "dvm-gh"},
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "missing source",
			yaml: CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   CredentialMetadata{Name: "TOKEN", Ecosystem: "testlab"},
				Spec:       CredentialSpec{VaultSecret: "dvm-gh"},
			},
			wantErr: true,
			errMsg:  "source is required",
		},
		{
			name: "invalid source",
			yaml: CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   CredentialMetadata{Name: "TOKEN", Ecosystem: "testlab"},
				Spec:       CredentialSpec{Source: "plaintext"},
			},
			wantErr: true,
			errMsg:  "source must be 'vault' or 'env'",
		},
		{
			name: "vault missing vaultSecret",
			yaml: CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   CredentialMetadata{Name: "TOKEN", Ecosystem: "testlab"},
				Spec:       CredentialSpec{Source: "vault"},
			},
			wantErr: true,
			errMsg:  "vaultSecret is required for vault source",
		},
		{
			name: "env missing env-var",
			yaml: CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   CredentialMetadata{Name: "TOKEN", Ecosystem: "testlab"},
				Spec:       CredentialSpec{Source: "env"},
			},
			wantErr: true,
			errMsg:  "env-var is required for env source",
		},
		{
			name: "no scope specified",
			yaml: CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   CredentialMetadata{Name: "TOKEN"},
				Spec:       CredentialSpec{Source: "vault", VaultSecret: "dvm-gh"},
			},
			wantErr: true,
			errMsg:  "exactly one scope",
		},
		{
			name: "multiple scopes specified",
			yaml: CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   CredentialMetadata{Name: "TOKEN", Ecosystem: "testlab", App: "rust-svc"},
				Spec:       CredentialSpec{Source: "vault", VaultSecret: "dvm-gh"},
			},
			wantErr: true,
			errMsg:  "exactly one scope",
		},
		{
			name: "wrong kind",
			yaml: CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Registry",
				Metadata:   CredentialMetadata{Name: "TOKEN", Ecosystem: "testlab"},
				Spec:       CredentialSpec{Source: "vault", VaultSecret: "dvm-gh"},
			},
			wantErr: true,
			errMsg:  "kind must be 'Credential'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCredentialYAML(tt.yaml)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
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

func TestCredentialDB_ToUsernameConfig(t *testing.T) {
	vaultSecret := "github-pat"
	vaultUsernameSecret := "github-username"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"
	cred := &CredentialDB{
		Name:                "github-creds",
		Source:              "vault",
		VaultSecret:         &vaultSecret,
		VaultUsernameSecret: &vaultUsernameSecret,
		UsernameVar:         &usernameVar,
		PasswordVar:         &passwordVar,
	}

	cfg := cred.ToUsernameConfig()

	assert.Equal(t, config.SourceVault, cfg.Source)
	// ToUsernameConfig uses VaultUsernameSecret when available
	assert.Equal(t, "github-username", cfg.VaultSecret)
}

func TestCredentialDB_ToPasswordConfig(t *testing.T) {
	vaultSecret := "github-pat"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"
	cred := &CredentialDB{
		Name:        "github-creds",
		Source:      "vault",
		VaultSecret: &vaultSecret,
		UsernameVar: &usernameVar,
		PasswordVar: &passwordVar,
	}

	cfg := cred.ToPasswordConfig()

	assert.Equal(t, config.SourceVault, cfg.Source)
	assert.Equal(t, "github-pat", cfg.VaultSecret)
}

func TestCredentialsToMap_DualField_FanOut(t *testing.T) {
	vaultSecret := "github-pat"
	vaultUsernameSecret := "github-username"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"
	cred := &CredentialDB{
		Name:                "github-creds",
		Source:              "vault",
		VaultSecret:         &vaultSecret,
		VaultUsernameSecret: &vaultUsernameSecret,
		UsernameVar:         &usernameVar,
		PasswordVar:         &passwordVar,
	}

	result := CredentialsToMap([]*CredentialDB{cred})

	require.Len(t, result, 2, "dual-field credential should fan out to 2 map entries")

	userCfg, ok := result["GITHUB_USERNAME"]
	require.True(t, ok, "map should contain GITHUB_USERNAME key")
	assert.Equal(t, config.SourceVault, userCfg.Source)
	// ToUsernameConfig uses VaultUsernameSecret for the username lookup
	assert.Equal(t, "github-username", userCfg.VaultSecret)

	passCfg, ok := result["GITHUB_PAT"]
	require.True(t, ok, "map should contain GITHUB_PAT key")
	assert.Equal(t, config.SourceVault, passCfg.Source)
	assert.Equal(t, "github-pat", passCfg.VaultSecret)
}

func TestCredentialsToMap_MixedLegacyAndDualField(t *testing.T) {
	npmEnvVar := "NPM_TOKEN"
	legacyCred := &CredentialDB{
		Name:   "NPM_TOKEN",
		Source: "env",
		EnvVar: &npmEnvVar,
	}

	vaultSecret := "github-pat"
	vaultUsernameSecret := "github-username"
	ghUser := "GH_USER"
	ghPat := "GH_PAT"
	dualCred := &CredentialDB{
		Name:                "github-creds",
		Source:              "vault",
		VaultSecret:         &vaultSecret,
		VaultUsernameSecret: &vaultUsernameSecret,
		UsernameVar:         &ghUser,
		PasswordVar:         &ghPat,
	}

	result := CredentialsToMap([]*CredentialDB{legacyCred, dualCred})

	require.Len(t, result, 3, "mixed credentials should produce 3 map entries")
	assert.Contains(t, result, "NPM_TOKEN")
	assert.Contains(t, result, "GH_USER")
	assert.Contains(t, result, "GH_PAT")
}

func TestCredentialsToMap_DualField_PasswordOnly(t *testing.T) {
	vaultSecret := "github-pat"
	passwordVar := "GITHUB_PAT"
	cred := &CredentialDB{
		Name:        "github-creds",
		Source:      "vault",
		VaultSecret: &vaultSecret,
		UsernameVar: nil,
		PasswordVar: &passwordVar,
	}

	result := CredentialsToMap([]*CredentialDB{cred})

	require.Len(t, result, 1, "password-only dual-field should produce 1 map entry")
	passCfg, ok := result["GITHUB_PAT"]
	require.True(t, ok, "map should contain GITHUB_PAT key")
	assert.Equal(t, config.SourceVault, passCfg.Source)
	assert.Equal(t, "github-pat", passCfg.VaultSecret)
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
// Dual-Field Validation Tests
// =============================================================================

func TestValidateCredentialYAML_DualField(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		vaultSecret string
		envVar      string
		usernameVar string
		passwordVar string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid dual-field vault",
			source:      "vault",
			vaultSecret: "github-pat",
			usernameVar: "GH_USER",
			passwordVar: "GH_PAT",
			wantErr:     false,
		},
		{
			name:        "dual-field with env source rejected",
			source:      "env",
			envVar:      "X",
			usernameVar: "GH_USER",
			wantErr:     true,
			errMsg:      "vault",
		},
		{
			name:        "valid vault password-only",
			source:      "vault",
			vaultSecret: "svc-secret",
			passwordVar: "TOKEN",
			wantErr:     false,
		},
		{
			name:        "valid vault username-only",
			source:      "vault",
			vaultSecret: "svc-secret",
			usernameVar: "USER",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			y := CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   CredentialMetadata{Name: "test-cred", Ecosystem: "testlab"},
				Spec: CredentialSpec{
					Source:      tt.source,
					VaultSecret: tt.vaultSecret,
					EnvVar:      tt.envVar,
					UsernameVar: tt.usernameVar,
					PasswordVar: tt.passwordVar,
				},
			}
			err := ValidateCredentialYAML(y)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// =============================================================================
// Dual-Field Env Key Validation Tests (v0.37.1)
// =============================================================================

func TestValidateCredentialYAML_DualFieldEnvKeyValidation(t *testing.T) {
	tests := []struct {
		name        string
		usernameVar string
		passwordVar string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "invalid usernameVar - lowercase",
			usernameVar: "github_user",
			passwordVar: "GITHUB_PAT",
			wantErr:     true,
			errMsg:      "invalid usernameVar",
		},
		{
			name:        "invalid passwordVar - lowercase",
			usernameVar: "GITHUB_USERNAME",
			passwordVar: "github_pat",
			wantErr:     true,
			errMsg:      "invalid passwordVar",
		},
		{
			name:        "forbidden usernameVar - LD_PRELOAD denylist",
			usernameVar: "LD_PRELOAD",
			passwordVar: "GITHUB_PAT",
			wantErr:     true,
			errMsg:      "invalid usernameVar",
		},
		{
			name:        "reserved usernameVar - DVM_ prefix",
			usernameVar: "DVM_USER",
			passwordVar: "GITHUB_PAT",
			wantErr:     true,
			errMsg:      "invalid usernameVar",
		},
		{
			name:        "valid dual-field vars",
			usernameVar: "GITHUB_USERNAME",
			passwordVar: "GITHUB_PAT",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			y := CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   CredentialMetadata{Name: "test-cred", Ecosystem: "test"},
				Spec: CredentialSpec{
					Source:      "vault",
					VaultSecret: "test-secret",
					UsernameVar: tt.usernameVar,
					PasswordVar: tt.passwordVar,
				},
			}
			err := ValidateCredentialYAML(y)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// =============================================================================
// Test Helpers
// =============================================================================

func strPtr(s string) *string { return &s }

// =============================================================================
// TDD Phase 2 (RED): MaestroVault Credential Tests (v0.40.0)
// =============================================================================
// Replacing macOS Keychain with MaestroVault as the secrets backend.
//
// New fields on CredentialDB that MUST exist after implementation:
//
//	type CredentialDB struct {
//	    ...existing fields (keep: EnvVar, UsernameVar, PasswordVar, Description)...
//	    VaultSecret         *string `db:"vault_secret" json:"vault_secret,omitempty"`
//	    VaultEnv            *string `db:"vault_env" json:"vault_env,omitempty"`
//	    VaultUsernameSecret *string `db:"vault_username_secret" json:"vault_username_secret,omitempty"`
//	}
//
// New fields on CredentialSpec that MUST exist after implementation:
//
//	type CredentialSpec struct {
//	    ...existing fields (keep: EnvVar, Description, UsernameVar, PasswordVar)...
//	    VaultSecret         string `yaml:"vaultSecret,omitempty" json:"vaultSecret,omitempty"`
//	    VaultEnvironment    string `yaml:"vaultEnvironment,omitempty" json:"vaultEnvironment,omitempty"`
//	    VaultUsernameSecret string `yaml:"vaultUsernameSecret,omitempty" json:"vaultUsernameSecret,omitempty"`
//	}
//
// ALL tests in this section WILL FAIL TO COMPILE until the above fields are
// added to models/credential.go.
// =============================================================================

// ---------------------------------------------------------------------------
// Section: CredentialDB Vault Field Tests
// ---------------------------------------------------------------------------

// TestCredentialDB_VaultFields verifies that CredentialDB has the three vault
// fields: VaultSecret, VaultEnv, and VaultUsernameSecret, all of type *string.
//
// WILL FAIL TO COMPILE — the three VaultXxx fields do not exist yet.
func TestCredentialDB_VaultFields(t *testing.T) {
	vaultSecret := "github-pat"
	vaultEnv := "production"
	vaultUsernameSecret := "github-username"

	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	// CredentialDB.VaultSecret, .VaultEnv, .VaultUsernameSecret do not exist yet.
	cred := &CredentialDB{
		Name:                "github-creds",
		Source:              "vault",
		VaultSecret:         &vaultSecret,
		VaultEnv:            &vaultEnv,
		VaultUsernameSecret: &vaultUsernameSecret,
	}
	// ─────────────────────────────────────────────────────────────────────────

	require.NotNil(t, cred.VaultSecret, "VaultSecret must be non-nil when set")
	assert.Equal(t, "github-pat", *cred.VaultSecret,
		"VaultSecret must store and return the secret name")

	require.NotNil(t, cred.VaultEnv, "VaultEnv must be non-nil when set")
	assert.Equal(t, "production", *cred.VaultEnv,
		"VaultEnv must store and return the environment name")

	require.NotNil(t, cred.VaultUsernameSecret, "VaultUsernameSecret must be non-nil when set")
	assert.Equal(t, "github-username", *cred.VaultUsernameSecret,
		"VaultUsernameSecret must store and return the username secret name")
}

// TestCredentialDB_VaultFields_NilDefaults verifies that the three vault fields
// default to nil when not set (they are *string, not string).
//
// WILL FAIL TO COMPILE — the three VaultXxx fields do not exist yet.
func TestCredentialDB_VaultFields_NilDefaults(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	cred := &CredentialDB{
		Name:   "non-vault-cred",
		Source: "env",
	}
	assert.Nil(t, cred.VaultSecret,
		"VaultSecret must default to nil when not set")
	assert.Nil(t, cred.VaultEnv,
		"VaultEnv must default to nil when not set")
	assert.Nil(t, cred.VaultUsernameSecret,
		"VaultUsernameSecret must default to nil when not set")
	// ─────────────────────────────────────────────────────────────────────────
}

// ---------------------------------------------------------------------------
// Section: CredentialDB.ToConfig Vault Tests
// ---------------------------------------------------------------------------

// TestCredentialDB_ToConfig_Vault verifies that ToConfig correctly maps the
// vault source and VaultSecret / VaultEnv fields into a config.CredentialConfig.
//
// WILL FAIL TO COMPILE — CredentialDB.VaultSecret, .VaultEnv, and
// config.SourceVault, config.CredentialConfig.VaultSecret, .VaultEnv
// do not exist yet.
func TestCredentialDB_ToConfig_Vault(t *testing.T) {
	vaultSecret := "my-secret"
	vaultEnv := "staging"

	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	cred := &CredentialDB{
		Name:        "test-cred",
		Source:      "vault",
		VaultSecret: &vaultSecret,
		VaultEnv:    &vaultEnv,
	}
	cfg := cred.ToConfig()
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, config.SourceVault, cfg.Source,
		"ToConfig must set Source to SourceVault for vault-sourced credentials")
	assert.Equal(t, "my-secret", cfg.VaultSecret,
		"ToConfig must copy VaultSecret from CredentialDB to CredentialConfig")
	assert.Equal(t, "staging", cfg.VaultEnv,
		"ToConfig must copy VaultEnv from CredentialDB to CredentialConfig")
}

// TestCredentialDB_ToConfig_Vault_NilVaultEnv verifies that when VaultEnv is nil,
// ToConfig produces a CredentialConfig with an empty VaultEnv (zero value).
//
// WILL FAIL TO COMPILE — vault fields do not exist yet.
func TestCredentialDB_ToConfig_Vault_NilVaultEnv(t *testing.T) {
	vaultSecret := "my-secret"

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cred := &CredentialDB{
		Name:        "test-cred",
		Source:      "vault",
		VaultSecret: &vaultSecret,
		VaultEnv:    nil, // optional
	}
	cfg := cred.ToConfig()
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, config.SourceVault, cfg.Source)
	assert.Equal(t, "my-secret", cfg.VaultSecret)
	assert.Empty(t, cfg.VaultEnv,
		"nil VaultEnv in DB must produce empty string in CredentialConfig")
}

// ---------------------------------------------------------------------------
// Section: CredentialDB.ToUsernameConfig Vault Tests
// ---------------------------------------------------------------------------

// TestCredentialDB_ToUsernameConfig_Vault verifies that for vault-sourced
// dual-field credentials, ToUsernameConfig uses VaultUsernameSecret (not
// VaultSecret) as the secret name for the username lookup.
//
// WILL FAIL TO COMPILE — vault fields and config.SourceVault do not exist yet.
func TestCredentialDB_ToUsernameConfig_Vault(t *testing.T) {
	vaultSecret := "github-pat"
	vaultUsernameSecret := "github-user"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"

	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	cred := &CredentialDB{
		Name:                "github-creds",
		Source:              "vault",
		VaultSecret:         &vaultSecret,
		VaultUsernameSecret: &vaultUsernameSecret,
		UsernameVar:         &usernameVar,
		PasswordVar:         &passwordVar,
	}
	cfg := cred.ToUsernameConfig()
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, config.SourceVault, cfg.Source,
		"ToUsernameConfig must preserve vault source")
	// For vault dual-field, the username lookup uses VaultUsernameSecret
	// (the dedicated username secret), not VaultSecret (the password secret).
	assert.Equal(t, "github-user", cfg.VaultSecret,
		"ToUsernameConfig must use VaultUsernameSecret as the VaultSecret for username lookup")
}

// TestCredentialDB_ToPasswordConfig_Vault verifies that for vault-sourced
// dual-field credentials, ToPasswordConfig uses VaultSecret for the password
// lookup (the main/primary secret).
//
// WILL FAIL TO COMPILE — vault fields and config.SourceVault do not exist yet.
func TestCredentialDB_ToPasswordConfig_Vault(t *testing.T) {
	vaultSecret := "github-pat"
	vaultUsernameSecret := "github-user"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"

	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	cred := &CredentialDB{
		Name:                "github-creds",
		Source:              "vault",
		VaultSecret:         &vaultSecret,
		VaultUsernameSecret: &vaultUsernameSecret,
		UsernameVar:         &usernameVar,
		PasswordVar:         &passwordVar,
	}
	cfg := cred.ToPasswordConfig()
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, config.SourceVault, cfg.Source,
		"ToPasswordConfig must preserve vault source")
	assert.Equal(t, "github-pat", cfg.VaultSecret,
		"ToPasswordConfig must use VaultSecret (primary secret) for password lookup")
}

// ---------------------------------------------------------------------------
// Section: CredentialSpec Vault Field Tests
// ---------------------------------------------------------------------------

// TestCredentialSpec_VaultFields verifies that CredentialSpec has three vault
// fields: VaultSecret, VaultEnvironment (note: "Environment", not "Env"),
// and VaultUsernameSecret.
//
// WILL FAIL TO COMPILE — the vault fields on CredentialSpec do not exist yet.
func TestCredentialSpec_VaultFields(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	// CredentialSpec.VaultSecret, .VaultEnvironment, .VaultUsernameSecret
	// do not exist yet.
	spec := CredentialSpec{
		Source:              "vault",
		VaultSecret:         "my-secret",
		VaultEnvironment:    "production",
		VaultUsernameSecret: "my-user-secret",
	}
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, "my-secret", spec.VaultSecret,
		"VaultSecret must be stored and retrievable on CredentialSpec")
	assert.Equal(t, "production", spec.VaultEnvironment,
		"VaultEnvironment must be stored and retrievable on CredentialSpec")
	assert.Equal(t, "my-user-secret", spec.VaultUsernameSecret,
		"VaultUsernameSecret must be stored and retrievable on CredentialSpec")
}

// TestCredentialSpec_VaultFields_ZeroValues verifies that vault fields on
// CredentialSpec have empty string zero values (not pointer types).
//
// WILL FAIL TO COMPILE — the vault fields on CredentialSpec do not exist yet.
func TestCredentialSpec_VaultFields_ZeroValues(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	spec := CredentialSpec{
		Source: "env",
		EnvVar: "SOME_VAR",
	}
	assert.Empty(t, spec.VaultSecret, "VaultSecret must default to empty string")
	assert.Empty(t, spec.VaultEnvironment, "VaultEnvironment must default to empty string")
	assert.Empty(t, spec.VaultUsernameSecret, "VaultUsernameSecret must default to empty string")
	// ─────────────────────────────────────────────────────────────────────────
}

// ---------------------------------------------------------------------------
// Section: CredentialDB.ToYAML Vault Tests
// ---------------------------------------------------------------------------

// TestCredentialDB_ToYAML_VaultFull verifies that ToYAML correctly serializes
// vault fields from CredentialDB into CredentialSpec.
//
// WILL FAIL TO COMPILE — vault fields on CredentialDB and CredentialSpec
// do not exist yet.
func TestCredentialDB_ToYAML_VaultFull(t *testing.T) {
	vaultSecret := "my-github-pat"
	vaultEnv := "production"

	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	cred := &CredentialDB{
		Name:        "GITHUB_TOKEN",
		ScopeType:   CredentialScopeEcosystem,
		Source:      "vault",
		VaultSecret: &vaultSecret,
		VaultEnv:    &vaultEnv,
	}
	y := cred.ToYAML("testlab")
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, "devopsmaestro.io/v1", y.APIVersion)
	assert.Equal(t, "Credential", y.Kind)
	assert.Equal(t, "GITHUB_TOKEN", y.Metadata.Name)
	assert.Equal(t, "testlab", y.Metadata.Ecosystem)
	assert.Equal(t, "vault", y.Spec.Source)
	assert.Equal(t, "my-github-pat", y.Spec.VaultSecret,
		"ToYAML must serialize VaultSecret into Spec.VaultSecret")
	assert.Equal(t, "production", y.Spec.VaultEnvironment,
		"ToYAML must serialize VaultEnv into Spec.VaultEnvironment")
}

// TestCredentialDB_ToYAML_Vault_WithUsernameSecret verifies that ToYAML
// correctly serializes VaultUsernameSecret.
//
// WILL FAIL TO COMPILE — vault fields do not exist yet.
func TestCredentialDB_ToYAML_Vault_WithUsernameSecret(t *testing.T) {
	vaultSecret := "github-pat"
	vaultUsernameSecret := "github-username"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"

	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
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
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, "vault", y.Spec.Source)
	assert.Equal(t, "github-pat", y.Spec.VaultSecret)
	assert.Equal(t, "github-username", y.Spec.VaultUsernameSecret,
		"ToYAML must serialize VaultUsernameSecret into Spec.VaultUsernameSecret")
	assert.Equal(t, "GITHUB_USERNAME", y.Spec.UsernameVar)
	assert.Equal(t, "GITHUB_PAT", y.Spec.PasswordVar)
}

// TestCredentialDB_ToYAML_Vault_NilVaultEnv verifies that nil VaultEnv
// produces an empty VaultEnvironment in the YAML spec.
//
// WILL FAIL TO COMPILE — vault fields do not exist yet.
func TestCredentialDB_ToYAML_Vault_NilVaultEnv(t *testing.T) {
	vaultSecret := "my-secret"

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cred := &CredentialDB{
		Name:        "MY_SECRET",
		ScopeType:   CredentialScopeApp,
		Source:      "vault",
		VaultSecret: &vaultSecret,
		VaultEnv:    nil,
	}
	y := cred.ToYAML("test-app")
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, "vault", y.Spec.Source)
	assert.Equal(t, "my-secret", y.Spec.VaultSecret)
	assert.Empty(t, y.Spec.VaultEnvironment,
		"nil VaultEnv must produce empty VaultEnvironment in YAML spec")
}

// ---------------------------------------------------------------------------
// Section: CredentialDB.FromYAML Vault Tests
// ---------------------------------------------------------------------------

// TestCredentialDB_FromYAML_VaultFull verifies that FromYAML correctly
// deserializes vault fields from CredentialSpec into CredentialDB.
//
// WILL FAIL TO COMPILE — vault fields on CredentialDB and CredentialSpec
// do not exist yet.
func TestCredentialDB_FromYAML_VaultFull(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
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
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, "GITHUB_TOKEN", cred.Name)
	assert.Equal(t, "vault", cred.Source)
	require.NotNil(t, cred.VaultSecret,
		"FromYAML must populate VaultSecret from Spec.VaultSecret")
	assert.Equal(t, "my-github-pat", *cred.VaultSecret)
	require.NotNil(t, cred.VaultEnv,
		"FromYAML must populate VaultEnv from Spec.VaultEnvironment")
	assert.Equal(t, "production", *cred.VaultEnv)
	assert.Nil(t, cred.VaultUsernameSecret,
		"VaultUsernameSecret must be nil when not in spec")
}

// TestCredentialDB_FromYAML_Vault_WithUsernameSecret verifies that FromYAML
// correctly deserializes VaultUsernameSecret.
//
// WILL FAIL TO COMPILE — vault fields do not exist yet.
func TestCredentialDB_FromYAML_Vault_WithUsernameSecret(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
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
	// ─────────────────────────────────────────────────────────────────────────

	require.NotNil(t, cred.VaultSecret)
	assert.Equal(t, "github-pat", *cred.VaultSecret)
	require.NotNil(t, cred.VaultEnv)
	assert.Equal(t, "staging", *cred.VaultEnv)
	require.NotNil(t, cred.VaultUsernameSecret,
		"FromYAML must populate VaultUsernameSecret from Spec.VaultUsernameSecret")
	assert.Equal(t, "github-username", *cred.VaultUsernameSecret)
	require.NotNil(t, cred.UsernameVar)
	assert.Equal(t, "GITHUB_USERNAME", *cred.UsernameVar)
	require.NotNil(t, cred.PasswordVar)
	assert.Equal(t, "GITHUB_PAT", *cred.PasswordVar)
}

// TestCredentialDB_FromYAML_Vault_EmptyVaultEnvironment verifies that when
// VaultEnvironment is absent from the spec, VaultEnv is nil in the DB.
//
// WILL FAIL TO COMPILE — vault fields do not exist yet.
func TestCredentialDB_FromYAML_Vault_EmptyVaultEnvironment(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	y := CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata:   CredentialMetadata{Name: "MY_SECRET", App: "test-app"},
		Spec: CredentialSpec{
			Source:      "vault",
			VaultSecret: "my-secret",
			// VaultEnvironment intentionally omitted
		},
	}
	cred := &CredentialDB{}
	cred.FromYAML(y)
	// ─────────────────────────────────────────────────────────────────────────

	require.NotNil(t, cred.VaultSecret)
	assert.Equal(t, "my-secret", *cred.VaultSecret)
	assert.Nil(t, cred.VaultEnv,
		"absent VaultEnvironment in spec must produce nil VaultEnv in DB")
}

// ---------------------------------------------------------------------------
// Section: CredentialDB Vault RoundTrip Tests
// ---------------------------------------------------------------------------

// TestCredentialDB_RoundTrip_VaultFull verifies that vault credential data
// survives a complete ToYAML → FromYAML round trip.
//
// WILL FAIL TO COMPILE — vault fields do not exist yet.
func TestCredentialDB_RoundTrip_VaultFull(t *testing.T) {
	vaultSecret := "my-vault-secret"
	vaultEnv := "production"

	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
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
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, original.Name, restored.Name)
	assert.Equal(t, original.Source, restored.Source)
	require.NotNil(t, restored.VaultSecret,
		"VaultSecret must survive ToYAML → FromYAML round trip")
	assert.Equal(t, *original.VaultSecret, *restored.VaultSecret)
	require.NotNil(t, restored.VaultEnv,
		"VaultEnv must survive ToYAML → FromYAML round trip")
	assert.Equal(t, *original.VaultEnv, *restored.VaultEnv)
}

// TestCredentialDB_RoundTrip_Vault_WithUsernameSecret verifies a round trip
// for vault dual-field credentials (username + password secrets).
//
// WILL FAIL TO COMPILE — vault fields do not exist yet.
func TestCredentialDB_RoundTrip_Vault_WithUsernameSecret(t *testing.T) {
	vaultSecret := "github-pat-secret"
	vaultEnv := "staging"
	vaultUsernameSecret := "github-username-secret"
	usernameVar := "GH_USER"
	passwordVar := "GH_PAT"

	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
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
	// ─────────────────────────────────────────────────────────────────────────

	require.NotNil(t, restored.VaultSecret)
	assert.Equal(t, *original.VaultSecret, *restored.VaultSecret)
	require.NotNil(t, restored.VaultEnv)
	assert.Equal(t, *original.VaultEnv, *restored.VaultEnv)
	require.NotNil(t, restored.VaultUsernameSecret,
		"VaultUsernameSecret must survive round trip")
	assert.Equal(t, *original.VaultUsernameSecret, *restored.VaultUsernameSecret)
	require.NotNil(t, restored.UsernameVar)
	assert.Equal(t, *original.UsernameVar, *restored.UsernameVar)
	require.NotNil(t, restored.PasswordVar)
	assert.Equal(t, *original.PasswordVar, *restored.PasswordVar)
}

// ---------------------------------------------------------------------------
// Section: ValidateCredentialYAML Vault Tests
// ---------------------------------------------------------------------------

// TestValidateCredentialYAML_Vault validates that the "vault" source is
// accepted and that vault-specific validation rules are enforced.
//
// WILL FAIL AT RUNTIME (and possibly compile) — ValidateCredentialYAML
// currently rejects "vault" as an invalid source. After implementation it
// must accept it, and vault fields must be defined on CredentialSpec.
func TestValidateCredentialYAML_Vault(t *testing.T) {
	tests := []struct {
		name    string
		spec    CredentialSpec
		wantErr bool
		errMsg  string
	}{
		{
			// ── COMPILE ERRORS EXPECTED — vault fields don't exist yet ────────
			name: "valid vault credential",
			spec: CredentialSpec{
				Source:      "vault",
				VaultSecret: "my-github-pat",
			},
			wantErr: false,
		},
		{
			name: "valid vault credential with environment",
			spec: CredentialSpec{
				Source:           "vault",
				VaultSecret:      "my-github-pat",
				VaultEnvironment: "production",
			},
			wantErr: false,
		},
		{
			name: "valid vault dual-field credential",
			spec: CredentialSpec{
				Source:              "vault",
				VaultSecret:         "github-pat",
				VaultUsernameSecret: "github-username",
				UsernameVar:         "GITHUB_USERNAME",
				PasswordVar:         "GITHUB_PAT",
			},
			wantErr: false,
		},
		{
			name: "vault missing vaultSecret",
			spec: CredentialSpec{
				Source:           "vault",
				VaultEnvironment: "production",
				// VaultSecret intentionally omitted
			},
			wantErr: true,
			errMsg:  "vault",
		},
		// ─────────────────────────────────────────────────────────────────────
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			y := CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   CredentialMetadata{Name: "TEST_CRED", Ecosystem: "testlab"},
				Spec:       tt.spec,
			}
			err := ValidateCredentialYAML(y)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateCredentialYAML_Vault_SourceAccepted verifies that the string
// "vault" is accepted as a valid source value (currently rejected).
//
// WILL FAIL AT RUNTIME — ValidateCredentialYAML currently only accepts
// "keychain" and "env". After implementation it must also accept "vault".
func TestValidateCredentialYAML_Vault_SourceAccepted(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED — vault fields don't exist yet ───────────────
	y := CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata:   CredentialMetadata{Name: "TEST_CRED", Ecosystem: "testlab"},
		Spec: CredentialSpec{
			Source:      "vault",
			VaultSecret: "my-secret",
		},
	}
	// ─────────────────────────────────────────────────────────────────────────

	err := ValidateCredentialYAML(y)

	assert.NoError(t, err,
		"source='vault' must be accepted by ValidateCredentialYAML after implementation")
}

// TestValidateCredentialYAML_Vault_RequiresVaultSecret verifies that vault
// source requires the vaultSecret field to be set.
//
// WILL FAIL AT RUNTIME (the "vault" source itself will be rejected first)
// and WILL FAIL TO COMPILE — vault fields don't exist yet.
func TestValidateCredentialYAML_Vault_RequiresVaultSecret(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED — vault fields don't exist yet ───────────────
	y := CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata:   CredentialMetadata{Name: "TEST_CRED", Ecosystem: "testlab"},
		Spec: CredentialSpec{
			Source: "vault",
			// VaultSecret intentionally absent
		},
	}
	// ─────────────────────────────────────────────────────────────────────────

	err := ValidateCredentialYAML(y)

	assert.Error(t, err,
		"vault source without vaultSecret must be rejected by ValidateCredentialYAML")
	assert.Contains(t, err.Error(), "vault",
		"error must mention vault or vaultSecret")
}

// TestValidateCredentialYAML_Vault_UsernameSecretRequiresUsernameVar verifies
// that when VaultUsernameSecret is set, UsernameVar must also be set
// (cross-field validation).
//
// WILL FAIL AT RUNTIME — this cross-validation does not exist yet.
// Also WILL FAIL TO COMPILE — vault fields don't exist yet.
func TestValidateCredentialYAML_Vault_UsernameSecretRequiresUsernameVar(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED — vault fields don't exist yet ───────────────
	y := CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata:   CredentialMetadata{Name: "TEST_CRED", Ecosystem: "testlab"},
		Spec: CredentialSpec{
			Source:              "vault",
			VaultSecret:         "my-pat",
			VaultUsernameSecret: "my-username-secret",
			// UsernameVar intentionally absent — must trigger validation error
		},
	}
	// ─────────────────────────────────────────────────────────────────────────

	err := ValidateCredentialYAML(y)

	assert.Error(t, err,
		"vaultUsernameSecret requires usernameVar to be set")
	assert.Contains(t, err.Error(), "usernameVar",
		"error must mention usernameVar when vaultUsernameSecret is set without it")
}

// =============================================================================
// TDD Phase 2 (RED): VaultFields Tests (v0.41.0)
// =============================================================================
// New field on CredentialDB:
//
//	VaultFields *string `db:"vault_fields" json:"vault_fields,omitempty"`
//	// Stores JSON: {"ENV_VAR_NAME": "field_name", ...}
//
// New methods on CredentialDB:
//
//	func (c *CredentialDB) HasVaultFields() bool
//	func (c *CredentialDB) GetVaultFieldsMap() (map[string]string, error)
//	func (c *CredentialDB) ToMapEntries() []config.CredentialConfig
//
// New field on CredentialSpec:
//
//	VaultFields map[string]string `yaml:"vaultFields,omitempty"`
//
// ALL tests in this section WILL FAIL TO COMPILE until the above types/methods
// are added to models/credential.go.
// =============================================================================

// ---------------------------------------------------------------------------
// Section: CredentialDB.VaultFields Field Tests
// ---------------------------------------------------------------------------

// TestCredentialDB_VaultFields_FieldExists verifies that CredentialDB has a
// VaultFields *string field with db tag "vault_fields".
//
// WILL FAIL TO COMPILE — CredentialDB.VaultFields does not exist yet.
func TestCredentialDB_VaultFields_FieldExists(t *testing.T) {
	raw := `{"GITHUB_TOKEN":"token","GITHUB_USER":"username"}`

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cred := &CredentialDB{
		Name:        "github-creds",
		Source:      "vault",
		VaultSecret: strPtr("github/creds"),
		VaultFields: strPtr(raw),
	}
	// ─────────────────────────────────────────────────────────────────────────

	require.NotNil(t, cred.VaultFields, "VaultFields must not be nil when set")
	assert.Equal(t, raw, *cred.VaultFields, "VaultFields must hold the JSON blob")
}

// TestCredentialDB_VaultFields_NilByDefault verifies that VaultFields is nil
// for credentials that do not use vault fields.
//
// WILL FAIL TO COMPILE — CredentialDB.VaultFields does not exist yet.
func TestCredentialDB_VaultFields_NilByDefault(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cred := &CredentialDB{
		Name:        "env-cred",
		Source:      "env",
		EnvVar:      strPtr("MY_TOKEN"),
		VaultFields: nil,
	}
	// ─────────────────────────────────────────────────────────────────────────

	assert.Nil(t, cred.VaultFields, "VaultFields must be nil when not set")
}

// ---------------------------------------------------------------------------
// Section: CredentialDB.HasVaultFields Tests
// ---------------------------------------------------------------------------

// TestCredentialDB_HasVaultFields_True verifies that HasVaultFields returns
// true when VaultFields contains a non-empty JSON object.
//
// WILL FAIL TO COMPILE — HasVaultFields does not exist yet.
func TestCredentialDB_HasVaultFields_True(t *testing.T) {
	raw := `{"API_TOKEN":"token"}`

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cred := &CredentialDB{
		VaultFields: strPtr(raw),
	}
	result := cred.HasVaultFields()
	// ─────────────────────────────────────────────────────────────────────────

	assert.True(t, result, "HasVaultFields must return true when VaultFields is set")
}

// TestCredentialDB_HasVaultFields_FalseWhenNil verifies that HasVaultFields
// returns false when VaultFields is nil.
//
// WILL FAIL TO COMPILE — HasVaultFields does not exist yet.
func TestCredentialDB_HasVaultFields_FalseWhenNil(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cred := &CredentialDB{VaultFields: nil}
	result := cred.HasVaultFields()
	// ─────────────────────────────────────────────────────────────────────────

	assert.False(t, result, "HasVaultFields must return false when VaultFields is nil")
}

// TestCredentialDB_HasVaultFields_FalseWhenEmptyJSON verifies that HasVaultFields
// returns false when VaultFields is an empty JSON object "{}".
//
// WILL FAIL TO COMPILE — HasVaultFields does not exist yet.
func TestCredentialDB_HasVaultFields_FalseWhenEmptyJSON(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cred := &CredentialDB{VaultFields: strPtr("{}")}
	result := cred.HasVaultFields()
	// ─────────────────────────────────────────────────────────────────────────

	assert.False(t, result, "HasVaultFields must return false when VaultFields is empty JSON")
}

// ---------------------------------------------------------------------------
// Section: CredentialDB.GetVaultFieldsMap Tests
// ---------------------------------------------------------------------------

// TestCredentialDB_GetVaultFieldsMap_ParsesJSON verifies that GetVaultFieldsMap
// correctly deserializes the vault_fields JSON blob into a Go map.
//
// WILL FAIL TO COMPILE — GetVaultFieldsMap does not exist yet.
func TestCredentialDB_GetVaultFieldsMap_ParsesJSON(t *testing.T) {
	raw := `{"GITHUB_TOKEN":"token","GITHUB_USER":"username"}`

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cred := &CredentialDB{
		VaultFields: strPtr(raw),
	}
	result, err := cred.GetVaultFieldsMap()
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, err, "GetVaultFieldsMap must not return error for valid JSON")
	require.Len(t, result, 2, "map must contain 2 entries")
	assert.Equal(t, "token", result["GITHUB_TOKEN"])
	assert.Equal(t, "username", result["GITHUB_USER"])
}

// TestCredentialDB_GetVaultFieldsMap_ReturnsEmptyMapWhenNil verifies that
// GetVaultFieldsMap returns an empty (non-nil) map when VaultFields is nil.
//
// WILL FAIL TO COMPILE — GetVaultFieldsMap does not exist yet.
func TestCredentialDB_GetVaultFieldsMap_ReturnsEmptyMapWhenNil(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cred := &CredentialDB{VaultFields: nil}
	result, err := cred.GetVaultFieldsMap()
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, err, "GetVaultFieldsMap must not error when VaultFields is nil")
	assert.Nil(t, result, "GetVaultFieldsMap must return nil when VaultFields is nil")
	assert.Empty(t, result, "returned map must be empty when VaultFields is nil")
}

// TestCredentialDB_GetVaultFieldsMap_ErrorOnInvalidJSON verifies that
// GetVaultFieldsMap returns an error when VaultFields contains invalid JSON.
//
// WILL FAIL TO COMPILE — GetVaultFieldsMap does not exist yet.
func TestCredentialDB_GetVaultFieldsMap_ErrorOnInvalidJSON(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cred := &CredentialDB{VaultFields: strPtr("not-valid-json")}
	_, err := cred.GetVaultFieldsMap()
	// ─────────────────────────────────────────────────────────────────────────

	assert.Error(t, err, "GetVaultFieldsMap must return error for invalid JSON in VaultFields")
}

// ---------------------------------------------------------------------------
// Section: CredentialDB.ToMapEntries Tests
// ---------------------------------------------------------------------------

// TestCredentialDB_ToMapEntries_VaultFields_FansOut verifies that ToMapEntries
// expands a credential with VaultFields into one CredentialConfig per field
// (fan-out), where each config has VaultField set to the specific field name
// and the env var name as the key.
//
// WILL FAIL TO COMPILE — ToMapEntries does not exist yet.
func TestCredentialDB_ToMapEntries_VaultFields_FansOut(t *testing.T) {
	raw := `{"GITHUB_TOKEN":"token","GITHUB_USER":"username"}`

	cred := &CredentialDB{
		Name:        "github-creds",
		Source:      "vault",
		VaultSecret: strPtr("github/creds"),
		VaultEnv:    strPtr("production"),
		VaultFields: strPtr(raw),
	}
	entries := cred.ToMapEntries()

	assert.Len(t, entries, 2, "one CredentialConfig per vault field")

	assert.Contains(t, entries, "GITHUB_TOKEN",
		"ToMapEntries must produce an entry for GITHUB_TOKEN")
	assert.Contains(t, entries, "GITHUB_USER",
		"ToMapEntries must produce an entry for GITHUB_USER")

	// Verify VaultField is set correctly on each entry
	assert.Equal(t, "token", entries["GITHUB_TOKEN"].VaultField)
	assert.Equal(t, "username", entries["GITHUB_USER"].VaultField)
}

// TestCredentialDB_ToMapEntries_NoVaultFields_SingleEntry verifies that a
// credential without VaultFields produces exactly one CredentialConfig (the
// existing non-fan-out behaviour is preserved).
//
// WILL FAIL TO COMPILE — ToMapEntries does not exist yet.
func TestCredentialDB_ToMapEntries_NoVaultFields_SingleEntry(t *testing.T) {
	cred := &CredentialDB{
		Name:        "github-token",
		Source:      "vault",
		VaultSecret: strPtr("github/token"),
	}
	entries := cred.ToMapEntries()

	assert.Len(t, entries, 1, "non-fan-out credential must produce exactly 1 entry")
	assert.Contains(t, entries, "github-token", "key should be credential name for simple credentials")
}

// TestCredentialDB_ToMapEntries_EnvSource verifies that an env-sourced
// credential also produces a single CredentialConfig entry.
//
// WILL FAIL TO COMPILE — ToMapEntries does not exist yet.
func TestCredentialDB_ToMapEntries_EnvSource(t *testing.T) {
	cred := &CredentialDB{
		Name:   "my-token",
		Source: "env",
		EnvVar: strPtr("MY_API_TOKEN"),
	}
	entries := cred.ToMapEntries()

	assert.Len(t, entries, 1, "env credential must produce exactly 1 entry")
	assert.Contains(t, entries, "my-token", "key should be credential name for simple credentials")
}

// ---------------------------------------------------------------------------
// Section: CredentialSpec.VaultFields Field Tests
// ---------------------------------------------------------------------------

// TestCredentialSpec_VaultFields_FieldExists verifies that CredentialSpec has
// a VaultFields map[string]string field with yaml tag "vaultFields".
//
// WILL FAIL TO COMPILE — CredentialSpec.VaultFields does not exist yet.
func TestCredentialSpec_VaultFields_FieldExists(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	spec := CredentialSpec{
		Source:      "vault",
		VaultSecret: "github/creds",
		VaultFields: map[string]string{
			"GITHUB_TOKEN": "token",
			"GITHUB_USER":  "username",
		},
	}
	// ─────────────────────────────────────────────────────────────────────────

	assert.Len(t, spec.VaultFields, 2)
	assert.Equal(t, "token", spec.VaultFields["GITHUB_TOKEN"])
	assert.Equal(t, "username", spec.VaultFields["GITHUB_USER"])
}

// TestCredentialSpec_VaultFields_NilByDefault verifies that VaultFields is nil
// when not set.
//
// WILL FAIL TO COMPILE — CredentialSpec.VaultFields does not exist yet.
func TestCredentialSpec_VaultFields_NilByDefault(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	spec := CredentialSpec{
		Source: "env",
		EnvVar: "MY_TOKEN",
	}
	// ─────────────────────────────────────────────────────────────────────────

	// VaultFields should be nil (zero value for a map).
	assert.Nil(t, spec.VaultFields, "VaultFields must be nil when not set")
}

// ---------------------------------------------------------------------------
// Section: CredentialYAML VaultFields Round-Trip Tests
// ---------------------------------------------------------------------------

// TestCredentialYAML_VaultFields_RoundTrip verifies that a CredentialYAML with
// vaultFields can be marshalled to YAML and back with values intact.
//
// WILL FAIL TO COMPILE — CredentialSpec.VaultFields does not exist yet.
func TestCredentialYAML_VaultFields_RoundTrip(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
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
	// ─────────────────────────────────────────────────────────────────────────

	// Marshal to YAML.
	data, err := yamlv3.Marshal(original)
	require.NoError(t, err, "yaml.Marshal must not error")

	// Unmarshal back.
	var restored CredentialYAML
	err = yamlv3.Unmarshal(data, &restored)
	require.NoError(t, err, "yaml.Unmarshal must not error")

	assert.Equal(t, original.Spec.VaultFields, restored.Spec.VaultFields,
		"VaultFields must survive a marshal/unmarshal round-trip")
}

// ---------------------------------------------------------------------------
// Section: ValidateCredentialYAML VaultFields Validation Tests
// ---------------------------------------------------------------------------

// TestValidateCredentialYAML_VaultFields_RequiresVaultSource verifies that
// vaultFields is only valid with source=vault.
//
// WILL FAIL TO COMPILE and FAIL AT RUNTIME — VaultFields does not exist yet.
func TestValidateCredentialYAML_VaultFields_RequiresVaultSource(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	y := CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata:   CredentialMetadata{Name: "MY_CRED", Ecosystem: "testlab"},
		Spec: CredentialSpec{
			Source: "env",
			EnvVar: "MY_VAR",
			VaultFields: map[string]string{
				"SOME_VAR": "some-field",
			},
		},
	}
	// ─────────────────────────────────────────────────────────────────────────

	err := ValidateCredentialYAML(y)
	assert.Error(t, err,
		"vaultFields with env source must be rejected by ValidateCredentialYAML")
}

// TestValidateCredentialYAML_VaultFields_MutuallyExclusiveWithUsernameVar
// verifies that vaultFields cannot be combined with usernameVar/passwordVar
// (they represent different resolution modes).
//
// WILL FAIL TO COMPILE and FAIL AT RUNTIME.
func TestValidateCredentialYAML_VaultFields_MutuallyExclusiveWithUsernameVar(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	y := CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata:   CredentialMetadata{Name: "MY_CRED", Ecosystem: "testlab"},
		Spec: CredentialSpec{
			Source:      "vault",
			VaultSecret: "my-org/creds",
			UsernameVar: "MY_USER",
			VaultFields: map[string]string{
				"TOKEN": "token",
			},
		},
	}
	// ─────────────────────────────────────────────────────────────────────────

	err := ValidateCredentialYAML(y)
	assert.Error(t, err,
		"vaultFields and usernameVar are mutually exclusive")
}

// TestValidateCredentialYAML_VaultFields_MaxFiftyFields verifies that
// ValidateCredentialYAML rejects a credential with more than 50 vault fields.
//
// WILL FAIL TO COMPILE and FAIL AT RUNTIME.
func TestValidateCredentialYAML_VaultFields_MaxFiftyFields(t *testing.T) {
	// Build a map with 51 entries.
	fields := make(map[string]string, 51)
	for i := 0; i < 51; i++ {
		key := fmt.Sprintf("ENV_VAR_%03d", i)
		fields[key] = fmt.Sprintf("field_%03d", i)
	}

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	y := CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata:   CredentialMetadata{Name: "MY_CRED", Ecosystem: "testlab"},
		Spec: CredentialSpec{
			Source:      "vault",
			VaultSecret: "my-org/creds",
			VaultFields: fields,
		},
	}
	// ─────────────────────────────────────────────────────────────────────────

	err := ValidateCredentialYAML(y)
	assert.Error(t, err, "vaultFields with more than 50 entries must be rejected")
	assert.Contains(t, err.Error(), "50",
		"error message must mention the 50-field limit")
}

// TestValidateCredentialYAML_VaultFields_ExactlyFiftyFields verifies that
// a credential with exactly 50 vault fields is valid.
//
// WILL FAIL TO COMPILE and FAIL AT RUNTIME.
func TestValidateCredentialYAML_VaultFields_ExactlyFiftyFields(t *testing.T) {
	fields := make(map[string]string, 50)
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("ENV_VAR_%03d", i)
		fields[key] = fmt.Sprintf("field_%03d", i)
	}

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	y := CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata:   CredentialMetadata{Name: "MY_CRED", Ecosystem: "testlab"},
		Spec: CredentialSpec{
			Source:      "vault",
			VaultSecret: "my-org/creds",
			VaultFields: fields,
		},
	}
	// ─────────────────────────────────────────────────────────────────────────

	err := ValidateCredentialYAML(y)
	assert.NoError(t, err, "exactly 50 vault fields must be accepted")
}

// =============================================================================
// TDD Phase 2 (RED): Bug #157 — Missing VaultField in dual-field config
// =============================================================================
// Bug: ToUsernameConfig() and ToPasswordConfig() produce CredentialConfig
// objects WITHOUT VaultField set, causing vault resolution to call
// backend.Get() (whole-secret fetch) instead of backend.GetField()
// (field-level fetch). This makes dual-field credentials non-functional.
//
// Both tests WILL FAIL until models/credential.go is fixed to set VaultField
// in ToUsernameConfig() and ToPasswordConfig().
// =============================================================================

// TestCredentialDB_ToUsernameConfig_SetsVaultField verifies that
// ToUsernameConfig() sets VaultField to the UsernameVar value, enabling
// field-level vault access (GetField) rather than whole-secret access (Get).
//
// BUG #157 — WILL FAIL until ToUsernameConfig sets VaultField.
func TestCredentialDB_ToUsernameConfig_SetsVaultField(t *testing.T) {
	vaultSecret := "github-creds"
	vaultEnv := "production"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"

	cred := &CredentialDB{
		Name:        "github-creds",
		Source:      "vault",
		VaultSecret: &vaultSecret,
		VaultEnv:    &vaultEnv,
		UsernameVar: &usernameVar,
		PasswordVar: &passwordVar,
	}

	cfg := cred.ToUsernameConfig()

	assert.Equal(t, config.SourceVault, cfg.Source,
		"ToUsernameConfig must set Source to SourceVault")
	assert.Equal(t, "github-creds", cfg.VaultSecret,
		"ToUsernameConfig must copy VaultSecret")
	assert.Equal(t, "production", cfg.VaultEnv,
		"ToUsernameConfig must copy VaultEnv")
	// BUG: VaultField is "" — should be "GITHUB_USERNAME" to enable GetField()
	assert.Equal(t, "GITHUB_USERNAME", cfg.VaultField,
		"ToUsernameConfig MUST set VaultField to the UsernameVar name so "+
			"vault resolution calls GetField(secret, env, field) not Get(secret, env)")
}

// TestCredentialDB_ToPasswordConfig_SetsVaultField verifies that
// ToPasswordConfig() sets VaultField to the PasswordVar value, enabling
// field-level vault access (GetField) rather than whole-secret access (Get).
//
// BUG #157 — WILL FAIL until ToPasswordConfig sets VaultField.
func TestCredentialDB_ToPasswordConfig_SetsVaultField(t *testing.T) {
	vaultSecret := "github-creds"
	vaultEnv := "production"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"

	cred := &CredentialDB{
		Name:        "github-creds",
		Source:      "vault",
		VaultSecret: &vaultSecret,
		VaultEnv:    &vaultEnv,
		UsernameVar: &usernameVar,
		PasswordVar: &passwordVar,
	}

	cfg := cred.ToPasswordConfig()

	assert.Equal(t, config.SourceVault, cfg.Source,
		"ToPasswordConfig must set Source to SourceVault")
	assert.Equal(t, "github-creds", cfg.VaultSecret,
		"ToPasswordConfig must copy VaultSecret")
	assert.Equal(t, "production", cfg.VaultEnv,
		"ToPasswordConfig must copy VaultEnv")
	// BUG: VaultField is "" — should be "GITHUB_PAT" to enable GetField()
	assert.Equal(t, "GITHUB_PAT", cfg.VaultField,
		"ToPasswordConfig MUST set VaultField to the PasswordVar name so "+
			"vault resolution calls GetField(secret, env, field) not Get(secret, env)")
}

// TestCredentialDB_ToUsernameConfig_SetsVaultField_WithVaultUsernameSecret
// verifies that ToUsernameConfig() sets VaultField even when VaultUsernameSecret
// is used (i.e., username comes from a separate vault secret).
//
// BUG #157 — WILL FAIL until ToUsernameConfig sets VaultField.
func TestCredentialDB_ToUsernameConfig_SetsVaultField_WithVaultUsernameSecret(t *testing.T) {
	vaultSecret := "github-pat-secret"
	vaultUsernameSecret := "github-user-secret"
	usernameVar := "GITHUB_USERNAME"

	cred := &CredentialDB{
		Name:                "github-creds",
		Source:              "vault",
		VaultSecret:         &vaultSecret,
		VaultUsernameSecret: &vaultUsernameSecret,
		UsernameVar:         &usernameVar,
	}

	cfg := cred.ToUsernameConfig()

	assert.Equal(t, config.SourceVault, cfg.Source)
	// When VaultUsernameSecret is set, ToUsernameConfig uses it as the vault secret
	assert.Equal(t, "github-user-secret", cfg.VaultSecret,
		"ToUsernameConfig must use VaultUsernameSecret when available")
	// VaultField must be set to the username var name
	assert.Equal(t, "GITHUB_USERNAME", cfg.VaultField,
		"ToUsernameConfig MUST set VaultField to the UsernameVar name even when "+
			"VaultUsernameSecret is used")
}

// TestCredentialDB_ToMapEntries_DualField_SetsVaultField verifies the end-to-end
// path: when a dual-field credential (IsDualField() == true) is converted via
// ToMapEntries(), both resulting CredentialConfigs must have VaultField set.
//
// Without VaultField, ResolveCredentialWithBackend calls backend.Get() (fetches
// entire secret) instead of backend.GetField() (fetches specific field).
//
// BUG #157 — WILL FAIL until ToUsernameConfig/ToPasswordConfig set VaultField.
func TestCredentialDB_ToMapEntries_DualField_SetsVaultField(t *testing.T) {
	vaultSecret := "github-creds"
	vaultEnv := "staging"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"

	cred := &CredentialDB{
		Name:        "github-creds",
		Source:      "vault",
		VaultSecret: &vaultSecret,
		VaultEnv:    &vaultEnv,
		UsernameVar: &usernameVar,
		PasswordVar: &passwordVar,
	}

	entries := cred.ToMapEntries()

	require.Len(t, entries, 2,
		"dual-field credential must produce 2 map entries")

	userCfg, ok := entries["GITHUB_USERNAME"]
	require.True(t, ok, "map must contain GITHUB_USERNAME key")
	assert.Equal(t, config.SourceVault, userCfg.Source)
	assert.Equal(t, "github-creds", userCfg.VaultSecret)
	assert.Equal(t, "staging", userCfg.VaultEnv)
	// BUG: VaultField is currently "" — should be "GITHUB_USERNAME"
	assert.Equal(t, "GITHUB_USERNAME", userCfg.VaultField,
		"GITHUB_USERNAME entry MUST have VaultField='GITHUB_USERNAME' so vault "+
			"resolution calls GetField not Get")

	passCfg, ok := entries["GITHUB_PAT"]
	require.True(t, ok, "map must contain GITHUB_PAT key")
	assert.Equal(t, config.SourceVault, passCfg.Source)
	assert.Equal(t, "github-creds", passCfg.VaultSecret)
	assert.Equal(t, "staging", passCfg.VaultEnv)
	// BUG: VaultField is currently "" — should be "GITHUB_PAT"
	assert.Equal(t, "GITHUB_PAT", passCfg.VaultField,
		"GITHUB_PAT entry MUST have VaultField='GITHUB_PAT' so vault "+
			"resolution calls GetField not Get")
}
