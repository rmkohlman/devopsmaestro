// Package resolver provides factories for creating theme resolvers
package resolver

import (
	"devopsmaestro/db"
	"devopsmaestro/pkg/nvimops/theme"
)

// ResolverConfig contains configuration for theme resolvers
type ResolverConfig struct {
	DefaultTheme string `mapstructure:"default_theme"`
	CacheEnabled bool   `mapstructure:"cache_enabled"`
	CacheTTL     int    `mapstructure:"cache_ttl_seconds"` // seconds
}

// DefaultResolverConfig returns the default configuration
func DefaultResolverConfig() ResolverConfig {
	return ResolverConfig{
		DefaultTheme: DefaultTheme,
		CacheEnabled: false, // Start without caching
		CacheTTL:     300,   // 5 minutes
	}
}

// ThemeResolverFactory creates theme resolvers
type ThemeResolverFactory interface {
	Create(dataStore db.DataStore, themeStore theme.Store, config ResolverConfig) (ThemeResolver, error)
}

// DefaultThemeResolverFactory implements the factory interface
type DefaultThemeResolverFactory struct{}

// NewThemeResolverFactory returns a new factory instance
func NewThemeResolverFactory() ThemeResolverFactory {
	return &DefaultThemeResolverFactory{}
}

// Create creates a new theme resolver with the given dependencies
func (f *DefaultThemeResolverFactory) Create(dataStore db.DataStore, themeStore theme.Store, config ResolverConfig) (ThemeResolver, error) {
	resolver := &HierarchyThemeResolver{
		dataStore:    dataStore,
		themeStore:   themeStore,
		defaultTheme: config.DefaultTheme,
	}

	// TODO: Add caching layer if config.CacheEnabled is true
	// if config.CacheEnabled {
	//     resolver = &CachedThemeResolver{
	//         resolver: resolver,
	//         ttl:      time.Duration(config.CacheTTL) * time.Second,
	//     }
	// }

	return resolver, nil
}

// NewThemeResolver creates a theme resolver with default configuration
// This is a convenience function for common use cases
func NewThemeResolver(dataStore db.DataStore, themeStore theme.Store) (ThemeResolver, error) {
	factory := NewThemeResolverFactory()
	config := DefaultResolverConfig()
	return factory.Create(dataStore, themeStore, config)
}
