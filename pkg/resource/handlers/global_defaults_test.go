// Package handlers contains Phase 2 RED tests for GlobalDefaultsResource.Validate().
//
// RED PHASE: TestGlobalDefaultsResource_Validate_RejectsInvalidCACert tests that
// Validate() returns an error when the resource contains an invalid CA cert.
//
// Currently Validate() in global_defaults.go always returns nil (line 383-385):
//
//	func (r *GlobalDefaultsResource) Validate() error {
//	    return nil
//	}
//
// These tests WILL FAIL until Validate() calls models.ValidateCACerts(r.caCerts).
// The fix required: update Validate() to delegate cert validation.
package handlers

import (
	"strings"
	"testing"

	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGlobalDefaultsResource_Validate_RejectsInvalidCACert verifies that
// Validate() on a GlobalDefaultsResource returns an error when it holds a
// CACertConfig with an invalid name.
//
// RED: Validate() currently always returns nil — this test will FAIL.
func TestGlobalDefaultsResource_Validate_RejectsInvalidCACert(t *testing.T) {
	tests := []struct {
		name        string
		caCerts     []models.CACertConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "path traversal name rejected",
			caCerts: []models.CACertConfig{
				{Name: "../traversal", VaultSecret: "my-secret"},
			},
			wantErr:     true,
			errContains: "invalid",
		},
		{
			name: "name with spaces rejected",
			caCerts: []models.CACertConfig{
				{Name: "cert with spaces", VaultSecret: "my-secret"},
			},
			wantErr:     true,
			errContains: "invalid",
		},
		{
			name: "name over 64 chars rejected",
			caCerts: []models.CACertConfig{
				{Name: strings.Repeat("x", 65), VaultSecret: "my-secret"},
			},
			wantErr:     true,
			errContains: "exceeds maximum length",
		},
		{
			name: "empty vault secret rejected",
			caCerts: []models.CACertConfig{
				{Name: "valid-cert", VaultSecret: ""},
			},
			wantErr:     true,
			errContains: "vaultSecret is required",
		},
		{
			name: "valid cert passes validation",
			caCerts: []models.CACertConfig{
				{Name: "corp-root-ca", VaultSecret: "corp-vault-secret"},
			},
			wantErr: false,
		},
		{
			name:    "no certs passes validation",
			caCerts: nil,
			wantErr: false,
		},
		{
			name: "one valid one invalid cert fails",
			caCerts: []models.CACertConfig{
				{Name: "good-cert", VaultSecret: "vault-secret"},
				{Name: "../evil", VaultSecret: "vault-secret"},
			},
			wantErr:     true,
			errContains: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: create a GlobalDefaultsResource with the test CA certs
			res := NewGlobalDefaultsResource("", nil, tt.caCerts)
			require.NotNil(t, res)

			// Act: call Validate() — RED: currently always returns nil
			err := res.Validate()

			// Assert
			if tt.wantErr {
				require.Error(t, err,
					"Validate() must return error for invalid CA cert config")
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains,
						"error should mention %q", tt.errContains)
				}
			} else {
				assert.NoError(t, err,
					"Validate() must not error for valid CA cert config")
			}
		})
	}
}
