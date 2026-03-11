// Model conversion functions for bridging plugin.Plugin and models.NvimPluginDB.
package store

import (
	"database/sql"
	"encoding/json"
	"strings"

	"devopsmaestro/models"
	"devopsmaestro/pkg/nvimops/plugin"
)

// pluginToDBModel converts a plugin.Plugin to models.NvimPluginDB.
func pluginToDBModel(p *plugin.Plugin) *models.NvimPluginDB {
	db := &models.NvimPluginDB{
		Name:    p.Name,
		Repo:    p.Repo,
		Lazy:    p.Lazy,
		Enabled: p.Enabled,
	}

	// String fields with sql.NullString
	if p.Description != "" {
		db.Description = sql.NullString{String: p.Description, Valid: true}
	}
	if p.Branch != "" {
		db.Branch = sql.NullString{String: p.Branch, Valid: true}
	}
	if p.Version != "" {
		db.Version = sql.NullString{String: p.Version, Valid: true}
	}
	if p.Category != "" {
		db.Category = sql.NullString{String: p.Category, Valid: true}
	}
	if p.Build != "" {
		db.Build = sql.NullString{String: p.Build, Valid: true}
	}
	if p.Config != "" {
		db.Config = sql.NullString{String: p.Config, Valid: true}
	}
	if p.Init != "" {
		db.Init = sql.NullString{String: p.Init, Valid: true}
	}

	// Priority
	if p.Priority != 0 {
		db.Priority = sql.NullInt64{Int64: int64(p.Priority), Valid: true}
	}

	// JSON array fields
	if len(p.Event) > 0 {
		if data, err := json.Marshal(p.Event); err == nil {
			db.Event = sql.NullString{String: string(data), Valid: true}
		}
	}
	if len(p.Ft) > 0 {
		if data, err := json.Marshal(p.Ft); err == nil {
			db.Ft = sql.NullString{String: string(data), Valid: true}
		}
	}
	if len(p.Cmd) > 0 {
		if data, err := json.Marshal(p.Cmd); err == nil {
			db.Cmd = sql.NullString{String: string(data), Valid: true}
		}
	}
	if len(p.Tags) > 0 {
		if data, err := json.Marshal(p.Tags); err == nil {
			db.Tags = sql.NullString{String: string(data), Valid: true}
		}
	}

	// Keys (convert to compatible format)
	if len(p.Keys) > 0 {
		keysData := make([]models.PluginKeymap, len(p.Keys))
		for i, k := range p.Keys {
			keysData[i] = models.PluginKeymap{
				Key:    k.Key,
				Mode:   modeSliceToInterface(k.Mode),
				Action: k.Action,
				Desc:   k.Desc,
			}
		}
		if data, err := json.Marshal(keysData); err == nil {
			db.Keys = sql.NullString{String: string(data), Valid: true}
		}
	}

	// Keymaps
	if len(p.Keymaps) > 0 {
		keymapsData := make([]models.PluginKeymap, len(p.Keymaps))
		for i, k := range p.Keymaps {
			keymapsData[i] = models.PluginKeymap{
				Key:    k.Key,
				Mode:   modeSliceToInterface(k.Mode),
				Action: k.Action,
				Desc:   k.Desc,
			}
		}
		if data, err := json.Marshal(keymapsData); err == nil {
			db.Keymaps = sql.NullString{String: string(data), Valid: true}
		}
	}

	// Dependencies
	if len(p.Dependencies) > 0 {
		depsData := make([]interface{}, len(p.Dependencies))
		for i, d := range p.Dependencies {
			// Use full struct if has extra fields, otherwise just repo string
			if d.Build != "" || d.Version != "" || d.Branch != "" || d.Config {
				depsData[i] = models.PluginDependency{
					Repo:    d.Repo,
					Build:   d.Build,
					Version: d.Version,
					Branch:  d.Branch,
					Config:  d.Config,
				}
			} else {
				depsData[i] = d.Repo
			}
		}
		if data, err := json.Marshal(depsData); err == nil {
			db.Dependencies = sql.NullString{String: string(data), Valid: true}
		}
	}

	// Opts (map or string)
	if p.Opts != nil {
		// Handle both map and string types
		var shouldStore bool
		switch v := p.Opts.(type) {
		case map[string]interface{}:
			shouldStore = len(v) > 0
		case string:
			shouldStore = strings.TrimSpace(v) != ""
		default:
			shouldStore = true // For any other type, try to store it
		}

		if shouldStore {
			if data, err := json.Marshal(p.Opts); err == nil {
				db.Opts = sql.NullString{String: string(data), Valid: true}
			}
		}
	}

	// Timestamps
	if p.CreatedAt != nil {
		db.CreatedAt = *p.CreatedAt
	}
	if p.UpdatedAt != nil {
		db.UpdatedAt = *p.UpdatedAt
	}

	return db
}

// dbModelToPlugin converts a models.NvimPluginDB to plugin.Plugin.
func dbModelToPlugin(db *models.NvimPluginDB) *plugin.Plugin {
	p := &plugin.Plugin{
		Name:    db.Name,
		Repo:    db.Repo,
		Lazy:    db.Lazy,
		Enabled: db.Enabled,
	}

	// String fields
	if db.Description.Valid {
		p.Description = db.Description.String
	}
	if db.Branch.Valid {
		p.Branch = db.Branch.String
	}
	if db.Version.Valid {
		p.Version = db.Version.String
	}
	if db.Category.Valid {
		p.Category = db.Category.String
	}
	if db.Build.Valid {
		p.Build = db.Build.String
	}
	if db.Config.Valid {
		p.Config = db.Config.String
	}
	if db.Init.Valid {
		p.Init = db.Init.String
	}

	// Priority
	if db.Priority.Valid {
		p.Priority = int(db.Priority.Int64)
	}

	// JSON array fields
	if db.Event.Valid {
		var event []string
		if err := json.Unmarshal([]byte(db.Event.String), &event); err == nil {
			p.Event = event
		}
	}
	if db.Ft.Valid {
		var ft []string
		if err := json.Unmarshal([]byte(db.Ft.String), &ft); err == nil {
			p.Ft = ft
		}
	}
	if db.Cmd.Valid {
		var cmd []string
		if err := json.Unmarshal([]byte(db.Cmd.String), &cmd); err == nil {
			p.Cmd = cmd
		}
	}
	if db.Tags.Valid {
		var tags []string
		if err := json.Unmarshal([]byte(db.Tags.String), &tags); err == nil {
			p.Tags = tags
		}
	}

	// Keys
	if db.Keys.Valid {
		var keysData []models.PluginKeymap
		if err := json.Unmarshal([]byte(db.Keys.String), &keysData); err == nil {
			p.Keys = make([]plugin.Keymap, len(keysData))
			for i, k := range keysData {
				p.Keys[i] = plugin.Keymap{
					Key:    k.Key,
					Mode:   interfaceToModeSlice(k.Mode),
					Action: k.Action,
					Desc:   k.Desc,
				}
			}
		}
	}

	// Keymaps
	if db.Keymaps.Valid {
		var keymapsData []models.PluginKeymap
		if err := json.Unmarshal([]byte(db.Keymaps.String), &keymapsData); err == nil {
			p.Keymaps = make([]plugin.Keymap, len(keymapsData))
			for i, k := range keymapsData {
				p.Keymaps[i] = plugin.Keymap{
					Key:    k.Key,
					Mode:   interfaceToModeSlice(k.Mode),
					Action: k.Action,
					Desc:   k.Desc,
				}
			}
		}
	}

	// Dependencies
	if db.Dependencies.Valid {
		var depsRaw []interface{}
		if err := json.Unmarshal([]byte(db.Dependencies.String), &depsRaw); err == nil {
			for _, dep := range depsRaw {
				switch d := dep.(type) {
				case string:
					p.Dependencies = append(p.Dependencies, plugin.Dependency{Repo: d})
				case map[string]interface{}:
					depObj := plugin.Dependency{}
					if repo, ok := d["repo"].(string); ok {
						depObj.Repo = repo
					}
					if build, ok := d["build"].(string); ok {
						depObj.Build = build
					}
					if version, ok := d["version"].(string); ok {
						depObj.Version = version
					}
					if branch, ok := d["branch"].(string); ok {
						depObj.Branch = branch
					}
					if config, ok := d["config"].(bool); ok {
						depObj.Config = config
					}
					p.Dependencies = append(p.Dependencies, depObj)
				}
			}
		}
	}

	// Opts
	if db.Opts.Valid {
		var opts map[string]interface{}
		if err := json.Unmarshal([]byte(db.Opts.String), &opts); err == nil {
			p.Opts = opts
		}
	}

	// Timestamps
	if !db.CreatedAt.IsZero() {
		p.CreatedAt = &db.CreatedAt
	}
	if !db.UpdatedAt.IsZero() {
		p.UpdatedAt = &db.UpdatedAt
	}

	return p
}

// modeSliceToInterface converts []string mode to interface{} for JSON storage.
func modeSliceToInterface(modes []string) interface{} {
	if len(modes) == 0 {
		return nil
	}
	if len(modes) == 1 {
		return modes[0]
	}
	return modes
}

// interfaceToModeSlice converts interface{} mode from JSON to []string.
func interfaceToModeSlice(mode interface{}) []string {
	if mode == nil {
		return nil
	}
	switch m := mode.(type) {
	case string:
		return []string{m}
	case []interface{}:
		result := make([]string, len(m))
		for i, v := range m {
			if s, ok := v.(string); ok {
				result[i] = s
			}
		}
		return result
	case []string:
		return m
	}
	return nil
}
