package cmd

import (
	"strings"
	"testing"
)

// =============================================================================
// move system CLI smoke tests
// =============================================================================

// TestMoveSystemCmd_FlagWiring verifies that move system registers the expected
// flags and that the required --to-domain check fires before any DB access.
func TestMoveSystemCmd_RequiredFlag_ToDomain(t *testing.T) {
	// Reset global flag vars so previous test runs don't bleed over.
	moveToDomain = ""
	moveEcosystem = ""
	moveSystemDryRun = false

	err := moveSystemCmd.RunE(moveSystemCmd, []string{"payments"})
	if err == nil {
		t.Fatal("expected error when --to-domain is missing")
	}
	if !strings.Contains(err.Error(), "--to-domain") {
		t.Errorf("error %q should mention --to-domain", err.Error())
	}
}

func TestMoveSystemCmd_HasExpectedFlags(t *testing.T) {
	flags := []string{"to-domain", "ecosystem", "dry-run", "output"}
	for _, f := range flags {
		if moveSystemCmd.Flags().Lookup(f) == nil {
			t.Errorf("move system command missing flag --%s", f)
		}
	}
	// -e shorthand for --ecosystem
	if moveSystemCmd.Flags().ShorthandLookup("e") == nil {
		t.Error("move system command missing -e shorthand for --ecosystem")
	}
}

func TestMoveSystemCmd_DryRun_FlagRegistered(t *testing.T) {
	if moveSystemCmd.Flags().Lookup("dry-run") == nil {
		t.Error("move system command missing --dry-run flag")
	}
}

// =============================================================================
// move app CLI smoke tests
// =============================================================================

func TestMoveAppCmd_RequiredFlags_NeitherProvided(t *testing.T) {
	moveToSystem = ""
	moveToDomain = ""
	moveEcosystem = ""
	moveAppDryRun = false

	err := moveAppCmd.RunE(moveAppCmd, []string{"checkout"})
	if err == nil {
		t.Fatal("expected error when neither --to-system nor --to-domain provided")
	}
	if !strings.Contains(err.Error(), "--to-system") && !strings.Contains(err.Error(), "--to-domain") {
		t.Errorf("error %q should mention --to-system or --to-domain", err.Error())
	}
}

func TestMoveAppCmd_HasExpectedFlags(t *testing.T) {
	flags := []string{"to-system", "to-domain", "ecosystem", "dry-run", "output"}
	for _, f := range flags {
		if moveAppCmd.Flags().Lookup(f) == nil {
			t.Errorf("move app command missing flag --%s", f)
		}
	}
	if moveAppCmd.Flags().ShorthandLookup("e") == nil {
		t.Error("move app command missing -e shorthand for --ecosystem")
	}
}

func TestMoveAppCmd_DryRun_FlagRegistered(t *testing.T) {
	if moveAppCmd.Flags().Lookup("dry-run") == nil {
		t.Error("move app command missing --dry-run flag")
	}
}

func TestMoveAppCmd_DryRun_WithDomainOnly_FlagRegistered(t *testing.T) {
	if moveAppCmd.Flags().Lookup("dry-run") == nil || moveAppCmd.Flags().Lookup("to-domain") == nil {
		t.Error("move app command missing --dry-run or --to-domain flag")
	}
}
