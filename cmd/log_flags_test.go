package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Issue #96: --log-level, --log-format, --verbose flag tests
// =============================================================================

func TestRootCmd_HasLogLevelFlag(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("log-level")
	require.NotNil(t, flag, "root command should have --log-level flag")

	assert.Equal(t, "string", flag.Value.Type(), "log-level should be a string flag")
	assert.Equal(t, "warn", flag.DefValue, "log-level should default to warn")
}

func TestRootCmd_HasLogFormatFlag(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("log-format")
	require.NotNil(t, flag, "root command should have --log-format flag")

	assert.Equal(t, "string", flag.Value.Type(), "log-format should be a string flag")
	assert.Equal(t, "text", flag.DefValue, "log-format should default to text")
}

func TestRootCmd_HasVerboseFlag(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("verbose")
	require.NotNil(t, flag, "root command should have --verbose flag")

	assert.Equal(t, "v", flag.Shorthand, "verbose should have -v shorthand")
	assert.Equal(t, "bool", flag.Value.Type(), "verbose should be a bool flag")
	assert.Equal(t, "false", flag.DefValue, "verbose should default to false")
}

func TestRootCmd_HasLogFileFlag(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("log-file")
	require.NotNil(t, flag, "root command should have --log-file flag")

	assert.Equal(t, "string", flag.Value.Type(), "log-file should be a string flag")
	assert.Equal(t, "", flag.DefValue, "log-file should default to empty")
}

func TestRootCmd_LogFlagsArePersistent(t *testing.T) {
	// Persistent flags are inherited by all subcommands
	flags := []string{"log-level", "log-format", "verbose", "log-file"}
	for _, name := range flags {
		t.Run(name, func(t *testing.T) {
			// PersistentFlags().Lookup finds only persistent flags
			flag := rootCmd.PersistentFlags().Lookup(name)
			require.NotNil(t, flag, "flag %s should be persistent", name)
		})
	}
}
