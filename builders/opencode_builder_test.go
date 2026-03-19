package builders

// =============================================================================
// Issue #113 — TDD Phase 2: generateOpencodeBuilder() builder stage tests
//
// RED: These tests WILL NOT COMPILE until generateOpencodeBuilder() and the
// supporting constants are implemented:
//
//   - opencodeVersion constant does not exist yet (checksums.go)
//   - opencodeChecksumArm64 constant does not exist yet (checksums.go)
//   - opencodeChecksumAmd64 constant does not exist yet (checksums.go)
//   - DefaultDockerfileGenerator.generateOpencodeBuilder() does not exist yet
//   - WorkspaceSpec.Tools.Opencode field does not exist yet (models)
//
// Once the production code is added, all tests in this file MUST pass.
// =============================================================================

import (
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/paths"
	"strings"
	"testing"
)

// TestOpencodeBuilder_StageHeader verifies that generateOpencodeBuilder() emits
// the expected stage header comment and AS name.
//
// RED: WILL NOT COMPILE — generateOpencodeBuilder() does not exist yet.
func TestOpencodeBuilder_StageHeader(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	// ── COMPILE ERROR EXPECTED BELOW ──────────────────────────────────────────
	// models.ToolsConfig / WorkspaceSpec.Tools do not exist until #113.
	wsYAML := models.WorkspaceSpec{
		Tools: models.ToolsConfig{
			Opencode: true,
		},
	}
	// ──────────────────────────────────────────────────────────────────────────

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	wantPatterns := []string{
		// Stage header comment — consistent with other builder stages
		"# --- Parallel builder: opencode ---",
		// Stage AS name
		"AS opencode-builder",
	}
	for _, want := range wantPatterns {
		if !strings.Contains(dockerfile, want) {
			t.Errorf("Generate() missing expected opencode builder pattern: %q", want)
		}
	}
}

// TestOpencodeBuilder_SetE verifies that the opencode builder RUN command begins
// with `set -e` for fail-fast error handling — consistent with all other builders.
//
// RED: WILL NOT COMPILE — generateOpencodeBuilder() does not exist yet.
func TestOpencodeBuilder_SetE(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	// ── COMPILE ERROR EXPECTED BELOW ──────────────────────────────────────────
	wsYAML := models.WorkspaceSpec{
		Tools: models.ToolsConfig{Opencode: true},
	}
	// ──────────────────────────────────────────────────────────────────────────

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Isolate the opencode builder section
	opencodeStart := strings.Index(dockerfile, "# --- Parallel builder: opencode ---")
	if opencodeStart < 0 {
		t.Fatalf("opencode builder section not found in Dockerfile")
	}
	// Find next section boundary (the dev stage or another builder after opencode)
	nextSection := strings.Index(dockerfile[opencodeStart+1:], "\n# ")
	var opencodeSection string
	if nextSection > 0 {
		opencodeSection = dockerfile[opencodeStart : opencodeStart+1+nextSection]
	} else {
		opencodeSection = dockerfile[opencodeStart:]
	}

	if !strings.Contains(opencodeSection, "set -e") {
		t.Errorf("opencode builder stage missing 'set -e'.\n"+
			"Every builder stage RUN command must begin with 'set -e'.\n"+
			"opencode section:\n%s", opencodeSection)
	}
}

// TestOpencodeBuilder_HardenedCurl verifies that the opencode builder uses
// hardened curl flags: -fsSL --retry 3 --connect-timeout 30.
// No curl|sh patterns are allowed.
//
// RED: WILL NOT COMPILE — generateOpencodeBuilder() does not exist yet.
func TestOpencodeBuilder_HardenedCurl(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	// ── COMPILE ERROR EXPECTED BELOW ──────────────────────────────────────────
	wsYAML := models.WorkspaceSpec{
		Tools: models.ToolsConfig{Opencode: true},
	}
	// ──────────────────────────────────────────────────────────────────────────

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	opencodeStart := strings.Index(dockerfile, "# --- Parallel builder: opencode ---")
	if opencodeStart < 0 {
		t.Fatalf("opencode builder section not found in Dockerfile")
	}
	nextSection := strings.Index(dockerfile[opencodeStart+1:], "\n# ")
	var opencodeSection string
	if nextSection > 0 {
		opencodeSection = dockerfile[opencodeStart : opencodeStart+1+nextSection]
	} else {
		opencodeSection = dockerfile[opencodeStart:]
	}

	// MUST have hardened curl flags
	if !strings.Contains(opencodeSection, "curl -fsSL --retry 3 --connect-timeout 30") {
		t.Errorf("opencode builder missing hardened curl flags: 'curl -fsSL --retry 3 --connect-timeout 30'.\n"+
			"opencode section:\n%s", opencodeSection)
	}

	// MUST NOT use pipe-to-shell or install script pattern
	badPatterns := []string{
		"curl -sS",
		"curl -L ",
		"| sh",
		"| bash",
	}
	for _, bad := range badPatterns {
		if strings.Contains(opencodeSection, bad) {
			t.Errorf("opencode builder contains insecure curl pattern: %q.\n"+
				"Must use direct binary download with checksum verification.\n"+
				"opencode section:\n%s", bad, opencodeSection)
		}
	}
}

// TestOpencodeBuilder_SHA256Verification verifies that the opencode builder
// includes SHA256 checksum verification before executing the binary.
//
// RED: WILL NOT COMPILE — generateOpencodeBuilder() does not exist yet.
func TestOpencodeBuilder_SHA256Verification(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	// ── COMPILE ERROR EXPECTED BELOW ──────────────────────────────────────────
	wsYAML := models.WorkspaceSpec{
		Tools: models.ToolsConfig{Opencode: true},
	}
	// ──────────────────────────────────────────────────────────────────────────

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	opencodeStart := strings.Index(dockerfile, "# --- Parallel builder: opencode ---")
	if opencodeStart < 0 {
		t.Fatalf("opencode builder section not found in Dockerfile")
	}
	nextSection := strings.Index(dockerfile[opencodeStart+1:], "\n# ")
	var opencodeSection string
	if nextSection > 0 {
		opencodeSection = dockerfile[opencodeStart : opencodeStart+1+nextSection]
	} else {
		opencodeSection = dockerfile[opencodeStart:]
	}

	if !strings.Contains(opencodeSection, "sha256sum -c -") {
		t.Errorf("opencode builder missing SHA256 checksum verification: 'sha256sum -c -'.\n"+
			"All binary downloads must be verified against known checksums.\n"+
			"opencode section:\n%s", opencodeSection)
	}
}

// TestOpencodeBuilder_BinaryDestination verifies that the opencode builder
// installs the binary to /usr/local/bin/opencode.
//
// RED: WILL NOT COMPILE — generateOpencodeBuilder() does not exist yet.
func TestOpencodeBuilder_BinaryDestination(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	// ── COMPILE ERROR EXPECTED BELOW ──────────────────────────────────────────
	wsYAML := models.WorkspaceSpec{
		Tools: models.ToolsConfig{Opencode: true},
	}
	// ──────────────────────────────────────────────────────────────────────────

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	opencodeStart := strings.Index(dockerfile, "# --- Parallel builder: opencode ---")
	if opencodeStart < 0 {
		t.Fatalf("opencode builder section not found in Dockerfile")
	}
	nextSection := strings.Index(dockerfile[opencodeStart+1:], "\n# ")
	var opencodeSection string
	if nextSection > 0 {
		opencodeSection = dockerfile[opencodeStart : opencodeStart+1+nextSection]
	} else {
		opencodeSection = dockerfile[opencodeStart:]
	}

	// Binary must be placed at /usr/local/bin/opencode
	if !strings.Contains(opencodeSection, "/usr/local/bin/opencode") {
		t.Errorf("opencode builder missing binary destination '/usr/local/bin/opencode'.\n"+
			"opencode section:\n%s", opencodeSection)
	}
}

// TestOpencodeBuilder_BinaryVerification verifies that the opencode builder
// includes `test -x /usr/local/bin/opencode` to confirm installation.
//
// RED: WILL NOT COMPILE — generateOpencodeBuilder() does not exist yet.
func TestOpencodeBuilder_BinaryVerification(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	// ── COMPILE ERROR EXPECTED BELOW ──────────────────────────────────────────
	wsYAML := models.WorkspaceSpec{
		Tools: models.ToolsConfig{Opencode: true},
	}
	// ──────────────────────────────────────────────────────────────────────────

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	opencodeStart := strings.Index(dockerfile, "# --- Parallel builder: opencode ---")
	if opencodeStart < 0 {
		t.Fatalf("opencode builder section not found in Dockerfile")
	}
	nextSection := strings.Index(dockerfile[opencodeStart+1:], "\n# ")
	var opencodeSection string
	if nextSection > 0 {
		opencodeSection = dockerfile[opencodeStart : opencodeStart+1+nextSection]
	} else {
		opencodeSection = dockerfile[opencodeStart:]
	}

	if !strings.Contains(opencodeSection, "test -x /usr/local/bin/opencode") {
		t.Errorf("opencode builder missing binary verification: 'test -x /usr/local/bin/opencode'.\n"+
			"Each builder stage must verify the binary was installed correctly.\n"+
			"opencode section:\n%s", opencodeSection)
	}
}

// TestOpencodeBuilder_ArchitectureHandling verifies that the opencode builder
// handles both linux/amd64 and linux/arm64 architectures.
//
// RED: WILL NOT COMPILE — generateOpencodeBuilder() and checksum constants do not exist yet.
func TestOpencodeBuilder_ArchitectureHandling(t *testing.T) {
	tests := []struct {
		name     string
		language string
		version  string
	}{
		{name: "debian base (python)", language: "python", version: "3.11"},
		{name: "alpine base (golang)", language: "golang", version: "1.22"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &models.Workspace{
				ID:        1,
				Name:      "test-ws",
				ImageName: "test:latest",
			}
			// ── COMPILE ERROR EXPECTED BELOW ──────────────────────────────
			wsYAML := models.WorkspaceSpec{
				Tools: models.ToolsConfig{Opencode: true},
			}
			// ──────────────────────────────────────────────────────────────

			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
				Workspace:     ws,
				WorkspaceSpec: wsYAML,
				Language:      tt.language,
				Version:       tt.version,
				AppPath:       "/tmp/test",
				PathConfig:    paths.New(t.TempDir()),
			})

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			opencodeStart := strings.Index(dockerfile, "# --- Parallel builder: opencode ---")
			if opencodeStart < 0 {
				t.Fatalf("opencode builder section not found in Dockerfile")
			}
			nextSection := strings.Index(dockerfile[opencodeStart+1:], "\n# ")
			var opencodeSection string
			if nextSection > 0 {
				opencodeSection = dockerfile[opencodeStart : opencodeStart+1+nextSection]
			} else {
				opencodeSection = dockerfile[opencodeStart:]
			}

			// ── COMPILE ERROR EXPECTED BELOW ──────────────────────────────────
			// opencodeChecksumArm64 / opencodeChecksumAmd64 do not exist until #113.
			// Both checksums must appear in the builder (one per arch branch).
			if !strings.Contains(opencodeSection, opencodeChecksumArm64) {
				t.Errorf("opencode builder missing arm64 checksum for %s.\n"+
					"opencode section:\n%s", tt.name, opencodeSection)
			}
			if !strings.Contains(opencodeSection, opencodeChecksumAmd64) {
				t.Errorf("opencode builder missing amd64 checksum for %s.\n"+
					"opencode section:\n%s", tt.name, opencodeSection)
			}
			// ──────────────────────────────────────────────────────────────────
		})
	}
}

// TestOpencodeBuilder_PinnedVersion verifies that the opencode builder uses
// the pinned opencodeVersion constant in the download URL.
//
// RED: WILL NOT COMPILE — opencodeVersion constant does not exist yet.
func TestOpencodeBuilder_PinnedVersion(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	// ── COMPILE ERROR EXPECTED BELOW ──────────────────────────────────────────
	wsYAML := models.WorkspaceSpec{
		Tools: models.ToolsConfig{Opencode: true},
	}
	// ──────────────────────────────────────────────────────────────────────────

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	opencodeStart := strings.Index(dockerfile, "# --- Parallel builder: opencode ---")
	if opencodeStart < 0 {
		t.Fatalf("opencode builder section not found in Dockerfile")
	}
	nextSection := strings.Index(dockerfile[opencodeStart+1:], "\n# ")
	var opencodeSection string
	if nextSection > 0 {
		opencodeSection = dockerfile[opencodeStart : opencodeStart+1+nextSection]
	} else {
		opencodeSection = dockerfile[opencodeStart:]
	}

	// ── COMPILE ERROR EXPECTED BELOW ──────────────────────────────────────────
	// opencodeVersion does not exist until #113 is implemented.
	if !strings.Contains(opencodeSection, opencodeVersion) {
		t.Errorf("opencode builder missing pinned version %q in download URL.\n"+
			"opencode section:\n%s", opencodeVersion, opencodeSection)
	}
	// ──────────────────────────────────────────────────────────────────────────

	// MUST NOT use dynamic version resolution (e.g., GitHub API)
	if strings.Contains(opencodeSection, "api.github.com") {
		t.Errorf("opencode builder queries GitHub API for version — must use pinned version instead.\n"+
			"opencode section:\n%s", opencodeSection)
	}
}
