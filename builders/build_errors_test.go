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
