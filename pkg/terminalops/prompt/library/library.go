// Package library provides embedded prompt library for terminalops.
package library

import (
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"devopsmaestro/pkg/terminalops/prompt"
)

//go:embed prompts/*.yaml
var promptFS embed.FS

// PromptLibrary provides access to the embedded prompt library.
type PromptLibrary struct {
	prompts map[string]*prompt.Prompt
}

// NewPromptLibrary creates a new PromptLibrary by loading all embedded prompts.
func NewPromptLibrary() (*PromptLibrary, error) {
	lib := &PromptLibrary{
		prompts: make(map[string]*prompt.Prompt),
	}

	entries, err := promptFS.ReadDir("prompts")
	if err != nil {
		return nil, fmt.Errorf("failed to read prompts directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		data, err := promptFS.ReadFile(filepath.Join("prompts", entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", entry.Name(), err)
		}

		p, err := prompt.Parse(data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", entry.Name(), err)
		}

		lib.prompts[p.Name] = p
	}

	return lib, nil
}

// Get returns a prompt by name.
func (lib *PromptLibrary) Get(name string) (*prompt.Prompt, error) {
	p, ok := lib.prompts[name]
	if !ok {
		return nil, fmt.Errorf("prompt not found: %s", name)
	}
	return p, nil
}

// List returns all prompts in the library.
func (lib *PromptLibrary) List() []*prompt.Prompt {
	prompts := make([]*prompt.Prompt, 0, len(lib.prompts))
	for _, p := range lib.prompts {
		prompts = append(prompts, p)
	}
	// Sort by name for consistent ordering
	sort.Slice(prompts, func(i, j int) bool {
		return prompts[i].Name < prompts[j].Name
	})
	return prompts
}

// ListByCategory returns prompts in a specific category.
func (lib *PromptLibrary) ListByCategory(category string) []*prompt.Prompt {
	var prompts []*prompt.Prompt
	for _, p := range lib.prompts {
		if p.Category == category {
			prompts = append(prompts, p)
		}
	}
	sort.Slice(prompts, func(i, j int) bool {
		return prompts[i].Name < prompts[j].Name
	})
	return prompts
}

// ListByType returns prompts of a specific type.
func (lib *PromptLibrary) ListByType(promptType prompt.PromptType) []*prompt.Prompt {
	var prompts []*prompt.Prompt
	for _, p := range lib.prompts {
		if p.Type == promptType {
			prompts = append(prompts, p)
		}
	}
	sort.Slice(prompts, func(i, j int) bool {
		return prompts[i].Name < prompts[j].Name
	})
	return prompts
}

// Names returns all prompt names in the library.
func (lib *PromptLibrary) Names() []string {
	names := make([]string, 0, len(lib.prompts))
	for name := range lib.prompts {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Count returns the number of prompts in the library.
func (lib *PromptLibrary) Count() int {
	return len(lib.prompts)
}

// Categories returns all unique categories in the library.
func (lib *PromptLibrary) Categories() []string {
	categorySet := make(map[string]struct{})
	for _, p := range lib.prompts {
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
