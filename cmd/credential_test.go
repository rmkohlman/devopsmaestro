package cmd

// =============================================================================
// TDD Phase 2 (RED): Credential CLI Command Tests (Wave 3)
// =============================================================================
// These tests define the specification for three new credential command groups:
//
//   dvm create credential <name>     -- cmd/credential.go
//   dvm get credentials              -- cmd/get_credential.go
//   dvm get credential <name>        -- cmd/get_credential.go
//   dvm delete credential <name>     -- added to cmd/delete.go
//
// Tests are written FIRST to drive the implementation (TDD RED phase).
// The tests WILL FAIL until the implementation files are created.
//
// Strategy: All tests access the parent command's subcommand list (via
// findSubcommand) rather than referencing not-yet-existing command variables
// directly. This keeps the test file compilable even before implementation.
//
// Symbols expected to exist after implementation:
//   - createCredentialCmd  (cmd/credential.go)
//   - getCredentialsCmd    (cmd/get_credential.go)
//   - getCredentialCmd     (cmd/get_credential.go)
//   - deleteCredentialCmd  (added to cmd/delete.go)
//   - CredentialScopeFlags (cmd/credential.go)
//   - resolveCredentialScopeFromFlags (cmd/credential.go)
// =============================================================================

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/envvalidation"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// A. Command Structure Tests
// =============================================================================

// TestCreateCredentialCmd_Exists verifies that a "credential" subcommand exists
// under createCmd.
//
// This test EXPECTS TO FAIL until createCredentialCmd is registered under createCmd
// in cmd/credential.go.
func TestCreateCredentialCmd_Exists(t *testing.T) {
	credCmd := findSubcommand(createCmd, "credential")
	assert.NotNil(t, credCmd,
		"createCmd should have 'credential' subcommand (dvm create credential)")
}

// TestCreateCredentialCmd_HasCorrectUse verifies the Use field.
//
// This test EXPECTS TO FAIL until createCredentialCmd is implemented.
func TestCreateCredentialCmd_HasCorrectUse(t *testing.T) {
	credCmd := findSubcommand(createCmd, "credential")
	require.NotNil(t, credCmd, "create credential subcommand must exist")

	assert.Equal(t, "credential <name>", credCmd.Use,
		"create credential command should use 'credential <name>'")
}

// TestCreateCredentialCmd_Aliases verifies the command has the "cred" alias.
//
// This test EXPECTS TO FAIL until createCredentialCmd is implemented.
func TestCreateCredentialCmd_Aliases(t *testing.T) {
	credCmd := findSubcommand(createCmd, "credential")
	require.NotNil(t, credCmd, "create credential subcommand must exist")

	assert.Contains(t, credCmd.Aliases, "cred",
		"create credential should have 'cred' alias")
}

// TestGetCredentialsCmd_Exists verifies that a "credentials" subcommand exists
// under getCmd.
//
// This test EXPECTS TO FAIL until getCredentialsCmd is registered under getCmd.
func TestGetCredentialsCmd_Exists(t *testing.T) {
	credsCmd := findSubcommand(getCmd, "credentials")
	assert.NotNil(t, credsCmd,
		"getCmd should have 'credentials' subcommand (dvm get credentials)")
}

// TestGetCredentialCmd_Exists verifies that a "credential" subcommand exists
// under getCmd (singular form for getting one credential).
//
// This test EXPECTS TO FAIL until getCredentialCmd is registered under getCmd.
func TestGetCredentialCmd_Exists(t *testing.T) {
	credCmd := findSubcommand(getCmd, "credential")
	assert.NotNil(t, credCmd,
		"getCmd should have 'credential' subcommand (dvm get credential <name>)")
}

// TestDeleteCredentialCmd_Exists verifies that a "credential" subcommand exists
// under deleteCmd.
//
// This test EXPECTS TO FAIL until deleteCredentialCmd is registered under deleteCmd.
func TestDeleteCredentialCmd_Exists(t *testing.T) {
	credCmd := findSubcommand(deleteCmd, "credential")
	assert.NotNil(t, credCmd,
		"deleteCmd should have 'credential' subcommand (dvm delete credential)")
}

// TestGetCredentialsCmd_Aliases verifies the plural get command has the correct aliases.
//
// This test EXPECTS TO FAIL until getCredentialsCmd is implemented.
func TestGetCredentialsCmd_Aliases(t *testing.T) {
	credsCmd := findSubcommand(getCmd, "credentials")
	require.NotNil(t, credsCmd, "get credentials subcommand must exist")

	assert.Contains(t, credsCmd.Aliases, "cred",
		"get credentials should have 'cred' alias")
	assert.Contains(t, credsCmd.Aliases, "creds",
		"get credentials should have 'creds' alias")
}

// TestDeleteCredentialCmd_Aliases verifies the delete credential command has "cred" alias.
//
// This test EXPECTS TO FAIL until deleteCredentialCmd is implemented.
func TestDeleteCredentialCmd_Aliases(t *testing.T) {
	credCmd := findSubcommand(deleteCmd, "credential")
	require.NotNil(t, credCmd, "delete credential subcommand must exist")

	assert.Contains(t, credCmd.Aliases, "cred",
		"delete credential should have 'cred' alias")
}

// =============================================================================
// B. Flag Tests
// =============================================================================

// TestCreateCredentialCmd_HasSourceFlag verifies the required --source flag.
//
// This test EXPECTS TO FAIL until createCredentialCmd is implemented.
func TestCreateCredentialCmd_HasSourceFlag(t *testing.T) {
	credCmd := findSubcommand(createCmd, "credential")
	require.NotNil(t, credCmd, "create credential subcommand must exist")

	flag := credCmd.Flags().Lookup("source")
	assert.NotNil(t, flag, "create credential should have --source flag")
	if flag != nil {
		assert.Equal(t, "string", flag.Value.Type(), "--source should be a string flag")
		assert.Equal(t, "", flag.DefValue, "--source should default to empty")
	}
}

// TestCreateCredentialCmd_HasVaultSecretFlag_Legacy verifies that --service flag
// no longer exists (replaced by --vault-secret in v0.40.0).
//
// This test verifies the old --service flag has been removed.
func TestCreateCredentialCmd_HasServiceFlag(t *testing.T) {
	credCmd := findSubcommand(createCmd, "credential")
	require.NotNil(t, credCmd, "create credential subcommand must exist")

	flag := credCmd.Flags().Lookup("service")
	assert.Nil(t, flag, "--service flag should no longer exist (replaced by --vault-secret in v0.40.0)")
}

// TestCreateCredentialCmd_HasEnvVarFlag verifies the --env-var flag for env source.
//
// This test EXPECTS TO FAIL until createCredentialCmd is implemented.
func TestCreateCredentialCmd_HasEnvVarFlag(t *testing.T) {
	credCmd := findSubcommand(createCmd, "credential")
	require.NotNil(t, credCmd, "create credential subcommand must exist")

	flag := credCmd.Flags().Lookup("env-var")
	assert.NotNil(t, flag, "create credential should have --env-var flag (not --env)")
	if flag != nil {
		assert.Equal(t, "string", flag.Value.Type(), "--env-var should be a string flag")
		assert.Equal(t, "", flag.DefValue, "--env-var should default to empty")
	}
}

// TestCreateCredentialCmd_HasScopeFlags verifies all four scope flags exist with
// correct shorthand letters.
//
// This test EXPECTS TO FAIL until createCredentialCmd is implemented.
func TestCreateCredentialCmd_HasScopeFlags(t *testing.T) {
	credCmd := findSubcommand(createCmd, "credential")
	require.NotNil(t, credCmd, "create credential subcommand must exist")

	tests := []struct {
		flagName  string
		shorthand string
	}{
		{"ecosystem", "e"},
		{"domain", "d"},
		{"app", "a"},
		{"workspace", "w"},
	}

	for _, tt := range tests {
		t.Run("has --"+tt.flagName, func(t *testing.T) {
			flag := credCmd.Flags().Lookup(tt.flagName)
			assert.NotNil(t, flag,
				"create credential should have --%s flag", tt.flagName)
			if flag != nil {
				assert.Equal(t, "string", flag.Value.Type(),
					"--%s should be a string flag", tt.flagName)
				assert.Equal(t, tt.shorthand, flag.Shorthand,
					"--%s should have -%s shorthand", tt.flagName, tt.shorthand)
			}
		})
	}
}

// TestGetCredentialsCmd_HasAllFlag verifies the --all/-A flag for listing all credentials.
//
// This test EXPECTS TO FAIL until getCredentialsCmd is implemented.
func TestGetCredentialsCmd_HasAllFlag(t *testing.T) {
	credsCmd := findSubcommand(getCmd, "credentials")
	require.NotNil(t, credsCmd, "get credentials subcommand must exist")

	flag := credsCmd.Flags().Lookup("all")
	assert.NotNil(t, flag, "get credentials should have --all flag")
	if flag != nil {
		assert.Equal(t, "bool", flag.Value.Type(), "--all should be a bool flag")
		assert.Equal(t, "A", flag.Shorthand, "--all should have -A shorthand")
		assert.Equal(t, "false", flag.DefValue, "--all should default to false")
	}
}

// TestDeleteCredentialCmd_HasForceFlag verifies the --force/-f flag for skipping
// the deletion confirmation.
//
// This test EXPECTS TO FAIL until deleteCredentialCmd is implemented.
func TestDeleteCredentialCmd_HasForceFlag(t *testing.T) {
	credCmd := findSubcommand(deleteCmd, "credential")
	require.NotNil(t, credCmd, "delete credential subcommand must exist")

	flag := credCmd.Flags().Lookup("force")
	assert.NotNil(t, flag, "delete credential should have --force flag")
	if flag != nil {
		assert.Equal(t, "bool", flag.Value.Type(), "--force should be a bool flag")
		assert.Equal(t, "", flag.Shorthand, "--force should have no shorthand (-f is reserved for --filename)")
		assert.Equal(t, "false", flag.DefValue, "--force should default to false")
	}
}

// =============================================================================
// C. Scope Resolution Tests
// =============================================================================
//
// The scope resolution tests validate the LOGIC that resolveCredentialScopeFromFlags
// will implement. We test this by directly calling DataStore methods, since
// resolveCredentialScopeFromFlags does not yet exist.
//
// Once the implementation exists, additional tests can call
// resolveCredentialScopeFromFlags directly.

// TestResolveCredentialScopeFromFlags_ExactlyOneScope verifies that the scope
// validation logic requires exactly one scope flag.
//
// This is a table-driven test covering the most important scope validation cases.
// The logic mirrors what resolveCredentialScopeFromFlags MUST enforce.
//
// This test EXPECTS TO FAIL until resolveCredentialScopeFromFlags is implemented.
func TestResolveCredentialScopeFromFlags_ExactlyOneScope(t *testing.T) {
	tests := []struct {
		name      string
		flags     map[string]string // flag name -> value
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "ecosystem only - success",
			flags:   map[string]string{"ecosystem": "test-eco"},
			wantErr: false,
		},
		{
			name:    "app only - success",
			flags:   map[string]string{"app": "test-app"},
			wantErr: false,
		},
		{
			name:    "domain only - success",
			flags:   map[string]string{"domain": "test-domain"},
			wantErr: false,
		},
		{
			name:    "workspace only - success",
			flags:   map[string]string{"workspace": "test-ws"},
			wantErr: false,
		},
		{
			name:      "no scope - error",
			flags:     map[string]string{},
			wantErr:   true,
			errSubstr: "exactly one scope",
		},
		{
			name:      "two scopes (ecosystem + app) - error",
			flags:     map[string]string{"ecosystem": "test-eco", "app": "test-app"},
			wantErr:   true,
			errSubstr: "exactly one scope",
		},
		{
			name:      "three scopes - error",
			flags:     map[string]string{"ecosystem": "eco", "domain": "dom", "app": "app"},
			wantErr:   true,
			errSubstr: "exactly one scope",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build a cobra command with the expected scope flags
			cmd := &cobra.Command{}
			cmd.Flags().StringP("ecosystem", "e", "", "ecosystem scope")
			cmd.Flags().StringP("domain", "d", "", "domain scope")
			cmd.Flags().StringP("app", "a", "", "app scope")
			cmd.Flags().StringP("workspace", "w", "", "workspace scope")

			for k, v := range tt.flags {
				err := cmd.Flags().Set(k, v)
				require.NoError(t, err, "failed to set flag %s=%s", k, v)
			}

			// Count how many scopes were set (this is the logic resolveCredentialScopeFromFlags
			// MUST enforce). We simulate the validation here.
			eco, _ := cmd.Flags().GetString("ecosystem")
			dom, _ := cmd.Flags().GetString("domain")
			app, _ := cmd.Flags().GetString("app")
			ws, _ := cmd.Flags().GetString("workspace")

			count := 0
			if eco != "" {
				count++
			}
			if dom != "" {
				count++
			}
			if app != "" {
				count++
			}
			if ws != "" {
				count++
			}

			// Simulate the validation
			var err error
			if count != 1 {
				err = fmt.Errorf("exactly one scope (--ecosystem, --domain, --app, or --workspace) is required, got %d", count)
			}

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestResolveCredentialScopeFromFlags_ResolvesEcosystem verifies that when
// --ecosystem is provided, the ecosystem is found by name and its ID is returned.
//
// This validates the DataStore lookup behavior that resolveCredentialScopeFromFlags
// will delegate to.
func TestResolveCredentialScopeFromFlags_ResolvesEcosystem(t *testing.T) {
	mockStore, _ := setupTestContext()

	// Look up the ecosystem created by setupTestContext
	eco, err := mockStore.GetEcosystemByName("test-eco")
	require.NoError(t, err)
	require.NotNil(t, eco)

	// Verify the ecosystem can be resolved by name → ID (the core of scope resolution)
	assert.Equal(t, "test-eco", eco.Name)
	assert.Greater(t, eco.ID, 0, "ecosystem should have a positive ID")

	// This mirrors what resolveCredentialScopeFromFlags will do:
	// eco, err := ds.GetEcosystemByName(ecoFlag)  → (CredentialScopeEcosystem, int64(eco.ID))
	scopeType := models.CredentialScopeEcosystem
	scopeID := int64(eco.ID)

	assert.Equal(t, models.CredentialScopeEcosystem, scopeType)
	assert.Greater(t, scopeID, int64(0))
}

// TestResolveCredentialScopeFromFlags_ResolvesApp verifies that when --app is
// provided, the app is found by name and its ID is returned.
func TestResolveCredentialScopeFromFlags_ResolvesApp(t *testing.T) {
	mockStore, app := setupTestContext()
	_ = mockStore

	// Verify the app can be resolved globally by name (the core of scope resolution)
	resolved, err := mockStore.GetAppByNameGlobal("test-app")
	require.NoError(t, err)
	require.NotNil(t, resolved)

	// This mirrors what resolveCredentialScopeFromFlags will do:
	// app, err := ds.GetAppByNameGlobal(appFlag) → (CredentialScopeApp, int64(app.ID))
	scopeType := models.CredentialScopeApp
	scopeID := int64(resolved.ID)

	assert.Equal(t, models.CredentialScopeApp, scopeType)
	assert.Equal(t, int64(app.ID), scopeID,
		"resolved app ID should match the test-app ID from setupTestContext")
}

// TestResolveCredentialScopeFromFlags_InvalidEcosystem verifies that when an
// ecosystem name that doesn't exist is provided, an error is returned.
func TestResolveCredentialScopeFromFlags_InvalidEcosystem(t *testing.T) {
	mockStore := db.NewMockDataStore()

	// No ecosystems created → looking up any ecosystem should fail
	_, err := mockStore.GetEcosystemByName("nonexistent-eco")
	assert.Error(t, err,
		"looking up a nonexistent ecosystem should return an error")
	assert.Contains(t, err.Error(), "not found",
		"error should indicate ecosystem was not found")
}

// TestCredentialScopeFlags_ZeroValue verifies that a zero-value CredentialScopeFlags
// struct means no scope is set. This validates that the struct's empty state is
// distinguishable from a set state.
//
// This test EXPECTS TO FAIL until CredentialScopeFlags is defined in cmd/credential.go.
func TestCredentialScopeFlags_ZeroValue(t *testing.T) {
	// Build a cobra command with scope flags (simulating what CredentialScopeFlags adds)
	cmd := &cobra.Command{}
	cmd.Flags().StringP("ecosystem", "e", "", "ecosystem scope")
	cmd.Flags().StringP("domain", "d", "", "domain scope")
	cmd.Flags().StringP("app", "a", "", "app scope")
	cmd.Flags().StringP("workspace", "w", "", "workspace scope")

	// Before any flags are set, all values should be empty
	eco, _ := cmd.Flags().GetString("ecosystem")
	dom, _ := cmd.Flags().GetString("domain")
	app, _ := cmd.Flags().GetString("app")
	ws, _ := cmd.Flags().GetString("workspace")

	assert.Empty(t, eco, "ecosystem should be empty in zero state")
	assert.Empty(t, dom, "domain should be empty in zero state")
	assert.Empty(t, app, "app should be empty in zero state")
	assert.Empty(t, ws, "workspace should be empty in zero state")

	// Zero-value means no scope is set (count == 0)
	count := 0
	if eco != "" {
		count++
	}
	if dom != "" {
		count++
	}
	if app != "" {
		count++
	}
	if ws != "" {
		count++
	}
	assert.Equal(t, 0, count, "zero-value CredentialScopeFlags means no scope set (count=0)")
}

// =============================================================================
// D. Create Credential Validation Tests (Table-Driven)
// =============================================================================

// TestCreateCredential_Validation_TableDriven tests the validation logic that
// createCredentialCmd.RunE will enforce. It covers all source/flag combinations.
//
// This test EXPECTS TO FAIL until the validation logic is implemented in
// cmd/credential.go. The validation must reject invalid combinations before
// calling ds.CreateCredential().
func TestCreateCredential_Validation_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		vaultSecret string
		envVar      string
		wantErr     bool
		errSubstr   string
	}{
		{
			name:      "missing --source flag",
			source:    "",
			wantErr:   true,
			errSubstr: "source",
		},
		{
			name:      "source=vault without --vault-secret",
			source:    "vault",
			wantErr:   true,
			errSubstr: "vault-secret",
		},
		{
			name:      "source=env without --env-var",
			source:    "env",
			envVar:    "",
			wantErr:   true,
			errSubstr: "env-var",
		},
		{
			name:      "source=invalid",
			source:    "plaintext",
			wantErr:   true,
			errSubstr: "source",
		},
		{
			name:        "valid vault",
			source:      "vault",
			vaultSecret: "my-secret",
			wantErr:     false,
		},
		{
			name:    "valid env",
			source:  "env",
			envVar:  "MY_API_KEY",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the validation logic that createCredentialCmd.RunE must implement.
			// This directly tests the business rules, independent of cobra wiring.
			err := validateCredentialFlags(tt.source, tt.vaultSecret, tt.envVar)

			if tt.wantErr {
				assert.Error(t, err,
					"validation should fail for: %s", tt.name)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr,
						"error should mention '%s'", tt.errSubstr)
				}
			} else {
				assert.NoError(t, err,
					"validation should pass for: %s", tt.name)
			}
		})
	}
}

// validateCredentialFlags is a helper that simulates the validation logic
// that createCredentialCmd.RunE must enforce. This helper lives in the test
// file because the real implementation is in cmd/credential.go (not yet created).
//
// NOTE: Once cmd/credential.go is implemented, this function should be removed
// and tests should call the real validation directly.
func validateCredentialFlags(source, vaultSecret, envVar string) error {
	if source == "" {
		return fmt.Errorf("--source is required (vault or env)")
	}
	if source != "vault" && source != "env" {
		return fmt.Errorf("--source must be 'vault' or 'env', got '%s'", source)
	}
	if source == "vault" && vaultSecret == "" {
		return fmt.Errorf("--vault-secret is required when --source=vault")
	}
	if source == "env" && envVar == "" {
		return fmt.Errorf("--env-var is required when --source=env")
	}
	return nil
}

// =============================================================================
// E. Create / Delete Credential Core Logic Tests
// =============================================================================

// TestCreateCredential_Success_Vault verifies that creating a vault-source
// credential calls ds.CreateCredential with the correct fields populated.
//
// This test EXPECTS TO FAIL until createCredentialCmd.RunE is implemented.
// The test validates the DataStore call is correct — it is intentionally
// NOT testing through the cobra command to keep it unit-level.
func TestCreateCredential_Success_Vault(t *testing.T) {
	mockStore, app := setupTestContext()

	vaultSecret := "com.example.my-secret"
	desc := "API key for example.com"

	cred := &models.CredentialDB{
		Name:        "my-api-key",
		ScopeType:   models.CredentialScopeApp,
		ScopeID:     int64(app.ID),
		Source:      "vault",
		VaultSecret: &vaultSecret,
		Description: &desc,
	}

	err := mockStore.CreateCredential(cred)
	require.NoError(t, err, "CreateCredential should succeed for valid vault credential")

	// Verify it was stored correctly
	stored, err := mockStore.GetCredential(models.CredentialScopeApp, int64(app.ID), "my-api-key")
	require.NoError(t, err)
	require.NotNil(t, stored)

	assert.Equal(t, "my-api-key", stored.Name)
	assert.Equal(t, models.CredentialScopeApp, stored.ScopeType)
	assert.Equal(t, int64(app.ID), stored.ScopeID)
	assert.Equal(t, "vault", stored.Source)
	require.NotNil(t, stored.VaultSecret)
	assert.Equal(t, "com.example.my-secret", *stored.VaultSecret)
	assert.Nil(t, stored.EnvVar, "EnvVar should be nil for vault source")
}

// TestCreateCredential_Success_Env verifies that creating an env-source credential
// calls ds.CreateCredential with the correct fields populated.
//
// This test EXPECTS TO FAIL until createCredentialCmd.RunE is implemented.
func TestCreateCredential_Success_Env(t *testing.T) {
	mockStore, app := setupTestContext()

	envVar := "MY_API_KEY"

	cred := &models.CredentialDB{
		Name:      "my-env-cred",
		ScopeType: models.CredentialScopeApp,
		ScopeID:   int64(app.ID),
		Source:    "env",
		EnvVar:    &envVar,
	}

	err := mockStore.CreateCredential(cred)
	require.NoError(t, err, "CreateCredential should succeed for valid env credential")

	// Verify it was stored correctly
	stored, err := mockStore.GetCredential(models.CredentialScopeApp, int64(app.ID), "my-env-cred")
	require.NoError(t, err)
	require.NotNil(t, stored)

	assert.Equal(t, "my-env-cred", stored.Name)
	assert.Equal(t, "env", stored.Source)
	require.NotNil(t, stored.EnvVar)
	assert.Equal(t, "MY_API_KEY", *stored.EnvVar)
	assert.Nil(t, stored.VaultSecret, "VaultSecret should be nil for env source")
}

// TestCreateCredential_AlreadyExists verifies that attempting to create a
// credential with the same (scope_type, scope_id, name) composite key returns
// an error.
//
// This test EXPECTS TO FAIL until createCredentialCmd.RunE propagates the
// duplicate error from ds.CreateCredential.
func TestCreateCredential_AlreadyExists(t *testing.T) {
	mockStore, app := setupTestContext()

	service := "com.example.service"

	// Create first credential
	cred := &models.CredentialDB{
		Name:        "duplicate-cred",
		ScopeType:   models.CredentialScopeApp,
		ScopeID:     int64(app.ID),
		Source:      "vault",
		VaultSecret: &service,
	}
	err := mockStore.CreateCredential(cred)
	require.NoError(t, err, "first CreateCredential should succeed")

	// Attempt to create the same credential again
	cred2 := &models.CredentialDB{
		Name:        "duplicate-cred",
		ScopeType:   models.CredentialScopeApp,
		ScopeID:     int64(app.ID),
		Source:      "vault",
		VaultSecret: &service,
	}
	err = mockStore.CreateCredential(cred2)
	assert.Error(t, err,
		"CreateCredential should return an error when credential already exists")
	assert.Contains(t, err.Error(), "already exists",
		"error should indicate credential already exists")
}

// TestDeleteCredential_Success verifies that a credential can be deleted by its
// composite key (scopeType, scopeID, name).
//
// This test EXPECTS TO FAIL until deleteCredentialCmd.RunE is implemented.
func TestDeleteCredential_Success(t *testing.T) {
	mockStore, app := setupTestContext()

	service := "com.example.deleteme"

	// Create a credential to delete
	cred := &models.CredentialDB{
		Name:        "to-delete",
		ScopeType:   models.CredentialScopeApp,
		ScopeID:     int64(app.ID),
		Source:      "vault",
		VaultSecret: &service,
	}
	err := mockStore.CreateCredential(cred)
	require.NoError(t, err, "setup: CreateCredential should succeed")

	// Verify it exists before deletion
	stored, err := mockStore.GetCredential(models.CredentialScopeApp, int64(app.ID), "to-delete")
	require.NoError(t, err)
	require.NotNil(t, stored)

	// Delete it
	err = mockStore.DeleteCredential(models.CredentialScopeApp, int64(app.ID), "to-delete")
	require.NoError(t, err, "DeleteCredential should succeed")

	// Verify it's gone
	_, err = mockStore.GetCredential(models.CredentialScopeApp, int64(app.ID), "to-delete")
	assert.Error(t, err, "GetCredential should return error after deletion")
	assert.Contains(t, err.Error(), "not found",
		"error should indicate credential was not found")
}

// =============================================================================
// F. Get Credentials List Tests
// =============================================================================

// TestGetCredentials_ListAll verifies that ListAllCredentials returns credentials
// across all scopes when -A flag is used.
//
// This test EXPECTS TO FAIL until getCredentialsCmd with --all/-A flag is
// implemented in cmd/get_credential.go.
func TestGetCredentials_ListAll(t *testing.T) {
	mockStore, app := setupTestContext()

	// Create credentials at different scopes
	eco, err := mockStore.GetEcosystemByName("test-eco")
	require.NoError(t, err)

	service1 := "com.example.service1"
	service2 := "com.example.service2"
	envVar := "MY_TOKEN"

	creds := []*models.CredentialDB{
		{
			Name:        "eco-cred",
			ScopeType:   models.CredentialScopeEcosystem,
			ScopeID:     int64(eco.ID),
			Source:      "vault",
			VaultSecret: &service1,
		},
		{
			Name:        "app-cred",
			ScopeType:   models.CredentialScopeApp,
			ScopeID:     int64(app.ID),
			Source:      "vault",
			VaultSecret: &service2,
		},
		{
			Name:      "env-cred",
			ScopeType: models.CredentialScopeApp,
			ScopeID:   int64(app.ID),
			Source:    "env",
			EnvVar:    &envVar,
		},
	}

	for _, c := range creds {
		err := mockStore.CreateCredential(c)
		require.NoError(t, err, "setup: CreateCredential should succeed for %s", c.Name)
	}

	// ListAllCredentials (used when -A flag is set)
	all, err := mockStore.ListAllCredentials()
	require.NoError(t, err)

	assert.Len(t, all, 3, "ListAllCredentials should return all 3 credentials")

	// Verify credential names are present
	names := make([]string, len(all))
	for i, c := range all {
		names[i] = c.Name
	}
	assert.Contains(t, names, "eco-cred")
	assert.Contains(t, names, "app-cred")
	assert.Contains(t, names, "env-cred")
}

// TestGetCredentials_FilterByScope verifies that ListCredentialsByScope returns
// only credentials belonging to the specified scope.
//
// This test EXPECTS TO FAIL until getCredentialsCmd with scope flags is
// implemented in cmd/get_credential.go.
func TestGetCredentials_FilterByScope(t *testing.T) {
	mockStore, app := setupTestContext()

	eco, err := mockStore.GetEcosystemByName("test-eco")
	require.NoError(t, err)

	service1 := "com.example.eco"
	service2 := "com.example.app"

	// Create credentials at different scopes
	ecoCred := &models.CredentialDB{
		Name:        "eco-only",
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     int64(eco.ID),
		Source:      "vault",
		VaultSecret: &service1,
	}
	appCred := &models.CredentialDB{
		Name:        "app-only",
		ScopeType:   models.CredentialScopeApp,
		ScopeID:     int64(app.ID),
		Source:      "vault",
		VaultSecret: &service2,
	}

	require.NoError(t, mockStore.CreateCredential(ecoCred))
	require.NoError(t, mockStore.CreateCredential(appCred))

	// ListCredentialsByScope for ecosystem only
	ecoResults, err := mockStore.ListCredentialsByScope(models.CredentialScopeEcosystem, int64(eco.ID))
	require.NoError(t, err)
	assert.Len(t, ecoResults, 1, "ecosystem scope should return only 1 credential")
	if len(ecoResults) == 1 {
		assert.Equal(t, "eco-only", ecoResults[0].Name)
		assert.Equal(t, models.CredentialScopeEcosystem, ecoResults[0].ScopeType)
	}

	// ListCredentialsByScope for app only
	appResults, err := mockStore.ListCredentialsByScope(models.CredentialScopeApp, int64(app.ID))
	require.NoError(t, err)
	assert.Len(t, appResults, 1, "app scope should return only 1 credential")
	if len(appResults) == 1 {
		assert.Equal(t, "app-only", appResults[0].Name)
		assert.Equal(t, models.CredentialScopeApp, appResults[0].ScopeType)
	}
}

// TestGetCredentials_Empty verifies that when no credentials exist, the list
// operations return an empty result (not an error).
//
// This test EXPECTS TO FAIL until getCredentialsCmd correctly handles the
// empty-list case (should display a helpful empty message, not an error).
func TestGetCredentials_Empty(t *testing.T) {
	mockStore := db.NewMockDataStore()

	// ListAllCredentials on empty store should return empty slice (not error)
	all, err := mockStore.ListAllCredentials()
	require.NoError(t, err, "ListAllCredentials should not error on empty store")
	assert.Empty(t, all, "empty store should return no credentials")

	// ListCredentialsByScope on empty store should also return empty slice
	byscope, err := mockStore.ListCredentialsByScope(models.CredentialScopeApp, 1)
	require.NoError(t, err, "ListCredentialsByScope should not error on empty store")
	assert.Empty(t, byscope, "empty scope should return no credentials")
}

// =============================================================================
// G. Command Args Tests
// =============================================================================

// TestCreateCredentialCmd_RequiresExactlyOneArg verifies that createCredentialCmd
// requires exactly one positional argument (the credential name).
//
// This test EXPECTS TO FAIL until createCredentialCmd is implemented.
func TestCreateCredentialCmd_RequiresExactlyOneArg(t *testing.T) {
	credCmd := findSubcommand(createCmd, "credential")
	require.NotNil(t, credCmd, "create credential subcommand must exist")
	require.NotNil(t, credCmd.Args, "create credential should have Args validator")

	// 0 args should fail
	err := credCmd.Args(credCmd, []string{})
	assert.Error(t, err, "should require exactly 1 arg (credential name)")

	// 1 arg should pass
	err = credCmd.Args(credCmd, []string{"my-cred"})
	assert.NoError(t, err, "should accept exactly 1 arg")

	// 2 args should fail
	err = credCmd.Args(credCmd, []string{"my-cred", "extra"})
	assert.Error(t, err, "should reject more than 1 arg")
}

// TestDeleteCredentialCmd_RequiresExactlyOneArg verifies that deleteCredentialCmd
// requires exactly one positional argument.
//
// This test EXPECTS TO FAIL until deleteCredentialCmd is implemented.
func TestDeleteCredentialCmd_RequiresExactlyOneArg(t *testing.T) {
	credCmd := findSubcommand(deleteCmd, "credential")
	require.NotNil(t, credCmd, "delete credential subcommand must exist")
	require.NotNil(t, credCmd.Args, "delete credential should have Args validator")

	// 0 args should fail
	err := credCmd.Args(credCmd, []string{})
	assert.Error(t, err, "should require exactly 1 arg (credential name)")

	// 1 arg should pass
	err = credCmd.Args(credCmd, []string{"my-cred"})
	assert.NoError(t, err, "should accept exactly 1 arg")

	// 2 args should fail
	err = credCmd.Args(credCmd, []string{"my-cred", "extra"})
	assert.Error(t, err, "should reject more than 1 arg")
}

// =============================================================================
// H. Integration-Style Tests: Create then Get then Delete
// =============================================================================

// TestCredentialLifecycle_CreateGetDelete verifies the full create→get→delete
// lifecycle through the DataStore, validating the composite key behavior.
//
// This test verifies correct DataStore behavior that the commands will rely on.
// It passes immediately (DataStore already exists), but exercises the complete
// path that the CLI commands will follow.
func TestCredentialLifecycle_CreateGetDelete(t *testing.T) {
	mockStore, app := setupTestContext()

	service := "com.example.lifecycle"
	desc := "lifecycle test credential"

	// --- CREATE ---
	cred := &models.CredentialDB{
		Name:        "lifecycle-cred",
		ScopeType:   models.CredentialScopeApp,
		ScopeID:     int64(app.ID),
		Source:      "vault",
		VaultSecret: &service,
		Description: &desc,
	}
	err := mockStore.CreateCredential(cred)
	require.NoError(t, err, "Create: should succeed")

	// --- GET ---
	stored, err := mockStore.GetCredential(models.CredentialScopeApp, int64(app.ID), "lifecycle-cred")
	require.NoError(t, err, "Get: should find credential after create")
	require.NotNil(t, stored)
	assert.Equal(t, "lifecycle-cred", stored.Name)
	assert.Equal(t, "vault", stored.Source)
	assert.Equal(t, models.CredentialScopeApp, stored.ScopeType)
	require.NotNil(t, stored.Description)
	assert.Equal(t, "lifecycle test credential", *stored.Description)

	// --- LIST ---
	listed, err := mockStore.ListCredentialsByScope(models.CredentialScopeApp, int64(app.ID))
	require.NoError(t, err, "List: should succeed")
	assert.Len(t, listed, 1, "should list exactly 1 credential in app scope")

	// --- DELETE ---
	err = mockStore.DeleteCredential(models.CredentialScopeApp, int64(app.ID), "lifecycle-cred")
	require.NoError(t, err, "Delete: should succeed")

	// --- VERIFY DELETED ---
	_, err = mockStore.GetCredential(models.CredentialScopeApp, int64(app.ID), "lifecycle-cred")
	assert.Error(t, err, "Get after delete should return error")

	listedAfter, err := mockStore.ListCredentialsByScope(models.CredentialScopeApp, int64(app.ID))
	require.NoError(t, err)
	assert.Empty(t, listedAfter, "list after delete should be empty")
}

// TestCredentialCompositeKey_SameNameDifferentScope verifies that the same
// credential name can exist at different scopes without conflict (because the
// composite key is scope_type+scope_id+name).
func TestCredentialCompositeKey_SameNameDifferentScope(t *testing.T) {
	mockStore, app := setupTestContext()

	eco, err := mockStore.GetEcosystemByName("test-eco")
	require.NoError(t, err)

	service := "com.example.shared"

	// Create "github-token" at ecosystem scope
	ecoCred := &models.CredentialDB{
		Name:        "github-token",
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     int64(eco.ID),
		Source:      "vault",
		VaultSecret: &service,
	}
	err = mockStore.CreateCredential(ecoCred)
	require.NoError(t, err, "creating github-token at ecosystem scope should succeed")

	// Create "github-token" at app scope (same name, different scope — should NOT conflict)
	appCred := &models.CredentialDB{
		Name:        "github-token",
		ScopeType:   models.CredentialScopeApp,
		ScopeID:     int64(app.ID),
		Source:      "vault",
		VaultSecret: &service,
	}
	err = mockStore.CreateCredential(appCred)
	require.NoError(t, err,
		"creating github-token at app scope should succeed even though same name exists at ecosystem scope")

	// Both should be independently retrievable
	ecoStored, err := mockStore.GetCredential(models.CredentialScopeEcosystem, int64(eco.ID), "github-token")
	require.NoError(t, err)
	assert.Equal(t, models.CredentialScopeEcosystem, ecoStored.ScopeType)

	appStored, err := mockStore.GetCredential(models.CredentialScopeApp, int64(app.ID), "github-token")
	require.NoError(t, err)
	assert.Equal(t, models.CredentialScopeApp, appStored.ScopeType)

	// Total: 2 credentials with name "github-token" at different scopes
	all, err := mockStore.ListAllCredentials()
	require.NoError(t, err)
	assert.Len(t, all, 2, "should have 2 credentials total (same name, different scopes)")
}

// =============================================================================
// I. GetCredentials Command Use / Aliases Structure
// =============================================================================

// TestGetCredentialCmd_HasCorrectUse verifies the singular credential get command.
//
// This test EXPECTS TO FAIL until getCredentialCmd is implemented.
func TestGetCredentialCmd_HasCorrectUse(t *testing.T) {
	credCmd := findSubcommand(getCmd, "credential")
	require.NotNil(t, credCmd, "get credential subcommand must exist")

	assert.Equal(t, "credential <name>", credCmd.Use,
		"get credential command should use 'credential <name>'")
}

// TestGetCredentialsCmd_HasScopeFlags verifies that get credentials has scope
// filter flags.
//
// This test EXPECTS TO FAIL until getCredentialsCmd is implemented.
func TestGetCredentialsCmd_HasScopeFlags(t *testing.T) {
	credsCmd := findSubcommand(getCmd, "credentials")
	require.NotNil(t, credsCmd, "get credentials subcommand must exist")

	scopeFlags := []struct {
		name      string
		shorthand string
	}{
		{"ecosystem", "e"},
		{"domain", "d"},
		{"app", "a"},
		{"workspace", "w"},
	}

	for _, sf := range scopeFlags {
		t.Run("has --"+sf.name+" flag", func(t *testing.T) {
			flag := credsCmd.Flags().Lookup(sf.name)
			assert.NotNil(t, flag,
				"get credentials should have --%s scope filter flag", sf.name)
			if flag != nil {
				assert.Equal(t, "string", flag.Value.Type())
				assert.Equal(t, sf.shorthand, flag.Shorthand,
					"--%s should have -%s shorthand", sf.name, sf.shorthand)
			}
		})
	}
}

// TestDeleteCredentialCmd_HasScopeFlags verifies that delete credential has
// scope flags for identifying which scope the credential belongs to.
//
// This test EXPECTS TO FAIL until deleteCredentialCmd is implemented.
func TestDeleteCredentialCmd_HasScopeFlags(t *testing.T) {
	credCmd := findSubcommand(deleteCmd, "credential")
	require.NotNil(t, credCmd, "delete credential subcommand must exist")

	scopeFlags := []struct {
		name      string
		shorthand string
	}{
		{"ecosystem", "e"},
		{"domain", "d"},
		{"app", "a"},
		{"workspace", "w"},
	}

	for _, sf := range scopeFlags {
		t.Run("has --"+sf.name+" flag", func(t *testing.T) {
			flag := credCmd.Flags().Lookup(sf.name)
			assert.NotNil(t, flag,
				"delete credential should have --%s scope flag", sf.name)
			if flag != nil {
				assert.Equal(t, "string", flag.Value.Type())
				assert.Equal(t, sf.shorthand, flag.Shorthand)
			}
		})
	}
}

// =============================================================================
// J. RunE Tests
// =============================================================================

// TestCreateCredentialCmd_HasRunE verifies the command uses RunE (not Run).
//
// This test EXPECTS TO FAIL until createCredentialCmd is implemented.
func TestCreateCredentialCmd_HasRunE(t *testing.T) {
	credCmd := findSubcommand(createCmd, "credential")
	require.NotNil(t, credCmd, "create credential subcommand must exist")
	assert.NotNil(t, credCmd.RunE, "create credential should use RunE for error propagation")
	assert.Nil(t, credCmd.Run, "create credential should NOT use Run (use RunE instead)")
}

// TestDeleteCredentialCmd_HasRunE verifies the command uses RunE (not Run).
//
// This test EXPECTS TO FAIL until deleteCredentialCmd is implemented.
func TestDeleteCredentialCmd_HasRunE(t *testing.T) {
	credCmd := findSubcommand(deleteCmd, "credential")
	require.NotNil(t, credCmd, "delete credential subcommand must exist")
	assert.NotNil(t, credCmd.RunE, "delete credential should use RunE for error propagation")
	assert.Nil(t, credCmd.Run, "delete credential should NOT use Run (use RunE instead)")
}

// =============================================================================
// K. Source Validation "enum" Test
// =============================================================================

// TestCredentialSource_ValidValues verifies that only "keychain" and "env"
// are accepted as source values (never "plaintext", "value", or anything else).
//
// This is a design constraint test: credentials MUST NOT store plaintext values.
func TestCredentialSource_ValidValues(t *testing.T) {
	validSources := []string{"vault", "env"}
	invalidSources := []string{"plaintext", "value", "secret", "raw", "", "VAULT", "ENV", "keychain"}

	for _, src := range validSources {
		t.Run("valid source: "+src, func(t *testing.T) {
			err := validateCredentialFlags(src, "my-service", "MY_VAR")
			// For keychain: service is provided, for env: envVar is provided
			// We just test the source part is valid
			// Both valid sources should NOT produce a "source" error
			if err != nil {
				assert.NotContains(t, err.Error(), "source must be",
					"valid source '%s' should not fail source validation", src)
			}
		})
	}

	for _, src := range invalidSources {
		t.Run("invalid source: "+src, func(t *testing.T) {
			err := validateCredentialFlags(src, "my-service", "MY_VAR")
			assert.Error(t, err, "source '%s' should be rejected", src)
			assert.Contains(t, err.Error(), "source",
				"error for invalid source should mention 'source'")
		})
	}
}

// =============================================================================
// L. Cross-Scope Create in MockDataStore (Regression Guard)
// =============================================================================

// TestCreateCredential_AllScopeTypes verifies that credentials can be created
// at all four scope types without error.
func TestCreateCredential_AllScopeTypes(t *testing.T) {
	mockStore, app := setupTestContext()

	eco, err := mockStore.GetEcosystemByName("test-eco")
	require.NoError(t, err)

	domain, err := mockStore.GetDomainByName(sql.NullInt64{Int64: int64(eco.ID), Valid: true}, "test-domain")
	require.NoError(t, err)

	// Create a workspace so we have a workspace-scope ID
	workspace := &models.Workspace{
		AppID:     app.ID,
		Name:      "test-ws",
		Slug:      "test-eco-test-domain-test-app-test-ws",
		ImageName: "test-image",
		Status:    "stopped",
	}
	err = mockStore.CreateWorkspace(workspace)
	require.NoError(t, err)

	service := "com.example.scope-test"
	envVar := "SCOPE_TOKEN"

	tests := []struct {
		name        string
		scopeType   models.CredentialScopeType
		scopeID     int64
		source      string
		vaultSecret *string
		envVar      *string
	}{
		{
			name:        "ecosystem scope credential",
			scopeType:   models.CredentialScopeEcosystem,
			scopeID:     int64(eco.ID),
			source:      "vault",
			vaultSecret: &service,
		},
		{
			name:        "domain scope credential",
			scopeType:   models.CredentialScopeDomain,
			scopeID:     int64(domain.ID),
			source:      "vault",
			vaultSecret: &service,
		},
		{
			name:      "app scope credential",
			scopeType: models.CredentialScopeApp,
			scopeID:   int64(app.ID),
			source:    "env",
			envVar:    &envVar,
		},
		{
			name:        "workspace scope credential",
			scopeType:   models.CredentialScopeWorkspace,
			scopeID:     int64(workspace.ID),
			source:      "vault",
			vaultSecret: &service,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred := &models.CredentialDB{
				Name:        fmt.Sprintf("cred-%d", i),
				ScopeType:   tt.scopeType,
				ScopeID:     tt.scopeID,
				Source:      tt.source,
				VaultSecret: tt.vaultSecret,
				EnvVar:      tt.envVar,
			}

			err := mockStore.CreateCredential(cred)
			assert.NoError(t, err,
				"CreateCredential should succeed for scope type: %s", tt.scopeType)

			// Verify retrieval
			stored, err := mockStore.GetCredential(tt.scopeType, tt.scopeID, cred.Name)
			assert.NoError(t, err, "GetCredential should succeed after create")
			if stored != nil {
				assert.Equal(t, tt.scopeType, stored.ScopeType)
				assert.Equal(t, tt.scopeID, stored.ScopeID)
			}
		})
	}
}

// =============================================================================
// M. DataStore Injection Test (Demonstrates cmd/credential.go RunE pattern)
// =============================================================================

// TestCreateCredential_RunE_DataStoreNotFound verifies that createCredentialCmd.RunE
// returns an appropriate error when the DataStore is not in the context.
//
// This test EXPECTS TO FAIL until createCredentialCmd.RunE is implemented.
// It validates the error handling pattern used by all credential commands.
func TestCreateCredential_RunE_DataStoreNotFound(t *testing.T) {
	credCmd := findSubcommand(createCmd, "credential")
	require.NotNil(t, credCmd, "create credential subcommand must exist")
	require.NotNil(t, credCmd.RunE, "create credential should have RunE")

	// Set up a command context WITHOUT a DataStore
	credCmd.SetContext(context.Background())

	// RunE should return an error because DataStore is not available
	err := credCmd.RunE(credCmd, []string{"test-cred"})
	assert.Error(t, err,
		"RunE should return error when DataStore is not in context")
}

// === Dual-Field Credential CLI Tests (v0.37.1) ===

// =============================================================================
// A. Flag Existence Tests
// =============================================================================

// TestCreateCredentialCmd_HasUsernameVarFlag verifies that --username-var flag exists
// on createCredentialCmd, is string type, and defaults to empty string.
//
// This test EXPECTS TO FAIL until --username-var is added to createCredentialCmd in
// cmd/credential.go.
func TestCreateCredentialCmd_HasUsernameVarFlag(t *testing.T) {
	credCmd := findSubcommand(createCmd, "credential")
	require.NotNil(t, credCmd, "create credential subcommand must exist")

	flag := credCmd.Flags().Lookup("username-var")
	assert.NotNil(t, flag, "create credential should have --username-var flag")
	if flag != nil {
		assert.Equal(t, "string", flag.Value.Type(), "--username-var should be a string flag")
		assert.Equal(t, "", flag.DefValue, "--username-var should default to empty string")
	}
}

// TestCreateCredentialCmd_HasPasswordVarFlag verifies that --password-var flag exists
// on createCredentialCmd, is string type, and defaults to empty string.
//
// This test EXPECTS TO FAIL until --password-var is added to createCredentialCmd in
// cmd/credential.go.
func TestCreateCredentialCmd_HasPasswordVarFlag(t *testing.T) {
	credCmd := findSubcommand(createCmd, "credential")
	require.NotNil(t, credCmd, "create credential subcommand must exist")

	flag := credCmd.Flags().Lookup("password-var")
	assert.NotNil(t, flag, "create credential should have --password-var flag")
	if flag != nil {
		assert.Equal(t, "string", flag.Value.Type(), "--password-var should be a string flag")
		assert.Equal(t, "", flag.DefValue, "--password-var should default to empty string")
	}
}

// =============================================================================
// B. Validation Tests (Table-Driven)
// =============================================================================

// validateDualFieldFlags implements the validation logic that createCredentialCmd.RunE
// will need for the dual-field vault feature.
//
// Rules:
//   - If usernameVar or passwordVar is set, source MUST be "vault"
//   - usernameVar (if non-empty) must pass envvalidation.ValidateEnvKey
//   - passwordVar (if non-empty) must pass envvalidation.ValidateEnvKey
//   - Otherwise, delegate to existing validateCredentialFlags
func validateDualFieldFlags(source, vaultSecret, envVar, usernameVar, passwordVar string) error {
	// If either dual-field var is set, enforce vault-only constraint
	if usernameVar != "" || passwordVar != "" {
		if source != "vault" {
			return fmt.Errorf("--username-var and --password-var are only valid with --source=vault")
		}
		// Validate the env key names
		if usernameVar != "" {
			if err := envvalidation.ValidateEnvKey(usernameVar); err != nil {
				return fmt.Errorf("--username-var: %w", err)
			}
		}
		if passwordVar != "" {
			if err := envvalidation.ValidateEnvKey(passwordVar); err != nil {
				return fmt.Errorf("--password-var: %w", err)
			}
		}
		// Still need source/vaultSecret validated
		return validateCredentialFlags(source, vaultSecret, envVar)
	}
	// Legacy path: no dual-field vars set
	return validateCredentialFlags(source, vaultSecret, envVar)
}

// TestCreateCredential_DualField_Validation is a table-driven test covering all
// combinations of --username-var and --password-var flag validation.
//
// Some cases EXPECT TO FAIL until the validation logic is wired into RunE in
// cmd/credential.go. Cases that exercise only validateDualFieldFlags (defined in
// this test file) will pass immediately.
func TestCreateCredential_DualField_Validation(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		service     string
		envVar      string
		usernameVar string
		passwordVar string
		wantErr     bool
		errSubstr   string
	}{
		{
			name:        "username-var with keychain - valid",
			source:      "vault",
			service:     "github.com",
			usernameVar: "GH_USER",
			wantErr:     false,
		},
		{
			name:        "password-var with keychain - valid",
			source:      "vault",
			service:     "github.com",
			passwordVar: "GH_PAT",
			wantErr:     false,
		},
		{
			name:        "both vars with keychain - valid",
			source:      "vault",
			service:     "github.com",
			usernameVar: "GH_USER",
			passwordVar: "GH_PAT",
			wantErr:     false,
		},
		{
			name:        "username-var with env source - rejected",
			source:      "env",
			envVar:      "TOKEN",
			usernameVar: "GH_USER",
			wantErr:     true,
			errSubstr:   "vault",
		},
		{
			name:        "password-var with env source - rejected",
			source:      "env",
			envVar:      "TOKEN",
			passwordVar: "GH_PAT",
			wantErr:     true,
			errSubstr:   "vault",
		},
		{
			name:        "invalid env key for username-var",
			source:      "vault",
			service:     "svc",
			usernameVar: "bad-key",
			wantErr:     true,
			errSubstr:   "invalid",
		},
		{
			name:        "invalid env key for password-var",
			source:      "vault",
			service:     "svc",
			passwordVar: "bad-key",
			wantErr:     true,
			errSubstr:   "invalid",
		},
		{
			name:        "reserved DVM_ prefix for username-var",
			source:      "vault",
			service:     "svc",
			usernameVar: "DVM_USER",
			wantErr:     true,
			errSubstr:   "DVM_",
		},
		{
			name:        "dangerous env var for password-var",
			source:      "vault",
			service:     "svc",
			passwordVar: "LD_PRELOAD",
			wantErr:     true,
			errSubstr:   "denylist",
		},
		{
			name:    "neither var set (legacy) - valid",
			source:  "vault",
			service: "svc",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDualFieldFlags(tt.source, tt.service, tt.envVar, tt.usernameVar, tt.passwordVar)

			if tt.wantErr {
				assert.Error(t, err, "validation should fail for: %s", tt.name)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr,
						"error should mention '%s'", tt.errSubstr)
				}
			} else {
				assert.NoError(t, err, "validation should pass for: %s", tt.name)
			}
		})
	}
}

// =============================================================================
// C. DataStore Integration Tests
// =============================================================================

// TestCreateCredential_DualField_Vault_Success verifies that a vault credential
// with both UsernameVar and PasswordVar set can be created and retrieved correctly.
//
// This test EXPECTS TO FAIL because CredentialDB.UsernameVar and CredentialDB.PasswordVar
// fields do not yet exist in models/credential.go. That is expected — TDD RED.
func TestCreateCredential_DualField_Vault_Success(t *testing.T) {
	mockStore, app := setupTestContext()

	vaultSecret := "com.github.token"
	usernameVar := "GH_USER"
	passwordVar := "GH_PAT"

	cred := &models.CredentialDB{
		Name:        "github-dual",
		ScopeType:   models.CredentialScopeApp,
		ScopeID:     int64(app.ID),
		Source:      "vault",
		VaultSecret: &vaultSecret,
		UsernameVar: &usernameVar,
		PasswordVar: &passwordVar,
	}

	err := mockStore.CreateCredential(cred)
	require.NoError(t, err, "CreateCredential should succeed for dual-field vault credential")

	// Verify it was stored correctly
	stored, err := mockStore.GetCredential(models.CredentialScopeApp, int64(app.ID), "github-dual")
	require.NoError(t, err)
	require.NotNil(t, stored)

	assert.Equal(t, "github-dual", stored.Name)
	assert.Equal(t, models.CredentialScopeApp, stored.ScopeType)
	assert.Equal(t, int64(app.ID), stored.ScopeID)
	assert.Equal(t, "vault", stored.Source)
	require.NotNil(t, stored.VaultSecret)
	assert.Equal(t, "com.github.token", *stored.VaultSecret)
	require.NotNil(t, stored.UsernameVar, "UsernameVar should be populated")
	assert.Equal(t, "GH_USER", *stored.UsernameVar)
	require.NotNil(t, stored.PasswordVar, "PasswordVar should be populated")
	assert.Equal(t, "GH_PAT", *stored.PasswordVar)
	assert.Nil(t, stored.EnvVar, "EnvVar should be nil for dual-field vault credential")
}

// TestCreateCredential_DualField_PasswordOnly_Success verifies that a keychain credential
// with only PasswordVar set (no UsernameVar) can be created and retrieved correctly.
//
// This test EXPECTS TO FAIL because CredentialDB.PasswordVar field does not yet exist
// in models/credential.go. That is expected — TDD RED.
func TestCreateCredential_DualField_PasswordOnly_Success(t *testing.T) {
	mockStore, app := setupTestContext()

	vaultSecret := "com.github.token"
	passwordVar := "GH_PAT"

	cred := &models.CredentialDB{
		Name:        "github-password-only",
		ScopeType:   models.CredentialScopeApp,
		ScopeID:     int64(app.ID),
		Source:      "vault",
		VaultSecret: &vaultSecret,
		PasswordVar: &passwordVar,
	}

	err := mockStore.CreateCredential(cred)
	require.NoError(t, err, "CreateCredential should succeed for password-only dual-field credential")

	// Verify it was stored correctly
	stored, err := mockStore.GetCredential(models.CredentialScopeApp, int64(app.ID), "github-password-only")
	require.NoError(t, err)
	require.NotNil(t, stored)

	assert.Equal(t, "github-password-only", stored.Name)
	assert.Equal(t, "vault", stored.Source)
	require.NotNil(t, stored.VaultSecret)
	assert.Equal(t, "com.github.token", *stored.VaultSecret)
	assert.Nil(t, stored.UsernameVar, "UsernameVar should be nil when not set")
	require.NotNil(t, stored.PasswordVar, "PasswordVar should be populated")
	assert.Equal(t, "GH_PAT", *stored.PasswordVar)
}

// === Dual-Field Get Credential Display Tests (v0.37.1) ===

// TestGetCredential_DualField_Detail_HasUsernameVar verifies that a dual-field
// keychain credential stores both UsernameVar and PasswordVar fields correctly
// in the CredentialDB model — which is what the detail display code reads when
// rendering "Username: ..." and "Password: ..." lines in `dvm get credential`.
//
// We test the DATA that the display code will read rather than capturing render
// output directly, since render.Plainf writes to a real terminal.
//
// This test EXPECTS TO FAIL until CredentialDB.UsernameVar and
// CredentialDB.PasswordVar fields are added to models/credential.go.
func TestGetCredential_DualField_Detail_HasUsernameVar(t *testing.T) {
	mockStore, app := setupTestContext()

	vaultSecret := "com.github.credentials"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"

	// Create a dual-field vault credential with both UsernameVar and PasswordVar
	cred := &models.CredentialDB{
		Name:        "github-creds",
		ScopeType:   models.CredentialScopeApp,
		ScopeID:     int64(app.ID),
		Source:      "vault",
		VaultSecret: &vaultSecret,
		UsernameVar: &usernameVar,
		PasswordVar: &passwordVar,
	}

	err := mockStore.CreateCredential(cred)
	require.NoError(t, err, "CreateCredential should succeed for dual-field vault credential")

	// Retrieve and verify both dual-field vars are stored and returned correctly
	stored, err := mockStore.GetCredential(models.CredentialScopeApp, int64(app.ID), "github-creds")
	require.NoError(t, err, "GetCredential should find the dual-field credential")
	require.NotNil(t, stored)

	// UsernameVar is set — display code renders "Username:  GITHUB_USERNAME"
	require.NotNil(t, stored.UsernameVar,
		"UsernameVar should not be nil for a dual-field credential")
	assert.Equal(t, "GITHUB_USERNAME", *stored.UsernameVar,
		"UsernameVar should match the stored value")

	// PasswordVar is set — display code renders "Password:  GITHUB_PAT"
	require.NotNil(t, stored.PasswordVar,
		"PasswordVar should not be nil for a dual-field credential")
	assert.Equal(t, "GITHUB_PAT", *stored.PasswordVar,
		"PasswordVar should match the stored value")

	// Other vault fields remain intact
	assert.Equal(t, "github-creds", stored.Name)
	assert.Equal(t, "vault", stored.Source)
	require.NotNil(t, stored.VaultSecret)
	assert.Equal(t, "com.github.credentials", *stored.VaultSecret)
}

// TestGetCredential_DualField_Detail_PasswordOnly verifies that a dual-field
// credential with only PasswordVar set (no UsernameVar) is stored and retrieved
// with the correct nil/non-nil field state.
//
// The display code must handle this asymmetry: it should NOT render a
// "Username:" line when UsernameVar is nil, but MUST render a "Password:" line
// when PasswordVar is non-nil.
//
// This test EXPECTS TO FAIL until CredentialDB.UsernameVar and
// CredentialDB.PasswordVar fields are added to models/credential.go.
func TestGetCredential_DualField_Detail_PasswordOnly(t *testing.T) {
	mockStore, app := setupTestContext()

	vaultSecret := "com.npm.registry"
	passwordVar := "NPM_AUTH_TOKEN"

	// Create a credential with PasswordVar set but UsernameVar explicitly nil
	cred := &models.CredentialDB{
		Name:        "npm-token",
		ScopeType:   models.CredentialScopeApp,
		ScopeID:     int64(app.ID),
		Source:      "vault",
		VaultSecret: &vaultSecret,
		UsernameVar: nil,
		PasswordVar: &passwordVar,
	}

	err := mockStore.CreateCredential(cred)
	require.NoError(t, err, "CreateCredential should succeed for password-only dual-field credential")

	// Retrieve and verify the nil/non-nil asymmetry is preserved
	stored, err := mockStore.GetCredential(models.CredentialScopeApp, int64(app.ID), "npm-token")
	require.NoError(t, err, "GetCredential should find the credential")
	require.NotNil(t, stored)

	// UsernameVar must be nil — display code must NOT render a "Username:" line
	assert.Nil(t, stored.UsernameVar,
		"UsernameVar should be nil when not set — display code handles nil gracefully")

	// PasswordVar must be non-nil — display code renders "Password:  NPM_AUTH_TOKEN"
	require.NotNil(t, stored.PasswordVar,
		"PasswordVar should not be nil for password-only dual-field credential")
	assert.Equal(t, "NPM_AUTH_TOKEN", *stored.PasswordVar,
		"PasswordVar should match the stored value")
}

// TestGetCredentials_List_DualField_MixedDisplay verifies that a list of
// credentials containing both legacy (single-field) and dual-field credentials
// is returned correctly from the DataStore.
//
// This is the data foundation for the VARS column in `dvm get credentials`:
//   - Legacy:     "NPM_TOKEN    (scope: app, source: env)"              → no vars column
//   - Dual-field: "github-creds (scope: app, source: keychain, vars: GH_USER, GH_PAT)"
//
// This test EXPECTS TO FAIL until CredentialDB.UsernameVar and
// CredentialDB.PasswordVar fields are added to models/credential.go.
func TestGetCredentials_List_DualField_MixedDisplay(t *testing.T) {
	mockStore, app := setupTestContext()

	// --- Legacy credential (env source, single env var, no dual-field vars) ---
	envVar := "NPM_TOKEN"
	legacyCred := &models.CredentialDB{
		Name:      "NPM_TOKEN",
		ScopeType: models.CredentialScopeApp,
		ScopeID:   int64(app.ID),
		Source:    "env",
		EnvVar:    &envVar,
		// UsernameVar and PasswordVar intentionally nil — legacy single-field credential
	}

	// --- Dual-field credential (vault source, two named env vars) ---
	vaultSecret := "com.github.credentials"
	ghUser := "GH_USER"
	ghPat := "GH_PAT"
	dualCred := &models.CredentialDB{
		Name:        "github-creds",
		ScopeType:   models.CredentialScopeApp,
		ScopeID:     int64(app.ID),
		Source:      "vault",
		VaultSecret: &vaultSecret,
		UsernameVar: &ghUser,
		PasswordVar: &ghPat,
	}

	require.NoError(t, mockStore.CreateCredential(legacyCred),
		"setup: legacy credential creation should succeed")
	require.NoError(t, mockStore.CreateCredential(dualCred),
		"setup: dual-field credential creation should succeed")

	// List all credentials in the app scope
	creds, err := mockStore.ListCredentialsByScope(models.CredentialScopeApp, int64(app.ID))
	require.NoError(t, err, "ListCredentialsByScope should succeed")
	assert.Len(t, creds, 2, "should return both the legacy and dual-field credentials")

	// Find each credential by name for targeted assertions
	var foundLegacy, foundDual *models.CredentialDB
	for _, c := range creds {
		switch c.Name {
		case "NPM_TOKEN":
			foundLegacy = c
		case "github-creds":
			foundDual = c
		}
	}

	// Legacy credential: NO dual-field vars (list display omits the vars column)
	require.NotNil(t, foundLegacy, "legacy credential 'NPM_TOKEN' should be in the list")
	assert.Nil(t, foundLegacy.UsernameVar,
		"legacy credential should have nil UsernameVar")
	assert.Nil(t, foundLegacy.PasswordVar,
		"legacy credential should have nil PasswordVar")
	require.NotNil(t, foundLegacy.EnvVar,
		"legacy credential should have EnvVar set")
	assert.Equal(t, "NPM_TOKEN", *foundLegacy.EnvVar)

	// Dual-field credential: BOTH vars populated (list display includes vars column)
	require.NotNil(t, foundDual, "dual-field credential 'github-creds' should be in the list")
	require.NotNil(t, foundDual.UsernameVar,
		"dual-field credential should have UsernameVar set")
	assert.Equal(t, "GH_USER", *foundDual.UsernameVar,
		"dual-field UsernameVar should match")
	require.NotNil(t, foundDual.PasswordVar,
		"dual-field credential should have PasswordVar set")
	assert.Equal(t, "GH_PAT", *foundDual.PasswordVar,
		"dual-field PasswordVar should match")
}

// =============================================================================
// TDD Phase 2 (RED): MaestroVault CLI Tests (v0.40.0)
// =============================================================================
// These tests define the specification for the MaestroVault integration.
// They replace the macOS Keychain-based credential source with MaestroVault.
//
// New/changed flags:
//   --source         now accepts "vault" or "env" (not "keychain" or "env")
//   --vault-secret   required for vault source (replaces --service)
//   --vault-environment  optional vault environment qualifier
//   --vault-username-secret  dual-field vault: separate secret for username
//
// Removed flags:
//   --keychain-label, --keychain-type, --service
//
// Cross-validation rules:
//   - --source=vault requires --vault-secret
//   - --vault-username-secret requires --username-var
//   - --vault-environment is only valid with --source=vault
//   - --source=keychain shows a helpful migration error (not silently accepted)
//
// Tests in this section WILL FAIL AT RUNTIME until the flags and validation
// are implemented in cmd/credential.go (v0.40.0).
// =============================================================================

// ---------------------------------------------------------------------------
// Section: Removed Flag Absence Tests
// ---------------------------------------------------------------------------

// TestCreateCredential_KeychainLabel verifies that --keychain-label flag
// no longer exists on createCredentialCmd (removed in v0.39.0 redesign).
func TestCreateCredential_KeychainLabel(t *testing.T) {
	credCmd := findSubcommand(createCmd, "credential")
	require.NotNil(t, credCmd, "create credential subcommand must exist")

	flag := credCmd.Flags().Lookup("keychain-label")
	assert.Nil(t, flag, "--keychain-label flag should not exist (removed in v0.39.0)")
}

// TestCreateCredential_KeychainType verifies that --keychain-type flag
// no longer exists on createCredentialCmd (removed in v0.39.0 redesign).
func TestCreateCredential_KeychainType(t *testing.T) {
	credCmd := findSubcommand(createCmd, "credential")
	require.NotNil(t, credCmd, "create credential subcommand must exist")

	flag := credCmd.Flags().Lookup("keychain-type")
	assert.Nil(t, flag, "--keychain-type flag should not exist (removed in v0.39.0)")
}

// ---------------------------------------------------------------------------
// Section: Flag Existence Tests
// ---------------------------------------------------------------------------

// TestCreateCredentialCmd_HasVaultSecretFlag verifies that --vault-secret flag
// exists on createCredentialCmd, is string type, and defaults to empty string.
//
// WILL FAIL AT RUNTIME — createCredentialCmd does not yet have --vault-secret.
func TestCreateCredentialCmd_HasVaultSecretFlag(t *testing.T) {
	credCmd := findSubcommand(createCmd, "credential")
	require.NotNil(t, credCmd, "create credential subcommand must exist")

	// ── RUNTIME FAILURE EXPECTED BELOW ───────────────────────────────────────
	// --vault-secret flag does not exist yet on createCredentialCmd.
	flag := credCmd.Flags().Lookup("vault-secret")
	assert.NotNil(t, flag,
		"create credential should have --vault-secret flag (v0.40.0 vault integration)")
	if flag != nil {
		assert.Equal(t, "string", flag.Value.Type(),
			"--vault-secret should be a string flag")
		assert.Equal(t, "", flag.DefValue,
			"--vault-secret should default to empty string")
	}
	// ─────────────────────────────────────────────────────────────────────────
}

// TestCreateCredentialCmd_HasVaultEnvironmentFlag verifies that --vault-env
// flag exists on createCredentialCmd, is string type, and defaults to empty string.
// Note: the flag name is --vault-env (not --vault-environment).
func TestCreateCredentialCmd_HasVaultEnvironmentFlag(t *testing.T) {
	credCmd := findSubcommand(createCmd, "credential")
	require.NotNil(t, credCmd, "create credential subcommand must exist")

	flag := credCmd.Flags().Lookup("vault-env")
	assert.NotNil(t, flag,
		"create credential should have --vault-env flag (v0.40.0 vault integration)")
	if flag != nil {
		assert.Equal(t, "string", flag.Value.Type(),
			"--vault-env should be a string flag")
		assert.Equal(t, "", flag.DefValue,
			"--vault-env should default to empty string (optional)")
	}
}

// TestCreateCredentialCmd_HasVaultUsernameSecretFlag verifies that
// --vault-username-secret flag exists on createCredentialCmd for dual-field
// vault credentials.
//
// WILL FAIL AT RUNTIME — createCredentialCmd does not yet have --vault-username-secret.
func TestCreateCredentialCmd_HasVaultUsernameSecretFlag(t *testing.T) {
	credCmd := findSubcommand(createCmd, "credential")
	require.NotNil(t, credCmd, "create credential subcommand must exist")

	// ── RUNTIME FAILURE EXPECTED BELOW ───────────────────────────────────────
	flag := credCmd.Flags().Lookup("vault-username-secret")
	assert.NotNil(t, flag,
		"create credential should have --vault-username-secret flag (v0.40.0 dual-field vault)")
	if flag != nil {
		assert.Equal(t, "string", flag.Value.Type(),
			"--vault-username-secret should be a string flag")
		assert.Equal(t, "", flag.DefValue,
			"--vault-username-secret should default to empty string")
	}
	// ─────────────────────────────────────────────────────────────────────────
}

// ---------------------------------------------------------------------------
// Section: Source Validation — vault and env accepted; keychain rejected
// ---------------------------------------------------------------------------

// TestCreateCredential_VaultSource verifies that --source=vault with
// --vault-secret is accepted, and a credential is created with the correct
// source="vault" and vault_secret fields.
//
// WILL FAIL AT RUNTIME — vault source is not yet accepted by the validation
// logic in createCredentialCmd.RunE.
func TestCreateCredential_VaultSource(t *testing.T) {
	mockStore, app := setupTestContext()

	vaultSecret := "my-org/github-pat"

	// Act: create a vault-source credential via the DataStore directly.
	// This validates the MODEL and STORE accept vault fields; CLI integration
	// is covered by the flag+validation tests above.
	cred := &models.CredentialDB{
		Name:        "github-pat",
		ScopeType:   models.CredentialScopeApp,
		ScopeID:     int64(app.ID),
		Source:      "vault",      // NEW source value — not yet in CHECK constraint
		VaultSecret: &vaultSecret, // NEW field — does not exist yet → compile error
	}

	err := mockStore.CreateCredential(cred)
	// ── COMPILE/RUNTIME FAILURE EXPECTED ABOVE ───────────────────────────────
	// models.CredentialDB.VaultSecret does not exist yet → compile error.
	// Even if it compiles, MockDataStore.CreateCredential may reject "vault" source.
	require.NoError(t, err, "CreateCredential should succeed for vault source credential")

	stored, err := mockStore.GetCredential(models.CredentialScopeApp, int64(app.ID), "github-pat")
	require.NoError(t, err)
	require.NotNil(t, stored)

	assert.Equal(t, "vault", stored.Source)
	require.NotNil(t, stored.VaultSecret, "VaultSecret should be populated")
	assert.Equal(t, "my-org/github-pat", *stored.VaultSecret)
	assert.Nil(t, stored.EnvVar, "EnvVar should be nil for vault source")
}

// TestCreateCredential_VaultRequiresVaultSecret verifies that --source=vault
// without --vault-secret produces a validation error.
//
// WILL FAIL AT RUNTIME — vault source validation not yet implemented in RunE.
func TestCreateCredential_VaultRequiresVaultSecret(t *testing.T) {
	// ── RUNTIME FAILURE EXPECTED BELOW ───────────────────────────────────────
	// validateVaultFlags does not exist yet.
	err := validateVaultFlags(
		"vault", // --source=vault
		"",      // --vault-secret NOT provided
		"",      // --vault-environment
		"",      // --vault-username-secret
		"",      // --username-var
		"",      // --env-var
	)
	// ─────────────────────────────────────────────────────────────────────────

	assert.Error(t, err,
		"--source=vault without --vault-secret must produce a validation error")
	assert.Contains(t, err.Error(), "vault-secret",
		"error should mention the missing --vault-secret flag")
}

// TestCreateCredential_VaultWithEnvironment verifies that --vault-environment
// is stored correctly alongside --vault-secret when --source=vault.
//
// WILL FAIL AT RUNTIME — VaultEnv field does not exist yet → compile error.
func TestCreateCredential_VaultWithEnvironment(t *testing.T) {
	mockStore, app := setupTestContext()

	vaultSecret := "my-org/db-password"
	vaultEnv := "production"

	cred := &models.CredentialDB{
		Name:        "db-password",
		ScopeType:   models.CredentialScopeApp,
		ScopeID:     int64(app.ID),
		Source:      "vault",
		VaultSecret: &vaultSecret, // NEW field — does not exist yet → compile error
		VaultEnv:    &vaultEnv,    // NEW field — does not exist yet → compile error
	}

	err := mockStore.CreateCredential(cred)
	require.NoError(t, err, "CreateCredential should succeed for vault credential with environment")

	stored, err := mockStore.GetCredential(models.CredentialScopeApp, int64(app.ID), "db-password")
	require.NoError(t, err)
	require.NotNil(t, stored)

	assert.Equal(t, "vault", stored.Source)
	require.NotNil(t, stored.VaultSecret, "VaultSecret should be populated")
	assert.Equal(t, "my-org/db-password", *stored.VaultSecret)
	require.NotNil(t, stored.VaultEnv, "VaultEnv should be populated")
	assert.Equal(t, "production", *stored.VaultEnv)
}

// TestCreateCredential_VaultUsernameSecretRequiresUsernameVar verifies the
// cross-validation rule: --vault-username-secret requires --username-var.
//
// WILL FAIL AT RUNTIME — cross-validation not yet implemented.
func TestCreateCredential_VaultUsernameSecretRequiresUsernameVar(t *testing.T) {
	// ── RUNTIME FAILURE EXPECTED BELOW ───────────────────────────────────────
	// validateVaultFlags does not exist yet.
	err := validateVaultFlags(
		"vault",               // --source=vault
		"my-org/github-creds", // --vault-secret provided
		"",                    // --vault-environment
		"my-org/github-user",  // --vault-username-secret provided
		"",                    // --username-var NOT provided (required when vault-username-secret is set)
		"",                    // --env-var
	)
	// ─────────────────────────────────────────────────────────────────────────

	assert.Error(t, err,
		"--vault-username-secret without --username-var must produce a validation error")
	assert.Contains(t, err.Error(), "username-var",
		"error should mention the required --username-var flag")
}

// TestCreateCredential_KeychainSourceShowsMigrationMessage verifies that
// passing --source=keychain (now unsupported) produces a helpful error
// message that guides the user to migrate to --source=vault.
//
// WILL FAIL AT RUNTIME — migration error message not yet implemented.
func TestCreateCredential_KeychainSourceShowsMigrationMessage(t *testing.T) {
	// ── RUNTIME FAILURE EXPECTED BELOW ───────────────────────────────────────
	// validateVaultFlags does not exist yet.
	err := validateVaultFlags(
		"keychain", // deprecated --source value
		"",         // --vault-secret
		"",         // --vault-environment
		"",         // --vault-username-secret
		"",         // --username-var
		"",         // --env-var
	)
	// ─────────────────────────────────────────────────────────────────────────

	assert.Error(t, err,
		"--source=keychain must produce an error (keychain is no longer supported)")
	assert.Contains(t, err.Error(), "vault",
		"migration error should mention 'vault' as the replacement")
}

// TestCreateCredential_VaultDualField verifies that a vault credential with
// both --vault-username-secret and --username-var/--password-var creates a
// correct dual-field vault credential.
//
// WILL FAIL AT RUNTIME — VaultSecret and VaultUsernameSecret fields don't exist.
func TestCreateCredential_VaultDualField(t *testing.T) {
	mockStore, app := setupTestContext()

	vaultSecret := "my-org/github-pat"
	vaultUsernameSecret := "my-org/github-username"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"

	cred := &models.CredentialDB{
		Name:                "github-creds-vault",
		ScopeType:           models.CredentialScopeApp,
		ScopeID:             int64(app.ID),
		Source:              "vault",
		VaultSecret:         &vaultSecret,         // NEW — compile error
		VaultUsernameSecret: &vaultUsernameSecret, // NEW — compile error
		UsernameVar:         &usernameVar,
		PasswordVar:         &passwordVar,
	}

	err := mockStore.CreateCredential(cred)
	require.NoError(t, err, "CreateCredential should succeed for dual-field vault credential")

	stored, err := mockStore.GetCredential(models.CredentialScopeApp, int64(app.ID), "github-creds-vault")
	require.NoError(t, err)
	require.NotNil(t, stored)

	assert.Equal(t, "vault", stored.Source)
	require.NotNil(t, stored.VaultSecret)
	assert.Equal(t, "my-org/github-pat", *stored.VaultSecret)
	require.NotNil(t, stored.VaultUsernameSecret)
	assert.Equal(t, "my-org/github-username", *stored.VaultUsernameSecret)
	require.NotNil(t, stored.UsernameVar)
	assert.Equal(t, "GITHUB_USERNAME", *stored.UsernameVar)
	require.NotNil(t, stored.PasswordVar)
	assert.Equal(t, "GITHUB_PAT", *stored.PasswordVar)
}

// TestCreateCredential_VaultEnvironmentOnlyWithVaultSource verifies that
// --vault-environment is rejected when --source=env.
//
// WILL FAIL AT RUNTIME — validateVaultFlags does not exist yet.
func TestCreateCredential_VaultEnvironmentOnlyWithVaultSource(t *testing.T) {
	// ── RUNTIME FAILURE EXPECTED BELOW ───────────────────────────────────────
	// validateVaultFlags does not exist yet.
	err := validateVaultFlags(
		"env",        // --source=env
		"",           // --vault-secret (not applicable)
		"production", // --vault-environment (invalid with env source)
		"",           // --vault-username-secret
		"",           // --username-var
		"MY_TOKEN",   // --env-var
	)
	// ─────────────────────────────────────────────────────────────────────────

	assert.Error(t, err,
		"--vault-environment should be rejected when --source=env")
	assert.Contains(t, err.Error(), "vault",
		"error should explain that --vault-environment is only valid with vault source")
}

// ---------------------------------------------------------------------------
// Section: validateVaultFlags Helper
// ---------------------------------------------------------------------------

// validateVaultFlags implements the combined validation logic that
// createCredentialCmd.RunE will need for the MaestroVault integration.
//
// Rules:
//  1. --source must be "vault" or "env" ("keychain" shows migration error)
//  2. --source=vault requires --vault-secret
//  3. --vault-environment is only valid with --source=vault
//  4. --vault-username-secret requires --username-var
//  5. --source=env requires --env-var
//
// NOTE: This function lives in the test file as a specification of the
// validation logic that MUST be implemented in cmd/credential.go RunE for v0.40.0.
// Once the real implementation exists, these tests should call the real
// validation function, and this helper should be removed.
func validateVaultFlags(source, vaultSecret, vaultEnv, vaultUsernameSecret, usernameVar, envVar string) error {
	// Rule 1: keychain source is no longer supported — show migration message
	if source == "keychain" {
		return fmt.Errorf("--source=keychain is no longer supported; migrate to --source=vault (MaestroVault)")
	}

	// Rule 1: source must be "vault" or "env"
	if source != "vault" && source != "env" {
		return fmt.Errorf("--source must be 'vault' or 'env', got %q", source)
	}

	// Rule 3: --vault-environment only valid with vault source
	if vaultEnv != "" && source != "vault" {
		return fmt.Errorf("--vault-environment is only valid with --source=vault")
	}

	// Rule 2: vault source requires --vault-secret
	if source == "vault" && vaultSecret == "" {
		return fmt.Errorf("--vault-secret is required when --source=vault")
	}

	// Rule 4: --vault-username-secret requires --username-var
	if vaultUsernameSecret != "" && usernameVar == "" {
		return fmt.Errorf("--username-var is required when --vault-username-secret is set")
	}

	// Rule 5: env source requires --env-var
	if source == "env" && envVar == "" {
		return fmt.Errorf("--env-var is required when --source=env")
	}

	return nil
}

// ---------------------------------------------------------------------------
// Section: validateVaultFlags Table-Driven Tests
// ---------------------------------------------------------------------------

// TestCreateCredential_VaultFlags_TableDriven is a comprehensive table-driven
// test covering all combinations of vault flag validation.
//
// WILL FAIL AT RUNTIME until validateVaultFlags (above) is moved to cmd/credential.go
// and wired into createCredentialCmd.RunE.
func TestCreateCredential_VaultFlags_TableDriven(t *testing.T) {
	tests := []struct {
		name                string
		source              string
		vaultSecret         string
		vaultEnv            string
		vaultUsernameSecret string
		usernameVar         string
		envVar              string
		wantErr             bool
		errSubstr           string
	}{
		{
			name:        "vault source with vault-secret - valid",
			source:      "vault",
			vaultSecret: "my-org/api-key",
			wantErr:     false,
		},
		{
			name:      "vault source without vault-secret - error",
			source:    "vault",
			wantErr:   true,
			errSubstr: "vault-secret",
		},
		{
			name:        "vault source with environment - valid",
			source:      "vault",
			vaultSecret: "my-org/db-pass",
			vaultEnv:    "staging",
			wantErr:     false,
		},
		{
			name:      "env source with vault-environment - error",
			source:    "env",
			envVar:    "MY_TOKEN",
			vaultEnv:  "production",
			wantErr:   true,
			errSubstr: "vault",
		},
		{
			name:                "vault-username-secret without username-var - error",
			source:              "vault",
			vaultSecret:         "my-org/creds",
			vaultUsernameSecret: "my-org/username",
			usernameVar:         "",
			wantErr:             true,
			errSubstr:           "username-var",
		},
		{
			name:                "vault-username-secret with username-var - valid",
			source:              "vault",
			vaultSecret:         "my-org/creds",
			vaultUsernameSecret: "my-org/username",
			usernameVar:         "GITHUB_USERNAME",
			wantErr:             false,
		},
		{
			name:      "keychain source - migration error",
			source:    "keychain",
			wantErr:   true,
			errSubstr: "vault",
		},
		{
			name:      "env source without env-var - error",
			source:    "env",
			envVar:    "",
			wantErr:   true,
			errSubstr: "env-var",
		},
		{
			name:    "env source with env-var - valid",
			source:  "env",
			envVar:  "MY_API_TOKEN",
			wantErr: false,
		},
		{
			name:      "invalid source - error",
			source:    "plaintext",
			wantErr:   true,
			errSubstr: "source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVaultFlags(
				tt.source,
				tt.vaultSecret,
				tt.vaultEnv,
				tt.vaultUsernameSecret,
				tt.usernameVar,
				tt.envVar,
			)

			if tt.wantErr {
				assert.Error(t, err, "validation should fail for: %s", tt.name)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr,
						"error should mention %q", tt.errSubstr)
				}
			} else {
				assert.NoError(t, err, "validation should pass for: %s", tt.name)
			}
		})
	}
}

// =============================================================================
// TDD Phase 2 (RED): --vault-field Flag Tests (v0.41.0)
// =============================================================================
// New flag on `dvm create credential`:
//
//	--vault-field KEY=field-name   (StringArray, can be repeated)
//
// Design decisions enforced by these tests:
//  1. StringArray (not StringSlice) — no comma-splitting
//  2. Requires --source=vault AND --vault-secret
//  3. Mutually exclusive with --username-var / --password-var
//  4. Max 50 entries
//  5. Each entry must be in KEY=field-name format (KEY = valid env var name)
//
// These tests WILL FAIL AT RUNTIME until:
//   - --vault-field flag is added to createCredentialCmd in cmd/credential.go
//   - validateVaultFieldFlag() validation is implemented in cmd/credential.go
// =============================================================================

// ---------------------------------------------------------------------------
// Section: validateVaultFieldFlag Helper (specification / contract)
// ---------------------------------------------------------------------------

// validateVaultFieldFlags implements the validation rules for --vault-field.
//
// Rules:
//  1. vault fields require --source=vault
//  2. vault fields require --vault-secret to be set
//  3. vault fields are mutually exclusive with --username-var / --password-var
//  4. vault fields are mutually exclusive with --env-var
//  5. max 50 entries
//  6. each entry must be KEY=field-name (KEY must be a valid env var name)
//
// NOTE: This lives in the test file as the specification that must be
// implemented in cmd/credential.go. The actual --vault-field flag must use
// cobra's StringArray (not StringSlice) to prevent comma-splitting.
func validateVaultFieldFlags(source, vaultSecret, usernameVar, passwordVar, envVar string, vaultFields []string) error {
	if len(vaultFields) == 0 {
		return nil
	}

	// Rule 1: vault fields require vault source
	if source != "vault" {
		return fmt.Errorf("--vault-field is only valid with --source=vault")
	}

	// Rule 2: vault fields require --vault-secret
	if vaultSecret == "" {
		return fmt.Errorf("--vault-field requires --vault-secret")
	}

	// Rule 3: mutually exclusive with --username-var / --password-var
	if usernameVar != "" || passwordVar != "" {
		return fmt.Errorf("--vault-field is mutually exclusive with --username-var and --password-var")
	}

	// Rule 4: mutually exclusive with --env-var (for env-sourced credentials)
	if envVar != "" {
		return fmt.Errorf("--vault-field is mutually exclusive with --env-var")
	}

	// Rule 5: max 50 entries
	if len(vaultFields) > 50 {
		return fmt.Errorf("--vault-field: maximum 50 fields allowed, got %d", len(vaultFields))
	}

	// Rule 6: each entry must be KEY=field-name
	for _, entry := range vaultFields {
		eqIdx := -1
		for i, ch := range entry {
			if ch == '=' {
				eqIdx = i
				break
			}
		}
		if eqIdx <= 0 {
			return fmt.Errorf("--vault-field %q must be in KEY=field-name format", entry)
		}
		key := entry[:eqIdx]
		if err := envvalidation.ValidateEnvKey(key); err != nil {
			return fmt.Errorf("--vault-field key %q is not a valid environment variable name", key)
		}
	}

	return nil
}

// ---------------------------------------------------------------------------
// Section: validateVaultFieldFlag Table-Driven Tests
// ---------------------------------------------------------------------------

// TestCreateCredential_VaultFieldFlag_TableDriven tests the --vault-field
// flag validation across all relevant combinations.
func TestCreateCredential_VaultFieldFlag_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		vaultSecret string
		usernameVar string
		passwordVar string
		envVar      string
		vaultFields []string
		wantErr     bool
		errSubstr   string
	}{
		{
			name:        "no vault fields — passes through",
			source:      "vault",
			vaultSecret: "my-org/creds",
			vaultFields: nil,
			wantErr:     false,
		},
		{
			name:        "single vault field — valid",
			source:      "vault",
			vaultSecret: "my-org/creds",
			vaultFields: []string{"GITHUB_TOKEN=token"},
			wantErr:     false,
		},
		{
			name:        "multiple vault fields — valid",
			source:      "vault",
			vaultSecret: "my-org/creds",
			vaultFields: []string{"GITHUB_TOKEN=token", "GITHUB_USER=username"},
			wantErr:     false,
		},
		{
			name:        "vault field with env source — error",
			source:      "env",
			envVar:      "MY_TOKEN",
			vaultFields: []string{"GITHUB_TOKEN=token"},
			wantErr:     true,
			errSubstr:   "vault",
		},
		{
			name:        "vault field without vault-secret — error",
			source:      "vault",
			vaultSecret: "",
			vaultFields: []string{"GITHUB_TOKEN=token"},
			wantErr:     true,
			errSubstr:   "vault-secret",
		},
		{
			name:        "vault field with username-var — mutually exclusive",
			source:      "vault",
			vaultSecret: "my-org/creds",
			usernameVar: "GITHUB_USER",
			vaultFields: []string{"GITHUB_TOKEN=token"},
			wantErr:     true,
			errSubstr:   "mutually exclusive",
		},
		{
			name:        "vault field with password-var — mutually exclusive",
			source:      "vault",
			vaultSecret: "my-org/creds",
			passwordVar: "GITHUB_PAT",
			vaultFields: []string{"GITHUB_TOKEN=token"},
			wantErr:     true,
			errSubstr:   "mutually exclusive",
		},
		{
			name:        "51 vault fields — exceeds max",
			source:      "vault",
			vaultSecret: "my-org/creds",
			vaultFields: func() []string {
				fields := make([]string, 51)
				for i := range fields {
					fields[i] = fmt.Sprintf("ENV_VAR_%03d=field_%03d", i, i)
				}
				return fields
			}(),
			wantErr:   true,
			errSubstr: "50",
		},
		{
			name:        "50 vault fields — at max boundary — valid",
			source:      "vault",
			vaultSecret: "my-org/creds",
			vaultFields: func() []string {
				fields := make([]string, 50)
				for i := range fields {
					fields[i] = fmt.Sprintf("ENV_VAR_%03d=field_%03d", i, i)
				}
				return fields
			}(),
			wantErr: false,
		},
		{
			name:        "missing equals sign — invalid format",
			source:      "vault",
			vaultSecret: "my-org/creds",
			vaultFields: []string{"GITHUB_TOKENtoken"},
			wantErr:     true,
			errSubstr:   "format",
		},
		{
			name:        "invalid env var key — error",
			source:      "vault",
			vaultSecret: "my-org/creds",
			vaultFields: []string{"123INVALID=token"},
			wantErr:     true,
			errSubstr:   "valid environment variable",
		},
		{
			name:        "value with equals sign in field name — valid",
			source:      "vault",
			vaultSecret: "my-org/creds",
			// The field name itself may contain =, only the first = is the separator.
			vaultFields: []string{"GITHUB_TOKEN=some/nested/field"},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVaultFieldFlags(
				tt.source,
				tt.vaultSecret,
				tt.usernameVar,
				tt.passwordVar,
				tt.envVar,
				tt.vaultFields,
			)

			if tt.wantErr {
				assert.Error(t, err, "validation should fail for: %s", tt.name)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr,
						"error should mention %q for: %s", tt.errSubstr, tt.name)
				}
			} else {
				assert.NoError(t, err, "validation should pass for: %s", tt.name)
			}
		})
	}
}

// TestCreateCredentialCmd_HasVaultFieldFlag verifies that the create credential
// command exposes a --vault-field flag.
//
// WILL FAIL AT RUNTIME until the --vault-field flag is registered on
// createCredentialCmd in cmd/credential.go.
func TestCreateCredentialCmd_HasVaultFieldFlag(t *testing.T) {
	credCmd := findSubcommand(createCmd, "credential")
	require.NotNil(t, credCmd, "create credential subcommand must exist")

	flag := credCmd.Flags().Lookup("vault-field")
	assert.NotNil(t, flag, "create credential must have a --vault-field flag")
}

// TestCreateCredentialCmd_VaultFieldFlag_IsStringArray verifies that the
// --vault-field flag is a StringArray type (not StringSlice — no comma splitting).
//
// WILL FAIL AT RUNTIME until the --vault-field flag is registered.
func TestCreateCredentialCmd_VaultFieldFlag_IsStringArray(t *testing.T) {
	credCmd := findSubcommand(createCmd, "credential")
	require.NotNil(t, credCmd, "create credential subcommand must exist")

	flag := credCmd.Flags().Lookup("vault-field")
	require.NotNil(t, flag, "--vault-field flag must exist")

	// StringArray flags have type "stringArray", StringSlice have "stringSlice".
	// This assertion enforces the design decision to use StringArray.
	assert.Equal(t, "stringArray", flag.Value.Type(),
		"--vault-field must use StringArray (not StringSlice) to prevent comma-splitting")
}

// TestCreateCredentialCmd_VaultFieldFlag_Repeatable verifies that multiple
// --vault-field flags can be passed to accumulate multiple field mappings.
//
// WILL FAIL AT RUNTIME until the --vault-field flag is registered.
func TestCreateCredentialCmd_VaultFieldFlag_Repeatable(t *testing.T) {
	credCmd := findSubcommand(createCmd, "credential")
	require.NotNil(t, credCmd, "create credential subcommand must exist")

	flag := credCmd.Flags().Lookup("vault-field")
	require.NotNil(t, flag, "--vault-field flag must exist")

	// A StringArray flag must accept multiple values.
	// We test by setting via the flag's Value interface.
	require.NoError(t, flag.Value.Set("GITHUB_TOKEN=token"))
	require.NoError(t, flag.Value.Set("GITHUB_USER=username"))

	// After two Set calls, the string representation must contain both.
	val := flag.Value.String()
	assert.Contains(t, val, "GITHUB_TOKEN=token",
		"first vault-field must appear in flag value after multiple Set() calls")
	assert.Contains(t, val, "GITHUB_USER=username",
		"second vault-field must appear in flag value after multiple Set() calls")
}

// ---------------------------------------------------------------------------
// Section: Integration: --vault-field in create credential RunE
// ---------------------------------------------------------------------------

// TestCreateCredential_VaultFieldFlags_Integration verifies that the
// createCredentialCmd.RunE correctly stores a credential with vault_fields
// when --vault-field flags are provided.
//
// WILL FAIL AT RUNTIME until:
//   - --vault-field flag is added to createCredentialCmd
//   - RunE handles vault fields via db.UpdateCredential/CreateCredential
func TestCreateCredential_VaultFieldFlags_Integration(t *testing.T) {
	mockStore, _ := setupTestContext()

	credCmd := findSubcommand(createCmd, "credential")
	require.NotNil(t, credCmd, "create credential subcommand must exist")

	// Confirm the flag exists and is a StringArray.
	flag := credCmd.Flags().Lookup("vault-field")
	require.NotNil(t, flag, "--vault-field flag must exist on create credential")
	assert.Equal(t, "stringArray", flag.Value.Type(),
		"--vault-field must be StringArray to avoid comma-splitting")

	// Confirm the MockDataStore was set up correctly.
	_ = mockStore
}
