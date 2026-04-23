package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttachCmd_HasEmergencyFlag(t *testing.T) {
	require.NotNil(t, attachCmd.Flags().Lookup("emergency"),
		"attachCmd must have an --emergency flag")
}

func TestAttachCmd_EmergencyFlagNoShort(t *testing.T) {
	f := attachCmd.Flags().Lookup("emergency")
	require.NotNil(t, f)
	assert.Equal(t, "", f.Shorthand,
		"--emergency must not have a short flag (-e is reserved for --ecosystem)")
}

func TestAttachCmd_EcosystemShortFlagPreserved(t *testing.T) {
	f := attachCmd.Flags().ShorthandLookup("e")
	require.NotNil(t, f, "-e shorthand must exist on attachCmd")
	assert.Equal(t, "ecosystem", f.Name,
		"-e must bind to --ecosystem, not --emergency")
}
