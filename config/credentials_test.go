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
	_, err := GetAccountFromKeychain("nonexistent-service-for-dvm-test")

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
