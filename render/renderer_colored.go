package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Icons holds the icon set for colored output
type Icons struct {
	Success  string
	Warning  string
	Error    string
	Info     string
	Progress string
	Bullet   string
	Section  string
}

// DefaultIcons returns Unicode icons
func DefaultIcons() Icons {
	return Icons{
		Success:  "✓",
		Warning:  "⚠",
		Error:    "✗",
		Info:     "ℹ",
		Progress: "→",
		Bullet:   "•",
		Section:  "▌",
	}
}

// NerdFontIcons returns Nerd Font icons
func NerdFontIcons() Icons {
	return Icons{
		Success:  "\uf00c", // nf-fa-check
		Warning:  "\uf071", // nf-fa-exclamation_triangle
		Error:    "\uf00d", // nf-fa-times
		Info:     "\uf05a", // nf-fa-info_circle
		Progress: "\uf061", // nf-fa-arrow_right
		Bullet:   "\uf111", // nf-fa-circle
		Section:  "\ue0b0", // powerline
	}
}

// PlainIcons returns ASCII-only icons
func PlainIcons() Icons {
	return Icons{
		Success:  "[OK]",
		Warning:  "[!]",
		Error:    "[X]",
		Info:     "[i]",
		Progress: "->",
		Bullet:   "*",
		Section:  "|",
	}
}

// styles holds lipgloss styles for colored output
type styles struct {
	success   lipgloss.Style
	warning   lipgloss.Style
	errStyle  lipgloss.Style
	info      lipgloss.Style
	muted     lipgloss.Style
	header    lipgloss.Style
	title     lipgloss.Style
	key       lipgloss.Style
	value     lipgloss.Style
	highlight lipgloss.Style
}

func defaultStyles() styles {
	return styles{
		success:   lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1")),
		warning:   lipgloss.NewStyle().Foreground(lipgloss.Color("#F9E2AF")),
		errStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8")),
		info:      lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA")),
		muted:     lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")),
		header:    lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7")).Bold(true),
		title:     lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7")).Bold(true),
		key:       lipgloss.NewStyle().Foreground(lipgloss.Color("#89DCEB")),
		value:     lipgloss.NewStyle().Foreground(lipgloss.Color("#F5E0DC")),
		highlight: lipgloss.NewStyle().Foreground(lipgloss.Color("#FAB387")).Bold(true),
	}
}

// ColoredRenderer outputs richly formatted text with colors and icons.
// This is the default renderer for interactive terminal use.
type ColoredRenderer struct {
	icons  Icons
	styles styles
}

// NewColoredRenderer creates a new colored renderer
func NewColoredRenderer() *ColoredRenderer {
	return &ColoredRenderer{
		icons:  DefaultIcons(),
		styles: defaultStyles(),
	}
}

// NewColoredRendererWithIcons creates a colored renderer with custom icons
func NewColoredRendererWithIcons(icons Icons) *ColoredRenderer {
	return &ColoredRenderer{
		icons:  icons,
		styles: defaultStyles(),
	}
}

// Name returns the renderer identifier
func (r *ColoredRenderer) Name() RendererName {
	return RendererColored
}

// SupportsColor returns true - this renderer uses colors
func (r *ColoredRenderer) SupportsColor() bool {
	return true
}

// Render outputs data with colors and formatting
func (r *ColoredRenderer) Render(w io.Writer, data any, opts Options) error {
	// Handle empty state
	if opts.Empty {
		if opts.EmptyMessage != "" {
			r.RenderMessage(w, Message{Level: LevelInfo, Content: opts.EmptyMessage})
		}
		if len(opts.EmptyHints) > 0 {
			fmt.Fprintln(w)
			fmt.Fprintln(w, "Set context with:")
			for _, hint := range opts.EmptyHints {
				fmt.Fprintf(w, "  %s %s\n", r.styles.info.Render(r.icons.Bullet), hint)
			}
		}
		return nil
	}

	// Render title if present
	if opts.Title != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, r.styles.title.Render(r.icons.Section+" "+opts.Title))
		fmt.Fprintln(w)
	}

	// Render based on type hint or data type
	switch v := data.(type) {
	case KeyValueData:
		return r.renderKeyValue(w, v)
	case TableData:
		return r.renderTable(w, v)
	case ListData:
		return r.renderList(w, v)
	case []string:
		return r.renderList(w, ListData{Items: v})
	case map[string]string:
		kv := NewKeyValueData(v)
		return r.renderKeyValue(w, kv)
	default:
		// For other types, just print
		fmt.Fprintf(w, "%v\n", data)
	}

	return nil
}

func (r *ColoredRenderer) renderKeyValue(w io.Writer, kv KeyValueData) error {
	for _, pair := range kv.Pairs {
		key := r.styles.key.Render(pair.Key + ":")
		value := r.styles.value.Render(pair.Value)
		fmt.Fprintf(w, "%s %s\n", key, value)
	}
	return nil
}

func (r *ColoredRenderer) renderTable(w io.Writer, t TableData) error {
	if len(t.Rows) == 0 {
		fmt.Fprintln(w, r.styles.muted.Render("No data"))
		return nil
	}

	// Calculate column widths
	widths := make([]int, len(t.Headers))
	for i, h := range t.Headers {
		widths[i] = len(h)
	}
	for _, row := range t.Rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print headers
	var headerParts []string
	for i, h := range t.Headers {
		headerParts = append(headerParts, r.styles.header.Render(fmt.Sprintf("%-*s", widths[i], h)))
	}
	fmt.Fprintln(w, strings.Join(headerParts, "   "))

	// Print separator
	var sepParts []string
	for _, width := range widths {
		sepParts = append(sepParts, r.styles.muted.Render(strings.Repeat("─", width)))
	}
	fmt.Fprintln(w, strings.Join(sepParts, "   "))

	// Print rows
	for _, row := range t.Rows {
		var cellParts []string
		for i, cell := range row {
			if i < len(widths) {
				cellParts = append(cellParts, fmt.Sprintf("%-*s", widths[i], cell))
			}
		}
		fmt.Fprintln(w, strings.Join(cellParts, "   "))
	}

	return nil
}

func (r *ColoredRenderer) renderList(w io.Writer, list ListData) error {
	for _, item := range list.Items {
		fmt.Fprintf(w, "  %s %s\n", r.styles.info.Render(r.icons.Bullet), item)
	}
	return nil
}

// RenderMessage outputs a styled message
func (r *ColoredRenderer) RenderMessage(w io.Writer, msg Message) error {
	var icon string
	var style lipgloss.Style

	switch msg.Level {
	case LevelSuccess:
		icon = r.icons.Success
		style = r.styles.success
	case LevelWarning:
		icon = r.icons.Warning
		style = r.styles.warning
	case LevelError:
		icon = r.icons.Error
		style = r.styles.errStyle
	case LevelProgress:
		icon = r.icons.Progress
		style = r.styles.info
	case LevelDebug:
		icon = ""
		style = r.styles.muted
	case LevelInfo:
		fallthrough
	default:
		icon = r.icons.Info
		style = r.styles.info
	}

	if icon != "" {
		fmt.Fprintln(w, style.Render(icon+" "+msg.Content))
	} else {
		fmt.Fprintln(w, style.Render(msg.Content))
	}
	return nil
}

func init() {
	Register(NewColoredRenderer())
}
