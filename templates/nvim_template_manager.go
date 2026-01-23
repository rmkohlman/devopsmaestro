package templates

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed nvim
var nvimTemplates embed.FS

// NvimTemplateConfig represents the nvim template configuration
type NvimTemplateConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Core        struct {
		Enabled  bool `yaml:"enabled"`
		Options  bool `yaml:"options"`
		Keymaps  bool `yaml:"keymaps"`
		Autocmds bool `yaml:"autocmds"`
	} `yaml:"core"`
	Plugins struct {
		Essentials       []string `yaml:"essentials"`
		UI               []string `yaml:"ui"`
		Navigation       []string `yaml:"navigation"`
		Editor           []string `yaml:"editor"`
		LSP              []string `yaml:"lsp"`
		Git              []string `yaml:"git"`
		Quality          []string `yaml:"quality"`
		AI               []string `yaml:"ai"`
		LanguageSpecific []string `yaml:"language_specific"`
		Terminal         []string `yaml:"terminal"`
	} `yaml:"plugins"`
	LanguageServers []string `yaml:"language_servers"`
	Tools           struct {
		Formatters []string `yaml:"formatters"`
		Linters    []string `yaml:"linters"`
	} `yaml:"tools"`
}

// NvimTemplateManager handles nvim template operations
type NvimTemplateManager struct {
	config NvimTemplateConfig
}

// NewNvimTemplateManager creates a new nvim template manager
func NewNvimTemplateManager() (*NvimTemplateManager, error) {
	// Load default config
	configData, err := nvimTemplates.ReadFile("nvim/config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read nvim config: %w", err)
	}

	var config NvimTemplateConfig
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse nvim config: %w", err)
	}

	return &NvimTemplateManager{
		config: config,
	}, nil
}

// GenerateDockerfileCommands generates Dockerfile commands to install and configure nvim
func (m *NvimTemplateManager) GenerateDockerfileCommands() []string {
	commands := []string{
		"",
		"# Install Neovim and dependencies",
		"RUN apt-get update && apt-get install -y --no-install-recommends \\",
		"    neovim \\",
		"    git \\",
		"    curl \\",
		"    unzip \\",
		"    build-essential \\",
		"    ripgrep \\",
		"    fd-find \\",
		"    && rm -rf /var/lib/apt/lists/*",
		"",
		"# Create nvim config directory",
		"RUN mkdir -p /home/dev/.config/nvim",
		"",
	}

	return commands
}

// CopyTemplatesToContainer generates commands to copy nvim config into the container
// This should be called during container build to embed the configuration
func (m *NvimTemplateManager) CopyTemplatesToContainer(destPath string) error {
	// Walk through embedded files and copy them
	return fs.WalkDir(nvimTemplates, "nvim", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory and config.yaml
		if path == "nvim" || strings.HasSuffix(path, "config.yaml") {
			return nil
		}

		// Get relative path from nvim/ directory
		relPath, err := filepath.Rel("nvim", path)
		if err != nil {
			return err
		}

		destFile := filepath.Join(destPath, relPath)

		if d.IsDir() {
			// Create directory
			return os.MkdirAll(destFile, 0755)
		}

		// Copy file
		content, err := nvimTemplates.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(destFile), 0755); err != nil {
			return err
		}

		return os.WriteFile(destFile, content, 0644)
	})
}

// GetEnabledPlugins returns a list of all enabled plugins
func (m *NvimTemplateManager) GetEnabledPlugins() []string {
	plugins := make([]string, 0)

	plugins = append(plugins, m.config.Plugins.Essentials...)
	plugins = append(plugins, m.config.Plugins.UI...)
	plugins = append(plugins, m.config.Plugins.Navigation...)
	plugins = append(plugins, m.config.Plugins.Editor...)
	plugins = append(plugins, m.config.Plugins.LSP...)
	plugins = append(plugins, m.config.Plugins.Git...)
	plugins = append(plugins, m.config.Plugins.Quality...)
	plugins = append(plugins, m.config.Plugins.AI...)
	plugins = append(plugins, m.config.Plugins.LanguageSpecific...)
	plugins = append(plugins, m.config.Plugins.Terminal...)

	return plugins
}

// ExportToPath exports all nvim templates to a given path
// This is useful for adding configs to a build context
func (m *NvimTemplateManager) ExportToPath(basePath string) error {
	nvimPath := filepath.Join(basePath, ".config", "nvim")
	return m.CopyTemplatesToContainer(nvimPath)
}
