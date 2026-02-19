// Package nvimops provides default configurations for Neovim workspaces.
package nvimops

import "devopsmaestro/models"

const (
	DefaultStructure = "lazyvim" // Most popular, well-maintained
	DefaultPackage   = "core"    // Minimal but useful base package
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
