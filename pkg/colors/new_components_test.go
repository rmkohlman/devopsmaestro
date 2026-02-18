package colors_test

import (
	"context"
	"os"
	"testing"

	"devopsmaestro/pkg/colors"
	"devopsmaestro/pkg/nvimops/theme"
	"devopsmaestro/pkg/palette"
)

// =============================================================================
// NoColorProvider Tests
// =============================================================================

func TestNoColorProvider_AllMethodsReturnEmptyString(t *testing.T) {
	provider := colors.NewNoColorProvider()

	// Test all color methods return empty strings
	methods := []struct {
		name   string
		method func() string
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
			result := m.method()
			if result != "" {
				t.Errorf("%s() = %q, want empty string", m.name, result)
			}
		})
	}
}

func TestNoColorProvider_Metadata(t *testing.T) {
	provider := colors.NewNoColorProvider()

	if provider.Name() != "no-color" {
		t.Errorf("Name() = %q, want %q", provider.Name(), "no-color")
	}

	if provider.IsLight() != false {
		t.Errorf("IsLight() = %v, want false", provider.IsLight())
	}
}

func TestNoColorProvider_ImplementsInterface(t *testing.T) {
	// Ensure NoColorProvider satisfies the ColorProvider interface
	var _ colors.ColorProvider = colors.NewNoColorProvider()
}

// =============================================================================
// ThemeStoreAdapter Tests
// =============================================================================

// mockThemeStore implements theme.Store for testing
type mockThemeStore struct {
	activeTheme *theme.Theme
	themes      map[string]*theme.Theme
	activeError error
	getError    error
	basePath    string
}

func (m *mockThemeStore) Path() string {
	return m.basePath
}

func (m *mockThemeStore) GetActive() (*theme.Theme, error) {
	if m.activeError != nil {
		return nil, m.activeError
	}
	return m.activeTheme, nil
}

func (m *mockThemeStore) Get(name string) (*theme.Theme, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	t, ok := m.themes[name]
	if !ok {
		return nil, &colors.NoProviderError{}
	}
	return t, nil
}

func (m *mockThemeStore) List() ([]*theme.Theme, error) {
	var themes []*theme.Theme
	for _, t := range m.themes {
		themes = append(themes, t)
	}
	return themes, nil
}

func (m *mockThemeStore) Save(theme *theme.Theme) error {
	return nil
}

func (m *mockThemeStore) Delete(name string) error {
	return nil
}

func (m *mockThemeStore) SetActive(name string) error {
	return nil
}

func (m *mockThemeStore) ClearActive() error {
	return nil
}

func TestThemeStoreAdapter_GetActivePalette(t *testing.T) {
	// Test with active theme
	activeTheme := &theme.Theme{
		Name: "test-active",
		Colors: map[string]string{
			palette.ColorPrimary: "#123456",
		},
	}

	store := &mockThemeStore{
		activeTheme: activeTheme,
		basePath:    "/test",
	}
	adapter := colors.NewThemeStoreAdapter(store)

	result, err := adapter.GetActivePalette()
	if err != nil {
		t.Errorf("GetActivePalette() error = %v, want nil", err)
	}
	if result.Name != "test-active" {
		t.Errorf("GetActivePalette().Name = %q, want %q", result.Name, "test-active")
	}
	if result.Colors[palette.ColorPrimary] != "#123456" {
		t.Errorf("GetActivePalette().Colors[primary] = %q, want %q",
			result.Colors[palette.ColorPrimary], "#123456")
	}
}

func TestThemeStoreAdapter_GetActivePalette_NoActive(t *testing.T) {
	store := &mockThemeStore{
		activeTheme: nil, // No active theme
		basePath:    "/test",
	}
	adapter := colors.NewThemeStoreAdapter(store)

	result, err := adapter.GetActivePalette()
	if err != nil {
		t.Errorf("GetActivePalette() error = %v, want nil", err)
	}
	if result != nil {
		t.Errorf("GetActivePalette() = %v, want nil when no active theme", result)
	}
}

func TestThemeStoreAdapter_GetActivePalette_NilStore(t *testing.T) {
	adapter := colors.NewThemeStoreAdapter(nil)

	result, err := adapter.GetActivePalette()
	if err != nil {
		t.Errorf("GetActivePalette() error = %v, want nil", err)
	}
	if result != nil {
		t.Errorf("GetActivePalette() = %v, want nil with nil store", result)
	}
}

func TestThemeStoreAdapter_GetPalette(t *testing.T) {
	testTheme := &theme.Theme{
		Name: "gruvbox",
		Colors: map[string]string{
			palette.ColorFg: "#ebdbb2",
			palette.ColorBg: "#282828",
		},
	}

	store := &mockThemeStore{
		themes: map[string]*theme.Theme{
			"gruvbox": testTheme,
		},
		basePath: "/test",
	}
	adapter := colors.NewThemeStoreAdapter(store)

	result, err := adapter.GetPalette("gruvbox")
	if err != nil {
		t.Errorf("GetPalette() error = %v, want nil", err)
	}
	if result.Name != "gruvbox" {
		t.Errorf("GetPalette().Name = %q, want %q", result.Name, "gruvbox")
	}
	if result.Colors[palette.ColorFg] != "#ebdbb2" {
		t.Errorf("GetPalette().Colors[fg] = %q, want %q",
			result.Colors[palette.ColorFg], "#ebdbb2")
	}
}

func TestThemeStoreAdapter_GetPalette_NotFound(t *testing.T) {
	store := &mockThemeStore{
		themes:   map[string]*theme.Theme{},
		basePath: "/test",
	}
	adapter := colors.NewThemeStoreAdapter(store)

	_, err := adapter.GetPalette("nonexistent")
	if err == nil {
		t.Error("GetPalette() expected error for nonexistent theme")
	}
}

func TestThemeStoreAdapter_GetPalette_NilStore(t *testing.T) {
	adapter := colors.NewThemeStoreAdapter(nil)

	_, err := adapter.GetPalette("any")
	if err == nil {
		t.Error("GetPalette() expected error with nil store")
	}
}

func TestThemeStoreAdapter_ImplementsInterface(t *testing.T) {
	// Ensure ThemeStoreAdapter satisfies the PaletteProvider interface
	store := &mockThemeStore{basePath: "/test"}
	var _ colors.PaletteProvider = colors.NewThemeStoreAdapter(store)
}

// =============================================================================
// Command Helper Tests
// =============================================================================

func TestInitColorProviderForCommand_NoColor(t *testing.T) {
	ctx := context.Background()

	// Test with noColor flag
	resultCtx, err := colors.InitColorProviderForCommand(ctx, "", true)
	if err != nil {
		t.Errorf("InitColorProviderForCommand() error = %v, want nil", err)
	}

	provider, ok := colors.FromContext(resultCtx)
	if !ok {
		t.Error("InitColorProviderForCommand() should inject provider into context")
	}
	if provider.Name() != "no-color" {
		t.Errorf("Provider.Name() = %q, want %q", provider.Name(), "no-color")
	}
}

func TestInitColorProviderForCommand_NoColorEnvVar(t *testing.T) {
	// Set NO_COLOR environment variable
	oldValue := os.Getenv("NO_COLOR")
	os.Setenv("NO_COLOR", "1")
	defer func() {
		if oldValue == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", oldValue)
		}
	}()

	ctx := context.Background()
	resultCtx, err := colors.InitColorProviderForCommand(ctx, "", false)
	if err != nil {
		t.Errorf("InitColorProviderForCommand() error = %v, want nil", err)
	}

	provider, ok := colors.FromContext(resultCtx)
	if !ok {
		t.Error("InitColorProviderForCommand() should inject provider into context")
	}
	if provider.Name() != "no-color" {
		t.Errorf("Provider.Name() = %q, want %q", provider.Name(), "no-color")
	}
}

func TestInitColorProviderForCommand_NoThemePath(t *testing.T) {
	ctx := context.Background()
	resultCtx, err := colors.InitColorProviderForCommand(ctx, "", false)
	if err != nil {
		t.Errorf("InitColorProviderForCommand() error = %v, want nil", err)
	}

	provider, ok := colors.FromContext(resultCtx)
	if !ok {
		t.Error("InitColorProviderForCommand() should inject provider into context")
	}
	if provider.Name() != "default" {
		t.Errorf("Provider.Name() = %q, want %q", provider.Name(), "default")
	}
}

func TestGetDefaultThemePath(t *testing.T) {
	// Save original environment
	originalDvmPath := os.Getenv("DVM_THEME_PATH")
	originalXdgConfig := os.Getenv("XDG_CONFIG_HOME")
	originalHome := os.Getenv("HOME")

	defer func() {
		// Restore environment
		if originalDvmPath == "" {
			os.Unsetenv("DVM_THEME_PATH")
		} else {
			os.Setenv("DVM_THEME_PATH", originalDvmPath)
		}
		if originalXdgConfig == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", originalXdgConfig)
		}
		if originalHome == "" {
			os.Unsetenv("HOME")
		} else {
			os.Setenv("HOME", originalHome)
		}
	}()

	// Test DVM_THEME_PATH takes precedence
	os.Setenv("DVM_THEME_PATH", "/custom/theme/path")
	path := colors.GetDefaultThemePath()
	if path != "/custom/theme/path" {
		t.Errorf("GetDefaultThemePath() = %q, want %q", path, "/custom/theme/path")
	}

	// Test XDG_CONFIG_HOME
	os.Unsetenv("DVM_THEME_PATH")
	os.Setenv("XDG_CONFIG_HOME", "/xdg/config")
	path = colors.GetDefaultThemePath()
	expected := "/xdg/config/dvm/themes"
	if path != expected {
		t.Errorf("GetDefaultThemePath() = %q, want %q", path, expected)
	}

	// Test HOME fallback
	os.Unsetenv("XDG_CONFIG_HOME")
	os.Setenv("HOME", "/home/user")
	path = colors.GetDefaultThemePath()
	expected = "/home/user/.config/dvm/themes"
	if path != expected {
		t.Errorf("GetDefaultThemePath() = %q, want %q", path, expected)
	}

	// Test last resort
	os.Unsetenv("HOME")
	path = colors.GetDefaultThemePath()
	expected = "./themes"
	if path != expected {
		t.Errorf("GetDefaultThemePath() = %q, want %q", path, expected)
	}
}

func TestIsNoColorRequested(t *testing.T) {
	// Save original environment
	originalNoColor := os.Getenv("NO_COLOR")
	originalTerm := os.Getenv("TERM")

	defer func() {
		// Restore environment
		if originalNoColor == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", originalNoColor)
		}
		if originalTerm == "" {
			os.Unsetenv("TERM")
		} else {
			os.Setenv("TERM", originalTerm)
		}
	}()

	tests := []struct {
		name        string
		noColorFlag bool
		noColorEnv  string
		termEnv     string
		expected    bool
	}{
		{"flag true", true, "", "", true},
		{"NO_COLOR set", false, "1", "", true},
		{"TERM=dumb", false, "", "dumb", true},
		{"all false", false, "", "xterm", false},
		{"multiple true", true, "1", "dumb", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment
			if tt.noColorEnv == "" {
				os.Unsetenv("NO_COLOR")
			} else {
				os.Setenv("NO_COLOR", tt.noColorEnv)
			}
			if tt.termEnv == "" {
				os.Unsetenv("TERM")
			} else {
				os.Setenv("TERM", tt.termEnv)
			}

			result := colors.IsNoColorRequested(tt.noColorFlag)
			if result != tt.expected {
				t.Errorf("IsNoColorRequested(%v) = %v, want %v",
					tt.noColorFlag, result, tt.expected)
			}
		})
	}
}
