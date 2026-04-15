package builders

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devopsmaestro/models"
	"devopsmaestro/utils"
	"github.com/rmkohlman/MaestroNvim/nvimops/plugin"
	"github.com/rmkohlman/MaestroSDK/paths"
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
				"FROM python:3.11-slim@sha256:",
				"AS base",
				"apt-get update",
				"build-essential",
			},
		},
		{
			name:    "python specific version",
			version: "3.10",
			wantContain: []string{
				"FROM python:3.10-slim",
			},
		},
		{
			name:    "python 3.12",
			version: "3.12",
			wantContain: []string{
				"FROM python:3.12-slim@sha256:",
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
				"FROM golang:1.22-alpine@sha256:",
				"AS base",
				"apk add git",
			},
		},
		{
			name:    "golang specific version",
			version: "1.21",
			wantContain: []string{
				"FROM golang:1.21-alpine",
			},
		},
		{
			name:    "golang 1.23",
			version: "1.23",
			wantContain: []string{
				"FROM golang:1.23-alpine@sha256:",
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
				"FROM node:20-alpine@sha256:",
			},
		},
		{
			name:    "nodejs specific version",
			version: "20",
			wantContain: []string{
				"FROM node:20-alpine@sha256:",
			},
		},
		{
			name:    "nodejs 16",
			version: "16",
			wantContain: []string{
				"FROM node:16-alpine",
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

func TestDockerfileGenerator_GenerateBaseStage_Dotnet(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		wantContain []string
	}{
		{
			name:    "dotnet default version",
			version: "",
			wantContain: []string{
				"mcr.microsoft.com/dotnet/sdk:9.0-alpine",
				"AS base",
				"apk add git ca-certificates",
				"dotnet restore",
			},
		},
		{
			name:    "dotnet specific version",
			version: "8.0",
			wantContain: []string{
				"mcr.microsoft.com/dotnet/sdk:8.0-alpine",
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
				Language:      "dotnet",
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
	if !strings.Contains(dockerfile, "FROM ubuntu:22.04@sha256:") {
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
		{"python", "", "python:3.11-slim"},
		{"python", "3.9", "python:3.9-slim"},
		{"python", "3.9.9", "python:3.9.9-slim"},
		{"python", "3.10", "python:3.10-slim"},
		{"python", "3.11", "python:3.11-slim"},
		{"python", "3.12", "python:3.12-slim"},
		{"python", "3.13", "python:3.13-slim"},

		// Go versions
		{"golang", "", "golang:1.22-alpine"},
		{"golang", "1.20", "golang:1.20-alpine"},
		{"golang", "1.21", "golang:1.21-alpine"},
		{"golang", "1.22", "golang:1.22-alpine"},
		{"golang", "1.23", "golang:1.23-alpine"},

		// Node.js versions
		{"nodejs", "", "node:20-alpine"},
		{"nodejs", "16", "node:16-alpine"},
		{"nodejs", "18", "node:18-alpine"},
		{"nodejs", "20", "node:20-alpine"},
		{"nodejs", "21", "node:21-alpine"},

		// Elixir versions
		{"elixir", "", "elixir:1.18-slim"},
		{"elixir", "1.16", "elixir:1.16-slim"},
		{"elixir", "1.17", "elixir:1.17-slim"},
		{"elixir", "1.18", "elixir:1.18-slim"},

		// Swift versions
		{"swift", "", "swift:6.1-slim"},
		{"swift", "5.10", "swift:5.10-slim"},
		{"swift", "6.0", "swift:6.0-slim"},
		{"swift", "6.1", "swift:6.1-slim"},

		// Dart versions
		{"dart", "", "dart:3.7"},
		{"dart", "3.6", "dart:3.6"},
		{"dart", "3.7", "dart:3.7"},

		// Haskell versions
		{"haskell", "", "haskell:9.12-slim"},
		{"haskell", "9.10", "haskell:9.10-slim"},
		{"haskell", "9.8", "haskell:9.8-slim"},

		// Perl versions
		{"perl", "", "perl:5.40-slim"},
		{"perl", "5.38", "perl:5.38-slim"},
		{"perl", "5.40", "perl:5.40-slim"},

		// Ruby versions
		{"ruby", "", "ruby:3.3-slim"},
		{"ruby", "3.2", "ruby:3.2-slim"},
		{"ruby", "3.3", "ruby:3.3-slim"},
		{"ruby", "3.4", "ruby:3.4-slim"},

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

	// MUST have: direct binary download with checksum verification (not install script)
	wantPatterns := []string{
		// Download binary tarball directly
		"-o /tmp/starship.tar.gz",
		// SHA256 checksum verification
		"sha256sum -c -",
		// Binary extraction
		"tar -C /usr/local/bin -xzf /tmp/starship.tar.gz starship",
		// Binary verification
		"test -x /usr/local/bin/starship",
		// Pinned version (not install script)
		"starship/releases/download/v",
	}
	for _, want := range wantPatterns {
		if !strings.Contains(starshipSection, want) {
			t.Errorf("Starship builder missing pattern: %q\n"+
				"Starship section:\n%s", want, starshipSection)
		}
	}

	// MUST NOT have: install script pattern (security risk - downloads and executes remote script)
	badPatterns := []string{
		"install-starship.sh",
		"starship.rs/install.sh",
	}
	for _, bad := range badPatterns {
		if strings.Contains(starshipSection, bad) {
			t.Errorf("Starship builder uses install script pattern: %q — this is insecure.\n"+
				"Must download pre-built binary directly with checksum verification.\n"+
				"Starship section:\n%s", bad, starshipSection)
		}
	}
}

// TestDockerfileGenerator_LazygitBuilder_VersionValidation verifies that the lazygit builder
// uses pinned versions with SHA256 checksum verification instead of dynamic version fetching.
// Tests BOTH Alpine and Debian paths.
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

			// MUST have pinned version in download URL (not dynamic from GitHub API)
			if !strings.Contains(lazygitSection, "lazygit/releases/download/v"+lazygitVersion) {
				t.Errorf("Lazygit builder (%s path) missing pinned version %q in download URL.\n"+
					"Lazygit section:\n%s",
					tt.path, lazygitVersion, lazygitSection)
			}

			// MUST have SHA256 checksum verification
			if !strings.Contains(lazygitSection, "sha256sum -c -") {
				t.Errorf("Lazygit builder (%s path) missing SHA256 checksum verification.\n"+
					"Lazygit section:\n%s", tt.path, lazygitSection)
			}

			// MUST use hardened curl flags
			if !strings.Contains(lazygitSection, "curl -fsSL --retry 3 --connect-timeout 30") {
				t.Errorf("Lazygit builder (%s path) missing hardened curl flags.\n"+
					"Lazygit section:\n%s", tt.path, lazygitSection)
			}

			// MUST NOT use dynamic version from GitHub API (pinned versions are more secure)
			if strings.Contains(lazygitSection, "api.github.com") {
				t.Errorf("Lazygit builder (%s path) uses dynamic version from GitHub API.\n"+
					"Should use pinned version with checksum verification instead.\n"+
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
					binaryPath:  "test -x /opt/nvim/usr/bin/nvim",
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

	// MUST have: direct binary download with checksum verification pattern
	wantPatterns := []string{
		// Direct binary download from GitHub releases
		"golangci/golangci-lint/releases/download/v",
		// SHA256 checksum verification
		"sha256sum -c -",
		// Hardened curl flags
		"curl -fsSL --retry 3 --connect-timeout 30",
		// Install to GOPATH
		"$(go env GOPATH)/bin/",
	}
	for _, want := range wantPatterns {
		if !strings.Contains(dockerfile, want) {
			t.Errorf("golangci-lint install missing pattern: %q\n"+
				"Must download binary directly with checksum verification (not pipe-to-shell)", want)
		}
	}

	// MUST NOT have: old pipe-to-shell pattern
	badPatterns := []string{
		"| sh -s --",
		// Also check for the raw pipe pattern with the golangci install script
		"golangci/golangci-lint/master/install.sh",
		// Must not download an install script at all
		"install-golangci.sh",
	}
	for _, bad := range badPatterns {
		if strings.Contains(dockerfile, bad) {
			t.Errorf("golangci-lint install uses pipe-to-shell or install-script pattern: %q\n"+
				"This is a security risk — must download binary directly with checksum verification.", bad)
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
			wantMason: "mason-registry",
			noMason:   "Mason not installed - skipping LSP pre-install",
			wantTS:    "runtimepath:prepend(ts_path)",
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
			noMason:   "mason-registry",
			wantTS:    "Treesitter not installed - skipping parser pre-install",
			noTS:      "runtimepath:prepend(ts_path)",
		},
		{
			name: "with Mason only",
			manifest: &plugin.PluginManifest{
				Features: plugin.PluginFeatures{
					HasMason:      true,
					HasTreesitter: false,
				},
			},
			wantMason: "mason-registry",
			noMason:   "Mason not installed - skipping LSP pre-install",
			wantTS:    "Treesitter not installed - skipping parser pre-install",
			noTS:      "runtimepath:prepend(ts_path)",
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
			noMason:   "mason-registry",
			wantTS:    "runtimepath:prepend(ts_path)",
			noTS:      "Treesitter not installed - skipping parser pre-install",
		},
		{
			name:      "nil manifest (backward compatibility)",
			manifest:  nil,
			wantMason: "mason-registry",
			noMason:   "",
			wantTS:    "runtimepath:prepend(ts_path)",
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
				if strings.Contains(line, "mason-install.lua") && strings.Contains(line, "|| true") {
					t.Errorf("Mason install should not use '|| true' fallback: %s", line)
				}
				if strings.Contains(line, "treesitter-install.lua") && strings.Contains(line, "|| true") {
					t.Errorf("Treesitter install should not use '|| true' fallback: %s", line)
				}
			}
		})
	}
}

// =============================================================================
// v0.37.5 Phase 2 Tests (RED) — Write failing tests before implementation
// =============================================================================

// TestDockerfileGenerator_TreeSitterBuilder_DynamicVersion verifies that the tree-sitter
// builder stage uses cargo to build from source with a pinned version instead of
// downloading pre-built binaries (which had GLIBC version mismatches — see #334).
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
	devStageStart := strings.Index(dockerfile, "# Development stage with additional tools")
	if devStageStart <= tsStart {
		devStageStart = strings.Index(dockerfile, "FROM base AS dev")
	}
	var tsSection string
	if devStageStart > tsStart {
		tsSection = dockerfile[tsStart:devStageStart]
	} else {
		tsSection = dockerfile[tsStart:]
	}

	// MUST use cargo install with pinned version
	if !strings.Contains(tsSection, "cargo install tree-sitter-cli@"+treeSitterVersion) {
		t.Errorf("tree-sitter builder must use 'cargo install tree-sitter-cli@%s'.\n"+
			"tree-sitter section:\n%s", treeSitterVersion, tsSection)
	}

	// MUST use a Rust base image
	if !strings.Contains(tsSection, "FROM rust:") {
		t.Errorf("tree-sitter builder must use a Rust base image.\n"+
			"tree-sitter section:\n%s", tsSection)
	}

	// MUST NOT download pre-built binaries (the old approach — see #334)
	if strings.Contains(tsSection, "tree-sitter/releases/download") {
		t.Errorf("tree-sitter builder should NOT download pre-built binaries.\n"+
			"Must build from source via cargo to avoid GLIBC mismatches.\n"+
			"tree-sitter section:\n%s", tsSection)
	}

	// MUST NOT use sha256sum (no binary checksums needed when building from source)
	if strings.Contains(tsSection, "sha256sum") {
		t.Errorf("tree-sitter builder should NOT use sha256sum (builds from source now).\n"+
			"tree-sitter section:\n%s", tsSection)
	}

	// MUST NOT query GitHub API dynamically
	if strings.Contains(tsSection, "api.github.com") {
		t.Errorf("tree-sitter builder queries GitHub API for version at build time.\n"+
			"Should use pinned version via cargo install instead.\n"+
			"tree-sitter section:\n%s", tsSection)
	}
}

// TestDockerfileGenerator_TreeSitterBuilder_DebianPath verifies that for Debian-based
// builds (e.g., Python), the tree-sitter builder uses rust:1-slim-bookworm, while
// Alpine-based builds (e.g., Go) use rust:1-alpine3.20. Both build from source via cargo
// to avoid GLIBC version mismatches with pre-built binaries (see #334).
func TestDockerfileGenerator_TreeSitterBuilder_DebianPath(t *testing.T) {
	tests := []struct {
		name         string
		language     string
		version      string
		wantImage    string
		wantCargo    string
		notWantImage string
	}{
		{
			name:         "python uses debian rust tree-sitter builder",
			language:     "python",
			version:      "3.11",
			wantImage:    "FROM rust:1-slim-bookworm@sha256:",
			wantCargo:    "cargo install tree-sitter-cli@",
			notWantImage: "FROM rust:1-alpine3.20",
		},
		{
			name:         "golang uses alpine rust tree-sitter builder",
			language:     "golang",
			version:      "1.22",
			wantImage:    "FROM rust:1-alpine3.20@sha256:",
			wantCargo:    "cargo install tree-sitter-cli@",
			notWantImage: "FROM rust:1-slim-bookworm",
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
			if !strings.Contains(tsSection, tt.wantCargo) {
				t.Errorf("tree-sitter builder should use %q for %s builds.\nGot section:\n%s", tt.wantCargo, tt.language, tsSection)
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

			// Both paths MUST use pinned version in download URL (not dynamic from GitHub API)
			if !strings.Contains(lazygitSection, "lazygit/releases/download/v"+lazygitVersion) {
				t.Errorf("[%s] lazygit builder missing pinned version %q in download URL.\n"+
					"Section:\n%s", tt.name, lazygitVersion, lazygitSection)
			}

			// Both paths MUST have SHA256 checksum verification
			if !strings.Contains(lazygitSection, "sha256sum -c -") {
				t.Errorf("[%s] lazygit builder missing SHA256 checksum verification.\n"+
					"Section:\n%s", tt.name, lazygitSection)
			}

			// Both paths MUST NOT query GitHub API dynamically (pinned versions are more secure)
			if strings.Contains(lazygitSection, "api.github.com") {
				t.Errorf("[%s] lazygit builder queries GitHub API for version at build time.\n"+
					"Should use pinned version with checksum verification instead.\n"+
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
			language:     "ruby",
			version:      "3.3",
			expectAlpine: false,
			debianMarker: "apt-get",
			debianAbsent: "apk add",
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
		{"ruby", "3.3"},
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
			// After digest pinning, the format is: FROM image:tag-alpine@sha256:... AS base
			hasAlpineBase := strings.Contains(dockerfile, "-alpine AS base") ||
				strings.Contains(dockerfile, "-alpine@sha256:")

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

// =============================================================================
// v0.44.0 RED-phase tests: Container Neovim Environment Fixes
// Tests are written to FAIL against the current implementation.
// These drive implementation of WI-1, WI-2, and WI-4.
// =============================================================================

// TestGenerateDevStage_DebianNodeSource verifies that for a Debian-based workspace
// (e.g., Python) with nvim enabled, Node.js is installed via NodeSource setup_22.x
// rather than being included in the merged apt-get install line.
//
// WI-1: Use NodeSource for Node.js on Debian (not the Debian packaged nodejs/npm).
// MUST FAIL against current code (currently adds nodejs/npm to merged apt-get install).
func TestGenerateDevStage_DebianNodeSource(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	// nvim enabled (not "none") triggers Mason toolchain installation
	wsYAML := models.WorkspaceSpec{
		Nvim: models.NvimConfig{
			Structure: "custom",
		},
	}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python", // Debian-based
		Version:       "3.11",
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// MUST: Use NodeSource setup_22.x for modern Node.js
	if !strings.Contains(dockerfile, "nodesource.com/setup_22.x") {
		t.Errorf("Debian workspace should use NodeSource setup_22.x for Node.js install.\n"+
			"WI-1: Replace 'apt-get install nodejs npm' with NodeSource.\n"+
			"Expected to find: nodesource.com/setup_22.x\n"+
			"Generated Dockerfile:\n%s", dockerfile)
	}

	// MUST: Install nodejs from the NodeSource repo in its own step
	if !strings.Contains(dockerfile, "apt-get install -y --no-install-recommends nodejs") {
		t.Errorf("Debian workspace should install nodejs (from NodeSource) via a dedicated apt-get step.\n"+
			"WI-1: After running the NodeSource setup script, install nodejs via apt-get.\n"+
			"Generated Dockerfile:\n%s", dockerfile)
	}

	// Extract the merged apt-get install block (the one that merges dev packages + nvim deps)
	// This is the block that should NOT include nodejs or npm
	mergedAptIdx := strings.Index(dockerfile, "apt-get install -y --no-install-recommends --fix-broken")
	if mergedAptIdx < 0 {
		t.Fatalf("Could not locate merged apt-get install block")
	}
	// Find the end of that specific apt-get block (newline after last package)
	blockEnd := strings.Index(dockerfile[mergedAptIdx:], "\n\n")
	var mergedAptBlock string
	if blockEnd > 0 {
		mergedAptBlock = dockerfile[mergedAptIdx : mergedAptIdx+blockEnd]
	} else {
		mergedAptBlock = dockerfile[mergedAptIdx:]
	}

	// MUST NOT: Include nodejs in the merged apt-get block (it comes from NodeSource separately)
	if strings.Contains(mergedAptBlock, "\n    nodejs") || strings.HasPrefix(mergedAptBlock, "    nodejs") {
		t.Errorf("Merged apt-get install block should NOT include 'nodejs'.\n"+
			"WI-1: nodejs must be installed via NodeSource (separate step), not merged apt-get.\n"+
			"Merged apt-get block:\n%s", mergedAptBlock)
	}

	// MUST NOT: Include npm in the merged apt-get block (NodeSource nodejs includes npm)
	if strings.Contains(mergedAptBlock, "\n    npm") || strings.HasPrefix(mergedAptBlock, "    npm") {
		t.Errorf("Merged apt-get install block should NOT include 'npm'.\n"+
			"WI-1: npm comes bundled with NodeSource nodejs, not installed separately.\n"+
			"Merged apt-get block:\n%s", mergedAptBlock)
	}
}

// TestGenerateDevStage_AlpineNodeUnchanged verifies that for an Alpine-based workspace
// (e.g., Golang), Node.js and npm are still installed via apk (unchanged behavior).
// NodeSource is Linux-distro specific and should NOT be used for Alpine.
//
// WI-1: Alpine path unchanged — apk add nodejs npm still used, no NodeSource.
// Should PASS against current code (Alpine behavior is not changing).
func TestGenerateDevStage_AlpineNodeUnchanged(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	// nvim enabled triggers Mason toolchain install (nodejs/npm via apk)
	wsYAML := models.WorkspaceSpec{
		Nvim: models.NvimConfig{
			Structure: "custom",
		},
	}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "golang", // Alpine-based
		Version:       "1.22",
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// MUST: Alpine still installs nodejs and npm via apk in the merged install
	if !strings.Contains(dockerfile, "nodejs") {
		t.Errorf("Alpine workspace should include 'nodejs' in apk install.\n"+
			"WI-1: Alpine path is unchanged — apk still installs nodejs.\n"+
			"Generated Dockerfile:\n%s", dockerfile)
	}
	if !strings.Contains(dockerfile, "npm") {
		t.Errorf("Alpine workspace should include 'npm' in apk install.\n"+
			"WI-1: Alpine path is unchanged — apk still installs npm.\n"+
			"Generated Dockerfile:\n%s", dockerfile)
	}

	// MUST NOT: Alpine does not use NodeSource (distro-specific to Debian/Ubuntu)
	if strings.Contains(dockerfile, "nodesource") {
		t.Errorf("Alpine workspace should NOT use nodesource.\n"+
			"WI-1: NodeSource is for Debian only; Alpine uses apk for nodejs.\n"+
			"Generated Dockerfile:\n%s", dockerfile)
	}
}

// TestEffectiveVersion_NodejsDefault20 verifies that effectiveVersion() returns "20"
// for the nodejs language when no version is explicitly specified.
//
// WI-2: Update nodejs default version from 18 → 20.
// MUST FAIL against current code (currently returns "18" for nodejs).
func TestEffectiveVersion_NodejsDefault20(t *testing.T) {
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
		Version:       "", // No explicit version — should default to "20"
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

	impl := gen.(*DefaultDockerfileGenerator)

	got := impl.effectiveVersion()
	want := "20"
	if got != want {
		t.Errorf("effectiveVersion() for nodejs with no version set = %q, want %q.\n"+
			"WI-2: Update nodejs default from '18' to '20' in effectiveVersion().",
			got, want)
	}

	// Also verify the generated FROM line uses node:20-alpine (with optional digest pin)
	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if !strings.Contains(dockerfile, "FROM node:20-alpine") {
		t.Errorf("nodejs workspace with no version should generate 'FROM node:20-alpine... AS base'.\n"+
			"WI-2: Default nodejs version must be 20.\n"+
			"Generated Dockerfile (first 300 chars):\n%s", dockerfile[:min(300, len(dockerfile))])
	}
}

// TestGetMasonToolsForLanguage_IncludesLinters verifies that getMasonToolsForLanguage()
// returns expanded tool lists that include linters and formatters for each language.
//
// WI-4: Expand Mason tool lists to include linters/formatters.
// MUST FAIL against current code (currently missing pylint, shellcheck, etc.).
func TestGetMasonToolsForLanguage_IncludesLinters(t *testing.T) {
	tests := []struct {
		name         string
		language     string
		wantTools    []string // tools that MUST be present
		notWantTools []string // tools that must NOT be present (sanity check)
	}{
		{
			name:     "python includes pyright, ruff, black, isort, and pylint",
			language: "python",
			wantTools: []string{
				"pyright",
				"ruff",
				"black",
				"isort",
				"pylint", // WI-4: new addition
			},
		},
		{
			name:     "golang includes gopls and golangci-lint-langserver",
			language: "golang",
			wantTools: []string{
				"gopls",
				"golangci-lint-langserver",
			},
		},
		{
			name:     "nodejs includes typescript-language-server, eslint-lsp, and prettier",
			language: "nodejs",
			wantTools: []string{
				"typescript-language-server",
				"eslint-lsp",
				"prettier",
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
				Language:      tt.language,
				AppPath:       "/tmp/test",
				PathConfig:    paths.New(t.TempDir()),
			})

			impl := gen.(*DefaultDockerfileGenerator)
			tools := impl.getMasonToolsForLanguage()

			toolSet := make(map[string]bool, len(tools))
			for _, tool := range tools {
				toolSet[tool] = true
			}

			for _, want := range tt.wantTools {
				if !toolSet[want] {
					t.Errorf("[%s] getMasonToolsForLanguage() missing %q.\n"+
						"WI-4: Expand Mason tool lists to include linters/formatters.\n"+
						"Got tools: %v",
						tt.name, want, tools)
				}
			}

			for _, notWant := range tt.notWantTools {
				if toolSet[notWant] {
					t.Errorf("[%s] getMasonToolsForLanguage() unexpectedly includes %q.\n"+
						"Got tools: %v",
						tt.name, notWant, tools)
				}
			}
		})
	}
}

// TestGetMasonToolsForLanguage_BaseToolsAlwaysPresent verifies that the Mason
// install Lua script in the generated Dockerfile always includes base Mason tools
// (lua-language-server, stylua) regardless of the workspace language.
//
// WI-4: Mason install must include lua-language-server and stylua for all languages.
func TestGetMasonToolsForLanguage_BaseToolsAlwaysPresent(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name     string
		language string
	}{
		{name: "python workspace", language: "python"},
		{name: "golang workspace", language: "golang"},
		{name: "nodejs workspace", language: "nodejs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create staging dir with nvim config so Mason install is generated
			repoName := "test-mason-base-" + tt.language
			stagingDir := filepath.Join(homeDir, ".devopsmaestro", "build-staging", repoName)
			nvimConfigPath := filepath.Join(stagingDir, ".config", "nvim")
			if err := os.MkdirAll(nvimConfigPath, 0755); err != nil {
				t.Fatalf("failed to create nvim config dir: %v", err)
			}
			defer os.RemoveAll(stagingDir)

			initLuaPath := filepath.Join(nvimConfigPath, "init.lua")
			if err := os.WriteFile(initLuaPath, []byte("-- test"), 0644); err != nil {
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

			manifest := &plugin.PluginManifest{
				Features: plugin.PluginFeatures{
					HasMason:      true,
					HasTreesitter: false,
				},
			}

			sourcePath := filepath.Join("/tmp", "dvm-clone-xyz", repoName)

			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
				Workspace:     ws,
				WorkspaceSpec: wsYAML,
				Language:      tt.language,
				AppPath:       sourcePath,
				PathConfig:    paths.New(homeDir),
			})
			gen.SetPluginManifest(manifest)

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			// Locate the Lua tools list in the mason-install.lua heredoc
			toolsIdx := strings.Index(dockerfile, "local tools = {")
			if toolsIdx < 0 {
				t.Fatalf("[%s] Generate() missing Mason Lua tools list", tt.name)
			}
			toolsLineEnd := strings.Index(dockerfile[toolsIdx:], "\n")
			var toolsLine string
			if toolsLineEnd > 0 {
				toolsLine = dockerfile[toolsIdx : toolsIdx+toolsLineEnd]
			} else {
				toolsLine = dockerfile[toolsIdx:]
			}

			// MUST include base tools regardless of language
			// WI-4: getBaseMasonTools() must be called by installMasonTools()
			baseTools := []string{"lua-language-server", "stylua"}
			for _, tool := range baseTools {
				if !strings.Contains(toolsLine, tool) {
					t.Errorf("[%s] Mason install missing base tool %q.\n"+
						"WI-4: Base tools (lua-language-server, stylua) must always be included.\n"+
						"Tools line: %s",
						tt.name, tool, toolsLine)
				}
			}
		})
	}
}

// TestInstallMasonLSPs_IncludesBaseTools verifies that the installMasonTools() output
// includes both language-specific LSPs AND base tools (lua-language-server, stylua) in the
// Mason install Lua script for any language.
//
// WI-4: Mason install must include base tools alongside language-specific tools.
func TestInstallMasonLSPs_IncludesBaseTools(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	// Create a staging dir with nvim config so generateNvimSection() runs installMasonTools()
	repoName := "test-mason-base-tools"
	stagingDir := filepath.Join(homeDir, ".devopsmaestro", "build-staging", repoName)
	nvimConfigPath := filepath.Join(stagingDir, ".config", "nvim")
	if err := os.MkdirAll(nvimConfigPath, 0755); err != nil {
		t.Fatalf("failed to create nvim config dir: %v", err)
	}
	defer os.RemoveAll(stagingDir)

	initLuaPath := filepath.Join(nvimConfigPath, "init.lua")
	if err := os.WriteFile(initLuaPath, []byte("-- test"), 0644); err != nil {
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

	// Use a manifest with Mason enabled to trigger MasonInstall
	manifest := &plugin.PluginManifest{
		Features: plugin.PluginFeatures{
			HasMason:      true,
			HasTreesitter: false,
		},
	}

	sourcePath := filepath.Join("/tmp", "dvm-clone-xyz", repoName)

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       sourcePath,
		PathConfig:    paths.New(homeDir),
	})
	gen.SetPluginManifest(manifest)

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Locate the Mason Lua tools list
	toolsIdx := strings.Index(dockerfile, "local tools = {")
	if toolsIdx < 0 {
		t.Fatalf("Generate() missing Mason Lua tools list — nvim section must be generated")
	}

	// Extract just the tools list line
	toolsLineEnd := strings.Index(dockerfile[toolsIdx:], "\n")
	var toolsLine string
	if toolsLineEnd > 0 {
		toolsLine = dockerfile[toolsIdx : toolsIdx+toolsLineEnd]
	} else {
		toolsLine = dockerfile[toolsIdx:]
	}

	// MUST include language-specific tools for Python
	pythonTools := []string{"pyright", "ruff", "black", "isort"}
	for _, tool := range pythonTools {
		if !strings.Contains(toolsLine, tool) {
			t.Errorf("Mason install missing Python tool %q.\n"+
				"WI-4: Language tools must be included in Mason install.\n"+
				"Tools line: %s", tool, toolsLine)
		}
	}

	// MUST include base tools (lua-language-server, stylua) alongside language-specific tools
	baseTools := []string{"lua-language-server", "stylua"}
	for _, tool := range baseTools {
		if !strings.Contains(toolsLine, tool) {
			t.Errorf("Mason install missing base tool %q.\n"+
				"WI-4: Base tools (lua-language-server, stylua) must always be included.\n"+
				"Tools line: %s", tool, toolsLine)
		}
	}
}

// TestGenerateDevStage_DebianNodeSource_OrderAfterMergedInstall verifies that for a
// Debian-based workspace with nvim enabled, the NodeSource curl command appears AFTER
// the dev-stage merged apt-get install block (which installs curl).
//
// Bug: The NodeSource RUN block currently runs BEFORE the merged apt-get install that
// provides curl, causing "curl: not found" failures at build time.
// Fix: Move the NodeSource block to after the merged apt-get install.
//
// Note: The generated Dockerfile has TWO "apt-get install --fix-broken" occurrences:
//  1. Base stage (early): installs build-essential only — curl is NOT here
//  2. Dev stage (later): the merged install that includes curl, git, wget, etc.
//
// We must anchor on the dev-stage merged install specifically, identified by its
// preceding comment "# Install all dev tools".
func TestGenerateDevStage_DebianNodeSource_OrderAfterMergedInstall(t *testing.T) {
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

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python", // Debian-based
		Version:       "3.11",
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Locate the dev-stage merged install block by its unique preceding comment.
	// This avoids false-matching the base stage's apt-get install line.
	mergedInstallComment := "# Install all dev tools, nvim dependencies, and Mason toolchains (merged)"
	mergedInstallIdx := strings.Index(dockerfile, mergedInstallComment)
	if mergedInstallIdx < 0 {
		t.Fatalf("Could not locate dev-stage merged install block comment %q in generated Dockerfile.\n"+
			"Generated Dockerfile:\n%s", mergedInstallComment, dockerfile)
	}

	// Locate the NodeSource setup script invocation
	nodeSourceIdx := strings.Index(dockerfile, "nodesource.com/setup_22.x")
	if nodeSourceIdx < 0 {
		t.Fatalf("Could not locate NodeSource setup_22.x in generated Dockerfile.\n"+
			"Generated Dockerfile:\n%s", dockerfile)
	}

	// The NodeSource block MUST appear AFTER the dev-stage merged apt-get install block.
	// If NodeSource comes before, curl is not yet installed and the build will fail
	// with "curl: not found".
	if nodeSourceIdx < mergedInstallIdx {
		t.Errorf("NodeSource curl command appears BEFORE the dev-stage merged apt-get install block.\n"+
			"Bug: curl is not installed until the merged apt-get install runs,\n"+
			"so the NodeSource RUN block must be placed AFTER it.\n"+
			"  mergedInstallIdx  = %d (comment: %q)\n"+
			"  nodeSourceIdx     = %d (nodesource.com/setup_22.x)\n"+
			"NodeSource must come AFTER merged install (nodeSourceIdx > mergedInstallIdx).\n"+
			"Generated Dockerfile:\n%s", mergedInstallIdx, mergedInstallComment, nodeSourceIdx, dockerfile)
	}
}

// TestGenerateDevStage_DebianNodeSource_Fallback verifies that the NodeSource install
// block includes a fallback to the Debian-packaged nodejs and npm when NodeSource is
// unreachable (corporate firewalls, Colima VM networking, etc.).
//
// Bug: The current NodeSource block has no fallback. If the NodeSource CDN is
// unreachable inside the build container the entire build fails with no recovery path.
// Fix: Use a shell || fallback so that if NodeSource fails, the Debian-packaged
// nodejs/npm are installed instead.
func TestGenerateDevStage_DebianNodeSource_Fallback(t *testing.T) {
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

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python", // Debian-based
		Version:       "3.11",
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// The NodeSource block MUST include a fallback that installs the Debian-packaged
	// nodejs and npm so that builds succeed even when NodeSource is unreachable.
	fallback := "apt-get install -y --no-install-recommends nodejs npm"
	if !strings.Contains(dockerfile, fallback) {
		t.Errorf("NodeSource install block is missing a network fallback.\n"+
			"Bug: If nodesource.com is unreachable (firewall, Colima networking),\n"+
			"the build will fail with no recovery path.\n"+
			"Fix: Add a || fallback that installs Debian-packaged nodejs and npm.\n"+
			"Expected Dockerfile to contain: %q\n"+
			"Generated Dockerfile:\n%s", fallback, dockerfile)
	}
}

// TestGetBaseMasonTools_UsesRegistryNames verifies that getBaseMasonTools() returns
// Mason registry package names (hyphenated), not nvim-lspconfig names (underscored).
//
// Bug: getBaseMasonTools() returns "lua_ls" which is the nvim-lspconfig name.
// Mason's registry uses "lua-language-server". The MasonInstall command fails with
// '"lua_ls" is not a valid package' at build time.
func TestGetBaseMasonTools_UsesRegistryNames(t *testing.T) {
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
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

	impl := gen.(*DefaultDockerfileGenerator)
	tools := impl.getBaseMasonTools()

	for _, tool := range tools {
		if strings.Contains(tool, "_") {
			t.Errorf("getBaseMasonTools() returned %q which contains underscores.\n"+
				"Mason registry uses hyphenated names (e.g., 'lua-language-server', not 'lua_ls').\n"+
				"MasonInstall will fail with '\"<name>\" is not a valid package'.\n"+
				"All tools: %v", tool, tools)
		}
	}

	// Specifically verify lua-language-server is present (not lua_ls)
	found := false
	for _, tool := range tools {
		if tool == "lua-language-server" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("getBaseMasonTools() should include 'lua-language-server' (Mason registry name).\n"+
			"Got: %v", tools)
	}
}

// TestPipInstall_ProxyFallback verifies that ALL 5 pip install sites in the generated
// Dockerfile include the proxy-aware fallback pattern:
//
//	pip install ... \
//	  || (unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy \
//	  && pip install ...)
//
// This is the RED-phase test for the proxy-fallback feature. It MUST FAIL until
// the implementation is complete.
//
// The 5 pip install sites are:
//  1. generateBaseStage() — default/no-private-repos case
//  2. generateBaseStage() — HTTPS private repos case
//  3. generateBaseStage() — SSH private repos case
//  4. generateBaseStage() — mixed HTTPS+SSH private repos case
//  5. installLanguageTools() — python dev tools (ruff, mypy, etc.)
func TestPipInstall_ProxyFallback(t *testing.T) {
	tests := []struct {
		name                string
		requirementsContent string
		devTools            []string
		wantContain         []string
	}{
		{
			// Site 1: generateBaseStage() default case — plain requirements.txt, no private repos
			name:                "site1: default plain pip install has proxy fallback",
			requirementsContent: "flask==3.0.0\n",
			wantContain: []string{
				"pip install -r /tmp/requirements.txt",
				"|| (unset",
				"unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy",
			},
		},
		{
			// Site 2: generateBaseStage() HTTPS private repos case
			name: "site2: HTTPS private repos pip install has proxy fallback",
			requirementsContent: "flask==3.0.0\n" +
				"git+https://${GITHUB_USERNAME}:${GITHUB_PAT}@github.com/Org/repo.git@v1.0\n",
			wantContain: []string{
				"pip install -r /tmp/requirements.txt",
				"|| (unset",
				"unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy",
			},
		},
		{
			// Site 3: generateBaseStage() SSH private repos case
			name:                "site3: SSH private repos pip install has proxy fallback",
			requirementsContent: "mylib @ git+ssh://git@github.com/Org/repo.git@v1.0\n",
			wantContain: []string{
				"--mount=type=ssh",
				"pip install -r /tmp/requirements.txt",
				"|| (unset",
				"unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy",
			},
		},
		{
			// Site 4: generateBaseStage() mixed HTTPS+SSH private repos case
			name: "site4: mixed HTTPS+SSH pip install has proxy fallback",
			requirementsContent: "git+https://${GITHUB_USERNAME}:${GITHUB_PAT}@github.com/Org/private-lib.git@v1.0\n" +
				"mylib @ git+ssh://git@github.com/Org/repo.git@v2.0\n",
			wantContain: []string{
				"--mount=type=ssh",
				"pip install -r /tmp/requirements.txt",
				"|| (unset",
				"unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy",
			},
		},
		{
			// Site 5: installLanguageTools() — python dev tools installed in the dev stage
			name:     "site5: installLanguageTools python dev tools has proxy fallback",
			devTools: []string{"ruff", "mypy"},
			wantContain: []string{
				"pip install ruff mypy",
				"|| (unset",
				"unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appPath := t.TempDir()

			// Write requirements.txt when this variant needs one
			if tt.requirementsContent != "" {
				reqPath := filepath.Join(appPath, "requirements.txt")
				if err := os.WriteFile(reqPath, []byte(tt.requirementsContent), 0644); err != nil {
					t.Fatalf("failed to write requirements.txt: %v", err)
				}
			}

			ws := &models.Workspace{
				ID:        1,
				Name:      "test-ws",
				ImageName: "test:latest",
			}
			wsYAML := models.WorkspaceSpec{}
			if len(tt.devTools) > 0 {
				wsYAML.Build.DevStage.DevTools = tt.devTools
			}

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
					t.Errorf("Generate() missing expected proxy-fallback content: %q\nDockerfile:\n%s", want, dockerfile)
				}
			}
		})
	}
}

// min returns the smaller of two ints. Used for safe string slicing in error messages.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// =============================================================================
// Python system dependency auto-detection tests (RED phase)
// These tests MUST FAIL until Phase 3 implementation is complete.
// They drive the implementation of detectPythonSystemDeps integration into
// the Dockerfile generator's base stage.
// =============================================================================

// TestDockerfileGenerator_SystemDeps verifies that Python packages requiring
// system libraries cause those libraries to appear in the base stage apt-get
// install command of the generated Dockerfile.
//
// RED phase: Tests FAIL because the generator does not yet call
// detectPythonSystemDeps or emit the detected system packages.
func TestDockerfileGenerator_SystemDeps(t *testing.T) {
	tests := []struct {
		name                string
		requirementsContent string   // written to requirements.txt in appPath
		baseStagePackages   []string // explicit packages via WorkspaceSpec.Build.BaseStage.Packages
		wantContain         []string // must appear in generated Dockerfile
		wantNotContain      []string // must NOT appear in generated Dockerfile
	}{
		{
			name:                "psycopg2 causes libpq-dev in base stage",
			requirementsContent: "psycopg2==2.9.9\nflask==2.3.0\n",
			wantContain:         []string{"libpq-dev"},
			wantNotContain:      nil,
		},
		{
			name:                "Pillow causes libjpeg-dev zlib1g-dev libfreetype6-dev in base stage",
			requirementsContent: "Pillow>=10.0\n",
			wantContain:         []string{"libjpeg-dev", "zlib1g-dev", "libfreetype6-dev"},
			wantNotContain:      nil,
		},
		{
			name:                "cryptography and cffi deduplicated — libffi-dev appears once",
			requirementsContent: "cryptography>=41.0\ncffi>=1.0\n",
			wantContain:         []string{"libffi-dev", "libssl-dev"},
			wantNotContain:      nil,
		},
		{
			name:                "psycopg2-binary needs no system deps",
			requirementsContent: "psycopg2-binary==2.9.9\n",
			wantContain:         nil,
			wantNotContain:      []string{"libpq-dev"},
		},
		{
			name:                "no system dep packages — no extra apt installs",
			requirementsContent: "flask==2.3.0\nrequests>=2.28\n",
			wantContain:         nil,
			wantNotContain:      []string{"libpq-dev", "libjpeg-dev", "libffi-dev"},
		},
		{
			name:                "lxml causes libxml2-dev and libxslt1-dev",
			requirementsContent: "lxml>=4.0\n",
			wantContain:         []string{"libxml2-dev", "libxslt1-dev"},
			wantNotContain:      nil,
		},
		{
			name:                "PyYAML causes libyaml-dev",
			requirementsContent: "PyYAML>=6.0\n",
			wantContain:         []string{"libyaml-dev"},
			wantNotContain:      nil,
		},
		{
			name:                "h5py causes libhdf5-dev",
			requirementsContent: "h5py>=3.0\n",
			wantContain:         []string{"libhdf5-dev"},
			wantNotContain:      nil,
		},
		{
			name:                "gevent causes libev-dev and libevent-dev",
			requirementsContent: "gevent>=23.0\n",
			wantContain:         []string{"libev-dev", "libevent-dev"},
			wantNotContain:      nil,
		},
		{
			name:                "mixed known and unknown packages",
			requirementsContent: "psycopg2==2.9.9\nflask==2.3.0\nlxml>=4.0\n",
			wantContain:         []string{"libpq-dev", "libxml2-dev", "libxslt1-dev"},
			wantNotContain:      nil,
		},
		{
			name:                "base stage packages merged with auto-detected system deps",
			requirementsContent: "psycopg2==2.9.9\n",
			baseStagePackages:   []string{"curl", "jq"},
			wantContain:         []string{"libpq-dev", "curl", "jq"},
			wantNotContain:      nil,
		},
		{
			name:                "no requirements.txt — no system dep packages added",
			requirementsContent: "", // do not create requirements.txt
			wantContain:         nil,
			wantNotContain:      []string{"libpq-dev", "libjpeg-dev", "libffi-dev"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appPath := t.TempDir()

			if tt.requirementsContent != "" {
				reqPath := filepath.Join(appPath, "requirements.txt")
				if err := os.WriteFile(reqPath, []byte(tt.requirementsContent), 0644); err != nil {
					t.Fatalf("failed to write requirements.txt: %v", err)
				}
			}

			ws := &models.Workspace{
				ID:        1,
				Name:      "test-ws",
				ImageName: "test:latest",
			}
			wsYAML := models.WorkspaceSpec{
				Build: models.DevBuildConfig{
					BaseStage: models.BaseStageConfig{
						Packages: tt.baseStagePackages,
					},
				},
			}

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
					t.Errorf("Generate() missing expected system dep %q\nDockerfile:\n%s", want, dockerfile)
				}
			}

			for _, notWant := range tt.wantNotContain {
				if strings.Contains(dockerfile, notWant) {
					t.Errorf("Generate() contains unexpected system dep %q\nDockerfile:\n%s", notWant, dockerfile)
				}
			}
		})
	}
}

// ── WI-1: USER directive tests ───────────────────────────────────────────────

// TestGenerate_CustomUser_USERDirectiveMatchesConfig verifies that when Container.User
// is configured, the final USER directive in the Dockerfile uses that value, not "dev".
func TestGenerate_CustomUser_USERDirectiveMatchesConfig(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{
		Container: models.ContainerConfig{
			User: "ray",
			UID:  2021,
			GID:  2021,
		},
	}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:       ws,
		WorkspaceSpec:   wsYAML,
		Language:        "python",
		AppPath:         "/tmp/test",
		PathConfig:      paths.New(t.TempDir()),
		PrivateRepoInfo: &utils.PrivateRepoInfo{},
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// The last USER directive should be "USER ray", not "USER dev"
	lastUserIdx := strings.LastIndex(dockerfile, "USER ")
	if lastUserIdx == -1 {
		t.Fatalf("Generate() output contains no USER directive\nDockerfile:\n%s", dockerfile)
	}
	lastUserLine := dockerfile[lastUserIdx:]
	// Extract the line
	if newline := strings.Index(lastUserLine, "\n"); newline != -1 {
		lastUserLine = lastUserLine[:newline]
	}
	if lastUserLine != "USER ray" {
		t.Errorf("Generate() last USER directive = %q, want %q\nDockerfile:\n%s", lastUserLine, "USER ray", dockerfile)
	}
}

// TestGenerate_DefaultUser_USERDirectiveIsDev is a regression guard: the default
// USER directive (no custom user configured) must remain "USER dev".
func TestGenerate_DefaultUser_USERDirectiveIsDev(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:       ws,
		WorkspaceSpec:   wsYAML,
		Language:        "python",
		AppPath:         "/tmp/test",
		PathConfig:      paths.New(t.TempDir()),
		PrivateRepoInfo: &utils.PrivateRepoInfo{},
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if !strings.Contains(dockerfile, "USER dev") {
		t.Errorf("Generate() missing 'USER dev' in default config\nDockerfile:\n%s", dockerfile)
	}
}

// ── WI-2: Additional build args tests ────────────────────────────────────────

// TestGenerate_AdditionalBuildArgs_EmitsARGDeclarations verifies that ARG directives
// are emitted for each name in AdditionalBuildArgs.
func TestGenerate_AdditionalBuildArgs_EmitsARGDeclarations(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:           ws,
		WorkspaceSpec:       wsYAML,
		Language:            "python",
		AppPath:             "/tmp/test",
		PathConfig:          paths.New(t.TempDir()),
		PrivateRepoInfo:     &utils.PrivateRepoInfo{},
		AdditionalBuildArgs: []string{"PIP_INDEX_URL", "GOPROXY"},
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	for _, want := range []string{"ARG PIP_INDEX_URL", "ARG GOPROXY"} {
		if !strings.Contains(dockerfile, want) {
			t.Errorf("Generate() missing expected ARG directive %q\nDockerfile:\n%s", want, dockerfile)
		}
	}
}

// TestGenerate_WorkspaceBuildArgs_EmitsARGDeclarations verifies that ARG directives
// are emitted for keys in WorkspaceSpec.Build.Args.
func TestGenerate_WorkspaceBuildArgs_EmitsARGDeclarations(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{
		Build: models.DevBuildConfig{
			Args: map[string]string{
				"CUSTOM_VAR": "value",
			},
		},
	}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:       ws,
		WorkspaceSpec:   wsYAML,
		Language:        "python",
		AppPath:         "/tmp/test",
		PathConfig:      paths.New(t.TempDir()),
		PrivateRepoInfo: &utils.PrivateRepoInfo{},
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if !strings.Contains(dockerfile, "ARG CUSTOM_VAR") {
		t.Errorf("Generate() missing 'ARG CUSTOM_VAR' from WorkspaceSpec.Build.Args\nDockerfile:\n%s", dockerfile)
	}
}

// TestGenerate_AdditionalBuildArgs_DeduplicatesWithRequiredBuildArgs verifies that when
// the same ARG name appears in both PrivateRepoInfo.RequiredBuildArgs and
// AdditionalBuildArgs, it is emitted exactly once.
func TestGenerate_AdditionalBuildArgs_DeduplicatesWithRequiredBuildArgs(t *testing.T) {
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
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
		PrivateRepoInfo: &utils.PrivateRepoInfo{
			RequiredBuildArgs: []string{"GITHUB_TOKEN"},
			NeedsGit:          true,
		},
		AdditionalBuildArgs: []string{"GITHUB_TOKEN", "PIP_INDEX_URL"},
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	count := strings.Count(dockerfile, "ARG GITHUB_TOKEN")
	if count != 1 {
		t.Errorf("Generate() contains %d occurrences of 'ARG GITHUB_TOKEN', want exactly 1\nDockerfile:\n%s", count, dockerfile)
	}

	if !strings.Contains(dockerfile, "ARG PIP_INDEX_URL") {
		t.Errorf("Generate() missing 'ARG PIP_INDEX_URL'\nDockerfile:\n%s", dockerfile)
	}
}

// TestGenerate_AdditionalBuildArgs_NoENVEmitted verifies that additional build args
// are declared as ARG only — NOT persisted as ENV (security: secrets must not leak).
func TestGenerate_AdditionalBuildArgs_NoENVEmitted(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:           ws,
		WorkspaceSpec:       wsYAML,
		Language:            "python",
		AppPath:             "/tmp/test",
		PathConfig:          paths.New(t.TempDir()),
		PrivateRepoInfo:     &utils.PrivateRepoInfo{},
		AdditionalBuildArgs: []string{"PIP_INDEX_URL"},
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if strings.Contains(dockerfile, "ENV PIP_INDEX_URL") {
		t.Errorf("Generate() must NOT emit 'ENV PIP_INDEX_URL' — ARG only, no ENV persistence\nDockerfile:\n%s", dockerfile)
	}
}

// TestGenerate_NoAdditionalBuildArgs_NoExtraARGs is a regression guard verifying that
// unknown ARG names do not appear when no additional build args are configured.
func TestGenerate_NoAdditionalBuildArgs_NoExtraARGs(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:       ws,
		WorkspaceSpec:   wsYAML,
		Language:        "python",
		AppPath:         "/tmp/test",
		PathConfig:      paths.New(t.TempDir()),
		PrivateRepoInfo: &utils.PrivateRepoInfo{},
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	for _, notWant := range []string{"ARG PIP_INDEX_URL", "ARG GOPROXY"} {
		if strings.Contains(dockerfile, notWant) {
			t.Errorf("Generate() unexpectedly contains %q with no additional build args configured\nDockerfile:\n%s", notWant, dockerfile)
		}
	}
}

// TestGenerate_AdditionalBuildArgs_Python_BeforePipInstall verifies that ARG declarations
// for additional build args appear before the pip install command so they are available
// when pip runs (e.g., PIP_INDEX_URL controlling the package index).
func TestGenerate_AdditionalBuildArgs_Python_BeforePipInstall(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:           ws,
		WorkspaceSpec:       wsYAML,
		Language:            "python",
		AppPath:             "/tmp/test",
		PathConfig:          paths.New(t.TempDir()),
		PrivateRepoInfo:     &utils.PrivateRepoInfo{},
		AdditionalBuildArgs: []string{"PIP_INDEX_URL"},
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	argIdx := strings.Index(dockerfile, "ARG PIP_INDEX_URL")
	pipIdx := strings.Index(dockerfile, "pip install")

	if argIdx == -1 {
		t.Fatalf("Generate() missing 'ARG PIP_INDEX_URL'\nDockerfile:\n%s", dockerfile)
	}
	if pipIdx == -1 {
		t.Fatalf("Generate() missing 'pip install'\nDockerfile:\n%s", dockerfile)
	}
	if argIdx >= pipIdx {
		t.Errorf("Generate() 'ARG PIP_INDEX_URL' (idx %d) must appear before 'pip install' (idx %d)\nDockerfile:\n%s", argIdx, pipIdx, dockerfile)
	}
}

// TestGenerate_AdditionalBuildArgs_Golang_EmitsARG verifies that additional build args
// are emitted for Golang language workspaces (GOPROXY controls the module proxy).
func TestGenerate_AdditionalBuildArgs_Golang_EmitsARG(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:           ws,
		WorkspaceSpec:       wsYAML,
		Language:            "golang",
		AppPath:             "/tmp/test",
		PathConfig:          paths.New(t.TempDir()),
		PrivateRepoInfo:     &utils.PrivateRepoInfo{},
		AdditionalBuildArgs: []string{"GOPROXY"},
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if !strings.Contains(dockerfile, "ARG GOPROXY") {
		t.Errorf("Generate() missing 'ARG GOPROXY' for Golang workspace\nDockerfile:\n%s", dockerfile)
	}
}

// TestGenerate_AdditionalBuildArgs_Nodejs_EmitsARG verifies that additional build args
// are emitted for NodeJS language workspaces (NPM_CONFIG_REGISTRY controls the npm registry).
func TestGenerate_AdditionalBuildArgs_Nodejs_EmitsARG(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:           ws,
		WorkspaceSpec:       wsYAML,
		Language:            "nodejs",
		AppPath:             "/tmp/test",
		PathConfig:          paths.New(t.TempDir()),
		PrivateRepoInfo:     &utils.PrivateRepoInfo{},
		AdditionalBuildArgs: []string{"NPM_CONFIG_REGISTRY"},
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if !strings.Contains(dockerfile, "ARG NPM_CONFIG_REGISTRY") {
		t.Errorf("Generate() missing 'ARG NPM_CONFIG_REGISTRY' for NodeJS workspace\nDockerfile:\n%s", dockerfile)
	}
}

// ── WI-3: CA certificate injection tests ─────────────────────────────────────

// TestGenerate_CACerts_EmitsCOPYAndUpdateCACertificates verifies the complete CA cert
// injection block: COPY into ca-certificates dir, update-ca-certificates, and all
// required ENV vars so Python/Node/curl all trust the corporate CA.
func TestGenerate_CACerts_EmitsCOPYAndUpdateCACertificates(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{
		Build: models.DevBuildConfig{
			CACerts: []models.CACertConfig{
				{
					Name:        "corporate-ca",
					VaultSecret: "cert",
				},
			},
		},
	}

	// Create a staging dir with certs/ so the generator finds it (#228)
	stagingDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(stagingDir, "certs"), 0755); err != nil {
		t.Fatalf("failed to create certs dir: %v", err)
	}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:       ws,
		WorkspaceSpec:   wsYAML,
		Language:        "python",
		AppPath:         "/tmp/test",
		PathConfig:      paths.New(t.TempDir()),
		StagingDir:      stagingDir,
		PrivateRepoInfo: &utils.PrivateRepoInfo{},
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	wantContain := []string{
		"COPY certs/ /usr/local/share/ca-certificates/custom/",
		"RUN update-ca-certificates",
		"ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt",
		"ENV REQUESTS_CA_BUNDLE=/etc/ssl/certs/ca-certificates.crt",
		"ENV NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt",
	}

	for _, want := range wantContain {
		if !strings.Contains(dockerfile, want) {
			t.Errorf("Generate() missing CA cert directive %q\nDockerfile:\n%s", want, dockerfile)
		}
	}
}

// TestGenerate_CACerts_BeforePipInstall verifies that the CA cert installation block
// appears before any pip install command so Python's pip trusts the CA during installs.
func TestGenerate_CACerts_BeforePipInstall(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{
		Build: models.DevBuildConfig{
			CACerts: []models.CACertConfig{
				{Name: "corporate-ca", VaultSecret: "cert"},
			},
		},
	}

	// Create staging dir with certs/ and requirements.txt (#228)
	stagingDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(stagingDir, "certs"), 0755); err != nil {
		t.Fatalf("failed to create certs dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stagingDir, "requirements.txt"), []byte("flask\n"), 0644); err != nil {
		t.Fatalf("failed to create requirements.txt: %v", err)
	}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:       ws,
		WorkspaceSpec:   wsYAML,
		Language:        "python",
		AppPath:         "/tmp/test",
		PathConfig:      paths.New(t.TempDir()),
		StagingDir:      stagingDir,
		PrivateRepoInfo: &utils.PrivateRepoInfo{},
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	certIdx := strings.Index(dockerfile, "COPY certs/")
	pipIdx := strings.Index(dockerfile, "pip install")

	if certIdx == -1 {
		t.Fatalf("Generate() missing 'COPY certs/' for CA cert injection\nDockerfile:\n%s", dockerfile)
	}
	if pipIdx == -1 {
		t.Fatalf("Generate() missing 'pip install'\nDockerfile:\n%s", dockerfile)
	}
	if certIdx >= pipIdx {
		t.Errorf("Generate() CA cert COPY (idx %d) must appear before pip install (idx %d)\nDockerfile:\n%s", certIdx, pipIdx, dockerfile)
	}
}

// TestGenerate_CACerts_Alpine_SamePathsAndCommands verifies that Golang (Alpine-based)
// workspaces use the same CA cert COPY destination and update-ca-certificates command
// as Debian-based workspaces (both distros support the same paths when ca-certificates is installed).
func TestGenerate_CACerts_Alpine_SamePathsAndCommands(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{
		Build: models.DevBuildConfig{
			CACerts: []models.CACertConfig{
				{Name: "corporate-ca", VaultSecret: "cert"},
			},
		},
	}

	// Create staging dir with certs/ (#228)
	stagingDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(stagingDir, "certs"), 0755); err != nil {
		t.Fatalf("failed to create certs dir: %v", err)
	}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:       ws,
		WorkspaceSpec:   wsYAML,
		Language:        "golang",
		AppPath:         "/tmp/test",
		PathConfig:      paths.New(t.TempDir()),
		StagingDir:      stagingDir,
		PrivateRepoInfo: &utils.PrivateRepoInfo{},
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	wantContain := []string{
		"COPY certs/ /usr/local/share/ca-certificates/custom/",
		"RUN update-ca-certificates",
	}

	for _, want := range wantContain {
		if !strings.Contains(dockerfile, want) {
			t.Errorf("Generate() (golang/alpine) missing CA cert directive %q\nDockerfile:\n%s", want, dockerfile)
		}
	}
}

// TestGenerate_CACerts_NoCerts_NoSection is a regression guard: when no caCerts are
// configured the Dockerfile must not contain any CA cert injection commands.
func TestGenerate_CACerts_NoCerts_NoSection(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:       ws,
		WorkspaceSpec:   wsYAML,
		Language:        "python",
		AppPath:         "/tmp/test",
		PathConfig:      paths.New(t.TempDir()),
		PrivateRepoInfo: &utils.PrivateRepoInfo{},
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	for _, notWant := range []string{"update-ca-certificates", "COPY certs/"} {
		if strings.Contains(dockerfile, notWant) {
			t.Errorf("Generate() unexpectedly contains %q when no caCerts are configured\nDockerfile:\n%s", notWant, dockerfile)
		}
	}
}

// TestGenerate_BuildArgs_RedeclaredInDevStage verifies that ARG declarations appear in
// BOTH the base stage and the dev stage. Docker ARG values do not carry across FROM
// boundaries, so the dev stage must re-declare them for proxy vars and other build args
// to be available to RUN commands like npm install.
func TestGenerate_BuildArgs_RedeclaredInDevStage(t *testing.T) {
	tests := []struct {
		name     string
		language string
		args     []string
	}{
		{
			name:     "python with proxy args",
			language: "python",
			args:     []string{"http_proxy", "https_proxy", "PIP_INDEX_URL"},
		},
		{
			name:     "golang with proxy args",
			language: "golang",
			args:     []string{"http_proxy", "https_proxy", "GOPROXY"},
		},
		{
			name:     "nodejs with proxy args",
			language: "nodejs",
			args:     []string{"http_proxy", "https_proxy", "NPM_CONFIG_REGISTRY"},
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
				Workspace:           ws,
				WorkspaceSpec:       wsYAML,
				Language:            tt.language,
				AppPath:             "/tmp/test",
				PathConfig:          paths.New(t.TempDir()),
				PrivateRepoInfo:     &utils.PrivateRepoInfo{},
				AdditionalBuildArgs: tt.args,
			})

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			// Split at "FROM base AS dev" to isolate base stage vs dev stage
			parts := strings.SplitN(dockerfile, "FROM base AS dev", 2)
			if len(parts) != 2 {
				t.Fatalf("Generate() output missing 'FROM base AS dev' boundary\nDockerfile:\n%s", dockerfile)
			}
			baseStage := parts[0]
			devStage := parts[1]

			for _, arg := range tt.args {
				argDecl := "ARG " + arg
				if !strings.Contains(baseStage, argDecl) {
					t.Errorf("base stage missing %q\nBase stage:\n%s", argDecl, baseStage)
				}
				if !strings.Contains(devStage, argDecl) {
					t.Errorf("dev stage missing %q\nDev stage:\n%s", argDecl, devStage)
				}
			}
		})
	}
}

// TestGenerate_NoAdditionalBuildArgs_DevStageHasNoExtraARGs verifies that when no
// additional build args are configured, the dev stage does not emit any extra ARG block.
func TestGenerate_NoAdditionalBuildArgs_DevStageHasNoExtraARGs(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:       ws,
		WorkspaceSpec:   wsYAML,
		Language:        "python",
		AppPath:         "/tmp/test",
		PathConfig:      paths.New(t.TempDir()),
		PrivateRepoInfo: &utils.PrivateRepoInfo{},
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Split at "FROM base AS dev" to isolate the dev stage
	parts := strings.SplitN(dockerfile, "FROM base AS dev", 2)
	if len(parts) != 2 {
		t.Fatalf("Generate() output missing 'FROM base AS dev' boundary\nDockerfile:\n%s", dockerfile)
	}
	devStage := parts[1]

	if strings.Contains(devStage, "# Additional build arguments") {
		t.Errorf("dev stage should not contain '# Additional build arguments' when no build args are configured\nDev stage:\n%s", devStage)
	}
}

// TestGenerate_WorkspaceBuildArgs_RedeclaredInDevStage verifies that ARG declarations
// from WorkspaceSpec.Build.Args (not just AdditionalBuildArgs) are also re-declared in
// the dev stage.
func TestGenerate_WorkspaceBuildArgs_RedeclaredInDevStage(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{
		Build: models.DevBuildConfig{
			Args: map[string]string{
				"CUSTOM_PROXY": "http://proxy:8080",
			},
		},
	}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:       ws,
		WorkspaceSpec:   wsYAML,
		Language:        "python",
		AppPath:         "/tmp/test",
		PathConfig:      paths.New(t.TempDir()),
		PrivateRepoInfo: &utils.PrivateRepoInfo{},
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Split at "FROM base AS dev" to isolate base stage vs dev stage
	parts := strings.SplitN(dockerfile, "FROM base AS dev", 2)
	if len(parts) != 2 {
		t.Fatalf("Generate() output missing 'FROM base AS dev' boundary\nDockerfile:\n%s", dockerfile)
	}
	baseStage := parts[0]
	devStage := parts[1]

	argDecl := "ARG CUSTOM_PROXY"
	if !strings.Contains(baseStage, argDecl) {
		t.Errorf("base stage missing %q\nBase stage:\n%s", argDecl, baseStage)
	}
	if !strings.Contains(devStage, argDecl) {
		t.Errorf("dev stage missing %q\nDev stage:\n%s", argDecl, devStage)
	}
}

// =============================================================================
// Issue #146 — npm install proxy-unset fallback (TDD Phase 2 — FAILING TESTS)
// =============================================================================

// TestDockerfileGenerator_NpmNeovimInstall_ProxyUnsetFallback verifies that the global
// `npm install -g neovim` command in the dev stage includes a proxy-unset fallback,
// matching the existing pip pattern. This prevents 503 failures when BuildKit injects
// NPM_CONFIG_REGISTRY / HTTP_PROXY env vars pointing at host.docker.internal registries
// that are unreachable inside the Colima heavy-local VM.
//
// Phase 2 failing test for Issue #146 — fix NOT yet implemented.
// MUST FAIL against current code (no fallback on npm install -g neovim at line 882).
func TestDockerfileGenerator_NpmNeovimInstall_ProxyUnsetFallback(t *testing.T) {
	tests := []struct {
		name     string
		language string
		version  string
	}{
		{
			name:     "python/debian — neovim npm package in dev stage",
			language: "python",
			version:  "3.11",
		},
		{
			name:     "nodejs — neovim npm package in dev stage",
			language: "nodejs",
			version:  "20",
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

			// Verify the npm install -g neovim command is present at all
			if !strings.Contains(dockerfile, "npm install -g neovim") {
				t.Fatalf("Generate() missing 'npm install -g neovim' — test requires it to be present for language=%s", tt.language)
			}

			// MUST have: fallback pattern on npm install -g neovim
			// Expected form (matching pip pattern):
			//   npm install -g neovim || \
			//   (unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy NPM_CONFIG_REGISTRY npm_config_registry && npm install -g neovim)
			wantPatterns := []string{
				"npm install -g neovim ||",
				"unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy NPM_CONFIG_REGISTRY npm_config_registry",
			}
			for _, want := range wantPatterns {
				if !strings.Contains(dockerfile, want) {
					t.Errorf("npm install -g neovim missing proxy-unset fallback pattern: %q\n"+
						"All npm install commands must have a proxy-unset fallback to handle\n"+
						"unreachable NPM_CONFIG_REGISTRY/HTTP_PROXY injected by BuildRegistryCoordinator.\n"+
						"See pip fallback pattern at installLanguageTools() for reference.\n"+
						"Language: %s", want, tt.language)
				}
			}

			// MUST NOT have: bare npm install without fallback
			// If the fallback is present, this check ensures it's actually a fallback
			// (i.e., the original install attempt comes first)
			if strings.Contains(dockerfile, "npm install -g neovim\n") {
				t.Errorf("npm install -g neovim has no fallback — line ends immediately after command.\n"+
					"Must use '|| (unset ... && npm install -g neovim)' fallback pattern.\n"+
					"Language: %s", tt.language)
			}
		})
	}
}

// TestDockerfileGenerator_NpmLanguageTools_ProxyUnsetFallback verifies that the
// installLanguageTools() nodejs case also includes a proxy-unset fallback for any
// `npm install -g` commands, matching the pip fallback pattern already used for python.
//
// Phase 2 failing test for Issue #146 — fix NOT yet implemented.
// MUST FAIL against current code (installLanguageTools nodejs case at line 1069 has no fallback).
func TestDockerfileGenerator_NpmLanguageTools_ProxyUnsetFallback(t *testing.T) {
	tests := []struct {
		name      string
		language  string
		version   string
		devTools  []string
		wantTools []string
	}{
		{
			name:      "nodejs with custom dev tools",
			language:  "nodejs",
			version:   "20",
			devTools:  []string{"typescript", "ts-node"},
			wantTools: []string{"typescript", "ts-node"},
		},
		{
			name:      "nodejs with default dev tools",
			language:  "nodejs",
			version:   "20",
			devTools:  []string{}, // empty → default tools
			wantTools: []string{}, // just verify fallback exists, not specific tools
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
				Build: models.DevBuildConfig{
					DevStage: models.DevStageConfig{
						DevTools: tt.devTools,
					},
				},
			}

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

			// Find the "Install language-specific tools" section for nodejs
			if !strings.Contains(dockerfile, "# Install language-specific tools") {
				// Some configs may not trigger this section — skip gracefully
				t.Skip("No language-specific tools section generated — skipping")
			}

			// Find all npm install -g occurrences (excluding the neovim one which has its own test)
			// and verify each has a fallback
			lines := strings.Split(dockerfile, "\n")
			for i, line := range lines {
				trimmed := strings.TrimSpace(line)
				// Match npm install -g lines that are NOT the neovim Mason install
				if strings.HasPrefix(trimmed, "npm install -g") && !strings.Contains(trimmed, "neovim") {
					// Check that the next meaningful line contains the fallback
					// i.e., the line should end with || \ or the next line should be || (
					hasTrailingFallback := strings.Contains(line, "||") ||
						(i+1 < len(lines) && strings.Contains(lines[i+1], "||"))
					if !hasTrailingFallback {
						t.Errorf("npm install -g command at line %d has no proxy-unset fallback: %q\n"+
							"All npm install commands must use '|| (unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy NPM_CONFIG_REGISTRY npm_config_registry && npm install ...)' fallback.\n"+
							"See pip fallback pattern in installLanguageTools() python case for reference.",
							i+1, line)
					}

					// Verify the specific vars that must be unset
					fallbackRegion := strings.Join(lines[i:min(i+5, len(lines))], "\n")
					requiredUnsets := []string{
						"HTTP_PROXY",
						"HTTPS_PROXY",
						"http_proxy",
						"https_proxy",
						"NPM_CONFIG_REGISTRY",
						"npm_config_registry",
					}
					for _, envVar := range requiredUnsets {
						if !strings.Contains(fallbackRegion, envVar) {
							t.Errorf("npm install -g fallback at line %d missing unset for %q.\n"+
								"Must unset: HTTP_PROXY HTTPS_PROXY http_proxy https_proxy NPM_CONFIG_REGISTRY npm_config_registry\n"+
								"Fallback region:\n%s",
								i+1, envVar, fallbackRegion)
						}
					}
				}
			}
		})
	}
}

// =============================================================================
// Issue #147 — nvim-section RUN commands missing proxy-unset prefix (TDD Phase 2)
// =============================================================================

// TestDockerfileGenerator_LazySyncStep_ProxyUnsetPrefix verifies that the
// `nvim --headless "+Lazy! sync"` RUN command in generateNvimSection() is prefixed
// with `unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy NPM_CONFIG_REGISTRY
// npm_config_registry &&` before the nvim invocation.
//
// Without the prefix, lazy.nvim runs `git clone` for plugins while inheriting
// HTTP_PROXY/HTTPS_PROXY ARG values that point at an unreachable Squid proxy
// inside the Colima heavy-local VM, causing plugin installation failures.
//
// Phase 2 failing test for Issue #147 — fix NOT yet implemented.
// MUST FAIL against current code (line 1194 has no unset prefix).
func TestDockerfileGenerator_LazySyncStep_ProxyUnsetPrefix(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name     string
		language string
		version  string
	}{
		{
			name:     "python/debian — lazy sync in dev stage",
			language: "python",
			version:  "3.11",
		},
		{
			name:     "golang — lazy sync in dev stage",
			language: "golang",
			version:  "1.22",
		},
		{
			name:     "nodejs — lazy sync in dev stage",
			language: "nodejs",
			version:  "20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create staging dir with nvim config to trigger generateNvimSection()
			repoName := "test-lazy-proxy-" + tt.language
			stagingDir := filepath.Join(homeDir, ".devopsmaestro", "build-staging", repoName)
			nvimConfigPath := filepath.Join(stagingDir, ".config", "nvim")
			if err := os.MkdirAll(nvimConfigPath, 0755); err != nil {
				t.Fatalf("failed to create nvim config dir: %v", err)
			}
			defer os.RemoveAll(stagingDir)

			initLuaPath := filepath.Join(nvimConfigPath, "init.lua")
			if err := os.WriteFile(initLuaPath, []byte("-- test"), 0644); err != nil {
				t.Fatalf("failed to create init.lua: %v", err)
			}

			sourcePath := filepath.Join("/tmp", "dvm-clone-xyz", repoName)

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

			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
				Workspace:     ws,
				WorkspaceSpec: wsYAML,
				Language:      tt.language,
				Version:       tt.version,
				AppPath:       sourcePath,
				PathConfig:    paths.New(homeDir),
			})

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			// Verify the Lazy! sync command is present at all
			if !strings.Contains(dockerfile, `"+Lazy! sync"`) {
				t.Fatalf("Generate() missing '+Lazy! sync' command — test requires it to be present for language=%s", tt.language)
			}

			// MUST have: unset prefix on the RUN containing "+Lazy! sync"
			// Expected form:
			//   RUN unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy NPM_CONFIG_REGISTRY npm_config_registry && \
			//       nvim --headless "+Lazy! sync" +qa 2>&1 | ...
			//
			// Locate the RUN block containing "+Lazy! sync" and check the preceding context
			lazyIdx := strings.Index(dockerfile, `"+Lazy! sync"`)
			if lazyIdx < 0 {
				t.Fatalf("cannot locate '+Lazy! sync' in generated Dockerfile")
			}

			// Look back up to 300 characters before the command for the unset prefix
			lookbackStart := lazyIdx - 300
			if lookbackStart < 0 {
				lookbackStart = 0
			}
			context := dockerfile[lookbackStart : lazyIdx+len(`"+Lazy! sync"`)]

			// Find the last RUN keyword before the +Lazy! sync command
			lastRunIdx := strings.LastIndex(context, "RUN ")
			if lastRunIdx < 0 {
				t.Fatalf("no RUN keyword found before '+Lazy! sync' in context:\n%s", context)
			}
			runBlock := context[lastRunIdx:]

			requiredUnsets := []string{
				"unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy NPM_CONFIG_REGISTRY npm_config_registry",
			}
			for _, want := range requiredUnsets {
				if !strings.Contains(runBlock, want) {
					t.Errorf("[%s] Lazy! sync RUN command missing proxy-unset prefix: %q\n"+
						"The '+Lazy! sync' step must unset all proxy/registry env vars before running\n"+
						"because lazy.nvim runs 'git clone' which honors HTTP_PROXY/HTTPS_PROXY.\n"+
						"Expected prefix: unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy NPM_CONFIG_REGISTRY npm_config_registry\n"+
						"RUN block found:\n%s",
						tt.name, want, runBlock)
				}
			}
		})
	}
}

// TestDockerfileGenerator_MasonInstallStep_ProxyUnsetPrefix verifies that the
// Mason install RUN command in installMasonTools() is prefixed with
// `unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy NPM_CONFIG_REGISTRY
// npm_config_registry &&` before the nvim invocation.
//
// Without the prefix, Mason internally spawns `npm install` for npm-based
// packages (pyright, typescript-language-server, etc.) which inherit the
// NPM_CONFIG_REGISTRY ARG pointing at an unreachable Verdaccio proxy, causing
// 503 errors at build time.
func TestDockerfileGenerator_MasonInstallStep_ProxyUnsetPrefix(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name     string
		language string
		version  string
	}{
		{
			name:     "python — mason install in dev stage",
			language: "python",
			version:  "3.11",
		},
		{
			name:     "golang — mason install in dev stage",
			language: "golang",
			version:  "1.22",
		},
		{
			name:     "nodejs — mason install in dev stage",
			language: "nodejs",
			version:  "20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create staging dir with nvim config to trigger installMasonTools()
			repoName := "test-mason-proxy-" + tt.language
			stagingDir := filepath.Join(homeDir, ".devopsmaestro", "build-staging", repoName)
			nvimConfigPath := filepath.Join(stagingDir, ".config", "nvim")
			if err := os.MkdirAll(nvimConfigPath, 0755); err != nil {
				t.Fatalf("failed to create nvim config dir: %v", err)
			}
			defer os.RemoveAll(stagingDir)

			initLuaPath := filepath.Join(nvimConfigPath, "init.lua")
			if err := os.WriteFile(initLuaPath, []byte("-- test"), 0644); err != nil {
				t.Fatalf("failed to create init.lua: %v", err)
			}

			sourcePath := filepath.Join("/tmp", "dvm-clone-xyz", repoName)

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

			// Enable Mason so installMasonTools() emits the MasonInstall RUN
			manifest := &plugin.PluginManifest{
				Features: plugin.PluginFeatures{
					HasMason:      true,
					HasTreesitter: false,
				},
			}

			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
				Workspace:     ws,
				WorkspaceSpec: wsYAML,
				Language:      tt.language,
				Version:       tt.version,
				AppPath:       sourcePath,
				PathConfig:    paths.New(homeDir),
			})
			gen.SetPluginManifest(manifest)

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			// Verify the Mason install Lua script is present at all
			if !strings.Contains(dockerfile, "mason-registry") {
				t.Fatalf("Generate() missing 'mason-registry' Lua script — test requires it to be present for language=%s", tt.language)
			}

			// MUST have: unset prefix on the RUN that executes nvim for Mason install.
			// With issue #204 fix, the Lua script is written via COPY heredoc and
			// the nvim execution is a separate RUN with the unset prefix.
			// Expected form:
			//   COPY <<'LUAEOF' /tmp/mason-install.lua
			//       ... lua content with mason-registry ...
			//   LUAEOF
			//
			//   RUN unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy NPM_CONFIG_REGISTRY npm_config_registry && \
			//       nvim --headless +"luafile /tmp/mason-install.lua" +qa 2>&1 && \
			//       rm -f /tmp/mason-install.lua
			nvimMasonIdx := strings.Index(dockerfile, "luafile /tmp/mason-install.lua")
			if nvimMasonIdx < 0 {
				t.Fatalf("cannot locate 'luafile /tmp/mason-install.lua' in generated Dockerfile")
			}

			// Look back up to 300 characters before the nvim command for the unset prefix
			lookbackStart := nvimMasonIdx - 300
			if lookbackStart < 0 {
				lookbackStart = 0
			}
			context := dockerfile[lookbackStart : nvimMasonIdx+len("luafile /tmp/mason-install.lua")]

			// Find the last RUN keyword before the nvim mason install command
			lastRunIdx := strings.LastIndex(context, "RUN ")
			if lastRunIdx < 0 {
				t.Fatalf("no RUN keyword found before 'luafile /tmp/mason-install.lua' in context:\n%s", context)
			}
			runBlock := context[lastRunIdx:]

			requiredUnsets := []string{
				"unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy NPM_CONFIG_REGISTRY npm_config_registry",
			}
			for _, want := range requiredUnsets {
				if !strings.Contains(runBlock, want) {
					t.Errorf("[%s] Mason install RUN command missing proxy-unset prefix: %q\n"+
						"Mason spawns 'npm install' internally for npm-based packages (pyright, tsserver, etc.).\n"+
						"Those subprocesses inherit NPM_CONFIG_REGISTRY/HTTP_PROXY ARG values that point at\n"+
						"unreachable host.docker.internal registries, causing 503 failures at build time.\n"+
						"Expected prefix: unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy NPM_CONFIG_REGISTRY npm_config_registry\n"+
						"RUN block found:\n%s",
						tt.name, want, runBlock)
				}
			}
		})
	}
}

// TestDockerfileGenerator_TreesitterInstallStep_ProxyUnsetPrefix verifies that the
// `nvim --headless -c "luafile /tmp/treesitter-install.lua"` RUN command in
// installTreesitterParsers() is prefixed with `unset HTTP_PROXY HTTPS_PROXY
// http_proxy https_proxy NPM_CONFIG_REGISTRY npm_config_registry &&` before the
// nvim invocation.
//
// Without the prefix, nvim-treesitter compiles parsers from source by fetching
// grammar repos via `git clone` from GitHub. With HTTP_PROXY/HTTPS_PROXY set to
// an unreachable Squid proxy, these git clones will timeout or fail.
//
// Phase 2 failing test for Issue #147.
func TestDockerfileGenerator_TreesitterInstallStep_ProxyUnsetPrefix(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name     string
		language string
		version  string
	}{
		{
			name:     "python — treesitter install in dev stage",
			language: "python",
			version:  "3.11",
		},
		{
			name:     "golang — treesitter install in dev stage",
			language: "golang",
			version:  "1.22",
		},
		{
			name:     "nodejs — treesitter install in dev stage",
			language: "nodejs",
			version:  "20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create staging dir with nvim config to trigger installTreesitterParsers()
			repoName := "test-ts-proxy-" + tt.language
			stagingDir := filepath.Join(homeDir, ".devopsmaestro", "build-staging", repoName)
			nvimConfigPath := filepath.Join(stagingDir, ".config", "nvim")
			if err := os.MkdirAll(nvimConfigPath, 0755); err != nil {
				t.Fatalf("failed to create nvim config dir: %v", err)
			}
			defer os.RemoveAll(stagingDir)

			initLuaPath := filepath.Join(nvimConfigPath, "init.lua")
			if err := os.WriteFile(initLuaPath, []byte("-- test"), 0644); err != nil {
				t.Fatalf("failed to create init.lua: %v", err)
			}

			sourcePath := filepath.Join("/tmp", "dvm-clone-xyz", repoName)

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

			// Enable Treesitter so installTreesitterParsers() emits the TSInstall RUN
			manifest := &plugin.PluginManifest{
				Features: plugin.PluginFeatures{
					HasMason:      false,
					HasTreesitter: true,
				},
			}

			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
				Workspace:     ws,
				WorkspaceSpec: wsYAML,
				Language:      tt.language,
				Version:       tt.version,
				AppPath:       sourcePath,
				PathConfig:    paths.New(homeDir),
			})
			gen.SetPluginManifest(manifest)

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			// Verify the treesitter install Lua script is present (master branch API)
			if !strings.Contains(dockerfile, "require('nvim-treesitter.configs').setup(") {
				t.Fatalf("Generate() missing require('nvim-treesitter.configs').setup( in Lua script — test requires it to be present for language=%s", tt.language)
			}

			// MUST have: unset prefix on the RUN containing the treesitter install
			// Expected form:
			//   RUN unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy NPM_CONFIG_REGISTRY npm_config_registry && \
			//       nvim --headless -c "luafile /tmp/treesitter-install.lua" 2>&1
			tsIdx := strings.Index(dockerfile, "luafile /tmp/treesitter-install.lua")
			if tsIdx < 0 {
				t.Fatalf("cannot locate treesitter luafile invocation in generated Dockerfile")
			}

			// Look back up to 300 characters before the command for the unset prefix
			lookbackStart := tsIdx - 300
			if lookbackStart < 0 {
				lookbackStart = 0
			}
			context := dockerfile[lookbackStart : tsIdx+len("luafile /tmp/treesitter-install.lua")]

			// Find the last RUN keyword before the treesitter install command
			lastRunIdx := strings.LastIndex(context, "RUN ")
			if lastRunIdx < 0 {
				t.Fatalf("no RUN keyword found before treesitter install in context:\n%s", context)
			}
			runBlock := context[lastRunIdx:]

			requiredUnsets := []string{
				"unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy NPM_CONFIG_REGISTRY npm_config_registry",
			}
			for _, want := range requiredUnsets {
				if !strings.Contains(runBlock, want) {
					t.Errorf("[%s] treesitter RUN command missing proxy-unset prefix: %q\n"+
						"nvim-treesitter compiles parsers from source by fetching grammar repos via 'git clone'.\n"+
						"With HTTP_PROXY/HTTPS_PROXY inherited from the dev stage ARG declarations, these\n"+
						"git clones route through an unreachable Squid proxy and fail.\n"+
						"Expected prefix: unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy NPM_CONFIG_REGISTRY npm_config_registry\n"+
						"RUN block found:\n%s",
						tt.name, want, runBlock)
				}
			}
		})
	}
}

// =============================================================================
// Issue #31: Mason install reliability + extensibility tests
// =============================================================================

// TestInstallMasonTools_SynchronousLuaScript verifies that installMasonTools()
// generates a synchronous Lua-based install using mason-registry and vim.wait()
// instead of the old unreliable `MasonInstall ... sleep 60` pattern.
func TestInstallMasonTools_SynchronousLuaScript(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	repoName := "test-mason-sync-lua"
	stagingDir := filepath.Join(homeDir, ".devopsmaestro", "build-staging", repoName)
	nvimConfigPath := filepath.Join(stagingDir, ".config", "nvim")
	if err := os.MkdirAll(nvimConfigPath, 0755); err != nil {
		t.Fatalf("failed to create nvim config dir: %v", err)
	}
	defer os.RemoveAll(stagingDir)

	initLuaPath := filepath.Join(nvimConfigPath, "init.lua")
	if err := os.WriteFile(initLuaPath, []byte("-- test"), 0644); err != nil {
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

	manifest := &plugin.PluginManifest{
		Features: plugin.PluginFeatures{
			HasMason:      true,
			HasTreesitter: false,
		},
	}

	sourcePath := filepath.Join("/tmp", "dvm-clone-xyz", repoName)

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       sourcePath,
		PathConfig:    paths.New(homeDir),
	})
	gen.SetPluginManifest(manifest)

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// MUST contain synchronous Lua-based install markers
	requiredStrings := []string{
		"mason-registry",
		"vim.wait(",
		"registry.refresh",
		"pkg:install()",
		"pkg:is_installed()",
		"mason-install.lua",
		"luafile /tmp/mason-install.lua",
	}
	for _, want := range requiredStrings {
		if !strings.Contains(dockerfile, want) {
			t.Errorf("Mason install missing required content %q.\n"+
				"Issue #31: Must use synchronous Lua-based install with mason-registry and vim.wait().",
				want)
		}
	}

	// MUST NOT contain the old unreliable pattern
	forbiddenStrings := []string{
		"sleep 60",
		"MasonInstall",
	}
	for _, bad := range forbiddenStrings {
		if strings.Contains(dockerfile, bad) {
			t.Errorf("Mason install still contains old unreliable pattern %q.\n"+
				"Issue #31: Replace 'MasonInstall + sleep 60' with synchronous Lua-based install.",
				bad)
		}
	}
}

// TestGetMasonToolsForLanguage_AllLanguages is a table-driven test verifying the
// exact Mason tool list for every supported language, including the expanded
// ruby (rubocop) and java (google-java-format) lists.
func TestGetMasonToolsForLanguage_AllLanguages(t *testing.T) {
	tests := []struct {
		name      string
		language  string
		wantTools []string
	}{
		{
			name:     "python",
			language: "python",
			wantTools: []string{
				"pyright", "ruff", "black", "isort", "pylint",
			},
		},
		{
			name:     "golang",
			language: "golang",
			wantTools: []string{
				"gopls", "golangci-lint-langserver", "goimports",
			},
		},
		{
			name:     "nodejs",
			language: "nodejs",
			wantTools: []string{
				"typescript-language-server", "eslint-lsp", "prettier",
			},
		},
		{
			name:     "rust",
			language: "rust",
			wantTools: []string{
				"rust-analyzer",
			},
		},
		{
			name:     "ruby includes rubocop",
			language: "ruby",
			wantTools: []string{
				"solargraph", "rubocop",
			},
		},
		{
			name:     "java includes google-java-format",
			language: "java",
			wantTools: []string{
				"jdtls", "google-java-format",
			},
		},
		{
			name:     "gleam",
			language: "gleam",
			wantTools: []string{
				"gleam",
			},
		},
		{
			name:     "dotnet includes omnisharp and netcoredbg",
			language: "dotnet",
			wantTools: []string{
				"omnisharp", "netcoredbg",
			},
		},
		{
			name:     "php includes intelephense and php-cs-fixer",
			language: "php",
			wantTools: []string{
				"intelephense", "php-cs-fixer",
			},
		},
		{
			name:     "swift includes sourcekit-lsp",
			language: "swift",
			wantTools: []string{
				"sourcekit-lsp",
			},
		},
		{
			name:     "dart includes dart-debug-adapter",
			language: "dart",
			wantTools: []string{
				"dart-debug-adapter",
			},
		},
		{
			name:     "haskell includes haskell-language-server",
			language: "haskell",
			wantTools: []string{
				"haskell-language-server",
			},
		},
		{
			name:     "perl includes perlnavigator",
			language: "perl",
			wantTools: []string{
				"perlnavigator",
			},
		},
		{
			name:      "unknown language returns empty",
			language:  "cobol",
			wantTools: []string{},
		},
		{
			name:      "empty language returns empty",
			language:  "",
			wantTools: []string{},
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
				AppPath:       "/tmp/test",
				PathConfig:    paths.New(t.TempDir()),
			})

			impl := gen.(*DefaultDockerfileGenerator)
			tools := impl.getMasonToolsForLanguage()

			// Verify exact length for non-empty expected lists
			if len(tt.wantTools) > 0 && len(tools) != len(tt.wantTools) {
				t.Errorf("[%s] getMasonToolsForLanguage() returned %d tools, want %d.\n"+
					"Got: %v\nWant: %v",
					tt.name, len(tools), len(tt.wantTools), tools, tt.wantTools)
			}

			// Verify all expected tools are present
			toolSet := make(map[string]bool, len(tools))
			for _, tool := range tools {
				toolSet[tool] = true
			}
			for _, want := range tt.wantTools {
				if !toolSet[want] {
					t.Errorf("[%s] getMasonToolsForLanguage() missing %q.\n"+
						"Got tools: %v", tt.name, want, tools)
				}
			}

			// For unknown/empty language, verify empty result
			if len(tt.wantTools) == 0 && len(tools) != 0 {
				t.Errorf("[%s] getMasonToolsForLanguage() should return empty for unknown language.\n"+
					"Got tools: %v", tt.name, tools)
			}
		})
	}
}

// TestInstallMasonTools_ExtraToolsFromConfig verifies that user-configured
// ExtraMasonTools from the workspace YAML are appended to the Mason install.
func TestInstallMasonTools_ExtraToolsFromConfig(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	repoName := "test-mason-extra-tools"
	stagingDir := filepath.Join(homeDir, ".devopsmaestro", "build-staging", repoName)
	nvimConfigPath := filepath.Join(stagingDir, ".config", "nvim")
	if err := os.MkdirAll(nvimConfigPath, 0755); err != nil {
		t.Fatalf("failed to create nvim config dir: %v", err)
	}
	defer os.RemoveAll(stagingDir)

	initLuaPath := filepath.Join(nvimConfigPath, "init.lua")
	if err := os.WriteFile(initLuaPath, []byte("-- test"), 0644); err != nil {
		t.Fatalf("failed to create init.lua: %v", err)
	}

	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{
		Nvim: models.NvimConfig{
			Structure:       "custom",
			ExtraMasonTools: []string{"shellcheck", "shfmt"},
		},
	}

	manifest := &plugin.PluginManifest{
		Features: plugin.PluginFeatures{
			HasMason:      true,
			HasTreesitter: false,
		},
	}

	sourcePath := filepath.Join("/tmp", "dvm-clone-xyz", repoName)

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       sourcePath,
		PathConfig:    paths.New(homeDir),
	})
	gen.SetPluginManifest(manifest)

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Locate the Lua tools list
	toolsIdx := strings.Index(dockerfile, "local tools = {")
	if toolsIdx < 0 {
		t.Fatalf("Generate() missing Mason Lua tools list")
	}
	toolsLineEnd := strings.Index(dockerfile[toolsIdx:], "\n")
	var toolsLine string
	if toolsLineEnd > 0 {
		toolsLine = dockerfile[toolsIdx : toolsIdx+toolsLineEnd]
	} else {
		toolsLine = dockerfile[toolsIdx:]
	}

	// Extra tools must appear in the Lua tools list
	extraTools := []string{"shellcheck", "shfmt"}
	for _, tool := range extraTools {
		if !strings.Contains(toolsLine, tool) {
			t.Errorf("Mason install missing extra tool %q from ExtraMasonTools config.\n"+
				"Issue #31: ExtraMasonTools from workspace YAML must be appended to Mason install.\n"+
				"Tools line: %s", tool, toolsLine)
		}
	}

	// Base tools must still be present
	baseTools := []string{"lua-language-server", "stylua"}
	for _, tool := range baseTools {
		if !strings.Contains(toolsLine, tool) {
			t.Errorf("Mason install missing base tool %q when extra tools are configured.\n"+
				"Tools line: %s", tool, toolsLine)
		}
	}

	// Language tools must still be present
	langTools := []string{"pyright", "ruff"}
	for _, tool := range langTools {
		if !strings.Contains(toolsLine, tool) {
			t.Errorf("Mason install missing language tool %q when extra tools are configured.\n"+
				"Tools line: %s", tool, toolsLine)
		}
	}
}

// TestInstallMasonTools_RubyIncludesRubocop verifies ruby language
// tools include rubocop in the generated Dockerfile.
func TestInstallMasonTools_RubyIncludesRubocop(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "ruby",
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

	impl := gen.(*DefaultDockerfileGenerator)
	tools := impl.getMasonToolsForLanguage()

	toolSet := make(map[string]bool, len(tools))
	for _, tool := range tools {
		toolSet[tool] = true
	}

	if !toolSet["solargraph"] {
		t.Errorf("Ruby tools missing 'solargraph'. Got: %v", tools)
	}
	if !toolSet["rubocop"] {
		t.Errorf("Ruby tools missing 'rubocop'.\n"+
			"Issue #31: Ruby language tools must include rubocop.\n"+
			"Got: %v", tools)
	}
}

// =============================================================================
// Issue #32: Treesitter parsers — extensibility, error handling, format tests
// =============================================================================

// TestGetTreesitterParsersForLanguage_AllLanguages verifies that each supported
// language returns the correct set of Treesitter parsers (base + language-specific),
// and that an unknown/empty language returns only base parsers.
func TestGetTreesitterParsersForLanguage_AllLanguages(t *testing.T) {
	baseParsers := []string{"lua", "vim", "vimdoc", "query", "markdown", "markdown_inline", "bash", "json", "yaml"}

	tests := []struct {
		name             string
		language         string
		wantLangSpecific []string
		wantNotPresent   []string
	}{
		{
			name:             "python includes python and toml",
			language:         "python",
			wantLangSpecific: []string{"python", "toml", "dockerfile", "gitignore"},
		},
		{
			name:             "golang includes go, gomod, gosum, gowork",
			language:         "golang",
			wantLangSpecific: []string{"go", "gomod", "gosum", "gowork", "dockerfile", "gitignore"},
		},
		{
			name:             "nodejs includes javascript, typescript, tsx, html, css",
			language:         "nodejs",
			wantLangSpecific: []string{"javascript", "typescript", "tsx", "html", "css", "dockerfile", "gitignore"},
		},
		{
			name:             "rust includes rust and toml",
			language:         "rust",
			wantLangSpecific: []string{"rust", "toml", "dockerfile", "gitignore"},
		},
		{
			name:             "ruby includes ruby",
			language:         "ruby",
			wantLangSpecific: []string{"ruby", "dockerfile", "gitignore"},
		},
		{
			name:             "java includes java and xml",
			language:         "java",
			wantLangSpecific: []string{"java", "xml", "dockerfile", "gitignore"},
		},
		{
			name:             "gleam includes gleam, erlang, elixir, toml",
			language:         "gleam",
			wantLangSpecific: []string{"gleam", "erlang", "elixir", "toml", "dockerfile", "gitignore"},
		},
		{
			name:             "dotnet includes c_sharp and xml",
			language:         "dotnet",
			wantLangSpecific: []string{"c_sharp", "xml", "dockerfile", "gitignore"},
		},
		{
			name:             "php includes php, phpdoc, html, css, javascript",
			language:         "php",
			wantLangSpecific: []string{"php", "phpdoc", "html", "css", "javascript", "dockerfile", "gitignore"},
		},
		{
			name:             "swift includes swift",
			language:         "swift",
			wantLangSpecific: []string{"swift", "dockerfile", "gitignore"},
		},
		{
			name:             "dart includes dart",
			language:         "dart",
			wantLangSpecific: []string{"dart", "dockerfile", "gitignore"},
		},
		{
			name:             "perl includes perl and pod",
			language:         "perl",
			wantLangSpecific: []string{"perl", "pod", "dockerfile", "gitignore"},
		},
		{
			name:             "unknown language returns only base parsers",
			language:         "unknown",
			wantLangSpecific: nil,
			wantNotPresent:   []string{"python", "go", "javascript", "rust", "ruby", "java", "gleam", "swift", "dart"},
		},
		{
			name:             "empty language returns only base parsers",
			language:         "",
			wantLangSpecific: nil,
			wantNotPresent:   []string{"python", "go", "javascript", "rust", "ruby", "java", "gleam"},
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
				AppPath:       "/tmp/test",
				PathConfig:    paths.New(t.TempDir()),
			})

			impl := gen.(*DefaultDockerfileGenerator)
			parsers := impl.getTreesitterParsersForLanguage()

			parserSet := make(map[string]bool, len(parsers))
			for _, p := range parsers {
				parserSet[p] = true
			}

			// Verify all base parsers are always present
			for _, base := range baseParsers {
				if !parserSet[base] {
					t.Errorf("base parser %q missing for language=%q. Got: %v", base, tt.language, parsers)
				}
			}

			// Verify language-specific parsers are present
			for _, lang := range tt.wantLangSpecific {
				if !parserSet[lang] {
					t.Errorf("language-specific parser %q missing for language=%q. Got: %v", lang, tt.language, parsers)
				}
			}

			// Verify unwanted parsers are NOT present
			for _, notWant := range tt.wantNotPresent {
				if parserSet[notWant] {
					t.Errorf("parser %q should NOT be present for language=%q. Got: %v", notWant, tt.language, parsers)
				}
			}
		})
	}
}

// TestInstallTreesitterParsers_ExtraParsersFromConfig verifies that
// ExtraTreesitterParsers from the workspace config are appended to the
// parser list in the generated Dockerfile.
func TestInstallTreesitterParsers_ExtraParsersFromConfig(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	repoName := "test-ts-extra-parsers"
	stagingDir := filepath.Join(homeDir, ".devopsmaestro", "build-staging", repoName)
	nvimConfigPath := filepath.Join(stagingDir, ".config", "nvim")
	if err := os.MkdirAll(nvimConfigPath, 0755); err != nil {
		t.Fatalf("failed to create nvim config dir: %v", err)
	}
	defer os.RemoveAll(stagingDir)

	initLuaPath := filepath.Join(nvimConfigPath, "init.lua")
	if err := os.WriteFile(initLuaPath, []byte("-- test"), 0644); err != nil {
		t.Fatalf("failed to create init.lua: %v", err)
	}

	sourcePath := filepath.Join("/tmp", "dvm-clone-xyz", repoName)

	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{
		Nvim: models.NvimConfig{
			Structure:              "custom",
			ExtraTreesitterParsers: []string{"hcl", "terraform"},
		},
	}

	manifest := &plugin.PluginManifest{
		Features: plugin.PluginFeatures{
			HasMason:      false,
			HasTreesitter: true,
		},
	}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "golang",
		Version:       "1.22",
		AppPath:       sourcePath,
		PathConfig:    paths.New(homeDir),
	})
	gen.SetPluginManifest(manifest)

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify extra parsers appear in the treesitter install command
	for _, parser := range []string{"hcl", "terraform"} {
		if !strings.Contains(dockerfile, parser) {
			t.Errorf("extra parser %q not found in generated Dockerfile treesitter install command", parser)
		}
	}

	// Also verify language-specific parsers are still present
	for _, parser := range []string{"go", "gomod"} {
		if !strings.Contains(dockerfile, parser) {
			t.Errorf("language parser %q should still be present alongside extra parsers", parser)
		}
	}
}

// TestInstallTreesitterParsers_ErrorHandling verifies that the generated
// treesitter install command includes error capture (tee + || exit 1) so
// that parser compilation failures cause the Docker build to fail.
func TestInstallTreesitterParsers_ErrorHandling(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	repoName := "test-ts-error-handling"
	stagingDir := filepath.Join(homeDir, ".devopsmaestro", "build-staging", repoName)
	nvimConfigPath := filepath.Join(stagingDir, ".config", "nvim")
	if err := os.MkdirAll(nvimConfigPath, 0755); err != nil {
		t.Fatalf("failed to create nvim config dir: %v", err)
	}
	defer os.RemoveAll(stagingDir)

	initLuaPath := filepath.Join(nvimConfigPath, "init.lua")
	if err := os.WriteFile(initLuaPath, []byte("-- test"), 0644); err != nil {
		t.Fatalf("failed to create init.lua: %v", err)
	}

	sourcePath := filepath.Join("/tmp", "dvm-clone-xyz", repoName)

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

	manifest := &plugin.PluginManifest{
		Features: plugin.PluginFeatures{
			HasMason:      false,
			HasTreesitter: true,
		},
	}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       sourcePath,
		PathConfig:    paths.New(homeDir),
	})
	gen.SetPluginManifest(manifest)

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Find the treesitter Lua script section (master branch API: configs.setup)
	tsIdx := strings.Index(dockerfile, "require('nvim-treesitter.configs').setup(")
	if tsIdx < 0 {
		t.Fatalf("treesitter install Lua script not found in Dockerfile (expected require('nvim-treesitter.configs').setup()")
	}

	// Verify tee pattern for log capture
	if !strings.Contains(dockerfile, "tee /tmp/treesitter-install.log") {
		t.Errorf("treesitter install command missing 'tee /tmp/treesitter-install.log' for error capture")
	}

	// Verify exit 1 on failure (in both Lua script and verification step)
	if !strings.Contains(dockerfile, "exit 1") {
		t.Errorf("treesitter install missing 'exit 1' failure handler")
	}

	// Verify cat of log on failure
	if !strings.Contains(dockerfile, "cat /tmp/treesitter-install.log") {
		t.Errorf("treesitter install missing 'cat /tmp/treesitter-install.log' for error output")
	}

	// Verify Lua script has error handling (vim.cmd('cq') on failure)
	if !strings.Contains(dockerfile, "vim.cmd('cq')") {
		t.Errorf("treesitter Lua script missing vim.cmd('cq') error exit")
	}
}

// TestInstallTreesitterParsers_OutputFormat verifies the command format
// uses a Lua script with the master branch nvim-treesitter API:
// require('nvim-treesitter.configs').setup({ ensure_installed, sync_install })
// to install parsers SYNCHRONOUSLY in headless mode. The Lua script explicitly prepends
// nvim-treesitter's path to runtimepath and package.path (issue #246).
//
// Docker builds pin Neovim 0.11.x. The nvim-treesitter `master` branch preserves
// the old configs API for Neovim <=0.11 compatibility. The `main` branch requires 0.12+.
func TestInstallTreesitterParsers_OutputFormat(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	repoName := "test-ts-output-format"
	stagingDir := filepath.Join(homeDir, ".devopsmaestro", "build-staging", repoName)
	nvimConfigPath := filepath.Join(stagingDir, ".config", "nvim")
	if err := os.MkdirAll(nvimConfigPath, 0755); err != nil {
		t.Fatalf("failed to create nvim config dir: %v", err)
	}
	defer os.RemoveAll(stagingDir)

	initLuaPath := filepath.Join(nvimConfigPath, "init.lua")
	if err := os.WriteFile(initLuaPath, []byte("-- test"), 0644); err != nil {
		t.Fatalf("failed to create init.lua: %v", err)
	}

	sourcePath := filepath.Join("/tmp", "dvm-clone-xyz", repoName)

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

	manifest := &plugin.PluginManifest{
		Features: plugin.PluginFeatures{
			HasMason:      false,
			HasTreesitter: true,
		},
	}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       sourcePath,
		PathConfig:    paths.New(homeDir),
	})
	gen.SetPluginManifest(manifest)

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify explicit runtimepath prepend is used instead of require('lazy').load()
	// (issue #246, attempt 4: lazy.load() does NOT synchronously update runtimepath/package.path)
	if !strings.Contains(dockerfile, "vim.opt.runtimepath:prepend(ts_path)") {
		t.Error("treesitter Lua script missing vim.opt.runtimepath:prepend(ts_path) — " +
			"explicit runtimepath manipulation required because require('lazy').load() does not " +
			"synchronously update runtimepath (issue #246, attempt 4)")
	}

	// Verify stdpath('data') is used to locate the nvim-treesitter install path
	if !strings.Contains(dockerfile, "vim.fn.stdpath('data')") {
		t.Error("treesitter Lua script missing vim.fn.stdpath('data') for locating nvim-treesitter")
	}

	// Verify package.path is updated for require() to find nvim-treesitter modules
	if !strings.Contains(dockerfile, "package.path") {
		t.Error("treesitter Lua script missing package.path update for nvim-treesitter modules")
	}

	// REGRESSION GUARD: require('lazy').load() must NOT be used — it doesn't synchronously
	// update runtimepath/package.path (issue #246, attempts 1-3)
	if strings.Contains(dockerfile, "require('lazy').load(") {
		t.Error("REGRESSION: treesitter must NOT use require('lazy').load() — " +
			"it does not synchronously update runtimepath/package.path (issue #246)")
	}

	// Verify Lazy! load is NOT on the nvim command line (it's now in the Lua script)
	// Look for the nvim --headless command and ensure it doesn't have -c "Lazy! load"
	if strings.Contains(dockerfile, `-c "Lazy! load nvim-treesitter"`) {
		t.Error("treesitter nvim command should NOT have -c 'Lazy! load' — loading is now done inside the Lua script (issue #246)")
	}

	// Verify synchronous Lua script is used (COPY heredoc pattern)
	if !strings.Contains(dockerfile, "COPY") || !strings.Contains(dockerfile, "treesitter-install.lua") {
		t.Error("treesitter install missing COPY heredoc for Lua script")
	}

	// Verify sync_install = true IS used (master branch configs API, issue #246)
	if !strings.Contains(dockerfile, "sync_install = true") {
		t.Error("treesitter Lua script missing sync_install = true — " +
			"required for synchronous parser install with master branch configs API (issue #246)")
	}

	// Verify ensure_installed IS used (master branch configs API, issue #246)
	if !strings.Contains(dockerfile, "ensure_installed") {
		t.Error("treesitter Lua script missing ensure_installed — " +
			"required for master branch configs API (issue #246)")
	}

	// Verify require('nvim-treesitter.configs') IS used (master branch API)
	if !strings.Contains(dockerfile, "require('nvim-treesitter.configs')") {
		t.Error("treesitter Lua script missing require('nvim-treesitter.configs') — " +
			"Docker pins Neovim 0.11.x which requires master branch API (issue #246)")
	}

	// REGRESSION GUARD: require('nvim-treesitter').setup({}) must NOT be used —
	// this is the Neovim 0.12+ API from the `main` branch (issue #246)
	if strings.Contains(dockerfile, "require('nvim-treesitter').setup({})") {
		t.Error("REGRESSION: treesitter must NOT use require('nvim-treesitter').setup({}) — " +
			"Docker pins Neovim 0.11.x; use master branch configs API instead (issue #246)")
	}

	// REGRESSION GUARD: require('nvim-treesitter').install() must NOT be used —
	// this is the Neovim 0.12+ API from the `main` branch (issue #246)
	if strings.Contains(dockerfile, "require('nvim-treesitter').install(parsers)") {
		t.Error("REGRESSION: treesitter must NOT use require('nvim-treesitter').install() — " +
			"Docker pins Neovim 0.11.x; use master branch configs API instead (issue #246)")
	}

	// REGRESSION GUARD: :wait() must NOT be used — this is the Neovim 0.12+ async API
	if strings.Contains(dockerfile, ":wait(") {
		t.Error("REGRESSION: treesitter must NOT use :wait() — " +
			"Docker pins Neovim 0.11.x; sync_install = true handles synchronous install (issue #246)")
	}

	// Verify luafile invocation uses -c flag (NOT + prefix).
	// REGRESSION GUARD (issue #246): The + prefix runs during Neovim STARTUP
	// before init.lua completes. The -c flag runs AFTER init.lua, ensuring Lazy
	// has synced plugins so nvim-treesitter exists on disk for runtimepath prepend.
	if !strings.Contains(dockerfile, `-c "luafile /tmp/treesitter-install.lua"`) {
		t.Error("treesitter command must use -c \"luafile ...\" (NOT +luafile) — " +
			"-c runs after init.lua completes Lazy sync (issue #246)")
	}

	// REGRESSION GUARD: ensure +luafile is NOT used for treesitter
	if strings.Contains(dockerfile, `+"luafile /tmp/treesitter-install.lua"`) ||
		strings.Contains(dockerfile, `+\"luafile /tmp/treesitter-install.lua\"`) {
		t.Error("REGRESSION: treesitter must NOT use +luafile — the + prefix runs during " +
			"startup BEFORE init.lua completes, so nvim-treesitter may not be synced yet (issue #246)")
	}

	// Verify +qa is NOT on the nvim command line (script controls exit via vim.cmd)
	if strings.Contains(dockerfile, "treesitter-install.lua\" +qa") {
		t.Error("treesitter nvim command should NOT have +qa — the Lua script must control exit timing")
	}

	// Verify .so file verification is still present
	if !strings.Contains(dockerfile, ".so") {
		t.Error("treesitter Lua script missing .so file check")
	}

	// Verify the script exits nvim after successful install
	if !strings.Contains(dockerfile, "vim.cmd('qa!')") {
		t.Error("treesitter Lua script missing vim.cmd('qa!') for clean exit after success")
	}

	// Verify parsers are present as Lua-quoted strings in the script
	// For python, we expect at minimum: lua, python
	for _, parser := range []string{"lua", "python"} {
		if !strings.Contains(dockerfile, "'"+parser+"'") {
			t.Errorf("parser %q not found as Lua string in generated Dockerfile", parser)
		}
	}

	// Verify error handling in Lua script
	if !strings.Contains(dockerfile, "vim.cmd('cq')") {
		t.Error("treesitter Lua script missing vim.cmd('cq') error exit")
	}
}

// TestInstallMasonTools_JavaIncludesGoogleJavaFormat verifies java language
// tools include google-java-format in the generated Dockerfile.
func TestInstallMasonTools_JavaIncludesGoogleJavaFormat(t *testing.T) {
	ws := &models.Workspace{
		ID:        1,
		Name:      "test-ws",
		ImageName: "test:latest",
	}
	wsYAML := models.WorkspaceSpec{}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "java",
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

	impl := gen.(*DefaultDockerfileGenerator)
	tools := impl.getMasonToolsForLanguage()

	toolSet := make(map[string]bool, len(tools))
	for _, tool := range tools {
		toolSet[tool] = true
	}

	if !toolSet["jdtls"] {
		t.Errorf("Java tools missing 'jdtls'. Got: %v", tools)
	}
	if !toolSet["google-java-format"] {
		t.Errorf("Java tools missing 'google-java-format'.\n"+
			"Issue #31: Java language tools must include google-java-format.\n"+
			"Got: %v", tools)
	}
}

// TestInstallMasonTools_NoUnknownDockerfileInstructions is a regression test for
// Issue #204: Dockerfile syntax error "unknown instruction: nvim".
//
// The original bug: installMasonTools() embedded a shell heredoc inside a multi-line
// RUN instruction. Docker line continuation (\) ended before the heredoc body, so
// Docker parsed Lua code lines and `nvim` as Dockerfile instructions.
//
// This test verifies that every line in the generated Dockerfile is either:
//   - A valid Dockerfile instruction (FROM, RUN, COPY, ENV, ARG, etc.)
//   - A comment (starts with #)
//   - A continuation line (previous line ended with \)
//   - Inside a BuildKit heredoc block (between `<< 'DELIM'` and `DELIM`)
//   - An empty line
//
// It specifically asserts that no line starts with bare `nvim`, `local`, `end`,
// `vim.wait`, or other Lua keywords that would indicate a broken heredoc.
func TestInstallMasonTools_NoUnknownDockerfileInstructions(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	repoName := "test-mason-dockerfile-syntax"
	stagingDir := filepath.Join(homeDir, ".devopsmaestro", "build-staging", repoName)
	nvimConfigPath := filepath.Join(stagingDir, ".config", "nvim")
	if err := os.MkdirAll(nvimConfigPath, 0755); err != nil {
		t.Fatalf("failed to create nvim config dir: %v", err)
	}
	defer os.RemoveAll(stagingDir)

	initLuaPath := filepath.Join(nvimConfigPath, "init.lua")
	if err := os.WriteFile(initLuaPath, []byte("-- test"), 0644); err != nil {
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

	manifest := &plugin.PluginManifest{
		Features: plugin.PluginFeatures{
			HasMason:      true,
			HasTreesitter: true,
		},
	}

	sourcePath := filepath.Join("/tmp", "dvm-clone-xyz", repoName)

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       sourcePath,
		PathConfig:    paths.New(homeDir),
	})
	gen.SetPluginManifest(manifest)

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Validate every line is a legal Dockerfile construct
	validInstructions := map[string]bool{
		"FROM": true, "RUN": true, "COPY": true, "ENV": true,
		"ARG": true, "WORKDIR": true, "USER": true, "LABEL": true,
		"EXPOSE": true, "ENTRYPOINT": true, "CMD": true, "SHELL": true,
		"ADD": true, "STOPSIGNAL": true, "HEALTHCHECK": true,
		"ONBUILD": true, "VOLUME": true,
	}

	lines := strings.Split(dockerfile, "\n")
	inContinuation := false // previous line ended with \
	inHeredoc := false      // inside a BuildKit heredoc block
	heredocDelimiter := ""  // the delimiter to close the heredoc

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Track heredoc state: if we're inside a heredoc, lines are content
		// until we see the closing delimiter
		if inHeredoc {
			if trimmed == heredocDelimiter {
				inHeredoc = false
				heredocDelimiter = ""
			}
			continue
		}

		// Empty lines are always valid
		if trimmed == "" {
			inContinuation = false
			continue
		}

		// Comments are always valid (including # syntax=docker/dockerfile:1)
		if strings.HasPrefix(trimmed, "#") {
			inContinuation = false
			continue
		}

		// If previous line ended with \, this is a continuation
		if inContinuation {
			inContinuation = strings.HasSuffix(trimmed, "\\")
			continue
		}

		// Check for heredoc start: look for << or <<- followed by a delimiter
		// in the instruction line (e.g., COPY <<'LUAEOF' /path or RUN <<EOF)
		if idx := strings.Index(trimmed, "<<"); idx >= 0 {
			rest := trimmed[idx+2:]
			// Skip <<- variant
			rest = strings.TrimPrefix(rest, "-")
			rest = strings.TrimSpace(rest)
			// Extract delimiter (may be quoted with ' or ")
			delim := rest
			if spaceIdx := strings.IndexAny(delim, " \t"); spaceIdx >= 0 {
				delim = delim[:spaceIdx]
			}
			delim = strings.Trim(delim, "'\"")
			if delim != "" {
				inHeredoc = true
				heredocDelimiter = delim
			}
		}

		// This should be a Dockerfile instruction — extract first word
		firstWord := trimmed
		if spaceIdx := strings.IndexAny(trimmed, " \t"); spaceIdx >= 0 {
			firstWord = trimmed[:spaceIdx]
		}
		firstWord = strings.ToUpper(firstWord)

		if !validInstructions[firstWord] {
			t.Errorf("Issue #204 regression: line %d starts with unknown "+
				"Dockerfile instruction %q.\nFull line: %s\n"+
				"This likely means a heredoc or RUN continuation is broken.",
				lineNum, firstWord, line)
		}

		inContinuation = strings.HasSuffix(trimmed, "\\")
	}

	// Explicit regression check: no bare `nvim` as a top-level Dockerfile instruction.
	// Lines that are continuations (preceded by a line ending with \) or inside
	// heredocs are valid — only flag `nvim` when Docker would parse it as an instruction.
	inCont := false
	inHDoc := false
	hDocDelim := ""
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if inHDoc {
			if trimmed == hDocDelim {
				inHDoc = false
				hDocDelim = ""
			}
			continue
		}

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			if !strings.HasSuffix(trimmed, "\\") {
				inCont = false
			}
			continue
		}

		if inCont {
			inCont = strings.HasSuffix(trimmed, "\\")
			continue
		}

		// Detect heredoc start
		if idx := strings.Index(trimmed, "<<"); idx >= 0 {
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed[idx+2:], "-"))
			delim := rest
			if sp := strings.IndexAny(delim, " \t"); sp >= 0 {
				delim = delim[:sp]
			}
			delim = strings.Trim(delim, "'\"")
			if delim != "" {
				inHDoc = true
				hDocDelim = delim
			}
		}

		// At this point the line is a top-level Dockerfile instruction
		if strings.HasPrefix(trimmed, "nvim ") || trimmed == "nvim" {
			t.Errorf("Issue #204 regression: line %d starts with bare 'nvim' "+
				"which Docker would interpret as an unknown instruction.\n"+
				"Line: %s", i+1, line)
		}

		inCont = strings.HasSuffix(trimmed, "\\")
	}

	// Verify Mason content IS present (test is meaningful)
	if !strings.Contains(dockerfile, "mason-registry") {
		t.Error("Test setup issue: generated Dockerfile missing mason-registry content")
	}
	if !strings.Contains(dockerfile, "luafile /tmp/mason-install.lua") {
		t.Error("Test setup issue: generated Dockerfile missing nvim mason-install execution")
	}
}

// =============================================================================
// UV Python integration tests (#273)
// Verify that UV is correctly wired into generated Python Dockerfiles.
// =============================================================================

// TestUVSetup_PythonBaseStage verifies that the UV binary COPY and all three
// ENV configuration variables are present in the base stage of Python Dockerfiles.
func TestUVSetup_PythonBaseStage(t *testing.T) {
	ws := &models.Workspace{ID: 1, Name: "test-ws", ImageName: "test:latest"}
	wsYAML := models.WorkspaceSpec{}

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

	uvLines := []string{
		"COPY --from=ghcr.io/astral-sh/uv:0.7.2 /uv /uvx /bin/",
		"ENV UV_LINK_MODE=copy",
		"ENV UV_COMPILE_BYTECODE=1",
		"ENV UV_SYSTEM_PYTHON=1",
	}
	for _, want := range uvLines {
		if !strings.Contains(dockerfile, want) {
			t.Errorf("Python base stage missing UV setup line: %q\nDockerfile:\n%s", want, dockerfile)
		}
	}
}

// TestUVSetup_NotPresentForNonPythonLanguages verifies that UV binary COPY is
// NOT injected into Golang or Node.js Dockerfiles.
func TestUVSetup_NotPresentForNonPythonLanguages(t *testing.T) {
	tests := []struct {
		name     string
		language string
		version  string
	}{
		{name: "golang", language: "golang", version: "1.22"},
		{name: "nodejs", language: "nodejs", version: "20"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &models.Workspace{ID: 1, Name: "test-ws", ImageName: "test:latest"}
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

			uvLines := []string{
				"ghcr.io/astral-sh/uv",
				"UV_LINK_MODE",
				"UV_COMPILE_BYTECODE",
				"UV_SYSTEM_PYTHON",
			}
			for _, notWant := range uvLines {
				if strings.Contains(dockerfile, notWant) {
					t.Errorf("%s Dockerfile should NOT contain UV line: %q", tt.language, notWant)
				}
			}
		})
	}
}

// TestUVInstall_CacheMount_UsesUVCache verifies that Python requirements.txt installation
// uses /root/.cache/uv (not /root/.cache/pip) for the BuildKit cache mount.
func TestUVInstall_CacheMount_UsesUVCache(t *testing.T) {
	appPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(appPath, "requirements.txt"), []byte("flask==3.0.0\n"), 0644); err != nil {
		t.Fatalf("failed to write requirements.txt: %v", err)
	}

	ws := &models.Workspace{ID: 1, Name: "test-ws", ImageName: "test:latest"}
	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: models.WorkspaceSpec{},
		Language:      "python",
		Version:       "3.11",
		AppPath:       appPath,
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if !strings.Contains(dockerfile, "target=/root/.cache/uv") {
		t.Errorf("Python Dockerfile should use /root/.cache/uv cache mount, got:\n%s", dockerfile)
	}
	if strings.Contains(dockerfile, "target=/root/.cache/pip") {
		t.Errorf("Python Dockerfile must NOT use legacy /root/.cache/pip cache mount\nDockerfile:\n%s", dockerfile)
	}
}

// TestUVInstall_ThreeTierFallback verifies the 3-tier install strategy:
// uv pip install → uv pip install (no proxy) → pip install (final fallback)
func TestUVInstall_ThreeTierFallback(t *testing.T) {
	appPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(appPath, "requirements.txt"), []byte("flask==3.0.0\n"), 0644); err != nil {
		t.Fatalf("failed to write requirements.txt: %v", err)
	}

	ws := &models.Workspace{ID: 1, Name: "test-ws", ImageName: "test:latest"}
	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: models.WorkspaceSpec{},
		Language:      "python",
		Version:       "3.11",
		AppPath:       appPath,
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Tier 1: uv pip install with cache mount
	if !strings.Contains(dockerfile, "uv pip install -r /tmp/requirements.txt") {
		t.Error("missing tier-1: uv pip install -r /tmp/requirements.txt")
	}
	// Tier 2: uv pip install after proxy unset
	if !strings.Contains(dockerfile, "unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy") {
		t.Error("missing proxy unset for tier-2 uv pip fallback")
	}
	// Tier 3: final pip fallback
	if !strings.Contains(dockerfile, "|| pip install -r /tmp/requirements.txt") {
		t.Error("missing tier-3 pip fallback: || pip install -r /tmp/requirements.txt")
	}
}

// TestUVInstall_DevTools_UsesUVWithPipFallback verifies that Python dev tools
// (e.g., ruff, mypy) installed in the dev stage use uv pip install with a pip fallback.
func TestUVInstall_DevTools_UsesUVWithPipFallback(t *testing.T) {
	ws := &models.Workspace{ID: 1, Name: "test-ws", ImageName: "test:latest"}
	wsYAML := models.WorkspaceSpec{
		Build: models.DevBuildConfig{
			DevStage: models.DevStageConfig{
				DevTools: []string{"ruff", "mypy"},
			},
		},
	}

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

	// Dev tools must use uv pip install (not bare pip)
	if !strings.Contains(dockerfile, "uv pip install ruff mypy") {
		t.Error("dev tools install should use: uv pip install ruff mypy")
	}
	// Must still have pip fallback for dev tools
	if !strings.Contains(dockerfile, "|| pip install ruff mypy") {
		t.Error("dev tools install missing pip fallback: || pip install ruff mypy")
	}
	// UV cache mount for dev tool install
	if !strings.Contains(dockerfile, "target=/root/.cache/uv") {
		t.Error("dev tools uv install missing /root/.cache/uv cache mount")
	}
}

// TestUVSetup_OrderBeforeRequirementsInstall verifies that the UV COPY and ENV
// lines appear in the Dockerfile before any requirements.txt installation.
func TestUVSetup_OrderBeforeRequirementsInstall(t *testing.T) {
	appPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(appPath, "requirements.txt"), []byte("flask==3.0.0\n"), 0644); err != nil {
		t.Fatalf("failed to write requirements.txt: %v", err)
	}

	ws := &models.Workspace{ID: 1, Name: "test-ws", ImageName: "test:latest"}
	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: models.WorkspaceSpec{},
		Language:      "python",
		Version:       "3.11",
		AppPath:       appPath,
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	uvCopyIdx := strings.Index(dockerfile, "COPY --from=ghcr.io/astral-sh/uv")
	uvInstallIdx := strings.Index(dockerfile, "uv pip install -r /tmp/requirements.txt")

	if uvCopyIdx == -1 {
		t.Fatal("missing UV COPY --from line")
	}
	if uvInstallIdx == -1 {
		t.Fatal("missing uv pip install line")
	}
	if uvCopyIdx >= uvInstallIdx {
		t.Errorf("UV COPY (idx %d) must appear before uv pip install (idx %d)", uvCopyIdx, uvInstallIdx)
	}
}

// TestDockerfileGenerator_PythonBaseImage_NoBookworm is a regression test for
// issue #324: Python base image must use "-slim" (not "-slim-bookworm") so that
// older patch versions like 3.9.9 resolve to a valid Docker Hub tag.
func TestDockerfileGenerator_PythonBaseImage_NoBookworm(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		// Exact reproduction from the bug report
		{"python 3.9.9 patch version", "3.9.9", "FROM python:3.9.9-slim"},
		{"python 3.10.4 patch version", "3.10.4", "FROM python:3.10.4-slim"},
		{"python 3.11.0 patch version", "3.11.0", "FROM python:3.11.0-slim"},
		{"python 3.12.1 patch version", "3.12.1", "FROM python:3.12.1-slim"},
		{"python 3.13.0 patch version", "3.13.0", "FROM python:3.13.0-slim"},
		// Minor-only versions
		{"python 3.9 minor", "3.9", "FROM python:3.9-slim"},
		{"python 3.13 minor", "3.13", "FROM python:3.13-slim"},
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

			if !strings.Contains(dockerfile, tt.want) {
				t.Errorf("Generate() expected base image %q not found in dockerfile", tt.want)
			}

			// Regression guard: ensure slim-bookworm is NEVER used for the Python base stage.
			// NOTE: we scan only the base stage FROM line, not the whole Dockerfile, because
			// the treesitter-builder parallel stage legitimately uses rust:1-slim-bookworm
			// (fix #334).  Checking the whole file would produce a false positive.
			baseFromPrefix := fmt.Sprintf("FROM python:%s", tt.version)
			for _, line := range strings.Split(dockerfile, "\n") {
				if strings.HasPrefix(line, baseFromPrefix) && strings.Contains(line, "slim-bookworm") {
					t.Errorf("Generate() python %s base FROM must not use 'slim-bookworm' tag (issue #324); got: %s", tt.version, line)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Issue #251 — nvim cache mount tests
// ---------------------------------------------------------------------------

// makeNvimStagingDir creates a staging dir with an init.lua so that
// generateNvimSection() is triggered. Returns the stagingDir path and a
// cleanup func.
func makeNvimStagingDir(t *testing.T, repoName string) (stagingDir string, cleanup func()) {
	t.Helper()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}
	stagingDir = filepath.Join(homeDir, ".devopsmaestro", "build-staging", repoName)
	nvimConfigPath := filepath.Join(stagingDir, ".config", "nvim")
	if err := os.MkdirAll(nvimConfigPath, 0755); err != nil {
		t.Fatalf("failed to create nvim config dir: %v", err)
	}
	initLua := filepath.Join(nvimConfigPath, "init.lua")
	if err := os.WriteFile(initLua, []byte("-- test"), 0644); err != nil {
		t.Fatalf("failed to create init.lua: %v", err)
	}
	return stagingDir, func() { os.RemoveAll(stagingDir) }
}

// TestNvimCacheMount_LazySync verifies the Lazy! sync RUN step carries the
// nvim cache mount directive (issue #251).
func TestNvimCacheMount_LazySync(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}
	_, cleanup := makeNvimStagingDir(t, "nvim-cache-lazy-test")
	defer cleanup()

	ws := &models.Workspace{ID: 1, Name: "myws", ImageName: "test:latest"}
	wsYAML := models.WorkspaceSpec{Nvim: models.NvimConfig{Structure: "custom"}}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       filepath.Join("/tmp", "dvm-clone", "nvim-cache-lazy-test"),
		PathConfig:    paths.New(homeDir),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Locate the Lazy! sync command
	lazyIdx := strings.Index(dockerfile, `"+Lazy! sync"`)
	if lazyIdx < 0 {
		t.Fatal("Generate() missing '+Lazy! sync' command — required for this test")
	}

	// Walk back up to 300 chars to find the enclosing RUN block
	start := lazyIdx - 300
	if start < 0 {
		start = 0
	}
	block := dockerfile[strings.LastIndex(dockerfile[start:lazyIdx], "RUN ")+start : lazyIdx+len(`"+Lazy! sync"`)]

	for _, want := range []string{
		"--mount=type=cache,target=/home/dev/.cache/nvim",
		"id=nvim-cache-myws",
		"uid=1000",
	} {
		if !strings.Contains(block, want) {
			t.Errorf("Lazy! sync RUN block missing nvim cache mount fragment %q\n"+
				"Expected the RUN --mount=type=cache,target=/home/<user>/.cache/nvim,id=nvim-cache-<ws>,uid=1000 directive.\n"+
				"RUN block:\n%s", want, block)
		}
	}
}

// TestNvimCacheMount_TreesitterInstall verifies the Treesitter install RUN step
// carries the nvim cache mount directive (issue #251).
func TestNvimCacheMount_TreesitterInstall(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}
	_, cleanup := makeNvimStagingDir(t, "nvim-cache-ts-test")
	defer cleanup()

	ws := &models.Workspace{ID: 1, Name: "myws", ImageName: "test:latest"}
	wsYAML := models.WorkspaceSpec{Nvim: models.NvimConfig{Structure: "custom"}}
	manifest := &plugin.PluginManifest{
		Features: plugin.PluginFeatures{HasTreesitter: true, HasMason: false},
	}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       filepath.Join("/tmp", "dvm-clone", "nvim-cache-ts-test"),
		PathConfig:    paths.New(homeDir),
	})
	gen.SetPluginManifest(manifest)

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Treesitter install uses "luafile /tmp/treesitter-install.lua"
	marker := "luafile /tmp/treesitter-install.lua"
	tsIdx := strings.Index(dockerfile, marker)
	if tsIdx < 0 {
		t.Fatal("Generate() missing 'luafile /tmp/treesitter-install.lua' — required for this test")
	}

	start := tsIdx - 300
	if start < 0 {
		start = 0
	}
	segment := dockerfile[start:tsIdx]
	lastRunOff := strings.LastIndex(segment, "RUN ")
	if lastRunOff < 0 {
		t.Fatalf("no RUN keyword found before treesitter lua marker; context:\n%s", segment)
	}
	block := dockerfile[start+lastRunOff : tsIdx+len(marker)]

	for _, want := range []string{
		"--mount=type=cache,target=/home/dev/.cache/nvim",
		"id=nvim-cache-myws",
		"uid=1000",
	} {
		if !strings.Contains(block, want) {
			t.Errorf("Treesitter install RUN block missing nvim cache mount fragment %q\n"+
				"RUN block:\n%s", want, block)
		}
	}
}

// TestNvimCacheMount_MasonInstall verifies the Mason install RUN step carries
// the nvim cache mount directive (issue #251).
func TestNvimCacheMount_MasonInstall(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}
	_, cleanup := makeNvimStagingDir(t, "nvim-cache-mason-test")
	defer cleanup()

	ws := &models.Workspace{ID: 1, Name: "myws", ImageName: "test:latest"}
	wsYAML := models.WorkspaceSpec{Nvim: models.NvimConfig{Structure: "custom"}}
	manifest := &plugin.PluginManifest{
		Features: plugin.PluginFeatures{HasMason: true, HasTreesitter: false},
	}

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       filepath.Join("/tmp", "dvm-clone", "nvim-cache-mason-test"),
		PathConfig:    paths.New(homeDir),
	})
	gen.SetPluginManifest(manifest)

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Mason install contains "luafile /tmp/mason-install.lua"
	masonIdx := strings.Index(dockerfile, "luafile /tmp/mason-install.lua")
	if masonIdx < 0 {
		t.Fatal("Generate() missing 'luafile /tmp/mason-install.lua' — required for this test")
	}

	start := masonIdx - 300
	if start < 0 {
		start = 0
	}
	segment := dockerfile[start:masonIdx]
	lastRunOff := strings.LastIndex(segment, "RUN ")
	if lastRunOff < 0 {
		t.Fatalf("no RUN keyword found before 'luafile /tmp/mason-install.lua'; context:\n%s", segment)
	}
	block := dockerfile[start+lastRunOff : masonIdx+len("luafile /tmp/mason-install.lua")]

	for _, want := range []string{
		"--mount=type=cache,target=/home/dev/.cache/nvim",
		"id=nvim-cache-myws",
		"uid=1000",
	} {
		if !strings.Contains(block, want) {
			t.Errorf("Mason install RUN block missing nvim cache mount fragment %q\n"+
				"RUN block:\n%s", want, block)
		}
	}
}

// TestNvimCacheMount_NotPresent_NoNvimConfig verifies that when no nvim config
// directory exists the nvim cache mount does not appear in the Dockerfile.
func TestNvimCacheMount_NotPresent_NoNvimConfig(t *testing.T) {
	ws := &models.Workspace{ID: 1, Name: "myws", ImageName: "test:latest"}
	wsYAML := models.WorkspaceSpec{} // no nvim config

	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       "/tmp/no-nvim-config-path",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if strings.Contains(dockerfile, "nvim-cache-") {
		t.Errorf("Dockerfile must not contain 'nvim-cache-' when nvim is not configured;\n"+
			"found it in output (means nvimCacheMount() was emitted unexpectedly).\n"+
			"Relevant fragment:\n%s",
			extractFragment(dockerfile, "nvim-cache-", 200))
	}
}

// TestNvimCacheMount_WorkspaceScoped verifies that different workspace names
// produce different cache mount IDs (e.g., nvim-cache-dev vs nvim-cache-staging).
func TestNvimCacheMount_WorkspaceScoped(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		wsName      string
		wantCacheID string
	}{
		{"dev", "nvim-cache-dev"},
		{"staging", "nvim-cache-staging"},
	}

	for _, tt := range tests {
		t.Run(tt.wsName, func(t *testing.T) {
			repoName := "nvim-cache-scope-" + tt.wsName
			_, cleanup := makeNvimStagingDir(t, repoName)
			defer cleanup()

			ws := &models.Workspace{ID: 1, Name: tt.wsName, ImageName: "test:latest"}
			wsYAML := models.WorkspaceSpec{Nvim: models.NvimConfig{Structure: "custom"}}

			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
				Workspace:     ws,
				WorkspaceSpec: wsYAML,
				Language:      "python",
				Version:       "3.11",
				AppPath:       filepath.Join("/tmp", "dvm-clone", repoName),
				PathConfig:    paths.New(homeDir),
			})

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			if !strings.Contains(dockerfile, tt.wantCacheID) {
				t.Errorf("expected cache mount ID %q not found in Dockerfile for workspace %q",
					tt.wantCacheID, tt.wsName)
			}

			// Ensure the other workspace's ID is NOT present
			for _, other := range tests {
				if other.wsName == tt.wsName {
					continue
				}
				if strings.Contains(dockerfile, other.wantCacheID) {
					t.Errorf("unexpected cache mount ID %q found in Dockerfile for workspace %q — IDs must be workspace-scoped",
						other.wantCacheID, tt.wsName)
				}
			}
		})
	}
}

// extractFragment returns up to `chars` characters surrounding the first
// occurrence of `needle` in `s`, useful for error messages.
func extractFragment(s, needle string, chars int) string {
	idx := strings.Index(s, needle)
	if idx < 0 {
		return "(not found)"
	}
	start := idx - chars/2
	if start < 0 {
		start = 0
	}
	end := idx + chars/2
	if end > len(s) {
		end = len(s)
	}
	return s[start:end]
}

// ---------------------------------------------------------------------------
// Issue #251 — APT lists path consistency (aptCacheMountsLocked)
// ---------------------------------------------------------------------------

// TestAptCacheMountsLocked_UsesListsPath verifies aptCacheMountsLocked() targets
// /var/lib/apt/lists (not the shorter /var/lib/apt). The missing /lists suffix
// caused the apt lock to target the wrong directory, breaking parallel builds.
func TestAptCacheMountsLocked_UsesListsPath(t *testing.T) {
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

	// Must use /var/lib/apt/lists (with /lists)
	if !strings.Contains(dockerfile, "target=/var/lib/apt/lists") {
		t.Errorf("aptCacheMountsLocked() must use 'target=/var/lib/apt/lists' but it was not found in Dockerfile")
	}

	// Regression guard: must NOT have /var/lib/apt without /lists as a cache target
	// Check that no mount targets exactly /var/lib/apt (i.e., not followed by /lists)
	if strings.Contains(dockerfile, "target=/var/lib/apt,") || strings.Contains(dockerfile, "target=/var/lib/apt ") {
		t.Errorf("aptCacheMountsLocked() must NOT target '/var/lib/apt' (missing '/lists') — regression detected")
	}
}

// ---------------------------------------------------------------------------
// Issue #321 — Scala signed-by GPG key pattern
// ---------------------------------------------------------------------------

// buildScalaDockerfile is a helper that generates a Dockerfile for a Scala
// workspace and returns its content.
func buildScalaDockerfile(t *testing.T) string {
	t.Helper()
	ws := &models.Workspace{ID: 1, Name: "scala-ws", ImageName: "test:latest"}
	gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: models.WorkspaceSpec{},
		Language:      "scala",
		Version:       "21",
		AppPath:       "/tmp/scala-test",
		PathConfig:    paths.New(t.TempDir()),
	})
	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	return dockerfile
}

// TestScalaBaseStage_UsesSignedBy verifies that the Scala base stage uses the
// modern signed-by pattern for the sbt APT repository (issue #321).
func TestScalaBaseStage_UsesSignedBy(t *testing.T) {
	dockerfile := buildScalaDockerfile(t)

	wants := []string{
		"signed-by=/usr/share/keyrings/sbt-archive-keyring.gpg",
		"gpg --dearmor -o /usr/share/keyrings/sbt-archive-keyring.gpg",
		"repo.scala-sbt.org/scalasbt/debian",
	}
	for _, want := range wants {
		if !strings.Contains(dockerfile, want) {
			t.Errorf("Scala Dockerfile missing expected signed-by pattern fragment: %q\n"+
				"The sbt APT source must use the modern 'signed-by' approach (issue #321).", want)
		}
	}
}

// TestScalaBaseStage_NoAptKeyAdd is a regression guard ensuring the deprecated
// 'apt-key add' command never appears in a Scala Dockerfile (issue #321).
func TestScalaBaseStage_NoAptKeyAdd(t *testing.T) {
	dockerfile := buildScalaDockerfile(t)

	if strings.Contains(dockerfile, "apt-key add") {
		t.Errorf("Scala Dockerfile must NOT use 'apt-key add' (deprecated, see issue #321).\n"+
			"Use 'gpg --dearmor' with 'signed-by=' in the sources.list entry instead.\n"+
			"Fragment with 'apt-key add':\n%s",
			extractFragment(dockerfile, "apt-key add", 300))
	}
}

// ---------------------------------------------------------------------------
// End Issue #251 / #321 tests
// ---------------------------------------------------------------------------

// TestUVSetup_PythonVersions verifies UV is injected across all supported Python versions.
func TestUVSetup_PythonVersions(t *testing.T) {
	versions := []string{"3.9", "3.10", "3.11", "3.12", "3.13"}

	for _, version := range versions {
		t.Run("python-"+version, func(t *testing.T) {
			ws := &models.Workspace{ID: 1, Name: "test-ws", ImageName: "test:latest"}
			gen := NewDockerfileGenerator(DockerfileGeneratorOptions{
				Workspace:     ws,
				WorkspaceSpec: models.WorkspaceSpec{},
				Language:      "python",
				Version:       version,
				AppPath:       "/tmp/test",
				PathConfig:    paths.New(t.TempDir()),
			})

			dockerfile, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			if !strings.Contains(dockerfile, "COPY --from=ghcr.io/astral-sh/uv:0.7.2 /uv /uvx /bin/") {
				t.Errorf("python %s missing UV COPY line", version)
			}
			if !strings.Contains(dockerfile, "ENV UV_SYSTEM_PYTHON=1") {
				t.Errorf("python %s missing ENV UV_SYSTEM_PYTHON=1", version)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Issue #332 — cacheID() uniqueness: workspace Slug eliminates collisions
// ---------------------------------------------------------------------------

// TestCacheID_UsesSlug_WhenAvailable verifies that when a Workspace has a non-empty
// Slug, the generated Dockerfile uses the slug (not just the workspace name) as the
// cache mount ID. Two workspaces with the same name but different apps must produce
// different cache IDs.
func TestCacheID_UsesSlug_WhenAvailable(t *testing.T) {
	// Simulate two workspaces both named "dev" but belonging to different apps.
	// The Slug is globally unique (ecosystem-domain-system-app-ws format).
	tests := []struct {
		name      string
		wsName    string
		slug      string
		wantID    string
		notWantID string
	}{
		{
			name:      "app-a dev workspace uses slug",
			wsName:    "dev",
			slug:      "myeco-core-mysys-app-a-dev",
			wantID:    "myeco-core-mysys-app-a-dev",
			notWantID: "myeco-core-mysys-app-b-dev",
		},
		{
			name:      "app-b dev workspace uses slug",
			wsName:    "dev",
			slug:      "myeco-core-mysys-app-b-dev",
			wantID:    "myeco-core-mysys-app-b-dev",
			notWantID: "myeco-core-mysys-app-a-dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := &models.Workspace{ID: 1, Name: tt.wsName, Slug: tt.slug, ImageName: "test:latest"}
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

			if !strings.Contains(dockerfile, tt.wantID) {
				t.Errorf("cache mount must use slug %q as ID, but it was not found in Dockerfile", tt.wantID)
			}
			if strings.Contains(dockerfile, tt.notWantID) {
				t.Errorf("cache mount must NOT contain %q (cross-app ID leak, issue #332)", tt.notWantID)
			}
		})
	}
}

// TestCacheID_FallsBackToName_WhenSlugEmpty verifies that when Slug is empty the
// cache mount ID degrades gracefully to the workspace Name (legacy behaviour).
func TestCacheID_FallsBackToName_WhenSlugEmpty(t *testing.T) {
	ws := &models.Workspace{ID: 1, Name: "dev", Slug: "", ImageName: "test:latest"}
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

	if !strings.Contains(dockerfile, "apt-cache-dev") {
		t.Errorf("without Slug, cache mount should use workspace name 'dev' (id=apt-cache-dev), not found in Dockerfile")
	}
}

// TestCacheID_TwoApps_SameWorkspaceName_DifferentIDs is a table-driven regression
// guard for issue #332: two workspaces that share the same name ("dev") but belong
// to different apps must generate *different* apt-cache mount IDs.
func TestCacheID_TwoApps_SameWorkspaceName_DifferentIDs(t *testing.T) {
	makeGen := func(slug string) string {
		ws := &models.Workspace{ID: 1, Name: "dev", Slug: slug, ImageName: "test:latest"}
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
		return dockerfile
	}

	slugA := "eco-domain-sys-app-a-dev"
	slugB := "eco-domain-sys-app-b-dev"

	dfA := makeGen(slugA)
	dfB := makeGen(slugB)

	// Each Dockerfile must contain its own slug-based cache ID
	if !strings.Contains(dfA, fmt.Sprintf("apt-cache-%s", slugA)) {
		t.Errorf("app-a Dockerfile should contain apt-cache-%s (issue #332)", slugA)
	}
	if !strings.Contains(dfB, fmt.Sprintf("apt-cache-%s", slugB)) {
		t.Errorf("app-b Dockerfile should contain apt-cache-%s (issue #332)", slugB)
	}

	// Neither Dockerfile should contain the other app's cache ID
	if strings.Contains(dfA, fmt.Sprintf("apt-cache-%s", slugB)) {
		t.Errorf("app-a Dockerfile must NOT contain app-b cache ID apt-cache-%s (cross-app pollution)", slugB)
	}
	if strings.Contains(dfB, fmt.Sprintf("apt-cache-%s", slugA)) {
		t.Errorf("app-b Dockerfile must NOT contain app-a cache ID apt-cache-%s (cross-app pollution)", slugA)
	}
}

// ---------------------------------------------------------------------------
// Issue #333 — APT lists cleanup: rm -rf before apt-get update
// ---------------------------------------------------------------------------

// TestAptGetUpdate_AlwaysPrecededByCleanup verifies that every apt-get update
// call in generated Dockerfiles is immediately preceded by
// "rm -rf /var/lib/apt/lists/*" on the same logical command line.
// This prevents stale/corrupted partial InRelease files from causing GPG
// signature failures on subsequent builds (issue #333).
func TestAptGetUpdate_AlwaysPrecededByCleanup(t *testing.T) {
	tests := []struct {
		name     string
		language string
		version  string
	}{
		{"python debian", "python", "3.11"},
		{"golang alpine", "golang", "1.22"},
		{"nodejs", "node", "20"},
		{"scala", "scala", "21"},
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

			// Find every occurrence of "apt-get update" and assert it's preceded
			// by the cleanup command on the same line (or combined with &&).
			remaining := dockerfile
			for {
				idx := strings.Index(remaining, "apt-get update")
				if idx < 0 {
					break
				}
				// Look at up to 60 chars before "apt-get update" for the cleanup
				start := idx - 60
				if start < 0 {
					start = 0
				}
				context := remaining[start:idx]
				if !strings.Contains(context, "rm -rf /var/lib/apt/lists/*") {
					// Show the surrounding line for debugging
					lineStart := strings.LastIndex(remaining[:idx], "\n")
					if lineStart < 0 {
						lineStart = 0
					}
					lineEnd := strings.Index(remaining[idx:], "\n")
					if lineEnd < 0 {
						lineEnd = len(remaining) - idx
					}
					line := remaining[lineStart : idx+lineEnd]
					t.Errorf("%s: found 'apt-get update' not preceded by 'rm -rf /var/lib/apt/lists/*' (issue #333).\nLine: %s", tt.name, line)
				}
				remaining = remaining[idx+len("apt-get update"):]
			}
		})
	}
}

// TestAptGetUpdate_NoOrphanUpdate is a strict regression guard: any apt-get update
// that is NOT preceded by the cache-clearing rm -rf is a latent corruption risk.
// This test generates a Python Dockerfile and asserts the pattern holds for every
// line containing apt-get update.
func TestAptGetUpdate_NoOrphanUpdate(t *testing.T) {
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

	for i, line := range strings.Split(dockerfile, "\n") {
		if strings.Contains(line, "apt-get update") {
			if !strings.Contains(line, "rm -rf /var/lib/apt/lists/*") {
				t.Errorf("line %d: apt-get update without preceding 'rm -rf /var/lib/apt/lists/*' (issue #333):\n  %s", i+1, line)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Issue #334 — tree-sitter built from source (cargo install, not binary download)
// ---------------------------------------------------------------------------

// TestTreeSitterBuilder_UsesCargoInstall verifies that the treesitter-builder
// stage uses "cargo install" instead of downloading a pre-built binary from
// GitHub releases. Pre-built binaries require GLIBC 2.39 which is not available
// on Debian Bookworm (GLIBC 2.36) — see issue #334.
func TestTreeSitterBuilder_UsesCargoInstall(t *testing.T) {
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

	tsStart := strings.Index(dockerfile, "# --- Parallel builder: tree-sitter CLI ---")
	if tsStart < 0 {
		t.Fatal("Generate() missing tree-sitter builder stage")
	}
	// Extract through the next blank-line-separated paragraph
	tsSection := dockerfile[tsStart:]
	if end := strings.Index(tsSection[50:], "\n\n"); end >= 0 {
		tsSection = tsSection[:end+50]
	}

	if !strings.Contains(tsSection, "cargo install tree-sitter-cli@") {
		t.Errorf("tree-sitter builder must use 'cargo install tree-sitter-cli@<version>' (issue #334).\nSection:\n%s", tsSection)
	}
}

// TestTreeSitterBuilder_NoPrebuiltBinaryURL asserts that no GitHub releases
// download URL for tree-sitter appears anywhere in the Dockerfile. The old
// approach downloaded a pre-built binary that required GLIBC 2.39 (issue #334).
func TestTreeSitterBuilder_NoPrebuiltBinaryURL(t *testing.T) {
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

	// The old pattern was: https://github.com/tree-sitter/tree-sitter/releases/download/...
	if strings.Contains(dockerfile, "tree-sitter/releases/download") {
		t.Errorf("Dockerfile must NOT download pre-built tree-sitter binary (GLIBC mismatch, issue #334).\n"+
			"Found 'tree-sitter/releases/download' in Dockerfile — use 'cargo install' instead.\n%s",
			extractFragment(dockerfile, "tree-sitter/releases/download", 300))
	}
}

// TestTreeSitterBuilder_AlpineAndDebian_BothUseCargo verifies that both the
// Alpine variant (golang) and the Debian variant (python) build tree-sitter
// from source via cargo, not from pre-built binaries (issue #334).
func TestTreeSitterBuilder_AlpineAndDebian_BothUseCargo(t *testing.T) {
	tests := []struct {
		name      string
		language  string
		version   string
		wantImage string
		wantCargo string
	}{
		{
			name:      "debian variant (python)",
			language:  "python",
			version:   "3.11",
			wantImage: "rust:1-slim-bookworm",
			wantCargo: "cargo install tree-sitter-cli@",
		},
		{
			name:      "alpine variant (golang)",
			language:  "golang",
			version:   "1.22",
			wantImage: "rust:1-alpine3.20",
			wantCargo: "cargo install tree-sitter-cli@",
		},
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

			tsStart := strings.Index(dockerfile, "# --- Parallel builder: tree-sitter CLI ---")
			if tsStart < 0 {
				t.Fatal("Generate() missing tree-sitter builder stage")
			}
			tsSection := dockerfile[tsStart:]
			if end := strings.Index(tsSection[50:], "\n\n"); end >= 0 {
				tsSection = tsSection[:end+50]
			}

			if !strings.Contains(tsSection, tt.wantImage) {
				t.Errorf("%s: tree-sitter builder should use Rust base image %q.\nSection:\n%s", tt.name, tt.wantImage, tsSection)
			}
			if !strings.Contains(tsSection, tt.wantCargo) {
				t.Errorf("%s: tree-sitter builder should use %q.\nSection:\n%s", tt.name, tt.wantCargo, tsSection)
			}
			if strings.Contains(tsSection, "tree-sitter/releases/download") {
				t.Errorf("%s: tree-sitter builder must NOT download pre-built binary (GLIBC mismatch, issue #334).\nSection:\n%s", tt.name, tsSection)
			}
		})
	}
}

// TestTreeSitterBuilder_BinaryVerified asserts that the builder stage includes a
// verification step (test -x) to confirm the installed binary is executable.
func TestTreeSitterBuilder_BinaryVerified(t *testing.T) {
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

	tsStart := strings.Index(dockerfile, "# --- Parallel builder: tree-sitter CLI ---")
	if tsStart < 0 {
		t.Fatal("Generate() missing tree-sitter builder stage")
	}
	tsSection := dockerfile[tsStart:]
	if end := strings.Index(tsSection[50:], "\n\n"); end >= 0 {
		tsSection = tsSection[:end+50]
	}

	if !strings.Contains(tsSection, "test -x /usr/local/bin/tree-sitter") {
		t.Errorf("tree-sitter builder stage should verify binary is executable with 'test -x /usr/local/bin/tree-sitter'.\nSection:\n%s", tsSection)
	}
}

// ---------------------------------------------------------------------------
// End Issue #332 / #333 / #334 tests
// ---------------------------------------------------------------------------

// =============================================================================
// Issue #220: go-tools-builder stage must use a sufficiently recent Go version
// =============================================================================

// TestCompareGoVersions verifies the helper that compares Go version strings.
func TestCompareGoVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.21", "1.25", -1},
		{"1.25", "1.21", 1},
		{"1.25", "1.25", 0},
		{"1.22", "1.25", -1},
		{"1.30", "1.25", 1},
		{"2.0", "1.25", 1},
		{"1.25", "2.0", -1},
		{"1.9", "1.10", -1},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			got := compareGoVersions(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("compareGoVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestGoToolsBuilderVersion_MinimumFloor verifies that the go-tools-builder
// stage uses at least Go 1.25, even when the project targets an older version.
// This prevents build failures from gopls@latest requiring Go >= 1.25.
func TestGoToolsBuilderVersion_MinimumFloor(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		wantToolsGo string
	}{
		{"old version 1.20 gets floor", "1.20", goToolsMinGoVersion},
		{"old version 1.21 gets floor", "1.21", goToolsMinGoVersion},
		{"old version 1.22 gets floor", "1.22", goToolsMinGoVersion},
		{"old version 1.24 gets floor", "1.24", goToolsMinGoVersion},
		{"exact minimum stays", "1.25", "1.25"},
		{"newer version kept", "1.26", "1.26"},
		{"much newer version kept", "1.30", "1.30"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := &DefaultDockerfileGenerator{
				language: "golang",
				version:  tt.version,
			}
			got := gen.goToolsBuilderVersion()
			if got != tt.wantToolsGo {
				t.Errorf("goToolsBuilderVersion() = %q, want %q", got, tt.wantToolsGo)
			}
		})
	}
}

// TestGoToolsBuilderStage_UsesMinGoVersion verifies the generated Dockerfile
// contains golang:<min>-alpine in the go-tools-builder FROM line when the
// project uses an older Go version, while the base stage still uses the
// project's Go version. End-to-end test for issue #220.
func TestGoToolsBuilderStage_UsesMinGoVersion(t *testing.T) {
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
		Version:       "1.21",
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Base stage should use the project's Go version (1.21)
	if !strings.Contains(dockerfile, "FROM golang:1.21-alpine") {
		t.Error("base stage should use golang:1.21-alpine for the project")
	}

	// go-tools-builder stage should use the minimum floor version
	wantToolsFrom := fmt.Sprintf("golang:%s-alpine", goToolsMinGoVersion)
	if !strings.Contains(dockerfile, wantToolsFrom+" AS go-tools-builder") &&
		!strings.Contains(dockerfile, wantToolsFrom+"@sha256:") {
		t.Errorf("go-tools-builder should use %s, got:\n%s",
			wantToolsFrom, extractGoToolsBuilderLine(dockerfile))
	}
}

// extractGoToolsBuilderLine returns the FROM line for go-tools-builder.
func extractGoToolsBuilderLine(dockerfile string) string {
	for _, line := range strings.Split(dockerfile, "\n") {
		if strings.Contains(line, "AS go-tools-builder") {
			return line
		}
	}
	return "<go-tools-builder FROM line not found>"
}

// =============================================================================
// Issue #338 Regression Test: tree-sitter builder uses $CARGO_HOME (not /root/.cargo)
// =============================================================================

// TestGenerateTreeSitterBuilder_UsesCargoHomeEnvVar verifies that the tree-sitter builder
// stage copies the compiled binary using $CARGO_HOME/bin/tree-sitter rather than the
// hardcoded path /root/.cargo/bin/tree-sitter (which breaks non-root Cargo installs).
// Regression test for the fix in generateTreeSitterBuilder() — Issue #338.
func TestGenerateTreeSitterBuilder_UsesCargoHomeEnvVar(t *testing.T) {
	tests := []struct {
		name     string
		language string
		version  string
		variant  string // "alpine" or "debian"
	}{
		{
			name:     "alpine variant (golang workspace)",
			language: "golang",
			version:  "1.22",
			variant:  "alpine",
		},
		{
			name:     "debian variant (python workspace)",
			language: "python",
			version:  "3.11",
			variant:  "debian",
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

			// Extract the tree-sitter builder section
			tsStart := strings.Index(dockerfile, "# --- Parallel builder: tree-sitter CLI ---")
			if tsStart < 0 {
				t.Fatalf("Generate() missing tree-sitter builder stage")
			}
			devStart := strings.Index(dockerfile, "FROM base AS dev")
			var tsSection string
			if devStart > tsStart {
				tsSection = dockerfile[tsStart:devStart]
			} else {
				tsSection = dockerfile[tsStart:]
			}

			// MUST use $CARGO_HOME env var (respects non-root and custom Cargo installs)
			if !strings.Contains(tsSection, "$CARGO_HOME/bin/tree-sitter") {
				t.Errorf("[%s] tree-sitter builder should use '$CARGO_HOME/bin/tree-sitter'.\n"+
					"Regression guard for Issue #338: $CARGO_HOME must be used instead of hardcoded path.\n"+
					"tree-sitter section:\n%s", tt.variant, tsSection)
			}

			// MUST NOT use the old hardcoded root-only path
			if strings.Contains(tsSection, "/root/.cargo/bin/tree-sitter") {
				t.Errorf("[%s] tree-sitter builder must NOT hardcode '/root/.cargo/bin/tree-sitter'.\n"+
					"Regression guard for Issue #338: hardcoded path breaks non-root Cargo installs.\n"+
					"Use '$CARGO_HOME/bin/tree-sitter' instead.\n"+
					"tree-sitter section:\n%s", tt.variant, tsSection)
			}
		})
	}
}
