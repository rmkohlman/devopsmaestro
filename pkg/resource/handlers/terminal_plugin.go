// Package handlers provides resource handlers for different resource types.
package handlers

import (
	"fmt"

	"devopsmaestro/models"
	"devopsmaestro/pkg/terminalbridge"

	"github.com/rmkohlman/MaestroSDK/resource"

	"gopkg.in/yaml.v3"
)

// KindTerminalPlugin is the resource kind for terminal plugins.
const KindTerminalPlugin = "TerminalPlugin"

// TerminalPluginHandler handles TerminalPlugin resources.
type TerminalPluginHandler struct{}

// NewTerminalPluginHandler creates a new TerminalPlugin handler.
func NewTerminalPluginHandler() *TerminalPluginHandler {
	return &TerminalPluginHandler{}
}

func (h *TerminalPluginHandler) Kind() string {
	return KindTerminalPlugin
}

// Apply creates or updates a terminal plugin from YAML data.
func (h *TerminalPluginHandler) Apply(ctx resource.Context, data []byte) (resource.Resource, error) {
	// Parse the YAML into the structured format
	var pluginYAML models.TerminalPluginYAML
	if err := yaml.Unmarshal(data, &pluginYAML); err != nil {
		return nil, fmt.Errorf("failed to parse terminal plugin YAML: %w", err)
	}

	// Convert YAML to DB model
	dbPlugin := &models.TerminalPluginDB{}
	if err := dbPlugin.FromYAML(pluginYAML); err != nil {
		return nil, fmt.Errorf("failed to convert terminal plugin YAML to DB model: %w", err)
	}

	// Get the plugin store from context
	pluginStore, err := h.getStore(ctx)
	if err != nil {
		return nil, err
	}

	// Upsert the plugin
	if err := pluginStore.UpsertTerminalPlugin(dbPlugin); err != nil {
		return nil, fmt.Errorf("failed to save terminal plugin: %w", err)
	}

	return &TerminalPluginResource{plugin: dbPlugin}, nil
}

// Get retrieves a terminal plugin by name.
func (h *TerminalPluginHandler) Get(ctx resource.Context, name string) (resource.Resource, error) {
	pluginStore, err := h.getStore(ctx)
	if err != nil {
		return nil, err
	}

	dbPlugin, err := pluginStore.GetTerminalPlugin(name)
	if err != nil {
		return nil, err
	}

	return &TerminalPluginResource{plugin: dbPlugin}, nil
}

// List returns all terminal plugins.
func (h *TerminalPluginHandler) List(ctx resource.Context) ([]resource.Resource, error) {
	pluginStore, err := h.getStore(ctx)
	if err != nil {
		return nil, err
	}

	dbPlugins, err := pluginStore.ListTerminalPlugins()
	if err != nil {
		return nil, err
	}

	result := make([]resource.Resource, len(dbPlugins))
	for i, p := range dbPlugins {
		result[i] = &TerminalPluginResource{plugin: p}
	}
	return result, nil
}

// Delete removes a terminal plugin by name.
func (h *TerminalPluginHandler) Delete(ctx resource.Context, name string) error {
	pluginStore, err := h.getStore(ctx)
	if err != nil {
		return err
	}

	return pluginStore.DeleteTerminalPlugin(name)
}

// ToYAML serializes a terminal plugin to YAML.
func (h *TerminalPluginHandler) ToYAML(res resource.Resource) ([]byte, error) {
	pr, ok := res.(*TerminalPluginResource)
	if !ok {
		return nil, fmt.Errorf("expected TerminalPluginResource, got %T", res)
	}

	yamlDoc, err := pr.plugin.ToYAML()
	if err != nil {
		return nil, fmt.Errorf("failed to convert terminal plugin to YAML: %w", err)
	}

	return yaml.Marshal(yamlDoc)
}

// getStore returns the PluginDataStore from context.
func (h *TerminalPluginHandler) getStore(ctx resource.Context) (terminalbridge.PluginDataStore, error) {
	if ctx.DataStore == nil {
		return nil, fmt.Errorf("no plugin store configured: DataStore is required")
	}

	ds, err := resource.DataStoreAs[terminalbridge.PluginDataStore](ctx)
	if err != nil {
		return nil, fmt.Errorf("DataStore does not implement PluginDataStore: %T", ctx.DataStore)
	}
	return ds, nil
}

// TerminalPluginResource wraps a models.TerminalPluginDB to implement resource.Resource.
type TerminalPluginResource struct {
	plugin *models.TerminalPluginDB
}

func (r *TerminalPluginResource) GetKind() string {
	return KindTerminalPlugin
}

func (r *TerminalPluginResource) GetName() string {
	return r.plugin.Name
}

func (r *TerminalPluginResource) Validate() error {
	if r.plugin.Name == "" {
		return fmt.Errorf("plugin name is required")
	}
	if r.plugin.Repo == "" {
		return fmt.Errorf("plugin repo is required")
	}
	return nil
}

// Plugin returns the underlying models.TerminalPluginDB.
func (r *TerminalPluginResource) Plugin() *models.TerminalPluginDB {
	return r.plugin
}

// NewTerminalPluginResource creates a new TerminalPluginResource from a DB model.
func NewTerminalPluginResource(db *models.TerminalPluginDB) *TerminalPluginResource {
	return &TerminalPluginResource{plugin: db}
}
