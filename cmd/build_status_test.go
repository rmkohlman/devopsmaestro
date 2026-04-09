package cmd

// =============================================================================
// TDD Phase 2 (RED): Build Status Subcommand — Issue #206
// =============================================================================
// These tests cover `dvm build status`:
//   - The subcommand is registered under buildCmd
//   - Output flags (-o table/json/yaml) are registered
//   - Table headers are ALL-CAPS
//   - Summary line format
//   - "no active build session" message when no build is running
//
// RED state: These tests FAIL because buildStatusCmd doesn't exist yet.
// GREEN state: After dvm-core implements Issue #206.
// =============================================================================

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =========================================================================
// R4: dvm build status subcommand must exist
// =========================================================================

// TestBuildStatusCmd_Exists verifies that `dvm build status` is registered
// as a subcommand of buildCmd.
//
// RED: buildStatusCmd does not exist in build.go yet.
func TestBuildStatusCmd_Exists(t *testing.T) {
	statusCmd := findSubcommand(buildCmd, "status")
	require.NotNil(t, statusCmd,
		"buildCmd should have a 'status' subcommand (dvm build status)")
}

// TestBuildStatusCmd_HasCorrectUse verifies the Use field.
func TestBuildStatusCmd_HasCorrectUse(t *testing.T) {
	statusCmd := findSubcommand(buildCmd, "status")
	require.NotNil(t, statusCmd)
	assert.Equal(t, "status", statusCmd.Use,
		"build status subcommand Use field should be 'status'")
}

// TestBuildStatusCmd_HasRunE verifies the subcommand has a RunE function.
func TestBuildStatusCmd_HasRunE(t *testing.T) {
	statusCmd := findSubcommand(buildCmd, "status")
	require.NotNil(t, statusCmd)
	assert.NotNil(t, statusCmd.RunE,
		"build status subcommand should have a RunE function")
}

// =========================================================================
// R4: -o / --output flag must be registered on build status
// =========================================================================

// TestBuildStatusCmd_HasOutputFlag verifies that -o/--output is registered
// on the status subcommand via AddOutputFlag().
//
// RED: flag not registered yet.
func TestBuildStatusCmd_HasOutputFlag(t *testing.T) {
	statusCmd := findSubcommand(buildCmd, "status")
	require.NotNil(t, statusCmd)

	flag := statusCmd.Flags().Lookup("output")
	require.NotNil(t, flag,
		"build status should have -o/--output flag (AddOutputFlag not called yet)")
	assert.Equal(t, "o", flag.Shorthand)
	assert.Equal(t, "string", flag.Value.Type())
}

// TestBuildStatusCmd_OutputFlag_SupportsTableDefault verifies that the default
// output format is "table" (empty string is also acceptable as table default).
func TestBuildStatusCmd_OutputFlag_SupportsTableDefault(t *testing.T) {
	statusCmd := findSubcommand(buildCmd, "status")
	require.NotNil(t, statusCmd)

	flag := statusCmd.Flags().Lookup("output")
	require.NotNil(t, flag)
	// Default should be "" (table is implicit) or "table"
	assert.True(t, flag.DefValue == "" || flag.DefValue == "table",
		"output flag default should be empty string or 'table', got %q", flag.DefValue)
}

// =========================================================================
// R5: Table column headers ALL-CAPS
// =========================================================================

// TestBuildStatusTableHeaders_AllCaps verifies that the build status table
// uses ALL-CAPS column headers as required by the CLI architect review.
// Headers must include: WORKSPACE, APP, STATUS, DURATION, ERROR
//
// RED: BuildStatusTableHeaders constant/function doesn't exist yet.
func TestBuildStatusTableHeaders_AllCaps(t *testing.T) {
	headers := BuildStatusTableHeaders()
	require.NotNil(t, headers, "BuildStatusTableHeaders() must return non-nil slice")
	require.NotEmpty(t, headers, "BuildStatusTableHeaders() must return at least one header")

	required := []string{"WORKSPACE", "APP", "STATUS", "DURATION", "ERROR"}
	for _, req := range required {
		found := false
		for _, h := range headers {
			if h == req {
				found = true
				break
			}
		}
		assert.True(t, found, "build status table must include %q column header", req)
	}
}

// =========================================================================
// R6: Summary line format
// =========================================================================

// TestBuildStatusSummaryLine_Format verifies the summary line format:
// "Build complete: N succeeded, M failed (T total) in Xs"
//
// RED: FormatBuildSummaryLine doesn't exist yet.
func TestBuildStatusSummaryLine_Format(t *testing.T) {
	tests := []struct {
		name      string
		succeeded int
		failed    int
		total     int
		wantParts []string
	}{
		{
			name:      "all success",
			succeeded: 3,
			failed:    0,
			total:     3,
			wantParts: []string{"3 succeeded", "0 failed", "3 total"},
		},
		{
			name:      "mixed results",
			succeeded: 2,
			failed:    1,
			total:     3,
			wantParts: []string{"2 succeeded", "1 failed", "3 total"},
		},
		{
			name:      "all failed",
			succeeded: 0,
			failed:    4,
			total:     4,
			wantParts: []string{"0 succeeded", "4 failed", "4 total"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := FormatBuildSummaryLine(tt.succeeded, tt.failed, tt.total)
			for _, part := range tt.wantParts {
				assert.Contains(t, line, part,
					"summary line should contain %q (got %q)", part, line)
			}
		})
	}
}

// =========================================================================
// No active build session → "no active build session" message
// =========================================================================

// TestBuildStatusCmd_NoActiveBuildSession verifies that when no build session
// is active, the status command outputs a clear message.
//
// RED: Requires the "no active build session" handling path in buildStatusCmd.
func TestBuildStatusCmd_NoActiveBuildSession(t *testing.T) {
	// The message must contain "no active build" to be useful to the user.
	msg := FormatNoActiveBuildSessionMessage()
	assert.Contains(t, msg, "no active build",
		"no-session message should contain 'no active build'")
}

// =========================================================================
// Dry-run table for hierarchy builds includes (dry-run) status column value
// =========================================================================

// TestBuildDryRunStatus_DryRunAnnotation verifies that the dry-run table uses
// "(dry-run)" in the STATUS column, consistent with the CLI architect requirement.
//
// RED: Requires FormatBuildDryRunStatusValue function.
func TestBuildDryRunStatus_DryRunAnnotation(t *testing.T) {
	val := FormatBuildDryRunStatusValue()
	assert.Equal(t, "(dry-run)", val,
		"dry-run table STATUS column must show '(dry-run)'")
}
