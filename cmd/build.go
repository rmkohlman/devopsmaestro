package cmd

import (
	"context"
	"devopsmaestro/builders"
	"devopsmaestro/config"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/operators"
	nvimconfig "devopsmaestro/pkg/nvimops/config"
	"devopsmaestro/pkg/nvimops/library"
	nvimpackage "devopsmaestro/pkg/nvimops/package"
	packagelibrary "devopsmaestro/pkg/nvimops/package/library"
	"devopsmaestro/pkg/nvimops/plugin"
	"devopsmaestro/pkg/nvimops/store"
	"devopsmaestro/pkg/nvimops/theme"
	"devopsmaestro/pkg/palette"
	"devopsmaestro/pkg/resolver"
	"devopsmaestro/pkg/resource"
	"devopsmaestro/pkg/resource/handlers"
	terminalplugin "devopsmaestro/pkg/terminalops/plugin"
	"devopsmaestro/pkg/terminalops/prompt"
	"devopsmaestro/pkg/terminalops/wezterm"
	"devopsmaestro/render"
	"devopsmaestro/utils"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	buildForce   bool
	buildNocache bool
	buildTarget  string
	buildFlags   HierarchyFlags
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build workspace container image",
	Long: `Build a development container image for the active workspace.

This command:
- Detects the app language
- Generates or extends Dockerfile with dev tools
- Builds the image using the detected container platform
- Tags as dvm-<workspace>-<app>:latest

Supports multiple platforms:
- OrbStack (uses Docker API)
- Docker Desktop (uses Docker API)
- Podman (uses Docker API)
- Colima with containerd (uses BuildKit API)

Use DVM_PLATFORM environment variable to select a specific platform.

Flags:
  -e, --ecosystem   Filter by ecosystem name
  -d, --domain      Filter by domain name  
  -a, --app         Filter by app name
  -w, --workspace   Filter by workspace name

Examples:
  dvm build
  dvm build --force
  dvm build --no-cache
  dvm build -a portal                 # Build workspace in 'portal' app
  dvm build -e healthcare -a portal   # Specify ecosystem and app
  DVM_PLATFORM=colima dvm build
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return buildWorkspace(cmd)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().BoolVar(&buildForce, "force", false, "Force rebuild even if image exists")
	buildCmd.Flags().BoolVar(&buildNocache, "no-cache", false, "Build without using cache")
	buildCmd.Flags().StringVar(&buildTarget, "target", "dev", "Build target stage (default: dev)")
	AddHierarchyFlags(buildCmd, &buildFlags)
}

func buildWorkspace(cmd *cobra.Command) error {
	slog.Info("starting build")

	// Get datastore
	sqlDS, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	var app *models.App
	var workspace *models.Workspace
	var appName, workspaceName string

	// Check if hierarchy flags were provided
	if buildFlags.HasAnyFlag() {
		// Use resolver to find workspace
		slog.Debug("using hierarchy flags", "ecosystem", buildFlags.Ecosystem,
			"domain", buildFlags.Domain, "app", buildFlags.App, "workspace", buildFlags.Workspace)

		wsResolver := resolver.NewWorkspaceResolver(sqlDS)
		result, err := wsResolver.Resolve(buildFlags.ToFilter())
		if err != nil {
			// Check if ambiguous and provide helpful output
			if ambiguousErr, ok := resolver.IsAmbiguousError(err); ok {
				render.Warning("Multiple workspaces match your criteria")
				fmt.Println(ambiguousErr.FormatDisambiguation())
				return fmt.Errorf("ambiguous workspace selection")
			}
			if resolver.IsNoWorkspaceFoundError(err) {
				render.Warning("No workspace found matching your criteria")
				render.Info("Hint: Use 'dvm get workspaces' to see available workspaces")
				return err
			}
			return fmt.Errorf("failed to resolve workspace: %w", err)
		}

		// Use resolved workspace and app
		workspace = result.Workspace
		app = result.App
		appName = app.Name
		workspaceName = workspace.Name

		// Update context to the resolved workspace
		if err := updateContextFromHierarchy(sqlDS, result); err != nil {
			slog.Warn("failed to update context", "error", err)
			// Continue anyway - this is not fatal
		}

		render.Info(fmt.Sprintf("Resolved: %s", result.FullPath()))
	} else {
		// Fall back to existing context-based behavior
		ctxMgr, err := operators.NewContextManager()
		if err != nil {
			slog.Error("failed to create context manager", "error", err)
			return fmt.Errorf("failed to create context manager: %w", err)
		}

		appName, err = ctxMgr.GetActiveApp()
		if err != nil {
			slog.Debug("no active app set")
			render.Info("Hint: Set active app with: dvm use app <name>")
			render.Info("      Or use flags: dvm build -a <app>")
			return fmt.Errorf("no active app set. Use 'dvm use app <name>' first")
		}

		workspaceName, err = ctxMgr.GetActiveWorkspace()
		if err != nil {
			slog.Debug("no active workspace set")
			render.Info("Hint: Set active workspace with: dvm use workspace <name>")
			render.Info("      Or use flags: dvm build -w <workspace>")
			return fmt.Errorf("no active workspace set. Use 'dvm use workspace <name>' first")
		}

		slog.Debug("build context", "app", appName, "workspace", workspaceName)

		// Get app (search globally across all domains)
		app, err = sqlDS.GetAppByNameGlobal(appName)
		if err != nil {
			slog.Error("failed to get app", "name", appName, "error", err)
			return fmt.Errorf("failed to get app: %w", err)
		}

		// Get workspace
		workspace, err = sqlDS.GetWorkspaceByName(app.ID, workspaceName)
		if err != nil {
			slog.Error("failed to get workspace", "name", workspaceName, "app_id", app.ID, "error", err)
			return fmt.Errorf("failed to get workspace: %w", err)
		}
	}

	render.Info(fmt.Sprintf("Building workspace: %s/%s", appName, workspaceName))
	render.Info(fmt.Sprintf("App path: %s", app.Path))
	fmt.Println()
	slog.Debug("app details", "path", app.Path, "id", app.ID)

	// Verify app path exists
	if _, err := os.Stat(app.Path); os.IsNotExist(err) {
		slog.Error("app path does not exist", "path", app.Path)
		return fmt.Errorf("app path does not exist: %s", app.Path)
	}

	// Step 1: Detect platform
	render.Progress("Detecting container platform...")
	platform, err := detectPlatform()
	if err != nil {
		return err
	}
	render.Info(fmt.Sprintf("Platform: %s", platform.Name))
	slog.Info("detected platform", "name", platform.Name, "type", platform.Type, "socket", platform.SocketPath)

	// Step 2: Detect language (use App.Language if set, fall back to auto-detection)
	fmt.Println()
	render.Progress("Detecting app language...")
	languageName, version, wasDetected := getLanguageFromApp(app)

	if !wasDetected {
		render.Info(fmt.Sprintf("Language: %s (from app config)", languageName))
		if version != "" {
			render.Info(fmt.Sprintf("Version: %s", version))
		}
		slog.Debug("using language from app config", "language", languageName, "version", version)
	} else if languageName != "unknown" {
		if version != "" {
			render.Info(fmt.Sprintf("Language: %s (version: %s)", languageName, version))
		} else {
			render.Info(fmt.Sprintf("Language: %s", languageName))
		}
		slog.Debug("detected language", "language", languageName, "version", version)
	} else {
		render.Info("Language: Unknown (will use generic base)")
		slog.Debug("language detection failed, using generic base")
	}

	// Step 3: Check for existing Dockerfile
	fmt.Println()
	render.Progress("Checking for Dockerfile...")
	hasDockerfile, dockerfilePath := utils.HasDockerfile(app.Path)
	if hasDockerfile {
		render.Info(fmt.Sprintf("Found: %s", dockerfilePath))
		slog.Debug("found existing Dockerfile", "path", dockerfilePath)
	} else {
		render.Info("No Dockerfile found, will generate from scratch")
		slog.Debug("no Dockerfile found, will generate")
	}

	// Step 4: Generate workspace spec (for now, use defaults)
	workspaceYAML := workspace.ToYAML(appName)

	// Set some sensible defaults if not configured
	if workspaceYAML.Spec.Shell.Type == "" {
		workspaceYAML.Spec.Shell.Type = "zsh"
		workspaceYAML.Spec.Shell.Framework = "oh-my-zsh"
		workspaceYAML.Spec.Shell.Theme = "starship"
	}

	if workspaceYAML.Spec.Container.WorkingDir == "" {
		workspaceYAML.Spec.Container.WorkingDir = "/workspace"
	}

	// Get home directory for later use
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Step 5: Generate nvim config BEFORE Dockerfile (so Dockerfile generator can see .config/nvim/)
	if workspaceYAML.Spec.Nvim.Structure != "" && workspaceYAML.Spec.Nvim.Structure != "none" {
		if err := copyNvimConfig(workspaceYAML.Spec.Nvim.Plugins, app.Path, homeDir, sqlDS, app, workspace, appName, workspaceName); err != nil {
			return err
		}
	}

	// Step 6: Generate Dockerfile (after nvim config so it can detect .config/nvim/)
	fmt.Println()
	render.Progress("Generating Dockerfile.dvm...")
	slog.Debug("generating Dockerfile", "language", languageName, "version", version)
	generator := builders.NewDockerfileGenerator(
		workspace,
		workspaceYAML.Spec,
		languageName,
		version,
		app.Path,
		dockerfilePath,
	)

	dockerfileContent, err := generator.Generate()
	if err != nil {
		slog.Error("failed to generate Dockerfile", "error", err)
		return fmt.Errorf("failed to generate Dockerfile: %w", err)
	}

	// Save Dockerfile
	dvmDockerfile, err := builders.SaveDockerfile(dockerfileContent, app.Path)
	if err != nil {
		slog.Error("failed to save Dockerfile", "error", err)
		return err
	}
	slog.Debug("saved Dockerfile", "path", dvmDockerfile)

	// Step 6: Build image
	// Use timestamp tag for versioning (enables container recreation on rebuild)
	timestamp := time.Now().Format("20060102-150405")
	imageName := fmt.Sprintf("dvm-%s-%s:%s", workspaceName, appName, timestamp)
	fmt.Println()
	render.Progress(fmt.Sprintf("Building image: %s", imageName))
	slog.Info("building image", "image", imageName, "dockerfile", dvmDockerfile)

	// Create image builder using the factory (decoupled from platform specifics)
	// Use staging directory as build context (contains app source + generated configs)
	stagingDir := filepath.Join(homeDir, ".devopsmaestro", "build-staging", filepath.Base(app.Path))
	buildContext := stagingDir // Use staging directory as build context

	// If staging directory doesn't exist, fall back to app path
	if _, err := os.Stat(stagingDir); os.IsNotExist(err) {
		buildContext = app.Path
		slog.Warn("staging directory not found, using app path as build context", "staging", stagingDir, "fallback", app.Path)
	}

	builder, err := builders.NewImageBuilder(builders.BuilderConfig{
		Platform:   platform,
		Namespace:  "devopsmaestro",
		AppPath:    buildContext,
		ImageName:  imageName,
		Dockerfile: dvmDockerfile,
	})
	if err != nil {
		return fmt.Errorf("failed to create builder: %w", err)
	}
	defer builder.Close()

	// Check if image exists (skip if --force)
	ctx := context.Background()
	if !buildForce {
		exists, err := builder.ImageExists(ctx)
		if err == nil && exists {
			slog.Debug("image already exists, skipping build", "image", imageName)
			render.Info(fmt.Sprintf("Image already exists: %s", imageName))
			render.Info("Use --force to rebuild")
			return nil
		}
	}

	// Prepare build args (from environment and config)
	buildArgs := make(map[string]string)

	// First, merge App's build args (lowest priority - can be overridden)
	if buildConfig := app.GetBuildConfig(); buildConfig != nil {
		for k, v := range buildConfig.Args {
			buildArgs[k] = v
			slog.Debug("using build arg from app config", "key", k)
		}
	}

	// Load and resolve credentials from the hierarchy
	resolvedCreds := loadBuildCredentials(sqlDS, app, workspace)
	for k, v := range resolvedCreds {
		buildArgs[k] = v
		slog.Debug("using credential", "key", k)
	}

	slog.Debug("starting image build", "target", buildTarget, "no_cache", buildNocache)

	// Build the image
	if err := builder.Build(ctx, builders.BuildOptions{
		BuildArgs: buildArgs,
		Target:    buildTarget,
		NoCache:   buildNocache,
	}); err != nil {
		slog.Error("build failed", "image", imageName, "error", err)
		return err
	}
	slog.Info("build completed", "image", imageName)

	// Step 6.5: For Colima/BuildKit, copy image to devopsmaestro namespace
	// BuildKit creates images in its own namespace
	if platform.IsContainerd() {
		if err := copyImageToNamespace(platform, imageName); err != nil {
			return err
		}
	}

	// Step 7: Update workspace image name in database
	workspace.ImageName = imageName
	if err := sqlDS.UpdateWorkspace(workspace); err != nil {
		render.Warning(fmt.Sprintf("Failed to update workspace image name: %v", err))
	}

	fmt.Println()
	render.Success("Build complete!")
	render.Info(fmt.Sprintf("Image: %s", imageName))
	render.Info(fmt.Sprintf("Dockerfile: %s", dvmDockerfile))
	fmt.Println()
	render.Info("Next: Attach to your workspace with: dvm attach")

	return nil
}

// detectPlatform detects and validates the container platform
func detectPlatform() (*operators.Platform, error) {
	detector, err := operators.NewPlatformDetector()
	if err != nil {
		return nil, fmt.Errorf("failed to create platform detector: %w", err)
	}

	platform, err := detector.Detect()
	if err != nil {
		return nil, fmt.Errorf("no container platform available: %w\n\n%s", err, getPlatformInstallHint())
	}

	return platform, nil
}

// getLanguageFromApp extracts language config from App, falls back to detection.
// Returns (languageName, version, wasDetected) - wasDetected is true if we fell back to auto-detection.
func getLanguageFromApp(app *models.App) (langName, version string, detected bool) {
	// Try App.Language first (uses model's GetLanguageConfig method)
	if langConfig := app.GetLanguageConfig(); langConfig != nil {
		slog.Debug("using language from app model", "language", langConfig.Name, "version", langConfig.Version)
		return langConfig.Name, langConfig.Version, false
	}

	// Fall back to auto-detection
	lang, err := utils.DetectLanguage(app.Path)
	if err != nil {
		slog.Debug("language detection error", "error", err)
		return "unknown", "", true
	}

	if lang != nil {
		ver := utils.DetectVersion(lang.Name, app.Path)
		return lang.Name, ver, true
	}

	return "unknown", "", true
}

// getPlatformInstallHint returns helpful installation instructions
func getPlatformInstallHint() string {
	return `Install one of the following:
  - OrbStack (recommended): brew install orbstack
  - Colima: brew install colima && colima start --runtime containerd
  - Docker Desktop: https://docker.com/products/docker-desktop
  - Podman: brew install podman && podman machine init && podman machine start`
}

// copyNvimConfig generates nvim configuration using nvp and copies to build context
// It filters plugins based on the workspace's configured plugin list
// Reads plugin data from the database (source of truth)
func copyNvimConfig(workspacePlugins []string, appPath, homeDir string, ds db.DataStore, app *models.App, workspace *models.Workspace, appName, workspaceName string) error {
	render.Progress("Preparing build staging directory...")

	// Create staging directory for build artifacts instead of placing in app directory
	stagingDir := filepath.Join(homeDir, ".devopsmaestro", "build-staging", filepath.Base(appPath))

	// Clean and recreate staging directory
	if err := os.RemoveAll(stagingDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clean staging directory: %w", err)
	}

	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return fmt.Errorf("failed to create staging directory: %w", err)
	}

	// Copy app source to staging directory (for Dockerfile COPY commands)
	render.Progress("Copying application source...")
	if err := copyAppSource(appPath, stagingDir); err != nil {
		return fmt.Errorf("failed to copy app source: %w", err)
	}

	nvimConfigPath := filepath.Join(stagingDir, ".config", "nvim")
	if err := os.MkdirAll(nvimConfigPath, 0755); err != nil {
		return fmt.Errorf("failed to create nvim config directory: %w", err)
	}

	// Load core config from ~/.nvp/core.yaml or use defaults
	nvpDir := filepath.Join(homeDir, ".nvp")
	coreConfigPath := filepath.Join(nvpDir, "core.yaml")

	var cfg *nvimconfig.CoreConfig
	var err error

	if _, statErr := os.Stat(coreConfigPath); statErr == nil {
		cfg, err = nvimconfig.ParseYAMLFile(coreConfigPath)
		if err != nil {
			slog.Warn("failed to parse core.yaml, using defaults", "error", err)
			cfg = nvimconfig.DefaultCoreConfig()
		}
	} else {
		slog.Debug("no core.yaml found, using defaults")
		cfg = nvimconfig.DefaultCoreConfig()
	}

	// Load plugins from database (source of truth)
	dbAdapter := store.NewDBStoreAdapter(ds)
	allPlugins, err := dbAdapter.List()
	if err != nil {
		slog.Warn("failed to list plugins from database", "error", err)
		allPlugins = []*plugin.Plugin{}
	}

	// Initialize plugin library for fallback
	pluginLibrary, err := library.NewLibrary()
	if err != nil {
		slog.Warn("failed to initialize plugin library", "error", err)
		pluginLibrary = nil
	}

	// Build a map of plugin names for quick lookup
	pluginMap := make(map[string]*plugin.Plugin)
	for _, p := range allPlugins {
		if p.Enabled {
			pluginMap[p.Name] = p
		}
	}

	// Filter plugins based on workspace configuration
	var enabledPlugins []*plugin.Plugin
	if len(workspacePlugins) > 0 {
		// Workspace has a specific plugin list - use only those
		for _, name := range workspacePlugins {
			if p, ok := pluginMap[name]; ok {
				enabledPlugins = append(enabledPlugins, p)
			} else {
				// Try loading from library as fallback
				if pluginLibrary != nil {
					if libPlugin, found := pluginLibrary.Get(name); found {
						enabledPlugins = append(enabledPlugins, libPlugin)
						slog.Debug("loaded workspace plugin from library", "plugin", name)
					} else {
						slog.Warn("workspace references unknown plugin", "plugin", name)
						render.Warning(fmt.Sprintf("Plugin '%s' not found in database or library (skipping)", name))
					}
				} else {
					slog.Warn("workspace references unknown plugin", "plugin", name)
					render.Warning(fmt.Sprintf("Plugin '%s' not found in database (skipping)", name))
				}
			}
		}
		slog.Debug("using workspace-specific plugins", "count", len(enabledPlugins), "requested", len(workspacePlugins))
	} else {
		// No plugins configured for workspace - check for default package
		defaultPkg, err := ds.GetDefault("nvim-package")
		if err == nil && defaultPkg != "" {
			// Try to resolve the default package
			packagePlugins, err := resolveDefaultPackagePlugins(defaultPkg, ds)
			if err != nil {
				slog.Warn("failed to resolve default package, falling back to all enabled plugins", "package", defaultPkg, "error", err)
				render.Warning(fmt.Sprintf("Failed to resolve default package '%s', using all enabled plugins", defaultPkg))
			} else {
				// Use plugins from the resolved package
				for _, pluginName := range packagePlugins {
					if p, ok := pluginMap[pluginName]; ok {
						enabledPlugins = append(enabledPlugins, p)
					} else {
						// Try loading from library as fallback
						if pluginLibrary != nil {
							if libPlugin, found := pluginLibrary.Get(pluginName); found {
								enabledPlugins = append(enabledPlugins, libPlugin)
								slog.Debug("loaded package plugin from library", "plugin", pluginName, "package", defaultPkg)
							} else {
								slog.Warn("default package references unknown plugin", "plugin", pluginName, "package", defaultPkg)
								render.Warning(fmt.Sprintf("Plugin '%s' from package '%s' not found in database or library (skipping)", pluginName, defaultPkg))
							}
						} else {
							slog.Warn("default package references unknown plugin", "plugin", pluginName, "package", defaultPkg)
							render.Warning(fmt.Sprintf("Plugin '%s' from package '%s' not found in database (skipping)", pluginName, defaultPkg))
						}
					}
				}
				slog.Debug("using plugins from default package", "package", defaultPkg, "count", len(enabledPlugins), "resolved_plugins", len(packagePlugins))
			}
		}

		// If no default package or resolution failed, fall back to all enabled plugins
		if len(enabledPlugins) == 0 {
			for _, p := range allPlugins {
				if p.Enabled {
					enabledPlugins = append(enabledPlugins, p)
				}
			}
			slog.Debug("no default package or resolution failed, using all enabled plugins", "count", len(enabledPlugins))
		}
	}

	slog.Debug("loaded nvp config", "plugins", len(enabledPlugins), "core_config", coreConfigPath)

	// Generate the full nvim config structure
	gen := nvimconfig.NewGenerator()
	if err := gen.WriteToDirectory(cfg, enabledPlugins, nvimConfigPath); err != nil {
		return fmt.Errorf("failed to generate nvim config: %w", err)
	}

	// Generate theme if active
	themeDir := filepath.Join(nvpDir, "themes")
	if _, statErr := os.Stat(themeDir); statErr == nil {
		themeStore := theme.NewFileStore(nvpDir)
		if activeTheme, _ := themeStore.GetActive(); activeTheme != nil {
			themeGen := theme.NewGenerator()
			generated, err := themeGen.Generate(activeTheme)
			if err == nil {
				ns := cfg.Namespace
				if ns == "" {
					ns = "workspace"
				}

				// Write theme files
				themeFiles := map[string]string{
					filepath.Join(nvimConfigPath, "lua", "theme", "palette.lua"):           generated.PaletteLua,
					filepath.Join(nvimConfigPath, "lua", "theme", "init.lua"):              generated.InitLua,
					filepath.Join(nvimConfigPath, "lua", ns, "plugins", "colorscheme.lua"): generated.PluginLua,
				}

				for path, content := range themeFiles {
					dir := filepath.Dir(path)
					if err := os.MkdirAll(dir, 0755); err != nil {
						slog.Warn("failed to create theme dir", "path", dir, "error", err)
						continue
					}
					if err := os.WriteFile(path, []byte(content), 0644); err != nil {
						slog.Warn("failed to write theme file", "path", path, "error", err)
						continue
					}
				}
				slog.Debug("generated theme", "name", activeTheme.Name)
			}
		}
	}

	// Generate shell configuration files (.zshrc and starship.toml)
	if err := generateShellConfig(stagingDir, appName, workspaceName, ds); err != nil {
		return fmt.Errorf("failed to generate shell config: %w", err)
	}

	render.Success(fmt.Sprintf("Neovim configuration generated (%d plugins)", len(enabledPlugins)))

	return nil
}

// copyImageToNamespace copies the built image from buildkit namespace to devopsmaestro namespace
// This is needed because BuildKit creates images in its own namespace
func copyImageToNamespace(platform *operators.Platform, imageName string) error {
	fmt.Println()
	render.Progress("Copying image to devopsmaestro namespace...")

	profile := platform.Profile
	if profile == "" {
		profile = "default"
	}

	tmpFile := fmt.Sprintf("/tmp/dvm-image-%d.tar", os.Getpid())

	// Save image from buildkit namespace
	saveCmd := exec.Command("colima", "--profile", profile, "ssh", "--",
		"sudo", "nerdctl", "--namespace", "buildkit", "image", "save", imageName, "-o", tmpFile)
	saveCmd.Stdout = os.Stdout
	saveCmd.Stderr = os.Stderr
	if err := saveCmd.Run(); err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	// Load image into devopsmaestro namespace
	loadCmd := exec.Command("colima", "--profile", profile, "ssh", "--",
		"sudo", "nerdctl", "--namespace", "devopsmaestro", "image", "load", "-i", tmpFile)
	loadCmd.Stdout = os.Stdout
	loadCmd.Stderr = os.Stderr
	if err := loadCmd.Run(); err != nil {
		return fmt.Errorf("failed to load image: %w", err)
	}

	// Clean up temp file
	cleanCmd := exec.Command("colima", "--profile", profile, "ssh", "--", "sudo", "rm", "-f", tmpFile)
	cleanCmd.Run() // Ignore errors on cleanup

	render.Success("Image copied to devopsmaestro namespace")
	return nil
}

// Helper function to get relative path for display
func getRelativePath(base, target string) string {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	return rel
}

// loadBuildCredentials loads and resolves credentials from the hierarchy:
// Global -> Ecosystem -> Domain -> App -> Workspace
// Environment variables always take highest priority.
func loadBuildCredentials(ds db.DataStore, app *models.App, workspace *models.Workspace) map[string]string {
	var scopes []config.CredentialScope

	// Layer 1: Global credentials from config file
	globalCreds := config.GetGlobalCredentials()
	if len(globalCreds) > 0 {
		scopes = append(scopes, config.CredentialScope{
			Type:        "global",
			ID:          0,
			Name:        "global",
			Credentials: globalCreds,
		})
		slog.Debug("loaded global credentials", "count", len(globalCreds))
	}

	// Layer 2: Ecosystem credentials (if app belongs to a domain with an ecosystem)
	if app.DomainID > 0 {
		domain, err := ds.GetDomainByID(app.DomainID)
		if err == nil && domain.EcosystemID > 0 {
			ecosystem, err := ds.GetEcosystemByID(domain.EcosystemID)
			if err == nil {
				ecoCreds, err := ds.ListCredentialsByScope(models.CredentialScopeEcosystem, int64(ecosystem.ID))
				if err == nil && len(ecoCreds) > 0 {
					scopes = append(scopes, config.CredentialScope{
						Type:        "ecosystem",
						ID:          int64(ecosystem.ID),
						Name:        ecosystem.Name,
						Credentials: models.CredentialsToMap(ecoCreds),
					})
					slog.Debug("loaded ecosystem credentials", "ecosystem", ecosystem.Name, "count", len(ecoCreds))
				}
			}

			// Layer 3: Domain credentials
			domainCreds, err := ds.ListCredentialsByScope(models.CredentialScopeDomain, int64(domain.ID))
			if err == nil && len(domainCreds) > 0 {
				scopes = append(scopes, config.CredentialScope{
					Type:        "domain",
					ID:          int64(domain.ID),
					Name:        domain.Name,
					Credentials: models.CredentialsToMap(domainCreds),
				})
				slog.Debug("loaded domain credentials", "domain", domain.Name, "count", len(domainCreds))
			}
		}
	}

	// Layer 4: App credentials
	appCreds, err := ds.ListCredentialsByScope(models.CredentialScopeApp, int64(app.ID))
	if err == nil && len(appCreds) > 0 {
		scopes = append(scopes, config.CredentialScope{
			Type:        "app",
			ID:          int64(app.ID),
			Name:        app.Name,
			Credentials: models.CredentialsToMap(appCreds),
		})
		slog.Debug("loaded app credentials", "app", app.Name, "count", len(appCreds))
	}

	// Layer 5: Workspace credentials
	if workspace != nil {
		wsCreds, err := ds.ListCredentialsByScope(models.CredentialScopeWorkspace, int64(workspace.ID))
		if err == nil && len(wsCreds) > 0 {
			scopes = append(scopes, config.CredentialScope{
				Type:        "workspace",
				ID:          int64(workspace.ID),
				Name:        workspace.Name,
				Credentials: models.CredentialsToMap(wsCreds),
			})
			slog.Debug("loaded workspace credentials", "workspace", workspace.Name, "count", len(wsCreds))
		}
	}

	// Resolve all credentials (env vars checked last internally)
	resolved, errors := config.ResolveCredentialsWithErrors(scopes...)

	// Log any resolution errors
	for name, err := range errors {
		slog.Warn("failed to resolve credential", "name", name, "error", err)
	}

	if len(resolved) > 0 {
		slog.Info("resolved build credentials", "count", len(resolved))
	}

	return resolved
}

// copyAppSource copies application source code to staging directory, excluding generated files
func copyAppSource(srcDir, dstDir string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// Skip certain directories and files
		if shouldSkipPath(relPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		dstPath := filepath.Join(dstDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath, info.Mode())
	})
}

// shouldSkipPath determines if a path should be skipped during app source copy
func shouldSkipPath(path string) bool {
	skipDirs := []string{".git", ".devopsmaestro", "node_modules", "vendor", "__pycache__", ".venv", "venv"}
	skipFiles := []string{".DS_Store", "Thumbs.db", "*.log", "Dockerfile.dvm"}

	for _, skip := range skipDirs {
		if strings.HasPrefix(path, skip+"/") || path == skip {
			return true
		}
	}

	for _, skip := range skipFiles {
		if matched, _ := filepath.Match(skip, filepath.Base(path)); matched {
			return true
		}
	}

	return false
}

// copyFile copies a single file
func copyFile(src, dst string, mode os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination directory if needed
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return os.Chmod(dst, mode)
}

// generateShellConfig creates .zshrc and starship.toml files in staging directory
func generateShellConfig(stagingDir, appName, workspaceName string, ds db.DataStore) error {
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

	// Use Resource/Handler pattern to get or create default prompt
	ctx := resource.Context{DataStore: ds}
	defaultPromptName := fmt.Sprintf("dvm-default-%s-%s", appName, workspaceName)

	// Declare promptYAML variable
	var promptYAML *prompt.PromptYAML

	// Try to get existing prompt first
	res, err := resource.Get(ctx, prompt.KindTerminalPrompt, defaultPromptName)
	if err != nil {
		// Prompt doesn't exist, create default and apply it
		slog.Debug("creating default terminal prompt", "name", defaultPromptName, "app", appName, "workspace", workspaceName)

		promptYAML = createDefaultTerminalPrompt(appName, workspaceName)
		yamlData, err := yaml.Marshal(promptYAML)
		if err != nil {
			// If YAML marshaling fails, fall back to hardcoded config
			slog.Warn("failed to marshal default prompt, using direct creation", "error", err)
			promptYAML = createDefaultTerminalPrompt(appName, workspaceName)
		} else {
			// Apply to database
			res, err = resource.Apply(ctx, yamlData, "build-default")
			if err != nil {
				// If database fails, fall back to direct creation
				slog.Warn("failed to apply default prompt to database, using direct creation", "error", err)
				promptYAML = createDefaultTerminalPrompt(appName, workspaceName)
			} else {
				// Successfully applied to database, extract the prompt
				promptRes, ok := res.(*handlers.TerminalPromptResource)
				if !ok {
					slog.Warn("unexpected resource type from handler, using direct creation")
					promptYAML = createDefaultTerminalPrompt(appName, workspaceName)
				} else {
					actualPrompt := promptRes.Prompt()
					promptYAML = actualPrompt.ToYAML()
					slog.Debug("applied default prompt to database", "name", defaultPromptName)
				}
			}
		}
	} else {
		// Always regenerate default prompt to ensure latest template changes
		// This fixes issue with stale prompts that have old double-quote format
		slog.Debug("regenerating default prompt to ensure latest template", "name", defaultPromptName)
		promptYAML = createDefaultTerminalPrompt(appName, workspaceName)

		// Update the database with the fresh prompt
		yamlData, err := yaml.Marshal(promptYAML)
		if err != nil {
			slog.Warn("failed to marshal updated prompt", "error", err)
			// Continue with the fresh prompt anyway
		} else {
			// Apply updated prompt to database
			_, err = resource.Apply(ctx, yamlData, "build-refresh")
			if err != nil {
				slog.Warn("failed to update prompt in database", "error", err)
				// Continue anyway - we have the fresh prompt to use
			} else {
				slog.Debug("updated prompt in database with latest template", "name", defaultPromptName)
			}
		}
	}

	// Create a default palette for now (future: get from theme system)
	defaultPalette := createDefaultPalette()

	// Use the new renderer to generate starship.toml
	renderer := prompt.NewRenderer()
	starshipPath := filepath.Join(configDir, "starship.toml")
	if err := renderer.RenderToFile(promptYAML, defaultPalette, starshipPath); err != nil {
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

// appendPluginLoading appends terminal plugin loading configuration to the .zshrc file.
func appendPluginLoading(zshrcPath string, ds db.DataStore) error {
	// Get enabled terminal plugins from database
	plugins, err := ds.ListTerminalPlugins()
	if err != nil {
		return fmt.Errorf("failed to list plugins: %w", err)
	}

	// Filter to only enabled plugins and convert to plugin.Plugin
	var enabledPlugins []*terminalplugin.Plugin
	for _, dbPlugin := range plugins {
		if dbPlugin.Enabled {
			// Convert to plugin.Plugin using conversion pattern
			p := dbModelToPlugin(dbPlugin)
			enabledPlugins = append(enabledPlugins, p)
		}
	}

	if len(enabledPlugins) == 0 {
		return nil // No plugins to load - non-fatal
	}

	// Use pkg/terminalops/plugin/generator.go to generate loading script
	generator := terminalplugin.NewZshGenerator("$HOME/.local/share/zsh/plugins")
	pluginScript, err := generator.Generate(enabledPlugins)
	if err != nil {
		return fmt.Errorf("failed to generate plugin script: %w", err)
	}

	// Append to existing .zshrc file
	file, err := os.OpenFile(zshrcPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open zshrc for appending: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString("\n" + pluginScript); err != nil {
		return fmt.Errorf("failed to append plugin script: %w", err)
	}

	return nil
}

// dbModelToPlugin converts a models.TerminalPluginDB to terminalplugin.Plugin.
// This is adapted from pkg/terminalops/store/db_adapter.go
func dbModelToPlugin(db *models.TerminalPluginDB) *terminalplugin.Plugin {
	p := &terminalplugin.Plugin{
		Name:    db.Name,
		Repo:    db.Repo,
		Enabled: db.Enabled,
	}

	// String fields
	if db.Description.Valid {
		p.Description = db.Description.String
	}
	if db.Category.Valid {
		p.Category = db.Category.String
	}

	// Plugin manager
	p.Manager = terminalplugin.PluginManager(db.Manager)

	// Load command and config
	if db.LoadCommand.Valid {
		p.Config = db.LoadCommand.String

		// If it looks like an oh-my-zsh plugin, extract the plugin name
		if strings.HasPrefix(db.LoadCommand.String, "plugins+=") {
			p.OhMyZshPlugin = strings.TrimPrefix(db.LoadCommand.String, "plugins+=")
		}
	}

	// Source file
	if db.SourceFile.Valid {
		p.SourceFiles = []string{db.SourceFile.String}
	}

	// Parse dependencies JSON
	if db.Dependencies != "" && db.Dependencies != "[]" {
		var deps []string
		if err := json.Unmarshal([]byte(db.Dependencies), &deps); err == nil {
			p.Dependencies = deps
		}
	}

	// Parse env vars JSON
	if db.EnvVars != "" && db.EnvVars != "{}" {
		var envVars map[string]string
		if err := json.Unmarshal([]byte(db.EnvVars), &envVars); err == nil {
			p.Env = envVars
		}
	}

	// Parse labels JSON and extract metadata
	if db.Labels != "" && db.Labels != "{}" {
		var labels map[string]string
		if err := json.Unmarshal([]byte(db.Labels), &labels); err == nil {
			// Extract tags
			var tags []string
			for key, value := range labels {
				if strings.HasPrefix(key, "tag:") && value == "true" {
					tags = append(tags, strings.TrimPrefix(key, "tag:"))
				}
			}
			p.Tags = tags

			// Extract other metadata
			if loadMode, ok := labels["load_mode"]; ok {
				p.LoadMode = terminalplugin.LoadMode(loadMode)
			}
			if branch, ok := labels["branch"]; ok {
				p.Branch = branch
			}
			if tag, ok := labels["tag"]; ok {
				p.Tag = tag
			}
			if priorityStr, ok := labels["priority"]; ok {
				var priority int
				fmt.Sscanf(priorityStr, "%d", &priority)
				p.Priority = priority
			}
		}
	}

	// Timestamps
	if !db.CreatedAt.IsZero() {
		p.CreatedAt = &db.CreatedAt
	}
	if !db.UpdatedAt.IsZero() {
		p.UpdatedAt = &db.UpdatedAt
	}

	return p
}

// resolveDefaultPackagePlugins resolves plugins from a default package name.
// It first checks the embedded library, then falls back to database packages.
func resolveDefaultPackagePlugins(packageName string, ds db.DataStore) ([]string, error) {
	// First, try to load from embedded library
	lib, err := packagelibrary.NewLibrary()
	if err != nil {
		return nil, fmt.Errorf("failed to create package library: %w", err)
	}

	if pkg, ok := lib.Get(packageName); ok {
		// Package found in library - resolve plugins including inheritance
		return resolvePackagePlugins(pkg, lib)
	}

	// Package not in library - try database
	dbPkg, err := ds.GetPackage(packageName)
	if err != nil {
		return nil, fmt.Errorf("package '%s' not found in library or database: %w", packageName, err)
	}

	// Convert database model to package model
	pkg := &nvimpackage.Package{
		Name:        dbPkg.Name,
		Description: dbPkg.Description.String,
		Category:    dbPkg.Category.String,
		Tags:        []string{}, // Database packages don't have tags in current schema
		Extends:     dbPkg.Extends.String,
		Plugins:     dbPkg.GetPlugins(),
		Enabled:     true, // Database packages are enabled by default
	}

	// Clean up plugins (they come from JSON so should already be clean, but just in case)
	var cleanPlugins []string
	for _, plugin := range pkg.Plugins {
		plugin = strings.TrimSpace(plugin)
		if plugin != "" {
			cleanPlugins = append(cleanPlugins, plugin)
		}
	}
	pkg.Plugins = cleanPlugins

	// For database packages, we need to handle inheritance manually
	// since we can't use the library's resolution logic
	if pkg.Extends != "" {
		// Try to resolve parent from library first
		if parentPkg, ok := lib.Get(pkg.Extends); ok {
			parentPlugins, err := resolvePackagePlugins(parentPkg, lib)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve parent package '%s' from library: %w", pkg.Extends, err)
			}
			// Combine parent plugins with current package plugins
			allPlugins := append(parentPlugins, pkg.Plugins...)
			return removeDuplicates(allPlugins), nil
		}

		// Parent not in library - try database
		parentDBPkg, err := ds.GetPackage(pkg.Extends)
		if err != nil {
			return nil, fmt.Errorf("parent package '%s' not found in library or database: %w", pkg.Extends, err)
		}

		// Simple inheritance for database packages (no deep recursion to avoid complexity)
		parentPlugins := parentDBPkg.GetPlugins()

		// Combine parent and current plugins
		allPlugins := append(parentPlugins, pkg.Plugins...)
		return removeDuplicates(allPlugins), nil
	}

	// No inheritance - return current package plugins
	return pkg.Plugins, nil
}

// resolvePackagePlugins resolves all plugins from a package including inheritance.
// This is based on the same function in cmd/nvp/package.go.
func resolvePackagePlugins(pkg *nvimpackage.Package, lib *packagelibrary.Library) ([]string, error) {
	var result []string
	visited := make(map[string]bool)

	var resolve func(p *nvimpackage.Package) error
	resolve = func(p *nvimpackage.Package) error {
		if visited[p.Name] {
			return fmt.Errorf("circular dependency detected: %s", p.Name)
		}
		visited[p.Name] = true
		defer func() { visited[p.Name] = false }()

		// If this package extends another, resolve parent first
		if p.Extends != "" {
			parent, ok := lib.Get(p.Extends)
			if !ok {
				return fmt.Errorf("package %s extends %s, but %s not found in library", p.Name, p.Extends, p.Extends)
			}
			if err := resolve(parent); err != nil {
				return err
			}
		}

		// Add this package's plugins
		for _, pluginName := range p.Plugins {
			if !contains(result, pluginName) {
				result = append(result, pluginName)
			}
		}

		return nil
	}

	err := resolve(pkg)
	return result, err
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// removeDuplicates removes duplicate strings from a slice while preserving order
func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

// generateWezTermConfig creates a WezTerm configuration file if a terminal emulator config exists in database
func generateWezTermConfig(stagingDir, appName, workspaceName string, ds db.DataStore) error {
	// 1. Look for workspace-specific emulator first
	//    Pattern: "{app}-{workspace}" or "{workspace}"
	workspaceEmulatorName := fmt.Sprintf("%s-%s", appName, workspaceName)
	emulatorDB, err := ds.GetTerminalEmulator(workspaceEmulatorName)
	if err != nil {
		// Try just workspace name
		emulatorDB, err = ds.GetTerminalEmulator(workspaceName)
		if err != nil {
			// 2. Fall back to default emulator if set
			defaultEmulatorName, err := ds.GetDefault("terminal-emulator")
			if err != nil || defaultEmulatorName == "" {
				// No emulator config found - not an error, just skip
				slog.Debug("no terminal emulator configuration found",
					"workspaceEmulator", workspaceEmulatorName,
					"workspace", workspaceName,
					"default", "not set")
				return nil
			}
			emulatorDB, err = ds.GetTerminalEmulator(defaultEmulatorName)
			if err != nil {
				return fmt.Errorf("default terminal emulator '%s' not found: %w", defaultEmulatorName, err)
			}
		}
	}

	// 3. Check if it's a wezterm emulator
	if emulatorDB.Type != "wezterm" {
		slog.Debug("terminal emulator is not wezterm type", "name", emulatorDB.Name, "type", emulatorDB.Type)
		return nil
	}

	// 4. Parse the configuration from JSON to WezTerm struct
	config, err := emulatorDB.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to parse emulator config: %w", err)
	}

	// Create WezTerm configuration
	weztermConfig := &wezterm.WezTerm{
		Name:        emulatorDB.Name,
		Description: emulatorDB.Description.String,
		Workspace:   workspaceName,
		Enabled:     emulatorDB.Enabled,
	}

	// Map JSON config to WezTerm struct fields
	if err := mapConfigToWezTerm(config, weztermConfig); err != nil {
		return fmt.Errorf("failed to map config to WezTerm struct: %w", err)
	}

	// 5. Use wezterm.LuaGenerator to generate config
	generator := wezterm.NewLuaGenerator()
	luaConfig, err := generator.GenerateFromConfig(weztermConfig)
	if err != nil {
		return fmt.Errorf("failed to generate wezterm lua config: %w", err)
	}

	// 6. Write to stagingDir/.wezterm.lua
	weztermPath := filepath.Join(stagingDir, ".wezterm.lua")
	if err := os.WriteFile(weztermPath, []byte(luaConfig), 0644); err != nil {
		return fmt.Errorf("failed to write wezterm config: %w", err)
	}

	slog.Debug("generated wezterm config", "name", emulatorDB.Name, "path", weztermPath)
	return nil
}

// mapConfigToWezTerm maps a generic config map to WezTerm struct fields
func mapConfigToWezTerm(config map[string]any, wt *wezterm.WezTerm) error {
	// Set defaults
	wt.Font = wezterm.FontConfig{
		Family: "MesloLGS Nerd Font Mono",
		Size:   14,
	}
	wt.Window = wezterm.WindowConfig{
		Opacity: 1.0,
	}

	// Map font configuration
	if fontConfig, ok := config["font"].(map[string]any); ok {
		if family, ok := fontConfig["family"].(string); ok {
			wt.Font.Family = family
		}
		if size, ok := fontConfig["size"].(float64); ok {
			wt.Font.Size = size
		} else if sizeInt, ok := fontConfig["size"].(int); ok {
			wt.Font.Size = float64(sizeInt)
		}
	}

	// Map window configuration
	if windowConfig, ok := config["window"].(map[string]any); ok {
		if opacity, ok := windowConfig["opacity"].(float64); ok {
			wt.Window.Opacity = opacity
		}
		if blur, ok := windowConfig["blur"].(int); ok {
			wt.Window.Blur = blur
		} else if blurFloat, ok := windowConfig["blur"].(float64); ok {
			wt.Window.Blur = int(blurFloat)
		}
		if decorations, ok := windowConfig["decorations"].(string); ok {
			wt.Window.Decorations = decorations
		}
		if initialRows, ok := windowConfig["initialRows"].(int); ok {
			wt.Window.InitialRows = initialRows
		} else if initialRowsFloat, ok := windowConfig["initialRows"].(float64); ok {
			wt.Window.InitialRows = int(initialRowsFloat)
		}
		if initialCols, ok := windowConfig["initialCols"].(int); ok {
			wt.Window.InitialCols = initialCols
		} else if initialColsFloat, ok := windowConfig["initialCols"].(float64); ok {
			wt.Window.InitialCols = int(initialColsFloat)
		}
		if closeOnExit, ok := windowConfig["closeOnExit"].(string); ok {
			wt.Window.CloseOnExit = closeOnExit
		}
		// Padding
		if paddingLeft, ok := windowConfig["paddingLeft"].(int); ok {
			wt.Window.PaddingLeft = paddingLeft
		} else if paddingLeftFloat, ok := windowConfig["paddingLeft"].(float64); ok {
			wt.Window.PaddingLeft = int(paddingLeftFloat)
		}
		if paddingRight, ok := windowConfig["paddingRight"].(int); ok {
			wt.Window.PaddingRight = paddingRight
		} else if paddingRightFloat, ok := windowConfig["paddingRight"].(float64); ok {
			wt.Window.PaddingRight = int(paddingRightFloat)
		}
		if paddingTop, ok := windowConfig["paddingTop"].(int); ok {
			wt.Window.PaddingTop = paddingTop
		} else if paddingTopFloat, ok := windowConfig["paddingTop"].(float64); ok {
			wt.Window.PaddingTop = int(paddingTopFloat)
		}
		if paddingBottom, ok := windowConfig["paddingBottom"].(int); ok {
			wt.Window.PaddingBottom = paddingBottom
		} else if paddingBottomFloat, ok := windowConfig["paddingBottom"].(float64); ok {
			wt.Window.PaddingBottom = int(paddingBottomFloat)
		}
	}

	// Map color configuration
	if colors, ok := config["colors"].(map[string]any); ok {
		colorConfig := &wezterm.ColorConfig{}

		if fg, ok := colors["foreground"].(string); ok {
			colorConfig.Foreground = fg
		}
		if bg, ok := colors["background"].(string); ok {
			colorConfig.Background = bg
		}
		if cursorBg, ok := colors["cursor_bg"].(string); ok {
			colorConfig.CursorBg = cursorBg
		}
		if cursorFg, ok := colors["cursor_fg"].(string); ok {
			colorConfig.CursorFg = cursorFg
		}
		if cursorBorder, ok := colors["cursor_border"].(string); ok {
			colorConfig.CursorBorder = cursorBorder
		}
		if selBg, ok := colors["selection_bg"].(string); ok {
			colorConfig.SelectionBg = selBg
		}
		if selFg, ok := colors["selection_fg"].(string); ok {
			colorConfig.SelectionFg = selFg
		}

		// ANSI colors (8 colors)
		if ansi, ok := colors["ansi"].([]any); ok {
			ansiColors := make([]string, 0, 8)
			for _, c := range ansi {
				if colorStr, ok := c.(string); ok {
					ansiColors = append(ansiColors, colorStr)
				}
			}
			colorConfig.ANSI = ansiColors
		}

		// Bright colors (8 colors)
		if brights, ok := colors["brights"].([]any); ok {
			brightColors := make([]string, 0, 8)
			for _, c := range brights {
				if colorStr, ok := c.(string); ok {
					brightColors = append(brightColors, colorStr)
				}
			}
			colorConfig.Brights = brightColors
		}

		wt.Colors = colorConfig
	}

	// Map theme reference
	if themeRef, ok := config["themeRef"].(string); ok {
		wt.ThemeRef = themeRef
	}

	// Map scrollback
	if scrollback, ok := config["scrollback"].(int); ok {
		wt.Scrollback = scrollback
	} else if scrollbackFloat, ok := config["scrollback"].(float64); ok {
		wt.Scrollback = int(scrollbackFloat)
	}

	// Map leader key
	if leader, ok := config["leader"].(map[string]any); ok {
		leaderKey := &wezterm.LeaderKey{}
		if key, ok := leader["key"].(string); ok {
			leaderKey.Key = key
		}
		if mods, ok := leader["mods"].(string); ok {
			leaderKey.Mods = mods
		}
		if timeout, ok := leader["timeout"].(int); ok {
			leaderKey.Timeout = timeout
		} else if timeoutFloat, ok := leader["timeout"].(float64); ok {
			leaderKey.Timeout = int(timeoutFloat)
		}
		wt.Leader = leaderKey
	}

	// Map key bindings
	if keys, ok := config["keys"].([]any); ok {
		keybindings := make([]wezterm.Keybinding, 0, len(keys))
		for _, k := range keys {
			if keyMap, ok := k.(map[string]any); ok {
				keybinding := wezterm.Keybinding{}
				if key, ok := keyMap["key"].(string); ok {
					keybinding.Key = key
				}
				if mods, ok := keyMap["mods"].(string); ok {
					keybinding.Mods = mods
				}
				if action, ok := keyMap["action"].(string); ok {
					keybinding.Action = action
				}
				if args, ok := keyMap["args"]; ok {
					keybinding.Args = args
				}
				keybindings = append(keybindings, keybinding)
			}
		}
		wt.Keys = keybindings
	}

	// Map tab bar configuration
	if tabBar, ok := config["tabBar"].(map[string]any); ok {
		tabBarConfig := &wezterm.TabBarConfig{}
		if enabled, ok := tabBar["enabled"].(bool); ok {
			tabBarConfig.Enabled = enabled
		}
		if position, ok := tabBar["position"].(string); ok {
			tabBarConfig.Position = position
		}
		if maxWidth, ok := tabBar["maxWidth"].(int); ok {
			tabBarConfig.MaxWidth = maxWidth
		} else if maxWidthFloat, ok := tabBar["maxWidth"].(float64); ok {
			tabBarConfig.MaxWidth = int(maxWidthFloat)
		}
		if showNewTab, ok := tabBar["showNewTab"].(bool); ok {
			tabBarConfig.ShowNewTab = showNewTab
		}
		if fancyTabBar, ok := tabBar["fancyTabBar"].(bool); ok {
			tabBarConfig.FancyTabBar = fancyTabBar
		}
		if hideTabBarIfOnly, ok := tabBar["hideTabBarIfOnly"].(bool); ok {
			tabBarConfig.HideTabBarIfOnly = hideTabBarIfOnly
		}
		wt.TabBar = tabBarConfig
	}

	// Map pane configuration
	if pane, ok := config["pane"].(map[string]any); ok {
		paneConfig := &wezterm.PaneConfig{}
		if inactiveSat, ok := pane["inactiveSaturation"].(float64); ok {
			paneConfig.InactiveSaturation = inactiveSat
		}
		if inactiveBright, ok := pane["inactiveBrightness"].(float64); ok {
			paneConfig.InactiveBrightness = inactiveBright
		}
		wt.Pane = paneConfig
	}

	return nil
}
