//go:build !integration
// +build !integration

// Package cmd - library_import_test.go contains tests for the
// `dvm library import` command.
//
// These tests verify the implementation of the `dvm library import` subcommand
// (Bug #5: "no dvm library import command" — now resolved).
//
// Run these tests with: go test ./cmd/... -run "TestLibraryImport" -v

package cmd

import (
	"bytes"
	"context"
	"testing"

	"devopsmaestro/db"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test 1: libraryCmd has an "import" subcommand
// =============================================================================

// TestLibraryImportCmd_Exists asserts that libraryCmd has a subcommand named "import".
func TestLibraryImportCmd_Exists(t *testing.T) {
	importCmd := findSubcommand(libraryCmd, "import")
	assert.NotNil(t, importCmd, "libraryCmd should have an 'import' subcommand (dvm library import)")
}

// =============================================================================
// Test 2: The import command has an --all flag
// =============================================================================

// TestLibraryImportCmd_AllFlag asserts the import command exposes an --all boolean flag.
func TestLibraryImportCmd_AllFlag(t *testing.T) {
	importCmd := findSubcommand(libraryCmd, "import")
	require.NotNil(t, importCmd, "libraryCmd must have an 'import' subcommand before checking flags")

	allFlag := importCmd.Flags().Lookup("all")
	assert.NotNil(t, allFlag, "library import should have an --all flag")

	if allFlag != nil {
		assert.Equal(t, "bool", allFlag.Value.Type(), "--all flag should be of type bool")
		assert.Equal(t, "false", allFlag.DefValue, "--all should default to false")
	}
}

// =============================================================================
// Test 3: dvm library import nvim-plugins writes to the DB
// =============================================================================

// TestLibraryImportCmd_SpecificType tests that `dvm library import nvim-plugins`
// calls UpsertPlugin for each plugin in the library.
func TestLibraryImportCmd_SpecificType(t *testing.T) {
	mockStore := db.NewMockDataStore()
	var ds db.DataStore = mockStore

	testRoot := newLibraryImportTestRoot(t, &ds)
	testRoot.SetArgs([]string{"library", "import", "nvim-plugins"})

	buf := new(bytes.Buffer)
	testRoot.SetOut(buf)
	testRoot.SetErr(buf)

	err := testRoot.Execute()
	assert.NoError(t, err, "dvm library import nvim-plugins should succeed")

	// After import, the DB should have plugins
	plugins, listErr := mockStore.ListPlugins()
	require.NoError(t, listErr)
	assert.Greater(t, len(plugins), 0, "nvim-plugins import should write plugins to the database")
}

// =============================================================================
// Test 4: --all imports all 6 resource types
// =============================================================================

// TestLibraryImportCmd_AllImportsAllTypes tests that `dvm library import --all`
// imports resources for all 6 types:
//   - nvim-plugins
//   - nvim-themes
//   - nvim-packages
//   - terminal-prompts
//   - terminal-plugins
//   - terminal-packages
func TestLibraryImportCmd_AllImportsAllTypes(t *testing.T) {
	mockStore := db.NewMockDataStore()
	var ds db.DataStore = mockStore

	testRoot := newLibraryImportTestRoot(t, &ds)
	testRoot.SetArgs([]string{"library", "import", "--all"})

	buf := new(bytes.Buffer)
	testRoot.SetOut(buf)
	testRoot.SetErr(buf)

	err := testRoot.Execute()
	assert.NoError(t, err, "dvm library import --all should succeed")

	// Verify all 6 resource types were imported (each should have >0 entries)
	plugins, err := mockStore.ListPlugins()
	require.NoError(t, err)
	assert.Greater(t, len(plugins), 0, "--all should import nvim-plugins")

	themes, err := mockStore.ListThemes()
	require.NoError(t, err)
	assert.Greater(t, len(themes), 0, "--all should import nvim-themes")

	packages, err := mockStore.ListPackages()
	require.NoError(t, err)
	assert.Greater(t, len(packages), 0, "--all should import nvim-packages")

	termPrompts, err := mockStore.ListTerminalPrompts()
	require.NoError(t, err)
	assert.Greater(t, len(termPrompts), 0, "--all should import terminal-prompts")

	termPlugins, err := mockStore.ListTerminalPlugins()
	require.NoError(t, err)
	assert.Greater(t, len(termPlugins), 0, "--all should import terminal-plugins")

	termPackages, err := mockStore.ListTerminalPackages()
	require.NoError(t, err)
	assert.Greater(t, len(termPackages), 0, "--all should import terminal-packages")
}

// =============================================================================
// Test 5: Unknown resource type returns a helpful error
// =============================================================================

// TestLibraryImportCmd_UnknownType tests that `dvm library import unknown-type`
// returns an error with a helpful message listing valid types.
func TestLibraryImportCmd_UnknownType(t *testing.T) {
	mockStore := db.NewMockDataStore()
	var ds db.DataStore = mockStore

	testRoot := newLibraryImportTestRoot(t, &ds)
	testRoot.SetArgs([]string{"library", "import", "unknown-type"})

	buf := new(bytes.Buffer)
	testRoot.SetOut(buf)
	testRoot.SetErr(buf)

	err := testRoot.Execute()
	assert.Error(t, err, "dvm library import unknown-type should return an error")

	// The error or output should contain helpful guidance
	combined := buf.String()
	if err != nil {
		combined += err.Error()
	}
	assert.Contains(t, combined, "unknown",
		"error should mention 'unknown' resource type")
}

// =============================================================================
// Test 6: No args and no --all flag shows an error or help
// =============================================================================

// TestLibraryImportCmd_NoArgs_NoAll_ShowsHelp tests that `dvm library import`
// (with no arguments and no --all) returns an error or shows usage help.
// The expected behavior is: require at least one resource type OR the --all flag.
func TestLibraryImportCmd_NoArgs_NoAll_ShowsHelp(t *testing.T) {
	mockStore := db.NewMockDataStore()
	var ds db.DataStore = mockStore

	testRoot := newLibraryImportTestRoot(t, &ds)
	testRoot.SetArgs([]string{"library", "import"})

	buf := new(bytes.Buffer)
	testRoot.SetOut(buf)
	testRoot.SetErr(buf)

	err := testRoot.Execute()

	// Should fail: no args, no --all — must require either a type or --all
	assert.Error(t, err,
		"dvm library import with no args and no --all should return an error")
}

// =============================================================================
// Test helpers
// =============================================================================

// newLibraryImportTestRoot builds a minimal isolated cobra root that:
//   - Has NO PersistentPreRunE (so it never touches the real SQLite DB)
//   - Injects the mock DataStore into the command context
//   - Mounts fresh command instances using the real runLibraryImport handler
//
// IMPORTANT: We create fresh cobra.Command instances instead of reusing the
// package-level libraryCmd/libraryImportCmd vars. Reusing shared package-level
// commands causes test pollution: Cobra sets a parent pointer on AddCommand,
// so the second test to call AddCommand(libraryCmd) gets a stale parent chain
// from the previous test, breaking context propagation.
func newLibraryImportTestRoot(t *testing.T, ds *db.DataStore) *cobra.Command {
	t.Helper()

	// Inject the mock DataStore into a context
	ctx := context.WithValue(context.Background(), CtxKeyDataStore, ds)

	root := &cobra.Command{
		Use:           "dvm",
		SilenceErrors: true,
		SilenceUsage:  true,
		// No PersistentPreRunE — we inject the dataStore via SetContext below.
	}

	// Fresh library parent command (mirrors real libraryCmd)
	libCmd := &cobra.Command{
		Use:     "library",
		Aliases: []string{"lib"},
		Short:   "Browse plugin and theme libraries",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Fresh import subcommand using the REAL runLibraryImport handler
	importCmd := &cobra.Command{
		Use:   "import [resource-type...]",
		Short: "Import library resources into the database",
		RunE:  runLibraryImport,
	}
	importCmd.Flags().Bool("all", false, "Import all resource types")
	importCmd.Flags().StringP("output", "o", "table", "Output format (table|yaml|json)")

	libCmd.AddCommand(importCmd)
	root.AddCommand(libCmd)

	// Inject the context so getDataStore(cmd) can resolve the mock store.
	root.SetContext(ctx)

	return root
}
