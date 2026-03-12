package builders

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devopsmaestro/models"
	"devopsmaestro/pkg/nvimops/plugin"
)

func TestDockerfileGenerator_GenerateBaseStage_Python(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		wantContain []string
	}{
		{
			name:    "python default version",
			version: "",
			wantContain: []string{
				"FROM python:3.11-slim-bookworm AS base",
				"apt-get update",
				"build-essential",
			},
		},
		{
			name:    "python specific version",
			version: "3.10",
			wantContain: []string{
				"FROM python:3.10-slim-bookworm AS base",
			},
		},
		{
			name:    "python 3.12",
			version: "3.12",
			wantContain: []string{
				"FROM python:3.12-slim-bookworm AS base",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &models.Workspace{
				ID:        1,
				Name:      "test-ws",
				ImageName: "test:latest",
			}
			wsYAML := models.WorkspaceSpec{}

			gen := NewDockerfileGenerator(ws, wsYAML, "python", tt.version, "/tmp/test", "")

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			for _, want := range tt.wantContain {
				if !strings.Contains(dockerfile, want) {
					t.Errorf("Generate() missing expected content: %q", want)
				}
			}
		})
	}
}

func TestDockerfileGenerator_GenerateBaseStage_Golang(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		wantContain []string
	}{
		{
			name:    "golang default version",
			version: "",
			wantContain: []string{
				"FROM golang:1.22-alpine AS base",
				"apk add git",
			},
		},
		{
			name:    "golang specific version",
			version: "1.21",
			wantContain: []string{
				"FROM golang:1.21-alpine AS base",
			},
		},
		{
			name:    "golang 1.23",
			version: "1.23",
			wantContain: []string{
				"FROM golang:1.23-alpine AS base",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &models.Workspace{
				ID:        1,
				Name:      "test-ws",
				ImageName: "test:latest",
			}
			wsYAML := models.WorkspaceSpec{}

			gen := NewDockerfileGenerator(ws, wsYAML, "golang", tt.version, "/tmp/test", "")

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			for _, want := range tt.wantContain {
				if !strings.Contains(dockerfile, want) {
					t.Errorf("Generate() missing expected content: %q", want)
				}
			}
		})
	}
}

func TestDockerfileGenerator_GenerateBaseStage_NodeJS(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		wantContain []string
	}{
		{
			name:    "nodejs default version",
			version: "",
			wantContain: []string{
				"FROM node:18-alpine AS base",
			},
		},
		{
			name:    "nodejs specific version",
			version: "20",
			wantContain: []string{
				"FROM node:20-alpine AS base",
			},
		},
		{
			name:    "nodejs 16",
			version: "16",
			wantContain: []string{
				"FROM node:16-alpine AS base",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &models.Workspace{
				ID:        1,
				Name:      "test-ws",
				ImageName: "test:latest",
			}
			wsYAML := models.WorkspaceSpec{}

			gen := NewDockerfileGenerator(ws, wsYAML, "nodejs", tt.version, "/tmp/test", "")

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			for _, want := range tt.wantContain {
				if !strings.Contains(dockerfile, want) {
					t.Errorf("Generate() missing expected content: %q", want)
				}
			}
		})
	}
}

func TestDockerfileGenerator_GenerateBaseStage_Unknown(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{}

	gen := NewDockerfileGenerator(ws, wsYAML, "unknown", "", "/tmp/test", "")

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Should use generic Ubuntu base
	if !strings.Contains(dockerfile, "FROM ubuntu:22.04 AS base") {
		t.Errorf("Generate() missing Ubuntu base for unknown language")
	}
}

func TestDockerfileGenerator_GenerateDevStage(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{}

	gen := NewDockerfileGenerator(ws, wsYAML, "python", "3.11", "/tmp/test", "")

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Should have dev stage
	wantContain := []string{
		"FROM base AS dev",
		"USER root",
		"USER dev",
		"WORKDIR /workspace",
		"zsh",
		"neovim",
		"git",
		"curl",
	}

	for _, want := range wantContain {
		if !strings.Contains(dockerfile, want) {
			t.Errorf("Generate() dev stage missing: %q", want)
		}
	}
}

func TestDockerfileGenerator_DevStage_CustomPackages(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{
		Build: models.DevBuildConfig{
			DevStage: models.DevStageConfig{
				Packages: []string{"htop", "vim", "tmux"},
			},
		},
	}

	gen := NewDockerfileGenerator(ws, wsYAML, "python", "3.11", "/tmp/test", "")

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Should have custom packages
	wantContain := []string{"htop", "vim", "tmux"}

	for _, want := range wantContain {
		if !strings.Contains(dockerfile, want) {
			t.Errorf("Generate() missing custom package: %q", want)
		}
	}
}

func TestDockerfileGenerator_DevStage_CustomDevTools(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{
		Build: models.DevBuildConfig{
			DevStage: models.DevStageConfig{
				DevTools: []string{"ruff", "mypy"},
			},
		},
	}

	gen := NewDockerfileGenerator(ws, wsYAML, "python", "3.11", "/tmp/test", "")

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Should install custom dev tools via pip
	wantContain := []string{"pip install", "ruff", "mypy"}

	for _, want := range wantContain {
		if !strings.Contains(dockerfile, want) {
			t.Errorf("Generate() missing dev tool: %q", want)
		}
	}
}

func TestDockerfileGenerator_DevStage_GolangTools(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{
		Build: models.DevBuildConfig{
			DevStage: models.DevStageConfig{
				DevTools: []string{"gopls", "delve"},
			},
		},
	}

	gen := NewDockerfileGenerator(ws, wsYAML, "golang", "1.22", "/tmp/test", "")

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Should install golang tools via go install
	wantContain := []string{
		"go install golang.org/x/tools/gopls@latest",
		"go install github.com/go-delve/delve/cmd/dlv@latest",
	}

	for _, want := range wantContain {
		if !strings.Contains(dockerfile, want) {
			t.Errorf("Generate() missing golang tool: %q", want)
		}
	}
}

func TestDockerfileGenerator_DevUser(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}

	tests := []struct {
		name        string
		wsYAML      models.WorkspaceSpec
		wantContain []string
	}{
		{
			name:   "default user settings",
			wsYAML: models.WorkspaceSpec{},
			wantContain: []string{
				"groupadd -g 1000 dev",
				"useradd -m -u 1000 -g dev -s /bin/zsh dev",
			},
		},
		{
			name: "custom uid/gid",
			wsYAML: models.WorkspaceSpec{
				Container: models.ContainerConfig{
					UID: 501,
					GID: 501,
				},
			},
			wantContain: []string{
				"groupadd -g 501 dev",
				"useradd -m -u 501 -g dev -s /bin/zsh dev",
			},
		},
		{
			name: "custom user name",
			wsYAML: models.WorkspaceSpec{
				Container: models.ContainerConfig{
					User: "myuser",
				},
			},
			wantContain: []string{
				"groupadd -g 1000 myuser",
				"useradd -m -u 1000 -g myuser -s /bin/zsh myuser",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewDockerfileGenerator(ws, tt.wsYAML, "python", "3.11", "/tmp/test", "")

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			for _, want := range tt.wantContain {
				if !strings.Contains(dockerfile, want) {
					t.Errorf("Generate() missing: %q", want)
				}
			}
		})
	}
}

func TestDockerfileGenerator_DevUserAlpine(t *testing.T) {
	// Test that Alpine-based images (golang) use addgroup/adduser
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}

	gen := NewDockerfileGenerator(ws, models.WorkspaceSpec{}, "golang", "1.23", "/tmp/test", "")

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Alpine uses addgroup/adduser
	wantContain := []string{
		"addgroup -g 1000 dev",
		"adduser -D -u 1000 -G dev -s /bin/zsh dev",
	}
	wantNotContain := []string{
		"groupadd",
		"useradd",
	}

	for _, want := range wantContain {
		if !strings.Contains(dockerfile, want) {
			t.Errorf("Generate() missing for Alpine: %q", want)
		}
	}
	for _, notWant := range wantNotContain {
		if strings.Contains(dockerfile, notWant) {
			t.Errorf("Generate() should not contain for Alpine: %q", notWant)
		}
	}
}

func TestDockerfileGenerator_CustomWorkDir(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{
		Container: models.ContainerConfig{
			WorkingDir: "/app",
		},
	}

	gen := NewDockerfileGenerator(ws, wsYAML, "python", "3.11", "/tmp/test", "")

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if !strings.Contains(dockerfile, "WORKDIR /app") {
		t.Errorf("Generate() missing custom WORKDIR")
	}
}

func TestDockerfileGenerator_CustomCommand(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{
		Container: models.ContainerConfig{
			Command: []string{"/bin/bash", "-c", "sleep infinity"},
		},
	}

	gen := NewDockerfileGenerator(ws, wsYAML, "python", "3.11", "/tmp/test", "")

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if !strings.Contains(dockerfile, `CMD ["/bin/bash", "-c", "sleep infinity"]`) {
		t.Errorf("Generate() missing custom CMD")
	}
}

func TestDockerfileGenerator_LanguageVersionTable(t *testing.T) {
	// Table-driven test for all language/version combinations
	tests := []struct {
		language string
		version  string
		want     string
	}{
		// Python versions
		{"python", "", "python:3.11-slim-bookworm"},
		{"python", "3.9", "python:3.9-slim-bookworm"},
		{"python", "3.10", "python:3.10-slim-bookworm"},
		{"python", "3.11", "python:3.11-slim-bookworm"},
		{"python", "3.12", "python:3.12-slim-bookworm"},

		// Go versions
		{"golang", "", "golang:1.22-alpine"},
		{"golang", "1.20", "golang:1.20-alpine"},
		{"golang", "1.21", "golang:1.21-alpine"},
		{"golang", "1.22", "golang:1.22-alpine"},
		{"golang", "1.23", "golang:1.23-alpine"},

		// Node.js versions
		{"nodejs", "", "node:18-alpine"},
		{"nodejs", "16", "node:16-alpine"},
		{"nodejs", "18", "node:18-alpine"},
		{"nodejs", "20", "node:20-alpine"},
		{"nodejs", "21", "node:21-alpine"},

		// Unknown language
		{"unknown", "", "ubuntu:22.04"},
		{"rust", "", "ubuntu:22.04"},
	}

	for _, tt := range tests {
		name := tt.language
		if tt.version != "" {
			name += "-" + tt.version
		}

		t.Run(name, func(t *testing.T) {
			ws := &models.Workspace{
				ID:        1,
				Name:      "test-ws",
				ImageName: "test:latest",
			}
			wsYAML := models.WorkspaceSpec{}

			gen := NewDockerfileGenerator(ws, wsYAML, tt.language, tt.version, "/tmp/test", "")

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			if !strings.Contains(dockerfile, "FROM "+tt.want) {
				t.Errorf("Generate() expected base image %q, got dockerfile:\n%s", tt.want, dockerfile[:200])
			}
		})
	}
}

func TestNewDockerfileGenerator(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{}

	gen := NewDockerfileGenerator(ws, wsYAML, "python", "3.11", "/app/path", "/app/path/Dockerfile")

	// Type assert to access internal fields for verification
	impl := gen.(*DefaultDockerfileGenerator)

	if impl.workspace != ws {
		t.Error("workspace not set")
	}
	if impl.language != "python" {
		t.Errorf("language = %q, want %q", impl.language, "python")
	}
	if impl.version != "3.11" {
		t.Errorf("version = %q, want %q", impl.version, "3.11")
	}
	if impl.appPath != "/app/path" {
		t.Errorf("appPath = %q, want %q", impl.appPath, "/app/path")
	}
	if impl.baseDockerfile != "/app/path/Dockerfile" {
		t.Errorf("baseDockerfile = %q, want %q", impl.baseDockerfile, "/app/path/Dockerfile")
	}
}

// TestDockerfileGenerator_NvimSection_WithGitRepo tests that nvim config is found
// when the appPath matches the staging directory path (Issue #18).
// This test verifies the fix for the bug where app.Path was passed instead of sourcePath.
func TestDockerfileGenerator_NvimSection_WithGitRepo(t *testing.T) {
	// Setup: Create a temporary staging directory structure
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	// Simulate a GitRepo source path (different from app.Path)
	repoName := "test-git-repo"
	stagingDir := filepath.Join(homeDir, ".devopsmaestro", "build-staging", repoName)
	nvimConfigPath := filepath.Join(stagingDir, ".config", "nvim")

	// Create the nvim config directory to simulate generateNvimConfig having run
	if err := os.MkdirAll(nvimConfigPath, 0755); err != nil {
		t.Fatalf("failed to create nvim config dir: %v", err)
	}
	defer os.RemoveAll(stagingDir)

	// Create a dummy init.lua to make it a valid nvim config
	initLuaPath := filepath.Join(nvimConfigPath, "init.lua")
	if err := os.WriteFile(initLuaPath, []byte("-- test config"), 0644); err != nil {
		t.Fatalf("failed to create init.lua: %v", err)
	}

	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{
		Nvim: models.NvimConfig{
			Structure: "custom", // Enable nvim section
		},
	}

	// KEY TEST: Pass the sourcePath (git repo path) NOT app.Path
	// This simulates the correct behavior after the fix
	sourcePath := filepath.Join("/tmp", "dvm-clone-xyz", repoName)

	gen := NewDockerfileGenerator(ws, wsYAML, "python", "3.11", sourcePath, "")

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// The generated Dockerfile should include COPY command for nvim config
	if !strings.Contains(dockerfile, "COPY .config/nvim /home/dev/.config/nvim") {
		t.Errorf("Generate() should include nvim COPY command when sourcePath matches staging dir basename.\nGenerated Dockerfile:\n%s", dockerfile)
	}
}

// TestDockerfileGenerator_NvimSection_AppPathMismatch demonstrates the bug (Issue #18)
// where passing app.Path instead of sourcePath causes nvim config to not be found.
func TestDockerfileGenerator_NvimSection_AppPathMismatch(t *testing.T) {
	// Setup: Create a temporary staging directory structure
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	// The ACTUAL staging dir uses the git repo name
	repoName := "test-git-repo"
	stagingDir := filepath.Join(homeDir, ".devopsmaestro", "build-staging", repoName)
	nvimConfigPath := filepath.Join(stagingDir, ".config", "nvim")

	// Create the nvim config directory
	if err := os.MkdirAll(nvimConfigPath, 0755); err != nil {
		t.Fatalf("failed to create nvim config dir: %v", err)
	}
	defer os.RemoveAll(stagingDir)

	// Create a dummy init.lua
	initLuaPath := filepath.Join(nvimConfigPath, "init.lua")
	if err := os.WriteFile(initLuaPath, []byte("-- test config"), 0644); err != nil {
		t.Fatalf("failed to create init.lua: %v", err)
	}

	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{
		Nvim: models.NvimConfig{
			Structure: "custom",
		},
	}

	// BUG SCENARIO: Pass the app.Path (which is different from staging dir basename)
	// This is the buggy behavior: app.Path = "/path/to/my-app" but staging uses "test-git-repo"
	appPath := "/Users/test/apps/my-app" // Different basename than "test-git-repo"

	gen := NewDockerfileGenerator(ws, wsYAML, "python", "3.11", appPath, "")

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// With the bug, the COPY command should NOT be present because the generator
	// looks for staging dir based on "my-app" but the nvim config is in "test-git-repo"
	if strings.Contains(dockerfile, "COPY .config/nvim /home/dev/.config/nvim") {
		t.Errorf("BUG TEST: When appPath doesn't match staging dir, nvim COPY should NOT be found.\nThis test documents the bug - it should fail when appPath != sourcePath")
	}

	// It should have the skip comment instead
	if !strings.Contains(dockerfile, "Skipping Neovim configuration") {
		t.Errorf("Expected 'Skipping Neovim configuration' comment when nvim config not found")
	}
}

// TestDockerfileGenerator_BuilderStage_SetE verifies that every builder stage RUN command
// begins with `set -e` so that any failed sub-command aborts the build immediately.
// This is a Phase 2 failing test — the hardening is NOT yet implemented.
func TestDockerfileGenerator_BuilderStage_SetE(t *testing.T) {
	tests := []struct {
		name         string
		language     string
		version      string
		wantBuilders []string // stage header comments to look for
		minSetECount int      // minimum number of `set -e` occurrences expected
	}{
		{
			name:     "python/debian — neovim, lazygit, starship, treesitter builders",
			language: "python",
			version:  "3.11",
			wantBuilders: []string{
				"# --- Parallel builder: Neovim ---",
				"# --- Parallel builder: lazygit ---",
				"# --- Parallel builder: Starship prompt ---",
				"# --- Parallel builder: tree-sitter CLI ---",
			},
			// At minimum one `set -e` per builder stage (neovim, lazygit, starship, treesitter)
			minSetECount: 4,
		},
		{
			name:     "golang/alpine — lazygit (alpine path), starship, treesitter builders",
			language: "golang",
			version:  "1.22",
			wantBuilders: []string{
				"# --- Parallel builder: lazygit ---",
				"# --- Parallel builder: Starship prompt ---",
				"# --- Parallel builder: tree-sitter CLI ---",
			},
			// At minimum one `set -e` per builder stage (lazygit, starship, treesitter)
			minSetECount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &models.Workspace{
				ID:        1,
				Name:      "test-ws",
				ImageName: "test:latest",
			}
			wsYAML := models.WorkspaceSpec{}

			gen := NewDockerfileGenerator(ws, wsYAML, tt.language, tt.version, "/tmp/test", "")

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			// Verify all expected builder stage headers are present
			for _, header := range tt.wantBuilders {
				if !strings.Contains(dockerfile, header) {
					t.Errorf("Generate() missing builder stage header: %q", header)
				}
			}

			// Verify `set -e` appears at least once per builder stage
			setECount := strings.Count(dockerfile, "set -e")
			if setECount < tt.minSetECount {
				t.Errorf("Generate() has %d occurrences of 'set -e', want at least %d.\n"+
					"Every builder stage RUN command must begin with 'set -e' for fail-fast behavior.",
					setECount, tt.minSetECount)
			}

			// Verify `set -e` appears specifically within builder stage sections
			// (not just in the dev stage or elsewhere)
			neovimIdx := strings.Index(dockerfile, "# --- Parallel builder: Neovim ---")
			lazygitIdx := strings.Index(dockerfile, "# --- Parallel builder: lazygit ---")
			starshipIdx := strings.Index(dockerfile, "# --- Parallel builder: Starship prompt ---")
			treesitterIdx := strings.Index(dockerfile, "# --- Parallel builder: tree-sitter CLI ---")
			devStageIdx := strings.Index(dockerfile, "FROM base AS dev")

			type stageCheck struct {
				name  string
				start int
				end   int
			}
			var stages []stageCheck

			if neovimIdx >= 0 && lazygitIdx > neovimIdx {
				stages = append(stages, stageCheck{"neovim", neovimIdx, lazygitIdx})
			}
			if lazygitIdx >= 0 && starshipIdx > lazygitIdx {
				stages = append(stages, stageCheck{"lazygit", lazygitIdx, starshipIdx})
			}
			if starshipIdx >= 0 && treesitterIdx > starshipIdx {
				stages = append(stages, stageCheck{"starship", starshipIdx, treesitterIdx})
			}
			if treesitterIdx >= 0 && devStageIdx > treesitterIdx {
				stages = append(stages, stageCheck{"treesitter", treesitterIdx, devStageIdx})
			}

			for _, stage := range stages {
				stageContent := dockerfile[stage.start:stage.end]
				if !strings.Contains(stageContent, "set -e") {
					t.Errorf("Builder stage %q is missing 'set -e'.\n"+
						"Stage content:\n%s", stage.name, stageContent)
				}
			}
		})
	}
}

// TestDockerfileGenerator_BuilderStage_CurlHardened verifies that ALL curl commands in
// builder stages use hardened flags: -f (fail on HTTP errors) and --retry 3.
// Also verifies that OLD vulnerable curl patterns are NOT present.
// This is a Phase 2 failing test — the hardening is NOT yet implemented.
func TestDockerfileGenerator_BuilderStage_CurlHardened(t *testing.T) {
	tests := []struct {
		name             string
		language         string
		version          string
		wantCurlPatterns []string // patterns that MUST be present
		badCurlPatterns  []string // patterns that must NOT be present
	}{
		{
			name:     "python/debian — all curl commands hardened",
			language: "python",
			version:  "3.11",
			wantCurlPatterns: []string{
				// Hardened curl flags must appear in builder stages
				"curl -fsSL --retry 3 --connect-timeout 30",
			},
			badCurlPatterns: []string{
				// Old neovim download pattern
				"curl -LO",
				// Old lazygit API call pattern
				`curl -s "https://api.github.com`,
				// Old lazygit download pattern (note: -Lo is the old pattern)
				"curl -Lo lazygit",
				// Old starship pattern (pipe-to-shell)
				"curl -sS https://starship.rs",
				// Old tree-sitter pattern
				"curl -sL ",
			},
		},
		{
			name:     "golang/alpine — all curl commands hardened",
			language: "golang",
			version:  "1.22",
			wantCurlPatterns: []string{
				"curl -fsSL --retry 3 --connect-timeout 30",
			},
			badCurlPatterns: []string{
				// Old lazygit API call pattern (Alpine path)
				`curl -s "https://api.github.com`,
				// Old lazygit download pattern
				"curl -Lo lazygit",
				// Old starship pattern
				"curl -sS https://starship.rs",
				// Old tree-sitter pattern
				"curl -sL ",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &models.Workspace{
				ID:        1,
				Name:      "test-ws",
				ImageName: "test:latest",
			}
			wsYAML := models.WorkspaceSpec{}

			gen := NewDockerfileGenerator(ws, wsYAML, tt.language, tt.version, "/tmp/test", "")

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			// Verify hardened patterns ARE present
			for _, want := range tt.wantCurlPatterns {
				if !strings.Contains(dockerfile, want) {
					t.Errorf("Generate() missing hardened curl pattern: %q\n"+
						"All curl commands in builder stages must use -f and --retry 3", want)
				}
			}

			// Verify vulnerable patterns are NOT present
			for _, bad := range tt.badCurlPatterns {
				if strings.Contains(dockerfile, bad) {
					t.Errorf("Generate() contains vulnerable curl pattern: %q\n"+
						"This pattern must be replaced with hardened curl flags", bad)
				}
			}
		})
	}
}

// TestDockerfileGenerator_StarshipBuilder_NoShellPipe verifies that the starship builder
// downloads the install script to a file first instead of piping directly to sh.
// This is a Phase 2 failing test — the hardening is NOT yet implemented.
func TestDockerfileGenerator_StarshipBuilder_NoShellPipe(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	// Default shell theme triggers starship builder
	wsYAML := models.WorkspaceSpec{}

	gen := NewDockerfileGenerator(ws, wsYAML, "python", "3.11", "/tmp/test", "")

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify the starship builder section is present
	if !strings.Contains(dockerfile, "# --- Parallel builder: Starship prompt ---") {
		t.Fatalf("Generate() missing starship builder stage — test requires it to be present")
	}

	// Extract the starship builder section
	starshipStart := strings.Index(dockerfile, "# --- Parallel builder: Starship prompt ---")
	treesitterStart := strings.Index(dockerfile, "# --- Parallel builder: tree-sitter CLI ---")
	if treesitterStart <= starshipStart {
		t.Fatalf("Could not isolate starship builder section")
	}
	starshipSection := dockerfile[starshipStart:treesitterStart]

	// MUST have: download to file (not pipe)
	wantPatterns := []string{
		// Download script to a temp file
		"-o /tmp/install-starship.sh",
		// Execute from file
		"sh /tmp/install-starship.sh --yes",
		// Binary verification
		"test -x /usr/local/bin/starship",
	}
	for _, want := range wantPatterns {
		if !strings.Contains(starshipSection, want) {
			t.Errorf("Starship builder missing pattern: %q\n"+
				"Starship section:\n%s", want, starshipSection)
		}
	}

	// MUST NOT have: pipe-to-shell pattern (security risk)
	if strings.Contains(starshipSection, "| sh") {
		t.Errorf("Starship builder uses pipe-to-shell pattern (| sh) — this is insecure.\n"+
			"Must download to file first, then execute.\n"+
			"Starship section:\n%s", starshipSection)
	}
}

// TestDockerfileGenerator_LazygitBuilder_VersionValidation verifies that the lazygit builder
// validates that $LAZYGIT_VERSION is non-empty before attempting to download the binary.
// Tests BOTH Alpine and Debian paths.
// This is a Phase 2 failing test — the hardening is NOT yet implemented.
func TestDockerfileGenerator_LazygitBuilder_VersionValidation(t *testing.T) {
	tests := []struct {
		name     string
		language string
		version  string
		path     string // "alpine" or "debian"
	}{
		{
			name:     "lazygit alpine path (golang workspace)",
			language: "golang",
			version:  "1.22",
			path:     "alpine",
		},
		{
			name:     "lazygit debian path (python workspace)",
			language: "python",
			version:  "3.11",
			path:     "debian",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &models.Workspace{
				ID:        1,
				Name:      "test-ws",
				ImageName: "test:latest",
			}
			wsYAML := models.WorkspaceSpec{}

			gen := NewDockerfileGenerator(ws, wsYAML, tt.language, tt.version, "/tmp/test", "")

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			// Verify lazygit builder section is present
			if !strings.Contains(dockerfile, "# --- Parallel builder: lazygit ---") {
				t.Fatalf("Generate() missing lazygit builder stage")
			}

			// Extract the lazygit builder section
			lazygitStart := strings.Index(dockerfile, "# --- Parallel builder: lazygit ---")
			starshipStart := strings.Index(dockerfile, "# --- Parallel builder: Starship prompt ---")
			if starshipStart <= lazygitStart {
				t.Fatalf("Could not isolate lazygit builder section")
			}
			lazygitSection := dockerfile[lazygitStart:starshipStart]

			// MUST validate that LAZYGIT_VERSION is non-empty
			// Pattern: [ -n "$LAZYGIT_VERSION" ] ensures we don't try to download ""
			if !strings.Contains(lazygitSection, `[ -n "$LAZYGIT_VERSION" ]`) {
				t.Errorf("Lazygit builder (%s path) missing version validation: %q\n"+
					"Without this check, a failed API call produces a broken download URL.\n"+
					"Lazygit section:\n%s",
					tt.path, `[ -n "$LAZYGIT_VERSION" ]`, lazygitSection)
			}

			// MUST use hardened curl for the GitHub API call
			if !strings.Contains(lazygitSection, "curl -fsSL --retry 3 --connect-timeout 30") {
				t.Errorf("Lazygit builder (%s path) API call missing hardened curl flags.\n"+
					"Lazygit section:\n%s", tt.path, lazygitSection)
			}

			// MUST have binary verification
			if !strings.Contains(lazygitSection, "test -x /usr/local/bin/lazygit") {
				t.Errorf("Lazygit builder (%s path) missing binary verification: 'test -x /usr/local/bin/lazygit'\n"+
					"Lazygit section:\n%s", tt.path, lazygitSection)
			}
		})
	}
}

// TestDockerfileGenerator_BuilderStage_BinaryVerification verifies that each builder stage
// includes a `test -x /path/to/binary` check to confirm the binary was installed correctly.
// This is a Phase 2 failing test — the hardening is NOT yet implemented.
func TestDockerfileGenerator_BuilderStage_BinaryVerification(t *testing.T) {
	tests := []struct {
		name         string
		language     string
		version      string
		binaryChecks []struct {
			stageName   string // for error messages
			stageHeader string // section header to locate
			nextHeader  string // next section header (to bound the search)
			binaryPath  string // expected `test -x` path
		}
	}{
		{
			name:     "python/debian — neovim, lazygit, starship, treesitter",
			language: "python",
			version:  "3.11",
			binaryChecks: []struct {
				stageName   string
				stageHeader string
				nextHeader  string
				binaryPath  string
			}{
				{
					stageName:   "neovim",
					stageHeader: "# --- Parallel builder: Neovim ---",
					nextHeader:  "# --- Parallel builder: lazygit ---",
					binaryPath:  "test -x /opt/nvim/bin/nvim",
				},
				{
					stageName:   "lazygit",
					stageHeader: "# --- Parallel builder: lazygit ---",
					nextHeader:  "# --- Parallel builder: Starship prompt ---",
					binaryPath:  "test -x /usr/local/bin/lazygit",
				},
				{
					stageName:   "starship",
					stageHeader: "# --- Parallel builder: Starship prompt ---",
					nextHeader:  "# --- Parallel builder: tree-sitter CLI ---",
					binaryPath:  "test -x /usr/local/bin/starship",
				},
				{
					stageName:   "tree-sitter",
					stageHeader: "# --- Parallel builder: tree-sitter CLI ---",
					nextHeader:  "# Development stage",
					binaryPath:  "test -x /usr/local/bin/tree-sitter",
				},
			},
		},
		{
			name:     "golang/alpine — lazygit, starship, treesitter",
			language: "golang",
			version:  "1.22",
			binaryChecks: []struct {
				stageName   string
				stageHeader string
				nextHeader  string
				binaryPath  string
			}{
				{
					stageName:   "lazygit",
					stageHeader: "# --- Parallel builder: lazygit ---",
					nextHeader:  "# --- Parallel builder: Starship prompt ---",
					binaryPath:  "test -x /usr/local/bin/lazygit",
				},
				{
					stageName:   "starship",
					stageHeader: "# --- Parallel builder: Starship prompt ---",
					nextHeader:  "# --- Parallel builder: tree-sitter CLI ---",
					binaryPath:  "test -x /usr/local/bin/starship",
				},
				{
					stageName:   "tree-sitter",
					stageHeader: "# --- Parallel builder: tree-sitter CLI ---",
					nextHeader:  "# Development stage",
					binaryPath:  "test -x /usr/local/bin/tree-sitter",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &models.Workspace{
				ID:        1,
				Name:      "test-ws",
				ImageName: "test:latest",
			}
			wsYAML := models.WorkspaceSpec{}

			gen := NewDockerfileGenerator(ws, wsYAML, tt.language, tt.version, "/tmp/test", "")

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			for _, check := range tt.binaryChecks {
				startIdx := strings.Index(dockerfile, check.stageHeader)
				endIdx := strings.Index(dockerfile, check.nextHeader)

				if startIdx < 0 {
					t.Errorf("Missing builder stage header: %q", check.stageHeader)
					continue
				}
				if endIdx <= startIdx {
					// Fall back to checking the full dockerfile section from the header
					endIdx = len(dockerfile)
				}

				section := dockerfile[startIdx:endIdx]

				if !strings.Contains(section, check.binaryPath) {
					t.Errorf("Builder stage %q missing binary verification: %q\n"+
						"Each builder stage must verify the binary was installed correctly.\n"+
						"Stage section:\n%s",
						check.stageName, check.binaryPath, section)
				}
			}
		})
	}
}

// TestDockerfileGenerator_GolangciLint_NoShellPipe verifies that the golangci-lint install
// in the dev stage uses download-to-file, NOT pipe-to-shell.
// This is a Phase 2 failing test — the hardening is NOT yet implemented.
func TestDockerfileGenerator_GolangciLint_NoShellPipe(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	// Default golang tools include golangci-lint
	wsYAML := models.WorkspaceSpec{}

	gen := NewDockerfileGenerator(ws, wsYAML, "golang", "1.22", "/tmp/test", "")

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify golangci-lint installer is referenced somewhere in the dockerfile
	if !strings.Contains(dockerfile, "golangci-lint") {
		t.Fatalf("Generate() missing golangci-lint install — golang workspaces must include it")
	}

	// MUST have: download to file pattern
	wantPatterns := []string{
		// Download install script to temp file
		"-o /tmp/install-golangci.sh",
		// Execute from file
		"sh /tmp/install-golangci.sh",
	}
	for _, want := range wantPatterns {
		if !strings.Contains(dockerfile, want) {
			t.Errorf("golangci-lint install missing pattern: %q\n"+
				"Must download install script to file before executing (not pipe-to-shell)", want)
		}
	}

	// MUST NOT have: old pipe-to-shell pattern
	// The old pattern was: curl -sSfL ... | sh -s -- -b $(go env GOPATH)/bin
	badPatterns := []string{
		"| sh -s --",
		// Also check for the raw pipe pattern with the golangci script URL
		"golangci/golangci-lint/master/install.sh | sh",
	}
	for _, bad := range badPatterns {
		if strings.Contains(dockerfile, bad) {
			t.Errorf("golangci-lint install uses pipe-to-shell pattern: %q\n"+
				"This is a security risk — must download to file first, then execute.", bad)
		}
	}
}

// TestDockerfileGenerator_PluginManifest tests conditional Mason/Treesitter pre-install
func TestDockerfileGenerator_PluginManifest(t *testing.T) {
	// Import needed for manifest
	// "devopsmaestro/pkg/nvimops/plugin"

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	// Create staging directory
	stagingDir := filepath.Join(homeDir, ".devopsmaestro", "build-staging", "test-manifest")
	nvimConfigPath := filepath.Join(stagingDir, ".config", "nvim")
	if err := os.MkdirAll(nvimConfigPath, 0755); err != nil {
		t.Fatalf("failed to create nvim config dir: %v", err)
	}
	defer os.RemoveAll(stagingDir)

	// Create init.lua so nvim config is detected
	initLuaPath := filepath.Join(nvimConfigPath, "init.lua")
	if err := os.WriteFile(initLuaPath, []byte("-- test"), 0644); err != nil {
		t.Fatalf("failed to create init.lua: %v", err)
	}

	tests := []struct {
		name      string
		manifest  *plugin.PluginManifest
		wantMason string
		noMason   string
		wantTS    string
		noTS      string
	}{
		{
			name: "with Mason and Treesitter",
			manifest: &plugin.PluginManifest{
				Features: plugin.PluginFeatures{
					HasMason:      true,
					HasTreesitter: true,
				},
			},
			wantMason: "RUN nvim --headless -c \"MasonInstall",
			noMason:   "Mason not installed - skipping LSP pre-install",
			wantTS:    "RUN nvim --headless -c \"lua require('nvim-treesitter').install",
			noTS:      "Treesitter not installed - skipping parser pre-install",
		},
		{
			name: "without Mason or Treesitter",
			manifest: &plugin.PluginManifest{
				Features: plugin.PluginFeatures{
					HasMason:      false,
					HasTreesitter: false,
				},
			},
			wantMason: "Mason not installed - skipping LSP pre-install",
			noMason:   "RUN nvim --headless -c \"MasonInstall",
			wantTS:    "Treesitter not installed - skipping parser pre-install",
			noTS:      "RUN nvim --headless -c \"lua require('nvim-treesitter').install",
		},
		{
			name: "with Mason only",
			manifest: &plugin.PluginManifest{
				Features: plugin.PluginFeatures{
					HasMason:      true,
					HasTreesitter: false,
				},
			},
			wantMason: "RUN nvim --headless -c \"MasonInstall",
			noMason:   "Mason not installed - skipping LSP pre-install",
			wantTS:    "Treesitter not installed - skipping parser pre-install",
			noTS:      "RUN nvim --headless -c \"lua require('nvim-treesitter').install",
		},
		{
			name: "with Treesitter only",
			manifest: &plugin.PluginManifest{
				Features: plugin.PluginFeatures{
					HasMason:      false,
					HasTreesitter: true,
				},
			},
			wantMason: "Mason not installed - skipping LSP pre-install",
			noMason:   "RUN nvim --headless -c \"MasonInstall",
			wantTS:    "RUN nvim --headless -c \"lua require('nvim-treesitter').install",
			noTS:      "Treesitter not installed - skipping parser pre-install",
		},
		{
			name:      "nil manifest (backward compatibility)",
			manifest:  nil,
			wantMason: "RUN nvim --headless -c \"MasonInstall",
			noMason:   "",
			wantTS:    "RUN nvim --headless -c \"lua require('nvim-treesitter').install",
			noTS:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &models.Workspace{
				ID:        1,
				Name:      "test-ws",
				ImageName: "test:latest",
			}
			wsYAML := models.WorkspaceSpec{
				Nvim: models.NvimConfig{
					Structure: "custom",
				},
			}

			gen := NewDockerfileGenerator(ws, wsYAML, "python", "3.11", stagingDir, "")

			// Set manifest if provided
			if tt.manifest != nil {
				gen.SetPluginManifest(tt.manifest)
			}

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			// Check Mason expectations
			if tt.wantMason != "" && !strings.Contains(dockerfile, tt.wantMason) {
				t.Errorf("Missing expected Mason content: %q", tt.wantMason)
			}
			if tt.noMason != "" && strings.Contains(dockerfile, tt.noMason) {
				t.Errorf("Unexpected Mason content found: %q", tt.noMason)
			}

			// Check Treesitter expectations
			if tt.wantTS != "" && !strings.Contains(dockerfile, tt.wantTS) {
				t.Errorf("Missing expected Treesitter content: %q", tt.wantTS)
			}
			if tt.noTS != "" && strings.Contains(dockerfile, tt.noTS) {
				t.Errorf("Unexpected Treesitter content found: %q", tt.noTS)
			}

			// Ensure no "|| true" fallback on Mason/Treesitter commands (should fail fast)
			// Check line-by-line to avoid false positives from user creation commands
			lines := strings.Split(dockerfile, "\n")
			for _, line := range lines {
				if strings.Contains(line, "MasonInstall") && strings.Contains(line, "|| true") {
					t.Errorf("MasonInstall should not use '|| true' fallback: %s", line)
				}
				if strings.Contains(line, "nvim-treesitter") && strings.Contains(line, "|| true") {
					t.Errorf("nvim-treesitter install should not use '|| true' fallback: %s", line)
				}
			}
		})
	}
}
