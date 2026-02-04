// Package cmd provides CLI commands for DevOpsMaestro.
// This file contains the WorkspacePluginManager interface and implementation
// for managing nvim plugins on a per-workspace basis.
package cmd

import (
	"database/sql"
	"strings"

	"devopsmaestro/models"
)

// =============================================================================
// Workspace Plugin Manager Interface and Implementation
// =============================================================================

// WorkspacePluginManager handles workspace plugin operations.
// This interface enables testing with mocks.
type WorkspacePluginManager interface {
	// ListPlugins returns the plugins configured for a workspace.
	ListPlugins(workspace *models.Workspace) []string

	// AddPlugins adds plugins to a workspace, returning added, skipped, and not-found lists.
	// globalPlugins is the list of valid plugin names to validate against.
	AddPlugins(workspace *models.Workspace, plugins []string, globalPlugins []string) (added, skipped, notFound []string)

	// RemovePlugins removes plugins from a workspace, returning removed and not-found lists.
	RemovePlugins(workspace *models.Workspace, plugins []string) (removed, notFound []string)

	// ClearPlugins removes all plugins from a workspace, returning the count cleared.
	ClearPlugins(workspace *models.Workspace) int
}

// DefaultWorkspacePluginManager is the default implementation of WorkspacePluginManager.
type DefaultWorkspacePluginManager struct{}

// NewWorkspacePluginManager creates a new DefaultWorkspacePluginManager.
func NewWorkspacePluginManager() (*DefaultWorkspacePluginManager, error) {
	return &DefaultWorkspacePluginManager{}, nil
}

// ListPlugins returns the plugins configured for a workspace.
func (m *DefaultWorkspacePluginManager) ListPlugins(workspace *models.Workspace) []string {
	if !workspace.NvimPlugins.Valid || workspace.NvimPlugins.String == "" {
		return nil
	}
	return strings.Split(workspace.NvimPlugins.String, ",")
}

// AddPlugins adds plugins to a workspace configuration.
// Returns lists of added, skipped (already present), and not-found (not in global) plugins.
func (m *DefaultWorkspacePluginManager) AddPlugins(workspace *models.Workspace, plugins []string, globalPlugins []string) (added, skipped, notFound []string) {
	// Build set of global plugin names for validation
	globalSet := make(map[string]bool)
	for _, name := range globalPlugins {
		globalSet[name] = true
	}

	// Get current workspace plugins
	currentPlugins := m.ListPlugins(workspace)
	currentSet := make(map[string]bool)
	for _, p := range currentPlugins {
		currentSet[p] = true
	}

	// Process each plugin
	for _, name := range plugins {
		if !globalSet[name] {
			notFound = append(notFound, name)
			continue
		}
		if currentSet[name] {
			skipped = append(skipped, name)
			continue
		}
		currentPlugins = append(currentPlugins, name)
		currentSet[name] = true
		added = append(added, name)
	}

	// Update workspace
	workspace.NvimPlugins = sql.NullString{
		String: strings.Join(currentPlugins, ","),
		Valid:  len(currentPlugins) > 0,
	}

	// Enable nvim config generation if not already set
	if !workspace.NvimStructure.Valid || workspace.NvimStructure.String == "" {
		workspace.NvimStructure = sql.NullString{
			String: "custom",
			Valid:  true,
		}
	}

	return added, skipped, notFound
}

// RemovePlugins removes plugins from a workspace configuration.
// Returns lists of removed and not-found plugins.
func (m *DefaultWorkspacePluginManager) RemovePlugins(workspace *models.Workspace, plugins []string) (removed, notFound []string) {
	currentPlugins := m.ListPlugins(workspace)
	if len(currentPlugins) == 0 {
		return nil, plugins
	}

	// Build removal set
	removeSet := make(map[string]bool)
	for _, name := range plugins {
		removeSet[name] = true
	}

	// Filter out removed plugins
	var remaining []string
	for _, p := range currentPlugins {
		if removeSet[p] {
			removed = append(removed, p)
			delete(removeSet, p)
		} else {
			remaining = append(remaining, p)
		}
	}

	// Remaining in removeSet are not found
	for name := range removeSet {
		notFound = append(notFound, name)
	}

	// Update workspace
	workspace.NvimPlugins = sql.NullString{
		String: strings.Join(remaining, ","),
		Valid:  len(remaining) > 0,
	}

	return removed, notFound
}

// ClearPlugins removes all plugins from a workspace, returning the count cleared.
func (m *DefaultWorkspacePluginManager) ClearPlugins(workspace *models.Workspace) int {
	plugins := m.ListPlugins(workspace)
	count := len(plugins)

	workspace.NvimPlugins = sql.NullString{
		String: "",
		Valid:  false,
	}

	return count
}
