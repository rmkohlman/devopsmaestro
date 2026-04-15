package cmd

import (
	"devopsmaestro/config"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/operators"
	"devopsmaestro/pkg/credentialbridge"
	ws "devopsmaestro/pkg/workspace"
	"devopsmaestro/utils"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/rmkohlman/MaestroSDK/paths"
	"github.com/rmkohlman/MaestroSDK/render"
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
func prepareStagingDirectory(stagingDir, appPath, appName, workspaceName string, ds db.DataStore, workspace *models.Workspace, out io.Writer) error {
	render.MsgTo(out, "", render.Message{Level: render.LevelProgress, Content: "Preparing build staging directory..."})

	// Clean and recreate staging directory.
	// Cleanup failure is non-fatal: a leftover directory from a previous build
	// should not prevent the current build from proceeding.
	if err := os.RemoveAll(stagingDir); err != nil && !os.IsNotExist(err) {
		slog.Warn("failed to clean staging directory (non-fatal, proceeding with build)",
			"path", stagingDir, "error", err)
	}

	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return fmt.Errorf("failed to create staging directory: %w", err)
	}

	// Copy app source to staging directory (for Dockerfile COPY commands)
	render.MsgTo(out, "", render.Message{Level: render.LevelProgress, Content: "Copying application source..."})
	if err := copyAppSource(appPath, stagingDir); err != nil {
		return fmt.Errorf("failed to copy app source: %w", err)
	}

	// Generate shell configuration files (.zshrc and starship.toml)
	// This is done here to ensure shell config is ALWAYS generated, even without nvim
	if err := generateShellConfig(stagingDir, appName, workspaceName, ds, workspace); err != nil {
		return fmt.Errorf("failed to generate shell config: %w", err)
	}

	render.MsgTo(out, "", render.Message{Level: render.LevelSuccess, Content: "Staging directory prepared"})
	return nil
}

// copyAppSource copies application source code to staging directory, excluding generated files.
// Symlinks are resolved and validated to ensure they don't escape the source directory tree,
// preventing symlink attacks where a link could point to sensitive files (e.g., /etc/passwd, ~/.ssh/).
func copyAppSource(srcDir, dstDir string) error {
	// Resolve the source directory to an absolute, symlink-free path for reliable comparisons
	absSrcDir, err := filepath.EvalSymlinks(srcDir)
	if err != nil {
		return fmt.Errorf("failed to resolve source directory: %w", err)
	}

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

		// Check for symlinks using the info from Walk (which uses Lstat, not Stat)
		if info.Mode()&os.ModeSymlink != 0 {
			// Resolve the symlink target and verify it stays within the source tree
			resolved, err := filepath.EvalSymlinks(path)
			if err != nil {
				slog.Warn("skipping symlink: failed to resolve target", "path", relPath, "error", err)
				return nil
			}

			if !isPathWithinDir(resolved, absSrcDir) {
				slog.Warn("skipping symlink: target escapes source directory", "path", relPath, "target", resolved)
				return nil
			}

			// Symlink target is within source tree — stat the resolved target to copy it
			resolvedInfo, err := os.Stat(resolved)
			if err != nil {
				slog.Warn("skipping symlink: cannot stat resolved target", "path", relPath, "target", resolved, "error", err)
				return nil
			}

			if resolvedInfo.IsDir() {
				// For directory symlinks within the source tree, create the directory
				return os.MkdirAll(dstPath, resolvedInfo.Mode())
			}

			// Copy the resolved file content
			return copyFile(resolved, dstPath, resolvedInfo.Mode())
		}

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath, info.Mode())
	})
}

// isPathWithinDir checks whether the given path is within (or equal to) the directory dir.
// Both paths should be absolute and cleaned. This is used to prevent symlink escape attacks.
func isPathWithinDir(path, dir string) bool {
	// Clean both paths for consistent comparison
	path = filepath.Clean(path)
	dir = filepath.Clean(dir)

	// The path must start with the directory prefix followed by a separator,
	// or be exactly the directory itself
	if path == dir {
		return true
	}
	return strings.HasPrefix(path, dir+string(filepath.Separator))
}

// shouldSkipPath determines if a path should be skipped during app source copy
func shouldSkipPath(path string) bool {
	skipDirs := []string{".git", paths.DVMDirName, "node_modules", "vendor", "__pycache__", ".venv", "venv"}
	skipFiles := []string{".DS_Store", "Thumbs.db", "*.log", "Dockerfile.dvm", ".dockerignore"}

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

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return os.Chmod(dst, mode)
}

// imageNameToSafeSlug converts an image name (e.g. "dvm-dev-daa-agents:20260414-230555")
// to a filesystem-safe slug by replacing non-alphanumeric characters with hyphens.
func imageNameToSafeSlug(imageName string) string {
	var b strings.Builder
	for _, c := range imageName {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' {
			b.WriteRune(c)
		} else {
			b.WriteRune('-')
		}
	}
	return b.String()
}

// copyImageToNamespace copies the built image from buildkit namespace to devopsmaestro namespace.
// This is needed because BuildKit creates images in its own namespace.
//
// The temp file uses a unique name per image (slug + 8-char UUID) to prevent
// collisions when multiple builds run concurrently under the same PID (#359).
func copyImageToNamespace(platform *operators.Platform, imageName string, out io.Writer) error {
	fmt.Fprintln(out)
	render.MsgTo(out, "", render.Message{Level: render.LevelProgress, Content: "Copying image to devopsmaestro namespace..."})

	profile := platform.Profile
	if profile == "" {
		profile = "default"
	}

	slug := imageNameToSafeSlug(imageName)
	tmpFile := fmt.Sprintf("/tmp/dvm-image-%s-%s.tar", slug, uuid.New().String()[:8])
	slog.Debug("image copy temp file", "path", tmpFile, "image", imageName)

	// Save image from buildkit namespace
	saveCmd := exec.Command("colima", "--profile", profile, "ssh", "--",
		"sudo", "nerdctl", "--namespace", "buildkit", "image", "save", imageName, "-o", tmpFile)
	saveCmd.Stdout = out
	saveCmd.Stderr = out
	if err := saveCmd.Run(); err != nil {
		return fmt.Errorf("failed to save image %s: %w", imageName, err)
	}

	// Verify the tar file exists before attempting to load it.
	// This catches silent export failures early with a clear message (#359).
	verifyCmd := exec.Command("colima", "--profile", profile, "ssh", "--",
		"sudo", "test", "-f", tmpFile)
	if err := verifyCmd.Run(); err != nil {
		return fmt.Errorf("image save reported success but tar file %s does not exist (image: %s)", tmpFile, imageName)
	}

	// Load image into devopsmaestro namespace
	loadCmd := exec.Command("colima", "--profile", profile, "ssh", "--",
		"sudo", "nerdctl", "--namespace", "devopsmaestro", "image", "load", "-i", tmpFile)
	loadCmd.Stdout = out
	loadCmd.Stderr = out
	if err := loadCmd.Run(); err != nil {
		return fmt.Errorf("failed to load image %s: %w", imageName, err)
	}

	// Clean up temp file
	cleanCmd := exec.Command("colima", "--profile", profile, "ssh", "--", "sudo", "rm", "-f", tmpFile)
	cleanCmd.Run() // Ignore errors on cleanup

	render.MsgTo(out, "", render.Message{Level: render.LevelSuccess, Content: "Image copied to devopsmaestro namespace"})
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
	if app.DomainID.Valid {
		domain, err := ds.GetDomainByID(int(app.DomainID.Int64))
		if err == nil && domain.EcosystemID.Valid {
			ecosystem, err := ds.GetEcosystemByID(int(domain.EcosystemID.Int64))
			if err == nil {
				ecoCreds, err := ds.ListCredentialsByScope(models.CredentialScopeEcosystem, int64(ecosystem.ID))
				if err == nil && len(ecoCreds) > 0 {
					scopes = append(scopes, config.CredentialScope{
						Type:        "ecosystem",
						ID:          int64(ecosystem.ID),
						Name:        ecosystem.Name,
						Credentials: credentialbridge.CredentialsToMap(ecoCreds),
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
					Credentials: credentialbridge.CredentialsToMap(domainCreds),
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
			Credentials: credentialbridge.CredentialsToMap(appCreds),
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
				Credentials: credentialbridge.CredentialsToMap(wsCreds),
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
func tagImageForRegistry(platform *operators.Platform, sourceImage, targetImage string, out io.Writer) error {
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

	cmd.Stdout = out
	cmd.Stderr = out
	return cmd.Run()
}

// pushImageToRegistry pushes an image to a registry.
// For Docker/OrbStack/Podman, uses docker push command.
// For Colima/containerd, uses nerdctl push command.
func pushImageToRegistry(platform *operators.Platform, image string, out io.Writer) error {
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

	cmd.Stdout = out
	cmd.Stderr = out
	return cmd.Run()
}
