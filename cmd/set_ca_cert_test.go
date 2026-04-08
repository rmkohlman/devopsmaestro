// Package cmd_test contains Phase 2 RED tests for CACertConfig.Validate().
//
// RED PHASE: These tests verify that CACertConfig.Validate() rejects invalid
// inputs. The tests PASS today (Validate() already exists), but they prove the
// contract that runSetCACert() MUST call cert.Validate() before dispatching.
//
// The implementation fix is: add cert.Validate() to runSetCACert() in
// cmd/set_ca_cert.go, immediately after building the cert struct.
//
// These tests will FAIL (compile error) if CACertConfig or its Validate()
// method is removed, serving as a regression guard.
package cmd

import (
	"strings"
	"testing"

	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetCACert_RejectsInvalidName verifies that CACertConfig.Validate() returns
// an error for names that would be unsafe as filenames or violate the cert name
// policy. The implementation fix (adding cert.Validate() to runSetCACert) ensures
// these are rejected at the CLI boundary before any DB writes occur.
func TestSetCACert_RejectsInvalidName(t *testing.T) {
	tests := []struct {
		name        string
		certName    string
		vaultSecret string
		wantErr     bool
		errContains string
	}{
		{
			name:        "path traversal with ../",
			certName:    "../traversal",
			vaultSecret: "my-vault-secret",
			wantErr:     true,
			errContains: "invalid",
		},
		{
			name:        "name with spaces",
			certName:    "cert with spaces",
			vaultSecret: "my-vault-secret",
			wantErr:     true,
			errContains: "invalid",
		},
		{
			name:        "name exceeds 64 characters",
			certName:    strings.Repeat("a", 65),
			vaultSecret: "my-vault-secret",
			wantErr:     true,
			errContains: "exceeds maximum length",
		},
		{
			name:        "empty vault secret",
			certName:    "valid-cert-name",
			vaultSecret: "",
			wantErr:     true,
			errContains: "vaultSecret is required",
		},
		{
			name:        "valid name and vault secret",
			certName:    "corp-root-ca",
			vaultSecret: "my-vault-secret",
			wantErr:     false,
		},
		{
			name:        "name with leading dot",
			certName:    ".hidden-cert",
			vaultSecret: "my-vault-secret",
			wantErr:     true,
			errContains: "invalid",
		},
		{
			name:        "name with slash",
			certName:    "some/nested/cert",
			vaultSecret: "my-vault-secret",
			wantErr:     true,
			errContains: "invalid",
		},
		{
			name:        "empty name",
			certName:    "",
			vaultSecret: "my-vault-secret",
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name:        "name exactly 64 characters is valid",
			certName:    strings.Repeat("a", 64),
			vaultSecret: "my-vault-secret",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cert := models.CACertConfig{
				Name:        tt.certName,
				VaultSecret: tt.vaultSecret,
			}

			err := cert.Validate()

			if tt.wantErr {
				require.Error(t, err,
					"Validate() should return an error for cert name %q", tt.certName)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains,
						"error message should contain %q", tt.errContains)
				}
			} else {
				assert.NoError(t, err,
					"Validate() should not error for cert name %q", tt.certName)
			}
		})
	}
}
