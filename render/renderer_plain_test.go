package render

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlainRenderer_Name(t *testing.T) {
	r := NewPlainRenderer()
	assert.Equal(t, RendererPlain, r.Name())
}

func TestPlainRenderer_SupportsColor(t *testing.T) {
	r := NewPlainRenderer()
	assert.False(t, r.SupportsColor())
}

func TestPlainRenderer_Render(t *testing.T) {
	r := NewPlainRenderer()

	t.Run("KeyValueData", func(t *testing.T) {
		var buf bytes.Buffer
		data := NewOrderedKeyValueData(
			KeyValue{Key: "Project", Value: "test"},
			KeyValue{Key: "Workspace", Value: "dev"},
		)

		err := r.Render(&buf, data, Options{})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Project")
		assert.Contains(t, output, "test")
		assert.Contains(t, output, "Workspace")
		assert.Contains(t, output, "dev")
	})

	t.Run("TableData", func(t *testing.T) {
		var buf bytes.Buffer
		data := TableData{
			Headers: []string{"Name", "Status"},
			Rows: [][]string{
				{"proj1", "active"},
				{"proj2", "stopped"},
			},
		}

		err := r.Render(&buf, data, Options{})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Name")
		assert.Contains(t, output, "Status")
		assert.Contains(t, output, "proj1")
		assert.Contains(t, output, "active")
		assert.Contains(t, output, "----") // separator
	})

	t.Run("ListData", func(t *testing.T) {
		var buf bytes.Buffer
		data := ListData{Items: []string{"item1", "item2"}}

		err := r.Render(&buf, data, Options{})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "* item1")
		assert.Contains(t, output, "* item2")
	})

	t.Run("with title", func(t *testing.T) {
		var buf bytes.Buffer
		data := map[string]string{"key": "value"}

		err := r.Render(&buf, data, Options{Title: "My Section"})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "== My Section ==")
	})

	t.Run("empty state with message", func(t *testing.T) {
		var buf bytes.Buffer

		err := r.Render(&buf, nil, Options{
			Empty:        true,
			EmptyMessage: "No data found",
			EmptyHints:   []string{"Try option A", "Try option B"},
		})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "No data found")
		assert.Contains(t, output, "* Try option A")
		assert.Contains(t, output, "* Try option B")
	})

	t.Run("empty table", func(t *testing.T) {
		var buf bytes.Buffer
		data := TableData{
			Headers: []string{"Name", "Status"},
			Rows:    [][]string{},
		}

		err := r.Render(&buf, data, Options{})
		require.NoError(t, err)

		assert.Contains(t, buf.String(), "No data")
	})
}

func TestPlainRenderer_RenderMessage(t *testing.T) {
	r := NewPlainRenderer()

	tests := []struct {
		level    MessageLevel
		expected string
	}{
		{LevelInfo, "[INFO]"},
		{LevelSuccess, "[OK]"},
		{LevelWarning, "[WARN]"},
		{LevelError, "[ERROR]"},
		{LevelDebug, "[DEBUG]"},
		{LevelProgress, "->"},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			var buf bytes.Buffer
			err := r.RenderMessage(&buf, Message{Level: tt.level, Content: "test"})
			require.NoError(t, err)
			assert.Contains(t, buf.String(), tt.expected)
			assert.Contains(t, buf.String(), "test")
		})
	}
}
