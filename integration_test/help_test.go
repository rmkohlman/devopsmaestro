// Package integration provides help text and error message validation tests.
// These tests ensure dvm follows CLI conventions for help output and error messages.
package integration

import (
	"strings"
	"testing"
)

// TestHelpText_RootCommand verifies that root help output includes all expected elements.
func TestHelpText_RootCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	tests := []struct {
		name  string
		args  []string
		check func(t *testing.T, stdout string)
	}{
		{
			name: "dvm --help",
			args: []string{"--help"},
			check: func(t *testing.T, stdout string) {
				// Should have usage line
				if !strings.Contains(stdout, "Usage:") {
					t.Error("Help text should contain 'Usage:' section")
				}
				// Should have description
				if !strings.Contains(stdout, "dvm") || !strings.Contains(stdout, "CLI") {
					t.Error("Help text should contain description")
				}
				// Should list main commands
				expectedCommands := []string{"create", "get", "delete", "use"}
				for _, cmd := range expectedCommands {
					if !strings.Contains(stdout, cmd) {
						t.Errorf("Help text should list '%s' command", cmd)
					}
				}
				// Should document flags
				if !strings.Contains(stdout, "Flags:") {
					t.Error("Help text should contain 'Flags:' section")
				}
				// Should have global flags
				expectedFlags := []string{"--verbose", "--help"}
				for _, flag := range expectedFlags {
					if !strings.Contains(stdout, flag) {
						t.Errorf("Help text should document '%s' flag", flag)
					}
				}
			},
		},
		{
			name: "dvm -h (short form)",
			args: []string{"-h"},
			check: func(t *testing.T, stdout string) {
				// Should have same content as --help
				if !strings.Contains(stdout, "Usage:") {
					t.Error("Help text should contain 'Usage:' section")
				}
			},
		},
		{
			name: "dvm help",
			args: []string{"help"},
			check: func(t *testing.T, stdout string) {
				// Should have same content as --help
				if !strings.Contains(stdout, "Usage:") {
					t.Error("Help text should contain 'Usage:' section")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _ := f.AssertExitCodeWithOutput(t, 0, tt.args...)
			tt.check(t, stdout)
		})
	}
}

// TestHelpText_CreateCommand verifies create command help text.
func TestHelpText_CreateCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	tests := []struct {
		name  string
		args  []string
		check func(t *testing.T, stdout string)
	}{
		{
			name: "dvm create --help",
			args: []string{"create", "--help"},
			check: func(t *testing.T, stdout string) {
				// Should have usage line
				if !strings.Contains(stdout, "Usage:") {
					t.Error("Help text should contain 'Usage:' section")
				}
				// Should have description
				if !strings.Contains(stdout, "Create") {
					t.Error("Help text should describe 'create' command")
				}
				// Should list subcommands
				expectedSubcommands := []string{"ecosystem", "domain", "app", "workspace"}
				for _, cmd := range expectedSubcommands {
					if !strings.Contains(stdout, cmd) {
						t.Errorf("Help text should list '%s' subcommand", cmd)
					}
				}
				// Should document aliases
				if !strings.Contains(stdout, "alias") || !strings.Contains(stdout, "ws") {
					t.Error("Help text should document aliases (e.g., 'ws' for workspace)")
				}
				// Should have examples
				if !strings.Contains(stdout, "Examples:") {
					t.Error("Help text should contain 'Examples:' section")
				}
			},
		},
		{
			name: "dvm create ecosystem --help",
			args: []string{"create", "ecosystem", "--help"},
			check: func(t *testing.T, stdout string) {
				// Should have usage with correct syntax
				if !strings.Contains(stdout, "Usage:") {
					t.Error("Help text should contain 'Usage:' section")
				}
				if !strings.Contains(stdout, "create ecosystem") {
					t.Error("Usage should show 'create ecosystem' syntax")
				}
				// Should have description
				if !strings.Contains(stdout, "ecosystem") {
					t.Error("Help text should describe ecosystem")
				}
				// Should have examples
				if !strings.Contains(stdout, "Examples:") {
					t.Error("Help text should contain 'Examples:' section")
				}
			},
		},
		{
			name: "dvm create workspace --help",
			args: []string{"create", "workspace", "--help"},
			check: func(t *testing.T, stdout string) {
				// Should have usage
				if !strings.Contains(stdout, "Usage:") {
					t.Error("Help text should contain 'Usage:' section")
				}
				// Should document flags
				expectedFlags := []string{"--app", "--repo", "--description"}
				for _, flag := range expectedFlags {
					if !strings.Contains(stdout, flag) {
						t.Errorf("Help text should document '%s' flag", flag)
					}
				}
				// Should show alias
				if !strings.Contains(stdout, "ws") {
					t.Error("Help text should show 'ws' alias")
				}
				// Should have examples
				if !strings.Contains(stdout, "Examples:") {
					t.Error("Help text should contain 'Examples:' section")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _ := f.AssertExitCodeWithOutput(t, 0, tt.args...)
			tt.check(t, stdout)
		})
	}
}

// TestHelpText_GetCommand verifies get command help text.
func TestHelpText_GetCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	tests := []struct {
		name  string
		args  []string
		check func(t *testing.T, stdout string)
	}{
		{
			name: "dvm get --help",
			args: []string{"get", "--help"},
			check: func(t *testing.T, stdout string) {
				// Should have usage line
				if !strings.Contains(stdout, "Usage:") {
					t.Error("Help text should contain 'Usage:' section")
				}
				// Should have description
				if !strings.Contains(stdout, "Get") {
					t.Error("Help text should describe 'get' command")
				}
				// Should list resource types
				expectedResources := []string{"apps", "workspaces", "context"}
				for _, res := range expectedResources {
					if !strings.Contains(stdout, res) {
						t.Errorf("Help text should list '%s' resource", res)
					}
				}
				// Should document aliases
				if !strings.Contains(stdout, "alias") {
					t.Error("Help text should document resource aliases")
				}
				// Should have examples
				if !strings.Contains(stdout, "Examples:") {
					t.Error("Help text should contain 'Examples:' section")
				}
			},
		},
		{
			name: "dvm get workspaces --help",
			args: []string{"get", "workspaces", "--help"},
			check: func(t *testing.T, stdout string) {
				// Should have usage
				if !strings.Contains(stdout, "Usage:") {
					t.Error("Help text should contain 'Usage:' section")
				}
				// Should document flags
				expectedFlags := []string{"--app", "--all"}
				for _, flag := range expectedFlags {
					if !strings.Contains(stdout, flag) {
						t.Errorf("Help text should document '%s' flag", flag)
					}
				}
				// Should show alias
				if !strings.Contains(stdout, "ws") {
					t.Error("Help text should show 'ws' alias")
				}
				// Should have examples
				if !strings.Contains(stdout, "Examples:") {
					t.Error("Help text should contain 'Examples:' section")
				}
			},
		},
		{
			name: "dvm get workspace --help",
			args: []string{"get", "workspace", "--help"},
			check: func(t *testing.T, stdout string) {
				// Should have usage
				if !strings.Contains(stdout, "Usage:") {
					t.Error("Help text should contain 'Usage:' section")
				}
				// Should document output format flag
				if !strings.Contains(stdout, "output") {
					t.Error("Help text should document output format flag")
				}
				// Should have examples
				if !strings.Contains(stdout, "Examples:") {
					t.Error("Help text should contain 'Examples:' section")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _ := f.AssertExitCodeWithOutput(t, 0, tt.args...)
			tt.check(t, stdout)
		})
	}
}

// TestHelpText_DeleteCommand verifies delete command help text.
func TestHelpText_DeleteCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	tests := []struct {
		name  string
		args  []string
		check func(t *testing.T, stdout string)
	}{
		{
			name: "dvm delete --help",
			args: []string{"delete", "--help"},
			check: func(t *testing.T, stdout string) {
				// Should have usage line
				if !strings.Contains(stdout, "Usage:") {
					t.Error("Help text should contain 'Usage:' section")
				}
				// Should have description
				if !strings.Contains(stdout, "Delete") {
					t.Error("Help text should describe 'delete' command")
				}
				// Should list resources
				expectedResources := []string{"app", "workspace"}
				for _, res := range expectedResources {
					if !strings.Contains(stdout, res) {
						t.Errorf("Help text should mention '%s' resource", res)
					}
				}
				// Should document aliases
				if !strings.Contains(stdout, "alias") {
					t.Error("Help text should document resource aliases")
				}
				// Should have examples
				if !strings.Contains(stdout, "Examples:") {
					t.Error("Help text should contain 'Examples:' section")
				}
			},
		},
		{
			name: "dvm delete workspace --help",
			args: []string{"delete", "workspace", "--help"},
			check: func(t *testing.T, stdout string) {
				// Should have usage
				if !strings.Contains(stdout, "Usage:") {
					t.Error("Help text should contain 'Usage:' section")
				}
				// Should document force flag
				if !strings.Contains(stdout, "force") {
					t.Error("Help text should document --force flag")
				}
				// Should have examples
				if !strings.Contains(stdout, "Examples:") {
					t.Error("Help text should contain 'Examples:' section")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _ := f.AssertExitCodeWithOutput(t, 0, tt.args...)
			tt.check(t, stdout)
		})
	}
}

// TestHelpText_UseCommand verifies use command help text.
func TestHelpText_UseCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	tests := []struct {
		name  string
		args  []string
		check func(t *testing.T, stdout string)
	}{
		{
			name: "dvm use --help",
			args: []string{"use", "--help"},
			check: func(t *testing.T, stdout string) {
				// Should have usage line
				if !strings.Contains(stdout, "Usage:") {
					t.Error("Help text should contain 'Usage:' section")
				}
				// Should describe context switching
				if !strings.Contains(stdout, "context") {
					t.Error("Help text should describe context switching")
				}
				// Should list resources
				expectedResources := []string{"app", "workspace"}
				for _, res := range expectedResources {
					if !strings.Contains(stdout, res) {
						t.Errorf("Help text should mention '%s' resource", res)
					}
				}
				// Should document aliases
				if !strings.Contains(stdout, "alias") {
					t.Error("Help text should document resource aliases")
				}
				// Should have examples
				if !strings.Contains(stdout, "Examples:") {
					t.Error("Help text should contain 'Examples:' section")
				}
			},
		},
		{
			name: "dvm use app --help",
			args: []string{"use", "app", "--help"},
			check: func(t *testing.T, stdout string) {
				// Should have usage
				if !strings.Contains(stdout, "Usage:") {
					t.Error("Help text should contain 'Usage:' section")
				}
				// Should mention 'none' to clear
				if !strings.Contains(stdout, "none") {
					t.Error("Help text should mention 'none' to clear context")
				}
				// Should have examples
				if !strings.Contains(stdout, "Examples:") {
					t.Error("Help text should contain 'Examples:' section")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _ := f.AssertExitCodeWithOutput(t, 0, tt.args...)
			tt.check(t, stdout)
		})
	}
}

// TestHelpText_ConsistencyAcrossCommands verifies that all subcommands follow consistent patterns.
func TestHelpText_ConsistencyAcrossCommands(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Test all major commands have consistent help structure
	commands := [][]string{
		{"create", "--help"},
		{"get", "--help"},
		{"delete", "--help"},
		{"use", "--help"},
		{"create", "ecosystem", "--help"},
		{"create", "domain", "--help"},
		{"create", "app", "--help"},
		{"create", "workspace", "--help"},
		{"get", "ecosystems", "--help"},
		{"get", "domains", "--help"},
		{"get", "apps", "--help"},
		{"get", "workspaces", "--help"},
	}

	for _, cmd := range commands {
		t.Run(strings.Join(cmd, " "), func(t *testing.T) {
			stdout, _ := f.AssertExitCodeWithOutput(t, 0, cmd...)

			// All help text should have these sections
			requiredSections := []string{"Usage:", "Flags:", "Examples:"}
			for _, section := range requiredSections {
				if !strings.Contains(stdout, section) {
					t.Errorf("Help text for '%s' should contain '%s' section",
						strings.Join(cmd, " "), section)
				}
			}

			// Should not be empty
			if len(strings.TrimSpace(stdout)) < 50 {
				t.Errorf("Help text for '%s' seems too short", strings.Join(cmd, " "))
			}
		})
	}
}

// TestErrorMessage_Format verifies that error messages follow consistent format.
func TestErrorMessage_Format(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	tests := []struct {
		name             string
		args             []string
		expectedExitCode int
		check            func(t *testing.T, stdout, stderr string)
	}{
		{
			name:             "not found error format",
			args:             []string{"get", "ecosystem", "nonexistent"},
			expectedExitCode: 1,
			check: func(t *testing.T, stdout, stderr string) {
				// Error should be in stderr
				if stderr == "" {
					t.Error("Error message should be written to stderr, not stdout")
				}
				// Should contain error information
				lowerStderr := strings.ToLower(stderr)
				if !strings.Contains(lowerStderr, "not found") && !strings.Contains(lowerStderr, "error") {
					t.Error("Error message should indicate resource not found")
				}
			},
		},
		{
			name:             "validation error format",
			args:             []string{"create", "ecosystem", ""},
			expectedExitCode: 1,
			check: func(t *testing.T, stdout, stderr string) {
				// Error should be in stderr
				if stderr == "" {
					t.Error("Validation error should be written to stderr")
				}
				// Should indicate what's wrong
				lowerStderr := strings.ToLower(stderr)
				if !strings.Contains(lowerStderr, "empty") && !strings.Contains(lowerStderr, "name") {
					t.Error("Validation error should explain the problem")
				}
			},
		},
		{
			name:             "missing context hint",
			args:             []string{"create", "domain", "test-domain"},
			expectedExitCode: 0, // Missing context returns 0 with hint
			check: func(t *testing.T, stdout, stderr string) {
				// Should have helpful message in stdout
				lowerStdout := strings.ToLower(stdout)
				if !strings.Contains(lowerStdout, "hint") && !strings.Contains(lowerStdout, "ecosystem") {
					t.Error("Missing context should provide helpful hint")
				}
			},
		},
		{
			name:             "invalid flag error",
			args:             []string{"get", "ecosystems", "--invalid-flag"},
			expectedExitCode: 1,
			check: func(t *testing.T, stdout, stderr string) {
				// Error should be in stderr
				if stderr == "" {
					t.Error("Invalid flag error should be written to stderr")
				}
				// Should mention the flag
				if !strings.Contains(stderr, "flag") && !strings.Contains(stderr, "unknown") {
					t.Error("Error should mention invalid flag")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr := f.AssertExitCodeWithOutput(t, tt.expectedExitCode, tt.args...)
			tt.check(t, stdout, stderr)
		})
	}
}

// TestErrorMessage_NoStackTraces verifies that error messages don't expose stack traces
// to users unless --verbose is enabled.
func TestErrorMessage_NoStackTraces(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Test various error scenarios
	errorScenarios := [][]string{
		{"get", "ecosystem", "nonexistent"},
		{"create", "ecosystem", ""},
		{"delete", "workspace", "nonexistent"},
		{"get", "ecosystems", "--invalid-flag"},
	}

	for _, args := range errorScenarios {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			exitCode, stdout, stderr := f.RunDVMWithExitCode(args...)

			// Should be an error
			if exitCode == 0 {
				t.Skip("Command succeeded, not an error scenario")
			}

			// Combine stdout and stderr for checking
			output := stdout + stderr

			// Should NOT contain stack trace indicators
			stackTraceIndicators := []string{
				"goroutine",
				"panic",
				".go:", // Go file line numbers
				"runtime.",
			}

			for _, indicator := range stackTraceIndicators {
				if strings.Contains(output, indicator) {
					t.Errorf("Error output should not contain stack trace (found '%s')\nOutput: %s",
						indicator, output)
				}
			}
		})
	}
}

// TestErrorMessage_ActionableMessages verifies that error messages provide actionable hints.
func TestErrorMessage_ActionableMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	tests := []struct {
		name              string
		args              []string
		expectedExitCode  int
		shouldContainHint string
	}{
		{
			name:              "missing app context provides hint",
			args:              []string{"create", "workspace", "dev"},
			expectedExitCode:  0, // Missing context returns 0 with hint
			shouldContainHint: "hint",
		},
		{
			name:              "missing required flags provides hint",
			args:              []string{"create", "app", "myapp", "--domain", "test-domain"},
			expectedExitCode:  0, // Returns 0 with actionable message
			shouldContainHint: "must specify",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr := f.AssertExitCodeWithOutput(t, tt.expectedExitCode, tt.args...)
			output := strings.ToLower(stdout + stderr)

			if !strings.Contains(output, tt.shouldContainHint) {
				t.Errorf("Error message should provide actionable hint (looking for '%s')\nOutput: %s",
					tt.shouldContainHint, output)
			}
		})
	}
}

// TestErrorMessage_ConsistentTerminology verifies error messages use consistent terminology.
func TestErrorMessage_ConsistentTerminology(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create resources for consistent error testing
	f.AssertCommandSuccess(t, "create", "ecosystem", "test-eco")
	f.AssertCommandSuccess(t, "create", "domain", "test-domain", "--ecosystem", "test-eco")

	tests := []struct {
		name         string
		args         []string
		expectedTerm string // Term that should appear in error
	}{
		{
			name:         "ecosystem not found",
			args:         []string{"get", "ecosystem", "nonexistent"},
			expectedTerm: "not found",
		},
		{
			name:         "domain not found",
			args:         []string{"get", "domain", "nonexistent", "--ecosystem", "test-eco"},
			expectedTerm: "not found",
		},
		{
			name:         "workspace not found in ecosystem",
			args:         []string{"get", "workspace", "nonexistent", "--ecosystem", "test-eco"},
			expectedTerm: "no workspace found", // Specific message for workspace
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exitCode, stdout, stderr := f.RunDVMWithExitCode(tt.args...)

			// Should fail
			if exitCode == 0 {
				t.Error("Expected command to fail")
				return
			}

			output := strings.ToLower(stdout + stderr)
			if !strings.Contains(output, tt.expectedTerm) {
				t.Errorf("Error message should use term '%s'\nOutput: %s",
					tt.expectedTerm, output)
			}
		})
	}
}

// TestHelpText_AliasesDocumented verifies that command aliases are properly documented.
func TestHelpText_AliasesDocumented(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	tests := []struct {
		name          string
		args          []string
		expectedAlias string
	}{
		{
			name:          "workspace alias 'ws' documented",
			args:          []string{"create", "workspace", "--help"},
			expectedAlias: "ws",
		},
		{
			name:          "app alias 'a' documented",
			args:          []string{"create", "app", "--help"},
			expectedAlias: "a",
		},
		{
			name:          "workspace list alias documented",
			args:          []string{"get", "workspaces", "--help"},
			expectedAlias: "ws",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _ := f.AssertExitCodeWithOutput(t, 0, tt.args...)

			if !strings.Contains(stdout, tt.expectedAlias) {
				t.Errorf("Help text should document alias '%s'\nOutput: %s",
					tt.expectedAlias, stdout)
			}
		})
	}
}
