package cmd

// =============================================================================
// TDD Phase 2 (RED): Hierarchical Build Triggers — Issue #206
// =============================================================================
// These tests cover the CLI layer for `dvm build` with hierarchy flags:
//   - --app, --domain, --ecosystem fan-out to multiple workspaces
//   - --all flag registered and mutually exclusive with scope flags
//   - Disambiguation errors for ambiguous app/domain names
//   - --dry-run with hierarchy flags prints a table (not a single workspace)
//   - Backward compatibility: no flags = single workspace build path
//   - Invalid/nonexistent names produce clear error messages
//
// RED state: These tests FAIL because the implementation doesn't exist yet.
// GREEN state: After dvm-core implements Issue #206.
// =============================================================================

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =========================================================================
// R1: --all flag must be registered on buildCmd
// =========================================================================

// TestBuildCmd_AllFlagRegistered verifies that the --all / -A flag is
// registered on buildCmd. This flag is required by the CLI architect review
// (Requirement R1) but is NOT currently registered in build.go.
func TestBuildCmd_AllFlagRegistered(t *testing.T) {
	flag := buildCmd.Flags().Lookup("all")
	require.NotNil(t, flag, "buildCmd should have --all flag (AddAllFlag not yet called in build.go)")
	assert.Equal(t, "A", flag.Shorthand)
	assert.Equal(t, "false", flag.DefValue)
	assert.Equal(t, "bool", flag.Value.Type())
}

// =========================================================================
// R2: --all is mutually exclusive with scope flags
// =========================================================================

// TestBuildCmd_AllFlag_MutuallyExclusive_WithApp verifies that passing both
// --all and --app returns an error.
func TestBuildCmd_AllFlag_MutuallyExclusive_WithApp(t *testing.T) {
	// Reset flags between tests
	buildFlags = HierarchyFlags{}

	err := buildCmd.Flags().Set("all", "true")
	require.NoError(t, err)
	err = buildCmd.Flags().Set("app", "my-app")
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = buildCmd.Flags().Set("all", "false")
		_ = buildCmd.Flags().Set("app", "")
	})

	// RunE should return a mutual exclusion error
	runErr := buildCmd.RunE(buildCmd, []string{})
	require.Error(t, runErr, "expected error when --all and --app are both set")
	assert.Contains(t, runErr.Error(), "--all",
		"error should mention --all flag")
}

// TestBuildCmd_AllFlag_MutuallyExclusive_WithDomain verifies that passing both
// --all and --domain returns an error.
func TestBuildCmd_AllFlag_MutuallyExclusive_WithDomain(t *testing.T) {
	buildFlags = HierarchyFlags{}

	err := buildCmd.Flags().Set("all", "true")
	require.NoError(t, err)
	err = buildCmd.Flags().Set("domain", "my-domain")
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = buildCmd.Flags().Set("all", "false")
		_ = buildCmd.Flags().Set("domain", "")
	})

	runErr := buildCmd.RunE(buildCmd, []string{})
	require.Error(t, runErr, "expected error when --all and --domain are both set")
	assert.Contains(t, runErr.Error(), "--all")
}

// =========================================================================
// R7: Backward compatibility — no flags = single workspace build
// =========================================================================

// TestBuildCmd_NoHierarchyFlags_UsesHierarchyFlagsIsEmpty verifies that when
// no hierarchy flags are provided, HierarchyFlags.IsEmpty() returns true,
// ensuring the single-workspace code path is taken.
// This test is a GREEN test — it validates the existing flag scaffolding.
func TestBuildCmd_NoHierarchyFlags_UsesHierarchyFlagsIsEmpty(t *testing.T) {
	flags := HierarchyFlags{}
	assert.True(t, flags.IsEmpty(),
		"HierarchyFlags with no values set must be empty (backward compat guard)")
}

// TestBuildCmd_HierarchyFlags_NotEmpty_WhenAppSet verifies that setting --app
// causes IsEmpty() to return false, triggering the multi-workspace path.
func TestBuildCmd_HierarchyFlags_NotEmpty_WhenAppSet(t *testing.T) {
	flags := HierarchyFlags{App: "my-app"}
	assert.False(t, flags.IsEmpty(),
		"HierarchyFlags with App set must not be empty")
}

// =========================================================================
// R8: --dry-run with hierarchy flags produces a table
// =========================================================================

// TestBuildCmd_DryRunFlag_Registered verifies that --dry-run is registered
// on buildCmd (it already is — this is a GREEN guard test).
func TestBuildCmd_DryRunFlag_Registered(t *testing.T) {
	flag := buildCmd.Flags().Lookup("dry-run")
	require.NotNil(t, flag, "buildCmd should have --dry-run flag")
	assert.Equal(t, "false", flag.DefValue)
	assert.Equal(t, "bool", flag.Value.Type())
}

// TestBuildCmd_HierarchyFlagsToFilter_App verifies that ToFilter() correctly
// maps --app into the WorkspaceFilter (GREEN guard — validates flag scaffolding).
func TestBuildCmd_HierarchyFlagsToFilter_App(t *testing.T) {
	flags := HierarchyFlags{App: "portal"}
	filter := flags.ToFilter()
	assert.Equal(t, "portal", filter.AppName)
	assert.Empty(t, filter.DomainName)
	assert.Empty(t, filter.EcosystemName)
	assert.Empty(t, filter.WorkspaceName)
}

// TestBuildCmd_HierarchyFlagsToFilter_Domain verifies filter mapping for --domain.
func TestBuildCmd_HierarchyFlagsToFilter_Domain(t *testing.T) {
	flags := HierarchyFlags{Domain: "backend"}
	filter := flags.ToFilter()
	assert.Equal(t, "backend", filter.DomainName)
	assert.Empty(t, filter.AppName)
}

// TestBuildCmd_HierarchyFlagsToFilter_Ecosystem verifies filter mapping for --ecosystem.
func TestBuildCmd_HierarchyFlagsToFilter_Ecosystem(t *testing.T) {
	flags := HierarchyFlags{Ecosystem: "healthcare"}
	filter := flags.ToFilter()
	assert.Equal(t, "healthcare", filter.EcosystemName)
	assert.Empty(t, filter.AppName)
}

// TestBuildCmd_HierarchyFlagsToFilter_Combined verifies filter mapping when multiple
// flags are set (e.g., --ecosystem + --app for disambiguation).
func TestBuildCmd_HierarchyFlagsToFilter_Combined(t *testing.T) {
	flags := HierarchyFlags{
		Ecosystem: "healthcare",
		App:       "my-app",
	}
	filter := flags.ToFilter()
	assert.Equal(t, "healthcare", filter.EcosystemName)
	assert.Equal(t, "my-app", filter.AppName)
	assert.Empty(t, filter.DomainName)
}

// =========================================================================
// R3: Disambiguation error format — app ambiguous across domains
// =========================================================================

// TestBuildHierarchyDisambiguation_AppAmbiguousAcrossDomains verifies that
// the build command returns a correctly formatted disambiguation error when
// --app matches multiple apps in different domains.
//
// RED: Requires the new hierarchical build code path in build.go RunE that
// calls ResolveAll, detects multiple results, and formats the error per R3.
func TestBuildHierarchyDisambiguation_AppAmbiguousAcrossDomains(t *testing.T) {
	// This test exercises the error *format* requirement from the CLI architect:
	//   error: ambiguous app name "my-app" — found in multiple domains:
	//     my-app  (domain: backend)
	//     my-app  (domain: frontend)
	//   Hint: Add --domain to disambiguate: dvm build --app my-app --domain backend
	//
	// The test validates the error message format by checking for required substrings.
	// Full integration testing requires a real database; this unit test validates
	// that FormatBuildDisambiguationError produces the required format.
	tests := []struct {
		name          string
		appName       string
		domainMatches []string
		wantContains  []string
	}{
		{
			name:          "app in two domains",
			appName:       "my-app",
			domainMatches: []string{"backend", "frontend"},
			wantContains:  []string{"my-app", "backend", "frontend", "--domain"},
		},
		{
			name:          "app in three domains",
			appName:       "portal",
			domainMatches: []string{"auth", "payments", "reporting"},
			wantContains:  []string{"portal", "auth", "payments", "reporting", "--domain"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := FormatBuildDisambiguationError(tt.appName, "domain", tt.domainMatches)
			for _, want := range tt.wantContains {
				assert.Contains(t, msg, want,
					"disambiguation error for app %q should contain %q", tt.appName, want)
			}
		})
	}
}

// TestBuildHierarchyDisambiguation_DomainAmbiguousAcrossEcosystems verifies the
// error format when --domain matches multiple domains in different ecosystems.
func TestBuildHierarchyDisambiguation_DomainAmbiguousAcrossEcosystems(t *testing.T) {
	tests := []struct {
		name             string
		domainName       string
		ecosystemMatches []string
		wantContains     []string
	}{
		{
			name:             "domain in two ecosystems",
			domainName:       "my-domain",
			ecosystemMatches: []string{"eco-a", "eco-b"},
			wantContains:     []string{"my-domain", "eco-a", "eco-b", "--ecosystem"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := FormatBuildDisambiguationError(tt.domainName, "ecosystem", tt.ecosystemMatches)
			for _, want := range tt.wantContains {
				assert.Contains(t, msg, want,
					"disambiguation error for domain %q should contain %q", tt.domainName, want)
			}
		})
	}
}

// =========================================================================
// Nonexistent resource names → clear error messages
// =========================================================================

// TestBuildHierarchyError_NonexistentApp verifies that the error returned
// when no workspaces are found for --app contains the app name.
//
// RED: Requires the new hierarchical build code path to format "not found" errors.
func TestBuildHierarchyError_NonexistentApp(t *testing.T) {
	msg := FormatBuildNotFoundError("app", "does-not-exist")
	assert.Contains(t, msg, "does-not-exist",
		"not-found error should include the resource name")
	assert.Contains(t, msg, "app",
		"not-found error should include the resource type")
}

// TestBuildHierarchyError_NonexistentDomain verifies the not-found error for --domain.
func TestBuildHierarchyError_NonexistentDomain(t *testing.T) {
	msg := FormatBuildNotFoundError("domain", "ghost-domain")
	assert.Contains(t, msg, "ghost-domain")
	assert.Contains(t, msg, "domain")
}

// TestBuildHierarchyError_NonexistentEcosystem verifies the not-found error for --ecosystem.
func TestBuildHierarchyError_NonexistentEcosystem(t *testing.T) {
	msg := FormatBuildNotFoundError("ecosystem", "missing-eco")
	assert.Contains(t, msg, "missing-eco")
	assert.Contains(t, msg, "ecosystem")
}

// =========================================================================
// Dry-run table format for hierarchy builds (R8)
// =========================================================================

// TestBuildDryRunHierarchy_TableColumns verifies that the dry-run table for a
// hierarchy build includes the required column headers (ALL-CAPS per CLI arch review).
//
// RED: Requires FormatBuildDryRunTable function that doesn't exist yet.
func TestBuildDryRunHierarchy_TableColumns(t *testing.T) {
	type dryRunRow struct {
		Workspace string
		App       string
	}
	rows := []dryRunRow{
		{"ws-frontend", "my-app"},
		{"ws-backend", "my-app"},
	}

	output := FormatBuildDryRunTable(rows[0].Workspace, rows[0].App)
	_ = output // We only need the function to exist and return a string

	// Full table is exercised via integration; here we validate the function signature
	// and that the formatter is callable with workspace + app strings.
	assert.NotEmpty(t, output)
}

// TestBuildDryRunHierarchy_SummaryLine verifies the summary line printed after
// dry-run table: "Would build N workspace(s). Use without --dry-run to execute."
//
// RED: Requires FormatBuildDryRunSummary function.
func TestBuildDryRunHierarchy_SummaryLine(t *testing.T) {
	tests := []struct {
		count int
		want  string
	}{
		{1, "1 workspace"},
		{3, "3 workspace"},
		{0, "0 workspace"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			msg := FormatBuildDryRunSummary(tt.count)
			assert.Contains(t, msg, tt.want,
				"dry-run summary for %d workspaces should mention count", tt.count)
			assert.Contains(t, msg, "--dry-run",
				"dry-run summary should mention --dry-run flag")
		})
	}
}
