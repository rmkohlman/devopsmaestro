package render

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	// Registry should have renderers from init()
	names := List()
	assert.Contains(t, names, RendererJSON)
	assert.Contains(t, names, RendererYAML)
	assert.Contains(t, names, RendererPlain)
	assert.Contains(t, names, RendererColored)
	assert.Contains(t, names, RendererTable)
	assert.Contains(t, names, RendererCompact)
}

func TestGet(t *testing.T) {
	t.Run("existing renderer", func(t *testing.T) {
		r := Get(RendererJSON)
		assert.NotNil(t, r)
		assert.Equal(t, RendererJSON, r.Name())
	})

	t.Run("non-existing renderer", func(t *testing.T) {
		r := Get("nonexistent")
		assert.Nil(t, r)
	})
}

func TestSetConfig(t *testing.T) {
	// Save original config
	original := GetConfig()
	defer SetConfig(original)

	newCfg := Config{
		Default:      RendererPlain,
		Verbose:      true,
		NoColor:      true,
		UseNerdFonts: true,
	}

	SetConfig(newCfg)
	cfg := GetConfig()

	assert.Equal(t, RendererPlain, cfg.Default)
	assert.True(t, cfg.Verbose)
	assert.True(t, cfg.NoColor)
	assert.True(t, cfg.UseNerdFonts)
}

func TestSetDefault(t *testing.T) {
	// Save original config
	original := GetConfig()
	defer SetConfig(original)

	SetDefault(RendererYAML)
	cfg := GetConfig()
	assert.Equal(t, RendererYAML, cfg.Default)
}

func TestResolveRenderer(t *testing.T) {
	// Save original config and env
	original := GetConfig()
	originalEnv := os.Getenv("DVM_RENDER")
	defer func() {
		SetConfig(original)
		os.Setenv("DVM_RENDER", originalEnv)
	}()

	t.Run("explicit override takes priority", func(t *testing.T) {
		SetDefault(RendererColored)
		os.Setenv("DVM_RENDER", "yaml")

		r := ResolveRenderer("json")
		assert.Equal(t, RendererJSON, r.Name())
	})

	t.Run("env var takes priority over default", func(t *testing.T) {
		SetDefault(RendererColored)
		os.Setenv("DVM_RENDER", "yaml")

		r := ResolveRenderer("")
		assert.Equal(t, RendererYAML, r.Name())
	})

	t.Run("default used when no override", func(t *testing.T) {
		SetDefault(RendererPlain)
		os.Unsetenv("DVM_RENDER")

		r := ResolveRenderer("")
		assert.Equal(t, RendererPlain, r.Name())
	})

	t.Run("NO_COLOR forces plain for colored", func(t *testing.T) {
		os.Setenv("NO_COLOR", "1")
		defer os.Unsetenv("NO_COLOR")

		r := ResolveRenderer("colored")
		assert.Equal(t, RendererPlain, r.Name())
	})
}

func TestOutput(t *testing.T) {
	var buf bytes.Buffer
	SetWriter(&buf)
	defer SetWriter(os.Stdout)

	data := map[string]string{"key": "value"}
	opts := Options{Type: TypeKeyValue}

	// Use JSON renderer for deterministic output
	err := OutputWith("json", data, opts)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "key")
	assert.Contains(t, buf.String(), "value")
}

func TestOutputTo(t *testing.T) {
	var buf bytes.Buffer

	data := KeyValueData{
		Pairs: []KeyValue{
			{Key: "name", Value: "test"},
		},
	}
	opts := Options{Type: TypeKeyValue}

	err := OutputTo(&buf, "json", data, opts)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "name")
	assert.Contains(t, buf.String(), "test")
}

func TestMsg(t *testing.T) {
	var buf bytes.Buffer
	SetWriter(&buf)
	defer SetWriter(os.Stdout)

	err := MsgWith("plain", LevelSuccess, "Operation completed")
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "[OK]")
	assert.Contains(t, buf.String(), "Operation completed")
}

func TestConvenienceFunctions(t *testing.T) {
	var buf bytes.Buffer
	SetWriter(&buf)
	defer SetWriter(os.Stdout)

	// Use plain renderer for predictable output
	SetDefault(RendererPlain)
	defer SetDefault(RendererColored)

	tests := []struct {
		name     string
		fn       func(string) error
		expected string
	}{
		{"Info", Info, "[INFO]"},
		{"Success", Success, "[OK]"},
		{"Warning", Warning, "[WARN]"},
		{"Error", Error, "[ERROR]"},
		{"Progress", Progress, "->"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			err := tt.fn("test message")
			require.NoError(t, err)
			assert.Contains(t, buf.String(), tt.expected)
			assert.Contains(t, buf.String(), "test message")
		})
	}
}
