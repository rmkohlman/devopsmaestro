// Package library provides embedded style library for terminalops.
package library

import (
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"devopsmaestro/pkg/terminalops/prompt/style"
)

//go:embed styles/*.yaml
var styleFS embed.FS

// StyleLibrary provides access to the embedded style library.
type StyleLibrary struct {
	styles map[string]*style.PromptStyle
}

// NewStyleLibrary creates a new StyleLibrary by loading all embedded styles.
func NewStyleLibrary() (*StyleLibrary, error) {
	lib := &StyleLibrary{
		styles: make(map[string]*style.PromptStyle),
	}

	entries, err := styleFS.ReadDir("styles")
	if err != nil {
		return nil, fmt.Errorf("failed to read styles directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		data, err := styleFS.ReadFile(filepath.Join("styles", entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", entry.Name(), err)
		}

		s, err := parseStyle(data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", entry.Name(), err)
		}

		lib.styles[s.Name] = s
	}

	return lib, nil
}

// parseStyle parses YAML data into a PromptStyle.
func parseStyle(data []byte) (*style.PromptStyle, error) {
	var styleYAML style.PromptStyleYAML
	if err := yaml.Unmarshal(data, &styleYAML); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return styleYAML.ToPromptStyle(), nil
}

// Get returns a style by name.
func (lib *StyleLibrary) Get(name string) (*style.PromptStyle, error) {
	s, ok := lib.styles[name]
	if !ok {
		return nil, fmt.Errorf("style not found: %s", name)
	}
	return s, nil
}

// List returns all styles in the library.
func (lib *StyleLibrary) List() []*style.PromptStyle {
	styles := make([]*style.PromptStyle, 0, len(lib.styles))
	for _, s := range lib.styles {
		styles = append(styles, s)
	}
	// Sort by name for consistent ordering
	sort.Slice(styles, func(i, j int) bool {
		return styles[i].Name < styles[j].Name
	})
	return styles
}

// ListByCategory returns styles in a specific category.
func (lib *StyleLibrary) ListByCategory(category string) []*style.PromptStyle {
	var styles []*style.PromptStyle
	for _, s := range lib.styles {
		if s.Category == category {
			styles = append(styles, s)
		}
	}
	sort.Slice(styles, func(i, j int) bool {
		return styles[i].Name < styles[j].Name
	})
	return styles
}

// Names returns all style names in the library.
func (lib *StyleLibrary) Names() []string {
	names := make([]string, 0, len(lib.styles))
	for name := range lib.styles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Count returns the number of styles in the library.
func (lib *StyleLibrary) Count() int {
	return len(lib.styles)
}

// Categories returns all unique categories in the library.
func (lib *StyleLibrary) Categories() []string {
	categorySet := make(map[string]struct{})
	for _, s := range lib.styles {
		if s.Category != "" {
			categorySet[s.Category] = struct{}{}
		}
	}
	categories := make([]string, 0, len(categorySet))
	for cat := range categorySet {
		categories = append(categories, cat)
	}
	sort.Strings(categories)
	return categories
}
