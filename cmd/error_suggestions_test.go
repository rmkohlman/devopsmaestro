package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorWithSuggestion(t *testing.T) {
	tests := []struct {
		name        string
		msg         string
		suggestions []string
		wantContain []string
		wantAbsent  []string
	}{
		{
			name:        "no suggestions returns plain error",
			msg:         "something broke",
			suggestions: nil,
			wantContain: []string{"something broke"},
			wantAbsent:  []string{"Suggestions:"},
		},
		{
			name:        "single suggestion",
			msg:         "no active app context",
			suggestions: []string{"Set active app: dvm use app <name>"},
			wantContain: []string{
				"no active app context",
				"Suggestions:",
				"→ Set active app: dvm use app <name>",
			},
		},
		{
			name: "multiple suggestions",
			msg:  "workspace not found",
			suggestions: []string{
				"Check the spelling",
				"List workspaces: dvm get workspaces",
			},
			wantContain: []string{
				"workspace not found",
				"Suggestions:",
				"→ Check the spelling",
				"→ List workspaces: dvm get workspaces",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ErrorWithSuggestion(tc.msg, tc.suggestions...)
			require.Error(t, err)
			errStr := err.Error()
			for _, want := range tc.wantContain {
				assert.Contains(t, errStr, want)
			}
			for _, absent := range tc.wantAbsent {
				assert.NotContains(t, errStr, absent)
			}
		})
	}
}

func TestFormatSuggestions(t *testing.T) {
	tests := []struct {
		name        string
		suggestions []string
		want        string
	}{
		{
			name:        "empty returns empty string",
			suggestions: nil,
			want:        "",
		},
		{
			name:        "single suggestion",
			suggestions: []string{"do this thing"},
			want:        "Suggestions:\n  → do this thing",
		},
		{
			name:        "multiple suggestions",
			suggestions: []string{"first", "second", "third"},
			want:        "Suggestions:\n  → first\n  → second\n  → third",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatSuggestions(tc.suggestions...)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestSuggestNoActiveApp(t *testing.T) {
	suggestions := SuggestNoActiveApp()
	require.Len(t, suggestions, 3)
	assert.Contains(t, suggestions[0], "dvm use app")
	assert.Contains(t, suggestions[1], "dvm get apps")
	assert.Contains(t, suggestions[2], "--app")
}

func TestSuggestNoActiveWorkspace(t *testing.T) {
	suggestions := SuggestNoActiveWorkspace()
	require.Len(t, suggestions, 3)
	assert.Contains(t, suggestions[0], "dvm use workspace")
	assert.Contains(t, suggestions[1], "dvm get workspaces")
	assert.Contains(t, suggestions[2], "--workspace")
}

func TestSuggestNoActiveEcosystem(t *testing.T) {
	suggestions := SuggestNoActiveEcosystem()
	require.Len(t, suggestions, 3)
	assert.Contains(t, suggestions[0], "dvm use ecosystem")
	assert.Contains(t, suggestions[1], "dvm get ecosystems")
	assert.Contains(t, suggestions[2], "--ecosystem")
}

func TestSuggestNoActiveDomain(t *testing.T) {
	suggestions := SuggestNoActiveDomain()
	require.Len(t, suggestions, 3)
	assert.Contains(t, suggestions[0], "dvm use domain")
	assert.Contains(t, suggestions[1], "dvm get domains")
	assert.Contains(t, suggestions[2], "--domain")
}

func TestSuggestResourceNotFound(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		resourceName string
		listCmd      string
	}{
		{"workspace", "workspace", "staging", "dvm get workspaces"},
		{"app", "app", "portal", "dvm get apps"},
		{"ecosystem", "ecosystem", "prod", "dvm get ecosystems"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			suggestions := SuggestResourceNotFound(tc.resourceType, tc.resourceName, tc.listCmd)
			require.Len(t, suggestions, 2)
			assert.Contains(t, suggestions[0], tc.resourceName)
			assert.Contains(t, suggestions[1], tc.listCmd)
		})
	}
}

func TestSuggestWorkspaceNotFound(t *testing.T) {
	suggestions := SuggestWorkspaceNotFound("staging")
	require.Len(t, suggestions, 2)
	assert.Contains(t, suggestions[0], "staging")
	assert.Contains(t, suggestions[1], "dvm get workspaces")
}

func TestSuggestAppNotFound(t *testing.T) {
	suggestions := SuggestAppNotFound("portal")
	require.Len(t, suggestions, 2)
	assert.Contains(t, suggestions[0], "portal")
	assert.Contains(t, suggestions[1], "dvm get apps")
}

func TestSuggestNoContainerRuntime(t *testing.T) {
	suggestions := SuggestNoContainerRuntime()
	require.Len(t, suggestions, 2)
	assert.Contains(t, suggestions[0], "OrbStack")
	assert.Contains(t, suggestions[1], "daemon")
}

func TestSuggestWorkspaceNotBuilt(t *testing.T) {
	suggestions := SuggestWorkspaceNotBuilt()
	require.Len(t, suggestions, 2)
	assert.Contains(t, suggestions[0], "dvm build")
}

func TestSuggestAmbiguousWorkspace(t *testing.T) {
	suggestions := SuggestAmbiguousWorkspace()
	require.Len(t, suggestions, 2)
	assert.Contains(t, suggestions[0], "-e, -d, -a, -w")
	assert.Contains(t, suggestions[1], "dvm get workspaces")
}

func TestErrorWithSuggestion_Integration(t *testing.T) {
	// Test that ErrorWithSuggestion works well with predefined suggestion sets
	err := ErrorWithSuggestion("no active app context", SuggestNoActiveApp()...)
	require.Error(t, err)

	errStr := err.Error()
	// Should contain the original message
	assert.True(t, strings.HasPrefix(errStr, "no active app context"))
	// Should contain suggestions
	assert.Contains(t, errStr, "Suggestions:")
	assert.Contains(t, errStr, "dvm use app")
	assert.Contains(t, errStr, "dvm get apps")
	assert.Contains(t, errStr, "--app")
}

func TestFormatSuggestions_UsableWithRender(t *testing.T) {
	// FormatSuggestions returns a string suitable for render.Plain()
	formatted := FormatSuggestions(SuggestNoActiveWorkspace()...)
	assert.NotEmpty(t, formatted)
	lines := strings.Split(formatted, "\n")
	// Header + 3 suggestions = 4 lines
	assert.Len(t, lines, 4)
	assert.Equal(t, "Suggestions:", lines[0])
	for _, line := range lines[1:] {
		assert.True(t, strings.HasPrefix(line, "  → "), "suggestion line should be prefixed with arrow: %q", line)
	}
}
