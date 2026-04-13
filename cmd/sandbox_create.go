package cmd

import (
	"context"
	"devopsmaestro/builders"
	"devopsmaestro/models"
	"devopsmaestro/operators"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/spf13/cobra"
)

// runSandboxCreate implements the sandbox create workflow:
// 1. Resolve language preset
// 2. Resolve version (flag, interactive, or default)
// 3. Generate Dockerfile and build image (or reuse cached)
// 4. Create container with sandbox labels
// 5. Attach TTY
// 6. On exit → stop + remove container
func runSandboxCreate(cmd *cobra.Command, lang string) error {
	ctx := context.Background()

	// 1. Resolve preset
	preset, ok := models.GetPreset(lang)
	if !ok {
		supported := models.ListPresets()
		render.Errorf("Unknown language %q. Supported: %s", lang, strings.Join(supported, ", "))
		return errSilent
	}

	// 2. Resolve version
	version := sandboxFlags.version
	if version == "" {
		version = resolveVersion(preset)
	}

	// Validate version is in preset's list
	if !isValidVersion(preset, version) {
		render.Errorf("Version %q not available for %s. Available: %s",
			version, preset.Language, strings.Join(preset.Versions, ", "))
		return errSilent
	}

	// 3. Validate deps file if provided
	depsFile := sandboxFlags.deps
	if depsFile != "" {
		absPath, err := filepath.Abs(depsFile)
		if err != nil {
			return fmt.Errorf("invalid deps file path: %w", err)
		}
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			render.Errorf("Dependency file not found: %s", absPath)
			return errSilent
		}
		depsFile = absPath
	}

	// 4. Create container runtime
	runtime, err := operators.NewContainerRuntime()
	if err != nil {
		render.Error("Failed to create container runtime")
		render.Plain(FormatSuggestions(SuggestNoContainerRuntime()...))
		return errSilent
	}

	slog.Info("sandbox create", "lang", preset.Language, "version", version,
		"platform", runtime.GetPlatformName())
	render.Infof("Language: %s | Version: %s | Platform: %s",
		preset.Language, version, runtime.GetPlatformName())

	// 5. Build or reuse image
	imageName := fmt.Sprintf("dvm-sandbox-%s:%s", preset.Language, version)
	if err := ensureSandboxImage(ctx, runtime, preset, version, depsFile, imageName); err != nil {
		render.Errorf("Failed to build sandbox image: %v", err)
		return errSilent
	}

	// 6. Generate container name
	containerName := sandboxFlags.name
	if containerName == "" {
		containerName = generateSandboxName(preset.Language)
	}

	// 7. Create and start container
	containerID, err := startSandboxContainer(ctx, runtime, imageName, containerName, preset, version)
	if err != nil {
		render.Errorf("Failed to start sandbox: %v", err)
		return errSilent
	}

	slog.Info("sandbox started", "container", containerName, "id", containerID)

	// 8. Clone repo if requested
	if sandboxFlags.repo != "" {
		render.Infof("Note: --repo cloning is not yet implemented. Clone manually inside the sandbox.")
	}

	// 9. Attach to container
	render.Progressf("Attaching to sandbox %s...", containerName)
	fmt.Fprintf(os.Stderr, "\x1b]0;[dvm-sandbox] %s %s\x07", preset.Language, version)

	attachErr := runtime.AttachToWorkspace(ctx, operators.AttachOptions{
		WorkspaceID: containerName,
		Shell:       "/bin/bash",
		LoginShell:  true,
		UID:         1000,
		GID:         1000,
		Env: map[string]string{
			"TERM":             "xterm-256color",
			"DVM_SANDBOX":      "true",
			"DVM_SANDBOX_LANG": preset.Language,
		},
	})

	// Reset terminal title
	fmt.Fprintf(os.Stderr, "\x1b]0;\x07")

	// 10. Auto-cleanup: stop and remove container
	render.Progress("Cleaning up sandbox...")
	if cleanupErr := runtime.RemoveContainer(ctx, containerName, true); cleanupErr != nil {
		slog.Warn("failed to remove sandbox container", "name", containerName, "error", cleanupErr)
		render.Warningf("Failed to clean up container %s: %v", containerName, cleanupErr)
		render.Plain(FormatSuggestions(
			fmt.Sprintf("Manual cleanup: dvm sandbox delete %s", containerName),
		))
	} else {
		render.Successf("Sandbox %s removed", containerName)
	}

	if attachErr != nil {
		return fmt.Errorf("attach error: %w", attachErr)
	}

	return nil
}

// ensureSandboxImage builds the sandbox image if it doesn't exist or --no-cache is set.
func ensureSandboxImage(
	ctx context.Context,
	runtime operators.ContainerRuntime,
	preset models.SandboxPreset,
	version, depsFile, imageName string,
) error {
	// Check if image already exists (skip if --no-cache)
	if !sandboxFlags.noCache {
		exists, err := runtime.ImageExists(ctx, imageName)
		if err != nil {
			slog.Warn("failed to check image existence", "error", err)
		}
		if exists {
			render.Infof("Using cached image %s", imageName)
			return nil
		}
	}

	// Generate Dockerfile
	render.Progressf("Building sandbox image %s...", imageName)
	dockerfile := builders.GenerateSandboxDockerfile(preset, version, depsFile)

	// Write Dockerfile to a temp directory
	tmpDir, err := os.MkdirTemp("", "dvm-sandbox-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	dfPath := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dfPath, []byte(dockerfile), 0644); err != nil {
		return fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	// Copy deps file into build context if provided
	if depsFile != "" {
		baseName := filepath.Base(depsFile)
		data, err := os.ReadFile(depsFile)
		if err != nil {
			return fmt.Errorf("failed to read deps file: %w", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, baseName), data, 0644); err != nil {
			return fmt.Errorf("failed to copy deps file to build context: %w", err)
		}
	}

	// Build image
	return runtime.BuildImage(ctx, operators.BuildOptions{
		ImageName:    imageName,
		Dockerfile:   dfPath,
		BuildContext: tmpDir,
	})
}

// startSandboxContainer creates and starts a sandbox container with proper labels.
func startSandboxContainer(
	ctx context.Context,
	runtime operators.ContainerRuntime,
	imageName, containerName string,
	preset models.SandboxPreset,
	version string,
) (string, error) {
	return runtime.StartWorkspace(ctx, operators.StartOptions{
		ImageName:     imageName,
		WorkspaceName: containerName,
		ContainerName: containerName,
		WorkingDir:    "/sandbox",
		Env: map[string]string{
			"DVM_SANDBOX":      "true",
			"DVM_SANDBOX_LANG": preset.Language,
		},
		Labels: buildSandboxLabels(preset.Language, version, containerName),
	})
}
