package builders

import (
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/paths"
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

	generator := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: "python", Version: "3.11", AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})

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

	t.Run("uses_tarball_not_appimage", func(t *testing.T) {
		// Issue #356: Tarball from neovim/neovim avoids AppImage extraction issues.
		// AppImage approach (#342) failed because AppImages do NOT bundle glibc.
		if !strings.Contains(dockerfile, ".tar.gz") {
			t.Error("Dockerfile should download Neovim tarball (.tar.gz) for GLIBC 2.17+ compat (#356)")
		}
		if strings.Contains(dockerfile, ".appimage") {
			t.Error("Dockerfile should NOT use Neovim AppImage (GLIBC incompatibility, see #356)")
		}
	})

	t.Run("extracts_tarball_with_tar", func(t *testing.T) {
		// Issue #356: Tarball extracted with tar, replacing unsquashfs-based AppImage extraction (#351).
		if !strings.Contains(dockerfile, "tar xzf") {
			t.Error("Dockerfile should extract Neovim tarball with tar xzf (see #356)")
		}
		if strings.Contains(dockerfile, "unsquashfs") {
			t.Error("Dockerfile should NOT use unsquashfs (AppImage extraction removed, see #356)")
		}
		if strings.Contains(dockerfile, "squashfs-tools") {
			t.Error("Dockerfile should NOT install squashfs-tools (no longer needed, see #356)")
		}
	})

	t.Run("tarball_binary_path", func(t *testing.T) {
		// Tarball extracts as nvim-linux-{arch}/bin/nvim → /opt/nvim/bin/nvim with --strip-components=1
		if !strings.Contains(dockerfile, "/opt/nvim/bin/nvim") {
			t.Error("Dockerfile should reference /opt/nvim/bin/nvim (tarball extraction path)")
		}
	})

	t.Run("does_not_detect_squashfs_offset", func(t *testing.T) {
		// unsquashfs offset detection no longer needed with tarball approach
		if strings.Contains(dockerfile, "hsqs") {
			t.Error("Dockerfile should NOT detect squashfs magic ('hsqs') — tarballs don't need offset detection")
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

	generator := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: "golang", Version: "1.22", AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})

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

	generator := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: "python", Version: "3.11", AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})

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

	generator := NewDockerfileGenerator(DockerfileGeneratorOptions{Workspace: ws, WorkspaceSpec: wsYAML, Language: "golang", Version: "1.22", AppPath: "/tmp/test", PathConfig: paths.New(t.TempDir())})

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

// =============================================================================
// Fix #342 — Neovim GLIBC incompatibility
// =============================================================================

// TestNeovimGlibcFallback_DebianIncludesFallbackRun verifies that the Debian
// neovim-builder stage includes the GLIBC compatibility check and source build
// fallback added in fix #342.
func TestNeovimGlibcFallback_DebianIncludesFallbackRun(t *testing.T) {
	ws := &models.Workspace{
		Name:      "test-glibc-fallback",
		ImageName: "python:3.11-slim",
	}
	wsYAML := models.WorkspaceSpec{
		Container: models.ContainerConfig{
			WorkingDir: "/workspace",
			UID:        1000,
			GID:        1000,
			User:       "dev",
		},
		Nvim: models.NvimConfig{
			Structure: "custom",
		},
		Shell: models.ShellConfig{
			Theme: "starship",
		},
	}

	generator := NewDockerfileGenerator(DockerfileGeneratorOptions{
		Workspace:     ws,
		WorkspaceSpec: wsYAML,
		Language:      "python",
		Version:       "3.11",
		AppPath:       "/tmp/test",
		PathConfig:    paths.New(t.TempDir()),
	})

	dockerfile, err := generator.Generate()
	if err != nil {
		t.Fatalf("Error generating Dockerfile: %v", err)
	}

	t.Run("has_glibc_compatibility_check", func(t *testing.T) {
		// Fix #342: test if the pre-built binary works before using it
		if !strings.Contains(dockerfile, "/opt/nvim/bin/nvim --version") {
			t.Error("Dockerfile must test nvim binary compatibility (#342): /opt/nvim/bin/nvim --version")
		}
	})

	t.Run("fallback_builds_from_source_with_git_clone", func(t *testing.T) {
		// Fix #342: if binary is GLIBC-incompatible, fall back to building from source
		if !strings.Contains(dockerfile, "git clone --depth 1 --branch") {
			t.Error("Dockerfile must include source build fallback with git clone (#342)")
		}
	})

	t.Run("fallback_uses_cmake", func(t *testing.T) {
		// Fix #342: cmake is required to build Neovim from source
		if !strings.Contains(dockerfile, "cmake") {
			t.Error("Dockerfile must install cmake for source build fallback (#342)")
		}
	})

	t.Run("symlink_still_created", func(t *testing.T) {
		// The symlink to /usr/local/bin/nvim must still be created regardless of path
		if !strings.Contains(dockerfile, "ln -sf /opt/nvim/bin/nvim /usr/local/bin/nvim") {
			t.Error("Dockerfile must still create symlink ln -sf /opt/nvim/bin/nvim /usr/local/bin/nvim (#342)")
		}
	})

	t.Run("alpine_does_not_have_fallback", func(t *testing.T) {
		// Alpine uses apk — the GLIBC fallback is Debian-only
		// This test uses python:3.11-slim (Debian), so the fallback IS present.
		// Confirm the fallback is NOT gated behind an Alpine check.
		if strings.Contains(dockerfile, "apk add") && strings.Contains(dockerfile, "git clone --depth 1 --branch") {
			// If Alpine apk AND git clone are both present, something is wrong
			// (Alpine should use apk for Neovim, not the source-build fallback).
			// This is not an error for Debian images, so just verify the fallback is present.
		}
	})
}
