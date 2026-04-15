package builders

import (
	"strings"
	"testing"

	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/paths"
)

// Issue #354 — APT timeout resilience
// ---------------------------------------------------------------------------
// When the squid proxy is unreachable, APT's default 60s timeout causes builds
// to appear hung (40–50 minutes for 6 workspaces). These tests verify that
// generated Dockerfiles include timeout reduction and proxy health checks.

// TestAPTTimeoutConfig_PresentInDevStage verifies that Debian-based Dockerfiles
// include APT timeout configuration in the dev stage to reduce the default 60s
// timeout to 10s with retries enabled.
func TestAPTTimeoutConfig_PresentInDevStage(t *testing.T) {
	tests := []struct {
		name     string
		language string
		version  string
		wantAPT  bool // true for Debian, false for Alpine
	}{
		{"python/debian", "python", "3.11", true},
		{"kotlin/debian", "kotlin", "21", true},
		{"scala/debian", "scala", "21", true},
		{"golang/alpine", "golang", "1.22", false},
		{"nodejs/alpine", "nodejs", "20", false},
		{"dotnet/alpine", "dotnet", "8.0", false},
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

			// Extract dev stage
			devStageIdx := strings.Index(dockerfile, "FROM base AS dev")
			if devStageIdx < 0 {
				t.Fatal("missing 'FROM base AS dev' in generated Dockerfile")
			}
			devStage := dockerfile[devStageIdx:]

			hasTimeoutConfig := strings.Contains(devStage, "99-dvm-timeout")
			hasHTTPTimeout := strings.Contains(devStage, `Acquire::http::Timeout "10"`)
			hasHTTPSTimeout := strings.Contains(devStage, `Acquire::https::Timeout "10"`)
			hasRetries := strings.Contains(devStage, `Acquire::Retries "2"`)

			if tt.wantAPT {
				if !hasTimeoutConfig {
					t.Errorf("Debian dev stage missing APT timeout config file (99-dvm-timeout).\n"+
						"Issue #354: APT timeout must be reduced from 60s to 10s.\n"+
						"Dev stage:\n%s", devStage[:min(len(devStage), 500)])
				}
				if !hasHTTPTimeout {
					t.Errorf("Debian dev stage missing Acquire::http::Timeout \"10\"")
				}
				if !hasHTTPSTimeout {
					t.Errorf("Debian dev stage missing Acquire::https::Timeout \"10\"")
				}
				if !hasRetries {
					t.Errorf("Debian dev stage missing Acquire::Retries \"2\"")
				}
			} else {
				if hasTimeoutConfig {
					t.Errorf("Alpine dev stage should NOT have APT timeout config.\n"+
						"Dev stage:\n%s", devStage[:min(len(devStage), 500)])
				}
			}
		})
	}
}

// TestProxyHealthCheck_PresentInDevStage verifies that Debian-based Dockerfiles
// include a proxy health check before the first APT operation in the dev stage.
// When proxy is unreachable, this sets APT to use DIRECT access instead of
// waiting for timeouts on every operation.
func TestProxyHealthCheck_PresentInDevStage(t *testing.T) {
	tests := []struct {
		name     string
		language string
		version  string
		wantAPT  bool
	}{
		{"python/debian", "python", "3.11", true},
		{"kotlin/debian", "kotlin", "21", true},
		{"golang/alpine", "golang", "1.22", false},
		{"nodejs/alpine", "nodejs", "20", false},
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

			// Extract dev stage
			devStageIdx := strings.Index(dockerfile, "FROM base AS dev")
			if devStageIdx < 0 {
				t.Fatal("missing 'FROM base AS dev' in generated Dockerfile")
			}
			devStage := dockerfile[devStageIdx:]

			hasProxyCheck := strings.Contains(devStage, "http_proxy")
			hasDirectFallback := strings.Contains(devStage, "DIRECT")
			hasHealthComment := strings.Contains(devStage, "Proxy health check")

			if tt.wantAPT {
				if !hasProxyCheck {
					t.Errorf("Debian dev stage missing proxy health check.\n"+
						"Issue #354: Must test proxy reachability before APT operations.\n"+
						"Dev stage:\n%s", devStage[:min(len(devStage), 500)])
				}
				if !hasDirectFallback {
					t.Errorf("Debian dev stage missing DIRECT fallback for unreachable proxy.\n" +
						"Issue #354: When proxy is unreachable, APT should use direct access.")
				}
				if !hasHealthComment {
					t.Errorf("Debian dev stage missing proxy health check comment")
				}
			} else {
				if hasHealthComment {
					t.Errorf("Alpine dev stage should NOT have proxy health check.\n"+
						"Dev stage:\n%s", devStage[:min(len(devStage), 500)])
				}
			}
		})
	}
}

// TestAPTTimeoutConfig_BeforeFirstAptGet verifies that the APT timeout
// configuration appears BEFORE the first apt-get command in the dev stage.
// This ensures the timeout reduction is active for all APT operations.
func TestAPTTimeoutConfig_BeforeFirstAptGet(t *testing.T) {
	ws := &models.Workspace{ID: 1, Name: "test-ws", ImageName: "test:latest"}
	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: models.WorkspaceSpec{},
		Language:      "python",
		Version:       "3.11",
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Extract dev stage only
	devStageIdx := strings.Index(dockerfile, "FROM base AS dev")
	if devStageIdx < 0 {
		t.Fatal("missing 'FROM base AS dev'")
	}
	devStage := dockerfile[devStageIdx:]

	timeoutIdx := strings.Index(devStage, "99-dvm-timeout")
	aptGetIdx := strings.Index(devStage, "apt-get update")

	if timeoutIdx < 0 {
		t.Fatal("missing APT timeout config in dev stage")
	}
	if aptGetIdx < 0 {
		t.Fatal("missing apt-get update in dev stage")
	}

	if timeoutIdx > aptGetIdx {
		t.Errorf("APT timeout config (pos %d) must appear BEFORE first apt-get update (pos %d) in dev stage.\n"+
			"Issue #354: Timeout reduction must be active before any APT operations.",
			timeoutIdx, aptGetIdx)
	}
}

// TestProxyHealthCheck_BeforeFirstAptGet verifies that the proxy health check
// appears BEFORE the first apt-get command in the dev stage.
func TestProxyHealthCheck_BeforeFirstAptGet(t *testing.T) {
	ws := &models.Workspace{ID: 1, Name: "test-ws", ImageName: "test:latest"}
	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: models.WorkspaceSpec{},
		Language:      "python",
		Version:       "3.11",
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	devStageIdx := strings.Index(dockerfile, "FROM base AS dev")
	if devStageIdx < 0 {
		t.Fatal("missing 'FROM base AS dev'")
	}
	devStage := dockerfile[devStageIdx:]

	healthCheckIdx := strings.Index(devStage, "Proxy health check")
	aptGetIdx := strings.Index(devStage, "apt-get update")

	if healthCheckIdx < 0 {
		t.Fatal("missing proxy health check in dev stage")
	}
	if aptGetIdx < 0 {
		t.Fatal("missing apt-get update in dev stage")
	}

	if healthCheckIdx > aptGetIdx {
		t.Errorf("Proxy health check (pos %d) must appear BEFORE first apt-get update (pos %d) in dev stage.\n"+
			"Issue #354: Proxy must be tested before APT operations to allow direct fallback.",
			healthCheckIdx, aptGetIdx)
	}
}
