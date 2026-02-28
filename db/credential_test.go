package db

import (
	"devopsmaestro/models"
	"strings"
	"testing"
)

// =============================================================================
// Task 2.2: Credential Validation Tests (v0.19.0)
// These tests verify that plaintext credentials are REJECTED
// =============================================================================

// TestCredentialRejectPlaintextSource verifies that credentials with source='value' are rejected
func TestCredentialRejectPlaintextSource(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create ecosystem for credential scope
	ecosystem := &models.Ecosystem{Name: "test-eco"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	tests := []struct {
		name    string
		cred    *models.CredentialDB
		wantErr bool
		errMsg  string
	}{
		{
			name: "reject plaintext source='value'",
			cred: &models.CredentialDB{
				ScopeType: models.CredentialScopeEcosystem,
				ScopeID:   int64(ecosystem.ID),
				Name:      "API_KEY",
				Source:    "value", // PROHIBITED
				Value:     stringPtr("secret123"),
			},
			wantErr: true,
			errMsg:  "plaintext credentials not allowed",
		},
		{
			name: "reject source='plaintext'",
			cred: &models.CredentialDB{
				ScopeType: models.CredentialScopeEcosystem,
				ScopeID:   int64(ecosystem.ID),
				Name:      "TOKEN",
				Source:    "plaintext", // PROHIBITED
				Value:     stringPtr("token123"),
			},
			wantErr: true,
			errMsg:  "plaintext credentials not allowed",
		},
		{
			name: "allow keychain source",
			cred: &models.CredentialDB{
				ScopeType: models.CredentialScopeEcosystem,
				ScopeID:   int64(ecosystem.ID),
				Name:      "GITHUB_TOKEN",
				Source:    "keychain", // ALLOWED
				Service:   stringPtr("github.com"),
			},
			wantErr: false,
		},
		{
			name: "allow env source",
			cred: &models.CredentialDB{
				ScopeType: models.CredentialScopeEcosystem,
				ScopeID:   int64(ecosystem.ID),
				Name:      "DATABASE_URL",
				Source:    "env", // ALLOWED
				EnvVar:    stringPtr("DATABASE_URL"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// FIXME: This test will FAIL - CreateCredential() doesn't validate source yet
			// After Phase 3, CreateCredential should reject source='value' and source='plaintext'
			err := ds.CreateCredential(tt.cred)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateCredential() expected error, got nil")
				} else if !strings.Contains(strings.ToLower(err.Error()), "plaintext") &&
					!strings.Contains(strings.ToLower(err.Error()), "not allowed") {
					t.Errorf("CreateCredential() error = %v, want error containing 'plaintext' or 'not allowed'", err)
				}
			} else {
				if err != nil {
					t.Errorf("CreateCredential() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestCredentialRejectValueField verifies that any non-nil value field is rejected
func TestCredentialRejectValueField(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create ecosystem for credential scope
	ecosystem := &models.Ecosystem{Name: "test-eco"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	tests := []struct {
		name    string
		cred    *models.CredentialDB
		wantErr bool
	}{
		{
			name: "reject keychain with value field",
			cred: &models.CredentialDB{
				ScopeType: models.CredentialScopeEcosystem,
				ScopeID:   int64(ecosystem.ID),
				Name:      "TEST_KEY",
				Source:    "keychain",
				Service:   stringPtr("test.service"),
				Value:     stringPtr("should-not-be-here"), // PROHIBITED
			},
			wantErr: true,
		},
		{
			name: "reject env with value field",
			cred: &models.CredentialDB{
				ScopeType: models.CredentialScopeEcosystem,
				ScopeID:   int64(ecosystem.ID),
				Name:      "TEST_ENV",
				Source:    "env",
				EnvVar:    stringPtr("TEST_ENV"),
				Value:     stringPtr("should-not-be-here"), // PROHIBITED
			},
			wantErr: true,
		},
		{
			name: "allow keychain without value",
			cred: &models.CredentialDB{
				ScopeType: models.CredentialScopeEcosystem,
				ScopeID:   int64(ecosystem.ID),
				Name:      "VALID_KEY",
				Source:    "keychain",
				Service:   stringPtr("valid.service"),
				Value:     nil, // Correct
			},
			wantErr: false,
		},
		{
			name: "allow env without value",
			cred: &models.CredentialDB{
				ScopeType: models.CredentialScopeEcosystem,
				ScopeID:   int64(ecosystem.ID),
				Name:      "VALID_ENV",
				Source:    "env",
				EnvVar:    stringPtr("VALID_ENV"),
				Value:     nil, // Correct
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// FIXME: This test will FAIL - CreateCredential() doesn't validate Value field yet
			// After Phase 3, CreateCredential should reject any non-nil Value field
			err := ds.CreateCredential(tt.cred)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateCredential() expected error for non-nil Value field, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("CreateCredential() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestCredentialAllowKeychain verifies keychain source with service succeeds
func TestCredentialAllowKeychain(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create ecosystem for credential scope
	ecosystem := &models.Ecosystem{Name: "test-eco"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	tests := []struct {
		name    string
		cred    *models.CredentialDB
		wantErr bool
		errMsg  string
	}{
		{
			name: "keychain with service",
			cred: &models.CredentialDB{
				ScopeType: models.CredentialScopeEcosystem,
				ScopeID:   int64(ecosystem.ID),
				Name:      "GITHUB_TOKEN",
				Source:    "keychain",
				Service:   stringPtr("github.com"),
			},
			wantErr: false,
		},
		{
			name: "keychain without service should fail",
			cred: &models.CredentialDB{
				ScopeType: models.CredentialScopeEcosystem,
				ScopeID:   int64(ecosystem.ID),
				Name:      "MISSING_SERVICE",
				Source:    "keychain",
				Service:   nil, // Missing required field
			},
			wantErr: true,
			errMsg:  "service required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// FIXME: This test may PASS now, but we need to verify validation
			err := ds.CreateCredential(tt.cred)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateCredential() expected error, got nil")
				} else if tt.errMsg != "" && !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errMsg)) {
					t.Errorf("CreateCredential() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("CreateCredential() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestCredentialAllowEnv verifies env source with env_var succeeds
func TestCredentialAllowEnv(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create ecosystem for credential scope
	ecosystem := &models.Ecosystem{Name: "test-eco"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	tests := []struct {
		name    string
		cred    *models.CredentialDB
		wantErr bool
		errMsg  string
	}{
		{
			name: "env with env_var",
			cred: &models.CredentialDB{
				ScopeType: models.CredentialScopeEcosystem,
				ScopeID:   int64(ecosystem.ID),
				Name:      "DATABASE_URL",
				Source:    "env",
				EnvVar:    stringPtr("DATABASE_URL"),
			},
			wantErr: false,
		},
		{
			name: "env without env_var should fail",
			cred: &models.CredentialDB{
				ScopeType: models.CredentialScopeEcosystem,
				ScopeID:   int64(ecosystem.ID),
				Name:      "MISSING_ENV",
				Source:    "env",
				EnvVar:    nil, // Missing required field
			},
			wantErr: true,
			errMsg:  "env_var required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// FIXME: This test may PASS now, but we need to verify validation
			err := ds.CreateCredential(tt.cred)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateCredential() expected error, got nil")
				} else if tt.errMsg != "" && !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errMsg)) {
					t.Errorf("CreateCredential() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("CreateCredential() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestCredentialValidationOnUpdate verifies update cannot change to plaintext
func TestCredentialValidationOnUpdate(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create ecosystem for credential scope
	ecosystem := &models.Ecosystem{Name: "test-eco"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Create a valid credential (keychain)
	cred := &models.CredentialDB{
		ScopeType: models.CredentialScopeEcosystem,
		ScopeID:   int64(ecosystem.ID),
		Name:      "TEST_CRED",
		Source:    "keychain",
		Service:   stringPtr("test.service"),
	}

	// FIXME: This will fail if CreateCredential validates properly
	// We're testing the UPDATE path here
	if err := ds.CreateCredential(cred); err != nil {
		t.Fatalf("Setup CreateCredential() error = %v", err)
	}

	tests := []struct {
		name       string
		updateFunc func(*models.CredentialDB)
		wantErr    bool
		errMsg     string
	}{
		{
			name: "update to plaintext source",
			updateFunc: func(c *models.CredentialDB) {
				c.Source = "value" // Try to change to plaintext
				c.Value = stringPtr("secret")
				c.Service = nil
			},
			wantErr: true,
			errMsg:  "plaintext",
		},
		{
			name: "update to add value field",
			updateFunc: func(c *models.CredentialDB) {
				// Keep keychain source but add value
				c.Value = stringPtr("should-not-work")
			},
			wantErr: true,
			errMsg:  "value",
		},
		{
			name: "valid update - change service",
			updateFunc: func(c *models.CredentialDB) {
				c.Service = stringPtr("new-service.com")
			},
			wantErr: false,
		},
		{
			name: "valid update - switch to env",
			updateFunc: func(c *models.CredentialDB) {
				c.Source = "env"
				c.Service = nil
				c.EnvVar = stringPtr("TEST_CRED_ENV")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy for this test
			testCred := *cred
			tt.updateFunc(&testCred)

			// FIXME: This test will FAIL - UpdateCredential() doesn't validate yet
			// After Phase 3, UpdateCredential should also validate
			err := ds.UpdateCredential(&testCred)

			if tt.wantErr {
				if err == nil {
					t.Errorf("UpdateCredential() expected error, got nil")
				} else if tt.errMsg != "" && !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errMsg)) {
					t.Errorf("UpdateCredential() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("UpdateCredential() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestCredentialDatabaseTriggerRejection tests that DB triggers reject plaintext
// This tests defense-in-depth: even if application code has a bug, DB should reject
func TestCredentialDatabaseTriggerRejection(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create ecosystem for credential scope
	ecosystem := &models.Ecosystem{Name: "test-eco"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Attempt to bypass application validation by inserting directly
	// FIXME: This test will FAIL - Database trigger doesn't exist yet
	// After Phase 3, database should have a trigger that prevents:
	// INSERT INTO credentials WHERE source = 'value' OR value IS NOT NULL
	query := `INSERT INTO credentials (scope_type, scope_id, name, source, value, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`

	_, err := ds.driver.Execute(query, "ecosystem", ecosystem.ID, "BYPASS_TEST", "value", "secret123")

	// Database trigger should prevent this
	if err == nil {
		t.Errorf("Direct INSERT with plaintext expected DB trigger to reject, got nil")
	}
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "check constraint") &&
		!strings.Contains(strings.ToLower(err.Error()), "trigger") &&
		!strings.Contains(strings.ToLower(err.Error()), "constraint") {
		t.Logf("Note: Got error %v, but not sure if it's from trigger", err)
	}
}

// TestCredentialOnlyKeychainAndEnvAllowed verifies CHECK constraint
func TestCredentialOnlyKeychainAndEnvAllowed(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create ecosystem for credential scope
	ecosystem := &models.Ecosystem{Name: "test-eco"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	tests := []struct {
		name       string
		source     string
		shouldFail bool
	}{
		{name: "keychain allowed", source: "keychain", shouldFail: false},
		{name: "env allowed", source: "env", shouldFail: false},
		{name: "value rejected", source: "value", shouldFail: true},
		{name: "plaintext rejected", source: "plaintext", shouldFail: true},
		{name: "file rejected", source: "file", shouldFail: true},
		{name: "vault rejected", source: "vault", shouldFail: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred := &models.CredentialDB{
				ScopeType: models.CredentialScopeEcosystem,
				ScopeID:   int64(ecosystem.ID),
				Name:      "TEST_" + tt.name,
				Source:    tt.source,
			}

			// Add required fields based on source
			if tt.source == "keychain" {
				cred.Service = stringPtr("test.service")
			} else if tt.source == "env" {
				cred.EnvVar = stringPtr("TEST_VAR")
			}

			// FIXME: This test will FAIL - CHECK constraint doesn't exist yet
			// After Phase 3, database schema should have:
			// CHECK (source IN ('keychain', 'env'))
			err := ds.CreateCredential(cred)

			if tt.shouldFail {
				if err == nil {
					t.Errorf("CreateCredential() with source=%q expected error, got nil", tt.source)
				}
			} else {
				if err != nil {
					t.Errorf("CreateCredential() with source=%q unexpected error = %v", tt.source, err)
				}
			}
		})
	}
}

// stringPtr is a helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
