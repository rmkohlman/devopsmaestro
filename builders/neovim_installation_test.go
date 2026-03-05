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
		// Neovim is now downloaded in a parallel builder stage
		if !strings.Contains(dockerfile, "Parallel builder: Neovim") {
			t.Error("Dockerfile should contain parallel Neovim builder stage")
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
		// With BuildKit cache mounts, apk uses --mount=type=cache instead of --no-cache.
		// build-base is still required for Mason/Treesitter compilation.
		if !strings.Contains(dockerfile, "apk add") || !strings.Contains(dockerfile, "build-base") {
			t.Error("Dockerfile should install Alpine-specific Neovim dependencies")
		}
	})

	t.Run("neovim_in_merged_apk_packages", func(t *testing.T) {
		// Alpine gets neovim via apk (GitHub releases are glibc-linked, won't work on musl).
		// The merged package install includes neovim and neovim-doc.
		if !strings.Contains(dockerfile, "neovim") {
			t.Error("Dockerfile should install neovim via apk for Alpine (musl can't use GitHub releases)")
		}
	})
}

func TestLazygitInstallation_PythonSlim(t *testing.T) {
	// Test that lazygit is installed from GitHub releases
	ws := &models.Workspace{
		Name:      "test-python-lazygit",
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
			Structure: "custom",
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

	t.Run("has_lazygit_github_installation", func(t *testing.T) {
		// Lazygit is now downloaded in a parallel builder stage
		if !strings.Contains(dockerfile, "Parallel builder: lazygit") {
			t.Error("Dockerfile should contain parallel lazygit builder stage")
		}
	})

	t.Run("has_lazygit_architecture_detection", func(t *testing.T) {
		if !strings.Contains(dockerfile, "LG_ARCH") {
			t.Error("Dockerfile should contain LG_ARCH variable for lazygit")
		}
	})

	t.Run("downloads_from_github", func(t *testing.T) {
		if !strings.Contains(dockerfile, "jesseduffield/lazygit") {
			t.Error("Dockerfile should download lazygit from GitHub")
		}
	})

	t.Run("installs_to_usr_local_bin", func(t *testing.T) {
		if !strings.Contains(dockerfile, "install lazygit /usr/local/bin") {
			t.Error("Dockerfile should install lazygit to /usr/local/bin")
		}
	})
}

func TestLazygitInstallation_GolangAlpine(t *testing.T) {
	// Test that lazygit is installed on Alpine (Go images)
	ws := &models.Workspace{
		Name:      "test-golang-lazygit",
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

	t.Run("has_lazygit_installation", func(t *testing.T) {
		// Lazygit is now downloaded in a parallel builder stage
		if !strings.Contains(dockerfile, "Parallel builder: lazygit") {
			t.Error("Dockerfile should contain parallel lazygit builder stage")
		}
	})

	t.Run("has_alpine_architecture_detection_for_lazygit", func(t *testing.T) {
		// Alpine uses uname -m, and we should see the LG_ARCH variable set
		if !strings.Contains(dockerfile, "LG_ARCH") {
			t.Error("Dockerfile should contain LG_ARCH variable for lazygit")
		}
	})
}
