// Package integration provides integration tests for DevOpsMaestro's
// wide format output feature.
package integration

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWideFormat_Workspace tests wide format output for workspaces.
// This is the most important test as workspaces have the most dynamic state.
//
// Standard output should have: NAME, APP, STATUS, IMAGE
// Wide output (-o wide) should add: CREATED, CONTAINER-ID
//
// Expected to FAIL initially (TDD Red phase) - wide format not yet implemented.
func TestWideFormat_Workspace(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create ecosystem → domain → app → workspace
	setupTestHierarchy(t, f)
	f.AssertCommandSuccess(t, "create", "workspace", "test-ws",
		"--description", "Test workspace for wide format")

	// Get standard output
	standardOutput, _, err := f.RunDVM("get", "workspaces")
	require.NoError(t, err, "Standard get workspaces should succeed")

	// Get wide output
	wideOutput, _, err := f.RunDVM("get", "workspaces", "-o", "wide")
	require.NoError(t, err, "Wide get workspaces should succeed")

	// Parse table structure
	standardHeaders := getTableHeaders(standardOutput)
	wideHeaders := getTableHeaders(wideOutput)
	standardCols := countTableColumns(standardOutput)
	wideCols := countTableColumns(wideOutput)

	// Verify wide has more columns than standard
	assert.Greater(t, wideCols, standardCols,
		"Wide format should have more columns than standard\nStandard: %d columns\nWide: %d columns",
		standardCols, wideCols)

	// Verify left-most columns are the same (wide should be superset)
	for i, h := range standardHeaders {
		assert.Equal(t, h, wideHeaders[i],
			"Left columns should match (standard column %d: %q should equal wide column %d: %q)",
			i, h, i, wideHeaders[i])
	}

	// Verify standard output contains expected columns
	assert.Contains(t, strings.ToUpper(standardOutput), "NAME",
		"Standard output should have NAME column")
	assert.Contains(t, strings.ToUpper(standardOutput), "APP",
		"Standard output should have APP column")
	assert.Contains(t, strings.ToUpper(standardOutput), "STATUS",
		"Standard output should have STATUS column")

	// Verify wide output contains additional columns
	assert.Contains(t, strings.ToUpper(wideOutput), "CREATED",
		"Wide output should have CREATED column")

	// Verify both outputs contain the workspace name
	assert.Contains(t, standardOutput, "test-ws",
		"Standard output should contain workspace name")
	assert.Contains(t, wideOutput, "test-ws",
		"Wide output should contain workspace name")
}

// TestWideFormat_App tests wide format output for applications.
//
// Standard output should have: NAME, DOMAIN, WORKSPACES
// Wide output should add: CREATED, ID, PATH, GITREPO
//
// Expected to FAIL initially (TDD Red phase).
func TestWideFormat_App(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create ecosystem → domain → app
	f.AssertCommandSuccess(t, "create", "ecosystem", "test-eco")
	f.AssertCommandSuccess(t, "use", "ecosystem", "test-eco")
	f.AssertCommandSuccess(t, "create", "domain", "test-domain")
	f.AssertCommandSuccess(t, "use", "domain", "test-domain")
	f.AssertCommandSuccess(t, "create", "app", "test-app",
		"--description", "Test app for wide format",
		"--from-cwd")

	// Get standard output
	standardOutput, _, err := f.RunDVM("get", "apps")
	require.NoError(t, err, "Standard get apps should succeed")

	// Get wide output
	wideOutput, _, err := f.RunDVM("get", "apps", "-o", "wide")
	require.NoError(t, err, "Wide get apps should succeed")

	// Parse table structure
	standardHeaders := getTableHeaders(standardOutput)
	wideHeaders := getTableHeaders(wideOutput)
	standardCols := countTableColumns(standardOutput)
	wideCols := countTableColumns(wideOutput)

	// Verify wide has more columns than standard
	assert.Greater(t, wideCols, standardCols,
		"Wide format should have more columns than standard\nStandard: %d columns\nWide: %d columns",
		standardCols, wideCols)

	// Verify left-most columns match
	for i, h := range standardHeaders {
		assert.Equal(t, h, wideHeaders[i],
			"Left columns should match (standard column %d: %q should equal wide column %d: %q)",
			i, h, i, wideHeaders[i])
	}

	// Verify standard output contains expected columns
	assert.Contains(t, strings.ToUpper(standardOutput), "NAME",
		"Standard output should have NAME column")

	// Verify wide output contains additional columns
	assert.Contains(t, strings.ToUpper(wideOutput), "CREATED",
		"Wide output should have CREATED column")
	assert.Contains(t, strings.ToUpper(wideOutput), "PATH",
		"Wide output should have PATH column")

	// Verify both outputs contain the app name
	assert.Contains(t, standardOutput, "test-app",
		"Standard output should contain app name")
	assert.Contains(t, wideOutput, "test-app",
		"Wide output should contain app name")
}

// TestWideFormat_Domain tests wide format output for domains.
//
// Standard output should have: NAME, ECOSYSTEM, APPS
// Wide output should add: CREATED, ID, DESCRIPTION
//
// Expected to FAIL initially (TDD Red phase).
func TestWideFormat_Domain(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create ecosystem → domain
	f.AssertCommandSuccess(t, "create", "ecosystem", "test-eco")
	f.AssertCommandSuccess(t, "use", "ecosystem", "test-eco")
	f.AssertCommandSuccess(t, "create", "domain", "test-domain",
		"--description", "Test domain for wide format")

	// Get standard output
	standardOutput, _, err := f.RunDVM("get", "domains")
	require.NoError(t, err, "Standard get domains should succeed")

	// Get wide output
	wideOutput, _, err := f.RunDVM("get", "domains", "-o", "wide")
	require.NoError(t, err, "Wide get domains should succeed")

	// Parse table structure
	standardHeaders := getTableHeaders(standardOutput)
	wideHeaders := getTableHeaders(wideOutput)
	standardCols := countTableColumns(standardOutput)
	wideCols := countTableColumns(wideOutput)

	// Verify wide has more columns than standard
	assert.Greater(t, wideCols, standardCols,
		"Wide format should have more columns than standard\nStandard: %d columns\nWide: %d columns",
		standardCols, wideCols)

	// Verify left-most columns match
	for i, h := range standardHeaders {
		assert.Equal(t, h, wideHeaders[i],
			"Left columns should match (standard column %d: %q should equal wide column %d: %q)",
			i, h, i, wideHeaders[i])
	}

	// Verify standard output contains expected columns
	assert.Contains(t, strings.ToUpper(standardOutput), "NAME",
		"Standard output should have NAME column")

	// Verify wide output contains additional columns
	assert.Contains(t, strings.ToUpper(wideOutput), "CREATED",
		"Wide output should have CREATED column")

	// Verify both outputs contain the domain name
	assert.Contains(t, standardOutput, "test-domain",
		"Standard output should contain domain name")
	assert.Contains(t, wideOutput, "test-domain",
		"Wide output should contain domain name")
}

// TestWideFormat_Ecosystem tests wide format output for ecosystems.
//
// Standard output should have: NAME, DOMAINS, APPS
// Wide output should add: CREATED, ID, DESCRIPTION
//
// Expected to FAIL initially (TDD Red phase).
func TestWideFormat_Ecosystem(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create ecosystem
	f.AssertCommandSuccess(t, "create", "ecosystem", "test-eco",
		"--description", "Test ecosystem for wide format")

	// Get standard output
	standardOutput, _, err := f.RunDVM("get", "ecosystems")
	require.NoError(t, err, "Standard get ecosystems should succeed")

	// Get wide output
	wideOutput, _, err := f.RunDVM("get", "ecosystems", "-o", "wide")
	require.NoError(t, err, "Wide get ecosystems should succeed")

	// Parse table structure
	standardHeaders := getTableHeaders(standardOutput)
	wideHeaders := getTableHeaders(wideOutput)
	standardCols := countTableColumns(standardOutput)
	wideCols := countTableColumns(wideOutput)

	// Verify wide has more columns than standard
	assert.Greater(t, wideCols, standardCols,
		"Wide format should have more columns than standard\nStandard: %d columns\nWide: %d columns",
		standardCols, wideCols)

	// Verify left-most columns match
	for i, h := range standardHeaders {
		assert.Equal(t, h, wideHeaders[i],
			"Left columns should match (standard column %d: %q should equal wide column %d: %q)",
			i, h, i, wideHeaders[i])
	}

	// Verify standard output contains expected columns
	assert.Contains(t, strings.ToUpper(standardOutput), "NAME",
		"Standard output should have NAME column")

	// Verify wide output contains additional columns
	assert.Contains(t, strings.ToUpper(wideOutput), "CREATED",
		"Wide output should have CREATED column")

	// Verify both outputs contain the ecosystem name
	assert.Contains(t, standardOutput, "test-eco",
		"Standard output should contain ecosystem name")
	assert.Contains(t, wideOutput, "test-eco",
		"Wide output should contain ecosystem name")
}

// TestWideFormat_ExitCodeSuccess verifies that wide format commands exit with code 0.
//
// All wide format commands should succeed with exit code 0 when resources exist.
//
// Expected to FAIL initially if -o wide flag is not recognized.
func TestWideFormat_ExitCodeSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create full hierarchy
	setupTestHierarchy(t, f)
	f.AssertCommandSuccess(t, "create", "workspace", "test-ws")

	// Test exit codes for all resource types with -o wide
	tests := []struct {
		name string
		args []string
	}{
		{"ecosystems wide", []string{"get", "ecosystems", "-o", "wide"}},
		{"domains wide", []string{"get", "domains", "-o", "wide"}},
		{"apps wide", []string{"get", "apps", "-o", "wide"}},
		{"workspaces wide", []string{"get", "workspaces", "-o", "wide"}},
		{"ecosystem wide", []string{"get", "ecosystem", "test-eco", "-o", "wide"}},
		{"domain wide", []string{"get", "domain", "test-domain", "-o", "wide"}},
		{"app wide", []string{"get", "app", "test-app", "-o", "wide"}},
		{"workspace wide", []string{"get", "workspace", "test-ws", "-o", "wide"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f.AssertExitCode(t, 0, tt.args...)
		})
	}
}

// TestWideFormat_MultipleResources tests wide format with multiple resources.
//
// Verifies that wide format works correctly when multiple resources are displayed.
//
// Expected to FAIL initially (TDD Red phase).
func TestWideFormat_MultipleResources(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create multiple workspaces
	setupTestHierarchy(t, f)
	f.AssertCommandSuccess(t, "create", "workspace", "ws-1")
	f.AssertCommandSuccess(t, "create", "workspace", "ws-2")
	f.AssertCommandSuccess(t, "create", "workspace", "ws-3")

	// Get wide output
	wideOutput, _, err := f.RunDVM("get", "workspaces", "-o", "wide")
	require.NoError(t, err, "Wide get workspaces should succeed")

	// Verify all workspaces appear in output
	assert.Contains(t, wideOutput, "ws-1", "Wide output should contain ws-1")
	assert.Contains(t, wideOutput, "ws-2", "Wide output should contain ws-2")
	assert.Contains(t, wideOutput, "ws-3", "Wide output should contain ws-3")

	// Count data rows (should be 3 workspaces)
	lines := strings.Split(strings.TrimSpace(wideOutput), "\n")
	dataRows := 0
	for i, line := range lines {
		// Skip header and separator lines
		if i == 0 || strings.Contains(line, "---") || strings.Contains(line, "===") {
			continue
		}
		if strings.TrimSpace(line) != "" {
			dataRows++
		}
	}
	assert.GreaterOrEqual(t, dataRows, 3, "Should have at least 3 data rows for 3 workspaces")
}

// TestWideFormat_ComparisonWithJSON tests that wide format shows same data as JSON.
//
// The wide format should not hallucinate data - it should only show data that
// exists in the underlying resource (verifiable via JSON output).
//
// Expected to FAIL initially (TDD Red phase).
func TestWideFormat_ComparisonWithJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create workspace
	setupTestHierarchy(t, f)
	f.AssertCommandSuccess(t, "create", "workspace", "test-ws",
		"--description", "Test workspace")

	// Get JSON output to see actual data
	workspaces, err := f.RunDVMJSONList("get", "workspaces")
	require.NoError(t, err, "JSON get workspaces should succeed")
	require.Len(t, workspaces, 1, "Should have 1 workspace")

	workspace := workspaces[0]

	// Get wide output
	wideOutput, _, err := f.RunDVM("get", "workspaces", "-o", "wide")
	require.NoError(t, err, "Wide get workspaces should succeed")

	// Verify workspace name appears in wide output
	name := f.GetResourceName(workspace)
	assert.Contains(t, wideOutput, name,
		"Wide output should contain workspace name from JSON: %s", name)

	// Note: We can't verify exact field values without knowing the wide format
	// column layout, but we ensure the name is present as a basic sanity check.
	// Once wide format is implemented, we can enhance this test to verify
	// specific fields match between JSON and wide output.
}

// Helper Functions

// countTableColumns counts the number of columns in table output by analyzing
// the header row. Returns 0 if no table structure is detected.
func countTableColumns(output string) int {
	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		return 0
	}

	// Find the header line (first non-empty line)
	var headerLine string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "---") {
			headerLine = trimmed
			break
		}
	}

	if headerLine == "" {
		return 0
	}

	// Split by whitespace and count non-empty parts
	// This is a simple heuristic - may need adjustment based on actual table format
	parts := strings.Fields(headerLine)
	return len(parts)
}

// getTableHeaders extracts the header row from table output.
// Returns empty slice if no table structure is detected.
func getTableHeaders(output string) []string {
	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		return []string{}
	}

	// Find the header line (first non-empty line)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "---") {
			// Split by whitespace to get headers
			return strings.Fields(trimmed)
		}
	}

	return []string{}
}
