package templates

import (
	"devopsmaestro/db"
	"devopsmaestro/models"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DBPluginManager handles loading plugins from database and generating Lua files
type DBPluginManager struct {
	datastore db.DataStore
}

// NewDBPluginManager creates a new manager
func NewDBPluginManager(datastore db.DataStore) *DBPluginManager {
	return &DBPluginManager{
		datastore: datastore,
	}
}

// GenerateLuaFilesForWorkspace loads plugins from DB and generates Lua files
func (pm *DBPluginManager) GenerateLuaFilesForWorkspace(pluginNames []string, outputDir string) error {
	if len(pluginNames) == 0 {
		return nil
	}

	// Create custom plugins directory
	customDir := filepath.Join(outputDir, "lua", "workspace", "plugins", "custom")
	if err := os.MkdirAll(customDir, 0755); err != nil {
		return fmt.Errorf("failed to create custom plugins dir: %v", err)
	}

	// Load each plugin from DB and generate Lua
	for _, name := range pluginNames {
		plugin, err := pm.datastore.GetPluginByName(name)
		if err != nil {
			return fmt.Errorf("failed to load plugin '%s': %v", name, err)
		}

		luaContent, err := pm.generateLuaForPlugin(plugin)
		if err != nil {
			return fmt.Errorf("failed to generate Lua for plugin '%s': %v", name, err)
		}

		// Write Lua file
		filename := filepath.Join(customDir, fmt.Sprintf("%s.lua", name))
		if err := os.WriteFile(filename, []byte(luaContent), 0644); err != nil {
			return fmt.Errorf("failed to write plugin file '%s': %v", filename, err)
		}

		fmt.Printf("  Generated plugin: %s\n", name)
	}

	return nil
}

// generateLuaForPlugin converts a database plugin to Lua code
func (pm *DBPluginManager) generateLuaForPlugin(plugin *models.NvimPluginDB) (string, error) {
	var lua strings.Builder

	// Start the plugin spec
	lua.WriteString("return {\n")
	lua.WriteString(fmt.Sprintf("  \"%s\",\n", plugin.Repo))

	// Branch
	if plugin.Branch.Valid {
		lua.WriteString(fmt.Sprintf("  branch = \"%s\",\n", plugin.Branch.String))
	}

	// Version
	if plugin.Version.Valid {
		lua.WriteString(fmt.Sprintf("  version = \"%s\",\n", plugin.Version.String))
	}

	// Priority
	if plugin.Priority.Valid {
		lua.WriteString(fmt.Sprintf("  priority = %d,\n", plugin.Priority.Int64))
	}

	// Event (lazy loading)
	if plugin.Event.Valid {
		var event interface{}
		if err := json.Unmarshal([]byte(plugin.Event.String), &event); err == nil {
			switch v := event.(type) {
			case string:
				lua.WriteString(fmt.Sprintf("  event = \"%s\",\n", v))
			case []interface{}:
				lua.WriteString("  event = {")
				for i, e := range v {
					if i > 0 {
						lua.WriteString(", ")
					}
					if eStr, ok := e.(string); ok {
						lua.WriteString(fmt.Sprintf("\"%s\"", eStr))
					}
				}
				lua.WriteString("},\n")
			}
		}
	}

	// Filetype
	if plugin.Ft.Valid {
		var ft interface{}
		if err := json.Unmarshal([]byte(plugin.Ft.String), &ft); err == nil {
			switch v := ft.(type) {
			case string:
				lua.WriteString(fmt.Sprintf("  ft = \"%s\",\n", v))
			case []interface{}:
				lua.WriteString("  ft = {")
				for i, f := range v {
					if i > 0 {
						lua.WriteString(", ")
					}
					if fStr, ok := f.(string); ok {
						lua.WriteString(fmt.Sprintf("\"%s\"", fStr))
					}
				}
				lua.WriteString("},\n")
			}
		}
	}

	// Command
	if plugin.Cmd.Valid {
		var cmd interface{}
		if err := json.Unmarshal([]byte(plugin.Cmd.String), &cmd); err == nil {
			switch v := cmd.(type) {
			case string:
				lua.WriteString(fmt.Sprintf("  cmd = \"%s\",\n", v))
			case []interface{}:
				lua.WriteString("  cmd = {")
				for i, c := range v {
					if i > 0 {
						lua.WriteString(", ")
					}
					if cStr, ok := c.(string); ok {
						lua.WriteString(fmt.Sprintf("\"%s\"", cStr))
					}
				}
				lua.WriteString("},\n")
			}
		}
	}

	// Dependencies
	if plugin.Dependencies.Valid {
		var deps []interface{}
		if err := json.Unmarshal([]byte(plugin.Dependencies.String), &deps); err == nil && len(deps) > 0 {
			lua.WriteString("  dependencies = {\n")
			for _, dep := range deps {
				switch d := dep.(type) {
				case string:
					lua.WriteString(fmt.Sprintf("    \"%s\",\n", d))
				case map[string]interface{}:
					// Complex dependency with build, etc.
					if repo, ok := d["repo"].(string); ok {
						lua.WriteString(fmt.Sprintf("    { \"%s\"", repo))
						if build, ok := d["build"].(string); ok {
							lua.WriteString(fmt.Sprintf(", build = \"%s\"", build))
						}
						if version, ok := d["version"].(string); ok {
							lua.WriteString(fmt.Sprintf(", version = \"%s\"", version))
						}
						if branch, ok := d["branch"].(string); ok {
							lua.WriteString(fmt.Sprintf(", branch = \"%s\"", branch))
						}
						if config, ok := d["config"].(bool); ok && config {
							lua.WriteString(", config = true")
						}
						lua.WriteString(" },\n")
					}
				}
			}
			lua.WriteString("  },\n")
		}
	}

	// Keys (lazy loading)
	if plugin.Keys.Valid {
		var keys []interface{}
		if err := json.Unmarshal([]byte(plugin.Keys.String), &keys); err == nil && len(keys) > 0 {
			lua.WriteString("  keys = {\n")
			for _, k := range keys {
				if keyMap, ok := k.(map[string]interface{}); ok {
					// Try lowercase "key" first, then uppercase "Key" (from JSON Marshal)
					var keyStr string
					if keyVal, ok := keyMap["key"]; ok && keyVal != nil {
						keyStr = keyVal.(string)
					} else if keyVal, ok := keyMap["Key"]; ok && keyVal != nil {
						keyStr = keyVal.(string)
					} else {
						continue // Skip if no key field
					}

					lua.WriteString(fmt.Sprintf("    { \"%s\"", keyStr))

					// Try to get action
					if action, ok := keyMap["action"].(string); ok {
						lua.WriteString(fmt.Sprintf(", \"%s\"", action))
					} else if action, ok := keyMap["Action"].(string); ok {
						lua.WriteString(fmt.Sprintf(", \"%s\"", action))
					}

					// Try to get desc
					if desc, ok := keyMap["desc"].(string); ok {
						lua.WriteString(fmt.Sprintf(", desc = \"%s\"", desc))
					} else if desc, ok := keyMap["Desc"].(string); ok {
						lua.WriteString(fmt.Sprintf(", desc = \"%s\"", desc))
					}

					// Try to get mode
					var mode interface{}
					if m, ok := keyMap["mode"]; ok && m != nil {
						mode = m
					} else if m, ok := keyMap["Mode"]; ok && m != nil {
						mode = m
					}

					if mode != nil {
						switch m := mode.(type) {
						case string:
							lua.WriteString(fmt.Sprintf(", mode = \"%s\"", m))
						case []interface{}:
							lua.WriteString(", mode = {")
							for i, md := range m {
								if i > 0 {
									lua.WriteString(", ")
								}
								if mdStr, ok := md.(string); ok {
									lua.WriteString(fmt.Sprintf("\"%s\"", mdStr))
								}
							}
							lua.WriteString("}")
						}
					}
					lua.WriteString(" },\n")
				}
			}
			lua.WriteString("  },\n")
		}
	}

	// Build
	if plugin.Build.Valid {
		lua.WriteString(fmt.Sprintf("  build = \"%s\",\n", plugin.Build.String))
	}

	// Init function
	if plugin.Init.Valid && plugin.Init.String != "" {
		lua.WriteString("  init = function()\n")
		// Indent each line
		for _, line := range strings.Split(plugin.Init.String, "\n") {
			if strings.TrimSpace(line) != "" {
				lua.WriteString("    " + line + "\n")
			}
		}
		lua.WriteString("  end,\n")
	}

	// Config function
	if plugin.Config.Valid && plugin.Config.String != "" {
		lua.WriteString("  config = function()\n")
		// Indent each line
		for _, line := range strings.Split(plugin.Config.String, "\n") {
			if strings.TrimSpace(line) != "" {
				lua.WriteString("    " + line + "\n")
			}
		}
		lua.WriteString("  end,\n")
	}

	// Opts
	if plugin.Opts.Valid && plugin.Opts.String != "" {
		lua.WriteString(fmt.Sprintf("  opts = %s,\n", plugin.Opts.String))
	}

	lua.WriteString("}\n")

	return lua.String(), nil
}
