// Package handlers provides resource handlers for different resource types.
// Each handler knows how to CRUD a specific resource type (NvimPlugin, NvimTheme, etc.)
package handlers

import (
	"fmt"

	"devopsmaestro/pkg/nvimops/plugin"
	"devopsmaestro/pkg/nvimops/store"
	"devopsmaestro/pkg/resource"

	"gopkg.in/yaml.v3"
)

const KindNvimPlugin = "NvimPlugin"

// NvimPluginHandler handles NvimPlugin resources.
type NvimPluginHandler struct{}

// NewNvimPluginHandler creates a new NvimPlugin handler.
func NewNvimPluginHandler() *NvimPluginHandler {
	return &NvimPluginHandler{}
}

func (h *NvimPluginHandler) Kind() string {
	return KindNvimPlugin
}

// Apply creates or updates a plugin from YAML data.
func (h *NvimPluginHandler) Apply(ctx resource.Context, data []byte) (resource.Resource, error) {
	// Parse the YAML
	p, err := plugin.ParseYAML(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse plugin YAML: %w", err)
	}

	// Get the appropriate store
	pluginStore, err := h.getStore(ctx)
	if err != nil {
		return nil, err
	}

	// Upsert the plugin
	if err := pluginStore.Upsert(p); err != nil {
		return nil, fmt.Errorf("failed to save plugin: %w", err)
	}

	return &NvimPluginResource{plugin: p}, nil
}

// Get retrieves a plugin by name.
func (h *NvimPluginHandler) Get(ctx resource.Context, name string) (resource.Resource, error) {
	pluginStore, err := h.getStore(ctx)
	if err != nil {
		return nil, err
	}

	p, err := pluginStore.Get(name)
	if err != nil {
		return nil, err
	}

	return &NvimPluginResource{plugin: p}, nil
}

// List returns all plugins.
func (h *NvimPluginHandler) List(ctx resource.Context) ([]resource.Resource, error) {
	pluginStore, err := h.getStore(ctx)
	if err != nil {
		return nil, err
	}

	plugins, err := pluginStore.List()
	if err != nil {
		return nil, err
	}

	result := make([]resource.Resource, len(plugins))
	for i, p := range plugins {
		result[i] = &NvimPluginResource{plugin: p}
	}
	return result, nil
}

// Delete removes a plugin by name.
func (h *NvimPluginHandler) Delete(ctx resource.Context, name string) error {
	pluginStore, err := h.getStore(ctx)
	if err != nil {
		return err
	}

	return pluginStore.Delete(name)
}

// ToYAML serializes a plugin to YAML.
func (h *NvimPluginHandler) ToYAML(res resource.Resource) ([]byte, error) {
	pr, ok := res.(*NvimPluginResource)
	if !ok {
		return nil, fmt.Errorf("expected NvimPluginResource, got %T", res)
	}

	yamlDoc := pr.plugin.ToYAML()
	return yaml.Marshal(yamlDoc)
}

// getStore returns the appropriate PluginStore based on context.
func (h *NvimPluginHandler) getStore(ctx resource.Context) (store.PluginStore, error) {
	// If PluginStore is directly provided, use it
	if ctx.PluginStore != nil {
		if ps, ok := ctx.PluginStore.(store.PluginStore); ok {
			return ps, nil
		}
		return nil, fmt.Errorf("invalid PluginStore type: %T", ctx.PluginStore)
	}

	// If DataStore is provided, use DBStoreAdapter
	if ctx.DataStore != nil {
		if ds, ok := ctx.DataStore.(store.PluginDataStore); ok {
			return store.NewDBStoreAdapter(ds), nil
		}
		return nil, fmt.Errorf("DataStore does not implement PluginDataStore: %T", ctx.DataStore)
	}

	// If ConfigDir is provided, use FileStore
	if ctx.ConfigDir != "" {
		return store.NewFileStore(ctx.ConfigDir)
	}

	// Default to standard file store location
	return store.DefaultFileStore()
}

// NvimPluginResource wraps a plugin.Plugin to implement resource.Resource.
type NvimPluginResource struct {
	plugin *plugin.Plugin
}

func (r *NvimPluginResource) GetKind() string {
	return KindNvimPlugin
}

func (r *NvimPluginResource) GetName() string {
	return r.plugin.Name
}

func (r *NvimPluginResource) Validate() error {
	if r.plugin.Name == "" {
		return fmt.Errorf("plugin name is required")
	}
	if r.plugin.Repo == "" {
		return fmt.Errorf("plugin repo is required")
	}
	return nil
}

// Plugin returns the underlying plugin.Plugin.
func (r *NvimPluginResource) Plugin() *plugin.Plugin {
	return r.plugin
}
