package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// TDD Phase 3 (GREEN): Rollout Command Tests
// =============================================================================
// These tests verify the rollout command structure:
//   - dvm rollout restart registry <name>
//   - dvm rollout status registry <name>
//   - dvm rollout history registry <name>
//   - dvm rollout undo registry <name>
// =============================================================================

// ========== ROLLOUT Parent Command Tests ==========

func TestRolloutCmd_Exists(t *testing.T) {
	// Test that rolloutCmd parent command exists
	assert.NotNil(t, rolloutCmd, "rolloutCmd should exist")
}

func TestRolloutCmd_HasCorrectUse(t *testing.T) {
	// Test Use field
	assert.Equal(t, "rollout", rolloutCmd.Use)
}

func TestRolloutCmd_RegisteredToRoot(t *testing.T) {
	// Test that rolloutCmd is registered as child of rootCmd
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "rollout" {
			found = true
			break
		}
	}
	assert.True(t, found, "rolloutCmd should be registered to rootCmd")
}

func TestRolloutCmd_HasSubcommands(t *testing.T) {
	// Test that rollout has restart, status, history, undo subcommands
	assert.NotNil(t, rolloutCmd, "rolloutCmd should exist")

	subcommands := []string{"restart", "status", "history", "undo"}
	for _, name := range subcommands {
		found := false
		for _, sub := range rolloutCmd.Commands() {
			if sub.Name() == name {
				found = true
				break
			}
		}
		assert.True(t, found, "rolloutCmd should have '%s' subcommand", name)
	}
}

// ========== ROLLOUT RESTART Tests ==========

func TestRolloutRestartCmd_Exists(t *testing.T) {
	// Test that rollout restart command exists
	assert.NotNil(t, rolloutCmd, "rolloutCmd should exist")
	found := false
	for _, sub := range rolloutCmd.Commands() {
		if sub.Name() == "restart" {
			found = true
			break
		}
	}
	assert.True(t, found, "rollout should have 'restart' subcommand")
}

func TestRolloutRestartCmd_HasCorrectUse(t *testing.T) {
	// Test Use field for rollout restart
	restartCmd := findSubcommand(rolloutCmd, "restart")
	assert.NotNil(t, restartCmd)
	assert.Equal(t, "restart", restartCmd.Use)
}

func TestRolloutRestartRegistryCmd_Exists(t *testing.T) {
	// Test that rollout restart registry subcommand exists
	restartCmd := findSubcommand(rolloutCmd, "restart")
	assert.NotNil(t, restartCmd)
	registryCmd := findSubcommand(restartCmd, "registry")
	assert.NotNil(t, registryCmd, "rollout restart should have 'registry' subcommand")
}

func TestRolloutRestartRegistryCmd_RequiresName(t *testing.T) {
	// Test that exactly 1 arg is required
	restartCmd := findSubcommand(rolloutCmd, "restart")
	assert.NotNil(t, restartCmd)
	registryCmd := findSubcommand(restartCmd, "registry")
	assert.NotNil(t, registryCmd)
	assert.NotNil(t, registryCmd.Args, "should have Args validator")

	// Test with 0 args (should fail)
	err := registryCmd.Args(registryCmd, []string{})
	assert.Error(t, err)

	// Test with 1 arg (should pass)
	err = registryCmd.Args(registryCmd, []string{"test-registry"})
	assert.NoError(t, err)
}

func TestRolloutRestartRegistryCmd_HasRunE(t *testing.T) {
	// Test that command has RunE function
	restartCmd := findSubcommand(rolloutCmd, "restart")
	assert.NotNil(t, restartCmd)
	registryCmd := findSubcommand(restartCmd, "registry")
	assert.NotNil(t, registryCmd)
	assert.NotNil(t, registryCmd.RunE, "should have RunE function")
}

// ========== ROLLOUT STATUS Tests ==========

func TestRolloutStatusCmd_Exists(t *testing.T) {
	// Test that rollout status command exists
	statusCmd := findSubcommand(rolloutCmd, "status")
	assert.NotNil(t, statusCmd, "rollout should have 'status' subcommand")
}

func TestRolloutStatusCmd_HasCorrectUse(t *testing.T) {
	// Test Use field for rollout status
	statusCmd := findSubcommand(rolloutCmd, "status")
	assert.NotNil(t, statusCmd)
	assert.Equal(t, "status", statusCmd.Use)
}

func TestRolloutStatusRegistryCmd_Exists(t *testing.T) {
	// Test that rollout status registry subcommand exists
	statusCmd := findSubcommand(rolloutCmd, "status")
	assert.NotNil(t, statusCmd)
	registryCmd := findSubcommand(statusCmd, "registry")
	assert.NotNil(t, registryCmd, "rollout status should have 'registry' subcommand")
}

func TestRolloutStatusRegistryCmd_RequiresName(t *testing.T) {
	// Test that exactly 1 arg is required
	statusCmd := findSubcommand(rolloutCmd, "status")
	assert.NotNil(t, statusCmd)
	registryCmd := findSubcommand(statusCmd, "registry")
	assert.NotNil(t, registryCmd)
	assert.NotNil(t, registryCmd.Args, "should have Args validator")
}

func TestRolloutStatusRegistryCmd_HasRunE(t *testing.T) {
	// Test that command has RunE function
	statusCmd := findSubcommand(rolloutCmd, "status")
	assert.NotNil(t, statusCmd)
	registryCmd := findSubcommand(statusCmd, "registry")
	assert.NotNil(t, registryCmd)
	assert.NotNil(t, registryCmd.RunE, "should have RunE function")
}

// ========== ROLLOUT HISTORY Tests ==========

func TestRolloutHistoryCmd_Exists(t *testing.T) {
	// Test that rollout history command exists
	historyCmd := findSubcommand(rolloutCmd, "history")
	assert.NotNil(t, historyCmd, "rollout should have 'history' subcommand")
}

func TestRolloutHistoryCmd_HasCorrectUse(t *testing.T) {
	// Test Use field for rollout history
	historyCmd := findSubcommand(rolloutCmd, "history")
	assert.NotNil(t, historyCmd)
	assert.Equal(t, "history", historyCmd.Use)
}

func TestRolloutHistoryRegistryCmd_Exists(t *testing.T) {
	// Test that rollout history registry subcommand exists
	historyCmd := findSubcommand(rolloutCmd, "history")
	assert.NotNil(t, historyCmd)
	registryCmd := findSubcommand(historyCmd, "registry")
	assert.NotNil(t, registryCmd, "rollout history should have 'registry' subcommand")
}

func TestRolloutHistoryRegistryCmd_RequiresName(t *testing.T) {
	// Test that exactly 1 arg is required
	historyCmd := findSubcommand(rolloutCmd, "history")
	assert.NotNil(t, historyCmd)
	registryCmd := findSubcommand(historyCmd, "registry")
	assert.NotNil(t, registryCmd)
	assert.NotNil(t, registryCmd.Args, "should have Args validator")
}

func TestRolloutHistoryRegistryCmd_HasRunE(t *testing.T) {
	// Test that command has RunE function
	historyCmd := findSubcommand(rolloutCmd, "history")
	assert.NotNil(t, historyCmd)
	registryCmd := findSubcommand(historyCmd, "registry")
	assert.NotNil(t, registryCmd)
	assert.NotNil(t, registryCmd.RunE, "should have RunE function")
}

// ========== ROLLOUT UNDO Tests ==========

func TestRolloutUndoCmd_Exists(t *testing.T) {
	// Test that rollout undo command exists
	undoCmd := findSubcommand(rolloutCmd, "undo")
	assert.NotNil(t, undoCmd, "rollout should have 'undo' subcommand")
}

func TestRolloutUndoCmd_HasCorrectUse(t *testing.T) {
	// Test Use field for rollout undo
	undoCmd := findSubcommand(rolloutCmd, "undo")
	assert.NotNil(t, undoCmd)
	assert.Equal(t, "undo", undoCmd.Use)
}

func TestRolloutUndoRegistryCmd_Exists(t *testing.T) {
	// Test that rollout undo registry subcommand exists
	undoCmd := findSubcommand(rolloutCmd, "undo")
	assert.NotNil(t, undoCmd)
	registryCmd := findSubcommand(undoCmd, "registry")
	assert.NotNil(t, registryCmd, "rollout undo should have 'registry' subcommand")
}

func TestRolloutUndoRegistryCmd_RequiresName(t *testing.T) {
	// Test that exactly 1 arg is required
	undoCmd := findSubcommand(rolloutCmd, "undo")
	assert.NotNil(t, undoCmd)
	registryCmd := findSubcommand(undoCmd, "registry")
	assert.NotNil(t, registryCmd)
	assert.NotNil(t, registryCmd.Args, "should have Args validator")
}

func TestRolloutUndoRegistryCmd_HasRunE(t *testing.T) {
	// Test that command has RunE function
	undoCmd := findSubcommand(rolloutCmd, "undo")
	assert.NotNil(t, undoCmd)
	registryCmd := findSubcommand(undoCmd, "registry")
	assert.NotNil(t, registryCmd)
	assert.NotNil(t, registryCmd.RunE, "should have RunE function")
}

// ========== Output Format Tests ==========

func TestRolloutStatusRegistry_SupportsOutputFormats(t *testing.T) {
	// Test that status supports --output flag
	statusCmd := findSubcommand(rolloutCmd, "status")
	assert.NotNil(t, statusCmd)
	registryCmd := findSubcommand(statusCmd, "registry")
	assert.NotNil(t, registryCmd)

	outputFlag := registryCmd.Flags().Lookup("output")
	assert.NotNil(t, outputFlag, "should have --output flag")
	assert.Equal(t, "string", outputFlag.Value.Type())
}

func TestRolloutHistoryRegistry_SupportsOutputFormats(t *testing.T) {
	// Test that history supports --output flag
	historyCmd := findSubcommand(rolloutCmd, "history")
	assert.NotNil(t, historyCmd)
	registryCmd := findSubcommand(historyCmd, "registry")
	assert.NotNil(t, registryCmd)

	outputFlag := registryCmd.Flags().Lookup("output")
	assert.NotNil(t, outputFlag, "should have --output flag")
	assert.Equal(t, "string", outputFlag.Value.Type())
}
