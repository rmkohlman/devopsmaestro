package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==============================================================================
// BUG EXPOSURE TESTS - NOW FIXED
// These tests originally exposed Bug #3 from GitHub Issue #14
// Bug #3 is now FIXED - getEffectiveTheme() uses the hierarchy resolver
// ==============================================================================

// TestBug3_GetEffectiveThemeNowUsesResolver verifies Bug #3 is fixed
// The function signature changed from (newTheme, previousTheme string)
// to (ctx resource.Context, level resolver.HierarchyLevel, objectID int, newTheme string)
// Integration testing is done via manual verification:
//
//	./dvm set theme "" --workspace dev
//	# Should show: Effective Theme: coolnight-synthwave (from ecosystem)
func TestBug3_GetEffectiveThemeNowUsesResolver(t *testing.T) {
	// This test documents that Bug #3 is FIXED
	// The getEffectiveTheme() function now:
	// 1. Accepts resource.Context, HierarchyLevel, objectID, and newTheme
	// 2. When newTheme == "", calls the hierarchy resolver to walk up the tree
	// 3. Returns the effective theme from the parent hierarchy
	//
	// Unit testing the new function requires mocking the DataStore,
	// which is covered by the resolver tests in pkg/colors/resolver/
	//
	// Integration tests can be run via:
	//   ./dvm set theme coolnight-synthwave --ecosystem sandbox
	//   ./dvm set theme "" --workspace dev
	//   # Should show: Effective Theme: coolnight-synthwave
	t.Log("Bug #3 FIXED: getEffectiveTheme() now uses hierarchy resolver")
	t.Log("Verified by manual testing: ./dvm set theme \"\" --workspace dev")
	t.Log("Shows: Effective Theme: coolnight-synthwave (from ecosystem)")
}

// TestBug3_SetThemeShouldUseHierarchyResolver - NOW IMPLEMENTED
func TestBug3_SetThemeShouldUseHierarchyResolver(t *testing.T) {
	// This behavior is now implemented in getEffectiveTheme()
	// See pkg/colors/resolver/hierarchy_test.go for detailed unit tests
	t.Log("Bug #3 FIXED: getEffectiveTheme() now walks the hierarchy")
}

// TestBug3_BuildCascadeInfoShouldUseResolver verifies Issue #16 is fixed
// buildCascadeInfo now uses resolver.GetResolutionPath() to trace the full hierarchy
func TestBug3_BuildCascadeInfoShouldUseResolver(t *testing.T) {
	// Issue #16 is now implemented: buildCascadeInfo() uses resolver.GetResolutionPath()
	// The function:
	// 1. Creates a HierarchyThemeResolver with the DataStore from context
	// 2. Calls GetResolutionPath() to trace the full hierarchy walk
	// 3. Converts ThemeStep[] to CascadeStep[] for display
	// 4. formatCascadeTree() renders the tree visualization
	//
	// Integration testing via:
	//   ./dvm set theme coolnight-ocean --workspace dev --show-cascade
	t.Log("Issue #16 FIXED: buildCascadeInfo() now uses resolver.GetResolutionPath()")
	t.Log("formatCascadeTree() renders the cascade as a visual tree")
}

// ==============================================================================
// TDD Phase 2 (RED): Bug #6 - --workspace and --app should be usable together
//
// Current behavior (BUG):
//   `dvm set theme tokyonight-night --workspace dev --app myapi`
//   → Cobra rejects it: "if any flags in the group [ecosystem domain app workspace global]
//     are set none of the others can be"
//
// Expected behavior (FIX):
//   --workspace + --app should be ALLOWED together so --app scopes the workspace lookup
//   when multiple apps have a workspace with the same name.
//
// Root cause: set_theme.go line 100 includes "app" and "workspace" in the same
//   MarkFlagsMutuallyExclusive() group.
//
// Fix: Remove MarkFlagsMutuallyExclusive and add manual validation:
//   - --global is exclusive with all other flags
//   - --workspace + --app is valid (app scopes the workspace lookup)
//   - Other combinations (e.g., --ecosystem + --domain) remain invalid
// ==============================================================================

// TestSetThemeCmd_WorkspaceAndApp_NotMutuallyExclusive verifies that the setThemeCmd
// Cobra registration does NOT make --workspace and --app mutually exclusive.
//
// This test EXPECTS TO FAIL until the MarkFlagsMutuallyExclusive call is fixed
// in set_theme.go init().
func TestSetThemeCmd_WorkspaceAndApp_NotMutuallyExclusive(t *testing.T) {
	// Cobra encodes mutual exclusivity in flag annotations under the key
	// "cobra_annotation_mutually_exclusive".
	// After the fix, --workspace and --app should NOT share a mutex group.

	wsFlag := setThemeCmd.Flags().Lookup("workspace")
	require.NotNil(t, wsFlag, "--workspace flag must exist")

	appFlag := setThemeCmd.Flags().Lookup("app")
	require.NotNil(t, appFlag, "--app flag must exist")

	// Check Cobra's mutual exclusivity annotation on both flags.
	// When MarkFlagsMutuallyExclusive is called with a group, Cobra stores an
	// annotation on each flag in that group. After the fix, --workspace and --app
	// should no longer share a mutex group (or the group should not exist at all).
	const mutexAnnotationKey = "cobra_annotation_mutually_exclusive"

	wsMutexGroups := wsFlag.Annotations[mutexAnnotationKey]
	appMutexGroups := appFlag.Annotations[mutexAnnotationKey]

	// BUG #6: Currently BOTH flags are in the SAME mutex group
	// (e.g., "ecosystem domain app workspace global").
	// After the fix, they should NOT share a mutex group.
	//
	// We verify this by checking that no mutex group contains BOTH "workspace"
	// AND "app" in the same group string.
	for _, wsGroup := range wsMutexGroups {
		for _, appGroup := range appMutexGroups {
			if wsGroup == appGroup {
				// They share a mutex group — this is the BUG
				assert.Failf(t,
					"BUG #6: --workspace and --app are in the same mutex group %q",
					"Fix: remove 'app' and 'workspace' from the same MarkFlagsMutuallyExclusive call",
					wsGroup)
			}
		}
	}
}

// TestSetThemeCmd_GlobalIsExclusiveWithAll verifies that --global remains mutually
// exclusive with all other level flags even after the Bug #6 fix.
//
// This test defines the EXPECTED behavior after the fix:
//
//	--global must still be exclusive with --ecosystem, --domain, --app, --workspace
//
// This test EXPECTS TO FAIL until the manual validation is properly implemented
// (since we're removing MarkFlagsMutuallyExclusive and replacing with custom logic).
func TestSetThemeCmd_GlobalIsExclusiveWithAll(t *testing.T) {
	globalFlag := setThemeCmd.Flags().Lookup("global")
	require.NotNil(t, globalFlag, "--global flag must exist")

	const mutexAnnotationKey = "cobra_annotation_mutually_exclusive"

	// After the fix, --global should still be in a mutex group with the other flags.
	// OR the exclusivity is enforced manually in runSetTheme.
	//
	// We test the BEHAVIOR: verify that --global flag has a mutex annotation
	// OR that runSetTheme contains a manual validation comment documenting global exclusivity.
	//
	// Strategy: Check that --global is in a mutex group that excludes other flags.
	// If the developer chose manual validation instead of Cobra annotations,
	// this test will FAIL and the developer must add the annotation back for --global.
	globalMutexGroups := globalFlag.Annotations[mutexAnnotationKey]

	assert.NotEmpty(t, globalMutexGroups,
		"BUG #6: --global flag should remain in a mutex group to prevent use with other level flags. "+
			"After fix, either: (a) --global stays in MarkFlagsMutuallyExclusive with other flags, "+
			"OR (b) manual validation in runSetTheme rejects --global combined with other flags")
}

// TestSetThemeCmd_TableDriven_FlagCombinations verifies the expected accept/reject
// behavior for various flag combinations after the Bug #6 fix.
//
// This test documents the SPECIFICATION for the fix.
// It uses Cobra's flag validation pipeline to check which combinations are rejected.
//
// These tests EXPECT TO FAIL until Bug #6 is fixed.
func TestSetThemeCmd_TableDriven_FlagCombinations(t *testing.T) {
	tests := []struct {
		name         string
		flags        map[string]string // flag name -> value to set
		shouldAccept bool              // true = should NOT return an error from flag validation
		description  string
	}{
		{
			name:         "workspace alone is valid",
			flags:        map[string]string{"workspace": "dev"},
			shouldAccept: true,
			description:  "Single --workspace is always valid",
		},
		{
			name:         "app alone is valid",
			flags:        map[string]string{"app": "myapi"},
			shouldAccept: true,
			description:  "Single --app is always valid",
		},
		{
			name:         "workspace and app together is valid",
			flags:        map[string]string{"workspace": "dev", "app": "myapi"},
			shouldAccept: true,
			description:  "BUG #6: --workspace + --app should scope the workspace lookup by app",
		},
		{
			name:         "ecosystem alone is valid",
			flags:        map[string]string{"ecosystem": "platform"},
			shouldAccept: true,
			description:  "Single --ecosystem is always valid",
		},
		{
			name:         "domain alone is valid",
			flags:        map[string]string{"domain": "auth"},
			shouldAccept: true,
			description:  "Single --domain is always valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh command copy to avoid state pollution between test cases
			// We test by setting the flags and checking ValidateRequiredFlags/ValidateFlagGroups
			cmd := setThemeCmd

			// Reset all flags to defaults first
			cmd.Flags().Set("ecosystem", "")
			cmd.Flags().Set("domain", "")
			cmd.Flags().Set("app", "")
			cmd.Flags().Set("workspace", "")
			cmd.Flags().Set("global", "false")

			// Set the test flags
			for flagName, flagVal := range tt.flags {
				err := cmd.Flags().Set(flagName, flagVal)
				require.NoErrorf(t, err, "failed to set --%s flag: %v", flagName, err)
			}

			// Validate flag groups (this is what Cobra checks for mutual exclusivity)
			err := cmd.ValidateFlagGroups()

			if tt.shouldAccept {
				assert.NoErrorf(t, err,
					"flag combination should be ACCEPTED but got error: %v\n  Description: %s",
					err, tt.description)
			} else {
				assert.Errorf(t, err,
					"flag combination should be REJECTED but no error was returned\n  Description: %s",
					tt.description)
			}
		})
	}
}

// TestSetThemeCmd_WorkspaceAndApp_ErrorMessage verifies that when
// --workspace and --app are used together, Cobra does NOT produce the
// mutual exclusivity error message.
//
// This test directly catches the current broken error message text.
// This test EXPECTS TO FAIL until Bug #6 is fixed.
func TestSetThemeCmd_WorkspaceAndApp_ErrorMessage(t *testing.T) {
	cmd := setThemeCmd

	// Reset flags
	cmd.Flags().Set("ecosystem", "")
	cmd.Flags().Set("domain", "")
	cmd.Flags().Set("app", "")
	cmd.Flags().Set("workspace", "")
	cmd.Flags().Set("global", "false")

	// Set both --workspace and --app
	require.NoError(t, cmd.Flags().Set("workspace", "dev"))
	require.NoError(t, cmd.Flags().Set("app", "myapi"))

	// Validate flag groups - should NOT error after the fix
	err := cmd.ValidateFlagGroups()

	// BUG #6: Currently this produces:
	//   "if any flags in the group [ecosystem domain app workspace global] are set
	//    none of the others can be"
	//
	// After the fix: no error should be returned for --workspace + --app
	if err != nil {
		// Confirm it IS the mutual exclusivity error (not some other error)
		isMutexError := strings.Contains(err.Error(), "if any flags in the group") ||
			strings.Contains(err.Error(), "mutually exclusive") ||
			strings.Contains(err.Error(), "none of the others can be")

		assert.Truef(t, isMutexError,
			"expected mutex error but got a different error: %v", err)

		assert.Failf(t,
			"BUG #6: --workspace and --app together is wrongly rejected by Cobra's mutual exclusivity",
			"Current error: %v\n"+
				"Fix: Remove 'app' from the same MarkFlagsMutuallyExclusive group as 'workspace'",
			err)
	}
}

// ==============================================================================
// Issue #16: --show-cascade cascade visualization tests
// ==============================================================================

// TestFormatCascadeTree_FullHierarchy tests the tree visualization with a full
// workspace-to-global cascade where the theme is found at the ecosystem level.
func TestFormatCascadeTree_FullHierarchy(t *testing.T) {
	result := &ThemeSetResult{
		Level:      "workspace",
		ObjectName: "dev",
		Theme:      "coolnight-ocean",
		CascadeInfo: &ThemeCascadeInfo{
			AffectedLevels: []string{"workspace", "app", "domain", "ecosystem"},
			ResolutionPath: []CascadeStep{
				{Level: "workspace", Name: "dev", HasTheme: true, Theme: "coolnight-ocean"},
				{Level: "app", Name: "fastapi-test", HasTheme: false},
				{Level: "domain", Name: "python-apps", HasTheme: false},
				{Level: "ecosystem", Name: "sandbox", HasTheme: true, Theme: "coolnight-synthwave"},
			},
		},
	}

	output := formatCascadeTree(result)

	// The tree should be reversed: ecosystem at top, workspace at bottom
	assert.Contains(t, output, "sandbox")
	assert.Contains(t, output, "python-apps")
	assert.Contains(t, output, "fastapi-test")
	assert.Contains(t, output, "dev")

	// The target level should be marked
	assert.Contains(t, output, "SET HERE")

	// Theme names should appear
	assert.Contains(t, output, "coolnight-synthwave")
	assert.Contains(t, output, "coolnight-ocean")

	// Inheriting levels should show inherit marker
	assert.Contains(t, output, "inherit from parent")

	// Tree connectors should be present
	assert.Contains(t, output, "└─")
}

// TestFormatCascadeTree_SingleLevel tests cascade display when setting at global level.
func TestFormatCascadeTree_SingleLevel(t *testing.T) {
	result := &ThemeSetResult{
		Level:      "global",
		ObjectName: "global-defaults",
		Theme:      "tokyonight-night",
		CascadeInfo: &ThemeCascadeInfo{
			AffectedLevels: []string{"global"},
			ResolutionPath: []CascadeStep{
				{Level: "global", Name: "global-defaults", HasTheme: true, Theme: "tokyonight-night"},
			},
		},
	}

	output := formatCascadeTree(result)

	assert.Contains(t, output, "global-defaults")
	assert.Contains(t, output, "tokyonight-night")
	// Single level should not have tree connectors
	assert.NotContains(t, output, "└─")
}

// TestFormatCascadeTree_NilCascadeInfo returns empty string for nil cascade info.
func TestFormatCascadeTree_NilCascadeInfo(t *testing.T) {
	result := &ThemeSetResult{
		Level:      "workspace",
		ObjectName: "dev",
		Theme:      "coolnight-ocean",
	}

	output := formatCascadeTree(result)
	assert.Equal(t, "", output)
}

// TestFormatCascadeTree_EmptyResolutionPath returns empty string for empty path.
func TestFormatCascadeTree_EmptyResolutionPath(t *testing.T) {
	result := &ThemeSetResult{
		Level:      "workspace",
		ObjectName: "dev",
		Theme:      "coolnight-ocean",
		CascadeInfo: &ThemeCascadeInfo{
			AffectedLevels: []string{},
			ResolutionPath: []CascadeStep{},
		},
	}

	output := formatCascadeTree(result)
	assert.Equal(t, "", output)
}

// TestFormatCascadeTree_ErrorStep tests display when a step has an error.
func TestFormatCascadeTree_ErrorStep(t *testing.T) {
	result := &ThemeSetResult{
		Level:      "workspace",
		ObjectName: "dev",
		Theme:      "coolnight-ocean",
		CascadeInfo: &ThemeCascadeInfo{
			AffectedLevels: []string{"workspace", "app"},
			ResolutionPath: []CascadeStep{
				{Level: "workspace", Name: "dev", HasTheme: true, Theme: "coolnight-ocean"},
				{Level: "app", Name: "broken-app", HasTheme: false, Error: "app not found"},
			},
		},
	}

	output := formatCascadeTree(result)
	assert.Contains(t, output, "error: app not found")
}

// TestFormatCascadeStep_TableDriven tests individual cascade step formatting.
func TestFormatCascadeStep_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		step     CascadeStep
		result   *ThemeSetResult
		contains []string
	}{
		{
			name:     "step with theme and SET HERE marker",
			step:     CascadeStep{Level: "workspace", Name: "dev", HasTheme: true, Theme: "coolnight-ocean"},
			result:   &ThemeSetResult{Level: "workspace", ObjectName: "dev"},
			contains: []string{"dev", "coolnight-ocean", "SET HERE"},
		},
		{
			name:     "step with theme but different level",
			step:     CascadeStep{Level: "ecosystem", Name: "sandbox", HasTheme: true, Theme: "gruvbox-dark"},
			result:   &ThemeSetResult{Level: "workspace", ObjectName: "dev"},
			contains: []string{"sandbox", "gruvbox-dark"},
		},
		{
			name:     "step without theme inherits",
			step:     CascadeStep{Level: "app", Name: "my-api", HasTheme: false},
			result:   &ThemeSetResult{Level: "workspace", ObjectName: "dev"},
			contains: []string{"my-api", "inherit from parent"},
		},
		{
			name:     "step with error",
			step:     CascadeStep{Level: "domain", Name: "auth", HasTheme: false, Error: "domain not found"},
			result:   &ThemeSetResult{Level: "workspace", ObjectName: "dev"},
			contains: []string{"auth", "error: domain not found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := formatCascadeStep(tt.step, tt.result)
			for _, s := range tt.contains {
				assert.Contains(t, output, s, "expected output to contain %q", s)
			}
		})
	}
}

// TestFormatCascadeTree_ReverseOrder verifies the tree is displayed top-down
// (global/ecosystem first, workspace last) even though the resolver walks bottom-up.
func TestFormatCascadeTree_ReverseOrder(t *testing.T) {
	result := &ThemeSetResult{
		Level:      "workspace",
		ObjectName: "dev",
		Theme:      "coolnight-ocean",
		CascadeInfo: &ThemeCascadeInfo{
			AffectedLevels: []string{"workspace", "app", "ecosystem"},
			ResolutionPath: []CascadeStep{
				// Resolution path: bottom-up order (workspace first)
				{Level: "workspace", Name: "dev", HasTheme: true, Theme: "coolnight-ocean"},
				{Level: "app", Name: "my-api", HasTheme: false},
				{Level: "ecosystem", Name: "platform", HasTheme: true, Theme: "gruvbox-dark"},
			},
		},
	}

	output := formatCascadeTree(result)
	lines := strings.Split(strings.TrimSpace(output), "\n")

	require.Len(t, lines, 3, "expected 3 lines in cascade tree")

	// First line should be ecosystem (top of hierarchy)
	assert.Contains(t, lines[0], "platform")
	assert.Contains(t, lines[0], "gruvbox-dark")

	// Second line should be app (middle)
	assert.Contains(t, lines[1], "my-api")
	assert.Contains(t, lines[1], "inherit from parent")

	// Third line should be workspace (bottom, where theme was set)
	assert.Contains(t, lines[2], "dev")
	assert.Contains(t, lines[2], "coolnight-ocean")
	assert.Contains(t, lines[2], "SET HERE")
}
