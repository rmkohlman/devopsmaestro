package nvimbridge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultNvimConfig(t *testing.T) {
	config := DefaultNvimConfig()

	assert.Equal(t, DefaultStructure, config.Structure)
	assert.Equal(t, "lazyvim", config.Structure)

	assert.Equal(t, DefaultPackage, config.PluginPackage)
	assert.Equal(t, "core", config.PluginPackage)

	assert.Equal(t, "", config.Theme) // Let theme resolution handle via cascade
	assert.Equal(t, "append", config.MergeMode)
	assert.Nil(t, config.Plugins) // Use package plugins only
}

func TestDefaultConstants(t *testing.T) {
	assert.Equal(t, "lazyvim", DefaultStructure)
	assert.Equal(t, "core", DefaultPackage)
}

func TestGetLanguagePackage(t *testing.T) {
	tests := []struct {
		name     string
		language string
		want     string
	}{
		{"golang maps to maestro-go", "golang", "maestro-go"},
		{"python maps to maestro-python", "python", "maestro-python"},
		{"rust maps to maestro-rust", "rust", "maestro-rust"},
		{"nodejs maps to maestro-node", "nodejs", "maestro-node"},
		{"java maps to maestro-java", "java", "maestro-java"},
		{"gleam maps to maestro-gleam", "gleam", "maestro-gleam"},
		{"dotnet maps to maestro-dotnet", "dotnet", "maestro-dotnet"},
		{"ruby maps to base maestro", "ruby", "maestro"},
		{"php maps to maestro-php", "php", "maestro-php"},
		{"kotlin maps to maestro-kotlin", "kotlin", "maestro-kotlin"},
		{"scala maps to maestro-scala", "scala", "maestro-scala"},
		{"elixir maps to maestro-elixir", "elixir", "maestro-elixir"},
		{"zig maps to maestro-zig", "zig", "maestro-zig"},
		{"dart maps to maestro-dart", "dart", "maestro-dart"},
		{"r maps to maestro-r", "r", "maestro-r"},
		{"haskell maps to maestro-haskell", "haskell", "maestro-haskell"},
		{"perl maps to maestro-perl", "perl", "maestro-perl"},
		{"unknown language returns empty", "cobol", ""},
		{"empty string returns empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetLanguagePackage(tt.language)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLanguagePackageMap_AllExpectedLanguages(t *testing.T) {
	// Verify every expected language has a mapping
	expectedLanguages := []string{"golang", "python", "rust", "nodejs", "java", "gleam", "dotnet", "ruby", "php", "kotlin", "scala", "elixir", "swift", "zig", "dart", "lua", "r", "haskell", "perl"}
	for _, lang := range expectedLanguages {
		_, ok := LanguagePackageMap[lang]
		assert.True(t, ok, "LanguagePackageMap should contain mapping for %q", lang)
	}
	// Verify exact count (no unexpected entries)
	assert.Equal(t, len(expectedLanguages), len(LanguagePackageMap),
		"LanguagePackageMap should have exactly %d entries", len(expectedLanguages))
}
