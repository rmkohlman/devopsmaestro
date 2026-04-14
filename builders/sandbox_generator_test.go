package builders

import (
	"devopsmaestro/models"
	"strings"
	"testing"
)

func TestGenerateSandboxDockerfile_Python(t *testing.T) {
	preset, ok := models.GetPreset("python")
	if !ok {
		t.Fatal("python preset not found")
	}

	df := GenerateSandboxDockerfile(preset, "3.12", "")

	assertContains(t, df, "FROM python:3.12-slim")
	assertContains(t, df, "git curl ca-certificates build-essential")
	assertContains(t, df, "useradd -m -u 1000 -g dev -s /bin/bash dev")
	assertContains(t, df, "WORKDIR /sandbox")
	assertContains(t, df, "USER dev")
	assertContains(t, df, `CMD ["/bin/bash"]`)
	assertNotContains(t, df, "COPY")
}

func TestGenerateSandboxDockerfile_Golang(t *testing.T) {
	preset, ok := models.GetPreset("golang")
	if !ok {
		t.Fatal("golang preset not found")
	}

	df := GenerateSandboxDockerfile(preset, "1.24", "")

	assertContains(t, df, "FROM golang:1.24-bookworm")
	assertContains(t, df, "WORKDIR /sandbox")
	assertContains(t, df, "USER dev")
}

func TestGenerateSandboxDockerfile_Rust(t *testing.T) {
	preset, ok := models.GetPreset("rust")
	if !ok {
		t.Fatal("rust preset not found")
	}

	df := GenerateSandboxDockerfile(preset, "1.86", "")

	assertContains(t, df, "FROM rust:1.86-slim-bookworm")
}

func TestGenerateSandboxDockerfile_Node(t *testing.T) {
	preset, ok := models.GetPreset("node")
	if !ok {
		t.Fatal("node preset not found")
	}

	df := GenerateSandboxDockerfile(preset, "22", "")

	assertContains(t, df, "FROM node:22-slim")
}

func TestGenerateSandboxDockerfile_Cpp(t *testing.T) {
	preset, ok := models.GetPreset("cpp")
	if !ok {
		t.Fatal("cpp preset not found")
	}

	df := GenerateSandboxDockerfile(preset, "14", "")

	assertContains(t, df, "FROM gcc:14-bookworm")
}

func TestGenerateSandboxDockerfile_Dotnet(t *testing.T) {
	preset, ok := models.GetPreset("dotnet")
	if !ok {
		t.Fatal("dotnet preset not found")
	}

	df := GenerateSandboxDockerfile(preset, "9.0", "")

	assertContains(t, df, "FROM mcr.microsoft.com/dotnet/sdk:9.0")
	assertContains(t, df, "WORKDIR /sandbox")
	assertContains(t, df, "USER dev")
}

func TestGenerateSandboxDockerfile_DotnetAlias(t *testing.T) {
	preset, ok := models.GetPreset("csharp")
	if !ok {
		t.Fatal("csharp alias should resolve to dotnet preset")
	}

	if preset.Language != "dotnet" {
		t.Errorf("csharp alias resolved to %q, want %q", preset.Language, "dotnet")
	}
}

func TestGenerateSandboxDockerfile_WithDepsFile(t *testing.T) {
	preset, ok := models.GetPreset("python")
	if !ok {
		t.Fatal("python preset not found")
	}

	df := GenerateSandboxDockerfile(preset, "3.12", "/home/user/project/requirements.txt")

	assertContains(t, df, "COPY requirements.txt /sandbox/requirements.txt")
	assertContains(t, df, "RUN pip install -r requirements.txt")
}

func TestGenerateSandboxDockerfile_WithGoDeps(t *testing.T) {
	preset, ok := models.GetPreset("golang")
	if !ok {
		t.Fatal("golang preset not found")
	}

	df := GenerateSandboxDockerfile(preset, "1.24", "/project/go.mod")

	assertContains(t, df, "COPY go.mod /sandbox/go.mod")
	// golang's DepsInstallCmd has no %s — it uses "go mod download" directly
	assertContains(t, df, "RUN go mod download")
}

func TestGenerateSandboxDockerfile_WithNodeDeps(t *testing.T) {
	preset, ok := models.GetPreset("node")
	if !ok {
		t.Fatal("node preset not found")
	}

	df := GenerateSandboxDockerfile(preset, "20", "/project/package.json")

	assertContains(t, df, "COPY package.json /sandbox/package.json")
	assertContains(t, df, "RUN npm install")
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected output to contain %q, got:\n%s", substr, s)
	}
}

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Errorf("expected output to NOT contain %q, got:\n%s", substr, s)
	}
}
