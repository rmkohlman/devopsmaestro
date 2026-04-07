package cmd

// =============================================================================
// TDD Phase 2 (RED): Build Hygiene — No testify in Production Code (Issue #70)
// =============================================================================
// Bug: cmd/executor.go imports "github.com/stretchr/testify/mock" in production
// code and defines a MockExecutor struct that is dead code. Test-only packages
// must never appear in non-test Go files.
//
// These tests WILL FAIL against the current code because executor.go contains
// the testify/mock import. That is the expected RED state.
//
// Expected fix: Remove the testify/mock import and MockExecutor definition from
// cmd/executor.go (move MockExecutor to a _test.go file if needed, or delete it
// entirely if it is unused).
// =============================================================================

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNoTestifyInProductionCode asserts that no non-test Go file in the cmd/
// package imports any testify package. This is a build-hygiene gate: testify
// is a test-only dependency and must never appear in production binaries.
//
// RED state: fails because cmd/executor.go imports "github.com/stretchr/testify/mock".
// GREEN state: passes after the import and MockExecutor dead code are removed.
func TestNoTestifyInProductionCode(t *testing.T) {
	// Collect all non-test .go files in the cmd/ directory (not subdirectories —
	// each sub-package is its own build unit and can be gated separately).
	entries, err := os.ReadDir(".")
	require.NoError(t, err, "should be able to read cmd/ directory")

	var violations []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Skip test files — testify imports are allowed there.
		if strings.HasSuffix(name, "_test.go") {
			continue
		}

		// Only inspect Go source files.
		if !strings.HasSuffix(name, ".go") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(".", name))
		require.NoErrorf(t, err, "should be able to read %s", name)

		src := string(content)

		// Any testify sub-package is a violation (mock, assert, require, suite…).
		if strings.Contains(src, `"github.com/stretchr/testify/`) {
			violations = append(violations, name)
		}
	}

	assert.Empty(t, violations,
		"production (non-test) Go files must not import testify packages — "+
			"found violations: %v", violations)
}

// TestNoMockStructInProductionCode asserts that no non-test Go file in cmd/
// defines a struct that embeds mock.Mock from testify. Mocks belong in test
// files only.
//
// RED state: fails because cmd/executor.go defines MockExecutor with mock.Mock.
// GREEN state: passes after MockExecutor is removed from executor.go.
func TestNoMockStructInProductionCode(t *testing.T) {
	entries, err := os.ReadDir(".")
	require.NoError(t, err, "should be able to read cmd/ directory")

	var violations []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		if strings.HasSuffix(name, "_test.go") {
			continue
		}

		if !strings.HasSuffix(name, ".go") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(".", name))
		require.NoErrorf(t, err, "should be able to read %s", name)

		src := string(content)

		// Detect embedding of mock.Mock (the testify mock embed pattern).
		if strings.Contains(src, "mock.Mock") {
			violations = append(violations, name)
		}
	}

	assert.Empty(t, violations,
		"production (non-test) Go files must not embed mock.Mock — "+
			"found violations: %v", violations)
}
