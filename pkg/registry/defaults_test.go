package registry

import (
	"context"
	"testing"

	"devopsmaestro/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRegistryDefaults_GetOCIRegistry tests getting the default OCI registry name
func TestRegistryDefaults_GetOCIRegistry(t *testing.T) {
	mockDB := db.NewMockDataStore()
	defaults := NewRegistryDefaults(mockDB)

	// Set a default OCI registry
	err := mockDB.SetDefault("registry-oci", "my-zot-registry")
	require.NoError(t, err)

	// Get the default OCI registry
	registryName, err := defaults.GetOCIRegistry(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "my-zot-registry", registryName)
}

// TestRegistryDefaults_GetOCIRegistry_NotSet tests when no default is set
func TestRegistryDefaults_GetOCIRegistry_NotSet(t *testing.T) {
	mockDB := db.NewMockDataStore()
	defaults := NewRegistryDefaults(mockDB)

	// Get the default OCI registry when not set
	registryName, err := defaults.GetOCIRegistry(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, registryName, "Should return empty string when no default set")
}

// TestRegistryDefaults_SetOCIRegistry tests setting the default OCI registry
func TestRegistryDefaults_SetOCIRegistry(t *testing.T) {
	mockDB := db.NewMockDataStore()
	defaults := NewRegistryDefaults(mockDB)

	// Set the default OCI registry
	err := defaults.SetOCIRegistry(context.Background(), "my-zot-registry")
	assert.NoError(t, err)

	// Verify it was set in the database
	value, err := mockDB.GetDefault("registry-oci")
	assert.NoError(t, err)
	assert.Equal(t, "my-zot-registry", value)
}

// TestRegistryDefaults_GetPyPIRegistry tests getting the default PyPI registry
func TestRegistryDefaults_GetPyPIRegistry(t *testing.T) {
	mockDB := db.NewMockDataStore()
	defaults := NewRegistryDefaults(mockDB)

	err := mockDB.SetDefault("registry-pypi", "my-devpi")
	require.NoError(t, err)

	registryName, err := defaults.GetPyPIRegistry(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "my-devpi", registryName)
}

// TestRegistryDefaults_SetPyPIRegistry tests setting the default PyPI registry
func TestRegistryDefaults_SetPyPIRegistry(t *testing.T) {
	mockDB := db.NewMockDataStore()
	defaults := NewRegistryDefaults(mockDB)

	err := defaults.SetPyPIRegistry(context.Background(), "my-devpi")
	assert.NoError(t, err)

	value, err := mockDB.GetDefault("registry-pypi")
	assert.NoError(t, err)
	assert.Equal(t, "my-devpi", value)
}

// TestRegistryDefaults_GetNPMRegistry tests getting the default NPM registry
func TestRegistryDefaults_GetNPMRegistry(t *testing.T) {
	mockDB := db.NewMockDataStore()
	defaults := NewRegistryDefaults(mockDB)

	err := mockDB.SetDefault("registry-npm", "my-verdaccio")
	require.NoError(t, err)

	registryName, err := defaults.GetNPMRegistry(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "my-verdaccio", registryName)
}

// TestRegistryDefaults_SetNPMRegistry tests setting the default NPM registry
func TestRegistryDefaults_SetNPMRegistry(t *testing.T) {
	mockDB := db.NewMockDataStore()
	defaults := NewRegistryDefaults(mockDB)

	err := defaults.SetNPMRegistry(context.Background(), "my-verdaccio")
	assert.NoError(t, err)

	value, err := mockDB.GetDefault("registry-npm")
	assert.NoError(t, err)
	assert.Equal(t, "my-verdaccio", value)
}

// TestRegistryDefaults_GetGoRegistry tests getting the default Go registry
func TestRegistryDefaults_GetGoRegistry(t *testing.T) {
	mockDB := db.NewMockDataStore()
	defaults := NewRegistryDefaults(mockDB)

	err := mockDB.SetDefault("registry-go", "my-athens")
	require.NoError(t, err)

	registryName, err := defaults.GetGoRegistry(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "my-athens", registryName)
}

// TestRegistryDefaults_SetGoRegistry tests setting the default Go registry
func TestRegistryDefaults_SetGoRegistry(t *testing.T) {
	mockDB := db.NewMockDataStore()
	defaults := NewRegistryDefaults(mockDB)

	err := defaults.SetGoRegistry(context.Background(), "my-athens")
	assert.NoError(t, err)

	value, err := mockDB.GetDefault("registry-go")
	assert.NoError(t, err)
	assert.Equal(t, "my-athens", value)
}

// TestRegistryDefaults_GetHTTPRegistry tests getting the default HTTP proxy registry
func TestRegistryDefaults_GetHTTPRegistry(t *testing.T) {
	mockDB := db.NewMockDataStore()
	defaults := NewRegistryDefaults(mockDB)

	err := mockDB.SetDefault("registry-http", "my-squid")
	require.NoError(t, err)

	registryName, err := defaults.GetHTTPRegistry(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "my-squid", registryName)
}

// TestRegistryDefaults_SetHTTPRegistry tests setting the default HTTP proxy registry
func TestRegistryDefaults_SetHTTPRegistry(t *testing.T) {
	mockDB := db.NewMockDataStore()
	defaults := NewRegistryDefaults(mockDB)

	err := defaults.SetHTTPRegistry(context.Background(), "my-squid")
	assert.NoError(t, err)

	value, err := mockDB.GetDefault("registry-http")
	assert.NoError(t, err)
	assert.Equal(t, "my-squid", value)
}

// TestRegistryDefaults_GetIdleTimeout tests getting the global idle timeout
func TestRegistryDefaults_GetIdleTimeout(t *testing.T) {
	mockDB := db.NewMockDataStore()
	defaults := NewRegistryDefaults(mockDB)

	// Set idle timeout
	err := mockDB.SetDefault("registry-idle-timeout", "45m")
	require.NoError(t, err)

	timeout, err := defaults.GetIdleTimeout(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "45m", timeout)
}

// TestRegistryDefaults_SetIdleTimeout tests setting the global idle timeout
func TestRegistryDefaults_SetIdleTimeout(t *testing.T) {
	mockDB := db.NewMockDataStore()
	defaults := NewRegistryDefaults(mockDB)

	err := defaults.SetIdleTimeout(context.Background(), "60m")
	assert.NoError(t, err)

	value, err := mockDB.GetDefault("registry-idle-timeout")
	assert.NoError(t, err)
	assert.Equal(t, "60m", value)
}

// TestRegistryDefaults_GetIdleTimeout_DefaultValue tests default idle timeout
func TestRegistryDefaults_GetIdleTimeout_DefaultValue(t *testing.T) {
	mockDB := db.NewMockDataStore()
	defaults := NewRegistryDefaults(mockDB)

	// Don't set any value - should return default
	timeout, err := defaults.GetIdleTimeout(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "30m", timeout, "Should return 30m as default when not set")
}

// TestRegistryDefaults_GetAllDefaults tests getting all registry defaults at once
func TestRegistryDefaults_GetAllDefaults(t *testing.T) {
	mockDB := db.NewMockDataStore()
	defaults := NewRegistryDefaults(mockDB)

	// Set multiple defaults
	require.NoError(t, mockDB.SetDefault("registry-oci", "my-zot"))
	require.NoError(t, mockDB.SetDefault("registry-pypi", "my-devpi"))
	require.NoError(t, mockDB.SetDefault("registry-npm", "my-verdaccio"))
	require.NoError(t, mockDB.SetDefault("registry-idle-timeout", "45m"))

	allDefaults, err := defaults.GetAllDefaults(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, allDefaults)

	// Verify all set defaults are present
	assert.Equal(t, "my-zot", allDefaults["registry-oci"])
	assert.Equal(t, "my-devpi", allDefaults["registry-pypi"])
	assert.Equal(t, "my-verdaccio", allDefaults["registry-npm"])
	assert.Equal(t, "45m", allDefaults["registry-idle-timeout"])

	// Defaults that weren't set should be empty or have default values
	assert.Contains(t, allDefaults, "registry-go", "Should include all registry types")
	assert.Contains(t, allDefaults, "registry-http", "Should include all registry types")
}

// TestRegistryDefaults_GetAllDefaults_Empty tests getting all defaults when none are set
func TestRegistryDefaults_GetAllDefaults_Empty(t *testing.T) {
	mockDB := db.NewMockDataStore()
	defaults := NewRegistryDefaults(mockDB)

	allDefaults, err := defaults.GetAllDefaults(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, allDefaults)

	// Should include keys for all registry types with empty/default values
	assert.Contains(t, allDefaults, "registry-oci")
	assert.Contains(t, allDefaults, "registry-pypi")
	assert.Contains(t, allDefaults, "registry-npm")
	assert.Contains(t, allDefaults, "registry-go")
	assert.Contains(t, allDefaults, "registry-http")
	assert.Contains(t, allDefaults, "registry-idle-timeout")

	// Idle timeout should have default value
	assert.Equal(t, "30m", allDefaults["registry-idle-timeout"])
}

// TestRegistryDefaults_ClearDefault tests clearing a specific default
func TestRegistryDefaults_ClearDefault(t *testing.T) {
	mockDB := db.NewMockDataStore()
	defaults := NewRegistryDefaults(mockDB)

	// Set a default
	err := defaults.SetOCIRegistry(context.Background(), "my-zot")
	require.NoError(t, err)

	// Clear the default
	err = defaults.ClearDefault(context.Background(), "registry-oci")
	assert.NoError(t, err)

	// Verify it's cleared
	registryName, err := defaults.GetOCIRegistry(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, registryName)
}

// TestRegistryDefaults_GetByAlias tests getting default by type alias
func TestRegistryDefaults_GetByAlias(t *testing.T) {
	tests := []struct {
		name      string
		alias     string
		setValue  string
		expectKey string
	}{
		{
			name:      "get by oci alias",
			alias:     "oci",
			setValue:  "my-zot",
			expectKey: "registry-oci",
		},
		{
			name:      "get by pypi alias",
			alias:     "pypi",
			setValue:  "my-devpi",
			expectKey: "registry-pypi",
		},
		{
			name:      "get by npm alias",
			alias:     "npm",
			setValue:  "my-verdaccio",
			expectKey: "registry-npm",
		},
		{
			name:      "get by go alias",
			alias:     "go",
			setValue:  "my-athens",
			expectKey: "registry-go",
		},
		{
			name:      "get by http alias",
			alias:     "http",
			setValue:  "my-squid",
			expectKey: "registry-http",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := db.NewMockDataStore()
			defaults := NewRegistryDefaults(mockDB)

			// Set the default using the underlying key
			err := mockDB.SetDefault(tt.expectKey, tt.setValue)
			require.NoError(t, err)

			// Get by alias
			registryName, err := defaults.GetByAlias(context.Background(), tt.alias)
			assert.NoError(t, err)
			assert.Equal(t, tt.setValue, registryName)
		})
	}
}

// TestRegistryDefaults_SetByAlias tests setting default by type alias
func TestRegistryDefaults_SetByAlias(t *testing.T) {
	tests := []struct {
		name      string
		alias     string
		setValue  string
		expectKey string
	}{
		{
			name:      "set by oci alias",
			alias:     "oci",
			setValue:  "my-zot",
			expectKey: "registry-oci",
		},
		{
			name:      "set by pypi alias",
			alias:     "pypi",
			setValue:  "my-devpi",
			expectKey: "registry-pypi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := db.NewMockDataStore()
			defaults := NewRegistryDefaults(mockDB)

			// Set by alias
			err := defaults.SetByAlias(context.Background(), tt.alias, tt.setValue)
			assert.NoError(t, err)

			// Verify it was set with the correct key
			value, err := mockDB.GetDefault(tt.expectKey)
			assert.NoError(t, err)
			assert.Equal(t, tt.setValue, value)
		})
	}
}
