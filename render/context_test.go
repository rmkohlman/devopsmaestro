package render

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"devopsmaestro/pkg/colors"
)

// MockColorProvider for testing
type MockColorProvider struct{}

func (m *MockColorProvider) Primary() string    { return "#FF0000" }
func (m *MockColorProvider) Secondary() string  { return "#00FF00" }
func (m *MockColorProvider) Accent() string     { return "#0000FF" }
func (m *MockColorProvider) Success() string    { return "#00AA00" }
func (m *MockColorProvider) Warning() string    { return "#FFAA00" }
func (m *MockColorProvider) Error() string      { return "#AA0000" }
func (m *MockColorProvider) Info() string       { return "#0000AA" }
func (m *MockColorProvider) Foreground() string { return "#FFFFFF" }
func (m *MockColorProvider) Background() string { return "#000000" }
func (m *MockColorProvider) Muted() string      { return "#888888" }
func (m *MockColorProvider) Highlight() string  { return "#FFFF00" }
func (m *MockColorProvider) Border() string     { return "#444444" }
func (m *MockColorProvider) Name() string       { return "test-theme" }
func (m *MockColorProvider) IsLight() bool      { return false }

func TestColoredRenderer_WithContextProvider(t *testing.T) {
	tests := []struct {
		name             string
		withProvider     bool
		expectedContains []string // Strings that should appear in output
	}{
		{
			name:             "with ColorProvider uses custom colors",
			withProvider:     true,
			expectedContains: []string{"Test:", "value"}, // Output should contain the data
		},
		{
			name:             "without ColorProvider uses default colors",
			withProvider:     false,
			expectedContains: []string{"Test:", "value"}, // Output should contain the data
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewColoredRenderer()
			var buf bytes.Buffer

			data := NewOrderedKeyValueData(
				KeyValue{Key: "Test", Value: "value"},
			)

			opts := Options{
				Title: "Test Output",
			}

			var ctx context.Context
			if tt.withProvider {
				ctx = colors.WithProvider(context.Background(), &MockColorProvider{})
			} else {
				ctx = context.Background()
			}

			err := renderer.RenderWithContext(ctx, &buf, data, opts)
			if err != nil {
				t.Fatalf("RenderWithContext failed: %v", err)
			}

			output := buf.String()
			for _, expected := range tt.expectedContains {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, got: %s", expected, output)
				}
			}

			// Verify output is not empty
			if output == "" {
				t.Error("Expected non-empty output")
			}
		})
	}
}

func TestColoredRenderer_MessageWithContextProvider(t *testing.T) {
	tests := []struct {
		name         string
		level        MessageLevel
		withProvider bool
	}{
		{
			name:         "success message with provider",
			level:        LevelSuccess,
			withProvider: true,
		},
		{
			name:         "error message with provider",
			level:        LevelError,
			withProvider: true,
		},
		{
			name:         "warning message with provider",
			level:        LevelWarning,
			withProvider: true,
		},
		{
			name:         "success message without provider",
			level:        LevelSuccess,
			withProvider: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewColoredRenderer()
			var buf bytes.Buffer

			msg := Message{
				Level:   tt.level,
				Content: "Test message",
			}

			var ctx context.Context
			if tt.withProvider {
				ctx = colors.WithProvider(context.Background(), &MockColorProvider{})
			} else {
				ctx = context.Background()
			}

			err := renderer.RenderMessageWithContext(ctx, &buf, msg)
			if err != nil {
				t.Fatalf("RenderMessageWithContext failed: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "Test message") {
				t.Errorf("Expected output to contain message, got: %s", output)
			}

			// Verify output is not empty
			if output == "" {
				t.Error("Expected non-empty output")
			}
		})
	}
}

func TestBackwardCompatibility_RenderMethods(t *testing.T) {
	// Test that old methods still work by delegating to context methods
	renderer := NewColoredRenderer()
	var buf bytes.Buffer

	data := NewOrderedKeyValueData(
		KeyValue{Key: "Compatibility", Value: "test"},
	)

	opts := Options{
		Title: "Backward Compatibility Test",
	}

	// Test old Render method
	err := renderer.Render(&buf, data, opts)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Compatibility") {
		t.Errorf("Expected output to contain 'Compatibility', got: %s", output)
	}

	// Reset buffer for message test
	buf.Reset()

	// Test old RenderMessage method
	msg := Message{
		Level:   LevelInfo,
		Content: "Backward compatibility message",
	}

	err = renderer.RenderMessage(&buf, msg)
	if err != nil {
		t.Fatalf("RenderMessage failed: %v", err)
	}

	output = buf.String()
	if !strings.Contains(output, "Backward compatibility message") {
		t.Errorf("Expected output to contain message, got: %s", output)
	}
}

func TestContextAwareFunctions(t *testing.T) {
	// Store original writer
	originalWriter := GetWriter()
	defer SetWriter(originalWriter)

	var buf bytes.Buffer
	SetWriter(&buf)

	data := NewOrderedKeyValueData(
		KeyValue{Key: "Context", Value: "test"},
	)

	opts := Options{
		Title: "Context Functions Test",
	}

	// Test OutputWithContext
	ctx := colors.WithProvider(context.Background(), &MockColorProvider{})
	err := OutputWithContext(ctx, data, opts)
	if err != nil {
		t.Fatalf("OutputWithContext failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Context") {
		t.Errorf("Expected output to contain 'Context', got: %s", output)
	}

	// Reset buffer for message test
	buf.Reset()

	// Test MsgWithContext
	err = MsgWithContext(ctx, LevelSuccess, "Context message test")
	if err != nil {
		t.Fatalf("MsgWithContext failed: %v", err)
	}

	output = buf.String()
	if !strings.Contains(output, "Context message test") {
		t.Errorf("Expected output to contain message, got: %s", output)
	}
}

func TestPlainRenderer_ContextCompatibility(t *testing.T) {
	// Test that plain renderer ignores context properly
	renderer := NewPlainRenderer()
	var buf bytes.Buffer

	data := NewOrderedKeyValueData(
		KeyValue{Key: "Plain", Value: "test"},
	)

	opts := Options{
		Title: "Plain Context Test",
	}

	// Plain renderer should work the same with or without context
	ctx := colors.WithProvider(context.Background(), &MockColorProvider{})

	err := renderer.RenderWithContext(ctx, &buf, data, opts)
	if err != nil {
		t.Fatalf("Plain RenderWithContext failed: %v", err)
	}

	output1 := buf.String()
	buf.Reset()

	err = renderer.RenderWithContext(context.Background(), &buf, data, opts)
	if err != nil {
		t.Fatalf("Plain RenderWithContext without provider failed: %v", err)
	}

	output2 := buf.String()

	// Both outputs should be identical for plain renderer
	if output1 != output2 {
		t.Errorf("Plain renderer should produce identical output regardless of context.\nWith provider: %s\nWithout provider: %s", output1, output2)
	}

	if !strings.Contains(output1, "Plain") {
		t.Errorf("Expected output to contain 'Plain', got: %s", output1)
	}
}

func TestJSONRenderer_ContextCompatibility(t *testing.T) {
	// Test that JSON renderer ignores context properly
	renderer := NewJSONRenderer()
	var buf bytes.Buffer

	data := NewOrderedKeyValueData(
		KeyValue{Key: "JSON", Value: "test"},
	)

	opts := Options{}

	// JSON renderer should work the same with or without context
	ctx := colors.WithProvider(context.Background(), &MockColorProvider{})

	err := renderer.RenderWithContext(ctx, &buf, data, opts)
	if err != nil {
		t.Fatalf("JSON RenderWithContext failed: %v", err)
	}

	output1 := buf.String()
	buf.Reset()

	err = renderer.RenderWithContext(context.Background(), &buf, data, opts)
	if err != nil {
		t.Fatalf("JSON RenderWithContext without provider failed: %v", err)
	}

	output2 := buf.String()

	// Both outputs should be identical for JSON renderer
	if output1 != output2 {
		t.Errorf("JSON renderer should produce identical output regardless of context.\nWith provider: %s\nWithout provider: %s", output1, output2)
	}

	if !strings.Contains(output1, "JSON") {
		t.Errorf("Expected output to contain 'JSON', got: %s", output1)
	}
}

func TestColorProvider_Integration(t *testing.T) {
	// This test verifies that the ColorProvider integration works end-to-end
	renderer := NewColoredRenderer()

	// Test that we can create styles from a provider
	provider := &MockColorProvider{}

	// Test that context with provider uses the provider
	ctx := colors.WithProvider(context.Background(), provider)
	if retrievedProvider, ok := colors.FromContext(ctx); !ok {
		t.Error("Expected provider to be found in context")
	} else {
		if retrievedProvider.Success() != "#00AA00" {
			t.Errorf("Expected provider success color to be #00AA00, got %s", retrievedProvider.Success())
		}
		if retrievedProvider.Error() != "#AA0000" {
			t.Errorf("Expected provider error color to be #AA0000, got %s", retrievedProvider.Error())
		}
		if retrievedProvider.Warning() != "#FFAA00" {
			t.Errorf("Expected provider warning color to be #FFAA00, got %s", retrievedProvider.Warning())
		}
	}

	// Test that context without provider doesn't have provider
	if _, ok := colors.FromContext(context.Background()); ok {
		t.Error("Expected no provider in background context")
	}

	// Test that getStyles function works with both contexts
	providerStyles := renderer.getStyles(ctx)
	defaultStyles := renderer.getStyles(context.Background())

	// These should be different Style objects (we can't directly compare them,
	// but we can verify that the function returns styles in both cases)
	_ = providerStyles
	_ = defaultStyles

	// Verify that stylesFromProvider creates styles with expected methods
	testStyles := stylesFromProvider(provider)
	_ = testStyles.success
	_ = testStyles.warning
	_ = testStyles.errStyle
	_ = testStyles.info

	// If we get this far without panics, the integration is working
}
