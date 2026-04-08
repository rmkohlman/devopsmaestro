// Package resolver provides factories for creating theme resolvers
package resolver

import (
	"devopsmaestro/db"
	theme "github.com/rmkohlman/MaestroTheme"
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

// ThemeResolverFactory creates ThemeResolver instances with injected dependencies.
// It follows the Interface → Implementation → Factory pattern from STANDARDS.md,
// allowing callers to swap resolver implementations without changing construction code.
//
// The default implementation is DefaultThemeResolverFactory, returned by
// NewThemeResolverFactory. For most callers, NewThemeResolver is sufficient.
//
// Example:
//
//	factory := resolver.NewThemeResolverFactory()
//	r, err := factory.Create(dataStore, themeStore, resolver.DefaultResolverConfig())
type ThemeResolverFactory interface {
	// Create constructs a ThemeResolver backed by the given data and theme stores.
	//
	// dataStore provides hierarchy entity lookups (app, domain, ecosystem, workspace).
	// themeStore provides theme loading by name (e.g., via MaestroTheme).
	// config controls resolver behaviour such as default theme name and cache settings.
	//
	// Returns an error if the resolver cannot be initialized with the provided config.
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
