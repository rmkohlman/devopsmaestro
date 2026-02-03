package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderTypes(t *testing.T) {
	// Verify render type constants
	assert.Equal(t, RenderType("auto"), TypeAuto)
	assert.Equal(t, RenderType("keyvalue"), TypeKeyValue)
	assert.Equal(t, RenderType("table"), TypeTable)
	assert.Equal(t, RenderType("list"), TypeList)
	assert.Equal(t, RenderType("detail"), TypeDetail)
	assert.Equal(t, RenderType("raw"), TypeRaw)
	assert.Equal(t, RenderType("progress"), TypeProgress)
}

func TestRendererNames(t *testing.T) {
	// Verify renderer name constants
	assert.Equal(t, RendererName("json"), RendererJSON)
	assert.Equal(t, RendererName("yaml"), RendererYAML)
	assert.Equal(t, RendererName("colored"), RendererColored)
	assert.Equal(t, RendererName("plain"), RendererPlain)
	assert.Equal(t, RendererName("table"), RendererTable)
	assert.Equal(t, RendererName("compact"), RendererCompact)
}

func TestMessageLevels(t *testing.T) {
	// Verify message level constants
	assert.Equal(t, MessageLevel("info"), LevelInfo)
	assert.Equal(t, MessageLevel("success"), LevelSuccess)
	assert.Equal(t, MessageLevel("warning"), LevelWarning)
	assert.Equal(t, MessageLevel("error"), LevelError)
	assert.Equal(t, MessageLevel("debug"), LevelDebug)
	assert.Equal(t, MessageLevel("progress"), LevelProgress)
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, RendererColored, cfg.Default)
	assert.False(t, cfg.Verbose)
	assert.False(t, cfg.NoColor)
	assert.False(t, cfg.UseNerdFonts)
}

func TestOptions(t *testing.T) {
	opts := Options{
		Type:         TypeKeyValue,
		Title:        "Test Title",
		Headers:      []string{"Col1", "Col2"},
		Empty:        false,
		EmptyMessage: "No data",
		EmptyHints:   []string{"Try this", "Or that"},
		Verbose:      true,
	}

	assert.Equal(t, TypeKeyValue, opts.Type)
	assert.Equal(t, "Test Title", opts.Title)
	assert.Len(t, opts.Headers, 2)
	assert.False(t, opts.Empty)
	assert.Equal(t, "No data", opts.EmptyMessage)
	assert.Len(t, opts.EmptyHints, 2)
	assert.True(t, opts.Verbose)
}

func TestMessage(t *testing.T) {
	msg := Message{
		Level:   LevelSuccess,
		Content: "Operation completed",
	}

	assert.Equal(t, LevelSuccess, msg.Level)
	assert.Equal(t, "Operation completed", msg.Content)
}
