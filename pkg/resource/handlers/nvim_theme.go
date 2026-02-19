package handlers

import (
	"fmt"

	"devopsmaestro/pkg/nvimops/theme"
	"devopsmaestro/pkg/nvimops/theme/library"
	"devopsmaestro/pkg/resource"
)

const KindNvimTheme = "NvimTheme"

// NvimThemeHandler handles NvimTheme resources.
type NvimThemeHandler struct{}

// NewNvimThemeHandler creates a new NvimTheme handler.
func NewNvimThemeHandler() *NvimThemeHandler {
	return &NvimThemeHandler{}
}

func (h *NvimThemeHandler) Kind() string {
	return KindNvimTheme
}

// Apply creates or updates a theme from YAML data.
func (h *NvimThemeHandler) Apply(ctx resource.Context, data []byte) (resource.Resource, error) {
	// Parse the YAML
	t, err := theme.ParseYAML(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse theme YAML: %w", err)
	}

	// Validate
	if err := t.Validate(); err != nil {
		return nil, fmt.Errorf("invalid theme: %w", err)
	}

	// Get the appropriate store
	themeStore, err := h.getStore(ctx)
	if err != nil {
		return nil, err
	}

	// Save the theme
	if err := themeStore.Save(t); err != nil {
		return nil, fmt.Errorf("failed to save theme: %w", err)
	}

	return &NvimThemeResource{theme: t}, nil
}

// Get retrieves a theme by name.
func (h *NvimThemeHandler) Get(ctx resource.Context, name string) (resource.Resource, error) {
	themeStore, err := h.getStore(ctx)
	if err != nil {
		return nil, err
	}

	// Try to get from user store first
	t, err := themeStore.Get(name)
	if err != nil {
		// If not found in store, try the embedded library as fallback
		libraryTheme, libraryErr := library.Get(name)
		if libraryErr != nil {
			// Return the original store error if library also doesn't have it
			return nil, err
		}
		t = libraryTheme
	}

	return &NvimThemeResource{theme: t}, nil
}

// List returns all themes.
func (h *NvimThemeHandler) List(ctx resource.Context) ([]resource.Resource, error) {
	themeStore, err := h.getStore(ctx)
	if err != nil {
		return nil, err
	}

	// Get user themes from store
	userThemes, err := themeStore.List()
	if err != nil {
		return nil, err
	}

	// Get library themes
	libraryInfo, err := library.List()
	if err != nil {
		// If library fails, just return user themes
		result := make([]resource.Resource, len(userThemes))
		for i, t := range userThemes {
			result[i] = &NvimThemeResource{theme: t}
		}
		return result, nil
	}

	// Create a map of user theme names for deduplication
	userThemeNames := make(map[string]bool)
	for _, t := range userThemes {
		userThemeNames[t.Name] = true
	}

	// Convert user themes to resources
	result := make([]resource.Resource, len(userThemes))
	for i, t := range userThemes {
		result[i] = &NvimThemeResource{theme: t}
	}

	// Add library themes that aren't overridden by user themes
	for _, info := range libraryInfo {
		if !userThemeNames[info.Name] {
			libraryTheme, err := library.Get(info.Name)
			if err == nil {
				result = append(result, &NvimThemeResource{theme: libraryTheme})
			}
		}
	}

	return result, nil
}

// Delete removes a theme by name.
func (h *NvimThemeHandler) Delete(ctx resource.Context, name string) error {
	themeStore, err := h.getStore(ctx)
	if err != nil {
		return err
	}

	return themeStore.Delete(name)
}

// ToYAML serializes a theme to YAML.
func (h *NvimThemeHandler) ToYAML(res resource.Resource) ([]byte, error) {
	tr, ok := res.(*NvimThemeResource)
	if !ok {
		return nil, fmt.Errorf("expected NvimThemeResource, got %T", res)
	}

	return tr.theme.ToYAML()
}

// getStore returns the appropriate theme.Store based on context.
func (h *NvimThemeHandler) getStore(ctx resource.Context) (theme.Store, error) {
	// If ThemeStore is directly provided, use it
	if ctx.ThemeStore != nil {
		if ts, ok := ctx.ThemeStore.(theme.Store); ok {
			return ts, nil
		}
		return nil, fmt.Errorf("invalid ThemeStore type: %T", ctx.ThemeStore)
	}

	// If DataStore is provided, use DBStoreAdapter
	if ctx.DataStore != nil {
		if ds, ok := ctx.DataStore.(theme.ThemeDataStore); ok {
			return theme.NewDBStoreAdapter(ds), nil
		}
		return nil, fmt.Errorf("DataStore does not implement ThemeDataStore: %T", ctx.DataStore)
	}

	// If ConfigDir is provided, use FileStore
	if ctx.ConfigDir != "" {
		fs := theme.NewFileStore(ctx.ConfigDir)
		if err := fs.Init(); err != nil {
			return nil, fmt.Errorf("failed to initialize theme store: %w", err)
		}
		return fs, nil
	}

	// Default to standard file store location
	return nil, fmt.Errorf("no theme store configured: provide ThemeStore, DataStore, or ConfigDir")
}

// NvimThemeResource wraps a theme.Theme to implement resource.Resource.
type NvimThemeResource struct {
	theme *theme.Theme
}

func (r *NvimThemeResource) GetKind() string {
	return KindNvimTheme
}

func (r *NvimThemeResource) GetName() string {
	return r.theme.Name
}

func (r *NvimThemeResource) Validate() error {
	return r.theme.Validate()
}

// Theme returns the underlying theme.Theme.
func (r *NvimThemeResource) Theme() *theme.Theme {
	return r.theme
}
