package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyValueData(t *testing.T) {
	t.Run("NewKeyValueData from map", func(t *testing.T) {
		m := map[string]string{
			"key1": "value1",
			"key2": "value2",
		}

		kv := NewKeyValueData(m)
		assert.Len(t, kv.Pairs, 2)

		// Verify all pairs are present (order not guaranteed from map)
		values := make(map[string]string)
		for _, pair := range kv.Pairs {
			values[pair.Key] = pair.Value
		}
		assert.Equal(t, "value1", values["key1"])
		assert.Equal(t, "value2", values["key2"])
	})

	t.Run("NewOrderedKeyValueData preserves order", func(t *testing.T) {
		kv := NewOrderedKeyValueData(
			KeyValue{Key: "first", Value: "1"},
			KeyValue{Key: "second", Value: "2"},
			KeyValue{Key: "third", Value: "3"},
		)

		assert.Len(t, kv.Pairs, 3)
		assert.Equal(t, "first", kv.Pairs[0].Key)
		assert.Equal(t, "second", kv.Pairs[1].Key)
		assert.Equal(t, "third", kv.Pairs[2].Key)
	})
}

func TestTableData(t *testing.T) {
	table := TableData{
		Headers: []string{"Name", "Status", "Age"},
		Rows: [][]string{
			{"project1", "active", "5d"},
			{"project2", "stopped", "10d"},
		},
	}

	assert.Len(t, table.Headers, 3)
	assert.Len(t, table.Rows, 2)
	assert.Equal(t, "project1", table.Rows[0][0])
}

func TestListData(t *testing.T) {
	list := ListData{
		Items: []string{"item1", "item2", "item3"},
	}

	assert.Len(t, list.Items, 3)
	assert.Equal(t, "item1", list.Items[0])
}

func TestRendererInterface(t *testing.T) {
	// Verify all renderers implement the interface
	renderers := []Renderer{
		NewJSONRenderer(),
		NewYAMLRenderer(),
		NewPlainRenderer(),
		NewColoredRenderer(),
		NewTableRenderer(),
		NewCompactRenderer(),
	}

	for _, r := range renderers {
		assert.NotEmpty(t, r.Name(), "Renderer should have a name")

		// SupportsColor should return a boolean (no panic)
		_ = r.SupportsColor()
	}
}

func TestRendererInterfaceCompliance(t *testing.T) {
	// Ensure all renderer types satisfy the Renderer interface at compile time
	var _ Renderer = (*JSONRenderer)(nil)
	var _ Renderer = (*YAMLRenderer)(nil)
	var _ Renderer = (*PlainRenderer)(nil)
	var _ Renderer = (*ColoredRenderer)(nil)
	var _ Renderer = (*TableRenderer)(nil)
	var _ Renderer = (*CompactRenderer)(nil)
}
