package colors_test

import (
	"context"
	"testing"

	"devopsmaestro/pkg/colors"
	"devopsmaestro/pkg/palette"
)

// =============================================================================
// ColorProvider Interface Tests
// =============================================================================

func TestColorProviderInterface(t *testing.T) {
	// Ensure all implementations satisfy the interface
	var _ colors.ColorProvider = &colors.MockColorProvider{}
	var _ colors.ColorProvider = colors.NewDefaultColorProvider()
	var _ colors.ColorProvider = colors.NewDefaultLightColorProvider()
}

// =============================================================================
// ThemeColorProvider Tests
// =============================================================================

func TestThemeColorProvider_FromPalette(t *testing.T) {
	p := &palette.Palette{
		Name:     "test-theme",
		Category: palette.CategoryDark,
		Colors: map[string]string{
			palette.ColorFg:          "#ffffff",
			palette.ColorBg:          "#000000",
			palette.ColorPrimary:     "#0000ff",
			palette.ColorSecondary:   "#ff00ff",
			palette.ColorAccent:      "#00ffff",
			palette.ColorSuccess:     "#00ff00",
			palette.ColorWarning:     "#ffff00",
			palette.ColorError:       "#ff0000",
			palette.ColorInfo:        "#00aaff",
			palette.ColorComment:     "#888888",
			palette.ColorBgHighlight: "#333333",
			palette.ColorBorder:      "#444444",
		},
	}

	provider := colors.NewThemeColorProvider(p)

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Foreground", provider.Foreground(), "#ffffff"},
		{"Background", provider.Background(), "#000000"},
		{"Primary", provider.Primary(), "#0000ff"},
		{"Secondary", provider.Secondary(), "#ff00ff"},
		{"Accent", provider.Accent(), "#00ffff"},
		{"Success", provider.Success(), "#00ff00"},
		{"Warning", provider.Warning(), "#ffff00"},
		{"Error", provider.Error(), "#ff0000"},
		{"Info", provider.Info(), "#00aaff"},
		{"Muted", provider.Muted(), "#888888"},
		{"Highlight", provider.Highlight(), "#333333"},
		{"Border", provider.Border(), "#444444"},
		{"Name", provider.Name(), "test-theme"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s() = %q, want %q", tt.name, tt.got, tt.expected)
			}
		})
	}

	if provider.IsLight() {
		t.Error("IsLight() = true, want false for dark theme")
	}
}

func TestThemeColorProvider_LightTheme(t *testing.T) {
	p := &palette.Palette{
		Name:     "light-theme",
		Category: palette.CategoryLight,
		Colors: map[string]string{
			palette.ColorFg: "#000000",
			palette.ColorBg: "#ffffff",
		},
	}

	provider := colors.NewThemeColorProvider(p)

	if !provider.IsLight() {
		t.Error("IsLight() = false, want true for light theme")
	}
	if provider.Name() != "light-theme" {
		t.Errorf("Name() = %q, want %q", provider.Name(), "light-theme")
	}
}

func TestThemeColorProvider_NilPalette(t *testing.T) {
	provider := colors.NewThemeColorProvider(nil)

	// Should return default values, not panic
	if provider.Name() != "default" {
		t.Errorf("Name() = %q, want %q", provider.Name(), "default")
	}
	if provider.Foreground() == "" {
		t.Error("Foreground() returned empty string, want default color")
	}
}

func TestThemeColorProvider_Fallbacks(t *testing.T) {
	// Empty palette - should use fallbacks
	p := &palette.Palette{
		Name:     "minimal",
		Category: palette.CategoryDark,
		Colors:   map[string]string{},
	}

	provider := colors.NewThemeColorProvider(p)

	// All methods should return default values
	if provider.Primary() == "" {
		t.Error("Primary() returned empty string, want fallback")
	}
	if provider.Success() == "" {
		t.Error("Success() returned empty string, want fallback")
	}
	if provider.Foreground() == "" {
		t.Error("Foreground() returned empty string, want fallback")
	}
}

// =============================================================================
// Default Color Provider Tests
// =============================================================================

func TestDefaultColorProvider(t *testing.T) {
	provider := colors.NewDefaultColorProvider()

	if provider.IsLight() {
		t.Error("IsLight() = true, want false for dark default")
	}
	if provider.Name() != "default" {
		t.Errorf("Name() = %q, want %q", provider.Name(), "default")
	}

	// Verify all colors are set
	methods := []struct {
		name string
		fn   func() string
	}{
		{"Primary", provider.Primary},
		{"Secondary", provider.Secondary},
		{"Accent", provider.Accent},
		{"Success", provider.Success},
		{"Warning", provider.Warning},
		{"Error", provider.Error},
		{"Info", provider.Info},
		{"Foreground", provider.Foreground},
		{"Background", provider.Background},
		{"Muted", provider.Muted},
		{"Highlight", provider.Highlight},
		{"Border", provider.Border},
	}

	for _, m := range methods {
		t.Run(m.name, func(t *testing.T) {
			color := m.fn()
			if color == "" {
				t.Errorf("%s() returned empty string", m.name)
			}
			if color[0] != '#' {
				t.Errorf("%s() = %q, want hex color starting with #", m.name, color)
			}
		})
	}
}

func TestDefaultLightColorProvider(t *testing.T) {
	provider := colors.NewDefaultLightColorProvider()

	if !provider.IsLight() {
		t.Error("IsLight() = false, want true for light default")
	}
	if provider.Name() != "default-light" {
		t.Errorf("Name() = %q, want %q", provider.Name(), "default-light")
	}

	// Light theme should have different colors than dark
	darkProvider := colors.NewDefaultColorProvider()
	if provider.Background() == darkProvider.Background() {
		t.Error("Light and dark providers should have different backgrounds")
	}
}

// =============================================================================
// Factory Tests
// =============================================================================

func TestProviderFactory_CreateDefault(t *testing.T) {
	factory := colors.NewProviderFactory(nil)

	dark := factory.CreateDefault(false)
	if dark.IsLight() {
		t.Error("CreateDefault(false) should return dark provider")
	}

	light := factory.CreateDefault(true)
	if !light.IsLight() {
		t.Error("CreateDefault(true) should return light provider")
	}
}

func TestProviderFactory_CreateFromActive_NoProvider(t *testing.T) {
	factory := colors.NewProviderFactory(nil)

	provider, err := factory.CreateFromActive()
	if err != nil {
		t.Errorf("CreateFromActive() error = %v, want nil", err)
	}
	if provider == nil {
		t.Error("CreateFromActive() returned nil provider")
	}
	// Should return default when no palette provider
	if provider.Name() != "default" {
		t.Errorf("Name() = %q, want %q", provider.Name(), "default")
	}
}

// mockPaletteProvider implements PaletteProvider for testing
type mockPaletteProvider struct {
	activePalette *palette.Palette
	palettes      map[string]*palette.Palette
	activeErr     error
	getErr        error
}

func (m *mockPaletteProvider) GetActivePalette() (*palette.Palette, error) {
	if m.activeErr != nil {
		return nil, m.activeErr
	}
	return m.activePalette, nil
}

func (m *mockPaletteProvider) GetPalette(name string) (*palette.Palette, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	p, ok := m.palettes[name]
	if !ok {
		return nil, &colors.NoProviderError{}
	}
	return p, nil
}

func TestProviderFactory_CreateFromActive_WithProvider(t *testing.T) {
	activePalette := &palette.Palette{
		Name:     "tokyo-night",
		Category: palette.CategoryDark,
		Colors: map[string]string{
			palette.ColorFg: "#c0caf5",
			palette.ColorBg: "#1a1b26",
		},
	}

	pp := &mockPaletteProvider{
		activePalette: activePalette,
	}

	factory := colors.NewProviderFactory(pp)

	provider, err := factory.CreateFromActive()
	if err != nil {
		t.Errorf("CreateFromActive() error = %v, want nil", err)
	}
	if provider.Name() != "tokyo-night" {
		t.Errorf("Name() = %q, want %q", provider.Name(), "tokyo-night")
	}
}

func TestProviderFactory_CreateFromTheme(t *testing.T) {
	gruvbox := &palette.Palette{
		Name:     "gruvbox",
		Category: palette.CategoryDark,
		Colors: map[string]string{
			palette.ColorFg: "#ebdbb2",
			palette.ColorBg: "#282828",
		},
	}

	pp := &mockPaletteProvider{
		palettes: map[string]*palette.Palette{
			"gruvbox": gruvbox,
		},
	}

	factory := colors.NewProviderFactory(pp)

	provider, err := factory.CreateFromTheme("gruvbox")
	if err != nil {
		t.Errorf("CreateFromTheme() error = %v, want nil", err)
	}
	if provider.Name() != "gruvbox" {
		t.Errorf("Name() = %q, want %q", provider.Name(), "gruvbox")
	}
}

func TestProviderFactory_CreateFromTheme_NotFound(t *testing.T) {
	pp := &mockPaletteProvider{
		palettes: map[string]*palette.Palette{},
	}

	factory := colors.NewProviderFactory(pp)

	_, err := factory.CreateFromTheme("nonexistent")
	if err == nil {
		t.Error("CreateFromTheme() expected error for nonexistent theme")
	}
}

// =============================================================================
// Static Factory Functions Tests
// =============================================================================

func TestFromPalette(t *testing.T) {
	p := &palette.Palette{
		Name: "static-test",
		Colors: map[string]string{
			palette.ColorPrimary: "#abcdef",
		},
	}

	provider := colors.FromPalette(p)
	if provider.Name() != "static-test" {
		t.Errorf("Name() = %q, want %q", provider.Name(), "static-test")
	}
	if provider.Primary() != "#abcdef" {
		t.Errorf("Primary() = %q, want %q", provider.Primary(), "#abcdef")
	}
}

func TestDefault(t *testing.T) {
	provider := colors.Default()
	if provider.IsLight() {
		t.Error("Default() should return dark provider")
	}
}

func TestDefaultLight(t *testing.T) {
	provider := colors.DefaultLight()
	if !provider.IsLight() {
		t.Error("DefaultLight() should return light provider")
	}
}

// =============================================================================
// Context Tests
// =============================================================================

func TestContext_WithProvider(t *testing.T) {
	mock := colors.NewMockColorProvider(colors.WithMockName("ctx-test"))
	ctx := colors.WithProvider(context.Background(), mock)

	provider, ok := colors.FromContext(ctx)
	if !ok {
		t.Error("FromContext() = false, want true")
	}
	if provider.Name() != "ctx-test" {
		t.Errorf("Name() = %q, want %q", provider.Name(), "ctx-test")
	}
}

func TestContext_FromContext_NotFound(t *testing.T) {
	ctx := context.Background()

	provider, ok := colors.FromContext(ctx)
	if ok {
		t.Error("FromContext() = true, want false for empty context")
	}
	if provider != nil {
		t.Error("provider should be nil when not found")
	}
}

func TestContext_MustFromContext_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustFromContext() should panic when provider not in context")
		}
	}()

	ctx := context.Background()
	colors.MustFromContext(ctx)
}

func TestContext_MustFromContext_Success(t *testing.T) {
	mock := colors.NewMockColorProvider()
	ctx := colors.WithProvider(context.Background(), mock)

	// Should not panic
	provider := colors.MustFromContext(ctx)
	if provider == nil {
		t.Error("MustFromContext() returned nil")
	}
}

func TestContext_FromContextOrDefault(t *testing.T) {
	// Without provider in context
	ctx := context.Background()
	provider := colors.FromContextOrDefault(ctx)
	if provider == nil {
		t.Error("FromContextOrDefault() returned nil")
	}
	if provider.Name() != "default" {
		t.Errorf("Name() = %q, want %q", provider.Name(), "default")
	}

	// With provider in context
	mock := colors.NewMockColorProvider(colors.WithMockName("custom"))
	ctx = colors.WithProvider(ctx, mock)
	provider = colors.FromContextOrDefault(ctx)
	if provider.Name() != "custom" {
		t.Errorf("Name() = %q, want %q", provider.Name(), "custom")
	}
}

func TestContext_FromContextOrDefaultLight(t *testing.T) {
	ctx := context.Background()
	provider := colors.FromContextOrDefaultLight(ctx)
	if !provider.IsLight() {
		t.Error("FromContextOrDefaultLight() should return light provider when not in context")
	}
}

func TestContext_HasProvider(t *testing.T) {
	ctx := context.Background()
	if colors.HasProvider(ctx) {
		t.Error("HasProvider() = true, want false for empty context")
	}

	mock := colors.NewMockColorProvider()
	ctx = colors.WithProvider(ctx, mock)
	if !colors.HasProvider(ctx) {
		t.Error("HasProvider() = false, want true after WithProvider")
	}
}

// =============================================================================
// Mock Provider Tests
// =============================================================================

func TestMockColorProvider_Defaults(t *testing.T) {
	mock := colors.NewMockColorProvider()

	if mock.Name() != "mock-theme" {
		t.Errorf("Name() = %q, want %q", mock.Name(), "mock-theme")
	}
	if mock.IsLight() {
		t.Error("IsLight() = true, want false for default mock")
	}
	if mock.Primary() == "" {
		t.Error("Primary() returned empty string")
	}
}

func TestMockColorProvider_WithOptions(t *testing.T) {
	mock := colors.NewMockColorProvider(
		colors.WithMockName("custom-mock"),
		colors.WithMockLight(),
		colors.WithMockColor("primary", "#123456"),
	)

	if mock.Name() != "custom-mock" {
		t.Errorf("Name() = %q, want %q", mock.Name(), "custom-mock")
	}
	if !mock.IsLight() {
		t.Error("IsLight() = false, want true after WithMockLight()")
	}
	if mock.Primary() != "#123456" {
		t.Errorf("Primary() = %q, want %q", mock.Primary(), "#123456")
	}
}

func TestMockColorProvider_WithMockColors(t *testing.T) {
	customColors := map[string]string{
		"primary":   "#aaaaaa",
		"secondary": "#bbbbbb",
	}

	mock := colors.NewMockColorProvider(colors.WithMockColors(customColors))

	if mock.Primary() != "#aaaaaa" {
		t.Errorf("Primary() = %q, want %q", mock.Primary(), "#aaaaaa")
	}
	if mock.Secondary() != "#bbbbbb" {
		t.Errorf("Secondary() = %q, want %q", mock.Secondary(), "#bbbbbb")
	}
}

func TestMockColorProvider_SetColor(t *testing.T) {
	mock := colors.NewMockColorProvider()
	original := mock.Success()

	mock.SetColor("success", "#fedcba")

	if mock.Success() == original {
		t.Error("SetColor() did not change the color")
	}
	if mock.Success() != "#fedcba" {
		t.Errorf("Success() = %q, want %q", mock.Success(), "#fedcba")
	}
}

func TestMockColorProvider_GetAllColors(t *testing.T) {
	mock := colors.NewMockColorProvider()
	allColors := mock.GetAllColors()

	expectedKeys := []string{
		"primary", "secondary", "accent",
		"success", "warning", "error", "info",
		"foreground", "background", "muted", "highlight", "border",
	}

	for _, key := range expectedKeys {
		if _, ok := allColors[key]; !ok {
			t.Errorf("GetAllColors() missing key %q", key)
		}
	}

	// Verify it's a copy (mutation doesn't affect original)
	allColors["primary"] = "#000000"
	if mock.Primary() == "#000000" {
		t.Error("GetAllColors() should return a copy, not the original map")
	}
}
