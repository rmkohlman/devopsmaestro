package models

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkspace_GetTerminalPlugins(t *testing.T) {
	tests := []struct {
		name     string
		plugins  sql.NullString
		expected []string
	}{
		{
			name:     "empty plugins returns empty slice",
			plugins:  sql.NullString{Valid: false},
			expected: []string{},
		},
		{
			name:     "null string returns empty slice",
			plugins:  sql.NullString{String: "", Valid: false},
			expected: []string{},
		},
		{
			name:     "valid JSON array",
			plugins:  sql.NullString{String: `["plugin1","plugin2"]`, Valid: true},
			expected: []string{"plugin1", "plugin2"},
		},
		{
			name:     "single plugin",
			plugins:  sql.NullString{String: `["zsh-autosuggestions"]`, Valid: true},
			expected: []string{"zsh-autosuggestions"},
		},
		{
			name:     "invalid JSON returns empty slice",
			plugins:  sql.NullString{String: `not-valid-json`, Valid: true},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Workspace{
				TerminalPlugins: tt.plugins,
			}

			result := w.GetTerminalPlugins()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWorkspace_SetTerminalPlugins(t *testing.T) {
	tests := []struct {
		name            string
		plugins         []string
		expectedValid   bool
		expectedContain string
	}{
		{
			name:          "empty slice sets null",
			plugins:       []string{},
			expectedValid: false,
		},
		{
			name:          "nil slice sets null",
			plugins:       nil,
			expectedValid: false,
		},
		{
			name:            "single plugin",
			plugins:         []string{"plugin1"},
			expectedValid:   true,
			expectedContain: "plugin1",
		},
		{
			name:            "multiple plugins",
			plugins:         []string{"plugin1", "plugin2", "plugin3"},
			expectedValid:   true,
			expectedContain: "plugin1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Workspace{}
			w.SetTerminalPlugins(tt.plugins)

			assert.Equal(t, tt.expectedValid, w.TerminalPlugins.Valid)

			if tt.expectedValid {
				assert.Contains(t, w.TerminalPlugins.String, tt.expectedContain)
				// Verify it can be unmarshaled back
				retrieved := w.GetTerminalPlugins()
				assert.Equal(t, tt.plugins, retrieved)
			}
		})
	}
}

func TestWorkspace_SetAndGetTerminalPlugins_RoundTrip(t *testing.T) {
	w := &Workspace{}

	// Test round trip with various plugin lists
	testCases := [][]string{
		{"zsh-autosuggestions"},
		{"zsh-autosuggestions", "zsh-syntax-highlighting"},
		{"plugin-a", "plugin-b", "plugin-c", "plugin-d"},
	}

	for _, plugins := range testCases {
		w.SetTerminalPlugins(plugins)
		retrieved := w.GetTerminalPlugins()
		assert.Equal(t, plugins, retrieved, "round trip should preserve plugin list")
	}

	// Test clearing
	w.SetTerminalPlugins([]string{})
	assert.False(t, w.TerminalPlugins.Valid, "clearing should set to null")
	assert.Empty(t, w.GetTerminalPlugins(), "cleared plugins should return empty slice")
}
