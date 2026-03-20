package models

// =============================================================================
// Issue #132 — TDD Phase 2: Workspace fields lost during round-trip
//
// These tests verify that ToolsConfig, ShellConfig, and SSHAgentForwarding
// survive a full FromYAML → model → ToYAML round-trip.
//
// RED: All tests below will FAIL until the fixes are implemented:
//   - ToolsConfig is not stored in BuildConfig JSON
//   - ShellConfig is not stored anywhere
//   - SSHAgentForwarding is not in WorkspaceSpec / ContainerConfig
// =============================================================================

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// =============================================================================
// ToolsConfig round-trip (#132 — fix 1 of 3)
// =============================================================================

// TestRoundTrip132_ToolsConfig_FromYAML_Persists verifies that when YAML with
// tools.opencode: true is parsed via FromYAML, the ToolsConfig is persisted
// into the BuildConfig JSON blob.
func TestRoundTrip132_ToolsConfig_FromYAML_Persists(t *testing.T) {
	wsYAML := WorkspaceYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Workspace",
		Metadata:   WorkspaceMetadata{Name: "tools-ws", App: "my-app"},
		Spec: WorkspaceSpec{
			Image: ImageConfig{Name: "ubuntu:22.04"},
			Tools: ToolsConfig{Opencode: true},
		},
	}

	ws := &Workspace{AppID: 1}
	ws.FromYAML(wsYAML)

	// ToolsConfig must be stored in the BuildConfig JSON blob
	require.True(t, ws.BuildConfig.Valid,
		"BuildConfig should be valid when ToolsConfig is set")

	var stored DevBuildConfig
	err := json.Unmarshal([]byte(ws.BuildConfig.String), &stored)
	require.NoError(t, err, "BuildConfig should be valid JSON")
	assert.True(t, stored.Tools.Opencode,
		"ToolsConfig.Opencode should be true in stored BuildConfig JSON")
}

// TestRoundTrip132_ToolsConfig_ToYAML_Restores verifies that ToYAML reads
// ToolsConfig from the BuildConfig JSON and populates spec.Tools.
func TestRoundTrip132_ToolsConfig_ToYAML_Restores(t *testing.T) {
	// Simulate a workspace with ToolsConfig stored in BuildConfig JSON
	buildCfg := DevBuildConfig{
		Tools: ToolsConfig{Opencode: true},
	}
	buildJSON, _ := json.Marshal(buildCfg)

	ws := &Workspace{
		Name:      "tools-restore-ws",
		ImageName: "ubuntu:22.04",
	}
	ws.BuildConfig.Valid = true
	ws.BuildConfig.String = string(buildJSON)

	result := ws.ToYAML("my-app", "")

	assert.True(t, result.Spec.Tools.Opencode,
		"ToYAML should restore ToolsConfig.Opencode from BuildConfig JSON")
}

// TestRoundTrip132_ToolsConfig_FullRoundTrip verifies the complete cycle:
// YAML → FromYAML → model → ToYAML → YAML with ToolsConfig preserved.
func TestRoundTrip132_ToolsConfig_FullRoundTrip(t *testing.T) {
	originalYAML := `
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: tools-roundtrip
  app: myapp
spec:
  image:
    name: ubuntu:22.04
  tools:
    opencode: true
`
	var wsYAML WorkspaceYAML
	require.NoError(t, yaml.Unmarshal([]byte(originalYAML), &wsYAML))

	// FromYAML
	ws := &Workspace{AppID: 1}
	ws.FromYAML(wsYAML)

	// ToYAML
	result := ws.ToYAML("myapp", "")

	// The tools config must survive
	assert.True(t, result.Spec.Tools.Opencode,
		"ToolsConfig.Opencode must survive FromYAML → ToYAML round-trip")

	// Serialize to YAML bytes and verify text
	data, err := yaml.Marshal(result)
	require.NoError(t, err)
	yamlStr := string(data)
	assert.Contains(t, yamlStr, "opencode: true",
		"serialized YAML must contain 'opencode: true'")
}

// =============================================================================
// ShellConfig round-trip (#132 — fix 2 of 3)
// =============================================================================

// TestRoundTrip132_ShellConfig_FromYAML_Persists verifies that when YAML with
// shell config is parsed via FromYAML, the ShellConfig is persisted into the
// BuildConfig JSON blob.
func TestRoundTrip132_ShellConfig_FromYAML_Persists(t *testing.T) {
	wsYAML := WorkspaceYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Workspace",
		Metadata:   WorkspaceMetadata{Name: "shell-ws", App: "my-app"},
		Spec: WorkspaceSpec{
			Image: ImageConfig{Name: "ubuntu:22.04"},
			Shell: ShellConfig{
				Type:      "zsh",
				Framework: "oh-my-zsh",
				Theme:     "starship",
			},
		},
	}

	ws := &Workspace{AppID: 1}
	ws.FromYAML(wsYAML)

	require.True(t, ws.BuildConfig.Valid,
		"BuildConfig should be valid when ShellConfig is set")

	var stored DevBuildConfig
	err := json.Unmarshal([]byte(ws.BuildConfig.String), &stored)
	require.NoError(t, err, "BuildConfig should be valid JSON")
	assert.Equal(t, "zsh", stored.Shell.Type,
		"ShellConfig.Type should be 'zsh' in stored BuildConfig JSON")
	assert.Equal(t, "oh-my-zsh", stored.Shell.Framework,
		"ShellConfig.Framework should be 'oh-my-zsh' in stored BuildConfig JSON")
	assert.Equal(t, "starship", stored.Shell.Theme,
		"ShellConfig.Theme should be 'starship' in stored BuildConfig JSON")
}

// TestRoundTrip132_ShellConfig_ToYAML_Restores verifies that ToYAML reads
// ShellConfig from the BuildConfig JSON and populates spec.Shell.
func TestRoundTrip132_ShellConfig_ToYAML_Restores(t *testing.T) {
	buildCfg := DevBuildConfig{
		Shell: ShellConfig{
			Type:      "bash",
			Framework: "",
			Theme:     "powerlevel10k",
		},
	}
	buildJSON, _ := json.Marshal(buildCfg)

	ws := &Workspace{
		Name:      "shell-restore-ws",
		ImageName: "ubuntu:22.04",
	}
	ws.BuildConfig.Valid = true
	ws.BuildConfig.String = string(buildJSON)

	result := ws.ToYAML("my-app", "")

	assert.Equal(t, "bash", result.Spec.Shell.Type,
		"ToYAML should restore ShellConfig.Type from BuildConfig JSON")
	assert.Equal(t, "powerlevel10k", result.Spec.Shell.Theme,
		"ToYAML should restore ShellConfig.Theme from BuildConfig JSON")
}

// TestRoundTrip132_ShellConfig_FullRoundTrip verifies the complete cycle:
// YAML → FromYAML → model → ToYAML → YAML with ShellConfig preserved.
func TestRoundTrip132_ShellConfig_FullRoundTrip(t *testing.T) {
	originalYAML := `
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: shell-roundtrip
  app: myapp
spec:
  image:
    name: ubuntu:22.04
  shell:
    type: zsh
    framework: oh-my-zsh
    theme: starship
`
	var wsYAML WorkspaceYAML
	require.NoError(t, yaml.Unmarshal([]byte(originalYAML), &wsYAML))

	ws := &Workspace{AppID: 1}
	ws.FromYAML(wsYAML)

	result := ws.ToYAML("myapp", "")

	assert.Equal(t, "zsh", result.Spec.Shell.Type,
		"ShellConfig.Type must survive round-trip")
	assert.Equal(t, "oh-my-zsh", result.Spec.Shell.Framework,
		"ShellConfig.Framework must survive round-trip")
	assert.Equal(t, "starship", result.Spec.Shell.Theme,
		"ShellConfig.Theme must survive round-trip")

	// Serialize to YAML bytes and check text
	data, err := yaml.Marshal(result)
	require.NoError(t, err)
	yamlStr := string(data)
	assert.Contains(t, yamlStr, "type: zsh")
	assert.Contains(t, yamlStr, "framework: oh-my-zsh")
	assert.Contains(t, yamlStr, "theme: starship")
}

// =============================================================================
// SSHAgentForwarding round-trip (#132 — fix 3 of 3)
// =============================================================================

// TestRoundTrip132_SSHAgentForwarding_FromYAML verifies that when YAML with
// sshAgentForwarding: true in the container config is parsed, FromYAML
// persists it to the model.
func TestRoundTrip132_SSHAgentForwarding_FromYAML(t *testing.T) {
	wsYAML := WorkspaceYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Workspace",
		Metadata:   WorkspaceMetadata{Name: "ssh-ws", App: "my-app"},
		Spec: WorkspaceSpec{
			Image: ImageConfig{Name: "ubuntu:22.04"},
			Container: ContainerConfig{
				SSHAgentForwarding: true,
			},
		},
	}

	ws := &Workspace{AppID: 1}
	ws.FromYAML(wsYAML)

	assert.True(t, ws.SSHAgentForwarding,
		"FromYAML should set SSHAgentForwarding from YAML container config")
}

// TestRoundTrip132_SSHAgentForwarding_ToYAML verifies that ToYAML includes
// sshAgentForwarding in the container config when set on the model.
func TestRoundTrip132_SSHAgentForwarding_ToYAML(t *testing.T) {
	ws := &Workspace{
		Name:               "ssh-export-ws",
		ImageName:          "ubuntu:22.04",
		SSHAgentForwarding: true,
	}

	result := ws.ToYAML("my-app", "")

	assert.True(t, result.Spec.Container.SSHAgentForwarding,
		"ToYAML should include SSHAgentForwarding in container config")

	data, err := yaml.Marshal(result)
	require.NoError(t, err)
	yamlStr := string(data)
	assert.Contains(t, yamlStr, "sshAgentForwarding: true",
		"serialized YAML must contain 'sshAgentForwarding: true'")
}

// TestRoundTrip132_SSHAgentForwarding_FullRoundTrip verifies the complete cycle:
// YAML → FromYAML → model → ToYAML → YAML with SSHAgentForwarding preserved.
func TestRoundTrip132_SSHAgentForwarding_FullRoundTrip(t *testing.T) {
	originalYAML := `
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: ssh-roundtrip
  app: myapp
spec:
  image:
    name: ubuntu:22.04
  container:
    sshAgentForwarding: true
`
	var wsYAML WorkspaceYAML
	require.NoError(t, yaml.Unmarshal([]byte(originalYAML), &wsYAML))

	ws := &Workspace{AppID: 1}
	ws.FromYAML(wsYAML)

	assert.True(t, ws.SSHAgentForwarding,
		"FromYAML must capture SSHAgentForwarding")

	result := ws.ToYAML("myapp", "")

	assert.True(t, result.Spec.Container.SSHAgentForwarding,
		"SSHAgentForwarding must survive round-trip")

	data, err := yaml.Marshal(result)
	require.NoError(t, err)
	yamlStr := string(data)
	assert.Contains(t, yamlStr, "sshAgentForwarding: true",
		"serialized YAML must contain 'sshAgentForwarding: true'")
}

// TestRoundTrip132_SSHAgentForwarding_FalseOmitted verifies that when
// SSHAgentForwarding is false (default), it is omitted from YAML output.
func TestRoundTrip132_SSHAgentForwarding_FalseOmitted(t *testing.T) {
	ws := &Workspace{
		Name:               "ssh-default-ws",
		ImageName:          "ubuntu:22.04",
		SSHAgentForwarding: false,
	}

	result := ws.ToYAML("my-app", "")
	data, err := yaml.Marshal(result)
	require.NoError(t, err)
	yamlStr := string(data)

	assert.NotContains(t, yamlStr, "sshAgentForwarding",
		"sshAgentForwarding should be omitted when false (default)")
}

// =============================================================================
// Combined round-trip: all three fields together
// =============================================================================

// TestRoundTrip132_AllThreeFields verifies that when a workspace has all three
// fields (ToolsConfig, ShellConfig, SSHAgentForwarding) set, they all survive
// a complete FromYAML → model → ToYAML round-trip simultaneously.
func TestRoundTrip132_AllThreeFields(t *testing.T) {
	originalYAML := `
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: combined-roundtrip
  app: myapp
spec:
  image:
    name: ubuntu:22.04
  tools:
    opencode: true
  shell:
    type: zsh
    framework: oh-my-zsh
    theme: starship
  container:
    sshAgentForwarding: true
    user: dev
    uid: 1000
    gid: 1000
    workingDir: /workspace
`
	var wsYAML WorkspaceYAML
	require.NoError(t, yaml.Unmarshal([]byte(originalYAML), &wsYAML))

	ws := &Workspace{AppID: 1}
	ws.FromYAML(wsYAML)

	result := ws.ToYAML("myapp", "")

	// ToolsConfig
	assert.True(t, result.Spec.Tools.Opencode, "ToolsConfig.Opencode must survive")

	// ShellConfig
	assert.Equal(t, "zsh", result.Spec.Shell.Type, "Shell.Type must survive")
	assert.Equal(t, "oh-my-zsh", result.Spec.Shell.Framework, "Shell.Framework must survive")
	assert.Equal(t, "starship", result.Spec.Shell.Theme, "Shell.Theme must survive")

	// SSHAgentForwarding
	assert.True(t, result.Spec.Container.SSHAgentForwarding, "SSHAgentForwarding must survive")
}
