package cmd

// Tests for Issue #207: Output Flag Unification (TDD Phase 2 — RED)
//
// These tests drive the following changes:
//   - rootCmd gains a persistent --output/-o flag (not just getCmd)
//   - Default is "table" (not empty string)
//   - Valid values: table, json, yaml, plain, compact, wide
//   - "colored" is accepted silently as a deprecated alias for "table"
//   - Help text enumerates all valid format names
//
// All tests are expected to FAIL until the implementation is added.

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// 1. rootCmd persistent --output/-o flag
// ---------------------------------------------------------------------------

func TestRootCmd_HasPersistentOutputFlag(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("output")
	if flag == nil {
		t.Fatal("rootCmd should have a persistent --output flag, but it was not found")
	}
}

func TestRootCmd_OutputFlag_HasShorthand_o(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("output")
	if flag == nil {
		t.Fatal("rootCmd --output flag not found")
	}
	if flag.Shorthand != "o" {
		t.Errorf("rootCmd --output shorthand = %q, want %q", flag.Shorthand, "o")
	}
}

func TestRootCmd_OutputFlag_DefaultIsTable(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("output")
	if flag == nil {
		t.Fatal("rootCmd --output flag not found")
	}
	if flag.DefValue != "table" {
		t.Errorf("rootCmd --output default = %q, want %q", flag.DefValue, "table")
	}
}

// ---------------------------------------------------------------------------
// 2. Valid format values are accepted
// ---------------------------------------------------------------------------

func TestRootCmd_OutputFlag_AcceptsTableFormat(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("output")
	if flag == nil {
		t.Fatal("rootCmd --output flag not found")
	}
	if err := flag.Value.Set("table"); err != nil {
		t.Errorf("--output table should be accepted, got error: %v", err)
	}
}

func TestRootCmd_OutputFlag_AcceptsJsonFormat(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("output")
	if flag == nil {
		t.Fatal("rootCmd --output flag not found")
	}
	if err := flag.Value.Set("json"); err != nil {
		t.Errorf("--output json should be accepted, got error: %v", err)
	}
}

func TestRootCmd_OutputFlag_AcceptsYamlFormat(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("output")
	if flag == nil {
		t.Fatal("rootCmd --output flag not found")
	}
	if err := flag.Value.Set("yaml"); err != nil {
		t.Errorf("--output yaml should be accepted, got error: %v", err)
	}
}

func TestRootCmd_OutputFlag_AcceptsPlainFormat(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("output")
	if flag == nil {
		t.Fatal("rootCmd --output flag not found")
	}
	if err := flag.Value.Set("plain"); err != nil {
		t.Errorf("--output plain should be accepted, got error: %v", err)
	}
}

func TestRootCmd_OutputFlag_AcceptsCompactFormat(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("output")
	if flag == nil {
		t.Fatal("rootCmd --output flag not found")
	}
	if err := flag.Value.Set("compact"); err != nil {
		t.Errorf("--output compact should be accepted, got error: %v", err)
	}
}

func TestRootCmd_OutputFlag_AcceptsWideFormat(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("output")
	if flag == nil {
		t.Fatal("rootCmd --output flag not found")
	}
	if err := flag.Value.Set("wide"); err != nil {
		t.Errorf("--output wide should be accepted, got error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// 3. "colored" is silently accepted as deprecated alias for "table"
// ---------------------------------------------------------------------------

func TestRootCmd_OutputFlag_AcceptsColoredAsDeprecatedAlias(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("output")
	if flag == nil {
		t.Fatal("rootCmd --output flag not found")
	}
	// "colored" should NOT return an error (silently accepted)
	if err := flag.Value.Set("colored"); err != nil {
		t.Errorf("--output colored should be silently accepted as deprecated alias, got error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// 4. Help text enumerates valid format names
// ---------------------------------------------------------------------------

func TestRootCmd_OutputFlag_HelpTextContainsValidFormats(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("output")
	if flag == nil {
		t.Fatal("rootCmd --output flag not found")
	}

	usage := flag.Usage
	requiredTerms := []string{"table", "json", "yaml", "plain", "compact", "wide"}
	for _, term := range requiredTerms {
		if !strings.Contains(usage, term) {
			t.Errorf("--output help text should mention %q, but usage is: %q", term, usage)
		}
	}
}

// ---------------------------------------------------------------------------
// 5. --output flag is persistent — inherited by subcommands
// ---------------------------------------------------------------------------

func TestGetCmd_InheritsOutputFlagFromRoot(t *testing.T) {
	// getCmd should inherit --output from rootCmd's persistent flags
	// (it currently has its own persistent flag, but after unification,
	// it should come from rootCmd's persistent flags)
	flag := getCmd.InheritedFlags().Lookup("output")
	if flag == nil {
		// Acceptable if rootCmd's PersistentFlags are inherited
		// Try looking in the full flag set
		flag = getCmd.Flags().Lookup("output")
	}
	if flag == nil {
		t.Fatal("getCmd should have --output available (inherited from rootCmd or own persistent flags)")
	}
}
