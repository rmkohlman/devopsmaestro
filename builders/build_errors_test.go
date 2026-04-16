package builders

import (
	"errors"
	"strings"
	"testing"
)

func TestEnhanceBuildError_NilError(t *testing.T) {
	got := EnhanceBuildError(nil)
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestEnhanceBuildError_NetworkTimeout(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		wantHint bool
	}{
		{"i/o timeout", `dial tcp 98.90.0.75:443: i/o timeout`, true},
		{"dial tcp", `failed to do request: dial tcp 1.2.3.4:443: connect: connection refused`, true},
		{"connection refused", `connection refused`, true},
		{"no such host", `no such host`, true},
		{"network unreachable", `network is unreachable`, true},
		{"unrelated error", `permission denied: /etc/shadow`, false},
		{"generic build failure", `exit status 1`, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := EnhanceBuildError(errors.New(tc.errMsg))
			if err == nil {
				t.Fatal("expected non-nil error")
			}
			hasHint := strings.Contains(err.Error(), "dvm restart registries")
			if hasHint != tc.wantHint {
				t.Errorf("wantHint=%v, got err: %v", tc.wantHint, err)
			}
			// All errors should be wrapped with "build failed:"
			if !strings.HasPrefix(err.Error(), "build failed:") {
				t.Errorf("expected 'build failed:' prefix, got: %v", err)
			}
		})
	}
}

func TestEnhanceBuildError_WrapsOriginalError(t *testing.T) {
	orig := errors.New("dial tcp 1.2.3.4:443: i/o timeout")
	enhanced := EnhanceBuildError(orig)
	if !errors.Is(enhanced, orig) {
		t.Errorf("EnhanceBuildError should wrap original error via errors.Is chain")
	}
}

func TestIsCacheCorruption_NilError(t *testing.T) {
	if IsCacheCorruption(nil) {
		t.Error("expected false for nil error")
	}
}

func TestIsCacheCorruption_PositiveCases(t *testing.T) {
	positives := []string{
		"blob not found: not found",
		"ERROR: failed to get reader from content store: blob sha256:abc123: blob not found",
		"parent snapshot does not exist",
		"failed to get reader from content store",
		"missing content for digest sha256:deadbeef",
		"failed to verify sha256 checksum",
		"index.json: no such file or directory",
		// Case-insensitive variants
		"BLOB NOT FOUND",
		"Parent Snapshot Does Not Exist",
	}
	for _, msg := range positives {
		t.Run(msg, func(t *testing.T) {
			if !IsCacheCorruption(errors.New(msg)) {
				t.Errorf("IsCacheCorruption(%q) = false, want true", msg)
			}
		})
	}
}

func TestIsCacheCorruption_NegativeCases(t *testing.T) {
	negatives := []string{
		"exit status 1",
		"permission denied",
		"i/o timeout",
		"dial tcp 1.2.3.4:443: connection refused",
		"context deadline exceeded",
		"",
	}
	for _, msg := range negatives {
		t.Run(msg, func(t *testing.T) {
			if IsCacheCorruption(errors.New(msg)) {
				t.Errorf("IsCacheCorruption(%q) = true, want false", msg)
			}
		})
	}
}

func TestIsCacheCorruptionMsg_AllPatterns(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want bool
	}{
		{"blob not found", "blob not found", true},
		{"parent snapshot does not exist", "parent snapshot does not exist", true},
		{"failed to get reader from content store", "failed to get reader from content store", true},
		{"missing content", "missing content", true},
		{"failed to verify", "failed to verify", true},
		{"unexpected end of json input", "unexpected end of json input", true},
		{"index.json no such file", "index.json: no such file or directory", true},
		{"case insensitive", "BLOB NOT FOUND", true},
		{"unrelated error", "something else entirely", false},
		{"empty", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isCacheCorruptionMsg(tc.msg)
			if got != tc.want {
				t.Errorf("isCacheCorruptionMsg(%q) = %v, want %v", tc.msg, got, tc.want)
			}
		})
	}
}

func TestEnhanceBuildError_CacheCorruption(t *testing.T) {
	corruptionErrors := []string{
		"blob not found: not found",
		"parent snapshot does not exist",
		"failed to get reader from content store: blob sha256:abc: blob not found",
	}
	for _, msg := range corruptionErrors {
		t.Run(msg, func(t *testing.T) {
			orig := errors.New(msg)
			enhanced := EnhanceBuildError(orig)
			if enhanced == nil {
				t.Fatal("expected non-nil error")
			}
			enhanced_str := enhanced.Error()
			if !strings.HasPrefix(enhanced_str, "build failed:") {
				t.Errorf("expected 'build failed:' prefix, got: %v", enhanced_str)
			}
			if !strings.Contains(enhanced_str, "clean-cache") {
				t.Errorf("expected '--clean-cache' hint in error, got: %v", enhanced_str)
			}
			if !strings.Contains(enhanced_str, "BuildKit cache corruption") {
				t.Errorf("expected cache corruption hint in error, got: %v", enhanced_str)
			}
			if !errors.Is(enhanced, orig) {
				t.Error("expected enhanced error to wrap original via errors.Is chain")
			}
		})
	}
}

func TestIsNetworkTimeout(t *testing.T) {
	tests := []struct {
		msg  string
		want bool
	}{
		{"i/o timeout", true},
		{"I/O TIMEOUT", true}, // case-insensitive
		{"dial tcp 1.2.3.4:80: connect: connection refused", true},
		{"no such host registry-1.docker.io", true},
		{"network is unreachable", true},
		{"connection refused", true},
		{"exit status 2", false},
		{"", false},
	}
	for _, tc := range tests {
		t.Run(tc.msg, func(t *testing.T) {
			got := isNetworkTimeout(tc.msg)
			if got != tc.want {
				t.Errorf("isNetworkTimeout(%q) = %v, want %v", tc.msg, got, tc.want)
			}
		})
	}
}
