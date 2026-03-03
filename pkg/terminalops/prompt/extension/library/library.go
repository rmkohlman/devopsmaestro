// Package library provides embedded extension library for terminalops.
package library

import (
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"devopsmaestro/pkg/terminalops/prompt/extension"
)

//go:embed extensions/*.yaml
var extensionFS embed.FS

// ExtensionLibrary provides access to the embedded extension library.
type ExtensionLibrary struct {
	extensions map[string]*extension.PromptExtension
}

// NewExtensionLibrary creates a new ExtensionLibrary by loading all embedded extensions.
func NewExtensionLibrary() (*ExtensionLibrary, error) {
	lib := &ExtensionLibrary{
		extensions: make(map[string]*extension.PromptExtension),
	}

	entries, err := extensionFS.ReadDir("extensions")
	if err != nil {
		return nil, fmt.Errorf("failed to read extensions directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		data, err := extensionFS.ReadFile(filepath.Join("extensions", entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", entry.Name(), err)
		}

		ext, err := parseExtension(data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", entry.Name(), err)
		}

		lib.extensions[ext.Name] = ext
	}

	return lib, nil
}

// parseExtension parses YAML data into a PromptExtension.
func parseExtension(data []byte) (*extension.PromptExtension, error) {
	var extYAML extension.PromptExtensionYAML
	if err := yaml.Unmarshal(data, &extYAML); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return extYAML.ToPromptExtension(), nil
}

// Get returns an extension by name.
func (lib *ExtensionLibrary) Get(name string) (*extension.PromptExtension, error) {
	ext, ok := lib.extensions[name]
	if !ok {
		return nil, fmt.Errorf("extension not found: %s", name)
	}
	return ext, nil
}

// List returns all extensions in the library.
func (lib *ExtensionLibrary) List() []*extension.PromptExtension {
	extensions := make([]*extension.PromptExtension, 0, len(lib.extensions))
	for _, ext := range lib.extensions {
		extensions = append(extensions, ext)
	}
	// Sort by name for consistent ordering
	sort.Slice(extensions, func(i, j int) bool {
		return extensions[i].Name < extensions[j].Name
	})
	return extensions
}

// ListBySegment returns extensions targeting a specific segment.
func (lib *ExtensionLibrary) ListBySegment(segment string) []*extension.PromptExtension {
	var extensions []*extension.PromptExtension
	for _, ext := range lib.extensions {
		if ext.Segment == segment {
			extensions = append(extensions, ext)
		}
	}
	sort.Slice(extensions, func(i, j int) bool {
		return extensions[i].Name < extensions[j].Name
	})
	return extensions
}

// ListByCategory returns extensions in a specific category.
func (lib *ExtensionLibrary) ListByCategory(category string) []*extension.PromptExtension {
	var extensions []*extension.PromptExtension
	for _, ext := range lib.extensions {
		if ext.Category == category {
			extensions = append(extensions, ext)
		}
	}
	sort.Slice(extensions, func(i, j int) bool {
		return extensions[i].Name < extensions[j].Name
	})
	return extensions
}

// Names returns all extension names in the library.
func (lib *ExtensionLibrary) Names() []string {
	names := make([]string, 0, len(lib.extensions))
	for name := range lib.extensions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Count returns the number of extensions in the library.
func (lib *ExtensionLibrary) Count() int {
	return len(lib.extensions)
}

// Categories returns all unique categories in the library.
func (lib *ExtensionLibrary) Categories() []string {
	categorySet := make(map[string]struct{})
	for _, ext := range lib.extensions {
		if ext.Category != "" {
			categorySet[ext.Category] = struct{}{}
		}
	}
	categories := make([]string, 0, len(categorySet))
	for cat := range categorySet {
		categories = append(categories, cat)
	}
	sort.Strings(categories)
	return categories
}

// Segments returns all unique segments targeted by extensions.
func (lib *ExtensionLibrary) Segments() []string {
	segmentSet := make(map[string]struct{})
	for _, ext := range lib.extensions {
		if ext.Segment != "" {
			segmentSet[ext.Segment] = struct{}{}
		}
	}
	segments := make([]string, 0, len(segmentSet))
	for seg := range segmentSet {
		segments = append(segments, seg)
	}
	sort.Strings(segments)
	return segments
}
