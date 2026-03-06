package models

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestWorkspace_FromYAML_AllFields(t *testing.T) {
	yamlContent := `
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: dev
  app: fastapi-test
  annotations:
    description: "Development workspace"
spec:
  image:
    name: golang:1.21
  nvim:
    structure: lazyvim
    theme: tokyonight-night
    pluginPackage: maestro
    plugins:
      - nvim-treesitter
      - telescope.nvim
  terminal:
    prompt: starship
    package: maestro
    plugins:
      - zsh-syntax-highlighting
      - zsh-autosuggestions
  gitrepo: fastapi-test
  container:
    user: dev
    uid: 1000
    gid: 1000
    workingDir: /workspace
`

	var wsYAML WorkspaceYAML
	err := yaml.Unmarshal([]byte(yamlContent), &wsYAML)
	require.NoError(t, err, "failed to parse YAML")

	// Convert to workspace model
	workspace := &Workspace{AppID: 1}
	workspace.FromYAML(wsYAML)

	// Verify all fields were mapped
	assert.Equal(t, "dev", workspace.Name)
	assert.Equal(t, "golang:1.21", workspace.ImageName)
	assert.Equal(t, "created", workspace.Status)

	// Description
	require.True(t, workspace.Description.Valid)
	assert.Equal(t, "Development workspace", workspace.Description.String)

	// Nvim fields
	require.True(t, workspace.Theme.Valid)
	assert.Equal(t, "tokyonight-night", workspace.Theme.String)

	require.True(t, workspace.NvimStructure.Valid)
	assert.Equal(t, "lazyvim", workspace.NvimStructure.String)

	require.True(t, workspace.NvimPlugins.Valid)
	assert.Equal(t, "nvim-treesitter,telescope.nvim", workspace.NvimPlugins.String)

	// Terminal fields
	require.True(t, workspace.TerminalPrompt.Valid)
	assert.Equal(t, "starship", workspace.TerminalPrompt.String)

	require.True(t, workspace.TerminalPackage.Valid)
	assert.Equal(t, "maestro", workspace.TerminalPackage.String)

	require.True(t, workspace.TerminalPlugins.Valid)
	termPlugins := workspace.GetTerminalPlugins()
	assert.Equal(t, []string{"zsh-syntax-highlighting", "zsh-autosuggestions"}, termPlugins)

	// Note: GitRepo resolution happens in the handler, not in FromYAML
	// So GitRepoID is not set here
}

func TestWorkspace_ToYAML_AllFields(t *testing.T) {
	workspace := &Workspace{
		ID:              1,
		AppID:           2,
		Name:            "dev",
		ImageName:       "golang:1.21",
		Status:          "running",
		Description:     sql.NullString{String: "Test workspace", Valid: true},
		Theme:           sql.NullString{String: "catppuccin-mocha", Valid: true},
		NvimStructure:   sql.NullString{String: "lazyvim", Valid: true},
		NvimPlugins:     sql.NullString{String: "plugin1,plugin2", Valid: true},
		TerminalPrompt:  sql.NullString{String: "starship", Valid: true},
		TerminalPackage: sql.NullString{String: "my-package", Valid: true},
	}
	workspace.SetTerminalPlugins([]string{"zsh-autosuggestions", "zsh-syntax-highlighting"})

	// Convert to YAML
	yamlDoc := workspace.ToYAML("fastapi-test", "my-gitrepo")

	// Verify structure
	assert.Equal(t, "devopsmaestro.io/v1", yamlDoc.APIVersion)
	assert.Equal(t, "Workspace", yamlDoc.Kind)
	assert.Equal(t, "dev", yamlDoc.Metadata.Name)
	assert.Equal(t, "fastapi-test", yamlDoc.Metadata.App)
	assert.Equal(t, "Test workspace", yamlDoc.Metadata.Annotations["description"])

	// Verify nvim config
	assert.Equal(t, "lazyvim", yamlDoc.Spec.Nvim.Structure)
	assert.Equal(t, "catppuccin-mocha", yamlDoc.Spec.Nvim.Theme)
	assert.Equal(t, []string{"plugin1", "plugin2"}, yamlDoc.Spec.Nvim.Plugins)

	// Verify terminal config
	assert.Equal(t, "starship", yamlDoc.Spec.Terminal.Prompt)
	assert.Equal(t, "my-package", yamlDoc.Spec.Terminal.Package)
	assert.Equal(t, []string{"zsh-autosuggestions", "zsh-syntax-highlighting"}, yamlDoc.Spec.Terminal.Plugins)

	// Verify gitrepo
	assert.Equal(t, "my-gitrepo", yamlDoc.Spec.GitRepo)
}

func TestWorkspace_YAML_RoundTrip(t *testing.T) {
	// Start with a YAML document
	originalYAML := `
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: roundtrip-test
  app: myapp
  annotations:
    description: "Round trip test"
spec:
  image:
    name: ubuntu:22.04
  nvim:
    structure: custom
    theme: gruvbox
    plugins:
      - plugin-a
      - plugin-b
  terminal:
    prompt: oh-my-posh
    plugins:
      - terminal-plugin-1
    package: my-terminal-package
  gitrepo: my-repo
  container:
    user: dev
    uid: 1000
    gid: 1000
    workingDir: /workspace
`

	// Parse YAML
	var wsYAML WorkspaceYAML
	err := yaml.Unmarshal([]byte(originalYAML), &wsYAML)
	require.NoError(t, err)

	// Convert to model
	workspace := &Workspace{AppID: 1}
	workspace.FromYAML(wsYAML)

	// Convert back to YAML
	resultYAML := workspace.ToYAML("myapp", "my-repo")

	// Verify all fields survived the round trip
	assert.Equal(t, "roundtrip-test", resultYAML.Metadata.Name)
	assert.Equal(t, "myapp", resultYAML.Metadata.App)
	assert.Equal(t, "Round trip test", resultYAML.Metadata.Annotations["description"])
	assert.Equal(t, "ubuntu:22.04", resultYAML.Spec.Image.Name)

	// Nvim
	assert.Equal(t, "custom", resultYAML.Spec.Nvim.Structure)
	assert.Equal(t, "gruvbox", resultYAML.Spec.Nvim.Theme)
	assert.Equal(t, []string{"plugin-a", "plugin-b"}, resultYAML.Spec.Nvim.Plugins)

	// Terminal
	assert.Equal(t, "oh-my-posh", resultYAML.Spec.Terminal.Prompt)
	assert.Equal(t, "my-terminal-package", resultYAML.Spec.Terminal.Package)
	assert.Equal(t, []string{"terminal-plugin-1"}, resultYAML.Spec.Terminal.Plugins)

	// GitRepo
	assert.Equal(t, "my-repo", resultYAML.Spec.GitRepo)
}

func TestWorkspace_FromYAML_PartialFields(t *testing.T) {
	// Test that FromYAML works even with minimal fields
	yamlContent := `
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: minimal
  app: myapp
spec:
  image:
    name: alpine:latest
  container:
    user: dev
`

	var wsYAML WorkspaceYAML
	err := yaml.Unmarshal([]byte(yamlContent), &wsYAML)
	require.NoError(t, err)

	workspace := &Workspace{AppID: 1}
	workspace.FromYAML(wsYAML)

	assert.Equal(t, "minimal", workspace.Name)
	assert.Equal(t, "alpine:latest", workspace.ImageName)
	assert.False(t, workspace.Description.Valid)
	assert.False(t, workspace.Theme.Valid)
	assert.False(t, workspace.NvimStructure.Valid)
	assert.False(t, workspace.NvimPlugins.Valid)
	assert.False(t, workspace.TerminalPrompt.Valid)
	assert.False(t, workspace.TerminalPackage.Valid)
	assert.False(t, workspace.TerminalPlugins.Valid)
}

func TestWorkspace_ToYAML_MarshalToBytes(t *testing.T) {
	workspace := &Workspace{
		Name:      "yaml-bytes-test",
		ImageName: "golang:1.21",
		Theme:     sql.NullString{String: "tokyonight", Valid: true},
	}

	yamlDoc := workspace.ToYAML("testapp", "")
	bytes, err := yaml.Marshal(yamlDoc)
	require.NoError(t, err)

	// Verify it's valid YAML
	yamlStr := string(bytes)
	assert.Contains(t, yamlStr, "apiVersion: devopsmaestro.io/v1")
	assert.Contains(t, yamlStr, "kind: Workspace")
	assert.Contains(t, yamlStr, "name: yaml-bytes-test")
	assert.Contains(t, yamlStr, "app: testapp")
	assert.Contains(t, yamlStr, "theme: tokyonight")

	// When gitRepoName is empty, gitrepo should not appear in YAML
	assert.False(t, strings.Contains(yamlStr, "gitrepo:"), "gitrepo should be omitted when empty")
}
