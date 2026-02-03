package render

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestColoredRenderer_Name(t *testing.T) {
	r := NewColoredRenderer()
	assert.Equal(t, RendererColored, r.Name())
}

func TestColoredRenderer_SupportsColor(t *testing.T) {
	r := NewColoredRenderer()
	assert.True(t, r.SupportsColor())
}

func TestColoredRenderer_Render(t *testing.T) {
	r := NewColoredRenderer()

	t.Run("KeyValueData", func(t *testing.T) {
		var buf bytes.Buffer
		data := NewOrderedKeyValueData(
			KeyValue{Key: "Project", Value: "test"},
			KeyValue{Key: "Workspace", Value: "dev"},
		)

		err := r.Render(&buf, data, Options{})
		require.NoError(t, err)

		output := buf.String()
		// Content should be present (with ANSI codes)
		assert.Contains(t, output, "Project")
		assert.Contains(t, output, "test")
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
		assert.Contains(t, output, "proj1")
	})

	t.Run("with title", func(t *testing.T) {
		var buf bytes.Buffer
		data := map[string]string{"key": "value"}

		err := r.Render(&buf, data, Options{Title: "Test Section"})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Test Section")
	})

	t.Run("empty state", func(t *testing.T) {
		var buf bytes.Buffer

		err := r.Render(&buf, nil, Options{
			Empty:        true,
			EmptyMessage: "No items found",
			EmptyHints:   []string{"Add an item"},
		})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "No items found")
		assert.Contains(t, output, "Add an item")
	})
}

func TestColoredRenderer_RenderMessage(t *testing.T) {
	r := NewColoredRenderer()

	levels := []MessageLevel{
		LevelInfo,
		LevelSuccess,
		LevelWarning,
		LevelError,
		LevelProgress,
		LevelDebug,
	}

	for _, level := range levels {
		t.Run(string(level), func(t *testing.T) {
			var buf bytes.Buffer
			err := r.RenderMessage(&buf, Message{Level: level, Content: "test message"})
			require.NoError(t, err)
			assert.Contains(t, buf.String(), "test message")
		})
	}
}

func TestColoredRendererWithIcons(t *testing.T) {
	nerdIcons := NerdFontIcons()
	r := NewColoredRendererWithIcons(nerdIcons)

	var buf bytes.Buffer
	err := r.RenderMessage(&buf, Message{Level: LevelSuccess, Content: "done"})
	require.NoError(t, err)

	// Verify it renders (specific icon chars may not display in test output)
	assert.Contains(t, buf.String(), "done")
}

func TestIcons(t *testing.T) {
	t.Run("DefaultIcons", func(t *testing.T) {
		icons := DefaultIcons()
		assert.NotEmpty(t, icons.Success)
		assert.NotEmpty(t, icons.Warning)
		assert.NotEmpty(t, icons.Error)
		assert.NotEmpty(t, icons.Info)
		assert.NotEmpty(t, icons.Progress)
		assert.NotEmpty(t, icons.Bullet)
	})

	t.Run("NerdFontIcons", func(t *testing.T) {
		icons := NerdFontIcons()
		assert.NotEmpty(t, icons.Success)
		assert.NotEmpty(t, icons.Warning)
	})

	t.Run("PlainIcons", func(t *testing.T) {
		icons := PlainIcons()
		assert.Equal(t, "[OK]", icons.Success)
		assert.Equal(t, "[!]", icons.Warning)
		assert.Equal(t, "[X]", icons.Error)
	})
}
