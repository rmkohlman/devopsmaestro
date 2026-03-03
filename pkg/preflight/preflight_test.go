package preflight

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPreflightRunner_AddCheck tests adding checks to runner
func TestPreflightRunner_AddCheck(t *testing.T) {
	runner := NewPreflightRunner()

	check := &MockCheck{
		NameValue: "test-check",
		RunFunc: func(ctx context.Context) CheckResult {
			return CheckResult{Status: StatusOK, Message: "OK"}
		},
	}

	runner.AddCheck(check)

	// Verify check was added
	checks := runner.GetChecks()
	assert.Len(t, checks, 1)
	assert.Equal(t, "test-check", checks[0].Name())
}

// TestPreflightRunner_Run tests executing all checks
func TestPreflightRunner_Run(t *testing.T) {
	runner := NewPreflightRunner()

	// Add multiple checks
	check1 := &MockCheck{
		NameValue: "check1",
		RunFunc: func(ctx context.Context) CheckResult {
			return CheckResult{Status: StatusOK, Message: "Check 1 passed"}
		},
	}
	check2 := &MockCheck{
		NameValue: "check2",
		RunFunc: func(ctx context.Context) CheckResult {
			return CheckResult{Status: StatusOK, Message: "Check 2 passed"}
		},
	}

	runner.AddCheck(check1)
	runner.AddCheck(check2)

	// Run all checks
	results := runner.Run(context.Background())

	assert.Len(t, results, 2)
	assert.Equal(t, StatusOK, results[0].Status)
	assert.Equal(t, StatusOK, results[1].Status)
}

// TestPreflightRunner_HasErrors tests detecting failed checks
func TestPreflightRunner_HasErrors(t *testing.T) {
	tests := []struct {
		name       string
		checks     []Check
		wantErrors bool
	}{
		{
			name: "all checks pass",
			checks: []Check{
				&MockCheck{
					NameValue: "check1",
					RunFunc: func(ctx context.Context) CheckResult {
						return CheckResult{Status: StatusOK, Message: "OK"}
					},
				},
				&MockCheck{
					NameValue: "check2",
					RunFunc: func(ctx context.Context) CheckResult {
						return CheckResult{Status: StatusOK, Message: "OK"}
					},
				},
			},
			wantErrors: false,
		},
		{
			name: "one check fails",
			checks: []Check{
				&MockCheck{
					NameValue: "check1",
					RunFunc: func(ctx context.Context) CheckResult {
						return CheckResult{Status: StatusOK, Message: "OK"}
					},
				},
				&MockCheck{
					NameValue: "check2",
					RunFunc: func(ctx context.Context) CheckResult {
						return CheckResult{Status: StatusError, Message: "Failed"}
					},
				},
			},
			wantErrors: true,
		},
		{
			name: "warnings do not count as errors",
			checks: []Check{
				&MockCheck{
					NameValue: "check1",
					RunFunc: func(ctx context.Context) CheckResult {
						return CheckResult{Status: StatusWarning, Message: "Warning"}
					},
				},
			},
			wantErrors: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := NewPreflightRunner()
			for _, check := range tt.checks {
				runner.AddCheck(check)
			}

			results := runner.Run(context.Background())
			hasErrors := runner.HasErrors(results)

			assert.Equal(t, tt.wantErrors, hasErrors)
		})
	}
}

// TestPreflightRunner_HasWarnings tests detecting warning checks
func TestPreflightRunner_HasWarnings(t *testing.T) {
	tests := []struct {
		name         string
		checks       []Check
		wantWarnings bool
	}{
		{
			name: "no warnings",
			checks: []Check{
				&MockCheck{
					NameValue: "check1",
					RunFunc: func(ctx context.Context) CheckResult {
						return CheckResult{Status: StatusOK, Message: "OK"}
					},
				},
			},
			wantWarnings: false,
		},
		{
			name: "has warning",
			checks: []Check{
				&MockCheck{
					NameValue: "check1",
					RunFunc: func(ctx context.Context) CheckResult {
						return CheckResult{Status: StatusWarning, Message: "Warning"}
					},
				},
			},
			wantWarnings: true,
		},
		{
			name: "errors do not count as warnings",
			checks: []Check{
				&MockCheck{
					NameValue: "check1",
					RunFunc: func(ctx context.Context) CheckResult {
						return CheckResult{Status: StatusError, Message: "Error"}
					},
				},
			},
			wantWarnings: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := NewPreflightRunner()
			for _, check := range tt.checks {
				runner.AddCheck(check)
			}

			results := runner.Run(context.Background())
			hasWarnings := runner.HasWarnings(results)

			assert.Equal(t, tt.wantWarnings, hasWarnings)
		})
	}
}

// TestPreflightRunner_SkippedChecks tests handling skipped checks
func TestPreflightRunner_SkippedChecks(t *testing.T) {
	runner := NewPreflightRunner()

	check := &MockCheck{
		NameValue: "skipped-check",
		RunFunc: func(ctx context.Context) CheckResult {
			return CheckResult{Status: StatusSkipped, Message: "Skipped due to condition"}
		},
	}

	runner.AddCheck(check)
	results := runner.Run(context.Background())

	assert.Len(t, results, 1)
	assert.Equal(t, StatusSkipped, results[0].Status)

	// Skipped checks should not count as errors
	assert.False(t, runner.HasErrors(results))
}

// TestPreflightRunner_CheckTimeout tests check timeout handling
func TestPreflightRunner_CheckTimeout(t *testing.T) {
	runner := NewPreflightRunner()
	runner.SetTimeout(100) // 100ms timeout

	slowCheck := &MockCheck{
		NameValue: "slow-check",
		RunFunc: func(ctx context.Context) CheckResult {
			// Simulate slow check
			select {
			case <-ctx.Done():
				return CheckResult{Status: StatusError, Message: "Timeout"}
			}
		},
	}

	runner.AddCheck(slowCheck)
	results := runner.Run(context.Background())

	// Should timeout and return error
	require.Len(t, results, 1)
	assert.Equal(t, StatusError, results[0].Status)
	assert.Contains(t, results[0].Message, "timeout", "Should indicate timeout")
}

// TestPreflightRunner_ParallelExecution tests parallel check execution
func TestPreflightRunner_ParallelExecution(t *testing.T) {
	runner := NewPreflightRunner()
	runner.SetParallel(true)

	// Add multiple checks
	for i := 0; i < 5; i++ {
		check := &MockCheck{
			NameValue: "check-" + string(rune(i)),
			RunFunc: func(ctx context.Context) CheckResult {
				return CheckResult{Status: StatusOK, Message: "OK"}
			},
		}
		runner.AddCheck(check)
	}

	results := runner.Run(context.Background())

	// All checks should complete
	assert.Len(t, results, 5)
	for _, result := range results {
		assert.Equal(t, StatusOK, result.Status)
	}
}

// TestPreflightRunner_EmptyRunner tests runner with no checks
func TestPreflightRunner_EmptyRunner(t *testing.T) {
	runner := NewPreflightRunner()

	results := runner.Run(context.Background())

	assert.Empty(t, results, "Runner with no checks should return empty results")
	assert.False(t, runner.HasErrors(results))
	assert.False(t, runner.HasWarnings(results))
}

// TestPreflightRunner_CheckOrder tests checks run in order
func TestPreflightRunner_CheckOrder(t *testing.T) {
	runner := NewPreflightRunner()
	runner.SetParallel(false) // Ensure sequential execution

	var order []string

	check1 := &MockCheck{
		NameValue: "check1",
		RunFunc: func(ctx context.Context) CheckResult {
			order = append(order, "check1")
			return CheckResult{Status: StatusOK, Message: "OK"}
		},
	}
	check2 := &MockCheck{
		NameValue: "check2",
		RunFunc: func(ctx context.Context) CheckResult {
			order = append(order, "check2")
			return CheckResult{Status: StatusOK, Message: "OK"}
		},
	}

	runner.AddCheck(check1)
	runner.AddCheck(check2)

	runner.Run(context.Background())

	// Checks should run in order
	assert.Equal(t, []string{"check1", "check2"}, order)
}

// TestPreflightRunner_ContextCancellation tests context cancellation
func TestPreflightRunner_ContextCancellation(t *testing.T) {
	runner := NewPreflightRunner()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	check := &MockCheck{
		NameValue: "check",
		RunFunc: func(ctx context.Context) CheckResult {
			if ctx.Err() != nil {
				return CheckResult{Status: StatusError, Message: "Context cancelled"}
			}
			return CheckResult{Status: StatusOK, Message: "OK"}
		},
	}

	runner.AddCheck(check)
	results := runner.Run(ctx)

	// Check should detect cancellation
	require.Len(t, results, 1)
	assert.Equal(t, StatusError, results[0].Status)
}

// TestCheckResult_IsSuccess tests result success detection
func TestCheckResult_IsSuccess(t *testing.T) {
	tests := []struct {
		name    string
		status  CheckStatus
		success bool
	}{
		{"OK is success", StatusOK, true},
		{"Warning is success", StatusWarning, true},
		{"Skipped is success", StatusSkipped, true},
		{"Error is not success", StatusError, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckResult{Status: tt.status}
			assert.Equal(t, tt.success, result.IsSuccess())
		})
	}
}

// MockCheck implements Check interface for testing
type MockCheck struct {
	NameValue string
	RunFunc   func(ctx context.Context) CheckResult
}

func (m *MockCheck) Name() string {
	return m.NameValue
}

func (m *MockCheck) Run(ctx context.Context) CheckResult {
	if m.RunFunc != nil {
		return m.RunFunc(ctx)
	}
	return CheckResult{Status: StatusOK, Message: "OK"}
}
