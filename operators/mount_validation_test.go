package operators

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// =============================================================================
// Issue #106: Adversarial input tests for ValidateMountSource()
// =============================================================================

func TestValidateMountSource_AdversarialInputs(t *testing.T) {
	// Create a real temporary directory so "valid path exists" cases can pass
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		path      string
		wantError bool
		errorHint string // substring expected in error message (if wantError)
	}{
		// ---------------------------------------------------------------
		// Category: empty / whitespace
		// ---------------------------------------------------------------
		{
			name:      "empty string",
			path:      "",
			wantError: true,
			errorHint: "empty",
		},
		{
			name:      "whitespace only — spaces",
			path:      "   ",
			wantError: true,
			// Whitespace-only resolves to CWD/<spaces> which won't exist,
			// so we expect either "does not exist" or a path error.
		},

		// ---------------------------------------------------------------
		// Category: path traversal
		// ---------------------------------------------------------------
		{
			name:      "classic traversal — ../etc/passwd",
			path:      "../etc/passwd",
			wantError: true,
		},
		{
			name:      "deep traversal — ./../../secret",
			path:      "./../../secret",
			wantError: true,
		},
		{
			name:      "traversal into /etc via relative path",
			path:      "../../../../../../etc",
			wantError: true,
		},
		{
			name:      "traversal with encoded dot sequences — ..%2Fetc",
			path:      "..%2Fetc",
			wantError: true,
			// Not a real traversal after filepath.Abs, but should not exist on disk.
		},

		// ---------------------------------------------------------------
		// Category: sensitive path blocklist — direct hits
		// ---------------------------------------------------------------
		{
			name:      "blocked — /etc directly",
			path:      "/etc",
			wantError: true,
			errorHint: "sensitive",
		},
		{
			name:      "blocked — subpath under /etc",
			path:      "/etc/passwd",
			wantError: true,
			errorHint: "sensitive",
		},
		{
			name:      "blocked — filesystem root /",
			path:      "/",
			wantError: true,
			errorHint: "root",
		},
		{
			name:      "blocked — /proc/self/mem",
			path:      "/proc/self/mem",
			wantError: true,
			errorHint: "sensitive",
		},
		{
			name:      "blocked — /dev/null (under /dev)",
			path:      "/dev/null",
			wantError: true,
			errorHint: "sensitive",
		},

		// ---------------------------------------------------------------
		// Category: sensitive path blocklist — bypass attempts
		// ---------------------------------------------------------------
		{
			name:      "bypass attempt — trailing slash on /etc/",
			path:      "/etc/",
			wantError: true,
			errorHint: "sensitive",
		},
		{
			name:      "bypass attempt — double slash //etc",
			path:      "//etc",
			wantError: true,
			errorHint: "sensitive",
		},
		{
			name:      "bypass attempt — /etc with embedded dot //etc/..",
			path:      "/etc/..",
			wantError: true,
			// filepath.Clean resolves /etc/.. → / which is blocked as root
		},

		// ---------------------------------------------------------------
		// Category: command injection characters in path strings
		// ---------------------------------------------------------------
		{
			name:      "command injection — semicolon",
			path:      "/tmp/data;rm -rf /",
			wantError: true,
			// Path won't exist on disk, so "does not exist" error expected.
		},
		{
			name:      "command injection — backtick",
			path:      "/tmp/`whoami`",
			wantError: true,
		},
		{
			name:      "command injection — subshell $(...)",
			path:      "/tmp/$(cat /etc/passwd)",
			wantError: true,
		},
		{
			name:      "command injection — newline in path",
			path:      "/tmp/data\nrm -rf /",
			wantError: true,
		},

		// ---------------------------------------------------------------
		// Category: special / boundary values
		// ---------------------------------------------------------------
		{
			name:      "null byte in path",
			path:      "/tmp/data\x00evil",
			wantError: true,
			// filepath.Abs/os.Stat should reject or fail on null byte.
		},
		{
			name:      "extremely long path (>4096 chars)",
			path:      "/" + strings.Repeat("a", 4097),
			wantError: true,
			// Does not exist on disk; OS may also reject path length.
		},

		// ---------------------------------------------------------------
		// Category: macOS temp dirs — under /var (correctly rejected)
		// On macOS, t.TempDir() returns /var/folders/... which is under
		// the sensitive /var prefix and is therefore rejected by design.
		// ---------------------------------------------------------------
		{
			name:      "macOS temp directory — under /var, correctly rejected",
			path:      tmpDir,
			wantError: true,
			errorHint: "sensitive",
		},
		{
			name:      "macOS temp subdirectory — under /var, correctly rejected",
			path:      createSubdir(t, tmpDir, "project"),
			wantError: true,
			errorHint: "sensitive",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateMountSource(tc.path)

			if tc.wantError && err == nil {
				t.Errorf("ValidateMountSource(%q) returned nil; want an error — SECURITY FINDING: bad input accepted", tc.path)
				return
			}
			if !tc.wantError && err != nil {
				t.Errorf("ValidateMountSource(%q) returned unexpected error: %v", tc.path, err)
				return
			}
			if tc.wantError && tc.errorHint != "" {
				if !strings.Contains(err.Error(), tc.errorHint) {
					t.Errorf("ValidateMountSource(%q) error = %q; want it to contain %q", tc.path, err.Error(), tc.errorHint)
				}
			}
		})
	}
}

// TestValidateMountSource_HomeDirectoryBlocking verifies that sensitive
// home-relative paths (~/.ssh, ~/.aws, etc.) are blocked.
func TestValidateMountSource_HomeDirectoryBlocking(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory:", err)
	}

	sensitiveDirs := []string{
		filepath.Join(home, ".ssh"),
		filepath.Join(home, ".aws"),
		filepath.Join(home, ".kube"),
		filepath.Join(home, ".docker"),
		filepath.Join(home, ".gnupg"),
		filepath.Join(home, ".azure"),
	}

	for _, dir := range sensitiveDirs {
		t.Run(dir, func(t *testing.T) {
			err := ValidateMountSource(dir)
			if err == nil {
				t.Errorf("ValidateMountSource(%q) returned nil — SECURITY FINDING: sensitive home dir accepted", dir)
			}
			if err != nil && !strings.Contains(err.Error(), "sensitive") {
				t.Errorf("ValidateMountSource(%q) error = %q; expected 'sensitive' in message", dir, err.Error())
			}
		})
	}
}

// TestValidateMountSource_SensitiveSubpaths ensures that paths *beneath*
// a blocked prefix are also rejected (not just exact matches).
func TestValidateMountSource_SensitiveSubpaths(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory:", err)
	}

	subpaths := []string{
		filepath.Join(home, ".ssh", "id_rsa"),
		filepath.Join(home, ".ssh", "authorized_keys"),
		filepath.Join(home, ".aws", "credentials"),
		filepath.Join(home, ".kube", "config"),
		"/etc/shadow",
		"/etc/sudoers",
		"/var/log",
		"/sys/kernel",
		"/proc/1/environ",
	}

	for _, p := range subpaths {
		t.Run(p, func(t *testing.T) {
			err := ValidateMountSource(p)
			if err == nil {
				t.Errorf("ValidateMountSource(%q) returned nil — SECURITY FINDING: sensitive subpath accepted", p)
			}
		})
	}
}

// createSubdir creates a subdirectory under parent and returns its path.
// Skips the test if creation fails.
func createSubdir(t *testing.T, parent, name string) string {
	t.Helper()
	path := filepath.Join(parent, name)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Skipf("could not create subdir %s: %v", path, err)
	}
	return path
}
