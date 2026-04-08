package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddOutputFlag(t *testing.T) {
	tests := []struct {
		name       string
		defaultVal string
	}{
		{"with table default", "table"},
		{"with empty default", ""},
		{"with yaml default", "yaml"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			AddOutputFlag(cmd, tt.defaultVal)

			flag := cmd.Flags().Lookup("output")
			require.NotNil(t, flag)
			assert.Equal(t, "o", flag.Shorthand)
			assert.Equal(t, tt.defaultVal, flag.DefValue)
			assert.Equal(t, "string", flag.Value.Type())
		})
	}
}

func TestAddForceConfirmFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	AddForceConfirmFlag(cmd)

	flag := cmd.Flags().Lookup("force")
	require.NotNil(t, flag)
	assert.Equal(t, "", flag.Shorthand)
	assert.Equal(t, "false", flag.DefValue)
	assert.Equal(t, "bool", flag.Value.Type())
	assert.Equal(t, "Skip confirmation prompt", flag.Usage)
}

func TestAddDryRunFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	var dest bool
	AddDryRunFlag(cmd, &dest)

	flag := cmd.Flags().Lookup("dry-run")
	require.NotNil(t, flag)
	assert.Equal(t, "false", flag.DefValue)
	assert.Equal(t, "bool", flag.Value.Type())
	assert.Equal(t, "Preview changes without applying", flag.Usage)

	// Verify BoolVar binding
	require.NoError(t, cmd.Flags().Set("dry-run", "true"))
	assert.True(t, dest)
}

func TestAddAllFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	AddAllFlag(cmd, "Show everything")

	flag := cmd.Flags().Lookup("all")
	require.NotNil(t, flag)
	assert.Equal(t, "A", flag.Shorthand)
	assert.Equal(t, "false", flag.DefValue)
	assert.Equal(t, "bool", flag.Value.Type())
	assert.Equal(t, "Show everything", flag.Usage)
}
