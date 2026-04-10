package cmd

import (
	"fmt"
	"strings"
	"sync"

	"devopsmaestro/models"

	"github.com/google/uuid"
)

// buildWorkspacesInParallel executes the given build function for each
// workspace using a bounded worker pool. Concurrency controls the maximum
// number of simultaneous builds. All workspaces are attempted regardless
// of individual failures (failure isolation). Returns nil if all succeed,
// or an aggregate error describing which workspaces failed.
func buildWorkspacesInParallel(
	workspaces []*models.WorkspaceWithHierarchy,
	concurrency int,
	buildFn func(ws *models.WorkspaceWithHierarchy) error,
) error {
	if len(workspaces) == 0 {
		return nil
	}
	if concurrency < 1 {
		concurrency = 1
	}

	// Semaphore channel bounds concurrency
	semaphore := make(chan struct{}, concurrency)

	var mu sync.Mutex
	var failedNames []string

	var wg sync.WaitGroup
	for _, ws := range workspaces {
		wg.Add(1)
		go func(w *models.WorkspaceWithHierarchy) {
			defer wg.Done()

			// Acquire semaphore slot
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := buildFn(w); err != nil {
				mu.Lock()
				failedNames = append(failedNames, w.Workspace.Name)
				mu.Unlock()
			}
		}(ws)
	}
	wg.Wait()

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
	go func() {
		_ = buildWorkspacesInParallel(workspaces, concurrency, buildFn)
	}()

	return sessionID, nil
}
