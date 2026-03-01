package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// TDD Phase 2 (RED): Lifecycle Command Tests
// =============================================================================
// These tests verify the NEW lifecycle command structure:
//   - dvm start registry <name>
//   - dvm stop registry <name>
//
// These tests will FAIL until implementation is added.
// =============================================================================

// ========== START Command Tests ==========

func TestStartCmd_Exists(t *testing.T) {
	// Test that startCmd parent command exists
	assert.NotNil(t, startCmd, "startCmd should exist")
}

func TestStartCmd_HasCorrectUse(t *testing.T) {
	// Test Use field
	assert.Equal(t, "start [resource]", startCmd.Use)
}

func TestStartCmd_RegisteredToRoot(t *testing.T) {
	// Test that startCmd is registered as child of rootCmd
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "start" {
			found = true
			break
		}
	}
	assert.True(t, found, "startCmd should be registered to rootCmd")
}

func TestStartRegistryCmd_Exists(t *testing.T) {
	// Test that startRegistryCmd subcommand exists
	assert.NotNil(t, startCmd, "startCmd should exist")
	found := false
	for _, sub := range startCmd.Commands() {
		if sub.Name() == "registry" {
			found = true
			break
		}
	}
	assert.True(t, found, "startCmd should have 'registry' subcommand")
}

func TestStartRegistryCmd_HasCorrectUse(t *testing.T) {
	// Test Use field for start registry
	var startRegistryCmd *cobra.Command
	for _, sub := range startCmd.Commands() {
		if sub.Name() == "registry" {
			startRegistryCmd = sub
			break
		}
	}
	assert.NotNil(t, startRegistryCmd)
	assert.Equal(t, "registry <name>", startRegistryCmd.Use)
}

func TestStartRegistryCmd_RequiresName(t *testing.T) {
	// Test that exactly 1 arg is required
	var startRegistryCmd *cobra.Command
	for _, sub := range startCmd.Commands() {
		if sub.Name() == "registry" {
			startRegistryCmd = sub
			break
		}
	}
	assert.NotNil(t, startRegistryCmd)
	assert.NotNil(t, startRegistryCmd.Args, "should have Args validator")

	// Test with 0 args (should fail)
	err := startRegistryCmd.Args(startRegistryCmd, []string{})
	assert.Error(t, err, "should return error with 0 args")

	// Test with 1 arg (should pass)
	err = startRegistryCmd.Args(startRegistryCmd, []string{"test-registry"})
	assert.NoError(t, err, "should accept 1 arg")

	// Test with 2 args (should fail)
	err = startRegistryCmd.Args(startRegistryCmd, []string{"reg1", "reg2"})
	assert.Error(t, err, "should return error with 2 args")
}

func TestStartRegistryCmd_HasRunE(t *testing.T) {
	// Test that command has RunE (not Run)
	var startRegistryCmd *cobra.Command
	for _, sub := range startCmd.Commands() {
		if sub.Name() == "registry" {
			startRegistryCmd = sub
			break
		}
	}
	assert.NotNil(t, startRegistryCmd)
	assert.NotNil(t, startRegistryCmd.RunE, "should have RunE function")
}

func TestStartRegistryCmd_HasShortDescription(t *testing.T) {
	// Test that command has Short description
	var startRegistryCmd *cobra.Command
	for _, sub := range startCmd.Commands() {
		if sub.Name() == "registry" {
			startRegistryCmd = sub
			break
		}
	}
	assert.NotNil(t, startRegistryCmd)
	assert.NotEmpty(t, startRegistryCmd.Short, "should have Short description")
}

// ========== STOP Command Tests ==========

func TestStopCmd_Exists(t *testing.T) {
	// Test that stopCmd parent command exists
	assert.NotNil(t, stopCmd, "stopCmd should exist")
}

func TestStopCmd_HasCorrectUse(t *testing.T) {
	// Test Use field
	assert.Equal(t, "stop [resource]", stopCmd.Use)
}

func TestStopCmd_RegisteredToRoot(t *testing.T) {
	// Test that stopCmd is registered as child of rootCmd
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "stop" {
			found = true
			break
		}
	}
	assert.True(t, found, "stopCmd should be registered to rootCmd")
}

func TestStopRegistryCmd_Exists(t *testing.T) {
	// Test that stopRegistryCmd subcommand exists
	assert.NotNil(t, stopCmd, "stopCmd should exist")
	found := false
	for _, sub := range stopCmd.Commands() {
		if sub.Name() == "registry" {
			found = true
			break
		}
	}
	assert.True(t, found, "stopCmd should have 'registry' subcommand")
}

func TestStopRegistryCmd_HasCorrectUse(t *testing.T) {
	// Test Use field for stop registry
	var stopRegistryCmd *cobra.Command
	for _, sub := range stopCmd.Commands() {
		if sub.Name() == "registry" {
			stopRegistryCmd = sub
			break
		}
	}
	assert.NotNil(t, stopRegistryCmd)
	assert.Equal(t, "registry <name>", stopRegistryCmd.Use)
}

func TestStopRegistryCmd_RequiresName(t *testing.T) {
	// Test that exactly 1 arg is required
	var stopRegistryCmd *cobra.Command
	for _, sub := range stopCmd.Commands() {
		if sub.Name() == "registry" {
			stopRegistryCmd = sub
			break
		}
	}
	assert.NotNil(t, stopRegistryCmd)
	assert.NotNil(t, stopRegistryCmd.Args, "should have Args validator")

	// Test with 0 args (should fail)
	err := stopRegistryCmd.Args(stopRegistryCmd, []string{})
	assert.Error(t, err, "should return error with 0 args")

	// Test with 1 arg (should pass)
	err = stopRegistryCmd.Args(stopRegistryCmd, []string{"test-registry"})
	assert.NoError(t, err, "should accept 1 arg")

	// Test with 2 args (should fail)
	err = stopRegistryCmd.Args(stopRegistryCmd, []string{"reg1", "reg2"})
	assert.Error(t, err, "should return error with 2 args")
}

func TestStopRegistryCmd_HasRunE(t *testing.T) {
	// Test that command has RunE (not Run)
	var stopRegistryCmd *cobra.Command
	for _, sub := range stopCmd.Commands() {
		if sub.Name() == "registry" {
			stopRegistryCmd = sub
			break
		}
	}
	assert.NotNil(t, stopRegistryCmd)
	assert.NotNil(t, stopRegistryCmd.RunE, "should have RunE function")
}

func TestStopRegistryCmd_HasShortDescription(t *testing.T) {
	// Test that command has Short description
	var stopRegistryCmd *cobra.Command
	for _, sub := range stopCmd.Commands() {
		if sub.Name() == "registry" {
			stopRegistryCmd = sub
			break
		}
	}
	assert.NotNil(t, stopRegistryCmd)
	assert.NotEmpty(t, stopRegistryCmd.Short, "should have Short description")
}

// ========== Helper Functions ==========

// Helper to find subcommand by name
func findSubcommand(parent *cobra.Command, name string) *cobra.Command {
	for _, sub := range parent.Commands() {
		if sub.Name() == name {
			return sub
		}
	}
	return nil
}
