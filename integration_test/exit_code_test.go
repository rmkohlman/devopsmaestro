// Package integration provides exit code verification tests.
// These tests ensure dvm follows CLI conventions:
// - Exit code 0: Success
// - Exit code 1: Error (not found, validation, etc.)
package integration

import (
	"strings"
	"testing"
)

// TestExitCode_SuccessfulOperations verifies that successful operations
// return exit code 0.
func TestExitCode_SuccessfulOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Run tests sequentially with shared state
	t.Run("get ecosystems (empty list)", func(t *testing.T) {
		f.AssertExitCode(t, 0, "get", "ecosystems")
	})

	t.Run("create ecosystem", func(t *testing.T) {
		f.AssertExitCode(t, 0, "create", "ecosystem", "test-eco")
	})

	t.Run("get ecosystem (exists)", func(t *testing.T) {
		f.AssertExitCode(t, 0, "get", "ecosystem", "test-eco")
	})

	t.Run("get ecosystems (with results)", func(t *testing.T) {
		f.AssertExitCode(t, 0, "get", "ecosystems")
	})

	t.Run("create domain", func(t *testing.T) {
		f.AssertExitCode(t, 0, "create", "domain", "test-domain", "--ecosystem", "test-eco")
	})

	t.Run("get domain", func(t *testing.T) {
		f.AssertExitCode(t, 0, "get", "domain", "test-domain")
	})

	t.Run("create app", func(t *testing.T) {
		f.AssertExitCode(t, 0, "create", "app", "test-app", "--domain", "test-domain", "--from-cwd")
	})

	t.Run("get app", func(t *testing.T) {
		f.AssertExitCode(t, 0, "get", "app", "test-app")
	})

	t.Run("delete app", func(t *testing.T) {
		f.AssertExitCode(t, 0, "delete", "app", "test-app", "--domain", "test-domain")
	})

	t.Run("delete domain", func(t *testing.T) {
		f.AssertExitCode(t, 0, "delete", "domain", "test-domain", "--ecosystem", "test-eco")
	})

	t.Run("delete ecosystem", func(t *testing.T) {
		f.AssertExitCode(t, 0, "delete", "ecosystem", "test-eco")
	})

	t.Run("version command", func(t *testing.T) {
		f.AssertExitCode(t, 0, "version")
	})

	t.Run("help command", func(t *testing.T) {
		f.AssertExitCode(t, 0, "help")
	})
}

// TestExitCode_NotFoundErrors verifies that "not found" errors return exit code 1.
func TestExitCode_NotFoundErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create ecosystem and domain for context
	f.AssertCommandSuccess(t, "create", "ecosystem", "test-eco")
	f.AssertCommandSuccess(t, "create", "domain", "test-domain", "--ecosystem", "test-eco")
	f.AssertCommandSuccess(t, "create", "app", "test-app", "--domain", "test-domain", "--from-cwd")

	tests := []struct {
		name           string
		args           []string
		stderrContains string
	}{
		{
			name:           "get nonexistent ecosystem",
			args:           []string{"get", "ecosystem", "nonexistent"},
			stderrContains: "not found",
		},
		{
			name:           "get nonexistent domain",
			args:           []string{"get", "domain", "nonexistent", "--ecosystem", "test-eco"},
			stderrContains: "not found",
		},
		{
			name:           "get nonexistent workspace",
			args:           []string{"get", "workspace", "nonexistent", "--app", "test-app"},
			stderrContains: "no workspace found",
		},
		{
			name:           "delete nonexistent ecosystem",
			args:           []string{"delete", "ecosystem", "nonexistent"},
			stderrContains: "not found",
		},
		{
			name:           "delete nonexistent domain",
			args:           []string{"delete", "domain", "nonexistent", "--ecosystem", "test-eco"},
			stderrContains: "not found",
		},
		{
			name:           "delete nonexistent app",
			args:           []string{"delete", "app", "nonexistent", "--domain", "test-domain"},
			stderrContains: "not found",
		},
		{
			name:           "delete nonexistent workspace",
			args:           []string{"delete", "workspace", "nonexistent", "--app", "test-app"},
			stderrContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, stderr := f.AssertExitCodeWithOutput(t, 1, tt.args...)
			if !strings.Contains(strings.ToLower(stderr), tt.stderrContains) {
				t.Errorf("Expected stderr to contain %q, got: %s", tt.stderrContains, stderr)
			}
		})
	}
}

// TestExitCode_ValidationErrors verifies that validation errors return exit code 1.
func TestExitCode_ValidationErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create ecosystem for some tests
	f.AssertCommandSuccess(t, "create", "ecosystem", "test-eco")

	tests := []struct {
		name           string
		args           []string
		stderrContains string
	}{
		{
			name:           "create ecosystem with empty name",
			args:           []string{"create", "ecosystem", ""},
			stderrContains: "cannot be empty",
		},
		{
			name:           "create domain with nonexistent ecosystem",
			args:           []string{"create", "domain", "test-domain", "--ecosystem", "nonexistent"},
			stderrContains: "not found",
		},
		{
			name:           "create app with nonexistent domain",
			args:           []string{"create", "app", "test-app", "--domain", "nonexistent", "--from-cwd"},
			stderrContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, stderr := f.AssertExitCodeWithOutput(t, 1, tt.args...)
			if !strings.Contains(strings.ToLower(stderr), tt.stderrContains) {
				t.Errorf("Expected stderr to contain %q, got: %s", tt.stderrContains, stderr)
			}
		})
	}
}

// TestExitCode_DuplicateResourceErrors verifies that creating duplicate resources
// returns exit code 1.
func TestExitCode_DuplicateResourceErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create initial resources
	f.AssertCommandSuccess(t, "create", "ecosystem", "dup-test-eco")

	tests := []struct {
		name           string
		args           []string
		stderrContains string
	}{
		{
			name:           "create duplicate ecosystem",
			args:           []string{"create", "ecosystem", "dup-test-eco"},
			stderrContains: "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, stderr := f.AssertExitCodeWithOutput(t, 1, tt.args...)
			if !strings.Contains(strings.ToLower(stderr), tt.stderrContains) {
				t.Errorf("Expected stderr to contain %q, got: %s", tt.stderrContains, stderr)
			}
		})
	}
}

// TestExitCode_MissingContextHints verifies that missing context returns exit 0 with hints.
// The CLI uses exit 0 for "missing context" errors to distinguish them from actual errors.
func TestExitCode_MissingContextHints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	tests := []struct {
		name           string
		args           []string
		stdoutContains string
	}{
		{
			name:           "create domain without ecosystem flag",
			args:           []string{"create", "domain", "test-domain"},
			stdoutContains: "no ecosystem specified",
		},
		{
			name:           "create app without domain flag",
			args:           []string{"create", "app", "test-app", "--from-cwd"},
			stdoutContains: "no domain specified",
		},
		{
			name:           "get domain without ecosystem flag",
			args:           []string{"get", "domain", "test-domain"},
			stdoutContains: "no ecosystem specified",
		},
		{
			name:           "get app without domain flag",
			args:           []string{"get", "app", "test-app"},
			stdoutContains: "no domain specified",
		},
		{
			name:           "delete domain without ecosystem flag",
			args:           []string{"delete", "domain", "test-domain"},
			stdoutContains: "no ecosystem specified",
		},
		{
			name:           "delete app without domain flag",
			args:           []string{"delete", "app", "test-app"},
			stdoutContains: "no domain specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _ := f.AssertExitCodeWithOutput(t, 0, tt.args...)
			if !strings.Contains(strings.ToLower(stdout), tt.stdoutContains) {
				t.Errorf("Expected stdout to contain %q, got: %s", tt.stdoutContains, stdout)
			}
		})
	}
}

// TestExitCode_InvalidFlagCombinations verifies that invalid flag combinations
// return exit code 1.
func TestExitCode_InvalidFlagCombinations(t *testing.T) {
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
			name: "unknown flag",
			args: []string{"get", "ecosystems", "--unknown-flag"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exitCode, _, _ := f.RunDVMWithExitCode(tt.args...)
			if exitCode == 0 {
				t.Errorf("Expected non-zero exit code for invalid input, got 0")
			}
		})
	}
}

// TestExitCode_InvalidOutputFormat verifies that invalid output formats
// fall back to table format (returns exit 0, not an error).
func TestExitCode_InvalidOutputFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Create ecosystem for test
	f.AssertCommandSuccess(t, "create", "ecosystem", "test-eco")

	// Invalid output format falls back to table format (exit 0)
	exitCode, stdout, _ := f.RunDVMWithExitCode("get", "ecosystems", "-o", "invalid")
	if exitCode != 0 {
		t.Errorf("Invalid output format should return exit 0 (fallback to table), got %d", exitCode)
	}
	// Should display table output with the ecosystem
	if !strings.Contains(stdout, "test-eco") {
		t.Errorf("Expected stdout to contain 'test-eco' (table fallback), got: %s", stdout)
	}
}

// TestExitCode_HelpAlwaysSucceeds verifies that help commands always return exit code 0.
func TestExitCode_HelpAlwaysSucceeds(t *testing.T) {
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
			name: "help",
			args: []string{"help"},
		},
		{
			name: "--help flag",
			args: []string{"--help"},
		},
		{
			name: "-h flag",
			args: []string{"-h"},
		},
		{
			name: "create --help",
			args: []string{"create", "--help"},
		},
		{
			name: "get --help",
			args: []string{"get", "--help"},
		},
		{
			name: "delete --help",
			args: []string{"delete", "--help"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f.AssertExitCode(t, 0, tt.args...)
		})
	}
}

// TestExitCode_OutputFormats verifies that different output formats succeed.
func TestExitCode_OutputFormats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create a test ecosystem
	f.AssertCommandSuccess(t, "create", "ecosystem", "format-test")

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "json output",
			args: []string{"get", "ecosystem", "format-test", "-o", "json"},
		},
		{
			name: "yaml output",
			args: []string{"get", "ecosystem", "format-test", "-o", "yaml"},
		},
		{
			name: "table output (default)",
			args: []string{"get", "ecosystem", "format-test"},
		},
		{
			name: "list json output",
			args: []string{"get", "ecosystems", "-o", "json"},
		},
		{
			name: "list yaml output",
			args: []string{"get", "ecosystems", "-o", "yaml"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f.AssertExitCode(t, 0, tt.args...)
		})
	}
}
