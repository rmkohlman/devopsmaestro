package cmd

import (
	"context"
	"devopsmaestro/operators"
	"fmt"
	"log/slog"
	"os"

	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/spf13/cobra"
)

// runSandboxAttach re-attaches to a running sandbox container.
func runSandboxAttach(cmd *cobra.Command, name string) error {
	ctx := context.Background()

	runtime, err := operators.NewContainerRuntime()
	if err != nil {
		render.Error("Failed to create container runtime")
		render.Plain(FormatSuggestions(SuggestNoContainerRuntime()...))
		return errSilent
	}

	// Verify sandbox exists
	containers, err := runtime.ListContainers(ctx, sandboxLabels)
	if err != nil {
		render.Errorf("Failed to list sandboxes: %v", err)
		return errSilent
	}

	found := false
	for _, c := range containers {
		if c.Name == name {
			found = true
			break
		}
	}
	if !found {
		render.Errorf("Sandbox %q not found", name)
		render.Plain(FormatSuggestions(
			"List sandboxes: dvm sandbox get",
			fmt.Sprintf("Create a new sandbox: dvm sandbox create <language>"),
		))
		return errSilent
	}

	render.Progressf("Attaching to sandbox %s...", name)
	fmt.Fprintf(os.Stderr, "\x1b]0;[dvm-sandbox] %s\x07", name)

	attachErr := runtime.AttachToWorkspace(ctx, operators.AttachOptions{
		WorkspaceID: name,
		Shell:       "/bin/bash",
		LoginShell:  true,
		UID:         1000,
		GID:         1000,
		Env: map[string]string{
			"TERM":        "xterm-256color",
			"DVM_SANDBOX": "true",
		},
	})

	// Reset terminal title
	fmt.Fprintf(os.Stderr, "\x1b]0;\x07")

	if attachErr != nil {
		slog.Warn("sandbox attach error", "name", name, "error", attachErr)
		return fmt.Errorf("failed to attach to sandbox: %w", attachErr)
	}

	render.Info("Detached from sandbox")
	return nil
}
