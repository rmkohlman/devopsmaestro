// Package resolver provides mock implementations for testing
package resolver

import (
	"context"
	"time"

	"devopsmaestro/pkg/nvimops/theme"
)

// MockThemeResolver provides a mock implementation for testing
type MockThemeResolver struct {
	themes      map[string]*theme.Theme
	defaultName string
	resolutions map[string]*ThemeResolution // key: "level:objectID"
}

// NewMockThemeResolver creates a mock resolver with default themes
func NewMockThemeResolver() *MockThemeResolver {
	return &MockThemeResolver{
		themes:      make(map[string]*theme.Theme),
		defaultName: DefaultTheme,
		resolutions: make(map[string]*ThemeResolution),
	}
}

// SetTheme adds a theme to the mock resolver
func (m *MockThemeResolver) SetTheme(name string, theme *theme.Theme) {
	m.themes[name] = theme
}

// SetResolution sets a predetermined resolution for a specific level and object ID
func (m *MockThemeResolver) SetResolution(level HierarchyLevel, objectID int, resolution *ThemeResolution) {
	key := m.makeKey(level, objectID)
	m.resolutions[key] = resolution
}

// SetDefaultTheme sets the default theme name
func (m *MockThemeResolver) SetDefaultTheme(name string) {
	m.defaultName = name
}

// Resolve implements ThemeResolver interface
func (m *MockThemeResolver) Resolve(ctx context.Context, level HierarchyLevel, objectID int) (*ThemeResolution, error) {
	key := m.makeKey(level, objectID)
	if resolution, ok := m.resolutions[key]; ok {
		return resolution, nil
	}

	// Return default resolution
	return m.ResolveDefault()
}

// ResolveDefault implements ThemeResolver interface
func (m *MockThemeResolver) ResolveDefault() (*ThemeResolution, error) {
	defaultTheme := m.themes[m.defaultName]
	if defaultTheme == nil {
		// Create a minimal theme if not found
		defaultTheme = &theme.Theme{
			Name:     m.defaultName,
			Category: "dark",
			Plugin:   theme.ThemePlugin{Repo: "test/mock"},
			Colors: map[string]string{
				"bg": "#1a1b26",
				"fg": "#c0caf5",
			},
		}
	}

	return &ThemeResolution{
		Theme:      defaultTheme,
		Source:     LevelGlobal,
		SourceName: "global default",
		SourceID:   0,
		Path: []ThemeStep{
			{
				Level:     LevelGlobal,
				Name:      "global default",
				ThemeName: m.defaultName,
				Found:     true,
			},
		},
		ResolvedAt: time.Now(),
	}, nil
}

// GetResolutionPath implements ThemeResolver interface
func (m *MockThemeResolver) GetResolutionPath(ctx context.Context, level HierarchyLevel, objectID int) (*ThemeResolution, error) {
	key := m.makeKey(level, objectID)
	if resolution, ok := m.resolutions[key]; ok {
		// Return copy without theme loaded
		pathOnlyResolution := &ThemeResolution{
			Source:     resolution.Source,
			SourceName: resolution.SourceName,
			SourceID:   resolution.SourceID,
			Path:       make([]ThemeStep, len(resolution.Path)),
			ResolvedAt: time.Now(),
		}
		copy(pathOnlyResolution.Path, resolution.Path)
		return pathOnlyResolution, nil
	}

	// Return default path
	return &ThemeResolution{
		Source:     LevelGlobal,
		SourceName: "global default",
		SourceID:   0,
		Path: []ThemeStep{
			{
				Level:     LevelGlobal,
				Name:      "global default",
				ThemeName: m.defaultName,
				Found:     true,
			},
		},
		ResolvedAt: time.Now(),
	}, nil
}

// makeKey creates a unique key for level and object ID combinations
func (m *MockThemeResolver) makeKey(level HierarchyLevel, objectID int) string {
	return level.String() + ":" + string(rune('0'+objectID))
}
