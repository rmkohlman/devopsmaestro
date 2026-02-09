package builders

import (
	"strings"
	"testing"

	"devopsmaestro/models"
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
				"FROM python:3.11-slim AS base",
				"apt-get update",
				"gcc",
				"python3-dev",
			},
		},
		{
			name:    "python specific version",
			version: "3.10",
			wantContain: []string{
				"FROM python:3.10-slim AS base",
			},
		},
		{
			name:    "python 3.12",
			version: "3.12",
			wantContain: []string{
				"FROM python:3.12-slim AS base",
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
				"apk add --no-cache git",
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
				"useradd -m -u 1000 -g 1000 -s /bin/zsh dev",
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
				"useradd -m -u 501 -g 501 -s /bin/zsh dev",
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
				"useradd -m -u 1000 -g 1000 -s /bin/zsh myuser",
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
		{"python", "", "python:3.11-slim"},
		{"python", "3.9", "python:3.9-slim"},
		{"python", "3.10", "python:3.10-slim"},
		{"python", "3.11", "python:3.11-slim"},
		{"python", "3.12", "python:3.12-slim"},

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
