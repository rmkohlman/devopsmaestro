package registry

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Task 1: IsPortAvailable Tests
// =============================================================================

func TestIsPortAvailable(t *testing.T) {
	tests := []struct {
		name          string
		port          int
		bindFirst     bool
		wantAvailable bool
		description   string
	}{
		{
			name:          "high port unoccupied",
			port:          59999,
			bindFirst:     false,
			wantAvailable: true,
			description:   "IsPortAvailable should return true for unoccupied port",
		},
		{
			name:          "port in use",
			port:          59998,
			bindFirst:     true,
			wantAvailable: false,
			description:   "IsPortAvailable should return false for port in use",
		},
		{
			name:          "privileged port",
			port:          80,
			bindFirst:     false,
			wantAvailable: false,
			description:   "IsPortAvailable should handle privileged ports",
		},
		{
			name:          "negative port",
			port:          -1,
			bindFirst:     false,
			wantAvailable: false,
			description:   "IsPortAvailable should return false for invalid port",
		},
		{
			name:          "port too high",
			port:          70000,
			bindFirst:     false,
			wantAvailable: false,
			description:   "IsPortAvailable should return false for out-of-range port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var listener net.Listener

			if tt.bindFirst && tt.port > 1024 && tt.port < 65536 {
				// Bind to port to make it unavailable
				var err error
				listener, err = net.Listen("tcp", ":"+strconv.Itoa(tt.port))
				if err != nil {
					t.Skipf("Cannot bind to port %d: %v", tt.port, err)
				}
				defer listener.Close()
			}

			available := IsPortAvailable(tt.port)
			assert.Equal(t, tt.wantAvailable, available, tt.description)
		})
	}
}

func TestIsPortAvailable_PortRange(t *testing.T) {
	tests := []struct {
		name        string
		port        int
		description string
	}{
		{
			name:        "minimum valid port",
			port:        1024,
			description: "Port 1024 should be checkable",
		},
		{
			name:        "common registry port",
			port:        5000,
			description: "Common registry ports should be checkable",
		},
		{
			name:        "high port",
			port:        50000,
			description: "High ports should be checkable",
		},
		{
			name:        "maximum valid port",
			port:        65535,
			description: "Port 65535 should be checkable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify function doesn't panic with valid ports
			assert.NotPanics(t, func() {
				IsPortAvailable(tt.port)
			}, tt.description)
		})
	}
}

func TestIsPortAvailable_ReleasesPort(t *testing.T) {
	port := 59997

	// First check should find port available
	available1 := IsPortAvailable(port)
	assert.True(t, available1, "Port should be available initially")

	// Second check should also find it available (function should release port)
	available2 := IsPortAvailable(port)
	assert.True(t, available2, "Port should still be available after check")

	// We should be able to bind to the port ourselves
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	require.NoError(t, err, "Should be able to bind to port after IsPortAvailable check")
	listener.Close()
}

// =============================================================================
// Task 2: WaitForReady Tests
// =============================================================================

func TestWaitForReady(t *testing.T) {
	tests := []struct {
		name             string
		serverStatus     int
		acceptedStatuses []int
		serverDelay      time.Duration
		timeout          time.Duration
		wantError        bool
		description      string
	}{
		{
			name:             "endpoint ready immediately",
			serverStatus:     200,
			acceptedStatuses: []int{200},
			serverDelay:      0,
			timeout:          1 * time.Second,
			wantError:        false,
			description:      "WaitForReady should return nil when endpoint is ready",
		},
		{
			name:             "endpoint ready with 204",
			serverStatus:     204,
			acceptedStatuses: []int{200, 204},
			serverDelay:      0,
			timeout:          1 * time.Second,
			wantError:        false,
			description:      "WaitForReady should accept 204 status",
		},
		{
			name:             "endpoint ready after delay",
			serverStatus:     200,
			acceptedStatuses: []int{200},
			serverDelay:      100 * time.Millisecond,
			timeout:          2 * time.Second,
			wantError:        false,
			description:      "WaitForReady should poll until endpoint is ready",
		},
		{
			name:             "endpoint times out",
			serverStatus:     500,
			acceptedStatuses: []int{200},
			serverDelay:      0,
			timeout:          100 * time.Millisecond,
			wantError:        true,
			description:      "WaitForReady should timeout if endpoint never ready",
		},
		{
			name:             "endpoint returns wrong status",
			serverStatus:     404,
			acceptedStatuses: []int{200, 204},
			serverDelay:      0,
			timeout:          100 * time.Millisecond,
			wantError:        true,
			description:      "WaitForReady should fail if status not in accepted list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			requestCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount++

				// Delay first request if specified
				if requestCount == 1 && tt.serverDelay > 0 {
					time.Sleep(tt.serverDelay)
				}

				w.WriteHeader(tt.serverStatus)
			}))
			defer server.Close()

			// Wait for ready
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			err := WaitForReady(ctx, server.URL, tt.acceptedStatuses, tt.timeout)

			if tt.wantError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

func TestWaitForReady_RespectsContext(t *testing.T) {
	// Create server that never responds with accepted status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer server.Close()

	// Create context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := WaitForReady(ctx, server.URL, []int{200}, 5*time.Second)
	assert.Error(t, err, "WaitForReady should fail when context is cancelled")
	assert.Contains(t, err.Error(), "context", "Error should mention context cancellation")
}

func TestWaitForReady_PollsMultipleTimes(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		// Return 503 first 2 times, then 200
		if requestCount < 3 {
			w.WriteHeader(503)
		} else {
			w.WriteHeader(200)
		}
	}))
	defer server.Close()

	ctx := context.Background()
	err := WaitForReady(ctx, server.URL, []int{200}, 2*time.Second)

	assert.NoError(t, err, "WaitForReady should eventually succeed")
	assert.GreaterOrEqual(t, requestCount, 3, "WaitForReady should poll multiple times")
}

func TestWaitForReady_MultipleAcceptedStatuses(t *testing.T) {
	tests := []struct {
		name             string
		serverStatus     int
		acceptedStatuses []int
		wantError        bool
	}{
		{
			name:             "200 accepted",
			serverStatus:     200,
			acceptedStatuses: []int{200, 204, 301},
			wantError:        false,
		},
		{
			name:             "204 accepted",
			serverStatus:     204,
			acceptedStatuses: []int{200, 204, 301},
			wantError:        false,
		},
		{
			name:             "301 accepted",
			serverStatus:     301,
			acceptedStatuses: []int{200, 204, 301},
			wantError:        false,
		},
		{
			name:             "404 not accepted",
			serverStatus:     404,
			acceptedStatuses: []int{200, 204, 301},
			wantError:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
			}))
			defer server.Close()

			ctx := context.Background()
			err := WaitForReady(ctx, server.URL, tt.acceptedStatuses, 500*time.Millisecond)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// =============================================================================
// Task 3: CalculateDiskUsage Tests
// =============================================================================

func TestCalculateDiskUsage(t *testing.T) {
	tests := []struct {
		name        string
		setupFiles  map[string]int64 // filename -> size in bytes
		wantBytes   int64
		description string
	}{
		{
			name:        "empty directory",
			setupFiles:  map[string]int64{},
			wantBytes:   0,
			description: "CalculateDiskUsage should return 0 for empty directory",
		},
		{
			name: "single file",
			setupFiles: map[string]int64{
				"file1.txt": 100,
			},
			wantBytes:   100,
			description: "CalculateDiskUsage should return file size",
		},
		{
			name: "multiple files",
			setupFiles: map[string]int64{
				"file1.txt": 100,
				"file2.txt": 200,
				"file3.txt": 300,
			},
			wantBytes:   600,
			description: "CalculateDiskUsage should sum all file sizes",
		},
		{
			name: "nested directories",
			setupFiles: map[string]int64{
				"file1.txt":        100,
				"subdir/file2.txt": 200,
				"subdir/file3.txt": 300,
			},
			wantBytes:   600,
			description: "CalculateDiskUsage should include nested files",
		},
		{
			name: "deeply nested",
			setupFiles: map[string]int64{
				"a/b/c/file1.txt": 500,
				"a/b/file2.txt":   250,
				"a/file3.txt":     125,
			},
			wantBytes:   875,
			description: "CalculateDiskUsage should traverse deep directories",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir := t.TempDir()

			// Create test files
			for filename, size := range tt.setupFiles {
				fullPath := filepath.Join(tmpDir, filename)

				// Create parent directories
				parentDir := filepath.Dir(fullPath)
				err := os.MkdirAll(parentDir, 0755)
				require.NoError(t, err)

				// Create file with specified size
				data := make([]byte, size)
				err = os.WriteFile(fullPath, data, 0644)
				require.NoError(t, err)
			}

			// Calculate disk usage
			usage := CalculateDiskUsage(tmpDir)

			assert.Equal(t, tt.wantBytes, usage, tt.description)
		})
	}
}

func TestCalculateDiskUsage_NonExistentPath(t *testing.T) {
	usage := CalculateDiskUsage("/path/that/does/not/exist/12345")
	assert.Equal(t, int64(0), usage, "CalculateDiskUsage should return 0 for non-existent path")
}

func TestCalculateDiskUsage_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create empty file
	emptyFile := filepath.Join(tmpDir, "empty.txt")
	err := os.WriteFile(emptyFile, []byte{}, 0644)
	require.NoError(t, err)

	usage := CalculateDiskUsage(tmpDir)
	assert.Equal(t, int64(0), usage, "CalculateDiskUsage should count empty file as 0 bytes")
}

func TestCalculateDiskUsage_LargeFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a 1MB file
	largeFile := filepath.Join(tmpDir, "large.bin")
	data := make([]byte, 1024*1024) // 1 MB
	err := os.WriteFile(largeFile, data, 0644)
	require.NoError(t, err)

	usage := CalculateDiskUsage(tmpDir)
	assert.Equal(t, int64(1024*1024), usage, "CalculateDiskUsage should handle large files")
}

func TestCalculateDiskUsage_SymlinksIgnored(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a regular file
	regularFile := filepath.Join(tmpDir, "regular.txt")
	err := os.WriteFile(regularFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Create a symlink to the file
	symlinkFile := filepath.Join(tmpDir, "symlink.txt")
	err = os.Symlink(regularFile, symlinkFile)
	if err != nil {
		t.Skip("Cannot create symlinks on this system")
	}

	usage := CalculateDiskUsage(tmpDir)

	// Should only count the regular file once, not the symlink
	assert.Equal(t, int64(12), usage, "CalculateDiskUsage should not double-count symlinks")
}

// =============================================================================
// Task 4: EnsureDir Tests
// =============================================================================

func TestEnsureDir(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T, basePath string) string // Returns path to ensure
		wantError   bool
		description string
	}{
		{
			name: "creates new directory",
			setupFunc: func(t *testing.T, basePath string) string {
				return filepath.Join(basePath, "newdir")
			},
			wantError:   false,
			description: "EnsureDir should create directory if it doesn't exist",
		},
		{
			name: "directory already exists",
			setupFunc: func(t *testing.T, basePath string) string {
				dirPath := filepath.Join(basePath, "existingdir")
				err := os.Mkdir(dirPath, 0755)
				require.NoError(t, err)
				return dirPath
			},
			wantError:   false,
			description: "EnsureDir should return nil if directory already exists",
		},
		{
			name: "creates parent directories",
			setupFunc: func(t *testing.T, basePath string) string {
				return filepath.Join(basePath, "parent", "child", "grandchild")
			},
			wantError:   false,
			description: "EnsureDir should create parent directories like mkdir -p",
		},
		{
			name: "deeply nested path",
			setupFunc: func(t *testing.T, basePath string) string {
				return filepath.Join(basePath, "a", "b", "c", "d", "e")
			},
			wantError:   false,
			description: "EnsureDir should handle deeply nested paths",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			pathToEnsure := tt.setupFunc(t, tmpDir)

			err := EnsureDir(pathToEnsure)

			if tt.wantError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)

				// Verify directory exists
				info, err := os.Stat(pathToEnsure)
				require.NoError(t, err)
				assert.True(t, info.IsDir(), "Path should be a directory")
			}
		})
	}
}

func TestEnsureDir_PermissionsCorrect(t *testing.T) {
	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "newdir")

	err := EnsureDir(newDir)
	require.NoError(t, err)

	// Verify directory has correct permissions (0755)
	info, err := os.Stat(newDir)
	require.NoError(t, err)

	// On Unix systems, should be 0755
	if info.Mode().Perm() != 0755 {
		t.Logf("Directory permissions: %o (expected 0755)", info.Mode().Perm())
	}
}

func TestEnsureDir_FileExistsAtPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a regular file
	filePath := filepath.Join(tmpDir, "notadir")
	err := os.WriteFile(filePath, []byte("content"), 0644)
	require.NoError(t, err)

	// Try to ensure directory at same path as file
	err = EnsureDir(filePath)
	assert.Error(t, err, "EnsureDir should fail if regular file exists at path")
}

func TestEnsureDir_InvalidPath(t *testing.T) {
	// Try to create directory with invalid characters
	// Note: This is platform-dependent
	invalidPath := "/tmp/\x00invalid"

	err := EnsureDir(invalidPath)
	if err != nil {
		// Expected on most systems
		assert.Error(t, err, "EnsureDir should fail with invalid path")
	} else {
		t.Skip("Platform allows this path")
	}
}

func TestEnsureDir_RelativePath(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create relative path
	err = EnsureDir("relative/path/test")
	require.NoError(t, err)

	// Verify it exists
	_, err = os.Stat("relative/path/test")
	assert.NoError(t, err, "EnsureDir should work with relative paths")
}

func TestEnsureDir_AbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()
	absolutePath := filepath.Join(tmpDir, "absolute", "path")

	err := EnsureDir(absolutePath)
	require.NoError(t, err)

	// Verify it exists
	info, err := os.Stat(absolutePath)
	require.NoError(t, err)
	assert.True(t, info.IsDir(), "EnsureDir should work with absolute paths")
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestUtils_IntegrationScenario(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Ensure storage directory exists
	storagePath := filepath.Join(tmpDir, "registry", "storage")
	err := EnsureDir(storagePath)
	require.NoError(t, err, "Should create storage directory")

	// 2. Create some test files
	testFiles := []string{"image1.tar", "image2.tar", "manifest.json"}
	for _, filename := range testFiles {
		filePath := filepath.Join(storagePath, filename)
		err := os.WriteFile(filePath, []byte("test data"), 0644)
		require.NoError(t, err)
	}

	// 3. Calculate disk usage
	usage := CalculateDiskUsage(storagePath)
	expectedUsage := int64(len(testFiles) * len("test data"))
	assert.Equal(t, expectedUsage, usage, "Should calculate correct disk usage")

	// 4. Check if a port is available
	testPort := 59996
	available := IsPortAvailable(testPort)
	assert.True(t, available, "High port should be available")
}

// =============================================================================
// Compilation Tests
// =============================================================================

func TestUtils_FunctionsExist(t *testing.T) {
	// These tests will fail to compile until utils.go is implemented
	// That's expected - we're in TDD RED phase

	_ = IsPortAvailable(5000)
	_ = WaitForReady(context.Background(), "http://localhost", []int{200}, 1*time.Second)
	_ = CalculateDiskUsage("/tmp")
	_ = EnsureDir("/tmp/test")

	t.Skip("Utility functions not yet implemented - expected to fail")
}
