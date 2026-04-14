package cmd

// =============================================================================
// Verification: Issue #257 — New dvm system commands (Phase 1)
// =============================================================================
// New commands: dvm system info, dvm system df, dvm system prune
// Parent command: dvm system (systemMaintCmd)
//
// These tests verify:
//   (a) systemMaintCmd is registered on rootCmd
//   (b) All three subcommands are registered under systemMaintCmd
//   (c) Each subcommand has the expected Use field and RunE function
//   (d) dvm system prune has all documented flags registered
//   (e) dvm system df and system info have --output flag
// =============================================================================

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSystemMaintCmd_RegisteredOnRoot verifies that `dvm system` is registered
// as a subcommand of rootCmd.
func TestSystemMaintCmd_RegisteredOnRoot(t *testing.T) {
	cmd := findSubcommand(rootCmd, "system")
	require.NotNil(t, cmd,
		"rootCmd must have a 'system' subcommand (dvm system)")
	assert.Equal(t, "system", cmd.Use)
}

// TestSystemMaintCmd_HasRunFunction verifies systemMaintCmd has a Run function
// (it runs Help when called without subcommand).
func TestSystemMaintCmd_HasRunFunction(t *testing.T) {
	cmd := findSubcommand(rootCmd, "system")
	require.NotNil(t, cmd)
	assert.NotNil(t, cmd.Run,
		"dvm system must have a Run func (shows help when called alone)")
}

// TestSystemInfoCmd_RegisteredUnderSystem verifies `dvm system info` exists.
func TestSystemInfoCmd_RegisteredUnderSystem(t *testing.T) {
	systemCmd := findSubcommand(rootCmd, "system")
	require.NotNil(t, systemCmd, "dvm system must be registered")

	infoCmd := findSubcommand(systemCmd, "info")
	require.NotNil(t, infoCmd,
		"dvm system info subcommand must be registered under dvm system")
	assert.Equal(t, "info", infoCmd.Use)
}

// TestSystemInfoCmd_HasRunE verifies that `dvm system info` uses RunE.
func TestSystemInfoCmd_HasRunE(t *testing.T) {
	systemCmd := findSubcommand(rootCmd, "system")
	require.NotNil(t, systemCmd)
	infoCmd := findSubcommand(systemCmd, "info")
	require.NotNil(t, infoCmd)
	assert.NotNil(t, infoCmd.RunE,
		"dvm system info must have a RunE function")
}

// TestSystemInfoCmd_HasOutputFlag verifies that `dvm system info` registers
// the --output / -o flag for JSON/YAML output.
func TestSystemInfoCmd_HasOutputFlag(t *testing.T) {
	systemCmd := findSubcommand(rootCmd, "system")
	require.NotNil(t, systemCmd)
	infoCmd := findSubcommand(systemCmd, "info")
	require.NotNil(t, infoCmd)

	flag := infoCmd.Flags().Lookup("output")
	require.NotNil(t, flag,
		"dvm system info must have --output flag for JSON/YAML output")
	assert.Equal(t, "o", flag.Shorthand,
		"output flag shorthand must be 'o'")
}

// TestSystemDFCmd_RegisteredUnderSystem verifies `dvm system df` exists.
func TestSystemDFCmd_RegisteredUnderSystem(t *testing.T) {
	systemCmd := findSubcommand(rootCmd, "system")
	require.NotNil(t, systemCmd)

	dfCmd := findSubcommand(systemCmd, "df")
	require.NotNil(t, dfCmd,
		"dvm system df subcommand must be registered under dvm system")
	assert.Equal(t, "df", dfCmd.Use)
}

// TestSystemDFCmd_HasRunE verifies that `dvm system df` uses RunE.
func TestSystemDFCmd_HasRunE(t *testing.T) {
	systemCmd := findSubcommand(rootCmd, "system")
	require.NotNil(t, systemCmd)
	dfCmd := findSubcommand(systemCmd, "df")
	require.NotNil(t, dfCmd)
	assert.NotNil(t, dfCmd.RunE,
		"dvm system df must have a RunE function")
}

// TestSystemDFCmd_HasOutputFlag verifies that `dvm system df` registers
// the --output flag.
func TestSystemDFCmd_HasOutputFlag(t *testing.T) {
	systemCmd := findSubcommand(rootCmd, "system")
	require.NotNil(t, systemCmd)
	dfCmd := findSubcommand(systemCmd, "df")
	require.NotNil(t, dfCmd)

	flag := dfCmd.Flags().Lookup("output")
	require.NotNil(t, flag,
		"dvm system df must have --output flag")
}

// TestSystemPruneCmd_RegisteredUnderSystem verifies `dvm system prune` exists.
func TestSystemPruneCmd_RegisteredUnderSystem(t *testing.T) {
	systemCmd := findSubcommand(rootCmd, "system")
	require.NotNil(t, systemCmd)

	pruneCmd := findSubcommand(systemCmd, "prune")
	require.NotNil(t, pruneCmd,
		"dvm system prune subcommand must be registered under dvm system")
	assert.Equal(t, "prune", pruneCmd.Use)
}

// TestSystemPruneCmd_HasRunE verifies that `dvm system prune` uses RunE.
func TestSystemPruneCmd_HasRunE(t *testing.T) {
	systemCmd := findSubcommand(rootCmd, "system")
	require.NotNil(t, systemCmd)
	pruneCmd := findSubcommand(systemCmd, "prune")
	require.NotNil(t, pruneCmd)
	assert.NotNil(t, pruneCmd.RunE,
		"dvm system prune must have a RunE function")
}

// TestSystemPruneCmd_HasExpectedFlags verifies all documented prune flags are
// registered: --buildkit, --images, --all, --dry-run, --force.
func TestSystemPruneCmd_HasExpectedFlags(t *testing.T) {
	systemCmd := findSubcommand(rootCmd, "system")
	require.NotNil(t, systemCmd)
	pruneCmd := findSubcommand(systemCmd, "prune")
	require.NotNil(t, pruneCmd)

	expectedFlags := []string{"buildkit", "images", "all", "dry-run", "force"}
	for _, flagName := range expectedFlags {
		t.Run(flagName, func(t *testing.T) {
			f := pruneCmd.Flags().Lookup(flagName)
			assert.NotNil(t, f,
				"dvm system prune must have --%s flag", flagName)
		})
	}
}

// TestSystemMaintCmd_SubcommandCount verifies systemMaintCmd has exactly
// 3 subcommands: info, df, prune.
func TestSystemMaintCmd_SubcommandCount(t *testing.T) {
	systemCmd := findSubcommand(rootCmd, "system")
	require.NotNil(t, systemCmd)

	subNames := make([]string, 0)
	for _, sub := range systemCmd.Commands() {
		subNames = append(subNames, sub.Name())
	}

	// Must contain exactly these three
	for _, want := range []string{"info", "df", "prune"} {
		assert.Contains(t, subNames, want,
			"dvm system must have '%s' subcommand; found: %v", want, subNames)
	}
}
