package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompletionCommand_GeneratesBash verifies that dvm can generate bash
// completion scripts via the Cobra GenBashCompletion method.
func TestCompletionCommand_GeneratesBash(t *testing.T) {
	buf := new(bytes.Buffer)
	err := rootCmd.GenBashCompletion(buf)
	require.NoError(t, err)
	assert.NotEmpty(t, buf.String())
	assert.Contains(t, buf.String(), "bash completion", "output should be a bash completion script")
}

// TestCompletionCommand_GeneratesZsh verifies zsh completion generation.
func TestCompletionCommand_GeneratesZsh(t *testing.T) {
	buf := new(bytes.Buffer)
	err := rootCmd.GenZshCompletion(buf)
	require.NoError(t, err)
	assert.NotEmpty(t, buf.String())
	assert.Contains(t, buf.String(), "zsh completion", "output should be a zsh completion script")
}

// TestCompletionCommand_GeneratesFish verifies fish completion generation.
func TestCompletionCommand_GeneratesFish(t *testing.T) {
	buf := new(bytes.Buffer)
	err := rootCmd.GenFishCompletion(buf, true)
	require.NoError(t, err)
	assert.NotEmpty(t, buf.String())
	assert.Contains(t, buf.String(), "fish", "output should be a fish completion script")
}

// TestCompletionCommand_GeneratesPowerShell verifies powershell completion generation.
func TestCompletionCommand_GeneratesPowerShell(t *testing.T) {
	buf := new(bytes.Buffer)
	err := rootCmd.GenPowerShellCompletionWithDesc(buf)
	require.NoError(t, err)
	assert.NotEmpty(t, buf.String())
}

// TestCompletionCommand_SubcommandExists verifies that 'dvm completion bash'
// runs successfully via Execute, confirming the completion command is registered.
func TestCompletionCommand_SubcommandExists(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(new(bytes.Buffer))
	rootCmd.SetArgs([]string{"completion", "bash"})

	err := rootCmd.Execute()
	require.NoError(t, err, "dvm completion bash should not error")
	assert.NotEmpty(t, buf.String(), "dvm completion bash should produce non-empty output")

	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
}

// TestCompletionCommand_HelpWorks verifies that 'dvm completion --help'
// runs without error, confirming the completion command is registered.
func TestCompletionCommand_HelpWorks(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(new(bytes.Buffer))
	rootCmd.SetArgs([]string{"completion", "--help"})

	err := rootCmd.Execute()
	require.NoError(t, err, "dvm completion --help should not error")
	assert.Contains(t, buf.String(), "completion", "help output should mention completion")

	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
}
