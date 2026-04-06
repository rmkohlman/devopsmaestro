package builders

// Tests for Issue #190: workspace containers missing openssh-client
//
// The function getDefaultPackages() in dockerfile_generator.go (~line 1020)
// returns ["git", "curl", "wget", "zsh"] but does NOT include "openssh-client".
// This means SSH-based git operations (git clone via SSH, git push via SSH, etc.)
// fail inside every workspace container regardless of language.
//
// The fix is to add "openssh-client" to the base slice in getDefaultPackages().
//
// All four tests below are RED (failing) against current code — they assert
// openssh-client appears in the generated Dockerfile's apt-get/apk install
// command and will only turn GREEN after the fix lands.

import (
	"os"
	"strings"
	"testing"

	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/paths"
)

// TestDevStage_IncludesOpensshClient_Python verifies that a Python workspace
// Dockerfile contains openssh-client in its package install directive.
//
// Phase 2 (RED) — must FAIL against current code because getDefaultPackages()
// does not include openssh-client.
func TestDevStage_IncludesOpensshClient_Python(t *testing.T) {
	// Arrange: temp dir with a minimal Python project
	appDir := t.TempDir()
	if err := os.WriteFile(appDir+"/main.py", []byte("print('hello')\n"), 0644); err != nil {
		t.Fatalf("setup: failed to create main.py: %v", err)
	}

	ws := &models.Workspace{
		ID:        1,
		Name:      "test-python-ssh",
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

	// Assert: openssh-client must appear in an apt-get or apk install line
	assertOpensshClientPresent(t, "python", dockerfile)
}

// TestDevStage_IncludesOpensshClient_Golang verifies that a Golang workspace
// Dockerfile contains openssh-client in its package install directive.
//
// Phase 2 (RED) — must FAIL against current code.
func TestDevStage_IncludesOpensshClient_Golang(t *testing.T) {
	// Arrange: temp dir with a minimal Go project
	appDir := t.TempDir()
	if err := os.WriteFile(appDir+"/main.go", []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatalf("setup: failed to create main.go: %v", err)
	}

	ws := &models.Workspace{
		ID:        2,
		Name:      "test-golang-ssh",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "golang",
		Version:       "1.22",
		AppPath:       appDir,
		PathConfig:    paths.New(t.TempDir()),
	})

	// Act
	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Assert
	assertOpensshClientPresent(t, "golang", dockerfile)
}

// TestDevStage_IncludesOpensshClient_Nodejs verifies that a Node.js workspace
// Dockerfile contains openssh-client in its package install directive.
//
// Phase 2 (RED) — must FAIL against current code.
func TestDevStage_IncludesOpensshClient_Nodejs(t *testing.T) {
	// Arrange: temp dir with a minimal Node.js project
	appDir := t.TempDir()
	if err := os.WriteFile(appDir+"/index.js", []byte("console.log('hello');\n"), 0644); err != nil {
		t.Fatalf("setup: failed to create index.js: %v", err)
	}

	ws := &models.Workspace{
		ID:        3,
		Name:      "test-nodejs-ssh",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "nodejs",
		Version:       "20",
		AppPath:       appDir,
		PathConfig:    paths.New(t.TempDir()),
	})

	// Act
	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Assert
	assertOpensshClientPresent(t, "nodejs", dockerfile)
}

// TestDevStage_IncludesOpensshClient_Default verifies that an unknown/default
// language workspace Dockerfile contains openssh-client in its package install
// directive. The default/fallback branch in getDefaultPackages() must also include
// openssh-client so that any language variation is covered.
//
// Phase 2 (RED) — must FAIL against current code.
func TestDevStage_IncludesOpensshClient_Default(t *testing.T) {
	// Arrange: temp dir (no language-specific file needed)
	appDir := t.TempDir()

	ws := &models.Workspace{
		ID:        4,
		Name:      "test-default-ssh",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "unknown",
		Version:       "",
		AppPath:       appDir,
		PathConfig:    paths.New(t.TempDir()),
	})

	// Act
	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Assert
	assertOpensshClientPresent(t, "unknown/default", dockerfile)
}

// assertOpensshClientPresent is a shared assertion helper that checks for
// "openssh-client" anywhere in the generated Dockerfile and emits a
// diagnostic error message pointing at the dev-stage install block.
func assertOpensshClientPresent(t *testing.T, language, dockerfile string) {
	t.Helper()

	if strings.Contains(dockerfile, "openssh-client") {
		return // PASS — package is present
	}

	t.Errorf(
		"language=%q: Generate() Dockerfile does not contain 'openssh-client'.\n"+
			"Bug (Issue #190): getDefaultPackages() in builders/dockerfile_generator.go (~line 1020)\n"+
			"returns [\"git\", \"curl\", \"wget\", \"zsh\"] but omits \"openssh-client\".\n"+
			"SSH-based git operations will fail inside the workspace container.\n"+
			"Fix: add \"openssh-client\" to the base slice in getDefaultPackages().\n\n"+
			"Dev-stage install section of generated Dockerfile:\n%s",
		language,
		extractDevStageInstallSection(dockerfile),
	)
}

// extractDevStageInstallSection pulls out the merged package-install block
// from the dev stage for cleaner diagnostic output in test failures.
func extractDevStageInstallSection(dockerfile string) string {
	const marker = "# Install all dev tools"
	idx := strings.Index(dockerfile, marker)
	if idx < 0 {
		// Fallback: return first 80 lines
		lines := strings.Split(dockerfile, "\n")
		if len(lines) > 80 {
			lines = lines[:80]
		}
		return strings.Join(lines, "\n")
	}

	// Grab from the marker until the next blank-line-terminated RUN block
	section := dockerfile[idx:]
	lines := strings.Split(section, "\n")
	// Return up to 30 lines starting at the marker — enough to see the full
	// apt-get/apk install command plus a few lines of context.
	if len(lines) > 30 {
		lines = lines[:30]
	}
	return strings.Join(lines, "\n")
}
