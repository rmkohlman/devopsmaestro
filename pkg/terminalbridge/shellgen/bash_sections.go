package shellgen

import (
	"fmt"
	"strings"

	"github.com/rmkohlman/MaestroTerminal/terminalops/plugin"
	"github.com/rmkohlman/MaestroTerminal/terminalops/shell"
)

// writeShellConfigs includes any additional shell configurations.
func (g *BashGenerator) writeShellConfigs(sb *strings.Builder, config ShellConfig) {
	if len(config.ShellConfigs) == 0 {
		return
	}

	gen := shell.NewGenerator()
	for _, s := range config.ShellConfigs {
		if s == nil || !s.Enabled {
			continue
		}
		output, err := gen.Generate(s)
		if err != nil {
			sb.WriteString(fmt.Sprintf("# Warning: failed to generate shell config %q: %v\n", s.Name, err))
			continue
		}
		sb.WriteString(output)
	}
}

// writePlugins generates plugin source lines for Bash.
// Uses conditional sourcing: if [ -f file ]; then source file; fi
func (g *BashGenerator) writePlugins(sb *strings.Builder, config ShellConfig) {
	if len(config.Plugins) == 0 {
		return
	}

	var enabled []*plugin.Plugin
	for _, p := range config.Plugins {
		if p != nil && p.Enabled {
			enabled = append(enabled, p)
		}
	}
	if len(enabled) == 0 {
		return
	}

	pluginDir := config.PluginDir
	if pluginDir == "" {
		pluginDir = "${HOME}/.local/share/bash/plugins"
	}

	sb.WriteString("# === Shell Plugins ===\n")
	for _, p := range enabled {
		pluginPath := fmt.Sprintf("%s/%s/%s.plugin.bash", pluginDir, p.Name, p.Name)
		sb.WriteString(fmt.Sprintf("if [ -f \"%s\" ]; then source \"%s\"; fi\n", pluginPath, pluginPath))
	}
	sb.WriteString("\n")
}

// writePromptConfig appends prompt initialization for Bash.
func (g *BashGenerator) writePromptConfig(sb *strings.Builder, config ShellConfig) {
	if config.PromptConfig == "" {
		return
	}

	sb.WriteString("# === Prompt Configuration ===\n")
	sb.WriteString(strings.TrimRight(config.PromptConfig, "\n"))
	sb.WriteString("\n\n")
}

// writeFooter writes the end-of-file marker.
func (g *BashGenerator) writeFooter(sb *strings.Builder) {
	sb.WriteString("# End of .bashrc.workspace\n")
}
