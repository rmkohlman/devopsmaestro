package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// TDD Phase 2 (RED): Registry Defaults Command Tests
// =============================================================================
// These tests verify the NEW registry defaults command structure:
//   - dvm registry set-default <type> <registry-name>
//   - dvm registry get-defaults
//
// These commands manage default registry assignments for each type.
//
// These tests will FAIL until implementation is added.
// =============================================================================

// ========== Registry Set-Default Command Tests ==========

func TestRegistrySetDefaultCmd_Exists(t *testing.T) {
	// Test that set-default subcommand exists
	setDefaultCmd := findSubcommand(registryCmd, "set-default")
	assert.NotNil(t, setDefaultCmd, "registryCmd should have 'set-default' subcommand")
}

func TestRegistrySetDefaultCmd_HasCorrectUse(t *testing.T) {
	// Test Use field for registry set-default
	setDefaultCmd := findSubcommand(registryCmd, "set-default")
	assert.NotNil(t, setDefaultCmd)
	assert.Equal(t, "set-default <type> <registry-name>", setDefaultCmd.Use)
}

func TestRegistrySetDefaultCmd_IsSubcommandOfRegistry(t *testing.T) {
	// Test that set-default is a child of registryCmd
	setDefaultCmd := findSubcommand(registryCmd, "set-default")
	assert.NotNil(t, setDefaultCmd)
	assert.Equal(t, "registry", setDefaultCmd.Parent().Name())
}

func TestRegistrySetDefaultCmd_RequiresTwoArgs(t *testing.T) {
	// Test that exactly 2 args (type and registry-name) are required
	setDefaultCmd := findSubcommand(registryCmd, "set-default")
	assert.NotNil(t, setDefaultCmd)
	assert.NotNil(t, setDefaultCmd.Args, "should have Args validator")

	// Test with 0 args (should fail)
	err := setDefaultCmd.Args(setDefaultCmd, []string{})
	assert.Error(t, err, "should return error with 0 args")

	// Test with 1 arg (should fail)
	err = setDefaultCmd.Args(setDefaultCmd, []string{"oci"})
	assert.Error(t, err, "should return error with 1 arg")

	// Test with 2 args (should pass)
	err = setDefaultCmd.Args(setDefaultCmd, []string{"oci", "zot-local"})
	assert.NoError(t, err, "should accept 2 args")

	// Test with 3 args (should fail)
	err = setDefaultCmd.Args(setDefaultCmd, []string{"oci", "zot-local", "extra"})
	assert.Error(t, err, "should return error with 3 args")
}

func TestRegistrySetDefaultCmd_HasRunE(t *testing.T) {
	// Test that command has RunE (not Run)
	setDefaultCmd := findSubcommand(registryCmd, "set-default")
	assert.NotNil(t, setDefaultCmd)
	assert.NotNil(t, setDefaultCmd.RunE, "should have RunE function")
}

// ========== Registry Get-Defaults Command Tests ==========

func TestRegistryGetDefaultsCmd_Exists(t *testing.T) {
	// Test that get-defaults subcommand exists
	getDefaultsCmd := findSubcommand(registryCmd, "get-defaults")
	assert.NotNil(t, getDefaultsCmd, "registryCmd should have 'get-defaults' subcommand")
}

func TestRegistryGetDefaultsCmd_HasCorrectUse(t *testing.T) {
	// Test Use field for registry get-defaults
	getDefaultsCmd := findSubcommand(registryCmd, "get-defaults")
	assert.NotNil(t, getDefaultsCmd)
	assert.Equal(t, "get-defaults", getDefaultsCmd.Use)
}

func TestRegistryGetDefaultsCmd_IsSubcommandOfRegistry(t *testing.T) {
	// Test that get-defaults is a child of registryCmd
	getDefaultsCmd := findSubcommand(registryCmd, "get-defaults")
	assert.NotNil(t, getDefaultsCmd)
	assert.Equal(t, "registry", getDefaultsCmd.Parent().Name())
}

func TestRegistryGetDefaultsCmd_NoArgsRequired(t *testing.T) {
	// Test that 0 args is valid (no Args validator, or allows 0)
	getDefaultsCmd := findSubcommand(registryCmd, "get-defaults")
	assert.NotNil(t, getDefaultsCmd)

	// If Args is set, it should allow 0 args
	if getDefaultsCmd.Args != nil {
		err := getDefaultsCmd.Args(getDefaultsCmd, []string{})
		assert.NoError(t, err, "should accept 0 args")
	}
}

func TestRegistryGetDefaultsCmd_HasRunE(t *testing.T) {
	// Test that command has RunE (not Run)
	getDefaultsCmd := findSubcommand(registryCmd, "get-defaults")
	assert.NotNil(t, getDefaultsCmd)
	assert.NotNil(t, getDefaultsCmd.RunE, "should have RunE function")
}
