package render

import (
	"context"
	"fmt"
	"io"
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
	// Table renderer ignores context - delegates to ColoredRenderer
	return r.RenderWithContext(context.Background(), w, data, opts)
}

// RenderWithContext outputs only table data, suppressing other elements
func (r *TableRenderer) RenderWithContext(ctx context.Context, w io.Writer, data any, opts Options) error {
	// For table renderer, we ignore Empty state messages
	// and only output actual table data
	if opts.Empty {
		return nil
	}

	// Render based on data type - only tables get output
	switch v := data.(type) {
	case TableData:
		return r.colored.renderTableWithStyles(w, v, r.colored.getStyles(ctx))
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

// RenderMessageWithContext is a no-op for table renderer - we suppress messages
func (r *TableRenderer) RenderMessageWithContext(ctx context.Context, w io.Writer, msg Message) error {
	// Suppress messages in table-only mode
	return nil
}

func init() {
	Register(NewTableRenderer())
}
