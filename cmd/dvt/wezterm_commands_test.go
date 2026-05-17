package main

// wezterm_commands_test.go — Tests for Issue #433
//
// Tests verify the new WezTerm commands: parse, diff, merge, and use with backup.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rmkohlman/MaestroTerminal/terminalops/wezterm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// dvt wezterm parse command tests
// =============================================================================

// TestWeztermParse_SubcommandExists verifies that `dvt wezterm parse` exists.
func TestWeztermParse_SubcommandExists(t *testing.T) {
	parseCmd := findDvtSubcommand(weztermCmd, "parse")
	assert.NotNil(t, parseCmd, "weztermCmd should have a 'parse' subcommand")
}

// TestWeztermParse_HasOutputFlag verifies the parse command has --output flag.
func TestWeztermParse_HasOutputFlag(t *testing.T) {
	parseCmd := findDvtSubcommand(weztermCmd, "parse")
	require.NotNil(t, parseCmd, "weztermCmd must have 'parse' subcommand")

	outputFlag := parseCmd.Flags().Lookup("output")
	assert.NotNil(t, outputFlag, "dvt wezterm parse should have --output flag")
}

// TestWeztermParse_DefaultOutputIsYAML verifies default output format is yaml.
func TestWeztermParse_DefaultOutputIsYAML(t *testing.T) {
	parseCmd := findDvtSubcommand(weztermCmd, "parse")
	require.NotNil(t, parseCmd, "weztermCmd must have 'parse' subcommand")

	outputFlag := parseCmd.Flags().Lookup("output")
	require.NotNil(t, outputFlag, "dvt wezterm parse must have --output flag")
	assert.Equal(t, "yaml", outputFlag.DefValue, "default output should be yaml")
}

// =============================================================================
// dvt wezterm diff command tests
// =============================================================================

// TestWeztermDiff_SubcommandExists verifies that `dvt wezterm diff` exists.
func TestWeztermDiff_SubcommandExists(t *testing.T) {
	diffCmd := findDvtSubcommand(weztermCmd, "diff")
	assert.NotNil(t, diffCmd, "weztermCmd should have a 'diff' subcommand")
}

// =============================================================================
// dvt wezterm merge command tests
// =============================================================================

// TestWeztermMerge_SubcommandExists verifies that `dvt wezterm merge` exists.
func TestWeztermMerge_SubcommandExists(t *testing.T) {
	mergeCmd := findDvtSubcommand(weztermCmd, "merge")
	assert.NotNil(t, mergeCmd, "weztermCmd should have a 'merge' subcommand")
}

// TestWeztermMerge_HasMergeFileFlag verifies merge has --merge-file flag.
func TestWeztermMerge_HasMergeFileFlag(t *testing.T) {
	mergeCmd := findDvtSubcommand(weztermCmd, "merge")
	require.NotNil(t, mergeCmd, "weztermCmd must have 'merge' subcommand")

	mergeFileFlag := mergeCmd.Flags().Lookup("merge-file")
	assert.NotNil(t, mergeFileFlag, "dvt wezterm merge should have --merge-file flag")
}

// TestWeztermMerge_HasDryRunFlag verifies merge has --dry-run flag.
func TestWeztermMerge_HasDryRunFlag(t *testing.T) {
	mergeCmd := findDvtSubcommand(weztermCmd, "merge")
	require.NotNil(t, mergeCmd, "weztermCmd must have 'merge' subcommand")

	dryRunFlag := mergeCmd.Flags().Lookup("dry-run")
	assert.NotNil(t, dryRunFlag, "dvt wezterm merge should have --dry-run flag")
}

// TestWeztermMerge_HasOutputFileFlag verifies merge has --output-file flag.
func TestWeztermMerge_HasOutputFileFlag(t *testing.T) {
	mergeCmd := findDvtSubcommand(weztermCmd, "merge")
	require.NotNil(t, mergeCmd, "weztermCmd must have 'merge' subcommand")

	outputFileFlag := mergeCmd.Flags().Lookup("output-file")
	assert.NotNil(t, outputFileFlag, "dvt wezterm merge should have --output-file flag")
}

// =============================================================================
// dvt wezterm use command tests (backup functionality)
// =============================================================================

// TestWeztermUse_HasNoBackupFlag verifies use has --no-backup flag.
func TestWeztermUse_HasNoBackupFlag(t *testing.T) {
	noBackupFlag := weztermUseCmd.Flags().Lookup("no-backup")
	assert.NotNil(t, noBackupFlag, "dvt wezterm use should have --no-backup flag")
}

// TestWeztermUse_HasDryRunFlag verifies use has --dry-run flag.
func TestWeztermUse_HasDryRunFlag(t *testing.T) {
	dryRunFlag := weztermUseCmd.Flags().Lookup("dry-run")
	assert.NotNil(t, dryRunFlag, "dvt wezterm use should have --dry-run flag")
}

// TestWeztermUse_HasOutputFileFlag verifies use has --output-file flag.
func TestWeztermUse_HasOutputFileFlag(t *testing.T) {
	outputFileFlag := weztermUseCmd.Flags().Lookup("output-file")
	assert.NotNil(t, outputFileFlag, "dvt wezterm use should have --output-file flag")
}

// =============================================================================
// Lua Parser tests (in MaestroTerminal)
// =============================================================================

// TestLuaParser_ParseSimpleConfig verifies the Lua parser can parse a simple config.
func TestLuaParser_ParseSimpleConfig(t *testing.T) {
	// This test verifies the parser exists and can be instantiated
	// The actual parsing is tested in MaestroTerminal package
	parser := wezterm.NewLuaParser()
	assert.NotNil(t, parser, "NewLuaParser should return a non-nil parser")
}

// TestBackupFile_CreatesBackup verifies backup file creation.
func TestBackupFile_CreatesBackup(t *testing.T) {
	// Create a temp file to backup
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "test.lua")
	err := os.WriteFile(srcPath, []byte("local test = true"), 0644)
	require.NoError(t, err, "should create temp file")

	// Test backup
	backupPath, err := wezterm.BackupFile(srcPath)
	require.NoError(t, err, "backup should succeed")
	assert.NotEmpty(t, backupPath, "backup path should not be empty")

	// Verify backup file exists
	_, err = os.Stat(backupPath)
	assert.NoError(t, err, "backup file should exist")
}

// TestBackupFile_NoFileReturnsNil verifies backup returns nil for non-existent file.
func TestBackupFile_NoFileReturnsNil(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "nonexistent.lua")

	backupPath, err := wezterm.BackupFile(srcPath)
	assert.NoError(t, err, "backup of non-existent file should not error")
	assert.Empty(t, backupPath, "backup path should be empty for non-existent file")
}