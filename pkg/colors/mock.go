package colors

// MockColorProvider provides a mock implementation for testing.
// All colors are configurable and default to dark theme colors.
type MockColorProvider struct {
	name    string
	isLight bool
	colors  map[string]string
}

// MockOption is a functional option for configuring MockColorProvider.
type MockOption func(*MockColorProvider)

// NewMockColorProvider creates a mock provider with default dark colors.
// Options can be used to customize the mock.
//
// Example usage:
//
//	mock := colors.NewMockColorProvider()
//	mock := colors.NewMockColorProvider(colors.WithMockName("test-theme"))
//	mock := colors.NewMockColorProvider(colors.WithMockLight())
func NewMockColorProvider(opts ...MockOption) *MockColorProvider {
	m := &MockColorProvider{
		name:    "mock-theme",
		isLight: false,
		colors:  cloneColorMap(DefaultDarkColors),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// WithMockName sets the theme name.
func WithMockName(name string) MockOption {
	return func(m *MockColorProvider) {
		m.name = name
	}
}

// WithMockLight configures the mock as a light theme with light defaults.
func WithMockLight() MockOption {
	return func(m *MockColorProvider) {
		m.isLight = true
		m.colors = cloneColorMap(DefaultLightColors)
	}
}

// WithMockColor overrides a specific color.
func WithMockColor(key, value string) MockOption {
	return func(m *MockColorProvider) {
		m.colors[key] = value
	}
}

// WithMockColors overrides multiple colors.
func WithMockColors(colors map[string]string) MockOption {
	return func(m *MockColorProvider) {
		for k, v := range colors {
			m.colors[k] = v
		}
	}
}

// cloneColorMap creates a copy of a color map.
func cloneColorMap(m map[string]string) map[string]string {
	clone := make(map[string]string, len(m))
	for k, v := range m {
		clone[k] = v
	}
	return clone
}

// =============================================================================
// ColorProvider Interface Implementation
// =============================================================================

// Primary returns the primary color.
func (m *MockColorProvider) Primary() string { return m.colors["primary"] }

// Secondary returns the secondary color.
func (m *MockColorProvider) Secondary() string { return m.colors["secondary"] }

// Accent returns the accent color.
func (m *MockColorProvider) Accent() string { return m.colors["accent"] }

// Success returns the success color.
func (m *MockColorProvider) Success() string { return m.colors["success"] }

// Warning returns the warning color.
func (m *MockColorProvider) Warning() string { return m.colors["warning"] }

// Error returns the error color.
func (m *MockColorProvider) Error() string { return m.colors["error"] }

// Info returns the info color.
func (m *MockColorProvider) Info() string { return m.colors["info"] }

// Foreground returns the foreground color.
func (m *MockColorProvider) Foreground() string { return m.colors["foreground"] }

// Background returns the background color.
func (m *MockColorProvider) Background() string { return m.colors["background"] }

// Muted returns the muted color.
func (m *MockColorProvider) Muted() string { return m.colors["muted"] }

// Highlight returns the highlight color.
func (m *MockColorProvider) Highlight() string { return m.colors["highlight"] }

// Border returns the border color.
func (m *MockColorProvider) Border() string { return m.colors["border"] }

// Name returns the theme name.
func (m *MockColorProvider) Name() string { return m.name }

// IsLight returns whether this is a light theme.
func (m *MockColorProvider) IsLight() bool { return m.isLight }

// =============================================================================
// Test Helpers
// =============================================================================

// SetColor allows changing a color in tests (mutation).
// This is useful for testing specific color scenarios.
func (m *MockColorProvider) SetColor(key, value string) {
	m.colors[key] = value
}

// GetAllColors returns all colors for inspection in tests.
func (m *MockColorProvider) GetAllColors() map[string]string {
	return cloneColorMap(m.colors)
}
