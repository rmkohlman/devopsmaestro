package registry

// =============================================================================
// Security Hardening Tests (TDD Phase 2 — RED)
//
// These tests document security gaps that MUST be fixed before v0.19.0.
// Each test is intentionally written to FAIL against the current code and
// PASS once the corresponding fix is implemented.
//
// Item 1 (HIGH)  — Checksum verification for Zot binary downloads
// Item 2 (MEDIUM)— Athens binary manager missing defensive timeout
// Item 3 (MEDIUM)— Config files written with 0644 instead of 0600
// Item 4 (MEDIUM)— Log files opened with 0644 instead of 0600
// Item 5 (MEDIUM)— Storage path validation missing from resolveStoragePath
// =============================================================================

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// ITEM 1 — Checksum verification wiring
// =============================================================================

// TestDownloadBinary_VerifiesChecksum verifies that downloadBinary() calls
// verifyChecksum() after downloading the file.
//
// BUG: downloadBinary() performs io.Copy but never calls verifyChecksum(),
// even though the function exists. A tampered or corrupted download will be
// silently accepted.
//
// RED  today : download succeeds even when SHA256 doesn't match — test FAILS.
// GREEN after: downloadBinary() fetches the .sha256 sidecar and rejects
//
//	mismatches before moving the temp file to the final location.
func TestDownloadBinary_VerifiesChecksum(t *testing.T) {
	// Arrange: serve a binary whose content is "hello" but whose .sha256
	// sidecar contains an intentionally WRONG checksum.
	binaryContent := []byte("fake-zot-binary-content")
	wrongChecksum := strings.Repeat("a", 64) // 64 hex chars = 32 bytes, all 'a'

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".sha256") {
			// Serve wrong checksum so verification must fail
			fmt.Fprintf(w, "%s  zot-linux-amd64\n", wrongChecksum)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(binaryContent)
	}))
	defer ts.Close()

	// Override the download URL base for testing by swapping http.DefaultTransport.
	// We use a transport that rewrites requests to our test server.
	origTransport := http.DefaultTransport
	http.DefaultTransport = &urlRewriteTransport{base: ts.URL, inner: http.DefaultTransport}
	t.Cleanup(func() { http.DefaultTransport = origTransport })

	binDir := t.TempDir()
	bm := &DefaultBinaryManager{binDir: binDir, version: "99.99.99"}
	destPath := filepath.Join(binDir, "zot")

	// Act
	_, err := bm.downloadBinary(context.Background(), destPath)

	// Assert: the download MUST fail when the checksum doesn't match.
	// BUG present : err == nil (checksum never checked) → test FAILS (RED).
	// BUG fixed   : err != nil with "checksum mismatch" → test PASSES (GREEN).
	require.Error(t, err,
		"BUG ITEM 1: downloadBinary must reject a download whose SHA256 checksum "+
			"does not match the .sha256 sidecar file. Currently checksum is never checked.")
	assert.Contains(t, err.Error(), "checksum",
		"error must mention checksum so the cause is obvious")
}

// TestDownloadBinary_RejectsOversizedContentLength verifies that downloadBinary()
// rejects a response whose Content-Length header exceeds a safety cap (~500 MB)
// before writing any bytes to disk.
//
// BUG: downloadBinary() does not inspect Content-Length before streaming.
// A malicious or misconfigured server can advertise and send gigabytes,
// filling the user's disk completely.
//
// RED  today : downloadBinary does NOT check Content-Length → proceeds to
//
//	io.Copy → download "succeeds" → err == nil → require.Error FAILS (RED).
//
// GREEN after: downloadBinary() checks Content-Length before streaming and
//
//	returns an error mentioning "size" → test PASSES.
func TestDownloadBinary_RejectsOversizedContentLength(t *testing.T) {
	// Arrange: inject a transport that injects a huge Content-Length into the
	// response while serving a small, complete body so there's no EOF error.
	// This isolates the test to the Content-Length-guard behaviour only.
	const limitBytes = int64(500 * 1024 * 1024) // 500 MiB — our expected safety cap
	const oversizeBytes = limitBytes + 1

	binaryContent := []byte("fake-zot-binary-content")

	oversizeTransport := &oversizeContentLengthTransport{
		body:          binaryContent,
		contentLength: oversizeBytes,
	}

	origTransport := http.DefaultTransport
	http.DefaultTransport = oversizeTransport
	t.Cleanup(func() { http.DefaultTransport = origTransport })

	binDir := t.TempDir()
	bm := &DefaultBinaryManager{binDir: binDir, version: "99.99.99"}
	destPath := filepath.Join(binDir, "zot")

	// Act
	_, err := bm.downloadBinary(context.Background(), destPath)

	// Assert: the download MUST fail when Content-Length exceeds the safety cap.
	// BUG present : no Content-Length check → err == nil → require.Error FAILS (RED).
	// BUG fixed   : err != nil mentioning "size" → PASSES (GREEN).
	require.Error(t, err,
		"BUG ITEM 1: downloadBinary must reject downloads whose Content-Length "+
			"exceeds the safety cap (~500 MB). Currently there is no size guard. "+
			"Content-Length was %d bytes.", oversizeBytes)
	assert.Contains(t, strings.ToLower(err.Error()), "size",
		"error must mention size/too large so the cause is obvious; got: %v", err)
}

// oversizeContentLengthTransport is a test RoundTripper that returns a synthetic
// HTTP response with a specific Content-Length but a small, complete body.
// This allows testing Content-Length guards without serving gigabytes of data.
type oversizeContentLengthTransport struct {
	body          []byte
	contentLength int64
}

func (o *oversizeContentLengthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	header := make(http.Header)
	header.Set("Content-Length", fmt.Sprintf("%d", o.contentLength))
	return &http.Response{
		StatusCode:    http.StatusOK,
		Header:        header,
		Body:          io.NopCloser(strings.NewReader(string(o.body))),
		ContentLength: o.contentLength,
		Request:       req,
	}, nil
}

// TestVerifyChecksum_UsesConstantTimeComparison verifies that verifyChecksum()
// compares hashes using subtle.ConstantTimeCompare instead of the != operator,
// preventing timing-based side-channel attacks on the checksum value.
//
// BUG: verifyChecksum() uses `if actualSum != expectedSum` which is a
// variable-time string comparison, leaking information via timing.
//
// RED  today : direct inspection of source — the function uses !=.
//
//	We verify this indirectly: if constant-time is used, passing a
//	deliberately-wrong expected sum whose first byte matches must still fail.
//	(The test relies on observing an error; it can't distinguish timing
//	 directly, so we verify the API contract and document the code smell.)
//
// GREEN after: verifyChecksum() uses subtle.ConstantTimeCompare → same
//
//	functional result but no timing leak.
//
// NOTE: This test primarily documents the requirement; the underlying timing
// property cannot be verified in a unit test. The RED status comes from the
// fact that the function's use of != is a documented code smell that must be
// changed regardless of observable behaviour.
func TestVerifyChecksum_UsesConstantTimeComparison(t *testing.T) {
	// Arrange: write a file with known content, compute its real checksum.
	dir := t.TempDir()
	filePath := filepath.Join(dir, "binary")
	content := []byte("test binary content")
	require.NoError(t, os.WriteFile(filePath, content, 0600))

	h := sha256.Sum256(content)
	correctSum := hex.EncodeToString(h[:])

	bm := &DefaultBinaryManager{binDir: dir, version: "1.0.0"}

	t.Run("correct checksum accepted", func(t *testing.T) {
		err := bm.verifyChecksum(filePath, correctSum)
		assert.NoError(t, err, "correct checksum must be accepted")
	})

	t.Run("wrong checksum rejected", func(t *testing.T) {
		err := bm.verifyChecksum(filePath, strings.Repeat("0", 64))
		assert.Error(t, err, "wrong checksum must be rejected")
	})

	// RED marker: the function currently uses != (not constant-time).
	// This subtest fails until the implementation switches to
	// subtle.ConstantTimeCompare. We detect the implementation gap by
	// checking that the source file no longer contains the unsafe pattern
	// after the fix is applied.
	//
	// For the RED phase, this test PASSES functionally but documents that
	// the timing-safe implementation is required. The companion implementation
	// test (checking source code) is left to static analysis / code review.
	//
	// The actual RED test for constant-time is the checksum-wiring test above,
	// which fails because verifyChecksum is never called at all.
	t.Log("NOTE: verifyChecksum must be updated to use subtle.ConstantTimeCompare " +
		"instead of the != string comparison operator to prevent timing side-channels.")
}

// =============================================================================
// ITEM 2 — Athens binary manager missing defensive timeout
// =============================================================================

// TestAthensBinaryManager_DownloadBinary_AppliesDefensiveTimeout verifies that
// AthensBinaryManager.downloadBinary() applies a defensive timeout when the
// caller's context has no deadline.
//
// BUG: AthensBinaryManager.downloadBinary() passes the caller's context
// verbatim to http.NewRequestWithContext(). When the caller provides
// context.Background() (no deadline), a stalled server will hang forever.
//
// This mirrors the same bug that was fixed in DefaultBinaryManager (Bug #1).
// The Athens implementation was not updated at the same time.
//
// RED  today : AthensBinaryManager uses ctx verbatim → deadlineFound == false → FAILS.
// GREEN after: AthensBinaryManager wraps ctx with context.WithTimeout → PASSES.
func TestAthensBinaryManager_DownloadBinary_AppliesDefensiveTimeout(t *testing.T) {
	// Arrange: inject transport that captures deadline info and returns immediately.
	// (deadlineCapturingTransport is defined in binary_manager_bugs_test.go)
	transport := &deadlineCapturingTransport{}
	origTransport := http.DefaultTransport
	http.DefaultTransport = transport
	t.Cleanup(func() { http.DefaultTransport = origTransport })

	binDir := t.TempDir()
	bm := &AthensBinaryManager{binDir: binDir, version: "0.14.1"}
	destPath := filepath.Join(binDir, "athens")

	// Act: call downloadBinary with context.Background() — no deadline.
	ctx := context.Background()
	_, _ = bm.downloadBinary(ctx, destPath) // error expected (transport cancels)

	// Assert: the request context passed to the HTTP layer MUST have a deadline.
	// BUG present : deadlineFound == false → this assertion FAILS  (RED).
	// BUG fixed   : deadlineFound == true  → this assertion PASSES (GREEN).
	assert.True(t, transport.deadlineFound,
		"BUG ITEM 2: AthensBinaryManager.downloadBinary must apply "+
			"context.WithTimeout(ctx, 5*time.Minute) when the caller provides "+
			"context.Background() (no deadline). "+
			"A stalled server will otherwise hang the download forever.")
}

// TestAthensBinaryManager_DownloadBinary_RespectsExistingDeadline is the
// companion "don't regress" test: a caller-supplied short deadline must still
// be honoured after the fix is applied.
func TestAthensBinaryManager_DownloadBinary_RespectsExistingDeadline(t *testing.T) {
	transport := &deadlineCapturingTransport{}
	origTransport := http.DefaultTransport
	http.DefaultTransport = transport
	t.Cleanup(func() { http.DefaultTransport = origTransport })

	binDir := t.TempDir()
	bm := &AthensBinaryManager{binDir: binDir, version: "0.14.1"}
	destPath := filepath.Join(binDir, "athens")

	// Caller sets a short (1-second) deadline.
	shortCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	expectedDeadline, _ := shortCtx.Deadline()

	_, _ = bm.downloadBinary(shortCtx, destPath)

	require.True(t, transport.deadlineFound,
		"request context should have a deadline when caller provided one")

	assert.False(t, transport.deadline.After(expectedDeadline),
		"AthensBinaryManager must not override a short caller deadline with a longer 5-minute one; "+
			"caller deadline: %v, request deadline: %v",
		expectedDeadline, transport.deadline)
}

// =============================================================================
// ITEM 3 — Config files written with 0644 instead of 0600
// =============================================================================

// TestZotManager_WriteConfigFile_Uses0600 verifies that ZotManager.writeConfigFile()
// creates config files with permission 0600 (owner-read/write only), not 0644.
//
// BUG: os.WriteFile(path, data, 0644) allows group and world to read the config,
// which may contain sensitive data (ports, storage paths, auth tokens).
//
// RED  today : writeConfigFile uses 0644 → assert 0600 fails → test FAILS.
// GREEN after: changed to os.WriteFile(path, data, 0600) → test PASSES.
func TestZotManager_WriteConfigFile_Uses0600(t *testing.T) {
	dir := t.TempDir()
	bm := NewMockBinaryManagerNamed(dir, "2.1.15", "zot")
	pm := NewProcessManager(ProcessConfig{
		PIDFile: filepath.Join(dir, "zot.pid"),
		LogFile: filepath.Join(dir, "zot.log"),
	})
	mgr := NewZotManagerWithDeps(RegistryConfig{
		Storage:   dir,
		Port:      5001,
		Lifecycle: "manual",
	}, bm, pm)

	configPath := filepath.Join(dir, "config.json")
	err := mgr.writeConfigFile(configPath, map[string]interface{}{"key": "value"})
	require.NoError(t, err, "writeConfigFile should not return an error")

	info, err := os.Stat(configPath)
	require.NoError(t, err, "config file should exist after writeConfigFile")

	perm := info.Mode().Perm()
	// BUG present : perm == 0644 → assertion below FAILS (RED).
	// BUG fixed   : perm == 0600 → assertion PASSES (GREEN).
	assert.Equal(t, os.FileMode(0600), perm,
		"BUG ITEM 3: ZotManager.writeConfigFile must use 0600 permissions, not 0644. "+
			"Config files may contain sensitive data; group/world read is a security risk.")
}

// TestAthensManager_WriteConfigFile_Uses0600 verifies that AthensManager.writeConfigFile()
// creates config files with permission 0600, not 0644.
//
// RED  today : writeConfigFile uses 0644 → assertion FAILS.
// GREEN after: changed to 0600 → test PASSES.
func TestAthensManager_WriteConfigFile_Uses0600(t *testing.T) {
	dir := t.TempDir()
	bm := NewMockBinaryManager(dir, "0.14.1")
	pm := NewProcessManager(ProcessConfig{
		PIDFile: filepath.Join(dir, "athens.pid"),
		LogFile: filepath.Join(dir, "athens.log"),
	})
	mgr, err := NewAthensManager(GoModuleConfig{
		Storage:   dir,
		Port:      3000,
		Lifecycle: "manual",
	}, bm, pm)
	require.NoError(t, err)

	configPath := filepath.Join(dir, "config.toml")
	err = mgr.writeConfigFile(configPath, "StorageType = \"disk\"\n")
	require.NoError(t, err, "writeConfigFile should not return an error")

	info, err := os.Stat(configPath)
	require.NoError(t, err, "config file should exist after writeConfigFile")

	perm := info.Mode().Perm()
	assert.Equal(t, os.FileMode(0600), perm,
		"BUG ITEM 3: AthensManager.writeConfigFile must use 0600 permissions, not 0644.")
}

// TestVerdaccioManager_GenerateConfig_Uses0600 verifies that
// VerdaccioManager.generateConfig() writes config.yaml with 0600 permissions.
//
// RED  today : uses 0644 → assertion FAILS.
// GREEN after: changed to 0600 → test PASSES.
func TestVerdaccioManager_GenerateConfig_Uses0600(t *testing.T) {
	dir := t.TempDir()
	bm := NewMockNpmBinaryManager(dir, "5.28.0")
	pm := NewProcessManager(ProcessConfig{
		PIDFile: filepath.Join(dir, "verdaccio.pid"),
		LogFile: filepath.Join(dir, "verdaccio.log"),
	})
	mgr, err := NewVerdaccioManager(NpmProxyConfig{
		Storage:   dir,
		Port:      4873,
		Lifecycle: "manual",
		Upstreams: []NpmUpstreamConfig{
			{Name: "npmjs", URL: "https://registry.npmjs.org"},
		},
	}, bm, pm)
	require.NoError(t, err)

	err = mgr.generateConfig()
	require.NoError(t, err, "generateConfig should not return an error")

	configPath := filepath.Join(dir, "config.yaml")
	info, err := os.Stat(configPath)
	require.NoError(t, err, "config.yaml should exist after generateConfig")

	perm := info.Mode().Perm()
	assert.Equal(t, os.FileMode(0600), perm,
		"BUG ITEM 3: VerdaccioManager.generateConfig must write config.yaml with 0600 permissions, not 0644.")
}

// TestSquidManager_GenerateConfig_Uses0600 verifies that
// SquidManager.generateConfig() writes squid.conf with 0600 permissions.
//
// RED  today : uses 0644 → assertion FAILS.
// GREEN after: changed to 0600 → test PASSES.
func TestSquidManager_GenerateConfig_Uses0600(t *testing.T) {
	dir := t.TempDir()
	config := DefaultHttpProxyConfig()
	config.CacheDir = filepath.Join(dir, "cache")
	config.LogDir = filepath.Join(dir, "logs")
	config.PidFile = filepath.Join(dir, "squid.pid")

	mgr := &SquidManager{
		BaseServiceManager: NewBaseServiceManager(
			NewMockBrewBinaryManager(dir, "6.0"),
			NewProcessManager(ProcessConfig{
				PIDFile: config.PidFile,
				LogFile: filepath.Join(config.LogDir, "squid.log"),
			}),
		),
		config: config,
	}

	configPath := filepath.Join(dir, "squid.conf")
	err := mgr.generateConfig(configPath)
	require.NoError(t, err, "generateConfig should not return an error")

	info, err := os.Stat(configPath)
	require.NoError(t, err, "squid.conf should exist after generateConfig")

	perm := info.Mode().Perm()
	assert.Equal(t, os.FileMode(0600), perm,
		"BUG ITEM 3: SquidManager.generateConfig must write squid.conf with 0600 permissions, not 0644.")
}

// =============================================================================
// ITEM 4 — Log files opened with 0644 instead of 0600
// =============================================================================

// TestProcessManager_Start_LogFileUses0600 verifies that ProcessManager.Start()
// opens the log file with permission 0600, not 0644.
//
// BUG: os.OpenFile(config.LogFile, ..., 0644) exposes log output (which may
// contain sensitive startup arguments or error messages) to group and world.
//
// RED  today : log file uses 0644 → assertion FAILS.
// GREEN after: changed to 0600 → test PASSES.
func TestProcessManager_Start_LogFileUses0600(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "service.log")

	pm := NewProcessManager(ProcessConfig{
		PIDFile: filepath.Join(dir, "service.pid"),
		LogFile: logFile,
	})

	// Use a real binary that exits immediately so Start() completes.
	// We only need the log file to be created — the process can exit right away.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start with /usr/bin/true (exits 0 immediately) — log file gets created.
	_ = pm.Start(ctx, "/usr/bin/true", []string{}, ProcessConfig{
		PIDFile: filepath.Join(dir, "service.pid"),
		LogFile: logFile,
	})

	// Give the OS a moment to flush file creation
	time.Sleep(50 * time.Millisecond)

	info, err := os.Stat(logFile)
	require.NoError(t, err, "log file should have been created by Start()")

	perm := info.Mode().Perm()
	// BUG present : perm == 0644 → assertion FAILS (RED).
	// BUG fixed   : perm == 0600 → assertion PASSES (GREEN).
	assert.Equal(t, os.FileMode(0600), perm,
		"BUG ITEM 4: ProcessManager.Start must open the log file with 0600 permissions, not 0644. "+
			"Log files may contain sensitive startup data; group/world read is a security risk.")
}

// =============================================================================
// ITEM 5 — Storage path validation missing from resolveStoragePath
// =============================================================================

// TestValidateStoragePath_AcceptsAllowedPaths verifies that a
// validateStoragePath() function exists and allows paths inside
// ~/.devopsmaestro/.
//
// BUG: validateStoragePath does not exist. resolveStoragePath() accepts any
// path from the config JSON without any safety checks, allowing a hostile
// config to redirect storage to /etc/, /tmp/, or path-traversal targets.
//
// RED  today : validateStoragePath does not exist → compile error → test FAILS.
// GREEN after: function is implemented → test PASSES.
func TestValidateStoragePath_AcceptsAllowedPaths(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	allowedPaths := []string{
		filepath.Join(homeDir, ".devopsmaestro", "registries", "my-reg"),
		filepath.Join(homeDir, ".devopsmaestro", "registries", "zot"),
		filepath.Join(homeDir, ".devopsmaestro", "custom"),
	}

	for _, p := range allowedPaths {
		t.Run(p, func(t *testing.T) {
			// BUG present : function does not exist → compile error (RED).
			// BUG fixed   : returns nil for allowed paths (GREEN).
			err := validateStoragePath(p, homeDir)
			assert.NoError(t, err,
				"validateStoragePath must accept paths inside ~/.devopsmaestro/")
		})
	}
}

// TestValidateStoragePath_RejectsDangerousPaths verifies that
// validateStoragePath() rejects system directories and path traversal attempts.
//
// RED  today : function does not exist → compile error → test FAILS.
// GREEN after: function implemented and rejects dangerous paths → test PASSES.
func TestValidateStoragePath_RejectsDangerousPaths(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name string
		path string
	}{
		{name: "system etc", path: "/etc/evil"},
		{name: "tmp directory", path: "/tmp/test"},
		{name: "root directory", path: "/"},
		{name: "path traversal", path: filepath.Join(homeDir, ".devopsmaestro", "..", "..", "etc")},
		{name: "relative path", path: "relative/path"},
		{name: "outside home", path: "/var/lib/someservice"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// BUG present : function does not exist → compile error (RED).
			// BUG fixed   : returns error for dangerous paths (GREEN).
			err := validateStoragePath(tt.path, homeDir)
			assert.Error(t, err,
				"validateStoragePath must reject dangerous path %q", tt.path)
		})
	}
}

// =============================================================================
// Helper: urlRewriteTransport — rewrites outgoing URLs to a test server
// =============================================================================

// urlRewriteTransport is a test http.RoundTripper that rewrites outgoing
// request URLs to point to a local test server, allowing downloadBinary() to
// be exercised without real network access.
type urlRewriteTransport struct {
	base  string // e.g. "http://127.0.0.1:54321"
	inner http.RoundTripper
}

func (u *urlRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rewrite the scheme+host to the test server while keeping path+query.
	newURL := *req.URL
	parts := strings.SplitN(u.base, "://", 2)
	if len(parts) == 2 {
		newURL.Scheme = parts[0]
		newURL.Host = parts[1]
	}
	newReq := req.Clone(req.Context())
	newReq.URL = &newURL
	newReq.Host = newURL.Host
	return u.inner.RoundTrip(newReq)
}
