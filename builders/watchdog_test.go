package builders

import (
	"context"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultWatchdogConfig(t *testing.T) {
	cfg := DefaultWatchdogConfig()

	assert.Equal(t, 2*time.Second, cfg.PollInterval, "PollInterval should be 2 seconds")
	assert.Equal(t, 30*time.Minute, cfg.Timeout, "Timeout should be 30 minutes")
	assert.Equal(t, 5*time.Second, cfg.CleanupWait, "CleanupWait should be 5 seconds")
}

func TestRunWithWatchdog_CompletesNormally(t *testing.T) {
	tests := []struct {
		name         string
		command      []string
		wantResult   WatchdogResult
		wantErr      bool
		conditionMet bool
	}{
		{
			name:         "successful command completes",
			command:      []string{"echo", "hello"},
			wantResult:   WatchdogCompleted,
			wantErr:      false,
			conditionMet: false,
		},
		{
			name:         "command with short sleep completes",
			command:      []string{"sh", "-c", "sleep 0.1"},
			wantResult:   WatchdogCompleted,
			wantErr:      false,
			conditionMet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cmd := exec.Command(tt.command[0], tt.command[1:]...)

			cfg := WatchdogConfig{
				PollInterval: 50 * time.Millisecond,
				Timeout:      5 * time.Second,
				CleanupWait:  100 * time.Millisecond,
			}

			checkCondition := func(ctx context.Context) bool {
				return tt.conditionMet
			}

			result, err := RunWithWatchdog(ctx, cmd, checkCondition, cfg)

			assert.Equal(t, tt.wantResult, result, "Result should match expected")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRunWithWatchdog_ConditionDetected(t *testing.T) {
	tests := []struct {
		name           string
		conditionDelay time.Duration
		wantResult     WatchdogResult
		wantErr        bool
	}{
		{
			name:           "condition met after 100ms",
			conditionDelay: 100 * time.Millisecond,
			wantResult:     WatchdogDetected,
			wantErr:        false,
		},
		{
			name:           "condition met after 200ms",
			conditionDelay: 200 * time.Millisecond,
			wantResult:     WatchdogDetected,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Use a long-running command that won't exit on its own
			var cmd *exec.Cmd
			if runtime.GOOS == "windows" {
				cmd = exec.Command("timeout", "10")
			} else {
				cmd = exec.Command("sleep", "10")
			}

			cfg := WatchdogConfig{
				PollInterval: 50 * time.Millisecond,
				Timeout:      5 * time.Second,
				CleanupWait:  100 * time.Millisecond,
			}

			// Condition becomes true after delay
			start := time.Now()
			checkCondition := func(ctx context.Context) bool {
				return time.Since(start) >= tt.conditionDelay
			}

			result, err := RunWithWatchdog(ctx, cmd, checkCondition, cfg)

			assert.Equal(t, tt.wantResult, result, "Result should be WatchdogDetected")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err, "No error should be returned when condition detected")
			}
		})
	}
}

func TestRunWithWatchdog_Timeout(t *testing.T) {
	tests := []struct {
		name       string
		timeout    time.Duration
		wantResult WatchdogResult
	}{
		{
			name:       "timeout after 200ms",
			timeout:    200 * time.Millisecond,
			wantResult: WatchdogTimedOut,
		},
		{
			name:       "timeout after 300ms",
			timeout:    300 * time.Millisecond,
			wantResult: WatchdogTimedOut,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Use a long-running command
			var cmd *exec.Cmd
			if runtime.GOOS == "windows" {
				cmd = exec.Command("timeout", "10")
			} else {
				cmd = exec.Command("sleep", "10")
			}

			cfg := WatchdogConfig{
				PollInterval: 50 * time.Millisecond,
				Timeout:      tt.timeout,
				CleanupWait:  100 * time.Millisecond,
			}

			// Condition never becomes true
			checkCondition := func(ctx context.Context) bool {
				return false
			}

			start := time.Now()
			result, err := RunWithWatchdog(ctx, cmd, checkCondition, cfg)
			elapsed := time.Since(start)

			assert.Equal(t, tt.wantResult, result, "Result should be WatchdogTimedOut")
			assert.NoError(t, err, "No error should be returned on timeout")
			assert.GreaterOrEqual(t, elapsed, tt.timeout, "Should wait at least the timeout duration")
			assert.Less(t, elapsed, tt.timeout+500*time.Millisecond, "Should not wait much longer than timeout")
		})
	}
}

func TestRunWithWatchdog_ParentCancelled(t *testing.T) {
	tests := []struct {
		name        string
		cancelAfter time.Duration
		wantResult  WatchdogResult
		wantErr     bool
	}{
		{
			name:        "parent cancelled after 100ms",
			cancelAfter: 100 * time.Millisecond,
			wantResult:  WatchdogCancelled,
			wantErr:     true,
		},
		{
			name:        "parent cancelled after 200ms",
			cancelAfter: 200 * time.Millisecond,
			wantResult:  WatchdogCancelled,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Use a long-running command
			var cmd *exec.Cmd
			if runtime.GOOS == "windows" {
				cmd = exec.Command("timeout", "10")
			} else {
				cmd = exec.Command("sleep", "10")
			}

			cfg := WatchdogConfig{
				PollInterval: 50 * time.Millisecond,
				Timeout:      5 * time.Second,
				CleanupWait:  100 * time.Millisecond,
			}

			// Cancel parent context after delay
			go func() {
				time.Sleep(tt.cancelAfter)
				cancel()
			}()

			// Condition never becomes true
			checkCondition := func(ctx context.Context) bool {
				return false
			}

			start := time.Now()
			result, err := RunWithWatchdog(ctx, cmd, checkCondition, cfg)
			elapsed := time.Since(start)

			assert.Equal(t, tt.wantResult, result, "Result should be WatchdogCancelled")
			if tt.wantErr {
				assert.Error(t, err, "Should return context error")
				assert.ErrorIs(t, err, context.Canceled, "Error should be context.Canceled")
			} else {
				assert.NoError(t, err)
			}
			assert.GreaterOrEqual(t, elapsed, tt.cancelAfter, "Should wait at least the cancel duration")
			assert.Less(t, elapsed, tt.cancelAfter+500*time.Millisecond, "Should not wait much longer than cancel")
		})
	}
}

func TestRunWithWatchdog_CommandFailsBeforeCondition(t *testing.T) {
	tests := []struct {
		name       string
		command    []string
		wantResult WatchdogResult
		wantErr    bool
	}{
		{
			name:       "command with non-zero exit code",
			command:    []string{"sh", "-c", "exit 1"},
			wantResult: WatchdogCompleted,
			wantErr:    true,
		},
		{
			name:       "command not found",
			command:    []string{"nonexistent-command-12345"},
			wantResult: WatchdogCompleted,
			wantErr:    true,
		},
		{
			name:       "command with invalid flag",
			command:    []string{"ls", "--invalid-flag-xyz"},
			wantResult: WatchdogCompleted,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cmd := exec.Command(tt.command[0], tt.command[1:]...)

			cfg := WatchdogConfig{
				PollInterval: 50 * time.Millisecond,
				Timeout:      5 * time.Second,
				CleanupWait:  100 * time.Millisecond,
			}

			// Condition never becomes true
			checkCondition := func(ctx context.Context) bool {
				return false
			}

			result, err := RunWithWatchdog(ctx, cmd, checkCondition, cfg)

			assert.Equal(t, tt.wantResult, result, "Result should be WatchdogCompleted")
			if tt.wantErr {
				assert.Error(t, err, "Should return command error")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRunWithWatchdog_ZeroConfigUsesDefaults(t *testing.T) {
	ctx := context.Background()
	cmd := exec.Command("echo", "hello")

	// Pass zero config - should apply defaults
	cfg := WatchdogConfig{}

	checkCondition := func(ctx context.Context) bool {
		return false
	}

	result, err := RunWithWatchdog(ctx, cmd, checkCondition, cfg)

	assert.Equal(t, WatchdogCompleted, result, "Should complete normally")
	assert.NoError(t, err)
	// The fact that it completes without panic or timeout verifies defaults were applied
}

func TestRunWithWatchdog_CommandStartError(t *testing.T) {
	ctx := context.Background()

	// Create a command with invalid working directory
	cmd := exec.Command("echo", "hello")
	cmd.Dir = "/nonexistent/directory/that/does/not/exist"

	cfg := WatchdogConfig{
		PollInterval: 50 * time.Millisecond,
		Timeout:      5 * time.Second,
		CleanupWait:  100 * time.Millisecond,
	}

	checkCondition := func(ctx context.Context) bool {
		return false
	}

	result, err := RunWithWatchdog(ctx, cmd, checkCondition, cfg)

	// When command fails to start, should return WatchdogCompleted with error
	assert.Equal(t, WatchdogCompleted, result, "Result should be WatchdogCompleted when start fails")
	assert.Error(t, err, "Should return error from cmd.Start()")
}

func TestRunWithWatchdog_CleanupWaitBehavior(t *testing.T) {
	t.Run("cleanup wait allows process to exit cleanly", func(t *testing.T) {
		ctx := context.Background()

		// Use a long-running command
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("timeout", "10")
		} else {
			cmd = exec.Command("sleep", "10")
		}

		cfg := WatchdogConfig{
			PollInterval: 50 * time.Millisecond,
			Timeout:      200 * time.Millisecond,
			CleanupWait:  100 * time.Millisecond, // Give time for cleanup
		}

		checkCondition := func(ctx context.Context) bool {
			return false
		}

		start := time.Now()
		result, err := RunWithWatchdog(ctx, cmd, checkCondition, cfg)
		elapsed := time.Since(start)

		assert.Equal(t, WatchdogTimedOut, result)
		assert.NoError(t, err)

		// Should include cleanup wait time
		expectedMin := cfg.Timeout
		expectedMax := cfg.Timeout + cfg.CleanupWait + 200*time.Millisecond
		assert.GreaterOrEqual(t, elapsed, expectedMin, "Should wait at least timeout")
		assert.Less(t, elapsed, expectedMax, "Should not exceed timeout + cleanup + buffer")
	})
}

func TestRunWithWatchdog_ConditionCheckRespectsPollInterval(t *testing.T) {
	ctx := context.Background()

	// Use a command that completes quickly
	cmd := exec.Command("sh", "-c", "sleep 0.5")

	pollInterval := 100 * time.Millisecond
	cfg := WatchdogConfig{
		PollInterval: pollInterval,
		Timeout:      2 * time.Second,
		CleanupWait:  50 * time.Millisecond,
	}

	// Count how many times condition is checked
	checkCount := 0
	checkCondition := func(ctx context.Context) bool {
		checkCount++
		return false
	}

	result, err := RunWithWatchdog(ctx, cmd, checkCondition, cfg)

	assert.Equal(t, WatchdogCompleted, result)
	assert.NoError(t, err)

	// Should check condition approximately every 100ms for ~500ms
	// Expected: ~5 checks (500ms / 100ms), allow some variance
	assert.GreaterOrEqual(t, checkCount, 3, "Should poll multiple times")
	assert.LessOrEqual(t, checkCount, 8, "Should not poll excessively")
}

func TestWatchdogResult_Constants(t *testing.T) {
	// Verify the enum values are distinct
	results := []WatchdogResult{
		WatchdogCompleted,
		WatchdogDetected,
		WatchdogTimedOut,
		WatchdogCancelled,
	}

	// Check all values are unique
	seen := make(map[WatchdogResult]bool)
	for _, r := range results {
		require.False(t, seen[r], "WatchdogResult values should be unique")
		seen[r] = true
	}

	// Verify expected values (based on iota starting at 0)
	assert.Equal(t, WatchdogResult(0), WatchdogCompleted)
	assert.Equal(t, WatchdogResult(1), WatchdogDetected)
	assert.Equal(t, WatchdogResult(2), WatchdogTimedOut)
	assert.Equal(t, WatchdogResult(3), WatchdogCancelled)
}
