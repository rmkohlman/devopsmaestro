package registry

import (
	"regexp"
	"strings"
)

// semverRe matches version strings like v0.14.1, 2.1.15, v1.0.0-rc1, etc.
var semverRe = regexp.MustCompile(`v?\d+\.\d+\.\d+(?:[-+][A-Za-z0-9._-]+)?`)

// versionKeyRe matches "Version:" or "version:" followed by a version string.
var versionKeyRe = regexp.MustCompile(`(?i)version\s*:\s*(v?\d+\.\d+\.\d+(?:[-+][A-Za-z0-9._-]+)?)`)

// sanitizeVersion extracts a clean version string from potentially multi-line
// command output. It handles cases like Athens' verbose build details:
//
//	Build Details:
//	    Version:    v0.14.1
//	    Date:       2024-06-02-23:35:29-UTC
//
// Returns the extracted version (with leading "v" stripped), or empty string if
// no version could be found.
func sanitizeVersion(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	// Single-line simple case: just a version number
	if !strings.Contains(raw, "\n") {
		s := strings.TrimSpace(raw)
		if m := semverRe.FindString(s); m != "" {
			return strings.TrimPrefix(m, "v")
		}
		// Might be "toolname X.Y.Z"
		return strings.TrimPrefix(s, "v")
	}

	// Multi-line: look for "Version: vX.Y.Z" pattern first
	if m := versionKeyRe.FindStringSubmatch(raw); len(m) > 1 {
		return strings.TrimPrefix(strings.TrimSpace(m[1]), "v")
	}

	// Fall back to first semver-like string in entire output
	if m := semverRe.FindString(raw); m != "" {
		return strings.TrimPrefix(m, "v")
	}

	return ""
}
