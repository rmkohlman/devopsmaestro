// Package integration provides comprehensive CLI flag testing for DevOpsMaestro.
// These tests verify that all CLI flags work correctly across different commands
// and output formats, following kubectl conventions.
package integration

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestOutputFormats_AllFormats tests all supported output formats (-o flag).
// DevOpsMaestro supports: json, yaml, plain, table, colored (default)
// Note: 'wide' format is documented but currently falls back to table format
func TestOutputFormats_AllFormats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create test data
	f.AssertCommandSuccess(t, "create", "ecosystem", "format-test")

	tests := []struct {
		name           string
		format         string
		validateFunc   func(t *testing.T, stdout string, format string)
		shouldFallback bool // whether this format is expected to fallback to table
	}{
		{
			name:   "json format",
			format: "json",
			validateFunc: func(t *testing.T, stdout string, format string) {
				var result map[string]interface{}
				err := json.Unmarshal([]byte(stdout), &result)
				require.NoError(t, err, "JSON output should be valid JSON")
				// Verify it contains the ecosystem name
				if metadata, ok := result["Metadata"].(map[string]interface{}); ok {
					assert.Equal(t, "format-test", metadata["Name"])
				} else if name, ok := result["name"].(string); ok {
					assert.Equal(t, "format-test", name)
				} else {
					t.Fatalf("Could not find name in JSON output: %+v", result)
				}
			},
		},
		{
			name:   "yaml format",
			format: "yaml",
			validateFunc: func(t *testing.T, stdout string, format string) {
				var result map[string]interface{}
				err := yaml.Unmarshal([]byte(stdout), &result)
				require.NoError(t, err, "YAML output should be valid YAML")
				// Verify it contains the ecosystem name
				// YAML uses lowercase 'metadata' not 'Metadata'
				if metadata, ok := result["metadata"].(map[string]interface{}); ok {
					assert.Equal(t, "format-test", metadata["name"])
				} else if name, ok := result["name"].(string); ok {
					assert.Equal(t, "format-test", name)
				} else {
					t.Fatalf("Could not find name in YAML output: %+v", result)
				}
			},
		},
		{
			name:   "table format (explicit)",
			format: "table",
			validateFunc: func(t *testing.T, stdout string, format string) {
				// Table format should contain column headers and data
				assert.Contains(t, stdout, "format-test", "Table should contain ecosystem name")
				// Should have some structure (spaces/alignment)
				assert.True(t, len(strings.Split(stdout, "\n")) >= 2, "Table should have multiple lines")
			},
		},
		{
			name:   "plain format",
			format: "plain",
			validateFunc: func(t *testing.T, stdout string, format string) {
				// Plain format should be readable text without ANSI color codes
				assert.Contains(t, stdout, "format-test", "Plain output should contain ecosystem name")
				// Should not contain ANSI escape sequences
				assert.NotContains(t, stdout, "\033[", "Plain output should not have ANSI codes")
			},
		},
		{
			name:   "colored format (explicit)",
			format: "colored",
			validateFunc: func(t *testing.T, stdout string, format string) {
				// Colored format should contain the data
				assert.Contains(t, stdout, "format-test", "Colored output should contain ecosystem name")
			},
		},
		{
			name:           "wide format (documented, falls back to table)",
			format:         "wide",
			shouldFallback: true,
			validateFunc: func(t *testing.T, stdout string, format string) {
				// Wide format is documented but not fully implemented - falls back to table
				assert.Contains(t, stdout, "format-test", "Wide output should contain ecosystem name")
				// Should look like table format
				assert.True(t, len(strings.Split(stdout, "\n")) >= 2, "Wide output should have multiple lines")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := f.RunDVM("get", "ecosystem", "format-test", "-o", tt.format)
			if err != nil {
				t.Fatalf("Command failed: %v\nStderr: %s", err, stderr)
			}
			tt.validateFunc(t, stdout, tt.format)
		})
	}
}

// TestOutputFormats_InvalidFormat tests that invalid output formats fallback gracefully.
// Per exit_code_test.go:310-330, invalid formats return exit 0 with fallback to table.
func TestOutputFormats_InvalidFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create test data
	f.AssertCommandSuccess(t, "create", "ecosystem", "invalid-format-test")

	// Invalid output format should fallback to table format (exit 0, not error)
	exitCode, stdout, _ := f.RunDVMWithExitCode("get", "ecosystems", "-o", "invalid-format-xyz")

	// Should succeed (fallback behavior)
	assert.Equal(t, 0, exitCode, "Invalid format should fallback gracefully with exit 0")

	// Should display table output with the ecosystem
	assert.Contains(t, stdout, "invalid-format-test", "Fallback should show table output")
}

// TestOutputFormats_ListCommands tests output formats with list commands.
// List commands return arrays in JSON/YAML format.
func TestOutputFormats_ListCommands(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create test hierarchy
	f.AssertCommandSuccess(t, "create", "ecosystem", "list-eco-1")
	f.AssertCommandSuccess(t, "create", "ecosystem", "list-eco-2")
	f.AssertCommandSuccess(t, "create", "domain", "list-domain-1", "--ecosystem", "list-eco-1")
	f.AssertCommandSuccess(t, "create", "domain", "list-domain-2", "--ecosystem", "list-eco-1")

	tests := []struct {
		name    string
		command []string
		format  string
		verify  func(t *testing.T, stdout string)
	}{
		{
			name:    "ecosystems json",
			command: []string{"get", "ecosystems"},
			format:  "json",
			verify: func(t *testing.T, stdout string) {
				var result []map[string]interface{}
				err := json.Unmarshal([]byte(stdout), &result)
				require.NoError(t, err, "List JSON should be valid")
				assert.GreaterOrEqual(t, len(result), 2, "Should have at least 2 ecosystems")
			},
		},
		{
			name:    "ecosystems yaml",
			command: []string{"get", "ecosystems"},
			format:  "yaml",
			verify: func(t *testing.T, stdout string) {
				var result []map[string]interface{}
				err := yaml.Unmarshal([]byte(stdout), &result)
				require.NoError(t, err, "List YAML should be valid")
				assert.GreaterOrEqual(t, len(result), 2, "Should have at least 2 ecosystems")
			},
		},
		{
			name:    "domains json with filter",
			command: []string{"get", "domains", "--ecosystem", "list-eco-1"},
			format:  "json",
			verify: func(t *testing.T, stdout string) {
				var result []map[string]interface{}
				err := json.Unmarshal([]byte(stdout), &result)
				require.NoError(t, err, "Filtered list JSON should be valid")
				assert.Equal(t, 2, len(result), "Should have exactly 2 domains in ecosystem")
			},
		},
		{
			name:    "domains yaml with filter",
			command: []string{"get", "domains", "--ecosystem", "list-eco-1"},
			format:  "yaml",
			verify: func(t *testing.T, stdout string) {
				var result []map[string]interface{}
				err := yaml.Unmarshal([]byte(stdout), &result)
				require.NoError(t, err, "Filtered list YAML should be valid")
				assert.Equal(t, 2, len(result), "Should have exactly 2 domains in ecosystem")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := append(tt.command, "-o", tt.format)
			stdout, stderr, err := f.RunDVM(args...)
			if err != nil {
				t.Fatalf("Command failed: %v\nStderr: %s", err, stderr)
			}
			tt.verify(t, stdout)
		})
	}
}

// TestFlagCombinations_AllFlagWithOutputFormats tests -A flag combined with -o formats.
// The -A (--all) flag lists resources across all parent contexts.
func TestFlagCombinations_AllFlagWithOutputFormats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create cross-hierarchy test data
	// Ecosystem 1 → Domain 1 → App 1
	f.AssertCommandSuccess(t, "create", "ecosystem", "all-eco-1")
	f.AssertCommandSuccess(t, "create", "domain", "all-domain-1", "--ecosystem", "all-eco-1")
	f.AssertCommandSuccess(t, "create", "app", "all-app-1", "--domain", "all-domain-1", "--from-cwd")

	// Ecosystem 2 → Domain 2 → App 2
	f.AssertCommandSuccess(t, "create", "ecosystem", "all-eco-2")
	f.AssertCommandSuccess(t, "create", "domain", "all-domain-2", "--ecosystem", "all-eco-2")
	f.AssertCommandSuccess(t, "create", "app", "all-app-2", "--domain", "all-domain-2", "--from-cwd")

	tests := []struct {
		name     string
		command  []string
		format   string
		minCount int // minimum expected items
	}{
		{
			name:     "get domains -A (all ecosystems) with json",
			command:  []string{"get", "domains", "-A"},
			format:   "json",
			minCount: 2,
		},
		{
			name:     "get domains -A (all ecosystems) with yaml",
			command:  []string{"get", "domains", "-A"},
			format:   "yaml",
			minCount: 2,
		},
		{
			name:     "get apps -A (all domains) with json",
			command:  []string{"get", "apps", "-A"},
			format:   "json",
			minCount: 2,
		},
		{
			name:     "get apps -A (all domains) with yaml",
			command:  []string{"get", "apps", "-A"},
			format:   "yaml",
			minCount: 2,
		},
		{
			name:     "get domains -A with table (default)",
			command:  []string{"get", "domains", "-A"},
			format:   "table",
			minCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := append(tt.command, "-o", tt.format)
			stdout, stderr, err := f.RunDVM(args...)
			if err != nil {
				t.Fatalf("Command failed: %v\nStderr: %s", err, stderr)
			}

			// Verify based on format
			if tt.format == "json" {
				var result []map[string]interface{}
				err := json.Unmarshal([]byte(stdout), &result)
				require.NoError(t, err, "JSON output should be valid")
				assert.GreaterOrEqual(t, len(result), tt.minCount,
					"Should have at least %d items across all contexts", tt.minCount)
			} else if tt.format == "yaml" {
				var result []map[string]interface{}
				err := yaml.Unmarshal([]byte(stdout), &result)
				require.NoError(t, err, "YAML output should be valid")
				assert.GreaterOrEqual(t, len(result), tt.minCount,
					"Should have at least %d items across all contexts", tt.minCount)
			} else {
				// Table/colored format - check content appears
				lines := strings.Split(strings.TrimSpace(stdout), "\n")
				// Account for header row in table format
				assert.GreaterOrEqual(t, len(lines), tt.minCount+1,
					"Table should have at least %d data rows plus header", tt.minCount)
			}
		})
	}
}

// TestFlagCombinations_FilterFlags tests filter flags (--ecosystem, --domain, --app).
func TestFlagCombinations_FilterFlags(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create filtered test hierarchy
	f.AssertCommandSuccess(t, "create", "ecosystem", "filter-eco")
	f.AssertCommandSuccess(t, "create", "domain", "filter-domain-1", "--ecosystem", "filter-eco")
	f.AssertCommandSuccess(t, "create", "domain", "filter-domain-2", "--ecosystem", "filter-eco")
	f.AssertCommandSuccess(t, "create", "app", "filter-app-1", "--domain", "filter-domain-1", "--from-cwd")
	f.AssertCommandSuccess(t, "create", "app", "filter-app-2", "--domain", "filter-domain-2", "--from-cwd")

	tests := []struct {
		name         string
		args         []string
		expectedName string
	}{
		{
			name:         "get domains filtered by ecosystem",
			args:         []string{"get", "domains", "--ecosystem", "filter-eco", "-o", "json"},
			expectedName: "filter-domain-1",
		},
		{
			name:         "get apps filtered by domain",
			args:         []string{"get", "apps", "--domain", "filter-domain-1", "-o", "json"},
			expectedName: "filter-app-1",
		},
		{
			name:         "get app filtered by domain only",
			args:         []string{"get", "app", "filter-app-2", "--domain", "filter-domain-2", "-o", "json"},
			expectedName: "filter-app-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := f.RunDVM(tt.args...)
			if err != nil {
				t.Fatalf("Command failed: %v\nStderr: %s", err, stderr)
			}

			// Verify the expected name appears in JSON output
			assert.Contains(t, stdout, tt.expectedName,
				"Filtered output should contain %s", tt.expectedName)
		})
	}
}

// TestFlagCombinations_ConflictingFlags tests behavior with conflicting flag combinations.
func TestFlagCombinations_ConflictingFlags(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create test data
	f.AssertCommandSuccess(t, "create", "ecosystem", "conflict-eco-1")
	f.AssertCommandSuccess(t, "create", "ecosystem", "conflict-eco-2")
	f.AssertCommandSuccess(t, "create", "domain", "conflict-domain", "--ecosystem", "conflict-eco-1")

	// Test: -A (all) with specific ecosystem filter
	// Behavior: specific filter should take precedence over -A
	t.Run("all flag with specific ecosystem filter", func(t *testing.T) {
		stdout, stderr, err := f.RunDVM("get", "domains", "-A", "--ecosystem", "conflict-eco-1", "-o", "json")
		if err != nil {
			t.Fatalf("Command failed: %v\nStderr: %s", err, stderr)
		}

		// Should return domains from conflict-eco-1 only
		var result []map[string]interface{}
		err = json.Unmarshal([]byte(stdout), &result)
		require.NoError(t, err, "JSON output should be valid")

		// Verify filter worked (should have exactly 1 domain from conflict-eco-1)
		assert.Equal(t, 1, len(result), "Specific filter should override -A flag")
		assert.Contains(t, stdout, "conflict-domain")
	})
}

// TestFlagValidation_UnknownFlags tests that unknown flags are rejected.
func TestFlagValidation_UnknownFlags(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "unknown flag on get ecosystems",
			args: []string{"get", "ecosystems", "--unknown-flag-xyz"},
		},
		{
			name: "unknown flag on get domains",
			args: []string{"get", "domains", "-A", "--invalid-option"},
		},
		{
			name: "typo in output flag",
			args: []string{"get", "ecosystems", "--outpt", "json"}, // typo: outpt vs output
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exitCode, _, stderr := f.RunDVMWithExitCode(tt.args...)

			// Unknown flags should cause non-zero exit
			assert.NotEqual(t, 0, exitCode, "Unknown flag should fail with non-zero exit code")

			// Stderr should contain helpful error message
			assert.NotEmpty(t, stderr, "Should provide error message for unknown flag")
		})
	}
}

// TestOutputFormat_EmptyResults tests output formats with empty result sets.
func TestOutputFormat_EmptyResults(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// No setup - test with empty database

	tests := []struct {
		name   string
		args   []string
		format string
	}{
		{
			name:   "empty ecosystems json",
			args:   []string{"get", "ecosystems"},
			format: "json",
		},
		{
			name:   "empty ecosystems yaml",
			args:   []string{"get", "ecosystems"},
			format: "yaml",
		},
		{
			name:   "empty ecosystems table",
			args:   []string{"get", "ecosystems"},
			format: "table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := append(tt.args, "-o", tt.format)
			stdout, stderr, err := f.RunDVM(args...)

			// Empty results should succeed (exit 0)
			require.NoError(t, err, "Empty results should not error")

			// Verify format-specific empty handling
			if tt.format == "json" {
				// Should be valid JSON (either [] or {} depending on implementation)
				stdout = strings.TrimSpace(stdout)
				// Handle various empty patterns
				if stdout == "{}" || stdout == "[]" {
					// Valid empty JSON
					return
				}
				// Or may have empty message - check it's not an error
				assert.NotContains(t, strings.ToLower(stderr), "error")
			} else if tt.format == "yaml" {
				// Should be valid YAML (could be empty or have message)
				assert.NotContains(t, strings.ToLower(stderr), "error")
			} else {
				// Table format likely shows "No resources found" message
				// This is acceptable for empty results
				assert.NotContains(t, strings.ToLower(stderr), "error")
			}
		})
	}
}

// TestOutputFormat_DefaultBehavior tests that default output format is used when -o is omitted.
func TestOutputFormat_DefaultBehavior(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create test data
	f.AssertCommandSuccess(t, "create", "ecosystem", "default-test")

	// Test without -o flag - should use default (colored/table)
	stdout, stderr, err := f.RunDVM("get", "ecosystems")
	require.NoError(t, err, "Default format should succeed")

	// Should display human-readable output (not JSON/YAML)
	assert.Contains(t, stdout, "default-test", "Default output should contain ecosystem name")

	// Should not be JSON
	assert.False(t, strings.HasPrefix(strings.TrimSpace(stdout), "{") || strings.HasPrefix(strings.TrimSpace(stdout), "["),
		"Default output should not be JSON")

	// Should not be YAML
	assert.False(t, strings.Contains(stdout, "apiVersion:") || strings.Contains(stdout, "kind:"),
		"Default output should not be YAML")

	// Should not have errors
	assert.Empty(t, stderr, "Default output should not have errors")
}

// TestFlagCombinations_MultipleOutputFormats tests behavior when multiple -o flags are provided.
func TestFlagCombinations_MultipleOutputFormats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create test data
	f.AssertCommandSuccess(t, "create", "ecosystem", "multi-format-test")

	// Test with multiple -o flags - last one should win
	stdout, _, err := f.RunDVM("get", "ecosystems", "-o", "yaml", "-o", "json")
	require.NoError(t, err, "Multiple -o flags should not error")

	// Last flag (json) should take precedence
	var result []map[string]interface{}
	err = json.Unmarshal([]byte(stdout), &result)
	assert.NoError(t, err, "Output should be JSON (last -o flag wins)")
	// Cobra silently accepts multiple flag values, last one wins - no warning expected
}

// TestFlagCombinations_WorkspaceAllFlag tests -A flag specifically for workspaces.
// Workspaces have the most complex hierarchy flags.
func TestFlagCombinations_WorkspaceAllFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create cross-hierarchy workspaces
	// Ecosystem 1 → Domain 1 → App 1 → Workspace 1
	f.AssertCommandSuccess(t, "create", "ecosystem", "ws-eco-1")
	f.AssertCommandSuccess(t, "create", "domain", "ws-domain-1", "--ecosystem", "ws-eco-1")
	f.AssertCommandSuccess(t, "create", "app", "ws-app-1", "--domain", "ws-domain-1", "--from-cwd")
	f.AssertCommandSuccess(t, "create", "workspace", "ws-1", "--app", "ws-app-1")

	// Ecosystem 2 → Domain 2 → App 2 → Workspace 2
	f.AssertCommandSuccess(t, "create", "ecosystem", "ws-eco-2")
	f.AssertCommandSuccess(t, "create", "domain", "ws-domain-2", "--ecosystem", "ws-eco-2")
	f.AssertCommandSuccess(t, "create", "app", "ws-app-2", "--domain", "ws-domain-2", "--from-cwd")
	f.AssertCommandSuccess(t, "create", "workspace", "ws-2", "--app", "ws-app-2")

	tests := []struct {
		name     string
		args     []string
		minCount int
	}{
		{
			name:     "get workspaces -A with json",
			args:     []string{"get", "workspaces", "-A", "-o", "json"},
			minCount: 2,
		},
		{
			name:     "get workspaces -A with yaml",
			args:     []string{"get", "workspaces", "-A", "-o", "yaml"},
			minCount: 2,
		},
		{
			name:     "get workspaces -A with table",
			args:     []string{"get", "workspaces", "-A", "-o", "table"},
			minCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := f.RunDVM(tt.args...)
			if err != nil {
				t.Fatalf("Command failed: %v\nStderr: %s", err, stderr)
			}

			// Verify we got workspaces from both hierarchies
			if strings.Contains(tt.args[len(tt.args)-1], "json") {
				var result []map[string]interface{}
				err := json.Unmarshal([]byte(stdout), &result)
				require.NoError(t, err, "JSON output should be valid")
				assert.GreaterOrEqual(t, len(result), tt.minCount,
					"Should have at least %d workspaces across all hierarchies", tt.minCount)
			} else if strings.Contains(tt.args[len(tt.args)-1], "yaml") {
				var result []map[string]interface{}
				err := yaml.Unmarshal([]byte(stdout), &result)
				require.NoError(t, err, "YAML output should be valid")
				assert.GreaterOrEqual(t, len(result), tt.minCount,
					"Should have at least %d workspaces across all hierarchies", tt.minCount)
			} else {
				// Table format
				assert.Contains(t, stdout, "ws-1")
				assert.Contains(t, stdout, "ws-2")
			}
		})
	}
}

// TestFlagCombinations_WorkspaceHierarchyFilters tests workspace-specific hierarchy flags.
func TestFlagCombinations_WorkspaceHierarchyFilters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create nested hierarchy
	f.AssertCommandSuccess(t, "create", "ecosystem", "filter-ws-eco")
	f.AssertCommandSuccess(t, "create", "domain", "filter-ws-domain-1", "--ecosystem", "filter-ws-eco")
	f.AssertCommandSuccess(t, "create", "domain", "filter-ws-domain-2", "--ecosystem", "filter-ws-eco")
	f.AssertCommandSuccess(t, "create", "app", "filter-ws-app-1", "--domain", "filter-ws-domain-1", "--from-cwd")
	f.AssertCommandSuccess(t, "create", "app", "filter-ws-app-2", "--domain", "filter-ws-domain-2", "--from-cwd")
	f.AssertCommandSuccess(t, "create", "workspace", "filter-ws-1", "--app", "filter-ws-app-1")
	f.AssertCommandSuccess(t, "create", "workspace", "filter-ws-2", "--app", "filter-ws-app-2")

	tests := []struct {
		name          string
		args          []string
		shouldFind    []string
		shouldNotFind []string
	}{
		{
			name:          "filter by app",
			args:          []string{"get", "workspaces", "--app", "filter-ws-app-1", "-o", "json"},
			shouldFind:    []string{"filter-ws-1"},
			shouldNotFind: []string{"filter-ws-2"},
		},
		{
			name:          "filter by domain (finds all apps in domain)",
			args:          []string{"get", "workspaces", "--domain", "filter-ws-domain-1", "-o", "json"},
			shouldFind:    []string{"filter-ws-1"},
			shouldNotFind: []string{"filter-ws-2"},
		},
		{
			name:       "filter by ecosystem (finds all in hierarchy)",
			args:       []string{"get", "workspaces", "--ecosystem", "filter-ws-eco", "-o", "json"},
			shouldFind: []string{"filter-ws-1", "filter-ws-2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := f.RunDVM(tt.args...)
			if err != nil {
				t.Fatalf("Command failed: %v\nStderr: %s", err, stderr)
			}

			// Check what should be found
			for _, expected := range tt.shouldFind {
				assert.Contains(t, stdout, expected,
					"Output should contain workspace %s", expected)
			}

			// Check what should NOT be found
			for _, unexpected := range tt.shouldNotFind {
				assert.NotContains(t, stdout, unexpected,
					"Output should NOT contain workspace %s", unexpected)
			}
		})
	}
}

// TestOutputFormat_WideFormat tests the documented -o wide format behavior.
// According to documentation, 'wide' should show additional columns but currently falls back.
func TestOutputFormat_WideFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create test resources
	f.AssertCommandSuccess(t, "create", "ecosystem", "wide-test")
	f.AssertCommandSuccess(t, "create", "domain", "wide-domain", "--ecosystem", "wide-test")
	f.AssertCommandSuccess(t, "create", "app", "wide-app", "--domain", "wide-domain", "--from-cwd")
	f.AssertCommandSuccess(t, "create", "workspace", "wide-ws", "--app", "wide-app")

	tests := []struct {
		name    string
		command []string
	}{
		{
			name:    "ecosystems wide",
			command: []string{"get", "ecosystems", "-o", "wide"},
		},
		{
			name:    "domains wide",
			command: []string{"get", "domains", "-A", "-o", "wide"},
		},
		{
			name:    "apps wide",
			command: []string{"get", "apps", "-A", "-o", "wide"},
		},
		{
			name:    "workspaces wide",
			command: []string{"get", "workspaces", "-A", "-o", "wide"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := f.RunDVM(tt.command...)
			require.NoError(t, err, "Wide format should succeed (or fallback)")

			// Wide format should produce tabular output
			assert.NotEmpty(t, stdout, "Wide format should produce output")

			// Should not error
			if stderr != "" {
				// Only check if stderr contains actual errors (not just info messages)
				assert.NotContains(t, strings.ToLower(stderr), "error:",
					"Wide format should not produce errors")
			}

			// Should look like table format (has multiple lines, contains resource name)
			lines := strings.Split(strings.TrimSpace(stdout), "\n")
			assert.Greater(t, len(lines), 1, "Wide format should have multiple lines like table")
		})
	}
}
