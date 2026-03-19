package builders

// =============================================================================
// Issue #113 — TDD Phase 2: opencode checksum constant tests
//
// RED: These tests WILL NOT COMPILE until the following are added to checksums.go:
//
//   - opencodeVersion constant
//   - opencodeChecksumArm64 constant
//   - opencodeChecksumAmd64 constant
//
// Once the production code is added, all tests in this file MUST pass.
// =============================================================================

import (
	"regexp"
	"testing"
)

// sha256HexPattern matches a valid 64-character lowercase hex SHA256 checksum.
var sha256HexPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// TestOpencodeChecksums_VersionConstantExists verifies that opencodeVersion
// is defined and non-empty.
//
// RED: WILL NOT COMPILE — opencodeVersion constant does not exist yet.
func TestOpencodeChecksums_VersionConstantExists(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ──────────────────────────────────────────
	// opencodeVersion does not exist until #113 is implemented in checksums.go.
	if opencodeVersion == "" {
		t.Error("opencodeVersion constant must be non-empty — a pinned version is required")
	}
	// ──────────────────────────────────────────────────────────────────────────
}

// TestOpencodeChecksums_Arm64IsValidSHA256 verifies that opencodeChecksumArm64
// is a valid 64-character lowercase hex SHA256 checksum.
//
// RED: WILL NOT COMPILE — opencodeChecksumArm64 constant does not exist yet.
func TestOpencodeChecksums_Arm64IsValidSHA256(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ──────────────────────────────────────────
	// opencodeChecksumArm64 does not exist until #113 is implemented in checksums.go.
	if !sha256HexPattern.MatchString(opencodeChecksumArm64) {
		t.Errorf("opencodeChecksumArm64 = %q is not a valid SHA256 checksum.\n"+
			"Expected 64 lowercase hex characters, got %d characters.",
			opencodeChecksumArm64, len(opencodeChecksumArm64))
	}
	// ──────────────────────────────────────────────────────────────────────────
}

// TestOpencodeChecksums_Amd64IsValidSHA256 verifies that opencodeChecksumAmd64
// is a valid 64-character lowercase hex SHA256 checksum.
//
// RED: WILL NOT COMPILE — opencodeChecksumAmd64 constant does not exist yet.
func TestOpencodeChecksums_Amd64IsValidSHA256(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ──────────────────────────────────────────
	// opencodeChecksumAmd64 does not exist until #113 is implemented in checksums.go.
	if !sha256HexPattern.MatchString(opencodeChecksumAmd64) {
		t.Errorf("opencodeChecksumAmd64 = %q is not a valid SHA256 checksum.\n"+
			"Expected 64 lowercase hex characters, got %d characters.",
			opencodeChecksumAmd64, len(opencodeChecksumAmd64))
	}
	// ──────────────────────────────────────────────────────────────────────────
}

// TestOpencodeChecksums_Arm64AndAmd64AreDistinct verifies that the two
// architecture checksums are not equal — they must be different binaries.
//
// RED: WILL NOT COMPILE — checksum constants do not exist yet.
func TestOpencodeChecksums_Arm64AndAmd64AreDistinct(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ──────────────────────────────────────────
	// opencodeChecksumArm64 / opencodeChecksumAmd64 do not exist until #113.
	if opencodeChecksumArm64 == opencodeChecksumAmd64 {
		t.Error("opencodeChecksumArm64 and opencodeChecksumAmd64 must be different values.\n" +
			"arm64 and amd64 are distinct binaries with distinct checksums.")
	}
	// ──────────────────────────────────────────────────────────────────────────
}

// TestOpencodeChecksums_VersionNotEmpty verifies that opencodeVersion follows
// a semver-like pattern (non-empty, starts with a digit).
//
// RED: WILL NOT COMPILE — opencodeVersion constant does not exist yet.
func TestOpencodeChecksums_VersionFormat(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ──────────────────────────────────────────
	// opencodeVersion does not exist until #113 is implemented in checksums.go.
	version := opencodeVersion
	// ──────────────────────────────────────────────────────────────────────────

	if len(version) == 0 {
		t.Error("opencodeVersion must not be empty")
		return
	}

	// Version should start with a digit (e.g., "0.1.0", "1.0.0") — not "v0.1.0"
	// (consistent with other constants in checksums.go which omit the leading 'v')
	if version[0] == 'v' {
		t.Errorf("opencodeVersion = %q should NOT start with 'v'.\n"+
			"All version constants in checksums.go omit the leading 'v' (e.g., '0.1.0' not 'v0.1.0').\n"+
			"The builder stage prepends 'v' in the download URL.", version)
	}
}
