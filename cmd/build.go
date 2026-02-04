package cmd

import (
	"context"
	"devopsmaestro/builders"
	"devopsmaestro/db"
	"devopsmaestro/operators"
	nvimconfig "devopsmaestro/pkg/nvimops/config"
	"devopsmaestro/pkg/nvimops/plugin"
	"devopsmaestro/pkg/nvimops/store"
	"devopsmaestro/pkg/nvimops/theme"
	"devopsmaestro/render"
	"devopsmaestro/utils"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	buildForce   bool
	buildNocache bool
	buildTarget  string
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build workspace container image",
	Long: `Build a development container image for the active workspace.

This command:
- Detects the project language
- Generates or extends Dockerfile with dev tools
- Builds the image using the detected container platform
- Tags as dvm-<workspace>-<project>:latest

Supports multiple platforms:
- OrbStack (uses Docker API)
- Docker Desktop (uses Docker API)
- Podman (uses Docker API)
- Colima with containerd (uses BuildKit API)

Use DVM_PLATFORM environment variable to select a specific platform.

Examples:
  dvm build
  dvm build --force
  dvm build --no-cache
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
}

func buildWorkspace(cmd *cobra.Command) error {
	slog.Info("starting build")

	// Get current context
	ctxMgr, err := operators.NewContextManager()
	if err != nil {
		slog.Error("failed to create context manager", "error", err)
		return fmt.Errorf("failed to create context manager: %w", err)
	}

	projectName, err := ctxMgr.GetActiveProject()
	if err != nil {
		slog.Debug("no active project set")
		return fmt.Errorf("no active project set. Use 'dvm use project <name>' first")
	}

	workspaceName, err := ctxMgr.GetActiveWorkspace()
	if err != nil {
		slog.Debug("no active workspace set")
		return fmt.Errorf("no active workspace set. Use 'dvm use workspace <name>' first")
	}

	slog.Debug("build context", "project", projectName, "workspace", workspaceName)

	// Get datastore
	sqlDS, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// Get project
	project, err := sqlDS.GetProjectByName(projectName)
	if err != nil {
		slog.Error("failed to get project", "name", projectName, "error", err)
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Get workspace
	workspace, err := sqlDS.GetWorkspaceByName(project.ID, workspaceName)
	if err != nil {
		slog.Error("failed to get workspace", "name", workspaceName, "project_id", project.ID, "error", err)
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	render.Info(fmt.Sprintf("Building workspace: %s/%s", projectName, workspaceName))
	render.Info(fmt.Sprintf("Project path: %s", project.Path))
	fmt.Println()
	slog.Debug("project details", "path", project.Path, "id", project.ID)

	// Verify project path exists
	if _, err := os.Stat(project.Path); os.IsNotExist(err) {
		slog.Error("project path does not exist", "path", project.Path)
		return fmt.Errorf("project path does not exist: %s", project.Path)
	}

	// Step 1: Detect platform
	render.Progress("Detecting container platform...")
	platform, err := detectPlatform()
	if err != nil {
		return err
	}
	render.Info(fmt.Sprintf("Platform: %s", platform.Name))
	slog.Info("detected platform", "name", platform.Name, "type", platform.Type, "socket", platform.SocketPath)

	// Step 2: Detect language
	fmt.Println()
	render.Progress("Detecting project language...")
	lang, err := utils.DetectLanguage(project.Path)
	if err != nil {
		slog.Error("failed to detect language", "error", err)
		return fmt.Errorf("failed to detect language: %w", err)
	}

	var languageName, version string
	if lang != nil {
		languageName = lang.Name
		version = utils.DetectVersion(languageName, project.Path)
		if version != "" {
			render.Info(fmt.Sprintf("Language: %s (version: %s)", languageName, version))
		} else {
			render.Info(fmt.Sprintf("Language: %s", languageName))
		}
		slog.Debug("detected language", "language", languageName, "version", version)
	} else {
		languageName = "unknown"
		render.Info("Language: Unknown (will use generic base)")
		slog.Debug("language detection failed, using generic base")
	}

	// Step 3: Check for existing Dockerfile
	fmt.Println()
	render.Progress("Checking for Dockerfile...")
	hasDockerfile, dockerfilePath := utils.HasDockerfile(project.Path)
	if hasDockerfile {
		render.Info(fmt.Sprintf("Found: %s", dockerfilePath))
		slog.Debug("found existing Dockerfile", "path", dockerfilePath)
	} else {
		render.Info("No Dockerfile found, will generate from scratch")
		slog.Debug("no Dockerfile found, will generate")
	}

	// Step 4: Generate workspace spec (for now, use defaults)
	workspaceYAML := workspace.ToYAML(projectName)

	// Set some sensible defaults if not configured
	if workspaceYAML.Spec.Shell.Type == "" {
		workspaceYAML.Spec.Shell.Type = "zsh"
		workspaceYAML.Spec.Shell.Framework = "oh-my-zsh"
		workspaceYAML.Spec.Shell.Theme = "starship"
	}

	if workspaceYAML.Spec.Container.WorkingDir == "" {
		workspaceYAML.Spec.Container.WorkingDir = "/workspace"
	}

	// Step 5: Generate Dockerfile
	fmt.Println()
	render.Progress("Generating Dockerfile.dvm...")
	slog.Debug("generating Dockerfile", "language", languageName, "version", version)
	generator := builders.NewDockerfileGenerator(
		workspace,
		workspaceYAML.Spec,
		languageName,
		version,
		project.Path,
		dockerfilePath,
	)

	dockerfileContent, err := generator.Generate()
	if err != nil {
		slog.Error("failed to generate Dockerfile", "error", err)
		return fmt.Errorf("failed to generate Dockerfile: %w", err)
	}

	// Save Dockerfile
	dvmDockerfile, err := builders.SaveDockerfile(dockerfileContent, project.Path)
	if err != nil {
		slog.Error("failed to save Dockerfile", "error", err)
		return err
	}
	slog.Debug("saved Dockerfile", "path", dvmDockerfile)

	// Get home directory for later use
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Step 5.5: Copy nvim templates to build context if nvim config is enabled
	if workspaceYAML.Spec.Nvim.Structure != "" && workspaceYAML.Spec.Nvim.Structure != "none" {
		if err := copyNvimConfig(workspaceYAML.Spec.Nvim.Plugins, project.Path, homeDir, sqlDS); err != nil {
			return err
		}
	}

	// Step 6: Build image
	imageName := fmt.Sprintf("dvm-%s-%s:latest", workspaceName, projectName)
	fmt.Println()
	render.Progress(fmt.Sprintf("Building image: %s", imageName))
	slog.Info("building image", "image", imageName, "dockerfile", dvmDockerfile)

	// Create image builder using the factory (decoupled from platform specifics)
	builder, err := builders.NewImageBuilder(builders.BuilderConfig{
		Platform:    platform,
		Namespace:   "devopsmaestro",
		ProjectPath: project.Path,
		ImageName:   imageName,
		Dockerfile:  dvmDockerfile,
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
	if ghUser := os.Getenv("GITHUB_USERNAME"); ghUser != "" {
		buildArgs["GITHUB_USERNAME"] = ghUser
		slog.Debug("using GITHUB_USERNAME from environment")
	}
	if ghPat := os.Getenv("GITHUB_PAT"); ghPat != "" {
		buildArgs["GITHUB_PAT"] = ghPat
		slog.Debug("using GITHUB_PAT from environment")
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
func copyNvimConfig(workspacePlugins []string, projectPath, homeDir string, ds db.DataStore) error {
	render.Progress("Generating Neovim configuration for container...")

	nvimConfigPath := filepath.Join(projectPath, ".config", "nvim")
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
