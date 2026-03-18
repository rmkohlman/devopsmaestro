package models

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// v0.55.0 Phase 2 RED Tests: WI-1/WI-3/WI-5 — Build Arg Key Validation
//
// Design Decision 14 (sprint plan): "Validate keys at all entry points + call
// IsDangerousEnvVar() on cascade args; validate KEY with ValidateEnvKey()."
//
// These tests verify two new public functions:
//   - ValidateBuildArgKey(key string) error
//     • Must be a valid environment variable name: [A-Z][A-Z0-9_]* (uppercase)
//     • Empty string is an error
//     • Keys starting with digits are invalid
//     • Keys containing hyphens or spaces are invalid
//     • Keys with DVM_ prefix are reserved (error)
//   - IsDangerousEnvVar(key string) bool
//     • Returns true for known dangerous/security-sensitive env var names
//     • Dangerous set includes at minimum: LD_PRELOAD, BASH_ENV, LD_LIBRARY_PATH,
//       ENV, DYLD_INSERT_LIBRARIES, DYLD_LIBRARY_PATH
//
// RED: ALL tests WILL NOT COMPILE until WI-1/WI-3 implementation adds these
// functions to the models package (or a shared validation package imported here).
//
// =============================================================================

// TestValidateBuildArgKey verifies that ValidateBuildArgKey accepts valid
// uppercase env-style keys and rejects malformed ones.
//
// RED: WILL NOT COMPILE — ValidateBuildArgKey does not exist yet.
func TestValidateBuildArgKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
		errMsg  string
	}{
		// ── Valid keys ────────────────────────────────────────────────────────
		{
			name:    "simple uppercase key",
			key:     "CGO_ENABLED",
			wantErr: false,
		},
		{
			name:    "key with numbers (not leading)",
			key:     "PIP_INDEX_URL",
			wantErr: false,
		},
		{
			name:    "key with trailing number",
			key:     "MY_ARG_1",
			wantErr: false,
		},
		{
			name:    "single letter key",
			key:     "X",
			wantErr: false,
		},
		{
			name:    "key starting with letter then digits",
			key:     "V8_FLAGS",
			wantErr: false,
		},
		// ── Invalid keys ─────────────────────────────────────────────────────
		{
			name:    "empty string",
			key:     "",
			wantErr: true,
			errMsg:  "empty",
		},
		{
			name:    "leading digit",
			key:     "123BAD",
			wantErr: true,
			errMsg:  "invalid",
		},
		{
			name:    "contains hyphen",
			key:     "pip-index",
			wantErr: true,
			errMsg:  "invalid",
		},
		{
			name:    "contains space",
			key:     "MY KEY",
			wantErr: true,
			errMsg:  "invalid",
		},
		{
			name:    "lowercase letters",
			key:     "my_arg",
			wantErr: true,
			errMsg:  "invalid",
		},
		{
			name:    "mixed case",
			key:     "My_Arg",
			wantErr: true,
			errMsg:  "invalid",
		},
		// ── Reserved prefix ──────────────────────────────────────────────────
		{
			name:    "DVM_ prefix is reserved",
			key:     "DVM_SECRET",
			wantErr: true,
			errMsg:  "reserved",
		},
		{
			name:    "DVM_ prefix case exact",
			key:     "DVM_INTERNAL",
			wantErr: true,
			errMsg:  "reserved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────
			// ValidateBuildArgKey does not exist yet — WI-1/WI-3.
			err := ValidateBuildArgKey(tt.key)
			// ─────────────────────────────────────────────────────────────────
			if tt.wantErr {
				require.Error(t, err,
					"ValidateBuildArgKey(%q) should return error", tt.key)
				assert.Contains(t, err.Error(), tt.errMsg,
					"error message for key %q should contain %q", tt.key, tt.errMsg)
			} else {
				assert.NoError(t, err,
					"ValidateBuildArgKey(%q) should return nil for valid key", tt.key)
			}
		})
	}
}

// TestIsDangerousEnvVar verifies that IsDangerousEnvVar correctly identifies
// known dangerous environment variable names that could be exploited for
// privilege escalation or code injection when used as build args.
//
// RED: WILL NOT COMPILE — IsDangerousEnvVar does not exist yet (WI-1/WI-3).
func TestIsDangerousEnvVar(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		wantDanger bool
	}{
		// ── Known dangerous vars ─────────────────────────────────────────────
		{
			name:       "LD_PRELOAD is dangerous",
			key:        "LD_PRELOAD",
			wantDanger: true,
		},
		{
			name:       "BASH_ENV is dangerous",
			key:        "BASH_ENV",
			wantDanger: true,
		},
		{
			name:       "LD_LIBRARY_PATH is dangerous",
			key:        "LD_LIBRARY_PATH",
			wantDanger: true,
		},
		{
			name:       "ENV is dangerous",
			key:        "ENV",
			wantDanger: true,
		},
		{
			name:       "DYLD_INSERT_LIBRARIES is dangerous",
			key:        "DYLD_INSERT_LIBRARIES",
			wantDanger: true,
		},
		{
			name:       "DYLD_LIBRARY_PATH is dangerous",
			key:        "DYLD_LIBRARY_PATH",
			wantDanger: true,
		},
		// ── Safe vars ────────────────────────────────────────────────────────
		{
			name:       "PIP_INDEX_URL is not dangerous",
			key:        "PIP_INDEX_URL",
			wantDanger: false,
		},
		{
			name:       "CGO_ENABLED is not dangerous",
			key:        "CGO_ENABLED",
			wantDanger: false,
		},
		{
			name:       "DEBUG_BUILD is not dangerous",
			key:        "DEBUG_BUILD",
			wantDanger: false,
		},
		{
			name:       "EXTRA_PIP_PACKAGES is not dangerous",
			key:        "EXTRA_PIP_PACKAGES",
			wantDanger: false,
		},
		{
			name:       "NPM_REGISTRY is not dangerous",
			key:        "NPM_REGISTRY",
			wantDanger: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────
			// IsDangerousEnvVar does not exist yet — WI-1/WI-3.
			got := IsDangerousEnvVar(tt.key)
			// ─────────────────────────────────────────────────────────────────
			assert.Equal(t, tt.wantDanger, got,
				"IsDangerousEnvVar(%q) = %v, want %v", tt.key, got, tt.wantDanger)
		})
	}
}

// =============================================================================
// v0.54.0: CACertConfig Validation Tests  [RED Phase]
//
// These tests reference validation methods that do NOT exist yet:
//   - CACertConfig.Validate() error
//   - ValidateCACerts(certs []CACertConfig) error
//   - ValidatePEMContent(content string) error
//
// They will fail to COMPILE until the implementation is added in Phase 3.
// =============================================================================

// TestCACertConfig_Validation_RequiresName verifies that CACertConfig.Validate
// returns an error when Name is empty.
func TestCACertConfig_Validation_RequiresName(t *testing.T) {
	cert := CACertConfig{Name: "", VaultSecret: "test-secret"}
	err := cert.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

// TestCACertConfig_Validation_RequiresVaultSecret verifies that
// CACertConfig.Validate returns an error when VaultSecret is empty.
func TestCACertConfig_Validation_RequiresVaultSecret(t *testing.T) {
	cert := CACertConfig{Name: "my-cert", VaultSecret: ""}
	err := cert.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "vaultSecret")
}

// TestCACertConfig_Validation_MaxCerts verifies that ValidateCACerts enforces
// a maximum of 10 CA certificates.
func TestCACertConfig_Validation_MaxCerts(t *testing.T) {
	// Create 11 certs (over the limit of 10)
	certs := make([]CACertConfig, 11)
	for i := range certs {
		certs[i] = CACertConfig{
			Name:        fmt.Sprintf("cert-%d", i),
			VaultSecret: fmt.Sprintf("secret-%d", i),
		}
	}
	err := ValidateCACerts(certs)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "10")
}

// TestCACertConfig_Validation_PEMFormat verifies that ValidatePEMContent
// rejects content that is not in PEM format (no BEGIN/END markers).
func TestCACertConfig_Validation_PEMFormat(t *testing.T) {
	// Non-PEM content (no BEGIN/END markers)
	err := ValidatePEMContent("this is not a PEM certificate")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "PEM")
}

// TestCACertConfig_Validation_PEMTruncated verifies that ValidatePEMContent
// rejects PEM content that has a BEGIN marker but no END marker.
func TestCACertConfig_Validation_PEMTruncated(t *testing.T) {
	// PEM with BEGIN but no END marker
	truncated := "-----BEGIN CERTIFICATE-----\nMIIBkTCB+wIJ"
	err := ValidatePEMContent(truncated)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "END")
}

// TestCACertConfig_Validation_RejectsPathTraversal verifies that
// CACertConfig.Validate rejects names containing path traversal patterns.
func TestCACertConfig_Validation_RejectsPathTraversal(t *testing.T) {
	tests := []struct {
		name     string
		certName string
	}{
		{"dot-dot-slash", "../etc/passwd"},
		{"absolute-path", "/etc/certs/ca.crt"},
		{"leading-dot", ".hidden-cert"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cert := CACertConfig{Name: tt.certName, VaultSecret: "secret"}
			err := cert.Validate()
			require.Error(t, err)
		})
	}
}
