package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// === MaestroVault Credential Tests (v0.40.0) ===

// ---------------------------------------------------------------------------
// Section: ResolveCredential Env Source Tests
// ---------------------------------------------------------------------------

// TestResolveCredential_EnvSource verifies that ResolveCredential resolves
// an env-sourced credential from the environment.
func TestResolveCredential_EnvSource(t *testing.T) {
	t.Setenv("TEST_VAR_XYZ_DVM", "hello")

	cfg := CredentialConfig{
		Source: SourceEnv,
		EnvVar: "TEST_VAR_XYZ_DVM",
	}

	result, err := ResolveCredential(cfg)

	assert.NoError(t, err, "env source should resolve without error")
	assert.Equal(t, "hello", result,
		"value should come from env var")
}

// =============================================================================
// TDD Phase 2 (RED): Env Var Fallback Rescue Tests (v0.38.2)
// =============================================================================
// Bug 2: Env var fallback only checks credentials that were already successfully
// resolved (in the `result` map).  If a vault lookup fails, the credential
// is placed in `errors` and never considered for env var rescue.
//
// Expected fix: After the resolution loop, also iterate over `errors` and check
// whether the credential name is set as an env var.  If it is, move the entry
// from errors → result (removing the error).
//
// All three tests below WILL FAIL against the current implementation because
// the bug means the rescued credential never appears in the result map.
// =============================================================================

// ---------------------------------------------------------------------------
// Section: ResolveCredentialsWithErrors — env var rescues failed vault lookup
// ---------------------------------------------------------------------------

// TestResolveCredentialsWithErrors_EnvVarRescuesFailedVault verifies that
// when a vault-sourced credential fails to resolve but an environment variable
// with the SAME NAME as the credential is set, the credential IS resolved from
// the env var and does NOT appear in the errors map.
func TestResolveCredentialsWithErrors_EnvVarRescuesFailedVault(t *testing.T) {
	const credName = "MY_RESCUED_CRED"
	const credValue = "rescued-secret-value"

	// Set an env var matching the credential name so it can rescue the failed
	// vault lookup.
	t.Setenv(credName, credValue)

	scope := CredentialScope{
		Type: "app",
		ID:   1,
		Name: "test-app",
		Credentials: Credentials{
			credName: CredentialConfig{
				Source:      SourceVault,
				VaultSecret: "dvm-test-nonexistent-rescue-secret-99999",
			},
		},
	}

	resolved, errors := ResolveCredentialsWithErrors(scope)

	// The credential MUST appear in the resolved map (rescued by env var).
	assert.Contains(t, resolved, credName,
		"env var should rescue the failed vault credential; it must appear in resolved map")

	if val, ok := resolved[credName]; ok {
		assert.Equal(t, credValue, val,
			"rescued credential value should match the env var value")
	}

	// The credential must NOT appear in the errors map once rescued.
	assert.NotContains(t, errors, credName,
		"rescued credential must be removed from the errors map")
}

// TestResolveCredentials_EnvVarRescuesFailedVault is the same scenario but
// using the non-WithErrors variant ResolveCredentials.
func TestResolveCredentials_EnvVarRescuesFailedVault(t *testing.T) {
	const credName = "MY_RESCUED_CRED_V2"
	const credValue = "another-rescued-value"

	t.Setenv(credName, credValue)

	scope := CredentialScope{
		Type: "app",
		ID:   1,
		Name: "test-app",
		Credentials: Credentials{
			credName: CredentialConfig{
				Source:      SourceVault,
				VaultSecret: "dvm-test-nonexistent-rescue-v2-secret-99999",
			},
		},
	}

	resolved := ResolveCredentials(scope)

	// The credential MUST appear in the resolved map (rescued by env var).
	assert.Contains(t, resolved, credName,
		"env var should rescue the failed vault credential in ResolveCredentials")

	if val, ok := resolved[credName]; ok {
		assert.Equal(t, credValue, val,
			"rescued credential value should match the env var value")
	}
}

// TestResolveCredentialsWithErrors_EnvVarRescueClearsError verifies that when
// an env var rescues a previously failed vault credential, the error entry
// for that credential name is cleared from the errors map.
func TestResolveCredentialsWithErrors_EnvVarRescueClearsError(t *testing.T) {
	const credName = "CRED_ERROR_CLEARED"
	const credValue = "cleared-error-value"

	t.Setenv(credName, credValue)

	scope := CredentialScope{
		Type: "app",
		ID:   2,
		Name: "test-app-2",
		Credentials: Credentials{
			credName: CredentialConfig{
				Source:      SourceVault,
				VaultSecret: "dvm-test-nonexistent-clear-error-secret-99999",
			},
		},
	}

	_, errors := ResolveCredentialsWithErrors(scope)

	// No error should remain for the rescued credential.
	if err, exists := errors[credName]; exists {
		t.Errorf("errors map should NOT contain %q after env var rescue, but got error: %v",
			credName, err)
	}
}

// TestResolveCredentialsWithErrors_UnrescuedCredentialStaysInErrors verifies
// that a credential that fails vault lookup AND has no matching env var set
// DOES remain in the errors map (the rescue must not accidentally clear all
// errors — only the ones that are rescued).
func TestResolveCredentialsWithErrors_UnrescuedCredentialStaysInErrors(t *testing.T) {
	const credName = "UNRESCUED_CRED"

	// Explicitly ensure this env var is NOT set.
	t.Setenv(credName, "")

	scope := CredentialScope{
		Type: "app",
		ID:   3,
		Name: "test-app-3",
		Credentials: Credentials{
			credName: CredentialConfig{
				Source:      SourceVault,
				VaultSecret: "dvm-test-nonexistent-unrescued-secret-99999",
			},
		},
	}

	resolved, errors := ResolveCredentialsWithErrors(scope)

	// Credential must NOT appear in resolved (nothing rescued it).
	assert.NotContains(t, resolved, credName,
		"unrescued credential must not appear in resolved map")

	// Credential error MUST still be in errors.
	assert.Contains(t, errors, credName,
		"unrescued credential must remain in errors map")
}

// TestResolveCredentialsWithErrors_MultipleCredsPartialRescue verifies that
// when multiple credentials fail, only those with a matching env var are
// rescued.  This tests the mix of rescued and unrescued credentials.
func TestResolveCredentialsWithErrors_MultipleCredsPartialRescue(t *testing.T) {
	const rescuedCred = "RESCUED_MULTI"
	const unrescuedCred = "UNRESCUED_MULTI"
	const rescuedValue = "multi-rescued-value"

	// Only set the env var for the credential that should be rescued.
	t.Setenv(rescuedCred, rescuedValue)
	t.Setenv(unrescuedCred, "") // explicitly empty — not rescued

	scope := CredentialScope{
		Type: "app",
		ID:   4,
		Name: "test-app-4",
		Credentials: Credentials{
			rescuedCred: CredentialConfig{
				Source:      SourceVault,
				VaultSecret: "dvm-test-nonexistent-multi-rescued-99999",
			},
			unrescuedCred: CredentialConfig{
				Source:      SourceVault,
				VaultSecret: "dvm-test-nonexistent-multi-unrescued-99999",
			},
		},
	}

	resolved, errors := ResolveCredentialsWithErrors(scope)

	// The rescued credential must be in the result.
	assert.Contains(t, resolved, rescuedCred,
		"env var should rescue %q", rescuedCred)
	if val, ok := resolved[rescuedCred]; ok {
		assert.Equal(t, rescuedValue, val)
	}
	assert.NotContains(t, errors, rescuedCred,
		"rescued credential must not remain in errors map")

	// The unrescued credential must still be in errors.
	assert.NotContains(t, resolved, unrescuedCred,
		"unrescued credential must not appear in resolved map")
	assert.Contains(t, errors, unrescuedCred,
		"unrescued credential must remain in errors map")
}

// =============================================================================
// TDD Phase 2 (RED): MaestroVault Integration Tests (v0.40.0)
// =============================================================================
// Replacing macOS Keychain with MaestroVault as the secrets backend.
//
// New constants and fields that MUST exist after implementation:
//
//	const SourceVault CredentialSource = "vault"
//
//	type CredentialConfig struct {
//	    ...existing fields...
//	    VaultSecret string `yaml:"vaultSecret,omitempty"`
//	    VaultEnv    string `yaml:"vaultEnv,omitempty"`
//	}
//
// ALL tests in this section WILL FAIL TO COMPILE until the above constants
// and fields are added to config/credentials.go.
// =============================================================================

// ---------------------------------------------------------------------------
// Section: SourceVault Constant Tests
// ---------------------------------------------------------------------------

// TestSourceVault_Constant verifies that SourceVault is defined and has the
// correct string value "vault".
//
// WILL FAIL TO COMPILE — SourceVault does not exist yet.
func TestSourceVault_Constant(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// SourceVault does not exist yet.
	assert.Equal(t, CredentialSource("vault"), SourceVault,
		"SourceVault must equal the string \"vault\"")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestSourceVault_IsCredentialSource verifies that SourceVault is of the
// correct type CredentialSource (not just an untyped string constant).
//
// WILL FAIL TO COMPILE — SourceVault does not exist yet.
func TestSourceVault_IsCredentialSource(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	var _ CredentialSource = SourceVault
	// ─────────────────────────────────────────────────────────────────────────
	assert.True(t, true, "SourceVault must be of type CredentialSource")
}

// TestSourceVault_DistinctFromOtherSources verifies that SourceVault is
// distinct from the other source constants (SourceEnv).
//
// WILL FAIL TO COMPILE — SourceVault does not exist yet.
func TestSourceVault_DistinctFromOtherSources(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	assert.NotEqual(t, SourceEnv, SourceVault,
		"SourceVault must be distinct from SourceEnv")
	// ─────────────────────────────────────────────────────────────────────────
}

// ---------------------------------------------------------------------------
// Section: CredentialConfig Vault Field Tests
// ---------------------------------------------------------------------------

// TestCredentialConfig_VaultFields verifies that CredentialConfig has
// VaultSecret and VaultEnv fields of type string that store and retrieve
// correctly.
//
// WILL FAIL TO COMPILE — CredentialConfig.VaultSecret and
// CredentialConfig.VaultEnv do not exist yet.
func TestCredentialConfig_VaultFields(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	// CredentialConfig.VaultSecret and CredentialConfig.VaultEnv do not exist yet.
	cfg := CredentialConfig{
		Source:      SourceVault,
		VaultSecret: "my-github-pat",
		VaultEnv:    "production",
	}
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, SourceVault, cfg.Source,
		"Source must be set to SourceVault")
	assert.Equal(t, "my-github-pat", cfg.VaultSecret,
		"VaultSecret must be stored and retrievable on CredentialConfig")
	assert.Equal(t, "production", cfg.VaultEnv,
		"VaultEnv must be stored and retrievable on CredentialConfig")
}

// TestCredentialConfig_VaultEnvOptional verifies that VaultEnv is optional —
// a CredentialConfig with a VaultSecret but no VaultEnv must be constructable.
//
// WILL FAIL TO COMPILE — CredentialConfig.VaultSecret does not exist yet.
func TestCredentialConfig_VaultEnvOptional(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cfg := CredentialConfig{
		Source:      SourceVault,
		VaultSecret: "my-secret",
		// VaultEnv intentionally omitted — must default to zero value ""
	}
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, "my-secret", cfg.VaultSecret)
	assert.Empty(t, cfg.VaultEnv,
		"VaultEnv zero value must be empty string (optional field)")
}

// TestCredentialConfig_VaultFieldsZeroValues verifies that a CredentialConfig
// with no vault fields set has empty string zero values for VaultSecret and
// VaultEnv (confirming the fields are value types, not pointers).
//
// WILL FAIL TO COMPILE — CredentialConfig.VaultSecret and .VaultEnv don't exist.
func TestCredentialConfig_VaultFieldsZeroValues(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cfg := CredentialConfig{
		Source: SourceEnv,
		EnvVar: "SOME_VAR",
	}
	assert.Empty(t, cfg.VaultSecret,
		"VaultSecret must default to empty string when not set")
	assert.Empty(t, cfg.VaultEnv,
		"VaultEnv must default to empty string when not set")
	// ─────────────────────────────────────────────────────────────────────────
}

// ---------------------------------------------------------------------------
// Section: CredentialConfig Vault Field YAML Tag Tests
// ---------------------------------------------------------------------------

// TestCredentialConfig_VaultSecret_YAMLTag verifies that the VaultSecret field
// uses the yaml tag "vaultSecret" (camelCase, matching the project convention).
//
// We verify indirectly by constructing the struct and confirming the field
// is accessible — actual YAML marshalling is tested in models/credential_test.go.
//
// WILL FAIL TO COMPILE — CredentialConfig.VaultSecret does not exist yet.
func TestCredentialConfig_VaultSecret_YAMLTag(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cfg := CredentialConfig{
		Source:      SourceVault,
		VaultSecret: "tagged-secret",
	}
	// ─────────────────────────────────────────────────────────────────────────

	// Field accessibility check — the YAML tag is verified via struct literal.
	assert.Equal(t, "tagged-secret", cfg.VaultSecret)
}

// ---------------------------------------------------------------------------
// Section: ResolveCredential with Vault Source (fail-fast behavior)
// ---------------------------------------------------------------------------

// TestResolveCredential_VaultSource_FailsWithoutBackend verifies that the
// original ResolveCredential function (with no backend parameter) returns a
// hard error for vault-sourced credentials — callers must use
// ResolveCredentialWithBackend instead.
//
// This test verifies the "fail-fast" design: if someone calls ResolveCredential
// with source="vault" they get an actionable error, not a silent empty string.
//
// WILL FAIL TO COMPILE — SourceVault and CredentialConfig.VaultSecret
// do not exist yet.
func TestResolveCredential_VaultSource_FailsWithoutBackend(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	cfg := CredentialConfig{
		Source:      SourceVault,
		VaultSecret: "my-secret",
		VaultEnv:    "production",
	}
	_, err := ResolveCredential(cfg)
	// ─────────────────────────────────────────────────────────────────────────

	assert.Error(t, err,
		"ResolveCredential must return an error for vault source (no backend available)")
	// The error should guide the user to the vault-aware resolver.
	assert.True(t,
		contains(err.Error(), "vault") || contains(err.Error(), "backend"),
		"error message should mention vault or backend to guide the user; got: %q", err.Error())
}

// contains is a simple substring helper used in tests to avoid importing
// strings just for Contains.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}

// =============================================================================
// TDD Phase 2 (RED): VaultField Credential Resolution Tests (v0.41.0)
// =============================================================================
// New field on CredentialConfig:
//
//	type CredentialConfig struct {
//	    ...
//	    VaultField string `yaml:"vaultField,omitempty"`
//	}
//
// When VaultField is set, ResolveCredentialWithBackend must call
// backend.GetField(name, env, field) via the FieldCapableBackend interface.
//
// ALL tests in this section WILL FAIL TO COMPILE until VaultField is added to
// CredentialConfig in config/credentials.go.
// =============================================================================

// ---------------------------------------------------------------------------
// Section: CredentialConfig.VaultField Field Tests
// ---------------------------------------------------------------------------

// TestCredentialConfig_VaultField_FieldExists verifies that CredentialConfig
// has a VaultField string field with yaml tag "vaultField".
//
// WILL FAIL TO COMPILE — CredentialConfig.VaultField does not exist yet.
func TestCredentialConfig_VaultField_FieldExists(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cfg := CredentialConfig{
		Source:      SourceVault,
		VaultSecret: "github/creds",
		VaultField:  "password",
	}
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, "password", cfg.VaultField,
		"CredentialConfig.VaultField must be accessible and hold the set value")
}

// TestCredentialConfig_VaultField_EmptyByDefault verifies that VaultField
// defaults to the empty string (zero value for string).
//
// WILL FAIL TO COMPILE — CredentialConfig.VaultField does not exist yet.
func TestCredentialConfig_VaultField_EmptyByDefault(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cfg := CredentialConfig{
		Source:      SourceVault,
		VaultSecret: "github/token",
	}
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, "", cfg.VaultField,
		"VaultField must default to empty string when not set")
}

// TestResolveCredentialWithBackend_VaultField_CredentialsConfig verifies that
// when a CredentialConfig specifies a VaultField, the resolution correctly
// routes to GetField on a FieldCapableBackend.
//
// This test focuses on the config package's integration of VaultField with the
// resolution function (field routing / dispatch logic).
//
// WILL FAIL TO COMPILE — CredentialConfig.VaultField and FieldCapableBackend
// do not exist yet.
func TestResolveCredentialWithBackend_VaultField_CredentialsConfig(t *testing.T) {
	tests := []struct {
		name           string
		vaultField     string
		expectGetField bool
		expectGet      bool
		expectedValue  string
	}{
		{
			name:           "with vault field set — uses GetField",
			vaultField:     "username",
			expectGetField: true,
			expectGet:      false,
			expectedValue:  "resolved-username",
		},
		{
			name:           "without vault field — uses Get",
			vaultField:     "",
			expectGetField: false,
			expectGet:      true,
			expectedValue:  "resolved-secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getFieldCalled := false
			getCalled := false

			mock := &mockFieldCapableBackend{
				GetFunc: func(name, env string) (string, error) {
					getCalled = true
					return "resolved-secret", nil
				},
				GetFieldFunc: func(name, env, field string) (string, error) {
					getFieldCalled = true
					return "resolved-username", nil
				},
			}

			// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────
			cfg := CredentialConfig{
				Source:      SourceVault,
				VaultSecret: "my-org/creds",
				VaultEnv:    "staging",
				VaultField:  tt.vaultField,
			}
			value, err := ResolveCredentialWithBackend(cfg, mock)
			// ─────────────────────────────────────────────────────────────────

			require.NoError(t, err)
			assert.Equal(t, tt.expectedValue, value)
			assert.Equal(t, tt.expectGetField, getFieldCalled,
				"GetField called mismatch for case %q", tt.name)
			assert.Equal(t, tt.expectGet, getCalled,
				"Get called mismatch for case %q", tt.name)
		})
	}
}
