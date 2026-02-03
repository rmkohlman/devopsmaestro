package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestNvimCommandsStructure verifies the namespaced nvim command hierarchy
func TestNvimCommandsStructure(t *testing.T) {
	// Test 'get nvim' command structure
	t.Run("get nvim command hierarchy", func(t *testing.T) {
		// Verify nvimGetCmd exists and has expected subcommands
		assert.NotNil(t, nvimGetCmd, "nvimGetCmd should exist")
		assert.Equal(t, "nvim", nvimGetCmd.Use, "nvimGetCmd should have Use 'nvim'")

		// Check subcommands exist
		subcommands := nvimGetCmd.Commands()
		subcommandNames := make([]string, len(subcommands))
		for i, cmd := range subcommands {
			subcommandNames[i] = cmd.Name()
		}

		assert.Contains(t, subcommandNames, "plugins", "should have 'plugins' subcommand")
		assert.Contains(t, subcommandNames, "plugin", "should have 'plugin' subcommand")
		assert.Contains(t, subcommandNames, "themes", "should have 'themes' subcommand")
		assert.Contains(t, subcommandNames, "theme", "should have 'theme' subcommand")
	})

	// Test 'apply nvim' command structure
	t.Run("apply nvim command hierarchy", func(t *testing.T) {
		assert.NotNil(t, applyNvimCmd, "applyNvimCmd should exist")
		assert.Equal(t, "nvim", applyNvimCmd.Use, "applyNvimCmd should have Use 'nvim'")

		subcommands := applyNvimCmd.Commands()
		subcommandNames := make([]string, len(subcommands))
		for i, cmd := range subcommands {
			subcommandNames[i] = cmd.Name()
		}

		assert.Contains(t, subcommandNames, "plugin", "should have 'plugin' subcommand")
		assert.Contains(t, subcommandNames, "theme", "should have 'theme' subcommand")
	})

	// Test 'delete nvim' command structure
	t.Run("delete nvim command hierarchy", func(t *testing.T) {
		assert.NotNil(t, deleteNvimCmd, "deleteNvimCmd should exist")
		assert.Equal(t, "nvim", deleteNvimCmd.Use, "deleteNvimCmd should have Use 'nvim'")

		subcommands := deleteNvimCmd.Commands()
		subcommandNames := make([]string, len(subcommands))
		for i, cmd := range subcommands {
			subcommandNames[i] = cmd.Name()
		}

		assert.Contains(t, subcommandNames, "plugin", "should have 'plugin' subcommand")
		assert.Contains(t, subcommandNames, "theme", "should have 'theme' subcommand")
	})

	// Test 'edit nvim' command structure
	t.Run("edit nvim command hierarchy", func(t *testing.T) {
		assert.NotNil(t, editNvimCmd, "editNvimCmd should exist")
		assert.Equal(t, "nvim", editNvimCmd.Use, "editNvimCmd should have Use 'nvim'")

		subcommands := editNvimCmd.Commands()
		subcommandNames := make([]string, len(subcommands))
		for i, cmd := range subcommands {
			subcommandNames[i] = cmd.Name()
		}

		assert.Contains(t, subcommandNames, "plugin", "should have 'plugin' subcommand")
		assert.Contains(t, subcommandNames, "theme", "should have 'theme' subcommand")
	})
}

// TestNvimCommandsHelp verifies help text is displayed properly
func TestNvimCommandsHelp(t *testing.T) {
	tests := []struct {
		name           string
		cmd            *cobra.Command
		expectedUsage  string
		expectedInHelp []string
	}{
		{
			name:          "get nvim help",
			cmd:           nvimGetCmd,
			expectedUsage: "nvim",
			expectedInHelp: []string{
				"plugins",
				"plugin",
				"themes",
				"theme",
			},
		},
		{
			name:          "apply nvim help",
			cmd:           applyNvimCmd,
			expectedUsage: "nvim",
			expectedInHelp: []string{
				"plugin",
				"theme",
			},
		},
		{
			name:          "delete nvim help",
			cmd:           deleteNvimCmd,
			expectedUsage: "nvim",
			expectedInHelp: []string{
				"plugin",
				"theme",
			},
		},
		{
			name:          "edit nvim help",
			cmd:           editNvimCmd,
			expectedUsage: "nvim",
			expectedInHelp: []string{
				"plugin",
				"theme",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedUsage, tt.cmd.Use)

			// Get help text
			buf := new(bytes.Buffer)
			tt.cmd.SetOut(buf)
			tt.cmd.Help()
			helpText := buf.String()

			for _, expected := range tt.expectedInHelp {
				assert.Contains(t, helpText, expected, "help should contain '%s'", expected)
			}
		})
	}
}

// TestApplyNvimPluginRequiresFilename verifies the -f flag is required
func TestApplyNvimPluginRequiresFilename(t *testing.T) {
	// applyNvimPluginCmd should require the filename flag
	filenameFlag := applyNvimPluginCmd.Flags().Lookup("filename")
	assert.NotNil(t, filenameFlag, "should have 'filename' flag")
	assert.Equal(t, "f", filenameFlag.Shorthand, "filename flag should have 'f' shorthand")
}

// TestDeleteNvimPluginRequiresName verifies the command requires a name argument
func TestDeleteNvimPluginRequiresName(t *testing.T) {
	// deleteNvimPluginCmd should require exactly 1 argument
	assert.NotNil(t, deleteNvimPluginCmd.Args, "should have Args validator")
}

// TestEditNvimPluginRequiresName verifies the command requires a name argument
func TestEditNvimPluginRequiresName(t *testing.T) {
	// editNvimPluginCmd should require exactly 1 argument
	assert.NotNil(t, editNvimPluginCmd.Args, "should have Args validator")
}

// TestDeleteNvimPluginHasForceFlag verifies the --force flag exists
func TestDeleteNvimPluginHasForceFlag(t *testing.T) {
	forceFlag := deleteNvimPluginCmd.Flags().Lookup("force")
	assert.NotNil(t, forceFlag, "should have 'force' flag")
	assert.Equal(t, "f", forceFlag.Shorthand, "force flag should have 'f' shorthand")
}

// TestNoDeprecatedPluginCommandsInGet verifies deprecated commands were removed
func TestNoDeprecatedPluginCommandsInGet(t *testing.T) {
	subcommands := getCmd.Commands()
	for _, cmd := range subcommands {
		// 'plugins' and 'plugin' should NOT exist as direct children of 'get'
		// They should only exist under 'get nvim'
		if cmd.Name() != "nvim" {
			assert.NotEqual(t, "plugins", cmd.Name(), "'dvm get plugins' should be removed (use 'dvm get nvim plugins')")
			assert.NotEqual(t, "plugin", cmd.Name(), "'dvm get plugin' should be removed (use 'dvm get nvim plugin')")
		}
	}
}

// TestKubectlStyleCommandPatterns verifies kubectl-style command patterns
func TestKubectlStyleCommandPatterns(t *testing.T) {
	// Verify the command patterns follow kubectl conventions:
	// dvm <verb> <resource-type> <resource-name> [flags]

	// get nvim plugins (list)
	assert.Equal(t, "plugins", nvimGetPluginsCmd.Use)
	assert.Contains(t, nvimGetPluginsCmd.Short, "List")

	// get nvim plugin <name> (get specific)
	assert.Equal(t, "plugin [name]", nvimGetPluginCmd.Use)
	assert.Contains(t, nvimGetPluginCmd.Short, "specific")

	// apply nvim plugin (apply from file)
	assert.Equal(t, "plugin", applyNvimPluginCmd.Use)

	// delete nvim plugin <name>
	assert.Equal(t, "plugin [name]", deleteNvimPluginCmd.Use)

	// edit nvim plugin <name>
	assert.Equal(t, "plugin [name]", editNvimPluginCmd.Use)
}

// TestNvimPluginAliases verifies short aliases exist
func TestNvimPluginAliases(t *testing.T) {
	// nvimGetPluginsCmd should have 'np' alias
	assert.Contains(t, nvimGetPluginsCmd.Aliases, "np", "plugins should have 'np' alias")

	// nvimGetThemesCmd should have 'nt' alias
	assert.Contains(t, nvimGetThemesCmd.Aliases, "nt", "themes should have 'nt' alias")
}
