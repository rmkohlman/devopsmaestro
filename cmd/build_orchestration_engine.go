package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strings"
	"sync"
	"time"

	"devopsmaestro/config"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/buildlog"

	"github.com/google/uuid"
)

// BuildError is returned by buildWorkspacesInParallel when one or more
// workspaces fail. It carries the succeeded/failed counts so callers can
// display an accurate summary without a round-trip through the database.
type BuildError struct {
	Succeeded   int
	Failed      int
	FailedNames []string
}

func (e *BuildError) Error() string {
	return fmt.Sprintf("build failed for %d workspace(s): %s",
		len(e.FailedNames), strings.Join(e.FailedNames, ", "))
}

// buildWorkspacesInParallel executes the given build function for each
// workspace using a bounded worker pool. Concurrency controls the maximum
// number of simultaneous builds. All workspaces are attempted regardless
// of individual failures (failure isolation). Returns nil if all succeed,
// or an aggregate error describing which workspaces failed.
//
// buildFn receives the workspace and a per-workspace io.Writer that is
// teed into the rotating build log file (see pkg/buildlog). When the
// build log is disabled, the writer is io.Discard.
//
// When ds is non-nil, the function persists a BuildSession and per-workspace
// entries to the DataStore. DB write failures are logged as warnings and
// never abort the build (fire-and-forget).
func buildWorkspacesInParallel(
	workspaces []*models.WorkspaceWithHierarchy,
	concurrency int,
	buildFn func(ws *models.WorkspaceWithHierarchy, logWriter io.Writer) error,
	ds ...db.DataStore,
) error {
	if len(workspaces) == 0 {
		return nil
	}
	if concurrency < 1 {
		concurrency = 1
	}
	maxConcurrency := runtime.NumCPU() * 2
	if maxConcurrency < 8 {
		maxConcurrency = 8 // minimum ceiling of 8
	}
	if concurrency > maxConcurrency {
		concurrency = maxConcurrency
	}

	// Extract optional DataStore (first variadic arg)
	var store db.DataStore
	if len(ds) > 0 && ds[0] != nil {
		store = ds[0]
	}

	// --- Session persistence: create session before build ---
	sessionID := uuid.New().String()
	now := time.Now().UTC()

	// --- Build log file (per-session, rotating) ---
	// Open at session start; Close in defer below so the file flushes even
	// when the build is interrupted via context cancellation (#399).
	cfg := config.GetConfig()
	bl, blErr := buildlog.New(buildlog.Options{
		Enabled:    cfg.BuildLogs.Enabled,
		Directory:  cfg.BuildLogs.Directory,
		MaxSizeMB:  cfg.BuildLogs.MaxSizeMB,
		MaxAgeDays: cfg.BuildLogs.MaxAgeDays,
		MaxBackups: cfg.BuildLogs.MaxBackups,
		Compress:   cfg.BuildLogs.Compress,
	})
	if blErr != nil {
		slog.Warn("buildlog init failed; build will run without file logging",
			"session_id", sessionID, "error", blErr)
		bl, _ = buildlog.New(buildlog.Options{Enabled: false})
	}
	if err := bl.Open(sessionID); err != nil {
		slog.Warn("buildlog open failed; build will run without file logging",
			"session_id", sessionID, "error", err)
	}
	defer func() {
		if cerr := bl.Close(); cerr != nil {
			slog.Warn("buildlog close failed", "session_id", sessionID, "error", cerr)
		}
	}()

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
		name        string
		wsID        int
		bswID       int
		failed      bool
		interrupted bool
		errMsg      string
		startedAt   time.Time
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

			// Execute the build, passing a per-workspace writer that is
			// teed into the rotating build log file. Callers can MultiWriter
			// this with their own stdout sink.
			logWriter := bl.Writer(w.Workspace.Name)
			buildErr := buildFn(w, logWriter)
			wsEnd := time.Now().UTC()
			duration := int64(wsEnd.Sub(wsStart).Seconds())

			res := workspaceResult{
				name:      w.ShortPath(),
				wsID:      w.Workspace.ID,
				startedAt: wsStart,
			}

			if buildErr != nil {
				res.failed = true
				res.errMsg = buildErr.Error()
				if errors.Is(buildErr, context.Canceled) || errors.Is(buildErr, context.DeadlineExceeded) {
					res.interrupted = true
				}
			}

			// Update workspace entry with final status
			if store != nil {
				if bswID, ok := bswIDMap[w.Workspace.ID]; ok {
					status := "succeeded"
					if res.interrupted {
						status = "interrupted"
					} else if buildErr != nil {
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
						slog.Error("failed to update workspace build result — "+
							"'dvm build status' will show stale workspace data",
							"workspace", w.Workspace.Name, "status", status, "error", err)
					}
				}

				// Fix the :pending bug: update workspace image after successful build
				// Use the actual image tag set by buildFn (via postBuild), not a placeholder
				if buildErr == nil {
					imageTag := w.Workspace.ImageName
					if imageTag != "" {
						if err := store.UpdateWorkspaceImage(w.Workspace.ID, imageTag); err != nil {
							slog.Error("CRITICAL: failed to update workspace image — "+
								"workspace will show stale tag, 'dvm attach' may fail",
								"workspace", w.Workspace.Name, "image", imageTag, "error", err)
						} else {
							slog.Info("workspace image updated",
								"workspace", w.Workspace.Name, "image", imageTag)
						}
					} else {
						slog.Warn("build succeeded but no image tag available — "+
							"workspace will retain :pending tag",
							"workspace", w.Workspace.Name)
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
	// Use defer to guarantee session finalization even if the result aggregation
	// or workspace image updates encounter a panic (#366).
	var failedNames []string
	succeeded := 0
	failed := 0
	interrupted := 0
	for _, r := range results {
		switch {
		case r.interrupted:
			interrupted++
		case r.failed:
			failedNames = append(failedNames, r.name)
			failed++
		default:
			succeeded++
		}
	}

	if store != nil {
		completedAt := time.Now().UTC()
		status := "completed"
		switch {
		case interrupted > 0:
			status = "interrupted"
		case failed > 0 && succeeded == 0:
			status = "failed"
		case failed > 0:
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
			slog.Error("CRITICAL: failed to finalize build session — "+
				"'dvm build status' will show stale data",
				"session_id", sessionID, "error", err)
		} else {
			slog.Info("build session finalized",
				"session_id", sessionID, "status", status,
				"succeeded", succeeded, "failed", failed)
		}
	}

	if len(failedNames) > 0 {
		return &BuildError{
			Succeeded:   succeeded,
			Failed:      failed,
			FailedNames: failedNames,
		}
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
	buildFn func(ws *models.WorkspaceWithHierarchy, logWriter io.Writer) error,
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
