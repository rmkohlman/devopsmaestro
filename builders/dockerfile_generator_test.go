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

	if gen.workspace != ws {
		t.Error("workspace not set")
	}
	if gen.language != "python" {
		t.Errorf("language = %q, want %q", gen.language, "python")
	}
	if gen.version != "3.11" {
		t.Errorf("version = %q, want %q", gen.version, "3.11")
	}
	if gen.appPath != "/app/path" {
		t.Errorf("appPath = %q, want %q", gen.appPath, "/app/path")
	}
	if gen.baseDockerfile != "/app/path/Dockerfile" {
		t.Errorf("baseDockerfile = %q, want %q", gen.baseDockerfile, "/app/path/Dockerfile")
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
