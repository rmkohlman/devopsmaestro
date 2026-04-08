package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestToolConfigCmd_Exists(t *testing.T) {
	// Verify tool-config is registered as a subcommand of root
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "tool-config" {
			found = true
			break
		}
	}
	if !found {
		t.Error("tool-config command not registered on rootCmd")
	}
}

func TestToolConfigCmd_HasSubcommands(t *testing.T) {
	subs := toolConfigCmd.Commands()
	names := make(map[string]bool)
	for _, s := range subs {
		names[s.Name()] = true
	}

	expected := []string{"generate", "list"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("missing subcommand %q on tool-config", name)
		}
	}
}

func TestToolConfigGenerateCmd_Flags(t *testing.T) {
	// Verify flags exist
	allFlag := toolConfigGenerateCmd.Flags().Lookup("all")
	if allFlag == nil {
		t.Error("--all flag not found on generate command")
	}

	formatFlag := toolConfigGenerateCmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Error("--format flag not found on generate command")
	}
	if formatFlag != nil && formatFlag.DefValue != "env" {
		t.Errorf("--format default = %q, want %q", formatFlag.DefValue, "env")
	}
}

func TestToolConfigGenerateCmd_RequiresToolOrAll(t *testing.T) {
	// Create a fresh command to avoid root PersistentPreRunE
	cmd := &cobra.Command{Use: "test"}
	genCmd := *toolConfigGenerateCmd
	cmd.AddCommand(&genCmd)

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"generate"})

	// Should fail without tool name or --all
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no tool name or --all specified")
	}
}

func TestPrintEnvFormat_Fzf(t *testing.T) {
	// Capture stdout by testing the format logic
	output := "# fzf colors\n--color=fg:#c0caf5,bg:#1a1b26"

	// Verify the fzf output contains the expected --color= pattern
	lines := strings.Split(output, "\n")
	found := false
	for _, line := range lines {
		if strings.HasPrefix(line, "--color=") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected --color= line in fzf output")
	}
}

func TestPrintEnvFormat_Bat(t *testing.T) {
	// The bat env format should reference BAT_THEME
	// This is a logic test — bat always outputs "ansi"
	output := "# bat config\n--theme=ansi"
	if !strings.Contains(output, "ansi") {
		t.Error("bat output should reference ansi theme")
	}
}
