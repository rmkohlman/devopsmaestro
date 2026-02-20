package cmd

import (
	"context"
	"devopsmaestro/builders"
	"devopsmaestro/config"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/operators"
	nvimconfig "devopsmaestro/pkg/nvimops/config"
	nvimpackage "devopsmaestro/pkg/nvimops/package"
	packagelibrary "devopsmaestro/pkg/nvimops/package/library"
	"devopsmaestro/pkg/nvimops/plugin"
	"devopsmaestro/pkg/nvimops/store"
	"devopsmaestro/pkg/nvimops/theme"
	"devopsmaestro/pkg/palette"
	"devopsmaestro/pkg/resolver"
	"devopsmaestro/pkg/resource"
	"devopsmaestro/pkg/resource/handlers"
	"devopsmaestro/pkg/terminalops/prompt"
	"devopsmaestro/render"
	"devopsmaestro/utils"
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
				slog.Warn("workspace references unknown plugin", "plugin", name)
				render.Warning(fmt.Sprintf("Plugin '%s' not found in database (skipping)", name))
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
						slog.Warn("default package references unknown plugin", "plugin", pluginName, "package", defaultPkg)
						render.Warning(fmt.Sprintf("Plugin '%s' from package '%s' not found in database (skipping)", pluginName, defaultPkg))
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
		// Prompt exists in database, use it
		promptRes, ok := res.(*handlers.TerminalPromptResource)
		if !ok {
			slog.Warn("unexpected resource type from database, using direct creation")
			promptYAML = createDefaultTerminalPrompt(appName, workspaceName)
		} else {
			actualPrompt := promptRes.Prompt()
			promptYAML = actualPrompt.ToYAML()
			slog.Debug("using existing prompt from database", "name", defaultPromptName)
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
