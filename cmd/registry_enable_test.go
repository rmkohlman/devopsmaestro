package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// TDD Phase 2 (RED): Registry Enable/Disable Command Tests
// =============================================================================
// These tests verify the NEW registry enable/disable command structure:
//   - dvm registry enable <type>
//   - dvm registry disable <type>
//
// These commands enable/disable registry types (oci, pypi, npm, go, http)
// and manage lifecycle settings (persistent, on-demand, manual).
//
// These tests will FAIL until implementation is added.
// =============================================================================

// ========== Registry Parent Command Tests ==========

func TestRegistryCmd_Exists(t *testing.T) {
	// Test that registryCmd parent command exists
	assert.NotNil(t, registryCmd, "registryCmd should exist")
}

func TestRegistryCmd_HasCorrectUse(t *testing.T) {
	// Test Use field
	assert.Equal(t, "registry", registryCmd.Use)
}

func TestRegistryCmd_RegisteredToRoot(t *testing.T) {
	// Test that registryCmd is registered as child of rootCmd
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "registry" {
			found = true
			break
		}
	}
	assert.True(t, found, "registryCmd should be registered to rootCmd")
}

func TestRegistryCmd_HasCorrectSubcommands(t *testing.T) {
	// Test that registryCmd has enable, disable, set-default, get-defaults subcommands
	subcommands := []string{"enable", "disable", "set-default", "get-defaults"}
	for _, expectedSub := range subcommands {
		found := false
		for _, sub := range registryCmd.Commands() {
			if sub.Name() == expectedSub {
				found = true
				break
			}
		}
		assert.True(t, found, "registryCmd should have '%s' subcommand", expectedSub)
	}
}

// ========== Registry Enable Command Tests ==========

func TestRegistryEnableCmd_Exists(t *testing.T) {
	// Test that enable subcommand exists
	enableCmd := findSubcommand(registryCmd, "enable")
	assert.NotNil(t, enableCmd, "registryCmd should have 'enable' subcommand")
}

func TestRegistryEnableCmd_HasCorrectUse(t *testing.T) {
	// Test Use field for registry enable
	enableCmd := findSubcommand(registryCmd, "enable")
	assert.NotNil(t, enableCmd)
	assert.Equal(t, "enable <type>", enableCmd.Use)
}

func TestRegistryEnableCmd_IsSubcommandOfRegistry(t *testing.T) {
	// Test that enable is a child of registryCmd
	enableCmd := findSubcommand(registryCmd, "enable")
	assert.NotNil(t, enableCmd)
	assert.Equal(t, "registry", enableCmd.Parent().Name())
}

func TestRegistryEnableCmd_RequiresTypeArg(t *testing.T) {
	// Test that exactly 1 arg (type) is required
	enableCmd := findSubcommand(registryCmd, "enable")
	assert.NotNil(t, enableCmd)
	assert.NotNil(t, enableCmd.Args, "should have Args validator")

	// Test with 0 args (should fail)
	err := enableCmd.Args(enableCmd, []string{})
	assert.Error(t, err, "should return error with 0 args")

	// Test with 1 arg (should pass)
	err = enableCmd.Args(enableCmd, []string{"oci"})
	assert.NoError(t, err, "should accept 1 arg")

	// Test with 2 args (should fail)
	err = enableCmd.Args(enableCmd, []string{"oci", "extra"})
	assert.Error(t, err, "should return error with 2 args")
}

func TestRegistryEnableCmd_HasLifecycleFlag(t *testing.T) {
	// Test that enable command has --lifecycle flag
	enableCmd := findSubcommand(registryCmd, "enable")
	assert.NotNil(t, enableCmd)

	lifecycleFlag := enableCmd.Flags().Lookup("lifecycle")
	assert.NotNil(t, lifecycleFlag, "should have --lifecycle flag")
	assert.Equal(t, "string", lifecycleFlag.Value.Type(), "lifecycle flag should be string")
}

func TestRegistryEnableCmd_HasRunE(t *testing.T) {
	// Test that command has RunE (not Run)
	enableCmd := findSubcommand(registryCmd, "enable")
	assert.NotNil(t, enableCmd)
	assert.NotNil(t, enableCmd.RunE, "should have RunE function")
}

// ========== Registry Disable Command Tests ==========

func TestRegistryDisableCmd_Exists(t *testing.T) {
	// Test that disable subcommand exists
	disableCmd := findSubcommand(registryCmd, "disable")
	assert.NotNil(t, disableCmd, "registryCmd should have 'disable' subcommand")
}

func TestRegistryDisableCmd_HasCorrectUse(t *testing.T) {
	// Test Use field for registry disable
	disableCmd := findSubcommand(registryCmd, "disable")
	assert.NotNil(t, disableCmd)
	assert.Equal(t, "disable <type>", disableCmd.Use)
}

func TestRegistryDisableCmd_IsSubcommandOfRegistry(t *testing.T) {
	// Test that disable is a child of registryCmd
	disableCmd := findSubcommand(registryCmd, "disable")
	assert.NotNil(t, disableCmd)
	assert.Equal(t, "registry", disableCmd.Parent().Name())
}

func TestRegistryDisableCmd_RequiresTypeArg(t *testing.T) {
	// Test that exactly 1 arg (type) is required
	disableCmd := findSubcommand(registryCmd, "disable")
	assert.NotNil(t, disableCmd)
	assert.NotNil(t, disableCmd.Args, "should have Args validator")

	// Test with 0 args (should fail)
	err := disableCmd.Args(disableCmd, []string{})
	assert.Error(t, err, "should return error with 0 args")

	// Test with 1 arg (should pass)
	err = disableCmd.Args(disableCmd, []string{"npm"})
	assert.NoError(t, err, "should accept 1 arg")

	// Test with 2 args (should fail)
	err = disableCmd.Args(disableCmd, []string{"npm", "extra"})
	assert.Error(t, err, "should return error with 2 args")
}

func TestRegistryDisableCmd_HasRunE(t *testing.T) {
	// Test that command has RunE (not Run)
	disableCmd := findSubcommand(registryCmd, "disable")
	assert.NotNil(t, disableCmd)
	assert.NotNil(t, disableCmd.RunE, "should have RunE function")
}
