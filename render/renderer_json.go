package render

import (
	"encoding/json"
	"io"
)

// JSONRenderer outputs data as formatted JSON.
// It ignores titles, hints, and other human-readable options.
type JSONRenderer struct {
	indent bool
}

// NewJSONRenderer creates a new JSON renderer
func NewJSONRenderer() *JSONRenderer {
	return &JSONRenderer{indent: true}
}

// Name returns the renderer identifier
func (r *JSONRenderer) Name() RendererName {
	return RendererJSON
}

// SupportsColor returns false - JSON doesn't use colors
func (r *JSONRenderer) SupportsColor() bool {
	return false
}

// Render outputs data as JSON
func (r *JSONRenderer) Render(w io.Writer, data any, opts Options) error {
	// Handle empty state
	if opts.Empty {
		// For JSON, output an empty object or the data as-is
		if data == nil {
			data = map[string]any{}
		}
	}

	// Handle special data types
	switch v := data.(type) {
	case KeyValueData:
		// Convert to map for JSON
		m := make(map[string]string)
		for _, kv := range v.Pairs {
			m[kv.Key] = kv.Value
		}
		data = m
	case TableData:
		// Convert to array of maps
		rows := make([]map[string]string, len(v.Rows))
		for i, row := range v.Rows {
			rows[i] = make(map[string]string)
			for j, cell := range row {
				if j < len(v.Headers) {
					rows[i][v.Headers[j]] = cell
				}
			}
		}
		data = rows
	case ListData:
		data = v.Items
	}

	encoder := json.NewEncoder(w)
	if r.indent {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(data)
}

// RenderMessage outputs a message as JSON
func (r *JSONRenderer) RenderMessage(w io.Writer, msg Message) error {
	data := map[string]string{
		"level":   string(msg.Level),
		"message": msg.Content,
	}
	encoder := json.NewEncoder(w)
	if r.indent {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(data)
}

func init() {
	Register(NewJSONRenderer())
}
