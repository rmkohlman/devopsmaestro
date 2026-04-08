// Package cmd_test contains Phase 2 RED tests for the delete ca-cert command.
//
// RED PHASE: TestDeleteCACertCmd_HasForceFlag is a structural test that verifies
// the --force flag is registered on deleteCACertCmd.
//
// The test currently PASSES (the flag exists). It serves as a regression guard:
// if someone removes the --force flag registration from init(), this test fails.
//
// The deeper security concern driving issue #138 is that WITHOUT --force,
// deletion requires interactive confirmation (reading from os.Stdin). The
// --force flag is the machine-friendly bypass. Tests that invoke runDeleteCACert
// directly must use --force or mock stdin to avoid hanging.
package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeleteCACertCmd_HasForceFlag verifies that the delete ca-cert command
// has a --force flag registered, enabling non-interactive deletion in scripts
// and automated workflows.
func TestDeleteCACertCmd_HasForceFlag(t *testing.T) {
	require.NotNil(t, deleteCACertCmd,
		"deleteCACertCmd must be registered")

	flag := deleteCACertCmd.Flags().Lookup("force")
	assert.NotNil(t, flag,
		"deleteCACertCmd must have a --force flag to bypass confirmation prompt")

	if flag != nil {
		assert.Equal(t, "bool", flag.Value.Type(),
			"--force flag must be a bool")
		assert.Equal(t, "false", flag.DefValue,
			"--force flag must default to false (prompt by default)")
	}
}

// TestDeleteCACertCmd_HasRequiredFlags verifies all expected flags are registered
// on the delete ca-cert command.
func TestDeleteCACertCmd_HasRequiredFlags(t *testing.T) {
	require.NotNil(t, deleteCACertCmd,
		"deleteCACertCmd must be registered")

	expectedFlags := []struct {
		name     string
		flagType string
	}{
		{"ecosystem", "string"},
		{"domain", "string"},
		{"app", "string"},
		{"workspace", "string"},
		{"global", "bool"},
		{"force", "bool"},
	}

	for _, ef := range expectedFlags {
		t.Run("flag --"+ef.name, func(t *testing.T) {
			flag := deleteCACertCmd.Flags().Lookup(ef.name)
			require.NotNil(t, flag,
				"deleteCACertCmd must have --%s flag", ef.name)
			assert.Equal(t, ef.flagType, flag.Value.Type(),
				"--%s must be of type %s", ef.name, ef.flagType)
		})
	}
}
