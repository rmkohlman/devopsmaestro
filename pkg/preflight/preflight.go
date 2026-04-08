package preflight

import (
	"context"
	"time"
)

// CheckStatus represents the result status of a pre-flight check
type CheckStatus int

const (
	StatusOK      CheckStatus = iota // Check passed
	StatusWarning                    // Check passed with warnings
	StatusError                      // Check failed
	StatusSkipped                    // Check was skipped
)

// CheckResult contains the result of running a check
type CheckResult struct {
	Status  CheckStatus
	Message string
	Details map[string]interface{}
}

// IsSuccess returns true if the check was successful (OK, Warning, or Skipped)
func (cr *CheckResult) IsSuccess() bool {
	return cr.Status != StatusError
}

// Check represents a single pre-flight validation step that can be run before
// a command proceeds. Implementations test a specific system condition such as
// container runtime availability, database connectivity, or binary presence.
//
// Checks are registered with a PreflightRunner, which executes them sequentially
// or in parallel and aggregates results. Implementations should be safe for
// concurrent use when registered in a parallel runner.
//
// Example implementation:
//
//	type RuntimeCheck struct{}
//
//	func (c *RuntimeCheck) Name() string { return "container-runtime" }
//
//	func (c *RuntimeCheck) Run(ctx context.Context) preflight.CheckResult {
//	    _, err := operators.NewPlatformDetector()
//	    if err != nil {
//	        return preflight.CheckResult{Status: preflight.StatusError, Message: err.Error()}
//	    }
//	    return preflight.CheckResult{Status: preflight.StatusOK, Message: "runtime detected"}
//	}
type Check interface {
	// Name returns a short human-readable identifier for the check (e.g., "container-runtime").
	// Used in output display and logging. Must be unique within a PreflightRunner.
	Name() string

	// Run executes the check and returns its result.
	// Implementations must respect ctx cancellation and timeouts.
	// The result Status indicates OK, Warning, Error, or Skipped.
	Run(ctx context.Context) CheckResult
}

// PreflightRunner executes multiple checks and aggregates results
type PreflightRunner struct {
	checks   []Check
	timeout  time.Duration
	parallel bool
}

// NewPreflightRunner creates a new PreflightRunner
func NewPreflightRunner() *PreflightRunner {
	return &PreflightRunner{
		checks:   make([]Check, 0),
		timeout:  30 * time.Second, // default 30s
		parallel: false,
	}
}

// AddCheck adds a check to the runner
func (pr *PreflightRunner) AddCheck(check Check) {
	pr.checks = append(pr.checks, check)
}

// GetChecks returns all registered checks
func (pr *PreflightRunner) GetChecks() []Check {
	return pr.checks
}

// SetTimeout sets the timeout for each check
func (pr *PreflightRunner) SetTimeout(milliseconds int) {
	pr.timeout = time.Duration(milliseconds) * time.Millisecond
}

// SetParallel enables or disables parallel execution
func (pr *PreflightRunner) SetParallel(parallel bool) {
	pr.parallel = parallel
}

// Run executes all checks and returns results
func (pr *PreflightRunner) Run(ctx context.Context) []CheckResult {
	if len(pr.checks) == 0 {
		return []CheckResult{}
	}

	if pr.parallel {
		return pr.runParallel(ctx)
	}
	return pr.runSequential(ctx)
}

// runSequential executes checks one by one in order
func (pr *PreflightRunner) runSequential(ctx context.Context) []CheckResult {
	results := make([]CheckResult, 0, len(pr.checks))

	for _, check := range pr.checks {
		result := pr.runCheckWithTimeout(ctx, check)
		results = append(results, result)
	}

	return results
}

// runParallel executes checks concurrently
func (pr *PreflightRunner) runParallel(ctx context.Context) []CheckResult {
	results := make([]CheckResult, len(pr.checks))

	// Use channel to collect results with their index
	type indexedResult struct {
		index  int
		result CheckResult
	}
	resultsChan := make(chan indexedResult, len(pr.checks))

	// Launch all checks concurrently
	for i, check := range pr.checks {
		go func(idx int, chk Check) {
			result := pr.runCheckWithTimeout(ctx, chk)
			resultsChan <- indexedResult{index: idx, result: result}
		}(i, check)
	}

	// Collect results
	for i := 0; i < len(pr.checks); i++ {
		indexed := <-resultsChan
		results[indexed.index] = indexed.result
	}

	return results
}

// runCheckWithTimeout runs a single check with the configured timeout
func (pr *PreflightRunner) runCheckWithTimeout(ctx context.Context, check Check) CheckResult {
	// Create a timeout context for this check
	checkCtx, cancel := context.WithTimeout(ctx, pr.timeout)
	defer cancel()

	// Run check in a goroutine to handle timeout
	resultChan := make(chan CheckResult, 1)
	go func() {
		resultChan <- check.Run(checkCtx)
	}()

	// Wait for either result or timeout
	select {
	case result := <-resultChan:
		return result
	case <-checkCtx.Done():
		// Check if it's a timeout or context cancellation
		if ctx.Err() != nil {
			// Parent context was cancelled
			return CheckResult{
				Status:  StatusError,
				Message: "Check cancelled: " + ctx.Err().Error(),
			}
		}
		// Timeout
		return CheckResult{
			Status:  StatusError,
			Message: "Check timeout after " + pr.timeout.String(),
		}
	}
}

// HasErrors returns true if any check returned an error
func (pr *PreflightRunner) HasErrors(results []CheckResult) bool {
	for _, result := range results {
		if result.Status == StatusError {
			return true
		}
	}
	return false
}

// HasWarnings returns true if any check returned a warning
func (pr *PreflightRunner) HasWarnings(results []CheckResult) bool {
	for _, result := range results {
		if result.Status == StatusWarning {
			return true
		}
	}
	return false
}
