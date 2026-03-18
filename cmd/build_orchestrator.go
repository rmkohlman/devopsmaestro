package cmd

import (
	"context"
	"devopsmaestro/builders"
	"devopsmaestro/config"
	"devopsmaestro/models"
	"devopsmaestro/pkg/buildargs/resolver"
	"devopsmaestro/pkg/envvalidation"
	"devopsmaestro/pkg/nvimops/plugin"
	"devopsmaestro/pkg/paths"
	"devopsmaestro/pkg/registry"
	"devopsmaestro/pkg/registry/envinjector"
	wsresolver "devopsmaestro/pkg/resolver"
	"devopsmaestro/render"
	"devopsmaestro/utils"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

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

		wsResolver := wsresolver.NewWorkspaceResolver(sqlDS)
		result, err := wsResolver.Resolve(buildFlags.ToFilter())
		if err != nil {
			// Check if ambiguous and provide helpful output
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
		// Fall back to existing context-based behavior (DB-backed)
		appName, err = getActiveAppFromContext(sqlDS)
		if err != nil {
			slog.Debug("no active app set")
			render.Info("Hint: Set active app with: dvm use app <name>")
			render.Info("      Or use flags: dvm build -a <app>")
			return fmt.Errorf("no active app set. Use 'dvm use app <name>' first")
		}

		workspaceName, err = getActiveWorkspaceFromContext(sqlDS)
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
	render.Blank()
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

	// Step 1.5: Ensure registries are running if enabled
	var registryEndpoint string
	var registryEnvVars map[string]string
	ctx := context.Background()
	if config.IsRegistryEnabled() {
		coordinator := registry.NewBuildRegistryCoordinator(
			sqlDS,
			registry.NewServiceFactory(),
			envinjector.NewEnvironmentInjector(),
		)
		regResult, regErr := coordinator.Prepare(ctx)
		if regErr != nil {
			render.Warning(fmt.Sprintf("Registry preparation failed: %v", regErr))
			render.Info("Continuing build without registry cache")
			slog.Warn("registry preparation failed", "error", regErr)
		} else {
			registryEndpoint = regResult.OCIEndpoint
			registryEnvVars = regResult.EnvVars
			for _, w := range regResult.Warnings {
				render.Warning(w)
			}
			if len(regResult.Managers) > 0 {
				render.Info(fmt.Sprintf("Started %d registry cache(s)", len(regResult.Managers)))
			}
		}
	}

	// Step 2: Check for existing Dockerfile
	render.Blank()
	render.Progress("Checking for Dockerfile...")
	hasDockerfile, dockerfilePath := utils.HasDockerfile(app.Path)
	if hasDockerfile {
		render.Info(fmt.Sprintf("Found: %s", dockerfilePath))
		slog.Debug("found existing Dockerfile", "path", dockerfilePath)
	} else {
		render.Info("No Dockerfile found, will generate from scratch")
		slog.Debug("no Dockerfile found, will generate")
	}

	// Step 3: Generate workspace spec (for now, use defaults)
	// Resolve GitRepo name if GitRepoID is set
	gitRepoName := ""
	if workspace.GitRepoID.Valid {
		gitRepo, err := sqlDS.GetGitRepoByID(workspace.GitRepoID.Int64)
		if err == nil && gitRepo != nil {
			gitRepoName = gitRepo.Name
		}
	}
	workspaceYAML := workspace.ToYAML(appName, gitRepoName)

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

	// Step 4: Prepare staging directory and generate shell config (ALWAYS)
	// This must happen before Dockerfile generation so configs are available
	// Get the correct source path (workspace repo path if GitRepoID set, else app.Path)
	sourcePath, err := getBuildSourcePath(sqlDS, workspace, app.Path)
	if err != nil {
		return fmt.Errorf("failed to determine build source path: %w", err)
	}

	// Step 4b: Detect language (use App.Language if set, fall back to auto-detection)
	// This must happen AFTER sourcePath is computed so detection uses the worktree
	// checkout (sourcePath) instead of the bare git mirror (app.Path).
	render.Blank()
	render.Progress("Detecting app language...")
	languageName, version, wasDetected := getLanguageFromApp(app, sourcePath)

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

	stagingDir := paths.New(homeDir).BuildStagingDir(filepath.Base(sourcePath))
	if err := prepareStagingDirectory(stagingDir, sourcePath, appName, workspaceName, sqlDS, workspace); err != nil {
		return err
	}

	// Step 4.5: Prepare CA certificates if configured
	if len(workspaceYAML.Spec.Build.CACerts) > 0 {
		render.Progress("Resolving CA certificates from vault...")
		if err := prepareCACerts(stagingDir, workspaceYAML.Spec.Build.CACerts); err != nil {
			return err
		}
		render.Info(fmt.Sprintf("Injecting %d CA certificate(s) into build context", len(workspaceYAML.Spec.Build.CACerts)))
	}

	// Step 5b: Generate nvim config BEFORE Dockerfile (so Dockerfile generator can see .config/nvim/)
	var pluginManifest *plugin.PluginManifest
	if workspaceYAML.Spec.Nvim.Structure != "" && workspaceYAML.Spec.Nvim.Structure != "none" {
		manifest, err := generateNvimConfig(workspaceYAML.Spec.Nvim.Plugins, stagingDir, homeDir, sqlDS, app, workspace, appName, workspaceName, languageName)
		if err != nil {
			return err
		}
		pluginManifest = manifest
	}

	// Step 6: Generate Dockerfile (after nvim config so it can detect .config/nvim/)
	render.Blank()
	render.Progress("Generating Dockerfile.dvm...")
	slog.Debug("generating Dockerfile", "language", languageName, "version", version)

	// Detect private repos and system dependencies
	privateRepoInfo := utils.DetectPrivateRepos(sourcePath, languageName)

	// Log auto-detected system dependencies for visibility
	if len(privateRepoInfo.SystemDeps) > 0 {
		render.Info(fmt.Sprintf("Auto-detected system dependencies: %s", strings.Join(privateRepoInfo.SystemDeps, ", ")))
		slog.Debug("auto-detected system dependencies",
			"deps", privateRepoInfo.SystemDeps,
			"sources", privateRepoInfo.SystemDepSources)
	}

	// Pre-compute additional build arg names from registry env vars and the cascade resolver.
	// The Dockerfile generator needs these names (not values) to emit ARG declarations.
	// Deduplication is handled by the generator's collectAdditionalBuildArgs().
	var additionalBuildArgNames []string
	if registryEnvVars != nil {
		for k := range registryEnvVars {
			additionalBuildArgNames = append(additionalBuildArgNames, k)
		}
	}

	// Resolve hierarchical build args (global < ecosystem < domain < app < workspace)
	buildArgsResolver := resolver.NewHierarchyBuildArgsResolver(sqlDS)
	cascadeResolution, cascadeErr := buildArgsResolver.Resolve(context.Background(), workspace.ID)
	if cascadeErr != nil {
		slog.Warn("failed to resolve hierarchical build args, continuing without them", "error", cascadeErr)
		cascadeResolution = nil
	}
	if cascadeResolution != nil {
		for k := range cascadeResolution.Args {
			additionalBuildArgNames = append(additionalBuildArgNames, k)
		}
	}

	generator := builders.NewDockerfileGenerator(builders.DockerfileGeneratorOptions{
		Workspace:           workspace,
		WorkspaceSpec:       workspaceYAML.Spec,
		Language:            languageName,
		Version:             version,
		AppPath:             sourcePath, // Use sourcePath (not app.Path) so nvim config staging dir is found correctly (Issue #18)
		BaseDockerfile:      dockerfilePath,
		PathConfig:          paths.New(homeDir),
		PrivateRepoInfo:     privateRepoInfo,
		AdditionalBuildArgs: additionalBuildArgNames,
	})

	// Set plugin manifest for conditional feature detection
	if pluginManifest != nil {
		generator.SetPluginManifest(pluginManifest)
	}

	dockerfileContent, err := generator.Generate()
	if err != nil {
		slog.Error("failed to generate Dockerfile", "error", err)
		return fmt.Errorf("failed to generate Dockerfile: %w", err)
	}

	// Save Dockerfile to STAGING directory (not app directory)
	// This ensures the Dockerfile is in the same directory as the build context
	// so COPY commands can find .config/starship.toml and other generated files
	dvmDockerfile, err := builders.SaveDockerfile(dockerfileContent, stagingDir)
	if err != nil {
		slog.Error("failed to save Dockerfile", "error", err)
		return err
	}
	slog.Debug("saved Dockerfile", "path", dvmDockerfile)

	// Step 6: Build image
	// Use timestamp tag for versioning (enables container recreation on rebuild)
	timestamp := time.Now().Format("20060102-150405")
	imageName := fmt.Sprintf("dvm-%s-%s:%s", workspaceName, appName, timestamp)
	render.Blank()
	render.Progress(fmt.Sprintf("Building image: %s", imageName))
	slog.Info("building image", "image", imageName, "dockerfile", dvmDockerfile)

	// Create image builder using the factory (decoupled from platform specifics)
	// Use staging directory as build context (contains app source + generated configs)
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
	// Priority (lowest to highest): registry env vars -> cascade resolver (global < eco < domain < app < workspace) -> credentials
	buildArgs := make(map[string]string)

	// Layer 1: Registry env vars (lowest priority - can be overridden by cascade resolver)
	if registryEnvVars != nil {
		for k, v := range registryEnvVars {
			buildArgs[k] = v
			slog.Debug("using registry env var", "key", k)
		}
	}

	// Layer 2: Hierarchical cascade build args (global < ecosystem < domain < app < workspace)
	// cascadeResolution was already computed above for Dockerfile ARG generation.
	if cascadeResolution != nil {
		for k, v := range cascadeResolution.Args {
			buildArgs[k] = v
			slog.Debug("using cascaded build arg", "key", k, "source", cascadeResolution.Sources[k].String())
		}
	}

	// Layer 3: Credentials from hierarchy (highest priority)
	resolvedCreds, credWarnings := loadBuildCredentials(sqlDS, app, workspace)
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

	// Step 8: Push to registry if --push flag is set and registry is available
	if buildPush && registryEndpoint != "" {
		render.Blank()
		render.Progress(fmt.Sprintf("Pushing image to registry: %s", registryEndpoint))

		// Tag image for registry
		registryImage := fmt.Sprintf("%s/%s", registryEndpoint, imageName)
		if err := tagImageForRegistry(platform, imageName, registryImage); err != nil {
			render.Warning(fmt.Sprintf("Failed to tag image for registry: %v", err))
			render.Info("Skipping push to registry")
		} else {
			// Push image to registry
			if err := pushImageToRegistry(platform, registryImage); err != nil {
				render.Warning(fmt.Sprintf("Failed to push image to registry: %v", err))
			} else {
				render.Success(fmt.Sprintf("Pushed to registry: %s", registryImage))
				slog.Info("pushed image to registry", "image", registryImage)
			}
		}
	} else if buildPush && registryEndpoint == "" {
		render.Warning("Cannot push: registry is not available")
		render.Info("Start the registry with: dvm registry start")
	}

	render.Blank()
	render.Success("Build complete!")
	render.Info(fmt.Sprintf("Image: %s", imageName))
	render.Info(fmt.Sprintf("Dockerfile: %s", dvmDockerfile))
	if registryEndpoint != "" {
		render.Info(fmt.Sprintf("Registry cache: %s", registryEndpoint))
	}
	render.Blank()
	render.Info("Next: Attach to your workspace with: dvm attach")

	return nil
}

// prepareCACerts resolves CA certificates from MaestroVault and writes them
// to the staging directory's certs/ subdirectory. The Dockerfile generator
// will COPY these into the image and update the system trust store.
// All errors are fatal — a missing or invalid cert should fail the build.
func prepareCACerts(stagingDir string, caCerts []models.CACertConfig) error {
	// Validate cert configs
	if err := models.ValidateCACerts(caCerts); err != nil {
		return fmt.Errorf("invalid CA certificate configuration: %w", err)
	}

	// Resolve vault token (fatal if missing — certs require vault)
	token, tokenErr := config.ResolveVaultToken()
	if tokenErr != nil || token == "" {
		return fmt.Errorf("CA certificates require MaestroVault but no vault token is configured. Hint: Configure vault with: dvm admin vault configure")
	}
	if err := config.EnsureVaultDaemon(); err != nil {
		return fmt.Errorf("failed to start vault daemon for CA cert resolution: %w", err)
	}
	vb, err := config.NewVaultBackend(token)
	if err != nil {
		return fmt.Errorf("failed to create vault backend for CA cert resolution: %w", err)
	}

	// Store as interface so type assertions work for FieldCapableBackend
	var backend config.SecretBackend = vb

	// Create certs directory in staging
	certsDir := filepath.Join(stagingDir, "certs")
	if err := os.MkdirAll(certsDir, 0755); err != nil {
		return fmt.Errorf("failed to create certs directory: %w", err)
	}

	// Resolve each cert from vault and write to staging
	for _, cert := range caCerts {
		var pemContent string
		var err error

		// Check if this is a field-level request
		if cert.VaultField != "" {
			// Use GetField for field-level access
			if fb, ok := backend.(config.FieldCapableBackend); ok {
				pemContent, err = fb.GetField(cert.VaultSecret, cert.VaultEnvironment, cert.VaultField)
			} else {
				return fmt.Errorf("vault backend does not support field-level access for cert %q", cert.Name)
			}
		} else {
			pemContent, err = backend.Get(cert.VaultSecret, cert.VaultEnvironment)
		}
		if err != nil {
			return fmt.Errorf("failed to resolve CA certificate %q from vault: %w", cert.Name, err)
		}

		// Validate PEM content
		if err := models.ValidatePEMContent(pemContent); err != nil {
			return fmt.Errorf("CA certificate %q has invalid content: %w", cert.Name, err)
		}

		// Path traversal defense: ensure name is just a filename
		safeName := filepath.Base(cert.Name)
		if safeName != cert.Name {
			return fmt.Errorf("CA certificate name %q contains path separators", cert.Name)
		}

		certPath := filepath.Join(certsDir, safeName+".crt")

		// Verify the resolved path is within certsDir (defense in depth)
		cleanPath := filepath.Clean(certPath)
		if !strings.HasPrefix(cleanPath, filepath.Clean(certsDir)) {
			return fmt.Errorf("CA certificate path %q escapes certs directory", cert.Name)
		}

		if err := os.WriteFile(certPath, []byte(pemContent), 0644); err != nil {
			return fmt.Errorf("failed to write CA certificate %q: %w", cert.Name, err)
		}

		slog.Debug("wrote CA certificate", "name", cert.Name, "path", certPath)
	}

	return nil
}
