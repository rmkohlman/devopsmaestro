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
	assert.Equal(t, "A", allFlag.Shorthand) // Changed from -a to -A to free up -a for --app
	assert.Equal(t, "false", allFlag.DefValue)

	// Verify hierarchy flags are registered
	ecoFlag := detachCmd.Flags().Lookup("ecosystem")
	assert.NotNil(t, ecoFlag)
	assert.Equal(t, "e", ecoFlag.Shorthand)

	appFlag := detachCmd.Flags().Lookup("app")
	assert.NotNil(t, appFlag)
	assert.Equal(t, "a", appFlag.Shorthand)
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

func TestContainsRunning(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "running status",
			status:   "Up 5 minutes",
			expected: true,
		},
		{
			name:     "up status short",
			status:   "Up",
			expected: true,
		},
		{
			name:     "exited status",
			status:   "Exited (0) 5 minutes ago",
			expected: false,
		},
		{
			name:     "created status",
			status:   "Created",
			expected: false,
		},
		{
			name:     "empty status",
			status:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsRunning(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}
