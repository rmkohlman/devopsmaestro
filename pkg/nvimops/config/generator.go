package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"devopsmaestro/pkg/nvimops/plugin"
)

// Generator creates Neovim configuration files from CoreConfig.
type Generator struct {
	// IndentSize is spaces per indent level (default: 2 for Lua, tabs for some)
	IndentSize int
	// UseTabs uses tabs instead of spaces
	UseTabs bool
}

// NewGenerator creates a new config generator with default settings.
func NewGenerator() *Generator {
	return &Generator{
		IndentSize: 2,
		UseTabs:    true, // Match nvim-config style (uses tabs)
	}
}

// GeneratedConfig holds all generated Lua files.
type GeneratedConfig struct {
	// InitLua is the root init.lua content
	InitLua string

	// LazyLua is the lazy.nvim bootstrap (lua/{namespace}/lazy.lua)
	LazyLua string

	// CoreInitLua is lua/{namespace}/core/init.lua
	CoreInitLua string

	// OptionsLua is lua/{namespace}/core/options.lua
	OptionsLua string

	// KeymapsLua is lua/{namespace}/core/keymaps.lua
	KeymapsLua string

	// AutocmdsLua is lua/{namespace}/core/autocmds.lua
	AutocmdsLua string

	// PluginsInitLua is lua/{namespace}/plugins/init.lua (base plugins)
	PluginsInitLua string
}

// Generate creates all Lua configuration files from a CoreConfig.
func (g *Generator) Generate(cfg *CoreConfig) (*GeneratedConfig, error) {
	ns := cfg.Namespace
	if ns == "" {
		ns = "workspace"
	}

	return &GeneratedConfig{
		InitLua:        g.generateInitLua(cfg, ns),
		LazyLua:        g.generateLazyLua(ns),
		CoreInitLua:    g.generateCoreInitLua(ns),
		OptionsLua:     g.generateOptionsLua(cfg),
		KeymapsLua:     g.generateKeymapsLua(cfg),
		AutocmdsLua:    g.generateAutocmdsLua(cfg),
		PluginsInitLua: g.generatePluginsInitLua(cfg),
	}, nil
}

// WriteToDirectory writes all generated files to the specified nvim config directory.
// This creates the full structure:
//
//	{dir}/init.lua
//	{dir}/lua/{namespace}/lazy.lua
//	{dir}/lua/{namespace}/core/init.lua
//	{dir}/lua/{namespace}/core/options.lua
//	{dir}/lua/{namespace}/core/keymaps.lua
//	{dir}/lua/{namespace}/core/autocmds.lua
//	{dir}/lua/{namespace}/plugins/init.lua
func (g *Generator) WriteToDirectory(cfg *CoreConfig, plugins []*plugin.Plugin, dir string) error {
	ns := cfg.Namespace
	if ns == "" {
		ns = "workspace"
	}

	generated, err := g.Generate(cfg)
	if err != nil {
		return err
	}

	// Create directory structure
	dirs := []string{
		dir,
		filepath.Join(dir, "lua", ns),
		filepath.Join(dir, "lua", ns, "core"),
		filepath.Join(dir, "lua", ns, "plugins"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", d, err)
		}
	}

	// Write core files
	files := map[string]string{
		filepath.Join(dir, "init.lua"):                        generated.InitLua,
		filepath.Join(dir, "lua", ns, "lazy.lua"):             generated.LazyLua,
		filepath.Join(dir, "lua", ns, "core", "init.lua"):     generated.CoreInitLua,
		filepath.Join(dir, "lua", ns, "core", "options.lua"):  generated.OptionsLua,
		filepath.Join(dir, "lua", ns, "core", "keymaps.lua"):  generated.KeymapsLua,
		filepath.Join(dir, "lua", ns, "core", "autocmds.lua"): generated.AutocmdsLua,
		filepath.Join(dir, "lua", ns, "plugins", "init.lua"):  generated.PluginsInitLua,
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", path, err)
		}
	}

	// Write plugin files
	if len(plugins) > 0 {
		pluginGen := plugin.NewGenerator()
		for _, p := range plugins {
			if !p.Enabled {
				continue
			}
			lua, err := pluginGen.GenerateLuaFile(p)
			if err != nil {
				return fmt.Errorf("failed to generate plugin %s: %w", p.Name, err)
			}
			pluginPath := filepath.Join(dir, "lua", ns, "plugins", p.Name+".lua")
			if err := os.WriteFile(pluginPath, []byte(lua), 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", pluginPath, err)
			}
		}
	}

	return nil
}

// indent returns the indentation string
func (g *Generator) indent(level int) string {
	if g.UseTabs {
		return strings.Repeat("\t", level)
	}
	return strings.Repeat(" ", g.IndentSize*level)
}

// generateInitLua creates the root init.lua
// Leader key MUST be set before lazy.nvim loads for mappings to work
func (g *Generator) generateInitLua(cfg *CoreConfig, namespace string) string {
	leader := cfg.Leader
	if leader == "" {
		leader = " "
	}
	return fmt.Sprintf(`-- Set leader key BEFORE loading lazy.nvim (required for mappings)
vim.g.mapleader = "%s"
vim.g.maplocalleader = "%s"

require("%s.core")
require("%s.lazy")
`, leader, leader, namespace, namespace)
}

// generateLazyLua creates the lazy.nvim bootstrap file
func (g *Generator) generateLazyLua(namespace string) string {
	return fmt.Sprintf(`local lazypath = vim.fn.stdpath("data") .. "/lazy/lazy.nvim"
if not vim.loop.fs_stat(lazypath) then
%[1]svim.fn.system({
%[1]s%[1]s"git",
%[1]s%[1]s"clone",
%[1]s%[1]s"--filter=blob:none",
%[1]s%[1]s"https://github.com/folke/lazy.nvim.git",
%[1]s%[1]s"--branch=stable", -- latest stable release
%[1]s%[1]slazypath,
%[1]s})
end
vim.opt.rtp:prepend(lazypath)

require("lazy").setup({ { import = "%[2]s.plugins" } }, {
%[1]schange_detection = {
%[1]s%[1]snotify = false,
%[1]s},
})
`, g.indent(1), namespace)
}

// generateCoreInitLua creates the core/init.lua that requires other core modules
func (g *Generator) generateCoreInitLua(namespace string) string {
	return fmt.Sprintf(`require("%s.core.options")
require("%s.core.keymaps")
require("%s.core.autocmds")
`, namespace, namespace, namespace)
}

// generateOptionsLua creates the core/options.lua file
func (g *Generator) generateOptionsLua(cfg *CoreConfig) string {
	var lua strings.Builder

	// vim.cmd for global settings
	if v, ok := cfg.Globals["netrw_liststyle"]; ok {
		lua.WriteString(fmt.Sprintf("vim.cmd(\"let g:netrw_liststyle = %v\")\n\n", v))
	}

	lua.WriteString("local opt = vim.opt\n\n")

	// Group options by category for readability
	categories := map[string][]string{
		"numbers":    {"relativenumber", "number"},
		"tabs":       {"tabstop", "shiftwidth", "expandtab", "autoindent"},
		"search":     {"ignorecase", "smartcase", "cursorline"},
		"appearance": {"termguicolors", "background", "signcolumn", "wrap"},
		"behavior":   {"backspace", "splitright", "splitbelow", "timeoutlen"},
		"clipboard":  {"clipboard"},
	}

	categoryOrder := []string{"numbers", "tabs", "search", "appearance", "behavior", "clipboard"}
	written := make(map[string]bool)

	for _, cat := range categoryOrder {
		opts := categories[cat]
		wroteAny := false
		for _, key := range opts {
			if val, ok := cfg.Options[key]; ok {
				if !wroteAny {
					// Add comment for category (except first)
					if lua.Len() > 50 {
						lua.WriteString("\n")
					}
				}
				g.writeOption(&lua, key, val)
				written[key] = true
				wroteAny = true
			}
		}
	}

	// Write any remaining options not in categories
	var remaining []string
	for key := range cfg.Options {
		if !written[key] {
			remaining = append(remaining, key)
		}
	}
	sort.Strings(remaining)

	if len(remaining) > 0 {
		lua.WriteString("\n")
		for _, key := range remaining {
			g.writeOption(&lua, key, cfg.Options[key])
		}
	}

	return lua.String()
}

func (g *Generator) writeOption(lua *strings.Builder, key string, val interface{}) {
	switch v := val.(type) {
	case bool:
		lua.WriteString(fmt.Sprintf("opt.%s = %t\n", key, v))
	case int, int64, float64:
		lua.WriteString(fmt.Sprintf("opt.%s = %v\n", key, v))
	case string:
		// Handle special cases like clipboard:append
		if strings.HasPrefix(v, "append:") {
			appendVal := strings.TrimPrefix(v, "append:")
			lua.WriteString(fmt.Sprintf("opt.%s:append(\"%s\")\n", key, appendVal))
		} else {
			lua.WriteString(fmt.Sprintf("opt.%s = \"%s\"\n", key, v))
		}
	default:
		lua.WriteString(fmt.Sprintf("opt.%s = \"%v\"\n", key, v))
	}
}

// generateKeymapsLua creates the core/keymaps.lua file
func (g *Generator) generateKeymapsLua(cfg *CoreConfig) string {
	var lua strings.Builder

	// Leader key
	leader := cfg.Leader
	if leader == "" {
		leader = " "
	}
	lua.WriteString(fmt.Sprintf("vim.g.mapleader = \"%s\"\n\n", escapeString(leader)))
	lua.WriteString("local keymap = vim.keymap\n\n")

	// Write keymaps
	for _, km := range cfg.Keymaps {
		g.writeKeymap(&lua, &km)
	}

	return lua.String()
}

func (g *Generator) writeKeymap(lua *strings.Builder, km *Keymap) {
	mode := km.Mode
	if mode == "" {
		mode = "n"
	}

	// Build options table
	var opts []string
	if km.Desc != "" {
		opts = append(opts, fmt.Sprintf("desc = \"%s\"", escapeString(km.Desc)))
	}
	if km.Silent {
		opts = append(opts, "silent = true")
	}

	optsStr := ""
	if len(opts) > 0 {
		optsStr = fmt.Sprintf(", { %s }", strings.Join(opts, ", "))
	}

	lua.WriteString(fmt.Sprintf("keymap.set(\"%s\", \"%s\", \"%s\"%s)\n",
		mode,
		escapeString(km.Key),
		escapeString(km.Action),
		optsStr,
	))
}

// generateAutocmdsLua creates the core/autocmds.lua file
func (g *Generator) generateAutocmdsLua(cfg *CoreConfig) string {
	var lua strings.Builder

	// Group autocmds by group name
	groups := make(map[string][]Autocmd)
	var groupOrder []string

	for _, ac := range cfg.Autocmds {
		if _, exists := groups[ac.Group]; !exists {
			groupOrder = append(groupOrder, ac.Group)
		}
		groups[ac.Group] = append(groups[ac.Group], ac)
	}

	for _, groupName := range groupOrder {
		autocmds := groups[groupName]

		// Create the augroup
		varName := strings.ToLower(groupName) + "_group"
		varName = strings.ReplaceAll(varName, "-", "_")
		lua.WriteString(fmt.Sprintf("local %s = vim.api.nvim_create_augroup(\"%s\", { clear = true })\n\n",
			varName, groupName))

		// Create autocmds in this group
		for _, ac := range autocmds {
			lua.WriteString("vim.api.nvim_create_autocmd(")

			// Events
			if len(ac.Events) == 1 {
				lua.WriteString(fmt.Sprintf("\"%s\"", ac.Events[0]))
			} else {
				lua.WriteString("{ ")
				for i, e := range ac.Events {
					if i > 0 {
						lua.WriteString(", ")
					}
					lua.WriteString(fmt.Sprintf("\"%s\"", e))
				}
				lua.WriteString(" }")
			}

			lua.WriteString(", {\n")
			lua.WriteString(fmt.Sprintf("%sgroup = %s,\n", g.indent(1), varName))

			if ac.Pattern != "" {
				lua.WriteString(fmt.Sprintf("%spattern = \"%s\",\n", g.indent(1), ac.Pattern))
			}

			if ac.Callback != "" {
				lua.WriteString(fmt.Sprintf("%scallback = function()\n", g.indent(1)))
				// Indent callback code
				lines := strings.Split(ac.Callback, "\n")
				for _, line := range lines {
					if strings.TrimSpace(line) != "" {
						lua.WriteString(fmt.Sprintf("%s%s\n", g.indent(2), strings.TrimSpace(line)))
					}
				}
				lua.WriteString(fmt.Sprintf("%send,\n", g.indent(1)))
			} else if ac.Command != "" {
				lua.WriteString(fmt.Sprintf("%scommand = \"%s\",\n", g.indent(1), escapeString(ac.Command)))
			}

			lua.WriteString("})\n\n")
		}
	}

	return lua.String()
}

// generatePluginsInitLua creates the plugins/init.lua for base plugins
func (g *Generator) generatePluginsInitLua(cfg *CoreConfig) string {
	var lua strings.Builder

	lua.WriteString("return {\n")
	for _, repo := range cfg.BasePlugins {
		lua.WriteString(fmt.Sprintf("%s\"%s\",\n", g.indent(1), repo))
	}
	lua.WriteString("}\n")

	return lua.String()
}

// escapeString escapes special characters in a Lua string
func escapeString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}
