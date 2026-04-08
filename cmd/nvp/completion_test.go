package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompletionCommand_Exists verifies that the 'completion' command is
// registered on the nvp root command.
func TestCompletionCommand_Exists(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "completion" {
			found = true
			break
		}
	}
	assert.True(t, found, "nvp should have a 'completion' command registered")
}

// TestCompletionCommand_ValidArgs verifies that the completion command
// accepts bash, zsh, fish, and powershell as valid arguments.
func TestCompletionCommand_ValidArgs(t *testing.T) {
	assert.Equal(t, []string{"bash", "zsh", "fish", "powershell"}, completionCmd.ValidArgs)
}

// TestCompletionCommand_ProducesOutput verifies that each shell completion
// generates non-empty output.
func TestCompletionCommand_ProducesOutput(t *testing.T) {
	shells := []string{"bash", "zsh", "fish", "powershell"}
	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(new(bytes.Buffer))
			rootCmd.SetArgs([]string{"completion", shell})

			err := rootCmd.Execute()
			require.NoError(t, err, "nvp completion %s should not error", shell)
			assert.NotEmpty(t, buf.String(), "nvp completion %s should produce non-empty output", shell)
		})
	}
}

// TestCompletionCommand_HelpText verifies the Long description contains
// useful installation instructions.
func TestCompletionCommand_HelpText(t *testing.T) {
	assert.Contains(t, completionCmd.Long, "source <(nvp completion",
		"Long description should include source <() usage example")
	assert.Contains(t, completionCmd.Long, "powershell",
		"Long description should mention powershell")
}
