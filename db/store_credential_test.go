package db

// === Dual-Field Credential DB Tests (v0.37.1) ===
//
// These tests verify that the DB store methods correctly read/write the new
// username_var and password_var columns added in migration 010.
//
// RED PHASE: These tests DO NOT COMPILE until the following are added:
//   - models.CredentialDB.UsernameVar  (*string, db:"username_var")
//   - models.CredentialDB.PasswordVar  (*string, db:"password_var")
//
// They also fail at runtime until store_credential.go is updated to:
//   - include username_var/password_var in INSERT (CreateCredential)
//   - include username_var/password_var in SELECT + Scan (Get*, List*)
//   - include username_var/password_var in UPDATE SET (UpdateCredential)

import (
	"devopsmaestro/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------------
// Test Setup
// -----------------------------------------------------------------------------

// createTestCredentialStore creates an in-memory SQLite DataStore whose
// credentials table already contains the migration-010 columns
// (username_var, password_var).  Only the tables required by these tests
// are created so the helper stays lightweight and self-contained.
func createTestCredentialStore(t *testing.T) *SQLDataStore {
	t.Helper()

	cfg := DriverConfig{Type: DriverMemory}
	driver, err := NewMemorySQLiteDriver(cfg)
	require.NoError(t, err)
	require.NoError(t, driver.Connect())

	stmts := []string{
		// Ecosystems — needed for scope_type='ecosystem' credentials
		`CREATE TABLE ecosystems (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			name        TEXT NOT NULL UNIQUE,
			description TEXT,
			theme       TEXT,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Credentials — includes the post-migration-010 columns
		`CREATE TABLE credentials (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			scope_type   TEXT NOT NULL CHECK(scope_type IN ('ecosystem','domain','app','workspace')),
			scope_id     INTEGER NOT NULL,
			name         TEXT NOT NULL,
			source       TEXT NOT NULL CHECK(source IN ('keychain','env')),
			service      TEXT,
			env_var      TEXT,
			description  TEXT,
			username_var TEXT,
			password_var TEXT,
			created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(scope_type, scope_id, name)
		)`,
	}

	for _, stmt := range stmts {
		_, err := driver.Execute(stmt)
		require.NoError(t, err, "failed to execute schema statement: %s", stmt)
	}

	return NewSQLDataStore(driver, nil)
}

// strPtr is a helper that takes a string value and returns a pointer to it.
func strPtr(s string) *string { return &s }

// -----------------------------------------------------------------------------
// TestSQLDataStore_CreateCredential_DualField
//
// Creates a keychain credential that has both UsernameVar and PasswordVar set.
// After creation, GetCredential must return those values intact.
// -----------------------------------------------------------------------------

func TestSQLDataStore_CreateCredential_DualField(t *testing.T) {
	ds := createTestCredentialStore(t)
	defer ds.Close()

	// Arrange: insert a scope entity so the FK-like scope_id makes semantic sense
	eco := &models.Ecosystem{Name: "df-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))

	cred := &models.CredentialDB{
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     int64(eco.ID),
		Name:        "gh-dual",
		Source:      "keychain",
		Service:     strPtr("github.com"),
		UsernameVar: strPtr("GH_USER"), // NEW — does not exist yet → compile error
		PasswordVar: strPtr("GH_PAT"),  // NEW — does not exist yet → compile error
	}

	// Act
	err := ds.CreateCredential(cred)

	// Assert — create must succeed
	require.NoError(t, err)
	assert.NotZero(t, cred.ID, "CreateCredential must populate ID")

	// Read back via GetCredential
	got, err := ds.GetCredential(models.CredentialScopeEcosystem, int64(eco.ID), "gh-dual")
	require.NoError(t, err)

	require.NotNil(t, got.UsernameVar, "UsernameVar must not be nil after round-trip")
	assert.Equal(t, "GH_USER", *got.UsernameVar)

	require.NotNil(t, got.PasswordVar, "PasswordVar must not be nil after round-trip")
	assert.Equal(t, "GH_PAT", *got.PasswordVar)
}

// -----------------------------------------------------------------------------
// TestSQLDataStore_CreateCredential_DualField_NilVars
//
// Creates a legacy credential (no UsernameVar/PasswordVar).
// After creation, GetCredential must return nil for both new fields.
// -----------------------------------------------------------------------------

func TestSQLDataStore_CreateCredential_DualField_NilVars(t *testing.T) {
	ds := createTestCredentialStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "nil-vars-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))

	cred := &models.CredentialDB{
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     int64(eco.ID),
		Name:        "legacy-keychain",
		Source:      "keychain",
		Service:     strPtr("legacy.service"),
		UsernameVar: nil, // intentionally absent
		PasswordVar: nil, // intentionally absent
	}

	require.NoError(t, ds.CreateCredential(cred))

	got, err := ds.GetCredential(models.CredentialScopeEcosystem, int64(eco.ID), "legacy-keychain")
	require.NoError(t, err)

	assert.Nil(t, got.UsernameVar, "legacy credential must have nil UsernameVar")
	assert.Nil(t, got.PasswordVar, "legacy credential must have nil PasswordVar")
}

// -----------------------------------------------------------------------------
// TestSQLDataStore_GetCredentialByName_DualField
//
// Creates a dual-field credential, then retrieves it via GetCredentialByName.
// Both UsernameVar and PasswordVar must be populated.
// -----------------------------------------------------------------------------

func TestSQLDataStore_GetCredentialByName_DualField(t *testing.T) {
	ds := createTestCredentialStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "byname-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))

	cred := &models.CredentialDB{
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     int64(eco.ID),
		Name:        "npm-registry",
		Source:      "keychain",
		Service:     strPtr("registry.npmjs.org"),
		UsernameVar: strPtr("NPM_USER"),
		PasswordVar: strPtr("NPM_TOKEN"),
	}
	require.NoError(t, ds.CreateCredential(cred))

	// Act — retrieve by name (cross-scope lookup)
	got, err := ds.GetCredentialByName("npm-registry")
	require.NoError(t, err)

	assert.Equal(t, "npm-registry", got.Name)

	require.NotNil(t, got.UsernameVar, "GetCredentialByName must return UsernameVar")
	assert.Equal(t, "NPM_USER", *got.UsernameVar)

	require.NotNil(t, got.PasswordVar, "GetCredentialByName must return PasswordVar")
	assert.Equal(t, "NPM_TOKEN", *got.PasswordVar)
}

// -----------------------------------------------------------------------------
// TestSQLDataStore_UpdateCredential_DualField
//
// Creates a legacy credential (nil vars), then updates it to set UsernameVar
// and PasswordVar.  A subsequent GetCredential must return the updated values.
// -----------------------------------------------------------------------------

func TestSQLDataStore_UpdateCredential_DualField(t *testing.T) {
	ds := createTestCredentialStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "update-dual-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))

	// Create a legacy credential (no dual-field vars)
	cred := &models.CredentialDB{
		ScopeType: models.CredentialScopeEcosystem,
		ScopeID:   int64(eco.ID),
		Name:      "docker-hub",
		Source:    "keychain",
		Service:   strPtr("hub.docker.com"),
	}
	require.NoError(t, ds.CreateCredential(cred))

	// Verify starting state: both vars are nil
	before, err := ds.GetCredential(models.CredentialScopeEcosystem, int64(eco.ID), "docker-hub")
	require.NoError(t, err)
	assert.Nil(t, before.UsernameVar, "pre-update: UsernameVar should be nil")
	assert.Nil(t, before.PasswordVar, "pre-update: PasswordVar should be nil")

	// Act — update to add dual-field vars
	before.UsernameVar = strPtr("DOCKER_USER")
	before.PasswordVar = strPtr("DOCKER_TOKEN")
	require.NoError(t, ds.UpdateCredential(before))

	// Assert — values must persist after re-read
	after, err := ds.GetCredential(models.CredentialScopeEcosystem, int64(eco.ID), "docker-hub")
	require.NoError(t, err)

	require.NotNil(t, after.UsernameVar, "post-update: UsernameVar must not be nil")
	assert.Equal(t, "DOCKER_USER", *after.UsernameVar)

	require.NotNil(t, after.PasswordVar, "post-update: PasswordVar must not be nil")
	assert.Equal(t, "DOCKER_TOKEN", *after.PasswordVar)
}

// -----------------------------------------------------------------------------
// TestSQLDataStore_ListCredentialsByScope_DualField
//
// Creates two credentials in the same scope — one legacy and one dual-field.
// ListCredentialsByScope must return both, and the dual-field one must have its
// UsernameVar and PasswordVar correctly populated.
// -----------------------------------------------------------------------------

func TestSQLDataStore_ListCredentialsByScope_DualField(t *testing.T) {
	ds := createTestCredentialStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "list-scope-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))
	scopeID := int64(eco.ID)

	// Legacy credential — no dual-field vars
	legacy := &models.CredentialDB{
		ScopeType: models.CredentialScopeEcosystem,
		ScopeID:   scopeID,
		Name:      "legacy-svc",
		Source:    "keychain",
		Service:   strPtr("legacy.internal"),
	}
	require.NoError(t, ds.CreateCredential(legacy))

	// Dual-field credential
	dual := &models.CredentialDB{
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     scopeID,
		Name:        "github-svc",
		Source:      "keychain",
		Service:     strPtr("github.com"),
		UsernameVar: strPtr("GH_USER"),
		PasswordVar: strPtr("GH_PAT"),
	}
	require.NoError(t, ds.CreateCredential(dual))

	// Act
	results, err := ds.ListCredentialsByScope(models.CredentialScopeEcosystem, scopeID)
	require.NoError(t, err)
	require.Len(t, results, 2, "expected 2 credentials in scope")

	// Build a lookup by name
	byName := make(map[string]*models.CredentialDB, len(results))
	for _, c := range results {
		byName[c.Name] = c
	}

	// Legacy credential must have nil vars
	legacyResult, ok := byName["legacy-svc"]
	require.True(t, ok, "legacy-svc must appear in list")
	assert.Nil(t, legacyResult.UsernameVar, "legacy credential: UsernameVar must be nil")
	assert.Nil(t, legacyResult.PasswordVar, "legacy credential: PasswordVar must be nil")

	// Dual-field credential must have vars populated
	dualResult, ok := byName["github-svc"]
	require.True(t, ok, "github-svc must appear in list")
	require.NotNil(t, dualResult.UsernameVar, "dual credential: UsernameVar must not be nil")
	assert.Equal(t, "GH_USER", *dualResult.UsernameVar)
	require.NotNil(t, dualResult.PasswordVar, "dual credential: PasswordVar must not be nil")
	assert.Equal(t, "GH_PAT", *dualResult.PasswordVar)
}

// -----------------------------------------------------------------------------
// TestSQLDataStore_ListAllCredentials_DualField
//
// Creates 3 credentials across two ecosystem scopes:
//   - eco1/legacy-one   (no dual vars)
//   - eco1/dual-one     (UsernameVar + PasswordVar set)
//   - eco2/legacy-two   (no dual vars)
//
// ListAllCredentials must return all 3, and dual-one must have its vars
// populated while the legacy entries must have nil vars.
// -----------------------------------------------------------------------------

func TestSQLDataStore_ListAllCredentials_DualField(t *testing.T) {
	ds := createTestCredentialStore(t)
	defer ds.Close()

	// Create two ecosystems
	eco1 := &models.Ecosystem{Name: "listall-eco1"}
	require.NoError(t, ds.CreateEcosystem(eco1))

	eco2 := &models.Ecosystem{Name: "listall-eco2"}
	require.NoError(t, ds.CreateEcosystem(eco2))

	// eco1 — one legacy, one dual-field
	legacyOne := &models.CredentialDB{
		ScopeType: models.CredentialScopeEcosystem,
		ScopeID:   int64(eco1.ID),
		Name:      "legacy-one",
		Source:    "env",
		EnvVar:    strPtr("LEGACY_ONE_VAR"),
	}
	require.NoError(t, ds.CreateCredential(legacyOne))

	dualOne := &models.CredentialDB{
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     int64(eco1.ID),
		Name:        "dual-one",
		Source:      "keychain",
		Service:     strPtr("dual.svc"),
		UsernameVar: strPtr("DUAL_USER"),
		PasswordVar: strPtr("DUAL_PASS"),
	}
	require.NoError(t, ds.CreateCredential(dualOne))

	// eco2 — one legacy
	legacyTwo := &models.CredentialDB{
		ScopeType: models.CredentialScopeEcosystem,
		ScopeID:   int64(eco2.ID),
		Name:      "legacy-two",
		Source:    "env",
		EnvVar:    strPtr("LEGACY_TWO_VAR"),
	}
	require.NoError(t, ds.CreateCredential(legacyTwo))

	// Act
	all, err := ds.ListAllCredentials()
	require.NoError(t, err)
	assert.Len(t, all, 3, "expected 3 credentials across all scopes")

	// Build lookup by name
	byName := make(map[string]*models.CredentialDB, len(all))
	for _, c := range all {
		byName[c.Name] = c
	}

	// legacy-one — nil vars
	l1, ok := byName["legacy-one"]
	require.True(t, ok, "legacy-one must appear in ListAllCredentials")
	assert.Nil(t, l1.UsernameVar, "legacy-one: UsernameVar must be nil")
	assert.Nil(t, l1.PasswordVar, "legacy-one: PasswordVar must be nil")

	// dual-one — vars populated
	d1, ok := byName["dual-one"]
	require.True(t, ok, "dual-one must appear in ListAllCredentials")
	require.NotNil(t, d1.UsernameVar, "dual-one: UsernameVar must not be nil")
	assert.Equal(t, "DUAL_USER", *d1.UsernameVar)
	require.NotNil(t, d1.PasswordVar, "dual-one: PasswordVar must not be nil")
	assert.Equal(t, "DUAL_PASS", *d1.PasswordVar)

	// legacy-two — nil vars
	l2, ok := byName["legacy-two"]
	require.True(t, ok, "legacy-two must appear in ListAllCredentials")
	assert.Nil(t, l2.UsernameVar, "legacy-two: UsernameVar must be nil")
	assert.Nil(t, l2.PasswordVar, "legacy-two: PasswordVar must be nil")
}
