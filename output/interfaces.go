// Package output provides a decoupled output formatting system for CLI applications.
//
// Architecture:
//   - Formatter interface defines the contract for all output formatters
//   - Theme interface allows swappable color/style themes
//   - Factory function creates formatters based on configuration
//
// Output Styles:
//   - PlainFormatter: Minimal, no colors, suitable for piping
//   - TableFormatter: Structured tables with alignment
//   - ColoredFormatter: Rich colors and icons (default)
//   - JSONFormatter: Machine-readable JSON output
//   - YAMLFormatter: Human-readable structured output
//
// Example usage:
//
//	formatter := output.NewFormatter(output.FormatterConfig{
//	    Style: output.StyleColored,
//	    Theme: output.ThemeCatppuccin,
//	})
//	formatter.Success("Build complete!")
//	formatter.Table(headers, rows)
package output

import (
	"io"
)

// Style identifies the output formatting style
type Style string

const (
	// StylePlain - minimal output, no colors, no icons
	StylePlain Style = "plain"

	// StyleTable - structured tables with alignment, minimal decoration
	StyleTable Style = "table"

	// StyleColored - rich colors, icons, and formatting (default)
	StyleColored Style = "colored"

	// StyleCompact - colored but condensed for smaller terminals
	StyleCompact Style = "compact"

	// StyleVerbose - extra details, timestamps, full paths
	StyleVerbose Style = "verbose"

	// StyleJSON - machine-readable JSON
	StyleJSON Style = "json"

	// StyleYAML - human-readable structured YAML
	StyleYAML Style = "yaml"
)

// Formatter is the main interface for all output formatting.
// Implementations handle different output styles while maintaining
// a consistent API for the application.
type Formatter interface {
	// Message Output
	// These methods output various types of messages

	// Info outputs an informational message
	Info(message string)

	// Success outputs a success message
	Success(message string)

	// Warning outputs a warning message
	Warning(message string)

	// Error outputs an error message
	Error(message string)

	// Debug outputs a debug message (only shown in verbose mode)
	Debug(message string)

	// Progress outputs a progress/step message
	Progress(message string)

	// Step outputs a numbered step message
	Step(number int, total int, message string)

	// Structured Output
	// These methods output structured data

	// Table outputs tabular data with headers and rows
	Table(headers []string, rows [][]string)

	// List outputs a list of items
	List(items []string)

	// KeyValue outputs key-value pairs
	KeyValue(pairs map[string]string)

	// Object outputs a structured object (for JSON/YAML)
	Object(v interface{}) error

	// Sections and Grouping
	// These methods help organize output

	// Section starts a new section with a header
	Section(title string)

	// Subsection starts a subsection
	Subsection(title string)

	// Separator outputs a visual separator
	Separator()

	// NewLine outputs a blank line
	NewLine()

	// Raw Output
	// For cases where direct output is needed

	// Print outputs raw text
	Print(text string)

	// Printf outputs formatted text
	Printf(format string, args ...interface{})

	// Println outputs text with newline
	Println(text string)

	// Configuration

	// SetWriter sets the output writer (default: os.Stdout)
	SetWriter(w io.Writer)

	// SetVerbose enables/disables verbose output
	SetVerbose(verbose bool)

	// GetStyle returns the formatter's style
	GetStyle() Style
}

// TableOptions provides options for table rendering
type TableOptions struct {
	// ShowHeaders controls whether headers are displayed
	ShowHeaders bool

	// ShowBorders controls whether table borders are shown
	ShowBorders bool

	// ShowRowNumbers adds row numbers
	ShowRowNumbers bool

	// MaxWidth limits the table width (0 = unlimited)
	MaxWidth int

	// ColumnWidths specifies fixed column widths (nil = auto)
	ColumnWidths []int

	// HighlightColumn highlights a specific column
	HighlightColumn int

	// ActiveRow marks a row as "active" with special styling
	ActiveRow int
}

// ProgressOptions provides options for progress display
type ProgressOptions struct {
	// ShowSpinner shows an animated spinner
	ShowSpinner bool

	// ShowPercentage shows completion percentage
	ShowPercentage bool

	// ShowElapsed shows elapsed time
	ShowElapsed bool

	// Total is the total number of items (0 = indeterminate)
	Total int

	// Current is the current progress
	Current int
}

// ListOptions provides options for list rendering
type ListOptions struct {
	// Numbered uses numbered list instead of bullets
	Numbered bool

	// Indent level for nested lists
	Indent int

	// BulletStyle specifies the bullet character
	BulletStyle string
}
