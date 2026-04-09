package cmd

// Tests for Issue #207: Theme Flag (TDD Phase 2 — RED)
//
// These tests drive the creation of:
//   - A persistent --theme flag on rootCmd (global, affects all rendered output)
//   - Default: empty string (uses config/env chain)
//   - Valid theme names are accepted
//   - Priority chain: --theme flag > DVM_THEME env > config > default (catppuccin-mocha)
//
// All tests are expected to FAIL until the implementation is added.

import (
	"testing"
)

// ---------------------------------------------------------------------------
// 1. rootCmd persistent --theme flag exists
// ---------------------------------------------------------------------------

func TestRootCmd_HasPersistentThemeFlag(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("theme")
	if flag == nil {
		t.Fatal("rootCmd should have a persistent --theme flag, but it was not found")
	}
}

func TestRootCmd_ThemeFlag_DefaultIsEmpty(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("theme")
	if flag == nil {
		t.Fatal("rootCmd --theme flag not found")
	}
	// Default should be empty string — uses config/env chain to resolve theme
	if flag.DefValue != "" {
		t.Errorf("rootCmd --theme default = %q, want empty string (resolves via config chain)", flag.DefValue)
	}
}

func TestRootCmd_ThemeFlag_TypeIsString(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("theme")
	if flag == nil {
		t.Fatal("rootCmd --theme flag not found")
	}
	if flag.Value.Type() != "string" {
		t.Errorf("rootCmd --theme flag type = %q, want %q", flag.Value.Type(), "string")
	}
}

// ---------------------------------------------------------------------------
// 2. Valid theme names are accepted
// ---------------------------------------------------------------------------

func TestRootCmd_ThemeFlag_AcceptsCatppuccinMocha(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("theme")
	if flag == nil {
		t.Fatal("rootCmd --theme flag not found")
	}
	if err := flag.Value.Set("catppuccin-mocha"); err != nil {
		t.Errorf("--theme catppuccin-mocha should be accepted, got error: %v", err)
	}
}

func TestRootCmd_ThemeFlag_AcceptsTokyonight(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("theme")
	if flag == nil {
		t.Fatal("rootCmd --theme flag not found")
	}
	if err := flag.Value.Set("tokyonight"); err != nil {
		t.Errorf("--theme tokyonight should be accepted, got error: %v", err)
	}
}

func TestRootCmd_ThemeFlag_AcceptsArbitraryThemeName(t *testing.T) {
	// Theme names are arbitrary strings — don't validate against a fixed list
	flag := rootCmd.PersistentFlags().Lookup("theme")
	if flag == nil {
		t.Fatal("rootCmd --theme flag not found")
	}
	if err := flag.Value.Set("custom-dark-theme"); err != nil {
		t.Errorf("--theme should accept arbitrary theme names, got error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// 3. Theme flag is persistent — inherited by subcommands
// ---------------------------------------------------------------------------

func TestGetCmd_InheritsThemeFlagFromRoot(t *testing.T) {
	// getCmd should inherit --theme from rootCmd's persistent flags
	flag := getCmd.InheritedFlags().Lookup("theme")
	if flag == nil {
		t.Fatal("getCmd should inherit --theme from rootCmd persistent flags")
	}
}

// ---------------------------------------------------------------------------
// 4. Global outputFormat variable reflects --output default of "table"
// ---------------------------------------------------------------------------

func TestOutputFormat_GlobalVariable_DefaultIsTable(t *testing.T) {
	// After unification, the global output format variable should default to "table"
	// This is checked by looking at the rootCmd persistent flag default, not a
	// package-level var, since the unification may change how the value is stored.
	flag := rootCmd.PersistentFlags().Lookup("output")
	if flag == nil {
		t.Fatal("rootCmd --output flag not found")
	}
	if flag.DefValue != "table" {
		t.Errorf("Global output format default should be 'table', got %q", flag.DefValue)
	}
}
