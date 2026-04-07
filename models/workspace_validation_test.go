package models_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"testing"
	"time"

	"devopsmaestro/models"
	"devopsmaestro/pkg/envvalidation"
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
			err := envvalidation.ValidateEnvKey(tt.key)
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
			got := envvalidation.IsDangerousEnvVar(tt.key)
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
//   - models.ValidateCACerts(certs []CACertConfig) error
//   - models.ValidatePEMContent(content string) error
//
// They will fail to COMPILE until the implementation is added in Phase 3.
// =============================================================================

// TestCACertConfig_Validation_RequiresName verifies that CACertConfig.Validate
// returns an error when Name is empty.
func TestCACertConfig_Validation_RequiresName(t *testing.T) {
	cert := models.CACertConfig{Name: "", VaultSecret: "test-secret"}
	err := cert.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

// TestCACertConfig_Validation_RequiresVaultSecret verifies that
// CACertConfig.Validate returns an error when VaultSecret is empty.
func TestCACertConfig_Validation_RequiresVaultSecret(t *testing.T) {
	cert := models.CACertConfig{Name: "my-cert", VaultSecret: ""}
	err := cert.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "vaultSecret")
}

// TestCACertConfig_Validation_MaxCerts verifies that ValidateCACerts enforces
// a maximum of 10 CA certificates.
func TestCACertConfig_Validation_MaxCerts(t *testing.T) {
	// Create 11 certs (over the limit of 10)
	certs := make([]models.CACertConfig, 11)
	for i := range certs {
		certs[i] = models.CACertConfig{
			Name:        fmt.Sprintf("cert-%d", i),
			VaultSecret: fmt.Sprintf("secret-%d", i),
		}
	}
	err := models.ValidateCACerts(certs)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "10")
}

// TestCACertConfig_Validation_PEMFormat verifies that ValidatePEMContent
// rejects content that is not in PEM format (no BEGIN/END markers).
func TestCACertConfig_Validation_PEMFormat(t *testing.T) {
	// Non-PEM content (no BEGIN/END markers)
	err := models.ValidatePEMContent("this is not a PEM certificate")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "PEM")
}

// TestCACertConfig_Validation_PEMTruncated verifies that ValidatePEMContent
// rejects PEM content that has a BEGIN marker but no END marker.
func TestCACertConfig_Validation_PEMTruncated(t *testing.T) {
	// PEM with BEGIN but no END marker
	truncated := "-----BEGIN CERTIFICATE-----\nMIIBkTCB+wIJ"
	err := models.ValidatePEMContent(truncated)
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
			cert := models.CACertConfig{Name: tt.certName, VaultSecret: "secret"}
			err := cert.Validate()
			require.Error(t, err)
		})
	}
}

// =============================================================================
// v0.56.0 Phase 2 RED Tests: CA Certs Cascade — Enhanced PEM Validation
//
// These tests drive crypto/x509-based PEM parsing in ValidatePEMContent.
// The current implementation only checks for PEM markers; it does NOT:
//   - Parse the DER-encoded certificate body
//   - Verify BasicConstraints.IsCA == true
//   - Reject non-CERTIFICATE PEM block types (e.g. PRIVATE KEY)
//   - Detect garbage/invalid base64 in the PEM body
//
// RED: Tests marked "WILL FAIL" pass today ONLY because the existing
// implementation is too permissive. They MUST fail once the stricter
// x509 parsing is added in Phase 3.
//
// Tests marked "WILL NOT COMPILE" reference functions that don't exist yet.
// =============================================================================

// TestValidatePEMContent_ValidCA verifies that ValidatePEMContent accepts a
// well-formed PEM block containing a self-signed CA certificate
// (BasicConstraints CA=true).
//
// RED: WILL FAIL — current implementation only checks markers; once x509
// parsing is added, a real CA cert must pass. Until then this test passes
// vacuously (the marker check succeeds). After Phase 3 it must still pass
// because it IS a valid CA cert.
//
// Note: The PEM below is a minimal self-signed CA cert generated for testing.
func TestValidatePEMContent_ValidCA(t *testing.T) {
	// Real self-signed CA certificate generated with:
	//   openssl req -x509 -newkey rsa:2048 -nodes -days 3650 \
	//     -subj "/CN=Test CA" -addext "basicConstraints=critical,CA:true"
	validCAPEM := `-----BEGIN CERTIFICATE-----
MIIDBTCCAe2gAwIBAgIUWfu0fjrzeodYjtwXRuXDqw+7dqgwDQYJKoZIhvcNAQEL
BQAwEjEQMA4GA1UEAwwHVGVzdCBDQTAeFw0yNjAzMTgxODI2NDBaFw0zNjAzMTUx
ODI2NDBaMBIxEDAOBgNVBAMMB1Rlc3QgQ0EwggEiMA0GCSqGSIb3DQEBAQUAA4IB
DwAwggEKAoIBAQDV3bF7wPo1am3SgsAC0ZbVS88fUZye7mu8sy0MGirIzZPkxG9B
S9/f3BraBUwIpRC2UN+/NYEprxClrPFwqkFW8bT0CoOju72wvlHWp86ezDVV+OzK
X7vgvKvVn7z7dookj1tk06bV2Bg4sx/XhtfHT2TRIH1Rh6csOmCIz4upQkO43wj5
zCE1ASiqn4nfvaegEN1Y9ktvJXMrSzMeR4x/DYBZJ2neVx3lYiVXu/Fzmy68pMi0
BRIR5+VAJN/O2tDPxfX5Qw1FLjGgl0oQiZ3Cbgh8JoRm9ITrvJAan72CyIcf6bMn
186MGQGWkREdy62TDpJS/AHhJwEFlRFD7JT1AgMBAAGjUzBRMB0GA1UdDgQWBBTq
/iPgUXxjAZ8xhWvX9JUQvCZ3MzAfBgNVHSMEGDAWgBTq/iPgUXxjAZ8xhWvX9JUQ
vCZ3MzAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQC3Z+3YlDQa
ZZwCdcJzYFR5+KrisbMMscO3DNXb/svhBkcON62E8RJztD6IcHtw90YUyvCPuWyJ
Ew+Wp0EVAw2KMhE32+bqSJdly/gkC44xANZ2gi6mW+v3ed0kqpwc/glslBx+eU1U
rf9HGME4S1i7il/4n1bwj1Dn4TaZsToT1TA0kjEnzoUKTbmlQcseqcu7fTOuex2N
lQ4H7WfgYwU1LHg3JMM9gZT8uPR8VdDjcz3Lbnbc40DJnAXDmGkkSN62p14crE1C
FFqZqdhjOhsCGn9OpTKSS9kuZloj/94ePzTcUOOj+yIqTCVl3hNYVhzLzyDBg1QX
OUhBijtAaeDg
-----END CERTIFICATE-----`

	err := models.ValidatePEMContent(validCAPEM)
	assert.NoError(t, err, "valid CA certificate PEM should be accepted")
}

// TestValidatePEMContent_LeafCert verifies that ValidatePEMContent accepts
// a PEM containing a leaf (end-entity) certificate — one where
// BasicConstraints.IsCA is false. Per issue #143, users may legitimately
// trust server certificates in their ca-certs store. A debug-level warning
// is logged but the cert is accepted.
func TestValidatePEMContent_LeafCert(t *testing.T) {
	// Real leaf certificate (CA:FALSE) generated with:
	//   openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out key.pem
	//   openssl req -new -key key.pem -out csr.pem -subj "/CN=Test Leaf"
	//   openssl x509 -req -in csr.pem -signkey key.pem -days 3650 \
	//     -extfile <(echo "basicConstraints=CA:FALSE") -out leaf.pem
	leafCertPEM := `-----BEGIN CERTIFICATE-----
MIIC4jCCAcqgAwIBAgIUOa2KDpQL2W/lHdR2Vj/r37BJiqowDQYJKoZIhvcNAQEL
BQAwFDESMBAGA1UEAwwJVGVzdCBMZWFmMB4XDTI2MDMxODE4MjY1M1oXDTM2MDMx
NTE4MjY1M1owFDESMBAGA1UEAwwJVGVzdCBMZWFmMIIBIjANBgkqhkiG9w0BAQEF
AAOCAQ8AMIIBCgKCAQEA2cOpyx2Kpz19UppXOHqsA7112/f++AW9Tt+RJW/CJfDQ
iKyaMUyncUSMlTpNWiHlEmg2wwqbxrrONSJGDM0hlBQCfeJMVCS6NB945/1WSIEp
KA58zaQPpR5Hmc3YlnpajxEkS1zFjhkwsyHxhRqJ0rIGy/8a7acEPJj6xHCIzNVm
s70ODkez5qJGVO5FNRjgyYJGjApxPiZ5G3wUOWw174hh3HPLrlWhX+1oQgcXt4Ee
WgPlhDxwMkt6NycTskUKp2dJAP7UejNH8FCfe0XcgqV5NLRPsjxRhXmHvj3+e3ir
CusqdcDQcB4LagTDNijtgBTD0GE/TSSl8y4VE2Q4XwIDAQABoywwKjAJBgNVHRME
AjAAMB0GA1UdDgQWBBQ8f0Iwh2H0BKD+1TC0Z97WfYWOBDANBgkqhkiG9w0BAQsF
AAOCAQEAoZgLOjuBEaQocM/Gcz2+lHkryo2hpHv6V9CyZ+qhxJnenRr24gHxJyji
D0JeCp/suhsde9gInsAljySfmEKDUwPAb2XmyUfs0fmcP8FBU+/NGBDsyPH5IsSZ
KCyOfYCoMliYsoWZjYBKsszeLJBfgsF/bKJEzAUrtMHMDFDHUTXEQU1vCWHgV3q9
fUsOV7QG4O+5Rdf9kzJuOZZLrdN8qAhGOfYeI2VJWUZqmIn6Bn2VDezUXMWiSLZZ
kwOiuCbHlQWkBigSvodkDIGT+tSSjU0jh6o1q77+7/hdinzuEYbBYuJ3zg3TBG/l
GRp7Y6Z0Xr9qzitRKCKyJptV56q2Eg==
-----END CERTIFICATE-----`

	err := models.ValidatePEMContent(leafCertPEM)
	assert.NoError(t, err,
		"leaf certificate (IsCA=false) should be accepted by ValidatePEMContent (issue #143)")
}

// TestValidatePEMContent_PrivateKeyRejected verifies that ValidatePEMContent
// rejects PEM content whose block type is PRIVATE KEY (or RSA PRIVATE KEY),
// even if the block is well-formed.
//
// RED: WILL FAIL — current implementation checks for "BEGIN CERTIFICATE" prefix
// only. A private key PEM does NOT start with "BEGIN CERTIFICATE", so the
// current code will return an error — but for the WRONG reason (marker check).
// After Phase 3 the error message must specifically mention the block type.
func TestValidatePEMContent_PrivateKeyRejected(t *testing.T) {
	privateKeyPEM := `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA2a2rwplBQLzHPZe5TNJNK7bLzQKFRDtyePFuUBOQHNagFqNs
n91bCFMFDRGmHkSV0xE0lG1iRITgMjTaEJ9p5butAHzOGBnf3OmBOI7OUmskbfvL
-----END RSA PRIVATE KEY-----`

	err := models.ValidatePEMContent(privateKeyPEM)
	require.Error(t, err,
		"PEM block with type RSA PRIVATE KEY should be rejected")
	// After Phase 3, the error must mention "CERTIFICATE" or "block type":
	assert.Contains(t, err.Error(), "CERTIFICATE",
		"error message should indicate only CERTIFICATE blocks are accepted")
}

// TestValidatePEMContent_InvalidDER verifies that ValidatePEMContent rejects
// PEM content that has valid BEGIN/END markers but whose base64-encoded body
// is not valid DER (i.e., garbage bytes).
//
// RED: WILL FAIL — current implementation only checks markers; it will return
// nil for this input. After Phase 3 adds x509.ParseCertificate(), this must
// return an error (DER parse failure).
func TestValidatePEMContent_InvalidDER(t *testing.T) {
	// Well-formed PEM envelope but the body is not a valid DER certificate.
	garbageDERPEM := `-----BEGIN CERTIFICATE-----
dGhpcyBpcyBub3QgYSB2YWxpZCBERVIgY2VydGlmaWNhdGUgYXQgYWxsCg==
-----END CERTIFICATE-----`

	// RED: This INCORRECTLY returns nil today (marker check only).
	// After Phase 3, ValidatePEMContent must return an error for invalid DER.
	err := models.ValidatePEMContent(garbageDERPEM)
	require.Error(t, err,
		"PEM with invalid DER body should be rejected by ValidatePEMContent")
	assert.Contains(t, err.Error(), "parse",
		"error message should indicate a parse failure")
}

// TestCACertConfig_NameMaxLength verifies that CACertConfig.Validate rejects
// cert names longer than 64 characters.
//
// RED: WILL FAIL — the current certNameRegex (`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)
// has NO length limit. A 65-char valid name currently passes. After Phase 3
// adds a length check to Validate(), this test must pass.
func TestCACertConfig_NameMaxLength(t *testing.T) {
	tests := []struct {
		name     string
		certName string
		wantErr  bool
	}{
		{
			name:     "64 char name is valid",
			certName: "A" + fmt.Sprintf("%063d", 0), // 1 + 63 = 64 chars
			wantErr:  false,
		},
		{
			name:     "65 char name is too long",
			certName: "A" + fmt.Sprintf("%064d", 0), // 1 + 64 = 65 chars
			wantErr:  true,
		},
		{
			name:     "100 char name is too long",
			certName: fmt.Sprintf("A%099d", 0), // 1 + 99 = 100 chars
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cert := models.CACertConfig{Name: tt.certName, VaultSecret: "some-secret"}
			err := cert.Validate()
			if tt.wantErr {
				// RED: currently passes (no length check) — must fail after Phase 3.
				require.Error(t, err,
					"cert name of length %d should be rejected", len(tt.certName))
				assert.Contains(t, err.Error(), "64",
					"error should mention the 64-char limit")
			} else {
				assert.NoError(t, err,
					"cert name of length %d should be accepted", len(tt.certName))
			}
		})
	}
}

// TestValidateCACerts_MergedExceedsMax verifies that ValidateCACerts enforces
// the 10-cert maximum on a MERGED cascade result (not just per-level).
//
// This tests the scenario where individual levels are each under the limit
// but the merged result from CascadeResolver exceeds 10. ValidateCACerts must
// be called on the merged slice and must return an error.
//
// This test exercises the cascade resolver's output being validated — ensuring
// Phase 3 wires ValidateCACerts into the resolver pipeline.
//
// RED: The resolver package does not exist yet; once pkg/cacerts/resolver is
// implemented in Phase 3, the cascade result must be validated. For now this
// test exercises ValidateCACerts directly with 10+ entries to confirm the
// boundary condition.
func TestValidateCACerts_MergedExceedsMax(t *testing.T) {
	// Simulate a merged result from 3 hierarchy levels that together produce 11 certs.
	// e.g., Ecosystem contributes 4, Domain adds 4 more, App adds 3 more unique names.
	mergedCerts := make([]models.CACertConfig, 11)
	for i := range mergedCerts {
		mergedCerts[i] = models.CACertConfig{
			Name:        fmt.Sprintf("merged-cert-%d", i),
			VaultSecret: fmt.Sprintf("vault-secret-%d", i),
		}
	}

	// Verify the 11-cert merged result is rejected.
	err := models.ValidateCACerts(mergedCerts)
	require.Error(t, err,
		"merged cascade result with 11 certs should exceed the 10-cert maximum")
	assert.Contains(t, err.Error(), "10",
		"error message should reference the 10-cert limit")

	// Also verify that exactly 10 is accepted (boundary condition).
	exactlyTen := mergedCerts[:10]
	err = models.ValidateCACerts(exactlyTen)
	assert.NoError(t, err,
		"merged result with exactly 10 certs should be accepted")
}

// =============================================================================
// NormalizePEMContent Tests
//
// NormalizePEMContent restores proper newline formatting to PEM content that
// has had its newlines collapsed to spaces (a known MaestroVault behaviour).
// =============================================================================

// TestNormalizePEMContent verifies all normalization cases using a table-driven
// approach: already-valid PEM is untouched, space-collapsed PEM is restored,
// cert chains with spaces are repaired, mixed content is handled, and
// leading/trailing whitespace is stripped.
func TestNormalizePEMContent(t *testing.T) {
	// validPEM is a well-formed single-certificate PEM with correct newlines.
	// All other cases are derived from this canonical form.
	const validPEM = `-----BEGIN CERTIFICATE-----
MIIBIDCB6wIJANS5fxVvbrSBMA0GCSqGSIb3DQEBCwUAMCExHzAdBgNVBAMT
FkRldk9wc01hZXN0cm9UZXN0Q0EwHhcNMjMwMTAxMDAwMDAwWhcNMzMwMTAx
MDAwMDAwWjAhMR8wHQYDVQQDExZEZXZPcHNNYWVzdHJvVGVzdENBMFwwDQYJ
KoZIhvcNAQEBBQADSwAwSAJBALRiMLAHudeSA/xKMSNtHRpfqo5PQKL8I9gD
AiDkX10kMwIDAQABMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQAD
QQCcLSHAhMHqSqRLQ9oGCePg/FK2KNhvNMNEYLPrBm8gnyHFe1yMBbvFkBpk
vGwc4H7VFtKlcCa8IxlMiOas
-----END CERTIFICATE-----`

	// secondCert simulates a second certificate in a chain (same body, different
	// label to keep it distinct from the first in string comparisons).
	const secondCert = `-----BEGIN CERTIFICATE-----
MIIBIDCB6wIJANS5fxVvbrSBMB0GCSqGSIb3DQEBCwUAMCIxIDAeBgNVBAMT
F0RlVk9wc01hZXN0cm9DaGFpbkNBMHhcNMjMwMTAxMDAwMDAwWhcNMzMwMTAx
MDAwMDAwWjAiMSAwHgYDVQQDExdEZXZPcHNNYWVzdHJvQ2hhaW5DQQIBAcwDQ
YJKoZIhvcNAQELBQADQQCcLSHAhMHqSqRLQ9oGCePg/FK2KNhvNMNEYLPrBm
8gnyHFe1yMBbvFkBpkvGwc4H7VFtKlcCa8IxlMiOas==
-----END CERTIFICATE-----`

	tests := []struct {
		name  string
		input string
		want  string
	}{
		// ── Case 1: already-valid PEM — must be a no-op ───────────────────────
		{
			name:  "already valid pem is unchanged",
			input: validPEM,
			want:  validPEM,
		},

		// ── Case 2: vault bug — spaces instead of newlines (single cert) ──────
		// MaestroVault collapses every \n to a single space.  The BEGIN marker
		// suffix space and the pre-END marker space must both be restored.
		// Note: the space in the middle of the base64 body is NOT touched —
		// NormalizePEMContent only fixes the BEGIN/END boundaries.
		{
			name: "spaces instead of newlines (vault bug)",
			input: "-----BEGIN CERTIFICATE----- " +
				"MIIBIDCB6wIJANS5fxVvbrSBMA0GCSqGSIb3DQEBCwUAMCExHzAdBgNVBAMT " +
				"FkRldk9wc01hZXN0cm9UZXN0Q0E= " +
				"-----END CERTIFICATE-----",
			// "-----BEGIN CERTIFICATE----- " → "-----BEGIN CERTIFICATE-----\n"  (space consumed)
			// " -----END CERTIFICATE-----"   → "\n-----END CERTIFICATE-----"   (space consumed)
			want: "-----BEGIN CERTIFICATE-----\n" +
				"MIIBIDCB6wIJANS5fxVvbrSBMA0GCSqGSIb3DQEBCwUAMCExHzAdBgNVBAMT " +
				"FkRldk9wc01hZXN0cm9UZXN0Q0E=" +
				"\n-----END CERTIFICATE-----",
		},

		// ── Case 3: cert chain with spaces ────────────────────────────────────
		// Two complete PEM blocks joined by a space instead of a newline.
		// "-----END CERTIFICATE----- -----BEGIN CERTIFICATE-----" is handled by
		// two sequential replacements:
		//   " -----END CERTIFICATE-----"  → "\n-----END CERTIFICATE-----"
		//   " -----BEGIN CERTIFICATE-----" → "\n-----BEGIN CERTIFICATE-----"
		{
			name: "cert chain with spaces between blocks",
			input: "-----BEGIN CERTIFICATE-----\n" +
				"MIIBIDCB6wIJANS5fxVvbrSBMA0GCSqGSIb3DQEBCwUAMCExHzAdBgNVBAMT\n" +
				"-----END CERTIFICATE----- -----BEGIN CERTIFICATE-----\n" +
				"MIIBIDCB6wIJANS5fxVvbrSBMB0GCSqGSIb3DQEBCwUAMCIxIDAeBgNVBAMT\n" +
				"-----END CERTIFICATE-----",
			// The "-----END CERTIFICATE----- -----BEGIN CERTIFICATE-----" span:
			//   step 1 (" -----END"  → "\n-----END"):  "-----END CERTIFICATE-----\n-----BEGIN CERTIFICATE-----"
			//   step 2 (" -----BEGIN" → no longer matches — already replaced above)
			// Result: END and BEGIN are separated by a single \n (no double-blank-line).
			want: "-----BEGIN CERTIFICATE-----\n" +
				"MIIBIDCB6wIJANS5fxVvbrSBMA0GCSqGSIb3DQEBCwUAMCExHzAdBgNVBAMT\n" +
				"-----END CERTIFICATE-----\n-----BEGIN CERTIFICATE-----\n" +
				"MIIBIDCB6wIJANS5fxVvbrSBMB0GCSqGSIb3DQEBCwUAMCIxIDAeBgNVBAMT\n" +
				"-----END CERTIFICATE-----",
		},

		// ── Case 4: mixed — some newlines correct, some replaced by spaces ────
		// The BEGIN→body transition is already a newline, but body→END uses a
		// space (vault bug), and the inter-cert separator is also a space.
		{
			name: "mixed correct newlines and space separators",
			input: "-----BEGIN CERTIFICATE-----\n" +
				"MIIBIDCB6wIJANS5fxVvbrSBMA0GCSqGSIb3DQEBCwUAMCExHzAdBgNVBAMT= " +
				"-----END CERTIFICATE----- -----BEGIN CERTIFICATE-----\n" +
				"MIIBIDCB6wIJANS5fxVvbrSBMB0GCSqGSIb3DQEBCwUAMCIxIDAeBgNVBAMT=\n" +
				"-----END CERTIFICATE-----",
			// " -----END CERTIFICATE-----" → "\n-----END CERTIFICATE-----" (space consumed)
			// " -----BEGIN CERTIFICATE-----" → already handled by prior replacement leaving "\n-----BEGIN"
			want: "-----BEGIN CERTIFICATE-----\n" +
				"MIIBIDCB6wIJANS5fxVvbrSBMA0GCSqGSIb3DQEBCwUAMCExHzAdBgNVBAMT=" +
				"\n-----END CERTIFICATE-----\n-----BEGIN CERTIFICATE-----\n" +
				"MIIBIDCB6wIJANS5fxVvbrSBMB0GCSqGSIb3DQEBCwUAMCIxIDAeBgNVBAMT=\n" +
				"-----END CERTIFICATE-----",
		},

		// ── Case 5: leading and trailing whitespace is stripped ───────────────
		{
			name:  "leading and trailing whitespace stripped",
			input: "   \n\t" + validPEM + "\n\t   ",
			want:  validPEM,
		},

		// ── Case 6: leading/trailing whitespace plus vault-style space collapse─
		{
			name: "leading/trailing whitespace with space-collapsed pem",
			input: "  \n-----BEGIN CERTIFICATE----- " +
				"MIIBIDCB6wIJANS5fxVvbrSBMA0= " +
				"-----END CERTIFICATE-----\n  ",
			want: "-----BEGIN CERTIFICATE-----\n" +
				"MIIBIDCB6wIJANS5fxVvbrSBMA0=" +
				"\n-----END CERTIFICATE-----",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange + Act
			got := models.NormalizePEMContent(tt.input)

			// Assert
			assert.Equal(t, tt.want, got,
				"NormalizePEMContent(%q) produced unexpected output", tt.name)
		})
	}

	// Verify the already-valid case is truly a no-op using a stricter identity
	// check — ensures the function doesn't silently mutate correct input.
	t.Run("no-op identity for valid pem", func(t *testing.T) {
		got := models.NormalizePEMContent(validPEM)
		require.Equal(t, validPEM, got,
			"NormalizePEMContent must not alter already-valid PEM content")
	})

	// Verify a two-cert chain round-trip produces both BEGIN/END markers in order.
	t.Run("two-cert chain markers preserved after normalization", func(t *testing.T) {
		// Build a fully space-collapsed two-cert chain (worst-case vault output).
		chain := validPEM + "\n" + secondCert
		got := models.NormalizePEMContent(chain)

		require.Contains(t, got, "-----BEGIN CERTIFICATE-----",
			"normalized chain must contain BEGIN marker")
		require.Contains(t, got, "-----END CERTIFICATE-----",
			"normalized chain must contain END marker")

		// Both certs must be present: count occurrences of the BEGIN marker.
		const beginMarker = "-----BEGIN CERTIFICATE-----"
		count := 0
		s := got
		for {
			idx := len(s) - len(s[len(beginMarker):]) // avoid import of strings
			_ = idx
			i := 0
			for i < len(s)-len(beginMarker)+1 {
				if s[i:i+len(beginMarker)] == beginMarker {
					count++
					s = s[i+len(beginMarker):]
					break
				}
				i++
			}
			if i >= len(s)-len(beginMarker)+1 {
				break
			}
		}
		assert.Equal(t, 2, count,
			"normalized two-cert chain should contain exactly 2 BEGIN CERTIFICATE markers")
	})
}

// =============================================================================
// Issue #143: Certificate chain and non-CA cert acceptance tests
//
// These tests verify that ValidatePEMContent:
//   - Accepts leaf/server certs (CA=false) with a debug warning
//   - Accepts certificate chains (leaf + intermediate + root)
//   - Rejects chains containing an expired certificate
//   - Still accepts single valid CA certs (regression check)
//   - Still rejects empty PEM / garbage data (regression check)
// =============================================================================

// generateTestCert creates a self-signed X.509 certificate for testing.
// If isCA is true, BasicConstraints CA:TRUE is set. The cert is valid from
// notBefore to notAfter. Returns the PEM-encoded certificate string.
func generateTestCert(t *testing.T, cn string, isCA bool, notBefore, notAfter time.Time) string {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "failed to generate RSA key for test cert")

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		IsCA:         isCA,
		KeyUsage:     x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
	}
	if isCA {
		template.BasicConstraintsValid = true
	} else {
		template.BasicConstraintsValid = true
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err, "failed to create test certificate")

	pemBlock := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	return string(pemBlock)
}

// TestValidatePEMContent_CertificateChain verifies that a PEM bundle containing
// multiple certificates (leaf + intermediate + root) is accepted when all certs
// are valid.
func TestValidatePEMContent_CertificateChain(t *testing.T) {
	now := time.Now()
	leaf := generateTestCert(t, "Leaf Cert", false, now.Add(-time.Hour), now.Add(24*time.Hour))
	intermediate := generateTestCert(t, "Intermediate CA", true, now.Add(-time.Hour), now.Add(24*time.Hour))
	root := generateTestCert(t, "Root CA", true, now.Add(-time.Hour), now.Add(24*time.Hour))

	chain := leaf + intermediate + root

	err := models.ValidatePEMContent(chain)
	assert.NoError(t, err,
		"certificate chain (leaf + intermediate + root) should be accepted when all certs are valid")
}

// TestValidatePEMContent_ChainWithExpiredCert verifies that a PEM bundle
// containing at least one expired certificate is rejected with a clear error.
func TestValidatePEMContent_ChainWithExpiredCert(t *testing.T) {
	now := time.Now()
	validCert := generateTestCert(t, "Valid CA", true, now.Add(-time.Hour), now.Add(24*time.Hour))
	expiredCert := generateTestCert(t, "Expired CA", true, now.Add(-48*time.Hour), now.Add(-24*time.Hour))

	chain := validCert + expiredCert

	err := models.ValidatePEMContent(chain)
	require.Error(t, err,
		"chain with an expired certificate should be rejected")
	assert.Contains(t, err.Error(), "expired",
		"error message should mention expiration")
	assert.Contains(t, err.Error(), "Expired CA",
		"error message should name the expired certificate")
}

// TestValidatePEMContent_SingleValidCA is a regression check ensuring a
// single valid CA certificate continues to be accepted after the IsCA
// check removal.
func TestValidatePEMContent_SingleValidCA(t *testing.T) {
	now := time.Now()
	caPEM := generateTestCert(t, "Regression Test CA", true, now.Add(-time.Hour), now.Add(24*time.Hour))

	err := models.ValidatePEMContent(caPEM)
	assert.NoError(t, err,
		"single valid CA certificate should be accepted (regression check)")
}

// TestValidatePEMContent_EmptyAndGarbage is a regression check ensuring
// empty PEM content and garbage data are still rejected.
func TestValidatePEMContent_EmptyAndGarbage(t *testing.T) {
	tests := []struct {
		name    string
		content string
		errMsg  string
	}{
		{
			name:    "empty string",
			content: "",
			errMsg:  "BEGIN CERTIFICATE",
		},
		{
			name:    "whitespace only",
			content: "   \n\t  ",
			errMsg:  "BEGIN CERTIFICATE",
		},
		{
			name:    "random garbage",
			content: "this is totally not a certificate!!!",
			errMsg:  "BEGIN CERTIFICATE",
		},
		{
			name:    "looks like PEM but garbage body",
			content: "-----BEGIN CERTIFICATE-----\ndGhpcyBpcyBub3QgYSB2YWxpZCBERVIgY2VydGlmaWNhdGUgYXQgYWxsCg==\n-----END CERTIFICATE-----",
			errMsg:  "parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := models.ValidatePEMContent(tt.content)
			require.Error(t, err,
				"invalid content %q should be rejected", tt.name)
			assert.Contains(t, err.Error(), tt.errMsg,
				"error message should contain %q", tt.errMsg)
		})
	}
}
