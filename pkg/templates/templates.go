// Package templates provides access to embedded YAML resource templates.
// It follows the Interface → Implementation → Factory pattern.
package templates

import (
	"bytes"
	"embed"
	"fmt"
	"sort"
)

//go:embed yaml/*.yaml
var yamlFS embed.FS

// TemplateStore provides access to embedded YAML resource templates.
type TemplateStore interface {
	// Get returns the annotated YAML template for the given resource kind.
	// Kind uses the CLI kebab-case form (e.g., "workspace", "nvim-plugin").
	// PascalCase (e.g., "NvimPlugin") is also accepted.
	Get(kind string) ([]byte, error)

	// List returns all available template kind names, sorted alphabetically.
	List() []string

	// GetAll returns a multi-document YAML string with all templates
	// separated by "---" document separators in hierarchy order.
	GetAll() ([]byte, error)
}

// embeddedTemplateStore is the default TemplateStore backed by embedded YAML.
type embeddedTemplateStore struct {
	// templates maps kebab-case kind → file contents
	templates map[string][]byte
	// pascalToKebab maps PascalCase kind → kebab-case kind
	pascalToKebab map[string]string
	// sortedKinds is the alphabetically sorted list of kebab-case kinds
	sortedKinds []string
}

// hierarchyOrder defines the output ordering for GetAll().
var hierarchyOrder = []string{
	"ecosystem",
	"domain",
	"app",
	"workspace",
	"credential",
	"registry",
	"gitrepo",
	"nvim-plugin",
	"nvim-theme",
	"nvim-package",
	"terminal-prompt",
	"terminal-package",
	"terminal-plugin",
	"global-defaults",
	"custom-resource-definition",
}

// kindMappings maps kebab-case to PascalCase for normalization.
var kindMappings = map[string]string{
	"ecosystem":                  "Ecosystem",
	"domain":                     "Domain",
	"app":                        "App",
	"workspace":                  "Workspace",
	"credential":                 "Credential",
	"registry":                   "Registry",
	"gitrepo":                    "Gitrepo",
	"nvim-plugin":                "NvimPlugin",
	"nvim-theme":                 "NvimTheme",
	"nvim-package":               "NvimPackage",
	"terminal-prompt":            "TerminalPrompt",
	"terminal-package":           "TerminalPackage",
	"terminal-plugin":            "TerminalPlugin",
	"global-defaults":            "GlobalDefaults",
	"custom-resource-definition": "CustomResourceDefinition",
}

// NewTemplateStore creates a TemplateStore backed by embedded YAML files.
func NewTemplateStore() (TemplateStore, error) {
	store := &embeddedTemplateStore{
		templates:     make(map[string][]byte),
		pascalToKebab: make(map[string]string),
	}

	// Build the reverse mapping (PascalCase → kebab-case)
	for kebab, pascal := range kindMappings {
		store.pascalToKebab[pascal] = kebab
	}

	// Load all embedded templates
	for kebab := range kindMappings {
		filename := "yaml/" + kebab + ".yaml"
		data, err := yamlFS.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read embedded template %q: %w", filename, err)
		}
		store.templates[kebab] = data
	}

	// Build sorted kind list
	store.sortedKinds = make([]string, 0, len(store.templates))
	for kind := range store.templates {
		store.sortedKinds = append(store.sortedKinds, kind)
	}
	sort.Strings(store.sortedKinds)

	return store, nil
}

// Get returns the template bytes for the given kind.
// Accepts both kebab-case and PascalCase.
func (s *embeddedTemplateStore) Get(kind string) ([]byte, error) {
	normalized := s.normalize(kind)
	if normalized == "" {
		return nil, fmt.Errorf("unknown template kind: %q", kind)
	}

	data, ok := s.templates[normalized]
	if !ok {
		return nil, fmt.Errorf("unknown template kind: %q", kind)
	}

	return data, nil
}

// List returns all available template kinds, sorted alphabetically.
func (s *embeddedTemplateStore) List() []string {
	result := make([]string, len(s.sortedKinds))
	copy(result, s.sortedKinds)
	return result
}

// GetAll returns all templates concatenated with "---" separators in hierarchy order.
func (s *embeddedTemplateStore) GetAll() ([]byte, error) {
	var buf bytes.Buffer

	for i, kind := range hierarchyOrder {
		if i > 0 {
			buf.WriteString("\n---\n\n")
		}
		data, ok := s.templates[kind]
		if !ok {
			return nil, fmt.Errorf("missing template for kind %q in hierarchy", kind)
		}
		buf.Write(data)
	}

	return buf.Bytes(), nil
}

// normalize converts a kind string to its kebab-case form.
// Accepts exact kebab-case or exact PascalCase. All-caps, snake_case,
// and other forms are rejected.
func (s *embeddedTemplateStore) normalize(kind string) string {
	// Direct match (kebab-case)
	if _, ok := s.templates[kind]; ok {
		return kind
	}

	// Try PascalCase → kebab-case
	if kebab, ok := s.pascalToKebab[kind]; ok {
		return kebab
	}

	return ""
}
