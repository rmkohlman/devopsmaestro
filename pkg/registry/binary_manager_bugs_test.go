package registry

// =============================================================================
// BUG #1 (P0): Binary download has no HTTP timeout
//
// Root cause: downloadBinary() passes the caller's context verbatim to
// http.NewRequestWithContext(). When the caller provides context.Background()
// (which has no deadline), a stalled server will hang the download forever.
//
// Expected fix: inside downloadBinary(), wrap the caller's context with
//   context.WithTimeout(ctx, 5*time.Minute) when ctx has no deadline.
//
// Test strategy: inject a custom http.RoundTripper (deadlineCapturingTransport)
// into http.DefaultTransport. Call downloadBinary() with context.Background()
// (no deadline). Assert that the outgoing request's context HAS a deadline.
//
//   RED  today : transport.deadlineFound == false → test FAILS.
//   GREEN after: transport.deadlineFound == true  → test PASSES.
// =============================================================================

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// deadlineCapturingTransport is a test http.RoundTripper that records whether
// the outgoing request's context has a deadline, then immediately returns
// context.Canceled so the call completes quickly.
type deadlineCapturingTransport struct {
	deadlineFound bool
	deadline      time.Time
}

func (d *deadlineCapturingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	dl, ok := req.Context().Deadline()
	d.deadlineFound = ok
	d.deadline = dl
	// Return immediately — the specific error doesn't matter for this assertion.
	return nil, context.Canceled
}

// TestDownloadBinary_AppliesDefensiveTimeout verifies that downloadBinary()
// applies a defensive timeout to the HTTP request even when the caller's
// context has no deadline (context.Background()).
//
// BUG #1 (P0): Currently downloadBinary uses the caller's context verbatim.
// When the caller passes context.Background() (no deadline), the request
// context also has no deadline → the download can hang forever.
//
// RED  today : capturingTransport.deadlineFound == false → test FAILS.
// GREEN after: capturingTransport.deadlineFound == true  → test PASSES.
func TestDownloadBinary_AppliesDefensiveTimeout(t *testing.T) {
	// Arrange: inject transport that captures deadline info and returns immediately.
	transport := &deadlineCapturingTransport{}
	origTransport := http.DefaultTransport
	http.DefaultTransport = transport
	t.Cleanup(func() { http.DefaultTransport = origTransport })

	binDir := t.TempDir()
	bm := &DefaultBinaryManager{binDir: binDir, version: "99.99.99"}
	destPath := t.TempDir() + "/zot"

	// Act: call downloadBinary with context.Background() — no deadline.
	ctx := context.Background()
	_, _ = bm.downloadBinary(ctx, destPath) // error expected (transport cancels)

	// Assert: the request context passed to the HTTP layer MUST have a deadline.
	// BUG present : deadlineFound == false → this assertion FAILS  (RED).
	// BUG fixed   : deadlineFound == true  → this assertion PASSES (GREEN).
	assert.True(t, transport.deadlineFound,
		"BUG #1: downloadBinary must apply context.WithTimeout(ctx, 5*time.Minute) "+
			"when the caller provides context.Background() (no deadline). "+
			"A stalled server will otherwise hang the download forever.")
}

// TestDownloadBinary_RespectsExistingDeadline is the companion "don't regress"
// test: once the fix is applied, a caller-supplied short deadline must still
// be honoured (not replaced with a longer 5-minute one).
func TestDownloadBinary_RespectsExistingDeadline(t *testing.T) {
	transport := &deadlineCapturingTransport{}
	origTransport := http.DefaultTransport
	http.DefaultTransport = transport
	t.Cleanup(func() { http.DefaultTransport = origTransport })

	// Use a real test server so the URL is valid (transport short-circuits anyway).
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	binDir := t.TempDir()
	bm := &DefaultBinaryManager{binDir: binDir, version: "99.99.99"}
	destPath := t.TempDir() + "/zot"

	// Caller sets a short (1-second) deadline.
	shortCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	expectedDeadline, _ := shortCtx.Deadline()

	_, _ = bm.downloadBinary(shortCtx, destPath)

	require.True(t, transport.deadlineFound,
		"request context should have a deadline when caller provided one")

	// The deadline used must NOT be later than the caller's deadline.
	// The fix must prefer the shorter of the two deadlines.
	assert.False(t, transport.deadline.After(expectedDeadline),
		"downloadBinary must not override a short caller deadline with a longer 5-minute one; "+
			"caller deadline: %v, request deadline: %v",
		expectedDeadline, transport.deadline)
}
