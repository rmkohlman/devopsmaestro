//go:build !integration

package builders

import (
	"strings"
	"testing"

	"devopsmaestro/models"
	"devopsmaestro/utils/appkind"
	"github.com/rmkohlman/MaestroSDK/paths"
)

// TestDockerfileGenerator_CICD_NoAptGet asserts that a KindCICD Dockerfile
// contains NO apt-get, backports, python3-pip, or cargo invocations.
// These are Debian/Ubuntu toolchain artifacts that must not appear in an Alpine-based
// CICD image. Fails until generateCICDStage() is implemented (#404).
func TestDockerfileGenerator_CICD_NoAptGet(t *testing.T) {
	ws := &models.Workspace{ID: 1, Name: "cicd-ws", ImageName: "cicd:latest"}
	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: models.WorkspaceSpec{},
		AppKind:       string(appkind.KindCICD),
		AppPath:       "/tmp/cicd-test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	forbidden := []string{"apt-get", "backports", "python3-pip", "cargo"}
	for _, bad := range forbidden {
		if strings.Contains(dockerfile, bad) {
			t.Errorf("CICD Dockerfile must NOT contain %q — found in generated output", bad)
		}
	}
}

// TestDockerfileGenerator_CICD_PinnedAlpineBase asserts that the KindCICD Dockerfile
// uses a digest-pinned alpine:3.20 FROM line (via pinnedImage()).
// Fails until generateCICDStage() is implemented (#404).
func TestDockerfileGenerator_CICD_PinnedAlpineBase(t *testing.T) {
	ws := &models.Workspace{ID: 1, Name: "cicd-ws", ImageName: "cicd:latest"}
	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: models.WorkspaceSpec{},
		AppKind:       string(appkind.KindCICD),
		AppPath:       "/tmp/cicd-test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Must use digest-pinned alpine:3.20 exactly as pinnedImage() produces
	wantPrefix := "FROM alpine:3.20@sha256:"
	if !strings.Contains(dockerfile, wantPrefix) {
		t.Errorf("CICD Dockerfile must contain %q — missing or not digest-pinned", wantPrefix)
	}
}

// TestDockerfileGenerator_CICD_KubectlHelmKustomizeBuilderStages asserts that
// kubectl, helm, and kustomize are installed via pinned builder stages with
// COPY --from= directives (supply-chain policy, see security review H1 in #404).
// Fails until generateCICDStage() is implemented (#404).
func TestDockerfileGenerator_CICD_KubectlHelmKustomizeBuilderStages(t *testing.T) {
	ws := &models.Workspace{ID: 1, Name: "cicd-ws", ImageName: "cicd:latest"}
	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: models.WorkspaceSpec{},
		AppKind:       string(appkind.KindCICD),
		AppPath:       "/tmp/cicd-test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Each tool must be copied from a dedicated builder stage
	requiredCopies := []string{
		"COPY --from=kubectl-builder",
		"COPY --from=helm-builder",
		"COPY --from=kustomize-builder",
	}
	for _, want := range requiredCopies {
		if !strings.Contains(dockerfile, want) {
			t.Errorf("CICD Dockerfile missing %q — tools must use builder stage pattern", want)
		}
	}
}

// TestDockerfileGenerator_CICD_NonRootUser asserts that the KindCICD Dockerfile
// ends with a non-root USER directive (existing generateDevUser flow, refs #56, #97).
// Fails until generateCICDStage() is implemented (#404).
func TestDockerfileGenerator_CICD_NonRootUser(t *testing.T) {
	ws := &models.Workspace{ID: 1, Name: "cicd-ws", ImageName: "cicd:latest"}
	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: models.WorkspaceSpec{},
		AppKind:       string(appkind.KindCICD),
		AppPath:       "/tmp/cicd-test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Final USER must be non-root dev user
	if !strings.Contains(dockerfile, "USER dev") {
		t.Error("CICD Dockerfile must switch to non-root USER dev (refs #56, #97)")
	}
	// Must not leave USER root as final user directive
	lines := strings.Split(dockerfile, "\n")
	lastUserLine := ""
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "USER ") {
			lastUserLine = trimmed
		}
	}
	if lastUserLine != "USER dev" {
		t.Errorf("last USER directive = %q, want %q", lastUserLine, "USER dev")
	}
}

// TestDockerfileGenerator_CICD_ArgoCDConditional asserts that:
// - When ArgoCDDetected=true, the argocd CLI builder stage IS present
// - When ArgoCDDetected=false, the argocd CLI builder stage is ABSENT
// Fails until generateCICDStage() is implemented (#404).
func TestDockerfileGenerator_CICD_ArgoCDConditional(t *testing.T) {
	tests := []struct {
		name           string
		argoCDDetected bool
		wantArgoCD     bool
	}{
		{
			name:           "argocd_present_when_dot_argocd_detected",
			argoCDDetected: true,
			wantArgoCD:     true,
		},
		{
			name:           "argocd_absent_when_no_dot_argocd",
			argoCDDetected: false,
			wantArgoCD:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &models.Workspace{ID: 1, Name: "cicd-ws", ImageName: "cicd:latest"}
			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
				Workspace:      ws,
				WorkspaceSpec:  models.WorkspaceSpec{},
				AppKind:        string(appkind.KindCICD),
				AppPath:        "/tmp/cicd-test",
				PathConfig:     paths.New(t.TempDir()),
				ArgoCDDetected: tt.argoCDDetected,
			})

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			hasArgoCD := strings.Contains(dockerfile, "argocd-builder") ||
				strings.Contains(dockerfile, "argocd-linux")
			if tt.wantArgoCD && !hasArgoCD {
				t.Error("argocd builder stage should be present when .argocd/ detected")
			}
			if !tt.wantArgoCD && hasArgoCD {
				t.Error("argocd builder stage must be absent when .argocd/ not detected")
			}
		})
	}
}

// TestDockerfileGenerator_CICD_ActiveBuilderStagesGating asserts that for KindCICD,
// Neovim / Tree-sitter / Go-tools / opencode builder stages are NOT emitted,
// while Lazygit and Starship still are.
// Fails until generateCICDStage() is implemented and activeBuilderStages() is gated (#404).
func TestDockerfileGenerator_CICD_ActiveBuilderStagesGating(t *testing.T) {
	ws := &models.Workspace{ID: 1, Name: "cicd-ws", ImageName: "cicd:latest"}
	wsSpec := models.WorkspaceSpec{}
	wsSpec.Tools.Opencode = true // force opencode on to confirm it's gated out

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsSpec,
		AppKind:       string(appkind.KindCICD),
		AppPath:       "/tmp/cicd-test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	skipped := []string{
		"neovim-builder",
		"treesitter-builder",
		"go-tools-builder",
		"opencode-builder",
	}
	for _, s := range skipped {
		if strings.Contains(dockerfile, s) {
			t.Errorf("CICD Dockerfile must NOT contain %q (gated out for KindCICD)", s)
		}
	}

	required := []string{
		"lazygit-builder",
		"starship-builder",
	}
	for _, r := range required {
		if !strings.Contains(dockerfile, r) {
			t.Errorf("CICD Dockerfile must contain %q (kept for KindCICD)", r)
		}
	}
}
