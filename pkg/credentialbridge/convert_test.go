package credentialbridge_test

import (
	"fmt"
	"testing"

	"devopsmaestro/config"
	"devopsmaestro/models"
	"devopsmaestro/pkg/credentialbridge"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string { return &s }

// =============================================================================
// ToUsernameConfig / ToPasswordConfig Tests
// =============================================================================

func TestCredentialDB_ToUsernameConfig(t *testing.T) {
	vaultSecret := "github-pat"
	vaultUsernameSecret := "github-username"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"
	cred := &models.CredentialDB{
		Name:                "github-creds",
		Source:              "vault",
		VaultSecret:         &vaultSecret,
		VaultUsernameSecret: &vaultUsernameSecret,
		UsernameVar:         &usernameVar,
		PasswordVar:         &passwordVar,
	}

	cfg := credentialbridge.ToUsernameConfig(cred)

	assert.Equal(t, config.SourceVault, cfg.Source)
	assert.Equal(t, "github-username", cfg.VaultSecret)
}

func TestCredentialDB_ToPasswordConfig(t *testing.T) {
	vaultSecret := "github-pat"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"
	cred := &models.CredentialDB{
		Name:        "github-creds",
		Source:      "vault",
		VaultSecret: &vaultSecret,
		UsernameVar: &usernameVar,
		PasswordVar: &passwordVar,
	}

	cfg := credentialbridge.ToPasswordConfig(cred)

	assert.Equal(t, config.SourceVault, cfg.Source)
	assert.Equal(t, "github-pat", cfg.VaultSecret)
}

// =============================================================================
// CredentialsToMap Tests
// =============================================================================

func TestCredentialsToMap_DualField_FanOut(t *testing.T) {
	vaultSecret := "github-pat"
	vaultUsernameSecret := "github-username"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"
	cred := &models.CredentialDB{
		Name:                "github-creds",
		Source:              "vault",
		VaultSecret:         &vaultSecret,
		VaultUsernameSecret: &vaultUsernameSecret,
		UsernameVar:         &usernameVar,
		PasswordVar:         &passwordVar,
	}

	result := credentialbridge.CredentialsToMap([]*models.CredentialDB{cred})

	require.Len(t, result, 2, "dual-field credential should fan out to 2 map entries")

	userCfg, ok := result["GITHUB_USERNAME"]
	require.True(t, ok, "map should contain GITHUB_USERNAME key")
	assert.Equal(t, config.SourceVault, userCfg.Source)
	assert.Equal(t, "github-username", userCfg.VaultSecret)

	passCfg, ok := result["GITHUB_PAT"]
	require.True(t, ok, "map should contain GITHUB_PAT key")
	assert.Equal(t, config.SourceVault, passCfg.Source)
	assert.Equal(t, "github-pat", passCfg.VaultSecret)
}

func TestCredentialsToMap_MixedLegacyAndDualField(t *testing.T) {
	npmEnvVar := "NPM_TOKEN"
	legacyCred := &models.CredentialDB{
		Name:   "NPM_TOKEN",
		Source: "env",
		EnvVar: &npmEnvVar,
	}

	vaultSecret := "github-pat"
	vaultUsernameSecret := "github-username"
	ghUser := "GH_USER"
	ghPat := "GH_PAT"
	dualCred := &models.CredentialDB{
		Name:                "github-creds",
		Source:              "vault",
		VaultSecret:         &vaultSecret,
		VaultUsernameSecret: &vaultUsernameSecret,
		UsernameVar:         &ghUser,
		PasswordVar:         &ghPat,
	}

	result := credentialbridge.CredentialsToMap([]*models.CredentialDB{legacyCred, dualCred})

	require.Len(t, result, 3, "mixed credentials should produce 3 map entries")
	assert.Contains(t, result, "NPM_TOKEN")
	assert.Contains(t, result, "GH_USER")
	assert.Contains(t, result, "GH_PAT")
}

func TestCredentialsToMap_DualField_PasswordOnly(t *testing.T) {
	vaultSecret := "github-pat"
	passwordVar := "GITHUB_PAT"
	cred := &models.CredentialDB{
		Name:        "github-creds",
		Source:      "vault",
		VaultSecret: &vaultSecret,
		UsernameVar: nil,
		PasswordVar: &passwordVar,
	}

	result := credentialbridge.CredentialsToMap([]*models.CredentialDB{cred})

	require.Len(t, result, 1, "password-only dual-field should produce 1 map entry")
	passCfg, ok := result["GITHUB_PAT"]
	require.True(t, ok, "map should contain GITHUB_PAT key")
	assert.Equal(t, config.SourceVault, passCfg.Source)
	assert.Equal(t, "github-pat", passCfg.VaultSecret)
}

// =============================================================================
// ValidateCredentialYAML Tests
// =============================================================================

func TestValidateCredentialYAML(t *testing.T) {
	tests := []struct {
		name    string
		yaml    models.CredentialYAML
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid vault credential",
			yaml: models.CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   models.CredentialMetadata{Name: "GITHUB_TOKEN", Ecosystem: "testlab"},
				Spec:       models.CredentialSpec{Source: "vault", VaultSecret: "dvm-gh"},
			},
			wantErr: false,
		},
		{
			name: "valid env credential",
			yaml: models.CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   models.CredentialMetadata{Name: "NPM_TOKEN", App: "rust-svc"},
				Spec:       models.CredentialSpec{Source: "env", EnvVar: "MY_NPM"},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			yaml: models.CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   models.CredentialMetadata{Ecosystem: "testlab"},
				Spec:       models.CredentialSpec{Source: "vault", VaultSecret: "dvm-gh"},
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "missing source",
			yaml: models.CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   models.CredentialMetadata{Name: "TOKEN", Ecosystem: "testlab"},
				Spec:       models.CredentialSpec{VaultSecret: "dvm-gh"},
			},
			wantErr: true,
			errMsg:  "source is required",
		},
		{
			name: "invalid source",
			yaml: models.CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   models.CredentialMetadata{Name: "TOKEN", Ecosystem: "testlab"},
				Spec:       models.CredentialSpec{Source: "plaintext"},
			},
			wantErr: true,
			errMsg:  "source must be 'vault' or 'env'",
		},
		{
			name: "vault missing vaultSecret",
			yaml: models.CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   models.CredentialMetadata{Name: "TOKEN", Ecosystem: "testlab"},
				Spec:       models.CredentialSpec{Source: "vault"},
			},
			wantErr: true,
			errMsg:  "vaultSecret is required for vault source",
		},
		{
			name: "env missing env-var",
			yaml: models.CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   models.CredentialMetadata{Name: "TOKEN", Ecosystem: "testlab"},
				Spec:       models.CredentialSpec{Source: "env"},
			},
			wantErr: true,
			errMsg:  "env-var is required for env source",
		},
		{
			name: "no scope specified",
			yaml: models.CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   models.CredentialMetadata{Name: "TOKEN"},
				Spec:       models.CredentialSpec{Source: "vault", VaultSecret: "dvm-gh"},
			},
			wantErr: true,
			errMsg:  "exactly one scope",
		},
		{
			name: "multiple scopes specified",
			yaml: models.CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   models.CredentialMetadata{Name: "TOKEN", Ecosystem: "testlab", App: "rust-svc"},
				Spec:       models.CredentialSpec{Source: "vault", VaultSecret: "dvm-gh"},
			},
			wantErr: true,
			errMsg:  "exactly one scope",
		},
		{
			name: "wrong kind",
			yaml: models.CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Registry",
				Metadata:   models.CredentialMetadata{Name: "TOKEN", Ecosystem: "testlab"},
				Spec:       models.CredentialSpec{Source: "vault", VaultSecret: "dvm-gh"},
			},
			wantErr: true,
			errMsg:  "kind must be 'Credential'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := credentialbridge.ValidateCredentialYAML(tt.yaml)
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
// ValidateCredentialYAML Dual-Field Tests
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
			y := models.CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   models.CredentialMetadata{Name: "test-cred", Ecosystem: "testlab"},
				Spec: models.CredentialSpec{
					Source:      tt.source,
					VaultSecret: tt.vaultSecret,
					EnvVar:      tt.envVar,
					UsernameVar: tt.usernameVar,
					PasswordVar: tt.passwordVar,
				},
			}
			err := credentialbridge.ValidateCredentialYAML(y)
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
// ValidateCredentialYAML Dual-Field Env Key Validation Tests
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
			y := models.CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   models.CredentialMetadata{Name: "test-cred", Ecosystem: "test"},
				Spec: models.CredentialSpec{
					Source:      "vault",
					VaultSecret: "test-secret",
					UsernameVar: tt.usernameVar,
					PasswordVar: tt.passwordVar,
				},
			}
			err := credentialbridge.ValidateCredentialYAML(y)
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
// ToConfig Vault Tests
// =============================================================================

func TestCredentialDB_ToConfig_Vault(t *testing.T) {
	vaultSecret := "my-secret"
	vaultEnv := "staging"

	cred := &models.CredentialDB{
		Name:        "test-cred",
		Source:      "vault",
		VaultSecret: &vaultSecret,
		VaultEnv:    &vaultEnv,
	}
	cfg := credentialbridge.ToConfig(cred)

	assert.Equal(t, config.SourceVault, cfg.Source,
		"ToConfig must set Source to SourceVault for vault-sourced credentials")
	assert.Equal(t, "my-secret", cfg.VaultSecret,
		"ToConfig must copy VaultSecret from CredentialDB to CredentialConfig")
	assert.Equal(t, "staging", cfg.VaultEnv,
		"ToConfig must copy VaultEnv from CredentialDB to CredentialConfig")
}

func TestCredentialDB_ToConfig_Vault_NilVaultEnv(t *testing.T) {
	vaultSecret := "my-secret"

	cred := &models.CredentialDB{
		Name:        "test-cred",
		Source:      "vault",
		VaultSecret: &vaultSecret,
		VaultEnv:    nil,
	}
	cfg := credentialbridge.ToConfig(cred)

	assert.Equal(t, config.SourceVault, cfg.Source)
	assert.Equal(t, "my-secret", cfg.VaultSecret)
	assert.Empty(t, cfg.VaultEnv,
		"nil VaultEnv in DB must produce empty string in CredentialConfig")
}

// =============================================================================
// ToUsernameConfig / ToPasswordConfig Vault Tests
// =============================================================================

func TestCredentialDB_ToUsernameConfig_Vault(t *testing.T) {
	vaultSecret := "github-pat"
	vaultUsernameSecret := "github-user"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"

	cred := &models.CredentialDB{
		Name:                "github-creds",
		Source:              "vault",
		VaultSecret:         &vaultSecret,
		VaultUsernameSecret: &vaultUsernameSecret,
		UsernameVar:         &usernameVar,
		PasswordVar:         &passwordVar,
	}
	cfg := credentialbridge.ToUsernameConfig(cred)

	assert.Equal(t, config.SourceVault, cfg.Source,
		"ToUsernameConfig must preserve vault source")
	assert.Equal(t, "github-user", cfg.VaultSecret,
		"ToUsernameConfig must use VaultUsernameSecret as the VaultSecret for username lookup")
}

func TestCredentialDB_ToPasswordConfig_Vault(t *testing.T) {
	vaultSecret := "github-pat"
	vaultUsernameSecret := "github-user"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"

	cred := &models.CredentialDB{
		Name:                "github-creds",
		Source:              "vault",
		VaultSecret:         &vaultSecret,
		VaultUsernameSecret: &vaultUsernameSecret,
		UsernameVar:         &usernameVar,
		PasswordVar:         &passwordVar,
	}
	cfg := credentialbridge.ToPasswordConfig(cred)

	assert.Equal(t, config.SourceVault, cfg.Source,
		"ToPasswordConfig must preserve vault source")
	assert.Equal(t, "github-pat", cfg.VaultSecret,
		"ToPasswordConfig must use VaultSecret (primary secret) for password lookup")
}

// =============================================================================
// ValidateCredentialYAML Vault Tests
// =============================================================================

func TestValidateCredentialYAML_Vault(t *testing.T) {
	tests := []struct {
		name    string
		spec    models.CredentialSpec
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid vault credential",
			spec: models.CredentialSpec{
				Source:      "vault",
				VaultSecret: "my-github-pat",
			},
			wantErr: false,
		},
		{
			name: "valid vault credential with environment",
			spec: models.CredentialSpec{
				Source:           "vault",
				VaultSecret:      "my-github-pat",
				VaultEnvironment: "production",
			},
			wantErr: false,
		},
		{
			name: "valid vault dual-field credential",
			spec: models.CredentialSpec{
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
			spec: models.CredentialSpec{
				Source:           "vault",
				VaultEnvironment: "production",
			},
			wantErr: true,
			errMsg:  "vault",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			y := models.CredentialYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "Credential",
				Metadata:   models.CredentialMetadata{Name: "TEST_CRED", Ecosystem: "testlab"},
				Spec:       tt.spec,
			}
			err := credentialbridge.ValidateCredentialYAML(y)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCredentialYAML_Vault_SourceAccepted(t *testing.T) {
	y := models.CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata:   models.CredentialMetadata{Name: "TEST_CRED", Ecosystem: "testlab"},
		Spec: models.CredentialSpec{
			Source:      "vault",
			VaultSecret: "my-secret",
		},
	}

	err := credentialbridge.ValidateCredentialYAML(y)

	assert.NoError(t, err,
		"source='vault' must be accepted by ValidateCredentialYAML after implementation")
}

func TestValidateCredentialYAML_Vault_RequiresVaultSecret(t *testing.T) {
	y := models.CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata:   models.CredentialMetadata{Name: "TEST_CRED", Ecosystem: "testlab"},
		Spec: models.CredentialSpec{
			Source: "vault",
		},
	}

	err := credentialbridge.ValidateCredentialYAML(y)

	assert.Error(t, err,
		"vault source without vaultSecret must be rejected by ValidateCredentialYAML")
	assert.Contains(t, err.Error(), "vault",
		"error must mention vault or vaultSecret")
}

func TestValidateCredentialYAML_Vault_UsernameSecretRequiresUsernameVar(t *testing.T) {
	y := models.CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata:   models.CredentialMetadata{Name: "TEST_CRED", Ecosystem: "testlab"},
		Spec: models.CredentialSpec{
			Source:              "vault",
			VaultSecret:         "my-pat",
			VaultUsernameSecret: "my-username-secret",
		},
	}

	err := credentialbridge.ValidateCredentialYAML(y)

	assert.Error(t, err,
		"vaultUsernameSecret requires usernameVar to be set")
	assert.Contains(t, err.Error(), "usernameVar",
		"error must mention usernameVar when vaultUsernameSecret is set without it")
}

// =============================================================================
// ToMapEntries Tests
// =============================================================================

func TestCredentialDB_ToMapEntries_VaultFields_FansOut(t *testing.T) {
	raw := `{"GITHUB_TOKEN":"token","GITHUB_USER":"username"}`

	cred := &models.CredentialDB{
		Name:        "github-creds",
		Source:      "vault",
		VaultSecret: strPtr("github/creds"),
		VaultEnv:    strPtr("production"),
		VaultFields: strPtr(raw),
	}
	entries := credentialbridge.ToMapEntries(cred)

	assert.Len(t, entries, 2, "one CredentialConfig per vault field")

	assert.Contains(t, entries, "GITHUB_TOKEN",
		"ToMapEntries must produce an entry for GITHUB_TOKEN")
	assert.Contains(t, entries, "GITHUB_USER",
		"ToMapEntries must produce an entry for GITHUB_USER")

	assert.Equal(t, "token", entries["GITHUB_TOKEN"].VaultField)
	assert.Equal(t, "username", entries["GITHUB_USER"].VaultField)
}

func TestCredentialDB_ToMapEntries_NoVaultFields_SingleEntry(t *testing.T) {
	cred := &models.CredentialDB{
		Name:        "github-token",
		Source:      "vault",
		VaultSecret: strPtr("github/token"),
	}
	entries := credentialbridge.ToMapEntries(cred)

	assert.Len(t, entries, 1, "non-fan-out credential must produce exactly 1 entry")
	assert.Contains(t, entries, "github-token", "key should be credential name for simple credentials")
}

func TestCredentialDB_ToMapEntries_EnvSource(t *testing.T) {
	cred := &models.CredentialDB{
		Name:   "my-token",
		Source: "env",
		EnvVar: strPtr("MY_API_TOKEN"),
	}
	entries := credentialbridge.ToMapEntries(cred)

	assert.Len(t, entries, 1, "env credential must produce exactly 1 entry")
	assert.Contains(t, entries, "my-token", "key should be credential name for simple credentials")
}

// =============================================================================
// ValidateCredentialYAML VaultFields Validation Tests
// =============================================================================

func TestValidateCredentialYAML_VaultFields_RequiresVaultSource(t *testing.T) {
	y := models.CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata:   models.CredentialMetadata{Name: "MY_CRED", Ecosystem: "testlab"},
		Spec: models.CredentialSpec{
			Source: "env",
			EnvVar: "MY_VAR",
			VaultFields: map[string]string{
				"SOME_VAR": "some-field",
			},
		},
	}

	err := credentialbridge.ValidateCredentialYAML(y)
	assert.Error(t, err,
		"vaultFields with env source must be rejected by ValidateCredentialYAML")
}

func TestValidateCredentialYAML_VaultFields_MutuallyExclusiveWithUsernameVar(t *testing.T) {
	y := models.CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata:   models.CredentialMetadata{Name: "MY_CRED", Ecosystem: "testlab"},
		Spec: models.CredentialSpec{
			Source:      "vault",
			VaultSecret: "my-org/creds",
			UsernameVar: "MY_USER",
			VaultFields: map[string]string{
				"TOKEN": "token",
			},
		},
	}

	err := credentialbridge.ValidateCredentialYAML(y)
	assert.Error(t, err,
		"vaultFields and usernameVar are mutually exclusive")
}

func TestValidateCredentialYAML_VaultFields_MaxFiftyFields(t *testing.T) {
	fields := make(map[string]string, 51)
	for i := 0; i < 51; i++ {
		key := fmt.Sprintf("ENV_VAR_%03d", i)
		fields[key] = fmt.Sprintf("field_%03d", i)
	}

	y := models.CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata:   models.CredentialMetadata{Name: "MY_CRED", Ecosystem: "testlab"},
		Spec: models.CredentialSpec{
			Source:      "vault",
			VaultSecret: "my-org/creds",
			VaultFields: fields,
		},
	}

	err := credentialbridge.ValidateCredentialYAML(y)
	assert.Error(t, err, "vaultFields with more than 50 entries must be rejected")
	assert.Contains(t, err.Error(), "50",
		"error message must mention the 50-field limit")
}

func TestValidateCredentialYAML_VaultFields_ExactlyFiftyFields(t *testing.T) {
	fields := make(map[string]string, 50)
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("ENV_VAR_%03d", i)
		fields[key] = fmt.Sprintf("field_%03d", i)
	}

	y := models.CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata:   models.CredentialMetadata{Name: "MY_CRED", Ecosystem: "testlab"},
		Spec: models.CredentialSpec{
			Source:      "vault",
			VaultSecret: "my-org/creds",
			VaultFields: fields,
		},
	}

	err := credentialbridge.ValidateCredentialYAML(y)
	assert.NoError(t, err, "exactly 50 vault fields must be accepted")
}

// =============================================================================
// Bug #157 — VaultField Tests
// =============================================================================

func TestCredentialDB_ToUsernameConfig_SetsVaultField(t *testing.T) {
	vaultSecret := "github-creds"
	vaultEnv := "production"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"

	cred := &models.CredentialDB{
		Name:        "github-creds",
		Source:      "vault",
		VaultSecret: &vaultSecret,
		VaultEnv:    &vaultEnv,
		UsernameVar: &usernameVar,
		PasswordVar: &passwordVar,
	}

	cfg := credentialbridge.ToUsernameConfig(cred)

	assert.Equal(t, config.SourceVault, cfg.Source,
		"ToUsernameConfig must set Source to SourceVault")
	assert.Equal(t, "github-creds", cfg.VaultSecret,
		"ToUsernameConfig must copy VaultSecret")
	assert.Equal(t, "production", cfg.VaultEnv,
		"ToUsernameConfig must copy VaultEnv")
	assert.Equal(t, "GITHUB_USERNAME", cfg.VaultField,
		"ToUsernameConfig MUST set VaultField to the UsernameVar name so "+
			"vault resolution calls GetField(secret, env, field) not Get(secret, env)")
}

func TestCredentialDB_ToPasswordConfig_SetsVaultField(t *testing.T) {
	vaultSecret := "github-creds"
	vaultEnv := "production"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"

	cred := &models.CredentialDB{
		Name:        "github-creds",
		Source:      "vault",
		VaultSecret: &vaultSecret,
		VaultEnv:    &vaultEnv,
		UsernameVar: &usernameVar,
		PasswordVar: &passwordVar,
	}

	cfg := credentialbridge.ToPasswordConfig(cred)

	assert.Equal(t, config.SourceVault, cfg.Source,
		"ToPasswordConfig must set Source to SourceVault")
	assert.Equal(t, "github-creds", cfg.VaultSecret,
		"ToPasswordConfig must copy VaultSecret")
	assert.Equal(t, "production", cfg.VaultEnv,
		"ToPasswordConfig must copy VaultEnv")
	assert.Equal(t, "GITHUB_PAT", cfg.VaultField,
		"ToPasswordConfig MUST set VaultField to the PasswordVar name so "+
			"vault resolution calls GetField(secret, env, field) not Get(secret, env)")
}

func TestCredentialDB_ToUsernameConfig_SetsVaultField_WithVaultUsernameSecret(t *testing.T) {
	vaultSecret := "github-pat-secret"
	vaultUsernameSecret := "github-user-secret"
	usernameVar := "GITHUB_USERNAME"

	cred := &models.CredentialDB{
		Name:                "github-creds",
		Source:              "vault",
		VaultSecret:         &vaultSecret,
		VaultUsernameSecret: &vaultUsernameSecret,
		UsernameVar:         &usernameVar,
	}

	cfg := credentialbridge.ToUsernameConfig(cred)

	assert.Equal(t, config.SourceVault, cfg.Source)
	assert.Equal(t, "github-user-secret", cfg.VaultSecret,
		"ToUsernameConfig must use VaultUsernameSecret when available")
	assert.Equal(t, "GITHUB_USERNAME", cfg.VaultField,
		"ToUsernameConfig MUST set VaultField to the UsernameVar name even when "+
			"VaultUsernameSecret is used")
}

func TestCredentialDB_ToMapEntries_DualField_SetsVaultField(t *testing.T) {
	vaultSecret := "github-creds"
	vaultEnv := "staging"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"

	cred := &models.CredentialDB{
		Name:        "github-creds",
		Source:      "vault",
		VaultSecret: &vaultSecret,
		VaultEnv:    &vaultEnv,
		UsernameVar: &usernameVar,
		PasswordVar: &passwordVar,
	}

	entries := credentialbridge.ToMapEntries(cred)

	require.Len(t, entries, 2,
		"dual-field credential must produce 2 map entries")

	userCfg, ok := entries["GITHUB_USERNAME"]
	require.True(t, ok, "map must contain GITHUB_USERNAME key")
	assert.Equal(t, config.SourceVault, userCfg.Source)
	assert.Equal(t, "github-creds", userCfg.VaultSecret)
	assert.Equal(t, "staging", userCfg.VaultEnv)
	assert.Equal(t, "GITHUB_USERNAME", userCfg.VaultField,
		"GITHUB_USERNAME entry MUST have VaultField='GITHUB_USERNAME' so vault "+
			"resolution calls GetField not Get")

	passCfg, ok := entries["GITHUB_PAT"]
	require.True(t, ok, "map must contain GITHUB_PAT key")
	assert.Equal(t, config.SourceVault, passCfg.Source)
	assert.Equal(t, "github-creds", passCfg.VaultSecret)
	assert.Equal(t, "staging", passCfg.VaultEnv)
	assert.Equal(t, "GITHUB_PAT", passCfg.VaultField,
		"GITHUB_PAT entry MUST have VaultField='GITHUB_PAT' so vault "+
			"resolution calls GetField not Get")
}
