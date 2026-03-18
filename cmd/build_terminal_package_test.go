package cmd

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	terminalpkg "github.com/rmkohlman/MaestroTerminal/terminalops/package"
	"github.com/rmkohlman/MaestroTerminal/terminalops/prompt"
	promptextension "github.com/rmkohlman/MaestroTerminal/terminalops/prompt/extension"
	promptextensionlibrary "github.com/rmkohlman/MaestroTerminal/terminalops/prompt/extension/library"
	promptstyle "github.com/rmkohlman/MaestroTerminal/terminalops/prompt/style"
	promptstylelibrary "github.com/rmkohlman/MaestroTerminal/terminalops/prompt/style/library"
	"github.com/rmkohlman/MaestroPalette"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==============================================================================
// TDD Phase 2 - RED: Failing Tests for Terminal Package Bug Fix
// ==============================================================================
// These tests expose the bug where `generateShellConfig()` in build.go always
// calls `createDefaultTerminalPrompt()` instead of reading the terminal package
// from the database when `terminal-package` default is set.
//
// GitHub Issue: Terminal Package Not Used During Build
//
// Expected behavior:
// 1. getPromptFromPackageOrDefault() should check for terminal-package default
// 2. If set, read the package and compose prompt from style + extensions
// 3. If not set or package doesn't exist, fall back to createDefaultTerminalPrompt()
// 4. Package must have promptStyle and promptExtensions set (UsesModularPrompt() == true)
// ==============================================================================

// MockTerminalPackageStore is a minimal mock for terminal package storage
type MockTerminalPackageStore struct {
	packages map[string]*terminalpkg.Package
}

func NewMockTerminalPackageStore() *MockTerminalPackageStore {
	return &MockTerminalPackageStore{
		packages: make(map[string]*terminalpkg.Package),
	}
}

func (m *MockTerminalPackageStore) Get(name string) (*terminalpkg.Package, bool) {
	if pkg, ok := m.packages[name]; ok {
		return pkg, true
	}
	return nil, false
}

// MockPromptStyleStore is a minimal mock for prompt style storage
type MockPromptStyleStore struct {
	styles map[string]*promptstyle.PromptStyle
}

func NewMockPromptStyleStore() *MockPromptStyleStore {
	return &MockPromptStyleStore{
		styles: make(map[string]*promptstyle.PromptStyle),
	}
}

func (m *MockPromptStyleStore) Get(name string) (*promptstyle.PromptStyle, error) {
	if style, ok := m.styles[name]; ok {
		return style, nil
	}
	return nil, sql.ErrNoRows
}

// MockPromptExtensionStore is a minimal mock for prompt extension storage
type MockPromptExtensionStore struct {
	extensions map[string]*promptextension.PromptExtension
}

func NewMockPromptExtensionStore() *MockPromptExtensionStore {
	return &MockPromptExtensionStore{
		extensions: make(map[string]*promptextension.PromptExtension),
	}
}

func (m *MockPromptExtensionStore) Get(name string) (*promptextension.PromptExtension, error) {
	if ext, ok := m.extensions[name]; ok {
		return ext, nil
	}
	return nil, sql.ErrNoRows
}

// ==============================================================================
// Test 1: Terminal package is set in defaults and should be used
// ==============================================================================
// When terminal-package default is set to a valid package name that has
// promptStyle and promptExtensions, the prompt should be composed from the
// package, NOT the hardcoded default.

func TestGetPromptFromPackage_WithPackageSet(t *testing.T) {
	// Setup: Database has terminal-package default set to "maestro"
	ds := NewMockDataStoreForBuild()
	ds.defaults["terminal-package"] = "maestro"

	// Create mock stores
	pkgStore := NewMockTerminalPackageStore()
	styleStore := NewMockPromptStyleStore()
	extStore := NewMockPromptExtensionStore()

	// Load embedded library data
	styleLib, err := promptstylelibrary.NewStyleLibrary()
	require.NoError(t, err, "Failed to load style library")
	extLib, err := promptextensionlibrary.NewExtensionLibrary()
	require.NoError(t, err, "Failed to load extension library")

	// Add powerline-segments style to mock store
	powerlineStyle, err := styleLib.Get("powerline-segments")
	require.NoError(t, err)
	styleStore.styles["powerline-segments"] = powerlineStyle

	// Add extensions to mock store
	extensionNames := []string{"kubernetes", "colima", "git-detailed"}
	for _, name := range extensionNames {
		ext, err := extLib.Get(name)
		require.NoError(t, err)
		extStore.extensions[name] = ext
	}

	// Create terminal package "maestro" with modular prompt config
	maestroPkg := &terminalpkg.Package{
		Name:             "maestro",
		Description:      "Rich prompt configuration",
		PromptStyle:      "powerline-segments",
		PromptExtensions: []string{"kubernetes", "colima", "git-detailed"},
		Enabled:          true,
	}
	pkgStore.packages["maestro"] = maestroPkg

	// Verify package uses modular prompt system
	assert.True(t, maestroPkg.UsesModularPrompt(), "Package should use modular prompt system")

	ctx := context.Background()
	appName := "test-app"
	workspaceName := "dev"

	// THIS FUNCTION DOESN'T EXIST YET - Test should fail
	// Expected signature: getPromptFromPackageOrDefault(ctx context.Context, ds db.DataStore, pkgStore, styleStore, extStore, appName, workspaceName string) (*prompt.PromptYAML, error)
	promptYAML, err := getPromptFromPackageOrDefault(ctx, ds, pkgStore, styleStore, extStore, appName, workspaceName)

	require.NoError(t, err)
	require.NotNil(t, promptYAML)

	// Assertions: Verify prompt was composed from package, not default
	assert.Contains(t, promptYAML.Metadata.Name, "dvm-pkg-maestro",
		"Prompt name should indicate it came from package 'maestro'")

	// Verify format is NOT the hardcoded 3-module default
	defaultFormat := `$custom\
$directory\
$git_branch\
$git_status\
$character`
	assert.NotEqual(t, defaultFormat, promptYAML.Spec.Format,
		"Format should NOT be the hardcoded default (should be composed from style+extensions)")

	// Verify format contains kubernetes module (from extension)
	assert.Contains(t, promptYAML.Spec.Format, "$kubernetes",
		"Format should contain kubernetes module from extension")

	// Verify format contains colima module (from extension) - check for either format
	hasColimaSimple := strings.Contains(promptYAML.Spec.Format, "$custom.colima_profile")
	hasColimaBraces := strings.Contains(promptYAML.Spec.Format, "${custom.colima_profile}")
	assert.True(t, hasColimaSimple || hasColimaBraces,
		"Format should contain colima module from extension (found: %s)", promptYAML.Spec.Format)

	// Verify modules config exists for git-detailed
	assert.NotEmpty(t, promptYAML.Spec.Modules,
		"Modules should be configured from composed extensions")
}

// ==============================================================================
// Test 2: No terminal-package default set
// ==============================================================================
// When terminal-package default is NOT set, should fall back to
// createDefaultTerminalPrompt() (existing behavior).

func TestGetPromptFromPackage_NoPackageSet(t *testing.T) {
	// Setup: Database has NO terminal-package default
	ds := NewMockDataStoreForBuild()
	// ds.defaults is empty

	// Create empty mock stores (won't be used)
	pkgStore := NewMockTerminalPackageStore()
	styleStore := NewMockPromptStyleStore()
	extStore := NewMockPromptExtensionStore()

	ctx := context.Background()
	appName := "test-app"
	workspaceName := "dev"

	// Should fall back to default prompt
	promptYAML, err := getPromptFromPackageOrDefault(ctx, ds, pkgStore, styleStore, extStore, appName, workspaceName)

	require.NoError(t, err)
	require.NotNil(t, promptYAML)

	// Assertions: Verify it's the default prompt
	expectedName := "dvm-default-test-app-dev"
	assert.Equal(t, expectedName, promptYAML.Metadata.Name,
		"Should use default prompt name when no package is set")

	// Verify format IS the hardcoded default
	expectedFormat := `$custom\
$directory\
$git_branch\
$git_status\
$character`
	assert.Equal(t, expectedFormat, promptYAML.Spec.Format,
		"Should use hardcoded default format when no package is set")

	// Verify description mentions it's a default
	assert.Contains(t, promptYAML.Metadata.Description, "Default DevOpsMaestro prompt",
		"Description should indicate this is the default prompt")
}

// ==============================================================================
// Test 3: Terminal package is set but doesn't exist in database
// ==============================================================================
// When terminal-package default is set to a non-existent package name,
// should log warning and fall back to createDefaultTerminalPrompt().

func TestGetPromptFromPackage_PackageNotFound(t *testing.T) {
	// Setup: Database has terminal-package default but package doesn't exist
	ds := NewMockDataStoreForBuild()
	ds.defaults["terminal-package"] = "nonexistent-package"

	// Create empty mock stores (package won't be found)
	pkgStore := NewMockTerminalPackageStore()
	styleStore := NewMockPromptStyleStore()
	extStore := NewMockPromptExtensionStore()

	ctx := context.Background()
	appName := "test-app"
	workspaceName := "dev"

	// Should fall back to default prompt
	promptYAML, err := getPromptFromPackageOrDefault(ctx, ds, pkgStore, styleStore, extStore, appName, workspaceName)

	require.NoError(t, err, "Should not error, but fall back gracefully")
	require.NotNil(t, promptYAML)

	// Assertions: Verify it's the default prompt
	expectedName := "dvm-default-test-app-dev"
	assert.Equal(t, expectedName, promptYAML.Metadata.Name,
		"Should use default prompt name when package not found")

	// Verify format IS the hardcoded default
	expectedFormat := `$custom\
$directory\
$git_branch\
$git_status\
$character`
	assert.Equal(t, expectedFormat, promptYAML.Spec.Format,
		"Should use hardcoded default format when package not found")
}

// ==============================================================================
// Test 4: Package exists but doesn't use modular prompt system
// ==============================================================================
// When package exists but doesn't have promptStyle/promptExtensions set
// (i.e., UsesModularPrompt() returns false), should fall back to default.

func TestGetPromptFromPackage_PackageWithoutModularPrompt(t *testing.T) {
	// Setup: Database has terminal-package default set
	ds := NewMockDataStoreForBuild()
	ds.defaults["terminal-package"] = "legacy-package"

	// Create mock stores
	pkgStore := NewMockTerminalPackageStore()
	styleStore := NewMockPromptStyleStore()
	extStore := NewMockPromptExtensionStore()

	// Create terminal package WITHOUT modular prompt config
	legacyPkg := &terminalpkg.Package{
		Name:        "legacy-package",
		Description: "Legacy package without modular prompts",
		Plugins:     []string{"zsh-autosuggestions"},
		// NO PromptStyle or PromptExtensions set
		Enabled: true,
	}
	pkgStore.packages["legacy-package"] = legacyPkg

	// Verify package does NOT use modular prompt system
	assert.False(t, legacyPkg.UsesModularPrompt(),
		"Legacy package should not use modular prompt system")

	ctx := context.Background()
	appName := "test-app"
	workspaceName := "dev"

	// Should fall back to default prompt
	promptYAML, err := getPromptFromPackageOrDefault(ctx, ds, pkgStore, styleStore, extStore, appName, workspaceName)

	require.NoError(t, err, "Should not error, but fall back gracefully")
	require.NotNil(t, promptYAML)

	// Assertions: Verify it's the default prompt
	expectedName := "dvm-default-test-app-dev"
	assert.Equal(t, expectedName, promptYAML.Metadata.Name,
		"Should use default prompt name when package doesn't have modular prompt config")

	// Verify format IS the hardcoded default
	expectedFormat := `$custom\
$directory\
$git_branch\
$git_status\
$character`
	assert.Equal(t, expectedFormat, promptYAML.Spec.Format,
		"Should use hardcoded default format when package doesn't have modular prompt config")
}

// ==============================================================================
// Test 5: Package has style but style doesn't exist in store
// ==============================================================================
// When package references a style that doesn't exist, should fall back to default.

func TestGetPromptFromPackage_StyleNotFound(t *testing.T) {
	// Setup: Database has terminal-package default set
	ds := NewMockDataStoreForBuild()
	ds.defaults["terminal-package"] = "broken-package"

	// Create mock stores
	pkgStore := NewMockTerminalPackageStore()
	styleStore := NewMockPromptStyleStore()
	extStore := NewMockPromptExtensionStore()

	// Create package with non-existent style
	brokenPkg := &terminalpkg.Package{
		Name:             "broken-package",
		Description:      "Package with non-existent style",
		PromptStyle:      "nonexistent-style",
		PromptExtensions: []string{"git-detailed"},
		Enabled:          true,
	}
	pkgStore.packages["broken-package"] = brokenPkg

	ctx := context.Background()
	appName := "test-app"
	workspaceName := "dev"

	// Should fall back to default prompt
	promptYAML, err := getPromptFromPackageOrDefault(ctx, ds, pkgStore, styleStore, extStore, appName, workspaceName)

	require.NoError(t, err, "Should not error, but fall back gracefully")
	require.NotNil(t, promptYAML)

	// Assertions: Verify it's the default prompt
	expectedName := "dvm-default-test-app-dev"
	assert.Equal(t, expectedName, promptYAML.Metadata.Name,
		"Should use default prompt name when style not found")
}

// ==============================================================================
// Test 6: Package has extensions but one extension doesn't exist
// ==============================================================================
// When package references extensions and one doesn't exist, should compose
// with available extensions and log warning about missing one.

func TestGetPromptFromPackage_ExtensionNotFound(t *testing.T) {
	// Setup: Database has terminal-package default set
	ds := NewMockDataStoreForBuild()
	ds.defaults["terminal-package"] = "partial-package"

	// Create mock stores
	pkgStore := NewMockTerminalPackageStore()
	styleStore := NewMockPromptStyleStore()
	extStore := NewMockPromptExtensionStore()

	// Load embedded library data
	styleLib, err := promptstylelibrary.NewStyleLibrary()
	require.NoError(t, err, "Failed to load style library")
	extLib, err := promptextensionlibrary.NewExtensionLibrary()
	require.NoError(t, err, "Failed to load extension library")

	// Add powerline-segments style
	powerlineStyle, err := styleLib.Get("powerline-segments")
	require.NoError(t, err)
	styleStore.styles["powerline-segments"] = powerlineStyle

	// Add only git-detailed extension (missing nonexistent-extension)
	gitExt, err := extLib.Get("git-detailed")
	require.NoError(t, err)
	extStore.extensions["git-detailed"] = gitExt

	// Create package with one valid and one invalid extension
	partialPkg := &terminalpkg.Package{
		Name:             "partial-package",
		Description:      "Package with missing extension",
		PromptStyle:      "powerline-segments",
		PromptExtensions: []string{"git-detailed", "nonexistent-extension"},
		Enabled:          true,
	}
	pkgStore.packages["partial-package"] = partialPkg

	ctx := context.Background()
	appName := "test-app"
	workspaceName := "dev"

	// Should compose with available extensions
	promptYAML, err := getPromptFromPackageOrDefault(ctx, ds, pkgStore, styleStore, extStore, appName, workspaceName)

	require.NoError(t, err, "Should not error, but compose with available extensions")
	require.NotNil(t, promptYAML)

	// Assertions: Verify prompt was composed from package
	assert.Contains(t, promptYAML.Metadata.Name, "dvm-pkg-partial-package",
		"Prompt name should indicate it came from package 'partial-package'")

	// Verify format contains git modules (from valid extension)
	assert.Contains(t, promptYAML.Spec.Format, "$git_branch",
		"Format should contain git modules from valid extension")
}

// ==============================================================================
// Test 7: Integration with palette - verify theme colors are used
// ==============================================================================
// When package is used with modular prompts, verify that theme colors
// are still applied correctly via the palette parameter.

func TestGetPromptFromPackage_WithThemePalette(t *testing.T) {
	// This test verifies that the composed prompt from package
	// still works with the theme palette system.

	// Setup similar to Test 1
	ds := NewMockDataStoreForBuild()
	ds.defaults["terminal-package"] = "maestro"

	pkgStore := NewMockTerminalPackageStore()
	styleStore := NewMockPromptStyleStore()
	extStore := NewMockPromptExtensionStore()

	// Load embedded libraries
	styleLib, err := promptstylelibrary.NewStyleLibrary()
	require.NoError(t, err, "Failed to load style library")
	extLib, err := promptextensionlibrary.NewExtensionLibrary()
	require.NoError(t, err, "Failed to load extension library")

	powerlineStyle, err := styleLib.Get("powerline-segments")
	require.NoError(t, err)
	styleStore.styles["powerline-segments"] = powerlineStyle

	gitExt, err := extLib.Get("git-detailed")
	require.NoError(t, err)
	extStore.extensions["git-detailed"] = gitExt

	maestroPkg := &terminalpkg.Package{
		Name:             "maestro",
		PromptStyle:      "powerline-segments",
		PromptExtensions: []string{"git-detailed"},
		Enabled:          true,
	}
	pkgStore.packages["maestro"] = maestroPkg

	ctx := context.Background()
	appName := "test-app"
	workspaceName := "dev"

	// Get composed prompt
	promptYAML, err := getPromptFromPackageOrDefault(ctx, ds, pkgStore, styleStore, extStore, appName, workspaceName)
	require.NoError(t, err)
	require.NotNil(t, promptYAML)

	// Verify palette reference is set to "theme" (for theme color interpolation)
	assert.Equal(t, "theme", promptYAML.Spec.Palette,
		"Palette should be set to 'theme' for color interpolation")

	// Verify format contains palette color references (not ${theme.} - that's added during rendering)
	// The composer uses palette keys like 'sky', 'sapphire', etc.
	hasPaletteKey := strings.Contains(promptYAML.Spec.Format, "(sky)") ||
		strings.Contains(promptYAML.Spec.Format, "(sapphire)") ||
		strings.Contains(promptYAML.Spec.Format, "bg:") ||
		strings.Contains(promptYAML.Spec.Format, "fg:")
	assert.True(t, hasPaletteKey,
		"Format should contain palette color keys for theme interpolation")
}

// ==============================================================================
// Test 8: Verify prompt name format for package-based prompts
// ==============================================================================
// Package-based prompts should have a distinct name format to differentiate
// them from manually created prompts and default prompts.

func TestGetPromptFromPackage_PromptNamingConvention(t *testing.T) {
	tests := []struct {
		name            string
		packageName     string
		appName         string
		workspaceName   string
		wantNamePattern string
	}{
		{
			name:            "standard package prompt",
			packageName:     "maestro",
			appName:         "test-app",
			workspaceName:   "dev",
			wantNamePattern: "dvm-pkg-maestro-test-app-dev",
		},
		{
			name:            "package with hyphens",
			packageName:     "dev-essentials",
			appName:         "api-service",
			workspaceName:   "staging",
			wantNamePattern: "dvm-pkg-dev-essentials-api-service-staging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ds := NewMockDataStoreForBuild()
			ds.defaults["terminal-package"] = tt.packageName

			pkgStore := NewMockTerminalPackageStore()
			styleStore := NewMockPromptStyleStore()
			extStore := NewMockPromptExtensionStore()

			// Load embedded libraries
			styleLib, err := promptstylelibrary.NewStyleLibrary()
			require.NoError(t, err, "Failed to load style library")
			extLib, err := promptextensionlibrary.NewExtensionLibrary()
			require.NoError(t, err, "Failed to load extension library")

			powerlineStyle, err := styleLib.Get("powerline-segments")
			require.NoError(t, err)
			styleStore.styles["powerline-segments"] = powerlineStyle

			gitExt, err := extLib.Get("git-detailed")
			require.NoError(t, err)
			extStore.extensions["git-detailed"] = gitExt

			pkg := &terminalpkg.Package{
				Name:             tt.packageName,
				PromptStyle:      "powerline-segments",
				PromptExtensions: []string{"git-detailed"},
				Enabled:          true,
			}
			pkgStore.packages[tt.packageName] = pkg

			ctx := context.Background()

			// Execute
			promptYAML, err := getPromptFromPackageOrDefault(ctx, ds, pkgStore, styleStore, extStore, tt.appName, tt.workspaceName)

			// Assert
			require.NoError(t, err)
			require.NotNil(t, promptYAML)
			assert.Equal(t, tt.wantNamePattern, promptYAML.Metadata.Name,
				"Prompt name should follow package naming convention")
		})
	}
}

// ==============================================================================
// Test 9: Verify palette parameter passed to renderer
// ==============================================================================
// After getting the prompt (from package or default), the palette should
// still be passed to the renderer for color interpolation.

func TestRenderPromptWithPalette(t *testing.T) {
	// This is a documentation test to verify the expected workflow:
	// 1. getPromptFromPackageOrDefault() returns *prompt.PromptYAML
	// 2. Caller resolves theme to *palette.Palette (existing code in generateShellConfig)
	// 3. Caller passes both to renderer.RenderToFile(promptYAML, themePalette, path)

	// Create a sample prompt
	appName := "test-app"
	workspaceName := "dev"
	promptYAML := createDefaultTerminalPrompt(appName, workspaceName)

	// Create a sample palette
	themePalette := &palette.Palette{
		Name:        "test-theme",
		Description: "Test theme",
		Category:    palette.CategoryDark,
		Colors: map[string]string{
			"red":   "#ff0000",
			"green": "#00ff00",
			"blue":  "#0000ff",
			"cyan":  "#00ffff",
		},
	}

	// Verify renderer interface exists and accepts these types
	renderer := prompt.NewRenderer()
	require.NotNil(t, renderer, "Renderer should be created successfully")

	// Note: We don't actually render to a file in this test,
	// just verify the types are compatible.
	// The actual rendering is tested in renderer_test.go

	// Verify prompt has palette reference
	assert.NotEmpty(t, promptYAML.Spec.Palette,
		"Prompt should have palette reference for renderer")

	// Verify palette has required colors
	assert.NotEmpty(t, themePalette.Colors,
		"Palette should have colors for interpolation")
}
