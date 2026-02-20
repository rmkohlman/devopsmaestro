package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetStoredVersion_NoFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "dvm-version-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Temporarily change HOME to our test directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	version, err := GetStoredVersion()
	if err != nil {
		t.Errorf("Expected no error for missing file, got: %v", err)
	}
	if version != "" {
		t.Errorf("Expected empty version for missing file, got: %s", version)
	}
}

func TestSaveAndGetStoredVersion(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "dvm-version-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Temporarily change HOME to our test directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	testVersion := "v1.2.3"

	// Save version
	err = SaveCurrentVersion(testVersion)
	if err != nil {
		t.Errorf("Failed to save version: %v", err)
	}

	// Check that .devopsmaestro directory was created
	configDir := filepath.Join(tmpDir, ".devopsmaestro")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Config directory was not created")
	}

	// Get version
	storedVersion, err := GetStoredVersion()
	if err != nil {
		t.Errorf("Failed to get stored version: %v", err)
	}
	if storedVersion != testVersion {
		t.Errorf("Expected version %s, got %s", testVersion, storedVersion)
	}
}

func TestSaveStoredVersion_UpdatesExisting(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "dvm-version-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Temporarily change HOME to our test directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Save first version
	firstVersion := "v1.0.0"
	err = SaveCurrentVersion(firstVersion)
	if err != nil {
		t.Errorf("Failed to save first version: %v", err)
	}

	// Save second version
	secondVersion := "v2.0.0"
	err = SaveCurrentVersion(secondVersion)
	if err != nil {
		t.Errorf("Failed to save second version: %v", err)
	}

	// Get version - should be the second one
	storedVersion, err := GetStoredVersion()
	if err != nil {
		t.Errorf("Failed to get stored version: %v", err)
	}
	if storedVersion != secondVersion {
		t.Errorf("Expected version %s, got %s", secondVersion, storedVersion)
	}
}

func TestGetStoredVersion_HandlesWhitespace(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "dvm-version-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Temporarily change HOME to our test directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Create config directory manually
	configDir := filepath.Join(tmpDir, ".devopsmaestro")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Write version with whitespace
	versionFile := filepath.Join(configDir, ".version")
	testVersion := "v1.2.3"
	err = os.WriteFile(versionFile, []byte("  "+testVersion+"\n  "), 0644)
	if err != nil {
		t.Fatalf("Failed to write version file: %v", err)
	}

	// Get version - should be trimmed
	storedVersion, err := GetStoredVersion()
	if err != nil {
		t.Errorf("Failed to get stored version: %v", err)
	}
	if storedVersion != testVersion {
		t.Errorf("Expected trimmed version %s, got %s", testVersion, storedVersion)
	}
}
