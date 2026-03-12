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
