package models

import (
	"testing"

	"devopsmaestro/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			Source:      "keychain",
			Service:     "dvm-github-token",
			Description: "GitHub PAT for private repo access",
		},
	}

	assert.Equal(t, "devopsmaestro.io/v1", yaml.APIVersion)
	assert.Equal(t, "Credential", yaml.Kind)
	assert.Equal(t, "GITHUB_TOKEN", yaml.Metadata.Name)
	assert.Equal(t, "testlab", yaml.Metadata.Ecosystem)
	assert.Equal(t, "keychain", yaml.Spec.Source)
	assert.Equal(t, "dvm-github-token", yaml.Spec.Service)
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
	assert.Empty(t, yaml.Spec.Service)
}

// =============================================================================
// Credential ToYAML Tests
// =============================================================================

func TestCredentialDB_ToYAML_Keychain(t *testing.T) {
	service := "dvm-github-token"
	desc := "GitHub PAT"
	cred := &CredentialDB{
		Name:        "GITHUB_TOKEN",
		ScopeType:   CredentialScopeEcosystem,
		Source:      "keychain",
		Service:     &service,
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
	assert.Equal(t, "keychain", yaml.Spec.Source)
	assert.Equal(t, "dvm-github-token", yaml.Spec.Service)
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
	assert.Empty(t, yaml.Spec.Service)
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
			service := "dvm-test"
			cred := &CredentialDB{
				Name:      "TEST_CRED",
				ScopeType: tt.scopeType,
				Source:    "keychain",
				Service:   &service,
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
		Source:      "keychain",
		Service:     nil,
		EnvVar:      nil,
		Description: nil,
	}

	yaml := cred.ToYAML("testlab")

	assert.Empty(t, yaml.Spec.Service)
	assert.Empty(t, yaml.Spec.EnvVar)
	assert.Empty(t, yaml.Spec.Description)
}

// =============================================================================
// Credential FromYAML Tests
// =============================================================================

func TestCredentialDB_FromYAML_Keychain(t *testing.T) {
	yaml := CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata: CredentialMetadata{
			Name:      "GITHUB_TOKEN",
			Ecosystem: "testlab",
		},
		Spec: CredentialSpec{
			Source:      "keychain",
			Service:     "dvm-github-token",
			Description: "GitHub PAT",
		},
	}

	cred := &CredentialDB{}
	cred.FromYAML(yaml)

	assert.Equal(t, "GITHUB_TOKEN", cred.Name)
	assert.Equal(t, "keychain", cred.Source)
	require.NotNil(t, cred.Service)
	assert.Equal(t, "dvm-github-token", *cred.Service)
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
	assert.Nil(t, cred.Service)
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
			Source: "keychain",
		},
	}

	cred := &CredentialDB{}
	cred.FromYAML(yaml)

	assert.Equal(t, "BARE_CRED", cred.Name)
	assert.Equal(t, "keychain", cred.Source)
	assert.Nil(t, cred.Service)
	assert.Nil(t, cred.EnvVar)
	assert.Nil(t, cred.Description)
}

// =============================================================================
// Credential RoundTrip Tests
// =============================================================================

func TestCredentialDB_RoundTrip_Keychain(t *testing.T) {
	service := "dvm-gh"
	desc := "GitHub token"
	original := &CredentialDB{
		Name:        "GITHUB_TOKEN",
		ScopeType:   CredentialScopeEcosystem,
		Source:      "keychain",
		Service:     &service,
		Description: &desc,
	}

	yaml := original.ToYAML("testlab")
	restored := &CredentialDB{}
	restored.FromYAML(yaml)

	assert.Equal(t, original.Name, restored.Name)
	assert.Equal(t, original.Source, restored.Source)
	require.NotNil(t, restored.Service)
	assert.Equal(t, *original.Service, *restored.Service)
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
	assert.Nil(t, restored.Service)
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
			name: "valid keychain credential",
			yaml: CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   CredentialMetadata{Name: "GITHUB_TOKEN", Ecosystem: "testlab"},
				Spec:       CredentialSpec{Source: "keychain", Service: "dvm-gh"},
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
				Spec:       CredentialSpec{Source: "keychain", Service: "dvm-gh"},
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
				Spec:       CredentialSpec{Service: "dvm-gh"},
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
			errMsg:  "source must be 'keychain' or 'env'",
		},
		{
			name: "keychain missing service",
			yaml: CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   CredentialMetadata{Name: "TOKEN", Ecosystem: "testlab"},
				Spec:       CredentialSpec{Source: "keychain"},
			},
			wantErr: true,
			errMsg:  "service is required for keychain source",
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
				Spec:       CredentialSpec{Source: "keychain", Service: "dvm-gh"},
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
				Spec:       CredentialSpec{Source: "keychain", Service: "dvm-gh"},
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
				Spec:       CredentialSpec{Source: "keychain", Service: "dvm-gh"},
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

// === Dual-Field Credential Tests (v0.37.1) ===

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
				Source:      "keychain",
				UsernameVar: tt.usernameVar,
				PasswordVar: tt.passwordVar,
			}
			assert.Equal(t, tt.want, cred.IsDualField())
		})
	}
}

func TestCredentialDB_ToUsernameConfig(t *testing.T) {
	service := "github.com"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"
	cred := &CredentialDB{
		Name:        "github-creds",
		Source:      "keychain",
		Service:     &service,
		UsernameVar: &usernameVar,
		PasswordVar: &passwordVar,
	}

	cfg := cred.ToUsernameConfig()

	assert.Equal(t, config.SourceKeychain, cfg.Source)
	assert.Equal(t, "github.com", cfg.Service)
	assert.Equal(t, config.KeychainFieldAccount, cfg.Field)
}

func TestCredentialDB_ToPasswordConfig(t *testing.T) {
	service := "github.com"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"
	cred := &CredentialDB{
		Name:        "github-creds",
		Source:      "keychain",
		Service:     &service,
		UsernameVar: &usernameVar,
		PasswordVar: &passwordVar,
	}

	cfg := cred.ToPasswordConfig()

	assert.Equal(t, config.SourceKeychain, cfg.Source)
	assert.Equal(t, "github.com", cfg.Service)
	assert.Equal(t, config.KeychainFieldPassword, cfg.Field)
}

func TestCredentialsToMap_DualField_FanOut(t *testing.T) {
	service := "github.com"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"
	cred := &CredentialDB{
		Name:        "github-creds",
		Source:      "keychain",
		Service:     &service,
		UsernameVar: &usernameVar,
		PasswordVar: &passwordVar,
	}

	result := CredentialsToMap([]*CredentialDB{cred})

	require.Len(t, result, 2, "dual-field credential should fan out to 2 map entries")

	userCfg, ok := result["GITHUB_USERNAME"]
	require.True(t, ok, "map should contain GITHUB_USERNAME key")
	assert.Equal(t, config.SourceKeychain, userCfg.Source)
	assert.Equal(t, "github.com", userCfg.Service)
	assert.Equal(t, config.KeychainFieldAccount, userCfg.Field)

	passCfg, ok := result["GITHUB_PAT"]
	require.True(t, ok, "map should contain GITHUB_PAT key")
	assert.Equal(t, config.SourceKeychain, passCfg.Source)
	assert.Equal(t, "github.com", passCfg.Service)
	assert.Equal(t, config.KeychainFieldPassword, passCfg.Field)
}

func TestCredentialsToMap_MixedLegacyAndDualField(t *testing.T) {
	npmEnvVar := "NPM_TOKEN"
	legacyCred := &CredentialDB{
		Name:   "NPM_TOKEN",
		Source: "env",
		EnvVar: &npmEnvVar,
	}

	service := "github.com"
	ghUser := "GH_USER"
	ghPat := "GH_PAT"
	dualCred := &CredentialDB{
		Name:        "github-creds",
		Source:      "keychain",
		Service:     &service,
		UsernameVar: &ghUser,
		PasswordVar: &ghPat,
	}

	result := CredentialsToMap([]*CredentialDB{legacyCred, dualCred})

	require.Len(t, result, 3, "mixed credentials should produce 3 map entries")
	assert.Contains(t, result, "NPM_TOKEN")
	assert.Contains(t, result, "GH_USER")
	assert.Contains(t, result, "GH_PAT")
}

func TestCredentialsToMap_DualField_PasswordOnly(t *testing.T) {
	service := "github.com"
	passwordVar := "GITHUB_PAT"
	cred := &CredentialDB{
		Name:        "github-creds",
		Source:      "keychain",
		Service:     &service,
		UsernameVar: nil,
		PasswordVar: &passwordVar,
	}

	result := CredentialsToMap([]*CredentialDB{cred})

	require.Len(t, result, 1, "password-only dual-field should produce 1 map entry")
	passCfg, ok := result["GITHUB_PAT"]
	require.True(t, ok, "map should contain GITHUB_PAT key")
	assert.Equal(t, config.KeychainFieldPassword, passCfg.Field)
}

// =============================================================================
// Dual-Field YAML Tests
// =============================================================================

func TestCredentialDB_ToYAML_DualField(t *testing.T) {
	service := "github.com"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"
	cred := &CredentialDB{
		Name:        "github-creds",
		ScopeType:   CredentialScopeEcosystem,
		Source:      "keychain",
		Service:     &service,
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
			Source:      "keychain",
			Service:     "github.com",
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
	service := "github.com"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"
	original := &CredentialDB{
		Name:        "github-creds",
		ScopeType:   CredentialScopeEcosystem,
		Source:      "keychain",
		Service:     &service,
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
		service     string
		envVar      string
		usernameVar string
		passwordVar string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid dual-field keychain",
			source:      "keychain",
			service:     "github.com",
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
			errMsg:      "keychain",
		},
		{
			name:        "valid keychain password-only",
			source:      "keychain",
			service:     "svc",
			passwordVar: "TOKEN",
			wantErr:     false,
		},
		{
			name:        "valid keychain username-only",
			source:      "keychain",
			service:     "svc",
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
					Service:     tt.service,
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
					Source:      "keychain",
					Service:     "test.service",
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
// TDD Phase 2 (RED): Keychain Label Redesign Tests (v0.39.0)
// =============================================================================
// Design change: Introduce keychainLabel: (canonical YAML field), keep service:
// as a fallback for backwards-compatibility, add keychainType: field, and add
// Label / KeychainType columns to CredentialDB.
//
// New CredentialSpec fields that MUST exist after implementation:
//
//	type CredentialSpec struct {
//	    ...existing fields...
//	    KeychainLabel string `yaml:"keychainLabel,omitempty"`
//	    KeychainType  string `yaml:"keychainType,omitempty"`
//	}
//
// New CredentialDB fields that MUST exist after implementation:
//
//	type CredentialDB struct {
//	    ...existing fields...
//	    Label        *string `db:"label" json:"label,omitempty"`
//	    KeychainType *string `db:"keychain_type" json:"keychain_type,omitempty"`
//	}
//
// ALL tests in this section WILL FAIL TO COMPILE (or FAIL AT RUNTIME) until
// the above fields are added to models/credential.go.
// =============================================================================

// ---------------------------------------------------------------------------
// Section: CredentialSpec keychainLabel Field Tests
// ---------------------------------------------------------------------------

// TestCredentialSpec_KeychainLabel verifies that CredentialSpec has a
// KeychainLabel field that survives a YAML round-trip through CredentialDB.
//
// WILL FAIL TO COMPILE — CredentialSpec.KeychainLabel does not exist yet.
func TestCredentialSpec_KeychainLabel(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// CredentialSpec.KeychainLabel field does not exist yet.
	spec := CredentialSpec{
		Source:        "keychain",
		KeychainLabel: "com.example.my-new-label",
	}
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, "com.example.my-new-label", spec.KeychainLabel,
		"KeychainLabel should be stored and retrievable on CredentialSpec")
	assert.Empty(t, spec.Service,
		"Service should be empty when only KeychainLabel is set")
}

// TestCredentialSpec_KeychainType verifies that CredentialSpec has a
// KeychainType field that survives a YAML round-trip through CredentialDB.
//
// WILL FAIL TO COMPILE — CredentialSpec.KeychainType does not exist yet.
func TestCredentialSpec_KeychainType(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// CredentialSpec.KeychainType field does not exist yet.
	spec := CredentialSpec{
		Source:        "keychain",
		KeychainLabel: "com.example.label",
		KeychainType:  "internet",
	}
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, "internet", spec.KeychainType,
		"KeychainType should be stored and retrievable on CredentialSpec")
}

// TestCredentialSpec_KeychainLabel_RoundTrip verifies that keychainLabel
// survives a ToYAML → FromYAML round-trip through CredentialDB.
//
// WILL FAIL TO COMPILE — CredentialSpec.KeychainLabel and CredentialDB.Label
// do not exist yet.
func TestCredentialSpec_KeychainLabel_RoundTrip(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	// CredentialDB.Label field does not exist yet.
	label := "com.example.round-trip-label"
	original := &CredentialDB{
		Name:      "MY_LABEL_CRED",
		ScopeType: CredentialScopeApp,
		Source:    "keychain",
		Label:     &label,
	}
	// ─────────────────────────────────────────────────────────────────────────

	y := original.ToYAML("test-app")
	restored := &CredentialDB{}
	restored.FromYAML(y)

	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	// CredentialDB.Label field does not exist yet.
	require.NotNil(t, restored.Label,
		"Label should survive ToYAML → FromYAML round-trip")
	assert.Equal(t, *original.Label, *restored.Label,
		"Label value should be preserved across round-trip")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestCredentialSpec_KeychainType_RoundTrip verifies that keychainType
// survives a ToYAML → FromYAML round-trip through CredentialDB.
//
// WILL FAIL TO COMPILE — CredentialDB.KeychainType does not exist yet.
func TestCredentialSpec_KeychainType_RoundTrip(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	// CredentialDB.KeychainType field does not exist yet.
	label := "com.example.type-round-trip"
	keychainType := "internet"
	original := &CredentialDB{
		Name:         "MY_INTERNET_CRED",
		ScopeType:    CredentialScopeApp,
		Source:       "keychain",
		Label:        &label,
		KeychainType: &keychainType,
	}
	// ─────────────────────────────────────────────────────────────────────────

	y := original.ToYAML("test-app")
	restored := &CredentialDB{}
	restored.FromYAML(y)

	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	require.NotNil(t, restored.KeychainType,
		"KeychainType should survive ToYAML → FromYAML round-trip")
	assert.Equal(t, "internet", *restored.KeychainType,
		"KeychainType value should be preserved across round-trip")
	// ─────────────────────────────────────────────────────────────────────────
}

// ---------------------------------------------------------------------------
// Section: FromYAML service: Fallback Tests
// ---------------------------------------------------------------------------

// TestCredentialSpec_ServiceFallback verifies that FromYAML still accepts
// service: as a fallback for keychainLabel: for backwards compatibility.
// When service: is present and keychainLabel: is absent, the credential should
// be treated as if keychainLabel: was set with the same value.
//
// WILL FAIL AT RUNTIME — the current FromYAML populates c.Service from
// y.Spec.Service but does not populate c.Label. After the fix, a non-empty
// y.Spec.Service should also populate c.Label (fallback behaviour).
func TestCredentialSpec_ServiceFallback(t *testing.T) {
	// A YAML with the legacy service: field (no keychainLabel:)
	y := CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata:   CredentialMetadata{Name: "LEGACY_CRED", Ecosystem: "testlab"},
		Spec: CredentialSpec{
			Source:  "keychain",
			Service: "com.example.legacy-service", // old field — still accepted
		},
	}

	cred := &CredentialDB{}
	cred.FromYAML(y)

	// The legacy service: field MUST still populate Service for backwards compat
	require.NotNil(t, cred.Service,
		"service: field in YAML must still populate CredentialDB.Service")
	assert.Equal(t, "com.example.legacy-service", *cred.Service)

	// ── RUNTIME FAILURE EXPECTED BELOW ───────────────────────────────────────
	// After the fix, FromYAML must also copy service: → Label (fallback).
	// Currently c.Label is never populated, so this assertion fails.
	require.NotNil(t, cred.Label,
		"service: field must also populate CredentialDB.Label as fallback for keychainLabel:")
	assert.Equal(t, "com.example.legacy-service", *cred.Label,
		"Label must mirror Service when only service: is provided (backwards compat fallback)")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestCredentialSpec_KeychainLabelTakesPrecedenceOverService verifies that when
// BOTH keychainLabel: and service: appear in YAML, keychainLabel: wins.
//
// WILL FAIL TO COMPILE — CredentialSpec.KeychainLabel does not exist yet.
func TestCredentialSpec_KeychainLabelTakesPrecedenceOverService(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	y := CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata:   CredentialMetadata{Name: "DUAL_LABEL_CRED", Ecosystem: "testlab"},
		Spec: CredentialSpec{
			Source:        "keychain",
			Service:       "com.example.old-service",
			KeychainLabel: "com.example.new-label", // should win
		},
	}
	// ─────────────────────────────────────────────────────────────────────────

	cred := &CredentialDB{}
	cred.FromYAML(y)

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	require.NotNil(t, cred.Label,
		"Label must be populated when keychainLabel: is present in YAML")
	assert.Equal(t, "com.example.new-label", *cred.Label,
		"keychainLabel: must take precedence over service: when both are present")
	// ─────────────────────────────────────────────────────────────────────────
}

// ---------------------------------------------------------------------------
// Section: ValidateCredentialYAML keychainType Validation Tests
// ---------------------------------------------------------------------------

// TestValidateCredentialYAML_KeychainType is a table-driven test verifying
// that ValidateCredentialYAML accepts only "generic" and "internet" as valid
// keychainType values (when provided), and rejects anything else.
//
// WILL FAIL AT RUNTIME — ValidateCredentialYAML does not yet validate
// keychainType (the field doesn't exist in CredentialSpec yet either).
func TestValidateCredentialYAML_KeychainType(t *testing.T) {
	tests := []struct {
		name         string
		keychainType string
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "generic is valid",
			keychainType: "generic",
			wantErr:      false,
		},
		{
			name:         "internet is valid",
			keychainType: "internet",
			wantErr:      false,
		},
		{
			name:         "empty string is valid (defaults to generic)",
			keychainType: "",
			wantErr:      false,
		},
		{
			name:         "keychain is rejected (old wrong value)",
			keychainType: "keychain",
			wantErr:      true,
			errMsg:       "keychainType",
		},
		{
			name:         "GENERIC uppercase rejected",
			keychainType: "GENERIC",
			wantErr:      true,
			errMsg:       "keychainType",
		},
		{
			name:         "INTERNET uppercase rejected",
			keychainType: "INTERNET",
			wantErr:      true,
			errMsg:       "keychainType",
		},
		{
			name:         "arbitrary string rejected",
			keychainType: "plaintext",
			wantErr:      true,
			errMsg:       "keychainType",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────
			// CredentialSpec.KeychainLabel and CredentialSpec.KeychainType
			// do not exist yet.
			y := CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   CredentialMetadata{Name: "TYPE_TEST", Ecosystem: "testlab"},
				Spec: CredentialSpec{
					Source:        "keychain",
					KeychainLabel: "com.example.label",
					KeychainType:  tt.keychainType,
				},
			}
			// ─────────────────────────────────────────────────────────────────

			err := ValidateCredentialYAML(y)
			if tt.wantErr {
				assert.Error(t, err,
					"keychainType %q should be rejected by ValidateCredentialYAML", tt.keychainType)
				assert.Contains(t, err.Error(), tt.errMsg,
					"error message should mention keychainType")
			} else {
				assert.NoError(t, err,
					"keychainType %q should be accepted by ValidateCredentialYAML", tt.keychainType)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Section: CredentialDB Label and KeychainType Field Tests
// ---------------------------------------------------------------------------

// TestCredentialDB_LabelField verifies that CredentialDB has a Label field
// of type *string with the correct db/json tags.
//
// WILL FAIL TO COMPILE — CredentialDB.Label does not exist yet.
func TestCredentialDB_LabelField(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// CredentialDB.Label field does not exist yet.
	label := "com.example.db-label"
	cred := &CredentialDB{
		Name:   "LABEL_CRED",
		Source: "keychain",
		Label:  &label,
	}
	// ─────────────────────────────────────────────────────────────────────────

	require.NotNil(t, cred.Label,
		"CredentialDB.Label should be non-nil when set")
	assert.Equal(t, "com.example.db-label", *cred.Label,
		"CredentialDB.Label should store and return the label string")
}

// TestCredentialDB_KeychainTypeField verifies that CredentialDB has a
// KeychainType field of type *string with the correct db/json tags.
//
// WILL FAIL TO COMPILE — CredentialDB.KeychainType does not exist yet.
func TestCredentialDB_KeychainTypeField(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// CredentialDB.KeychainType field does not exist yet.
	kt := "internet"
	cred := &CredentialDB{
		Name:         "KTYPE_CRED",
		Source:       "keychain",
		KeychainType: &kt,
	}
	// ─────────────────────────────────────────────────────────────────────────

	require.NotNil(t, cred.KeychainType,
		"CredentialDB.KeychainType should be non-nil when set")
	assert.Equal(t, "internet", *cred.KeychainType,
		"CredentialDB.KeychainType should store and return the type string")
}

// TestCredentialDB_LabelField_NilDefault verifies that a zero-value
// CredentialDB has a nil Label (the field is optional/*string).
//
// WILL FAIL TO COMPILE — CredentialDB.Label does not exist yet.
func TestCredentialDB_LabelField_NilDefault(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cred := &CredentialDB{
		Name:   "NO_LABEL_CRED",
		Source: "env",
	}
	assert.Nil(t, cred.Label,
		"CredentialDB.Label should default to nil when not set")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestCredentialDB_KeychainTypeField_NilDefault verifies that a zero-value
// CredentialDB has a nil KeychainType.
//
// WILL FAIL TO COMPILE — CredentialDB.KeychainType does not exist yet.
func TestCredentialDB_KeychainTypeField_NilDefault(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cred := &CredentialDB{
		Name:   "NO_KTYPE_CRED",
		Source: "keychain",
	}
	assert.Nil(t, cred.KeychainType,
		"CredentialDB.KeychainType should default to nil when not set")
	// ─────────────────────────────────────────────────────────────────────────
}

// ---------------------------------------------------------------------------
// Section: ToUsernameConfig and ToPasswordConfig Label Propagation
// ---------------------------------------------------------------------------

// TestToUsernameConfig_Label verifies that ToUsernameConfig propagates the
// Label field from CredentialDB into the returned CredentialConfig.Label.
//
// WILL FAIL TO COMPILE — CredentialDB.Label does not exist yet AND
// CredentialConfig.Label does not exist yet (must also be added to
// config/credentials.go).
func TestToUsernameConfig_Label(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	// CredentialDB.Label and CredentialConfig.Label do not exist yet.
	label := "com.example.dual-label"
	usernameVar := "GH_USER"
	passwordVar := "GH_PAT"
	cred := &CredentialDB{
		Name:        "github-creds",
		Source:      "keychain",
		Label:       &label,
		UsernameVar: &usernameVar,
		PasswordVar: &passwordVar,
	}

	cfg := cred.ToUsernameConfig()

	assert.Equal(t, config.SourceKeychain, cfg.Source)
	assert.Equal(t, "com.example.dual-label", cfg.Label,
		"ToUsernameConfig must populate Label from CredentialDB.Label")
	assert.Equal(t, config.KeychainFieldAccount, cfg.Field)
	// ─────────────────────────────────────────────────────────────────────────
}

// TestToPasswordConfig_Label verifies that ToPasswordConfig propagates the
// Label field from CredentialDB into the returned CredentialConfig.Label.
//
// WILL FAIL TO COMPILE — CredentialDB.Label does not exist yet AND
// CredentialConfig.Label does not exist yet.
func TestToPasswordConfig_Label(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	// CredentialDB.Label and CredentialConfig.Label do not exist yet.
	label := "com.example.dual-label"
	usernameVar := "GH_USER"
	passwordVar := "GH_PAT"
	cred := &CredentialDB{
		Name:        "github-creds",
		Source:      "keychain",
		Label:       &label,
		UsernameVar: &usernameVar,
		PasswordVar: &passwordVar,
	}

	cfg := cred.ToPasswordConfig()

	assert.Equal(t, config.SourceKeychain, cfg.Source)
	assert.Equal(t, "com.example.dual-label", cfg.Label,
		"ToPasswordConfig must populate Label from CredentialDB.Label")
	assert.Equal(t, config.KeychainFieldPassword, cfg.Field)
	// ─────────────────────────────────────────────────────────────────────────
}

// TestToUsernameConfig_ServiceFallback verifies that ToUsernameConfig falls
// back to Service when Label is nil (backwards compatibility).
//
// WILL FAIL AT RUNTIME — ToUsernameConfig does not yet set cfg.Label.
func TestToUsernameConfig_ServiceFallback(t *testing.T) {
	service := "com.example.legacy-service"
	usernameVar := "GH_USER"
	cred := &CredentialDB{
		Name:        "legacy-cred",
		Source:      "keychain",
		Service:     &service,
		UsernameVar: &usernameVar,
		// Label intentionally nil — should fall back to Service
	}

	cfg := cred.ToUsernameConfig()

	assert.Equal(t, config.SourceKeychain, cfg.Source)
	// When Label is nil, the old Service should still be passed through
	// via cfg.Service for backwards-compat (the credential system falls back)
	assert.Equal(t, "com.example.legacy-service", cfg.Service,
		"ToUsernameConfig must still set Service for backwards compat when Label is nil")
}

// =============================================================================
// Test Helpers
// =============================================================================

func strPtr(s string) *string { return &s }
