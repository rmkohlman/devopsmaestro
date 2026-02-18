package render

import (
	"context"
	"io"
)

// Renderer is the interface that all renderers must implement.
// Renderers are responsible for deciding how to display data based on
// the render type and options provided.
type Renderer interface {
	// Render outputs the data to the writer according to this renderer's style.
	// The Options provide hints about the data structure and display preferences.
	Render(w io.Writer, data any, opts Options) error

	// RenderWithContext outputs the data with context available for theming.
	// New context-aware method that allows access to ColorProvider.
	RenderWithContext(ctx context.Context, w io.Writer, data any, opts Options) error

	// RenderMessage outputs a status message (info, success, warning, error, etc.)
	RenderMessage(w io.Writer, msg Message) error

	// RenderMessageWithContext outputs a status message with context available for theming.
	// New context-aware method that allows access to ColorProvider.
	RenderMessageWithContext(ctx context.Context, w io.Writer, msg Message) error

	// Name returns the renderer's identifier
	Name() RendererName

	// SupportsColor returns true if this renderer uses colors
	SupportsColor() bool
}

// TableData represents data that should be rendered as a table.
// Commands can use this to explicitly structure table data.
type TableData struct {
	Headers []string
	Rows    [][]string
}

// ListData represents data that should be rendered as a list.
type ListData struct {
	Items []string
}

// KeyValueData represents key-value pairs for rendering.
// Using ordered slice to preserve key order (maps don't guarantee order).
type KeyValueData struct {
	Pairs []KeyValue
}

// KeyValue is a single key-value pair
type KeyValue struct {
	Key   string
	Value string
}

// NewKeyValueData creates KeyValueData from a map (order not guaranteed)
func NewKeyValueData(m map[string]string) KeyValueData {
	kv := KeyValueData{Pairs: make([]KeyValue, 0, len(m))}
	for k, v := range m {
		kv.Pairs = append(kv.Pairs, KeyValue{Key: k, Value: v})
	}
	return kv
}

// NewOrderedKeyValueData creates KeyValueData with guaranteed order
func NewOrderedKeyValueData(pairs ...KeyValue) KeyValueData {
	return KeyValueData{Pairs: pairs}
}
