package registry

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// MockAthensBinaryManager is a mock implementation of BinaryManager for Athens testing.
type MockAthensBinaryManager struct {
	binDir  string
	version string

	// Hooks for customizing behavior in tests
	EnsureBinaryFunc func(ctx context.Context) (string, error)
	GetVersionFunc   func(ctx context.Context) (string, error)
	NeedsUpdateFunc  func(ctx context.Context) (bool, error)
	UpdateFunc       func(ctx context.Context) error
}

// NewMockAthensBinaryManager creates a MockAthensBinaryManager for testing.
func NewMockAthensBinaryManager(binDir, version string) *MockAthensBinaryManager {
	return &MockAthensBinaryManager{
		binDir:  binDir,
		version: version,
	}
}

// EnsureBinary creates a fake Athens binary file for testing.
func (m *MockAthensBinaryManager) EnsureBinary(ctx context.Context) (string, error) {
	if m.EnsureBinaryFunc != nil {
		return m.EnsureBinaryFunc(ctx)
	}

	// Default behavior: create a fake Athens binary
	binaryPath := filepath.Join(m.binDir, "athens")

	// Check if already exists
	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath, nil
	}

	// Create directory
	if err := os.MkdirAll(m.binDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Create fake executable script that simulates Athens
	// It should respond to HTTP requests for testing
	script := fmt.Sprintf(`#!/bin/bash
# Mock Athens binary for testing
if [[ "$1" == "--version" ]]; then
    echo "athens v%s"
    exit 0
fi

# Extract port from config file (simple grep)
PORT=3000
if [[ -f "config.toml" ]]; then
    PORT=$(grep -oP 'Port = ":\K\d+' config.toml 2>/dev/null || echo 3000)
fi

# Start a simple HTTP server that responds to /healthz
python3 -c "
import http.server
import socketserver
import sys

class HealthCheckHandler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        if self.path == '/healthz' or self.path == '/readyz':
            self.send_response(200)
            self.send_header('Content-type', 'text/plain')
            self.end_headers()
            self.wfile.write(b'OK')
        else:
            self.send_response(404)
            self.end_headers()
    
    def log_message(self, format, *args):
        pass  # Silence logs

PORT = int('$PORT')
with socketserver.TCPServer(('', PORT), HealthCheckHandler) as httpd:
    httpd.serve_forever()
" &
PID=$!
echo $PID
wait $PID
`, m.version)

	if err := os.WriteFile(binaryPath, []byte(script), 0755); err != nil {
		return "", fmt.Errorf("failed to create fake binary: %w", err)
	}

	return binaryPath, nil
}

// GetVersion returns the mock version.
func (m *MockAthensBinaryManager) GetVersion(ctx context.Context) (string, error) {
	if m.GetVersionFunc != nil {
		return m.GetVersionFunc(ctx)
	}

	return strings.TrimPrefix(m.version, "v"), nil
}

// NeedsUpdate always returns false for mock.
func (m *MockAthensBinaryManager) NeedsUpdate(ctx context.Context) (bool, error) {
	if m.NeedsUpdateFunc != nil {
		return m.NeedsUpdateFunc(ctx)
	}

	return false, nil
}

// Update is a no-op for mock.
func (m *MockAthensBinaryManager) Update(ctx context.Context) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx)
	}

	return nil
}

// StartMockAthensServer starts a simple HTTP server for testing without using the binary.
// This is useful for unit tests that need a quick health check endpoint.
func StartMockAthensServer(port int) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go server.ListenAndServe()

	return server
}
