package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra/doc"
)

// TestGenerateDocsNvpCmd_IsRegistered verifies the generate-docs command is
// registered on the nvp root command.
func TestGenerateDocsNvpCmd_IsRegistered(t *testing.T) {
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "generate-docs" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("generate-docs command not registered on nvp root command")
	}
}

// TestGenerateDocsNvpCmd_IsHidden verifies the command is hidden from normal users.
func TestGenerateDocsNvpCmd_IsHidden(t *testing.T) {
	if !generateDocsNvpCmd.Hidden {
		t.Error("generate-docs should be hidden from normal nvp users")
	}
}

// TestGenerateDocsNvpCmd_Flags verifies required flags exist with correct types.
func TestGenerateDocsNvpCmd_Flags(t *testing.T) {
	flags := []struct {
		name     string
		wantType string
	}{
		{"output-dir", "string"},
		{"man-pages", "bool"},
		{"markdown", "bool"},
	}

	for _, tt := range flags {
		t.Run(tt.name, func(t *testing.T) {
			f := generateDocsNvpCmd.Flags().Lookup(tt.name)
			if f == nil {
				t.Fatalf("--%-12s flag not found on nvp generate-docs", tt.name)
			}
			if f.Value.Type() != tt.wantType {
				t.Errorf("--%s type = %q, want %q", tt.name, f.Value.Type(), tt.wantType)
			}
		})
	}
}

// TestGenerateDocsNvpCmd_NoFormatFlagErrors verifies that calling generate-docs
// without --man-pages or --markdown returns an error.
func TestGenerateDocsNvpCmd_NoFormatFlagErrors(t *testing.T) {
	tmpDir := t.TempDir()
	cmd := generateDocsNvpCmd
	cmd.ResetFlags()
	cmd.Flags().String("output-dir", tmpDir, "")
	cmd.Flags().Bool("man-pages", false, "")
	cmd.Flags().Bool("markdown", false, "")

	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when no format flag is set for nvp generate-docs")
	}
}

// TestGenerateNvpManPages_SmokeTest verifies that man pages are written to disk.
func TestGenerateNvpManPages_SmokeTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping nvp man page smoke test in short mode")
	}

	tmpDir := t.TempDir()

	if err := generateNvpManPages(rootCmd, tmpDir); err != nil {
		t.Fatalf("generateNvpManPages returned error: %v", err)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read output dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected man page files to be written, got none")
	}

	// Verify the root man page for nvp exists
	rootPage := filepath.Join(tmpDir, "nvp.1")
	if _, err := os.Stat(rootPage); os.IsNotExist(err) {
		t.Errorf("expected root man page %s to exist", rootPage)
	}
}

// TestNvpGenMarkdownTree_SmokeTest verifies that markdown files are written for nvp.
func TestNvpGenMarkdownTree_SmokeTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping nvp markdown smoke test in short mode")
	}

	tmpDir := t.TempDir()

	if err := doc.GenMarkdownTree(rootCmd, tmpDir); err != nil {
		t.Fatalf("GenMarkdownTree returned error: %v", err)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read output dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected markdown files to be written, got none")
	}

	// Verify root markdown exists
	rootMd := filepath.Join(tmpDir, "nvp.md")
	if _, err := os.Stat(rootMd); os.IsNotExist(err) {
		t.Errorf("expected root markdown %s to exist", rootMd)
	}
}

// TestNvpShouldSkipAutoMigration_GenerateDocs verifies generate-docs skips DB init.
func TestNvpShouldSkipAutoMigration_GenerateDocs(t *testing.T) {
	if !shouldSkipAutoMigration(generateDocsNvpCmd) {
		t.Error("shouldSkipAutoMigration should return true for nvp generate-docs")
	}
}
