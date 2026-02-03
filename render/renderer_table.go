package render

import (
	"fmt"
	"io"
	"strings"
)

// TableRenderer focuses on clean table output.
// It suppresses messages and other decorations, outputting only tables.
// Useful for piping to other tools.
type TableRenderer struct {
	colored *ColoredRenderer
}

// NewTableRenderer creates a new table-focused renderer
func NewTableRenderer() *TableRenderer {
	return &TableRenderer{
		colored: NewColoredRenderer(),
	}
}

// Name returns the renderer identifier
func (r *TableRenderer) Name() RendererName {
	return RendererTable
}

// SupportsColor returns true - uses colored table output
func (r *TableRenderer) SupportsColor() bool {
	return true
}

// Render outputs only table data, suppressing other elements
func (r *TableRenderer) Render(w io.Writer, data any, opts Options) error {
	// For table renderer, we ignore Empty state messages
	// and only output actual table data
	if opts.Empty {
		return nil
	}

	// Render based on data type - only tables get output
	switch v := data.(type) {
	case TableData:
		return r.colored.renderTable(w, v)
	case KeyValueData:
		// Render key-value as simple table
		return r.renderKeyValueAsTable(w, v)
	case map[string]string:
		kv := NewKeyValueData(v)
		return r.renderKeyValueAsTable(w, kv)
	default:
		// For non-table data, output nothing
		return nil
	}
}

func (r *TableRenderer) renderKeyValueAsTable(w io.Writer, kv KeyValueData) error {
	for _, pair := range kv.Pairs {
		fmt.Fprintf(w, "%s: %s\n", pair.Key, pair.Value)
	}
	return nil
}

// RenderMessage is a no-op for table renderer - we suppress messages
func (r *TableRenderer) RenderMessage(w io.Writer, msg Message) error {
	// Suppress messages in table-only mode
	return nil
}

func init() {
	Register(NewTableRenderer())
}

// CompactRenderer is like ColoredRenderer but more condensed.
// Good for smaller terminals or when you want less visual noise.
type CompactRenderer struct {
	*ColoredRenderer
}

// NewCompactRenderer creates a new compact renderer
func NewCompactRenderer() *CompactRenderer {
	return &CompactRenderer{
		ColoredRenderer: NewColoredRenderer(),
	}
}

// Name returns the renderer identifier
func (r *CompactRenderer) Name() RendererName {
	return RendererCompact
}

// Render outputs data in compact format
func (r *CompactRenderer) Render(w io.Writer, data any, opts Options) error {
	// Handle empty state
	if opts.Empty {
		if opts.EmptyMessage != "" {
			fmt.Fprintln(w, r.styles.muted.Render(opts.EmptyMessage))
		}
		return nil
	}

	// Compact title
	if opts.Title != "" {
		fmt.Fprintln(w, r.styles.title.Render("▸ "+opts.Title))
	}

	// Render based on data type
	switch v := data.(type) {
	case KeyValueData:
		return r.renderCompactKeyValue(w, v)
	case TableData:
		return r.renderCompactTable(w, v)
	case ListData:
		return r.renderCompactList(w, v)
	default:
		return r.ColoredRenderer.Render(w, data, opts)
	}
}

func (r *CompactRenderer) renderCompactKeyValue(w io.Writer, kv KeyValueData) error {
	for _, pair := range kv.Pairs {
		fmt.Fprintf(w, "%s: %s\n",
			r.styles.muted.Render(pair.Key),
			pair.Value)
	}
	return nil
}

func (r *CompactRenderer) renderCompactTable(w io.Writer, t TableData) error {
	if len(t.Rows) == 0 {
		fmt.Fprintln(w, r.styles.muted.Render("(empty)"))
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

	// Headers - muted style, no separator
	var headerParts []string
	for i, h := range t.Headers {
		headerParts = append(headerParts, r.styles.muted.Render(fmt.Sprintf("%-*s", widths[i], h)))
	}
	fmt.Fprintln(w, strings.Join(headerParts, " "))

	// Rows - tighter spacing
	for _, row := range t.Rows {
		var cellParts []string
		for i, cell := range row {
			if i < len(widths) {
				cellParts = append(cellParts, fmt.Sprintf("%-*s", widths[i], cell))
			}
		}
		fmt.Fprintln(w, strings.Join(cellParts, " "))
	}

	return nil
}

func (r *CompactRenderer) renderCompactList(w io.Writer, list ListData) error {
	for _, item := range list.Items {
		fmt.Fprintf(w, "  - %s\n", item)
	}
	return nil
}

func init() {
	Register(NewCompactRenderer())
}
