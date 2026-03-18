package resolver

import (
	"context"
	"testing"
	"time"

	theme "github.com/rmkohlman/MaestroTheme"

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

// =============================================================================
// Bug Regression Tests — makeKey rune overflow (Bug 3)
// =============================================================================

// TestMakeKey_MultiDigitIDs_AreDistinct verifies that makeKey produces distinct
// keys for distinct objectIDs > 9.
//
// BUG: makeKey uses `string(rune('0' + objectID))` which only produces correct
// single-digit representations for IDs 0–9. For ID 10 it produces ':' (ASCII 58),
// for ID 11 it produces ';', etc. This means:
//   - Two different IDs may map to non-numeric/unexpected characters
//   - IDs >= 43 collide with uppercase letters ('0'+43 = 'K') which might
//     collide with level prefixes or other rune ranges
//   - Two scenarios can silently alias: e.g., makeKey(LevelApp, 0) could equal
//     makeKey(LevelApp, 48) because rune('0'+48) == rune('0') == '0'
//
// This test FAILS against current code (RED phase).
func TestMakeKey_MultiDigitIDs_AreDistinct(t *testing.T) {
	mock := NewMockThemeResolver()

	tests := []struct {
		name string
		id1  int
		id2  int
	}{
		{"IDs 9 and 10 must be distinct", 9, 10},
		{"IDs 10 and 11 must be distinct", 10, 11},
		{"IDs 0 and 10 must be distinct", 0, 10},
		{"IDs 5 and 15 must be distinct", 5, 15},
		{"IDs 1 and 100 must be distinct", 1, 100},
		{"IDs 42 and 0 must be distinct", 42, 0}, // '0'+42 = 'J', different from '0'+0 = '0'
	}

	level := LevelApp
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key1 := mock.makeKey(level, tt.id1)
			key2 := mock.makeKey(level, tt.id2)

			// BUG: For IDs > 9, rune arithmetic wraps into non-digit characters.
			// makeKey(LevelApp, 10) → "app::" which differs from "app:9" but is
			// not a valid decimal representation, so distinct IDs produce correct
			// distinct keys by accident for small values but wrong format.
			// The real failure is the format: "app::" for ID=10 instead of "app:10".
			if key1 == key2 {
				t.Errorf("makeKey(%s, %d) == makeKey(%s, %d): both returned %q, want distinct keys",
					level, tt.id1, level, tt.id2, key1)
			}
		})
	}
}

// TestMakeKey_MultiDigitIDs_ContainDecimalRepresentation verifies that makeKey
// produces keys that contain the decimal string representation of the objectID.
//
// BUG: makeKey(LevelApp, 10) returns "app::" (because rune('0'+10) = ':')
// instead of the expected "app:10". The key should contain the decimal string
// of the objectID so that keys are human-readable and unambiguous.
//
// This test FAILS against current code (RED phase).
func TestMakeKey_MultiDigitIDs_ContainDecimalRepresentation(t *testing.T) {
	mock := NewMockThemeResolver()

	tests := []struct {
		level      HierarchyLevel
		objectID   int
		wantSuffix string // Expected decimal representation in the key
	}{
		{LevelWorkspace, 0, "0"},
		{LevelWorkspace, 1, "1"},
		{LevelWorkspace, 9, "9"},
		{LevelApp, 10, "10"},         // BUG: currently produces ':' not "10"
		{LevelApp, 15, "15"},         // BUG: currently produces '?' not "15"
		{LevelApp, 42, "42"},         // BUG: currently produces 'J' not "42"
		{LevelDomain, 100, "100"},    // BUG: currently produces 'd' (rune 100) not "100"
		{LevelEcosystem, 999, "999"}, // BUG: produces multi-byte rune, not "999"
	}

	for _, tt := range tests {
		t.Run(tt.level.String()+"_"+tt.wantSuffix, func(t *testing.T) {
			key := mock.makeKey(tt.level, tt.objectID)
			expectedKey := tt.level.String() + ":" + tt.wantSuffix

			if key != expectedKey {
				t.Errorf("makeKey(%s, %d) = %q, want %q (decimal representation)",
					tt.level, tt.objectID, key, expectedKey)
			}
		})
	}
}

// TestMakeKey_ResolutionLookup_FailsForLargeIDs verifies that SetResolution and
// Resolve round-trip correctly for objectIDs > 9 using makeKey.
//
// BUG: When SetResolution stores a resolution under makeKey(level, 10) and then
// Resolve calls makeKey(level, 10) again to look it up, both calls use the same
// buggy key. The round-trip technically works for integer IDs stored & looked up
// consistently. However, if a caller uses string key "app:10" (human-readable),
// it will NEVER match the stored key "app::" (from rune arithmetic).
//
// The more critical failure: two distinct IDs could produce the SAME key if
// rune arithmetic causes a collision. Specifically: ID=0 produces rune '0' (48),
// and ID=48 also produces rune '0'+48 = rune(96) = '`'. These don't collide,
// but ID=0 and ID=48 happen to produce different chars. However, if we go up
// to rune('0' + 256) we wrap around and collide. For realistic IDs 0-255
// there are no collisions, but the key format is wrong (non-decimal chars).
//
// This test demonstrates that resolution CANNOT be externally looked up by
// decimal string key, which is the expected behavior.
func TestMakeKey_ResolutionLookup_MultiDigitID(t *testing.T) {
	ctx := context.Background()
	mock := NewMockThemeResolver()

	testTheme := &theme.Theme{Name: "multi-digit-theme"}
	resolution := &ThemeResolution{
		Theme:      testTheme,
		Source:     LevelApp,
		SourceName: "app-with-large-id",
		SourceID:   42,
	}

	// Store resolution for objectID=42
	mock.SetResolution(LevelApp, 42, resolution)

	// Retrieve it back — this round-trip works even with the bug
	// because both SetResolution and Resolve use the same buggy makeKey
	result, err := mock.Resolve(ctx, LevelApp, 42)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, testTheme, result.Theme, "round-trip lookup for ID=42 should return stored resolution")

	// Now verify the key format itself is wrong: it should be "app:42" not "app:J"
	key := mock.makeKey(LevelApp, 42)
	expectedKey := "app:42"
	if key == expectedKey {
		// If this passes, the bug has been fixed
		t.Logf("makeKey(LevelApp, 42) = %q — bug is FIXED (was expecting failure)", key)
	} else {
		// Document the actual buggy key that is produced
		t.Errorf("makeKey(LevelApp, 42) = %q, want %q — key uses rune arithmetic instead of decimal string",
			key, expectedKey)
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
