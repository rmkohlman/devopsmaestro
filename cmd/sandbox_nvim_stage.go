package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	nvimconfig "github.com/rmkohlman/MaestroNvim/nvimops/config"
	"github.com/rmkohlman/MaestroNvim/nvimops/library"
	"github.com/rmkohlman/MaestroNvim/nvimops/plugin"
	theme "github.com/rmkohlman/MaestroTheme"
	themelib "github.com/rmkohlman/MaestroTheme/library"
)

// sandboxDefaultTheme is the theme applied to sandbox workspaces by default.
//
// We pick `catppuccin-mocha` because:
//   - it ships in MaestroTheme/library/themes/ (no resolution needed),
//   - it's a popular, dark, well-supported colorscheme that works on the
//     terminal-color-256 baseline that all sandbox base images provide,
//   - it's the most commonly-used theme across the DevOpsMaestro ecosystem.
//
// Sandboxes are intentionally hierarchy-less (no workspace -> app -> domain
// -> ecosystem cascade), so we substitute a sensible constant default here
// rather than calling resolveWorkspaceTheme() which requires a *models.Workspace.
const sandboxDefaultTheme = "catppuccin-mocha"

// sandboxCorePluginNames is the set of "core" plugins installed by default in a
// sandbox. This mirrors the final fallback list inside generateNvimConfig() in
// build_nvim.go (see `corePluginNames` there) — we keep them in sync intentionally
// so the sandbox experience matches the regular workspace experience when no
// hierarchy override is configured.
var sandboxCorePluginNames = []string{
	"treesitter",
	"telescope",
	"which-key",
	"lspconfig",
	"nvim-cmp",
	"gitsigns",
}

// generateSandboxNvimStaging populates `stagingDir` with a complete nvim config
// tree (including theme files) suitable for COPY into a sandbox image. The
// resulting layout matches what GenerateSandboxDockerfileWithNvim expects:
//
//	{stagingDir}/.config/nvim/init.lua
//	{stagingDir}/.config/nvim/lua/<ns>/...
//	{stagingDir}/.config/nvim/lua/theme/palette.lua
//	{stagingDir}/.config/nvim/lua/theme/init.lua
//	{stagingDir}/.config/nvim/lua/<ns>/plugins/colorscheme.lua
//
// This function is deliberately DB-free: it pulls plugins from the embedded
// MaestroNvim library and themes from the embedded MaestroTheme library, so it
// works for sandboxes which have no database record and no app/workspace context.
func generateSandboxNvimStaging(stagingDir string) error {
	nvimConfigPath := filepath.Join(stagingDir, ".config", "nvim")
	if err := os.MkdirAll(nvimConfigPath, 0755); err != nil {
		return fmt.Errorf("create nvim config dir: %w", err)
	}

	// Load the embedded plugin library — same source the regular workspace path
	// uses as its final fallback when no hierarchy package is resolved.
	pluginLibrary, err := library.NewLibrary()
	if err != nil {
		return fmt.Errorf("init embedded plugin library: %w", err)
	}

	plugins, err := collectSandboxCorePlugins(pluginLibrary)
	if err != nil {
		return err
	}

	// Use the default core config (sensible Vim options, leader=space, etc).
	cfg := nvimconfig.DefaultCoreConfig()

	gen := nvimconfig.NewGenerator()
	if err := gen.WriteToDirectory(cfg, plugins, nvimConfigPath); err != nil {
		return fmt.Errorf("write nvim config: %w", err)
	}

	// Apply the default theme. Failure here is non-fatal — the sandbox is still
	// usable without the colorscheme files; nvim will just fall back to default.
	if err := writeSandboxTheme(nvimConfigPath, cfg.Namespace); err != nil {
		// Wrap and return so the caller can decide. Currently sandbox_create.go
		// surfaces this as a warning rather than aborting the build.
		return fmt.Errorf("apply default theme %q: %w", sandboxDefaultTheme, err)
	}

	return nil
}

// collectSandboxCorePlugins resolves the sandbox core plugin list from the
// embedded library. Plugins not found in the library are silently skipped —
// the embedded list should always satisfy sandboxCorePluginNames, but we
// degrade gracefully if a future library refactor renames one.
func collectSandboxCorePlugins(lib *library.Library) ([]*plugin.Plugin, error) {
	var plugins []*plugin.Plugin
	for _, name := range sandboxCorePluginNames {
		if p, ok := lib.Get(name); ok {
			plugins = append(plugins, p)
		}
	}
	if len(plugins) == 0 {
		return nil, fmt.Errorf("no core plugins resolved from embedded library (looked for %v)", sandboxCorePluginNames)
	}
	return plugins, nil
}

// writeSandboxTheme generates the default theme files into the staged nvim
// config. The layout matches what generateNvimConfig() in build_nvim.go writes
// for the regular workspace path so any future plugin code that reads
// `require("theme.palette")` works identically inside a sandbox.
func writeSandboxTheme(nvimConfigPath, namespace string) error {
	if namespace == "" {
		namespace = "workspace"
	}

	t, err := themelib.Get(sandboxDefaultTheme)
	if err != nil {
		return fmt.Errorf("load embedded theme: %w", err)
	}

	generated, err := theme.NewGenerator().Generate(t)
	if err != nil {
		return fmt.Errorf("generate theme lua: %w", err)
	}

	files := map[string]string{
		filepath.Join(nvimConfigPath, "lua", "theme", "palette.lua"):                 generated.PaletteLua,
		filepath.Join(nvimConfigPath, "lua", "theme", "init.lua"):                    generated.InitLua,
		filepath.Join(nvimConfigPath, "lua", namespace, "plugins", "colorscheme.lua"): generated.PluginLua,
	}
	if generated.ColorschemeLua != "" {
		files[filepath.Join(nvimConfigPath, "lua", "theme", "colorscheme.lua")] = generated.ColorschemeLua
	}

	for path, content := range files {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}

	return nil
}
