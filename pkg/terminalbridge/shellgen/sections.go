package shellgen

import (
	"fmt"
	"strings"

	"github.com/rmkohlman/MaestroTerminal/terminalops/plugin"
	"github.com/rmkohlman/MaestroTerminal/terminalops/shell"
)

// writeShellConfigs includes any additional shell configurations (aliases, functions, etc.).
func (g *ZshrcGenerator) writeShellConfigs(sb *strings.Builder, config ShellConfig) {
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

// writePlugins generates plugin source lines in correct load order.
func (g *ZshrcGenerator) writePlugins(sb *strings.Builder, config ShellConfig) {
	if len(config.Plugins) == 0 {
		return
	}

	// Filter enabled plugins only
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
		pluginDir = "${HOME}/.local/share/zsh/plugins"
	}

	gen := plugin.NewZshGenerator(pluginDir)
	output, err := gen.Generate(enabled)
	if err != nil {
		sb.WriteString(fmt.Sprintf("# Warning: failed to generate plugin config: %v\n", err))
		return
	}
	sb.WriteString(output)
	sb.WriteString("\n")
}

// writePromptConfig appends prompt initialization (e.g., starship init zsh).
func (g *ZshrcGenerator) writePromptConfig(sb *strings.Builder, config ShellConfig) {
	if config.PromptConfig == "" {
		return
	}

	sb.WriteString("# === Prompt Configuration ===\n")
	sb.WriteString(strings.TrimRight(config.PromptConfig, "\n"))
	sb.WriteString("\n\n")
}

// writeFooter writes the end-of-file marker.
func (g *ZshrcGenerator) writeFooter(sb *strings.Builder) {
	sb.WriteString("# End of .zshrc.workspace\n")
}
