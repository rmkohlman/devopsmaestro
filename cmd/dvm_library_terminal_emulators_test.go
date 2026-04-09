// Package cmd - dvm_library_terminal_emulators_test.go
//
// TDD Phase 2 (RED) tests for Issue #29, Change #5:
//   "Add dvm library get terminal-emulators (+ describe/import)"
//
// These tests FAIL until the implementation adds terminal-emulators support
// to the dvm library commands.
//
// Run with: go test ./cmd/... -run "TestDvmLibraryTerminalEmulators" -v

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
// Change #5: dvm library get terminal-emulators
// =============================================================================

// TestDvmLibraryTerminalEmulators_GetList verifies that
// `dvm library get terminal-emulators` lists available emulators from the library.
func TestDvmLibraryTerminalEmulators_GetList(t *testing.T) {
	testRoot := newLibraryTerminalEmulatorsTestRoot(t)
	testRoot.SetArgs([]string{"library", "get", "terminal-emulators"})

	buf := new(bytes.Buffer)
	testRoot.SetOut(buf)
	testRoot.SetErr(buf)

	err := testRoot.Execute()
	assert.NoError(t, err, "dvm library get terminal-emulators should succeed")

	output := buf.String()
	assert.NotEmpty(t, output, "dvm library get terminal-emulators should produce output")
}

// TestDvmLibraryTerminalEmulators_GetListWithYaml verifies that
// `dvm library get terminal-emulators -o yaml` produces valid output.
func TestDvmLibraryTerminalEmulators_GetListWithYaml(t *testing.T) {
	testRoot := newLibraryTerminalEmulatorsTestRoot(t)
	testRoot.SetArgs([]string{"library", "get", "terminal-emulators", "-o", "yaml"})

	buf := new(bytes.Buffer)
	testRoot.SetOut(buf)
	testRoot.SetErr(buf)

	err := testRoot.Execute()
	assert.NoError(t, err, "dvm library get terminal-emulators -o yaml should succeed")
}

// TestDvmLibraryTerminalEmulators_NormalizeType verifies that
// normalizeResourceType maps "terminal-emulators" to the canonical form.
func TestDvmLibraryTerminalEmulators_NormalizeType(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "hyphenated terminal-emulators",
			args:     []string{"terminal-emulators"},
			expected: "terminal-emulators",
		},
		{
			name:     "spaced terminal emulators",
			args:     []string{"terminal", "emulators"},
			expected: "terminal-emulators",
		},
		{
			name:     "quoted spaced terminal emulators",
			args:     []string{"terminal emulators"},
			expected: "terminal-emulators",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeResourceType(tt.args)
			assert.Equal(t, tt.expected, result,
				"normalizeResourceType(%v) = %q, want %q", tt.args, result, tt.expected)
		})
	}
}

// TestDvmLibraryTerminalEmulators_DescribeEmulator verifies that
// `dvm library describe terminal-emulator <name>` shows emulator details.
func TestDvmLibraryTerminalEmulators_DescribeEmulator(t *testing.T) {
	testRoot := newLibraryTerminalEmulatorsTestRoot(t)
	testRoot.SetArgs([]string{"library", "describe", "terminal-emulator", "maestro"})

	buf := new(bytes.Buffer)
	testRoot.SetOut(buf)
	testRoot.SetErr(buf)

	err := testRoot.Execute()
	assert.NoError(t, err, "dvm library describe terminal-emulator maestro should succeed")
}

// TestDvmLibraryTerminalEmulators_DescribeNormalized verifies that
// normalizeResourceType maps "terminal-emulator" (singular) for describe.
func TestDvmLibraryTerminalEmulators_DescribeNormalized(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "singular terminal-emulator",
			args:     []string{"terminal-emulator"},
			expected: "terminal-emulator",
		},
		{
			name:     "spaced terminal emulator (singular)",
			args:     []string{"terminal", "emulator"},
			expected: "terminal-emulator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeResourceType(tt.args)
			assert.Equal(t, tt.expected, result,
				"normalizeResourceType(%v) = %q, want %q", tt.args, result, tt.expected)
		})
	}
}

// TestDvmLibraryTerminalEmulators_Import verifies that
// `dvm library import terminal-emulators` imports emulators into the database.
func TestDvmLibraryTerminalEmulators_Import(t *testing.T) {
	mockStore := db.NewMockDataStore()
	var ds db.DataStore = mockStore

	testRoot := newLibraryTerminalEmulatorsImportTestRoot(t, &ds)
	testRoot.SetArgs([]string{"library", "import", "terminal-emulators"})

	buf := new(bytes.Buffer)
	testRoot.SetOut(buf)
	testRoot.SetErr(buf)

	err := testRoot.Execute()
	assert.NoError(t, err, "dvm library import terminal-emulators should succeed")
}

// TestDvmLibraryTerminalEmulators_AllImportsIncludesEmulators verifies that
// `dvm library import --all` now also imports terminal-emulators.
func TestDvmLibraryTerminalEmulators_AllImportsIncludesEmulators(t *testing.T) {
	mockStore := db.NewMockDataStore()
	var ds db.DataStore = mockStore

	testRoot := newLibraryTerminalEmulatorsImportTestRoot(t, &ds)
	testRoot.SetArgs([]string{"library", "import", "--all"})

	buf := new(bytes.Buffer)
	testRoot.SetOut(buf)
	testRoot.SetErr(buf)

	err := testRoot.Execute()
	assert.NoError(t, err, "dvm library import --all should succeed (including terminal-emulators)")

	// After --all, the DB should have emulators too
	emulators, listErr := mockStore.ListTerminalEmulators()
	require.NoError(t, listErr)
	assert.Greater(t, len(emulators), 0,
		"dvm library import --all should import terminal-emulators into the database")
}

// TestDvmLibraryTerminalEmulators_RunLibraryListDispatch verifies that
// runLibraryList correctly dispatches "terminal-emulators" to the list function.
// This test exercises the switch statement in runLibraryList.
func TestDvmLibraryTerminalEmulators_RunLibraryListDispatch(t *testing.T) {
	// Build a test root command using the real runLibraryList handler
	rootCmd := &cobra.Command{
		Use:           "dvm",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	libCmd := &cobra.Command{
		Use:   "library",
		Short: "Browse libraries",
		Run:   func(cmd *cobra.Command, args []string) { cmd.Help() },
	}

	listCmd := &cobra.Command{
		Use:  "get [resource-type]",
		Args: cobra.MinimumNArgs(1),
		RunE: runLibraryList,
	}
	listCmd.Flags().StringP("output", "o", "table", "Output format")

	libCmd.AddCommand(listCmd)
	rootCmd.AddCommand(libCmd)
	rootCmd.SetArgs([]string{"library", "get", "terminal-emulators"})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	assert.NoError(t, err, "runLibraryList should handle 'terminal-emulators' type without error")
}

// =============================================================================
// Test Helpers
// =============================================================================

// newLibraryTerminalEmulatorsTestRoot creates a minimal isolated root command
// for testing dvm library get terminal-emulators (no DB needed for list/describe).
func newLibraryTerminalEmulatorsTestRoot(t *testing.T) *cobra.Command {
	t.Helper()

	root := &cobra.Command{
		Use:           "dvm",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	libCmd := &cobra.Command{
		Use:   "library",
		Short: "Browse libraries",
		Run:   func(cmd *cobra.Command, args []string) { cmd.Help() },
	}

	listCmd := &cobra.Command{
		Use:  "get [resource-type]",
		Args: cobra.MinimumNArgs(1),
		RunE: runLibraryList,
	}
	listCmd.Flags().StringP("output", "o", "table", "Output format (table|yaml|json)")

	showCmd := &cobra.Command{
		Use:  "describe [resource-type] [name]",
		Args: cobra.MinimumNArgs(2),
		RunE: runLibraryShow,
	}
	showCmd.Flags().StringP("output", "o", "yaml", "Output format (yaml|json)")

	libCmd.AddCommand(listCmd)
	libCmd.AddCommand(showCmd)
	root.AddCommand(libCmd)

	return root
}

// newLibraryTerminalEmulatorsImportTestRoot creates a minimal isolated root command
// for testing dvm library import terminal-emulators (DB needed for import).
func newLibraryTerminalEmulatorsImportTestRoot(t *testing.T, ds *db.DataStore) *cobra.Command {
	t.Helper()

	ctx := context.WithValue(context.Background(), CtxKeyDataStore, ds)

	root := &cobra.Command{
		Use:           "dvm",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	libCmd := &cobra.Command{
		Use:   "library",
		Short: "Browse libraries",
		Run:   func(cmd *cobra.Command, args []string) { cmd.Help() },
	}

	importCmd := &cobra.Command{
		Use:  "import [resource-type...]",
		RunE: runLibraryImport,
	}
	importCmd.Flags().Bool("all", false, "Import all resource types")
	importCmd.Flags().StringP("output", "o", "table", "Output format (table|yaml|json)")

	libCmd.AddCommand(importCmd)
	root.AddCommand(libCmd)
	root.SetContext(ctx)

	return root
}
