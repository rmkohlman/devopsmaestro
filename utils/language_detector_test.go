//go:build !integration

package utils

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDetectLanguage_Go verifies that a directory containing go.mod is detected as "golang".
func TestDetectLanguage_Go(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\ngo 1.22\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error: %v", err)
	}
	if lang == nil {
		t.Fatal("DetectLanguage returned nil, expected *Language")
	}
	if lang.Name != "golang" {
		t.Errorf("expected Name == %q, got %q", "golang", lang.Name)
	}
}

// TestDetectLanguage_Rust verifies that a directory containing Cargo.toml is detected as "rust".
func TestDetectLanguage_Rust(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]\nname = \"test\"\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write Cargo.toml: %v", err)
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error: %v", err)
	}
	if lang == nil {
		t.Fatal("DetectLanguage returned nil, expected *Language")
	}
	if lang.Name != "rust" {
		t.Errorf("expected Name == %q, got %q", "rust", lang.Name)
	}
}

// TestDetectLanguage_Python verifies that a directory containing requirements.txt is detected as "python".
func TestDetectLanguage_Python(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("flask==3.0\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write requirements.txt: %v", err)
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error: %v", err)
	}
	if lang == nil {
		t.Fatal("DetectLanguage returned nil, expected *Language")
	}
	if lang.Name != "python" {
		t.Errorf("expected Name == %q, got %q", "python", lang.Name)
	}
}

// TestDetectLanguage_Gleam verifies that a directory containing gleam.toml is detected as "gleam".
func TestDetectLanguage_Gleam(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "gleam.toml"), []byte("name = \"test\"\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write gleam.toml: %v", err)
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error: %v", err)
	}
	if lang == nil {
		t.Fatal("DetectLanguage returned nil, expected *Language for gleam project")
	}
	if lang.Name != "gleam" {
		t.Errorf("expected Name == %q, got %q", "gleam", lang.Name)
	}
}

// TestDetectLanguage_EmptyDir verifies that an empty directory returns nil, nil
// (no language detected, no error).
func TestDetectLanguage_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error for empty dir: %v", err)
	}
	if lang != nil {
		t.Errorf("expected nil for empty dir, got %+v", lang)
	}
}

// TestDetectLanguage_BareRepo verifies that a directory with no language files
// (simulating a bare git repository) returns nil, nil.
// This documents the expected behavior: bare repos have no source code files,
// so language detection correctly returns nothing.
func TestDetectLanguage_BareRepo(t *testing.T) {
	dir := t.TempDir()

	// Simulate a bare git repo structure — contains only git metadata, no source files
	gitDir := filepath.Join(dir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0644); err != nil {
		t.Fatalf("failed to write HEAD: %v", err)
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error for bare repo: %v", err)
	}
	if lang != nil {
		t.Errorf("expected nil for bare repo dir, got %+v", lang)
	}
}
