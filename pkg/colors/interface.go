// Package colors provides a decoupled interface for accessing theme colors.
// This package defines the ColorProvider interface that consumers use instead of
// importing theme internals directly. This allows the render package and other
// consumers to remain decoupled from the theme implementation.
//
// Dependency Flow:
//
//	cmd/ → injects ColorProvider via context
//	render/ → uses ColorProvider interface (no theme import)
//	pkg/colors/ → defines interface, implements via palette
//	pkg/palette/ → pure data model
//	pkg/nvimops/theme/ → manages themes, creates palettes
package colors

// ColorProvider defines the interface for accessing theme colors.
// This is the decoupled interface that consumers use instead of importing theme internals.
// All color values are returned as hex strings (e.g., "#7aa2f7").
type ColorProvider interface {
	// Primary colors - main theme colors
	Primary() string   // Main brand/accent color
	Secondary() string // Secondary brand color
	Accent() string    // Highlight/focus color

	// Status colors - semantic meaning
	Success() string // Green-ish success state
	Warning() string // Yellow-ish warning state
	Error() string   // Red-ish error state
	Info() string    // Blue-ish info state

	// UI colors - interface elements
	Foreground() string // Main text color
	Background() string // Main background color
	Muted() string      // Subdued/disabled text
	Highlight() string  // Selection/hover background
	Border() string     // Border/separator color

	// Theme metadata
	Name() string  // Theme name (e.g., "tokyonight-night")
	IsLight() bool // Whether this is a light theme
}
