package builders

// Tests for Issue #186: dvm build fails for Python projects without requirements.txt
//
// These tests verify that the Dockerfile generator conditionally emits the
// `COPY requirements.txt` and `pip install -r` directives based on whether
// requirements.txt exists in the appPath directory.
//
// Test 1 (TestPythonDockerfile_NoRequirementsTxt_SkipsPipInstall) — RED:
//   When requirements.txt is absent, pip install directives must be omitted.
//
// Test 2 (TestPythonDockerfile_WithRequirementsTxt_IncludesPipInstall) — GREEN:
//   When requirements.txt is present, pip install directives must be included.
//
// Test 3 (TestPythonDockerfile_EmptyRequirementsTxt_IncludesPipInstall) — GREEN:
//   An empty requirements.txt (file exists, 0 bytes) still triggers pip install.

import (
	"os"
	"strings"
	"testing"

	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/paths"
)

// TestPythonDockerfile_NoRequirementsTxt_SkipsPipInstall verifies that when a Python
// project has NO requirements.txt the generator does NOT emit COPY/pip-install lines.
//
// Phase 2 (RED) — must FAIL against current code because the generator unconditionally
// emits COPY requirements.txt / pip install regardless of file existence.
func TestPythonDockerfile_NoRequirementsTxt_SkipsPipInstall(t *testing.T) {
	// Arrange: temp dir with a Python file but NO requirements.txt
	appDir := t.TempDir()
	if err := os.WriteFile(appDir+"/main.py", []byte("print('hello')\n"), 0644); err != nil {
		t.Fatalf("failed to create main.py: %v", err)
	}

	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       appDir,
		PathConfig:    paths.New(t.TempDir()),
	})

	// Act
	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Assert: pip install directives must be absent when requirements.txt doesn't exist
	if strings.Contains(dockerfile, "COPY requirements.txt") {
		t.Errorf("Generate() should NOT emit 'COPY requirements.txt' when requirements.txt is absent.\n"+
			"Bug: dockerfile generator unconditionally copies requirements.txt regardless of file existence.\n"+
			"Relevant Dockerfile section:\n%s",
			extractPythonBaseSection(dockerfile))
	}
	if strings.Contains(dockerfile, "pip install -r") {
		t.Errorf("Generate() should NOT emit 'pip install -r' when requirements.txt is absent.\n"+
			"Bug: dockerfile generator unconditionally runs pip install -r regardless of file existence.\n"+
			"Relevant Dockerfile section:\n%s",
			extractPythonBaseSection(dockerfile))
	}

	// Assert: a skip/absent indicator comment must be present instead
	noReqsIndicators := []string{
		"# No requirements.txt",
		"# requirements.txt not found",
		"# Skipping pip install",
		"# No requirements",
	}
	found := false
	for _, indicator := range noReqsIndicators {
		if strings.Contains(dockerfile, indicator) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Generate() should emit a comment indicating requirements.txt was skipped.\n"+
			"Expected one of: %v\n"+
			"Relevant Dockerfile section:\n%s",
			noReqsIndicators, extractPythonBaseSection(dockerfile))
	}
}

// TestPythonDockerfile_WithRequirementsTxt_IncludesPipInstall verifies that when a
// Python project HAS a requirements.txt the generator DOES emit COPY/pip-install lines.
//
// This is the existing (correct) behaviour — the test should PASS against current code
// and continue to pass after the fix.
func TestPythonDockerfile_WithRequirementsTxt_IncludesPipInstall(t *testing.T) {
	// Arrange: temp dir WITH requirements.txt containing a package
	appDir := t.TempDir()
	if err := os.WriteFile(appDir+"/requirements.txt", []byte("flask==2.0\n"), 0644); err != nil {
		t.Fatalf("failed to create requirements.txt: %v", err)
	}
	if err := os.WriteFile(appDir+"/main.py", []byte("from flask import Flask\n"), 0644); err != nil {
		t.Fatalf("failed to create main.py: %v", err)
	}

	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       appDir,
		PathConfig:    paths.New(t.TempDir()),
	})

	// Act
	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Assert: pip install directives MUST be present when requirements.txt exists
	if !strings.Contains(dockerfile, "COPY requirements.txt") {
		t.Errorf("Generate() should emit 'COPY requirements.txt' when requirements.txt is present.\n"+
			"Relevant Dockerfile section:\n%s",
			extractPythonBaseSection(dockerfile))
	}
	if !strings.Contains(dockerfile, "pip install -r") {
		t.Errorf("Generate() should emit 'pip install -r' when requirements.txt is present.\n"+
			"Relevant Dockerfile section:\n%s",
			extractPythonBaseSection(dockerfile))
	}
}

// TestPythonDockerfile_EmptyRequirementsTxt_IncludesPipInstall verifies that when a
// Python project has an EMPTY requirements.txt (file exists but 0 bytes) the generator
// still emits COPY/pip-install lines — the file's existence, not content, is what matters.
//
// This is the existing (correct) behaviour — the test should PASS against current code
// and continue to pass after the fix.
func TestPythonDockerfile_EmptyRequirementsTxt_IncludesPipInstall(t *testing.T) {
	// Arrange: temp dir with an EMPTY requirements.txt (0 bytes)
	appDir := t.TempDir()
	if err := os.WriteFile(appDir+"/requirements.txt", []byte{}, 0644); err != nil {
		t.Fatalf("failed to create empty requirements.txt: %v", err)
	}
	if err := os.WriteFile(appDir+"/main.py", []byte("print('hello')\n"), 0644); err != nil {
		t.Fatalf("failed to create main.py: %v", err)
	}

	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       appDir,
		PathConfig:    paths.New(t.TempDir()),
	})

	// Act
	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Assert: pip install directives MUST be present when requirements.txt exists (even empty)
	if !strings.Contains(dockerfile, "COPY requirements.txt") {
		t.Errorf("Generate() should emit 'COPY requirements.txt' when requirements.txt exists (even if empty).\n"+
			"An empty requirements.txt is still a valid file that the user explicitly placed.\n"+
			"Relevant Dockerfile section:\n%s",
			extractPythonBaseSection(dockerfile))
	}
	if !strings.Contains(dockerfile, "pip install -r") {
		t.Errorf("Generate() should emit 'pip install -r' when requirements.txt exists (even if empty).\n"+
			"An empty requirements.txt is still a valid file that the user explicitly placed.\n"+
			"Relevant Dockerfile section:\n%s",
			extractPythonBaseSection(dockerfile))
	}
}

// extractPythonBaseSection pulls out the base stage section of the Dockerfile for
// cleaner error messages — avoids dumping the entire multi-hundred-line file.
func extractPythonBaseSection(dockerfile string) string {
	// The base stage runs from the top of the file until the first parallel builder
	builderMarker := "# --- Parallel builder:"
	if idx := strings.Index(dockerfile, builderMarker); idx > 0 {
		section := dockerfile[:idx]
		// Only return last 40 lines to keep error messages manageable
		lines := strings.Split(section, "\n")
		if len(lines) > 40 {
			lines = lines[len(lines)-40:]
		}
		return strings.Join(lines, "\n")
	}
	// Fall back: first 60 lines
	lines := strings.Split(dockerfile, "\n")
	if len(lines) > 60 {
		lines = lines[:60]
	}
	return strings.Join(lines, "\n")
}
