package models

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"
)

// Workspace represents a workspace entity associated with an app.
type Workspace struct {
	ID                 int            `db:"id" json:"id" yaml:"-"`
	AppID              int            `db:"app_id" json:"app_id" yaml:"-"`
	Name               string         `db:"name" json:"name" yaml:"name"`
	Slug               string         `db:"slug" json:"slug" yaml:"slug"`
	Description        sql.NullString `db:"description" json:"description,omitempty" yaml:"description,omitempty"`
	ImageName          string         `db:"image_name" json:"image_name" yaml:"image_name"`
	ContainerID        sql.NullString `db:"container_id" json:"container_id,omitempty" yaml:"-"`
	Status             string         `db:"status" json:"status" yaml:"status"`
	SSHAgentForwarding bool           `db:"ssh_agent_forwarding" json:"ssh_agent_forwarding" yaml:"ssh_agent_forwarding"`
	Theme              sql.NullString `db:"theme" json:"theme,omitempty" yaml:"theme,omitempty"`
	NvimStructure      sql.NullString `db:"nvim_structure" json:"nvim_structure,omitempty" yaml:"-"`
	NvimPlugins        sql.NullString `db:"nvim_plugins" json:"nvim_plugins,omitempty" yaml:"-"` // Comma-separated plugin names
	TerminalPrompt     sql.NullString `db:"terminal_prompt" json:"terminal_prompt,omitempty" yaml:"-"`
	TerminalPlugins    sql.NullString `db:"terminal_plugins" json:"terminal_plugins,omitempty" yaml:"-"` // JSON array
	TerminalPackage    sql.NullString `db:"terminal_package" json:"terminal_package,omitempty" yaml:"-"`
	GitRepoID          sql.NullInt64  `db:"git_repo_id" json:"git_repo_id,omitempty" yaml:"-"`
	Env                sql.NullString `db:"env" json:"env,omitempty" yaml:"-"`
	CreatedAt          time.Time      `db:"created_at" json:"created_at" yaml:"-"`
	UpdatedAt          time.Time      `db:"updated_at" json:"updated_at" yaml:"-"`
}

// WorkspaceYAML represents the YAML serialization format for a workspace
type WorkspaceYAML struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   WorkspaceMetadata `yaml:"metadata"`
	Spec       WorkspaceSpec     `yaml:"spec"`
}

// WorkspaceMetadata contains workspace metadata
type WorkspaceMetadata struct {
	Name        string            `yaml:"name"`
	App         string            `yaml:"app"`
	Domain      string            `yaml:"domain,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// WorkspaceSpec contains the complete workspace specification.
// Workspaces focus on DEVELOPER ENVIRONMENT concerns:
// - Editor configuration (nvim)
// - Shell setup (zsh, bash, oh-my-zsh)
// - Terminal multiplexer (tmux)
// - Dev user setup (UID/GID mapping)
// - Dev mounts (SSH keys, gitconfig)
//
// App-level concerns (language, build, services, ports) belong in AppSpec.
type WorkspaceSpec struct {
	Image     ImageConfig       `yaml:"image"`
	Build     DevBuildConfig    `yaml:"build"`
	Shell     ShellConfig       `yaml:"shell"`
	Terminal  TerminalConfig    `yaml:"terminal,omitempty"`
	Nvim      NvimConfig        `yaml:"nvim"`
	Mounts    []MountConfig     `yaml:"mounts,omitempty"`
	SSHKey    SSHKeyConfig      `yaml:"sshKey,omitempty"`
	Env       map[string]string `yaml:"env,omitempty"`
	Container ContainerConfig   `yaml:"container"`
	GitRepo   string            `yaml:"gitrepo,omitempty"` // Name of GitRepo resource to clone
}

// ImageConfig defines the container image configuration
type ImageConfig struct {
	Name      string `yaml:"name"`
	BuildFrom string `yaml:"buildFrom,omitempty"`
	BaseImage string `yaml:"baseImage,omitempty"`
}

// BaseStageConfig defines configuration for the base (app) build stage.
// Packages listed here are installed via apt-get in the base stage,
// alongside any auto-detected system dependencies.
type BaseStageConfig struct {
	Packages []string `yaml:"packages,omitempty"`
}

// DevBuildConfig defines the build configuration for the dev environment.
// This focuses on developer tools added on top of the app's base image.
type DevBuildConfig struct {
	Args      map[string]string `yaml:"args,omitempty"`
	BaseStage BaseStageConfig   `yaml:"baseStage,omitempty"`
	DevStage  DevStageConfig    `yaml:"devStage"`
}

// DevStageConfig defines what developer tools to add in the dev stage.
// These are tools for the developer, not the app itself.
type DevStageConfig struct {
	Packages       []string `yaml:"packages,omitempty"`       // System packages (ripgrep, fd, etc.)
	DevTools       []string `yaml:"devTools,omitempty"`       // Language dev tools (gopls, delve, pylsp, etc.)
	CustomCommands []string `yaml:"customCommands,omitempty"` // Custom setup commands
}

// TerminalConfig defines terminal multiplexer configuration
type TerminalConfig struct {
	Type       string   `yaml:"type,omitempty"`       // tmux, zellij, screen
	ConfigPath string   `yaml:"configPath,omitempty"` // Path to config file to mount
	Autostart  bool     `yaml:"autostart,omitempty"`  // Start on attach
	Prompt     string   `yaml:"prompt,omitempty"`     // Terminal prompt name (e.g., "starship")
	Plugins    []string `yaml:"plugins,omitempty"`    // Terminal plugins to install
	Package    string   `yaml:"package,omitempty"`    // Reference to a terminal package by name
}

// ShellConfig defines shell configuration
type ShellConfig struct {
	Type      string   `yaml:"type"`                // zsh, bash
	Framework string   `yaml:"framework,omitempty"` // oh-my-zsh
	Theme     string   `yaml:"theme,omitempty"`     // starship, powerlevel10k
	Plugins   []string `yaml:"plugins,omitempty"`
	CustomRc  string   `yaml:"customRc,omitempty"`
}

// NvimConfig defines Neovim configuration
type NvimConfig struct {
	Structure     string   `yaml:"structure"`               // lazyvim, custom, nvchad, astronvim
	Theme         string   `yaml:"theme,omitempty"`         // Theme name (e.g., "tokyonight-night", "catppuccin-mocha")
	PluginPackage string   `yaml:"pluginPackage,omitempty"` // Reference to a plugin package by name (e.g., "go-dev")
	Plugins       []string `yaml:"plugins,omitempty"`       // List of plugin names (references to DB)
	MergeMode     string   `yaml:"mergeMode,omitempty"`     // How to merge package + plugins: "append" (default), "replace"
	CustomConfig  string   `yaml:"customConfig,omitempty"`  // Raw Lua config
}

// MountConfig defines a container mount
type MountConfig struct {
	Type        string `yaml:"type"` // bind, volume, tmpfs
	Source      string `yaml:"source"`
	Destination string `yaml:"destination"`
	ReadOnly    bool   `yaml:"readOnly,omitempty"`
}

// SSHKeyConfig defines SSH key configuration
type SSHKeyConfig struct {
	Mode string `yaml:"mode"` // mount_host, global_dvm, per_project, generate
	Path string `yaml:"path,omitempty"`
}

// ContainerConfig defines container runtime settings for the dev environment.
// Port exposure is handled at the App level, not here.
type ContainerConfig struct {
	User       string         `yaml:"user,omitempty"`
	UID        int            `yaml:"uid,omitempty"`
	GID        int            `yaml:"gid,omitempty"`
	WorkingDir string         `yaml:"workingDir,omitempty"`
	Command    []string       `yaml:"command,omitempty"`
	Entrypoint []string       `yaml:"entrypoint,omitempty"`
	Resources  ResourceLimits `yaml:"resources,omitempty"`
}

// ResourceLimits defines container resource limits
type ResourceLimits struct {
	CPUs   string `yaml:"cpus,omitempty"`
	Memory string `yaml:"memory,omitempty"`
}

// ToYAML converts a Workspace to YAML format
func (w *Workspace) ToYAML(appName string, gitRepoName string) WorkspaceYAML {
	description := ""
	if w.Description.Valid {
		description = w.Description.String
	}

	annotations := make(map[string]string)
	if description != "" {
		annotations["description"] = description
	}

	// Parse nvim config from database
	nvimConfig := NvimConfig{}
	if w.NvimStructure.Valid {
		nvimConfig.Structure = w.NvimStructure.String
	}
	if w.NvimPlugins.Valid && w.NvimPlugins.String != "" {
		// Split comma-separated plugin names
		pluginNames := strings.Split(w.NvimPlugins.String, ",")
		nvimConfig.Plugins = pluginNames
	}
	// Include theme in nvim config if set at workspace level
	if w.Theme.Valid && w.Theme.String != "" {
		nvimConfig.Theme = w.Theme.String
	}

	// Parse terminal config from database
	terminalConfig := TerminalConfig{}
	if w.TerminalPrompt.Valid {
		terminalConfig.Prompt = w.TerminalPrompt.String
	}
	if w.TerminalPlugins.Valid {
		terminalConfig.Plugins = w.GetTerminalPlugins()
	}
	if w.TerminalPackage.Valid {
		terminalConfig.Package = w.TerminalPackage.String
	}

	// Include env variables if any are set
	envMap := w.GetEnv()
	if len(envMap) == 0 {
		envMap = nil // Ensure omitempty works for YAML serialization
	}

	// Create default spec with minimal configuration
	// This will be enhanced when we implement config storage in DB
	spec := WorkspaceSpec{
		Image: ImageConfig{
			Name: w.ImageName,
		},
		Nvim:     nvimConfig,
		Terminal: terminalConfig,
		Env:      envMap,
		Container: ContainerConfig{
			User:       "dev",
			UID:        1000,
			GID:        1000,
			WorkingDir: "/workspace",
			Command:    []string{"/bin/zsh", "-l"},
		},
	}

	// Add gitrepo if provided
	if gitRepoName != "" {
		spec.GitRepo = gitRepoName
	}

	return WorkspaceYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Workspace",
		Metadata: WorkspaceMetadata{
			Name:        w.Name,
			App:         appName,
			Labels:      make(map[string]string),
			Annotations: annotations,
		},
		Spec: spec,
	}
}

// FromYAML converts YAML format to a Workspace
func (w *Workspace) FromYAML(yaml WorkspaceYAML) {
	w.Name = yaml.Metadata.Name
	w.ImageName = yaml.Spec.Image.Name
	w.Status = "created"

	if desc, ok := yaml.Metadata.Annotations["description"]; ok {
		w.Description = sql.NullString{String: desc, Valid: true}
	}

	// Nvim configuration
	if yaml.Spec.Nvim.Theme != "" {
		w.Theme = sql.NullString{String: yaml.Spec.Nvim.Theme, Valid: true}
	}
	if yaml.Spec.Nvim.Structure != "" {
		w.NvimStructure = sql.NullString{String: yaml.Spec.Nvim.Structure, Valid: true}
	}
	if len(yaml.Spec.Nvim.Plugins) > 0 {
		w.NvimPlugins = sql.NullString{String: strings.Join(yaml.Spec.Nvim.Plugins, ","), Valid: true}
	}
	// Note: PluginPackage is stored in YAML but needs separate handling (package lookup or store as name)

	// Terminal configuration
	if yaml.Spec.Terminal.Prompt != "" {
		w.TerminalPrompt = sql.NullString{String: yaml.Spec.Terminal.Prompt, Valid: true}
	}
	if len(yaml.Spec.Terminal.Plugins) > 0 {
		w.SetTerminalPlugins(yaml.Spec.Terminal.Plugins)
	}
	if yaml.Spec.Terminal.Package != "" {
		w.TerminalPackage = sql.NullString{String: yaml.Spec.Terminal.Package, Valid: true}
	}

	// Environment variables
	if len(yaml.Spec.Env) > 0 {
		w.SetEnv(yaml.Spec.Env)
	}
	// Note: GitRepo resolution (name→ID) happens in the handler, not here
}

// GetTerminalPlugins returns the list of terminal plugins configured for this workspace.
// Returns an empty slice if no plugins are configured or if the JSON is invalid.
func (w *Workspace) GetTerminalPlugins() []string {
	if !w.TerminalPlugins.Valid || w.TerminalPlugins.String == "" {
		return []string{}
	}
	var plugins []string
	if err := json.Unmarshal([]byte(w.TerminalPlugins.String), &plugins); err != nil {
		// If unmarshal fails, return empty slice
		return []string{}
	}
	return plugins
}

// SetTerminalPlugins sets the terminal plugins for this workspace.
// Stores as a JSON array. If plugins is empty, sets to NULL.
func (w *Workspace) SetTerminalPlugins(plugins []string) {
	if len(plugins) == 0 {
		w.TerminalPlugins = sql.NullString{Valid: false}
		return
	}
	data, err := json.Marshal(plugins)
	if err != nil {
		// If marshal fails, set to invalid
		w.TerminalPlugins = sql.NullString{Valid: false}
		return
	}
	w.TerminalPlugins = sql.NullString{String: string(data), Valid: true}
}

// GetEnv returns the environment variables configured for this workspace.
// Returns an empty (non-nil) map if no env vars are configured.
func (w *Workspace) GetEnv() map[string]string {
	if !w.Env.Valid || w.Env.String == "" || w.Env.String == "{}" {
		return map[string]string{}
	}
	var env map[string]string
	if err := json.Unmarshal([]byte(w.Env.String), &env); err != nil {
		return map[string]string{}
	}
	return env
}

// SetEnv stores the environment variables for this workspace.
func (w *Workspace) SetEnv(env map[string]string) {
	if len(env) == 0 {
		w.Env = sql.NullString{String: "{}", Valid: true}
		return
	}
	data, err := json.Marshal(env)
	if err != nil {
		w.Env = sql.NullString{String: "{}", Valid: true}
		return
	}
	w.Env = sql.NullString{String: string(data), Valid: true}
}
