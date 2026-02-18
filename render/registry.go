package render

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
)

var (
	// Global registry and configuration
	registry     = make(map[RendererName]Renderer)
	registryMu   sync.RWMutex
	globalConfig = DefaultConfig()
	configMu     sync.RWMutex

	// Default writer (can be changed for testing)
	defaultWriter io.Writer = os.Stdout
)

// Register adds a renderer to the global registry.
// Renderers are typically registered in init() functions.
func Register(r Renderer) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[r.Name()] = r
}

// Get returns a renderer by name from the registry.
// Returns nil if the renderer is not found.
func Get(name RendererName) Renderer {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return registry[name]
}

// List returns all registered renderer names
func List() []RendererName {
	registryMu.RLock()
	defer registryMu.RUnlock()
	names := make([]RendererName, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

// SetConfig updates the global renderer configuration
func SetConfig(cfg Config) {
	configMu.Lock()
	defer configMu.Unlock()
	globalConfig = cfg
}

// GetConfig returns the current global configuration
func GetConfig() Config {
	configMu.RLock()
	defer configMu.RUnlock()
	return globalConfig
}

// SetDefault sets the default renderer
func SetDefault(name RendererName) {
	configMu.Lock()
	defer configMu.Unlock()
	globalConfig.Default = name
}

// SetWriter sets the default output writer (useful for testing)
func SetWriter(w io.Writer) {
	defaultWriter = w
}

// GetWriter returns the current default writer
func GetWriter() io.Writer {
	return defaultWriter
}

// ResolveRenderer determines which renderer to use based on:
// 1. Explicit override (from flag)
// 2. Environment variable DVM_RENDER
// 3. Global config default
func ResolveRenderer(override string) Renderer {
	var name RendererName

	// Priority 1: Explicit override (from -r flag)
	if override != "" {
		name = RendererName(override)
	} else if envRender := os.Getenv("DVM_RENDER"); envRender != "" {
		// Priority 2: Environment variable
		name = RendererName(envRender)
	} else {
		// Priority 3: Global config default
		cfg := GetConfig()
		name = cfg.Default
	}

	// Check NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		// Force plain renderer if NO_COLOR is set
		if name == RendererColored || name == RendererCompact {
			name = RendererPlain
		}
	}

	r := Get(name)
	if r == nil {
		// Fallback to colored if requested renderer not found
		r = Get(RendererColored)
	}
	if r == nil {
		// Last resort fallback to plain
		r = Get(RendererPlain)
	}

	return r
}

// OutputWithContext renders data using the resolved renderer to the default writer with context.
// This is the new context-aware entry point for commands to output data.
func OutputWithContext(ctx context.Context, data any, opts Options) error {
	return OutputToWithContext(ctx, defaultWriter, "", data, opts)
}

// OutputWithContextAndRenderer renders data using a specific renderer (by name or from flag) with context.
func OutputWithContextAndRenderer(ctx context.Context, rendererOverride string, data any, opts Options) error {
	return OutputToWithContext(ctx, defaultWriter, rendererOverride, data, opts)
}

// OutputToWithContext renders data to a specific writer with optional renderer override and context.
func OutputToWithContext(ctx context.Context, w io.Writer, rendererOverride string, data any, opts Options) error {
	r := ResolveRenderer(rendererOverride)
	if r == nil {
		return fmt.Errorf("no renderer available")
	}
	return r.RenderWithContext(ctx, w, data, opts)
}

// Output renders data using the resolved renderer to the default writer.
// This is the main entry point for commands to output data.
func Output(data any, opts Options) error {
	return OutputTo(defaultWriter, "", data, opts)
}

// OutputWith renders data using a specific renderer (by name or from flag).
func OutputWith(rendererOverride string, data any, opts Options) error {
	return OutputTo(defaultWriter, rendererOverride, data, opts)
}

// OutputTo renders data to a specific writer with optional renderer override.
func OutputTo(w io.Writer, rendererOverride string, data any, opts Options) error {
	r := ResolveRenderer(rendererOverride)
	if r == nil {
		return fmt.Errorf("no renderer available")
	}
	return r.Render(w, data, opts)
}

// MsgWithContext outputs a status message using the resolved renderer with context.
func MsgWithContext(ctx context.Context, level MessageLevel, content string) error {
	return MsgToWithContext(ctx, defaultWriter, "", Message{Level: level, Content: content})
}

// MsgWithContextAndRenderer outputs a status message with a specific renderer and context.
func MsgWithContextAndRenderer(ctx context.Context, rendererOverride string, level MessageLevel, content string) error {
	return MsgToWithContext(ctx, defaultWriter, rendererOverride, Message{Level: level, Content: content})
}

// MsgToWithContext outputs a status message to a specific writer with context.
func MsgToWithContext(ctx context.Context, w io.Writer, rendererOverride string, msg Message) error {
	r := ResolveRenderer(rendererOverride)
	if r == nil {
		return fmt.Errorf("no renderer available")
	}
	return r.RenderMessageWithContext(ctx, w, msg)
}

// Message outputs a status message using the resolved renderer.
func Msg(level MessageLevel, content string) error {
	return MsgTo(defaultWriter, "", Message{Level: level, Content: content})
}

// MsgWith outputs a status message with a specific renderer.
func MsgWith(rendererOverride string, level MessageLevel, content string) error {
	return MsgTo(defaultWriter, rendererOverride, Message{Level: level, Content: content})
}

// MsgTo outputs a status message to a specific writer.
func MsgTo(w io.Writer, rendererOverride string, msg Message) error {
	r := ResolveRenderer(rendererOverride)
	if r == nil {
		return fmt.Errorf("no renderer available")
	}
	return r.RenderMessage(w, msg)
}

// Convenience functions for common message types

// Info outputs an info message
func Info(content string) error {
	return Msg(LevelInfo, content)
}

// Success outputs a success message
func Success(content string) error {
	return Msg(LevelSuccess, content)
}

// Warning outputs a warning message
func Warning(content string) error {
	return Msg(LevelWarning, content)
}

// Error outputs an error message
func Error(content string) error {
	return Msg(LevelError, content)
}

// Progress outputs a progress message
func Progress(content string) error {
	return Msg(LevelProgress, content)
}
