package builders

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devopsmaestro/models"
	"devopsmaestro/pkg/nvimops/plugin"
	"devopsmaestro/pkg/paths"
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

			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
				Workspace:     ws,
				WorkspaceSpec: wsYAML,
				Language:      "python",
				Version:       tt.version,
				AppPath:       "/tmp/test",
				PathConfig:    paths.New(t.TempDir()),
			})

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

			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
				Workspace:     ws,
				WorkspaceSpec: wsYAML,
				Language:      "golang",
				Version:       tt.version,
				AppPath:       "/tmp/test",
				PathConfig:    paths.New(t.TempDir()),
			})

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

			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
				Workspace:     ws,
				WorkspaceSpec: wsYAML,
				Language:      "nodejs",
				Version:       tt.version,
				AppPath:       "/tmp/test",
				PathConfig:    paths.New(t.TempDir()),
			})

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

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "unknown",
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

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

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: "python", Version: "3.11", AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})

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

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: "python", Version: "3.11", AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})

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

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: "python", Version: "3.11", AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})

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

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: "golang", Version: "1.22", AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})

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
			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: tt.wsYAML, Language: "python", Version: "3.11", AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})

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

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: models.WorkspaceSpec{}, Language: "golang", Version: "1.23", AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})

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

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: "python", Version: "3.11", AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})

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

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: "python", Version: "3.11", AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})

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

			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: tt.language, Version: tt.version, AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})

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

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: "python", Version: "3.11", AppPath: "/app/path", BaseDockerfile: "/app/path/Dockerfile", PathConfig: paths.New(t.TempDir())})

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

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: "python", Version: "3.11", AppPath: sourcePath, PathConfig: paths.New(homeDir)})

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

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: "python", Version: "3.11", AppPath: appPath, PathConfig: paths.New(homeDir)})

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

			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: tt.language, Version: tt.version, AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})

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

			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: tt.language, Version: tt.version, AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})

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

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: "python", Version: "3.11", AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})

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

			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: tt.language, Version: tt.version, AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})

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

			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: tt.language, Version: tt.version, AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})

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

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: "golang", Version: "1.22", AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})

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

			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: "python", Version: "3.11", AppPath: stagingDir, PathConfig: paths.New(homeDir)})

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

// =============================================================================
// v0.37.5 Phase 2 Tests (RED) — Write failing tests before implementation
// =============================================================================

// TestDockerfileGenerator_TreeSitterBuilder_DynamicVersion verifies that the tree-sitter
// builder stage queries the GitHub Releases API for the latest version instead of
// hardcoding a specific version (e.g., v0.24.6).
//
// Phase 2 failing test for H3: tree-sitter dynamic versioning.
// MUST FAIL against current code (hardcodes v0.24.6).
func TestDockerfileGenerator_TreeSitterBuilder_DynamicVersion(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: "python", Version: "3.11", AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify the tree-sitter builder section exists
	if !strings.Contains(dockerfile, "# --- Parallel builder: tree-sitter CLI ---") {
		t.Fatalf("Generate() missing tree-sitter builder stage")
	}

	// Extract the tree-sitter builder section
	tsStart := strings.Index(dockerfile, "# --- Parallel builder: tree-sitter CLI ---")
	// The tree-sitter builder is the last builder stage before the dev stage
	devStageStart := strings.Index(dockerfile, "# Development stage with additional tools")
	if devStageStart <= tsStart {
		// Fall back: search for go-tools-builder or FROM base AS dev
		devStageStart = strings.Index(dockerfile, "FROM base AS dev")
	}
	var tsSection string
	if devStageStart > tsStart {
		tsSection = dockerfile[tsStart:devStageStart]
	} else {
		tsSection = dockerfile[tsStart:]
	}

	// MUST NOT contain hardcoded version
	if strings.Contains(tsSection, "v0.24.6") {
		t.Errorf("tree-sitter builder hardcodes version 'v0.24.6'.\n"+
			"Must use dynamic version from GitHub API instead.\n"+
			"tree-sitter section:\n%s", tsSection)
	}

	// MUST query GitHub API for latest version
	if !strings.Contains(tsSection, "api.github.com/repos/tree-sitter/tree-sitter/releases/latest") {
		t.Errorf("tree-sitter builder does not query GitHub API for latest version.\n"+
			"Expected: api.github.com/repos/tree-sitter/tree-sitter/releases/latest\n"+
			"tree-sitter section:\n%s", tsSection)
	}

	// MUST validate version is non-empty before using it
	if !strings.Contains(tsSection, `[ -n "$TREESITTER_VERSION" ]`) {
		t.Errorf("tree-sitter builder missing version validation: %q\n"+
			"Without this check, a failed API call produces a broken download URL.\n"+
			"tree-sitter section:\n%s",
			`[ -n "$TREESITTER_VERSION" ]`, tsSection)
	}

	// MUST use the variable ${TREESITTER_VERSION} in the download URL
	if !strings.Contains(tsSection, "${TREESITTER_VERSION}") {
		t.Errorf("tree-sitter builder download URL must use ${TREESITTER_VERSION} variable.\n"+
			"tree-sitter section:\n%s", tsSection)
	}
}

// TestDockerfileGenerator_TreeSitterBuilder_DebianPath verifies that for Debian-based
// builds (e.g., Python), the tree-sitter builder uses debian:bookworm-slim with apt-get
// and ca-certificates, while Alpine-based builds (e.g., Go) use alpine:3.20 with apk.
func TestDockerfileGenerator_TreeSitterBuilder_DebianPath(t *testing.T) {
	tests := []struct {
		name         string
		language     string
		version      string
		wantImage    string
		wantPkgMgr   string
		notWantImage string
	}{
		{
			name:         "python uses debian tree-sitter builder",
			language:     "python",
			version:      "3.11",
			wantImage:    "FROM debian:bookworm-slim AS treesitter-builder",
			wantPkgMgr:   "apt-get update && apt-get install -y --no-install-recommends curl ca-certificates sed",
			notWantImage: "FROM alpine:3.20 AS treesitter-builder",
		},
		{
			name:         "golang uses alpine tree-sitter builder",
			language:     "golang",
			version:      "1.22",
			wantImage:    "FROM alpine:3.20 AS treesitter-builder",
			wantPkgMgr:   "apk add --no-cache curl sed",
			notWantImage: "FROM debian:bookworm-slim AS treesitter-builder",
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

			// Extract tree-sitter section
			tsStart := strings.Index(dockerfile, "# --- Parallel builder: tree-sitter CLI ---")
			if tsStart < 0 {
				t.Fatalf("Generate() missing tree-sitter builder stage")
			}
			devStart := strings.Index(dockerfile[tsStart+1:], "\n# ")
			var tsSection string
			if devStart > 0 {
				tsSection = dockerfile[tsStart : tsStart+1+devStart]
			} else {
				tsSection = dockerfile[tsStart:]
			}

			if !strings.Contains(tsSection, tt.wantImage) {
				t.Errorf("tree-sitter builder should use %q for %s builds.\nGot section:\n%s", tt.wantImage, tt.language, tsSection)
			}
			if strings.Contains(tsSection, tt.notWantImage) {
				t.Errorf("tree-sitter builder should NOT use %q for %s builds.\nGot section:\n%s", tt.notWantImage, tt.language, tsSection)
			}
			if !strings.Contains(tsSection, tt.wantPkgMgr) {
				t.Errorf("tree-sitter builder should use %q for %s builds.\nGot section:\n%s", tt.wantPkgMgr, tt.language, tsSection)
			}
		})
	}
}

// TestDockerfileGenerator_Generate_NilWorkspace verifies that Generate() returns a
// non-nil error when the workspace is nil, rather than panicking.
//
// Phase 2 failing test for M3: nil workspace guard.
// MUST FAIL against current code (panics at g.workspace.ImageName in isAlpineImage()).
func TestDockerfileGenerator_Generate_NilWorkspace(t *testing.T) {
	wsYAML := models.WorkspaceSpec{}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: nil, WorkspaceSpec: wsYAML, Language: "python", Version: "3.11", AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})

	// Recover from panic so the test can report failure rather than crashing the suite
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Generate() panicked with nil workspace instead of returning an error: %v", r)
		}
	}()

	_, err := gen.Generate()
	if err == nil {
		t.Error("Generate() with nil workspace should return a non-nil error, got nil")
	}
}

// TestDockerfileGenerator_BuilderStageConsistency verifies that every builder stage
// emitted in the Dockerfile has a corresponding COPY --from= directive, and vice versa.
//
// This is a regression guard for M4's refactor (should PASS against current code).
// It also documents the invariant that must hold after the activeBuilderStages() refactor.
func TestDockerfileGenerator_BuilderStageConsistency(t *testing.T) {
	tests := []struct {
		name     string
		language string
		version  string
	}{
		{name: "python/debian", language: "python", version: "3.11"},
		{name: "golang/alpine", language: "golang", version: "1.22"},
	}

	// Known stage-name to FROM-stage mapping
	builderHeaders := map[string]string{
		"# --- Parallel builder: Neovim ---":          "neovim-builder",
		"# --- Parallel builder: lazygit ---":         "lazygit-builder",
		"# --- Parallel builder: Starship prompt ---": "starship-builder",
		"# --- Parallel builder: tree-sitter CLI ---": "treesitter-builder",
		"# --- Parallel builder: Go tools ---":        "go-tools-builder",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &models.Workspace{
				ID:        1,
				Name:      "test-ws",
				ImageName: "test:latest",
			}
			wsYAML := models.WorkspaceSpec{}

			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: tt.language, Version: tt.version, AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			// For every builder header present, there must be a COPY --from= for that stage
			for header, stageName := range builderHeaders {
				hasHeader := strings.Contains(dockerfile, header)
				copyDirective := "COPY --from=" + stageName
				hasCopy := strings.Contains(dockerfile, copyDirective)

				if hasHeader && !hasCopy {
					t.Errorf("Builder stage %q is emitted but has no corresponding %q directive.\n"+
						"Builders and COPY --from= directives must always be in sync.",
						header, copyDirective)
				}
				if !hasHeader && hasCopy {
					t.Errorf("COPY directive %q is present but the corresponding builder stage %q is missing.\n"+
						"Builders and COPY --from= directives must always be in sync.",
						copyDirective, header)
				}
			}
		})
	}
}

// TestDockerfileGenerator_ArchDetection_NoSilentFallback verifies that the else branch
// in every builder's architecture detection block fails explicitly (echo ERROR + exit 1)
// rather than silently falling back to x86_64.
//
// Phase 2 failing test for L2: remove silent arch fallback.
// MUST FAIL against current code (silently falls back to x86_64 in else branches).
func TestDockerfileGenerator_ArchDetection_NoSilentFallback(t *testing.T) {
	tests := []struct {
		name          string
		language      string
		version       string
		builderStages []struct {
			header     string
			nextHeader string
		}
	}{
		{
			name:     "python/debian — neovim, lazygit, tree-sitter builders",
			language: "python",
			version:  "3.11",
			builderStages: []struct {
				header     string
				nextHeader string
			}{
				{
					header:     "# --- Parallel builder: Neovim ---",
					nextHeader: "# --- Parallel builder: lazygit ---",
				},
				{
					header:     "# --- Parallel builder: lazygit ---",
					nextHeader: "# --- Parallel builder: Starship prompt ---",
				},
				{
					header:     "# --- Parallel builder: tree-sitter CLI ---",
					nextHeader: "# Development stage with additional tools",
				},
			},
		},
		{
			name:     "golang/alpine — lazygit, tree-sitter builders",
			language: "golang",
			version:  "1.22",
			builderStages: []struct {
				header     string
				nextHeader string
			}{
				{
					header:     "# --- Parallel builder: lazygit ---",
					nextHeader: "# --- Parallel builder: Starship prompt ---",
				},
				{
					header:     "# --- Parallel builder: tree-sitter CLI ---",
					nextHeader: "# Development stage with additional tools",
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

			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: tt.language, Version: tt.version, AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			for _, stage := range tt.builderStages {
				startIdx := strings.Index(dockerfile, stage.header)
				if startIdx < 0 {
					t.Errorf("Missing builder stage: %q", stage.header)
					continue
				}

				endIdx := strings.Index(dockerfile, stage.nextHeader)
				var section string
				if endIdx > startIdx {
					section = dockerfile[startIdx:endIdx]
				} else {
					// Fall back to rest of file
					section = dockerfile[startIdx:]
				}

				// The else branch must NOT silently set an arch variable to a default
				// Silent fallback patterns look like: else \n    NVIM_ARCH="nvim-linux-x86_64"; \
				silentFallbackPatterns := []string{
					// Neovim builder silent fallback
					"else \\\n        NVIM_ARCH=\"nvim-linux-x86_64\"",
					// Lazygit builder silent fallback (Alpine-style single-line)
					"else \\\n        LG_ARCH=\"x86_64\"",
					// Lazygit builder silent fallback (Debian-style)
					"else \\\n        LG_ARCH=\"x86_64\"",
					// Tree-sitter builder silent fallback (inline style)
					"else TS_ARCH=\"x64\"",
				}

				for _, silentPattern := range silentFallbackPatterns {
					if strings.Contains(section, silentPattern) {
						t.Errorf("Builder stage %q has silent arch fallback: %q\n"+
							"The else branch must fail explicitly with 'echo ERROR' and 'exit 1'.\n"+
							"Stage section:\n%s",
							stage.header, silentPattern, section)
					}
				}

				// The else branch MUST contain explicit failure
				// Check that if there's any arch detection (if/elif/else), the else has exit 1
				if strings.Contains(section, "if [ \"$ARCH\"") || strings.Contains(section, "if [ \"$ARCH\" =") {
					if !strings.Contains(section, "exit 1") {
						t.Errorf("Builder stage %q has arch detection but no 'exit 1' in else branch.\n"+
							"Unknown architectures must fail explicitly, not silently fall back.\n"+
							"Stage section:\n%s",
							stage.header, section)
					}
					if !strings.Contains(section, "echo \"ERROR") && !strings.Contains(section, "echo \"Unsupported") {
						t.Errorf("Builder stage %q has arch detection but no error message in else branch.\n"+
							"Unknown architectures must print an error message before failing.\n"+
							"Stage section:\n%s",
							stage.header, section)
					}
				}
			}
		})
	}
}

// TestDockerfileGenerator_LazygitBuilder_UnifiedDownload verifies that the lazygit builder
// produces consistent download logic across both Alpine and Debian paths — same URL pattern,
// same version validation, same binary verification.
//
// This is a regression guard for L1's refactor (should PASS against current code).
// The test ensures L1's unification of the two lazygit builder paths doesn't break anything.
func TestDockerfileGenerator_LazygitBuilder_UnifiedDownload(t *testing.T) {
	tests := []struct {
		name     string
		language string
		version  string
	}{
		{name: "alpine path (golang workspace)", language: "golang", version: "1.22"},
		{name: "debian path (python workspace)", language: "python", version: "3.11"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &models.Workspace{
				ID:        1,
				Name:      "test-ws",
				ImageName: "test:latest",
			}
			wsYAML := models.WorkspaceSpec{}

			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: tt.language, Version: tt.version, AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			// Extract lazygit builder section
			lazygitStart := strings.Index(dockerfile, "# --- Parallel builder: lazygit ---")
			starshipStart := strings.Index(dockerfile, "# --- Parallel builder: Starship prompt ---")
			if lazygitStart < 0 {
				t.Fatalf("Missing lazygit builder stage")
			}
			if starshipStart <= lazygitStart {
				t.Fatalf("Could not isolate lazygit builder section")
			}
			lazygitSection := dockerfile[lazygitStart:starshipStart]

			// Both paths MUST query the GitHub Releases API for version
			if !strings.Contains(lazygitSection, "api.github.com/repos/jesseduffield/lazygit/releases/latest") {
				t.Errorf("[%s] lazygit builder missing GitHub API query for version.\n"+
					"Section:\n%s", tt.name, lazygitSection)
			}

			// Both paths MUST validate the version is non-empty
			if !strings.Contains(lazygitSection, `[ -n "$LAZYGIT_VERSION" ]`) {
				t.Errorf("[%s] lazygit builder missing version validation: %q\n"+
					"Section:\n%s", tt.name, `[ -n "$LAZYGIT_VERSION" ]`, lazygitSection)
			}

			// Both paths MUST use the standard lazygit download URL pattern
			if !strings.Contains(lazygitSection, "jesseduffield/lazygit/releases/latest/download/lazygit_") {
				t.Errorf("[%s] lazygit builder missing standard download URL.\n"+
					"Section:\n%s", tt.name, lazygitSection)
			}

			// Both paths MUST verify the binary
			if !strings.Contains(lazygitSection, "test -x /usr/local/bin/lazygit") {
				t.Errorf("[%s] lazygit builder missing binary verification.\n"+
					"Section:\n%s", tt.name, lazygitSection)
			}

			// Both paths MUST use hardened curl flags
			if !strings.Contains(lazygitSection, "curl -fsSL --retry 3 --connect-timeout 30") {
				t.Errorf("[%s] lazygit builder missing hardened curl flags.\n"+
					"Section:\n%s", tt.name, lazygitSection)
			}
		})
	}
}

// TestDockerfileGenerator_NvimSection_CustomUser verifies that when Container.User is set
// to a custom username, the nvim COPY and chown commands use that username instead of
// the hardcoded "dev" username.
//
// Phase 2 failing test for A3: nvim section respects configured user.
// MUST FAIL against current code (hardcodes /home/dev/ on line 732).
func TestDockerfileGenerator_NvimSection_CustomUser(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	// Use a unique workspace name for the staging directory
	repoName := "test-custom-user-ws"
	stagingDir := filepath.Join(homeDir, ".devopsmaestro", "build-staging", repoName)
	nvimConfigPath := filepath.Join(stagingDir, ".config", "nvim")

	if err := os.MkdirAll(nvimConfigPath, 0755); err != nil {
		t.Fatalf("failed to create nvim config dir: %v", err)
	}
	defer os.RemoveAll(stagingDir)

	// Create a dummy init.lua to make the nvim config directory detected
	initLuaPath := filepath.Join(nvimConfigPath, "init.lua")
	if err := os.WriteFile(initLuaPath, []byte("-- test config for custom user"), 0644); err != nil {
		t.Fatalf("failed to create init.lua: %v", err)
	}

	tests := []struct {
		name           string
		containerUser  string
		wantUser       string
		wantNotContain string
	}{
		{
			name:           "custom user myuser",
			containerUser:  "myuser",
			wantUser:       "myuser",
			wantNotContain: "/home/dev/",
		},
		{
			name:           "custom user appdev",
			containerUser:  "appdev",
			wantUser:       "appdev",
			wantNotContain: "/home/dev/",
		},
		{
			name:           "explicit dev user",
			containerUser:  "dev",
			wantUser:       "dev",
			wantNotContain: "", // No negative assertion needed for default
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
				Container: models.ContainerConfig{
					User: tt.containerUser,
				},
			}

			// Pass the sourcePath whose basename matches the staging dir name
			sourcePath := filepath.Join("/tmp", "dvm-clone-xyz", repoName)

			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: "python", Version: "3.11", AppPath: sourcePath, PathConfig: paths.New(homeDir)})

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			// Verify the nvim COPY uses the correct user's home directory
			wantCopyLine := "COPY .config/nvim /home/" + tt.wantUser + "/.config/nvim"
			if !strings.Contains(dockerfile, wantCopyLine) {
				t.Errorf("[%s] nvim COPY should use /home/%s/ but it doesn't.\n"+
					"Want: %q\n"+
					"Generated Dockerfile (nvim section):\n%s",
					tt.name, tt.wantUser, wantCopyLine,
					extractNvimSection(dockerfile))
			}

			// Verify the chown uses the correct user
			wantChownLine := "RUN chown -R " + tt.wantUser + ":" + tt.wantUser + " /home/" + tt.wantUser + "/.config"
			if !strings.Contains(dockerfile, wantChownLine) {
				t.Errorf("[%s] nvim chown should reference /home/%s/ but it doesn't.\n"+
					"Want: %q\n"+
					"Generated Dockerfile (nvim section):\n%s",
					tt.name, tt.wantUser, wantChownLine,
					extractNvimSection(dockerfile))
			}

			// For non-default users, verify we don't have the hardcoded /home/dev/ path
			if tt.wantNotContain != "" {
				// Only check in the nvim section (the dev user section itself uses "dev" literally)
				nvimSec := extractNvimSection(dockerfile)
				if strings.Contains(nvimSec, tt.wantNotContain) {
					t.Errorf("[%s] nvim section hardcodes %q instead of using configured user %q.\n"+
						"Nvim section:\n%s",
						tt.name, tt.wantNotContain, tt.wantUser, nvimSec)
				}
			}
		})
	}
}

// extractNvimSection returns the portion of the Dockerfile between the
// "# Copy Neovim configuration" comment and "USER root" (end of nvim section).
// Returns the full dockerfile as fallback if the section is not found.
func extractNvimSection(dockerfile string) string {
	start := strings.Index(dockerfile, "# Copy Neovim configuration")
	if start < 0 {
		return dockerfile
	}
	// Find the last USER root after the nvim section starts
	end := strings.Index(dockerfile[start:], "USER root\n\n")
	if end < 0 {
		return dockerfile[start:]
	}
	return dockerfile[start : start+end+len("USER root\n\n")]
}

// NOTE: L4 (cache mount comment fix) is a documentation-only change to the source comment
// in generateBuilderStages(). Comments are not testable, so no test is written for L4.
// The fix updates the comment to note that go-tools-builder IS an exception that uses
// cache mounts (--mount=type=cache), which is already verified by existing tests.

// =============================================================================
// v0.38.0 RED-phase tests: Dockerfile Generator Purity & Robustness
// These tests are written to FAIL against the current implementation, exposing
// bugs and gaps that the v0.38.0 sprint items (D2, A1, A2, A4) will fix.
// =============================================================================

// TestIsAlpine_ComputedPerLanguage verifies that the isAlpine computed field
// (set in generateBaseStage) correctly reflects the base image chosen for each
// language. This is a regression guard replacing the former D2 test that expected
// Debian behavior for golang workspaces — golang always uses Alpine base images,
// so isAlpine=true is correct behavior, not a bug.
func TestIsAlpine_ComputedPerLanguage(t *testing.T) {
	tests := []struct {
		language     string
		version      string
		expectAlpine bool
		alpineMarker string // command that MUST be present when alpine
		debianMarker string // command that MUST be present when debian
		alpineAbsent string // command that must NOT be present when alpine
		debianAbsent string // command that must NOT be present when debian
	}{
		{
			language:     "golang",
			version:      "1.22",
			expectAlpine: true,
			alpineMarker: "apk add",
			alpineAbsent: "apt-get",
		},
		{
			language:     "python",
			version:      "3.11",
			expectAlpine: false,
			debianMarker: "apt-get",
			debianAbsent: "apk add",
		},
		{
			language:     "nodejs",
			version:      "18",
			expectAlpine: true,
			alpineMarker: "apk add",
			alpineAbsent: "apt-get",
		},
		{
			language:     "", // default/ubuntu
			version:      "",
			expectAlpine: false,
			debianMarker: "apt-get",
			debianAbsent: "apk add",
		},
	}

	for _, tt := range tests {
		name := tt.language
		if name == "" {
			name = "default"
		}
		t.Run(name, func(t *testing.T) {
			ws := &models.Workspace{
				ID:        1,
				Name:      "test-ws",
				ImageName: "test:latest",
			}
			wsYAML := models.WorkspaceSpec{}

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

			// Extract dev stage to check package manager commands
			devStageIdx := strings.Index(dockerfile, "FROM base AS dev")
			if devStageIdx < 0 {
				t.Fatalf("missing dev stage in generated Dockerfile")
			}
			devStage := dockerfile[devStageIdx:]

			if tt.expectAlpine {
				if tt.alpineMarker != "" && !strings.Contains(devStage, tt.alpineMarker) {
					t.Errorf("language=%q: expected Alpine marker %q in dev stage", tt.language, tt.alpineMarker)
				}
				if tt.alpineAbsent != "" && strings.Contains(devStage, tt.alpineAbsent) {
					t.Errorf("language=%q: Alpine dev stage should not contain %q", tt.language, tt.alpineAbsent)
				}
			} else {
				if tt.debianMarker != "" && !strings.Contains(devStage, tt.debianMarker) {
					t.Errorf("language=%q: expected Debian marker %q in dev stage", tt.language, tt.debianMarker)
				}
				if tt.debianAbsent != "" && strings.Contains(devStage, tt.debianAbsent) {
					t.Errorf("language=%q: Debian dev stage should not contain %q", tt.language, tt.debianAbsent)
				}
			}

			// Verify the computed field via type assertion
			impl := gen.(*DefaultDockerfileGenerator)
			if impl.isAlpine != tt.expectAlpine {
				t.Errorf("language=%q: isAlpine=%v, want %v", tt.language, impl.isAlpine, tt.expectAlpine)
			}
		})
	}
}

// TestIsAlpine_FieldMatchesGeneratedImage verifies end-to-end consistency: the generated
// base image (FROM line) and the package manager commands in the dev stage must agree.
// If the FROM line uses an Alpine image tag, the dev stage must use apk; if it uses a
// Debian/Ubuntu image, the dev stage must use apt-get. This is a stronger invariant than
// checking the language alone.
func TestIsAlpine_FieldMatchesGeneratedImage(t *testing.T) {
	tests := []struct {
		language string
		version  string
	}{
		{"golang", "1.22"},
		{"python", "3.11"},
		{"nodejs", "18"},
		{"", ""},     // default/ubuntu
		{"rust", ""}, // unknown → ubuntu
	}

	for _, tt := range tests {
		name := tt.language
		if name == "" {
			name = "default"
		}
		t.Run(name, func(t *testing.T) {
			ws := &models.Workspace{
				ID:        1,
				Name:      "test-ws",
				ImageName: "test:latest",
			}
			wsYAML := models.WorkspaceSpec{}

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

			// Determine if the FROM line is Alpine-based
			hasAlpineBase := strings.Contains(dockerfile, "-alpine AS base")

			// Extract dev stage
			devStageIdx := strings.Index(dockerfile, "FROM base AS dev")
			if devStageIdx < 0 {
				t.Fatalf("missing dev stage")
			}
			devStage := dockerfile[devStageIdx:]

			hasApk := strings.Contains(devStage, "apk add")
			hasAptGet := strings.Contains(devStage, "apt-get")

			if hasAlpineBase {
				if !hasApk {
					t.Errorf("FROM *-alpine but dev stage missing 'apk add'")
				}
				if hasAptGet {
					t.Errorf("FROM *-alpine but dev stage has 'apt-get' (inconsistent)")
				}
			} else {
				if !hasAptGet {
					t.Errorf("FROM non-Alpine base but dev stage missing 'apt-get'")
				}
				if hasApk {
					t.Errorf("FROM non-Alpine base but dev stage has 'apk add' (inconsistent)")
				}
			}

			// Also verify user creation commands are consistent
			hasAdduser := strings.Contains(devStage, "adduser -D")
			hasUseradd := strings.Contains(devStage, "useradd -m")

			if hasAlpineBase {
				if !hasAdduser {
					t.Errorf("FROM *-alpine but dev stage missing Alpine 'adduser -D'")
				}
				if hasUseradd {
					t.Errorf("FROM *-alpine but dev stage has Debian 'useradd' (inconsistent)")
				}
			} else {
				if !hasUseradd {
					t.Errorf("FROM non-Alpine base but dev stage missing Debian 'useradd -m'")
				}
				if hasAdduser {
					t.Errorf("FROM non-Alpine base but dev stage has Alpine 'adduser -D' (inconsistent)")
				}
			}
		})
	}
}

// TestEffectiveGoVersion_PythonGenerator_ExposesMissingUnification exposes A1 gap:
// effectiveGoVersion() has no language awareness — it returns the Go default "1.22"
// even when called on a python generator with no version set. After A1, a unified
// effectiveVersion() will return language-appropriate defaults (python→3.11, etc.).
func TestEffectiveGoVersion_PythonGenerator_ExposesMissingUnification(t *testing.T) {
	ws := &models.Workspace{ID: 1, Name: "test-ws", ImageName: "test:latest"}
	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: models.WorkspaceSpec{}, Language: "python", AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})
	impl := gen.(*DefaultDockerfileGenerator)

	// effectiveGoVersion() is language-agnostic — returns Go default for any language.
	// After A1, there will be a unified effectiveVersion() that returns "3.11" for python.
	// For now, effectiveGoVersion() returns "1.22" even for python — that's the bug.
	got := impl.effectiveGoVersion()
	want := "3.11"
	if got != want {
		t.Errorf("effectiveGoVersion() on python generator = %q, want %q.\n"+
			"A1 fix: implement effectiveVersion() that returns language-appropriate defaults:\n"+
			"  python → 3.11, golang → 1.22, nodejs → 18, unknown → \"\"",
			got, want)
	}
}

// TestActiveBuilderStages_GolangDebian_IncludesNeovimBuilder exposes D2+A4 interaction:
// For a golang workspace with a Debian ImageName, neovim-builder should be included
// (Debian needs compiled neovim from GitHub releases). Currently FAILS because
// isAlpineImage() hardcodes golang=Alpine, excluding the neovim-builder stage.
func TestActiveBuilderStages_GolangDebian_IncludesNeovimBuilder(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "golang:1.22-bookworm", // Debian
	}
	wsYAML := models.WorkspaceSpec{}
	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: "golang", Version: "1.22", AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})
	impl := gen.(*DefaultDockerfileGenerator)

	stages := impl.activeBuilderStages()

	// With D2 fix: Debian golang image → isAlpine=false → neovim-builder included
	// Expected stages: neovim-builder, lazygit-builder, starship-builder, treesitter-builder, go-tools-builder = 5
	stageNames := make([]string, len(stages))
	for i, s := range stages {
		stageNames[i] = s.name
	}

	hasNeovim := false
	for _, name := range stageNames {
		if name == "neovim-builder" {
			hasNeovim = true
			break
		}
	}

	if !hasNeovim {
		t.Errorf("activeBuilderStages() for golang:1.22-bookworm (Debian) should include neovim-builder.\n"+
			"After D2: isAlpine must be computed from ImageName, not hardcoded from language.\n"+
			"Got stages: %v", stageNames)
	}
}

// =============================================================================
// D1 Tests: generateNvimSection() error propagation
// =============================================================================

// TestGenerate_NvimConfig_GracefulSkip verifies that when no nvim staging directory
// exists (the normal case for a fresh workspace), Generate() succeeds and includes the
// skip comment. This is a regression guard confirming the D1 error-return refactor
// didn't break the happy path.
func TestGenerate_NvimConfig_GracefulSkip(t *testing.T) {
	// Use a temp dir as home — no nvim staging dir will exist
	tmpHome := t.TempDir()

	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{
		Nvim: models.NvimConfig{
			Structure: "custom", // Enable nvim section (not "none")
		},
	}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       "/tmp/test-app",
		PathConfig:    paths.New(tmpHome),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() should succeed when nvim config is missing (graceful skip), got error: %v", err)
	}

	// Must contain the skip comment
	if !strings.Contains(dockerfile, "Skipping Neovim configuration (no config generated)") {
		t.Errorf("Generate() should include skip comment when no nvim staging dir exists.\n"+
			"Generated Dockerfile:\n%s", dockerfile)
	}

	// Must NOT contain nvim COPY directive
	if strings.Contains(dockerfile, "COPY .config/nvim") {
		t.Errorf("Generate() should not include nvim COPY when staging dir doesn't exist")
	}
}

// TestGenerateNvimSection_UsesPathConfig verifies that generateNvimSection() uses the
// injected PathConfig to locate the nvim staging directory and generates the correct
// COPY directive when the nvim config exists. This validates D3 (PathConfig injection)
// end-to-end through the nvim section.
func TestGenerateNvimSection_UsesPathConfig(t *testing.T) {
	// Create a temp dir that acts as home, with a real nvim config inside
	tmpHome := t.TempDir()
	appName := "my-test-app"

	// Create the nvim config structure that generateNvimSection looks for
	pc := paths.New(tmpHome)
	stagingDir := pc.BuildStagingDir(appName)
	nvimConfigPath := filepath.Join(stagingDir, ".config", "nvim")
	if err := os.MkdirAll(nvimConfigPath, 0755); err != nil {
		t.Fatalf("failed to create nvim config dir: %v", err)
	}

	// Create a dummy init.lua so the directory is detected as valid nvim config
	initLuaPath := filepath.Join(nvimConfigPath, "init.lua")
	if err := os.WriteFile(initLuaPath, []byte("-- test nvim config"), 0644); err != nil {
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

	// appPath basename must match the staging dir name
	appPath := filepath.Join("/tmp", "some-clone-dir", appName)

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       appPath,
		PathConfig:    pc,
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// The output must include the COPY directive for the nvim config
	if !strings.Contains(dockerfile, "COPY .config/nvim /home/dev/.config/nvim") {
		t.Errorf("Generate() should include nvim COPY when PathConfig points to existing config.\n"+
			"Expected: COPY .config/nvim /home/dev/.config/nvim\n"+
			"Generated Dockerfile:\n%s", dockerfile)
	}

	// Must NOT contain the skip comment
	if strings.Contains(dockerfile, "Skipping Neovim configuration") {
		t.Errorf("Generate() should not skip nvim when config exists at PathConfig staging dir")
	}

	// Verify chown uses the default user
	if !strings.Contains(dockerfile, "RUN chown -R dev:dev /home/dev/.config") {
		t.Errorf("Generate() should include chown for nvim config")
	}
}

// TestDockerfileGenerator_PythonPrivateRepos verifies that the Dockerfile generator
// correctly handles all GitURLType scenarios (https, ssh, mixed, none) for Python workspaces.
// This is the regression test for the v0.38.1 bug where the NeedsSSH flag incorrectly
// shadowed the HTTPS token-substitution path in the if/else if/else chain.
func TestDockerfileGenerator_PythonPrivateRepos(t *testing.T) {
	tests := []struct {
		name                string
		requirementsContent string
		wantContain         []string
		wantNotContain      []string
	}{
		{
			name: "HTTPS-only — pip expands build args natively",
			requirementsContent: "flask==2.3.0\n" +
				"git+https://${GITHUB_USERNAME}:${GITHUB_PAT}@github.com/Org/repo.git@v1.0\n",
			wantContain: []string{
				"COPY requirements.txt /tmp/",
				"pip install -r /tmp/requirements.txt",
				"ARG GITHUB_USERNAME",
				"ARG GITHUB_PAT",
				"pip expands ${VAR} from build args",
			},
			wantNotContain: []string{
				"--mount=type=ssh",
				"requirements-template.txt",
				"sed \"s/",
			},
		},
		{
			name:                "SSH-only — SSH mount path",
			requirementsContent: "mylib @ git+ssh://git@github.com/Org/repo.git@v1.0\n",
			wantContain: []string{
				"--mount=type=ssh",
				"ssh-keyscan",
				"openssh-client",
			},
			wantNotContain: []string{
				"requirements-template.txt",
				"sed \"s/",
			},
		},
		{
			// Mixed case: pip natively expands ${VAR} from build args (declared after FROM),
			// and SSH mount provides key-based auth for git+ssh:// URLs.
			name: "Mixed HTTPS+SSH — pip expands build args with SSH mount",
			requirementsContent: "git+https://${GITHUB_USERNAME}:${GITHUB_PAT}@github.com/Org/private-lib.git@v1.0\n" +
				"mylib @ git+ssh://git@github.com/Org/repo.git@v2.0\n",
			wantContain: []string{
				"--mount=type=ssh",
				"GITHUB_USERNAME",
				"ssh-keyscan",
				"COPY requirements.txt /tmp/",
				"pip install -r /tmp/requirements.txt",
				"pip expands ${VAR} from build args",
			},
			wantNotContain: []string{
				"requirements-template.txt",
				"sed \"s/",
			},
		},
		{
			name:                "No private repos — plain pip install",
			requirementsContent: "flask==2.3.0\nrequests>=2.28\n",
			wantContain: []string{
				"COPY requirements.txt /tmp/",
				"pip install -r /tmp/requirements.txt",
			},
			wantNotContain: []string{
				"--mount=type=ssh",
				"requirements-template.txt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appPath := t.TempDir()

			// Write requirements.txt so DetectPrivateRepos can scan it
			reqPath := filepath.Join(appPath, "requirements.txt")
			if err := os.WriteFile(reqPath, []byte(tt.requirementsContent), 0644); err != nil {
				t.Fatalf("failed to write requirements.txt: %v", err)
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
				AppPath:       appPath,
				PathConfig:    paths.New(t.TempDir()),
			})

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			for _, want := range tt.wantContain {
				if !strings.Contains(dockerfile, want) {
					t.Errorf("Generate() missing expected content: %q\nDockerfile:\n%s", want, dockerfile)
				}
			}

			for _, notWant := range tt.wantNotContain {
				if strings.Contains(dockerfile, notWant) {
					t.Errorf("Generate() contains unexpected content: %q\nDockerfile:\n%s", notWant, dockerfile)
				}
			}
		})
	}
}
