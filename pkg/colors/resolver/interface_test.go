package resolver

import (
	"testing"
	"time"

	"devopsmaestro/pkg/nvimops/theme"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHierarchyLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    HierarchyLevel
		expected string
	}{
		{
			name:     "workspace level",
			level:    LevelWorkspace,
			expected: "workspace",
		},
		{
			name:     "app level",
			level:    LevelApp,
			expected: "app",
		},
		{
			name:     "domain level",
			level:    LevelDomain,
			expected: "domain",
		},
		{
			name:     "ecosystem level",
			level:    LevelEcosystem,
			expected: "ecosystem",
		},
		{
			name:     "global level",
			level:    LevelGlobal,
			expected: "global",
		},
		{
			name:     "unknown level",
			level:    HierarchyLevel(999),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestThemeResolution_IsLight(t *testing.T) {
	tests := []struct {
		name     string
		theme    *theme.Theme
		expected bool
	}{
		{
			name: "light theme",
			theme: &theme.Theme{
				Name:     "light-theme",
				Category: "light",
			},
			expected: true,
		},
		{
			name: "dark theme",
			theme: &theme.Theme{
				Name:     "dark-theme",
				Category: "dark",
			},
			expected: false,
		},
		{
			name: "theme with no category",
			theme: &theme.Theme{
				Name: "no-category-theme",
			},
			expected: false,
		},
		{
			name:     "nil theme",
			theme:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolution := &ThemeResolution{
				Theme: tt.theme,
			}
			result := resolution.IsLight()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestThemeResolution_GetEffectiveThemeName(t *testing.T) {
	tests := []struct {
		name     string
		theme    *theme.Theme
		expected string
	}{
		{
			name: "theme with name",
			theme: &theme.Theme{
				Name: "tokyonight-night",
			},
			expected: "tokyonight-night",
		},
		{
			name:     "nil theme returns default",
			theme:    nil,
			expected: DefaultTheme,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolution := &ThemeResolution{
				Theme: tt.theme,
			}
			result := resolution.GetEffectiveThemeName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestThemeResolution_GetSourceDescription(t *testing.T) {
	tests := []struct {
		name       string
		source     HierarchyLevel
		sourceName string
		expected   string
	}{
		{
			name:       "with source name",
			source:     LevelApp,
			sourceName: "my-app",
			expected:   "app 'my-app'",
		},
		{
			name:       "empty source name",
			source:     LevelEcosystem,
			sourceName: "",
			expected:   "ecosystem",
		},
		{
			name:       "global level",
			source:     LevelGlobal,
			sourceName: "global default",
			expected:   "global 'global default'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolution := &ThemeResolution{
				Source:     tt.source,
				SourceName: tt.sourceName,
			}
			result := resolution.GetSourceDescription()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestThemeStep_Structure(t *testing.T) {
	// Test that ThemeStep can be created and accessed properly
	step := ThemeStep{
		Level:     LevelApp,
		ObjectID:  123,
		Name:      "test-app",
		ThemeName: "tokyonight-night",
		Found:     true,
		Error:     "",
	}

	assert.Equal(t, LevelApp, step.Level)
	assert.Equal(t, 123, step.ObjectID)
	assert.Equal(t, "test-app", step.Name)
	assert.Equal(t, "tokyonight-night", step.ThemeName)
	assert.True(t, step.Found)
	assert.Empty(t, step.Error)
}

func TestThemeStep_WithError(t *testing.T) {
	step := ThemeStep{
		Level:     LevelWorkspace,
		ObjectID:  456,
		Name:      "test-workspace",
		ThemeName: "",
		Found:     false,
		Error:     "workspace not found",
	}

	assert.Equal(t, LevelWorkspace, step.Level)
	assert.Equal(t, 456, step.ObjectID)
	assert.Equal(t, "test-workspace", step.Name)
	assert.Empty(t, step.ThemeName)
	assert.False(t, step.Found)
	assert.Equal(t, "workspace not found", step.Error)
}

func TestThemeResolution_Complete(t *testing.T) {
	// Test a complete ThemeResolution structure
	now := time.Now()
	testTheme := &theme.Theme{
		Name:     "catppuccin-mocha",
		Category: "dark",
		Plugin:   theme.ThemePlugin{Repo: "catppuccin/nvim"},
		Colors: map[string]string{
			"bg": "#1e1e2e",
			"fg": "#cdd6f4",
		},
	}

	resolution := &ThemeResolution{
		Theme:      testTheme,
		Source:     LevelDomain,
		SourceName: "web-services",
		SourceID:   42,
		Path: []ThemeStep{
			{
				Level:    LevelWorkspace,
				ObjectID: 1,
				Name:     "my-workspace",
				Found:    false,
			},
			{
				Level:    LevelApp,
				ObjectID: 2,
				Name:     "my-app",
				Found:    false,
			},
			{
				Level:     LevelDomain,
				ObjectID:  42,
				Name:      "web-services",
				ThemeName: "catppuccin-mocha",
				Found:     true,
			},
		},
		CacheHit:   false,
		ResolvedAt: now,
	}

	// Verify all fields
	assert.Equal(t, testTheme, resolution.Theme)
	assert.Equal(t, LevelDomain, resolution.Source)
	assert.Equal(t, "web-services", resolution.SourceName)
	assert.Equal(t, 42, resolution.SourceID)
	assert.Len(t, resolution.Path, 3)
	assert.False(t, resolution.CacheHit)
	assert.Equal(t, now, resolution.ResolvedAt)

	// Test helper methods
	assert.False(t, resolution.IsLight())
	assert.Equal(t, "catppuccin-mocha", resolution.GetEffectiveThemeName())
	assert.Equal(t, "domain 'web-services'", resolution.GetSourceDescription())

	// Verify path structure
	require.Len(t, resolution.Path, 3)

	// Workspace step (not found)
	workspaceStep := resolution.Path[0]
	assert.Equal(t, LevelWorkspace, workspaceStep.Level)
	assert.Equal(t, 1, workspaceStep.ObjectID)
	assert.Equal(t, "my-workspace", workspaceStep.Name)
	assert.Empty(t, workspaceStep.ThemeName)
	assert.False(t, workspaceStep.Found)

	// App step (not found)
	appStep := resolution.Path[1]
	assert.Equal(t, LevelApp, appStep.Level)
	assert.Equal(t, 2, appStep.ObjectID)
	assert.Equal(t, "my-app", appStep.Name)
	assert.Empty(t, appStep.ThemeName)
	assert.False(t, appStep.Found)

	// Domain step (found)
	domainStep := resolution.Path[2]
	assert.Equal(t, LevelDomain, domainStep.Level)
	assert.Equal(t, 42, domainStep.ObjectID)
	assert.Equal(t, "web-services", domainStep.Name)
	assert.Equal(t, "catppuccin-mocha", domainStep.ThemeName)
	assert.True(t, domainStep.Found)
}

func TestDefaultTheme(t *testing.T) {
	// Verify the default theme constant
	assert.Equal(t, "coolnight-ocean", DefaultTheme)
}
