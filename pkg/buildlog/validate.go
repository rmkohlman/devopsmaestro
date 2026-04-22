package buildlog

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// sessionIDPattern is the strict allowlist for session ID filename components.
// It explicitly forbids "/", "\", "..", NUL, control characters, and any other
// character that could cause path traversal or shell injection. Length is
// bounded to 128 characters.
var sessionIDPattern = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,128}$`)

// forbiddenSystemPathPrefixes are absolute path prefixes that may never be
// used as the build log directory. Mirrors the pattern from
// operators/mount_validation.go.
var forbiddenSystemPathPrefixes = []string{
	"/etc",
	"/var",
	"/usr",
	"/bin",
	"/sbin",
	"/proc",
	"/sys",
	"/dev",
	"/boot",
	"/lib",
}

// validateSessionID rejects any sessionID that could be unsafe as a filename
// component. Returns a descriptive error on failure.
func validateSessionID(sid string) error {
	if sid == "" {
		return fmt.Errorf("buildlog: session ID is empty")
	}
	if len(sid) > 128 {
		return fmt.Errorf("buildlog: session ID exceeds 128 characters")
	}
	if strings.HasPrefix(sid, ".") {
		return fmt.Errorf("buildlog: session ID may not start with '.'")
	}
	if !sessionIDPattern.MatchString(sid) {
		return fmt.Errorf("buildlog: session ID contains invalid characters: %q (allowed: a-zA-Z0-9._-)", sid)
	}
	return nil
}

// validateNumericBounds enforces sane bounds on the rotation/retention knobs.
func validateNumericBounds(opts Options) error {
	if opts.MaxSizeMB <= 0 {
		return fmt.Errorf("buildlog: MaxSizeMB must be > 0, got %d", opts.MaxSizeMB)
	}
	if opts.MaxSizeMB > 100_000 {
		return fmt.Errorf("buildlog: MaxSizeMB must be <= 100000, got %d", opts.MaxSizeMB)
	}
	if opts.MaxAgeDays <= 0 {
		return fmt.Errorf("buildlog: MaxAgeDays must be > 0, got %d", opts.MaxAgeDays)
	}
	if opts.MaxBackups <= 0 {
		return fmt.Errorf("buildlog: MaxBackups must be > 0, got %d", opts.MaxBackups)
	}
	return nil
}

// expandHome expands a leading "~" to the current user's home directory.
func expandHome(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("buildlog: directory is empty")
	}
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("buildlog: cannot resolve home directory: %w", err)
	}
	if path == "~" {
		return home, nil
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

// validateDirectory expands, cleans, and validates the requested log
// directory. It rejects forbidden system paths and refuses to operate on a
// path whose final component is itself a symlink.
//
// Returns the cleaned absolute path on success.
func validateDirectory(dir string) (string, error) {
	expanded, err := expandHome(dir)
	if err != nil {
		return "", err
	}
	abs, err := filepath.Abs(expanded)
	if err != nil {
		return "", fmt.Errorf("buildlog: cannot resolve directory %q: %w", dir, err)
	}
	abs = filepath.Clean(abs)

	if strings.Contains(abs, "..") {
		return "", fmt.Errorf("buildlog: directory %q contains path traversal", dir)
	}
	if abs == "/" {
		return "", fmt.Errorf("buildlog: filesystem root is not allowed as log directory")
	}

	// Resolve symlinks in any existing parent component before checking
	// forbidden prefixes. On macOS, /var, /tmp, and /etc are symlinks to
	// /private/..., so the forbidden-prefix check must operate on canonical
	// paths to avoid both bypass (attacker-planted symlink) and false
	// positives (legitimate /var/folders tempdirs).
	canonical := canonicalize(abs)

	for _, prefix := range forbiddenSystemPathPrefixes {
		if canonical == prefix || strings.HasPrefix(canonical, prefix+string(filepath.Separator)) {
			return "", fmt.Errorf("buildlog: directory %q is under forbidden system path %q", dir, prefix)
		}
	}

	// If the path already exists and its final component is a symlink, refuse.
	if info, lerr := os.Lstat(abs); lerr == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("buildlog: directory %q is a symlink (refusing to follow)", dir)
		}
	}

	return abs, nil
}

// canonicalize returns the path with symlinks in its existing prefix
// resolved. If no parent exists yet (entirely new path), the original abs
// path is returned. This is used only for the forbidden-prefix check —
// not for the actual file operations.
func canonicalize(abs string) string {
	cur := abs
	var trailing []string
	for cur != "/" && cur != "." {
		if real, err := filepath.EvalSymlinks(cur); err == nil {
			if len(trailing) == 0 {
				return real
			}
			parts := append([]string{real}, reverse(trailing)...)
			return filepath.Join(parts...)
		}
		trailing = append(trailing, filepath.Base(cur))
		cur = filepath.Dir(cur)
	}
	return abs
}

func reverse(in []string) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[len(in)-1-i] = v
	}
	return out
}
