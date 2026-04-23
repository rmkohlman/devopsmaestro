package cmd

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"devopsmaestro/builders/emergency"
)

func TestGenerateEmergencyName_Prefix(t *testing.T) {
	name := generateEmergencyName()
	assert.True(t, strings.HasPrefix(name, emergency.ContainerNamePrefix),
		"expected prefix %q, got %q", emergency.ContainerNamePrefix, name)
}

func TestGenerateEmergencyName_Unique(t *testing.T) {
	seen := make(map[string]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		name := generateEmergencyName()
		_, dup := seen[name]
		require.False(t, dup, "collision detected at iteration %d: %s", i, name)
		seen[name] = struct{}{}
	}
}

func TestGenerateEmergencyName_Length(t *testing.T) {
	name := generateEmergencyName()
	// prefix (14) + 8 hex chars = 22
	expected := len(emergency.ContainerNamePrefix) + 8
	assert.Equal(t, expected, len(name),
		"expected length %d, got %d (%q)", expected, len(name), name)
}

func TestResolveEmergencyMountPath_FallsBackToCwd(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)

	// Build a minimal cobra command with no hierarchy flags set.
	cmd := newTestAttachCmd()
	cmd.SetContext(context.Background())

	path, err := resolveEmergencyMountPath(cmd)
	require.NoError(t, err)
	assert.Equal(t, tmp, path)
}
