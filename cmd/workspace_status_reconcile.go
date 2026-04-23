package cmd

import (
	"context"
	"log/slog"

	"devopsmaestro/models"
	"devopsmaestro/operators"
)

// reconcileWorkspaceStatuses updates the Status field of each workspace in-place
// to reflect the current state of the container runtime.
//
// Background: `dvm status` queries the container runtime (containerd/docker)
// directly for the list of running workspaces, while `dvm get workspaces` was
// previously rendering the cached Status column from the SQLite database. When
// a container is stopped/started outside dvm — or even just between dvm
// invocations across reboots — the cached DB value drifts from the runtime
// truth, producing the divergence reported in issue #405.
//
// The container runtime is the authoritative source for whether a container is
// currently running, so this helper consults it once per command invocation
// and rewrites Status on each workspace in the slice. On runtime error
// (runtime unavailable, etc.) we leave the DB-cached values untouched so
// listing still works in offline scenarios — best effort, not strict.
//
// Matching strategy:
//  1. By full container ID
//  2. By 12-char short container ID prefix (Docker/containerd convention)
//  3. By workspace name — matched against both the runtime's container Name
//     (Docker uses the workspace name as the container name) and the
//     `io.devopsmaestro.workspace` label (containerd uses the container ID
//     hash as the Name, so the label is the only reliable mapping back to
//     the workspace name there). See issue #418.
func reconcileWorkspaceStatuses(workspaces []*models.Workspace) {
	if len(workspaces) == 0 {
		return
	}

	runtime, err := operators.NewContainerRuntime()
	if err != nil {
		slog.Debug("workspace status reconcile: failed to create runtime", "error", err)
		return
	}

	infos, err := runtime.ListWorkspaces(context.Background())
	if err != nil {
		slog.Debug("workspace status reconcile: failed to list workspaces", "error", err)
		return
	}

	applyWorkspaceStatusReconcile(workspaces, infos)
}

// applyWorkspaceStatusReconcile applies the matching logic against a pre-fetched
// slice of runtime WorkspaceInfos. Extracted for unit-testability.
func applyWorkspaceStatusReconcile(workspaces []*models.Workspace, infos []operators.WorkspaceInfo) {
	runningByID := make(map[string]bool, len(infos))
	runningByShortID := make(map[string]bool, len(infos))
	runningByName := make(map[string]bool, len(infos))
	for _, info := range infos {
		if !isRunning(info.Status) {
			continue
		}
		if info.ID != "" {
			runningByID[info.ID] = true
			if len(info.ID) >= 12 {
				runningByShortID[info.ID[:12]] = true
			}
		}
		if info.Name != "" {
			runningByName[info.Name] = true
		}
		// Containerd's WorkspaceInfo.Name is the full container ID hash, not
		// the workspace name. The workspace name lives in the
		// io.devopsmaestro.workspace label (surfaced as info.Workspace).
		// Index by it so reconcile works on containerd as well as Docker
		// (issue #418).
		if info.Workspace != "" {
			runningByName[info.Workspace] = true
		}
	}

	for _, ws := range workspaces {
		if ws == nil {
			continue
		}
		running := false
		if ws.ContainerID.Valid && ws.ContainerID.String != "" {
			cid := ws.ContainerID.String
			if runningByID[cid] {
				running = true
			} else if len(cid) >= 12 && runningByShortID[cid[:12]] {
				running = true
			}
		}
		if !running && runningByName[ws.Name] {
			running = true
		}
		if running {
			ws.Status = "running"
		} else {
			ws.Status = "stopped"
		}
	}
}

// reconcileWorkspaceHierarchyStatuses is a convenience wrapper that reconciles
// statuses for resolver results (which wrap *models.Workspace in a hierarchy
// envelope).
func reconcileWorkspaceHierarchyStatuses(results []*models.WorkspaceWithHierarchy) {
	if len(results) == 0 {
		return
	}
	workspaces := make([]*models.Workspace, 0, len(results))
	for _, wh := range results {
		if wh != nil && wh.Workspace != nil {
			workspaces = append(workspaces, wh.Workspace)
		}
	}
	reconcileWorkspaceStatuses(workspaces)
}
