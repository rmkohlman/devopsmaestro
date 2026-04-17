package cmd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Issue #94: --timeout flag existence tests for build, attach, detach commands
// =============================================================================

func TestBuildCmd_HasTimeoutFlag(t *testing.T) {
	flag := buildCmd.Flags().Lookup("timeout")
	require.NotNil(t, flag, "build command should have --timeout flag")

	assert.Equal(t, "duration", flag.Value.Type(), "timeout flag should be duration type")
	assert.Equal(t, (10 * time.Minute).String(), flag.DefValue, "build timeout should default to 10m")
}

func TestAttachCmd_HasTimeoutFlag(t *testing.T) {
	flag := attachCmd.Flags().Lookup("timeout")
	require.NotNil(t, flag, "attach command should have --timeout flag")

	assert.Equal(t, "duration", flag.Value.Type(), "timeout flag should be duration type")
	assert.Equal(t, (10 * time.Minute).String(), flag.DefValue, "attach timeout should default to 10m")
}

func TestDetachCmd_HasTimeoutFlag(t *testing.T) {
	flag := detachCmd.Flags().Lookup("timeout")
	require.NotNil(t, flag, "detach command should have --timeout flag")

	assert.Equal(t, "duration", flag.Value.Type(), "timeout flag should be duration type")
	assert.Equal(t, (5 * time.Minute).String(), flag.DefValue, "detach timeout should default to 5m")
}

func TestTimeoutFlags_AcceptDurationStrings(t *testing.T) {
	tests := []struct {
		command string
		input   string
		want    time.Duration
	}{
		{"build", "5m", 5 * time.Minute},
		{"build", "1h", 1 * time.Hour},
		{"build", "30s", 30 * time.Second},
		{"attach", "2m30s", 2*time.Minute + 30*time.Second},
		{"detach", "90s", 90 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.command+"_"+tt.input, func(t *testing.T) {
			var flag = buildCmd.Flags().Lookup("timeout")
			switch tt.command {
			case "attach":
				flag = attachCmd.Flags().Lookup("timeout")
			case "detach":
				flag = detachCmd.Flags().Lookup("timeout")
			}
			require.NotNil(t, flag)

			err := flag.Value.Set(tt.input)
			assert.NoError(t, err, "should accept duration string %q", tt.input)

			parsed, err := time.ParseDuration(flag.Value.String())
			assert.NoError(t, err)
			assert.Equal(t, tt.want, parsed)
		})
	}
}
