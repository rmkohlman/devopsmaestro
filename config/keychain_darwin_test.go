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

	_, err := GetAccountFromKeychain("dvm-test-definitely-nonexistent-service-xyz-99999")

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
	_, realErr := GetAccountFromKeychain(nonExistentService)
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
