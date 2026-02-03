package render

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTableRenderer_Name(t *testing.T) {
	r := NewTableRenderer()
	assert.Equal(t, RendererTable, r.Name())
}

func TestTableRenderer_SupportsColor(t *testing.T) {
	r := NewTableRenderer()
	assert.True(t, r.SupportsColor())
}

func TestTableRenderer_Render(t *testing.T) {
	r := NewTableRenderer()

	t.Run("TableData", func(t *testing.T) {
		var buf bytes.Buffer
		data := TableData{
			Headers: []string{"Name", "Status"},
			Rows: [][]string{
				{"proj1", "active"},
			},
		}

		err := r.Render(&buf, data, Options{})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Name")
		assert.Contains(t, output, "proj1")
	})

	t.Run("KeyValueData as table", func(t *testing.T) {
		var buf bytes.Buffer
		data := NewOrderedKeyValueData(
			KeyValue{Key: "Project", Value: "test"},
		)

		err := r.Render(&buf, data, Options{})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Project")
		assert.Contains(t, output, "test")
	})

	t.Run("empty state suppressed", func(t *testing.T) {
		var buf bytes.Buffer

		err := r.Render(&buf, nil, Options{
			Empty:        true,
			EmptyMessage: "Should not appear",
		})
		require.NoError(t, err)

		// Table renderer suppresses empty messages
		assert.Empty(t, buf.String())
	})

	t.Run("non-table data suppressed", func(t *testing.T) {
		var buf bytes.Buffer
		data := "just a string"

		err := r.Render(&buf, data, Options{})
		require.NoError(t, err)

		assert.Empty(t, buf.String())
	})
}

func TestTableRenderer_RenderMessage(t *testing.T) {
	r := NewTableRenderer()
	var buf bytes.Buffer

	// Messages should be suppressed
	err := r.RenderMessage(&buf, Message{Level: LevelSuccess, Content: "done"})
	require.NoError(t, err)

	assert.Empty(t, buf.String())
}

func TestCompactRenderer_Name(t *testing.T) {
	r := NewCompactRenderer()
	assert.Equal(t, RendererCompact, r.Name())
}

func TestCompactRenderer_Render(t *testing.T) {
	r := NewCompactRenderer()

	t.Run("KeyValueData", func(t *testing.T) {
		var buf bytes.Buffer
		data := NewOrderedKeyValueData(
			KeyValue{Key: "Project", Value: "test"},
		)

		err := r.Render(&buf, data, Options{})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Project")
		assert.Contains(t, output, "test")
	})

	t.Run("with title uses compact format", func(t *testing.T) {
		var buf bytes.Buffer
		data := map[string]string{"key": "value"}

		err := r.Render(&buf, data, Options{Title: "Compact Title"})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "▸") // Compact uses different section marker
		assert.Contains(t, output, "Compact Title")
	})

	t.Run("empty state", func(t *testing.T) {
		var buf bytes.Buffer

		err := r.Render(&buf, nil, Options{
			Empty:        true,
			EmptyMessage: "Nothing here",
		})
		require.NoError(t, err)

		assert.Contains(t, buf.String(), "Nothing here")
	})
}
