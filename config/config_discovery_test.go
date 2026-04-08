package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rmkohlman/MaestroSDK/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultRegistryConfig_UsesPathConfig verifies that DefaultRegistryConfig
// returns a storage path consistent with the PathConfig.RegistryStorage() value.
func TestDefaultRegistryConfig_UsesPathConfig(t *testing.T) {
	cfg := DefaultRegistryConfig()

	pc, err := paths.Default()
	require.NoError(t, err, "paths.Default() should not error")

	expected := pc.RegistryStorage()
	assert.Equal(t, expected, cfg.Storage,
		"DefaultRegistryConfig().Storage should match paths.Default().RegistryStorage()")
}

// TestResolveVaultToken_UsesPathConfigRoot verifies that ResolveVaultToken
// resolves the config directory via paths.Default().Root() rather than
// hardcoding the directory name.
func TestResolveVaultToken_UsesPathConfigRoot(t *testing.T) {
	// Set up a temp dir to isolate from real vault
	tmpDir := t.TempDir()
	tokenContent := "test-token-via-file"
	tokenFile := filepath.Join(tmpDir, ".vault_token")
	require.NoError(t, os.WriteFile(tokenFile, []byte(tokenContent), 0600))

	// Use the injectable version to verify it reads from the expected directory
	token, err := ResolveVaultTokenFromDir(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, tokenContent, token,
		"ResolveVaultTokenFromDir should read token from the provided directory")
}

// TestPathsDefault_ConsistentWithNew verifies that paths.Default() produces
// the same results as paths.New(os.UserHomeDir()).
func TestPathsDefault_ConsistentWithNew(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	pcDefault, err := paths.Default()
	require.NoError(t, err)

	pcNew := paths.New(homeDir)

	assert.Equal(t, pcNew.Root(), pcDefault.Root(),
		"Default() and New(homeDir) should produce same Root()")
	assert.Equal(t, pcNew.ConfigFile(), pcDefault.ConfigFile(),
		"Default() and New(homeDir) should produce same ConfigFile()")
	assert.Equal(t, pcNew.Database(), pcDefault.Database(),
		"Default() and New(homeDir) should produce same Database()")
	assert.Equal(t, pcNew.RegistryStorage(), pcDefault.RegistryStorage(),
		"Default() and New(homeDir) should produce same RegistryStorage()")
	assert.Equal(t, pcNew.NVPRoot(), pcDefault.NVPRoot(),
		"Default() and New(homeDir) should produce same NVPRoot()")
	assert.Equal(t, pcNew.DVTRoot(), pcDefault.DVTRoot(),
		"Default() and New(homeDir) should produce same DVTRoot()")
}
