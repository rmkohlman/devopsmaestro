package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// FilePackageCreator creates packages as YAML files in the file system.
// This is the default implementation for nvp standalone operation.
type FilePackageCreator struct {
	// PackagesDir is the directory where package YAML files will be written
	PackagesDir string
}

// NewFilePackageCreator creates a new FilePackageCreator.
func NewFilePackageCreator(packagesDir string) *FilePackageCreator {
	return &FilePackageCreator{
		PackagesDir: packagesDir,
	}
}

// CreatePackage creates or updates a package YAML file with the given plugins.
func (f *FilePackageCreator) CreatePackage(sourceName string, plugins []string) error {
	if f.PackagesDir == "" {
		return fmt.Errorf("packages directory not set")
	}

	// Create packages directory if it doesn't exist
	if err := os.MkdirAll(f.PackagesDir, 0755); err != nil {
		return fmt.Errorf("failed to create packages directory: %w", err)
	}

	// Create package
	now := time.Now()
	pkg := &Package{
		Name:        sourceName,
		Description: fmt.Sprintf("Auto-generated package from %s source sync", sourceName),
		Category:    "source-sync",
		Plugins:     plugins,
		Enabled:     true,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}

	// Convert to YAML
	packageYAML := pkg.ToYAML()
	packageYAML.Metadata.Labels = map[string]string{
		"source":         sourceName,
		"auto-generated": "true",
		"sync-time":      now.Format(time.RFC3339),
	}

	// Write to file
	filename := filepath.Join(f.PackagesDir, sourceName+".yaml")
	data, err := yaml.Marshal(packageYAML)
	if err != nil {
		return fmt.Errorf("failed to serialize package: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write package file: %w", err)
	}

	return nil
}

// FilePackageCreator implements sync.PackageCreator interface.
// Interface verification is done at the call site to avoid import cycles.
