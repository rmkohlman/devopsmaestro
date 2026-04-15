package builders

import (
	"context"
	"log/slog"
	"os/exec"
	"time"
)

// WatchdogConfig configures the watchdog behavior for process monitoring.
type WatchdogConfig struct {
	// PollInterval is how often to check if the condition is met.
	// Default: 2 seconds
	PollInterval time.Duration

	// Timeout is the maximum time to wait for the process to complete.
	// Default: 45 minutes (matches --timeout CLI flag default)
	Timeout time.Duration

	// CleanupWait is how long to wait for process cleanup after cancellation.
	// Default: 5 seconds
	CleanupWait time.Duration
}

// DefaultWatchdogConfig returns sensible defaults for production use.
func DefaultWatchdogConfig() WatchdogConfig {
	return WatchdogConfig{
		PollInterval: 2 * time.Second,
		Timeout:      45 * time.Minute,
		CleanupWait:  5 * time.Second,
	}
}

// WatchdogResult indicates how the watched process terminated.
type WatchdogResult int

const (
	// WatchdogCompleted means the process completed normally
	WatchdogCompleted WatchdogResult = iota
	// WatchdogDetected means the condition was detected and process was terminated
	WatchdogDetected
	// WatchdogTimedOut means the process exceeded the timeout
	WatchdogTimedOut
	// WatchdogCancelled means the parent context was cancelled
	WatchdogCancelled
)

// RunWithWatchdog executes a command while polling a condition function.
// If the condition returns true before the command exits, the command is cancelled.
// This is useful for handling processes that complete their work but don't exit.
//
// Parameters:
//   - ctx: Parent context for cancellation
//   - cmd: The command to execute (must not be started yet)
//   - checkCondition: Function that returns true when the work is complete
//   - cfg: Watchdog configuration
//
// Returns:
//   - result: How the process terminated
//   - err: Any error from the process (nil if condition detected or cancelled)
func RunWithWatchdog(
	ctx context.Context,
	cmd *exec.Cmd,
	checkCondition func(ctx context.Context) bool,
	cfg WatchdogConfig,
) (WatchdogResult, error) {
	// Apply defaults for zero values
	if cfg.PollInterval == 0 {
		cfg.PollInterval = 2 * time.Second
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 45 * time.Minute
	}
	if cfg.CleanupWait == 0 {
		cfg.CleanupWait = 5 * time.Second
	}

	// Channel to receive command result
	cmdDone := make(chan error, 1)

	// Start the command in a goroutine
	if err := cmd.Start(); err != nil {
		return WatchdogCompleted, err
	}

	go func() {
		cmdDone <- cmd.Wait()
	}()

	// Helper to kill the process (logs errors but doesn't fail)
	killProcess := func() {
		if cmd.Process != nil {
			if err := cmd.Process.Kill(); err != nil {
				slog.Debug("watchdog: failed to kill process", "pid", cmd.Process.Pid, "error", err)
			}
		}
	}

	// Set up polling ticker
	ticker := time.NewTicker(cfg.PollInterval)
	defer ticker.Stop()

	// Set up timeout
	timeout := time.NewTimer(cfg.Timeout)
	defer timeout.Stop()

	// Track if we killed the process (to distinguish from other errors)
	var killedByWatchdog bool

	for {
		select {
		case err := <-cmdDone:
			// Command completed
			if err != nil {
				// Check if we killed it (which means success via watchdog)
				if killedByWatchdog {
					return WatchdogDetected, nil
				}
				// Final condition check: the process exited with error, but the
				// work may have completed (e.g., Docker buildx built the image
				// then exited non-zero due to a post-export hang or cleanup
				// issue on Colima). If the condition is satisfied, the build
				// actually succeeded — treat it as a watchdog detection.
				if checkCondition(ctx) {
					slog.Info("watchdog: process exited with error but condition is satisfied, treating as success",
						"error", err)
					return WatchdogDetected, nil
				}
				return WatchdogCompleted, err
			}
			return WatchdogCompleted, nil

		case <-ticker.C:
			// Check if condition is met (e.g., image exists)
			if checkCondition(ctx) {
				killedByWatchdog = true
				killProcess() // Kill the hung process
				// Wait briefly for cleanup
				select {
				case <-cmdDone:
				case <-time.After(cfg.CleanupWait):
				}
				return WatchdogDetected, nil
			}

		case <-timeout.C:
			killProcess()
			// Wait briefly for cleanup
			select {
			case <-cmdDone:
			case <-time.After(cfg.CleanupWait):
			}
			return WatchdogTimedOut, nil

		case <-ctx.Done():
			killProcess()
			return WatchdogCancelled, ctx.Err()
		}
	}
}
