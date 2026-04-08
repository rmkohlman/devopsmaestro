package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"devopsmaestro/builders"
	"devopsmaestro/config"
	"devopsmaestro/models"
	"devopsmaestro/pkg/buildargs/resolver"
	cacertsresolver "devopsmaestro/pkg/cacerts/resolver"
	"devopsmaestro/pkg/envvalidation"
	"devopsmaestro/pkg/registry"
	"devopsmaestro/pkg/registry/envinjector"
	wsresolver "devopsmaestro/pkg/resolver"
	"devopsmaestro/utils"

	"github.com/rmkohlman/MaestroSDK/paths"
	"github.com/rmkohlman/MaestroSDK/render"
)

// resolveWorkspaceTarget resolves the workspace from hierarchy flags or active context.
// Sets bc.app, bc.workspace, bc.appName, bc.workspaceName.
func (bc *buildContext) resolveWorkspaceTarget() error {
	if buildFlags.HasAnyFlag() {
		return bc.resolveFromHierarchyFlags()
	}
	return bc.resolveFromActiveContext()
}

// resolveFromHierarchyFlags uses the workspace resolver with hierarchy flags.
func (bc *buildContext) resolveFromHierarchyFlags() error {
	slog.Debug("using hierarchy flags", "ecosystem", buildFlags.Ecosystem,
		"domain", buildFlags.Domain, "app", buildFlags.App, "workspace", buildFlags.Workspace)

	wsResolver := wsresolver.NewWorkspaceResolver(bc.ds)
	result, err := wsResolver.Resolve(buildFlags.ToFilter())
	if err != nil {
		if ambiguousErr, ok := wsresolver.IsAmbiguousError(err); ok {
			render.Warning("Multiple workspaces match your criteria")
			render.Plain(ambiguousErr.FormatDisambiguation())
			return fmt.Errorf("ambiguous workspace selection")
		}
		if wsresolver.IsNoWorkspaceFoundError(err) {
			render.Warning("No workspace found matching your criteria")
			render.Info("Hint: Use 'dvm get workspaces' to see available workspaces")
			return err
		}
		return fmt.Errorf("failed to resolve workspace: %w", err)
	}

	bc.workspace = result.Workspace
	bc.app = result.App
	bc.appName = bc.app.Name
	bc.workspaceName = bc.workspace.Name

	render.Info(fmt.Sprintf("Resolved: %s", result.FullPath()))
	return nil
}

// resolveFromActiveContext falls back to DB-backed active app/workspace context.
func (bc *buildContext) resolveFromActiveContext() error {
	var err error
	bc.appName, err = getActiveAppFromContext(bc.ds)
	if err != nil {
		slog.Debug("no active app set")
		render.Info("Hint: Set active app with: dvm use app <name>")
		render.Info("      Or use flags: dvm build -a <app>")
		return fmt.Errorf("no active app set. Use 'dvm use app <name>' first")
	}

	bc.workspaceName, err = getActiveWorkspaceFromContext(bc.ds)
	if err != nil {
		slog.Debug("no active workspace set")
		render.Info("Hint: Set active workspace with: dvm use workspace <name>")
		render.Info("      Or use flags: dvm build -w <workspace>")
		return fmt.Errorf("no active workspace set. Use 'dvm use workspace <name>' first")
	}

	slog.Debug("build context", "app", bc.appName, "workspace", bc.workspaceName)

	bc.app, err = bc.ds.GetAppByNameGlobal(bc.appName)
	if err != nil {
		slog.Error("failed to get app", "name", bc.appName, "error", err)
		return fmt.Errorf("failed to get app: %w", err)
	}

	bc.workspace, err = bc.ds.GetWorkspaceByName(bc.app.ID, bc.workspaceName)
	if err != nil {
		slog.Error("failed to get workspace", "name", bc.workspaceName, "app_id", bc.app.ID, "error", err)
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	return nil
}

// validateAppPath verifies the app's source path exists on disk.
func (bc *buildContext) validateAppPath() error {
	if _, err := os.Stat(bc.app.Path); os.IsNotExist(err) {
		slog.Error("app path does not exist", "path", bc.app.Path)
		return fmt.Errorf("app path does not exist: %s", bc.app.Path)
	}
	return nil
}

// detectBuildPlatform detects the container platform (Docker/Colima/etc.).
// Sets bc.platform.
func (bc *buildContext) detectBuildPlatform() error {
	render.Progress("Detecting container platform...")
	platform, err := detectPlatform()
	if err != nil {
		return err
	}
	bc.platform = platform
	render.Info(fmt.Sprintf("Platform: %s", platform.Name))
	slog.Info("detected platform", "name", platform.Name, "type", platform.Type, "socket", platform.SocketPath)
	return nil
}

// prepareRegistry starts registry caches if registry is enabled.
// Sets bc.registryEndpoint and bc.registryEnvVars.
func (bc *buildContext) prepareRegistry() error {
	if !config.IsRegistryEnabled() {
		return nil
	}

	coordinator := registry.NewBuildRegistryCoordinator(
		bc.ds,
		registry.NewServiceFactory(),
		envinjector.NewEnvironmentInjector(),
	)
	regResult, regErr := coordinator.Prepare(bc.ctx)
	if regErr != nil {
		render.Warning(fmt.Sprintf("Registry preparation failed: %v", regErr))
		render.Info("Continuing build without registry cache")
		slog.Warn("registry preparation failed", "error", regErr)
		return nil
	}

	bc.registryEndpoint = regResult.OCIEndpoint
	bc.registryEnvVars = regResult.EnvVars
	for _, w := range regResult.Warnings {
		render.Warning(w)
	}
	if len(regResult.Managers) > 0 {
		render.Info(fmt.Sprintf("Started %d registry cache(s)", len(regResult.Managers)))
	}
	return nil
}

// checkDockerfile looks for an existing Dockerfile in the app directory.
// Sets bc.hasDockerfile and bc.dockerfilePath.
func (bc *buildContext) checkDockerfile() {
	render.Blank()
	render.Progress("Checking for Dockerfile...")
	bc.hasDockerfile, bc.dockerfilePath = utils.HasDockerfile(bc.app.Path)
	if bc.hasDockerfile {
		render.Info(fmt.Sprintf("Found: %s", bc.dockerfilePath))
		slog.Debug("found existing Dockerfile", "path", bc.dockerfilePath)
	} else {
		render.Info("No Dockerfile found, will generate from scratch")
		slog.Debug("no Dockerfile found, will generate")
	}
}

// prepareWorkspaceSpec converts the workspace to YAML and applies sensible defaults.
// Sets bc.workspaceYAML and bc.homeDir.
func (bc *buildContext) prepareWorkspaceSpec() error {
	// Resolve GitRepo name if GitRepoID is set
	gitRepoName := ""
	if bc.workspace.GitRepoID.Valid {
		gitRepo, err := bc.ds.GetGitRepoByID(bc.workspace.GitRepoID.Int64)
		if err == nil && gitRepo != nil {
			gitRepoName = gitRepo.Name
		}
	}
	bc.workspaceYAML = bc.workspace.ToYAML(bc.appName, gitRepoName)

	// Set some sensible defaults if not configured
	if bc.workspaceYAML.Spec.Shell.Type == "" {
		bc.workspaceYAML.Spec.Shell.Type = "zsh"
		bc.workspaceYAML.Spec.Shell.Framework = "oh-my-zsh"
		bc.workspaceYAML.Spec.Shell.Theme = "starship"
	}

	if bc.workspaceYAML.Spec.Container.WorkingDir == "" {
		bc.workspaceYAML.Spec.Container.WorkingDir = "/workspace"
	}

	var err error
	bc.homeDir, err = os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	return nil
}

// prepareSourceAndStaging determines the build source path, detects the language,
// and creates the staging directory with shell configuration.
// Sets bc.sourcePath, bc.languageName, bc.version, bc.stagingDir.
func (bc *buildContext) prepareSourceAndStaging() error {
	var err error
	bc.sourcePath, err = getBuildSourcePath(bc.ds, bc.workspace, bc.app.Path)
	if err != nil {
		return fmt.Errorf("failed to determine build source path: %w", err)
	}

	// Detect language (use App.Language if set, fall back to auto-detection)
	render.Blank()
	render.Progress("Detecting app language...")
	bc.languageName, bc.version, _ = bc.detectLanguageAndReport()

	bc.stagingDir = paths.New(bc.homeDir).BuildStagingDir(filepath.Base(bc.sourcePath))
	if err := prepareStagingDirectory(bc.stagingDir, bc.sourcePath, bc.appName, bc.workspaceName, bc.ds, bc.workspace); err != nil {
		return err
	}
	return nil
}

// detectLanguageAndReport detects the app language and emits render messages.
// Returns languageName, version, and whether it was auto-detected.
func (bc *buildContext) detectLanguageAndReport() (string, string, bool) {
	languageName, version, wasDetected := getLanguageFromApp(bc.app, bc.sourcePath)

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

	return languageName, version, wasDetected
}

// resolveCACerts resolves hierarchical CA certificates and prepares them for the build.
// Modifies bc.workspaceYAML.Spec.Build.CACerts if cascade resolution succeeds.
func (bc *buildContext) resolveCACerts() error {
	caCertsResolver := cacertsresolver.NewHierarchyCACertsResolver(bc.ds)
	caCertsResolution, caCertsErr := caCertsResolver.Resolve(context.Background(), bc.workspace.ID)
	if caCertsErr != nil {
		slog.Warn("failed to resolve hierarchical CA certs, continuing with workspace-only certs", "error", caCertsErr)
	} else if len(caCertsResolution.Certs) > 0 {
		mergedCerts := make([]models.CACertConfig, 0, len(caCertsResolution.Certs))
		for _, entry := range caCertsResolution.Certs {
			mergedCerts = append(mergedCerts, models.CACertConfig{
				Name:             entry.Name,
				VaultSecret:      entry.VaultSecret,
				VaultEnvironment: entry.VaultEnvironment,
				VaultField:       entry.VaultField,
			})
			slog.Debug("using cascaded CA cert", "name", entry.Name, "source", entry.Source.String())
		}
		bc.workspaceYAML.Spec.Build.CACerts = mergedCerts
	}

	// Prepare CA certificates if configured
	if len(bc.workspaceYAML.Spec.Build.CACerts) > 0 {
		render.Progress("Resolving CA certificates from vault...")
		if err := prepareCACerts(bc.stagingDir, bc.workspaceYAML.Spec.Build.CACerts); err != nil {
			return err
		}
		render.Info(fmt.Sprintf("Injecting %d CA certificate(s) into build context", len(bc.workspaceYAML.Spec.Build.CACerts)))
	}
	return nil
}

// generateNvimConfiguration generates nvim config if a structure is configured.
// Sets bc.pluginManifest.
func (bc *buildContext) generateNvimConfiguration() error {
	if bc.workspaceYAML.Spec.Nvim.Structure == "" || bc.workspaceYAML.Spec.Nvim.Structure == "none" {
		return nil
	}
	manifest, err := generateNvimConfig(
		bc.workspaceYAML.Spec.Nvim.Plugins, bc.stagingDir, bc.homeDir, bc.ds,
		bc.app, bc.workspace, bc.appName, bc.workspaceName, bc.languageName,
	)
	if err != nil {
		return err
	}
	bc.pluginManifest = manifest
	return nil
}

// generateDockerfileAndResolveArgs generates the Dockerfile, resolves cascade build args,
// and saves the Dockerfile to the staging directory.
// Sets bc.cascadeResolution, bc.dvmDockerfile.
func (bc *buildContext) generateDockerfileAndResolveArgs() error {
	render.Blank()
	render.Progress("Generating Dockerfile.dvm...")
	slog.Debug("generating Dockerfile", "language", bc.languageName, "version", bc.version)

	// Detect private repos and system dependencies
	privateRepoInfo := utils.DetectPrivateRepos(bc.sourcePath, bc.languageName)
	if len(privateRepoInfo.SystemDeps) > 0 {
		render.Info(fmt.Sprintf("Auto-detected system dependencies: %s", strings.Join(privateRepoInfo.SystemDeps, ", ")))
		slog.Debug("auto-detected system dependencies",
			"deps", privateRepoInfo.SystemDeps,
			"sources", privateRepoInfo.SystemDepSources)
	}

	// Pre-compute additional build arg names for Dockerfile ARG declarations
	additionalBuildArgNames := bc.resolveBuildArgNames()

	generator := builders.NewDockerfileGenerator(builders.DockerfileGeneratorOptions{
		Workspace:           bc.workspace,
		WorkspaceSpec:       bc.workspaceYAML.Spec,
		Language:            bc.languageName,
		Version:             bc.version,
		AppPath:             bc.sourcePath,
		BaseDockerfile:      bc.dockerfilePath,
		PathConfig:          paths.New(bc.homeDir),
		PrivateRepoInfo:     privateRepoInfo,
		AdditionalBuildArgs: additionalBuildArgNames,
	})

	if bc.pluginManifest != nil {
		generator.SetPluginManifest(bc.pluginManifest)
	}

	dockerfileContent, err := generator.Generate()
	if err != nil {
		slog.Error("failed to generate Dockerfile", "error", err)
		return fmt.Errorf("failed to generate Dockerfile: %w", err)
	}

	bc.dvmDockerfile, err = builders.SaveDockerfile(dockerfileContent, bc.stagingDir)
	if err != nil {
		slog.Error("failed to save Dockerfile", "error", err)
		return err
	}
	slog.Debug("saved Dockerfile", "path", bc.dvmDockerfile)
	return nil
}

// resolveBuildArgNames resolves hierarchical build args and collects all
// additional build arg names (from registry env vars and the cascade resolver)
// for Dockerfile ARG declarations.
// Sets bc.cascadeResolution; returns the collected names.
func (bc *buildContext) resolveBuildArgNames() []string {
	var names []string
	if bc.registryEnvVars != nil {
		for k := range bc.registryEnvVars {
			names = append(names, k)
		}
	}

	// Resolve hierarchical build args (global < ecosystem < domain < app < workspace)
	buildArgsResolver := resolver.NewHierarchyBuildArgsResolver(bc.ds)
	cascadeResolution, cascadeErr := buildArgsResolver.Resolve(context.Background(), bc.workspace.ID)
	if cascadeErr != nil {
		slog.Warn("failed to resolve hierarchical build args, continuing without them", "error", cascadeErr)
		cascadeResolution = nil
	}
	bc.cascadeResolution = cascadeResolution

	if cascadeResolution != nil {
		for k := range cascadeResolution.Args {
			names = append(names, k)
		}
	}
	return names
}

// buildImage creates the image builder, checks for existing images, assembles
// build args, and executes the container image build.
// Sets bc.imageName, bc.builder. Returns true if build was skipped (image exists).
func (bc *buildContext) buildImage() (skipped bool, err error) {
	// Generate image name with timestamp tag
	timestamp := time.Now().Format("20060102-150405")
	bc.imageName = fmt.Sprintf("dvm-%s-%s:%s", bc.workspaceName, bc.appName, timestamp)
	render.Blank()
	render.Progress(fmt.Sprintf("Building image: %s", bc.imageName))
	slog.Info("building image", "image", bc.imageName, "dockerfile", bc.dvmDockerfile)

	if err := bc.createBuilder(); err != nil {
		return false, err
	}

	// Check if image exists (skip if --force)
	if !buildForce {
		exists, existsErr := bc.builder.ImageExists(bc.ctx)
		if existsErr == nil && exists {
			slog.Debug("image already exists, skipping build", "image", bc.imageName)
			render.Info(fmt.Sprintf("Image already exists: %s", bc.imageName))
			render.Info("Use --force to rebuild")
			return true, nil
		}
	}

	buildArgs := bc.assembleBuildArgs()

	slog.Debug("starting image build", "target", buildTarget, "no_cache", buildNocache)
	if err := bc.builder.Build(bc.ctx, builders.BuildOptions{
		BuildArgs: buildArgs,
		Target:    buildTarget,
		NoCache:   buildNocache,
	}); err != nil {
		slog.Error("build failed", "image", bc.imageName, "error", err)
		return false, err
	}
	slog.Info("build completed", "image", bc.imageName)

	// For Colima/BuildKit, copy image to devopsmaestro namespace
	if bc.platform.IsContainerd() {
		if err := copyImageToNamespace(bc.platform, bc.imageName); err != nil {
			return false, err
		}
	}

	return false, nil
}

// createBuilder creates the image builder, using staging dir as build context
// with a fallback to app path.
// Sets bc.builder.
func (bc *buildContext) createBuilder() error {
	buildCtxDir := bc.stagingDir
	if _, statErr := os.Stat(bc.stagingDir); os.IsNotExist(statErr) {
		buildCtxDir = bc.app.Path
		slog.Warn("staging directory not found, using app path as build context", "staging", bc.stagingDir, "fallback", bc.app.Path)
	}

	var err error
	bc.builder, err = builders.NewImageBuilder(builders.BuilderConfig{
		Platform:   bc.platform,
		Namespace:  "devopsmaestro",
		AppPath:    buildCtxDir,
		ImageName:  bc.imageName,
		Dockerfile: bc.dvmDockerfile,
	})
	if err != nil {
		return fmt.Errorf("failed to create builder: %w", err)
	}
	return nil
}

// assembleBuildArgs layers build args from registry, cascade resolver, and credentials.
// Priority (lowest to highest): registry env vars → cascade → credentials.
func (bc *buildContext) assembleBuildArgs() map[string]string {
	buildArgs := make(map[string]string)

	// Layer 1: Registry env vars (lowest priority)
	if bc.registryEnvVars != nil {
		for k, v := range bc.registryEnvVars {
			buildArgs[k] = v
			slog.Debug("using registry env var", "key", k)
		}
	}

	// Layer 2: Hierarchical cascade build args
	if bc.cascadeResolution != nil {
		for k, v := range bc.cascadeResolution.Args {
			buildArgs[k] = v
			slog.Debug("using cascaded build arg", "key", k, "source", bc.cascadeResolution.Sources[k].String())
		}
	}

	// Layer 3: Credentials from hierarchy (highest priority)
	resolvedCreds, credWarnings := loadBuildCredentials(bc.ds, bc.app, bc.workspace)
	for _, w := range credWarnings {
		render.Warning(w)
	}
	for k, v := range resolvedCreds {
		if envvalidation.IsDangerousEnvVar(k) {
			slog.Warn("blocked dangerous credential build arg", "key", k)
			continue
		}
		if err := envvalidation.ValidateEnvKey(k); err != nil {
			slog.Warn("skipped credential with invalid env key", "key", k, "error", err)
			continue
		}
		buildArgs[k] = v
		slog.Debug("using credential", "key", k)
	}

	return buildArgs
}

// postBuild updates the workspace in the database, pushes to registry if requested,
// and prints the build summary.
func (bc *buildContext) postBuild() {
	// Update workspace image name in database
	bc.workspace.ImageName = bc.imageName
	if err := bc.ds.UpdateWorkspace(bc.workspace); err != nil {
		render.Warning(fmt.Sprintf("Failed to update workspace image name: %v", err))
	}

	// Push to registry if --push flag is set and registry is available
	if buildPush && bc.registryEndpoint != "" {
		bc.pushToRegistry()
	} else if buildPush && bc.registryEndpoint == "" {
		render.Warning("Cannot push: registry is not available")
		render.Info("Start the registry with: dvm registry start")
	}

	render.Blank()
	render.Success("Build complete!")
	render.Info(fmt.Sprintf("Image: %s", bc.imageName))
	render.Info(fmt.Sprintf("Dockerfile: %s", bc.dvmDockerfile))
	if bc.registryEndpoint != "" {
		render.Info(fmt.Sprintf("Registry cache: %s", bc.registryEndpoint))
	}
	render.Blank()
	render.Info("Next: Attach to your workspace with: dvm attach")
}

// pushToRegistry tags and pushes the built image to the registry.
func (bc *buildContext) pushToRegistry() {
	render.Blank()
	render.Progress(fmt.Sprintf("Pushing image to registry: %s", bc.registryEndpoint))

	registryImage := fmt.Sprintf("%s/%s", bc.registryEndpoint, bc.imageName)
	if err := tagImageForRegistry(bc.platform, bc.imageName, registryImage); err != nil {
		render.Warning(fmt.Sprintf("Failed to tag image for registry: %v", err))
		render.Info("Skipping push to registry")
		return
	}

	if err := pushImageToRegistry(bc.platform, registryImage); err != nil {
		render.Warning(fmt.Sprintf("Failed to push image to registry: %v", err))
	} else {
		render.Success(fmt.Sprintf("Pushed to registry: %s", registryImage))
		slog.Info("pushed image to registry", "image", registryImage)
	}
}
