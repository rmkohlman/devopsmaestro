package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDryRunFlagRegistered verifies that every mutating command has a --dry-run
// flag registered with default=false.
func TestDryRunFlagRegistered(t *testing.T) {
	tests := []struct {
		name string
		cmd  *cobra.Command
	}{
		// Lifecycle commands
		{"init", initCmd},
		{"build", buildCmd},
		{"attach", attachCmd},
		{"detach", detachCmd},
		{"cache clear", cacheClearCmd},

		// Create commands
		{"create ecosystem", createEcosystemCmd},
		{"create domain", createDomainCmd},
		{"create app", createAppCmd},
		{"create workspace", createWorkspaceCmd},
		{"create gitrepo", createGitRepoCmd},
		{"create credential", createCredentialCmd},

		// Use commands
		{"use ecosystem", useEcosystemCmd},
		{"use domain", useDomainCmd},
		{"use app", useAppCmd},
		{"use workspace", useWorkspaceCmd},

		// Delete commands
		{"delete ecosystem", deleteEcosystemCmd},
		{"delete domain", deleteDomainCmd},
		{"delete app", deleteAppCmd},
		{"delete workspace", deleteWorkspaceCmd},
		{"delete credential", deleteCredentialCmd},
		{"delete ca-cert", deleteCACertCmd},
		{"delete build-arg", deleteBuildArgCmd},
		{"delete gitrepo", deleteGitRepoCmd},

		// Set commands
		{"set credential", setCredentialCmd},
		{"set ca-cert", setCACertCmd},
		{"set build-arg", setBuildArgCmd},
		{"set theme", setThemeCmd},
		{"set nvim plugin", setNvimPluginCmd},
		{"set terminal prompt", setTerminalPromptCmd},
		{"set terminal plugin", setTerminalPluginCmd},
		{"set terminal package", setTerminalPackageCmd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := tt.cmd.Flags().Lookup("dry-run")
			require.NotNil(t, flag, "command %q should have --dry-run flag", tt.name)
			assert.Equal(t, "false", flag.DefValue, "command %q --dry-run default should be false", tt.name)
			assert.Equal(t, "bool", flag.Value.Type(), "command %q --dry-run should be bool", tt.name)
		})
	}
}
