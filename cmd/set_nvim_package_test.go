// Package cmd — CLI tests for 'dvm set nvim-package' command.
// These tests are in RED state — setNvimPackageCmd does not exist yet.
// File is .pending — CI skips it until the implementation exists.
package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetNvimPackageCmd_Exists verifies the command variable is declared.
func TestSetNvimPackageCmd_Exists(t *testing.T) {
	assert.NotNil(t, setNvimPackageCmd, "setNvimPackageCmd should not be nil")
}

// TestSetNvimPackageCmd_Use verifies the Use field is "nvim-package".
func TestSetNvimPackageCmd_Use(t *testing.T) {
	require.NotNil(t, setNvimPackageCmd)
	assert.Equal(t, "nvim-package <name>", setNvimPackageCmd.Use,
		"Use should follow the '<resource> <arg>' pattern")
}

// TestSetNvimPackageCmd_RegisteredOnSetCmd verifies it is a subcommand of setCmd.
func TestSetNvimPackageCmd_RegisteredOnSetCmd(t *testing.T) {
	require.NotNil(t, setCmd)
	require.NotNil(t, setNvimPackageCmd)

	found := false
	for _, sub := range setCmd.Commands() {
		if sub == setNvimPackageCmd {
			found = true
			break
		}
	}
	assert.True(t, found, "setNvimPackageCmd should be registered as a subcommand of setCmd")
}

// TestSetNvimPackageCmd_HasHierarchyFlags verifies all scope flags are present.
func TestSetNvimPackageCmd_HasHierarchyFlags(t *testing.T) {
	require.NotNil(t, setNvimPackageCmd)

	flags := setNvimPackageCmd.Flags()

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

// TestSetNvimPackageCmd_HasDryRunFlag verifies --dry-run is present.
func TestSetNvimPackageCmd_HasDryRunFlag(t *testing.T) {
	require.NotNil(t, setNvimPackageCmd)
	assert.NotNil(t, setNvimPackageCmd.Flags().Lookup("dry-run"),
		"--dry-run flag should be present")
}

// TestSetNvimPackageCmd_HasShowCascadeFlag verifies --show-cascade is present.
func TestSetNvimPackageCmd_HasShowCascadeFlag(t *testing.T) {
	require.NotNil(t, setNvimPackageCmd)
	assert.NotNil(t, setNvimPackageCmd.Flags().Lookup("show-cascade"),
		"--show-cascade flag should be present")
}
