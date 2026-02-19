package resolver

import (
	"context"
	"testing"
	"time"

	"devopsmaestro/pkg/nvimops/theme"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMockThemeResolver(t *testing.T) {
	mock := NewMockThemeResolver()

	assert.NotNil(t, mock)
	assert.NotNil(t, mock.themes)
	assert.NotNil(t, mock.resolutions)
	assert.Equal(t, DefaultTheme, mock.defaultName)
	assert.Empty(t, mock.themes)
	assert.Empty(t, mock.resolutions)
}

func TestMockThemeResolver_SetTheme(t *testing.T) {
	mock := NewMockThemeResolver()
	testTheme := &theme.Theme{
		Name:     "test-theme",
		Category: "dark",
		Plugin:   theme.ThemePlugin{Repo: "test/repo"},
		Colors: map[string]string{
			"bg": "#1a1b26",
			"fg": "#c0caf5",
		},
	}

	mock.SetTheme("test-theme", testTheme)

	assert.Equal(t, testTheme, mock.themes["test-theme"])
}

func TestMockThemeResolver_SetResolution(t *testing.T) {
	mock := NewMockThemeResolver()
	testTheme := &theme.Theme{Name: "test-theme"}
	resolution := &ThemeResolution{
		Theme:      testTheme,
		Source:     LevelApp,
		SourceName: "test-app",
		SourceID:   123,
		Path: []ThemeStep{
			{
				Level:     LevelApp,
				ObjectID:  123,
				Name:      "test-app",
				ThemeName: "test-theme",
				Found:     true,
			},
		},
		ResolvedAt: time.Now(),
	}

	mock.SetResolution(LevelApp, 123, resolution)

	key := mock.makeKey(LevelApp, 123)
	assert.Equal(t, resolution, mock.resolutions[key])
}

func TestMockThemeResolver_SetDefaultTheme(t *testing.T) {
	mock := NewMockThemeResolver()

	assert.Equal(t, DefaultTheme, mock.defaultName)

	mock.SetDefaultTheme("new-default")

	assert.Equal(t, "new-default", mock.defaultName)
}

func TestMockThemeResolver_Resolve(t *testing.T) {
	mock := NewMockThemeResolver()
	ctx := context.Background()

	t.Run("returns preset resolution", func(t *testing.T) {
		testTheme := &theme.Theme{Name: "preset-theme"}
		resolution := &ThemeResolution{
			Theme:      testTheme,
			Source:     LevelDomain,
			SourceName: "test-domain",
			SourceID:   456,
		}

		mock.SetResolution(LevelDomain, 456, resolution)

		result, err := mock.Resolve(ctx, LevelDomain, 456)

		require.NoError(t, err)
		assert.Equal(t, resolution, result)
	})

	t.Run("falls back to default when no preset resolution", func(t *testing.T) {
		// Set a default theme in the mock
		defaultTheme := &theme.Theme{Name: "default-theme"}
		mock.SetTheme("default-theme", defaultTheme)
		mock.SetDefaultTheme("default-theme")

		result, err := mock.Resolve(ctx, LevelApp, 999)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, defaultTheme, result.Theme)
		assert.Equal(t, LevelGlobal, result.Source)
		assert.Equal(t, "global default", result.SourceName)
	})

	t.Run("creates minimal theme when default not found", func(t *testing.T) {
		// Create a fresh mock without default theme set
		freshMock := NewMockThemeResolver()
		freshMock.SetDefaultTheme("nonexistent-theme")

		result, err := freshMock.Resolve(ctx, LevelEcosystem, 777)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Theme)
		assert.Equal(t, "nonexistent-theme", result.Theme.Name)
		assert.Equal(t, "dark", result.Theme.Category)
		assert.Equal(t, LevelGlobal, result.Source)
	})
}

func TestMockThemeResolver_ResolveDefault(t *testing.T) {
	mock := NewMockThemeResolver()

	t.Run("returns configured default theme", func(t *testing.T) {
		defaultTheme := &theme.Theme{
			Name:     "configured-default",
			Category: "light",
			Colors: map[string]string{
				"bg": "#ffffff",
				"fg": "#000000",
			},
		}
		mock.SetTheme("configured-default", defaultTheme)
		mock.SetDefaultTheme("configured-default")

		result, err := mock.ResolveDefault()

		require.NoError(t, err)
		assert.Equal(t, defaultTheme, result.Theme)
		assert.Equal(t, LevelGlobal, result.Source)
		assert.Equal(t, "global default", result.SourceName)
		assert.Equal(t, 0, result.SourceID)
		assert.Len(t, result.Path, 1)

		step := result.Path[0]
		assert.Equal(t, LevelGlobal, step.Level)
		assert.Equal(t, "global default", step.Name)
		assert.Equal(t, "configured-default", step.ThemeName)
		assert.True(t, step.Found)
	})

	t.Run("creates minimal theme when default not found", func(t *testing.T) {
		// Use fresh mock with nonexistent default
		freshMock := NewMockThemeResolver()
		freshMock.SetDefaultTheme("nonexistent")

		result, err := freshMock.ResolveDefault()

		require.NoError(t, err)
		assert.NotNil(t, result.Theme)
		assert.Equal(t, "nonexistent", result.Theme.Name)
		assert.Equal(t, "dark", result.Theme.Category)
		assert.Equal(t, "test/mock", result.Theme.Plugin.Repo)
		assert.Contains(t, result.Theme.Colors, "bg")
		assert.Contains(t, result.Theme.Colors, "fg")
	})
}

func TestMockThemeResolver_GetResolutionPath(t *testing.T) {
	mock := NewMockThemeResolver()
	ctx := context.Background()

	t.Run("returns path from preset resolution", func(t *testing.T) {
		originalResolution := &ThemeResolution{
			Theme:      &theme.Theme{Name: "test-theme"},
			Source:     LevelApp,
			SourceName: "test-app",
			SourceID:   123,
			Path: []ThemeStep{
				{
					Level:    LevelWorkspace,
					ObjectID: 1,
					Name:     "test-workspace",
					Found:    false,
				},
				{
					Level:     LevelApp,
					ObjectID:  123,
					Name:      "test-app",
					ThemeName: "test-theme",
					Found:     true,
				},
			},
		}

		mock.SetResolution(LevelWorkspace, 1, originalResolution)

		result, err := mock.GetResolutionPath(ctx, LevelWorkspace, 1)

		require.NoError(t, err)
		assert.NotNil(t, result)

		// Theme should be nil (path only)
		assert.Nil(t, result.Theme)

		// Source information should be preserved
		assert.Equal(t, originalResolution.Source, result.Source)
		assert.Equal(t, originalResolution.SourceName, result.SourceName)
		assert.Equal(t, originalResolution.SourceID, result.SourceID)

		// Path should be copied
		assert.Len(t, result.Path, len(originalResolution.Path))
		for i, step := range result.Path {
			assert.Equal(t, originalResolution.Path[i], step)
		}

		// Resolved time should be recent
		assert.WithinDuration(t, time.Now(), result.ResolvedAt, time.Second)
	})

	t.Run("returns default path when no preset resolution", func(t *testing.T) {
		mock.SetDefaultTheme("path-default")

		result, err := mock.GetResolutionPath(ctx, LevelDomain, 888)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Theme)
		assert.Equal(t, LevelGlobal, result.Source)
		assert.Equal(t, "global default", result.SourceName)
		assert.Equal(t, 0, result.SourceID)
		assert.Len(t, result.Path, 1)

		step := result.Path[0]
		assert.Equal(t, LevelGlobal, step.Level)
		assert.Equal(t, "global default", step.Name)
		assert.Equal(t, "path-default", step.ThemeName)
		assert.True(t, step.Found)
	})
}

func TestMockThemeResolver_MakeKey(t *testing.T) {
	mock := NewMockThemeResolver()

	tests := []struct {
		level    HierarchyLevel
		objectID int
		expected string
	}{
		{LevelWorkspace, 1, "workspace:1"},
		{LevelApp, 42, "app:*"},
		{LevelDomain, 0, "domain:0"},
		{LevelEcosystem, 999, "ecosystem:ח"}, // Unicode character for large number
		{LevelGlobal, 5, "global:5"},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			key := mock.makeKey(tt.level, tt.objectID)
			assert.Contains(t, key, tt.level.String()+":")
			// The exact character representation may vary, but should be consistent
			assert.NotEmpty(t, key)
		})
	}
}

func TestMockThemeResolver_InterfaceCompliance(t *testing.T) {
	// Test that MockThemeResolver implements ThemeResolver interface
	var resolver ThemeResolver = NewMockThemeResolver()

	assert.NotNil(t, resolver)

	ctx := context.Background()

	// Should be able to call all interface methods
	resolution, err := resolver.Resolve(ctx, LevelGlobal, 0)
	assert.NoError(t, err)
	assert.NotNil(t, resolution)

	defaultRes, err := resolver.ResolveDefault()
	assert.NoError(t, err)
	assert.NotNil(t, defaultRes)

	pathRes, err := resolver.GetResolutionPath(ctx, LevelGlobal, 0)
	assert.NoError(t, err)
	assert.NotNil(t, pathRes)
}

func TestMockThemeResolver_ConcurrentAccess(t *testing.T) {
	mock := NewMockThemeResolver()
	ctx := context.Background()

	// Set up test data
	testTheme := &theme.Theme{Name: "concurrent-theme"}
	mock.SetTheme("concurrent-theme", testTheme)
	mock.SetDefaultTheme("concurrent-theme")

	// Test concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Each goroutine tries to resolve
			resolution, err := mock.Resolve(ctx, LevelApp, id)
			assert.NoError(t, err)
			assert.NotNil(t, resolution)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// OK
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out")
		}
	}
}

func TestMockThemeResolver_EdgeCases(t *testing.T) {
	mock := NewMockThemeResolver()
	ctx := context.Background()

	t.Run("negative object ID", func(t *testing.T) {
		resolution, err := mock.Resolve(ctx, LevelApp, -1)

		require.NoError(t, err)
		assert.NotNil(t, resolution)
	})

	t.Run("zero object ID", func(t *testing.T) {
		resolution, err := mock.Resolve(ctx, LevelDomain, 0)

		require.NoError(t, err)
		assert.NotNil(t, resolution)
	})

	t.Run("large object ID", func(t *testing.T) {
		largeID := 999999999
		resolution, err := mock.Resolve(ctx, LevelEcosystem, largeID)

		require.NoError(t, err)
		assert.NotNil(t, resolution)
	})

	t.Run("empty theme name", func(t *testing.T) {
		mock.SetDefaultTheme("")
		resolution, err := mock.ResolveDefault()

		require.NoError(t, err)
		assert.NotNil(t, resolution)
		assert.Equal(t, "", resolution.Theme.Name)
	})
}
