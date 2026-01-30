// Package library provides access to pre-built plugin definitions.
// These are curated plugin configurations that work well together.
package library

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"devopsmaestro/pkg/nvimops/plugin"
)

//go:embed plugins/*.yaml
var embeddedPlugins embed.FS

// Library provides access to pre-built plugin definitions.
type Library struct {
	plugins map[string]*plugin.Plugin
}

// NewLibrary creates a new Library, loading embedded plugins.
func NewLibrary() (*Library, error) {
	lib := &Library{
		plugins: make(map[string]*plugin.Plugin),
	}

	// Load embedded plugins
	err := fs.WalkDir(embeddedPlugins, "plugins", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}

		data, err := embeddedPlugins.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		p, err := plugin.ParseYAML(data)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		lib.plugins[p.Name] = p
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load library: %w", err)
	}

	return lib, nil
}

// NewLibraryFromDir creates a Library from a directory of YAML files.
func NewLibraryFromDir(dir string) (*Library, error) {
	lib := &Library{
		plugins: make(map[string]*plugin.Plugin),
	}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}

		p, err := plugin.ParseYAMLFile(path)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		lib.plugins[p.Name] = p
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load library from %s: %w", dir, err)
	}

	return lib, nil
}

// Get retrieves a plugin by name from the library.
func (l *Library) Get(name string) (*plugin.Plugin, bool) {
	p, ok := l.plugins[name]
	return p, ok
}

// List returns all plugins in the library, sorted by name.
func (l *Library) List() []*plugin.Plugin {
	result := make([]*plugin.Plugin, 0, len(l.plugins))
	for _, p := range l.plugins {
		result = append(result, p)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// ListByCategory returns plugins in a specific category.
func (l *Library) ListByCategory(category string) []*plugin.Plugin {
	var result []*plugin.Plugin
	for _, p := range l.plugins {
		if p.Category == category {
			result = append(result, p)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// ListByTag returns plugins that have a specific tag.
func (l *Library) ListByTag(tag string) []*plugin.Plugin {
	var result []*plugin.Plugin
	for _, p := range l.plugins {
		for _, t := range p.Tags {
			if t == tag {
				result = append(result, p)
				break
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// Categories returns all unique categories in the library.
func (l *Library) Categories() []string {
	categories := make(map[string]bool)
	for _, p := range l.plugins {
		if p.Category != "" {
			categories[p.Category] = true
		}
	}

	result := make([]string, 0, len(categories))
	for c := range categories {
		result = append(result, c)
	}
	sort.Strings(result)
	return result
}

// Tags returns all unique tags in the library.
func (l *Library) Tags() []string {
	tags := make(map[string]bool)
	for _, p := range l.plugins {
		for _, t := range p.Tags {
			tags[t] = true
		}
	}

	result := make([]string, 0, len(tags))
	for t := range tags {
		result = append(result, t)
	}
	sort.Strings(result)
	return result
}

// Count returns the number of plugins in the library.
func (l *Library) Count() int {
	return len(l.plugins)
}

// PluginInfo provides summary information about a plugin.
type PluginInfo struct {
	Name        string
	Description string
	Category    string
	Tags        []string
	Repo        string
}

// Info returns summary information for all plugins.
func (l *Library) Info() []PluginInfo {
	result := make([]PluginInfo, 0, len(l.plugins))
	for _, p := range l.plugins {
		result = append(result, PluginInfo{
			Name:        p.Name,
			Description: p.Description,
			Category:    p.Category,
			Tags:        p.Tags,
			Repo:        p.Repo,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}
