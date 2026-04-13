package cmd

import (
	"context"
	"devopsmaestro/operators"
	"log/slog"

	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/spf13/cobra"
)

// runSandboxDelete removes a specific sandbox container.
func runSandboxDelete(cmd *cobra.Command, name string) error {
	ctx := context.Background()

	runtime, err := operators.NewContainerRuntime()
	if err != nil {
		render.Error("Failed to create container runtime")
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
		render.Plain(FormatSuggestions("List sandboxes: dvm sandbox get"))
		return errSilent
	}

	if err := runtime.RemoveContainer(ctx, name, true); err != nil {
		render.Errorf("Failed to delete sandbox %s: %v", name, err)
		return errSilent
	}

	render.Successf("Sandbox %s deleted", name)
	return nil
}

// runSandboxDeleteAll removes all sandbox containers.
func runSandboxDeleteAll(cmd *cobra.Command) error {
	ctx := context.Background()

	runtime, err := operators.NewContainerRuntime()
	if err != nil {
		render.Error("Failed to create container runtime")
		return errSilent
	}

	containers, err := runtime.ListContainers(ctx, sandboxLabels)
	if err != nil {
		render.Errorf("Failed to list sandboxes: %v", err)
		return errSilent
	}

	if len(containers) == 0 {
		render.Info("No sandboxes to delete")
		return nil
	}

	deleted := 0
	for _, c := range containers {
		if err := runtime.RemoveContainer(ctx, c.Name, true); err != nil {
			slog.Warn("failed to delete sandbox", "name", c.Name, "error", err)
			render.Warningf("Failed to delete %s: %v", c.Name, err)
			continue
		}
		deleted++
	}

	render.Successf("Deleted %d sandbox(es)", deleted)
	if deleted < len(containers) {
		render.Warningf("%d sandbox(es) could not be deleted", len(containers)-deleted)
	}
	return nil
}
