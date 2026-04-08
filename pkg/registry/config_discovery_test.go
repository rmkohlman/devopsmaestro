package registry

import (
	"testing"

	"github.com/rmkohlman/MaestroSDK/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultRegistryConfig_StorageUsesPathConfig verifies that the OCI registry
// default config uses PathConfig for storage path resolution.
func TestDefaultRegistryConfig_StorageUsesPathConfig(t *testing.T) {
	cfg := DefaultRegistryConfig()
	pc, err := paths.Default()
	require.NoError(t, err)

	assert.Equal(t, pc.RegistryStorage(), cfg.Storage,
		"DefaultRegistryConfig().Storage should match PathConfig.RegistryStorage()")
}

// TestDefaultNpmProxyConfig_StorageUsesPathConfig verifies that the npm proxy
// default config uses PathConfig for storage path resolution.
func TestDefaultNpmProxyConfig_StorageUsesPathConfig(t *testing.T) {
	cfg := DefaultNpmProxyConfig()
	pc, err := paths.Default()
	require.NoError(t, err)

	assert.Equal(t, pc.VerdaccioStorage(), cfg.Storage,
		"DefaultNpmProxyConfig().Storage should match PathConfig.VerdaccioStorage()")
}

// TestDefaultGoModuleConfig_StorageUsesPathConfig verifies that the Go module
// proxy default config uses PathConfig for storage path resolution.
func TestDefaultGoModuleConfig_StorageUsesPathConfig(t *testing.T) {
	cfg := DefaultGoModuleConfig()
	pc, err := paths.Default()
	require.NoError(t, err)

	assert.Equal(t, pc.AthensStorage(), cfg.Storage,
		"DefaultGoModuleConfig().Storage should match PathConfig.AthensStorage()")
}

// TestDefaultPyPIProxyConfig_StorageUsesPathConfig verifies that the PyPI proxy
// default config uses PathConfig for storage path resolution.
func TestDefaultPyPIProxyConfig_StorageUsesPathConfig(t *testing.T) {
	cfg := DefaultPyPIProxyConfig()
	pc, err := paths.Default()
	require.NoError(t, err)

	assert.Equal(t, pc.DevpiStorage(), cfg.Storage,
		"DefaultPyPIProxyConfig().Storage should match PathConfig.DevpiStorage()")
}

// TestDefaultHttpProxyConfig_PathsUsePathConfig verifies that the HTTP proxy
// default config uses PathConfig for its directories.
func TestDefaultHttpProxyConfig_PathsUsePathConfig(t *testing.T) {
	cfg := DefaultHttpProxyConfig()
	pc, err := paths.Default()
	require.NoError(t, err)

	squidDir := pc.SquidDir()
	assert.Contains(t, cfg.CacheDir, squidDir,
		"DefaultHttpProxyConfig().CacheDir should be under PathConfig.SquidDir()")
	assert.Contains(t, cfg.LogDir, squidDir,
		"DefaultHttpProxyConfig().LogDir should be under PathConfig.SquidDir()")
	assert.Contains(t, cfg.PidFile, squidDir,
		"DefaultHttpProxyConfig().PidFile should be under PathConfig.SquidDir()")
}
