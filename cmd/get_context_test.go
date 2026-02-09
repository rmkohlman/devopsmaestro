package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetContextCommand(t *testing.T) {
	// Test that the context command exists
	assert.NotNil(t, getContextCmd)
	assert.Equal(t, "context", getContextCmd.Use)
	assert.Contains(t, getContextCmd.Short, "context")
}

func TestGetContextCommandHelp(t *testing.T) {
	// Verify help text contains useful information
	helpText := getContextCmd.Long

	assert.Contains(t, helpText, "app")
	assert.Contains(t, helpText, "workspace")
	assert.Contains(t, helpText, "dvm use app")
	assert.Contains(t, helpText, "dvm use workspace")
	assert.Contains(t, helpText, "DVM_APP")
	assert.Contains(t, helpText, "DVM_WORKSPACE")
}

func TestGetContextCommandExamples(t *testing.T) {
	// Verify examples are present
	helpText := getContextCmd.Long

	assert.Contains(t, helpText, "dvm get context")
	assert.Contains(t, helpText, "-o yaml")
	assert.Contains(t, helpText, "-o json")
}

func TestGetContextIsSubcommandOfGet(t *testing.T) {
	// Verify context is registered as a subcommand of get
	found := false
	for _, cmd := range getCmd.Commands() {
		if cmd.Use == "context" {
			found = true
			break
		}
	}
	assert.True(t, found, "context should be a subcommand of get")
}
