package cmd

import (
	"context"
	"devopsmaestro/db"
	"devopsmaestro/models"
	colorresolver "devopsmaestro/pkg/colors/resolver"
	"devopsmaestro/pkg/nvimops/theme"
	"devopsmaestro/pkg/paths"
	"devopsmaestro/pkg/resource/handlers"
	terminalpkg "devopsmaestro/pkg/terminalops/package"
	terminalpkglib "devopsmaestro/pkg/terminalops/package/library"
	"devopsmaestro/pkg/terminalops/prompt"
	"devopsmaestro/pkg/terminalops/prompt/composer"
	promptextension "devopsmaestro/pkg/terminalops/prompt/extension"
	extlib "devopsmaestro/pkg/terminalops/prompt/extension/library"
	promptstyle "devopsmaestro/pkg/terminalops/prompt/style"
	stylelib "devopsmaestro/pkg/terminalops/prompt/style/library"
	"fmt"
	"github.com/rmkohlman/MaestroPalette"
	"log/slog"
	"os"
	"path/filepath"
)

// TerminalPackageStore provides access to terminal packages.
type TerminalPackageStore interface {
	Get(name string) (*terminalpkg.Package, bool)
}

// PromptStyleStore provides access to prompt styles.
type PromptStyleStore interface {
	Get(name string) (*promptstyle.PromptStyle, error)
}

// PromptExtensionStore provides access to prompt extensions.
type PromptExtensionStore interface {
	Get(name string) (*promptextension.PromptExtension, error)
}

// getPromptFromPackageOrDefault returns a PromptYAML either from the configured terminal package
// or falls back to createDefaultTerminalPrompt().
//
// Logic:
// 1. Check ds.GetDefault("terminal-package")
// 2. If not set -> fall back to createDefaultTerminalPrompt()
// 3. If set -> load package from pkgStore
// 4. Check UsesModularPrompt() - if false, fall back
// 5. Load style from styleStore
// 6. Load extensions from extStore
// 7. Compose using composer.NewStarshipComposer().Compose()
// 8. Convert ComposedPrompt to PromptYAML
// 9. Set prompt name to dvm-pkg-{packageName}-{appName}-{workspaceName}
// 10. Set Spec.Palette to "theme"
func getPromptFromPackageOrDefault(
	ctx context.Context,
	ds db.DataStore,
	pkgStore TerminalPackageStore,
	styleStore PromptStyleStore,
	extStore PromptExtensionStore,
	appName, workspaceName string,
) (*prompt.PromptYAML, error) {
	// Step 1: Check for terminal-package default
	packageName, err := ds.GetDefault("terminal-package")
	if err != nil || packageName == "" {
		// No terminal package set, use default
		slog.Debug("no terminal-package default set, using hardcoded default")
		return createDefaultTerminalPrompt(appName, workspaceName), nil
	}

	slog.Debug("found terminal-package default", "package", packageName)

	// Step 2: Load package from store
	pkg, found := pkgStore.Get(packageName)
	if !found {
		slog.Warn("terminal package not found, falling back to default", "package", packageName)
		return createDefaultTerminalPrompt(appName, workspaceName), nil
	}

	// Step 3: Check if package uses modular prompt system
	if !pkg.UsesModularPrompt() {
		slog.Debug("terminal package does not use modular prompt system, falling back to default", "package", packageName)
		return createDefaultTerminalPrompt(appName, workspaceName), nil
	}

	slog.Debug("composing prompt from terminal package",
		"package", packageName,
		"style", pkg.PromptStyle,
		"extensions", pkg.PromptExtensions)

	// Step 4: Load style
	style, err := styleStore.Get(pkg.PromptStyle)
	if err != nil {
		slog.Warn("failed to load prompt style, falling back to default",
			"package", packageName,
			"style", pkg.PromptStyle,
			"error", err)
		return createDefaultTerminalPrompt(appName, workspaceName), nil
	}

	// Step 5: Load extensions (skip missing ones, continue with what we have)
	var extensions []*promptextension.PromptExtension
	for _, extName := range pkg.PromptExtensions {
		ext, err := extStore.Get(extName)
		if err != nil {
			slog.Warn("failed to load prompt extension, skipping",
				"package", packageName,
				"extension", extName,
				"error", err)
			continue
		}
		extensions = append(extensions, ext)
	}

	// If no extensions loaded successfully, fall back
	if len(extensions) == 0 {
		slog.Warn("no prompt extensions loaded successfully, falling back to default", "package", packageName)
		return createDefaultTerminalPrompt(appName, workspaceName), nil
	}

	// Step 6: Compose prompt
	starshipComposer := composer.NewStarshipComposer()
	composedPrompt, err := starshipComposer.Compose(style, extensions)
	if err != nil {
		slog.Warn("failed to compose prompt, falling back to default",
			"package", packageName,
			"error", err)
		return createDefaultTerminalPrompt(appName, workspaceName), nil
	}

	// Step 7: Convert ComposedPrompt to PromptYAML
	promptName := fmt.Sprintf("dvm-pkg-%s-%s-%s", packageName, appName, workspaceName)
	promptYAML := prompt.NewTerminalPrompt(promptName)
	promptYAML.Metadata.Description = fmt.Sprintf("Composed from terminal package '%s'", packageName)
	promptYAML.Spec.Format = composedPrompt.Format
	promptYAML.Spec.Palette = "theme" // Use theme colors

	// Convert composer.ModuleConfig to prompt.ModuleConfig
	promptYAML.Spec.Modules = make(map[string]prompt.ModuleConfig)
	for moduleName, moduleConfig := range composedPrompt.Modules {
		promptYAML.Spec.Modules[moduleName] = prompt.ModuleConfig{
			Disabled: moduleConfig.Disabled,
			Symbol:   moduleConfig.Symbol,
			Format:   moduleConfig.Format,
			Style:    moduleConfig.Style,
			Options:  moduleConfig.Options,
		}
	}

	slog.Debug("successfully composed prompt from terminal package",
		"package", packageName,
		"prompt", promptName,
		"modules", len(promptYAML.Spec.Modules))

	return promptYAML, nil
}

// generateShellConfig creates .zshrc and starship.toml files in staging directory
func generateShellConfig(stagingDir, appName, workspaceName string, ds db.DataStore, workspace *models.Workspace) error {
	// Create .config directory for starship.toml
	configDir := filepath.Join(stagingDir, ".config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Generate .zshrc
	zshrc := `# DevOpsMaestro Container Shell
export TERM=xterm-256color
export EDITOR=nvim
export DVM_APP=` + appName + `

# Starship prompt
eval "$(starship init zsh)"

# Aliases
alias vim=nvim
alias ll='ls -la'
alias la='ls -la'
alias l='ls -l'

# Set up completion system
autoload -U compinit
compinit
`

	zshrcPath := filepath.Join(stagingDir, ".zshrc")
	if err := os.WriteFile(zshrcPath, []byte(zshrc), 0644); err != nil {
		return fmt.Errorf("failed to write .zshrc: %w", err)
	}

	// Append plugin loading (non-fatal if it fails)
	if err := appendPluginLoading(zshrcPath, ds); err != nil {
		slog.Warn("failed to append plugin loading to zshrc", "error", err)
		// Continue - this is non-fatal
	}

	// Ensure handlers are registered (idempotent)
	handlers.RegisterAll()

	// Create library stores for terminal packages, styles, and extensions
	pkgLib, err := terminalpkglib.NewLibrary()
	if err != nil {
		slog.Warn("failed to load terminal package library", "error", err)
		pkgLib = nil
	}

	styleLibrary, err := stylelib.NewStyleLibrary()
	if err != nil {
		slog.Warn("failed to load prompt style library", "error", err)
		styleLibrary = nil
	}

	extLibrary, err := extlib.NewExtensionLibrary()
	if err != nil {
		slog.Warn("failed to load prompt extension library", "error", err)
		extLibrary = nil
	}

	// Get prompt from package or default
	ctx := context.Background()
	promptYAML, err := getPromptFromPackageOrDefault(ctx, ds, pkgLib, styleLibrary, extLibrary, appName, workspaceName)
	if err != nil {
		slog.Warn("failed to get prompt from package, using hardcoded default", "error", err)
		promptYAML = createDefaultTerminalPrompt(appName, workspaceName)
	}

	// Resolve theme from hierarchy (workspace -> app -> domain -> ecosystem -> global)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	nvpDir := paths.New(homeDir).NVPRoot()
	themeStore := theme.NewFileStore(nvpDir)

	themeCtx := context.Background()
	resolvedTheme, themeErr := resolveWorkspaceTheme(themeCtx, ds, themeStore, workspace)
	if themeErr != nil {
		slog.Warn("failed to resolve theme from hierarchy, using default palette", "error", themeErr)
		// Fall back to default palette if resolution fails
		resolvedTheme = nil
	}

	// Convert theme to palette, or use default if resolution failed
	var themePalette *palette.Palette
	if resolvedTheme != nil {
		themePalette = resolvedTheme.ToPalette()
		slog.Debug("using hierarchy-resolved theme for shell config", "theme", resolvedTheme.Name)
	} else {
		themePalette = createDefaultPalette()
		slog.Debug("using default palette for shell config")
	}

	// Use the new renderer to generate starship.toml
	renderer := prompt.NewRenderer()
	starshipPath := filepath.Join(configDir, "starship.toml")
	if err := renderer.RenderToFile(promptYAML, themePalette, starshipPath); err != nil {
		return fmt.Errorf("failed to render starship.toml: %w", err)
	}

	// Generate WezTerm config if terminal emulator exists in database
	if err := generateWezTermConfig(stagingDir, appName, workspaceName, ds); err != nil {
		slog.Warn("failed to generate wezterm config", "error", err)
		// Non-fatal - continue with build
	}

	return nil
}

// createDefaultTerminalPrompt creates a default TerminalPrompt configuration
// that matches the previous hardcoded behavior.
func createDefaultTerminalPrompt(appName, workspaceName string) *prompt.PromptYAML {
	defaultPromptName := fmt.Sprintf("dvm-default-%s-%s", appName, workspaceName)
	py := prompt.NewTerminalPrompt(defaultPromptName)
	py.Metadata.Description = fmt.Sprintf("Default DevOpsMaestro prompt for %s/%s", appName, workspaceName)

	// Set format matching the original hardcoded config
	py.Spec.Format = `$custom\
$directory\
$git_branch\
$git_status\
$character`

	// Configure custom module for app name
	py.Spec.Modules = map[string]prompt.ModuleConfig{
		"custom.dvm": {
			Format: "[$output](bold ${theme.cyan}) ",
			Options: map[string]any{
				"command": fmt.Sprintf(`echo '[%s]'`, appName),
				"when":    `test -n "$DVM_APP"`,
				"shell":   []string{"bash", "--noprofile", "--norc"},
			},
		},
		"directory": {
			Options: map[string]any{
				"truncation_length": 3,
			},
		},
		"character": {
			Options: map[string]any{
				"success_symbol": "[➜](bold ${theme.green})",
				"error_symbol":   "[✗](bold ${theme.red})",
			},
		},
	}

	return py
}

// createDefaultPalette creates a default palette for starship prompt rendering.
// This provides basic colors that work well in most terminal environments.
func createDefaultPalette() *palette.Palette {
	return &palette.Palette{
		Name:        "default",
		Description: "Default DevOpsMaestro colors",
		Category:    palette.CategoryDark,
		Colors: map[string]string{
			// Basic background/foreground
			palette.ColorBg: "#1a1b26",
			palette.ColorFg: "#c0caf5",

			// Standard terminal colors
			palette.TermRed:     "#f7768e",
			palette.TermGreen:   "#9ece6a",
			palette.TermYellow:  "#e0af68",
			palette.TermBlue:    "#7aa2f7",
			palette.TermMagenta: "#bb9af7",
			palette.TermCyan:    "#7dcfff",
			palette.TermWhite:   "#c0caf5",
			palette.TermBlack:   "#15161e",

			// Bright variants
			palette.TermBrightRed:     "#f7768e",
			palette.TermBrightGreen:   "#9ece6a",
			palette.TermBrightYellow:  "#e0af68",
			palette.TermBrightBlue:    "#7aa2f7",
			palette.TermBrightMagenta: "#bb9af7",
			palette.TermBrightCyan:    "#7dcfff",
			palette.TermBrightWhite:   "#c0caf5",
			palette.TermBrightBlack:   "#414868",

			// Standard theme color names (needed for ToTerminalColors mapping)
			"red":     "#f7768e",
			"green":   "#9ece6a",
			"yellow":  "#e0af68",
			"blue":    "#7aa2f7",
			"magenta": "#bb9af7",
			"cyan":    "#7dcfff",
			"white":   "#c0caf5",
			"black":   "#15161e",

			// Semantic colors
			palette.ColorError:   "#f7768e",
			palette.ColorWarning: "#e0af68",
			palette.ColorInfo:    "#7aa2f7",
			palette.ColorHint:    "#1abc9c",
			palette.ColorSuccess: "#9ece6a",
			palette.ColorComment: "#565f89",
			palette.ColorBorder:  "#27a1b9",

			// Accent colors
			palette.ColorPrimary:   "#7aa2f7",
			palette.ColorSecondary: "#bb9af7",
			palette.ColorAccent:    "#7dcfff",
		},
	}
}

// resolveWorkspaceTheme resolves the theme for a workspace using hierarchy resolution.
// It walks the hierarchy: workspace -> app -> domain -> ecosystem -> global default.
// If workspace is nil, it returns the global default theme from the database.
// On any error, it falls back to the database default.
func resolveWorkspaceTheme(ctx context.Context, ds db.DataStore, themeStore theme.Store, workspace *models.Workspace) (*theme.Theme, error) {
	// Create resolver
	resolverInst := colorresolver.NewHierarchyThemeResolver(ds, themeStore)

	// Handle nil workspace - return global default from database
	if workspace == nil {
		resolution, err := resolverInst.ResolveDefault()
		if err != nil {
			return nil, err
		}
		return resolution.Theme, nil
	}

	// Resolve theme starting from workspace level
	resolution, err := resolverInst.Resolve(ctx, colorresolver.LevelWorkspace, workspace.ID)
	if err != nil {
		// Fall back to default on error
		defaultResolution, defaultErr := resolverInst.ResolveDefault()
		if defaultErr != nil {
			return nil, defaultErr
		}
		return defaultResolution.Theme, nil
	}

	return resolution.Theme, nil
}
