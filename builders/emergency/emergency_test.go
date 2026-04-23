package emergency_test

import (
	"testing"

	"devopsmaestro/builders/emergency"

	"github.com/stretchr/testify/assert"
)

func TestImageName_Stable(t *testing.T) {
	assert.Equal(t, "dvm-emergency:v1", emergency.ImageName,
		"ImageName must not change without bumping the version suffix (forces rebuild)")
}

func TestContainerNamePrefix_Stable(t *testing.T) {
	assert.Equal(t, "dvm-emergency-", emergency.ContainerNamePrefix)
}

func TestLabelKey_Stable(t *testing.T) {
	assert.Equal(t, "dvm.emergency", emergency.LabelKey)
}

func TestDockerfile_NotEmpty(t *testing.T) {
	assert.NotEmpty(t, emergency.Dockerfile())
}

func TestDockerfile_AlpineBase(t *testing.T) {
	assert.Contains(t, emergency.Dockerfile(), "FROM alpine")
}

func TestDockerfile_RequiredTools(t *testing.T) {
	df := emergency.Dockerfile()
	for _, tool := range []string{"bash", "git", "vim", "nano", "curl", "ca-certificates"} {
		assert.Contains(t, df, tool, "Dockerfile must install %s", tool)
	}
}

func TestDockerfile_NonRootUser(t *testing.T) {
	df := emergency.Dockerfile()
	assert.Contains(t, df, "addgroup -g 1000 dev")
	assert.Contains(t, df, "adduser -D -u 1000 -G dev")
}

func TestDockerfile_WorkspaceMountPoint(t *testing.T) {
	assert.Contains(t, emergency.Dockerfile(), "WORKDIR /workspace")
}
