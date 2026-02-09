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
		// Get command aliases (Apps replaced Projects)
		{"get apps", getAppsCmd, []string{"app", "application", "applications"}},
		{"get app", getAppCmd, []string{"application"}},
		{"get workspaces", getWorkspacesCmd, []string{"ws"}},
		{"get workspace", getWorkspaceCmd, []string{"ws"}},
		{"get context", getContextCmd, []string{"ctx"}},
		{"get platforms", getPlatformsCmd, []string{"plat"}},

		// Create command aliases (Apps replaced Projects)
		{"create app", createAppCmd, []string{"application"}},
		{"create workspace", createWorkspaceCmd, []string{"ws"}},

		// Delete command aliases (Apps replaced Projects)
		{"delete app", deleteAppCmd, []string{"application"}},
		{"delete workspace", deleteWorkspaceCmd, []string{"ws"}},

		// Use command aliases
		{"use app", useAppCmd, []string{"a"}},
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
		// Apps replaced Projects - check for "app" or "a" aliases
		{"get command help contains app", getCmd.Long, "app"},
		{"get command help contains ws", getCmd.Long, "ws"},
		{"get command help contains ctx", getCmd.Long, "ctx"},
		{"create command help contains app", createCmd.Long, "app"},
		{"create command help contains ws", createCmd.Long, "ws"},
		{"delete command help contains app", deleteCmd.Long, "app"},
		{"delete command help contains ws", deleteCmd.Long, "ws"},
		{"use command help contains app", useCmd.Long, "app"},
		{"use command help contains ws", useCmd.Long, "ws"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	// NOTE: Apps replaced Projects in the migration
	aliases := map[string]string{
		"apps":       "app, application, applications",
		"app":        "application",
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
	// NOTE: Apps replaced Projects - now using "app" and "application" aliases
	tests := []struct {
		parent *cobra.Command
		alias  string
	}{
		{getCmd, "app"},
		{getCmd, "ws"},
		{getCmd, "ctx"},
		{getCmd, "plat"},
		{createCmd, "application"},
		{createCmd, "ws"},
		{deleteCmd, "application"},
		{deleteCmd, "ws"},
		{useCmd, "a"},
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
