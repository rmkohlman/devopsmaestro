// Package nvimbridge provides database adapters that bridge MaestroNvim types
// with dvm's database layer (models, db packages).
package nvimbridge

import "devopsmaestro/models"

const (
	DefaultStructure     = "lazyvim" // Most popular, well-maintained
	DefaultPackage       = "core"    // Minimal but useful base package
	DefaultPluginPackage = "core"    // Alias for consistency
)

// DefaultNvimConfig returns the default Neovim configuration for new workspaces.
// This provides a minimal but useful setup with essential plugins.
func DefaultNvimConfig() models.NvimConfig {
	return models.NvimConfig{
		Structure:     DefaultStructure,
		Theme:         "", // Let theme resolution handle via cascade
		PluginPackage: DefaultPackage,
		Plugins:       nil, // Use package plugins only
		MergeMode:     "append",
	}
}

// LanguagePackageMap maps detected language names to their recommended maestro nvim package.
// Used during workspace builds to auto-select the right package when no explicit default is set.
var LanguagePackageMap = map[string]string{
	"golang":  "maestro-go",
	"python":  "maestro-python",
	"rust":    "maestro-rust",
	"nodejs":  "maestro-node",
	"java":    "maestro-java",
	"gleam":   "maestro-gleam",
	"dotnet":  "maestro-dotnet",
	"ruby":    "maestro", // No Ruby-specific package yet, use base maestro
	"php":     "maestro-php",
	"kotlin":  "maestro-kotlin",
	"scala":   "maestro-scala",
	"elixir":  "maestro-elixir",
	"swift":   "maestro-swift",
	"zig":     "maestro-zig",
	"dart":    "maestro-dart",
	"lua":     "maestro-lua",
	"r":       "maestro-r",
	"haskell": "maestro-haskell",
	"perl":    "maestro-perl",
}

// GetLanguagePackage returns the recommended nvim package for a detected language.
// Returns empty string if no mapping exists.
func GetLanguagePackage(language string) string {
	if pkg, ok := LanguagePackageMap[language]; ok {
		return pkg
	}
	return ""
}

// GetDefaults returns the default nvim configuration settings
func GetDefaults() map[string]interface{} {
	return map[string]interface{}{
		"structure":     DefaultStructure,
		"pluginPackage": DefaultPluginPackage,
		"mergeMode":     "append",
		"corePlugins": []string{
			"treesitter",
			"telescope",
			"which-key",
			"lspconfig",
			"nvim-cmp",
			"gitsigns",
		},
	}
}
