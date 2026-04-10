package cmd

// =============================================================================
// TDD Phase 2 (RED): Parallel Build Execution — Issue #213
// =============================================================================
// These tests define the contract for parallel build execution as specified in
// the CLI architect review comment on issue #213.
//
// Key behaviors under test:
//   - --detach flag runs build in background and returns session ID
//   - --concurrency <n> controls worker pool size (default 4)
//   - --all / -A builds all workspaces in parallel
//   - Scope flags (-e, -d, -a) are NOT mutually exclusive with --all
//   - --dry-run previews workspaces without building
//   - Exit code non-zero when ANY workspace fails
//   - Helpful error when no workspaces match scope
//
// RED state: Tests FAIL because --detach, --concurrency, and parallel
//            execution logic do not exist yet.
// GREEN state: After dvm-core implements Issue #213.
// =============================================================================

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Flag Registration Tests
// =============================================================================

// TestBuildCommandHasDetachFlag verifies that --detach is registered on buildCmd.
// --detach runs the build in background; use 'dvm build status' to monitor.
//
// RED: --detach flag does not exist yet in build.go init().
func TestBuildCommandHasDetachFlag(t *testing.T) {
	flag := buildCmd.Flags().Lookup("detach")
	require.NotNil(t, flag,
		"buildCmd must have --detach flag (not yet registered in build.go)")
	assert.Equal(t, "bool", flag.Value.Type(),
		"--detach must be a boolean flag")
	assert.Equal(t, "false", flag.DefValue,
		"--detach default must be false (foreground is the default)")
}

// TestBuildCommandHasConcurrencyFlag verifies that --concurrency is registered
// on buildCmd with a default of 4.
//
// RED: --concurrency flag does not exist yet in build.go init().
func TestBuildCommandHasConcurrencyFlag(t *testing.T) {
	flag := buildCmd.Flags().Lookup("concurrency")
	require.NotNil(t, flag,
		"buildCmd must have --concurrency flag (not yet registered in build.go)")
	assert.Equal(t, "int", flag.Value.Type(),
		"--concurrency must be an integer flag")
	assert.Equal(t, "4", flag.DefValue,
		"--concurrency default must be 4")
}

// TestBuildCommandHasAllFlag verifies that --all / -A is registered on buildCmd.
// This is a GREEN guard — AddAllFlag is already called in build.go.
func TestBuildCommandHasAllFlag(t *testing.T) {
	flag := buildCmd.Flags().Lookup("all")
	require.NotNil(t, flag,
		"buildCmd must have --all flag")
	assert.Equal(t, "A", flag.Shorthand,
		"--all shorthand must be -A (kubectl convention)")
	assert.Equal(t, "bool", flag.Value.Type())
	assert.Equal(t, "false", flag.DefValue)
}

// TestBuildCommandNoBackgroundFlag verifies that --background does NOT exist
// on buildCmd. Per CLI architect review, --background is rejected in favor of
// --detach.
//
// GREEN: --background was never added; this is a guard against regression.
func TestBuildCommandNoBackgroundFlag(t *testing.T) {
	flag := buildCmd.Flags().Lookup("background")
	assert.Nil(t, flag,
		"buildCmd must NOT have --background flag (rejected by CLI architect — use --detach)")
}

// TestBuildCommandNoForegroundFlag verifies that --foreground does NOT exist
// on buildCmd. Foreground is the default; there is no flag to select it.
//
// GREEN: --foreground was never added; this is a guard against regression.
func TestBuildCommandNoForegroundFlag(t *testing.T) {
	flag := buildCmd.Flags().Lookup("foreground")
	assert.Nil(t, flag,
		"buildCmd must NOT have --foreground flag (foreground is the default, not a flag)")
}

// =============================================================================
// Mutual Exclusion Removal Tests
// =============================================================================

// TestBuildParallelNoMutualExclusion verifies that --all combined with scope
// flags does NOT produce an error. Per CLI architect review, validateBuildAllMutualExclusion
// must be removed and scope flags + --all are additive, not exclusive.
//
// RED: Currently buildCmd.RunE calls validateBuildAllMutualExclusion which
//
//	returns an error when --all is combined with any scope flag.
func TestBuildParallelNoMutualExclusion(t *testing.T) {
	tests := []struct {
		name      string
		flagName  string
		flagValue string
	}{
		{"all + ecosystem", "ecosystem", "healthcare"},
		{"all + domain", "domain", "payments"},
		{"all + app", "app", "portal"},
		{"all + workspace", "workspace", "dev"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set --all and a scope flag
			require.NoError(t, buildCmd.Flags().Set("all", "true"))
			require.NoError(t, buildCmd.Flags().Set(tt.flagName, tt.flagValue))

			t.Cleanup(func() {
				_ = buildCmd.Flags().Set("all", "false")
				_ = buildCmd.Flags().Set(tt.flagName, "")
				buildFlags = HierarchyFlags{}
			})

			// validateBuildAllMutualExclusion must NOT fire — the combination
			// is valid. We test the validation function directly.
			// After #213 is implemented, this function should return nil always.
			allSet, _ := buildCmd.Flags().GetBool("all")
			err := validateBuildAllMutualExclusion(allSet, buildFlags)
			assert.NoError(t, err,
				"--all + --%s must NOT be an error after #213 removes mutual exclusion",
				tt.flagName)
		})
	}
}

// =============================================================================
// Concurrency Behavior Tests
// =============================================================================

// TestBuildParallelDefaultConcurrency verifies that the default worker pool
// size is 4 when --concurrency is not specified.
//
// RED: buildConcurrency variable and --concurrency flag don't exist yet.
func TestBuildParallelDefaultConcurrency(t *testing.T) {
	flag := buildCmd.Flags().Lookup("concurrency")
	require.NotNil(t, flag, "buildCmd must have --concurrency flag")
	assert.Equal(t, "4", flag.DefValue,
		"default concurrency must be 4 workers")
}

// TestBuildParallelConcurrencyFlagValue verifies that the --concurrency flag
// can be set to a custom value and is read back correctly.
//
// RED: --concurrency flag does not exist yet.
func TestBuildParallelConcurrencyFlagValue(t *testing.T) {
	flag := buildCmd.Flags().Lookup("concurrency")
	require.NotNil(t, flag, "buildCmd must have --concurrency flag")

	require.NoError(t, buildCmd.Flags().Set("concurrency", "8"))
	t.Cleanup(func() {
		_ = buildCmd.Flags().Set("concurrency", "4")
	})

	val, err := buildCmd.Flags().GetInt("concurrency")
	require.NoError(t, err)
	assert.Equal(t, 8, val, "--concurrency 8 should be readable as int 8")
}

// =============================================================================
// Scope Resolution Format Tests
// =============================================================================

// TestBuildParallelNoMatchError verifies that FormatBuildNoMatchScopeError
// returns a helpful message when --ecosystem does not match any workspaces.
//
// RED: FormatBuildNoMatchScopeError does not exist yet.
func TestBuildParallelNoMatchError(t *testing.T) {
	tests := []struct {
		name       string
		scopeFlag  string
		scopeValue string
		wantParts  []string
	}{
		{
			name:       "nonexistent ecosystem",
			scopeFlag:  "ecosystem",
			scopeValue: "does-not-exist",
			wantParts: []string{
				"does-not-exist",
				"ecosystem",
				"dvm get workspaces",
			},
		},
		{
			name:       "nonexistent domain",
			scopeFlag:  "domain",
			scopeValue: "ghost-domain",
			wantParts: []string{
				"ghost-domain",
				"domain",
			},
		},
		{
			name:       "nonexistent app",
			scopeFlag:  "app",
			scopeValue: "missing-app",
			wantParts: []string{
				"missing-app",
				"app",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := FormatBuildNoMatchScopeError(tt.scopeFlag, tt.scopeValue)
			for _, part := range tt.wantParts {
				assert.Contains(t, msg, part,
					"no-match error for %s=%q should contain %q",
					tt.scopeFlag, tt.scopeValue, part)
			}
		})
	}
}

// TestBuildParallelNoMatchError_HintMentionsGetWorkspaces verifies the hint
// in the no-match error directs the user to 'dvm get workspaces -A'.
//
// RED: FormatBuildNoMatchScopeError does not exist yet.
func TestBuildParallelNoMatchError_HintMentionsGetWorkspaces(t *testing.T) {
	msg := FormatBuildNoMatchScopeError("ecosystem", "does-not-exist")
	assert.Contains(t, msg, "dvm get workspaces",
		"no-match error hint should suggest 'dvm get workspaces -A'")
}

// =============================================================================
// Dry-run Preview Tests
// =============================================================================

// TestBuildParallelDryRun_FlagRegistered verifies --dry-run is on buildCmd.
// GREEN guard — AddDryRunFlag is already called in build.go init().
func TestBuildParallelDryRun_FlagRegistered(t *testing.T) {
	flag := buildCmd.Flags().Lookup("dry-run")
	require.NotNil(t, flag, "buildCmd must have --dry-run flag")
	assert.Equal(t, "false", flag.DefValue)
}

// TestBuildParallelDryRun_TableFormatsWorkspaceAndApp verifies the dry-run
// table formatter includes workspace and app columns.
// GREEN — FormatBuildDryRunTable already exists in build_hierarchy.go.
func TestBuildParallelDryRun_TableFormatsWorkspaceAndApp(t *testing.T) {
	row := FormatBuildDryRunTable("api-dev", "portal")
	assert.Contains(t, row, "api-dev",
		"dry-run table row must include workspace name")
	assert.Contains(t, row, "portal",
		"dry-run table row must include app name")
	assert.Contains(t, row, "(dry-run)",
		"dry-run table row must include dry-run status marker")
}

// TestBuildParallelDryRun_SummaryCount verifies the summary line for multi-workspace
// dry-run shows the workspace count.
// GREEN — FormatBuildDryRunSummary already exists in build_hierarchy.go.
func TestBuildParallelDryRun_SummaryCount(t *testing.T) {
	tests := []struct {
		count int
		want  string
	}{
		{6, "6 workspace"},
		{1, "1 workspace"},
		{0, "0 workspace"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			msg := FormatBuildDryRunSummary(tt.count)
			assert.Contains(t, msg, tt.want)
			assert.Contains(t, msg, "--dry-run")
		})
	}
}

// =============================================================================
// Parallel Build Output Format Tests
// =============================================================================

// TestBuildParallelProgressHeader verifies FormatParallelBuildHeader produces
// the streaming table header shown during parallel builds.
//
// Expected format:
//
//	Building 6 workspace(s) in ecosystem "healthcare" (concurrency: 4)
//
// RED: FormatParallelBuildHeader does not exist yet.
func TestBuildParallelProgressHeader(t *testing.T) {
	tests := []struct {
		name        string
		count       int
		scopeLabel  string
		scopeValue  string
		concurrency int
		wantParts   []string
	}{
		{
			name:        "ecosystem scope",
			count:       6,
			scopeLabel:  "ecosystem",
			scopeValue:  "healthcare",
			concurrency: 4,
			wantParts:   []string{"6", "healthcare", "4"},
		},
		{
			name:        "all workspaces",
			count:       12,
			scopeLabel:  "",
			scopeValue:  "",
			concurrency: 8,
			wantParts:   []string{"12", "8"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := FormatParallelBuildHeader(tt.count, tt.scopeLabel, tt.scopeValue, tt.concurrency)
			for _, part := range tt.wantParts {
				assert.Contains(t, header, part,
					"parallel build header should contain %q", part)
			}
		})
	}
}

// TestBuildParallelExitCodeOnPartialFailure verifies the exit code contract:
// when any workspace fails, the aggregate exit code must be non-zero.
//
// This tests the AggregateBuildExitCode helper (RED: doesn't exist yet).
func TestBuildParallelExitCodeOnPartialFailure(t *testing.T) {
	tests := []struct {
		name        string
		succeeded   int
		failed      int
		wantNonZero bool
	}{
		{
			name:        "all success → exit 0",
			succeeded:   5,
			failed:      0,
			wantNonZero: false,
		},
		{
			name:        "one failure → non-zero",
			succeeded:   4,
			failed:      1,
			wantNonZero: true,
		},
		{
			name:        "all failed → non-zero",
			succeeded:   0,
			failed:      3,
			wantNonZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := AggregateBuildExitCode(tt.succeeded, tt.failed)
			if tt.wantNonZero {
				assert.NotEqual(t, 0, code,
					"exit code must be non-zero when %d workspace(s) failed", tt.failed)
			} else {
				assert.Equal(t, 0, code,
					"exit code must be 0 when all workspaces succeeded")
			}
		})
	}
}

// =============================================================================
// Detach / Background Execution Tests
// =============================================================================

// TestBuildParallelDetach_FlagCanBeSet verifies --detach can be set on buildCmd.
//
// RED: --detach flag does not exist yet.
func TestBuildParallelDetach_FlagCanBeSet(t *testing.T) {
	flag := buildCmd.Flags().Lookup("detach")
	require.NotNil(t, flag, "buildCmd must have --detach flag")

	require.NoError(t, buildCmd.Flags().Set("detach", "true"))
	t.Cleanup(func() {
		_ = buildCmd.Flags().Set("detach", "false")
	})

	val, err := buildCmd.Flags().GetBool("detach")
	require.NoError(t, err)
	assert.True(t, val, "--detach true should be readable as bool true")
}

// TestBuildParallelDetach_SessionIDFormat verifies FormatBuildSessionID produces
// a non-empty session ID string for background builds.
//
// RED: FormatBuildSessionID does not exist yet.
func TestBuildParallelDetach_SessionIDFormat(t *testing.T) {
	id := FormatBuildSessionID("abc123")
	assert.NotEmpty(t, id,
		"FormatBuildSessionID must return a non-empty string")
	assert.Contains(t, id, "abc123",
		"session ID output must include the session identifier")
	assert.Contains(t, id, "dvm build status",
		"session ID output must hint to use 'dvm build status' to monitor")
}

// =============================================================================
// Build Status with Session Tests
// =============================================================================

// TestBuildStatusNoSession_HintMentionsBuildAll verifies that the no-session
// message includes a hint to start a build with --all or a scope flag.
// Per CLI architect review, the hint should mention 'dvm build --all'.
//
// RED: FormatNoActiveBuildSessionMessage currently returns bare message
//
//	without the hint. Needs updating for #213.
func TestBuildStatusNoSession_HintMentionsBuildAll(t *testing.T) {
	msg := FormatNoActiveBuildSessionMessage()
	assert.Contains(t, msg, "no active build",
		"no-session message must contain 'no active build'")
	// RED: hint not yet added
	assert.Contains(t, msg, "dvm build",
		"no-session message should hint to run 'dvm build --all' or similar")
}
