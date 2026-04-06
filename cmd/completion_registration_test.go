package cmd

// completion_registration_test.go — Tests for flag completion registration gaps
//
// Issue: #187
// These tests verify that hierarchy flags (--ecosystem, --domain, --app, --workspace)
// on 7 commands have completion functions registered via registerAllFlagCompletions().
//
// Currently these tests FAIL because registerAllFlagCompletions() does not include:
//   - setBuildArgCmd    (--ecosystem, --domain, --app, --workspace)
//   - deleteBuildArgCmd (--ecosystem, --domain, --app, --workspace)
//   - getBuildArgsCmd   (--ecosystem, --domain, --app, --workspace)
//   - setCACertCmd      (--ecosystem, --domain, --app, --workspace)
//   - deleteCACertCmd   (--ecosystem, --domain, --app, --workspace)
//   - getCACertsCmd     (--ecosystem, --domain, --app, --workspace)
//   - getAllCmd         (--ecosystem, --domain, --app)
//
// Fix: Add registerHierarchyFlagCompletions() calls for each command in registerAllFlagCompletions().

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helper: assertFlagHasCompletion checks that a named flag on a command has
// a completion function registered. It calls t.Errorf (not t.Fatalf) so all
// flags are checked in a single sub-test run.
// ---------------------------------------------------------------------------

func assertFlagHasCompletion(t *testing.T, cmdName, flagName string) {
	t.Helper()

	// Each command under test is a package-level variable — access directly.
	// registerAllFlagCompletions() runs from completion.go init() automatically.
	switch cmdName {
	case "setBuildArgCmd":
		fn, ok := setBuildArgCmd.GetFlagCompletionFunc(flagName)
		assert.True(t, ok,
			"setBuildArgCmd: flag --%s has no completion function registered (gap in registerAllFlagCompletions)", flagName)
		assert.NotNil(t, fn,
			"setBuildArgCmd: flag --%s completion function is nil", flagName)

	case "deleteBuildArgCmd":
		fn, ok := deleteBuildArgCmd.GetFlagCompletionFunc(flagName)
		assert.True(t, ok,
			"deleteBuildArgCmd: flag --%s has no completion function registered (gap in registerAllFlagCompletions)", flagName)
		assert.NotNil(t, fn,
			"deleteBuildArgCmd: flag --%s completion function is nil", flagName)

	case "getBuildArgsCmd":
		fn, ok := getBuildArgsCmd.GetFlagCompletionFunc(flagName)
		assert.True(t, ok,
			"getBuildArgsCmd: flag --%s has no completion function registered (gap in registerAllFlagCompletions)", flagName)
		assert.NotNil(t, fn,
			"getBuildArgsCmd: flag --%s completion function is nil", flagName)

	case "setCACertCmd":
		fn, ok := setCACertCmd.GetFlagCompletionFunc(flagName)
		assert.True(t, ok,
			"setCACertCmd: flag --%s has no completion function registered (gap in registerAllFlagCompletions)", flagName)
		assert.NotNil(t, fn,
			"setCACertCmd: flag --%s completion function is nil", flagName)

	case "deleteCACertCmd":
		fn, ok := deleteCACertCmd.GetFlagCompletionFunc(flagName)
		assert.True(t, ok,
			"deleteCACertCmd: flag --%s has no completion function registered (gap in registerAllFlagCompletions)", flagName)
		assert.NotNil(t, fn,
			"deleteCACertCmd: flag --%s completion function is nil", flagName)

	case "getCACertsCmd":
		fn, ok := getCACertsCmd.GetFlagCompletionFunc(flagName)
		assert.True(t, ok,
			"getCACertsCmd: flag --%s has no completion function registered (gap in registerAllFlagCompletions)", flagName)
		assert.NotNil(t, fn,
			"getCACertsCmd: flag --%s completion function is nil", flagName)

	case "getAllCmd":
		fn, ok := getAllCmd.GetFlagCompletionFunc(flagName)
		assert.True(t, ok,
			"getAllCmd: flag --%s has no completion function registered (gap in registerAllFlagCompletions)", flagName)
		assert.NotNil(t, fn,
			"getAllCmd: flag --%s completion function is nil", flagName)

	default:
		t.Errorf("assertFlagHasCompletion: unknown command %q", cmdName)
	}
}

// ---------------------------------------------------------------------------
// TestCompletionRegistered_BuildArgCommands
//
// Verifies that setBuildArgCmd, deleteBuildArgCmd, and getBuildArgsCmd all have
// completion functions registered for their hierarchy flags.
//
// These commands define --ecosystem, --domain, --app, --workspace flags but
// are NOT in registerAllFlagCompletions() — this test documents and catches that gap.
// ---------------------------------------------------------------------------

func TestCompletionRegistered_BuildArgCommands(t *testing.T) {
	// Sanity: ensure the commands exist (they're package-level vars, always non-nil)
	require.NotNil(t, setBuildArgCmd, "setBuildArgCmd must not be nil")
	require.NotNil(t, deleteBuildArgCmd, "deleteBuildArgCmd must not be nil")
	require.NotNil(t, getBuildArgsCmd, "getBuildArgsCmd must not be nil")

	// All three build-arg commands share the same set of hierarchy flags.
	buildArgCmds := []string{"setBuildArgCmd", "deleteBuildArgCmd", "getBuildArgsCmd"}
	hierarchyFlags := []string{"ecosystem", "domain", "app", "workspace"}

	for _, cmdName := range buildArgCmds {
		cmdName := cmdName // capture for subtests
		t.Run(cmdName, func(t *testing.T) {
			for _, flag := range hierarchyFlags {
				flag := flag // capture for subtests
				t.Run("--"+flag, func(t *testing.T) {
					assertFlagHasCompletion(t, cmdName, flag)
				})
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestCompletionRegistered_CACertCommands
//
// Verifies that setCACertCmd, deleteCACertCmd, and getCACertsCmd all have
// completion functions registered for their hierarchy flags.
//
// These commands define --ecosystem, --domain, --app, --workspace flags but
// are NOT in registerAllFlagCompletions() — this test documents and catches that gap.
// ---------------------------------------------------------------------------

func TestCompletionRegistered_CACertCommands(t *testing.T) {
	require.NotNil(t, setCACertCmd, "setCACertCmd must not be nil")
	require.NotNil(t, deleteCACertCmd, "deleteCACertCmd must not be nil")
	require.NotNil(t, getCACertsCmd, "getCACertsCmd must not be nil")

	caCertCmds := []string{"setCACertCmd", "deleteCACertCmd", "getCACertsCmd"}
	hierarchyFlags := []string{"ecosystem", "domain", "app", "workspace"}

	for _, cmdName := range caCertCmds {
		cmdName := cmdName
		t.Run(cmdName, func(t *testing.T) {
			for _, flag := range hierarchyFlags {
				flag := flag
				t.Run("--"+flag, func(t *testing.T) {
					assertFlagHasCompletion(t, cmdName, flag)
				})
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestCompletionRegistered_GetAllCommand
//
// Verifies that getAllCmd has completion functions registered for --ecosystem,
// --domain, and --app.
//
// Note: getAllCmd does NOT have a --workspace flag (by design), so only 3 flags
// are tested here.
//
// getAllCmd is NOT in registerAllFlagCompletions() — this test documents and
// catches that gap.
// ---------------------------------------------------------------------------

func TestCompletionRegistered_GetAllCommand(t *testing.T) {
	require.NotNil(t, getAllCmd, "getAllCmd must not be nil")

	// Confirm the --workspace flag does not exist on getAllCmd (design constraint).
	// If it does exist, the test spec needs updating.
	t.Run("no --workspace flag (design constraint)", func(t *testing.T) {
		wsFlag := getAllCmd.Flags().Lookup("workspace")
		assert.Nil(t, wsFlag,
			"getAllCmd should not have a --workspace flag; if it was added, update this test")
	})

	// Flags that getAllCmd DOES have and that need completion registered.
	scopeFlags := []string{"ecosystem", "domain", "app"}

	for _, flag := range scopeFlags {
		flag := flag
		t.Run("--"+flag, func(t *testing.T) {
			assertFlagHasCompletion(t, "getAllCmd", flag)
		})
	}
}

// ---------------------------------------------------------------------------
// TestCompletionRegistered_FlagsExistOnCommands
//
// Sanity-check: verify that each command actually has the expected flags defined.
// If a flag doesn't exist, the completion test above would be testing the wrong thing.
// These should always PASS — they guard against refactoring removing flags.
// ---------------------------------------------------------------------------

func TestCompletionRegistered_FlagsExistOnCommands(t *testing.T) {
	type flagCheck struct {
		cmdName string
		cmd     func() interface {
			Flags() interface{ Lookup(string) interface{} }
		}
		flagName string
	}

	// Use direct checks since we can't easily make a generic helper without
	// importing pflag — just check each command's flags directly.

	t.Run("setBuildArgCmd flags exist", func(t *testing.T) {
		require.NotNil(t, setBuildArgCmd)
		for _, f := range []string{"ecosystem", "domain", "app", "workspace"} {
			assert.NotNil(t, setBuildArgCmd.Flags().Lookup(f),
				"setBuildArgCmd should have --%s flag", f)
		}
	})

	t.Run("deleteBuildArgCmd flags exist", func(t *testing.T) {
		require.NotNil(t, deleteBuildArgCmd)
		for _, f := range []string{"ecosystem", "domain", "app", "workspace"} {
			assert.NotNil(t, deleteBuildArgCmd.Flags().Lookup(f),
				"deleteBuildArgCmd should have --%s flag", f)
		}
	})

	t.Run("getBuildArgsCmd flags exist", func(t *testing.T) {
		require.NotNil(t, getBuildArgsCmd)
		for _, f := range []string{"ecosystem", "domain", "app", "workspace"} {
			assert.NotNil(t, getBuildArgsCmd.Flags().Lookup(f),
				"getBuildArgsCmd should have --%s flag", f)
		}
	})

	t.Run("setCACertCmd flags exist", func(t *testing.T) {
		require.NotNil(t, setCACertCmd)
		for _, f := range []string{"ecosystem", "domain", "app", "workspace"} {
			assert.NotNil(t, setCACertCmd.Flags().Lookup(f),
				"setCACertCmd should have --%s flag", f)
		}
	})

	t.Run("deleteCACertCmd flags exist", func(t *testing.T) {
		require.NotNil(t, deleteCACertCmd)
		for _, f := range []string{"ecosystem", "domain", "app", "workspace"} {
			assert.NotNil(t, deleteCACertCmd.Flags().Lookup(f),
				"deleteCACertCmd should have --%s flag", f)
		}
	})

	t.Run("getCACertsCmd flags exist", func(t *testing.T) {
		require.NotNil(t, getCACertsCmd)
		for _, f := range []string{"ecosystem", "domain", "app", "workspace"} {
			assert.NotNil(t, getCACertsCmd.Flags().Lookup(f),
				"getCACertsCmd should have --%s flag", f)
		}
	})

	t.Run("getAllCmd flags exist", func(t *testing.T) {
		require.NotNil(t, getAllCmd)
		for _, f := range []string{"ecosystem", "domain", "app"} {
			assert.NotNil(t, getAllCmd.Flags().Lookup(f),
				"getAllCmd should have --%s flag", f)
		}
	})
}
