//go:build !integration

package utils

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDetectPythonVersion verifies that DetectVersion("python", ...) correctly
// extracts a semantic version from various .python-version file contents.
//
// TDD Phase 2: These tests are written BEFORE the fix. Cases with non-standard
// inputs (prefixed, suffixed, lts-style) are expected to FAIL until
// detectPythonVersion() is updated to use regex extraction.
func TestDetectPythonVersion(t *testing.T) {
	const pythonDefault = "3.11"

	tests := []struct {
		name        string
		fileContent *string // nil means "don't create the file"
		expected    string
	}{
		{
			name:        "plain version",
			fileContent: strPtr("3.9.9"),
			expected:    "3.9.9",
		},
		{
			name:        "plain version with newline",
			fileContent: strPtr("3.9.9\n"),
			expected:    "3.9.9",
		},
		{
			name:        "prefixed version daa-api-3.9.9",
			fileContent: strPtr("daa-api-3.9.9"),
			expected:    "3.9.9",
		},
		{
			name:        "prefixed version myproject-3.11.2",
			fileContent: strPtr("myproject-3.11.2"),
			expected:    "3.11.2",
		},
		{
			name:        "suffixed version 3.10.5-dev",
			fileContent: strPtr("3.10.5-dev"),
			expected:    "3.10.5",
		},
		{
			name:        "complex prefix and suffix app-3.11.2-beta",
			fileContent: strPtr("app-3.11.2-beta"),
			expected:    "3.11.2",
		},
		{
			name:        "major.minor only 3.9",
			fileContent: strPtr("3.9"),
			expected:    "3.9",
		},
		{
			name:        "invalid string with no version digits",
			fileContent: strPtr("invalid-no-version"),
			expected:    pythonDefault,
		},
		{
			name:        "empty file",
			fileContent: strPtr(""),
			expected:    pythonDefault,
		},
		{
			name:        "no file exists",
			fileContent: nil,
			expected:    pythonDefault,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			if tt.fileContent != nil {
				versionFile := filepath.Join(dir, ".python-version")
				if err := os.WriteFile(versionFile, []byte(*tt.fileContent), 0644); err != nil {
					t.Fatalf("failed to write .python-version: %v", err)
				}
			}

			got := DetectVersion("python", dir)
			if got != tt.expected {
				t.Errorf("DetectVersion(\"python\", dir) = %q, want %q (input: %q)",
					got, tt.expected, safeDeref(tt.fileContent))
			}
		})
	}
}

// TestDetectNodeVersion verifies that DetectVersion("nodejs", ...) correctly
// extracts a semantic version from various .nvmrc file contents.
//
// TDD Phase 2: These tests are written BEFORE the fix. Cases with v-prefixed
// versions (e.g. "v18.17.0") or lts aliases ("lts/*", "lts/hydrogen", "node")
// are expected to FAIL until detectNodeVersion() is updated to strip the "v"
// prefix and return the default for non-numeric aliases.
func TestDetectNodeVersion(t *testing.T) {
	const nodeDefault = "20"

	tests := []struct {
		name        string
		fileContent *string // nil means "don't create the file"
		expected    string
	}{
		{
			name:        "plain numeric version",
			fileContent: strPtr("18.17.0"),
			expected:    "18.17.0",
		},
		{
			name:        "plain numeric version with newline",
			fileContent: strPtr("18.17.0\n"),
			expected:    "18.17.0",
		},
		{
			name:        "v-prefixed version v18.17.0",
			fileContent: strPtr("v18.17.0"),
			expected:    "18.17.0",
		},
		{
			name:        "lts alias lts/*",
			fileContent: strPtr("lts/*"),
			expected:    nodeDefault,
		},
		{
			name:        "lts named alias lts/hydrogen",
			fileContent: strPtr("lts/hydrogen"),
			expected:    nodeDefault,
		},
		{
			name:        "bare node alias",
			fileContent: strPtr("node"),
			expected:    nodeDefault,
		},
		{
			name:        "major.minor only 20.10",
			fileContent: strPtr("20.10"),
			expected:    "20.10",
		},
		{
			name:        "prefixed version myapp-16.14.0",
			fileContent: strPtr("myapp-16.14.0"),
			expected:    "16.14.0",
		},
		{
			name:        "no file exists",
			fileContent: nil,
			expected:    nodeDefault,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			if tt.fileContent != nil {
				nvmrc := filepath.Join(dir, ".nvmrc")
				if err := os.WriteFile(nvmrc, []byte(*tt.fileContent), 0644); err != nil {
					t.Fatalf("failed to write .nvmrc: %v", err)
				}
			}

			got := DetectVersion("nodejs", dir)
			if got != tt.expected {
				t.Errorf("DetectVersion(\"nodejs\", dir) = %q, want %q (input: %q)",
					got, tt.expected, safeDeref(tt.fileContent))
			}
		})
	}
}

// strPtr is a helper that returns a pointer to a string literal.
func strPtr(s string) *string {
	return &s
}

// safeDeref returns the string value of a pointer, or "<nil>" if the pointer is nil.
func safeDeref(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}
