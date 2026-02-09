package models

import (
	"database/sql"
	"strings"
	"time"
)

// Workspace represents a workspace entity associated with an app.
type Workspace struct {
	ID            int            `db:"id" json:"id" yaml:"-"`
	AppID         int            `db:"app_id" json:"app_id" yaml:"-"`
	Name          string         `db:"name" json:"name" yaml:"name"`
	Description   sql.NullString `db:"description" json:"description,omitempty" yaml:"description,omitempty"`
	ImageName     string         `db:"image_name" json:"image_name" yaml:"image_name"`
	ContainerID   sql.NullString `db:"container_id" json:"container_id,omitempty" yaml:"-"`
	Status        string         `db:"status" json:"status" yaml:"status"`
	NvimStructure sql.NullString `db:"nvim_structure" json:"nvim_structure,omitempty" yaml:"-"`
	NvimPlugins   sql.NullString `db:"nvim_plugins" json:"nvim_plugins,omitempty" yaml:"-"` // Comma-separated plugin names
	CreatedAt     time.Time      `db:"created_at" json:"created_at" yaml:"-"`
	UpdatedAt     time.Time      `db:"updated_at" json:"updated_at" yaml:"-"`
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
}

// ImageConfig defines the container image configuration
type ImageConfig struct {
	Name      string `yaml:"name"`
	BuildFrom string `yaml:"buildFrom,omitempty"`
	BaseImage string `yaml:"baseImage,omitempty"`
}

// DevBuildConfig defines the build configuration for the dev environment.
// This focuses on developer tools added on top of the app's base image.
type DevBuildConfig struct {
	Args     map[string]string `yaml:"args,omitempty"`
	DevStage DevStageConfig    `yaml:"devStage"`
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
	Type       string `yaml:"type,omitempty"`       // tmux, zellij, screen
	ConfigPath string `yaml:"configPath,omitempty"` // Path to config file to mount
	Autostart  bool   `yaml:"autostart,omitempty"`  // Start on attach
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
	Structure    string   `yaml:"structure"`              // lazyvim, custom, nvchad, astronvim
	Plugins      []string `yaml:"plugins,omitempty"`      // List of plugin names (references to DB)
	CustomConfig string   `yaml:"customConfig,omitempty"` // Raw Lua config
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
func (w *Workspace) ToYAML(appName string) WorkspaceYAML {
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

	// Create default spec with minimal configuration
	// This will be enhanced when we implement config storage in DB
	return WorkspaceYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Workspace",
		Metadata: WorkspaceMetadata{
			Name:        w.Name,
			App:         appName,
			Labels:      make(map[string]string),
			Annotations: annotations,
		},
		Spec: WorkspaceSpec{
			Image: ImageConfig{
				Name: w.ImageName,
			},
			Nvim: nvimConfig,
			Container: ContainerConfig{
				User:       "dev",
				UID:        1000,
				GID:        1000,
				WorkingDir: "/workspace",
				Command:    []string{"/bin/zsh", "-l"},
			},
		},
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
}
