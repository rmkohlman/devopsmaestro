package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
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

// TestCompletionZsh_NoCompdefLine verifies that "dvm completion zsh" does NOT
// include the bare "compdef _dvm dvm" line that Cobra emits by default.
// That line breaks zsh fpath-based autoloading, causing tab completion to
// stop working and requiring a shell restart (issue #292).
func TestCompletionZsh_NoCompdefLine(t *testing.T) {
	// Call genZshCompletionFixed directly to avoid rootCmd state leakage
	// from previous tests (e.g., --help flag persisting on completionCmd).
	cmd := &cobra.Command{Use: "test-wrapper"}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := genZshCompletionFixed(cmd)
	require.NoError(t, err, "genZshCompletionFixed should not error")

	output := buf.String()
	assert.NotEmpty(t, output)

	// The #compdef header MUST be present (zsh autoload requires it)
	assert.Contains(t, output, "#compdef dvm",
		"zsh completion must have #compdef dvm header for autoload")

	// The bare "compdef _dvm dvm" line must NOT be present
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "compdef _") {
			t.Errorf("found bare compdef line %q — this breaks zsh fpath autoload", trimmed)
		}
	}

	// The _dvm function must still be defined
	assert.Contains(t, output, "_dvm()",
		"zsh completion must define the _dvm function")
}

// TestCompletionZsh_AutoloadGuard verifies the autoload guard is present at
// the end of the zsh completion output. This guard ensures the function runs
// when autoloaded by zsh but not when merely sourced for evaluation.
func TestCompletionZsh_AutoloadGuard(t *testing.T) {
	cmd := &cobra.Command{Use: "test-wrapper"}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := genZshCompletionFixed(cmd)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `if [ "$funcstack[1]" = "_dvm" ]; then`,
		"zsh completion must have funcstack autoload guard")
}

// TestShouldSkipAutoMigration_CompleteCommands verifies that Cobra's hidden
// __complete commands are skipped for auto-migration. These commands run on
// every TAB press and must not trigger database operations.
func TestShouldSkipAutoMigration_CompleteCommands(t *testing.T) {
	// Build a command tree that mirrors production: dvm -> __complete
	// CommandPath() returns "dvm __complete" for child commands.
	root := &cobra.Command{Use: "dvm"}
	completeChild := &cobra.Command{Use: "__complete"}
	root.AddCommand(completeChild)

	assert.True(t, shouldSkipAutoMigration(completeChild),
		"__complete should skip auto-migration (runs on every TAB press)")

	noDescChild := &cobra.Command{Use: "__completeNoDesc"}
	root.AddCommand(noDescChild)
	assert.True(t, shouldSkipAutoMigration(noDescChild),
		"__completeNoDesc should skip auto-migration")

	// Regular commands should NOT skip
	buildChild := &cobra.Command{Use: "build"}
	root.AddCommand(buildChild)
	assert.False(t, shouldSkipAutoMigration(buildChild),
		"build command should not skip auto-migration")
}
