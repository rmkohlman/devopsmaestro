package render

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONRenderer_Name(t *testing.T) {
	r := NewJSONRenderer()
	assert.Equal(t, RendererJSON, r.Name())
}

func TestJSONRenderer_SupportsColor(t *testing.T) {
	r := NewJSONRenderer()
	assert.False(t, r.SupportsColor())
}

func TestJSONRenderer_Render(t *testing.T) {
	r := NewJSONRenderer()

	t.Run("map data", func(t *testing.T) {
		var buf bytes.Buffer
		data := map[string]string{"key": "value", "foo": "bar"}

		err := r.Render(&buf, data, Options{})
		require.NoError(t, err)

		var result map[string]string
		err = json.Unmarshal(buf.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "value", result["key"])
		assert.Equal(t, "bar", result["foo"])
	})

	t.Run("KeyValueData", func(t *testing.T) {
		var buf bytes.Buffer
		data := NewOrderedKeyValueData(
			KeyValue{Key: "project", Value: "test"},
			KeyValue{Key: "workspace", Value: "dev"},
		)

		err := r.Render(&buf, data, Options{})
		require.NoError(t, err)

		var result map[string]string
		err = json.Unmarshal(buf.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "test", result["project"])
		assert.Equal(t, "dev", result["workspace"])
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

		var result []map[string]string
		err = json.Unmarshal(buf.Bytes(), &result)
		require.NoError(t, err)

		assert.Len(t, result, 2)
		assert.Equal(t, "proj1", result[0]["Name"])
		assert.Equal(t, "stopped", result[1]["Status"])
	})

	t.Run("ListData", func(t *testing.T) {
		var buf bytes.Buffer
		data := ListData{Items: []string{"item1", "item2", "item3"}}

		err := r.Render(&buf, data, Options{})
		require.NoError(t, err)

		var result []string
		err = json.Unmarshal(buf.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, []string{"item1", "item2", "item3"}, result)
	})

	t.Run("empty state", func(t *testing.T) {
		var buf bytes.Buffer

		err := r.Render(&buf, nil, Options{Empty: true})
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(buf.Bytes(), &result)
		require.NoError(t, err)

		assert.Empty(t, result)
	})

	t.Run("ignores title", func(t *testing.T) {
		var buf bytes.Buffer
		data := map[string]string{"key": "value"}

		err := r.Render(&buf, data, Options{Title: "Should Not Appear"})
		require.NoError(t, err)

		assert.NotContains(t, buf.String(), "Should Not Appear")
	})
}

func TestJSONRenderer_RenderMessage(t *testing.T) {
	r := NewJSONRenderer()
	var buf bytes.Buffer

	err := r.RenderMessage(&buf, Message{Level: LevelSuccess, Content: "Done!"})
	require.NoError(t, err)

	var result map[string]string
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, "success", result["level"])
	assert.Equal(t, "Done!", result["message"])
}
