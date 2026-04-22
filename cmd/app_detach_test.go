package cmd

import (
	"strings"
	"testing"
)

// =============================================================================
// app detach CLI smoke tests
// =============================================================================

func TestAppDetachCmd_RequiredFlag_FromSystem(t *testing.T) {
	appDetachFromSystem = false
	appDetachEcosystem = ""
	appDetachDryRun = false

	err := appDetachCmd.RunE(appDetachCmd, []string{"checkout"})
	if err == nil {
		t.Fatal("expected error when --from-system is missing")
	}
	if !strings.Contains(err.Error(), "--from-system") {
		t.Errorf("error %q should mention --from-system", err.Error())
	}
}

func TestAppDetachCmd_HasExpectedFlags(t *testing.T) {
	flags := []string{"from-system", "ecosystem", "dry-run", "force", "output"}
	for _, f := range flags {
		if appDetachCmd.Flags().Lookup(f) == nil {
			t.Errorf("app detach command missing flag --%s", f)
		}
	}
	if appDetachCmd.Flags().ShorthandLookup("e") == nil {
		t.Error("app detach command missing -e shorthand for --ecosystem")
	}
}

func TestAppDetachCmd_DryRun_FlagRegistered(t *testing.T) {
	// The dry-run flag must be registered; the actual invocation skipping
	// DB access is tested at handler level. Here we only confirm the flag exists.
	if appDetachCmd.Flags().Lookup("dry-run") == nil {
		t.Error("app detach command missing --dry-run flag")
	}
}

func TestAppCmd_HasDetachSubcommand(t *testing.T) {
	found := false
	for _, sub := range appCmd.Commands() {
		if sub.Use == "detach <name>" {
			found = true
			break
		}
	}
	if !found {
		t.Error("app command should have 'detach' subcommand")
	}
}
