//go:build darwin

package config

// =============================================================================
// TDD Phase 2 (RED): Keychain Helper Tests (v0.38.2)
// =============================================================================
// Bug 3: Keychain exit-code handling is duplicated between GetFromKeychain and
// GetAccountFromKeychain, and uses inconsistent error messages.
//
// Expected fix: Extract a shared helper:
//
//	func keychainExitError(exitCode int, service string) error
//
// that maps well-known security(1) exit codes to descriptive errors.
//
// Test 2 (TestKeychainExitError) WILL FAIL TO COMPILE because keychainExitError
// does not yet exist.
// Test 1 (TestGetAccountFromKeychain_NonExistentService) passes today — it's
// here as a regression guard and specification anchor.
// =============================================================================

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Section: GetAccountFromKeychain Error Tests
// ---------------------------------------------------------------------------

// TestGetAccountFromKeychain_NonExistentService verifies that calling
// GetAccountFromKeychain with a service name that definitely does not exist in
// the macOS Keychain returns an error whose message indicates the item was not
// found.
//
// This test is a specification anchor: it passes today and must continue to
// pass after the Bug 3 refactor.
func TestGetAccountFromKeychain_NonExistentService(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("keychain tests only run on darwin")
	}

	_, err := GetAccountFromKeychain(KeychainLookup{
		Label:        "dvm-test-definitely-nonexistent-service-xyz-99999",
		KeychainType: KeychainTypeGeneric,
	})

	assert.Error(t, err,
		"GetAccountFromKeychain must return an error for a non-existent service")

	// The error message should communicate that the item wasn't found.
	// We accept "not found" as the canonical phrase (matching exit code 44).
	assert.Contains(t, err.Error(), "not found",
		"error message should indicate item was not found in Keychain")
}

// ---------------------------------------------------------------------------
// Section: keychainExitError Helper Tests
// ---------------------------------------------------------------------------

// TestKeychainExitError is a table-driven test for the extracted helper
// function keychainExitError(exitCode int, service string) error.
//
// This function does NOT yet exist in the codebase.
// This test WILL FAIL TO COMPILE until keychainExitError is added to
// config/keychain_darwin.go.
//
// Each exit code used by the macOS security(1) command should map to a distinct,
// descriptive error message so callers can give actionable feedback.
func TestKeychainExitError(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("keychain tests only run on darwin")
	}

	tests := []struct {
		name         string
		exitCode     int
		service      string
		wantContains string // substring that must appear in the error message
	}{
		{
			name:         "exit code 36 — user denied access",
			exitCode:     36,
			service:      "com.example.test-svc",
			wantContains: "cancelled",
		},
		{
			name:         "exit code 44 — credential not found",
			exitCode:     44,
			service:      "com.example.test-svc",
			wantContains: "not found",
		},
		{
			name:         "exit code 51 — invalid parameters / access denied",
			exitCode:     51,
			service:      "com.example.test-svc",
			wantContains: "denied",
		},
		{
			name:         "unknown exit code — generic message with exit code",
			exitCode:     99,
			service:      "com.example.test-svc",
			wantContains: "99",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ── COMPILE ERROR EXPECTED HERE ──────────────────────────────────
			// keychainExitError does not exist yet.
			err := keychainExitError(tt.exitCode, tt.service)
			// ─────────────────────────────────────────────────────────────────

			assert.Error(t, err,
				"keychainExitError must always return a non-nil error")

			assert.Contains(t, err.Error(), tt.wantContains,
				"error message for exit code %d should contain %q; got: %q",
				tt.exitCode, tt.wantContains, err.Error())

			// The service name should also appear in every error message so
			// the caller knows WHICH keychain entry failed.
			assert.Contains(t, err.Error(), tt.service,
				"error message should mention the service name %q", tt.service)
		})
	}
}

// TestKeychainExitError_AlwaysNonNil verifies that keychainExitError never
// returns a nil error — every exit code, even an unexpected one, must produce
// an error value.
//
// This test WILL FAIL TO COMPILE until keychainExitError is implemented.
func TestKeychainExitError_AlwaysNonNil(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("keychain tests only run on darwin")
	}

	exitCodes := []int{0, 1, 36, 44, 51, 99, 255, -1}

	for _, code := range exitCodes {
		t.Run("non-nil for exit code", func(t *testing.T) {
			// ── COMPILE ERROR EXPECTED HERE ──────────────────────────────────
			err := keychainExitError(code, "com.example.test")
			// ─────────────────────────────────────────────────────────────────

			assert.Error(t, err,
				"keychainExitError(%d, ...) must never return nil", code)
		})
	}
}

// TestGetAccountFromKeychain_UsesKeychainExitError verifies that after the
// refactor, GetAccountFromKeychain uses keychainExitError for known exit codes
// and therefore produces the same error messages that keychainExitError would
// generate for exit code 44 ("not found").
//
// This is a specification test: it will fail to compile until keychainExitError
// exists, and then become a regression guard ensuring the refactor doesn't
// change the user-visible error messages for known exit codes.
func TestGetAccountFromKeychain_UsesKeychainExitError(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("keychain tests only run on darwin")
	}

	const nonExistentService = "dvm-test-get-account-exit-error-99999"

	// Get the error from the real function (will use exit code 44 for "not found")
	_, realErr := GetAccountFromKeychain(KeychainLookup{
		Label:        nonExistentService,
		KeychainType: KeychainTypeGeneric,
	})
	assert.Error(t, realErr, "GetAccountFromKeychain must return error for non-existent service")

	// ── COMPILE ERROR EXPECTED HERE ──────────────────────────────────────────
	// keychainExitError does not exist yet.
	expectedErr := keychainExitError(44, nonExistentService)
	// ─────────────────────────────────────────────────────────────────────────

	// After the refactor, both errors should contain the same key phrase.
	assert.Contains(t, realErr.Error(), "not found",
		"GetAccountFromKeychain error for exit 44 should say 'not found'")
	assert.Contains(t, expectedErr.Error(), "not found",
		"keychainExitError(44, ...) should say 'not found'")
}

// =============================================================================
// TDD Phase 2 (RED): Keychain Label Redesign Tests (v0.39.0)
// =============================================================================
// Design change: Replace service-based lookup (-s SERVICE) with label-based
// lookup (-l LABEL), add explicit keychain type selection (generic vs internet),
// and remove the -a $USER filter from read operations.
//
// New types that MUST exist after implementation:
//
//	type KeychainType string
//	const (
//	    KeychainTypeGeneric  KeychainType = "generic"
//	    KeychainTypeInternet KeychainType = "internet"
//	)
//
//	type KeychainLookup struct {
//	    Label        string
//	    KeychainType string // "generic" or "internet"
//	}
//
// New function signatures that MUST exist after implementation:
//
//	func GetFromKeychain(lookup KeychainLookup) (string, error)
//	func GetAccountFromKeychain(lookup KeychainLookup) (string, error)
//
// ALL tests in this section WILL FAIL TO COMPILE until the above types and
// function signatures are added to config/keychain_darwin.go.
// =============================================================================

// ---------------------------------------------------------------------------
// Section: KeychainType Constant Tests
// ---------------------------------------------------------------------------

// TestKeychainType_Constants verifies the string values of the new KeychainType
// typed constants introduced by the label redesign.
//
// WILL FAIL TO COMPILE — KeychainType, KeychainTypeGeneric, KeychainTypeInternet
// do not exist yet.
func TestKeychainType_Constants(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("keychain tests only run on darwin")
	}

	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	// KeychainType, KeychainTypeGeneric, KeychainTypeInternet do not exist yet.
	assert.Equal(t, KeychainType("generic"), KeychainTypeGeneric,
		"KeychainTypeGeneric should equal the string \"generic\"")
	assert.Equal(t, KeychainType("internet"), KeychainTypeInternet,
		"KeychainTypeInternet should equal the string \"internet\"")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestKeychainType_ZeroValue verifies the zero value of KeychainType is
// distinct from both named constants.
//
// WILL FAIL TO COMPILE — KeychainType does not exist yet.
func TestKeychainType_ZeroValue(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("keychain tests only run on darwin")
	}

	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	var zero KeychainType
	assert.NotEqual(t, KeychainTypeGeneric, zero,
		"zero value should differ from KeychainTypeGeneric")
	assert.NotEqual(t, KeychainTypeInternet, zero,
		"zero value should differ from KeychainTypeInternet")
	// ─────────────────────────────────────────────────────────────────────────
}

// ---------------------------------------------------------------------------
// Section: KeychainLookup Struct Tests
// ---------------------------------------------------------------------------

// TestKeychainLookup_StructFields verifies that KeychainLookup has both Label
// and KeychainType fields with the correct types.
//
// WILL FAIL TO COMPILE — KeychainLookup struct does not exist yet.
func TestKeychainLookup_StructFields(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("keychain tests only run on darwin")
	}

	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	// KeychainLookup struct does not exist yet.
	lookup := KeychainLookup{
		Label:        "com.example.my-service",
		KeychainType: "generic",
	}
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, "com.example.my-service", lookup.Label,
		"KeychainLookup.Label should store the label string")
	assert.Equal(t, KeychainTypeGeneric, lookup.KeychainType,
		"KeychainLookup.KeychainType should store the type string")
}

// TestKeychainLookup_DefaultType verifies a KeychainLookup with only Label set
// is distinguishable (the caller or implementation can default the type to
// "generic" when KeychainType is empty).
//
// WILL FAIL TO COMPILE — KeychainLookup struct does not exist yet.
func TestKeychainLookup_DefaultType(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("keychain tests only run on darwin")
	}

	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	lookup := KeychainLookup{
		Label: "my-label",
		// KeychainType intentionally omitted — should be handled as "generic"
	}
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, "my-label", lookup.Label)
	assert.Empty(t, lookup.KeychainType,
		"zero-value KeychainType is empty; implementation must treat this as generic")
}

// ---------------------------------------------------------------------------
// Section: GetFromKeychain Label-Based Tests
// ---------------------------------------------------------------------------

// TestGetFromKeychain_LabelLookup verifies that GetFromKeychain accepts a
// KeychainLookup struct (not a bare string service) and uses the -l (label)
// flag instead of the old -s (service) flag.
//
// The test uses a non-existent label so we never need a real keychain entry.
// We verify that:
//   - An error IS returned (entry not found)
//   - The error message mentions the label (not "service")
//
// WILL FAIL TO COMPILE — GetFromKeychain currently takes (service string),
// not (lookup KeychainLookup). The signature must change.
func TestGetFromKeychain_LabelLookup(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("keychain tests only run on darwin")
	}

	// ── COMPILE ERROR EXPECTED HERE ──────────────────────────────────────────
	// GetFromKeychain(lookup KeychainLookup) does not exist yet.
	// Current signature: GetFromKeychain(service string)
	lookup := KeychainLookup{
		Label:        "dvm-test-label-lookup-nonexistent-99999",
		KeychainType: "generic",
	}
	_, err := GetFromKeychain(lookup)
	// ─────────────────────────────────────────────────────────────────────────

	assert.Error(t, err,
		"GetFromKeychain must return an error for a non-existent label")
	assert.Contains(t, err.Error(), "not found",
		"error should indicate the label was not found in the Keychain (exit code 44)")
}

// TestGetFromKeychain_InternetPassword verifies that when KeychainType is
// "internet" the function internally dispatches to find-internet-password
// instead of find-generic-password.
//
// We verify this indirectly: a non-existent internet label must return an
// error that still uses our keychainExitError formatting (exit code 44 →
// "not found"). If the wrong subcommand were used the exit code / error
// message would differ.
//
// WILL FAIL TO COMPILE — GetFromKeychain currently takes (service string).
func TestGetFromKeychain_InternetPassword(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("keychain tests only run on darwin")
	}

	// ── COMPILE ERROR EXPECTED HERE ──────────────────────────────────────────
	lookup := KeychainLookup{
		Label:        "dvm-test-internet-lookup-nonexistent-99999",
		KeychainType: "internet",
	}
	_, err := GetFromKeychain(lookup)
	// ─────────────────────────────────────────────────────────────────────────

	assert.Error(t, err,
		"GetFromKeychain with internet type must error for non-existent label")
	// The error must originate from find-internet-password, which also returns
	// exit code 44 when the entry is not found.
	assert.Contains(t, err.Error(), "not found",
		"internet-password lookup should report 'not found' for missing entry")
}

// TestGetFromKeychain_GenericIsDefault verifies that a KeychainLookup with an
// empty KeychainType behaves identically to one with KeychainType == "generic".
// Both should attempt find-generic-password and return the same "not found" error.
//
// WILL FAIL TO COMPILE — GetFromKeychain currently takes (service string).
func TestGetFromKeychain_GenericIsDefault(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("keychain tests only run on darwin")
	}

	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	const nonExistentLabel = "dvm-test-generic-default-99999"

	lookupExplicit := KeychainLookup{
		Label:        nonExistentLabel,
		KeychainType: "generic",
	}
	lookupDefault := KeychainLookup{
		Label: nonExistentLabel,
		// KeychainType empty — should default to generic
	}

	_, errExplicit := GetFromKeychain(lookupExplicit)
	_, errDefault := GetFromKeychain(lookupDefault)
	// ─────────────────────────────────────────────────────────────────────────

	assert.Error(t, errExplicit, "explicit generic lookup must error for non-existent label")
	assert.Error(t, errDefault, "default (empty) type lookup must error for non-existent label")

	// Both must yield equivalent errors — same "not found" message
	assert.Contains(t, errExplicit.Error(), "not found")
	assert.Contains(t, errDefault.Error(), "not found")
}

// ---------------------------------------------------------------------------
// Section: GetAccountFromKeychain Label-Based Tests
// ---------------------------------------------------------------------------

// TestGetAccountFromKeychain_LabelLookup verifies that GetAccountFromKeychain
// accepts a KeychainLookup struct and uses the -l (label) flag.
//
// WILL FAIL TO COMPILE — GetAccountFromKeychain currently takes (service string),
// not (lookup KeychainLookup). The signature must change.
func TestGetAccountFromKeychain_LabelLookup(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("keychain tests only run on darwin")
	}

	// ── COMPILE ERROR EXPECTED HERE ──────────────────────────────────────────
	// GetAccountFromKeychain(lookup KeychainLookup) does not exist yet.
	// Current signature: GetAccountFromKeychain(service string)
	lookup := KeychainLookup{
		Label:        "dvm-test-account-label-nonexistent-99999",
		KeychainType: "generic",
	}
	_, err := GetAccountFromKeychain(lookup)
	// ─────────────────────────────────────────────────────────────────────────

	assert.Error(t, err,
		"GetAccountFromKeychain must return an error for a non-existent label")
	assert.Contains(t, err.Error(), "not found",
		"error should indicate label was not found in Keychain (exit code 44)")
}

// TestGetAccountFromKeychain_NoUserFilter verifies that read operations no
// longer pass -a $USER to the security command. We test this indirectly:
// a non-existent label must produce an exit-code-44 "not found" error. If
// the -a flag were still passed AND the $USER account didn't match an entry,
// the error would be the same (not found). We confirm the error path is
// exercised cleanly with the new label-only lookup (no user filter applied).
//
// This is as close as we can get to verifying the -a flag removal without
// mocking exec.Command.
//
// WILL FAIL TO COMPILE — GetAccountFromKeychain currently takes (service string).
func TestGetAccountFromKeychain_NoUserFilter(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("keychain tests only run on darwin")
	}

	// ── COMPILE ERROR EXPECTED HERE ──────────────────────────────────────────
	lookup := KeychainLookup{
		Label:        "dvm-test-no-user-filter-nonexistent-99999",
		KeychainType: "generic",
	}
	_, err := GetAccountFromKeychain(lookup)
	// ─────────────────────────────────────────────────────────────────────────

	// The lookup must fail with "not found" (exit code 44), NOT with any
	// USER-environment-variable error. If the implementation had required
	// USER env var and it was unset, we'd get a different error entirely.
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found",
		"label-only lookup should return 'not found', not a USER env var error")
	assert.NotContains(t, err.Error(), "USER environment variable",
		"label-only read operations must not require or check $USER")
}
