package sync

import (
	"context"
	"fmt"
	"strings"
)

// BuiltinSources contains information about built-in sources that can be registered.
// These are sources that we know about but may not be implemented yet.
var BuiltinSources = []SourceInfo{
	{
		Name:        "lazyvim",
		Description: "LazyVim - A Neovim config for lazy people",
		URL:         "https://github.com/LazyVim/LazyVim",
		Type:        string(SourceTypeGitHub),
		ConfigKeys:  []string{"repo_url", "branch", "plugins_dir"},
	},
	{
		Name:        "astronvim",
		Description: "AstroNvim - An aesthetic and feature-rich neovim config",
		URL:         "https://github.com/AstroNvim/AstroNvim",
		Type:        string(SourceTypeGitHub),
		ConfigKeys:  []string{"repo_url", "branch", "plugins_dir"},
	},
	{
		Name:        "nvchad",
		Description: "NvChad - Blazing fast Neovim config",
		URL:         "https://github.com/NvChad/NvChad",
		Type:        string(SourceTypeGitHub),
		ConfigKeys:  []string{"repo_url", "branch", "plugins_dir"},
	},
	{
		Name:        "kickstart",
		Description: "Kickstart.nvim - A starting point for Neovim configuration",
		URL:         "https://github.com/nvim-lua/kickstart.nvim",
		Type:        string(SourceTypeGitHub),
		ConfigKeys:  []string{"repo_url", "branch"},
	},
	{
		Name:        "lunarvim",
		Description: "LunarVim - IDE layer for Neovim",
		URL:         "https://github.com/LunarVim/LunarVim",
		Type:        string(SourceTypeGitHub),
		ConfigKeys:  []string{"repo_url", "branch", "config_dir"},
	},
	{
		Name:        "local",
		Description: "Local filesystem plugins directory",
		URL:         "file://",
		Type:        string(SourceTypeLocal),
		ConfigKeys:  []string{"plugins_dir", "recursive"},
	},
}

// RegisterBuiltinSources registers all builtin source info in the registry.
// Note: This only registers the metadata, not the actual handlers.
// Handlers must be registered separately when their implementations are available.
func RegisterBuiltinSources(registry *SourceRegistry) error {
	for _, info := range BuiltinSources {
		// Create a placeholder registration that will error if used
		registration := HandlerRegistration{
			Name: info.Name,
			Info: info,
			CreateFunc: func() SourceHandler {
				return &NotImplementedHandler{info: info}
			},
		}

		// Only register if not already registered
		if !registry.IsRegistered(info.Name) {
			if err := registry.Register(registration); err != nil {
				return fmt.Errorf("failed to register builtin source %s: %w", info.Name, err)
			}
		}
	}
	return nil
}

// IsRegistered checks if a source is registered in the registry.
func (r *SourceRegistry) IsRegistered(sourceName string) bool {
	_, exists := r.GetRegistration(sourceName)
	return exists
}

// GetSourceInfo returns just the SourceInfo for a registered source.
func (r *SourceRegistry) GetSourceInfo(sourceName string) (*SourceInfo, error) {
	registration, exists := r.GetRegistration(sourceName)
	if !exists {
		return nil, &ErrSourceNotFound{Source: sourceName}
	}
	return &registration.Info, nil
}

// ListSourcesByType returns sources filtered by type.
func (r *SourceRegistry) ListSourcesByType(sourceType SourceType) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var sources []string
	for name, registration := range r.registrations {
		if registration.Info.Type == string(sourceType) {
			sources = append(sources, name)
		}
	}
	return sources
}

// SearchSources returns sources that match the search query (case-insensitive).
// Searches in name, description, and URL.
func (r *SourceRegistry) SearchSources(query string) []HandlerRegistration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query = strings.ToLower(query)
	var matches []HandlerRegistration

	for _, registration := range r.registrations {
		if r.matchesQuery(registration, query) {
			matches = append(matches, registration)
		}
	}
	return matches
}

func (r *SourceRegistry) matchesQuery(registration HandlerRegistration, query string) bool {
	return strings.Contains(strings.ToLower(registration.Name), query) ||
		strings.Contains(strings.ToLower(registration.Info.Description), query) ||
		strings.Contains(strings.ToLower(registration.Info.URL), query)
}

// NotImplementedHandler is a placeholder handler for sources that aren't implemented yet.
// It provides useful error messages and allows the registry to contain all known sources.
type NotImplementedHandler struct {
	info SourceInfo
}

func (h *NotImplementedHandler) Name() string {
	return h.info.Name
}

func (h *NotImplementedHandler) Description() string {
	return h.info.Description
}

func (h *NotImplementedHandler) Sync(ctx context.Context, options SyncOptions) (*SyncResult, error) {
	return nil, fmt.Errorf("source '%s' is not yet implemented - handler development needed", h.info.Name)
}

func (h *NotImplementedHandler) ListAvailable(ctx context.Context) ([]AvailablePlugin, error) {
	return nil, fmt.Errorf("source '%s' is not yet implemented - handler development needed", h.info.Name)
}

func (h *NotImplementedHandler) Validate(ctx context.Context) error {
	return fmt.Errorf("source '%s' is not yet implemented - handler development needed", h.info.Name)
}

// InitializeGlobalRegistry sets up the global registry with builtin sources.
// This should be called during application startup.
func InitializeGlobalRegistry() error {
	registry := GetGlobalRegistry()
	return RegisterBuiltinSources(registry)
}

// GetSourceStatus returns the implementation status of a source.
type SourceStatus struct {
	Name          string
	IsImplemented bool
	IsRegistered  bool
	HandlerType   string // "implemented", "placeholder", or "missing"
	SourceInfo    *SourceInfo
}

// GetSourceStatus returns detailed status information about a source.
func GetSourceStatus(sourceName string) (*SourceStatus, error) {
	registry := GetGlobalRegistry()

	status := &SourceStatus{
		Name:         sourceName,
		IsRegistered: registry.IsRegistered(sourceName),
	}

	if !status.IsRegistered {
		status.HandlerType = "missing"
		return status, nil
	}

	registration, _ := registry.GetRegistration(sourceName)
	status.SourceInfo = &registration.Info

	// Try to create a handler to see if it's implemented
	handler := registration.CreateFunc()
	switch handler.(type) {
	case *NotImplementedHandler:
		status.HandlerType = "placeholder"
		status.IsImplemented = false
	default:
		status.HandlerType = "implemented"
		status.IsImplemented = true
	}

	return status, nil
}

// ListAllSourceStatus returns status for all known sources (builtin + registered).
func ListAllSourceStatus() ([]*SourceStatus, error) {
	registry := GetGlobalRegistry()

	// Get all registered sources
	registrations := registry.ListRegistrations()
	statusList := make([]*SourceStatus, 0, len(registrations))

	for _, registration := range registrations {
		status, err := GetSourceStatus(registration.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get status for %s: %w", registration.Name, err)
		}
		statusList = append(statusList, status)
	}

	return statusList, nil
}
