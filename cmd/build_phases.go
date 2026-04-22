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
	"devopsmaestro/operators"
	"devopsmaestro/pkg/buildargs/resolver"
	cacertsresolver "devopsmaestro/pkg/cacerts/resolver"
	"devopsmaestro/pkg/envvalidation"
	"devopsmaestro/pkg/registry"
	"devopsmaestro/pkg/registry/envinjector"
	wsresolver "devopsmaestro/pkg/resolver"
	"devopsmaestro/utils"
	"devopsmaestro/utils/appkind"

	"github.com/google/uuid"
	"github.com/rmkohlman/MaestroSDK/paths"
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
		"domain", buildFlags.Domain, "system", buildFlags.System, "app", buildFlags.App, "workspace", buildFlags.Workspace)

	wsResolver := wsresolver.NewWorkspaceResolver(bc.ds)
	result, err := wsResolver.Resolve(buildFlags.ToFilter())
	if err != nil {
		if ambiguousErr, ok := wsresolver.IsAmbiguousError(err); ok {
			bc.renderWarning("Multiple workspaces match your criteria")
			bc.renderPlain(ambiguousErr.FormatDisambiguation())
			bc.renderPlain(FormatSuggestions(SuggestAmbiguousWorkspace()...))
			return fmt.Errorf("ambiguous workspace selection")
		}
		if wsresolver.IsNoWorkspaceFoundError(err) {
			bc.renderWarning("No workspace found matching your criteria")
			bc.renderPlain(FormatSuggestions(SuggestWorkspaceNotFound("")...))
			return err
		}
		return fmt.Errorf("failed to resolve workspace: %w", err)
	}

	bc.workspace = result.Workspace
	bc.app = result.App
	bc.appName = bc.app.Name
	bc.workspaceName = bc.workspace.Name

	bc.renderInfof("Resolved: %s", result.FullPath())
	return nil
}

// resolveFromActiveContext falls back to DB-backed active app/workspace context.
func (bc *buildContext) resolveFromActiveContext() error {
	var err error
	bc.appName, err = getActiveAppFromContext(bc.ds)
	if err != nil {
		slog.Debug("no active app set")
		bc.renderPlain(FormatSuggestions(SuggestNoActiveApp()...))
		return fmt.Errorf("no active app set. Use 'dvm use app <name>' first")
	}

	bc.workspaceName, err = getActiveWorkspaceFromContext(bc.ds)
	if err != nil {
		slog.Debug("no active workspace set")
		bc.renderPlain(FormatSuggestions(SuggestNoActiveWorkspace()...))
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

// buildKey returns a unique key for this workspace build, used for staging
// directories and build cache paths. Prefers the workspace slug (which encodes
// the full hierarchy: ecosystem-domain-system-app-workspace) for uniqueness.
// Falls back to appName-workspaceName for workspaces created before slug
// generation was introduced.
func (bc *buildContext) buildKey() string {
	if bc.workspace != nil && bc.workspace.Slug != "" {
		return bc.workspace.Slug
	}
	return bc.appName + "-" + bc.workspaceName
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
	bc.renderProgress("Detecting container platform...")
	platform, err := detectPlatform()
	if err != nil {
		return err
	}
	bc.platform = platform
	bc.renderInfof("Platform: %s", platform.Name)
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
		bc.renderWarningf("Registry preparation failed: %v", regErr)
		bc.renderInfo("Continuing build without registry cache")
		slog.Warn("registry preparation failed", "error", regErr)
		return nil
	}

	bc.registryEndpoint = regResult.OCIEndpoint
	bc.registryEnvVars = regResult.EnvVars
	bc.cacheReadiness = &regResult.CacheReadiness
	bc.buildKitConfigPath = regResult.BuildKitConfigPath
	bc.containerdCertsDir = regResult.ContainerdCertsDir
	for _, w := range regResult.Warnings {
		// Render registry warnings as errors for visual prominence —
		// these indicate missing or broken cache proxies that degrade
		// build performance. (Still non-fatal, but the user should see them.)
		bc.renderError(w)
	}
	if len(regResult.Managers) > 0 {
		bc.renderInfof("Started %d registry cache(s)", len(regResult.Managers))
	}
	if regResult.BuildKitConfigPath != "" {
		bc.renderInfo("BuildKit registry mirrors configured")
		slog.Info("buildkit mirror config", "path", regResult.BuildKitConfigPath)
	}
	return nil
}

// checkDockerfile looks for an existing Dockerfile in the app directory.
// Sets bc.hasDockerfile and bc.dockerfilePath.
func (bc *buildContext) checkDockerfile() {
	bc.renderBlank()
	bc.renderProgress("Checking for Dockerfile...")
	bc.hasDockerfile, bc.dockerfilePath = utils.HasDockerfile(bc.app.Path)
	if bc.hasDockerfile {
		bc.renderInfof("Found: %s", bc.dockerfilePath)
		slog.Debug("found existing Dockerfile", "path", bc.dockerfilePath)
	} else {
		bc.renderInfo("No Dockerfile found, will generate from scratch")
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

	// Detect AppKind first (#404). For CICD apps we skip language detection
	// entirely and use the alpine + kubectl/helm/kustomize build path.
	kind, evidence, _ := appkind.Detect(bc.sourcePath, bc.app, bc.app.GetKind())
	bc.appKind = string(kind)
	for _, s := range evidence.Signals {
		if strings.HasPrefix(s, "signal3:") {
			bc.argoCDDetected = true
			break
		}
	}

	if kind == appkind.KindCICD {
		bc.renderBlank()
		bc.renderProgress("Detected CICD app (YAML/Helm/Kustomize/Argo/Flux) — skipping language detection")
	} else {
		// Detect language (use App.Language if set, fall back to auto-detection)
		bc.renderBlank()
		bc.renderProgress("Detecting app language...")
		bc.languageName, bc.version, _ = bc.detectLanguageAndReport()
	}

	// Use the workspace slug (ecosystem-domain-system-app-workspace) plus a
	// per-invocation UUID to ensure each concurrent build gets its own
	// isolated staging directory.  The slug encodes the full hierarchy
	// (preventing collisions across different apps/domains — issue #227),
	// and the UUID suffix prevents races when two processes build overlapping
	// workspaces simultaneously (issue #256).
	stagingKey := bc.buildKey() + "-" + uuid.New().String()[:8]
	bc.stagingDir = paths.New(bc.homeDir).BuildStagingDir(stagingKey)
	if err := prepareStagingDirectory(bc.stagingDir, bc.sourcePath, bc.appName, bc.workspaceName, bc.ds, bc.workspace, bc.out()); err != nil {
		return err
	}
	return nil
}

// detectLanguageAndReport detects the app language and emits render messages.
// Returns languageName, version, and whether it was auto-detected.
func (bc *buildContext) detectLanguageAndReport() (string, string, bool) {
	languageName, version, wasDetected := getLanguageFromApp(bc.app, bc.sourcePath)

	if !wasDetected {
		bc.renderInfof("Language: %s (from app config)", languageName)
		if version != "" {
			bc.renderInfof("Version: %s", version)
		}
		slog.Debug("using language from app config", "language", languageName, "version", version)
	} else if languageName != "unknown" {
		if version != "" {
			bc.renderInfof("Language: %s (version: %s)", languageName, version)
		} else {
			bc.renderInfof("Language: %s", languageName)
		}
		slog.Debug("detected language", "language", languageName, "version", version)
	} else {
		bc.renderInfo("Language: Unknown (will use generic base)")
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
		bc.renderProgress("Resolving CA certificates from vault...")
		if err := prepareCACerts(bc.stagingDir, bc.workspaceYAML.Spec.Build.CACerts); err != nil {
			return err
		}
		bc.renderInfof("Injecting %d CA certificate(s) into build context", len(bc.workspaceYAML.Spec.Build.CACerts))
	}
	return nil
}

// generateNvimConfiguration generates nvim config if a structure is configured.
// Sets bc.pluginManifest.
// Before generating config, it auto-syncs embedded libraries to the DB
// if the library fingerprint has changed (issue #255).
func (bc *buildContext) generateNvimConfiguration() error {
	if bc.workspaceYAML.Spec.Nvim.Structure == "" || bc.workspaceYAML.Spec.Nvim.Structure == "none" {
		return nil
	}

	// Auto-sync embedded libraries to DB if fingerprint changed (#255).
	// This ensures builds always use the latest embedded plugin definitions.
	if err := EnsureLibrarySynced(bc.ds); err != nil {
		slog.Warn("library auto-sync failed, continuing with existing DB data", "error", err)
	}

	manifest, err := generateNvimConfig(
		bc.workspaceYAML.Spec.Nvim.Plugins, bc.stagingDir, bc.homeDir, bc.ds,
		bc.app, bc.workspace, bc.appName, bc.workspaceName, bc.languageName, bc.out(),
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
	bc.renderBlank()
	bc.renderProgress("Generating Dockerfile.dvm...")
	slog.Debug("generating Dockerfile", "language", bc.languageName, "version", bc.version)

	// Detect private repos and system dependencies
	privateRepoInfo := utils.DetectPrivateRepos(bc.sourcePath, bc.languageName)
	if len(privateRepoInfo.SystemDeps) > 0 {
		bc.renderInfof("Auto-detected system dependencies: %s", strings.Join(privateRepoInfo.SystemDeps, ", "))
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
		StagingDir:          bc.stagingDir,
		BaseDockerfile:      bc.dockerfilePath,
		PathConfig:          paths.New(bc.homeDir),
		PrivateRepoInfo:     privateRepoInfo,
		AdditionalBuildArgs: additionalBuildArgNames,
		AppKind:             bc.appKind,
		ArgoCDDetected:      bc.argoCDDetected,
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

// validateStagingDirectory parses the generated Dockerfile for COPY commands and
// verifies that every referenced source file/directory exists in the staging directory.
// This catches mismatches between Dockerfile generation and staging preparation
// before Docker even starts building (see #228).
func (bc *buildContext) validateStagingDirectory() error {
	if bc.dvmDockerfile == "" || bc.stagingDir == "" {
		return nil
	}

	content, err := os.ReadFile(bc.dvmDockerfile)
	if err != nil {
		slog.Warn("cannot read generated Dockerfile for validation", "error", err)
		return nil // Non-fatal: validation is best-effort
	}

	var missing []string
	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		// Skip comments and non-COPY lines
		if strings.HasPrefix(trimmed, "#") || !strings.HasPrefix(trimmed, "COPY") {
			continue
		}
		// Skip COPY --from= (multi-stage copies from builder stages)
		if strings.Contains(trimmed, "--from=") {
			continue
		}
		// Skip COPY heredocs (e.g., COPY <<'EOF' /path)
		if strings.Contains(trimmed, "<<") {
			continue
		}
		// Skip COPY --chown with heredoc
		if strings.Contains(trimmed, "--chown=") && strings.Contains(trimmed, "<<") {
			continue
		}

		// Parse: COPY [--chown=...] <src> <dst>
		parts := strings.Fields(trimmed)
		if len(parts) < 3 {
			continue
		}
		// Skip the COPY keyword and any flags (--chown, etc.)
		srcIdx := 1
		for srcIdx < len(parts)-1 && strings.HasPrefix(parts[srcIdx], "--") {
			srcIdx++
		}
		if srcIdx >= len(parts)-1 {
			continue
		}
		src := parts[srcIdx]

		// Check if the source exists in the staging directory
		fullPath := filepath.Join(bc.stagingDir, src)
		if _, statErr := os.Stat(fullPath); os.IsNotExist(statErr) {
			missing = append(missing, src)
		}
	}

	if len(missing) > 0 {
		slog.Warn("staging directory validation: missing files referenced by COPY",
			"staging_dir", bc.stagingDir,
			"missing", missing)
		bc.renderWarningf("Staging directory is missing %d file(s) referenced by Dockerfile COPY commands: %s",
			len(missing), strings.Join(missing, ", "))
		bc.renderInfo("This may cause Docker build failures. Check that all required files are generated.")
	}

	return nil
}

// buildImage creates the image builder, checks for existing images, assembles
// build args, and executes the container image build.
// Sets bc.imageName, bc.builder. Returns true if build was skipped (image exists).
func (bc *buildContext) buildImage() (skipped bool, err error) {
	// Generate image name with timestamp tag
	timestamp := time.Now().Format("20060102-150405")
	bc.imageName = fmt.Sprintf("dvm-%s-%s:%s", bc.workspaceName, bc.appName, timestamp)
	bc.renderBlank()
	bc.renderProgressf("Building image: %s", bc.imageName)
	slog.Info("building image", "image", bc.imageName, "dockerfile", bc.dvmDockerfile)

	if err := bc.createBuilder(); err != nil {
		return false, err
	}

	// Check if image exists (skip if --force)
	if !buildForce {
		exists, existsErr := bc.builder.ImageExists(bc.ctx)
		if existsErr == nil && exists {
			slog.Debug("image already exists, skipping build", "image", bc.imageName)
			bc.renderInfof("Image already exists: %s", bc.imageName)
			bc.renderInfo("Use --force to rebuild")
			return true, nil
		}
	}

	buildArgs := bc.assembleBuildArgs()

	slog.Debug("starting image build", "target", buildTarget, "no_cache", buildNocache)

	// Local directory build cache (type=local) for BuildKit.
	// Uses ~/.devopsmaestro/build-cache/<app>-<workspace>/ to persist layers
	// across builds, surviving docker system prune. This sidesteps the HTTP/HTTPS
	// mismatch that blocked registry-based caching via Zot (see #225).
	buildOpts := builders.BuildOptions{
		BuildArgs:          buildArgs,
		Target:             buildTarget,
		NoCache:            buildNocache,
		Timeout:            buildTimeout,
		Output:             bc.output,
		BuildKitConfigPath: bc.buildKitConfigPath,
		RegistryMirrorsDir: bc.containerdCertsDir,
	}

	if !buildNocache {
		pc := paths.New(bc.homeDir)
		cacheKey := bc.buildKey()
		cacheDir := filepath.Join(pc.Root(), "build-cache", cacheKey)

		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			slog.Warn("failed to create build cache directory, building without cache",
				"path", cacheDir, "error", err)
		} else {
			buildOpts.CacheFrom = fmt.Sprintf("type=local,src=%s", cacheDir)
			buildOpts.CacheTo = fmt.Sprintf("type=local,dest=%s,mode=max", cacheDir)
			slog.Info("build cache enabled", "cache_dir", cacheDir)
		}

		// Registry-based cache: use local zot registry for Docker layer
		// caching when available (#383). This keeps layers in the registry
		// rather than in BuildKit/Colima, minimizing disk footprint.
		if bc.registryEndpoint != "" {
			// Strip scheme — BuildKit registry refs use host:port, not URLs (#384).
			regHost := registry.EndpointFromURL(bc.registryEndpoint)
			regCacheRef := fmt.Sprintf("%s/dvm-cache/%s-%s:buildcache",
				regHost, bc.workspaceName, bc.appName)
			regCacheFrom := fmt.Sprintf("type=registry,ref=%s,registry.insecure=true", regCacheRef)
			regCacheTo := fmt.Sprintf("type=registry,ref=%s,mode=max,registry.insecure=true", regCacheRef)

			if buildOpts.CacheFrom != "" {
				// Combine local + registry cache sources
				buildOpts.CacheFrom = buildOpts.CacheFrom + "\n" + regCacheFrom
			} else {
				buildOpts.CacheFrom = regCacheFrom
			}
			buildOpts.CacheTo = regCacheTo // Prefer pushing to registry
			slog.Info("registry cache enabled", "ref", regCacheRef)
			bc.renderInfof("Registry layer cache: %s", regCacheRef)
		}
	} else {
		slog.Info("build cache disabled (--no-cache)")
	}

	// Pre-build cache cleanup if --clean-cache flag is set (#378, #383)
	if buildCleanCache {
		bc.renderProgress("Pruning BuildKit cache before build (--clean-cache)...")
		if pruneErr := bc.pruneBuildKitCache(); pruneErr != nil {
			slog.Warn("pre-build cache prune failed (continuing anyway)", "error", pruneErr)
			bc.renderWarning("Cache prune failed (continuing with build)")
		} else {
			bc.renderSuccess("BuildKit cache pruned successfully")
		}

		// Extended cleanup: remove old workspace images, prune dangling,
		// and verify registry health (#383).
		bc.preCleanCacheCleanup()
		bc.renderBlank()
	}

	if err := bc.builder.Build(bc.ctx, buildOpts); err != nil {
		// Check for BuildKit RPC connection errors (#385). These indicate the
		// BuildKit daemon crashed or dropped connections under load. Retry with
		// exponential backoff before giving up.
		if builders.IsBuildKitConnectionError(err) {
			slog.Warn("BuildKit connection error, retrying with backoff", "error", err)
			bc.renderWarning("BuildKit connection lost — retrying build...")
			retried := false
			for attempt := 1; attempt <= 2; attempt++ {
				backoff := time.Duration(attempt*5) * time.Second
				slog.Info("waiting before retry", "attempt", attempt, "backoff", backoff)
				select {
				case <-bc.ctx.Done():
					return false, bc.ctx.Err()
				case <-time.After(backoff):
				}
				bc.renderInfof("Retry attempt %d/2...", attempt)
				if retryErr := bc.builder.Build(bc.ctx, buildOpts); retryErr != nil {
					if builders.IsBuildKitConnectionError(retryErr) {
						slog.Warn("BuildKit connection error on retry", "attempt", attempt, "error", retryErr)
						continue
					}
					slog.Error("build failed on retry", "attempt", attempt, "error", retryErr)
					return false, builders.EnhanceBuildError(retryErr)
				}
				retried = true
				slog.Info("build succeeded on retry", "image", bc.imageName, "attempt", attempt)
				break
			}
			if !retried {
				slog.Error("build failed after all retries", "image", bc.imageName, "error", err)
				return false, builders.EnhanceBuildError(err)
			}
			goto buildSuccess
		}

		// Check for BuildKit cache corruption (#378). If detected, attempt
		// automatic recovery: prune the BuildKit cache and retry once.
		if builders.IsCacheCorruption(err) {
			bc.renderWarning("BuildKit cache corruption detected — attempting automatic recovery...")
			slog.Warn("cache corruption detected, attempting prune and retry", "error", err)
			if pruneErr := bc.pruneBuildKitCache(); pruneErr != nil {
				slog.Warn("automatic cache prune failed", "error", pruneErr)
				bc.renderWarning("Automatic cache cleanup failed. Try: dvm cache clear --buildkit --force")
			} else {
				bc.renderInfo("Cache pruned successfully, retrying build...")
				// Retry the build once after pruning
				if retryErr := bc.builder.Build(bc.ctx, buildOpts); retryErr != nil {
					slog.Error("build failed after cache prune", "image", bc.imageName, "error", retryErr)
					return false, builders.EnhanceBuildError(retryErr)
				}
				slog.Info("build succeeded after cache prune", "image", bc.imageName)
				goto buildSuccess
			}
		}

		// Defense-in-depth: the builder reported failure, but the image may
		// have been built successfully (e.g., Docker buildx on Colima exits
		// non-zero after completing image export). Verify the image exists
		// before propagating the error.
		if exists, existsErr := bc.builder.ImageExists(bc.ctx); existsErr == nil && exists {
			slog.Warn("builder returned error but image exists, treating as success",
				"image", bc.imageName, "original_error", err)
			bc.renderBlank()
			bc.renderSuccessf("Image built successfully (recovered): %s", bc.imageName)
		} else {
			slog.Error("build failed", "image", bc.imageName, "error", err)
			return false, err
		}
	} else {
		slog.Info("build completed", "image", bc.imageName)
	}

buildSuccess:
	if bc.platform.IsContainerd() {
		if err := copyImageToNamespace(bc.platform, bc.imageName, bc.out()); err != nil {
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

// pruneBuildKitCache attempts to prune the BuildKit cache via the builder's API.
// Falls back to CLI-based prune if the builder doesn't support direct pruning.
func (bc *buildContext) pruneBuildKitCache() error {
	if bkBuilder, ok := bc.builder.(*builders.BuildKitBuilder); ok {
		return bkBuilder.PruneBuildKitCache(bc.ctx)
	}
	// Fallback: use system cleaner for Docker-based platforms
	if bc.platform != nil {
		cleaner := operators.NewSystemCleaner(bc.platform)
		_, err := cleaner.PruneBuildKit(bc.ctx, false)
		return err
	}
	return fmt.Errorf("no builder available for cache prune")
}

// pruneBuildKitCacheLight prunes only unused BuildKit cache entries (#384).
// This is safe to call after every build — it preserves recently-used layers.
func (bc *buildContext) pruneBuildKitCacheLight() error {
	if bkBuilder, ok := bc.builder.(*builders.BuildKitBuilder); ok {
		return bkBuilder.PruneBuildKitCacheLight(bc.ctx)
	}
	// For Docker-based platforms, the light prune is the same as regular prune
	// (docker buildx prune doesn't have a "light" mode).
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
		bc.renderWarning(w)
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
	// Update workspace image name in database.
	// Use the targeted UpdateWorkspaceImage first (only updates image_name column),
	// then fall back to the full UpdateWorkspace. This ensures the image tag is
	// persisted even if the full row update fails due to concurrent access (#367).
	bc.workspace.ImageName = bc.imageName
	if err := bc.ds.UpdateWorkspaceImage(bc.workspace.ID, bc.imageName); err != nil {
		slog.Warn("targeted workspace image update failed, trying full update",
			"workspace_id", bc.workspace.ID, "image", bc.imageName, "error", err)
	}
	if err := bc.ds.UpdateWorkspace(bc.workspace); err != nil {
		bc.renderWarningf("Failed to update workspace: %v", err)
		slog.Warn("full workspace update failed",
			"workspace_id", bc.workspace.ID, "image", bc.imageName, "error", err)
	}

	// Push to registry if --push flag is set and registry is available
	if buildPush && bc.registryEndpoint != "" {
		bc.pushToRegistry()
	} else if buildPush && bc.registryEndpoint == "" {
		bc.renderWarning("Cannot push: registry is not available")
		bc.renderInfo("Start the registry with: dvm registry start")
	}

	// Prune old images for this workspace (auto-cleanup after successful build).
	bc.pruneOldImages()

	// Light BuildKit cache prune after every build to reclaim orphaned layers
	// without removing recently-used cache (#384).
	if !buildCleanCache {
		if pruneErr := bc.pruneBuildKitCacheLight(); pruneErr != nil {
			slog.Warn("post-build light cache prune failed (non-fatal)", "error", pruneErr)
		}
	}

	// Post-build cleanup when --clean-cache: prune dangling images and
	// BuildKit cache to minimize Colima footprint (#383).
	if buildCleanCache {
		bc.postCleanCacheCleanup()
	}

	bc.renderBlank()
	bc.renderSuccess("Build complete!")
	bc.renderInfof("Image: %s", bc.imageName)
	bc.renderInfof("Dockerfile: %s", bc.dvmDockerfile)
	if bc.registryEndpoint != "" {
		bc.renderInfof("Registry cache: %s", bc.registryEndpoint)
	}
	if bc.cacheReadiness != nil {
		summary := bc.cacheReadiness.FormatSummary()
		if bc.cacheReadiness.AllHealthy {
			bc.renderInfo(summary)
		} else {
			bc.renderWarning(summary)
		}
	}
	bc.renderBlank()
	bc.renderInfo("Next: Attach to your workspace with: dvm attach")
}

// pushToRegistry tags and pushes the built image to the registry.
func (bc *buildContext) pushToRegistry() {
	bc.renderBlank()
	bc.renderProgressf("Pushing image to registry: %s", bc.registryEndpoint)

	registryImage := fmt.Sprintf("%s/%s", bc.registryEndpoint, bc.imageName)
	if err := tagImageForRegistry(bc.platform, bc.imageName, registryImage, bc.out()); err != nil {
		bc.renderWarningf("Failed to tag image for registry: %v", err)
		bc.renderInfo("Skipping push to registry")
		return
	}

	if err := pushImageToRegistry(bc.platform, registryImage, bc.out()); err != nil {
		bc.renderWarningf("Failed to push image to registry: %v", err)
	} else {
		bc.renderSuccessf("Pushed to registry: %s", registryImage)
		slog.Info("pushed image to registry", "image", registryImage)
	}
}
