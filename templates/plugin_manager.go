package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// NvimPluginDef represents a user-defined nvim plugin
type NvimPluginDef struct {
	Name         string          `yaml:"name"`
	Description  string          `yaml:"description"`
	Repo         string          `yaml:"repo"`
	Dependencies []string        `yaml:"dependencies,omitempty"`
	Filetypes    []string        `yaml:"filetypes,omitempty"`
	Lazy         bool            `yaml:"lazy,omitempty"`
	Event        string          `yaml:"event,omitempty"`
	Keys         []NvimKeyDef    `yaml:"keys,omitempty"`
	Cmd          []string        `yaml:"cmd,omitempty"`
	Config       string          `yaml:"config,omitempty"`
	Init         string          `yaml:"init,omitempty"`
	Build        string          `yaml:"build,omitempty"`
	Priority     int             `yaml:"priority,omitempty"`
	Keymaps      []NvimKeymapDef `yaml:"keymaps,omitempty"`
	Enabled      bool            `yaml:"enabled,omitempty"`
}

// NvimKeyDef represents a lazy loading key trigger
type NvimKeyDef struct {
	Key  string `yaml:"key"`
	Mode string `yaml:"mode"`
	Desc string `yaml:"desc"`
}

// NvimKeymapDef represents a custom keymap
type NvimKeymapDef struct {
	Key     string `yaml:"key"`
	Mode    string `yaml:"mode"`
	Command string `yaml:"command"`
	Desc    string `yaml:"desc"`
}

// PluginManager handles plugin definitions
type PluginManager struct {
	plugins []NvimPluginDef
}

// NewPluginManager creates a new plugin manager
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make([]NvimPluginDef, 0),
	}
}

// LoadPluginsFromDirectory loads all plugin YAML files from a directory
func (pm *PluginManager) LoadPluginsFromDirectory(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read plugins directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}

		// Skip schema file
		if strings.Contains(file.Name(), "schema") {
			continue
		}

		pluginPath := filepath.Join(dir, file.Name())
		plugin, err := pm.LoadPlugin(pluginPath)
		if err != nil {
			return fmt.Errorf("failed to load plugin %s: %w", file.Name(), err)
		}

		pm.plugins = append(pm.plugins, *plugin)
	}

	return nil
}

// LoadPlugin loads a single plugin from a YAML file
func (pm *PluginManager) LoadPlugin(path string) (*NvimPluginDef, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var plugin NvimPluginDef
	if err := yaml.Unmarshal(data, &plugin); err != nil {
		return nil, fmt.Errorf("failed to parse plugin YAML: %w", err)
	}

	// Default enabled to true if not specified
	if plugin.Enabled == false && !strings.Contains(string(data), "enabled:") {
		plugin.Enabled = true
	}

	return &plugin, nil
}

// GenerateLuaPlugin generates a Lua file for the plugin
func (pm *PluginManager) GenerateLuaPlugin(plugin NvimPluginDef) string {
	var lua strings.Builder

	lua.WriteString("return {\n")
	lua.WriteString(fmt.Sprintf("  \"%s\",\n", plugin.Repo))

	// Dependencies
	if len(plugin.Dependencies) > 0 {
		lua.WriteString("  dependencies = {\n")
		for _, dep := range plugin.Dependencies {
			lua.WriteString(fmt.Sprintf("    \"%s\",\n", dep))
		}
		lua.WriteString("  },\n")
	}

	// Filetypes
	if len(plugin.Filetypes) > 0 {
		lua.WriteString("  ft = {")
		for i, ft := range plugin.Filetypes {
			if i > 0 {
				lua.WriteString(", ")
			}
			lua.WriteString(fmt.Sprintf("\"%s\"", ft))
		}
		lua.WriteString("},\n")
	}

	// Lazy loading
	if plugin.Lazy {
		lua.WriteString("  lazy = true,\n")
	}

	// Event
	if plugin.Event != "" {
		lua.WriteString(fmt.Sprintf("  event = \"%s\",\n", plugin.Event))
	}

	// Keys
	if len(plugin.Keys) > 0 {
		lua.WriteString("  keys = {\n")
		for _, key := range plugin.Keys {
			lua.WriteString("    {")
			lua.WriteString(fmt.Sprintf("\"%s\"", key.Key))
			if key.Mode != "" {
				lua.WriteString(fmt.Sprintf(", mode = \"%s\"", key.Mode))
			}
			if key.Desc != "" {
				lua.WriteString(fmt.Sprintf(", desc = \"%s\"", key.Desc))
			}
			lua.WriteString("},\n")
		}
		lua.WriteString("  },\n")
	}

	// Commands
	if len(plugin.Cmd) > 0 {
		lua.WriteString("  cmd = {")
		for i, cmd := range plugin.Cmd {
			if i > 0 {
				lua.WriteString(", ")
			}
			lua.WriteString(fmt.Sprintf("\"%s\"", cmd))
		}
		lua.WriteString("},\n")
	}

	// Build
	if plugin.Build != "" {
		lua.WriteString(fmt.Sprintf("  build = \"%s\",\n", plugin.Build))
	}

	// Priority
	if plugin.Priority > 0 {
		lua.WriteString(fmt.Sprintf("  priority = %d,\n", plugin.Priority))
	}

	// Init function
	if plugin.Init != "" {
		lua.WriteString("  init = function()\n")
		// Indent the init code
		initLines := strings.Split(plugin.Init, "\n")
		for _, line := range initLines {
			if strings.TrimSpace(line) != "" {
				lua.WriteString(fmt.Sprintf("    %s\n", line))
			}
		}
		lua.WriteString("  end,\n")
	}

	// Config function
	if plugin.Config != "" {
		lua.WriteString("  config = function()\n")
		// Indent the config code
		configLines := strings.Split(plugin.Config, "\n")
		for _, line := range configLines {
			if strings.TrimSpace(line) != "" {
				lua.WriteString(fmt.Sprintf("    %s\n", line))
			}
		}

		// Add keymaps if defined
		if len(plugin.Keymaps) > 0 {
			lua.WriteString("\n    -- Plugin keymaps\n")
			lua.WriteString("    local keymap = vim.keymap\n")
			for _, km := range plugin.Keymaps {
				mode := km.Mode
				if mode == "" {
					mode = "n"
				}
				lua.WriteString(fmt.Sprintf("    keymap.set(\"%s\", \"%s\", \"%s\", { desc = \"%s\" })\n",
					mode, km.Key, km.Command, km.Desc))
			}
		}

		lua.WriteString("  end,\n")
	}

	lua.WriteString("}\n")

	return lua.String()
}

// ExportPluginsToPath exports all loaded plugins as Lua files
func (pm *PluginManager) ExportPluginsToPath(basePath string) error {
	pluginsDir := filepath.Join(basePath, "lua", "workspace", "plugins", "custom")
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		return fmt.Errorf("failed to create custom plugins directory: %w", err)
	}

	for _, plugin := range pm.plugins {
		if !plugin.Enabled {
			continue
		}

		// Generate filename from plugin name
		filename := strings.ToLower(strings.ReplaceAll(plugin.Name, " ", "-"))
		if !strings.HasSuffix(filename, ".lua") {
			filename += ".lua"
		}

		luaContent := pm.GenerateLuaPlugin(plugin)
		luaPath := filepath.Join(pluginsDir, filename)

		if err := os.WriteFile(luaPath, []byte(luaContent), 0644); err != nil {
			return fmt.Errorf("failed to write plugin %s: %w", filename, err)
		}
	}

	return nil
}

// GetEnabledPlugins returns a list of enabled plugin names
func (pm *PluginManager) GetEnabledPlugins() []string {
	names := make([]string, 0, len(pm.plugins))
	for _, p := range pm.plugins {
		if p.Enabled {
			names = append(names, p.Name)
		}
	}
	return names
}
