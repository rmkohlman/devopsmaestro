package main

import (
	"testing"
)

func TestVersionCommand(t *testing.T) {
	// Reset args for each test
	rootCmd.SetArgs([]string{"version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
}

func TestPromptLibraryListCommand(t *testing.T) {
	rootCmd.SetArgs([]string{"prompt", "library", "list"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("prompt library list command failed: %v", err)
	}
}

func TestPluginLibraryListCommand(t *testing.T) {
	rootCmd.SetArgs([]string{"plugin", "library", "list"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("plugin library list command failed: %v", err)
	}
}

func TestProfilePresetListCommand(t *testing.T) {
	rootCmd.SetArgs([]string{"profile", "preset", "list"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("profile preset list command failed: %v", err)
	}
}

func TestPromptLibraryShowStarshipDefaultCommand(t *testing.T) {
	// Test the library show command instead of generate for a library item
	rootCmd.SetArgs([]string{"prompt", "library", "show", "starship-default"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("prompt library show command failed: %v", err)
	}
}

func TestPromptLibraryCategoriesCommand(t *testing.T) {
	rootCmd.SetArgs([]string{"prompt", "library", "categories"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("prompt library categories command failed: %v", err)
	}
}

func TestPluginLibraryCategoriesCommand(t *testing.T) {
	rootCmd.SetArgs([]string{"plugin", "library", "categories"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("plugin library categories command failed: %v", err)
	}
}

func TestPromptLibraryShowCommand(t *testing.T) {
	rootCmd.SetArgs([]string{"prompt", "library", "show", "starship-default"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("prompt library show command failed: %v", err)
	}
}

func TestPluginLibraryShowCommand(t *testing.T) {
	rootCmd.SetArgs([]string{"plugin", "library", "show", "zsh-autosuggestions"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("plugin library show command failed: %v", err)
	}
}
