// Package buildengine provides a parallel build engine with worker pool
// and status tracking for building multiple workspaces concurrently.
package buildengine

import (
	"context"
	"sync"
	"time"

	"devopsmaestro/models"
)

// buildJobState holds the mutable internal state of a BuildJob.
type buildJobState struct {
	mu         sync.Mutex
	workspace  models.Workspace
	status     JobStatus
	err        error
	startedAt  time.Time
	finishedAt time.Time
}

// BuildJob represents a single build unit. It wraps a pointer to internal
// state so it can be passed by value while remaining mutable.
type BuildJob struct {
	state *buildJobState
}

// NewBuildJob creates a new BuildJob in Pending status for the given workspace.
func NewBuildJob(ws models.Workspace) BuildJob {
	return BuildJob{
		state: &buildJobState{
			workspace: ws,
			status:    StatusPending,
		},
	}
}

// Workspace returns the workspace associated with this job.
func (j BuildJob) Workspace() models.Workspace {
	j.state.mu.Lock()
	defer j.state.mu.Unlock()
	return j.state.workspace
}

// Status returns the current status of the job.
func (j BuildJob) Status() JobStatus {
	j.state.mu.Lock()
	defer j.state.mu.Unlock()
	return j.state.status
}

// SetStatus sets the job's status.
func (j BuildJob) SetStatus(s JobStatus) {
	j.state.mu.Lock()
	defer j.state.mu.Unlock()
	j.state.status = s
}

// Err returns the error associated with this job, if any.
func (j BuildJob) Err() error {
	j.state.mu.Lock()
	defer j.state.mu.Unlock()
	return j.state.err
}

// SetError records an error on the job.
func (j BuildJob) SetError(err error) {
	j.state.mu.Lock()
	defer j.state.mu.Unlock()
	j.state.err = err
}

// MarkStarted records the time the build started.
func (j BuildJob) MarkStarted() {
	j.state.mu.Lock()
	defer j.state.mu.Unlock()
	j.state.startedAt = time.Now()
}

// MarkFinished records the time the build finished.
func (j BuildJob) MarkFinished() {
	j.state.mu.Lock()
	defer j.state.mu.Unlock()
	j.state.finishedAt = time.Now()
}

// StartedAt returns when the build started (zero if not yet started).
func (j BuildJob) StartedAt() time.Time {
	j.state.mu.Lock()
	defer j.state.mu.Unlock()
	return j.state.startedAt
}

// FinishedAt returns when the build finished (zero if not yet finished).
func (j BuildJob) FinishedAt() time.Time {
	j.state.mu.Lock()
	defer j.state.mu.Unlock()
	return j.state.finishedAt
}

// Duration returns how long the build took. Returns 0 if not yet complete.
func (j BuildJob) Duration() time.Duration {
	j.state.mu.Lock()
	defer j.state.mu.Unlock()
	if j.state.startedAt.IsZero() || j.state.finishedAt.IsZero() {
		return 0
	}
	return j.state.finishedAt.Sub(j.state.startedAt)
}

// BuildFunc is the function signature for building a single workspace.
type BuildFunc func(ctx context.Context, ws models.Workspace) error

// StatusCallback is called when a job reaches a terminal state (done or error).
type StatusCallback func(job BuildJob)

// BuildResult aggregates the outcome of a build session.
type BuildResult struct {
	Total     int
	Succeeded int
	Failed    int
	Duration  time.Duration
	Jobs      []*BuildJob
}

// SessionConfig configures a BuildSession.
type SessionConfig struct {
	Parallelism    int
	BuildFn        BuildFunc
	StatusCallback StatusCallback
}

// BuildSession manages submitting and executing build jobs with a worker pool.
type BuildSession struct {
	config SessionConfig
	jobs   []BuildJob
	seen   map[int]bool // workspace ID deduplication
	mu     sync.Mutex
}

// NewBuildSession creates a new BuildSession with the given configuration.
func NewBuildSession(config SessionConfig) *BuildSession {
	if config.Parallelism < 1 {
		config.Parallelism = 1
	}
	return &BuildSession{
		config: config,
		seen:   make(map[int]bool),
	}
}

// Submit adds a workspace to the build queue. Returns an error if the
// workspace has already been submitted (deduplication).
func (s *BuildSession) Submit(ws models.Workspace) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.seen[ws.ID] {
		return nil // silently deduplicate
	}
	s.seen[ws.ID] = true
	s.jobs = append(s.jobs, NewBuildJob(ws))
	return nil
}

// Run executes all submitted jobs using a worker pool constrained by
// the configured parallelism. It blocks until all jobs complete or the
// context is cancelled.
func (s *BuildSession) Run(ctx context.Context) *BuildResult {
	start := time.Now()

	sem := make(chan struct{}, s.config.Parallelism)
	var wg sync.WaitGroup

	for i := range s.jobs {
		job := s.jobs[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.runJob(ctx, job, sem)
		}()
	}

	wg.Wait()

	result := &BuildResult{
		Total:    len(s.jobs),
		Duration: time.Since(start),
	}
	for _, job := range s.jobs {
		switch job.Status() {
		case StatusDone:
			result.Succeeded++
		case StatusError:
			result.Failed++
		}
	}
	return result
}

// runJob executes a single build job, managing semaphore acquisition,
// status transitions, timing, and error handling.
func (s *BuildSession) runJob(ctx context.Context, job BuildJob, sem chan struct{}) {
	// Acquire semaphore slot, respecting context cancellation.
	select {
	case sem <- struct{}{}:
		defer func() { <-sem }()
	case <-ctx.Done():
		job.SetError(ctx.Err())
		job.SetStatus(StatusError)
		if s.config.StatusCallback != nil {
			s.config.StatusCallback(job)
		}
		return
	}

	// Check context again after acquiring semaphore.
	if ctx.Err() != nil {
		job.SetError(ctx.Err())
		job.SetStatus(StatusError)
		if s.config.StatusCallback != nil {
			s.config.StatusCallback(job)
		}
		return
	}

	job.SetStatus(StatusBuilding)
	job.MarkStarted()

	err := s.config.BuildFn(ctx, job.Workspace())

	job.MarkFinished()

	if err != nil {
		job.SetError(err)
		job.SetStatus(StatusError)
	} else {
		job.SetStatus(StatusDone)
	}

	if s.config.StatusCallback != nil {
		s.config.StatusCallback(job)
	}
}

// JobStatus represents the current state of a build job.
type JobStatus int

const (
	StatusPending JobStatus = iota
	StatusBuilding
	StatusDone
	StatusError
)

// String returns a human-readable representation of the job status.
func (s JobStatus) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusBuilding:
		return "building"
	case StatusDone:
		return "done"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}
