// Package library provides embedded pre-defined themes.
package library

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"

	"devopsmaestro/pkg/nvimops/theme"
)

//go:embed themes/*.yaml
var themesFS embed.FS

// ThemeInfo contains metadata about a library theme.
type ThemeInfo struct {
	Name        string
	Description string
	Author      string
	Category    string
	Plugin      string
}

// List returns all available themes in the library.
func List() ([]ThemeInfo, error) {
	entries, err := themesFS.ReadDir("themes")
	if err != nil {
		return nil, fmt.Errorf("failed to read themes directory: %w", err)
	}

	var themes []ThemeInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		t, err := Get(strings.TrimSuffix(entry.Name(), ".yaml"))
		if err != nil {
			continue
		}

		themes = append(themes, ThemeInfo{
			Name:        t.Name,
			Description: t.Description,
			Author:      t.Author,
			Category:    t.Category,
			Plugin:      t.Plugin.Repo,
		})
	}

	return themes, nil
}

// Get retrieves a theme by name from the library.
func Get(name string) (*theme.Theme, error) {
	// Try exact match first
	data, err := themesFS.ReadFile(filepath.Join("themes", name+".yaml"))
	if err != nil {
		// Try with common suffixes removed/added
		variants := []string{
			name,
			strings.TrimSuffix(name, "-dark"),
			strings.TrimSuffix(name, "-light"),
			name + "-dark",
			name + "-mocha",
			name + "-night",
		}

		for _, variant := range variants {
			data, err = themesFS.ReadFile(filepath.Join("themes", variant+".yaml"))
			if err == nil {
				break
			}
		}

		if err != nil {
			return nil, fmt.Errorf("theme %q not found in library", name)
		}
	}

	return theme.ParseYAML(data)
}

// GetRaw retrieves the raw YAML for a theme.
func GetRaw(name string) ([]byte, error) {
	data, err := themesFS.ReadFile(filepath.Join("themes", name+".yaml"))
	if err != nil {
		return nil, fmt.Errorf("theme %q not found in library", name)
	}
	return data, nil
}

// Has checks if a theme exists in the library.
func Has(name string) bool {
	_, err := Get(name)
	return err == nil
}

// Categories returns all unique categories in the library.
func Categories() ([]string, error) {
	themes, err := List()
	if err != nil {
		return nil, err
	}

	categoryMap := make(map[string]bool)
	for _, t := range themes {
		if t.Category != "" {
			categoryMap[t.Category] = true
		}
	}

	categories := make([]string, 0, len(categoryMap))
	for cat := range categoryMap {
		categories = append(categories, cat)
	}

	return categories, nil
}

// ListByCategory returns themes filtered by category.
func ListByCategory(category string) ([]ThemeInfo, error) {
	themes, err := List()
	if err != nil {
		return nil, err
	}

	var filtered []ThemeInfo
	for _, t := range themes {
		if t.Category == category {
			filtered = append(filtered, t)
		}
	}

	return filtered, nil
}
