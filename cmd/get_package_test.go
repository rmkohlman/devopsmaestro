// Package cmd — CLI tests for 'dvm get nvim-package' and 'dvm get terminal-package'.
// These tests are in RED state — the commands do not exist yet.
// File is .pending — CI skips it until the implementation exists.
package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetNvimPackage_CmdExists verifies the command variable is declared.
func TestGetNvimPackage_CmdExists(t *testing.T) {
	assert.NotNil(t, getNvimPackageCmd, "getNvimPackageCmd should not be nil")
}

// TestGetNvimPackage_RegisteredOnGetCmd verifies nvim-package is a get subcommand.
func TestGetNvimPackage_RegisteredOnGetCmd(t *testing.T) {
	require.NotNil(t, getCmd)
	require.NotNil(t, getNvimPackageCmd)

	found := false
	for _, sub := range getCmd.Commands() {
		if sub == getNvimPackageCmd {
			found = true
			break
		}
	}
	assert.True(t, found, "getNvimPackageCmd should be registered as a subcommand of getCmd")
}

// TestGetNvimPackage_Use verifies the Use field is "nvim-package".
func TestGetNvimPackage_Use(t *testing.T) {
	require.NotNil(t, getNvimPackageCmd)
	assert.Equal(t, "nvim-package", getNvimPackageCmd.Use)
}

// TestGetNvimPackage_HasShowCascadeFlag verifies --show-cascade flag is present.
func TestGetNvimPackage_HasShowCascadeFlag(t *testing.T) {
	require.NotNil(t, getNvimPackageCmd)
	assert.NotNil(t, getNvimPackageCmd.Flags().Lookup("show-cascade"),
		"--show-cascade flag should be present on get nvim-package")
}

// TestGetTerminalPackage_CmdExists verifies the command variable is declared.
func TestGetTerminalPackage_CmdExists(t *testing.T) {
	assert.NotNil(t, getTerminalPackageCmd, "getTerminalPackageCmd should not be nil")
}

// TestGetTerminalPackage_RegisteredOnGetCmd verifies terminal-package is a get subcommand.
func TestGetTerminalPackage_RegisteredOnGetCmd(t *testing.T) {
	require.NotNil(t, getCmd)
	require.NotNil(t, getTerminalPackageCmd)

	found := false
	for _, sub := range getCmd.Commands() {
		if sub == getTerminalPackageCmd {
			found = true
			break
		}
	}
	assert.True(t, found, "getTerminalPackageCmd should be registered as a subcommand of getCmd")
}

// TestGetTerminalPackage_Use verifies the Use field is "terminal-package".
func TestGetTerminalPackage_Use(t *testing.T) {
	require.NotNil(t, getTerminalPackageCmd)
	assert.Equal(t, "terminal-package", getTerminalPackageCmd.Use)
}

// TestGetTerminalPackage_HasShowCascadeFlag verifies --show-cascade flag is present.
func TestGetTerminalPackage_HasShowCascadeFlag(t *testing.T) {
	require.NotNil(t, getTerminalPackageCmd)
	assert.NotNil(t, getTerminalPackageCmd.Flags().Lookup("show-cascade"),
		"--show-cascade flag should be present on get terminal-package")
}
