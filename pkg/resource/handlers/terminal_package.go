// Package handlers provides resource handlers for different resource types.
// Each handler knows how to CRUD a specific resource type (TerminalPackage, etc.)
package handlers

import (
	"fmt"

	"devopsmaestro/models"
	"devopsmaestro/pkg/resource"
	terminalpkg "devopsmaestro/pkg/terminalops/package"

	"gopkg.in/yaml.v3"
)

const KindTerminalPackage = "TerminalPackage"

// TerminalPackageHandler handles TerminalPackage resources.
type TerminalPackageHandler struct{}

// NewTerminalPackageHandler creates a new TerminalPackage handler.
func NewTerminalPackageHandler() *TerminalPackageHandler {
	return &TerminalPackageHandler{}
}

func (h *TerminalPackageHandler) Kind() string {
	return KindTerminalPackage
}

// Apply creates or updates a package from YAML data.
func (h *TerminalPackageHandler) Apply(ctx resource.Context, data []byte) (resource.Resource, error) {
	// Parse the YAML
	pkg, err := terminalpkg.ParseYAML(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse package YAML: %w", err)
	}

	// Get the DataStore from context
	dataStore, err := h.getDataStore(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to DB model
	dbPkg, err := h.toDBModel(pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to convert package to DB model: %w", err)
	}

	// Check if package exists
	existing, err := dataStore.GetTerminalPackage(pkg.Name)
	if err == nil && existing != nil {
		// Update existing package
		dbPkg.ID = existing.ID
		dbPkg.CreatedAt = existing.CreatedAt
		if err := dataStore.UpdateTerminalPackage(dbPkg); err != nil {
			return nil, fmt.Errorf("failed to update package: %w", err)
		}
	} else {
		// Create new package
		if err := dataStore.CreateTerminalPackage(dbPkg); err != nil {
			return nil, fmt.Errorf("failed to create package: %w", err)
		}
	}

	return &TerminalPackageResource{pkg: pkg}, nil
}

// Get retrieves a package by name.
func (h *TerminalPackageHandler) Get(ctx resource.Context, name string) (resource.Resource, error) {
	dataStore, err := h.getDataStore(ctx)
	if err != nil {
		return nil, err
	}

	// Try to get from database
	dbPkg, err := dataStore.GetTerminalPackage(name)
	if err != nil {
		return nil, err
	}

	pkg, err := h.fromDBModel(dbPkg)
	if err != nil {
		return nil, fmt.Errorf("failed to convert DB model to package: %w", err)
	}

	return &TerminalPackageResource{pkg: pkg}, nil
}

// List returns all packages.
func (h *TerminalPackageHandler) List(ctx resource.Context) ([]resource.Resource, error) {
	dataStore, err := h.getDataStore(ctx)
	if err != nil {
		return nil, err
	}

	// Get packages from database
	dbPackages, err := dataStore.ListTerminalPackages()
	if err != nil {
		return nil, err
	}

	packages := make([]*terminalpkg.Package, len(dbPackages))
	for i, dbPkg := range dbPackages {
		pkg, err := h.fromDBModel(dbPkg)
		if err != nil {
			return nil, fmt.Errorf("failed to convert DB model to package: %w", err)
		}
		packages[i] = pkg
	}

	// Convert to resources
	result := make([]resource.Resource, len(packages))
	for i, p := range packages {
		result[i] = &TerminalPackageResource{pkg: p}
	}

	return result, nil
}

// Delete removes a package by name.
func (h *TerminalPackageHandler) Delete(ctx resource.Context, name string) error {
	dataStore, err := h.getDataStore(ctx)
	if err != nil {
		return err
	}

	return dataStore.DeleteTerminalPackage(name)
}

// ToYAML serializes a package to YAML.
func (h *TerminalPackageHandler) ToYAML(res resource.Resource) ([]byte, error) {
	pr, ok := res.(*TerminalPackageResource)
	if !ok {
		return nil, fmt.Errorf("expected TerminalPackageResource, got %T", res)
	}

	yamlDoc := pr.pkg.ToYAML()
	return yaml.Marshal(yamlDoc)
}

// getDataStore returns the DataStore with terminal package operations from context.
func (h *TerminalPackageHandler) getDataStore(ctx resource.Context) (interface {
	CreateTerminalPackage(pkg *models.TerminalPackageDB) error
	UpdateTerminalPackage(pkg *models.TerminalPackageDB) error
	GetTerminalPackage(name string) (*models.TerminalPackageDB, error)
	ListTerminalPackages() ([]*models.TerminalPackageDB, error)
	DeleteTerminalPackage(name string) error
}, error) {
	if ctx.DataStore == nil {
		return nil, fmt.Errorf("DataStore is required for terminal package operations")
	}

	// Type assert to ensure DataStore has terminal package methods
	dataStore, ok := ctx.DataStore.(interface {
		CreateTerminalPackage(pkg *models.TerminalPackageDB) error
		UpdateTerminalPackage(pkg *models.TerminalPackageDB) error
		GetTerminalPackage(name string) (*models.TerminalPackageDB, error)
		ListTerminalPackages() ([]*models.TerminalPackageDB, error)
		DeleteTerminalPackage(name string) error
	})
	if !ok {
		return nil, fmt.Errorf("DataStore does not support terminal package operations: %T", ctx.DataStore)
	}

	return dataStore, nil
}

// toDBModel converts a Package to TerminalPackageDB model.
func (h *TerminalPackageHandler) toDBModel(pkg *terminalpkg.Package) (*models.TerminalPackageDB, error) {
	dbPkg := &models.TerminalPackageDB{
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

	// Convert arrays to JSON
	if err := dbPkg.SetPlugins(pkg.Plugins); err != nil {
		return nil, fmt.Errorf("failed to set plugins: %w", err)
	}

	if err := dbPkg.SetPrompts(pkg.Prompts); err != nil {
		return nil, fmt.Errorf("failed to set prompts: %w", err)
	}

	if err := dbPkg.SetProfiles(pkg.Profiles); err != nil {
		return nil, fmt.Errorf("failed to set profiles: %w", err)
	}

	// Convert WezTerm config if present
	if pkg.WezTerm != nil {
		weztermMap := map[string]interface{}{
			"fontSize":    pkg.WezTerm.FontSize,
			"colorScheme": pkg.WezTerm.ColorScheme,
			"fontFamily":  pkg.WezTerm.FontFamily,
		}
		if err := dbPkg.SetWezTerm(weztermMap); err != nil {
			return nil, fmt.Errorf("failed to set wezterm config: %w", err)
		}
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

// fromDBModel converts TerminalPackageDB to Package.
func (h *TerminalPackageHandler) fromDBModel(dbPkg *models.TerminalPackageDB) (*terminalpkg.Package, error) {
	pkg := &terminalpkg.Package{
		Name:     dbPkg.Name,
		Plugins:  dbPkg.GetPlugins(),
		Prompts:  dbPkg.GetPrompts(),
		Profiles: dbPkg.GetProfiles(),
		Enabled:  true, // default to enabled
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

	// Convert WezTerm config if present
	if weztermMap := dbPkg.GetWezTerm(); weztermMap != nil {
		wezterm := &terminalpkg.WezTermConfig{}
		if fontSize, ok := weztermMap["fontSize"].(float64); ok {
			wezterm.FontSize = int(fontSize)
		}
		if colorScheme, ok := weztermMap["colorScheme"].(string); ok {
			wezterm.ColorScheme = colorScheme
		}
		if fontFamily, ok := weztermMap["fontFamily"].(string); ok {
			wezterm.FontFamily = fontFamily
		}
		pkg.WezTerm = wezterm
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

// Helper functions splitAndTrim and trimSpace are already defined in nvim_package.go

// TerminalPackageResource wraps a terminalpkg.Package to implement resource.Resource.
type TerminalPackageResource struct {
	pkg *terminalpkg.Package
}

func (r *TerminalPackageResource) GetKind() string {
	return KindTerminalPackage
}

func (r *TerminalPackageResource) GetName() string {
	return r.pkg.Name
}

func (r *TerminalPackageResource) Validate() error {
	if r.pkg.Name == "" {
		return fmt.Errorf("package name is required")
	}

	// A terminal package can be empty (no plugins, prompts, or profiles) - it might just define inheritance
	// So we don't require at least one item like nvim packages do

	// Validate plugin names are not empty
	for i, plugin := range r.pkg.Plugins {
		if plugin == "" {
			return fmt.Errorf("plugin at index %d cannot be empty", i)
		}
	}

	// Validate prompt names are not empty
	for i, prompt := range r.pkg.Prompts {
		if prompt == "" {
			return fmt.Errorf("prompt at index %d cannot be empty", i)
		}
	}

	// Validate profile names are not empty
	for i, profile := range r.pkg.Profiles {
		if profile == "" {
			return fmt.Errorf("profile at index %d cannot be empty", i)
		}
	}

	return nil
}

// Package returns the underlying terminalpkg.Package.
func (r *TerminalPackageResource) Package() *terminalpkg.Package {
	return r.pkg
}
