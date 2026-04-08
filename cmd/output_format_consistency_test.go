package cmd

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// =============================================================================
// Output format consistency tests (Issue #20)
//
// Ensures all commands that produce structured output use AddOutputFlag()
// consistently and that -o json/yaml produce valid output.
// =============================================================================

// ========== Flag Existence Tests ==========

func TestVersionCmd_HasOutputFlag(t *testing.T) {
	flag := versionCmd.Flags().Lookup("output")
	require.NotNil(t, flag, "version command should have --output flag")
	assert.Equal(t, "o", flag.Shorthand, "version --output should have -o shorthand")
	assert.Equal(t, "string", flag.Value.Type())
}

func TestStatusCmd_HasOutputFlag(t *testing.T) {
	flag := statusCmd.Flags().Lookup("output")
	require.NotNil(t, flag, "status command should have --output flag")
	assert.Equal(t, "o", flag.Shorthand, "status --output should have -o shorthand")
	assert.Equal(t, "string", flag.Value.Type())
}

func TestStatusCmd_UsesAddOutputFlag(t *testing.T) {
	// Verify status command uses AddOutputFlag (empty default = colored output)
	flag := statusCmd.Flags().Lookup("output")
	require.NotNil(t, flag)
	assert.Equal(t, "", flag.DefValue,
		"status --output should default to empty (colored output)")
}

func TestGetCmd_HasPersistentOutputFlag(t *testing.T) {
	// get command uses PersistentFlags so children inherit it
	flag := getCmd.PersistentFlags().Lookup("output")
	require.NotNil(t, flag, "get command should have persistent --output flag")
	assert.Equal(t, "o", flag.Shorthand)
}

func TestLibraryListCmd_HasOutputFlag(t *testing.T) {
	flag := libraryListCmd.Flags().Lookup("output")
	require.NotNil(t, flag, "library list should have --output flag")
	assert.Equal(t, "o", flag.Shorthand)
	assert.Equal(t, "table", flag.DefValue)
}

func TestLibraryShowCmd_HasOutputFlag(t *testing.T) {
	flag := libraryShowCmd.Flags().Lookup("output")
	require.NotNil(t, flag, "library show should have --output flag")
	assert.Equal(t, "o", flag.Shorthand)
	assert.Equal(t, "table", flag.DefValue)
}

func TestLibraryImportCmd_HasOutputFlag(t *testing.T) {
	flag := libraryImportCmd.Flags().Lookup("output")
	require.NotNil(t, flag, "library import should have --output flag")
	assert.Equal(t, "o", flag.Shorthand)
	assert.Equal(t, "table", flag.DefValue)
}

func TestRolloutStatusRegistryCmd_HasOutputFlag(t *testing.T) {
	statusSub := findSubcommand(rolloutCmd, "status")
	require.NotNil(t, statusSub)
	regSub := findSubcommand(statusSub, "registry")
	require.NotNil(t, regSub)

	flag := regSub.Flags().Lookup("output")
	require.NotNil(t, flag, "rollout status registry should have --output flag")
	assert.Equal(t, "o", flag.Shorthand)
}

func TestRolloutHistoryRegistryCmd_HasOutputFlag(t *testing.T) {
	historySub := findSubcommand(rolloutCmd, "history")
	require.NotNil(t, historySub)
	regSub := findSubcommand(historySub, "registry")
	require.NotNil(t, regSub)

	flag := regSub.Flags().Lookup("output")
	require.NotNil(t, flag, "rollout history registry should have --output flag")
	assert.Equal(t, "o", flag.Shorthand)
}

// ========== Version JSON/YAML Output Tests ==========

func TestVersionCmd_JSONOutput(t *testing.T) {
	// Simulate JSON output from VersionInfo struct
	info := VersionInfo{
		Version:   "v0.57.1",
		Commit:    "abc123",
		BuildTime: "2024-01-01T00:00:00Z",
		Go:        "go1.25.6",
		Platform:  "darwin/arm64",
	}

	data, err := json.Marshal(info)
	require.NoError(t, err, "VersionInfo should be JSON-serializable")

	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err, "VersionInfo JSON should be valid")

	assert.Equal(t, "v0.57.1", parsed["version"])
	assert.Equal(t, "abc123", parsed["commit"])
	assert.Equal(t, "go1.25.6", parsed["go"])
	assert.Equal(t, "darwin/arm64", parsed["platform"])
}

func TestVersionCmd_YAMLOutput(t *testing.T) {
	info := VersionInfo{
		Version:   "v0.57.1",
		Commit:    "abc123",
		BuildTime: "2024-01-01T00:00:00Z",
		Go:        "go1.25.6",
		Platform:  "darwin/arm64",
	}

	data, err := yaml.Marshal(info)
	require.NoError(t, err, "VersionInfo should be YAML-serializable")

	var parsed map[string]interface{}
	err = yaml.Unmarshal(data, &parsed)
	require.NoError(t, err, "VersionInfo YAML should be valid")

	assert.Equal(t, "v0.57.1", parsed["version"])
	assert.Equal(t, "abc123", parsed["commit"])
	assert.Equal(t, "go1.25.6", parsed["go"])
	assert.Equal(t, "darwin/arm64", parsed["platform"])
}

// ========== Status JSON/YAML Output Tests ==========

func TestStatusInfo_JSONOutput(t *testing.T) {
	status := StatusInfo{
		Context: ContextInfo{
			App:       "my-app",
			Workspace: "main",
		},
		Containers: []ContainerInfo{
			{
				Name:   "dvm-my-app-main",
				ID:     "abc123def456",
				Status: "running",
				Image:  "dvm-main-my-app:latest",
			},
		},
		Runtime: RuntimeInfo{
			Type:   "docker",
			Name:   "OrbStack",
			Status: "active",
		},
	}

	data, err := json.Marshal(status)
	require.NoError(t, err, "StatusInfo should be JSON-serializable")

	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err, "StatusInfo JSON should be valid")

	// Verify context
	ctx := parsed["context"].(map[string]interface{})
	assert.Equal(t, "my-app", ctx["app"])
	assert.Equal(t, "main", ctx["workspace"])

	// Verify containers
	containers := parsed["containers"].([]interface{})
	assert.Len(t, containers, 1)
	c := containers[0].(map[string]interface{})
	assert.Equal(t, "dvm-my-app-main", c["name"])
	assert.Equal(t, "running", c["status"])

	// Verify runtime
	rt := parsed["runtime"].(map[string]interface{})
	assert.Equal(t, "docker", rt["type"])
	assert.Equal(t, "active", rt["status"])
}

func TestStatusInfo_YAMLOutput(t *testing.T) {
	status := StatusInfo{
		Context: ContextInfo{
			App:       "my-app",
			Workspace: "main",
		},
		Containers: []ContainerInfo{},
		Runtime: RuntimeInfo{
			Type:   "docker",
			Status: "active",
		},
	}

	data, err := yaml.Marshal(status)
	require.NoError(t, err, "StatusInfo should be YAML-serializable")

	var parsed map[string]interface{}
	err = yaml.Unmarshal(data, &parsed)
	require.NoError(t, err, "StatusInfo YAML should be valid")

	ctx := parsed["context"].(map[string]interface{})
	assert.Equal(t, "my-app", ctx["app"])
	assert.Equal(t, "main", ctx["workspace"])
}

// ========== Consistency: No commands use manual flag registration ==========

func TestRegistryGetDefaultsCmd_UsesAddOutputFlag(t *testing.T) {
	// Verify registry get-defaults uses AddOutputFlag (empty default)
	flag := registryGetDefaultsCmd.Flags().Lookup("output")
	require.NotNil(t, flag)
	assert.Equal(t, "o", flag.Shorthand,
		"registry get-defaults --output should have -o shorthand from AddOutputFlag")
	assert.Equal(t, "", flag.DefValue,
		"registry get-defaults --output should have empty default from AddOutputFlag")
}
