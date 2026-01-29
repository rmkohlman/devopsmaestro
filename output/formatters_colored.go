package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ColoredFormatter outputs richly formatted text with colors and icons
type ColoredFormatter struct {
	baseFormatter
	styles coloredStyles
}

type coloredStyles struct {
	success   lipgloss.Style
	warning   lipgloss.Style
	error     lipgloss.Style
	info      lipgloss.Style
	muted     lipgloss.Style
	header    lipgloss.Style
	title     lipgloss.Style
	key       lipgloss.Style
	value     lipgloss.Style
	code      lipgloss.Style
	path      lipgloss.Style
	highlight lipgloss.Style
}

func defaultColoredStyles() coloredStyles {
	return coloredStyles{
		success:   lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1")),
		warning:   lipgloss.NewStyle().Foreground(lipgloss.Color("#F9E2AF")),
		error:     lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8")),
		info:      lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA")),
		muted:     lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")),
		header:    lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7")).Bold(true),
		title:     lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7")).Bold(true),
		key:       lipgloss.NewStyle().Foreground(lipgloss.Color("#89DCEB")),
		value:     lipgloss.NewStyle().Foreground(lipgloss.Color("#F5E0DC")),
		code:      lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1")).Background(lipgloss.Color("#313244")).Padding(0, 1),
		path:      lipgloss.NewStyle().Foreground(lipgloss.Color("#89DCEB")).Italic(true),
		highlight: lipgloss.NewStyle().Foreground(lipgloss.Color("#FAB387")).Bold(true),
	}
}

// NewColoredFormatter creates a new colored formatter
func NewColoredFormatter(cfg FormatterConfig, icons Icons) *ColoredFormatter {
	return &ColoredFormatter{
		baseFormatter: newBaseFormatter(cfg, icons),
		styles:        defaultColoredStyles(),
	}
}

func (f *ColoredFormatter) Info(message string) {
	f.writeln(f.styles.info.Render(f.icons.Info + " " + message))
}

func (f *ColoredFormatter) Success(message string) {
	f.writeln(f.styles.success.Render(f.icons.Success + " " + message))
}

func (f *ColoredFormatter) Warning(message string) {
	f.writeln(f.styles.warning.Render(f.icons.Warning + " " + message))
}

func (f *ColoredFormatter) Error(message string) {
	f.writeln(f.styles.error.Render(f.icons.Error + " " + message))
}

func (f *ColoredFormatter) Debug(message string) {
	if f.verbose {
		f.writeln(f.styles.muted.Render("[DEBUG] " + message))
	}
}

func (f *ColoredFormatter) Progress(message string) {
	f.writeln(f.styles.info.Render(f.icons.Arrow + " " + message))
}

func (f *ColoredFormatter) Step(num, total int, message string) {
	step := f.styles.muted.Render(fmt.Sprintf("[%d/%d]", num, total))
	f.writeln(step + " " + message)
}

func (f *ColoredFormatter) Table(headers []string, rows [][]string) {
	if len(rows) == 0 {
		f.writeln(f.styles.muted.Render("No data"))
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print headers
	var headerParts []string
	for i, h := range headers {
		headerParts = append(headerParts, f.styles.header.Render(fmt.Sprintf("%-*s", widths[i], h)))
	}
	f.writeln(strings.Join(headerParts, "   "))

	// Print separator
	var sepParts []string
	for _, w := range widths {
		sepParts = append(sepParts, f.styles.muted.Render(strings.Repeat("─", w)))
	}
	f.writeln(strings.Join(sepParts, "   "))

	// Print rows
	for _, row := range rows {
		var cellParts []string
		for i, cell := range row {
			if i < len(widths) {
				cellParts = append(cellParts, fmt.Sprintf("%-*s", widths[i], cell))
			}
		}
		f.writeln(strings.Join(cellParts, "   "))
	}
}

func (f *ColoredFormatter) List(items []string) {
	for _, item := range items {
		f.writeln("  " + f.styles.info.Render(f.icons.Bullet) + " " + item)
	}
}

func (f *ColoredFormatter) KeyValue(pairs map[string]string) {
	for k, v := range pairs {
		key := f.styles.key.Render(k + ":")
		value := f.styles.value.Render(v)
		f.writeln(key + " " + value)
	}
}

func (f *ColoredFormatter) Object(v interface{}) error {
	// For colored output, use the YAML formatter but with syntax highlighting
	yf := NewYAMLFormatter(f.config)
	return yf.Object(v)
}

func (f *ColoredFormatter) Section(title string) {
	f.writeln("")
	f.writeln(f.styles.title.Render("▌ " + title))
	f.writeln("")
}

func (f *ColoredFormatter) Subsection(title string) {
	f.writeln(f.styles.muted.Render("  ┌─ ") + f.styles.header.Render(title))
}

func (f *ColoredFormatter) Separator() {
	f.writeln(f.styles.muted.Render(strings.Repeat("─", 50)))
}

func (f *ColoredFormatter) NewLine()                                  { f.writeln("") }
func (f *ColoredFormatter) Print(text string)                         { f.write(text) }
func (f *ColoredFormatter) Printf(format string, args ...interface{}) { f.writef(format, args...) }
func (f *ColoredFormatter) Println(text string)                       { f.writeln(text) }

// CompactFormatter is like ColoredFormatter but more condensed
type CompactFormatter struct {
	ColoredFormatter
}

// NewCompactFormatter creates a new compact formatter
func NewCompactFormatter(cfg FormatterConfig, icons Icons) *CompactFormatter {
	return &CompactFormatter{
		ColoredFormatter: *NewColoredFormatter(cfg, icons),
	}
}

func (f *CompactFormatter) Section(title string) {
	f.writeln(f.styles.title.Render("▸ " + title))
}

func (f *CompactFormatter) Subsection(title string) {
	f.writeln(f.styles.muted.Render("  " + title + ":"))
}

func (f *CompactFormatter) Table(headers []string, rows [][]string) {
	// Compact table: no separator line, tighter spacing
	if len(rows) == 0 {
		f.writeln(f.styles.muted.Render("(empty)"))
		return
	}

	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Headers
	var headerParts []string
	for i, h := range headers {
		headerParts = append(headerParts, f.styles.muted.Render(fmt.Sprintf("%-*s", widths[i], h)))
	}
	f.writeln(strings.Join(headerParts, " "))

	// Rows
	for _, row := range rows {
		var cellParts []string
		for i, cell := range row {
			if i < len(widths) {
				cellParts = append(cellParts, fmt.Sprintf("%-*s", widths[i], cell))
			}
		}
		f.writeln(strings.Join(cellParts, " "))
	}
}

// VerboseFormatter adds extra details like timestamps
type VerboseFormatter struct {
	ColoredFormatter
}

// NewVerboseFormatter creates a new verbose formatter
func NewVerboseFormatter(cfg FormatterConfig, icons Icons) *VerboseFormatter {
	cfg.Verbose = true
	return &VerboseFormatter{
		ColoredFormatter: *NewColoredFormatter(cfg, icons),
	}
}

func (f *VerboseFormatter) timestamp() string {
	return f.styles.muted.Render(time.Now().Format("15:04:05") + " ")
}

func (f *VerboseFormatter) Info(message string) {
	f.writeln(f.timestamp() + f.styles.info.Render(f.icons.Info+" "+message))
}

func (f *VerboseFormatter) Success(message string) {
	f.writeln(f.timestamp() + f.styles.success.Render(f.icons.Success+" "+message))
}

func (f *VerboseFormatter) Warning(message string) {
	f.writeln(f.timestamp() + f.styles.warning.Render(f.icons.Warning+" "+message))
}

func (f *VerboseFormatter) Error(message string) {
	f.writeln(f.timestamp() + f.styles.error.Render(f.icons.Error+" "+message))
}

func (f *VerboseFormatter) Debug(message string) {
	// Always show debug in verbose mode
	f.writeln(f.timestamp() + f.styles.muted.Render("[DEBUG] "+message))
}

func (f *VerboseFormatter) Progress(message string) {
	f.writeln(f.timestamp() + f.styles.info.Render(f.icons.Arrow+" "+message))
}

// TableOnlyFormatter focuses on clean table output
type TableOnlyFormatter struct {
	ColoredFormatter
}

// NewTableOnlyFormatter creates a formatter optimized for tabular data
func NewTableOnlyFormatter(cfg FormatterConfig, icons Icons) *TableOnlyFormatter {
	return &TableOnlyFormatter{
		ColoredFormatter: *NewColoredFormatter(cfg, icons),
	}
}

func (f *TableOnlyFormatter) Info(message string)                 {} // Suppress non-table output
func (f *TableOnlyFormatter) Success(message string)              {}
func (f *TableOnlyFormatter) Warning(message string)              {}
func (f *TableOnlyFormatter) Error(message string)                {}
func (f *TableOnlyFormatter) Debug(message string)                {}
func (f *TableOnlyFormatter) Progress(message string)             {}
func (f *TableOnlyFormatter) Step(num, total int, message string) {}
func (f *TableOnlyFormatter) Section(title string)                {}
func (f *TableOnlyFormatter) Subsection(title string)             {}
func (f *TableOnlyFormatter) Separator()                          {}
func (f *TableOnlyFormatter) NewLine()                            {}

// Table is kept from ColoredFormatter - it's the main output method
