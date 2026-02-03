// Package library provides embedded plugin library for terminalops.
package library

import (
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"devopsmaestro/pkg/terminalops/plugin"
)

//go:embed plugins/*.yaml
var pluginFS embed.FS

// PluginLibrary provides access to the embedded plugin library.
type PluginLibrary struct {
	plugins map[string]*plugin.Plugin
}

// NewPluginLibrary creates a new PluginLibrary by loading all embedded plugins.
func NewPluginLibrary() (*PluginLibrary, error) {
	lib := &PluginLibrary{
		plugins: make(map[string]*plugin.Plugin),
	}

	entries, err := pluginFS.ReadDir("plugins")
	if err != nil {
		return nil, fmt.Errorf("failed to read plugins directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		data, err := pluginFS.ReadFile(filepath.Join("plugins", entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", entry.Name(), err)
		}

		p, err := plugin.Parse(data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", entry.Name(), err)
		}

		lib.plugins[p.Name] = p
	}

	return lib, nil
}

// Get returns a plugin by name.
func (lib *PluginLibrary) Get(name string) (*plugin.Plugin, error) {
	p, ok := lib.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}
	return p, nil
}

// List returns all plugins in the library.
func (lib *PluginLibrary) List() []*plugin.Plugin {
	plugins := make([]*plugin.Plugin, 0, len(lib.plugins))
	for _, p := range lib.plugins {
		plugins = append(plugins, p)
	}
	// Sort by priority then name
	sort.Slice(plugins, func(i, j int) bool {
		if plugins[i].Priority != plugins[j].Priority {
			return plugins[i].Priority < plugins[j].Priority
		}
		return plugins[i].Name < plugins[j].Name
	})
	return plugins
}

// ListByCategory returns plugins in a specific category.
func (lib *PluginLibrary) ListByCategory(category string) []*plugin.Plugin {
	var plugins []*plugin.Plugin
	for _, p := range lib.plugins {
		if p.Category == category {
			plugins = append(plugins, p)
		}
	}
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Name < plugins[j].Name
	})
	return plugins
}

// ListByManager returns plugins using a specific manager.
func (lib *PluginLibrary) ListByManager(manager plugin.PluginManager) []*plugin.Plugin {
	var plugins []*plugin.Plugin
	for _, p := range lib.plugins {
		if p.Manager == manager {
			plugins = append(plugins, p)
		}
	}
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Name < plugins[j].Name
	})
	return plugins
}

// Names returns all plugin names in the library.
func (lib *PluginLibrary) Names() []string {
	names := make([]string, 0, len(lib.plugins))
	for name := range lib.plugins {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Count returns the number of plugins in the library.
func (lib *PluginLibrary) Count() int {
	return len(lib.plugins)
}

// Categories returns all unique categories in the library.
func (lib *PluginLibrary) Categories() []string {
	categorySet := make(map[string]struct{})
	for _, p := range lib.plugins {
		if p.Category != "" {
			categorySet[p.Category] = struct{}{}
		}
	}
	categories := make([]string, 0, len(categorySet))
	for cat := range categorySet {
		categories = append(categories, cat)
	}
	sort.Strings(categories)
	return categories
}

// EssentialPlugins returns the most commonly used plugins.
func (lib *PluginLibrary) EssentialPlugins() []*plugin.Plugin {
	essentials := []string{
		"zsh-autosuggestions",
		"zsh-syntax-highlighting",
	}
	var plugins []*plugin.Plugin
	for _, name := range essentials {
		if p, err := lib.Get(name); err == nil {
			plugins = append(plugins, p)
		}
	}
	return plugins
}
