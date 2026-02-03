package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetachCommand(t *testing.T) {
	// Test that the detach command exists
	assert.NotNil(t, detachCmd)
	assert.Equal(t, "detach", detachCmd.Use)
	assert.Contains(t, detachCmd.Short, "Stop")
}

func TestDetachCommandHelp(t *testing.T) {
	// Verify help text contains useful information
	helpText := detachCmd.Long

	assert.Contains(t, helpText, "workspace")
	assert.Contains(t, helpText, "dvm attach")
	assert.Contains(t, helpText, "--all")
}

func TestDetachCommandFlags(t *testing.T) {
	// Verify flags are registered
	allFlag := detachCmd.Flags().Lookup("all")
	assert.NotNil(t, allFlag)
	assert.Equal(t, "a", allFlag.Shorthand)
	assert.Equal(t, "false", allFlag.DefValue)
}

func TestDetachIsSubcommandOfRoot(t *testing.T) {
	// Verify detach is registered as a root command
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "detach" {
			found = true
			break
		}
	}
	assert.True(t, found, "detach should be a subcommand of root")
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single line",
			input:    "container1",
			expected: []string{"container1"},
		},
		{
			name:     "multiple lines",
			input:    "container1\ncontainer2\ncontainer3",
			expected: []string{"container1", "container2", "container3"},
		},
		{
			name:     "with trailing newline",
			input:    "container1\ncontainer2\n",
			expected: []string{"container1", "container2"},
		},
		{
			name:     "with empty lines",
			input:    "container1\n\ncontainer2\n\n",
			expected: []string{"container1", "container2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitLines(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
