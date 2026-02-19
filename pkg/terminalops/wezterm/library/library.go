// Package library provides embedded WezTerm configuration library for terminalops.
package library

import (
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"devopsmaestro/pkg/terminalops/wezterm"
)

//go:embed presets/*.yaml
var presetFS embed.FS

// WeztermLibrary provides access to the embedded WezTerm configuration library.
type WeztermLibrary struct {
	configs map[string]*wezterm.WezTerm
}

// NewWeztermLibrary creates a new WeztermLibrary by loading all embedded presets.
func NewWeztermLibrary() (*WeztermLibrary, error) {
	lib := &WeztermLibrary{
		configs: make(map[string]*wezterm.WezTerm),
	}

	entries, err := presetFS.ReadDir("presets")
	if err != nil {
		return nil, fmt.Errorf("failed to read presets directory: %w", err)
	}

	parser := wezterm.NewParser()

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		data, err := presetFS.ReadFile(filepath.Join("presets", entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", entry.Name(), err)
		}

		config, err := parser.Parse(data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", entry.Name(), err)
		}

		lib.configs[config.Name] = config
	}

	return lib, nil
}

// Get returns a WezTerm configuration by name.
func (lib *WeztermLibrary) Get(name string) (*wezterm.WezTerm, error) {
	config, ok := lib.configs[name]
	if !ok {
		return nil, fmt.Errorf("wezterm config not found: %s", name)
	}
	return config, nil
}

// List returns all WezTerm configurations in the library.
func (lib *WeztermLibrary) List() []*wezterm.WezTerm {
	configs := make([]*wezterm.WezTerm, 0, len(lib.configs))
	for _, config := range lib.configs {
		configs = append(configs, config)
	}
	// Sort by name for consistent ordering
	sort.Slice(configs, func(i, j int) bool {
		return configs[i].Name < configs[j].Name
	})
	return configs
}

// ListByCategory returns WezTerm configurations in a specific category.
func (lib *WeztermLibrary) ListByCategory(category string) []*wezterm.WezTerm {
	var configs []*wezterm.WezTerm
	for _, config := range lib.configs {
		if config.Category == category {
			configs = append(configs, config)
		}
	}
	sort.Slice(configs, func(i, j int) bool {
		return configs[i].Name < configs[j].Name
	})
	return configs
}

// Names returns all configuration names in the library.
func (lib *WeztermLibrary) Names() []string {
	names := make([]string, 0, len(lib.configs))
	for name := range lib.configs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Count returns the number of configurations in the library.
func (lib *WeztermLibrary) Count() int {
	return len(lib.configs)
}

// Categories returns all unique categories in the library.
func (lib *WeztermLibrary) Categories() []string {
	categorySet := make(map[string]struct{})
	for _, config := range lib.configs {
		if config.Category != "" {
			categorySet[config.Category] = struct{}{}
		}
	}
	categories := make([]string, 0, len(categorySet))
	for cat := range categorySet {
		categories = append(categories, cat)
	}
	sort.Strings(categories)
	return categories
}
