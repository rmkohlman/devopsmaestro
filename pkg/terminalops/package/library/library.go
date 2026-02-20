// Package library provides access to pre-built package definitions.
// These are curated package configurations that work well together.
package library

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	pkg "devopsmaestro/pkg/terminalops/package"
)

//go:embed packages/*.yaml
var embeddedPackages embed.FS

// Library provides access to pre-built package definitions.
type Library struct {
	packages map[string]*pkg.Package
}

// NewLibrary creates a new Library, loading embedded packages.
func NewLibrary() (*Library, error) {
	lib := &Library{
		packages: make(map[string]*pkg.Package),
	}

	// Load embedded packages
	err := fs.WalkDir(embeddedPackages, "packages", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}

		data, err := embeddedPackages.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		p, err := pkg.ParseYAML(data)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		lib.packages[p.Name] = p
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
		packages: make(map[string]*pkg.Package),
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

		p, err := pkg.ParseYAMLFile(path)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		lib.packages[p.Name] = p
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load library from %s: %w", dir, err)
	}

	return lib, nil
}

// Get retrieves a package by name from the library.
func (l *Library) Get(name string) (*pkg.Package, bool) {
	p, ok := l.packages[name]
	return p, ok
}

// List returns all packages in the library, sorted by name.
func (l *Library) List() []*pkg.Package {
	result := make([]*pkg.Package, 0, len(l.packages))
	for _, p := range l.packages {
		result = append(result, p)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// ListByCategory returns packages in a specific category.
func (l *Library) ListByCategory(category string) []*pkg.Package {
	var result []*pkg.Package
	for _, p := range l.packages {
		if p.Category == category {
			result = append(result, p)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// ListByTag returns packages that have a specific tag.
func (l *Library) ListByTag(tag string) []*pkg.Package {
	var result []*pkg.Package
	for _, p := range l.packages {
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
	for _, p := range l.packages {
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
	for _, p := range l.packages {
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

// Count returns the number of packages in the library.
func (l *Library) Count() int {
	return len(l.packages)
}

// Has checks if a package with the given name exists in the library.
func (l *Library) Has(name string) bool {
	_, ok := l.packages[name]
	return ok
}

// PackageInfo provides summary information about a package.
type PackageInfo struct {
	Name         string
	Description  string
	Category     string
	Tags         []string
	Extends      string
	PluginCount  int
	PromptCount  int
	ProfileCount int
}

// Info returns summary information for all packages.
func (l *Library) Info() []PackageInfo {
	result := make([]PackageInfo, 0, len(l.packages))
	for _, p := range l.packages {
		result = append(result, PackageInfo{
			Name:         p.Name,
			Description:  p.Description,
			Category:     p.Category,
			Tags:         p.Tags,
			Extends:      p.Extends,
			PluginCount:  len(p.Plugins),
			PromptCount:  len(p.Prompts),
			ProfileCount: len(p.Profiles),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}
