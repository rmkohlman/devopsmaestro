//go:build !integration

package cmd

import (
	"testing"

	"devopsmaestro/db"
)

func TestResolveDefaultPackagePlugins_LanguagePackages(t *testing.T) {
	tests := []struct {
		packageName   string
		expectedCount int
	}{
		{"maestro-go", 42},
		{"maestro-python", 42},
		{"maestro-rust", 42},
		{"maestro-node", 40},
		{"maestro-java", 40},
		{"maestro-gleam", 37},
		{"maestro-dotnet", 39},
		{"maestro", 37},
		{"core", 6},
	}

	for _, tt := range tests {
		t.Run(tt.packageName, func(t *testing.T) {
			plugins, err := resolveDefaultPackagePlugins(tt.packageName, nil)
			if err != nil {
				t.Fatalf("resolveDefaultPackagePlugins(%q, nil) returned unexpected error: %v", tt.packageName, err)
			}
			if len(plugins) != tt.expectedCount {
				t.Errorf("resolveDefaultPackagePlugins(%q): got %d plugins, want %d\nplugins: %v",
					tt.packageName, len(plugins), tt.expectedCount, plugins)
			}
		})
	}
}

func TestResolveDefaultPackagePlugins_LanguagePackage_ContainsExpectedPlugins(t *testing.T) {
	plugins, err := resolveDefaultPackagePlugins("maestro-go", nil)
	if err != nil {
		t.Fatalf("resolveDefaultPackagePlugins(\"maestro-go\", nil) returned unexpected error: %v", err)
	}

	pluginSet := make(map[string]bool, len(plugins))
	for _, p := range plugins {
		pluginSet[p] = true
	}

	parentPlugins := []string{"telescope", "treesitter", "nvim-cmp", "copilot", "lazygit"}
	for _, p := range parentPlugins {
		if !pluginSet[p] {
			t.Errorf("expected parent plugin %q to be present in maestro-go resolved plugins, but it was not\nresolved: %v", p, plugins)
		}
	}

	childPlugins := []string{"nvim-dap", "nvim-dap-go", "neotest", "neotest-go", "gopher-nvim"}
	for _, p := range childPlugins {
		if !pluginSet[p] {
			t.Errorf("expected child plugin %q to be present in maestro-go resolved plugins, but it was not\nresolved: %v", p, plugins)
		}
	}
}

func TestResolveDefaultPackagePlugins_NonExistentPackage(t *testing.T) {
	mockDS := db.NewMockDataStore()
	_, err := resolveDefaultPackagePlugins("non-existent-package", mockDS)
	if err == nil {
		t.Fatal("resolveDefaultPackagePlugins(\"non-existent-package\", mockDS) expected an error but got nil")
	}
}
