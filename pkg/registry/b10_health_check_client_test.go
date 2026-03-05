// Package registry — TDD RED phase: B10 health-check client timeout.
//
// waitForReady() currently uses http.Get() directly (http.DefaultClient, no
// timeout, context not forwarded to the individual HTTP request).
//
// The fix must introduce a package-level variable:
//
//	var healthCheckClient = &http.Client{Timeout: 2 * time.Second}
//
// This file requires the build tag "b10" to include it, so the normal test
// suite compiles while this file remains as a permanent RED reminder.
// Run with: go test -tags b10 ./pkg/registry/...
//
// Once the healthCheckClient variable is added to the package, remove the
// build tag — the tests will then fail at runtime until the timeout is correct.

//go:build b10

package registry

import (
	"net/http"
	"testing"
	"time"
)

// TestHealthCheckClient_HasTimeout verifies the package-level healthCheckClient
// has a 2-second timeout so that hung health-check requests cannot block Start()
// indefinitely.
func TestHealthCheckClient_HasTimeout(t *testing.T) {
	// This test will fail to compile until healthCheckClient is created in the
	// registry package.  That is intentional: compilation failure = RED.
	if healthCheckClient.Timeout != 2*time.Second {
		t.Errorf("healthCheckClient.Timeout = %v, want 2s", healthCheckClient.Timeout)
	}
}

// TestHealthCheckClient_IsNotDefaultClient verifies that a dedicated client is
// used rather than http.DefaultClient (which has no timeout and is shared
// across all HTTP calls in the process).
func TestHealthCheckClient_IsNotDefaultClient(t *testing.T) {
	if healthCheckClient == http.DefaultClient {
		t.Error("healthCheckClient must not be http.DefaultClient (no timeout on DefaultClient)")
	}
}

// TestHealthCheckClient_TimeoutIsNonZero verifies the timeout is explicitly set
// to a non-zero value.  A zero Timeout means no timeout at all.
func TestHealthCheckClient_TimeoutIsNonZero(t *testing.T) {
	if healthCheckClient.Timeout == 0 {
		t.Error("healthCheckClient.Timeout must be non-zero; zero means no timeout (DoS risk)")
	}
}
