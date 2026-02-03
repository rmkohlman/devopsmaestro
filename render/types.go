// Package render provides a decoupled rendering system for CLI output.
//
// The render package separates data preparation from display logic, allowing
// commands to be completely agnostic about output format. Commands prepare
// structured data and pass it to the renderer, which decides how to display it.
//
// Architecture:
//
//	Command Layer (prepares data)
//	     │
//	     ▼
//	Render Package (Output function)
//	     │
//	     ▼
//	Renderer Interface (JSON, YAML, Colored, Plain, etc.)
//	     │
//	     ▼
//	io.Writer (stdout, file, buffer)
//
// Configuration hierarchy (highest to lowest priority):
//  1. Command flag (-r/--render)
//  2. Environment variable (DVM_RENDER)
//  3. Config file (~/.config/dvm/config.yaml)
//  4. Default (colored)
//
// Example usage:
//
//	// In a command - just prepare data and call Output
//	data := map[string]string{"project": "foo", "workspace": "bar"}
//	return render.Output(data, render.Options{
//	    Type:  render.TypeKeyValue,
//	    Title: "Current Context",
//	})
//
//	// Override renderer via flag
//	dvm get context -r json
//	dvm get context -r yaml
package render

// RenderType hints to the renderer what kind of data structure this is.
// Renderers use this to decide how to format the output.
type RenderType string

const (
	// TypeAuto lets the renderer inspect the data and decide
	TypeAuto RenderType = "auto"

	// TypeKeyValue is for key-value pairs (maps, structs with few fields)
	TypeKeyValue RenderType = "keyvalue"

	// TypeTable is for tabular data (slices of structs/maps)
	TypeTable RenderType = "table"

	// TypeList is for simple lists (slices of strings)
	TypeList RenderType = "list"

	// TypeDetail is for detailed single-object view
	TypeDetail RenderType = "detail"

	// TypeRaw passes data through with minimal formatting
	TypeRaw RenderType = "raw"

	// TypeProgress is for progress/status messages
	TypeProgress RenderType = "progress"
)

// RendererName identifies a renderer implementation
type RendererName string

const (
	RendererJSON    RendererName = "json"
	RendererYAML    RendererName = "yaml"
	RendererColored RendererName = "colored"
	RendererPlain   RendererName = "plain"
	RendererTable   RendererName = "table"
	RendererCompact RendererName = "compact"
)

// Options configures how data should be rendered.
// Commands set these options to provide hints to the renderer.
type Options struct {
	// Type hints what kind of data structure this is
	Type RenderType

	// Title is a section title (used by human-readable renderers, ignored by JSON/YAML)
	Title string

	// Headers are column headers for table type
	Headers []string

	// Empty indicates the data represents an empty state
	Empty bool

	// EmptyMessage is shown when Empty is true (human-readable renderers)
	EmptyMessage string

	// EmptyHints are suggestions shown when Empty is true
	EmptyHints []string

	// Verbose enables extra detail in output
	Verbose bool
}

// MessageLevel indicates the severity/type of a message
type MessageLevel string

const (
	LevelInfo     MessageLevel = "info"
	LevelSuccess  MessageLevel = "success"
	LevelWarning  MessageLevel = "warning"
	LevelError    MessageLevel = "error"
	LevelDebug    MessageLevel = "debug"
	LevelProgress MessageLevel = "progress"
)

// Message represents a status message to be rendered
type Message struct {
	Level   MessageLevel
	Content string
}

// Config holds renderer configuration
type Config struct {
	// Default renderer name
	Default RendererName

	// Verbose enables verbose output
	Verbose bool

	// NoColor disables colored output (respects NO_COLOR env var)
	NoColor bool

	// UseNerdFonts enables Nerd Font icons
	UseNerdFonts bool
}

// DefaultConfig returns the default renderer configuration
func DefaultConfig() Config {
	return Config{
		Default:      RendererColored,
		Verbose:      false,
		NoColor:      false,
		UseNerdFonts: false,
	}
}
