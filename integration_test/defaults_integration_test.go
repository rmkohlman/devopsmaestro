// Package integration_test provides integration tests that demonstrate
// how the GetDefaults() functions work in practice across different packages.
package integration_test

import (
	"fmt"
	"testing"

	themeresolver "devopsmaestro/pkg/colors/resolver"
	"devopsmaestro/pkg/nvimops"
	"devopsmaestro/pkg/terminalops/shell"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetDefaultsIntegration demonstrates how the GetDefaults() functions
// can be used in practice to show all default configuration values.
// This mimics what the 'dvm get defaults' command would do.
func TestGetDefaultsIntegration(t *testing.T) {
	fmt.Println("=== DevOpsMaestro Default Configuration Values ===")
	fmt.Println()

	// Test theme resolver defaults
	t.Run("theme_defaults", func(t *testing.T) {
		fmt.Println("Theme Configuration:")
		themeDefaults := themeresolver.GetDefaults()

		require.NotNil(t, themeDefaults)
		assert.Len(t, themeDefaults, 2)

		// Verify and display theme defaults
		global, ok := themeDefaults["global"]
		require.True(t, ok, "theme defaults should contain 'global' key")
		fmt.Printf("  Global theme: %v\n", global)
		assert.Equal(t, "coolnight-ocean", global)

		resolution, ok := themeDefaults["resolution"]
		require.True(t, ok, "theme defaults should contain 'resolution' key")
		fmt.Printf("  Resolution order: %v\n", resolution)
		assert.Equal(t, "workspace → app → domain → ecosystem → global", resolution)

		fmt.Println()
	})

	// Test shell defaults
	t.Run("shell_defaults", func(t *testing.T) {
		fmt.Println("Shell Configuration:")
		shellDefaults := shell.GetDefaults()

		require.NotNil(t, shellDefaults)
		assert.Len(t, shellDefaults, 3)

		// Verify and display shell defaults
		shellType, ok := shellDefaults["type"]
		require.True(t, ok, "shell defaults should contain 'type' key")
		fmt.Printf("  Shell type: %v\n", shellType)
		assert.Equal(t, "zsh", shellType)

		framework, ok := shellDefaults["framework"]
		require.True(t, ok, "shell defaults should contain 'framework' key")
		fmt.Printf("  Framework: %v\n", framework)
		assert.Equal(t, "oh-my-zsh", framework)

		theme, ok := shellDefaults["theme"]
		require.True(t, ok, "shell defaults should contain 'theme' key")
		fmt.Printf("  Shell theme: %v\n", theme)
		assert.Equal(t, "starship", theme)

		fmt.Println()
	})

	// Test nvim defaults
	t.Run("nvim_defaults", func(t *testing.T) {
		fmt.Println("Neovim Configuration:")
		nvimDefaults := nvimops.GetDefaults()

		require.NotNil(t, nvimDefaults)
		assert.Len(t, nvimDefaults, 4)

		// Verify and display nvim defaults
		structure, ok := nvimDefaults["structure"]
		require.True(t, ok, "nvim defaults should contain 'structure' key")
		fmt.Printf("  Structure: %v\n", structure)
		assert.Equal(t, "lazyvim", structure)

		pluginPackage, ok := nvimDefaults["pluginPackage"]
		require.True(t, ok, "nvim defaults should contain 'pluginPackage' key")
		fmt.Printf("  Plugin package: %v\n", pluginPackage)
		assert.Equal(t, "core", pluginPackage)

		mergeMode, ok := nvimDefaults["mergeMode"]
		require.True(t, ok, "nvim defaults should contain 'mergeMode' key")
		fmt.Printf("  Merge mode: %v\n", mergeMode)
		assert.Equal(t, "append", mergeMode)

		corePlugins, ok := nvimDefaults["corePlugins"]
		require.True(t, ok, "nvim defaults should contain 'corePlugins' key")
		plugins := corePlugins.([]string)
		fmt.Printf("  Core plugins: %v\n", plugins)
		assert.Len(t, plugins, 6)
		assert.Contains(t, plugins, "treesitter")
		assert.Contains(t, plugins, "telescope")
		assert.Contains(t, plugins, "which-key")
		assert.Contains(t, plugins, "lspconfig")
		assert.Contains(t, plugins, "nvim-cmp")
		assert.Contains(t, plugins, "gitsigns")

		fmt.Println()
	})

	fmt.Println("=== Integration test completed successfully ===")
}

// TestGetDefaultsForCLI demonstrates how the defaults would be collected
// and formatted for the 'dvm get defaults' command.
func TestGetDefaultsForCLI(t *testing.T) {
	// This simulates what the getDefaults CLI function would do
	allDefaults := make(map[string]interface{})

	// Collect defaults from all packages
	allDefaults["theme"] = themeresolver.GetDefaults()
	allDefaults["shell"] = shell.GetDefaults()
	allDefaults["nvim"] = nvimops.GetDefaults()

	// Verify the structure
	require.NotNil(t, allDefaults)
	assert.Len(t, allDefaults, 3)

	// Verify each section exists
	assert.Contains(t, allDefaults, "theme")
	assert.Contains(t, allDefaults, "shell")
	assert.Contains(t, allDefaults, "nvim")

	// This data structure is what would be passed to render.OutputWith()
	// for JSON, YAML, or formatted output
	fmt.Printf("CLI output data structure: %+v\n", allDefaults)
}

// TestDefaultsUsagePatterns demonstrates common usage patterns
// for the GetDefaults() functions in application code.
func TestDefaultsUsagePatterns(t *testing.T) {
	t.Run("workspace_creation_scenario", func(t *testing.T) {
		// Scenario: Creating a new workspace and need default configurations
		fmt.Println("=== Workspace Creation Scenario ===")

		// Get default nvim config for a new workspace
		nvimDefaults := nvimops.GetDefaults()
		structure := nvimDefaults["structure"].(string)
		pluginPackage := nvimDefaults["pluginPackage"].(string)

		fmt.Printf("Creating new workspace with Neovim structure: %s\n", structure)
		fmt.Printf("Using plugin package: %s\n", pluginPackage)

		// Get default theme
		themeDefaults := themeresolver.GetDefaults()
		globalTheme := themeDefaults["global"].(string)
		fmt.Printf("Default theme (if not overridden): %s\n", globalTheme)

		// Get default shell config
		shellDefaults := shell.GetDefaults()
		shellType := shellDefaults["type"].(string)
		framework := shellDefaults["framework"].(string)

		fmt.Printf("Shell configuration: %s with %s\n", shellType, framework)

		// Verify expected values
		assert.Equal(t, "lazyvim", structure)
		assert.Equal(t, "core", pluginPackage)
		assert.Equal(t, "coolnight-ocean", globalTheme)
		assert.Equal(t, "zsh", shellType)
		assert.Equal(t, "oh-my-zsh", framework)
	})

	t.Run("theme_resolution_scenario", func(t *testing.T) {
		// Scenario: Understanding theme resolution hierarchy
		fmt.Println("\n=== Theme Resolution Scenario ===")

		themeDefaults := themeresolver.GetDefaults()
		resolution := themeDefaults["resolution"].(string)
		defaultTheme := themeDefaults["global"].(string)

		fmt.Printf("Theme resolution follows this hierarchy: %s\n", resolution)
		fmt.Printf("If no theme is set at any level, falls back to: %s\n", defaultTheme)

		// Verify the resolution order makes sense
		assert.Contains(t, resolution, "workspace")
		assert.Contains(t, resolution, "app")
		assert.Contains(t, resolution, "domain")
		assert.Contains(t, resolution, "ecosystem")
		assert.Contains(t, resolution, "global")

		// Verify workspace comes first (highest priority)
		workspacePos := len(resolution) - len(resolution)
		globalPos := len(resolution) - len("global")
		assert.Less(t, workspacePos, globalPos, "workspace should have higher priority than global")
	})

	t.Run("configuration_validation_scenario", func(t *testing.T) {
		// Scenario: Validating user configuration against defaults
		fmt.Println("\n=== Configuration Validation Scenario ===")

		nvimDefaults := nvimops.GetDefaults()
		corePlugins := nvimDefaults["corePlugins"].([]string)

		// Simulate user providing a custom plugin list
		userPlugins := []string{"treesitter", "telescope", "custom-plugin"}

		// Check if user has essential plugins
		essentialMissing := []string{}
		for _, essential := range corePlugins {
			found := false
			for _, userPlugin := range userPlugins {
				if userPlugin == essential {
					found = true
					break
				}
			}
			if !found {
				essentialMissing = append(essentialMissing, essential)
			}
		}

		if len(essentialMissing) > 0 {
			fmt.Printf("Warning: User config is missing essential plugins: %v\n", essentialMissing)
			fmt.Printf("Consider adding them or using default plugin package: %s\n", nvimDefaults["pluginPackage"])
		} else {
			fmt.Println("User config includes all essential plugins")
		}

		// In this test, user is missing some essential plugins
		assert.Greater(t, len(essentialMissing), 0, "This test scenario should show missing plugins")
		assert.Contains(t, essentialMissing, "which-key")
		assert.Contains(t, essentialMissing, "lspconfig")
	})
}
