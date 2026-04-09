// Package cmd — CLI tests for 'dvm set terminal-package' command.
// These tests are in RED state — setTerminalPackageCmd does not exist yet.
// File is .pending — CI skips it until the implementation exists.
package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetTerminalPackageCmd_Exists verifies the command variable is declared.
func TestSetTerminalPackageCmd_Exists(t *testing.T) {
	assert.NotNil(t, setTerminalPackageCmd, "setTerminalPackageCmd should not be nil")
}

// TestSetTerminalPackageCmd_Use verifies the Use field is "terminal-package".
func TestSetTerminalPackageCmd_Use(t *testing.T) {
	require.NotNil(t, setTerminalPackageCmd)
	assert.Equal(t, "terminal-package <name>", setTerminalPackageCmd.Use,
		"Use should follow the '<resource> <arg>' pattern")
}

// TestSetTerminalPackageCmd_RegisteredOnSetCmd verifies it is a subcommand of setCmd.
func TestSetTerminalPackageCmd_RegisteredOnSetCmd(t *testing.T) {
	require.NotNil(t, setCmd)
	require.NotNil(t, setTerminalPackageCmd)

	found := false
	for _, sub := range setCmd.Commands() {
		if sub == setTerminalPackageCmd {
			found = true
			break
		}
	}
	assert.True(t, found, "setTerminalPackageCmd should be registered as a subcommand of setCmd")
}

// TestSetTerminalPackageCmd_HasHierarchyFlags verifies all scope flags are present.
func TestSetTerminalPackageCmd_HasHierarchyFlags(t *testing.T) {
	require.NotNil(t, setTerminalPackageCmd)

	flags := setTerminalPackageCmd.Flags()

	assert.NotNil(t, flags.Lookup("global"),
		"--global flag should be present")
	assert.NotNil(t, flags.Lookup("ecosystem"),
		"--ecosystem flag should be present")
	assert.NotNil(t, flags.Lookup("domain"),
		"--domain flag should be present")
	assert.NotNil(t, flags.Lookup("app"),
		"--app flag should be present")
	assert.NotNil(t, flags.Lookup("workspace"),
		"--workspace flag should be present")
}

// TestSetTerminalPackageCmd_HasDryRunFlag verifies --dry-run is present.
func TestSetTerminalPackageCmd_HasDryRunFlag(t *testing.T) {
	require.NotNil(t, setTerminalPackageCmd)
	assert.NotNil(t, setTerminalPackageCmd.Flags().Lookup("dry-run"),
		"--dry-run flag should be present")
}

// TestSetTerminalPackageCmd_HasShowCascadeFlag verifies --show-cascade is present.
func TestSetTerminalPackageCmd_HasShowCascadeFlag(t *testing.T) {
	require.NotNil(t, setTerminalPackageCmd)
	assert.NotNil(t, setTerminalPackageCmd.Flags().Lookup("show-cascade"),
		"--show-cascade flag should be present")
}
