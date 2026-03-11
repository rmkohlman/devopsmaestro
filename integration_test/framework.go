// Package integration provides a comprehensive system integration testing
// framework for DevOpsMaestro. It enables end-to-end testing of CLI commands,
// workflows, and state verification.
package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"devopsmaestro/pkg/paths"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFramework provides utilities for system integration tests.
// It creates an isolated test environment with its own database and
// compiled binary, ensuring tests don't interfere with each other
// or the user's production environment.
type TestFramework struct {
	// TempDir is the root temporary directory for this test
	TempDir string

	// DBPath is the path to the isolated test database
	DBPath string

	// BinaryPath is the path to the dvm executable
	BinaryPath string

	// HomeDir is the isolated home directory for this test
	HomeDir string

	// cleanup function to call after test completes
	cleanup func()

	// t is the testing context
	t *testing.T
}

// NewTestFramework creates an isolated test environment for integration testing.
// It compiles the dvm binary (if needed), creates a temporary database, and
// sets up necessary directories. The cleanup function MUST be called via defer.
//
// Example:
//
//	func TestMyWorkflow(t *testing.T) {
//	    f := NewTestFramework(t)
//	    defer f.Cleanup()
//
//	    f.AssertCommandSuccess(t, "create", "ecosystem", "test")
//	}
func NewTestFramework(t *testing.T) *TestFramework {
	t.Helper()

	// Create temporary directory for this test
	tempDir, err := os.MkdirTemp("", "dvm-integration-test-*")
	require.NoError(t, err, "Failed to create temp directory")

	// Create isolated home directory
	homeDir := filepath.Join(tempDir, "home")
	require.NoError(t, os.MkdirAll(homeDir, 0755), "Failed to create home directory")

	// Create .devopsmaestro directory structure
	dvmDir := filepath.Join(homeDir, paths.DVMDirName)
	require.NoError(t, os.MkdirAll(dvmDir, 0755), "Failed to create .devopsmaestro directory")

	// Database path
	dbPath := filepath.Join(dvmDir, paths.DatabaseFile)

	// Build the binary (in temp dir to avoid conflicts)
	binaryPath := filepath.Join(tempDir, "dvm")
	if err := buildBinary(binaryPath); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to build dvm binary: %v", err)
	}

	framework := &TestFramework{
		TempDir:    tempDir,
		DBPath:     dbPath,
		BinaryPath: binaryPath,
		HomeDir:    homeDir,
		t:          t,
	}

	// Setup cleanup function
	framework.cleanup = func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: Failed to cleanup temp directory %s: %v", tempDir, err)
		}
	}

	return framework
}

// Cleanup removes all temporary files and directories created by the test framework.
// This MUST be called via defer after creating the framework.
func (f *TestFramework) Cleanup() {
	if f.cleanup != nil {
		f.cleanup()
	}
}

// RunDVM executes a dvm command with the given arguments and returns
// stdout, stderr, and any error. The command runs in the isolated test
// environment with its own database and home directory.
//
// For workspace delete commands, --force is automatically added to skip
// confirmation prompts in the non-interactive test environment.
//
// Example:
//
//	stdout, stderr, err := f.RunDVM("get", "ecosystems", "-o", "json")
func (f *TestFramework) RunDVM(args ...string) (stdout, stderr string, err error) {
	f.t.Helper()

	// Auto-add --force to "delete workspace" commands to skip confirmation prompts
	// Only workspace delete has a confirmation prompt requiring --force
	if len(args) >= 2 && args[0] == "delete" && args[1] == "workspace" {
		hasForce := false
		for _, arg := range args {
			if arg == "--force" || arg == "-f" {
				hasForce = true
				break
			}
		}
		if !hasForce {
			args = append(args, "--force")
		}
	}

	cmd := exec.Command(f.BinaryPath, args...)

	// Set environment to use test database and home
	cmd.Env = []string{
		fmt.Sprintf("HOME=%s", f.HomeDir),
		fmt.Sprintf("DVM_DB_PATH=%s", f.DBPath),
		"PATH=" + os.Getenv("PATH"), // Preserve PATH for subcommands
	}

	var stdoutBuf, stderrBuf strings.Builder
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	return stdoutBuf.String(), stderrBuf.String(), err
}

// RunDVMJSON executes a dvm command with -o json and parses the result
// into a map. This is useful for verifying structured output.
// The output follows Kubernetes-style resource format with Metadata and Spec.
//
// Example:
//
//	result, err := f.RunDVMJSON("get", "ecosystem", "test")
//	assert.NoError(t, err)
//	assert.Equal(t, "test", result["Metadata"].(map[string]interface{})["Name"])
func (f *TestFramework) RunDVMJSON(args ...string) (map[string]interface{}, error) {
	f.t.Helper()

	// Add -o json flag if not already present
	hasOutputFlag := false
	for _, arg := range args {
		if arg == "-o" || arg == "--output" {
			hasOutputFlag = true
			break
		}
	}

	if !hasOutputFlag {
		args = append(args, "-o", "json")
	}

	stdout, stderr, err := f.RunDVM(args...)
	if err != nil {
		return nil, fmt.Errorf("command failed: %w\nstderr: %s", err, stderr)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w\nstdout: %s", err, stdout)
	}

	return result, nil
}

// GetResourceName extracts the name from a resource map.
// Supports both Kubernetes-style: {Metadata: {Name: "..."}}
// and flat JSON: {name: "..."}
func (f *TestFramework) GetResourceName(resource map[string]interface{}) string {
	f.t.Helper()

	// Try Kubernetes-style first: Metadata.Name
	if metadata, ok := resource["Metadata"].(map[string]interface{}); ok {
		if name, ok := metadata["Name"].(string); ok {
			return name
		}
	}
	// Try flat JSON: name
	if name, ok := resource["name"].(string); ok {
		return name
	}
	return ""
}

// GetResourceDescription extracts the description from a resource map.
// Supports Kubernetes-style: {Metadata: {Annotations: {description: "..."}}}
// and flat JSON: {description: "..."}
func (f *TestFramework) GetResourceDescription(resource map[string]interface{}) string {
	f.t.Helper()

	// Try Kubernetes-style first
	if metadata, ok := resource["Metadata"].(map[string]interface{}); ok {
		if annotations, ok := metadata["Annotations"].(map[string]interface{}); ok {
			if description, ok := annotations["description"].(string); ok {
				return description
			}
		}
	}
	// Try flat JSON
	if description, ok := resource["description"].(string); ok {
		return description
	}
	return ""
}

// GetResourceSpec extracts the Spec section from a resource map.
// For flat JSON resources, returns the resource itself (minus metadata fields).
func (f *TestFramework) GetResourceSpec(resource map[string]interface{}) map[string]interface{} {
	f.t.Helper()

	// Try Kubernetes-style first
	if spec, ok := resource["Spec"].(map[string]interface{}); ok {
		return spec
	}
	// For flat JSON, the resource itself contains the spec fields
	return resource
}

// GetResourceField extracts any field from a resource map (flat JSON style).
// Returns empty string if field doesn't exist or isn't a string.
func (f *TestFramework) GetResourceField(resource map[string]interface{}, field string) string {
	f.t.Helper()

	if value, ok := resource[field].(string); ok {
		return value
	}
	return ""
}

// RunDVMJSONList executes a dvm command that returns a JSON array and
// parses the result into a slice of maps. This is useful for list commands.
//
// Example:
//
//	ecosystems, err := f.RunDVMJSONList("get", "ecosystems")
//	assert.NoError(t, err)
//	assert.Len(t, ecosystems, 1)
func (f *TestFramework) RunDVMJSONList(args ...string) ([]map[string]interface{}, error) {
	f.t.Helper()

	// Add -o json flag if not already present
	hasOutputFlag := false
	for _, arg := range args {
		if arg == "-o" || arg == "--output" {
			hasOutputFlag = true
			break
		}
	}

	if !hasOutputFlag {
		args = append(args, "-o", "json")
	}

	stdout, stderr, err := f.RunDVM(args...)
	if err != nil {
		return nil, fmt.Errorf("command failed: %w\nstderr: %s", err, stderr)
	}

	// Trim whitespace and check for empty output patterns
	stdout = strings.TrimSpace(stdout)

	// Handle case where CLI returns "{}" for empty list (should be "[]")
	if stdout == "{}" {
		return []map[string]interface{}{}, nil
	}

	// Handle case where CLI returns text messages like "No git repositories found"
	// or "Applying database migrations..."
	if strings.HasPrefix(stdout, "No ") || strings.HasPrefix(stdout, "Applying ") {
		return []map[string]interface{}{}, nil
	}

	var result []map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		// If it's an object instead of array, check if it's empty
		var obj map[string]interface{}
		if objErr := json.Unmarshal([]byte(stdout), &obj); objErr == nil && len(obj) == 0 {
			return []map[string]interface{}{}, nil
		}
		return nil, fmt.Errorf("failed to parse JSON: %w\nstdout: %s", err, stdout)
	}

	return result, nil
}

// AssertCommandSuccess executes a command and asserts that it succeeds
// (exits with code 0). If the command fails, the test fails immediately
// with a descriptive error message including stderr output.
//
// Example:
//
//	f.AssertCommandSuccess(t, "create", "ecosystem", "test")
func (f *TestFramework) AssertCommandSuccess(t *testing.T, args ...string) {
	t.Helper()

	stdout, stderr, err := f.RunDVM(args...)
	if err != nil {
		t.Fatalf("Command failed: %v\nArgs: %v\nStdout: %s\nStderr: %s",
			err, args, stdout, stderr)
	}
}

// AssertCommandFails executes a command and asserts that it fails
// (exits with non-zero code). If the command succeeds, the test fails.
//
// Example:
//
//	f.AssertCommandFails(t, "create", "ecosystem", "") // empty name should fail
func (f *TestFramework) AssertCommandFails(t *testing.T, args ...string) {
	t.Helper()

	_, _, err := f.RunDVM(args...)
	if err == nil {
		t.Fatalf("Expected command to fail but it succeeded\nArgs: %v", args)
	}
}

// AssertOutput verifies that output contains all expected strings.
// This is useful for checking human-readable output.
//
// Example:
//
//	stdout, _, _ := f.RunDVM("get", "ecosystems")
//	f.AssertOutput(t, stdout, "test-ecosystem", "production")
func (f *TestFramework) AssertOutput(t *testing.T, output string, contains ...string) {
	t.Helper()

	for _, expected := range contains {
		assert.Contains(t, output, expected,
			"Output should contain %q\nFull output:\n%s", expected, output)
	}
}

// AssertOutputDoesNotContain verifies that output does NOT contain specified strings.
// This is useful for checking that deleted resources don't appear.
//
// Example:
//
//	stdout, _, _ := f.RunDVM("get", "ecosystems")
//	f.AssertOutputDoesNotContain(t, stdout, "deleted-ecosystem")
func (f *TestFramework) AssertOutputDoesNotContain(t *testing.T, output string, notContains ...string) {
	t.Helper()

	for _, unexpected := range notContains {
		assert.NotContains(t, output, unexpected,
			"Output should NOT contain %q\nFull output:\n%s", unexpected, output)
	}
}

// GetDatabasePath returns the path to the test database.
// This can be used for direct database inspection if needed.
func (f *TestFramework) GetDatabasePath() string {
	return f.DBPath
}

// RunDVMWithExitCode executes a dvm command and returns the exit code along with
// stdout and stderr. This is useful for testing that commands return proper exit codes.
//
// Example:
//
//	exitCode, stdout, stderr := f.RunDVMWithExitCode("get", "ecosystem", "nonexistent")
//	assert.Equal(t, 1, exitCode)
func (f *TestFramework) RunDVMWithExitCode(args ...string) (exitCode int, stdout, stderr string) {
	f.t.Helper()

	// Auto-add --force to "delete workspace" commands to skip confirmation prompts
	if len(args) >= 2 && args[0] == "delete" && args[1] == "workspace" {
		hasForce := false
		for _, arg := range args {
			if arg == "--force" || arg == "-f" {
				hasForce = true
				break
			}
		}
		if !hasForce {
			args = append(args, "--force")
		}
	}

	cmd := exec.Command(f.BinaryPath, args...)

	// Set environment to use test database and home
	cmd.Env = []string{
		fmt.Sprintf("HOME=%s", f.HomeDir),
		fmt.Sprintf("DVM_DB_PATH=%s", f.DBPath),
		"PATH=" + os.Getenv("PATH"),
	}

	var stdoutBuf, stderrBuf strings.Builder
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			// Non-exit error (e.g., binary not found)
			exitCode = -1
		}
	}

	return exitCode, stdoutBuf.String(), stderrBuf.String()
}

// AssertExitCode executes a command and asserts that it returns the expected exit code.
//
// Example:
//
//	f.AssertExitCode(t, 0, "get", "ecosystems")        // success
//	f.AssertExitCode(t, 1, "get", "workspace", "none") // error
func (f *TestFramework) AssertExitCode(t *testing.T, expectedCode int, args ...string) {
	t.Helper()

	exitCode, stdout, stderr := f.RunDVMWithExitCode(args...)
	if exitCode != expectedCode {
		t.Fatalf("Exit code mismatch\nExpected: %d\nActual: %d\nArgs: %v\nStdout: %s\nStderr: %s",
			expectedCode, exitCode, args, stdout, stderr)
	}
}

// AssertExitCodeWithOutput executes a command, asserts the exit code, and returns output.
// This is useful when you need to verify both exit code and output content.
//
// Example:
//
//	stdout, stderr := f.AssertExitCodeWithOutput(t, 1, "create", "ecosystem", "")
//	assert.Contains(t, stderr, "cannot be empty")
func (f *TestFramework) AssertExitCodeWithOutput(t *testing.T, expectedCode int, args ...string) (stdout, stderr string) {
	t.Helper()

	exitCode, stdout, stderr := f.RunDVMWithExitCode(args...)
	if exitCode != expectedCode {
		t.Fatalf("Exit code mismatch\nExpected: %d\nActual: %d\nArgs: %v\nStdout: %s\nStderr: %s",
			expectedCode, exitCode, args, stdout, stderr)
	}
	return stdout, stderr
}

// buildBinary compiles the dvm binary to the specified path.
// It uses go build with appropriate flags for testing.
func buildBinary(outputPath string) error {
	// Determine the repository root
	// When running tests, we're in integration_test/ directory
	// So we need to go up one level to find main.go
	repoRoot := "."
	if _, err := os.Stat("main.go"); err != nil {
		// Try parent directory
		repoRoot = ".."
		if _, err := os.Stat("../main.go"); err != nil {
			return fmt.Errorf("main.go not found - tests must be run from repository root or integration_test directory")
		}
	}

	cmd := exec.Command("go", "build", "-o", outputPath, ".")
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}

	return nil
}
