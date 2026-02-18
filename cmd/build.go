package cmd

import (
	"context"
	"devopsmaestro/builders"
	"devopsmaestro/config"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/operators"
	nvimconfig "devopsmaestro/pkg/nvimops/config"
	"devopsmaestro/pkg/nvimops/plugin"
	"devopsmaestro/pkg/nvimops/store"
	"devopsmaestro/pkg/nvimops/theme"
	"devopsmaestro/pkg/resolver"
	"devopsmaestro/render"
	"devopsmaestro/utils"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
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
		if err := copyNvimConfig(workspaceYAML.Spec.Nvim.Plugins, app.Path, homeDir, sqlDS); err != nil {
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
	builder, err := builders.NewImageBuilder(builders.BuilderConfig{
		Platform:   platform,
		Namespace:  "devopsmaestro",
		AppPath:    app.Path,
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
func copyNvimConfig(workspacePlugins []string, appPath, homeDir string, ds db.DataStore) error {
	render.Progress("Generating Neovim configuration for container...")

	nvimConfigPath := filepath.Join(appPath, ".config", "nvim")
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
		// No plugins configured for workspace - use all enabled plugins
		for _, p := range allPlugins {
			if p.Enabled {
				enabledPlugins = append(enabledPlugins, p)
			}
		}
		slog.Debug("no workspace plugins configured, using all enabled", "count", len(enabledPlugins))
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
