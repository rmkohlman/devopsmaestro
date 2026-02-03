package render

import (
	"fmt"
	"io"
	"strings"
)

// PlainRenderer outputs plain text without colors.
// Suitable for piping, CI/CD environments, and terminals without color support.
type PlainRenderer struct{}

// NewPlainRenderer creates a new plain text renderer
func NewPlainRenderer() *PlainRenderer {
	return &PlainRenderer{}
}

// Name returns the renderer identifier
func (r *PlainRenderer) Name() RendererName {
	return RendererPlain
}

// SupportsColor returns false - plain doesn't use colors
func (r *PlainRenderer) SupportsColor() bool {
	return false
}

// Render outputs data as plain text
func (r *PlainRenderer) Render(w io.Writer, data any, opts Options) error {
	// Handle empty state
	if opts.Empty {
		if opts.EmptyMessage != "" {
			fmt.Fprintln(w, opts.EmptyMessage)
		}
		if len(opts.EmptyHints) > 0 {
			fmt.Fprintln(w)
			for _, hint := range opts.EmptyHints {
				fmt.Fprintf(w, "  * %s\n", hint)
			}
		}
		return nil
	}

	// Render title if present
	if opts.Title != "" {
		fmt.Fprintf(w, "== %s ==\n", opts.Title)
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
		return r.renderMap(w, v)
	default:
		// For other types, try to render as key-value if it's a struct
		fmt.Fprintf(w, "%v\n", data)
	}

	return nil
}

func (r *PlainRenderer) renderKeyValue(w io.Writer, kv KeyValueData) error {
	// Find max key length for alignment
	maxLen := 0
	for _, pair := range kv.Pairs {
		if len(pair.Key) > maxLen {
			maxLen = len(pair.Key)
		}
	}

	for _, pair := range kv.Pairs {
		fmt.Fprintf(w, "%-*s: %s\n", maxLen, pair.Key, pair.Value)
	}
	return nil
}

func (r *PlainRenderer) renderMap(w io.Writer, m map[string]string) error {
	for k, v := range m {
		fmt.Fprintf(w, "%s: %s\n", k, v)
	}
	return nil
}

func (r *PlainRenderer) renderTable(w io.Writer, t TableData) error {
	if len(t.Rows) == 0 {
		fmt.Fprintln(w, "No data")
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
	for i, h := range t.Headers {
		fmt.Fprintf(w, "%-*s  ", widths[i], h)
	}
	fmt.Fprintln(w)

	// Print separator
	for i := range t.Headers {
		fmt.Fprint(w, strings.Repeat("-", widths[i])+"  ")
	}
	fmt.Fprintln(w)

	// Print rows
	for _, row := range t.Rows {
		for i, cell := range row {
			if i < len(widths) {
				fmt.Fprintf(w, "%-*s  ", widths[i], cell)
			}
		}
		fmt.Fprintln(w)
	}

	return nil
}

func (r *PlainRenderer) renderList(w io.Writer, list ListData) error {
	for _, item := range list.Items {
		fmt.Fprintf(w, "  * %s\n", item)
	}
	return nil
}

// RenderMessage outputs a message as plain text
func (r *PlainRenderer) RenderMessage(w io.Writer, msg Message) error {
	prefix := ""
	switch msg.Level {
	case LevelSuccess:
		prefix = "[OK] "
	case LevelWarning:
		prefix = "[WARN] "
	case LevelError:
		prefix = "[ERROR] "
	case LevelDebug:
		prefix = "[DEBUG] "
	case LevelProgress:
		prefix = "-> "
	case LevelInfo:
		prefix = "[INFO] "
	}
	fmt.Fprintf(w, "%s%s\n", prefix, msg.Content)
	return nil
}

func init() {
	Register(NewPlainRenderer())
}
