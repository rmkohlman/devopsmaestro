package config

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// === Keychain Dual-Field Config Tests (v0.37.1) ===

// ---------------------------------------------------------------------------
// Section: KeychainField Type Tests
// ---------------------------------------------------------------------------

// TestKeychainField_Constants verifies the string values of each constant.
func TestKeychainField_Constants(t *testing.T) {
	assert.Equal(t, KeychainField("password"), KeychainFieldPassword,
		"KeychainFieldPassword should equal \"password\"")
	assert.Equal(t, KeychainField("account"), KeychainFieldAccount,
		"KeychainFieldAccount should equal \"account\"")
}

// TestKeychainField_ZeroValue verifies that the zero value is distinct from
// both named constants.
func TestKeychainField_ZeroValue(t *testing.T) {
	var zero KeychainField
	assert.NotEqual(t, KeychainFieldPassword, zero,
		"zero value should be distinct from KeychainFieldPassword")
	assert.NotEqual(t, KeychainFieldAccount, zero,
		"zero value should be distinct from KeychainFieldAccount")
}

// ---------------------------------------------------------------------------
// Section: ResolveCredential Field Branching Tests
// ---------------------------------------------------------------------------

// TestResolveCredential_KeychainDefaultField verifies that a CredentialConfig
// with no Field set (zero value) routes through the password path.
// On darwin the call will fail with "not found"; on other platforms it will
// fail with "not available".  Both cases are errors — what matters is that the
// function accepts the config without complaining about missing metadata.
func TestResolveCredential_KeychainDefaultField(t *testing.T) {
	cfg := CredentialConfig{
		Source:  SourceKeychain,
		Service: "test-service-dvm-default-field",
		// Field intentionally omitted (zero value)
	}

	_, err := ResolveCredential(cfg)

	// We always expect an error (no real keychain entry / non-darwin).
	// What we must NOT get is the "service name" validation error, which would
	// indicate the branching code broke before even attempting the lookup.
	assert.Error(t, err, "expected error from keychain lookup (entry won't exist)")
	assert.NotContains(t, err.Error(), "requires service name",
		"branching should not fail on service-name validation for a non-empty service")
}

// TestResolveCredential_KeychainPasswordField verifies that explicitly setting
// Field to KeychainFieldPassword behaves identically to the default (zero-value)
// case — both should route to GetFromKeychain.
func TestResolveCredential_KeychainPasswordField(t *testing.T) {
	cfg := CredentialConfig{
		Source:  SourceKeychain,
		Service: "test-service-dvm-password-field",
		Field:   KeychainFieldPassword,
	}

	_, err := ResolveCredential(cfg)

	assert.Error(t, err, "expected error from keychain lookup (entry won't exist)")
	assert.NotContains(t, err.Error(), "requires service name",
		"explicit KeychainFieldPassword should route through GetFromKeychain without validation failure")
}

// TestResolveCredential_KeychainAccountField verifies that Field ==
// KeychainFieldAccount routes to GetAccountFromKeychain.
// On darwin: expect an item-not-found error.
// On non-darwin: expect a "not available" error.
func TestResolveCredential_KeychainAccountField(t *testing.T) {
	cfg := CredentialConfig{
		Source:  SourceKeychain,
		Service: "test-service-dvm-account-field",
		Field:   KeychainFieldAccount,
	}

	_, err := ResolveCredential(cfg)

	assert.Error(t, err, "expected error from keychain lookup (entry won't exist)")

	if runtime.GOOS == "darwin" {
		// The account lookup should fail with "not found", not a validation error.
		assert.NotContains(t, err.Error(), "requires service name",
			"account field routing should attempt keychain lookup, not fail on validation")
	} else {
		assert.Contains(t, err.Error(), "not available",
			"non-darwin should return 'not available' error")
	}
}

// TestResolveCredential_EnvSourceIgnoresField verifies that the Field
// discriminator is irrelevant when Source == SourceEnv.
func TestResolveCredential_EnvSourceIgnoresField(t *testing.T) {
	t.Setenv("TEST_VAR_XYZ_DVM", "hello")

	cfg := CredentialConfig{
		Source: SourceEnv,
		EnvVar: "TEST_VAR_XYZ_DVM",
		Field:  KeychainFieldAccount, // should be ignored for env source
	}

	result, err := ResolveCredential(cfg)

	assert.NoError(t, err, "env source should resolve without error")
	assert.Equal(t, "hello", result,
		"Field should be ignored for env source; value should come from env var")
}

// ---------------------------------------------------------------------------
// Section: CredentialConfig Field Tests
// ---------------------------------------------------------------------------

// TestCredentialConfig_HasFieldProperty verifies that CredentialConfig has a
// Field property of type KeychainField and that it round-trips correctly.
func TestCredentialConfig_HasFieldProperty(t *testing.T) {
	cfg := CredentialConfig{
		Source:  SourceKeychain,
		Service: "svc",
		Field:   KeychainFieldAccount,
	}

	assert.Equal(t, KeychainFieldAccount, cfg.Field,
		"Field should be stored and retrievable on CredentialConfig")
}

// ---------------------------------------------------------------------------
// Section: GetAccountFromKeychain Tests
// ---------------------------------------------------------------------------

// TestGetAccountFromKeychain_Exists verifies the function signature exists and
// is callable.  We don't assert on the success path because no real keychain
// entry exists; we only care about the type contract.
//
//   - darwin:     expect an error (item not found in Keychain)
//   - non-darwin: expect a "not available" error
func TestGetAccountFromKeychain_Exists(t *testing.T) {
	_, err := GetAccountFromKeychain(KeychainLookup{
		Label:        "nonexistent-service-for-dvm-test",
		KeychainType: KeychainTypeGeneric,
	})

	assert.Error(t, err,
		"GetAccountFromKeychain should return an error for a non-existent service")

	if runtime.GOOS != "darwin" {
		assert.Contains(t, err.Error(), "not available",
			"non-darwin stub should return 'not available' error")
	}

	// On darwin we just confirm an error is returned — the exact message
	// ("not found in Keychain" / exit code 44) is an implementation detail
	// tested elsewhere.
	_ = os.Getenv("USER") // no-op; ensure os package is used (avoids lint warning)
}

// =============================================================================
// TDD Phase 2 (RED): Env Var Fallback Rescue Tests (v0.38.2)
// =============================================================================
// Bug 2: Env var fallback only checks credentials that were already successfully
// resolved (in the `result` map).  If a keychain lookup fails, the credential
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
// Section: ResolveCredentialsWithErrors — env var rescues failed keychain
// ---------------------------------------------------------------------------

// TestResolveCredentialsWithErrors_EnvVarRescuesFailedKeychain verifies that
// when a keychain-sourced credential fails to resolve (service not found) but
// an environment variable with the SAME NAME as the credential is set, the
// credential IS resolved from the env var and does NOT appear in the errors map.
//
// Current behaviour (BUG): the credential stays in errors; the env var is never
// checked for failed credentials.
// Expected behaviour (FIX): the credential is rescued by the env var.
//
// This test WILL FAIL against current code.
func TestResolveCredentialsWithErrors_EnvVarRescuesFailedKeychain(t *testing.T) {
	const credName = "MY_RESCUED_CRED"
	const credValue = "rescued-secret-value"

	// Set an env var matching the credential name so it can rescue the failed
	// keychain lookup.
	t.Setenv(credName, credValue)

	scope := CredentialScope{
		Type: "app",
		ID:   1,
		Name: "test-app",
		Credentials: Credentials{
			credName: CredentialConfig{
				Source:  SourceKeychain,
				Service: "dvm-test-nonexistent-rescue-service-99999",
			},
		},
	}

	resolved, errors := ResolveCredentialsWithErrors(scope)

	// The credential MUST appear in the resolved map (rescued by env var).
	assert.Contains(t, resolved, credName,
		"env var should rescue the failed keychain credential; it must appear in resolved map")

	if val, ok := resolved[credName]; ok {
		assert.Equal(t, credValue, val,
			"rescued credential value should match the env var value")
	}

	// The credential must NOT appear in the errors map once rescued.
	assert.NotContains(t, errors, credName,
		"rescued credential must be removed from the errors map")
}

// TestResolveCredentials_EnvVarRescuesFailedKeychain is the same scenario but
// using the non-WithErrors variant ResolveCredentials.
//
// Current behaviour (BUG): the credential is absent from the result map.
// Expected behaviour (FIX): the credential appears in the result map.
//
// This test WILL FAIL against current code.
func TestResolveCredentials_EnvVarRescuesFailedKeychain(t *testing.T) {
	const credName = "MY_RESCUED_CRED_V2"
	const credValue = "another-rescued-value"

	t.Setenv(credName, credValue)

	scope := CredentialScope{
		Type: "app",
		ID:   1,
		Name: "test-app",
		Credentials: Credentials{
			credName: CredentialConfig{
				Source:  SourceKeychain,
				Service: "dvm-test-nonexistent-rescue-v2-service-99999",
			},
		},
	}

	resolved := ResolveCredentials(scope)

	// The credential MUST appear in the resolved map (rescued by env var).
	assert.Contains(t, resolved, credName,
		"env var should rescue the failed keychain credential in ResolveCredentials")

	if val, ok := resolved[credName]; ok {
		assert.Equal(t, credValue, val,
			"rescued credential value should match the env var value")
	}
}

// TestResolveCredentialsWithErrors_EnvVarRescueClearsError verifies that when
// an env var rescues a previously failed keychain credential, the error entry
// for that credential name is cleared from the errors map.
//
// This is a targeted test for the "clear error on rescue" behaviour.
//
// This test WILL FAIL against current code.
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
				Source:  SourceKeychain,
				Service: "dvm-test-nonexistent-clear-error-service-99999",
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
// that a credential that fails keychain lookup AND has no matching env var set
// DOES remain in the errors map (the rescue must not accidentally clear all
// errors — only the ones that are rescued).
//
// This test verifies the correct side of the fix and should PASS both before
// and after the bug fix.  It is included here as a regression guard.
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
				Source:  SourceKeychain,
				Service: "dvm-test-nonexistent-unrescued-service-99999",
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
//
// This test WILL FAIL against current code (the rescued credential will be
// absent from the result map).
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
				Source:  SourceKeychain,
				Service: "dvm-test-nonexistent-multi-rescued-99999",
			},
			unrescuedCred: CredentialConfig{
				Source:  SourceKeychain,
				Service: "dvm-test-nonexistent-multi-unrescued-99999",
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
// TDD Phase 2 (RED): ResolveCredential Label-Based Lookup Tests (v0.39.0)
// =============================================================================
// Design change: CredentialConfig gains a Label field (replaces Service for
// keychain lookups) and a KeychainType field. ResolveCredential must use Label
// (not Service) when calling keychain functions.
//
// New CredentialConfig fields that MUST exist after implementation:
//
//	type CredentialConfig struct {
//	    ...existing fields...
//	    Label        string       `yaml:"label,omitempty"`
//	    KeychainType KeychainType `yaml:"keychainType,omitempty"`
//	}
//
// New KeychainType type and constants (in keychain_darwin.go / keychain_stub.go):
//
//	type KeychainType string
//	const (
//	    KeychainTypeGeneric  KeychainType = "generic"
//	    KeychainTypeInternet KeychainType = "internet"
//	)
//
// ALL tests in this section WILL FAIL TO COMPILE until the above fields and
// types are added to config/credentials.go and config/keychain_*.go.
// =============================================================================

// ---------------------------------------------------------------------------
// Section: CredentialConfig Label Field Tests
// ---------------------------------------------------------------------------

// TestCredentialConfig_HasLabelProperty verifies that CredentialConfig has a
// Label field of type string.
//
// WILL FAIL TO COMPILE — CredentialConfig.Label does not exist yet.
func TestCredentialConfig_HasLabelProperty(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// CredentialConfig.Label field does not exist yet.
	cfg := CredentialConfig{
		Source: SourceKeychain,
		Label:  "com.example.my-label",
	}
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, "com.example.my-label", cfg.Label,
		"Label should be stored and retrievable on CredentialConfig")
	assert.Empty(t, cfg.Service,
		"Service should be empty when only Label is set")
}

// TestCredentialConfig_HasKeychainTypeProperty verifies that CredentialConfig
// has a KeychainType field.
//
// WILL FAIL TO COMPILE — CredentialConfig.KeychainType does not exist yet.
func TestCredentialConfig_HasKeychainTypeProperty(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// CredentialConfig.KeychainType and KeychainTypeGeneric do not exist yet.
	cfg := CredentialConfig{
		Source:       SourceKeychain,
		Label:        "com.example.label",
		KeychainType: KeychainTypeGeneric,
	}
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, KeychainTypeGeneric, cfg.KeychainType,
		"KeychainType should be stored and retrievable on CredentialConfig")
}

// ---------------------------------------------------------------------------
// Section: ResolveCredential Label-Based Routing Tests
// ---------------------------------------------------------------------------

// TestResolveCredential_UsesLabel verifies that ResolveCredential uses the
// Label field (not Service) when calling keychain functions. When Label is
// set, GetFromKeychain should be called with a KeychainLookup{Label: cfg.Label}.
//
// We test this indirectly: a non-existent label must produce an error that
// does NOT mention the old "requires service name" validation message (which
// would indicate it fell back to the old Service-based path).
//
// WILL FAIL TO COMPILE — CredentialConfig.Label does not exist yet.
func TestResolveCredential_UsesLabel(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// CredentialConfig.Label field does not exist yet.
	cfg := CredentialConfig{
		Source: SourceKeychain,
		Label:  "dvm-test-label-resolve-nonexistent-99999",
	}
	// ─────────────────────────────────────────────────────────────────────────

	_, err := ResolveCredential(cfg)

	// We always expect an error (no real keychain entry).
	assert.Error(t, err, "expected error from keychain lookup (entry won't exist)")

	// Must NOT get a validation error about missing label/service — the Label
	// field was provided, so the lookup should proceed to the keychain.
	assert.NotContains(t, err.Error(), "requires service name",
		"Label field should satisfy the keychain source requirement")
	assert.NotContains(t, err.Error(), "requires label",
		"Label field should satisfy the keychain source requirement")
}

// TestResolveCredential_UsesKeychainType verifies that ResolveCredential passes
// the KeychainType through to the underlying keychain function. When
// KeychainType is "internet", find-internet-password should be called;
// when "generic" (or empty), find-generic-password should be called.
//
// We verify indirectly via the error path: both types must produce "not found"
// errors (not validation errors) when the label doesn't exist.
//
// WILL FAIL TO COMPILE — CredentialConfig.Label and CredentialConfig.KeychainType
// do not exist yet.
func TestResolveCredential_UsesKeychainType(t *testing.T) {
	tests := []struct {
		name         string
		keychainType KeychainType
	}{
		{
			name:         "generic type",
			keychainType: KeychainTypeGeneric,
		},
		{
			name:         "internet type",
			keychainType: KeychainTypeInternet,
		},
		{
			name:         "empty type (defaults to generic)",
			keychainType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────
			// CredentialConfig.Label, CredentialConfig.KeychainType,
			// KeychainTypeGeneric, KeychainTypeInternet do not exist yet.
			cfg := CredentialConfig{
				Source:       SourceKeychain,
				Label:        "dvm-test-ktype-resolve-nonexistent-99999",
				KeychainType: tt.keychainType,
			}
			// ─────────────────────────────────────────────────────────────────

			_, err := ResolveCredential(cfg)

			assert.Error(t, err,
				"expected error from keychain lookup for type %q (entry won't exist)",
				tt.keychainType)
			// The error should be a "not found" error from the keychain,
			// not a validation error about the type value.
			assert.NotContains(t, err.Error(), "invalid keychain type",
				"valid KeychainType %q should not fail validation", tt.keychainType)
		})
	}
}

// TestResolveCredential_LabelFallbackToService verifies that when Label is
// empty but Service is set, ResolveCredential falls back to using Service for
// the keychain lookup (backwards compatibility).
//
// This tests the transition period where old configs still use Service.
//
// WILL FAIL AT RUNTIME — ResolveCredential currently requires cfg.Service and
// calls GetFromKeychain(cfg.Service). After the redesign, it should prefer
// cfg.Label but accept cfg.Service as a fallback.
//
// Note: This test passes today ONLY because the existing implementation uses
// Service. After the redesign, the test must CONTINUE to pass — verifying
// the fallback is preserved.
func TestResolveCredential_LabelFallbackToService(t *testing.T) {
	// Use the existing Service field (no Label set) — must still work
	cfg := CredentialConfig{
		Source:  SourceKeychain,
		Service: "dvm-test-service-fallback-nonexistent-99999",
		// Label intentionally empty — should fall back to Service
	}

	_, err := ResolveCredential(cfg)

	// We always expect an error (no real keychain entry exists).
	assert.Error(t, err, "expected error from keychain lookup (entry won't exist)")

	// Must NOT get a "requires service name" / "requires label" validation error.
	// The Service field must still satisfy the requirement.
	assert.NotContains(t, err.Error(), "requires service name",
		"Service field must still work as fallback when Label is empty")
	assert.NotContains(t, err.Error(), "requires label",
		"Service field must still work as fallback when Label is empty")
}

// TestResolveCredential_RequiresLabelOrService verifies that ResolveCredential
// returns a validation error when BOTH Label and Service are empty for a
// keychain source credential.
//
// WILL FAIL AT RUNTIME — the current error message says "requires service name"
// but after the redesign it should mention label/service. This test accepts
// either phrasing to be flexible.
func TestResolveCredential_RequiresLabelOrService(t *testing.T) {
	cfg := CredentialConfig{
		Source: SourceKeychain,
		// Both Label and Service are empty — must be rejected
	}

	_, err := ResolveCredential(cfg)

	assert.Error(t, err,
		"keychain source with neither Label nor Service should return validation error")

	// Accept any phrasing that communicates the missing identifier
	hasRelevantMsg := false
	if err != nil {
		msg := err.Error()
		hasRelevantMsg = contains(msg, "requires service") ||
			contains(msg, "requires label") ||
			contains(msg, "label or service") ||
			contains(msg, "service name")
	}
	assert.True(t, hasRelevantMsg,
		"error message should indicate that a label or service name is required")
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
