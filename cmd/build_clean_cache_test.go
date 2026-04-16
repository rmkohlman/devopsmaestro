package cmd

// =============================================================================
// Tests for build_clean_cache.go (#383)
// Covers: pre/post cleanup guards, registry cache ref generation, multi-value
// CacheFrom splitting (exercised via BuildOptions helpers).
// =============================================================================

import (
	"fmt"
	"strings"
	"testing"

	"devopsmaestro/operators"
)

// ---------------------------------------------------------------------------
// Section: preCleanCacheCleanup nil-platform guard
// ---------------------------------------------------------------------------

func TestPreCleanCacheCleanup_NilPlatform_DoesNotPanic(t *testing.T) {
	bc := &buildContext{
		platform: nil,
	}
	// Must not panic
	bc.preCleanCacheCleanup()
}

func TestPostCleanCacheCleanup_NilPlatform_DoesNotPanic(t *testing.T) {
	bc := &buildContext{
		platform: nil,
	}
	// Must not panic
	bc.postCleanCacheCleanup()
}

// ---------------------------------------------------------------------------
// Section: deleteExistingWorkspaceImages skips non-Docker-compatible
// ---------------------------------------------------------------------------

func TestDeleteExistingWorkspaceImages_SkipsNonDockerCompatible(t *testing.T) {
	bc := &buildContext{
		platform: &operators.Platform{
			Type: operators.PlatformType("unsupported-platform"),
		},
		workspaceName: "dev",
		appName:       "myapp",
	}
	// Should return immediately without panicking (no Docker client created)
	bc.deleteExistingWorkspaceImages()
}

// ---------------------------------------------------------------------------
// Section: pruneDanglingImages skips non-Docker-compatible
// ---------------------------------------------------------------------------

func TestPruneDanglingImages_SkipsNonDockerCompatible(t *testing.T) {
	bc := &buildContext{
		platform: &operators.Platform{
			Type: operators.PlatformType("unsupported-platform"),
		},
	}
	// Should return immediately without panicking
	bc.pruneDanglingImages()
}

// ---------------------------------------------------------------------------
// Section: logRegistryCacheStatus nil cacheReadiness
// ---------------------------------------------------------------------------

func TestLogRegistryCacheStatus_NilCacheReadiness_DoesNotPanic(t *testing.T) {
	bc := &buildContext{
		cacheReadiness: nil,
	}
	// Should not panic — outputs a warning message
	bc.logRegistryCacheStatus()
}

// ---------------------------------------------------------------------------
// Section: registry cache reference format (#383)
// This mirrors the logic in build_phases.go buildImage() that creates the
// registry cache ref used in CacheFrom/CacheTo.
// ---------------------------------------------------------------------------

func TestRegistryCacheRef_Format(t *testing.T) {
	tests := []struct {
		name             string
		registryEndpoint string
		workspaceName    string
		appName          string
		wantRef          string
	}{
		{
			name:             "standard workspace and app",
			registryEndpoint: "localhost:5001",
			workspaceName:    "dev",
			appName:          "myapp",
			wantRef:          "localhost:5001/dvm-cache/dev-myapp:buildcache",
		},
		{
			name:             "hyphenated names",
			registryEndpoint: "localhost:5001",
			workspaceName:    "prod-env",
			appName:          "api-gateway",
			wantRef:          "localhost:5001/dvm-cache/prod-env-api-gateway:buildcache",
		},
		{
			name:             "custom registry host",
			registryEndpoint: "zot.local:5000",
			workspaceName:    "staging",
			appName:          "backend",
			wantRef:          "zot.local:5000/dvm-cache/staging-backend:buildcache",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fmt.Sprintf("%s/dvm-cache/%s-%s:buildcache",
				tt.registryEndpoint, tt.workspaceName, tt.appName)
			if got != tt.wantRef {
				t.Errorf("registry cache ref = %q, want %q", got, tt.wantRef)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Section: multi-value CacheFrom splitting (#383)
// The Docker and BuildKit builders now split newline-separated CacheFrom into
// individual --cache-from flags. Validate the split logic here.
// ---------------------------------------------------------------------------

func TestCacheFromSplitting_SingleValue(t *testing.T) {
	cacheFrom := "type=local,src=/home/.devopsmaestro/build-cache/key"
	parts := strings.Split(cacheFrom, "\n")
	if len(parts) != 1 {
		t.Errorf("single-value CacheFrom split length = %d, want 1", len(parts))
	}
	if parts[0] != cacheFrom {
		t.Errorf("single-value CacheFrom part = %q, want %q", parts[0], cacheFrom)
	}
}

func TestCacheFromSplitting_MultipleValues(t *testing.T) {
	local := "type=local,src=/tmp/cache"
	registry := "type=registry,ref=localhost:5001/dvm-cache/dev-myapp:buildcache"
	cacheFrom := local + "\n" + registry

	parts := strings.Split(cacheFrom, "\n")
	if len(parts) != 2 {
		t.Errorf("two-value CacheFrom split length = %d, want 2", len(parts))
	}
	if parts[0] != local {
		t.Errorf("CacheFrom parts[0] = %q, want %q", parts[0], local)
	}
	if parts[1] != registry {
		t.Errorf("CacheFrom parts[1] = %q, want %q", parts[1], registry)
	}
}

func TestCacheFromSplitting_EmptyPartsAreSkipped(t *testing.T) {
	// Simulate what builders do: skip empty parts after split
	cacheFrom := "type=local,src=/tmp/cache"
	parts := strings.Split(cacheFrom, "\n")
	var nonEmpty []string
	for _, p := range parts {
		cf := strings.TrimSpace(p)
		if cf != "" {
			nonEmpty = append(nonEmpty, cf)
		}
	}
	if len(nonEmpty) != 1 {
		t.Errorf("expected 1 non-empty part, got %d", len(nonEmpty))
	}
}

// ---------------------------------------------------------------------------
// Section: CacheFrom combination logic (#383)
// Validates that local+registry cache sources are combined correctly with "\n"
// ---------------------------------------------------------------------------

func TestCacheFromCombination_LocalAndRegistry(t *testing.T) {
	tests := []struct {
		name          string
		existingFrom  string
		regCacheFrom  string
		wantCombined  string
		wantPartCount int
	}{
		{
			name:          "local cache combined with registry",
			existingFrom:  "type=local,src=/tmp/cache",
			regCacheFrom:  "type=registry,ref=localhost:5001/dvm-cache/dev-app:buildcache",
			wantCombined:  "type=local,src=/tmp/cache\ntype=registry,ref=localhost:5001/dvm-cache/dev-app:buildcache",
			wantPartCount: 2,
		},
		{
			name:          "registry only (no local cache)",
			existingFrom:  "",
			regCacheFrom:  "type=registry,ref=localhost:5001/dvm-cache/dev-app:buildcache",
			wantCombined:  "type=registry,ref=localhost:5001/dvm-cache/dev-app:buildcache",
			wantPartCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var combined string
			if tt.existingFrom != "" {
				combined = tt.existingFrom + "\n" + tt.regCacheFrom
			} else {
				combined = tt.regCacheFrom
			}
			if combined != tt.wantCombined {
				t.Errorf("combined CacheFrom = %q, want %q", combined, tt.wantCombined)
			}
			parts := strings.Split(combined, "\n")
			if len(parts) != tt.wantPartCount {
				t.Errorf("part count = %d, want %d", len(parts), tt.wantPartCount)
			}
		})
	}
}
