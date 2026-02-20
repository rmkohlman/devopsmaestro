package pkg

import (
	"testing"
)

func TestNewPackage(t *testing.T) {
	pkg := NewPackage("test-pkg")

	if pkg.Name != "test-pkg" {
		t.Errorf("Expected name 'test-pkg', got %s", pkg.Name)
	}

	if !pkg.Enabled {
		t.Error("Expected package to be enabled by default")
	}

	if pkg.Plugins == nil {
		t.Error("Expected plugins slice to be initialized")
	}

	if pkg.Prompts == nil {
		t.Error("Expected prompts slice to be initialized")
	}

	if pkg.Profiles == nil {
		t.Error("Expected profiles slice to be initialized")
	}
}

func TestNewPackageYAML(t *testing.T) {
	py := NewPackageYAML("test-pkg")

	if py.APIVersion != "devopsmaestro.io/v1" {
		t.Errorf("Expected API version 'devopsmaestro.io/v1', got %s", py.APIVersion)
	}

	if py.Kind != "TerminalPackage" {
		t.Errorf("Expected kind 'TerminalPackage', got %s", py.Kind)
	}

	if py.Metadata.Name != "test-pkg" {
		t.Errorf("Expected name 'test-pkg', got %s", py.Metadata.Name)
	}
}

func TestPackageYAMLToPackage(t *testing.T) {
	py := &PackageYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "TerminalPackage",
		Metadata: PackageMetadata{
			Name:        "test-pkg",
			Description: "Test package",
			Category:    "test",
			Tags:        []string{"test", "example"},
		},
		Spec: PackageSpec{
			Extends:  "core",
			Plugins:  StringOrSlice{"plugin1", "plugin2"},
			Prompts:  StringOrSlice{"prompt1"},
			Profiles: StringOrSlice{"profile1", "profile2"},
		},
	}

	pkg := py.ToPackage()

	if pkg.Name != "test-pkg" {
		t.Errorf("Expected name 'test-pkg', got %s", pkg.Name)
	}

	if pkg.Description != "Test package" {
		t.Errorf("Expected description 'Test package', got %s", pkg.Description)
	}

	if pkg.Extends != "core" {
		t.Errorf("Expected extends 'core', got %s", pkg.Extends)
	}

	if len(pkg.Plugins) != 2 || pkg.Plugins[0] != "plugin1" || pkg.Plugins[1] != "plugin2" {
		t.Errorf("Expected plugins [plugin1, plugin2], got %v", pkg.Plugins)
	}

	if len(pkg.Prompts) != 1 || pkg.Prompts[0] != "prompt1" {
		t.Errorf("Expected prompts [prompt1], got %v", pkg.Prompts)
	}

	if len(pkg.Profiles) != 2 || pkg.Profiles[0] != "profile1" || pkg.Profiles[1] != "profile2" {
		t.Errorf("Expected profiles [profile1, profile2], got %v", pkg.Profiles)
	}
}

func TestPackageToYAML(t *testing.T) {
	pkg := &Package{
		Name:        "test-pkg",
		Description: "Test package",
		Category:    "test",
		Tags:        []string{"test", "example"},
		Extends:     "core",
		Plugins:     []string{"plugin1", "plugin2"},
		Prompts:     []string{"prompt1"},
		Profiles:    []string{"profile1", "profile2"},
		Enabled:     true,
	}

	py := pkg.ToYAML()

	if py.APIVersion != "devopsmaestro.io/v1" {
		t.Errorf("Expected API version 'devopsmaestro.io/v1', got %s", py.APIVersion)
	}

	if py.Kind != "TerminalPackage" {
		t.Errorf("Expected kind 'TerminalPackage', got %s", py.Kind)
	}

	if py.Metadata.Name != "test-pkg" {
		t.Errorf("Expected name 'test-pkg', got %s", py.Metadata.Name)
	}

	if py.Spec.Extends != "core" {
		t.Errorf("Expected extends 'core', got %s", py.Spec.Extends)
	}

	if len(py.Spec.Plugins) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(py.Spec.Plugins))
	}

	// When enabled is true, it should not be included in YAML (to avoid clutter)
	if py.Spec.Enabled != nil {
		t.Error("Expected enabled field to be nil when true")
	}
}

func TestPackageToYAMLDisabled(t *testing.T) {
	pkg := &Package{
		Name:    "test-pkg",
		Enabled: false,
	}

	py := pkg.ToYAML()

	if py.Spec.Enabled == nil || *py.Spec.Enabled != false {
		t.Error("Expected enabled field to be set to false when package is disabled")
	}
}
