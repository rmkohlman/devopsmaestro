package builders

// =============================================================================
// Issue #113 — TDD Phase 2: activeBuilderStages() opencode pipeline tests
//
// RED: These tests WILL NOT COMPILE until the following are implemented:
//
//   - WorkspaceSpec.Tools field (models package)
//   - models.ToolsConfig struct with Opencode bool field
//   - Opencode conditional in activeBuilderStages() (builders package)
//   - generateOpencodeBuilder() method on DefaultDockerfileGenerator
//
// Once the production code is added, all tests in this file MUST pass.
// =============================================================================

import (
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/paths"
	"strings"
	"testing"
)

// TestActiveBuilderStages_Opencode_IncludedWhenEnabled verifies that
// activeBuilderStages() includes "opencode-builder" when Tools.Opencode = true.
//
// RED: WILL NOT COMPILE — WorkspaceSpec.Tools / ToolsConfig do not exist yet.
func TestActiveBuilderStages_Opencode_IncludedWhenEnabled(t *testing.T) {
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

	impl := gen.(*DefaultDockerfileGenerator)
	stages := impl.activeBuilderStages()

	var stageNames []string
	for _, s := range stages {
		stageNames = append(stageNames, s.name)
	}

	found := false
	for _, name := range stageNames {
		if name == "opencode-builder" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("activeBuilderStages() should include 'opencode-builder' when Tools.Opencode=true.\n"+
			"Got stages: %v", stageNames)
	}
}

// TestActiveBuilderStages_Opencode_ExcludedWhenDisabled verifies that
// activeBuilderStages() excludes "opencode-builder" when Tools.Opencode = false.
// Opencode is opt-in — it must NOT be included by default.
//
// RED: WILL NOT COMPILE — WorkspaceSpec.Tools / ToolsConfig do not exist yet.
func TestActiveBuilderStages_Opencode_ExcludedWhenDisabled(t *testing.T) {
	tests := []struct {
		name   string
		wsYAML models.WorkspaceSpec
	}{
		{
			name:   "zero value — no tools configured",
			wsYAML: models.WorkspaceSpec{},
		},
		{
			// ── COMPILE ERROR EXPECTED BELOW ──────────────────────────────
			// models.ToolsConfig / WorkspaceSpec.Tools do not exist until #113.
			name: "explicit opencode false",
			wsYAML: models.WorkspaceSpec{
				Tools: models.ToolsConfig{
					Opencode: false,
				},
			},
			// ──────────────────────────────────────────────────────────────
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &models.Workspace{
				ID:        1,
				Name:      "test-ws",
				ImageName: "test:latest",
			}

			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
				Workspace:     ws,
				WorkspaceSpec: tt.wsYAML,
				Language:      "python",
				Version:       "3.11",
				AppPath:       "/tmp/test",
				PathConfig:    paths.New(t.TempDir()),
			})

			impl := gen.(*DefaultDockerfileGenerator)
			stages := impl.activeBuilderStages()

			for _, s := range stages {
				if s.name == "opencode-builder" {
					t.Errorf("activeBuilderStages() should NOT include 'opencode-builder' when Tools.Opencode=false/unset.\n"+
						"opencode is opt-in — it must never appear unless explicitly enabled.\n"+
						"Got stages: %v", func() []string {
						var names []string
						for _, stage := range stages {
							names = append(names, stage.name)
						}
						return names
					}())
				}
			}
		})
	}
}

// TestActiveBuilderStages_Opencode_CopyLinePresent verifies that the opencode
// builderStage includes the correct COPY --from directive.
//
// RED: WILL NOT COMPILE — WorkspaceSpec.Tools / ToolsConfig do not exist yet.
func TestActiveBuilderStages_Opencode_CopyLinePresent(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	// ── COMPILE ERROR EXPECTED BELOW ──────────────────────────────────────────
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

	impl := gen.(*DefaultDockerfileGenerator)
	stages := impl.activeBuilderStages()

	for _, s := range stages {
		if s.name == "opencode-builder" {
			// Must have a COPY --from=opencode-builder line targeting /usr/local/bin/opencode
			wantCopy := "COPY --from=opencode-builder /usr/local/bin/opencode /usr/local/bin/opencode"
			found := false
			for _, line := range s.copyLines {
				if line == wantCopy {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("opencode builderStage.copyLines missing expected COPY directive.\n"+
					"Want: %q\n"+
					"Got:  %v", wantCopy, s.copyLines)
			}
			return
		}
	}
	t.Fatalf("opencode-builder stage not found in activeBuilderStages() output (Tools.Opencode=true)")
}

// TestGenerate_Opencode_COPYDirectivePresent verifies that the full Generate()
// output contains the COPY --from=opencode-builder directive in the dev stage
// when Tools.Opencode = true.
//
// RED: WILL NOT COMPILE — WorkspaceSpec.Tools / ToolsConfig do not exist yet.
func TestGenerate_Opencode_COPYDirectivePresent(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	// ── COMPILE ERROR EXPECTED BELOW ──────────────────────────────────────────
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

	wantCopy := "COPY --from=opencode-builder /usr/local/bin/opencode /usr/local/bin/opencode"
	if !strings.Contains(dockerfile, wantCopy) {
		t.Errorf("Generate() missing COPY directive for opencode.\n"+
			"Want: %q\n"+
			"When Tools.Opencode=true the dev stage must COPY the binary from the builder.", wantCopy)
	}
}

// TestGenerate_Opencode_COPYDirectiveAbsent verifies that the full Generate()
// output does NOT contain any opencode COPY directive when Tools.Opencode = false.
//
// RED: WILL NOT COMPILE — WorkspaceSpec.Tools / ToolsConfig do not exist yet.
func TestGenerate_Opencode_COPYDirectiveAbsent(t *testing.T) {
	tests := []struct {
		name     string
		language string
		version  string
		wsYAML   models.WorkspaceSpec
	}{
		{
			name:     "python/debian — opencode disabled",
			language: "python",
			version:  "3.11",
			wsYAML:   models.WorkspaceSpec{}, // zero value = opt-in disabled
		},
		{
			name:     "golang/alpine — opencode disabled",
			language: "golang",
			version:  "1.22",
			wsYAML:   models.WorkspaceSpec{},
		},
		{
			// ── COMPILE ERROR EXPECTED BELOW ──────────────────────────────
			name:     "python/debian — explicit opencode false",
			language: "python",
			version:  "3.11",
			wsYAML: models.WorkspaceSpec{
				Tools: models.ToolsConfig{Opencode: false},
			},
			// ──────────────────────────────────────────────────────────────
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &models.Workspace{
				ID:        1,
				Name:      "test-ws",
				ImageName: "test:latest",
			}

			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
				Workspace:     ws,
				WorkspaceSpec: tt.wsYAML,
				Language:      tt.language,
				Version:       tt.version,
				AppPath:       "/tmp/test",
				PathConfig:    paths.New(t.TempDir()),
			})

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			badPatterns := []string{
				"opencode-builder",
				"opencode",
			}
			for _, bad := range badPatterns {
				if strings.Contains(dockerfile, bad) {
					t.Errorf("Generate() should NOT contain %q when Tools.Opencode=false/unset.\n"+
						"opencode is opt-in — must not appear unless explicitly enabled.\n"+
						"Found in Dockerfile for test %q", bad, tt.name)
				}
			}
		})
	}
}

// TestGenerate_Opencode_StageConsistency verifies that when opencode is enabled,
// the builder stage and its COPY directive are both present (stage consistency
// invariant, same as TestDockerfileGenerator_BuilderStageConsistency for others).
//
// RED: WILL NOT COMPILE — WorkspaceSpec.Tools / ToolsConfig do not exist yet.
func TestGenerate_Opencode_StageConsistency(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	// ── COMPILE ERROR EXPECTED BELOW ──────────────────────────────────────────
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

	// Both must be present together
	hasBuilderStage := strings.Contains(dockerfile, "AS opencode-builder")
	hasCopyDirective := strings.Contains(dockerfile, "COPY --from=opencode-builder /usr/local/bin/opencode /usr/local/bin/opencode")

	if hasBuilderStage && !hasCopyDirective {
		t.Error("opencode-builder stage emitted but missing COPY --from=opencode-builder directive in dev stage")
	}
	if !hasBuilderStage && hasCopyDirective {
		t.Error("COPY --from=opencode-builder found but opencode-builder stage is missing")
	}
	if !hasBuilderStage && !hasCopyDirective {
		t.Error("Neither opencode-builder stage nor COPY directive found — stage was not emitted despite Tools.Opencode=true")
	}
}

// TestGenerate_Opencode_BothLanguages verifies opencode builder works for both
// Alpine (golang) and Debian (python) base images.
//
// RED: WILL NOT COMPILE — WorkspaceSpec.Tools / ToolsConfig do not exist yet.
func TestGenerate_Opencode_BothLanguages(t *testing.T) {
	tests := []struct {
		name     string
		language string
		version  string
	}{
		{name: "python/debian", language: "python", version: "3.11"},
		{name: "golang/alpine", language: "golang", version: "1.22"},
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

			// Both languages must include the opencode builder and COPY directive
			wantPatterns := []string{
				"# --- Parallel builder: opencode ---",
				"AS opencode-builder",
				"COPY --from=opencode-builder /usr/local/bin/opencode /usr/local/bin/opencode",
			}
			for _, want := range wantPatterns {
				if !strings.Contains(dockerfile, want) {
					t.Errorf("Generate() [%s] missing expected pattern: %q", tt.name, want)
				}
			}
		})
	}
}
