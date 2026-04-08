package shellgen

import (
	"fmt"
	"strings"

	"github.com/rmkohlman/MaestroTerminal/terminalops/plugin"
	"github.com/rmkohlman/MaestroTerminal/terminalops/shell"
)

// writeShellConfigs includes any additional shell configurations.
// Fish uses different syntax for aliases and functions, but the shell.Generator
// handles this based on the ShellType field.
func (g *FishGenerator) writeShellConfigs(sb *strings.Builder, config ShellConfig) {
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

// writePlugins generates plugin source lines for Fish.
// Uses Fish conditional sourcing: if test -f file; source file; end
func (g *FishGenerator) writePlugins(sb *strings.Builder, config ShellConfig) {
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
		pluginDir = "$HOME/.local/share/fish/plugins"
	}

	sb.WriteString("# === Shell Plugins ===\n")
	for _, p := range enabled {
		pluginPath := fmt.Sprintf("%s/%s/%s.plugin.fish", pluginDir, p.Name, p.Name)
		sb.WriteString(fmt.Sprintf("if test -f \"%s\"\n    source \"%s\"\nend\n", pluginPath, pluginPath))
	}
	sb.WriteString("\n")
}

// writePromptConfig appends prompt initialization for Fish.
func (g *FishGenerator) writePromptConfig(sb *strings.Builder, config ShellConfig) {
	if config.PromptConfig == "" {
		return
	}

	sb.WriteString("# === Prompt Configuration ===\n")
	sb.WriteString(strings.TrimRight(config.PromptConfig, "\n"))
	sb.WriteString("\n\n")
}

// writeFooter writes the end-of-file marker.
func (g *FishGenerator) writeFooter(sb *strings.Builder) {
	sb.WriteString("# End of config.fish.workspace\n")
}
