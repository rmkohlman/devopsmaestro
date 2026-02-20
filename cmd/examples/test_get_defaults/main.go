// Simple test program to demonstrate GetDefaults() function usage
// Run with: go run cmd/examples/test_get_defaults/main.go
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"devopsmaestro/builders"
	themeresolver "devopsmaestro/pkg/colors/resolver"
	"devopsmaestro/pkg/nvimops"
	"devopsmaestro/pkg/terminalops/shell"
)

func main() {
	fmt.Println("=== GetDefaults() Function Demo ===")
	fmt.Println("This demonstrates how the GetDefaults() functions work in DevOpsMaestro")
	fmt.Println()

	// Example 1: Get theme resolver defaults
	fmt.Println("1. Theme Resolver Defaults:")
	themeDefaults := themeresolver.GetDefaults()
	printMap(themeDefaults, "  ")
	fmt.Println()

	// Example 2: Get shell defaults
	fmt.Println("2. Shell Configuration Defaults:")
	shellDefaults := shell.GetDefaults()
	printMap(shellDefaults, "  ")
	fmt.Println()

	// Example 3: Get Neovim defaults
	fmt.Println("3. Neovim Configuration Defaults:")
	nvimDefaults := nvimops.GetDefaults()
	printMap(nvimDefaults, "  ")
	fmt.Println()

	// Example 4: Get container defaults
	fmt.Println("4. Container Configuration Defaults:")
	containerDefaults := builders.GetContainerDefaults()
	printMap(containerDefaults, "  ")
	fmt.Println()

	// Example 5: Collect all defaults (as done by CLI)
	fmt.Println("5. All Defaults Combined (CLI format):")
	allDefaults := map[string]interface{}{
		"theme":     themeDefaults,
		"shell":     shellDefaults,
		"nvim":      nvimDefaults,
		"container": containerDefaults,
	}

	// Pretty print as JSON
	jsonData, err := json.MarshalIndent(allDefaults, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonData))
	fmt.Println()

	// Example 6: Demonstrate practical usage
	fmt.Println("6. Practical Usage Examples:")
	fmt.Println()

	// Theme resolution example
	resolution := themeDefaults["resolution"].(string)
	fmt.Printf("Theme Resolution Hierarchy: %s\n", resolution)
	globalTheme := themeDefaults["global"].(string)
	fmt.Printf("Default Global Theme: %s\n", globalTheme)
	fmt.Println()

	// Workspace creation example
	nvimStructure := nvimDefaults["structure"].(string)
	nvimPackage := nvimDefaults["pluginPackage"].(string)
	fmt.Printf("New Workspace would use: Neovim structure '%s' with plugin package '%s'\n", nvimStructure, nvimPackage)

	shellType := shellDefaults["type"].(string)
	shellFramework := shellDefaults["framework"].(string)
	fmt.Printf("Shell would be: %s with %s framework\n", shellType, shellFramework)
	fmt.Println()

	// Essential plugins check
	corePlugins := nvimDefaults["corePlugins"].([]string)
	fmt.Printf("Essential plugins included by default: %v\n", corePlugins)
	fmt.Println()

	fmt.Println("=== Demo Complete ===")
	fmt.Println("This shows how the 'dvm get defaults' command gathers and displays default configurations.")
}

// printMap prints a map with proper indentation
func printMap(m map[string]interface{}, indent string) {
	for key, value := range m {
		switch v := value.(type) {
		case []string:
			fmt.Printf("%s%s: [%s]\n", indent, key, fmt.Sprintf("%v", v))
		default:
			fmt.Printf("%s%s: %v\n", indent, key, v)
		}
	}
}
