// Package handlers provides resource handlers for different resource types.
// Each handler knows how to CRUD a specific resource type (NvimPlugin, NvimTheme, etc.)
package handlers

import (
	"fmt"

	"devopsmaestro/models"
	nvimpkg "devopsmaestro/pkg/nvimops/package"
	"devopsmaestro/pkg/resource"

	"gopkg.in/yaml.v3"
)

const KindNvimPackage = "NvimPackage"

// NvimPackageHandler handles NvimPackage resources.
type NvimPackageHandler struct{}

// NewNvimPackageHandler creates a new NvimPackage handler.
func NewNvimPackageHandler() *NvimPackageHandler {
	return &NvimPackageHandler{}
}

func (h *NvimPackageHandler) Kind() string {
	return KindNvimPackage
}

// Apply creates or updates a pkg from YAML data.
func (h *NvimPackageHandler) Apply(ctx resource.Context, data []byte) (resource.Resource, error) {
	// Parse the YAML
	pkg, err := nvimpkg.ParseYAML(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pkg YAML: %w", err)
	}

	// Get the DataStore from context
	dataStore, err := h.getDataStore(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to DB model
	dbPkg, err := h.toDBModel(pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to convert pkg to DB model: %w", err)
	}

	// Check if pkg exists
	existing, err := dataStore.GetPackage(pkg.Name)
	if err == nil && existing != nil {
		// Update existing pkg
		dbPkg.ID = existing.ID
		dbPkg.CreatedAt = existing.CreatedAt
		if err := dataStore.UpdatePackage(dbPkg); err != nil {
			return nil, fmt.Errorf("failed to update pkg: %w", err)
		}
	} else {
		// Create new pkg
		if err := dataStore.CreatePackage(dbPkg); err != nil {
			return nil, fmt.Errorf("failed to create pkg: %w", err)
		}
	}

	return &NvimPackageResource{pkg: pkg}, nil
}

// Get retrieves a pkg by name.
func (h *NvimPackageHandler) Get(ctx resource.Context, name string) (resource.Resource, error) {
	dataStore, err := h.getDataStore(ctx)
	if err != nil {
		return nil, err
	}

	dbPkg, err := dataStore.GetPackage(name)
	if err != nil {
		return nil, err
	}

	pkg, err := h.fromDBModel(dbPkg)
	if err != nil {
		return nil, fmt.Errorf("failed to convert DB model to pkg: %w", err)
	}

	return &NvimPackageResource{pkg: pkg}, nil
}

// List returns all pkgs.
func (h *NvimPackageHandler) List(ctx resource.Context) ([]resource.Resource, error) {
	dataStore, err := h.getDataStore(ctx)
	if err != nil {
		return nil, err
	}

	dbPackages, err := dataStore.ListPackages()
	if err != nil {
		return nil, err
	}

	result := make([]resource.Resource, len(dbPackages))
	for i, dbPkg := range dbPackages {
		pkg, err := h.fromDBModel(dbPkg)
		if err != nil {
			return nil, fmt.Errorf("failed to convert DB model to pkg: %w", err)
		}
		result[i] = &NvimPackageResource{pkg: pkg}
	}
	return result, nil
}

// Delete removes a pkg by name.
func (h *NvimPackageHandler) Delete(ctx resource.Context, name string) error {
	dataStore, err := h.getDataStore(ctx)
	if err != nil {
		return err
	}

	return dataStore.DeletePackage(name)
}

// ToYAML serializes a pkg to YAML.
func (h *NvimPackageHandler) ToYAML(res resource.Resource) ([]byte, error) {
	pr, ok := res.(*NvimPackageResource)
	if !ok {
		return nil, fmt.Errorf("expected NvimPackageResource, got %T", res)
	}

	yamlDoc := pr.pkg.ToYAML()
	return yaml.Marshal(yamlDoc)
}

// getDataStore returns the DataStore with pkg operations from context.
func (h *NvimPackageHandler) getDataStore(ctx resource.Context) (interface {
	CreatePackage(pkg *models.NvimPackageDB) error
	UpdatePackage(pkg *models.NvimPackageDB) error
	GetPackage(name string) (*models.NvimPackageDB, error)
	ListPackages() ([]*models.NvimPackageDB, error)
	DeletePackage(name string) error
}, error) {
	if ctx.DataStore == nil {
		return nil, fmt.Errorf("DataStore is required for pkg operations")
	}

	// Type assert to ensure DataStore has pkg methods
	dataStore, ok := ctx.DataStore.(interface {
		CreatePackage(pkg *models.NvimPackageDB) error
		UpdatePackage(pkg *models.NvimPackageDB) error
		GetPackage(name string) (*models.NvimPackageDB, error)
		ListPackages() ([]*models.NvimPackageDB, error)
		DeletePackage(name string) error
	})
	if !ok {
		return nil, fmt.Errorf("DataStore does not support pkg operations: %T", ctx.DataStore)
	}

	return dataStore, nil
}

// toDBModel converts a Package to NvimPackageDB model.
func (h *NvimPackageHandler) toDBModel(pkg *nvimpkg.Package) (*models.NvimPackageDB, error) {
	dbPkg := &models.NvimPackageDB{
		Name: pkg.Name,
	}

	// Set optional fields
	if pkg.Description != "" {
		dbPkg.Description.String = pkg.Description
		dbPkg.Description.Valid = true
	}

	if pkg.Category != "" {
		dbPkg.Category.String = pkg.Category
		dbPkg.Category.Valid = true
	}

	if pkg.Extends != "" {
		dbPkg.Extends.String = pkg.Extends
		dbPkg.Extends.Valid = true
	}

	// Convert plugins to JSON array
	if err := dbPkg.SetPlugins(pkg.Plugins); err != nil {
		return nil, fmt.Errorf("failed to set plugins: %w", err)
	}

	// Convert tags to labels (tags become a "tags" label)
	labels := make(map[string]string)
	if len(pkg.Tags) > 0 {
		// Join tags with commas for the "tags" label
		tagStr := ""
		for i, tag := range pkg.Tags {
			if i > 0 {
				tagStr += ","
			}
			tagStr += tag
		}
		labels["tags"] = tagStr
	}

	// Add enabled status as label
	if !pkg.Enabled {
		labels["enabled"] = "false"
	}

	if err := dbPkg.SetLabels(labels); err != nil {
		return nil, fmt.Errorf("failed to set labels: %w", err)
	}

	return dbPkg, nil
}

// fromDBModel converts NvimPackageDB to Package.
func (h *NvimPackageHandler) fromDBModel(dbPkg *models.NvimPackageDB) (*nvimpkg.Package, error) {
	pkg := &nvimpkg.Package{
		Name:    dbPkg.Name,
		Plugins: dbPkg.GetPlugins(),
		Enabled: true, // default to enabled
	}

	// Set optional fields
	if dbPkg.Description.Valid {
		pkg.Description = dbPkg.Description.String
	}

	if dbPkg.Category.Valid {
		pkg.Category = dbPkg.Category.String
	}

	if dbPkg.Extends.Valid {
		pkg.Extends = dbPkg.Extends.String
	}

	// Convert labels back to tags and enabled status
	labels := dbPkg.GetLabels()
	if tagStr, exists := labels["tags"]; exists && tagStr != "" {
		// Split comma-separated tags
		tags := []string{}
		for _, tag := range splitAndTrim(tagStr, ",") {
			if tag != "" {
				tags = append(tags, tag)
			}
		}
		pkg.Tags = tags
	}

	if enabledStr, exists := labels["enabled"]; exists && enabledStr == "false" {
		pkg.Enabled = false
	}

	// Set timestamps if available
	if !dbPkg.CreatedAt.IsZero() {
		pkg.CreatedAt = &dbPkg.CreatedAt
	}
	if !dbPkg.UpdatedAt.IsZero() {
		pkg.UpdatedAt = &dbPkg.UpdatedAt
	}

	return pkg, nil
}

// splitAndTrim splits a string by delimiter and trims whitespace from each part.
func splitAndTrim(s, delim string) []string {
	if s == "" {
		return []string{}
	}

	parts := []string{}
	current := ""
	for _, char := range s {
		if string(char) == delim {
			if trimmed := trimSpace(current); trimmed != "" {
				parts = append(parts, trimmed)
			}
			current = ""
		} else {
			current += string(char)
		}
	}
	if trimmed := trimSpace(current); trimmed != "" {
		parts = append(parts, trimmed)
	}
	return parts
}

// trimSpace trims leading and trailing whitespace.
func trimSpace(s string) string {
	// Simple trim implementation without importing strings
	start := 0
	end := len(s)

	// Trim leading whitespace
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	// Trim trailing whitespace
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}

// NvimPackageResource wraps a nvimpkg.Package to implement resource.Resource.
type NvimPackageResource struct {
	pkg *nvimpkg.Package
}

func (r *NvimPackageResource) GetKind() string {
	return KindNvimPackage
}

func (r *NvimPackageResource) GetName() string {
	return r.pkg.Name
}

func (r *NvimPackageResource) Validate() error {
	if r.pkg.Name == "" {
		return fmt.Errorf("pkg name is required")
	}
	if len(r.pkg.Plugins) == 0 {
		return fmt.Errorf("pkg must contain at least one plugin")
	}
	// Validate plugin names are not empty
	for i, plugin := range r.pkg.Plugins {
		if plugin == "" {
			return fmt.Errorf("plugin at index %d cannot be empty", i)
		}
	}
	return nil
}

// Package returns the underlying nvimpkg.Package.
func (r *NvimPackageResource) Package() *nvimpkg.Package {
	return r.pkg
}
