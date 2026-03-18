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
			build_args  TEXT,
			ca_certs    TEXT,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Credentials — includes the post-migration-013 columns (vault_fields)
		`CREATE TABLE credentials (
			id                   INTEGER PRIMARY KEY AUTOINCREMENT,
			scope_type           TEXT NOT NULL CHECK(scope_type IN ('ecosystem','domain','app','workspace')),
			scope_id             INTEGER NOT NULL,
			name                 TEXT NOT NULL,
			source               TEXT NOT NULL CHECK(source IN ('vault','env')),
			env_var              TEXT,
			description          TEXT,
			username_var         TEXT,
			password_var         TEXT,
			vault_secret         TEXT,
			vault_env            TEXT,
			vault_username_secret TEXT,
			vault_fields         TEXT,
			created_at           DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at           DATETIME DEFAULT CURRENT_TIMESTAMP,
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
		Source:      "vault",
		VaultSecret: strPtr("github.com"),
		UsernameVar: strPtr("GH_USER"),
		PasswordVar: strPtr("GH_PAT"),
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
		Source:      "vault",
		VaultSecret: strPtr("legacy.service"),
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
		Source:      "vault",
		VaultSecret: strPtr("registry.npmjs.org"),
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
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     int64(eco.ID),
		Name:        "docker-hub",
		Source:      "vault",
		VaultSecret: strPtr("hub.docker.com"),
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
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     scopeID,
		Name:        "legacy-svc",
		Source:      "vault",
		VaultSecret: strPtr("legacy.internal"),
	}
	require.NoError(t, ds.CreateCredential(legacy))

	// Dual-field credential
	dual := &models.CredentialDB{
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     scopeID,
		Name:        "github-svc",
		Source:      "vault",
		VaultSecret: strPtr("github.com"),
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

// =============================================================================
// TDD Phase 2 (RED): MaestroVault DB Tests (v0.40.0)
// =============================================================================
// These tests verify that the DB store methods correctly read/write the new
// vault_secret, vault_env, and vault_username_secret columns added in
// migration 012 (MaestroVault integration).
//
// RED PHASE: These tests DO NOT COMPILE until the following are added:
//   - models.CredentialDB.VaultSecret         (*string, db:"vault_secret")
//   - models.CredentialDB.VaultEnv            (*string, db:"vault_env")
//   - models.CredentialDB.VaultUsernameSecret (*string, db:"vault_username_secret")
//
// They also fail at runtime until:
//   - credentials table source CHECK is updated to ('vault','env')
//   - store_credential.go INSERT/SELECT/UPDATE includes vault columns
// =============================================================================

// createTestVaultCredentialStore creates an in-memory SQLite DataStore whose
// credentials table has the post-migration-012 schema (vault columns, vault
// source in CHECK constraint).
func createTestVaultCredentialStore(t *testing.T) *SQLDataStore {
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
			build_args  TEXT,
			ca_certs    TEXT,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Credentials — post-migration-013: vault columns + vault_fields
		`CREATE TABLE credentials (
			id                   INTEGER PRIMARY KEY AUTOINCREMENT,
			scope_type           TEXT NOT NULL CHECK(scope_type IN ('ecosystem','domain','app','workspace')),
			scope_id             INTEGER NOT NULL,
			name                 TEXT NOT NULL,
			source               TEXT NOT NULL CHECK(source IN ('vault','env')),
			env_var              TEXT,
			description          TEXT,
			username_var         TEXT,
			password_var         TEXT,
			vault_secret         TEXT,
			vault_env            TEXT,
			vault_username_secret TEXT,
			vault_fields         TEXT,
			created_at           DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at           DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(scope_type, scope_id, name)
		)`,
	}

	for _, stmt := range stmts {
		_, err := driver.Execute(stmt)
		require.NoError(t, err, "failed to execute schema statement: %s", stmt)
	}

	return NewSQLDataStore(driver, nil)
}

// -----------------------------------------------------------------------------
// TestSQLDataStore_CreateCredential_VaultSource
//
// Creates a vault credential with vault_secret set.
// After creation, GetCredential must return the vault fields correctly.
// -----------------------------------------------------------------------------

func TestSQLDataStore_CreateCredential_VaultSource(t *testing.T) {
	ds := createTestVaultCredentialStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "vault-source-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))

	vaultSecret := "my-org/github-pat"

	// ── COMPILE FAILURE EXPECTED BELOW ───────────────────────────────────────
	// models.CredentialDB.VaultSecret does not exist yet.
	cred := &models.CredentialDB{
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     int64(eco.ID),
		Name:        "github-pat",
		Source:      "vault",
		VaultSecret: strPtr(vaultSecret), // NEW field — does not exist yet → compile error
	}
	// ─────────────────────────────────────────────────────────────────────────

	err := ds.CreateCredential(cred)
	require.NoError(t, err, "CreateCredential must succeed for vault source credential")
	assert.NotZero(t, cred.ID, "CreateCredential must populate ID")

	got, err := ds.GetCredential(models.CredentialScopeEcosystem, int64(eco.ID), "github-pat")
	require.NoError(t, err)

	assert.Equal(t, "vault", got.Source)
	require.NotNil(t, got.VaultSecret, "VaultSecret must not be nil after round-trip")
	assert.Equal(t, "my-org/github-pat", *got.VaultSecret)
	assert.Nil(t, got.VaultEnv, "VaultEnv must be nil when not set")
	assert.Nil(t, got.VaultUsernameSecret, "VaultUsernameSecret must be nil when not set")
}

// -----------------------------------------------------------------------------
// TestSQLDataStore_CreateCredential_VaultSource_RequiresVaultSecret
//
// Attempts to create a vault credential with VaultSecret=nil.
// The DB store must return an error (vault source requires vault_secret).
// -----------------------------------------------------------------------------

func TestSQLDataStore_CreateCredential_VaultSource_RequiresVaultSecret(t *testing.T) {
	ds := createTestVaultCredentialStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "vault-required-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))

	// ── COMPILE FAILURE EXPECTED BELOW ───────────────────────────────────────
	// models.CredentialDB.VaultSecret does not exist yet.
	cred := &models.CredentialDB{
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     int64(eco.ID),
		Name:        "no-vault-secret",
		Source:      "vault",
		VaultSecret: nil, // intentionally absent — should fail
	}
	// ─────────────────────────────────────────────────────────────────────────

	err := ds.CreateCredential(cred)
	assert.Error(t, err,
		"CreateCredential must return error when Source=vault and VaultSecret is nil")
}

// -----------------------------------------------------------------------------
// TestSQLDataStore_GetCredential_VaultFields
//
// Creates a vault credential with all vault fields set, then retrieves it.
// All vault fields must be populated correctly after a round-trip.
// -----------------------------------------------------------------------------

func TestSQLDataStore_GetCredential_VaultFields(t *testing.T) {
	ds := createTestVaultCredentialStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "vault-fields-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))

	// ── COMPILE FAILURE EXPECTED BELOW ───────────────────────────────────────
	// VaultSecret, VaultEnv, VaultUsernameSecret do not exist yet.
	cred := &models.CredentialDB{
		ScopeType:           models.CredentialScopeEcosystem,
		ScopeID:             int64(eco.ID),
		Name:                "full-vault-cred",
		Source:              "vault",
		VaultSecret:         strPtr("my-org/api-key"),
		VaultEnv:            strPtr("production"),
		VaultUsernameSecret: strPtr("my-org/api-username"),
	}
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, ds.CreateCredential(cred))

	got, err := ds.GetCredential(models.CredentialScopeEcosystem, int64(eco.ID), "full-vault-cred")
	require.NoError(t, err)

	assert.Equal(t, "vault", got.Source)

	require.NotNil(t, got.VaultSecret, "VaultSecret must not be nil after round-trip")
	assert.Equal(t, "my-org/api-key", *got.VaultSecret)

	require.NotNil(t, got.VaultEnv, "VaultEnv must not be nil after round-trip")
	assert.Equal(t, "production", *got.VaultEnv)

	require.NotNil(t, got.VaultUsernameSecret, "VaultUsernameSecret must not be nil after round-trip")
	assert.Equal(t, "my-org/api-username", *got.VaultUsernameSecret)
}

// -----------------------------------------------------------------------------
// TestSQLDataStore_UpdateCredential_VaultFields
//
// Creates a vault credential, then updates the vault_env field.
// The updated field must persist after re-read via GetCredential.
// -----------------------------------------------------------------------------

func TestSQLDataStore_UpdateCredential_VaultFields(t *testing.T) {
	ds := createTestVaultCredentialStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "vault-update-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))

	// ── COMPILE FAILURE EXPECTED BELOW ───────────────────────────────────────
	// VaultSecret does not exist yet.
	cred := &models.CredentialDB{
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     int64(eco.ID),
		Name:        "updateable-cred",
		Source:      "vault",
		VaultSecret: strPtr("my-org/db-password"),
		// VaultEnv intentionally nil initially
	}
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, ds.CreateCredential(cred))

	// Verify starting state: VaultEnv is nil
	before, err := ds.GetCredential(models.CredentialScopeEcosystem, int64(eco.ID), "updateable-cred")
	require.NoError(t, err)
	assert.Nil(t, before.VaultEnv, "pre-update: VaultEnv should be nil")

	// ── COMPILE FAILURE EXPECTED BELOW ───────────────────────────────────────
	// VaultEnv does not exist yet.
	before.VaultEnv = strPtr("staging")
	// ─────────────────────────────────────────────────────────────────────────
	require.NoError(t, ds.UpdateCredential(before))

	after, err := ds.GetCredential(models.CredentialScopeEcosystem, int64(eco.ID), "updateable-cred")
	require.NoError(t, err)

	require.NotNil(t, after.VaultEnv, "post-update: VaultEnv must not be nil")
	assert.Equal(t, "staging", *after.VaultEnv)
}

// -----------------------------------------------------------------------------
// TestSQLDataStore_CreateCredential_KeychainSourceRejected
//
// Attempts to create a credential with Source="keychain".
// The DB store must reject this — the CHECK constraint now only allows
// 'vault' and 'env'.
// -----------------------------------------------------------------------------

func TestSQLDataStore_CreateCredential_KeychainSourceRejected(t *testing.T) {
	ds := createTestVaultCredentialStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "keychain-reject-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))

	// Attempt to insert with the old "keychain" source value.
	// The CHECK constraint CHECK(source IN ('vault','env')) must reject this.
	cred := &models.CredentialDB{
		ScopeType: models.CredentialScopeEcosystem,
		ScopeID:   int64(eco.ID),
		Name:      "old-keychain-cred",
		Source:    "keychain", // INVALID in the new schema
	}

	err := ds.CreateCredential(cred)
	assert.Error(t, err,
		"CreateCredential must return error when Source=keychain (deprecated, use vault)")
}

// -----------------------------------------------------------------------------
// TestSQLDataStore_ListCredentialsByScope_VaultFields
//
// Creates vault and env credentials in the same scope.
// ListCredentialsByScope must return both and vault fields must be populated.
// -----------------------------------------------------------------------------

func TestSQLDataStore_ListCredentialsByScope_VaultFields(t *testing.T) {
	ds := createTestVaultCredentialStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "vault-list-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))
	scopeID := int64(eco.ID)

	// env credential — no vault fields
	envCred := &models.CredentialDB{
		ScopeType: models.CredentialScopeEcosystem,
		ScopeID:   scopeID,
		Name:      "env-cred",
		Source:    "env",
		EnvVar:    strPtr("MY_TOKEN"),
	}
	require.NoError(t, ds.CreateCredential(envCred))

	// ── COMPILE FAILURE EXPECTED BELOW ───────────────────────────────────────
	// VaultSecret does not exist yet.
	vaultCred := &models.CredentialDB{
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     scopeID,
		Name:        "vault-cred",
		Source:      "vault",
		VaultSecret: strPtr("my-org/api-key"),
		VaultEnv:    strPtr("production"),
	}
	// ─────────────────────────────────────────────────────────────────────────
	require.NoError(t, ds.CreateCredential(vaultCred))

	results, err := ds.ListCredentialsByScope(models.CredentialScopeEcosystem, scopeID)
	require.NoError(t, err)
	require.Len(t, results, 2, "expected 2 credentials in scope")

	byName := make(map[string]*models.CredentialDB, len(results))
	for _, c := range results {
		byName[c.Name] = c
	}

	// env credential: source=env, no vault fields
	envResult, ok := byName["env-cred"]
	require.True(t, ok, "env-cred must appear in list")
	assert.Equal(t, "env", envResult.Source)
	assert.Nil(t, envResult.VaultSecret, "env credential: VaultSecret must be nil")
	assert.Nil(t, envResult.VaultEnv, "env credential: VaultEnv must be nil")

	// vault credential: source=vault, vault fields populated
	vaultResult, ok := byName["vault-cred"]
	require.True(t, ok, "vault-cred must appear in list")
	assert.Equal(t, "vault", vaultResult.Source)
	require.NotNil(t, vaultResult.VaultSecret, "vault credential: VaultSecret must not be nil")
	assert.Equal(t, "my-org/api-key", *vaultResult.VaultSecret)
	require.NotNil(t, vaultResult.VaultEnv, "vault credential: VaultEnv must not be nil")
	assert.Equal(t, "production", *vaultResult.VaultEnv)
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
		Source:      "vault",
		VaultSecret: strPtr("dual.svc"),
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

// =============================================================================
// TDD Phase 2 (RED): VaultFields DB Tests (v0.41.0)
// =============================================================================
// These tests verify that the DB store methods correctly read/write the new
// vault_fields column added in migration 013.
//
// RED PHASE: These tests DO NOT COMPILE until the following are added:
//   - models.CredentialDB.VaultFields (*string, db:"vault_fields")
//
// They also fail at runtime until store_credential.go is updated to:
//   - include vault_fields in INSERT (CreateCredential)
//   - include vault_fields in SELECT + Scan (Get*, List*)
//   - include vault_fields in UPDATE SET (UpdateCredential)
// =============================================================================

// createTestVaultFieldsCredentialStore creates an in-memory SQLite DataStore
// whose credentials table has the post-migration-013 schema (vault_fields column).
func createTestVaultFieldsCredentialStore(t *testing.T) *SQLDataStore {
	t.Helper()
	// Reuse the createTestVaultCredentialStore since it already includes vault_fields.
	return createTestVaultCredentialStore(t)
}

// -----------------------------------------------------------------------------
// TestSQLDataStore_CreateCredential_VaultFields
//
// Creates a vault credential with vault_fields set (JSON map of env→field).
// After creation, GetCredential must return vault_fields intact.
// -----------------------------------------------------------------------------

// TestSQLDataStore_CreateCredential_VaultFields verifies that CreateCredential
// correctly persists the vault_fields JSON blob.
//
// WILL FAIL TO COMPILE — models.CredentialDB.VaultFields does not exist yet.
func TestSQLDataStore_CreateCredential_VaultFields(t *testing.T) {
	ds := createTestVaultFieldsCredentialStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "vf-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))

	vaultFieldsJSON := `{"GITHUB_TOKEN":"token","GITHUB_USER":"username"}`

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cred := &models.CredentialDB{
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     int64(eco.ID),
		Name:        "github-creds",
		Source:      "vault",
		VaultSecret: strPtr("github/creds"),
		VaultEnv:    strPtr("production"),
		VaultFields: strPtr(vaultFieldsJSON),
	}
	// ─────────────────────────────────────────────────────────────────────────

	err := ds.CreateCredential(cred)
	require.NoError(t, err, "CreateCredential with VaultFields must succeed")
	assert.NotZero(t, cred.ID, "CreateCredential must populate ID")

	// Read back.
	got, err := ds.GetCredential(models.CredentialScopeEcosystem, int64(eco.ID), "github-creds")
	require.NoError(t, err)

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	require.NotNil(t, got.VaultFields, "VaultFields must not be nil after round-trip")
	assert.Equal(t, vaultFieldsJSON, *got.VaultFields,
		"VaultFields JSON must be returned unchanged")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestSQLDataStore_UpdateCredential_VaultFieldsColumn verifies that UpdateCredential
// correctly updates the vault_fields column.
//
// WILL FAIL TO COMPILE — models.CredentialDB.VaultFields does not exist yet.
func TestSQLDataStore_UpdateCredential_VaultFieldsColumn(t *testing.T) {
	ds := createTestVaultFieldsCredentialStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "vf-update-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))

	initialJSON := `{"OLD_TOKEN":"token"}`
	updatedJSON := `{"NEW_TOKEN":"token","NEW_USER":"username"}`

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cred := &models.CredentialDB{
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     int64(eco.ID),
		Name:        "updatable-creds",
		Source:      "vault",
		VaultSecret: strPtr("my-org/creds"),
		VaultFields: strPtr(initialJSON),
	}
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, ds.CreateCredential(cred))

	// Update vault_fields.
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cred.VaultFields = strPtr(updatedJSON)
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, ds.UpdateCredential(cred), "UpdateCredential must succeed")

	got, err := ds.GetCredential(models.CredentialScopeEcosystem, int64(eco.ID), "updatable-creds")
	require.NoError(t, err)

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	require.NotNil(t, got.VaultFields, "VaultFields must not be nil after update")
	assert.Equal(t, updatedJSON, *got.VaultFields,
		"UpdateCredential must persist the new VaultFields value")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestSQLDataStore_UpdateCredential_ClearVaultFields verifies that setting
// VaultFields to nil in an update clears the column.
//
// WILL FAIL TO COMPILE — models.CredentialDB.VaultFields does not exist yet.
func TestSQLDataStore_UpdateCredential_ClearVaultFields(t *testing.T) {
	ds := createTestVaultFieldsCredentialStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "vf-clear-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cred := &models.CredentialDB{
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     int64(eco.ID),
		Name:        "clearable-creds",
		Source:      "vault",
		VaultSecret: strPtr("my-org/creds"),
		VaultFields: strPtr(`{"SOME_VAR":"field"}`),
	}
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, ds.CreateCredential(cred))

	// Clear vault_fields.
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cred.VaultFields = nil
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, ds.UpdateCredential(cred))

	got, err := ds.GetCredential(models.CredentialScopeEcosystem, int64(eco.ID), "clearable-creds")
	require.NoError(t, err)

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	assert.Nil(t, got.VaultFields,
		"VaultFields must be nil after being cleared via UpdateCredential")
	// ─────────────────────────────────────────────────────────────────────────
}

// TestSQLDataStore_ListCredentialsByScope_VaultFieldsColumn verifies that
// ListCredentialsByScope correctly returns vault_fields for each credential.
//
// WILL FAIL TO COMPILE — models.CredentialDB.VaultFields does not exist yet.
func TestSQLDataStore_ListCredentialsByScope_VaultFieldsColumn(t *testing.T) {
	ds := createTestVaultFieldsCredentialStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "vf-list-eco"}
	require.NoError(t, ds.CreateEcosystem(eco))
	scopeID := int64(eco.ID)

	fieldsJSON := `{"API_KEY":"api_key","API_SECRET":"api_secret"}`

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	withFields := &models.CredentialDB{
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     scopeID,
		Name:        "multi-field-cred",
		Source:      "vault",
		VaultSecret: strPtr("my-org/api"),
		VaultFields: strPtr(fieldsJSON),
	}
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, ds.CreateCredential(withFields))

	withoutFields := &models.CredentialDB{
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     scopeID,
		Name:        "simple-cred",
		Source:      "vault",
		VaultSecret: strPtr("my-org/simple"),
	}
	require.NoError(t, ds.CreateCredential(withoutFields))

	results, err := ds.ListCredentialsByScope(models.CredentialScopeEcosystem, scopeID)
	require.NoError(t, err)
	require.Len(t, results, 2)

	byName := make(map[string]*models.CredentialDB, len(results))
	for _, c := range results {
		byName[c.Name] = c
	}

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	mfc, ok := byName["multi-field-cred"]
	require.True(t, ok)
	require.NotNil(t, mfc.VaultFields, "multi-field-cred VaultFields must not be nil")
	assert.Equal(t, fieldsJSON, *mfc.VaultFields)

	sc, ok := byName["simple-cred"]
	require.True(t, ok)
	assert.Nil(t, sc.VaultFields, "simple-cred VaultFields must be nil")
	// ─────────────────────────────────────────────────────────────────────────
}
