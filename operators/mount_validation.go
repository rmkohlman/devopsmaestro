package operators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rmkohlman/MaestroSDK/paths"
)

// sensitivePathPrefixes are host paths that must never be bind-mounted into a
// container. Each entry is matched as a prefix against the cleaned, absolute
// source path. Home-directory-relative paths (e.g. ~/.ssh) are expanded when
// InitSensitivePaths is called.
var sensitivePathPrefixes []string

// initSensitiveOnce ensures InitSensitivePaths runs exactly once.
var initSensitiveOnce sync.Once

// InitSensitivePaths builds the list of sensitive path prefixes that must never
// be bind-mounted. It resolves home-relative paths using paths.Default().
// This function is idempotent — safe to call multiple times.
func InitSensitivePaths() {
	initSensitiveOnce.Do(func() {
		// Absolute system paths — always blocked.
		static := []string{
			"/etc",
			"/var",
			"/proc",
			"/sys",
			"/dev",
			"/boot",
			"/sbin",
			"/lib",
		}

		// Home-relative sensitive directories.
		homeRelative := []string{
			".ssh",
			".gnupg",
			".aws",
			".azure",
			".kube",
			".docker",
			".config/gcloud",
		}

		sensitivePathPrefixes = append(sensitivePathPrefixes, static...)

		if pc, err := paths.Default(); err == nil {
			home := filepath.Dir(pc.Root())
			for _, rel := range homeRelative {
				sensitivePathPrefixes = append(sensitivePathPrefixes, filepath.Join(home, rel))
			}
		}
	})
}

// ValidateMountSource checks that a host source path is safe to bind-mount.
//
// It performs the following checks in order:
//  1. Resolves the path to an absolute, clean form.
//  2. Rejects paths that still contain ".." components after cleaning (traversal).
//  3. Rejects the filesystem root ("/") — mounting the entire host is never safe.
//  4. Rejects paths under sensitive system or user directories.
//  5. Verifies the source path exists on disk.
//
// Returns nil if the path is acceptable, or a descriptive error otherwise.
func ValidateMountSource(path string) error {
	// Ensure sensitive paths are initialized (idempotent).
	InitSensitivePaths()

	if path == "" {
		return fmt.Errorf("mount source path is empty")
	}

	// Step 1: resolve to absolute + clean
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("cannot resolve mount source path %q: %w", path, err)
	}
	abs = filepath.Clean(abs)

	// Step 2: reject residual ".." — should not happen after Abs+Clean on a
	// real path, but guards against symlink tricks or unusual input.
	if strings.Contains(abs, "..") {
		return fmt.Errorf("mount source path %q contains path traversal", path)
	}

	// Step 3: reject filesystem root
	if abs == "/" {
		return fmt.Errorf("mounting the filesystem root is not allowed")
	}

	// Step 4: reject sensitive directories
	for _, prefix := range sensitivePathPrefixes {
		// Match the exact directory or anything beneath it.
		if abs == prefix || strings.HasPrefix(abs, prefix+string(filepath.Separator)) {
			return fmt.Errorf("mount source path %q is under sensitive directory %q", path, prefix)
		}
	}

	// Step 5: verify path exists
	if _, err := os.Stat(abs); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("mount source path %q does not exist", path)
		}
		return fmt.Errorf("cannot access mount source path %q: %w", path, err)
	}

	return nil
}

// validateStartOptionsMounts validates all mount source paths in the given
// StartOptions (both opts.AppPath and opts.Mounts[].Source). It returns the
// first validation error encountered, or nil if all paths are safe.
func validateStartOptionsMounts(opts StartOptions) error {
	// Validate legacy AppPath
	if opts.AppPath != "" {
		if err := ValidateMountSource(opts.AppPath); err != nil {
			return fmt.Errorf("AppPath: %w", err)
		}
	}

	// Validate each mount source
	for i, mount := range opts.Mounts {
		if err := ValidateMountSource(mount.Source); err != nil {
			return fmt.Errorf("Mounts[%d] (%s): %w", i, mount.Destination, err)
		}
	}

	return nil
}
