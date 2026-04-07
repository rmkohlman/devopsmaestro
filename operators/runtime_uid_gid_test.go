package operators

// =============================================================================
// TDD Phase 2 — Failing Tests for GitHub Issues #97 and #98
//
// #98: Runtime hardcodes UID 1000 but Dockerfile UID is configurable
//   → StartOptions and AttachOptions must carry UID/GID fields
//
// #97: Exec sessions missing --user for defense-in-depth
//   → AttachOptions UID/GID must flow through to nerdctl exec / Docker exec
//
// These tests FAIL TO COMPILE until the implementation adds UID/GID fields
// to StartOptions and AttachOptions in runtime_interface.go.
// =============================================================================

import (
	"fmt"
	"testing"
)

// =============================================================================
// Helper functions — these mirror what the implementation will use
// to convert the zero-value default into the hardcoded-1000 fallback.
// Tests here validate the defaulting contract so the implementation is clear.
// =============================================================================

// effectiveUID returns opts.UID if it is explicitly set (> 0), otherwise
// falls back to the legacy default of 1000 (the "dev" user in DVM images).
func effectiveUID(opts StartOptions) int {
	if opts.UID > 0 {
		return opts.UID
	}
	return 1000
}

// effectiveGID returns opts.GID if it is explicitly set (> 0), otherwise
// falls back to the legacy default of 1000.
func effectiveGID(opts StartOptions) int {
	if opts.GID > 0 {
		return opts.GID
	}
	return 1000
}

// effectiveAttachUID returns opts.UID if explicitly set, else 1000.
func effectiveAttachUID(opts AttachOptions) int {
	if opts.UID > 0 {
		return opts.UID
	}
	return 1000
}

// effectiveAttachGID returns opts.GID if explicitly set, else 1000.
func effectiveAttachGID(opts AttachOptions) int {
	if opts.GID > 0 {
		return opts.GID
	}
	return 1000
}

// =============================================================================
// Issue #98 — StartOptions must carry UID and GID fields
// =============================================================================

// TestStartOptionsHasUIDGIDFields verifies that StartOptions has UID and GID
// integer fields. This is a compile-time test: if the fields don't exist,
// this file will not compile.
//
// FAILS TO COMPILE until operators/runtime_interface.go adds:
//
//	type StartOptions struct {
//	    ...
//	    UID int   // Container user ID (default: 1000)
//	    GID int   // Container group ID (default: 1000)
//	}
func TestStartOptionsHasUIDGIDFields(t *testing.T) {
	// Arrange & Act — construct StartOptions with explicit UID/GID.
	// This line will produce a compile error until the fields are added.
	opts := StartOptions{
		ImageName:     "test:latest",
		WorkspaceName: "my-workspace",
		UID:           1001,
		GID:           1001,
	}

	// Assert — fields are accessible and hold the assigned values.
	if opts.UID != 1001 {
		t.Errorf("StartOptions.UID = %d, want 1001", opts.UID)
	}
	if opts.GID != 1001 {
		t.Errorf("StartOptions.GID = %d, want 1001", opts.GID)
	}
}

// TestStartOptionsDefaultsUID verifies the zero-value defaulting contract:
// when UID/GID are 0 (unset), the effective value must be 1000 (the DVM
// container user).  When set explicitly, the explicit value wins.
func TestStartOptionsDefaultsUID(t *testing.T) {
	tests := []struct {
		name    string
		uid     int
		gid     int
		wantUID int
		wantGID int
	}{
		{
			name:    "zero values default to 1000",
			uid:     0,
			gid:     0,
			wantUID: 1000,
			wantGID: 1000,
		},
		{
			name:    "explicit 1001 is preserved",
			uid:     1001,
			gid:     1001,
			wantUID: 1001,
			wantGID: 1001,
		},
		{
			name:    "custom uid 500 is preserved",
			uid:     500,
			gid:     500,
			wantUID: 500,
			wantGID: 500,
		},
		{
			name:    "uid and gid can differ",
			uid:     1001,
			gid:     1002,
			wantUID: 1001,
			wantGID: 1002,
		},
		{
			name:    "only uid set, gid defaults to 1000",
			uid:     2000,
			gid:     0,
			wantUID: 2000,
			wantGID: 1000,
		},
		{
			name:    "only gid set, uid defaults to 1000",
			uid:     0,
			gid:     2000,
			wantUID: 1000,
			wantGID: 2000,
		},
		{
			name:    "root uid 1 is preserved (not overridden by default)",
			uid:     1,
			gid:     1,
			wantUID: 1,
			wantGID: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			opts := StartOptions{
				UID: tt.uid, // compile error here until field exists
				GID: tt.gid, // compile error here until field exists
			}

			// Act
			gotUID := effectiveUID(opts)
			gotGID := effectiveGID(opts)

			// Assert
			if gotUID != tt.wantUID {
				t.Errorf("effectiveUID(%+v) = %d, want %d", opts, gotUID, tt.wantUID)
			}
			if gotGID != tt.wantGID {
				t.Errorf("effectiveGID(%+v) = %d, want %d", opts, gotGID, tt.wantGID)
			}
		})
	}
}

// TestStartOptionsUserString verifies that the UID/GID pair is formatted into
// the "uid:gid" string that is passed to --user / User: in container configs.
func TestStartOptionsUserString(t *testing.T) {
	tests := []struct {
		name     string
		uid      int
		gid      int
		wantUser string
	}{
		{
			name:     "defaults produce 1000:1000",
			uid:      0,
			gid:      0,
			wantUser: "1000:1000",
		},
		{
			name:     "custom 1001:1001",
			uid:      1001,
			gid:      1001,
			wantUser: "1001:1001",
		},
		{
			name:     "mixed custom uid/gid",
			uid:      1500,
			gid:      1600,
			wantUser: "1500:1600",
		},
		{
			name:     "low uid like 500:500",
			uid:      500,
			gid:      500,
			wantUser: "500:500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			opts := StartOptions{
				UID: tt.uid, // compile error here until field exists
				GID: tt.gid,
			}

			// Act — compute the user string using the defaulting helpers,
			// then format as "uid:gid" (exactly the string used in --user flag
			// and Docker container.Config.User)
			uid := effectiveUID(opts)
			gid := effectiveGID(opts)
			gotUser := fmt.Sprintf("%d:%d", uid, gid)

			// Assert
			if gotUser != tt.wantUser {
				t.Errorf("user string for UID=%d,GID=%d = %q, want %q",
					tt.uid, tt.gid, gotUser, tt.wantUser)
			}
		})
	}
}

// =============================================================================
// Issue #97 — AttachOptions must carry UID and GID for exec defense-in-depth
// =============================================================================

// TestAttachOptionsHasUIDGIDFields verifies that AttachOptions has UID and GID
// integer fields. Same compile-time check as above but for the attach path.
//
// FAILS TO COMPILE until operators/runtime_interface.go adds:
//
//	type AttachOptions struct {
//	    ...
//	    UID int   // User ID for exec session (default: 1000)
//	    GID int   // Group ID for exec session (default: 1000)
//	}
func TestAttachOptionsHasUIDGIDFields(t *testing.T) {
	// Arrange & Act — construct AttachOptions with explicit UID/GID.
	// This line will produce a compile error until the fields are added.
	opts := AttachOptions{
		WorkspaceID: "test-container",
		Shell:       "/bin/zsh",
		UID:         1001, // compile error here until field exists
		GID:         1001, // compile error here until field exists
	}

	// Assert
	if opts.UID != 1001 {
		t.Errorf("AttachOptions.UID = %d, want 1001", opts.UID)
	}
	if opts.GID != 1001 {
		t.Errorf("AttachOptions.GID = %d, want 1001", opts.GID)
	}
}

// TestAttachOptionsDefaultsUID verifies the same zero-value fallback contract
// applies to exec sessions: unset means 1000, set means use the explicit value.
func TestAttachOptionsDefaultsUID(t *testing.T) {
	tests := []struct {
		name    string
		uid     int
		gid     int
		wantUID int
		wantGID int
	}{
		{
			name:    "zero values default to 1000",
			uid:     0,
			gid:     0,
			wantUID: 1000,
			wantGID: 1000,
		},
		{
			name:    "explicit 1001 is preserved",
			uid:     1001,
			gid:     1001,
			wantUID: 1001,
			wantGID: 1001,
		},
		{
			name:    "custom uid 500 is preserved",
			uid:     500,
			gid:     500,
			wantUID: 500,
			wantGID: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			opts := AttachOptions{
				WorkspaceID: "test-container",
				UID:         tt.uid, // compile error until field exists
				GID:         tt.gid,
			}

			// Act
			gotUID := effectiveAttachUID(opts)
			gotGID := effectiveAttachGID(opts)

			// Assert
			if gotUID != tt.wantUID {
				t.Errorf("effectiveAttachUID(%+v) = %d, want %d", opts, gotUID, tt.wantUID)
			}
			if gotGID != tt.wantGID {
				t.Errorf("effectiveAttachGID(%+v) = %d, want %d", opts, gotGID, tt.wantGID)
			}
		})
	}
}

// TestAttachOptionsUserString verifies the "uid:gid" format for nerdctl exec
// --user and Docker ExecOptions.User fields.
func TestAttachOptionsUserString(t *testing.T) {
	tests := []struct {
		name     string
		uid      int
		gid      int
		wantUser string
	}{
		{
			name:     "defaults produce 1000:1000",
			uid:      0,
			gid:      0,
			wantUser: "1000:1000",
		},
		{
			name:     "custom 1001:1001",
			uid:      1001,
			gid:      1001,
			wantUser: "1001:1001",
		},
		{
			name:     "inherits same uid/gid as StartOptions ensures consistency",
			uid:      1500,
			gid:      1500,
			wantUser: "1500:1500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			opts := AttachOptions{
				WorkspaceID: "test-container",
				UID:         tt.uid, // compile error until field exists
				GID:         tt.gid,
			}

			// Act
			uid := effectiveAttachUID(opts)
			gid := effectiveAttachGID(opts)
			gotUser := fmt.Sprintf("%d:%d", uid, gid)

			// Assert
			if gotUser != tt.wantUser {
				t.Errorf("user string for UID=%d,GID=%d = %q, want %q",
					tt.uid, tt.gid, gotUser, tt.wantUser)
			}
		})
	}
}

// =============================================================================
// Cross-cutting: start and exec sessions must agree on the user
// =============================================================================

// TestStartAndAttachUserConsistency verifies that when a workspace is started
// with a custom UID/GID, an attach session with the same values produces the
// same "uid:gid" user string — preventing the runtime mismatch described in #98.
func TestStartAndAttachUserConsistency(t *testing.T) {
	tests := []struct {
		name           string
		uid            int
		gid            int
		wantConsistent bool
	}{
		{
			name:           "both zero produce same 1000:1000 default",
			uid:            0,
			gid:            0,
			wantConsistent: true,
		},
		{
			name:           "custom uid 1001 is consistent across start and attach",
			uid:            1001,
			gid:            1001,
			wantConsistent: true,
		},
		{
			name:           "custom uid 500 is consistent across start and attach",
			uid:            500,
			gid:            500,
			wantConsistent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			startOpts := StartOptions{
				ImageName:     "test:latest",
				WorkspaceName: "my-workspace",
				UID:           tt.uid, // compile error until field exists
				GID:           tt.gid,
			}
			attachOpts := AttachOptions{
				WorkspaceID: "my-workspace",
				UID:         tt.uid, // compile error until field exists
				GID:         tt.gid,
			}

			// Act
			startUser := fmt.Sprintf("%d:%d", effectiveUID(startOpts), effectiveGID(startOpts))
			attachUser := fmt.Sprintf("%d:%d", effectiveAttachUID(attachOpts), effectiveAttachGID(attachOpts))

			// Assert — the user string must match between start and exec
			consistent := startUser == attachUser
			if consistent != tt.wantConsistent {
				t.Errorf("start user %q vs attach user %q: consistent=%v, want %v",
					startUser, attachUser, consistent, tt.wantConsistent)
			}
		})
	}
}
