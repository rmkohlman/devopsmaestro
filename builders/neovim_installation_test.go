package builders

import (
	"devopsmaestro/models"
	"strings"
	"testing"
)

func TestNeovimInstallation_PythonSlim(t *testing.T) {
	// Test that Neovim is installed from GitHub releases instead of apt package
	ws := &models.Workspace{
		Name:      "test-python-neovim",
		ImageName: "python:3.11-slim",
	}

	wsYAML := models.WorkspaceSpec{
		Container: models.ContainerConfig{
			WorkingDir: "/workspace",
			UID:        1000,
			GID:        1000,
			User:       "dev",
		},
		Build: models.DevBuildConfig{
			DevStage: models.DevStageConfig{},
		},
		Nvim: models.NvimConfig{
			Structure: "custom", // Enable nvim
		},
		Shell: models.ShellConfig{
			Theme: "starship",
		},
	}

	generator := NewDockerfileGenerator(ws, wsYAML, "python", "3.11", "/tmp/test", "")

	dockerfile, err := generator.Generate()
	if err != nil {
		t.Fatalf("Error generating Dockerfile: %v", err)
	}

	t.Run("has_neovim_github_installation", func(t *testing.T) {
		if !strings.Contains(dockerfile, "Install Neovim from GitHub releases") {
			t.Error("Dockerfile should contain GitHub Neovim installation section")
		}
	})

	t.Run("has_architecture_detection", func(t *testing.T) {
		if !strings.Contains(dockerfile, "dpkg --print-architecture") {
			t.Error("Dockerfile should contain dpkg architecture detection for Debian/Ubuntu")
		}
	})

	t.Run("has_correct_download_urls", func(t *testing.T) {
		if !strings.Contains(dockerfile, "nvim-linux-x86_64") || !strings.Contains(dockerfile, "nvim-linux-arm64") {
			t.Error("Dockerfile should contain architecture-specific Neovim download URLs")
		}
	})

	t.Run("has_correct_symlink", func(t *testing.T) {
		if !strings.Contains(dockerfile, "/usr/local/bin/nvim") {
			t.Error("Dockerfile should create symlink to /usr/local/bin/nvim")
		}
	})

	t.Run("neovim_not_in_apt_packages", func(t *testing.T) {
		// This is the key fix - neovim should NOT be installed via apt-get
		lines := strings.Split(dockerfile, "\n")
		for _, line := range lines {
			if strings.Contains(line, "apt-get install") && strings.Contains(line, "neovim") {
				t.Errorf("Found neovim in apt-get install line (this will fail on slim images): %s", strings.TrimSpace(line))
			}
		}
	})
}

func TestNeovimInstallation_GolangAlpine(t *testing.T) {
	// Test that Neovim installation works with Alpine images
	ws := &models.Workspace{
		Name:      "test-golang-neovim",
		ImageName: "golang:1.22-alpine",
	}

	wsYAML := models.WorkspaceSpec{
		Container: models.ContainerConfig{
			WorkingDir: "/workspace",
			UID:        1000,
			GID:        1000,
			User:       "dev",
		},
		Build: models.DevBuildConfig{
			DevStage: models.DevStageConfig{},
		},
		Nvim: models.NvimConfig{
			Structure: "custom",
		},
		Shell: models.ShellConfig{
			Theme: "starship",
		},
	}

	generator := NewDockerfileGenerator(ws, wsYAML, "golang", "1.22", "/tmp/test", "")

	dockerfile, err := generator.Generate()
	if err != nil {
		t.Fatalf("Error generating Dockerfile: %v", err)
	}

	t.Run("has_alpine_architecture_detection", func(t *testing.T) {
		if !strings.Contains(dockerfile, "uname -m") {
			t.Error("Dockerfile should use uname -m for Alpine architecture detection")
		}
	})

	t.Run("has_alpine_neovim_dependencies", func(t *testing.T) {
		if !strings.Contains(dockerfile, "apk add --no-cache") || !strings.Contains(dockerfile, "build-base") {
			t.Error("Dockerfile should install Alpine-specific Neovim dependencies")
		}
	})

	t.Run("neovim_not_in_apk_packages", func(t *testing.T) {
		// Similar check for Alpine - neovim should NOT be installed via apk
		lines := strings.Split(dockerfile, "\n")
		for _, line := range lines {
			if strings.Contains(line, "apk add") && strings.Contains(line, "neovim") {
				t.Errorf("Found neovim in apk add line (should use GitHub releases): %s", strings.TrimSpace(line))
			}
		}
	})
}
