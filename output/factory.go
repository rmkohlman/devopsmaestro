package output

import (
	"io"
	"os"
)

// FormatterConfig provides configuration for creating a Formatter
type FormatterConfig struct {
	// Style determines the output formatting style
	Style Style

	// Theme determines the color theme (only for colored styles)
	Theme ThemeName

	// Writer is the output destination (default: os.Stdout)
	Writer io.Writer

	// Verbose enables verbose/debug output
	Verbose bool

	// UseNerdFonts enables Nerd Font icons (requires Nerd Font)
	UseNerdFonts bool

	// NoColor disables all color output (respects NO_COLOR env var)
	NoColor bool

	// Width specifies the terminal width (0 = auto-detect)
	Width int
}

// DefaultConfig returns the default formatter configuration
func DefaultConfig() FormatterConfig {
	return FormatterConfig{
		Style:        StyleColored,
		Theme:        ThemeAuto,
		Writer:       os.Stdout,
		Verbose:      false,
		UseNerdFonts: false,
		NoColor:      os.Getenv("NO_COLOR") != "",
		Width:        0,
	}
}

// NewFormatter creates a Formatter based on the provided configuration.
// This is the factory function that decouples consumers from specific implementations.
func NewFormatter(cfg FormatterConfig) Formatter {
	// Apply defaults
	if cfg.Writer == nil {
		cfg.Writer = os.Stdout
	}

	// Honor NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		cfg.NoColor = true
	}

	// Select icons
	var icons Icons
	if cfg.NoColor {
		icons = PlainIcons()
	} else if cfg.UseNerdFonts {
		icons = NerdFontIcons()
	} else {
		icons = DefaultIcons()
	}

	// Force plain style if NoColor is set
	if cfg.NoColor && cfg.Style != StyleJSON && cfg.Style != StyleYAML {
		cfg.Style = StylePlain
	}

	// Create the appropriate formatter
	switch cfg.Style {
	case StylePlain:
		return NewPlainFormatter(cfg, icons)
	case StyleJSON:
		return NewJSONFormatter(cfg)
	case StyleYAML:
		return NewYAMLFormatter(cfg)
	case StyleCompact:
		return NewCompactFormatter(cfg, icons)
	case StyleVerbose:
		return NewVerboseFormatter(cfg, icons)
	case StyleTable:
		return NewTableOnlyFormatter(cfg, icons)
	case StyleColored:
		fallthrough
	default:
		return NewColoredFormatter(cfg, icons)
	}
}

// Quick constructors for common use cases

// Plain returns a plain text formatter (no colors, no icons)
func Plain() Formatter {
	return NewFormatter(FormatterConfig{Style: StylePlain})
}

// Colored returns the default colored formatter
func Colored() Formatter {
	return NewFormatter(FormatterConfig{Style: StyleColored})
}

// JSON returns a JSON formatter
func JSON() Formatter {
	return NewFormatter(FormatterConfig{Style: StyleJSON})
}

// YAML returns a YAML formatter
func YAML() Formatter {
	return NewFormatter(FormatterConfig{Style: StyleYAML})
}

// Verbose returns a verbose formatter with extra details
func Verbose() Formatter {
	cfg := DefaultConfig()
	cfg.Style = StyleVerbose
	cfg.Verbose = true
	return NewFormatter(cfg)
}

// WithWriter returns a formatter writing to the specified writer
func WithWriter(w io.Writer, style Style) Formatter {
	return NewFormatter(FormatterConfig{
		Style:  style,
		Writer: w,
	})
}

// ForOutput returns a formatter for the specified output format string
// Useful for parsing -o/--output flags
func ForOutput(format string) Formatter {
	switch format {
	case "json":
		return JSON()
	case "yaml":
		return YAML()
	case "plain", "text":
		return Plain()
	case "table":
		return NewFormatter(FormatterConfig{Style: StyleTable})
	case "verbose":
		return Verbose()
	case "compact":
		return NewFormatter(FormatterConfig{Style: StyleCompact})
	default:
		return Colored()
	}
}
