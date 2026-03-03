package registry

import (
	"context"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRegistryResolver_ResolveCLIOverride tests CLI flag takes precedence
func TestRegistryResolver_ResolveCLIOverride(t *testing.T) {
	tests := []struct {
		name         string
		typeAlias    string
		cliOverride  string
		defaultValue string
		wantSource   ResolutionSource
		wantEndpoint string
		wantErr      bool
	}{
		{
			name:         "CLI override takes precedence over default",
			typeAlias:    "oci",
			cliOverride:  "localhost:5050",
			defaultValue: "my-zot-registry",
			wantSource:   SourceCLI,
			wantEndpoint: "localhost:5050",
			wantErr:      false,
		},
		{
			name:         "CLI override with http scheme",
			typeAlias:    "pypi",
			cliOverride:  "http://localhost:3141",
			defaultValue: "my-devpi",
			wantSource:   SourceCLI,
			wantEndpoint: "http://localhost:3141",
			wantErr:      false,
		},
		{
			name:         "CLI override empty string is treated as no override",
			typeAlias:    "npm",
			cliOverride:  "",
			defaultValue: "my-verdaccio",
			wantSource:   SourceDefault,
			wantEndpoint: "http://localhost:4873", // Should use default registry
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := db.NewMockDataStore()

			// Set up default registry if specified
			if tt.defaultValue != "" {
				registryType := ResolveAlias(tt.typeAlias)
				registry := &models.Registry{
					Name:    tt.defaultValue,
					Type:    registryType,
					Enabled: true,
					Port:    5000,
					Status:  "running",
					Storage: "/tmp/" + tt.defaultValue,
				}
				err := mockDB.CreateRegistry(registry)
				require.NoError(t, err)

				// Set as default
				err = mockDB.SetDefault("registry-"+tt.typeAlias, tt.defaultValue)
				require.NoError(t, err)
			}

			resolver := NewRegistryResolver(mockDB, nil)
			result, err := resolver.Resolve(context.Background(), tt.typeAlias, tt.cliOverride)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantSource, result.Source, "Resolution source should match")
			if tt.cliOverride != "" {
				assert.Equal(t, tt.wantEndpoint, result.Endpoint, "Endpoint should match CLI override")
			}
		})
	}
}

// TestRegistryResolver_ResolveDefault tests falling back to default from DB
func TestRegistryResolver_ResolveDefault(t *testing.T) {
	mockDB := db.NewMockDataStore()
	resolver := NewRegistryResolver(mockDB, nil)

	// Create a default OCI registry
	registry := &models.Registry{
		Name:    "default-zot",
		Type:    "zot",
		Enabled: true,
		Port:    5000,
		Status:  "running",
		Storage: "/tmp/default-zot",
	}
	err := mockDB.CreateRegistry(registry)
	require.NoError(t, err)

	// Set as default
	err = mockDB.SetDefault("registry-oci", "default-zot")
	require.NoError(t, err)

	// Resolve with no CLI override
	result, err := resolver.Resolve(context.Background(), "oci", "")
	assert.NoError(t, err)
	assert.Equal(t, SourceDefault, result.Source, "Should use default source")
	assert.Equal(t, "default-zot", result.RegistryName)
	assert.Equal(t, "http://localhost:5000", result.Endpoint)
}

// TestRegistryResolver_ResolveNone tests when no default is configured
func TestRegistryResolver_ResolveNone(t *testing.T) {
	mockDB := db.NewMockDataStore()
	resolver := NewRegistryResolver(mockDB, nil)

	// Don't set any default
	result, err := resolver.Resolve(context.Background(), "oci", "")
	assert.NoError(t, err)
	assert.Equal(t, SourceNone, result.Source, "Should return SourceNone when no default")
	assert.Empty(t, result.RegistryName)
	assert.Empty(t, result.Endpoint)
}

// TestRegistryResolver_ResolveAll tests getting all enabled registries
func TestRegistryResolver_ResolveAll(t *testing.T) {
	mockDB := db.NewMockDataStore()
	resolver := NewRegistryResolver(mockDB, nil)

	// Create multiple enabled registries
	registries := []*models.Registry{
		{Name: "zot-1", Type: "zot", Enabled: true, Port: 5000, Status: "running", Storage: "/tmp/zot-1"},
		{Name: "devpi-1", Type: "devpi", Enabled: true, Port: 3141, Status: "running", Storage: "/tmp/devpi-1"},
		{Name: "verdaccio-1", Type: "verdaccio", Enabled: true, Port: 4873, Status: "stopped", Storage: "/tmp/verdaccio-1"},
		{Name: "zot-2", Type: "zot", Enabled: false, Port: 5001, Status: "stopped", Storage: "/tmp/zot-2"}, // Disabled
	}

	for _, reg := range registries {
		err := mockDB.CreateRegistry(reg)
		require.NoError(t, err)
	}

	// Resolve all enabled registries
	results, err := resolver.ResolveAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, results, 3, "Should return only enabled registries")

	// Verify all results are enabled
	for _, result := range results {
		assert.True(t, result.Enabled, "All results should be enabled")
	}
}

// TestRegistryResolver_ResolveAll_Empty tests when no registries are enabled
func TestRegistryResolver_ResolveAll_Empty(t *testing.T) {
	mockDB := db.NewMockDataStore()
	resolver := NewRegistryResolver(mockDB, nil)

	// Don't create any registries
	results, err := resolver.ResolveAll(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, results, "Should return empty list when no registries enabled")
}

// TestRegistryResolver_BuildEndpoint tests endpoint URL construction
func TestRegistryResolver_BuildEndpoint(t *testing.T) {
	tests := []struct {
		name         string
		registryType string
		port         int
		wantEndpoint string
	}{
		{
			name:         "zot with port 5000",
			registryType: "zot",
			port:         5000,
			wantEndpoint: "http://localhost:5000",
		},
		{
			name:         "devpi with port 3141",
			registryType: "devpi",
			port:         3141,
			wantEndpoint: "http://localhost:3141",
		},
		{
			name:         "verdaccio with custom port",
			registryType: "verdaccio",
			port:         8080,
			wantEndpoint: "http://localhost:8080",
		},
		{
			name:         "athens with port 3000",
			registryType: "athens",
			port:         3000,
			wantEndpoint: "http://localhost:3000",
		},
		{
			name:         "squid with port 3128",
			registryType: "squid",
			port:         3128,
			wantEndpoint: "http://localhost:3128",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := db.NewMockDataStore()
			resolver := NewRegistryResolver(mockDB, nil)

			registry := &models.Registry{
				Name: "test-registry",
				Type: tt.registryType,
				Port: tt.port,
			}

			endpoint := resolver.BuildEndpoint(registry)
			assert.Equal(t, tt.wantEndpoint, endpoint)
		})
	}
}

// TestRegistryResolver_ResolveByType tests resolving registry by concrete type
func TestRegistryResolver_ResolveByType(t *testing.T) {
	mockDB := db.NewMockDataStore()
	resolver := NewRegistryResolver(mockDB, nil)

	// Create registry with concrete type "zot"
	registry := &models.Registry{
		Name:    "my-zot",
		Type:    "zot",
		Enabled: true,
		Port:    5000,
		Status:  "running",
		Storage: "/tmp/my-zot",
	}
	err := mockDB.CreateRegistry(registry)
	require.NoError(t, err)

	err = mockDB.SetDefault("registry-oci", "my-zot")
	require.NoError(t, err)

	// Should work with both alias and concrete type
	resultByAlias, err := resolver.Resolve(context.Background(), "oci", "")
	assert.NoError(t, err)
	assert.Equal(t, "my-zot", resultByAlias.RegistryName)

	resultByType, err := resolver.Resolve(context.Background(), "zot", "")
	assert.NoError(t, err)
	assert.Equal(t, "my-zot", resultByType.RegistryName)
}

// TestRegistryResolver_InvalidAlias tests resolving with invalid alias
func TestRegistryResolver_InvalidAlias(t *testing.T) {
	mockDB := db.NewMockDataStore()
	resolver := NewRegistryResolver(mockDB, nil)

	// Try to resolve with invalid alias
	result, err := resolver.Resolve(context.Background(), "invalid-type", "")

	// Should not error, but return SourceNone
	assert.NoError(t, err)
	assert.Equal(t, SourceNone, result.Source)
}

// TestRegistryResolver_DisabledRegistry tests resolving a disabled registry
func TestRegistryResolver_DisabledRegistry(t *testing.T) {
	mockDB := db.NewMockDataStore()
	resolver := NewRegistryResolver(mockDB, nil)

	// Create disabled registry
	registry := &models.Registry{
		Name:    "disabled-zot",
		Type:    "zot",
		Enabled: false, // Disabled
		Port:    5000,
		Status:  "stopped",
		Storage: "/tmp/disabled-zot",
	}
	err := mockDB.CreateRegistry(registry)
	require.NoError(t, err)

	// Set as default (even though disabled)
	err = mockDB.SetDefault("registry-oci", "disabled-zot")
	require.NoError(t, err)

	// Resolve should return the registry even if disabled
	// (Resolver doesn't filter by enabled - that's for lifecycle manager)
	result, err := resolver.Resolve(context.Background(), "oci", "")
	assert.NoError(t, err)
	assert.Equal(t, SourceDefault, result.Source)
	assert.Equal(t, "disabled-zot", result.RegistryName)
	assert.False(t, result.Enabled)
}
