package buildengine_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"devopsmaestro/models"
	"devopsmaestro/pkg/buildengine"
)

// --- helpers ---

func makeWorkspace(id int, name string) models.Workspace {
	return models.Workspace{
		ID:   id,
		Name: name,
	}
}

// successFunc returns a BuildFunc that succeeds immediately.
func successFunc() buildengine.BuildFunc {
	return func(ctx context.Context, ws models.Workspace) error {
		return nil
	}
}

// sleepFunc returns a BuildFunc that sleeps for d then succeeds.
func sleepFunc(d time.Duration) buildengine.BuildFunc {
	return func(ctx context.Context, ws models.Workspace) error {
		select {
		case <-time.After(d):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// errorFunc returns a BuildFunc that always returns the given error.
func errorFunc(err error) buildengine.BuildFunc {
	return func(ctx context.Context, ws models.Workspace) error {
		return err
	}
}

// --- TestBuildJob_Lifecycle ---

// TestBuildJob_Lifecycle verifies that a BuildJob starts in Pending status,
// transitions to Building when work begins, then ends in Done or Error.
func TestBuildJob_Lifecycle(t *testing.T) {
	ws := makeWorkspace(1, "dev")

	job := buildengine.NewBuildJob(ws)

	if job.Status() != buildengine.StatusPending {
		t.Errorf("expected StatusPending, got %v", job.Status())
	}

	job.SetStatus(buildengine.StatusBuilding)
	if job.Status() != buildengine.StatusBuilding {
		t.Errorf("expected StatusBuilding, got %v", job.Status())
	}

	job.SetStatus(buildengine.StatusDone)
	if job.Status() != buildengine.StatusDone {
		t.Errorf("expected StatusDone, got %v", job.Status())
	}
}

// TestBuildJob_Timing verifies that StartedAt and FinishedAt are recorded.
func TestBuildJob_Timing(t *testing.T) {
	ws := makeWorkspace(1, "dev")
	job := buildengine.NewBuildJob(ws)

	if !job.StartedAt().IsZero() {
		t.Error("expected zero StartedAt before build starts")
	}
	if !job.FinishedAt().IsZero() {
		t.Error("expected zero FinishedAt before build completes")
	}

	job.MarkStarted()
	if job.StartedAt().IsZero() {
		t.Error("expected non-zero StartedAt after MarkStarted")
	}

	job.MarkFinished()
	if job.FinishedAt().IsZero() {
		t.Error("expected non-zero FinishedAt after MarkFinished")
	}

	if job.Duration() <= 0 {
		t.Error("expected positive duration after build completes")
	}
}

// TestBuildJob_ErrorStatus verifies that a job set to Error status retains the error.
func TestBuildJob_ErrorStatus(t *testing.T) {
	ws := makeWorkspace(1, "dev")
	job := buildengine.NewBuildJob(ws)

	buildErr := errors.New("docker build failed")
	job.SetError(buildErr)
	job.SetStatus(buildengine.StatusError)

	if job.Status() != buildengine.StatusError {
		t.Errorf("expected StatusError, got %v", job.Status())
	}
	if job.Err() == nil {
		t.Fatal("expected non-nil error on job")
	}
	if job.Err().Error() != buildErr.Error() {
		t.Errorf("expected error %q, got %q", buildErr.Error(), job.Err().Error())
	}
}

// --- TestBuildSession_SingleJob ---

// TestBuildSession_SingleJob verifies that a session with one job completes
// and the job reaches StatusDone.
func TestBuildSession_SingleJob(t *testing.T) {
	ctx := context.Background()
	ws := makeWorkspace(1, "dev")

	session := buildengine.NewBuildSession(buildengine.SessionConfig{
		Parallelism: 1,
		BuildFn:     successFunc(),
	})

	err := session.Submit(ws)
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	result := session.Run(ctx)

	if result.Total != 1 {
		t.Errorf("expected Total=1, got %d", result.Total)
	}
	if result.Succeeded != 1 {
		t.Errorf("expected Succeeded=1, got %d", result.Succeeded)
	}
	if result.Failed != 0 {
		t.Errorf("expected Failed=0, got %d", result.Failed)
	}
}

// --- TestBuildSession_MultipleJobs ---

// TestBuildSession_MultipleJobs verifies that all submitted jobs complete.
func TestBuildSession_MultipleJobs(t *testing.T) {
	ctx := context.Background()

	workspaces := []models.Workspace{
		makeWorkspace(1, "dev"),
		makeWorkspace(2, "staging"),
		makeWorkspace(3, "prod"),
	}

	session := buildengine.NewBuildSession(buildengine.SessionConfig{
		Parallelism: 3,
		BuildFn:     successFunc(),
	})

	for _, ws := range workspaces {
		if err := session.Submit(ws); err != nil {
			t.Fatalf("Submit(%q) failed: %v", ws.Name, err)
		}
	}

	result := session.Run(ctx)

	if result.Total != 3 {
		t.Errorf("expected Total=3, got %d", result.Total)
	}
	if result.Succeeded != 3 {
		t.Errorf("expected Succeeded=3, got %d", result.Succeeded)
	}
	if result.Failed != 0 {
		t.Errorf("expected Failed=0, got %d", result.Failed)
	}
}

// --- TestBuildSession_ConcurrencyLimit ---

// TestBuildSession_ConcurrencyLimit verifies that with parallelism=2, at most
// 2 jobs execute simultaneously, even when more are queued.
func TestBuildSession_ConcurrencyLimit(t *testing.T) {
	ctx := context.Background()

	var (
		maxSeen int32
		current int32
	)

	// slowFn increments a counter, sleeps briefly, then decrements.
	// We track the maximum concurrent goroutines.
	slowFn := func(_ context.Context, _ models.Workspace) error {
		cur := atomic.AddInt32(&current, 1)
		for {
			old := atomic.LoadInt32(&maxSeen)
			if cur <= old || atomic.CompareAndSwapInt32(&maxSeen, old, cur) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
		atomic.AddInt32(&current, -1)
		return nil
	}

	session := buildengine.NewBuildSession(buildengine.SessionConfig{
		Parallelism: 2,
		BuildFn:     buildengine.BuildFunc(slowFn),
	})

	for i := 1; i <= 6; i++ {
		_ = session.Submit(makeWorkspace(i, "ws"))
	}

	session.Run(ctx)

	if atomic.LoadInt32(&maxSeen) > 2 {
		t.Errorf("concurrency exceeded limit: max concurrent=%d, want ≤2", atomic.LoadInt32(&maxSeen))
	}
}

// --- TestBuildSession_ErrorIsolation ---

// TestBuildSession_ErrorIsolation verifies that one failing job does not
// prevent other jobs from completing successfully.
func TestBuildSession_ErrorIsolation(t *testing.T) {
	ctx := context.Background()
	buildErr := errors.New("intentional failure")

	// Job IDs: 1 succeeds, 2 fails, 3 succeeds.
	fn := func(_ context.Context, ws models.Workspace) error {
		if ws.ID == 2 {
			return buildErr
		}
		return nil
	}

	session := buildengine.NewBuildSession(buildengine.SessionConfig{
		Parallelism: 3,
		BuildFn:     buildengine.BuildFunc(fn),
	})

	for i := 1; i <= 3; i++ {
		_ = session.Submit(makeWorkspace(i, "ws"))
	}

	result := session.Run(ctx)

	if result.Total != 3 {
		t.Errorf("expected Total=3, got %d", result.Total)
	}
	if result.Succeeded != 2 {
		t.Errorf("expected Succeeded=2, got %d", result.Succeeded)
	}
	if result.Failed != 1 {
		t.Errorf("expected Failed=1, got %d", result.Failed)
	}
}

// --- TestBuildSession_ResultAggregation ---

// TestBuildSession_ResultAggregation verifies that BuildResult correctly
// aggregates counts and records a non-zero total duration.
func TestBuildSession_ResultAggregation(t *testing.T) {
	ctx := context.Background()

	fn := func(_ context.Context, ws models.Workspace) error {
		if ws.ID%2 == 0 {
			return errors.New("even IDs fail")
		}
		return nil
	}

	session := buildengine.NewBuildSession(buildengine.SessionConfig{
		Parallelism: 4,
		BuildFn:     buildengine.BuildFunc(fn),
	})

	// Submit 4 jobs: IDs 1,2,3,4 → 2 succeed, 2 fail.
	for i := 1; i <= 4; i++ {
		_ = session.Submit(makeWorkspace(i, "ws"))
	}

	result := session.Run(ctx)

	if result.Total != 4 {
		t.Errorf("expected Total=4, got %d", result.Total)
	}
	if result.Succeeded != 2 {
		t.Errorf("expected Succeeded=2, got %d", result.Succeeded)
	}
	if result.Failed != 2 {
		t.Errorf("expected Failed=2, got %d", result.Failed)
	}
	if result.Duration <= 0 {
		t.Error("expected positive total Duration in BuildResult")
	}
}

// --- TestBuildSession_StatusCallback ---

// TestBuildSession_StatusCallback verifies that the StatusCallback is called
// for each status transition (building → done/error) for each job.
func TestBuildSession_StatusCallback(t *testing.T) {
	ctx := context.Background()

	var (
		mu       sync.Mutex
		statuses []buildengine.JobStatus
	)

	callback := func(job buildengine.BuildJob) {
		mu.Lock()
		defer mu.Unlock()
		statuses = append(statuses, job.Status())
	}

	session := buildengine.NewBuildSession(buildengine.SessionConfig{
		Parallelism:    2,
		BuildFn:        successFunc(),
		StatusCallback: callback,
	})

	for i := 1; i <= 3; i++ {
		_ = session.Submit(makeWorkspace(i, "ws"))
	}

	session.Run(ctx)

	mu.Lock()
	defer mu.Unlock()

	// Each job should have triggered at least one callback (on completion).
	if len(statuses) < 3 {
		t.Errorf("expected at least 3 status callbacks (one per job), got %d", len(statuses))
	}

	// All reported statuses should be terminal: Done or Error.
	for i, s := range statuses {
		if s != buildengine.StatusDone && s != buildengine.StatusError {
			t.Errorf("callback[%d]: expected terminal status, got %v", i, s)
		}
	}
}

// --- TestBuildSession_ContextCancellation ---

// TestBuildSession_ContextCancellation verifies that cancelling the context
// causes the session to stop processing pending jobs early.
func TestBuildSession_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Use a long sleep so jobs block until cancelled.
	session := buildengine.NewBuildSession(buildengine.SessionConfig{
		Parallelism: 1,
		BuildFn:     sleepFunc(10 * time.Second),
	})

	// Submit more jobs than the worker count.
	for i := 1; i <= 5; i++ {
		_ = session.Submit(makeWorkspace(i, "ws"))
	}

	// Cancel after a short delay so the first job starts but the rest are pending.
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	result := session.Run(ctx)

	// With cancellation, not all jobs should have succeeded.
	if result.Succeeded == 5 {
		t.Error("expected cancellation to prevent some jobs from succeeding, but all 5 succeeded")
	}
}

// --- TestBuildSession_Deduplication ---

// TestBuildSession_Deduplication verifies that submitting the same workspace
// twice only results in a single build execution.
func TestBuildSession_Deduplication(t *testing.T) {
	ctx := context.Background()

	var buildCount int32

	countFn := func(_ context.Context, _ models.Workspace) error {
		atomic.AddInt32(&buildCount, 1)
		return nil
	}

	session := buildengine.NewBuildSession(buildengine.SessionConfig{
		Parallelism: 2,
		BuildFn:     buildengine.BuildFunc(countFn),
	})

	ws := makeWorkspace(42, "duplicate-ws")

	// Submit the same workspace twice.
	err1 := session.Submit(ws)
	err2 := session.Submit(ws)

	if err1 != nil {
		t.Fatalf("first Submit failed unexpectedly: %v", err1)
	}
	// Second submit should return a deduplication error (or be silently dropped).
	// Either way, the build function must only run once.
	_ = err2

	session.Run(ctx)

	count := atomic.LoadInt32(&buildCount)
	if count != 1 {
		t.Errorf("expected build function to run exactly once for duplicate workspace, ran %d times", count)
	}
}

// --- Table-driven: status string representation ---

// TestJobStatus_String verifies that status constants produce meaningful strings.
func TestJobStatus_String(t *testing.T) {
	tests := []struct {
		status buildengine.JobStatus
		want   string
	}{
		{buildengine.StatusPending, "pending"},
		{buildengine.StatusBuilding, "building"},
		{buildengine.StatusDone, "done"},
		{buildengine.StatusError, "error"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			got := tc.status.String()
			if got != tc.want {
				t.Errorf("JobStatus.String(): got %q, want %q", got, tc.want)
			}
		})
	}
}
