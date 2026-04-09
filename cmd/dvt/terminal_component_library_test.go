package main

// terminal_component_library_test.go — TDD Phase 2 (RED) tests for Issue #29
//
// These tests verify the 7 required changes from the CLI Architect review.
// All tests in this file FAIL until the implementation is complete.
//
// Run with: go test ./cmd/dvt/... -run "TestDvtPackageLibrary|TestDvtEmulatorImport|TestDvtLibraryTag|TestDvtEmulatorCategory|TestDvtLibraryOutputDefaults" -v

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Change #1: dvt package library subcommand (get, describe, import)
// =============================================================================

// TestDvtPackageLibrary_SubcommandExists verifies that `dvt package library`
// is a registered subcommand of `dvt package`.
func TestDvtPackageLibrary_SubcommandExists(t *testing.T) {
	libraryCmd := findDvtSubcommand(packageCmd, "library")
	assert.NotNil(t, libraryCmd, "packageCmd should have a 'library' subcommand (dvt package library)")
}

// TestDvtPackageLibrary_GetSubcommand verifies that `dvt package library get`
// is registered and accessible.
func TestDvtPackageLibrary_GetSubcommand(t *testing.T) {
	libraryCmd := findDvtSubcommand(packageCmd, "library")
	require.NotNil(t, libraryCmd, "packageCmd must have 'library' subcommand")

	getCmd := findDvtSubcommand(libraryCmd, "get")
	assert.NotNil(t, getCmd, "package library should have a 'get' subcommand (dvt package library get)")
}

// TestDvtPackageLibrary_DescribeSubcommand verifies that `dvt package library describe`
// is registered and accessible.
func TestDvtPackageLibrary_DescribeSubcommand(t *testing.T) {
	libraryCmd := findDvtSubcommand(packageCmd, "library")
	require.NotNil(t, libraryCmd, "packageCmd must have 'library' subcommand")

	describeCmd := findDvtSubcommand(libraryCmd, "describe")
	assert.NotNil(t, describeCmd, "package library should have a 'describe' subcommand (dvt package library describe <name>)")
}

// TestDvtPackageLibrary_ImportSubcommand verifies that `dvt package library import`
// is registered and accessible.
func TestDvtPackageLibrary_ImportSubcommand(t *testing.T) {
	libraryCmd := findDvtSubcommand(packageCmd, "library")
	require.NotNil(t, libraryCmd, "packageCmd must have 'library' subcommand")

	importCmd := findDvtSubcommand(libraryCmd, "import")
	assert.NotNil(t, importCmd, "package library should have an 'import' subcommand (dvt package library import <name>)")
}

// TestDvtPackageLibrary_GetCommand_Executes verifies `dvt package library get`
// can be invoked without error (reads embedded YAML).
// This test requires the 'library' subcommand to exist on packageCmd.
func TestDvtPackageLibrary_GetCommand_Executes(t *testing.T) {
	libraryCmd := findDvtSubcommand(packageCmd, "library")
	require.NotNil(t, libraryCmd, "packageCmd must have 'library' subcommand (dvt package library)")

	rootCmd.SetArgs([]string{"package", "library", "get"})
	err := rootCmd.Execute()
	assert.NoError(t, err, "dvt package library get should succeed")
}

// TestDvtPackageLibrary_DescribeCommand_Executes verifies `dvt package library describe core`
// returns details for a known package.
// This test requires the 'library' subcommand to exist on packageCmd.
func TestDvtPackageLibrary_DescribeCommand_Executes(t *testing.T) {
	libraryCmd := findDvtSubcommand(packageCmd, "library")
	require.NotNil(t, libraryCmd, "packageCmd must have 'library' subcommand (dvt package library)")

	rootCmd.SetArgs([]string{"package", "library", "describe", "core"})
	err := rootCmd.Execute()
	assert.NoError(t, err, "dvt package library describe core should succeed")
}

// =============================================================================
// Change #2: dvt emulator install → dvt emulator library import
// =============================================================================

// TestDvtEmulatorImport_LibraryImportSubcommand verifies that
// `dvt emulator library import` exists as a subcommand.
func TestDvtEmulatorImport_LibraryImportSubcommand(t *testing.T) {
	libraryCmd := findDvtSubcommand(emulatorCmd, "library")
	require.NotNil(t, libraryCmd, "emulatorCmd must have 'library' subcommand")

	importCmd := findDvtSubcommand(libraryCmd, "import")
	assert.NotNil(t, importCmd, "emulator library should have an 'import' subcommand (dvt emulator library import <name>)")
}

// TestDvtEmulatorImport_HasForceFlag verifies `dvt emulator library import`
// has a --force flag.
func TestDvtEmulatorImport_HasForceFlag(t *testing.T) {
	libraryCmd := findDvtSubcommand(emulatorCmd, "library")
	require.NotNil(t, libraryCmd, "emulatorCmd must have 'library' subcommand")
	importCmd := findDvtSubcommand(libraryCmd, "import")
	require.NotNil(t, importCmd, "emulator library must have 'import' subcommand")

	forceFlag := importCmd.Flags().Lookup("force")
	assert.NotNil(t, forceFlag, "dvt emulator library import should have --force flag")
}

// TestDvtEmulatorImport_HasDryRunFlag verifies `dvt emulator library import`
// has a --dry-run flag.
func TestDvtEmulatorImport_HasDryRunFlag(t *testing.T) {
	libraryCmd := findDvtSubcommand(emulatorCmd, "library")
	require.NotNil(t, libraryCmd)
	importCmd := findDvtSubcommand(libraryCmd, "import")
	require.NotNil(t, importCmd)

	dryRunFlag := importCmd.Flags().Lookup("dry-run")
	assert.NotNil(t, dryRunFlag, "dvt emulator library import should have --dry-run flag")
}

// TestDvtEmulatorInstall_DeprecatedAlias verifies that the old `dvt emulator install`
// still works as a hidden deprecated alias.
func TestDvtEmulatorInstall_DeprecatedAlias(t *testing.T) {
	// The install command should still be present (hidden alias for backward compat)
	installCmd := findDvtSubcommand(emulatorCmd, "install")
	assert.NotNil(t, installCmd, "dvt emulator install should still exist as a deprecated hidden alias")
	if installCmd != nil {
		assert.True(t, installCmd.Hidden, "dvt emulator install should be hidden (deprecated alias)")
	}
}

// =============================================================================
// Change #3: dvt package install → dvt package library import
// =============================================================================

// TestDvtPackageInstall_DeprecatedAlias verifies that the old `dvt package install`
// still works as a hidden deprecated alias.
func TestDvtPackageInstall_DeprecatedAlias(t *testing.T) {
	// After migration, `dvt package install` should be a hidden deprecated alias
	installCmd := findDvtSubcommand(packageCmd, "install")
	assert.NotNil(t, installCmd, "dvt package install should still exist as a deprecated hidden alias")
	if installCmd != nil {
		assert.True(t, installCmd.Hidden, "dvt package install should be hidden (deprecated alias)")
	}
}

// =============================================================================
// Change #4: --tag flag on dvt prompt library get and dvt plugin library get
// =============================================================================

// TestDvtPromptLibraryGet_HasTagFlag verifies that `dvt prompt library get`
// accepts a --tag flag for filtering.
func TestDvtPromptLibraryGet_HasTagFlag(t *testing.T) {
	libraryCmd := findDvtSubcommand(promptCmd, "library")
	require.NotNil(t, libraryCmd, "promptCmd must have 'library' subcommand")
	getCmd := findDvtSubcommand(libraryCmd, "get")
	require.NotNil(t, getCmd, "prompt library must have 'get' subcommand")

	tagFlag := getCmd.Flags().Lookup("tag")
	assert.NotNil(t, tagFlag, "dvt prompt library get should have --tag flag for filtering")
}

// TestDvtPluginLibraryGet_HasTagFlag verifies that `dvt plugin library get`
// accepts a --tag flag for filtering.
func TestDvtPluginLibraryGet_HasTagFlag(t *testing.T) {
	libraryCmd := findDvtSubcommand(pluginCmd, "library")
	require.NotNil(t, libraryCmd, "pluginCmd must have 'library' subcommand")
	getCmd := findDvtSubcommand(libraryCmd, "get")
	require.NotNil(t, getCmd, "plugin library must have 'get' subcommand")

	tagFlag := getCmd.Flags().Lookup("tag")
	assert.NotNil(t, tagFlag, "dvt plugin library get should have --tag flag for filtering")
}

// TestDvtPromptLibraryGet_TagFlagFilters verifies that passing --tag to
// `dvt prompt library get` actually filters results.
func TestDvtPromptLibraryGet_TagFlagFilters(t *testing.T) {
	libraryCmd := findDvtSubcommand(promptCmd, "library")
	require.NotNil(t, libraryCmd)
	getCmd := findDvtSubcommand(libraryCmd, "get")
	require.NotNil(t, getCmd)

	// The --tag flag must exist before we can test filtering
	tagFlag := getCmd.Flags().Lookup("tag")
	require.NotNil(t, tagFlag, "dvt prompt library get must have --tag flag")

	rootCmd.SetArgs([]string{"prompt", "library", "get", "--tag", "starship"})
	err := rootCmd.Execute()
	assert.NoError(t, err, "dvt prompt library get --tag starship should succeed")
}

// TestDvtPluginLibraryGet_TagFlagFilters verifies that passing --tag to
// `dvt plugin library get` actually filters results.
func TestDvtPluginLibraryGet_TagFlagFilters(t *testing.T) {
	libraryCmd := findDvtSubcommand(pluginCmd, "library")
	require.NotNil(t, libraryCmd)
	getCmd := findDvtSubcommand(libraryCmd, "get")
	require.NotNil(t, getCmd)

	// The --tag flag must exist before we can test filtering
	tagFlag := getCmd.Flags().Lookup("tag")
	require.NotNil(t, tagFlag, "dvt plugin library get must have --tag flag")

	rootCmd.SetArgs([]string{"plugin", "library", "get", "--tag", "completion"})
	err := rootCmd.Execute()
	assert.NoError(t, err, "dvt plugin library get --tag completion should succeed")
}

// =============================================================================
// Change #6: --category filter on dvt emulator library get
// =============================================================================

// TestDvtEmulatorCategory_LibraryGetHasCategoryFlag verifies that
// `dvt emulator library get` has a --category flag.
func TestDvtEmulatorCategory_LibraryGetHasCategoryFlag(t *testing.T) {
	libraryCmd := findDvtSubcommand(emulatorCmd, "library")
	require.NotNil(t, libraryCmd, "emulatorCmd must have 'library' subcommand")
	getCmd := findDvtSubcommand(libraryCmd, "get")
	require.NotNil(t, getCmd, "emulator library must have 'get' subcommand")

	categoryFlag := getCmd.Flags().Lookup("category")
	assert.NotNil(t, categoryFlag, "dvt emulator library get should have --category flag")
}

// TestDvtEmulatorCategory_LibraryGetFilters verifies that passing --category to
// `dvt emulator library get` is accepted without error.
func TestDvtEmulatorCategory_LibraryGetFilters(t *testing.T) {
	libraryCmd := findDvtSubcommand(emulatorCmd, "library")
	require.NotNil(t, libraryCmd)
	getCmd := findDvtSubcommand(libraryCmd, "get")
	require.NotNil(t, getCmd)

	// The --category flag must exist before we can test filtering
	categoryFlag := getCmd.Flags().Lookup("category")
	require.NotNil(t, categoryFlag, "dvt emulator library get must have --category flag")

	rootCmd.SetArgs([]string{"emulator", "library", "get", "--category", "development"})
	err := rootCmd.Execute()
	assert.NoError(t, err, "dvt emulator library get --category development should succeed")
}

// =============================================================================
// Change #7: --output defaults: table for list commands, yaml for describe
// =============================================================================

// TestDvtLibraryOutputDefaults_PackageLibraryGet verifies that
// `dvt package library get` defaults --output to "table".
func TestDvtLibraryOutputDefaults_PackageLibraryGet(t *testing.T) {
	libraryCmd := findDvtSubcommand(packageCmd, "library")
	require.NotNil(t, libraryCmd)
	getCmd := findDvtSubcommand(libraryCmd, "get")
	require.NotNil(t, getCmd)

	outputFlag := getCmd.Flags().Lookup("output")
	require.NotNil(t, outputFlag, "dvt package library get should have --output flag")
	assert.Equal(t, "table", outputFlag.DefValue,
		"dvt package library get --output should default to 'table'")
}

// TestDvtLibraryOutputDefaults_PackageLibraryDescribe verifies that
// `dvt package library describe` defaults --output to "yaml".
func TestDvtLibraryOutputDefaults_PackageLibraryDescribe(t *testing.T) {
	libraryCmd := findDvtSubcommand(packageCmd, "library")
	require.NotNil(t, libraryCmd)
	describeCmd := findDvtSubcommand(libraryCmd, "describe")
	require.NotNil(t, describeCmd)

	outputFlag := describeCmd.Flags().Lookup("output")
	require.NotNil(t, outputFlag, "dvt package library describe should have --output flag")
	assert.Equal(t, "yaml", outputFlag.DefValue,
		"dvt package library describe --output should default to 'yaml'")
}

// TestDvtLibraryOutputDefaults_PromptLibraryGet verifies that
// `dvt prompt library get` defaults --output to "table".
func TestDvtLibraryOutputDefaults_PromptLibraryGet(t *testing.T) {
	libraryCmd := findDvtSubcommand(promptCmd, "library")
	require.NotNil(t, libraryCmd)
	getCmd := findDvtSubcommand(libraryCmd, "get")
	require.NotNil(t, getCmd)

	outputFlag := getCmd.Flags().Lookup("output")
	require.NotNil(t, outputFlag, "dvt prompt library get should have --output flag")
	assert.Equal(t, "table", outputFlag.DefValue,
		"dvt prompt library get --output should default to 'table'")
}

// TestDvtLibraryOutputDefaults_PromptLibraryDescribe verifies that
// `dvt prompt library describe` defaults --output to "yaml".
func TestDvtLibraryOutputDefaults_PromptLibraryDescribe(t *testing.T) {
	libraryCmd := findDvtSubcommand(promptCmd, "library")
	require.NotNil(t, libraryCmd)
	describeCmd := findDvtSubcommand(libraryCmd, "describe")
	require.NotNil(t, describeCmd)

	outputFlag := describeCmd.Flags().Lookup("output")
	require.NotNil(t, outputFlag, "dvt prompt library describe should have --output flag")
	assert.Equal(t, "yaml", outputFlag.DefValue,
		"dvt prompt library describe --output should default to 'yaml'")
}

// TestDvtLibraryOutputDefaults_PluginLibraryGet verifies that
// `dvt plugin library get` defaults --output to "table".
func TestDvtLibraryOutputDefaults_PluginLibraryGet(t *testing.T) {
	libraryCmd := findDvtSubcommand(pluginCmd, "library")
	require.NotNil(t, libraryCmd)
	getCmd := findDvtSubcommand(libraryCmd, "get")
	require.NotNil(t, getCmd)

	outputFlag := getCmd.Flags().Lookup("output")
	require.NotNil(t, outputFlag, "dvt plugin library get should have --output flag")
	assert.Equal(t, "table", outputFlag.DefValue,
		"dvt plugin library get --output should default to 'table'")
}

// TestDvtLibraryOutputDefaults_PluginLibraryDescribe verifies that
// `dvt plugin library describe` defaults --output to "yaml".
func TestDvtLibraryOutputDefaults_PluginLibraryDescribe(t *testing.T) {
	libraryCmd := findDvtSubcommand(pluginCmd, "library")
	require.NotNil(t, libraryCmd)
	describeCmd := findDvtSubcommand(libraryCmd, "describe")
	require.NotNil(t, describeCmd)

	outputFlag := describeCmd.Flags().Lookup("output")
	require.NotNil(t, outputFlag, "dvt plugin library describe should have --output flag")
	assert.Equal(t, "yaml", outputFlag.DefValue,
		"dvt plugin library describe --output should default to 'yaml'")
}

// TestDvtLibraryOutputDefaults_EmulatorLibraryGet verifies that
// `dvt emulator library get` defaults --output to "table".
func TestDvtLibraryOutputDefaults_EmulatorLibraryGet(t *testing.T) {
	libraryCmd := findDvtSubcommand(emulatorCmd, "library")
	require.NotNil(t, libraryCmd)
	getCmd := findDvtSubcommand(libraryCmd, "get")
	require.NotNil(t, getCmd)

	outputFlag := getCmd.Flags().Lookup("output")
	require.NotNil(t, outputFlag, "dvt emulator library get should have --output flag")
	assert.Equal(t, "table", outputFlag.DefValue,
		"dvt emulator library get --output should default to 'table'")
}

// TestDvtLibraryOutputDefaults_EmulatorLibraryDescribe verifies that
// `dvt emulator library describe` defaults --output to "yaml".
func TestDvtLibraryOutputDefaults_EmulatorLibraryDescribe(t *testing.T) {
	libraryCmd := findDvtSubcommand(emulatorCmd, "library")
	require.NotNil(t, libraryCmd)
	describeCmd := findDvtSubcommand(libraryCmd, "describe")
	require.NotNil(t, describeCmd)

	outputFlag := describeCmd.Flags().Lookup("output")
	require.NotNil(t, outputFlag, "dvt emulator library describe should have --output flag")
	assert.Equal(t, "yaml", outputFlag.DefValue,
		"dvt emulator library describe --output should default to 'yaml'")
}

// =============================================================================
// Test Helper
// =============================================================================

// findDvtSubcommand searches for a direct subcommand by name within a cobra command.
func findDvtSubcommand(parent *cobra.Command, name string) *cobra.Command {
	for _, sub := range parent.Commands() {
		if sub.Name() == name {
			return sub
		}
	}
	return nil
}
