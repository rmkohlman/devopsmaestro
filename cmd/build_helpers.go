package cmd

import (
	"devopsmaestro/config"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/operators"
	"devopsmaestro/pkg/paths"
	ws "devopsmaestro/pkg/workspace"
	"devopsmaestro/render"
	"devopsmaestro/utils"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// getBuildSourcePath determines the source path for building a workspace.
// When a workspace has a GitRepoID (created with --repo flag), the source code
// is in the workspace repo path (~/.devopsmaestro/workspaces/{slug}/repo/),
// not in the original app.Path. This function returns the correct path to use.
func getBuildSourcePath(ds db.DataStore, workspace *models.Workspace, appPath string) (string, error) {
	if workspace.GitRepoID.Valid {
		repoPath, err := ws.GetWorkspaceRepoPath(workspace.Slug)
		if err != nil {
			return "", fmt.Errorf("failed to get workspace repo path: %w", err)
		}
		return repoPath, nil
	}
	return appPath, nil
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
// sourcePath is used for auto-detection (should be the worktree checkout, not the bare mirror).
// Returns (languageName, version, wasDetected) - wasDetected is true if we fell back to auto-detection.
func getLanguageFromApp(app *models.App, sourcePath string) (langName, version string, detected bool) {
	// Try App.Language first (uses model's GetLanguageConfig method)
	if langConfig := app.GetLanguageConfig(); langConfig != nil {
		slog.Debug("using language from app model", "language", langConfig.Name, "version", langConfig.Version)
		return langConfig.Name, langConfig.Version, false
	}

	// Fall back to auto-detection using sourcePath (worktree checkout)
	lang, err := utils.DetectLanguage(sourcePath)
	if err != nil {
		slog.Debug("language detection error", "error", err)
		return "unknown", "", true
	}

	if lang != nil {
		ver := utils.DetectVersion(lang.Name, sourcePath)
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

// prepareStagingDirectory creates and populates the staging directory for container builds.
// This includes copying app source and generating shell configuration (starship.toml, .zshrc).
// This function is ALWAYS called during build, regardless of nvim configuration.
func prepareStagingDirectory(stagingDir, appPath, appName, workspaceName string, ds db.DataStore, workspace *models.Workspace) error {
	render.Progress("Preparing build staging directory...")

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

	// Generate shell configuration files (.zshrc and starship.toml)
	// This is done here to ensure shell config is ALWAYS generated, even without nvim
	if err := generateShellConfig(stagingDir, appName, workspaceName, ds, workspace); err != nil {
		return fmt.Errorf("failed to generate shell config: %w", err)
	}

	render.Success("Staging directory prepared")
	return nil
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
	skipDirs := []string{".git", paths.DVMDirName, "node_modules", "vendor", "__pycache__", ".venv", "venv"}
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

// copyImageToNamespace copies the built image from buildkit namespace to devopsmaestro namespace
// This is needed because BuildKit creates images in its own namespace
func copyImageToNamespace(platform *operators.Platform, imageName string) error {
	render.Blank()
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

// getRelativePath is a helper function to get relative path for display
func getRelativePath(base, target string) string {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	return rel
}

// loadBuildCredentials loads and resolves credentials from the hierarchy:
// Global -> Ecosystem -> Domain -> App -> Workspace
// Used for both build-time (--build-arg) and runtime (container env) injection.
// Environment variables always take highest priority.
func loadBuildCredentials(ds db.DataStore, app *models.App, workspace *models.Workspace) (map[string]string, []string) {
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

	// Initialize vault backend via auto-token resolution chain
	var backend config.SecretBackend
	token, tokenErr := config.ResolveVaultToken()
	if tokenErr != nil {
		slog.Warn("failed to resolve vault token", "error", tokenErr)
	}
	if token != "" {
		if err := config.EnsureVaultDaemon(); err != nil {
			slog.Warn("failed to start vault daemon", "error", err)
		} else {
			vb, err := config.NewVaultBackend(token)
			if err != nil {
				slog.Warn("failed to create vault backend", "error", err)
			} else {
				backend = vb
			}
		}
	}

	// Resolve all credentials (env vars checked last internally)
	resolved, errors := config.ResolveCredentialsWithBackend(backend, scopes...)

	// Collect warnings for failed credential resolutions
	var warnings []string
	for name, err := range errors {
		warnings = append(warnings, fmt.Sprintf("credential %q failed to resolve: %v", name, err))
		slog.Warn("failed to resolve credential", "name", name, "error", err)
	}

	if len(resolved) > 0 {
		slog.Info("resolved build credentials", "count", len(resolved))
	}

	return resolved, warnings
}

// tagImageForRegistry tags an image for pushing to a registry.
// For Docker/OrbStack/Podman, uses docker tag command.
// For Colima/containerd, uses nerdctl tag command.
func tagImageForRegistry(platform *operators.Platform, sourceImage, targetImage string) error {
	slog.Debug("tagging image for registry", "source", sourceImage, "target", targetImage)

	var cmd *exec.Cmd
	if platform.IsContainerd() {
		// Use nerdctl via colima ssh for containerd
		profile := platform.Profile
		if profile == "" {
			profile = "default"
		}
		cmd = exec.Command("colima", "--profile", profile, "ssh", "--",
			"sudo", "nerdctl", "--namespace", "devopsmaestro", "tag", sourceImage, targetImage)
	} else {
		// Use docker for Docker/OrbStack/Podman
		cmd = exec.Command("docker", "tag", sourceImage, targetImage)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// pushImageToRegistry pushes an image to a registry.
// For Docker/OrbStack/Podman, uses docker push command.
// For Colima/containerd, uses nerdctl push command.
func pushImageToRegistry(platform *operators.Platform, image string) error {
	slog.Debug("pushing image to registry", "image", image)

	var cmd *exec.Cmd
	if platform.IsContainerd() {
		// Use nerdctl via colima ssh for containerd
		profile := platform.Profile
		if profile == "" {
			profile = "default"
		}
		cmd = exec.Command("colima", "--profile", profile, "ssh", "--",
			"sudo", "nerdctl", "--namespace", "devopsmaestro", "push", "--insecure-registry", image)
	} else {
		// Use docker for Docker/OrbStack/Podman
		cmd = exec.Command("docker", "push", image)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
