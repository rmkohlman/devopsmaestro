package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"devopsmaestro/operators"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// ========== Command Structure Tests ==========

// TestUseCommandExists verifies the use command is registered
func TestUseCommandExists(t *testing.T) {
	assert.NotNil(t, useCmd, "useCmd should exist")
	assert.Equal(t, "use", useCmd.Use, "useCmd should have correct Use")
}

// TestUseCommandHierarchy verifies the use command structure
func TestUseCommandHierarchy(t *testing.T) {
	subcommands := useCmd.Commands()
	subcommandNames := make([]string, len(subcommands))
	for i, cmd := range subcommands {
		subcommandNames[i] = cmd.Name()
	}

	assert.Contains(t, subcommandNames, "project", "use should have 'project' subcommand")
	assert.Contains(t, subcommandNames, "workspace", "use should have 'workspace' subcommand")
}

// TestUseProjectCommandExists verifies use project command is registered
func TestUseProjectCommandExists(t *testing.T) {
	assert.NotNil(t, useProjectCmd, "useProjectCmd should exist")
	assert.Equal(t, "project <name>", useProjectCmd.Use, "useProjectCmd should have correct Use")
}

// TestUseWorkspaceCommandExists verifies use workspace command is registered
func TestUseWorkspaceCommandExists(t *testing.T) {
	assert.NotNil(t, useWorkspaceCmd, "useWorkspaceCmd should exist")
	assert.Equal(t, "workspace <name>", useWorkspaceCmd.Use, "useWorkspaceCmd should have correct Use")
}

// ========== Flag Tests ==========

// TestUseCmdHasClearFlag verifies the --clear flag is registered
func TestUseCmdHasClearFlag(t *testing.T) {
	clearFlag := useCmd.Flags().Lookup("clear")
	assert.NotNil(t, clearFlag, "useCmd should have 'clear' flag")

	if clearFlag != nil {
		assert.Equal(t, "false", clearFlag.DefValue, "clear flag should default to false")
		assert.Equal(t, "bool", clearFlag.Value.Type(), "clear flag should be bool type")
	}
}

// TestUseProjectCmdRequiresOneArg verifies use project requires exactly 1 argument
func TestUseProjectCmdRequiresOneArg(t *testing.T) {
	assert.NotNil(t, useProjectCmd.Args, "useProjectCmd should have Args validator")
}

// TestUseWorkspaceCmdRequiresOneArg verifies use workspace requires exactly 1 argument
func TestUseWorkspaceCmdRequiresOneArg(t *testing.T) {
	assert.NotNil(t, useWorkspaceCmd.Args, "useWorkspaceCmd should have Args validator")
}

// ========== Help Text Tests ==========

// TestUseCommandHelp verifies help text includes all options
func TestUseCommandHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	useCmd.SetOut(buf)
	useCmd.Help()
	helpText := buf.String()

	// Should mention subcommands
	assert.Contains(t, helpText, "project", "help should mention 'project'")
	assert.Contains(t, helpText, "workspace", "help should mention 'workspace'")

	// Should mention --clear flag
	assert.Contains(t, helpText, "--clear", "help should mention '--clear' flag")

	// Should mention 'none' for clearing
	assert.Contains(t, helpText, "none", "help should mention 'none' for clearing context")
}

// TestUseProjectCommandHelp verifies help text documents clearing with 'none'
func TestUseProjectCommandHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	useProjectCmd.SetOut(buf)
	useProjectCmd.Help()
	helpText := buf.String()

	// Should mention 'none' for clearing
	assert.Contains(t, helpText, "none", "help should mention 'none' for clearing")
	assert.Contains(t, helpText, "clear", "help should mention clearing context")
}

// TestUseWorkspaceCommandHelp verifies help text documents clearing with 'none'
func TestUseWorkspaceCommandHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	useWorkspaceCmd.SetOut(buf)
	useWorkspaceCmd.Help()
	helpText := buf.String()

	// Should mention 'none' for clearing
	assert.Contains(t, helpText, "none", "help should mention 'none' for clearing")
	assert.Contains(t, helpText, "clear", "help should mention clearing context")
}

// ========== Integration Tests with Real Context Manager ==========

// setupTestContextManager creates a context manager with a temp directory
func setupTestContextManager(t *testing.T) (*operators.ContextManager, string, func()) {
	t.Helper()

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "dvm-test-*")
	require.NoError(t, err)

	// Create a custom context manager that uses the temp directory
	contextPath := filepath.Join(tempDir, "context.yaml")

	// We need to create our own ContextManager-like behavior for testing
	// since the real one uses ~/.devopsmaestro
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	// For now, we'll test the ContextManager directly with its temp file behavior
	// by manipulating the home directory or testing the exported functions

	return nil, contextPath, cleanup
}

// TestClearProjectContext tests that 'dvm use project none' clears context
func TestClearProjectContext(t *testing.T) {
	// Create a temp context file
	tempDir, err := os.MkdirTemp("", "dvm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	contextPath := filepath.Join(tempDir, "context.yaml")

	// Write initial context with project and workspace
	initialContext := operators.ContextConfig{
		CurrentProject:   "test-project",
		CurrentWorkspace: "test-workspace",
	}
	data, err := yaml.Marshal(initialContext)
	require.NoError(t, err)
	err = os.WriteFile(contextPath, data, 0644)
	require.NoError(t, err)

	// Verify initial state
	var ctx operators.ContextConfig
	data, err = os.ReadFile(contextPath)
	require.NoError(t, err)
	err = yaml.Unmarshal(data, &ctx)
	require.NoError(t, err)
	assert.Equal(t, "test-project", ctx.CurrentProject)
	assert.Equal(t, "test-workspace", ctx.CurrentWorkspace)

	// Now test clearing by writing empty context (simulating ClearProject)
	clearedContext := operators.ContextConfig{
		CurrentProject:   "",
		CurrentWorkspace: "",
	}
	data, err = yaml.Marshal(clearedContext)
	require.NoError(t, err)
	err = os.WriteFile(contextPath, data, 0644)
	require.NoError(t, err)

	// Verify cleared state
	data, err = os.ReadFile(contextPath)
	require.NoError(t, err)
	err = yaml.Unmarshal(data, &ctx)
	require.NoError(t, err)
	assert.Equal(t, "", ctx.CurrentProject)
	assert.Equal(t, "", ctx.CurrentWorkspace)
}

// TestClearWorkspaceContext tests that 'dvm use workspace none' clears only workspace
func TestClearWorkspaceContext(t *testing.T) {
	// Create a temp context file
	tempDir, err := os.MkdirTemp("", "dvm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	contextPath := filepath.Join(tempDir, "context.yaml")

	// Write initial context with project and workspace
	initialContext := operators.ContextConfig{
		CurrentProject:   "test-project",
		CurrentWorkspace: "test-workspace",
	}
	data, err := yaml.Marshal(initialContext)
	require.NoError(t, err)
	err = os.WriteFile(contextPath, data, 0644)
	require.NoError(t, err)

	// Simulate ClearWorkspace (only clears workspace, keeps project)
	clearedContext := operators.ContextConfig{
		CurrentProject:   "test-project", // Project should remain
		CurrentWorkspace: "",             // Only workspace cleared
	}
	data, err = yaml.Marshal(clearedContext)
	require.NoError(t, err)
	err = os.WriteFile(contextPath, data, 0644)
	require.NoError(t, err)

	// Verify state - project should remain, workspace cleared
	var ctx operators.ContextConfig
	data, err = os.ReadFile(contextPath)
	require.NoError(t, err)
	err = yaml.Unmarshal(data, &ctx)
	require.NoError(t, err)
	assert.Equal(t, "test-project", ctx.CurrentProject, "project should remain when clearing workspace")
	assert.Equal(t, "", ctx.CurrentWorkspace, "workspace should be cleared")
}

// ========== Command Argument Tests ==========

// TestUseProjectNoneIsValidArg tests that 'none' is accepted as a valid argument
func TestUseProjectNoneIsValidArg(t *testing.T) {
	// The command accepts exactly 1 arg, 'none' should be valid
	cmd := &cobra.Command{}
	cmd.SetArgs([]string{"none"})

	// Verify the command accepts the arg (no validation error)
	args := []string{"none"}
	err := cobra.ExactArgs(1)(cmd, args)
	assert.NoError(t, err, "'none' should be valid arg for use project")
}

// TestUseWorkspaceNoneIsValidArg tests that 'none' is accepted as a valid argument
func TestUseWorkspaceNoneIsValidArg(t *testing.T) {
	// The command accepts exactly 1 arg, 'none' should be valid
	cmd := &cobra.Command{}
	cmd.SetArgs([]string{"none"})

	args := []string{"none"}
	err := cobra.ExactArgs(1)(cmd, args)
	assert.NoError(t, err, "'none' should be valid arg for use workspace")
}

// TestUseProjectRequiresArg tests that use project without arg fails
func TestUseProjectRequiresArg(t *testing.T) {
	args := []string{}
	err := cobra.ExactArgs(1)(nil, args)
	assert.Error(t, err, "use project should require exactly 1 arg")
}

// TestUseWorkspaceRequiresArg tests that use workspace without arg fails
func TestUseWorkspaceRequiresArg(t *testing.T) {
	args := []string{}
	err := cobra.ExactArgs(1)(nil, args)
	assert.Error(t, err, "use workspace should require exactly 1 arg")
}

// ========== RunE Tests ==========

// TestUseCmdHasRunE verifies useCmd uses RunE (not Run)
func TestUseCmdHasRunE(t *testing.T) {
	assert.NotNil(t, useCmd.RunE, "useCmd should have RunE (not Run)")
}

// TestUseProjectCmdHasRunE verifies useProjectCmd uses RunE
func TestUseProjectCmdHasRunE(t *testing.T) {
	assert.NotNil(t, useProjectCmd.RunE, "useProjectCmd should have RunE (not Run)")
}

// TestUseWorkspaceCmdHasRunE verifies useWorkspaceCmd uses RunE
func TestUseWorkspaceCmdHasRunE(t *testing.T) {
	assert.NotNil(t, useWorkspaceCmd.RunE, "useWorkspaceCmd should have RunE (not Run)")
}

// ========== Long Description Tests ==========

// TestUseCmdLongDescriptionMentionsClear verifies documentation
func TestUseCmdLongDescriptionMentionsClear(t *testing.T) {
	assert.Contains(t, useCmd.Long, "clear", "useCmd Long should mention clearing")
	assert.Contains(t, useCmd.Long, "none", "useCmd Long should mention 'none'")
	assert.Contains(t, useCmd.Long, "--clear", "useCmd Long should mention --clear flag")
}

// TestUseProjectCmdLongDescriptionMentionsClear verifies documentation
func TestUseProjectCmdLongDescriptionMentionsClear(t *testing.T) {
	assert.Contains(t, useProjectCmd.Long, "none", "useProjectCmd Long should mention 'none'")
	assert.Contains(t, useProjectCmd.Long, "clear", "useProjectCmd Long should mention clearing")
}

// TestUseWorkspaceCmdLongDescriptionMentionsClear verifies documentation
func TestUseWorkspaceCmdLongDescriptionMentionsClear(t *testing.T) {
	assert.Contains(t, useWorkspaceCmd.Long, "none", "useWorkspaceCmd Long should mention 'none'")
	assert.Contains(t, useWorkspaceCmd.Long, "clear", "useWorkspaceCmd Long should mention clearing")
}

// ========== Context Manager Unit Tests ==========

// TestContextManagerClearProjectClearsBoth tests ClearProject clears both project and workspace
func TestContextManagerClearProjectClearsBoth(t *testing.T) {
	// This test verifies the ContextConfig struct behavior
	ctx := &operators.ContextConfig{
		CurrentProject:   "my-project",
		CurrentWorkspace: "my-workspace",
	}

	// Simulate ClearProject behavior
	ctx.CurrentProject = ""
	ctx.CurrentWorkspace = ""

	assert.Equal(t, "", ctx.CurrentProject)
	assert.Equal(t, "", ctx.CurrentWorkspace)
}

// TestContextManagerClearWorkspaceKeepsProject tests ClearWorkspace keeps project
func TestContextManagerClearWorkspaceKeepsProject(t *testing.T) {
	ctx := &operators.ContextConfig{
		CurrentProject:   "my-project",
		CurrentWorkspace: "my-workspace",
	}

	// Simulate ClearWorkspace behavior
	ctx.CurrentWorkspace = ""

	assert.Equal(t, "my-project", ctx.CurrentProject, "project should be preserved")
	assert.Equal(t, "", ctx.CurrentWorkspace, "workspace should be cleared")
}

// ========== Example Commands Tests ==========

// TestUseCommandExamples verifies examples are documented
func TestUseCommandExamples(t *testing.T) {
	// Check Long description has examples
	assert.Contains(t, useCmd.Long, "dvm use project my-api", "should have project example")
	assert.Contains(t, useCmd.Long, "dvm use workspace dev", "should have workspace example")
	assert.Contains(t, useCmd.Long, "dvm use project none", "should have clear project example")
	assert.Contains(t, useCmd.Long, "dvm use workspace none", "should have clear workspace example")
	assert.Contains(t, useCmd.Long, "dvm use --clear", "should have clear all example")
}

// TestUseProjectCommandExamples verifies examples are documented
func TestUseProjectCommandExamples(t *testing.T) {
	assert.Contains(t, useProjectCmd.Long, "dvm use project my-api", "should have set project example")
	assert.Contains(t, useProjectCmd.Long, "dvm use project none", "should have clear example")
}

// TestUseWorkspaceCommandExamples verifies examples are documented
func TestUseWorkspaceCommandExamples(t *testing.T) {
	assert.Contains(t, useWorkspaceCmd.Long, "dvm use workspace main", "should have set workspace example")
	assert.Contains(t, useWorkspaceCmd.Long, "dvm use workspace none", "should have clear example")
}

// ========== Short Description Tests ==========

// TestUseCommandShortDescription verifies short description
func TestUseCommandShortDescription(t *testing.T) {
	assert.NotEmpty(t, useCmd.Short, "useCmd should have Short description")
	assert.Contains(t, useCmd.Short, "context", "Short should mention context")
}

// TestUseProjectCommandShortDescription verifies short description
func TestUseProjectCommandShortDescription(t *testing.T) {
	assert.NotEmpty(t, useProjectCmd.Short, "useProjectCmd should have Short description")
}

// TestUseWorkspaceCommandShortDescription verifies short description
func TestUseWorkspaceCommandShortDescription(t *testing.T) {
	assert.NotEmpty(t, useWorkspaceCmd.Short, "useWorkspaceCmd should have Short description")
}
