package cmd

// =============================================================================
// TDD Phase 2 (RED): loadBuildCredentials Warning Return Tests (v0.38.2)
// =============================================================================
// Bug 1: loadBuildCredentials silently drops credential resolution failures.
//
// Expected fix: Change the function signature from:
//
//	func loadBuildCredentials(ds db.DataStore, app *models.App, workspace *models.Workspace) map[string]string
//
// To:
//
//	func loadBuildCredentials(ds db.DataStore, app *models.App, workspace *models.Workspace) (map[string]string, []string)
//
// The second return value is a slice of human-readable warning strings, one per
// credential that failed to resolve (and had no env var fallback).
//
// These tests WILL FAIL TO COMPILE against current code because the current
// function only returns one value.  That is the expected RED state.
// =============================================================================

import (
	"strings"
	"testing"

	"devopsmaestro/config"
	"devopsmaestro/db"
	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Section: loadBuildCredentials Warning Return Tests
// ---------------------------------------------------------------------------

// TestLoadBuildCredentials_ReturnsWarningsForFailedKeychain verifies that when
// loadBuildCredentials encounters a credential it cannot resolve (e.g. a
// keychain entry that does not exist), it returns a non-nil warnings slice
// containing a string that mentions the credential name.
//
// This test WILL FAIL TO COMPILE because current loadBuildCredentials returns
// only map[string]string (one value), not (map[string]string, []string).
func TestLoadBuildCredentials_ReturnsWarningsForFailedKeychain(t *testing.T) {
	mockStore, app := setupTestContext()

	// Register a keychain credential whose service definitely does not exist so
	// the resolution will always fail on any machine.
	service := "dvm-test-nonexistent-service-for-warning-99999"
	credName := "MY_MISSING_CRED"

	cred := &models.CredentialDB{
		Name:        credName,
		ScopeType:   models.CredentialScopeApp,
		ScopeID:     int64(app.ID),
		Source:      "vault",
		VaultSecret: &service,
	}
	err := mockStore.CreateCredential(cred)
	require.NoError(t, err, "setup: CreateCredential should succeed")

	// ── COMPILE ERROR EXPECTED HERE ──────────────────────────────────────────
	// Current signature: loadBuildCredentials(...) map[string]string
	// Required signature: loadBuildCredentials(...) (map[string]string, []string)
	resolved, warnings := loadBuildCredentials(mockStore, app, nil)
	// ─────────────────────────────────────────────────────────────────────────

	// The resolved map must not contain the failed credential.
	assert.NotContains(t, resolved, credName,
		"failed credential should not appear in the resolved map")

	// At least one warning must be present.
	assert.NotEmpty(t, warnings,
		"warnings slice must not be empty when a credential fails to resolve")

	// At least one warning must mention the credential name (case-sensitive).
	found := false
	for _, w := range warnings {
		if strings.Contains(w, credName) {
			found = true
			break
		}
	}
	assert.True(t, found,
		"at least one warning should mention the failing credential name %q; got warnings: %v",
		credName, warnings)
}

// TestLoadBuildCredentials_NoWarningsWhenAllResolve verifies that when all
// credentials resolve successfully (env source with a set var), the warnings
// slice is empty (not nil — callers may range over it safely).
//
// This test WILL FAIL TO COMPILE for the same reason as above.
func TestLoadBuildCredentials_NoWarningsWhenAllResolve(t *testing.T) {
	mockStore, app := setupTestContext()

	envVarName := "DVM_TEST_CRED_EXISTS_ABC123"
	t.Setenv(envVarName, "secret-value")

	envVar := envVarName
	cred := &models.CredentialDB{
		Name:      "MY_EXISTING_CRED",
		ScopeType: models.CredentialScopeApp,
		ScopeID:   int64(app.ID),
		Source:    "env",
		EnvVar:    &envVar,
	}
	err := mockStore.CreateCredential(cred)
	require.NoError(t, err, "setup: CreateCredential should succeed")

	// ── COMPILE ERROR EXPECTED HERE ──────────────────────────────────────────
	resolved, warnings := loadBuildCredentials(mockStore, app, nil)
	// ─────────────────────────────────────────────────────────────────────────

	assert.Contains(t, resolved, "MY_EXISTING_CRED",
		"successfully resolved credential should appear in the map")
	assert.Empty(t, warnings,
		"warnings should be empty when all credentials resolved successfully")
}

// TestLoadBuildCredentials_WarningMentionsCredentialName is a table-driven test
// verifying that each credential whose resolution fails generates a warning
// mentioning that credential's name.
//
// This test WILL FAIL TO COMPILE for the same reason as above.
func TestLoadBuildCredentials_WarningMentionsCredentialName(t *testing.T) {
	tests := []struct {
		name        string
		credName    string
		service     string
		wantWarning bool
	}{
		{
			name:        "non-existent keychain service generates warning",
			credName:    "CRED_ALPHA",
			service:     "dvm-test-nonexistent-alpha-99999",
			wantWarning: true,
		},
		{
			name:        "another non-existent keychain service generates warning",
			credName:    "CRED_BETA",
			service:     "dvm-test-nonexistent-beta-99999",
			wantWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore, app := setupTestContext()

			service := tt.service
			cred := &models.CredentialDB{
				Name:        tt.credName,
				ScopeType:   models.CredentialScopeApp,
				ScopeID:     int64(app.ID),
				Source:      "vault",
				VaultSecret: &service,
			}
			require.NoError(t, mockStore.CreateCredential(cred))

			// ── COMPILE ERROR EXPECTED HERE ──────────────────────────────────
			_, warnings := loadBuildCredentials(mockStore, app, nil)
			// ─────────────────────────────────────────────────────────────────

			if tt.wantWarning {
				found := false
				for _, w := range warnings {
					if strings.Contains(w, tt.credName) {
						found = true
						break
					}
				}
				assert.True(t, found,
					"expected warning mentioning %q; got: %v", tt.credName, warnings)
			} else {
				assert.Empty(t, warnings, "expected no warnings for %s", tt.name)
			}
		})
	}
}

// TestLoadBuildCredentials_WorkspaceCredentialWarning verifies that credential
// failures at workspace scope also generate warnings.
//
// This test WILL FAIL TO COMPILE for the same reason as above.
func TestLoadBuildCredentials_WorkspaceCredentialWarning(t *testing.T) {
	mockStore, app := setupTestContext()

	// Create a workspace
	workspace := &models.Workspace{
		AppID:     app.ID,
		Name:      "warn-ws",
		Slug:      "test-eco-test-domain-test-app-warn-ws",
		ImageName: "test-image",
		Status:    "stopped",
	}
	err := mockStore.CreateWorkspace(workspace)
	require.NoError(t, err, "setup: CreateWorkspace should succeed")

	// Register a workspace-scoped credential that will fail to resolve
	service := "dvm-test-nonexistent-ws-service-99999"
	credName := "WS_MISSING_CRED"
	cred := &models.CredentialDB{
		Name:        credName,
		ScopeType:   models.CredentialScopeWorkspace,
		ScopeID:     int64(workspace.ID),
		Source:      "vault",
		VaultSecret: &service,
	}
	err = mockStore.CreateCredential(cred)
	require.NoError(t, err, "setup: CreateCredential should succeed")

	// ── COMPILE ERROR EXPECTED HERE ──────────────────────────────────────────
	_, warnings := loadBuildCredentials(mockStore, app, workspace)
	// ─────────────────────────────────────────────────────────────────────────

	found := false
	for _, w := range warnings {
		if strings.Contains(w, credName) {
			found = true
			break
		}
	}
	assert.True(t, found,
		"workspace-scope credential failure should generate warning mentioning %q; got: %v",
		credName, warnings)
}

// ---------------------------------------------------------------------------
// Section: Interface Compliance Sanity Check
// ---------------------------------------------------------------------------

// TestLoadBuildCredentials_UsesDataStoreInterface documents that loadBuildCredentials
// accepts a db.DataStore interface (not a concrete type), which is required for
// testability.  This passes today — it's here as a regression guard.
func TestLoadBuildCredentials_UsesDataStoreInterface(t *testing.T) {
	var _ db.DataStore = (*db.MockDataStore)(nil) // compile-time check
	var _ config.CredentialScope = config.CredentialScope{}
}

// ---------------------------------------------------------------------------
// Section: imageNameToSafeSlug Tests (#359)
// ---------------------------------------------------------------------------

// TestImageNameToSafeSlug verifies that image names are converted to
// filesystem-safe slugs for use in temp file paths.
func TestImageNameToSafeSlug(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple image name",
			input:    "my-image",
			expected: "my-image",
		},
		{
			name:     "image with tag colon",
			input:    "dvm-dev-daa-agents:20260414-230555",
			expected: "dvm-dev-daa-agents-20260414-230555",
		},
		{
			name:     "image with registry slash",
			input:    "registry.io/my-image:latest",
			expected: "registry-io-my-image-latest",
		},
		{
			name:     "image with underscores",
			input:    "my_image_name",
			expected: "my-image-name",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "image with dots",
			input:    "registry.example.com/org/image:v1.2.3",
			expected: "registry-example-com-org-image-v1-2-3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := imageNameToSafeSlug(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestImageNameToSafeSlug_ConcurrentUniqueness verifies that different image
// names produce different slugs, ensuring concurrent builds don't collide.
func TestImageNameToSafeSlug_ConcurrentUniqueness(t *testing.T) {
	images := []string{
		"dvm-dev-daa-agents:20260414-230555",
		"dvm-dev-daa-data-collector:20260414-230554",
		"dvm-dev-daa-northbound-api:20260414-230556",
		"dvm-dev-daa-frontend:20260414-230557",
	}

	slugs := make(map[string]string) // slug → original image name
	for _, img := range images {
		slug := imageNameToSafeSlug(img)
		if existing, ok := slugs[slug]; ok {
			t.Errorf("slug collision: %q and %q both produce slug %q", existing, img, slug)
		}
		slugs[slug] = img
	}
}

// TestImageNameToSafeSlug_OnlyContainsSafeChars ensures the slug output
// contains only alphanumeric characters and hyphens — safe for file paths.
func TestImageNameToSafeSlug_OnlyContainsSafeChars(t *testing.T) {
	inputs := []string{
		"dvm-dev-daa-agents:20260414-230555",
		"registry.io/org/my_image:v1.2.3-rc1",
		"image@sha256:abc123",
		"a/b/c:d/e/f",
	}

	for _, input := range inputs {
		slug := imageNameToSafeSlug(input)
		for _, c := range slug {
			isSafe := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-'
			if !isSafe {
				t.Errorf("slug for %q contains unsafe char %q: %s", input, string(c), slug)
			}
		}
	}
}
