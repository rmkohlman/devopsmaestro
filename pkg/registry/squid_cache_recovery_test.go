package registry

import (
	"os"
	"path/filepath"
	"testing"
)

// TestIsXCallocError verifies detection of the xcalloc integer overflow log line
// that indicates corrupted swap.state (issue #363, #377).
func TestIsXCallocError(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		{
			name:   "exact fatal line from squid",
			output: "2026/04/16 00:51:05| FATAL: xcalloc: Unable to allocate 18446744073709551615 blocks of 432 bytes!",
			want:   true,
		},
		{
			name:   "contains xcalloc keyword",
			output: "xcalloc failed",
			want:   true,
		},
		{
			name:   "contains Unable to allocate",
			output: "Unable to allocate 999 blocks",
			want:   true,
		},
		{
			name:   "terminated abnormally without xcalloc",
			output: "Squid Cache (Version 7.4): Terminated abnormally.",
			want:   false,
		},
		{
			name:   "empty string",
			output: "",
			want:   false,
		},
		{
			name:   "normal startup output",
			output: "2026/04/16 00:51:05| Starting Squid Cache version 7.4",
			want:   false,
		},
		{
			name:   "already running",
			output: "Squid is already running",
			want:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isXCallocError(tc.output)
			if got != tc.want {
				t.Errorf("isXCallocError(%q) = %v, want %v", tc.output, got, tc.want)
			}
		})
	}
}

// TestClearCacheDir verifies that clearCacheDir removes all entries inside the
// cache directory while leaving the directory itself intact.
func TestClearCacheDir(t *testing.T) {
	t.Run("removes files and subdirs", func(t *testing.T) {
		cacheDir := t.TempDir()

		// Populate the cache dir with files and a subdirectory
		if err := os.WriteFile(filepath.Join(cacheDir, "swap.state"), []byte("corrupt"), 0644); err != nil {
			t.Fatal(err)
		}
		subDir := filepath.Join(cacheDir, "00")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(subDir, "some_cache_file"), []byte("data"), 0644); err != nil {
			t.Fatal(err)
		}

		m := &SquidManager{config: HttpProxyConfig{CacheDir: cacheDir}}
		if err := m.clearCacheDir(); err != nil {
			t.Fatalf("clearCacheDir returned error: %v", err)
		}

		entries, err := os.ReadDir(cacheDir)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 0 {
			t.Errorf("expected empty cache dir, found %d entries", len(entries))
		}
	})

	t.Run("non-existent cache dir is a no-op", func(t *testing.T) {
		m := &SquidManager{config: HttpProxyConfig{CacheDir: "/tmp/devopsmaestro_test_no_such_dir_xyz"}}
		if err := m.clearCacheDir(); err != nil {
			t.Errorf("clearCacheDir on missing dir should not error, got: %v", err)
		}
	})

	t.Run("empty cache dir succeeds", func(t *testing.T) {
		cacheDir := t.TempDir()
		m := &SquidManager{config: HttpProxyConfig{CacheDir: cacheDir}}
		if err := m.clearCacheDir(); err != nil {
			t.Errorf("clearCacheDir on empty dir should not error, got: %v", err)
		}
	})
}
