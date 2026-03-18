package models

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
