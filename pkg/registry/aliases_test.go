package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestResolveAlias_KnownAliases tests resolution of known type aliases
func TestResolveAlias_KnownAliases(t *testing.T) {
	tests := []struct {
		name     string
		alias    string
		wantType string
	}{
		{
			name:     "oci resolves to zot",
			alias:    "oci",
			wantType: "zot",
		},
		{
			name:     "pypi resolves to devpi",
			alias:    "pypi",
			wantType: "devpi",
		},
		{
			name:     "npm resolves to verdaccio",
			alias:    "npm",
			wantType: "verdaccio",
		},
		{
			name:     "go resolves to athens",
			alias:    "go",
			wantType: "athens",
		},
		{
			name:     "http resolves to squid",
			alias:    "http",
			wantType: "squid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveAlias(tt.alias)
			assert.Equal(t, tt.wantType, got, "ResolveAlias(%q) should return %q", tt.alias, tt.wantType)
		})
	}
}

// TestResolveAlias_UnknownPassthrough tests that unknown aliases pass through unchanged
func TestResolveAlias_UnknownPassthrough(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "unknown type passes through",
			input: "custom-registry",
			want:  "custom-registry",
		},
		{
			name:  "zot passes through (already concrete type)",
			input: "zot",
			want:  "zot",
		},
		{
			name:  "athens passes through",
			input: "athens",
			want:  "athens",
		},
		{
			name:  "empty string passes through",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveAlias(tt.input)
			assert.Equal(t, tt.want, got, "ResolveAlias(%q) should pass through unchanged", tt.input)
		})
	}
}

// TestGetAliasForType tests reverse lookup of type to alias
func TestGetAliasForType(t *testing.T) {
	tests := []struct {
		name         string
		registryType string
		wantAlias    string
		wantFound    bool
	}{
		{
			name:         "zot has alias oci",
			registryType: "zot",
			wantAlias:    "oci",
			wantFound:    true,
		},
		{
			name:         "devpi has alias pypi",
			registryType: "devpi",
			wantAlias:    "pypi",
			wantFound:    true,
		},
		{
			name:         "verdaccio has alias npm",
			registryType: "verdaccio",
			wantAlias:    "npm",
			wantFound:    true,
		},
		{
			name:         "athens has alias go",
			registryType: "athens",
			wantAlias:    "go",
			wantFound:    true,
		},
		{
			name:         "squid has alias http",
			registryType: "squid",
			wantAlias:    "http",
			wantFound:    true,
		},
		{
			name:         "unknown type returns not found",
			registryType: "unknown",
			wantAlias:    "",
			wantFound:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alias, found := GetAliasForType(tt.registryType)
			assert.Equal(t, tt.wantFound, found, "GetAliasForType(%q) found flag", tt.registryType)
			if found {
				assert.Equal(t, tt.wantAlias, alias, "GetAliasForType(%q) should return alias %q", tt.registryType, tt.wantAlias)
			}
		})
	}
}

// TestAllAliases verifies all aliases are properly defined
func TestAllAliases(t *testing.T) {
	aliases := GetAllAliases()

	// Verify expected count
	assert.Equal(t, 5, len(aliases), "Should have exactly 5 type aliases")

	// Verify all expected aliases exist
	expectedAliases := map[string]string{
		"oci":  "zot",
		"pypi": "devpi",
		"npm":  "verdaccio",
		"go":   "athens",
		"http": "squid",
	}

	for alias, expectedType := range expectedAliases {
		actualType, found := aliases[alias]
		assert.True(t, found, "Alias %q should exist", alias)
		assert.Equal(t, expectedType, actualType, "Alias %q should map to %q", alias, expectedType)
	}
}

// TestAliasBidirectional verifies bidirectional alias <-> type mapping
func TestAliasBidirectional(t *testing.T) {
	tests := []struct {
		alias        string
		registryType string
	}{
		{"oci", "zot"},
		{"pypi", "devpi"},
		{"npm", "verdaccio"},
		{"go", "athens"},
		{"http", "squid"},
	}

	for _, tt := range tests {
		t.Run(tt.alias+"<->"+tt.registryType, func(t *testing.T) {
			// Forward: alias -> type
			resolvedType := ResolveAlias(tt.alias)
			assert.Equal(t, tt.registryType, resolvedType, "Alias %q should resolve to type %q", tt.alias, tt.registryType)

			// Reverse: type -> alias
			alias, found := GetAliasForType(tt.registryType)
			assert.True(t, found, "Type %q should have an alias", tt.registryType)
			assert.Equal(t, tt.alias, alias, "Type %q should map to alias %q", tt.registryType, tt.alias)
		})
	}
}
