package cmd

// =============================================================================
// TDD Phase 2 (RED): dvm cache command group (Issue #145)
// =============================================================================
// These tests drive the implementation of `dvm cache` and its subcommands.
// They WILL FAIL until cmd/cache.go is created with the required variables.
//
// Expected RED: undefined: cacheCmd, cacheClearCmd
// Expected GREEN: after dvm-core implements cmd/cache.go
// =============================================================================

import (
	"testing"
)

// TestCacheCmd_Exists verifies cacheCmd is declared and not nil.
func TestCacheCmd_Exists(t *testing.T) {
	if cacheCmd == nil {
		t.Fatal("cacheCmd must not be nil")
	}
}

// TestCacheCmd_Use verifies the Use field matches the kubectl-style verb.
func TestCacheCmd_Use(t *testing.T) {
	if cacheCmd.Use != "cache" {
		t.Errorf("cacheCmd.Use = %q, want %q", cacheCmd.Use, "cache")
	}
}

// TestCacheCmd_RegisteredOnRoot verifies cacheCmd is registered as a subcommand of rootCmd.
func TestCacheCmd_RegisteredOnRoot(t *testing.T) {
	for _, sub := range rootCmd.Commands() {
		if sub == cacheCmd {
			return
		}
	}
	t.Error("cacheCmd is not registered on rootCmd")
}

// TestCacheClearCmd_Exists verifies cacheClearCmd is declared and not nil.
func TestCacheClearCmd_Exists(t *testing.T) {
	if cacheClearCmd == nil {
		t.Fatal("cacheClearCmd must not be nil")
	}
}

// TestCacheClearCmd_Use verifies cacheClearCmd.Use is "clear".
func TestCacheClearCmd_Use(t *testing.T) {
	if cacheClearCmd.Use != "clear" {
		t.Errorf("cacheClearCmd.Use = %q, want %q", cacheClearCmd.Use, "clear")
	}
}

// TestCacheClearCmd_RegisteredOnCacheCmd verifies cacheClearCmd is a subcommand of cacheCmd.
func TestCacheClearCmd_RegisteredOnCacheCmd(t *testing.T) {
	for _, sub := range cacheCmd.Commands() {
		if sub == cacheClearCmd {
			return
		}
	}
	t.Error("cacheClearCmd is not registered on cacheCmd")
}

// TestCacheClearCmd_Flags verifies required flags are registered on cacheClearCmd.
func TestCacheClearCmd_Flags(t *testing.T) {
	tests := []struct {
		flag string
	}{
		{"all"},
		{"buildkit"},
		{"npm"},
		{"pip"},
		{"staging"},
		{"force"},
	}

	for _, tt := range tests {
		t.Run("flag --"+tt.flag, func(t *testing.T) {
			if f := cacheClearCmd.Flags().Lookup(tt.flag); f == nil {
				t.Errorf("flag --%s is not registered on cacheClearCmd", tt.flag)
			}
		})
	}
}
