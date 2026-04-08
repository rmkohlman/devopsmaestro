package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra/doc"
)

// TestGenerateDocsCmd_IsRegistered verifies the generate-docs command is
// registered on the root command.
func TestGenerateDocsCmd_IsRegistered(t *testing.T) {
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "generate-docs" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("generate-docs command not registered on dvm root command")
	}
}

// TestGenerateDocsCmd_IsHidden verifies the command is hidden from normal users.
func TestGenerateDocsCmd_IsHidden(t *testing.T) {
	if !generateDocsCmd.Hidden {
		t.Error("generate-docs should be hidden from normal users")
	}
}

// TestGenerateDocsCmd_Flags verifies required flags exist with correct types.
func TestGenerateDocsCmd_Flags(t *testing.T) {
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
			f := generateDocsCmd.Flags().Lookup(tt.name)
			if f == nil {
				t.Fatalf("--%-12s flag not found on generate-docs", tt.name)
			}
			if f.Value.Type() != tt.wantType {
				t.Errorf("--%s type = %q, want %q", tt.name, f.Value.Type(), tt.wantType)
			}
		})
	}
}

// TestGenerateDocsCmd_NoFormatFlagErrors verifies that calling generate-docs
// without --man-pages or --markdown returns an error.
func TestGenerateDocsCmd_NoFormatFlagErrors(t *testing.T) {
	tmpDir := t.TempDir()
	cmd := generateDocsCmd
	cmd.ResetFlags()
	cmd.Flags().String("output-dir", tmpDir, "")
	cmd.Flags().Bool("man-pages", false, "")
	cmd.Flags().Bool("markdown", false, "")

	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when no format flag is set")
	}
}

// TestGenerateManPages_SmokeTest verifies that man pages are written to disk.
// Uses a temp directory and calls generateManPages directly.
func TestGenerateManPages_SmokeTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping man page smoke test in short mode")
	}

	tmpDir := t.TempDir()

	if err := generateManPages(rootCmd, tmpDir); err != nil {
		t.Fatalf("generateManPages returned error: %v", err)
	}

	// Verify at least one .1 man page was created
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read output dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected man page files to be written, got none")
	}

	// Verify the root man page exists
	rootPage := filepath.Join(tmpDir, "dvm.1")
	if _, err := os.Stat(rootPage); os.IsNotExist(err) {
		t.Errorf("expected root man page %s to exist", rootPage)
	}
}

// TestGenMarkdownTree_SmokeTest verifies that markdown files are written for dvm.
func TestGenMarkdownTree_SmokeTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping markdown smoke test in short mode")
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
	rootMd := filepath.Join(tmpDir, "dvm.md")
	if _, err := os.Stat(rootMd); os.IsNotExist(err) {
		t.Errorf("expected root markdown %s to exist", rootMd)
	}
}

// TestShouldSkipAutoMigration_GenerateDocs verifies generate-docs skips DB init.
func TestShouldSkipAutoMigration_GenerateDocs(t *testing.T) {
	if !shouldSkipAutoMigration(generateDocsCmd) {
		t.Error("shouldSkipAutoMigration should return true for generate-docs")
	}
}
