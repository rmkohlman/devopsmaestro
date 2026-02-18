package colors

// NoColorProvider implements ColorProvider by returning empty strings for all colors.
// This effectively disables color output when used with rendering functions that
// check for empty color values. It's used when --no-color is specified or when
// the NO_COLOR environment variable is set.
type NoColorProvider struct{}

// NewNoColorProvider creates a ColorProvider that disables all colors.
// All color methods return empty strings, which should be interpreted
// by rendering code as "no color formatting".
func NewNoColorProvider() ColorProvider {
	return &NoColorProvider{}
}

// =============================================================================
// Primary Colors - all return empty strings
// =============================================================================

// Primary returns an empty string (no primary color).
func (n *NoColorProvider) Primary() string {
	return ""
}

// Secondary returns an empty string (no secondary color).
func (n *NoColorProvider) Secondary() string {
	return ""
}

// Accent returns an empty string (no accent color).
func (n *NoColorProvider) Accent() string {
	return ""
}

// =============================================================================
// Status Colors - all return empty strings
// =============================================================================

// Success returns an empty string (no success color).
func (n *NoColorProvider) Success() string {
	return ""
}

// Warning returns an empty string (no warning color).
func (n *NoColorProvider) Warning() string {
	return ""
}

// Error returns an empty string (no error color).
func (n *NoColorProvider) Error() string {
	return ""
}

// Info returns an empty string (no info color).
func (n *NoColorProvider) Info() string {
	return ""
}

// =============================================================================
// UI Colors - all return empty strings
// =============================================================================

// Foreground returns an empty string (no foreground color).
func (n *NoColorProvider) Foreground() string {
	return ""
}

// Background returns an empty string (no background color).
func (n *NoColorProvider) Background() string {
	return ""
}

// Muted returns an empty string (no muted color).
func (n *NoColorProvider) Muted() string {
	return ""
}

// Highlight returns an empty string (no highlight color).
func (n *NoColorProvider) Highlight() string {
	return ""
}

// Border returns an empty string (no border color).
func (n *NoColorProvider) Border() string {
	return ""
}

// =============================================================================
// Metadata
// =============================================================================

// Name returns the provider name.
func (n *NoColorProvider) Name() string {
	return "no-color"
}

// IsLight returns false since no colors are provided.
// This is somewhat arbitrary but false is a reasonable default.
func (n *NoColorProvider) IsLight() bool {
	return false
}
