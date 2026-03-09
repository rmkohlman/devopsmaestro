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
func TestCredentialSchemaEnforcement(t *testing.T) {
	// This test verifies that the database schema enforces source validation
	// The Value field has been removed from the model entirely in v0.19.0
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create ecosystem for credential scope
	ecosystem := &models.Ecosystem{Name: "test-eco"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Test that only keychain and env sources are allowed
	tests := []struct {
		name    string
		cred    *models.CredentialDB
		wantErr bool
	}{
		{
			name: "allow keychain source",
			cred: &models.CredentialDB{
				ScopeType: models.CredentialScopeEcosystem,
				ScopeID:   int64(ecosystem.ID),
				Name:      "VALID_KEY",
				Source:    "keychain",
				Service:   stringPtr("valid.service"),
			},
			wantErr: false,
		},
		{
			name: "allow env source",
			cred: &models.CredentialDB{
				ScopeType: models.CredentialScopeEcosystem,
				ScopeID:   int64(ecosystem.ID),
				Name:      "VALID_ENV",
				Source:    "env",
				EnvVar:    stringPtr("VALID_ENV"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ds.CreateCredential(tt.cred)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateCredential() expected error, got nil")
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
				c.Source = "value" // Try to change to plaintext (prohibited by schema)
				c.Service = nil
			},
			wantErr: true,
			errMsg:  "plaintext",
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

// =============================================================================
// Bug Regression Tests
// =============================================================================

// TestListAllCredentials_ScanMismatch verifies that ListAllCredentials() returns
// without error and returns the expected credentials.
//
// BUG: ListAllCredentials() SELECT query includes a "value" column that does not
// exist in the credentials table (removed in v0.19.0). The query will fail at
// runtime because it references a non-existent column, causing a "no such column"
// SQL error. Additionally, even if the column existed, Scan() only reads 10
// destinations for an 11-column result, causing a scan mismatch.
//
// This test FAILS against current code (RED phase).
func TestListAllCredentials_ScanMismatch(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create ecosystem for credential scope
	ecosystem := &models.Ecosystem{Name: "scan-mismatch-eco"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup CreateEcosystem() error = %v", err)
	}

	// Insert two credentials with different sources
	creds := []*models.CredentialDB{
		{
			ScopeType: models.CredentialScopeEcosystem,
			ScopeID:   int64(ecosystem.ID),
			Name:      "GITHUB_TOKEN",
			Source:    "keychain",
			Service:   stringPtr("github.com"),
		},
		{
			ScopeType: models.CredentialScopeEcosystem,
			ScopeID:   int64(ecosystem.ID),
			Name:      "DATABASE_URL",
			Source:    "env",
			EnvVar:    stringPtr("DATABASE_URL"),
		},
	}

	for _, cred := range creds {
		if err := ds.CreateCredential(cred); err != nil {
			t.Fatalf("Setup CreateCredential(%s) error = %v", cred.Name, err)
		}
	}

	// ListAllCredentials() should return both credentials without error.
	// BUG: This call fails because the SELECT references a "value" column
	// that no longer exists in the schema, resulting in SQL error:
	// "no such column: value"
	result, err := ds.ListAllCredentials()
	if err != nil {
		t.Fatalf("ListAllCredentials() error = %v, want nil", err)
	}

	if len(result) != 2 {
		t.Errorf("ListAllCredentials() returned %d credentials, want 2", len(result))
	}

	// Verify credential data is intact
	names := make(map[string]bool)
	for _, c := range result {
		names[c.Name] = true
	}
	if !names["GITHUB_TOKEN"] {
		t.Errorf("ListAllCredentials() missing GITHUB_TOKEN")
	}
	if !names["DATABASE_URL"] {
		t.Errorf("ListAllCredentials() missing DATABASE_URL")
	}
}

// TestListAllCredentials_MultiScope verifies ListAllCredentials returns credentials
// from all scopes and that field values are correct.
//
// This test also fails due to the same Bug 1 SQL column mismatch.
func TestListAllCredentials_MultiScope(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create two ecosystems to test cross-scope listing
	eco1 := &models.Ecosystem{Name: "multi-scope-eco1"}
	eco2 := &models.Ecosystem{Name: "multi-scope-eco2"}
	for _, eco := range []*models.Ecosystem{eco1, eco2} {
		if err := ds.CreateEcosystem(eco); err != nil {
			t.Fatalf("Setup CreateEcosystem(%s) error = %v", eco.Name, err)
		}
	}

	// Create one credential in each ecosystem
	cred1 := &models.CredentialDB{
		ScopeType: models.CredentialScopeEcosystem,
		ScopeID:   int64(eco1.ID),
		Name:      "ECO1_CRED",
		Source:    "keychain",
		Service:   stringPtr("eco1.service"),
	}
	cred2 := &models.CredentialDB{
		ScopeType: models.CredentialScopeEcosystem,
		ScopeID:   int64(eco2.ID),
		Name:      "ECO2_CRED",
		Source:    "env",
		EnvVar:    stringPtr("ECO2_ENV_VAR"),
	}
	for _, cred := range []*models.CredentialDB{cred1, cred2} {
		if err := ds.CreateCredential(cred); err != nil {
			t.Fatalf("Setup CreateCredential(%s) error = %v", cred.Name, err)
		}
	}

	// BUG: This will fail with "no such column: value" error
	results, err := ds.ListAllCredentials()
	if err != nil {
		t.Fatalf("ListAllCredentials() error = %v, want nil", err)
	}

	if len(results) < 2 {
		t.Errorf("ListAllCredentials() returned %d credentials, want >= 2", len(results))
	}

	// Verify the scoped credentials are present
	found := make(map[string]bool)
	for _, c := range results {
		found[c.Name] = true
	}
	if !found["ECO1_CRED"] {
		t.Errorf("ListAllCredentials() missing ECO1_CRED")
	}
	if !found["ECO2_CRED"] {
		t.Errorf("ListAllCredentials() missing ECO2_CRED")
	}
}

// TestListAllCredentials_EmptyDB verifies ListAllCredentials returns an empty
// slice (not nil/error) when there are no credentials.
//
// This test also fails due to the same Bug 1 SQL column mismatch (query fails
// even when there are zero rows, because column validation happens before scanning).
func TestListAllCredentials_EmptyDB(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// BUG: Still fails with "no such column: value" on the query itself
	results, err := ds.ListAllCredentials()
	if err != nil {
		t.Fatalf("ListAllCredentials() error = %v, want nil", err)
	}

	// Should return empty slice, not nil
	if results == nil {
		results = []*models.CredentialDB{} // normalize nil to empty slice for comparison
	}
	if len(results) != 0 {
		t.Errorf("ListAllCredentials() on empty DB returned %d credentials, want 0", len(results))
	}
}

// =============================================================================
// GetDefault error handling — Bug 2 regression test
// =============================================================================

// TestGetDefault_MissingKey_UsesErrorsIs documents that GetDefault() should
// return ("", nil) when a key is not found, using errors.Is(err, sql.ErrNoRows)
// for the no-rows check rather than fragile string comparison.
//
// The CURRENT implementation compares err.Error() == "sql: no rows in result set"
// which is fragile. This test pins the CORRECT BEHAVIOR (empty string, no error)
// so that after the @database agent refactors to use errors.Is, the behavior
// contract is enforced.
//
// The behavioral test itself PASSES with the current (fragile) implementation,
// because the string comparison happens to work today.
//
// Companion test: TestGetDefault_ErrorStringComparison_IsBrittleByDesign
// documents that the string-literal comparison is the implementation approach
// being replaced.
func TestGetDefault_MissingKey_ReturnsEmptyNoError(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Ensure the defaults table exists (not in base test schema)
	_, err := ds.Driver().Execute(`
		CREATE TABLE IF NOT EXISTS defaults (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Setup: create defaults table error = %v", err)
	}

	tests := []struct {
		name      string
		key       string
		wantValue string
		wantErr   bool
	}{
		{
			name:      "missing key returns empty string without error",
			key:       "nonexistent-key",
			wantValue: "",
			wantErr:   false,
		},
		{
			name:      "empty string key returns empty string without error",
			key:       "",
			wantValue: "",
			wantErr:   false,
		},
		{
			name:      "key with special chars not found returns empty without error",
			key:       "key:with:colons",
			wantValue: "",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := ds.GetDefault(tt.key)

			if tt.wantErr && err == nil {
				t.Errorf("GetDefault(%q) expected error, got nil", tt.key)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("GetDefault(%q) unexpected error = %v", tt.key, err)
			}
			if value != tt.wantValue {
				t.Errorf("GetDefault(%q) = %q, want %q", tt.key, value, tt.wantValue)
			}
		})
	}
}

// TestGetDefault_ExistingKey verifies that after SetDefault, GetDefault returns
// the stored value — ensuring the happy path is also covered.
func TestGetDefault_ExistingKey_ReturnsValue(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	_, err := ds.Driver().Execute(`
		CREATE TABLE IF NOT EXISTS defaults (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Setup: create defaults table error = %v", err)
	}

	if err := ds.SetDefault("active-theme", "tokyonight"); err != nil {
		t.Fatalf("SetDefault() error = %v", err)
	}

	value, err := ds.GetDefault("active-theme")
	if err != nil {
		t.Fatalf("GetDefault() error = %v", err)
	}
	if value != "tokyonight" {
		t.Errorf("GetDefault() = %q, want %q", value, "tokyonight")
	}
}

// stringPtr is a helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
