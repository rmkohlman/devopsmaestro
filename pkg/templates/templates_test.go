package templates_test

import (
	"strings"
	"testing"

	"devopsmaestro/pkg/templates"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// allExpectedKinds is the canonical set of 15 resource kinds for templates.
var allExpectedKinds = []string{
	"app",
	"credential",
	"custom-resource-definition",
	"domain",
	"ecosystem",
	"gitrepo",
	"global-defaults",
	"nvim-package",
	"nvim-plugin",
	"nvim-theme",
	"registry",
	"terminal-package",
	"terminal-plugin",
	"terminal-prompt",
	"workspace",
}

// hierarchyOrder is the expected ordering from GetAll() as specified in the architecture review.
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

// kindHeader is a minimal struct for validating apiVersion and kind fields.
type kindHeader struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
}

// TestNewTemplateStore verifies that the factory constructor succeeds.
func TestNewTemplateStore(t *testing.T) {
	store, err := templates.NewTemplateStore()
	require.NoError(t, err, "NewTemplateStore should not return an error")
	require.NotNil(t, store, "NewTemplateStore should return a non-nil store")
}

// TestTemplateStoreGet verifies that each known kind returns non-empty YAML bytes
// with valid apiVersion and kind fields.
func TestTemplateStoreGet(t *testing.T) {
	store, err := templates.NewTemplateStore()
	require.NoError(t, err)

	for _, kind := range allExpectedKinds {
		t.Run(kind, func(t *testing.T) {
			data, err := store.Get(kind)
			require.NoError(t, err, "Get(%q) should not error", kind)
			assert.NotEmpty(t, data, "Get(%q) should return non-empty bytes", kind)

			var header kindHeader
			err = yaml.Unmarshal(data, &header)
			require.NoError(t, err, "Get(%q) output must be valid YAML", kind)

			assert.NotEmpty(t, header.APIVersion,
				"template %q must have non-empty apiVersion field", kind)
			assert.NotEmpty(t, header.Kind,
				"template %q must have non-empty kind field", kind)
		})
	}
}

// TestTemplateStoreGetNormalization verifies both kebab-case and PascalCase
// kind names are accepted and return the same template bytes.
func TestTemplateStoreGetNormalization(t *testing.T) {
	store, err := templates.NewTemplateStore()
	require.NoError(t, err)

	tests := []struct {
		kebab  string
		pascal string
	}{
		{"nvim-plugin", "NvimPlugin"},
		{"nvim-theme", "NvimTheme"},
		{"nvim-package", "NvimPackage"},
		{"terminal-prompt", "TerminalPrompt"},
		{"terminal-package", "TerminalPackage"},
		{"terminal-plugin", "TerminalPlugin"},
		{"custom-resource-definition", "CustomResourceDefinition"},
		{"global-defaults", "GlobalDefaults"},
		{"ecosystem", "Ecosystem"},
		{"domain", "Domain"},
		{"app", "App"},
		{"workspace", "Workspace"},
		{"credential", "Credential"},
		{"registry", "Registry"},
		{"gitrepo", "Gitrepo"},
	}

	for _, tt := range tests {
		t.Run(tt.kebab, func(t *testing.T) {
			kebabData, err := store.Get(tt.kebab)
			require.NoError(t, err, "Get(%q) should succeed", tt.kebab)

			pascalData, err := store.Get(tt.pascal)
			require.NoError(t, err, "Get(%q) should succeed", tt.pascal)

			assert.Equal(t, kebabData, pascalData,
				"kebab-case %q and PascalCase %q should return identical template bytes",
				tt.kebab, tt.pascal)
		})
	}
}

// TestTemplateStoreGetUnknownKind verifies that Get returns an error for
// unrecognized kind names.
func TestTemplateStoreGetUnknownKind(t *testing.T) {
	store, err := templates.NewTemplateStore()
	require.NoError(t, err)

	unknownKinds := []string{
		"notakind",
		"",
		"ECOSYSTEM",
		"nvim_plugin",
		"nvim plugin",
		"unknown-resource",
	}

	for _, kind := range unknownKinds {
		t.Run(kind, func(t *testing.T) {
			data, err := store.Get(kind)
			assert.Error(t, err, "Get(%q) should return an error for unknown kind", kind)
			assert.Nil(t, data, "Get(%q) should return nil data for unknown kind", kind)
		})
	}
}

// TestTemplateStoreList verifies that List returns all 15 kind names, sorted.
func TestTemplateStoreList(t *testing.T) {
	store, err := templates.NewTemplateStore()
	require.NoError(t, err)

	kinds := store.List()
	require.Equal(t, len(allExpectedKinds), len(kinds),
		"List() should return exactly %d kinds", len(allExpectedKinds))

	// Verify sorted order
	for i, want := range allExpectedKinds {
		assert.Equal(t, want, kinds[i],
			"List()[%d] = %q, want %q (expected alphabetically sorted)", i, kinds[i], want)
	}
}

// TestTemplateStoreGetAll verifies that GetAll returns a multi-document YAML
// containing all templates separated by "---" separators.
func TestTemplateStoreGetAll(t *testing.T) {
	store, err := templates.NewTemplateStore()
	require.NoError(t, err)

	data, err := store.GetAll()
	require.NoError(t, err, "GetAll() should not error")
	assert.NotEmpty(t, data, "GetAll() should return non-empty bytes")

	output := string(data)

	// Must contain "---" document separators
	assert.Contains(t, output, "---",
		"GetAll() output must contain YAML document separators")

	// Verify all kinds appear somewhere in the combined output
	for _, kind := range allExpectedKinds {
		assert.True(t, strings.Contains(output, kind),
			"GetAll() output should contain kind %q", kind)
	}

	// The number of "---" separators should be (N-1) where N = number of templates
	separatorCount := strings.Count(output, "\n---\n")
	assert.Equal(t, len(allExpectedKinds)-1, separatorCount,
		"GetAll() should have %d document separators for %d templates",
		len(allExpectedKinds)-1, len(allExpectedKinds))
}

// TestTemplateStoreGetAllOrdering verifies that GetAll returns templates in the
// correct hierarchy order: ecosystem → domain → app → workspace → infra →
// nvim → terminal → meta.
func TestTemplateStoreGetAllOrdering(t *testing.T) {
	store, err := templates.NewTemplateStore()
	require.NoError(t, err)

	data, err := store.GetAll()
	require.NoError(t, err)

	output := string(data)

	// Verify each kind appears in the output in the expected order
	prevIdx := -1
	for _, kind := range hierarchyOrder {
		idx := strings.Index(output, kind)
		assert.True(t, idx >= 0, "GetAll() output must contain %q", kind)
		assert.Greater(t, idx, prevIdx,
			"kind %q must appear after previous kinds in GetAll() output", kind)
		if idx > prevIdx {
			prevIdx = idx
		}
	}
}

// TestAllTemplatesAreValidYAML iterates all embedded templates and verifies
// each can be parsed by yaml.Unmarshal with valid apiVersion and kind fields.
// This is the sync contract that dvm apply depends on.
func TestAllTemplatesAreValidYAML(t *testing.T) {
	store, err := templates.NewTemplateStore()
	require.NoError(t, err)

	for _, kind := range store.List() {
		t.Run(kind, func(t *testing.T) {
			data, err := store.Get(kind)
			require.NoError(t, err, "Get(%q) must succeed", kind)

			// Unmarshal into generic map to verify well-formed YAML
			var doc map[string]interface{}
			err = yaml.Unmarshal(data, &doc)
			require.NoError(t, err, "template %q must be valid YAML", kind)
			require.NotNil(t, doc, "template %q must unmarshal to non-nil document", kind)

			// Verify required top-level fields
			apiVersion, ok := doc["apiVersion"]
			require.True(t, ok, "template %q must have 'apiVersion' field", kind)
			assert.NotEmpty(t, apiVersion, "template %q apiVersion must not be empty", kind)

			kindField, ok := doc["kind"]
			require.True(t, ok, "template %q must have 'kind' field", kind)
			assert.NotEmpty(t, kindField, "template %q kind must not be empty", kind)
		})
	}
}
