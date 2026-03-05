package registry

import (
	"context"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// healthCheckClient is the shared HTTP client for health-check requests in
// waitForReady methods.  It has a short timeout to prevent hung requests from
// blocking service startup indefinitely.
var healthCheckClient = &http.Client{Timeout: 2 * time.Second}

// IsPortAvailable checks if a TCP port is available for binding.
func IsPortAvailable(port int) bool {
	// Validate port range - exclude privileged ports (< 1024)
	if port < 1024 || port > 65535 {
		return false
	}

	// Try to bind to the port
	addr := ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}

	// Port is available, release it
	listener.Close()
	return true
}

// WaitForReady polls an HTTP endpoint until it returns an accepted status code.
// It uses the shared healthCheckClient (2s per-request timeout) and a configurable
// overall timeout.
func WaitForReady(ctx context.Context, endpoint string, acceptedStatuses []int, timeout time.Duration) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout waiting for endpoint to be ready: %w", timeoutCtx.Err())
		case <-ticker.C:
			req, err := http.NewRequestWithContext(timeoutCtx, "GET", endpoint, nil)
			if err != nil {
				continue
			}

			resp, err := healthCheckClient.Do(req)
			if err != nil {
				continue
			}

			statusCode := resp.StatusCode
			resp.Body.Close()

			for _, accepted := range acceptedStatuses {
				if statusCode == accepted {
					return nil
				}
			}
		}
	}
}

// WaitForReadyTCP polls a TCP endpoint until a connection can be established.
// Used for services like Squid that do not expose an HTTP health endpoint.
func WaitForReadyTCP(ctx context.Context, address string, timeout time.Duration) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout waiting for endpoint to be ready: %w", timeoutCtx.Err())
		case <-ticker.C:
			conn, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
			if err == nil {
				conn.Close()
				return nil
			}
		}
	}
}

// CalculateDiskUsage walks a directory tree and returns total size in bytes.
func CalculateDiskUsage(storagePath string) int64 {
	var totalSize int64

	// Walk the directory tree
	err := filepath.WalkDir(storagePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Skip directories we can't read
			return nil
		}

		// Skip directories and symlinks
		if d.IsDir() {
			return nil
		}

		// Get file info
		info, err := d.Info()
		if err != nil {
			return nil
		}

		// Add file size (skip symlinks by checking mode)
		if info.Mode()&os.ModeSymlink == 0 {
			totalSize += info.Size()
		}

		return nil
	})

	// If walk fails, return 0
	if err != nil {
		return 0
	}

	return totalSize
}

// EnsureDir creates a directory and all parents if they don't exist.
func EnsureDir(path string) error {
	// Check if path already exists
	info, err := os.Stat(path)
	if err == nil {
		// Path exists - check if it's a directory
		if !info.IsDir() {
			return fmt.Errorf("path exists but is not a directory: %s", path)
		}
		return nil
	}

	// Path doesn't exist - create it with parents
	err = os.MkdirAll(path, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}

	return nil
}
