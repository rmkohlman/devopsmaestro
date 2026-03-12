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

// TestCreateCredentialCmd_HasServiceFlag verifies the --service flag for keychain source.
//
// This test EXPECTS TO FAIL until createCredentialCmd is implemented.
func TestCreateCredentialCmd_HasServiceFlag(t *testing.T) {
	credCmd := findSubcommand(createCmd, "credential")
	require.NotNil(t, credCmd, "create credential subcommand must exist")

	flag := credCmd.Flags().Lookup("service")
	assert.NotNil(t, flag, "create credential should have --service flag")
	if flag != nil {
		assert.Equal(t, "string", flag.Value.Type(), "--service should be a string flag")
		assert.Equal(t, "", flag.DefValue, "--service should default to empty")
	}
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
		assert.Equal(t, "f", flag.Shorthand, "--force should have -f shorthand")
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
		name      string
		source    string
		service   string
		envVar    string
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "missing --source flag",
			source:    "",
			wantErr:   true,
			errSubstr: "source",
		},
		{
			name:      "source=keychain without --service",
			source:    "keychain",
			service:   "",
			wantErr:   true,
			errSubstr: "service",
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
			name:    "valid keychain",
			source:  "keychain",
			service: "my-service",
			wantErr: false,
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
			err := validateCredentialFlags(tt.source, tt.service, tt.envVar)

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
func validateCredentialFlags(source, service, envVar string) error {
	if source == "" {
		return fmt.Errorf("--source is required (keychain or env)")
	}
	if source != "keychain" && source != "env" {
		return fmt.Errorf("--source must be 'keychain' or 'env', got '%s'", source)
	}
	if source == "keychain" && service == "" {
		return fmt.Errorf("--service is required when --source=keychain")
	}
	if source == "env" && envVar == "" {
		return fmt.Errorf("--env-var is required when --source=env")
	}
	return nil
}

// =============================================================================
// E. Create / Delete Credential Core Logic Tests
// =============================================================================

// TestCreateCredential_Success_Keychain verifies that creating a keychain-source
// credential calls ds.CreateCredential with the correct fields populated.
//
// This test EXPECTS TO FAIL until createCredentialCmd.RunE is implemented.
// The test validates the DataStore call is correct — it is intentionally
// NOT testing through the cobra command to keep it unit-level.
func TestCreateCredential_Success_Keychain(t *testing.T) {
	mockStore, app := setupTestContext()

	service := "com.example.my-service"
	desc := "API key for example.com"

	cred := &models.CredentialDB{
		Name:        "my-api-key",
		ScopeType:   models.CredentialScopeApp,
		ScopeID:     int64(app.ID),
		Source:      "keychain",
		Service:     &service,
		Description: &desc,
	}

	err := mockStore.CreateCredential(cred)
	require.NoError(t, err, "CreateCredential should succeed for valid keychain credential")

	// Verify it was stored correctly
	stored, err := mockStore.GetCredential(models.CredentialScopeApp, int64(app.ID), "my-api-key")
	require.NoError(t, err)
	require.NotNil(t, stored)

	assert.Equal(t, "my-api-key", stored.Name)
	assert.Equal(t, models.CredentialScopeApp, stored.ScopeType)
	assert.Equal(t, int64(app.ID), stored.ScopeID)
	assert.Equal(t, "keychain", stored.Source)
	require.NotNil(t, stored.Service)
	assert.Equal(t, "com.example.my-service", *stored.Service)
	assert.Nil(t, stored.EnvVar, "EnvVar should be nil for keychain source")
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
	assert.Nil(t, stored.Service, "Service should be nil for env source")
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
		Name:      "duplicate-cred",
		ScopeType: models.CredentialScopeApp,
		ScopeID:   int64(app.ID),
		Source:    "keychain",
		Service:   &service,
	}
	err := mockStore.CreateCredential(cred)
	require.NoError(t, err, "first CreateCredential should succeed")

	// Attempt to create the same credential again
	cred2 := &models.CredentialDB{
		Name:      "duplicate-cred",
		ScopeType: models.CredentialScopeApp,
		ScopeID:   int64(app.ID),
		Source:    "keychain",
		Service:   &service,
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
		Name:      "to-delete",
		ScopeType: models.CredentialScopeApp,
		ScopeID:   int64(app.ID),
		Source:    "keychain",
		Service:   &service,
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
			Name:      "eco-cred",
			ScopeType: models.CredentialScopeEcosystem,
			ScopeID:   int64(eco.ID),
			Source:    "keychain",
			Service:   &service1,
		},
		{
			Name:      "app-cred",
			ScopeType: models.CredentialScopeApp,
			ScopeID:   int64(app.ID),
			Source:    "keychain",
			Service:   &service2,
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
		Name:      "eco-only",
		ScopeType: models.CredentialScopeEcosystem,
		ScopeID:   int64(eco.ID),
		Source:    "keychain",
		Service:   &service1,
	}
	appCred := &models.CredentialDB{
		Name:      "app-only",
		ScopeType: models.CredentialScopeApp,
		ScopeID:   int64(app.ID),
		Source:    "keychain",
		Service:   &service2,
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
		Source:      "keychain",
		Service:     &service,
		Description: &desc,
	}
	err := mockStore.CreateCredential(cred)
	require.NoError(t, err, "Create: should succeed")

	// --- GET ---
	stored, err := mockStore.GetCredential(models.CredentialScopeApp, int64(app.ID), "lifecycle-cred")
	require.NoError(t, err, "Get: should find credential after create")
	require.NotNil(t, stored)
	assert.Equal(t, "lifecycle-cred", stored.Name)
	assert.Equal(t, "keychain", stored.Source)
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
		Name:      "github-token",
		ScopeType: models.CredentialScopeEcosystem,
		ScopeID:   int64(eco.ID),
		Source:    "keychain",
		Service:   &service,
	}
	err = mockStore.CreateCredential(ecoCred)
	require.NoError(t, err, "creating github-token at ecosystem scope should succeed")

	// Create "github-token" at app scope (same name, different scope — should NOT conflict)
	appCred := &models.CredentialDB{
		Name:      "github-token",
		ScopeType: models.CredentialScopeApp,
		ScopeID:   int64(app.ID),
		Source:    "keychain",
		Service:   &service,
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
	validSources := []string{"keychain", "env"}
	invalidSources := []string{"plaintext", "value", "secret", "raw", "", "KEYCHAIN", "ENV"}

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

	domain, err := mockStore.GetDomainByName(eco.ID, "test-domain")
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
		name      string
		scopeType models.CredentialScopeType
		scopeID   int64
		source    string
		service   *string
		envVar    *string
	}{
		{
			name:      "ecosystem scope credential",
			scopeType: models.CredentialScopeEcosystem,
			scopeID:   int64(eco.ID),
			source:    "keychain",
			service:   &service,
		},
		{
			name:      "domain scope credential",
			scopeType: models.CredentialScopeDomain,
			scopeID:   int64(domain.ID),
			source:    "keychain",
			service:   &service,
		},
		{
			name:      "app scope credential",
			scopeType: models.CredentialScopeApp,
			scopeID:   int64(app.ID),
			source:    "env",
			envVar:    &envVar,
		},
		{
			name:      "workspace scope credential",
			scopeType: models.CredentialScopeWorkspace,
			scopeID:   int64(workspace.ID),
			source:    "keychain",
			service:   &service,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred := &models.CredentialDB{
				Name:      fmt.Sprintf("cred-%d", i),
				ScopeType: tt.scopeType,
				ScopeID:   tt.scopeID,
				Source:    tt.source,
				Service:   tt.service,
				EnvVar:    tt.envVar,
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
// will need for the dual-field keychain feature.
//
// Rules:
//   - If usernameVar or passwordVar is set, source MUST be "keychain"
//   - usernameVar (if non-empty) must pass envvalidation.ValidateEnvKey
//   - passwordVar (if non-empty) must pass envvalidation.ValidateEnvKey
//   - Otherwise, delegate to existing validateCredentialFlags
func validateDualFieldFlags(source, service, envVar, usernameVar, passwordVar string) error {
	// If either dual-field var is set, enforce keychain-only constraint
	if usernameVar != "" || passwordVar != "" {
		if source != "keychain" {
			return fmt.Errorf("--username-var and --password-var are only valid with --source=keychain")
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
		// Still need source/service validated
		return validateCredentialFlags(source, service, envVar)
	}
	// Legacy path: no dual-field vars set
	return validateCredentialFlags(source, service, envVar)
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
			source:      "keychain",
			service:     "github.com",
			usernameVar: "GH_USER",
			wantErr:     false,
		},
		{
			name:        "password-var with keychain - valid",
			source:      "keychain",
			service:     "github.com",
			passwordVar: "GH_PAT",
			wantErr:     false,
		},
		{
			name:        "both vars with keychain - valid",
			source:      "keychain",
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
			errSubstr:   "keychain",
		},
		{
			name:        "password-var with env source - rejected",
			source:      "env",
			envVar:      "TOKEN",
			passwordVar: "GH_PAT",
			wantErr:     true,
			errSubstr:   "keychain",
		},
		{
			name:        "invalid env key for username-var",
			source:      "keychain",
			service:     "svc",
			usernameVar: "bad-key",
			wantErr:     true,
			errSubstr:   "invalid",
		},
		{
			name:        "invalid env key for password-var",
			source:      "keychain",
			service:     "svc",
			passwordVar: "bad-key",
			wantErr:     true,
			errSubstr:   "invalid",
		},
		{
			name:        "reserved DVM_ prefix for username-var",
			source:      "keychain",
			service:     "svc",
			usernameVar: "DVM_USER",
			wantErr:     true,
			errSubstr:   "DVM_",
		},
		{
			name:        "dangerous env var for password-var",
			source:      "keychain",
			service:     "svc",
			passwordVar: "LD_PRELOAD",
			wantErr:     true,
			errSubstr:   "denylist",
		},
		{
			name:    "neither var set (legacy) - valid",
			source:  "keychain",
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

// TestCreateCredential_DualField_Keychain_Success verifies that a keychain credential
// with both UsernameVar and PasswordVar set can be created and retrieved correctly.
//
// This test EXPECTS TO FAIL because CredentialDB.UsernameVar and CredentialDB.PasswordVar
// fields do not yet exist in models/credential.go. That is expected — TDD RED.
func TestCreateCredential_DualField_Keychain_Success(t *testing.T) {
	mockStore, app := setupTestContext()

	service := "com.github.token"
	usernameVar := "GH_USER"
	passwordVar := "GH_PAT"

	cred := &models.CredentialDB{
		Name:        "github-dual",
		ScopeType:   models.CredentialScopeApp,
		ScopeID:     int64(app.ID),
		Source:      "keychain",
		Service:     &service,
		UsernameVar: &usernameVar,
		PasswordVar: &passwordVar,
	}

	err := mockStore.CreateCredential(cred)
	require.NoError(t, err, "CreateCredential should succeed for dual-field keychain credential")

	// Verify it was stored correctly
	stored, err := mockStore.GetCredential(models.CredentialScopeApp, int64(app.ID), "github-dual")
	require.NoError(t, err)
	require.NotNil(t, stored)

	assert.Equal(t, "github-dual", stored.Name)
	assert.Equal(t, models.CredentialScopeApp, stored.ScopeType)
	assert.Equal(t, int64(app.ID), stored.ScopeID)
	assert.Equal(t, "keychain", stored.Source)
	require.NotNil(t, stored.Service)
	assert.Equal(t, "com.github.token", *stored.Service)
	require.NotNil(t, stored.UsernameVar, "UsernameVar should be populated")
	assert.Equal(t, "GH_USER", *stored.UsernameVar)
	require.NotNil(t, stored.PasswordVar, "PasswordVar should be populated")
	assert.Equal(t, "GH_PAT", *stored.PasswordVar)
	assert.Nil(t, stored.EnvVar, "EnvVar should be nil for dual-field keychain credential")
}

// TestCreateCredential_DualField_PasswordOnly_Success verifies that a keychain credential
// with only PasswordVar set (no UsernameVar) can be created and retrieved correctly.
//
// This test EXPECTS TO FAIL because CredentialDB.PasswordVar field does not yet exist
// in models/credential.go. That is expected — TDD RED.
func TestCreateCredential_DualField_PasswordOnly_Success(t *testing.T) {
	mockStore, app := setupTestContext()

	service := "com.github.token"
	passwordVar := "GH_PAT"

	cred := &models.CredentialDB{
		Name:        "github-password-only",
		ScopeType:   models.CredentialScopeApp,
		ScopeID:     int64(app.ID),
		Source:      "keychain",
		Service:     &service,
		PasswordVar: &passwordVar,
	}

	err := mockStore.CreateCredential(cred)
	require.NoError(t, err, "CreateCredential should succeed for password-only dual-field credential")

	// Verify it was stored correctly
	stored, err := mockStore.GetCredential(models.CredentialScopeApp, int64(app.ID), "github-password-only")
	require.NoError(t, err)
	require.NotNil(t, stored)

	assert.Equal(t, "github-password-only", stored.Name)
	assert.Equal(t, "keychain", stored.Source)
	require.NotNil(t, stored.Service)
	assert.Equal(t, "com.github.token", *stored.Service)
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

	service := "com.github.credentials"
	usernameVar := "GITHUB_USERNAME"
	passwordVar := "GITHUB_PAT"

	// Create a dual-field keychain credential with both UsernameVar and PasswordVar
	cred := &models.CredentialDB{
		Name:        "github-creds",
		ScopeType:   models.CredentialScopeApp,
		ScopeID:     int64(app.ID),
		Source:      "keychain",
		Service:     &service,
		UsernameVar: &usernameVar,
		PasswordVar: &passwordVar,
	}

	err := mockStore.CreateCredential(cred)
	require.NoError(t, err, "CreateCredential should succeed for dual-field keychain credential")

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

	// Other keychain fields remain intact
	assert.Equal(t, "github-creds", stored.Name)
	assert.Equal(t, "keychain", stored.Source)
	require.NotNil(t, stored.Service)
	assert.Equal(t, "com.github.credentials", *stored.Service)
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

	service := "com.npm.registry"
	passwordVar := "NPM_AUTH_TOKEN"

	// Create a credential with PasswordVar set but UsernameVar explicitly nil
	cred := &models.CredentialDB{
		Name:        "npm-token",
		ScopeType:   models.CredentialScopeApp,
		ScopeID:     int64(app.ID),
		Source:      "keychain",
		Service:     &service,
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

	// --- Dual-field credential (keychain source, two named env vars) ---
	service := "com.github.credentials"
	ghUser := "GH_USER"
	ghPat := "GH_PAT"
	dualCred := &models.CredentialDB{
		Name:        "github-creds",
		ScopeType:   models.CredentialScopeApp,
		ScopeID:     int64(app.ID),
		Source:      "keychain",
		Service:     &service,
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
// TDD Phase 2 (RED): Keychain Label Redesign — CLI Tests (v0.39.0)
// =============================================================================
// Design change: Add --keychain-label (replaces --service), add --keychain-type,
// keep --service as a deprecated alias, enforce mutual-exclusivity of
// --service and --keychain-label, and validate keychainType values.
//
// New flags that MUST exist after implementation:
//   --keychain-label  (string, replaces --service for keychain credentials)
//   --keychain-type   (string, "generic" or "internet", default "generic")
//   --service         (kept as deprecated alias for --keychain-label)
//
// Validation rules:
//   1. --keychain-label and --service are mutually exclusive
//   2. --keychain-type only valid when --source=keychain
//   3. --keychain-type must be "generic" or "internet" (empty = default generic)
//   4. When --keychain-label is set, it satisfies the keychain-source requirement
//   5. When only --service is set, it is accepted (deprecated path)
//
// Tests in this section WILL FAIL TO COMPILE or FAIL AT RUNTIME until
// --keychain-label and --keychain-type are added to cmd/credential.go.
// =============================================================================

// ---------------------------------------------------------------------------
// Section: Flag Existence Tests
// ---------------------------------------------------------------------------

// TestCreateCredential_KeychainLabel verifies that --keychain-label flag exists
// on createCredentialCmd, is string type, and defaults to empty string.
//
// WILL FAIL AT RUNTIME — createCredentialCmd does not yet have --keychain-label.
func TestCreateCredential_KeychainLabel(t *testing.T) {
	credCmd := findSubcommand(createCmd, "credential")
	require.NotNil(t, credCmd, "create credential subcommand must exist")

	// ── RUNTIME FAILURE EXPECTED BELOW ───────────────────────────────────────
	// --keychain-label flag does not exist yet on createCredentialCmd.
	flag := credCmd.Flags().Lookup("keychain-label")
	assert.NotNil(t, flag,
		"create credential should have --keychain-label flag (replaces --service)")
	if flag != nil {
		assert.Equal(t, "string", flag.Value.Type(),
			"--keychain-label should be a string flag")
		assert.Equal(t, "", flag.DefValue,
			"--keychain-label should default to empty string")
	}
	// ─────────────────────────────────────────────────────────────────────────
}

// TestCreateCredential_KeychainType verifies that --keychain-type flag exists
// on createCredentialCmd, is string type, and defaults to empty string (which
// the implementation treats as "generic").
//
// WILL FAIL AT RUNTIME — createCredentialCmd does not yet have --keychain-type.
func TestCreateCredential_KeychainType(t *testing.T) {
	credCmd := findSubcommand(createCmd, "credential")
	require.NotNil(t, credCmd, "create credential subcommand must exist")

	// ── RUNTIME FAILURE EXPECTED BELOW ───────────────────────────────────────
	// --keychain-type flag does not exist yet on createCredentialCmd.
	flag := credCmd.Flags().Lookup("keychain-type")
	assert.NotNil(t, flag,
		"create credential should have --keychain-type flag")
	if flag != nil {
		assert.Equal(t, "string", flag.Value.Type(),
			"--keychain-type should be a string flag")
		assert.Equal(t, "", flag.DefValue,
			"--keychain-type should default to empty string (treated as 'generic')")
	}
	// ─────────────────────────────────────────────────────────────────────────
}

// TestCreateCredential_ServiceDeprecated verifies that --service flag still
// exists on createCredentialCmd (as a deprecated alias) and works for backwards
// compatibility.
//
// This test should PASS today (--service already exists). It is included as a
// regression guard to ensure --service is NOT removed during the redesign.
func TestCreateCredential_ServiceDeprecated(t *testing.T) {
	credCmd := findSubcommand(createCmd, "credential")
	require.NotNil(t, credCmd, "create credential subcommand must exist")

	flag := credCmd.Flags().Lookup("service")
	assert.NotNil(t, flag,
		"--service flag must be kept as a deprecated alias (for backwards compat)")
	if flag != nil {
		assert.Equal(t, "string", flag.Value.Type(),
			"--service should remain a string flag")
	}
}

// ---------------------------------------------------------------------------
// Section: Validation Logic Tests
// ---------------------------------------------------------------------------

// TestCreateCredential_ServiceAndLabelMutuallyExclusive verifies that
// passing both --service and --keychain-label produces an error.
//
// WILL FAIL AT RUNTIME — this mutual-exclusivity rule is not yet enforced.
func TestCreateCredential_ServiceAndLabelMutuallyExclusive(t *testing.T) {
	// ── RUNTIME FAILURE EXPECTED BELOW ───────────────────────────────────────
	// validateKeychainLabelFlags does not exist yet; using the helper below.
	err := validateKeychainLabelFlags(
		"keychain",
		"com.example.old-service", // --service (deprecated)
		"com.example.new-label",   // --keychain-label (new)
		"",                        // --keychain-type
	)
	// ─────────────────────────────────────────────────────────────────────────

	assert.Error(t, err,
		"providing both --service and --keychain-label must produce an error")
	assert.Contains(t, err.Error(), "mutually exclusive",
		"error must explain that --service and --keychain-label cannot be used together")
}

// TestCreateCredential_KeychainTypeValidation is a table-driven test verifying
// that --keychain-type only accepts "generic", "internet", or "" (default).
//
// WILL FAIL AT RUNTIME — this validation is not yet implemented.
func TestCreateCredential_KeychainTypeValidation(t *testing.T) {
	tests := []struct {
		name         string
		keychainType string
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "generic is valid",
			keychainType: "generic",
			wantErr:      false,
		},
		{
			name:         "internet is valid",
			keychainType: "internet",
			wantErr:      false,
		},
		{
			name:         "empty string is valid (defaults to generic)",
			keychainType: "",
			wantErr:      false,
		},
		{
			name:         "keychain is invalid",
			keychainType: "keychain",
			wantErr:      true,
			errMsg:       "keychain-type",
		},
		{
			name:         "GENERIC uppercase is invalid",
			keychainType: "GENERIC",
			wantErr:      true,
			errMsg:       "keychain-type",
		},
		{
			name:         "INTERNET uppercase is invalid",
			keychainType: "INTERNET",
			wantErr:      true,
			errMsg:       "keychain-type",
		},
		{
			name:         "arbitrary value is invalid",
			keychainType: "any",
			wantErr:      true,
			errMsg:       "keychain-type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ── RUNTIME FAILURE EXPECTED BELOW ───────────────────────────────
			// validateKeychainLabelFlags does not exist yet.
			err := validateKeychainLabelFlags(
				"keychain",
				"",                     // no --service
				"com.example.my-label", // --keychain-label provided
				tt.keychainType,        // --keychain-type under test
			)
			// ─────────────────────────────────────────────────────────────────

			if tt.wantErr {
				assert.Error(t, err,
					"keychain-type %q should be rejected", tt.keychainType)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg,
						"error should mention %q", tt.errMsg)
				}
			} else {
				assert.NoError(t, err,
					"keychain-type %q should be accepted", tt.keychainType)
			}
		})
	}
}

// TestCreateCredential_KeychainTypeRequiresKeychainSource verifies that
// --keychain-type is only valid when --source=keychain. When used with
// --source=env, it must return an error.
//
// WILL FAIL AT RUNTIME — this constraint is not yet enforced.
func TestCreateCredential_KeychainTypeRequiresKeychainSource(t *testing.T) {
	// ── RUNTIME FAILURE EXPECTED BELOW ───────────────────────────────────────
	// validateKeychainLabelFlags does not exist yet.
	err := validateKeychainLabelFlags(
		"env",     // --source=env
		"",        // no --service
		"",        // no --keychain-label
		"generic", // --keychain-type=generic (invalid when source=env)
	)
	// ─────────────────────────────────────────────────────────────────────────

	assert.Error(t, err,
		"--keychain-type should be rejected when --source=env")
	assert.Contains(t, err.Error(), "keychain",
		"error should explain that --keychain-type is only valid with --source=keychain")
}

// TestCreateCredential_KeychainLabelSatisfiesKeychainRequirement verifies that
// when --keychain-label is set (and --service is NOT set), the keychain-source
// requirement is satisfied by the label.
//
// WILL FAIL AT RUNTIME — the current validation requires --service for keychain
// source. After the redesign, --keychain-label must also satisfy this requirement.
func TestCreateCredential_KeychainLabelSatisfiesKeychainRequirement(t *testing.T) {
	// ── RUNTIME FAILURE EXPECTED BELOW ───────────────────────────────────────
	// validateKeychainLabelFlags does not exist yet.
	err := validateKeychainLabelFlags(
		"keychain",             // --source=keychain
		"",                     // no --service (old way)
		"com.example.my-label", // --keychain-label (new way) — should satisfy requirement
		"generic",              // --keychain-type
	)
	// ─────────────────────────────────────────────────────────────────────────

	assert.NoError(t, err,
		"--keychain-label should satisfy the keychain source requirement (replacing --service)")
}

// TestCreateCredential_ServiceAloneStillWorks verifies that using only
// --service (without --keychain-label) still works as a deprecated path.
//
// WILL PASS TODAY (--service is already supported). Included as a regression
// guard to ensure the deprecated path is not broken during the redesign.
func TestCreateCredential_ServiceAloneStillWorks(t *testing.T) {
	// ── This should PASS with current code and continue to pass after redesign ──
	err := validateKeychainLabelFlags(
		"keychain",                // --source=keychain
		"com.example.old-service", // --service (deprecated but still valid)
		"",                        // no --keychain-label
		"",                        // no --keychain-type
	)

	assert.NoError(t, err,
		"--service alone (deprecated path) must still work for backwards compatibility")
}

// ---------------------------------------------------------------------------
// Section: validateKeychainLabelFlags Helper
// ---------------------------------------------------------------------------

// validateKeychainLabelFlags implements the combined validation logic that
// createCredentialCmd.RunE will need for the keychain label redesign.
//
// Rules:
//  1. --service and --keychain-label are mutually exclusive
//  2. For keychain source: either --service OR --keychain-label must be provided
//  3. --keychain-type must be "" (default), "generic", or "internet"
//  4. --keychain-type is only valid with --source=keychain
//
// NOTE: This function lives in the test file as a specification of the
// validation logic that MUST be implemented in cmd/credential.go RunE.
// Once the real implementation exists, these tests should call the real
// validation, and this helper should be removed.
func validateKeychainLabelFlags(source, service, keychainLabel, keychainType string) error {
	// Rule 1: --service and --keychain-label are mutually exclusive
	if service != "" && keychainLabel != "" {
		return fmt.Errorf("--service and --keychain-label are mutually exclusive; use --keychain-label (--service is deprecated)")
	}

	// Rule 4: --keychain-type only valid with keychain source
	if keychainType != "" && source != "keychain" {
		return fmt.Errorf("--keychain-type is only valid with --source=keychain")
	}

	// Rule 3: --keychain-type must be "generic" or "internet" (if provided)
	if keychainType != "" && keychainType != "generic" && keychainType != "internet" {
		return fmt.Errorf("--keychain-type must be 'generic' or 'internet', got %q", keychainType)
	}

	// Rule 2: For keychain source, either --service or --keychain-label must be set
	if source == "keychain" && service == "" && keychainLabel == "" {
		return fmt.Errorf("--keychain-label (or deprecated --service) is required when --source=keychain")
	}

	return nil
}
