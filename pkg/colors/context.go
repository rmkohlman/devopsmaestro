package colors

import "context"

// contextKey is the type for context keys in this package.
type contextKey string

// colorProviderKey is the context key for the ColorProvider.
const colorProviderKey contextKey = "devopsmaestro.colorProvider"

// WithProvider injects a ColorProvider into the context.
// This should be called early in the request lifecycle (e.g., in cmd/).
//
// Example usage:
//
//	ctx := colors.WithProvider(context.Background(), provider)
//	// Pass ctx to downstream functions
func WithProvider(ctx context.Context, provider ColorProvider) context.Context {
	return context.WithValue(ctx, colorProviderKey, provider)
}

// FromContext retrieves the ColorProvider from context.
// Returns the provider and true if found, or nil and false if not present.
//
// Example usage:
//
//	if provider, ok := colors.FromContext(ctx); ok {
//	    statusColor := provider.Success()
//	}
func FromContext(ctx context.Context) (ColorProvider, bool) {
	provider, ok := ctx.Value(colorProviderKey).(ColorProvider)
	return provider, ok
}

// MustFromContext retrieves the ColorProvider from context.
// Panics if the provider is not found.
//
// Use this when you're certain the provider has been injected upstream.
// For safer retrieval, use FromContext or FromContextOrDefault instead.
func MustFromContext(ctx context.Context) ColorProvider {
	provider, ok := FromContext(ctx)
	if !ok {
		panic("colors: ColorProvider not found in context - did you call WithProvider?")
	}
	return provider
}

// FromContextOrDefault retrieves the ColorProvider from context,
// returning a default provider if not found.
//
// This is the safest way to get a ColorProvider when you're not certain
// if one has been injected into the context.
func FromContextOrDefault(ctx context.Context) ColorProvider {
	provider, ok := FromContext(ctx)
	if !ok {
		return NewDefaultColorProvider()
	}
	return provider
}

// FromContextOrDefaultLight retrieves the ColorProvider from context,
// returning a default light provider if not found.
func FromContextOrDefaultLight(ctx context.Context) ColorProvider {
	provider, ok := FromContext(ctx)
	if !ok {
		return NewDefaultLightColorProvider()
	}
	return provider
}

// HasProvider checks if a ColorProvider exists in the context.
func HasProvider(ctx context.Context) bool {
	_, ok := FromContext(ctx)
	return ok
}
