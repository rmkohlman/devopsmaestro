package cmd

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"devopsmaestro/db"
	"devopsmaestro/models"

	"github.com/google/uuid"
)

// buildWorkspacesInParallel executes the given build function for each
// workspace using a bounded worker pool. Concurrency controls the maximum
// number of simultaneous builds. All workspaces are attempted regardless
// of individual failures (failure isolation). Returns nil if all succeed,
// or an aggregate error describing which workspaces failed.
//
// When ds is non-nil, the function persists a BuildSession and per-workspace
// entries to the DataStore. DB write failures are logged as warnings and
// never abort the build (fire-and-forget).
func buildWorkspacesInParallel(
	workspaces []*models.WorkspaceWithHierarchy,
	concurrency int,
	buildFn func(ws *models.WorkspaceWithHierarchy) error,
	ds ...db.DataStore,
) error {
	if len(workspaces) == 0 {
		return nil
	}
	if concurrency < 1 {
		concurrency = 1
	}

	// Extract optional DataStore (first variadic arg)
	var store db.DataStore
	if len(ds) > 0 && ds[0] != nil {
		store = ds[0]
	}

	// --- Session persistence: create session before build ---
	sessionID := uuid.New().String()
	now := time.Now().UTC()

	if store != nil {
		// GC: clean up sessions older than 30 days
		cleanupBuildSessions(store)

		session := &models.BuildSession{
			ID:              sessionID,
			StartedAt:       now,
			Status:          "running",
			TotalWorkspaces: len(workspaces),
		}
		if err := store.CreateBuildSession(session); err != nil {
			slog.Warn("failed to create build session", "error", err)
		}

		// Create per-workspace entries
		for _, ws := range workspaces {
			bsw := &models.BuildSessionWorkspace{
				SessionID:   sessionID,
				WorkspaceID: ws.Workspace.ID,
				Status:      "queued",
				StartedAt:   sql.NullTime{Time: now, Valid: true},
			}
			if err := store.CreateBuildSessionWorkspace(bsw); err != nil {
				slog.Warn("failed to create build session workspace entry",
					"workspace", ws.Workspace.Name, "error", err)
			}
		}
	}

	// --- Execute builds in parallel ---
	semaphore := make(chan struct{}, concurrency)

	type workspaceResult struct {
		name      string
		wsID      int
		bswID     int
		failed    bool
		errMsg    string
		startedAt time.Time
	}

	// Pre-map workspace IDs to build session workspace IDs
	bswIDMap := make(map[int]int) // workspace ID → build session workspace ID
	if store != nil {
		entries, err := store.GetBuildSessionWorkspaces(sessionID)
		if err == nil {
			for _, e := range entries {
				bswIDMap[e.WorkspaceID] = e.ID
			}
		}
	}

	var mu sync.Mutex
	var results []workspaceResult

	var wg sync.WaitGroup
	for _, ws := range workspaces {
		wg.Add(1)
		go func(w *models.WorkspaceWithHierarchy) {
			defer wg.Done()

			// Acquire semaphore slot
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			wsStart := time.Now().UTC()

			// Update workspace entry to "building"
			if store != nil {
				if bswID, ok := bswIDMap[w.Workspace.ID]; ok {
					bsw := &models.BuildSessionWorkspace{
						ID:        bswID,
						SessionID: sessionID,
						Status:    "building",
						StartedAt: sql.NullTime{Time: wsStart, Valid: true},
					}
					if err := store.UpdateBuildSessionWorkspace(bsw); err != nil {
						slog.Warn("failed to update workspace to building",
							"workspace", w.Workspace.Name, "error", err)
					}
				}
			}

			// Execute the build
			buildErr := buildFn(w)
			wsEnd := time.Now().UTC()
			duration := int64(wsEnd.Sub(wsStart).Seconds())

			res := workspaceResult{
				name:      w.Workspace.Name,
				wsID:      w.Workspace.ID,
				startedAt: wsStart,
			}

			if buildErr != nil {
				res.failed = true
				res.errMsg = buildErr.Error()
			}

			// Update workspace entry with final status
			if store != nil {
				if bswID, ok := bswIDMap[w.Workspace.ID]; ok {
					status := "succeeded"
					if buildErr != nil {
						status = "failed"
					}
					bsw := &models.BuildSessionWorkspace{
						ID:              bswID,
						SessionID:       sessionID,
						WorkspaceID:     w.Workspace.ID,
						Status:          status,
						StartedAt:       sql.NullTime{Time: wsStart, Valid: true},
						CompletedAt:     sql.NullTime{Time: wsEnd, Valid: true},
						DurationSeconds: sql.NullInt64{Int64: duration, Valid: true},
						ImageTag:        sql.NullString{String: w.Workspace.ImageName, Valid: w.Workspace.ImageName != ""},
						ErrorMessage:    sql.NullString{String: res.errMsg, Valid: res.errMsg != ""},
					}
					if err := store.UpdateBuildSessionWorkspace(bsw); err != nil {
						slog.Warn("failed to update workspace build result",
							"workspace", w.Workspace.Name, "error", err)
					}
				}

				// Fix the :pending bug: update workspace image after successful build
				// Use the actual image tag set by buildFn (via postBuild), not a placeholder
				if buildErr == nil {
					imageTag := w.Workspace.ImageName
					if imageTag != "" {
						if err := store.UpdateWorkspaceImage(w.Workspace.ID, imageTag); err != nil {
							slog.Warn("failed to update workspace image",
								"workspace", w.Workspace.Name, "error", err)
						}
					}
				}
			}

			mu.Lock()
			results = append(results, res)
			mu.Unlock()
		}(ws)
	}
	wg.Wait()

	// --- Session persistence: finalize session after build ---
	var failedNames []string
	succeeded := 0
	failed := 0
	for _, r := range results {
		if r.failed {
			failedNames = append(failedNames, r.name)
			failed++
		} else {
			succeeded++
		}
	}

	if store != nil {
		completedAt := time.Now().UTC()
		status := "completed"
		if failed > 0 && succeeded == 0 {
			status = "failed"
		} else if failed > 0 {
			status = "partial"
		}

		session := &models.BuildSession{
			ID:              sessionID,
			StartedAt:       now,
			CompletedAt:     sql.NullTime{Time: completedAt, Valid: true},
			Status:          status,
			TotalWorkspaces: len(workspaces),
			Succeeded:       succeeded,
			Failed:          failed,
		}
		if err := store.UpdateBuildSession(session); err != nil {
			slog.Warn("failed to update build session", "error", err)
		}
	}

	if len(failedNames) > 0 {
		return fmt.Errorf("build failed for %d workspace(s): %s",
			len(failedNames), strings.Join(failedNames, ", "))
	}
	return nil
}

// buildWorkspacesInParallelDetached launches the parallel build in a
// background goroutine and returns immediately with a unique session ID.
// The caller can monitor progress via 'dvm build status' using the
// returned session ID.
func buildWorkspacesInParallelDetached(
	workspaces []*models.WorkspaceWithHierarchy,
	concurrency int,
	buildFn func(ws *models.WorkspaceWithHierarchy) error,
) (string, error) {
	sessionID := uuid.New().String()

	// Launch in background goroutine — caller returns immediately
	// Note: detached builds don't persist sessions yet (see architecture review)
	go func() {
		_ = buildWorkspacesInParallel(workspaces, concurrency, buildFn)
	}()

	return sessionID, nil
}

// cleanupBuildSessions deletes build sessions older than 30 days.
// Failures are logged as warnings and never propagated.
func cleanupBuildSessions(ds db.DataStore) {
	cutoff := time.Now().UTC().AddDate(0, 0, -30)
	deleted, err := ds.DeleteBuildSessionsOlderThan(cutoff)
	if err != nil {
		slog.Warn("failed to cleanup old build sessions", "error", err)
		return
	}
	if deleted > 0 {
		slog.Info("cleaned up old build sessions", "deleted", deleted)
	}
}
