package builders

// Issue #393 — Build hangs indefinitely for Python 3.9 apps
// ---------------------------------------------------------------------------
// Python 3.9 uses the bullseye Debian base image, which is EOL. Its apt
// sources still point to deb.debian.org, which may redirect or stall,
// causing BuildKit to hang indefinitely. emitEOLDebianSourcesFix() rewrites
// sources.list to use archive.debian.org for EOL codenames.

import (
	"strings"
	"testing"

	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/paths"
)

// TestEOLDebianSourcesFix_PresentInPythonBaseStage verifies that Dockerfiles
// generated for Python include the EOL sources fix in the base stage.
func TestEOLDebianSourcesFix_PresentInPythonBaseStage(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantFix bool
	}{
		{"python_3.9_bullseye_eol", "3.9", true},
		{"python_3.10_bullseye_eol", "3.10", true},
		{"python_3.11_bookworm", "3.11", true}, // still emits fix (runtime codename check)
		{"python_3.12_bookworm", "3.12", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &models.Workspace{ID: 1, Name: "test-ws", ImageName: "test:latest"}
			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
				Workspace:     ws,
				WorkspaceSpec: models.WorkspaceSpec{},
				Language:      "python",
				Version:       tt.version,
				AppPath:       "/tmp/test",
				PathConfig:    paths.New(t.TempDir()),
			})

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			// Extract base stage (before "FROM base AS dev")
			devIdx := strings.Index(dockerfile, "FROM base AS dev")
			if devIdx < 0 {
				t.Fatal("missing 'FROM base AS dev' in generated Dockerfile")
			}
			baseStage := dockerfile[:devIdx]

			hasEOLFix := strings.Contains(baseStage, "archive.debian.org")
			hasEOLComment := strings.Contains(baseStage, "EOL Debian releases")
			hasBullseye := strings.Contains(baseStage, "bullseye")

			if tt.wantFix {
				if !hasEOLFix {
					t.Errorf("Python base stage missing archive.debian.org redirect.\n"+
						"Issue #393: EOL Debian sources fix must be present to prevent build hangs.\n"+
						"Base stage (first 600 chars):\n%s", baseStage[:min(len(baseStage), 600)])
				}
				if !hasEOLComment {
					t.Errorf("Python base stage missing EOL Debian comment.\n" +
						"Issue #393: Fix should be annotated with a comment referencing EOL releases.")
				}
				if !hasBullseye {
					t.Errorf("Python base stage must list 'bullseye' as an EOL codename.\n" +
						"Issue #393: Python 3.9 uses bullseye which is EOL.")
				}
			}
		})
	}
}

// TestEOLDebianSourcesFix_NotPresentForAlpine verifies that Alpine-based
// languages (golang, nodejs, dotnet) do NOT include the EOL sources fix,
// since they don't use apt or Debian sources.
func TestEOLDebianSourcesFix_NotPresentForAlpine(t *testing.T) {
	tests := []struct {
		name     string
		language string
		version  string
	}{
		{"golang_alpine", "golang", "1.22"},
		{"nodejs_alpine", "nodejs", "20"},
		{"dotnet_alpine", "dotnet", "8.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &models.Workspace{ID: 1, Name: "test-ws", ImageName: "test:latest"}
			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
				Workspace:     ws,
				WorkspaceSpec: models.WorkspaceSpec{},
				Language:      tt.language,
				Version:       tt.version,
				AppPath:       "/tmp/test",
				PathConfig:    paths.New(t.TempDir()),
			})

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			if strings.Contains(dockerfile, "archive.debian.org") {
				t.Errorf("%s Dockerfile should NOT contain archive.debian.org — Alpine images don't use apt/Debian sources", tt.language)
			}
		})
	}
}

// TestEOLDebianSourcesFix_EOLCodenames verifies that the fix covers all known
// EOL Debian codenames: jessie, stretch, buster, and bullseye.
func TestEOLDebianSourcesFix_EOLCodenames(t *testing.T) {
	ws := &models.Workspace{ID: 1, Name: "test-ws", ImageName: "test:latest"}
	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: models.WorkspaceSpec{},
		Language:      "python",
		Version:       "3.9",
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	eolCodenames := []string{"jessie", "stretch", "buster", "bullseye"}
	for _, codename := range eolCodenames {
		if !strings.Contains(dockerfile, codename) {
			t.Errorf("EOL sources fix missing codename %q — builds on %s Debian will still hang.\n"+
				"Issue #393: All EOL codenames must be handled.", codename, codename)
		}
	}
}

// TestEOLDebianSourcesFix_BeforeFirstAptGet verifies that the EOL sources fix
// appears BEFORE the first apt-get command in the Dockerfile. If apt runs
// before the sources are fixed, the build can still hang on EOL images.
func TestEOLDebianSourcesFix_BeforeFirstAptGet(t *testing.T) {
	ws := &models.Workspace{ID: 1, Name: "test-ws", ImageName: "test:latest"}
	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: models.WorkspaceSpec{},
		Language:      "python",
		Version:       "3.9",
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	eolFixIdx := strings.Index(dockerfile, "archive.debian.org")
	aptGetIdx := strings.Index(dockerfile, "apt-get")

	if eolFixIdx < 0 {
		t.Fatal("EOL sources fix (archive.debian.org) not found in Dockerfile")
	}
	if aptGetIdx < 0 {
		t.Fatal("apt-get not found in Dockerfile")
	}

	if eolFixIdx > aptGetIdx {
		t.Errorf("EOL sources fix (pos %d) must appear BEFORE first apt-get (pos %d).\n"+
			"Issue #393: Sources must be fixed before any apt operation to prevent hangs.",
			eolFixIdx, aptGetIdx)
	}
}

// TestEOLDebianSourcesFix_SecuritySourcesReplaced verifies the fix also
// replaces security.debian.org, which can also stall on EOL releases.
func TestEOLDebianSourcesFix_SecuritySourcesReplaced(t *testing.T) {
	ws := &models.Workspace{ID: 1, Name: "test-ws", ImageName: "test:latest"}
	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: models.WorkspaceSpec{},
		Language:      "python",
		Version:       "3.9",
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if !strings.Contains(dockerfile, "security.debian.org") {
		t.Errorf("EOL sources fix should replace security.debian.org entries.\n" +
			"Issue #393: security.debian.org for EOL releases can also hang builds.")
	}
}
