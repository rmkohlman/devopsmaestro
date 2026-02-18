package render

import (
	"context"
	"io"

	"gopkg.in/yaml.v3"
)

// YAMLRenderer outputs data as YAML.
// It ignores titles, hints, and other human-readable options.
type YAMLRenderer struct{}

// NewYAMLRenderer creates a new YAML renderer
func NewYAMLRenderer() *YAMLRenderer {
	return &YAMLRenderer{}
}

// Name returns the renderer identifier
func (r *YAMLRenderer) Name() RendererName {
	return RendererYAML
}

// SupportsColor returns false - YAML doesn't use colors
func (r *YAMLRenderer) SupportsColor() bool {
	return false
}

// Render outputs data as YAML
func (r *YAMLRenderer) Render(w io.Writer, data any, opts Options) error {
	// YAML renderer ignores context - no theming needed
	return r.RenderWithContext(context.Background(), w, data, opts)
}

// RenderWithContext outputs data as YAML (ignores context for YAML output)
func (r *YAMLRenderer) RenderWithContext(ctx context.Context, w io.Writer, data any, opts Options) error {
	// Handle empty state
	if opts.Empty {
		if data == nil {
			data = map[string]any{}
		}
	}

	// Handle special data types
	switch v := data.(type) {
	case KeyValueData:
		// Convert to map for YAML (note: order may not be preserved)
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

	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	defer encoder.Close()
	return encoder.Encode(data)
}

// RenderMessage outputs a message as YAML
func (r *YAMLRenderer) RenderMessage(w io.Writer, msg Message) error {
	// YAML renderer ignores context - no theming needed
	return r.RenderMessageWithContext(context.Background(), w, msg)
}

// RenderMessageWithContext outputs a message as YAML (ignores context)
func (r *YAMLRenderer) RenderMessageWithContext(ctx context.Context, w io.Writer, msg Message) error {
	data := map[string]string{
		"level":   string(msg.Level),
		"message": msg.Content,
	}
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	defer encoder.Close()
	return encoder.Encode(data)
}

func init() {
	Register(NewYAMLRenderer())
}
