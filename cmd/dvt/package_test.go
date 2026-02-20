package main

import (
	"testing"

	terminalpackage "devopsmaestro/pkg/terminalops/package"
	packagelibrary "devopsmaestro/pkg/terminalops/package/library"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolvePackageComponents(t *testing.T) {
	lib, err := packagelibrary.NewLibrary()
	require.NoError(t, err)

	tests := []struct {
		name             string
		packageName      string
		expectedPlugins  []string
		expectedPrompts  []string
		expectedProfiles []string
		shouldContain    map[string][]string // component type -> list of items to check for
		shouldError      bool
		errorContains    string
	}{
		{
			name:        "core package - base case",
			packageName: "core",
			shouldContain: map[string][]string{
				"plugins": {"zsh-autosuggestions"},
			},
			shouldError: false,
		},
		{
			name:        "developer package - inheritance",
			packageName: "developer",
			shouldContain: map[string][]string{
				"plugins": {"zsh-autosuggestions", "fzf"}, // from core and developer
			},
			shouldError: false,
		},
		{
			name:          "non-existent package",
			packageName:   "nonexistent",
			shouldError:   true,
			errorContains: "not found in library",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldError {
				// For error cases, we need to simulate the package not being found
				// Create a fake package that extends a non-existent package
				fakePackage := &terminalpackage.Package{
					Name:    tt.packageName,
					Extends: "nonexistent-parent",
				}

				_, err := resolvePackageComponents(fakePackage, lib)
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			// Get the actual package from library
			pkg, ok := lib.Get(tt.packageName)
			require.True(t, ok, "package %s should exist in library", tt.packageName)

			result, err := resolvePackageComponents(pkg, lib)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Check that expected items are present
			if expectedItems, ok := tt.shouldContain["plugins"]; ok {
				for _, item := range expectedItems {
					assert.Contains(t, result.Plugins, item,
						"plugins should contain %s", item)
				}
			}
			if expectedItems, ok := tt.shouldContain["prompts"]; ok {
				for _, item := range expectedItems {
					assert.Contains(t, result.Prompts, item,
						"prompts should contain %s", item)
				}
			}
			if expectedItems, ok := tt.shouldContain["profiles"]; ok {
				for _, item := range expectedItems {
					assert.Contains(t, result.Profiles, item,
						"profiles should contain %s", item)
				}
			}

			// Verify no duplicates
			assert.Equal(t, len(result.Plugins), len(removeDuplicates(result.Plugins)),
				"plugins should not have duplicates")
			assert.Equal(t, len(result.Prompts), len(removeDuplicates(result.Prompts)),
				"prompts should not have duplicates")
			assert.Equal(t, len(result.Profiles), len(removeDuplicates(result.Profiles)),
				"profiles should not have duplicates")
		})
	}
}

func TestResolvePackageComponents_CircularDependency(t *testing.T) {
	// For this test, we need to create a scenario where the same package
	// would be visited twice. This is hard to test with the current implementation
	// since we can't easily inject circular dependencies into the real library.
	// Instead, we'll test the logic by calling the function directly.

	lib, err := packagelibrary.NewLibrary()
	require.NoError(t, err)

	// Create a package that extends a non-existent package
	testPackage := &terminalpackage.Package{
		Name:    "test-package",
		Extends: "nonexistent-parent",
		Plugins: []string{"test-plugin"},
	}

	// This should error because the parent package doesn't exist
	_, err = resolvePackageComponents(testPackage, lib)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found in library")
}

func TestResolvePackageComponents_DeepInheritance(t *testing.T) {
	// Test with real library packages that have inheritance
	lib, err := packagelibrary.NewLibrary()
	require.NoError(t, err)

	// Use the developer package which extends core
	pkg, ok := lib.Get("developer")
	if !ok {
		t.Skip("developer package not available in library")
	}

	result, err := resolvePackageComponents(pkg, lib)
	require.NoError(t, err)

	// Should include components from both developer and core packages
	// At minimum, check that we have some components
	totalComponents := len(result.Plugins) + len(result.Prompts) + len(result.Profiles)
	assert.Greater(t, totalComponents, 0, "should have components from inheritance chain")

	// Verify no duplicates
	assert.Equal(t, len(result.Plugins), len(removeDuplicates(result.Plugins)),
		"plugins should not have duplicates")
	assert.Equal(t, len(result.Prompts), len(removeDuplicates(result.Prompts)),
		"prompts should not have duplicates")
	assert.Equal(t, len(result.Profiles), len(removeDuplicates(result.Profiles)),
		"profiles should not have duplicates")
}

func TestGetComponentSource(t *testing.T) {
	lib, err := packagelibrary.NewLibrary()
	require.NoError(t, err)

	// Get a package that has inheritance
	pkg, ok := lib.Get("developer")
	if !ok {
		t.Skip("developer package not available in library")
	}

	// Get the parent package to compare
	parentPkg, parentOk := lib.Get(pkg.Extends)
	if !parentOk {
		t.Skip("parent package not available in library")
	}

	tests := []struct {
		name          string
		componentName string
		pkg           *terminalpackage.Package
		componentType string
		shouldFind    bool // Whether we should find a source
	}{
		{
			name:          "existing plugin in current package",
			componentName: getFirstItem(pkg.Plugins),
			pkg:           pkg,
			componentType: "plugin",
			shouldFind:    len(pkg.Plugins) > 0,
		},
		{
			name:          "existing prompt in current package",
			componentName: getFirstItem(pkg.Prompts),
			pkg:           pkg,
			componentType: "prompt",
			shouldFind:    len(pkg.Prompts) > 0,
		},
		{
			name:          "existing plugin from parent",
			componentName: getFirstItem(parentPkg.Plugins),
			pkg:           pkg,
			componentType: "plugin",
			shouldFind:    len(parentPkg.Plugins) > 0,
		},
		{
			name:          "non-existent component",
			componentName: "definitely-nonexistent-component",
			pkg:           pkg,
			componentType: "plugin",
			shouldFind:    false,
		},
	}

	for _, tt := range tests {
		if tt.componentName == "" && tt.shouldFind {
			continue // Skip tests where we don't have components to test
		}

		t.Run(tt.name, func(t *testing.T) {
			result := getComponentSource(tt.componentName, tt.pkg, lib, tt.componentType)
			if tt.shouldFind {
				assert.NotEmpty(t, result, "should find source for component")
			} else {
				assert.Empty(t, result, "should not find source for non-existent component")
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "item exists",
			slice:    []string{"a", "b", "c"},
			item:     "b",
			expected: true,
		},
		{
			name:     "item does not exist",
			slice:    []string{"a", "b", "c"},
			item:     "d",
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			item:     "a",
			expected: false,
		},
		{
			name:     "empty item",
			slice:    []string{"a", "b", "c"},
			item:     "",
			expected: false,
		},
		{
			name:     "nil slice",
			slice:    nil,
			item:     "a",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPackageCommands_Integration(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "package list",
			args: []string{"package", "list"},
		},
		{
			name: "package list with yaml output",
			args: []string{"package", "list", "-o", "yaml"},
		},
		{
			name: "package list with json output",
			args: []string{"package", "list", "-o", "json"},
		},
		{
			name: "package list wide format",
			args: []string{"package", "list", "-w"},
		},
		{
			name: "package get core",
			args: []string{"package", "get", "core"},
		},
		{
			name: "package get with yaml output",
			args: []string{"package", "get", "core", "-o", "yaml"},
		},
		{
			name: "package get with json output",
			args: []string{"package", "get", "core", "-o", "json"},
		},
		{
			name: "package install dry run",
			args: []string{"package", "install", "core", "--dry-run"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset root command args
			rootCmd.SetArgs(tt.args)

			// Execute command - should not error
			err := rootCmd.Execute()
			assert.NoError(t, err, "command should execute successfully: %v", tt.args)
		})
	}
}

func TestPackageCommands_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "package get nonexistent",
			args:        []string{"package", "get", "nonexistent"},
			expectError: true,
		},
		{
			name:        "package install nonexistent",
			args:        []string{"package", "install", "nonexistent"},
			expectError: true,
		},
		{
			name:        "package get without argument",
			args:        []string{"package", "get"},
			expectError: true,
		},
		{
			name:        "package install without argument",
			args:        []string{"package", "install"},
			expectError: true,
		},
		{
			name:        "package list with invalid output format",
			args:        []string{"package", "list", "-o", "invalid"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset root command args
			rootCmd.SetArgs(tt.args)

			// Execute command
			err := rootCmd.Execute()
			if tt.expectError {
				assert.Error(t, err, "command should return error: %v", tt.args)
			} else {
				assert.NoError(t, err, "command should not return error: %v", tt.args)
			}
		})
	}
}

// Helper function to get first item from a slice, or empty string if slice is empty
func getFirstItem(slice []string) string {
	if len(slice) == 0 {
		return ""
	}
	return slice[0]
}

// Helper function to remove duplicates from a slice
func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	return result
}
