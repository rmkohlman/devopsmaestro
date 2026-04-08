// Package themebridge provides inheritance resolution for Neovim themes.
package themebridge

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"devopsmaestro/models"
)

// MaxInheritanceDepth is the maximum depth of the inheritance chain
// to prevent infinite loops.
const MaxInheritanceDepth = 5

// ResolveWithInheritance resolves a theme by walking its inheritance chain.
// The returned ThemeSpec has all inherited values merged: child values override
// parent values, and unset values inherit from the parent.
// Resolution is lazy — only called when the theme is actually used.
func ResolveWithInheritance(themeName string, store ThemeDataStore) (*models.NvimThemeDB, error) {
	visited := make(map[string]bool)
	return resolveChain(themeName, store, visited, 0)
}

// resolveChain recursively walks the inheritance chain.
func resolveChain(themeName string, store ThemeDataStore, visited map[string]bool, depth int) (*models.NvimThemeDB, error) {
	if depth >= MaxInheritanceDepth {
		return nil, fmt.Errorf("theme inheritance depth exceeded maximum of %d", MaxInheritanceDepth)
	}

	if visited[themeName] {
		return nil, fmt.Errorf("circular theme inheritance detected: %q already visited", themeName)
	}
	visited[themeName] = true

	theme, err := store.GetThemeByName(themeName)
	if err != nil {
		return nil, fmt.Errorf("failed to get theme %q: %w", themeName, err)
	}

	// No inheritance — return as-is.
	if !theme.Inherits.Valid || theme.Inherits.String == "" {
		return theme, nil
	}

	// Resolve parent recursively.
	parent, err := resolveChain(theme.Inherits.String, store, visited, depth+1)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve parent theme %q: %w", theme.Inherits.String, err)
	}

	// Merge: child overrides parent.
	return mergeThemes(theme, parent)
}

// mergeThemes merges a child theme on top of a parent theme.
// Child values take precedence; unset child values inherit from parent.
// The returned theme has the child's identity (name, ID, etc.) but merged content.
func mergeThemes(child, parent *models.NvimThemeDB) (*models.NvimThemeDB, error) {
	result := *child // Copy child as base

	// String fields: inherit from parent if child is unset
	if !result.Description.Valid && parent.Description.Valid {
		result.Description = parent.Description
	}
	if !result.Author.Valid && parent.Author.Valid {
		result.Author = parent.Author
	}
	if !result.Category.Valid && parent.Category.Valid {
		result.Category = parent.Category
	}
	if result.PluginRepo == "" && parent.PluginRepo != "" {
		result.PluginRepo = parent.PluginRepo
	}
	if !result.PluginBranch.Valid && parent.PluginBranch.Valid {
		result.PluginBranch = parent.PluginBranch
	}
	if !result.PluginTag.Valid && parent.PluginTag.Valid {
		result.PluginTag = parent.PluginTag
	}
	if !result.Style.Valid && parent.Style.Valid {
		result.Style = parent.Style
	}

	// Colors: deep merge (child colors override parent colors)
	mergedColors, err := mergeJSONMaps(child.Colors, parent.Colors)
	if err != nil {
		return nil, fmt.Errorf("failed to merge colors: %w", err)
	}
	result.Colors = mergedColors

	// Options: deep merge (child options override parent options)
	mergedOptions, err := mergeJSONMaps(child.Options, parent.Options)
	if err != nil {
		return nil, fmt.Errorf("failed to merge options: %w", err)
	}
	result.Options = mergedOptions

	// CustomHighlights: deep merge by group name
	mergedHL, err := mergeJSONMaps(child.CustomHighlights, parent.CustomHighlights)
	if err != nil {
		return nil, fmt.Errorf("failed to merge custom highlights: %w", err)
	}
	result.CustomHighlights = mergedHL

	return &result, nil
}

// mergeJSONMaps deep-merges two sql.NullString values containing JSON objects.
// Child entries override parent entries. If only one side is valid, it is
// returned as-is. If neither is valid, an empty NullString is returned.
func mergeJSONMaps(child, parent sql.NullString) (sql.NullString, error) {
	if !child.Valid && !parent.Valid {
		return sql.NullString{}, nil
	}
	if !child.Valid {
		return parent, nil
	}
	if !parent.Valid {
		return child, nil
	}

	// Both are valid — deep merge the JSON maps.
	var childMap, parentMap map[string]json.RawMessage
	if err := json.Unmarshal([]byte(parent.String), &parentMap); err != nil {
		return sql.NullString{}, fmt.Errorf("failed to unmarshal parent JSON: %w", err)
	}
	if err := json.Unmarshal([]byte(child.String), &childMap); err != nil {
		return sql.NullString{}, fmt.Errorf("failed to unmarshal child JSON: %w", err)
	}

	// Start with parent, overlay child entries.
	merged := make(map[string]json.RawMessage, len(parentMap)+len(childMap))
	for k, v := range parentMap {
		merged[k] = v
	}
	for k, v := range childMap {
		merged[k] = v
	}

	data, err := json.Marshal(merged)
	if err != nil {
		return sql.NullString{}, fmt.Errorf("failed to marshal merged JSON: %w", err)
	}
	return sql.NullString{String: string(data), Valid: true}, nil
}
