package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// TDD Phase 2 (RED): Auto-Token Resolution Tests (v0.41.0)
// =============================================================================
// New types and functions being introduced to support automatic creation and
// persistence of MaestroVault tokens when none is available:
//
//	type VaultConfig struct {
//	    Token string `mapstructure:"token"`
//	}
//
//	type TokenCreator func() (string, error)
//
//	func resolveExistingToken(configDir string) string
//	func createToken(creator TokenCreator) (string, error)
//	func persistToken(configDir, token string) error
//	func ResolveVaultToken(creator ...TokenCreator) (string, error)
//	func defaultTokenCreator() (string, error)
//
// Resolution chain (highest priority first):
//   1. MAV_TOKEN environment variable
//   2. vault.token in viper config
//   3. ~/.devopsmaestro/.vault_token file
//   4. Auto-create via TokenCreator (mav token create)
//
// ALL tests in this section WILL FAIL TO COMPILE until the above types and
// functions are added to config/vault_token.go (new file).
// =============================================================================

// ---------------------------------------------------------------------------
// Section 1: Type/Interface Tests (compile-time existence checks)
// ---------------------------------------------------------------------------

// TestVaultConfig_StructExists verifies that VaultConfig exists as a struct
// with a Token string field.
//
// WILL FAIL TO COMPILE — VaultConfig struct does not exist yet.
func TestVaultConfig_StructExists(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// VaultConfig type does not exist yet.
	var cfg VaultConfig
	cfg.Token = "dvm-test-token"
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, "dvm-test-token", cfg.Token,
		"VaultConfig.Token must be a string field that stores and retrieves correctly")
}

// TestConfig_HasVaultField verifies that the Config struct has a Vault field
// of type VaultConfig.
//
// WILL FAIL TO COMPILE — Config.Vault field does not exist yet.
func TestConfig_HasVaultField(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// Config.Vault and VaultConfig do not exist yet.
	cfg := Config{
		Vault: VaultConfig{
			Token: "some-token",
		},
	}
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, "some-token", cfg.Vault.Token,
		"Config.Vault.Token must be accessible through the nested VaultConfig struct")
}

// TestTokenCreator_TypeExists verifies that TokenCreator is a function type
// with the exact signature: func() (string, error).
//
// WILL FAIL TO COMPILE — TokenCreator type does not exist yet.
func TestTokenCreator_TypeExists(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// TokenCreator type does not exist yet.
	var creator TokenCreator = func() (string, error) {
		return "test-token-abc123", nil
	}
	// ─────────────────────────────────────────────────────────────────────────

	require.NotNil(t, creator, "TokenCreator variable must be assignable from a func() (string, error)")
	token, err := creator()
	require.NoError(t, err)
	assert.Equal(t, "test-token-abc123", token,
		"TokenCreator must be callable and return (string, error)")
}

// ---------------------------------------------------------------------------
// Section 2: resolveExistingToken Tests
// ---------------------------------------------------------------------------

// TestResolveExistingToken_EnvVarWins verifies that when the MAV_TOKEN
// environment variable is set, resolveExistingToken returns it immediately.
//
// WILL FAIL TO COMPILE — resolveExistingToken does not exist yet.
func TestResolveExistingToken_EnvVarWins(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	t.Setenv("MAV_TOKEN", "env-token-value")

	tmpDir := t.TempDir()

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// resolveExistingToken does not exist yet.
	token := resolveExistingToken(tmpDir)
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, "env-token-value", token,
		"resolveExistingToken must return the MAV_TOKEN env var when set")
}

// TestResolveExistingToken_ViperConfigFallback verifies that when MAV_TOKEN
// env is empty but vault.token is set in viper, resolveExistingToken returns
// the viper value.
//
// WILL FAIL TO COMPILE — resolveExistingToken does not exist yet.
func TestResolveExistingToken_ViperConfigFallback(t *testing.T) {
	t.Setenv("MAV_TOKEN", "") // explicitly unset

	viper.Reset()
	defer viper.Reset()
	viper.Set("vault.token", "viper-token-value")

	tmpDir := t.TempDir()

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	token := resolveExistingToken(tmpDir)
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, "viper-token-value", token,
		"resolveExistingToken must fall back to vault.token in viper config when MAV_TOKEN is unset")
}

// TestResolveExistingToken_FileTokenFallback verifies that when both MAV_TOKEN
// env and viper vault.token are empty, resolveExistingToken reads the token
// from the .vault_token file in the configDir.
//
// WILL FAIL TO COMPILE — resolveExistingToken does not exist yet.
func TestResolveExistingToken_FileTokenFallback(t *testing.T) {
	t.Setenv("MAV_TOKEN", "") // explicitly unset

	viper.Reset()
	defer viper.Reset()
	// vault.token NOT set in viper

	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, ".vault_token")
	require.NoError(t, os.WriteFile(tokenFile, []byte("file-token-value"), 0600))

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	token := resolveExistingToken(tmpDir)
	// ─────────────────────────────────────────────────────────────────────────

	assert.Equal(t, "file-token-value", token,
		"resolveExistingToken must read the token from <configDir>/.vault_token when env and viper are empty")
}

// TestResolveExistingToken_EmptyEverywhere verifies that when no token is
// available from any source, resolveExistingToken returns an empty string.
//
// WILL FAIL TO COMPILE — resolveExistingToken does not exist yet.
func TestResolveExistingToken_EmptyEverywhere(t *testing.T) {
	t.Setenv("MAV_TOKEN", "") // explicitly unset

	viper.Reset()
	defer viper.Reset()
	// vault.token NOT set in viper

	tmpDir := t.TempDir()
	// No .vault_token file written

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	token := resolveExistingToken(tmpDir)
	// ─────────────────────────────────────────────────────────────────────────

	assert.Empty(t, token,
		"resolveExistingToken must return empty string when no token is available from any source")
}

// ---------------------------------------------------------------------------
// Section 3: persistToken Tests
// ---------------------------------------------------------------------------

// TestPersistToken_WritesFile verifies that persistToken writes the token to
// <configDir>/.vault_token.
//
// WILL FAIL TO COMPILE — persistToken does not exist yet.
func TestPersistToken_WritesFile(t *testing.T) {
	tmpDir := t.TempDir()

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// persistToken does not exist yet.
	err := persistToken(tmpDir, "my-persisted-token")
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, err, "persistToken must write the token without error")

	tokenFile := filepath.Join(tmpDir, ".vault_token")
	content, readErr := os.ReadFile(tokenFile)
	require.NoError(t, readErr, ".vault_token file must exist after persistToken")
	assert.Equal(t, "my-persisted-token", string(content),
		"token file content must match the token passed to persistToken")
}

// TestPersistToken_FilePermissions verifies that the .vault_token file written
// by persistToken has 0600 permissions (owner read/write only).
//
// WILL FAIL TO COMPILE — persistToken does not exist yet.
func TestPersistToken_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	err := persistToken(tmpDir, "perms-test-token")
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, err)

	tokenFile := filepath.Join(tmpDir, ".vault_token")
	info, statErr := os.Stat(tokenFile)
	require.NoError(t, statErr, ".vault_token file must exist")
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm(),
		".vault_token file must have 0600 permissions (owner read/write only)")
}

// TestPersistToken_OverwritesExisting verifies that persistToken overwrites
// an existing .vault_token file with the new token.
//
// WILL FAIL TO COMPILE — persistToken does not exist yet.
func TestPersistToken_OverwritesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, ".vault_token")

	// Write an existing token file
	require.NoError(t, os.WriteFile(tokenFile, []byte("old-token"), 0600))

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	err := persistToken(tmpDir, "new-token")
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, err, "persistToken must overwrite existing token file without error")

	content, readErr := os.ReadFile(tokenFile)
	require.NoError(t, readErr)
	assert.Equal(t, "new-token", string(content),
		"persistToken must overwrite the existing token file with the new value")
}

// ---------------------------------------------------------------------------
// Section 4: ResolveVaultToken Integration Tests
// ---------------------------------------------------------------------------

// TestResolveVaultToken_EnvVarShortCircuits verifies that when MAV_TOKEN is
// set, ResolveVaultToken returns it immediately without calling the TokenCreator.
//
// WILL FAIL TO COMPILE — ResolveVaultToken does not exist yet.
func TestResolveVaultToken_EnvVarShortCircuits(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	t.Setenv("MAV_TOKEN", "env-shortcircuit-token")

	creatorCalled := false
	mockCreator := TokenCreator(func() (string, error) {
		creatorCalled = true
		return "should-not-be-created", nil
	})

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// ResolveVaultToken does not exist yet.
	token, err := ResolveVaultToken(mockCreator)
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, err)
	assert.Equal(t, "env-shortcircuit-token", token,
		"ResolveVaultToken must return MAV_TOKEN env var without calling the creator")
	assert.False(t, creatorCalled,
		"TokenCreator must NOT be called when MAV_TOKEN env var is set")
}

// TestResolveVaultToken_ConfigShortCircuits verifies that when vault.token is
// set in viper, ResolveVaultToken returns it without calling the TokenCreator.
//
// WILL FAIL TO COMPILE — ResolveVaultToken does not exist yet.
func TestResolveVaultToken_ConfigShortCircuits(t *testing.T) {
	t.Setenv("MAV_TOKEN", "") // explicitly unset

	viper.Reset()
	defer viper.Reset()
	viper.Set("vault.token", "viper-shortcircuit-token")

	creatorCalled := false
	mockCreator := TokenCreator(func() (string, error) {
		creatorCalled = true
		return "should-not-be-created", nil
	})

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	token, err := ResolveVaultToken(mockCreator)
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, err)
	assert.Equal(t, "viper-shortcircuit-token", token,
		"ResolveVaultToken must return viper vault.token without calling the creator")
	assert.False(t, creatorCalled,
		"TokenCreator must NOT be called when vault.token is set in viper config")
}

// TestResolveVaultToken_FileTokenShortCircuits verifies that when a .vault_token
// file exists, ResolveVaultToken returns the file token without calling the TokenCreator.
//
// This test manipulates the config dir used by ResolveVaultToken. Because
// ResolveVaultToken uses the real home directory by default, we test the
// short-circuit behaviour by ensuring no creator is called when a token is
// already available via the file path. The test uses env var short-circuit
// detection; file-specific short-circuiting is implicitly covered by the
// persistToken+resolveExistingToken contract.
//
// WILL FAIL TO COMPILE — ResolveVaultToken does not exist yet.
func TestResolveVaultToken_FileTokenShortCircuits(t *testing.T) {
	t.Setenv("MAV_TOKEN", "") // explicitly unset

	viper.Reset()
	defer viper.Reset()
	// vault.token NOT set in viper

	creatorCalled := false
	mockCreator := TokenCreator(func() (string, error) {
		creatorCalled = true
		return "should-not-be-created", nil
	})

	// Verify that when a valid token is already resolved (from file),
	// the creator is not called. We test this via ResolveVaultTokenFromDir
	// to avoid touching real ~/.devopsmaestro.
	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, ".vault_token")
	require.NoError(t, os.WriteFile(tokenFile, []byte("file-shortcircuit-token"), 0600))

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// ResolveVaultTokenFromDir does not exist yet.
	token, err := ResolveVaultTokenFromDir(tmpDir, mockCreator)
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, err)
	assert.Equal(t, "file-shortcircuit-token", token,
		"ResolveVaultTokenFromDir must return file token without calling the creator")
	assert.False(t, creatorCalled,
		"TokenCreator must NOT be called when .vault_token file exists")
}

// TestResolveVaultToken_CallsCreatorWhenNoToken verifies that when no token
// exists anywhere, ResolveVaultToken calls the TokenCreator and returns the
// created token.
//
// WILL FAIL TO COMPILE — ResolveVaultTokenFromDir does not exist yet.
func TestResolveVaultToken_CallsCreatorWhenNoToken(t *testing.T) {
	t.Setenv("MAV_TOKEN", "") // explicitly unset

	viper.Reset()
	defer viper.Reset()
	// vault.token NOT set in viper

	tmpDir := t.TempDir()
	// No .vault_token file written

	creatorCalled := false
	mockCreator := TokenCreator(func() (string, error) {
		creatorCalled = true
		return "newly-created-token", nil
	})

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	token, err := ResolveVaultTokenFromDir(tmpDir, mockCreator)
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, err)
	assert.Equal(t, "newly-created-token", token,
		"ResolveVaultTokenFromDir must call the creator and return the created token when no token exists")
	assert.True(t, creatorCalled,
		"TokenCreator MUST be called when no token is available from any source")
}

// TestResolveVaultToken_PersistsCreatedToken verifies that after the
// TokenCreator successfully creates a token, ResolveVaultToken persists
// it to the .vault_token file so future calls don't need to create again.
//
// WILL FAIL TO COMPILE — ResolveVaultTokenFromDir does not exist yet.
func TestResolveVaultToken_PersistsCreatedToken(t *testing.T) {
	t.Setenv("MAV_TOKEN", "") // explicitly unset

	viper.Reset()
	defer viper.Reset()

	tmpDir := t.TempDir()
	// No .vault_token file written initially

	mockCreator := TokenCreator(func() (string, error) {
		return "auto-created-persist-token", nil
	})

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	_, err := ResolveVaultTokenFromDir(tmpDir, mockCreator)
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, err)

	// Verify the token was persisted to disk
	tokenFile := filepath.Join(tmpDir, ".vault_token")
	content, readErr := os.ReadFile(tokenFile)
	require.NoError(t, readErr, ".vault_token file must be created after auto-token creation")
	assert.Equal(t, "auto-created-persist-token", string(content),
		"persisted token must match the value returned by the TokenCreator")
}

// TestResolveVaultToken_CreatorError_ReturnsEmpty verifies graceful degradation:
// if the TokenCreator fails (e.g., mav not installed), ResolveVaultToken
// returns ("", nil) rather than propagating the error. The build can proceed
// without credentials; individual credential resolution will fail with a
// more specific error.
//
// WILL FAIL TO COMPILE — ResolveVaultTokenFromDir does not exist yet.
func TestResolveVaultToken_CreatorError_ReturnsEmpty(t *testing.T) {
	t.Setenv("MAV_TOKEN", "") // explicitly unset

	viper.Reset()
	defer viper.Reset()

	tmpDir := t.TempDir()

	failingCreator := TokenCreator(func() (string, error) {
		return "", os.ErrNotExist // simulate mav not found
	})

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	token, err := ResolveVaultTokenFromDir(tmpDir, failingCreator)
	// ─────────────────────────────────────────────────────────────────────────

	// Graceful degradation: no hard error propagated to caller
	assert.NoError(t, err,
		"ResolveVaultTokenFromDir must NOT return an error when the TokenCreator fails — graceful degradation")
	assert.Empty(t, token,
		"ResolveVaultTokenFromDir must return empty string when the TokenCreator fails")
}

// TestResolveVaultToken_DefaultCreator_MavNotInPath verifies that when no
// TokenCreator is injected AND mav is not in PATH, ResolveVaultToken returns
// ("", nil) without panicking — graceful degradation.
//
// WILL FAIL TO COMPILE — ResolveVaultTokenFromDir does not exist yet.
func TestResolveVaultToken_DefaultCreator_MavNotInPath(t *testing.T) {
	t.Setenv("MAV_TOKEN", "") // explicitly unset
	// Explicitly set PATH to something that won't have mav
	t.Setenv("PATH", t.TempDir())

	viper.Reset()
	defer viper.Reset()

	tmpDir := t.TempDir()
	// No .vault_token file

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// Call with no creator → uses defaultTokenCreator internally
	token, err := ResolveVaultTokenFromDir(tmpDir)
	// ─────────────────────────────────────────────────────────────────────────

	assert.NoError(t, err,
		"ResolveVaultTokenFromDir without creator must not return an error when mav is not in PATH")
	assert.Empty(t, token,
		"ResolveVaultTokenFromDir without creator must return empty string when mav is not in PATH")
}

// TestResolveVaultToken_CreatedTokenHasCorrectFilePerms verifies that a token
// persisted after auto-creation has 0600 file permissions.
//
// WILL FAIL TO COMPILE — ResolveVaultTokenFromDir does not exist yet.
func TestResolveVaultToken_CreatedTokenHasCorrectFilePerms(t *testing.T) {
	t.Setenv("MAV_TOKEN", "") // explicitly unset

	viper.Reset()
	defer viper.Reset()

	tmpDir := t.TempDir()

	mockCreator := TokenCreator(func() (string, error) {
		return "perm-check-token", nil
	})

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	_, err := ResolveVaultTokenFromDir(tmpDir, mockCreator)
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, err)

	tokenFile := filepath.Join(tmpDir, ".vault_token")
	info, statErr := os.Stat(tokenFile)
	require.NoError(t, statErr, ".vault_token file must be created after auto-token creation")
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm(),
		"auto-created .vault_token file must have 0600 permissions")
}

// ---------------------------------------------------------------------------
// Section 5: defaultTokenCreator Tests
// ---------------------------------------------------------------------------

// TestDefaultTokenCreator_Exists verifies that the defaultTokenCreator
// function exists and is assignable to TokenCreator.
//
// WILL FAIL TO COMPILE — defaultTokenCreator does not exist yet.
func TestDefaultTokenCreator_Exists(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// defaultTokenCreator does not exist yet.
	var creator TokenCreator = defaultTokenCreator
	// ─────────────────────────────────────────────────────────────────────────

	require.NotNil(t, creator,
		"defaultTokenCreator must exist and be assignable to TokenCreator")
}

// TestDefaultTokenCreator_ChecksMavInPath verifies that defaultTokenCreator
// returns an error when `mav` is not found in PATH.
//
// WILL FAIL TO COMPILE — defaultTokenCreator does not exist yet.
func TestDefaultTokenCreator_ChecksMavInPath(t *testing.T) {
	// Point PATH at an empty temp dir — mav won't be there.
	t.Setenv("PATH", t.TempDir())

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	token, err := defaultTokenCreator()
	// ─────────────────────────────────────────────────────────────────────────

	assert.Error(t, err,
		"defaultTokenCreator must return an error when mav is not found in PATH")
	assert.Empty(t, token,
		"defaultTokenCreator must return empty string when mav is not found in PATH")
}

// ---------------------------------------------------------------------------
// Section 6: Resolution Priority Table Tests
// ---------------------------------------------------------------------------

// TestResolveExistingToken_PriorityOrder verifies the full priority chain of
// resolveExistingToken using a table-driven approach: env > viper > file > empty.
//
// WILL FAIL TO COMPILE — resolveExistingToken does not exist yet.
func TestResolveExistingToken_PriorityOrder(t *testing.T) {
	tests := []struct {
		name          string
		envToken      string
		viperToken    string
		fileToken     string
		expectedToken string
	}{
		{
			name:          "env wins over viper and file",
			envToken:      "env-wins",
			viperToken:    "viper-loses",
			fileToken:     "file-loses",
			expectedToken: "env-wins",
		},
		{
			name:          "viper wins over file when env is empty",
			envToken:      "",
			viperToken:    "viper-wins",
			fileToken:     "file-loses",
			expectedToken: "viper-wins",
		},
		{
			name:          "file wins when env and viper are empty",
			envToken:      "",
			viperToken:    "",
			fileToken:     "file-wins",
			expectedToken: "file-wins",
		},
		{
			name:          "empty everywhere returns empty string",
			envToken:      "",
			viperToken:    "",
			fileToken:     "",
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up env
			t.Setenv("MAV_TOKEN", tt.envToken)

			// Set up viper
			viper.Reset()
			defer viper.Reset()
			if tt.viperToken != "" {
				viper.Set("vault.token", tt.viperToken)
			}

			// Set up file
			tmpDir := t.TempDir()
			if tt.fileToken != "" {
				tokenFile := filepath.Join(tmpDir, ".vault_token")
				require.NoError(t, os.WriteFile(tokenFile, []byte(tt.fileToken), 0600))
			}

			// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────
			token := resolveExistingToken(tmpDir)
			// ─────────────────────────────────────────────────────────────────

			assert.Equal(t, tt.expectedToken, token,
				"resolveExistingToken priority mismatch for case %q", tt.name)
		})
	}
}
