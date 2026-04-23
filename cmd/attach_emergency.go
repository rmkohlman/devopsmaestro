package cmd

import (
	"context"
	"crypto/rand"
	"devopsmaestro/builders"
	"devopsmaestro/builders/emergency"
	"devopsmaestro/db"
	"devopsmaestro/operators"
	"devopsmaestro/pkg/resolver"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/spf13/cobra"
)

// attachEmergency is the cobra flag bound to `dvm attach --emergency`.
//
// Short-form decision: there is NO short flag. `-e` is already taken by the
// global `--ecosystem` hierarchy flag (see cmd/flags.go) and `-E` was rejected
// because uppercase short flags are not used elsewhere in the dvm CLI. Users
// must spell `--emergency` in full — appropriate for an unusual escape hatch.
var attachEmergency bool

// runAttachEmergency drops the user into a lightweight Alpine fallback
// container with the workspace mounted at /workspace. It is invoked by
// `dvm attach --emergency` when the normal build is broken and the user
// just needs to edit files quickly.
//
// Flow:
//  1. Resolve a mount path (workspace if flags/context match, else $PWD).
//  2. Detect platform and ensure the emergency image exists locally,
//     building it on demand the first time.
//  3. Start a uniquely-named ephemeral container with the workspace mounted.
//  4. Attach an interactive bash shell.
//  5. On exit, force-remove the container so nothing is left behind.
func runAttachEmergency(cmd *cobra.Command) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	render.Warning("Entering EMERGENCY mode — using lightweight fallback container.")
	render.Plain("This is a degraded environment intended for quick edits only.")
	render.Blank()

	mountPath, err := resolveEmergencyMountPath(cmd)
	if err != nil {
		return err
	}
	render.Infof("Workspace mount: %s -> /workspace", mountPath)

	runtime, err := operators.NewContainerRuntime()
	if err != nil {
		render.Plain(FormatSuggestions(SuggestNoContainerRuntime()...))
		return fmt.Errorf("failed to create container runtime: %w", err)
	}
	render.Infof("Platform: %s", runtime.GetPlatformName())

	if err := ensureEmergencyImage(ctx); err != nil {
		return fmt.Errorf("failed to prepare emergency image: %w", err)
	}

	containerName := generateEmergencyName()
	slog.Info("starting emergency container", "name", containerName, "mount", mountPath)

	if _, err := runtime.StartWorkspace(ctx, operators.StartOptions{
		ImageName:     emergency.ImageName,
		WorkspaceName: containerName,
		ContainerName: containerName,
		AppName:       "emergency",
		AppPath:       mountPath,
		WorkingDir:    "/workspace",
		UID:           1000,
		GID:           1000,
		Env: map[string]string{
			"DVM_EMERGENCY": "1",
			"TERM":          "xterm-256color",
		},
		Labels: map[string]string{
			emergency.LabelKey: "true",
			"dvm.ephemeral":    "true",
		},
	}); err != nil {
		return fmt.Errorf("failed to start emergency container: %w", err)
	}

	// Always tear the container down on exit — emergency sessions are ephemeral
	// and should never leak. This handles both clean exits and panics from attach.
	defer func() {
		render.Progress("Cleaning up emergency container...")
		if rmErr := runtime.RemoveContainer(context.Background(), containerName, true); rmErr != nil {
			slog.Warn("failed to remove emergency container",
				"name", containerName, "error", rmErr)
			render.Warningf("Could not remove %s: %v", containerName, rmErr)
			render.Plain(fmt.Sprintf("Manual cleanup: dvm container rm -f %s", containerName))
			return
		}
		render.Successf("Emergency container %s removed.", containerName)
	}()

	render.Progress("Attaching to emergency shell...")
	fmt.Fprintf(os.Stderr, "\x1b]0;[dvm-emergency] %s\x07", filepath.Base(mountPath))

	attachErr := runtime.AttachToWorkspace(ctx, operators.AttachOptions{
		WorkspaceID: containerName,
		Shell:       "/bin/bash",
		LoginShell:  true,
		UID:         1000,
		GID:         1000,
		Env: map[string]string{
			"TERM":          "xterm-256color",
			"DVM_EMERGENCY": "1",
		},
	})

	// Reset terminal title regardless of attach result.
	fmt.Fprintf(os.Stderr, "\x1b]0;\x07")

	if attachErr != nil {
		return fmt.Errorf("attach failed: %w", attachErr)
	}

	render.Info("Emergency session ended.")
	return nil
}

// resolveEmergencyMountPath picks the directory to mount at /workspace.
//
// Priority:
//  1. If hierarchy flags resolve to a real workspace, use that workspace's
//     repo path (so emergency edits land in the same place normal builds use).
//  2. Otherwise fall back to the current working directory — the user is most
//     likely sitting in a broken repo when they reach for --emergency.
func resolveEmergencyMountPath(cmd *cobra.Command) (string, error) {
	ds, err := getDataStore(cmd)
	if err == nil && attachFlags.HasAnyFlag() {
		// Best-effort workspace resolution. Failures here are NOT fatal — the
		// whole point of emergency mode is that the workspace may be in a bad
		// state. We just fall back to $PWD if anything looks off.
		if path, ok := tryResolveWorkspacePath(ds); ok {
			return path, nil
		}
		render.Warning("Could not resolve a workspace from flags; falling back to current directory.")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot determine current directory: %w", err)
	}
	return cwd, nil
}

// tryResolveWorkspacePath attempts to translate the active hierarchy flags
// into a host path suitable for mounting. Returns ok=false on any error so
// the caller can fall back to a safer default.
func tryResolveWorkspacePath(ds db.DataStore) (string, bool) {
	wsResolver := resolver.NewWorkspaceResolver(ds)
	result, err := wsResolver.Resolve(attachFlags.ToFilter())
	if err != nil || result == nil || result.Workspace == nil || result.App == nil {
		return "", false
	}
	path, err := getMountPath(ds, result.Workspace, result.App.Path)
	if err != nil || path == "" {
		return "", false
	}
	if _, statErr := os.Stat(path); statErr != nil {
		return "", false
	}
	return path, true
}

// generateEmergencyName produces a unique container name per session. The
// random suffix lets several emergency sessions coexist without colliding.
func generateEmergencyName() string {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Fall back to a timestamp-based suffix if /dev/urandom is unavailable.
		return fmt.Sprintf("%s%d", emergency.ContainerNamePrefix, time.Now().UnixNano())
	}
	return emergency.ContainerNamePrefix + hex.EncodeToString(b[:])
}

// ensureEmergencyImage builds the emergency fallback image if it is not
// already present locally. The image is cached under emergency.ImageName so
// subsequent invocations skip the build entirely.
//
// This mirrors the pattern used by `dvm sandbox create`: the embedded
// Dockerfile is written into a temp dir and built via the standard
// builders.ImageBuilder factory so it works on every supported platform
// (Docker-API and BuildKit/containerd alike).
func ensureEmergencyImage(ctx context.Context) error {
	platform, err := detectPlatform()
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "dvm-emergency-build-*")
	if err != nil {
		return fmt.Errorf("failed to create build context dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	dfPath := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dfPath, []byte(emergency.Dockerfile()), 0644); err != nil {
		return fmt.Errorf("failed to write emergency Dockerfile: %w", err)
	}

	builder, err := builders.NewImageBuilder(builders.BuilderConfig{
		Platform:   platform,
		Namespace:  "devopsmaestro",
		AppPath:    tmpDir,
		ImageName:  emergency.ImageName,
		Dockerfile: dfPath,
	})
	if err != nil {
		return fmt.Errorf("failed to create image builder: %w", err)
	}
	defer builder.Close()

	exists, existsErr := builder.ImageExists(ctx)
	if existsErr != nil {
		slog.Warn("could not check emergency image existence; will rebuild",
			"error", existsErr)
	}
	if existsErr == nil && exists {
		render.Infof("Using cached emergency image %s", emergency.ImageName)
		return nil
	}

	render.Progressf("Building emergency image %s (one-time, ~30s)...", emergency.ImageName)
	if err := builder.Build(ctx, builders.BuildOptions{}); err != nil {
		return fmt.Errorf("emergency image build failed: %w", err)
	}

	// On containerd-based platforms, the BuildKit builder writes the image
	// into BuildKit's namespace. Copy it into the devopsmaestro namespace
	// where the runtime expects to find it (same as sandbox create).
	if platform.IsContainerd() {
		if err := copyImageToNamespace(platform, emergency.ImageName, os.Stdout); err != nil {
			return fmt.Errorf("failed to copy emergency image to namespace: %w", err)
		}
	}

	render.Successf("Emergency image ready: %s", emergency.ImageName)
	return nil
}
