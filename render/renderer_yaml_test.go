package render

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestYAMLRenderer_Name(t *testing.T) {
	r := NewYAMLRenderer()
	assert.Equal(t, RendererYAML, r.Name())
}

func TestYAMLRenderer_SupportsColor(t *testing.T) {
	r := NewYAMLRenderer()
	assert.False(t, r.SupportsColor())
}

func TestYAMLRenderer_Render(t *testing.T) {
	r := NewYAMLRenderer()

	t.Run("map data", func(t *testing.T) {
		var buf bytes.Buffer
		data := map[string]string{"key": "value", "foo": "bar"}

		err := r.Render(&buf, data, Options{})
		require.NoError(t, err)

		var result map[string]string
		err = yaml.Unmarshal(buf.Bytes(), &result)
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
		err = yaml.Unmarshal(buf.Bytes(), &result)
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
		err = yaml.Unmarshal(buf.Bytes(), &result)
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
		err = yaml.Unmarshal(buf.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, []string{"item1", "item2", "item3"}, result)
	})

	t.Run("empty state", func(t *testing.T) {
		var buf bytes.Buffer

		err := r.Render(&buf, nil, Options{Empty: true})
		require.NoError(t, err)

		var result map[string]any
		err = yaml.Unmarshal(buf.Bytes(), &result)
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

func TestYAMLRenderer_RenderMessage(t *testing.T) {
	r := NewYAMLRenderer()
	var buf bytes.Buffer

	err := r.RenderMessage(&buf, Message{Level: LevelWarning, Content: "Watch out!"})
	require.NoError(t, err)

	var result map[string]string
	err = yaml.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, "warning", result["level"])
	assert.Equal(t, "Watch out!", result["message"])
}
