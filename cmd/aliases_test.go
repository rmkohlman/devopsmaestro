package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestResourceAliases verifies that all kubectl-style resource aliases are configured correctly
func TestResourceAliases(t *testing.T) {
	tests := []struct {
		name     string
		cmd      *cobra.Command
		expected []string
	}{
		// Get command aliases
		{"get projects", getProjectsCmd, []string{"proj"}},
		{"get project", getProjectCmd, []string{"proj"}},
		{"get workspaces", getWorkspacesCmd, []string{"ws"}},
		{"get workspace", getWorkspaceCmd, []string{"ws"}},
		{"get context", getContextCmd, []string{"ctx"}},
		{"get platforms", getPlatformsCmd, []string{"plat"}},

		// Create command aliases
		{"create project", createProjectCmd, []string{"proj"}},
		{"create workspace", createWorkspaceCmd, []string{"ws"}},

		// Delete command aliases
		{"delete project", deleteProjectCmd, []string{"proj"}},
		{"delete workspace", deleteWorkspaceCmd, []string{"ws"}},

		// Use command aliases
		{"use project", useProjectCmd, []string{"proj"}},
		{"use workspace", useWorkspaceCmd, []string{"ws"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aliases := tt.cmd.Aliases
			if len(aliases) != len(tt.expected) {
				t.Errorf("expected %d aliases, got %d", len(tt.expected), len(aliases))
				return
			}
			for i, expected := range tt.expected {
				if aliases[i] != expected {
					t.Errorf("alias[%d]: expected %q, got %q", i, expected, aliases[i])
				}
			}
		})
	}
}

// TestAliasesInHelpText verifies that help text mentions aliases
func TestAliasesInHelpText(t *testing.T) {
	tests := []struct {
		name        string
		helpText    string
		shouldMatch string
	}{
		{"get command help", getCmd.Long, "proj"},
		{"get command help", getCmd.Long, "ws"},
		{"get command help", getCmd.Long, "ctx"},
		{"create command help", createCmd.Long, "proj"},
		{"create command help", createCmd.Long, "ws"},
		{"delete command help", deleteCmd.Long, "proj"},
		{"delete command help", deleteCmd.Long, "ws"},
		{"use command help", useCmd.Long, "proj"},
		{"use command help", useCmd.Long, "ws"},
	}

	for _, tt := range tests {
		t.Run(tt.name+" contains "+tt.shouldMatch, func(t *testing.T) {
			if !strings.Contains(tt.helpText, tt.shouldMatch) {
				t.Errorf("help text should contain %q", tt.shouldMatch)
			}
		})
	}
}

// TestAliasReference documents all supported aliases for reference
func TestAliasReference(t *testing.T) {
	// This test documents all supported kubectl-style aliases
	// Format: full name → alias
	aliases := map[string]string{
		"projects":   "proj",
		"project":    "proj",
		"workspaces": "ws",
		"workspace":  "ws",
		"context":    "ctx",
		"platforms":  "plat",
	}

	t.Logf("Supported kubectl-style resource aliases:")
	for full, short := range aliases {
		t.Logf("  %s → %s", full, short)
	}

	// Verify the count
	expectedCount := 6
	if len(aliases) != expectedCount {
		t.Errorf("expected %d aliases, got %d", expectedCount, len(aliases))
	}
}

// TestAliasCommandExecution tests that aliases work through command execution
func TestAliasCommandExecution(t *testing.T) {
	// Verify that the subcommands are properly registered with their aliases
	// Note: For commands where both singular and plural share the same alias (proj, ws),
	// we just verify that an alias exists, not which specific command it maps to
	tests := []struct {
		parent *cobra.Command
		alias  string
	}{
		{getCmd, "proj"},
		{getCmd, "ws"},
		{getCmd, "ctx"},
		{getCmd, "plat"},
		{createCmd, "proj"},
		{createCmd, "ws"},
		{deleteCmd, "proj"},
		{deleteCmd, "ws"},
		{useCmd, "proj"},
		{useCmd, "ws"},
	}

	for _, tt := range tests {
		t.Run(tt.parent.Use+" "+tt.alias, func(t *testing.T) {
			// Find command by alias
			var found *cobra.Command
			for _, sub := range tt.parent.Commands() {
				for _, alias := range sub.Aliases {
					if alias == tt.alias {
						found = sub
						break
					}
				}
				if found != nil {
					break
				}
			}

			if found == nil {
				t.Errorf("alias %q not found in %s subcommands", tt.alias, tt.parent.Use)
			}
		})
	}
}
