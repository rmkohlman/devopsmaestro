package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestWorkspaceCommandsHaveProjectFlag verifies all workspace commands have the -p flag
func TestWorkspaceCommandsHaveProjectFlag(t *testing.T) {
	tests := []struct {
		name        string
		cmd         *cobra.Command
		cmdName     string
		description string
	}{
		{
			name:        "get workspaces has -p flag",
			cmd:         getWorkspacesCmd,
			cmdName:     "workspaces",
			description: "dvm get workspaces -p <project>",
		},
		{
			name:        "get workspace has -p flag",
			cmd:         getWorkspaceCmd,
			cmdName:     "workspace",
			description: "dvm get workspace <name> -p <project>",
		},
		{
			name:        "create workspace has -p flag",
			cmd:         createWorkspaceCmd,
			cmdName:     "workspace",
			description: "dvm create workspace <name> -p <project>",
		},
		{
			name:        "delete workspace has -p flag",
			cmd:         deleteWorkspaceCmd,
			cmdName:     "workspace",
			description: "dvm delete workspace <name> -p <project>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check flag exists
			projectFlag := tt.cmd.Flags().Lookup("project")
			assert.NotNil(t, projectFlag, "%s should have 'project' flag", tt.cmdName)

			if projectFlag != nil {
				// Check shorthand is 'p'
				assert.Equal(t, "p", projectFlag.Shorthand,
					"%s project flag should have 'p' shorthand", tt.cmdName)

				// Check flag is optional (has empty default)
				assert.Equal(t, "", projectFlag.DefValue,
					"%s project flag should default to empty (uses context)", tt.cmdName)
			}
		})
	}
}

// TestDeleteProjectCommandExists verifies delete project command is registered
func TestDeleteProjectCommandExists(t *testing.T) {
	// Check deleteProjectCmd exists and is properly configured
	assert.NotNil(t, deleteProjectCmd, "deleteProjectCmd should exist")
	assert.Equal(t, "project [name]", deleteProjectCmd.Use, "deleteProjectCmd should have correct Use")

	// Check it's a subcommand of deleteCmd
	subcommands := deleteCmd.Commands()
	found := false
	for _, cmd := range subcommands {
		if cmd.Name() == "project" {
			found = true
			break
		}
	}
	assert.True(t, found, "delete should have 'project' subcommand")
}

// TestDeleteWorkspaceCommandExists verifies delete workspace command is registered
func TestDeleteWorkspaceCommandExists(t *testing.T) {
	// Check deleteWorkspaceCmd exists and is properly configured
	assert.NotNil(t, deleteWorkspaceCmd, "deleteWorkspaceCmd should exist")
	assert.Equal(t, "workspace [name]", deleteWorkspaceCmd.Use, "deleteWorkspaceCmd should have correct Use")

	// Check it's a subcommand of deleteCmd
	subcommands := deleteCmd.Commands()
	found := false
	for _, cmd := range subcommands {
		if cmd.Name() == "workspace" {
			found = true
			break
		}
	}
	assert.True(t, found, "delete should have 'workspace' subcommand")
}

// TestDeleteCommandsHaveForceFlag verifies delete commands have --force flag
func TestDeleteCommandsHaveForceFlag(t *testing.T) {
	tests := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"delete project has --force", deleteProjectCmd},
		{"delete workspace has --force", deleteWorkspaceCmd},
		{"delete nvim plugin has --force", deleteNvimPluginCmd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			forceFlag := tt.cmd.Flags().Lookup("force")
			assert.NotNil(t, forceFlag, "should have 'force' flag")

			if forceFlag != nil {
				assert.Equal(t, "f", forceFlag.Shorthand, "force flag should have 'f' shorthand")
				assert.Equal(t, "false", forceFlag.DefValue, "force flag should default to false")
			}
		})
	}
}

// TestDeleteCommandHierarchy verifies the delete command structure
func TestDeleteCommandHierarchy(t *testing.T) {
	subcommands := deleteCmd.Commands()
	subcommandNames := make([]string, len(subcommands))
	for i, cmd := range subcommands {
		subcommandNames[i] = cmd.Name()
	}

	// Should have nvim, project, and workspace as direct children
	assert.Contains(t, subcommandNames, "nvim", "delete should have 'nvim' subcommand")
	assert.Contains(t, subcommandNames, "project", "delete should have 'project' subcommand")
	assert.Contains(t, subcommandNames, "workspace", "delete should have 'workspace' subcommand")
}

// TestDeleteCommandHelp verifies help text includes all options
func TestDeleteCommandHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	deleteCmd.SetOut(buf)
	deleteCmd.Help()
	helpText := buf.String()

	// Should mention all subcommands
	assert.Contains(t, helpText, "project", "help should mention 'project'")
	assert.Contains(t, helpText, "workspace", "help should mention 'workspace'")
	assert.Contains(t, helpText, "nvim", "help should mention 'nvim'")
}

// TestDeleteWorkspaceHelpShowsProjectFlag verifies -p flag is documented
func TestDeleteWorkspaceHelpShowsProjectFlag(t *testing.T) {
	buf := new(bytes.Buffer)
	deleteWorkspaceCmd.SetOut(buf)
	deleteWorkspaceCmd.Help()
	helpText := buf.String()

	// Should show -p flag in help
	assert.Contains(t, helpText, "-p", "help should show '-p' flag")
	assert.Contains(t, helpText, "--project", "help should show '--project' flag")
}

// TestGetWorkspacesHelpShowsProjectFlag verifies -p flag is documented
func TestGetWorkspacesHelpShowsProjectFlag(t *testing.T) {
	buf := new(bytes.Buffer)
	getWorkspacesCmd.SetOut(buf)
	getWorkspacesCmd.Help()
	helpText := buf.String()

	// Should show -p flag in help
	assert.Contains(t, helpText, "-p", "help should show '-p' flag")
	assert.Contains(t, helpText, "--project", "help should show '--project' flag")
}

// TestGetWorkspaceHelpShowsProjectFlag verifies -p flag is documented
func TestGetWorkspaceHelpShowsProjectFlag(t *testing.T) {
	buf := new(bytes.Buffer)
	getWorkspaceCmd.SetOut(buf)
	getWorkspaceCmd.Help()
	helpText := buf.String()

	// Should show -p flag in help
	assert.Contains(t, helpText, "-p", "help should show '-p' flag")
	assert.Contains(t, helpText, "--project", "help should show '--project' flag")
}

// TestCreateWorkspaceHelpShowsProjectFlag verifies -p flag is documented
func TestCreateWorkspaceHelpShowsProjectFlag(t *testing.T) {
	buf := new(bytes.Buffer)
	createWorkspaceCmd.SetOut(buf)
	createWorkspaceCmd.Help()
	helpText := buf.String()

	// Should show -p flag in help
	assert.Contains(t, helpText, "-p", "help should show '-p' flag")
	assert.Contains(t, helpText, "--project", "help should show '--project' flag")
}

// TestWorkspaceCommandsRequireName verifies workspace commands require name argument
func TestWorkspaceCommandsRequireName(t *testing.T) {
	// These commands should require exactly 1 argument
	cmdsRequiringName := []*cobra.Command{
		getWorkspaceCmd,
		deleteWorkspaceCmd,
		createWorkspaceCmd,
	}

	for _, cmd := range cmdsRequiringName {
		assert.NotNil(t, cmd.Args, "%s should have Args validator", cmd.Name())
	}
}

// TestProjectFlagDefaultsToEmpty verifies -p flag defaults to empty (uses context)
func TestProjectFlagDefaultsToEmpty(t *testing.T) {
	cmds := []*cobra.Command{
		getWorkspacesCmd,
		getWorkspaceCmd,
		createWorkspaceCmd,
		deleteWorkspaceCmd,
	}

	for _, cmd := range cmds {
		flag := cmd.Flags().Lookup("project")
		if flag != nil {
			assert.Equal(t, "", flag.DefValue,
				"%s project flag should default to empty", cmd.Name())
		}
	}
}
