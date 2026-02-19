package resolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultResolverConfig(t *testing.T) {
	config := DefaultResolverConfig()

	assert.Equal(t, DefaultTheme, config.DefaultTheme)
	assert.False(t, config.CacheEnabled)
	assert.Equal(t, 300, config.CacheTTL)
}

func TestNewThemeResolverFactory(t *testing.T) {
	factory := NewThemeResolverFactory()

	assert.NotNil(t, factory)
	assert.IsType(t, &DefaultThemeResolverFactory{}, factory)
}

func TestDefaultThemeResolverFactory_Create(t *testing.T) {
	factory := NewThemeResolverFactory()
	dataStore := NewMockDataStore()
	themeStore := NewMockThemeStore()

	t.Run("create with default config", func(t *testing.T) {
		config := DefaultResolverConfig()
		resolver, err := factory.Create(dataStore, themeStore, config)

		require.NoError(t, err)
		assert.NotNil(t, resolver)
		assert.IsType(t, &HierarchyThemeResolver{}, resolver)

		// Verify resolver is properly initialized
		hierarchyResolver, ok := resolver.(*HierarchyThemeResolver)
		require.True(t, ok)
		assert.Equal(t, dataStore, hierarchyResolver.dataStore)
		assert.Equal(t, themeStore, hierarchyResolver.themeStore)
		assert.Equal(t, config.DefaultTheme, hierarchyResolver.defaultTheme)
	})

	t.Run("create with custom config", func(t *testing.T) {
		config := ResolverConfig{
			DefaultTheme: "custom-theme",
			CacheEnabled: true,
			CacheTTL:     600,
		}
		resolver, err := factory.Create(dataStore, themeStore, config)

		require.NoError(t, err)
		assert.NotNil(t, resolver)
		assert.IsType(t, &HierarchyThemeResolver{}, resolver)

		// Verify custom config is applied
		hierarchyResolver, ok := resolver.(*HierarchyThemeResolver)
		require.True(t, ok)
		assert.Equal(t, "custom-theme", hierarchyResolver.defaultTheme)
	})

	t.Run("create with nil datastore", func(t *testing.T) {
		config := DefaultResolverConfig()
		resolver, err := factory.Create(nil, themeStore, config)

		require.NoError(t, err)
		assert.NotNil(t, resolver)

		hierarchyResolver, ok := resolver.(*HierarchyThemeResolver)
		require.True(t, ok)
		assert.Nil(t, hierarchyResolver.dataStore)
	})

	t.Run("create with nil themestore", func(t *testing.T) {
		config := DefaultResolverConfig()
		resolver, err := factory.Create(dataStore, nil, config)

		require.NoError(t, err)
		assert.NotNil(t, resolver)

		hierarchyResolver, ok := resolver.(*HierarchyThemeResolver)
		require.True(t, ok)
		assert.Nil(t, hierarchyResolver.themeStore)
	})
}

func TestNewThemeResolver(t *testing.T) {
	dataStore := NewMockDataStore()
	themeStore := NewMockThemeStore()

	resolver, err := NewThemeResolver(dataStore, themeStore)

	require.NoError(t, err)
	assert.NotNil(t, resolver)
	assert.IsType(t, &HierarchyThemeResolver{}, resolver)

	// Verify it uses default config
	hierarchyResolver, ok := resolver.(*HierarchyThemeResolver)
	require.True(t, ok)
	assert.Equal(t, dataStore, hierarchyResolver.dataStore)
	assert.Equal(t, themeStore, hierarchyResolver.themeStore)
	assert.Equal(t, DefaultTheme, hierarchyResolver.defaultTheme)
}

func TestResolverConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config ResolverConfig
		valid  bool
	}{
		{
			name: "valid default config",
			config: ResolverConfig{
				DefaultTheme: DefaultTheme,
				CacheEnabled: false,
				CacheTTL:     300,
			},
			valid: true,
		},
		{
			name: "valid custom config",
			config: ResolverConfig{
				DefaultTheme: "custom-theme",
				CacheEnabled: true,
				CacheTTL:     600,
			},
			valid: true,
		},
		{
			name: "empty default theme",
			config: ResolverConfig{
				DefaultTheme: "",
				CacheEnabled: false,
				CacheTTL:     300,
			},
			valid: true, // Empty theme name is valid, will fallback to DefaultTheme
		},
		{
			name: "zero cache TTL",
			config: ResolverConfig{
				DefaultTheme: DefaultTheme,
				CacheEnabled: true,
				CacheTTL:     0,
			},
			valid: true, // Zero TTL is valid
		},
		{
			name: "negative cache TTL",
			config: ResolverConfig{
				DefaultTheme: DefaultTheme,
				CacheEnabled: true,
				CacheTTL:     -300,
			},
			valid: true, // Negative TTL is valid (infinite cache)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewThemeResolverFactory()
			dataStore := NewMockDataStore()
			themeStore := NewMockThemeStore()

			resolver, err := factory.Create(dataStore, themeStore, tt.config)

			if tt.valid {
				require.NoError(t, err)
				assert.NotNil(t, resolver)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// Test that the factory implements the interface correctly
func TestThemeResolverFactory_Interface(t *testing.T) {
	var factory ThemeResolverFactory = NewThemeResolverFactory()

	assert.NotNil(t, factory)

	// Should be able to call interface methods
	dataStore := NewMockDataStore()
	themeStore := NewMockThemeStore()
	config := DefaultResolverConfig()

	resolver, err := factory.Create(dataStore, themeStore, config)

	require.NoError(t, err)
	assert.NotNil(t, resolver)
}

func TestResolverConfig_Fields(t *testing.T) {
	config := ResolverConfig{
		DefaultTheme: "test-theme",
		CacheEnabled: true,
		CacheTTL:     1200,
	}

	// Test field access
	assert.Equal(t, "test-theme", config.DefaultTheme)
	assert.True(t, config.CacheEnabled)
	assert.Equal(t, 1200, config.CacheTTL)

	// Test field modification
	config.DefaultTheme = "modified-theme"
	config.CacheEnabled = false
	config.CacheTTL = 600

	assert.Equal(t, "modified-theme", config.DefaultTheme)
	assert.False(t, config.CacheEnabled)
	assert.Equal(t, 600, config.CacheTTL)
}
