package builders

import (
	"devopsmaestro/models"
	"strings"
	"testing"
)

// TestGenerateSandboxDockerfileWithNvim_NoNvim verifies that opting out of nvim
// produces output identical to the legacy minimal generator. This protects the
// --no-nvim flag from accidental drift.
func TestGenerateSandboxDockerfileWithNvim_NoNvim(t *testing.T) {
	preset, ok := models.GetPreset("python")
	if !ok {
		t.Fatal("python preset not found")
	}

	withFlag := GenerateSandboxDockerfileWithNvim(preset, "3.12", "",
		SandboxNvimOptions{IncludeNvim: false})
	legacy := GenerateSandboxDockerfile(preset, "3.12", "")

	if withFlag != legacy {
		t.Errorf("--no-nvim path must match legacy GenerateSandboxDockerfile output exactly\n"+
			"--- legacy ---\n%s\n--- with flag ---\n%s", legacy, withFlag)
	}
}

// TestGenerateSandboxDockerfileWithNvim_IncludesNeovimBuilder verifies that the
// default (with-nvim) sandbox emits the parallel neovim-builder stage with the
// pinned tarball + checksum verification, matching the regular workspace path.
func TestGenerateSandboxDockerfileWithNvim_IncludesNeovimBuilder(t *testing.T) {
	preset, ok := models.GetPreset("python")
	if !ok {
		t.Fatal("python preset not found")
	}

	df := GenerateSandboxDockerfileWithNvim(preset, "3.12", "",
		SandboxNvimOptions{IncludeNvim: true})

	// BuildKit syntax directive — required for cache mounts and parallel stages.
	assertContains(t, df, "# syntax=docker/dockerfile:1")

	// Parallel Neovim builder stage with checksum-verified tarball.
	assertContains(t, df, "FROM debian:bookworm-slim")
	assertContains(t, df, "AS neovim-builder")
	assertContains(t, df, "neovim/neovim/releases/download/v")
	assertContains(t, df, "sha256sum -c -")
	assertContains(t, df, "test -x /opt/nvim/bin/nvim")

	// Final image still uses the language base image.
	assertContains(t, df, "FROM python:3.12-slim")

	// Nvim copy + GLIBC fallback + symlink onto PATH.
	assertContains(t, df, "COPY --from=neovim-builder /opt/nvim/ /opt/nvim/")
	assertContains(t, df, "Neovim pre-built binary OK")
	assertContains(t, df, "ln -sf /opt/nvim/bin/nvim /usr/local/bin/nvim")

	// Runtime dependencies that nvim plugins (telescope, treesitter) need.
	assertContains(t, df, "ripgrep")
	assertContains(t, df, "fd-find")
	assertContains(t, df, "unzip")

	// Pre-staged config copied from build context.
	assertContains(t, df, "COPY --chown=dev:dev .config/nvim /home/dev/.config/nvim")

	// Lazy.nvim plugin sync as the dev user (must come after USER dev).
	assertContains(t, df, "USER dev")
	assertContains(t, df, `nvim --headless "+Lazy! sync" +qa`)
}

// TestGenerateSandboxDockerfileWithNvim_PreservesPresetExtras checks that
// preset-specific extras (UV for Python, deps install) still emit when nvim
// is enabled — i.e. the nvim layer is additive, not a replacement.
func TestGenerateSandboxDockerfileWithNvim_PreservesPresetExtras(t *testing.T) {
	preset, ok := models.GetPreset("python")
	if !ok {
		t.Fatal("python preset not found")
	}

	df := GenerateSandboxDockerfileWithNvim(preset, "3.12", "",
		SandboxNvimOptions{IncludeNvim: true})

	// UV is still copied in for Python presets.
	assertContains(t, df, "COPY --from=ghcr.io/astral-sh/uv:0.7.2 /uv /uvx /bin/")
	assertContains(t, df, "ENV UV_LINK_MODE=copy")

	// Non-root dev user is still created (and chowns nvim config correctly).
	assertContains(t, df, "useradd -m -u 1000 -g dev -s /bin/bash dev")
	assertContains(t, df, "WORKDIR /sandbox")
}

// TestGenerateSandboxDockerfileWithNvim_OrderingDevUserBeforeNvimCopy is a
// regression guard: the COPY --chown=dev:dev step must come AFTER the dev user
// is created, otherwise the chown silently fails and lazy.nvim sync hits
// permission errors at runtime.
func TestGenerateSandboxDockerfileWithNvim_OrderingDevUserBeforeNvimCopy(t *testing.T) {
	preset, ok := models.GetPreset("golang")
	if !ok {
		t.Fatal("golang preset not found")
	}

	df := GenerateSandboxDockerfileWithNvim(preset, "1.24", "",
		SandboxNvimOptions{IncludeNvim: true})

	userIdx := strings.Index(df, "useradd -m -u 1000 -g dev")
	copyIdx := strings.Index(df, "COPY --chown=dev:dev .config/nvim")
	syncIdx := strings.Index(df, `nvim --headless "+Lazy! sync"`)

	if userIdx < 0 || copyIdx < 0 || syncIdx < 0 {
		t.Fatalf("expected useradd/COPY/Lazy sync all present (user=%d copy=%d sync=%d)",
			userIdx, copyIdx, syncIdx)
	}
	if !(userIdx < copyIdx && copyIdx < syncIdx) {
		t.Errorf("ordering violated: useradd@%d must precede COPY@%d which must precede Lazy sync@%d",
			userIdx, copyIdx, syncIdx)
	}
}

// TestGenerateSandboxDockerfileWithNvim_DepsFileStillWorks verifies that the
// optional deps install (e.g. requirements.txt) coexists with the nvim layer.
func TestGenerateSandboxDockerfileWithNvim_DepsFileStillWorks(t *testing.T) {
	preset, ok := models.GetPreset("python")
	if !ok {
		t.Fatal("python preset not found")
	}

	df := GenerateSandboxDockerfileWithNvim(preset, "3.12",
		"/host/path/to/requirements.txt",
		SandboxNvimOptions{IncludeNvim: true})

	assertContains(t, df, "COPY requirements.txt /sandbox/requirements.txt")
	// Python deps should still get the UV-then-pip fallback wrap.
	assertContains(t, df, "uv pip install")
}
