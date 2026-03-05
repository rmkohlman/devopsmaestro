//go:build !integration

// Package cmd - build_language_test.go contains tests for the getLanguageFromApp
// function, specifically verifying that language detection uses the worktree
// sourcePath rather than the bare git mirror path (Bug #7 fix).
//
// Run these tests with: go test ./cmd/... -run "TestGetLanguageFromApp" -v

package cmd

import (
	"devopsmaestro/models"
	"os"
	"path/filepath"
	"testing"
)

// TestGetLanguageFromApp_UsesSourcePath tests that language detection uses the
// sourcePath (worktree checkout) rather than app.Path (bare git mirror).
// This verifies Bug #7 is fixed: getLanguageFromApp(app, sourcePath) accepts
// a sourcePath parameter and uses it for detection instead of app.Path.
func TestGetLanguageFromApp_UsesSourcePath(t *testing.T) {
	// Arrange: Create two temp directories simulating the bug scenario
	//
	// barePath = app.Path = bare git mirror (contains NO source files)
	// sourcePath = worktree checkout (contains go.mod and source files)
	//
	// The bug was: getLanguageFromApp used app.Path for detection.
	// Since app.Path points to the bare mirror (no source files), detection
	// returned "unknown". The fix accepts a sourcePath parameter and
	// uses that for detection instead.

	barePath := t.TempDir()
	sourcePath := t.TempDir()

	// barePath has no language indicator files — simulates bare git mirror
	// (bare repos have .git metadata only, no working tree)

	// sourcePath has a go.mod — simulates the worktree checkout where code lives
	err := os.WriteFile(filepath.Join(sourcePath, "go.mod"), []byte("module test\ngo 1.22\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// Create App with Path pointing to the bare mirror (the bug scenario)
	app := &models.App{
		Name: "test-app",
		Path: barePath, // Points to bare mirror — has NO source files
	}
	// No Language config set — forces auto-detection path

	// Act: Call with the updated signature that accepts sourcePath
	langName, version, detected := getLanguageFromApp(app, sourcePath)

	// Assert: Language should be detected from sourcePath, NOT from app.Path
	if !detected {
		t.Error("expected detected == true (auto-detection path), got false")
	}
	if langName != "golang" {
		t.Errorf("expected langName == %q (from sourcePath), got %q", "golang", langName)
	}
	// go.mod says "go 1.22", so version should be "1.22"
	if version == "" {
		t.Error("expected non-empty version from go.mod, got empty string")
	}
}
