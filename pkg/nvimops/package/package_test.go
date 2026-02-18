package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPackage(t *testing.T) {
	pkg := NewPackage("test-package")

	assert.Equal(t, "test-package", pkg.Name)
	assert.True(t, pkg.Enabled)
	assert.Empty(t, pkg.Plugins)
	assert.Empty(t, pkg.Description)
	assert.Empty(t, pkg.Category)
	assert.Empty(t, pkg.Tags)
	assert.Empty(t, pkg.Extends)
}

func TestNewPackageYAML(t *testing.T) {
	pkg := NewPackageYAML("test-package")

	assert.Equal(t, "devopsmaestro.io/v1", pkg.APIVersion)
	assert.Equal(t, "NvimPackage", pkg.Kind)
	assert.Equal(t, "test-package", pkg.Metadata.Name)
	assert.Empty(t, pkg.Spec.Plugins)
	assert.Nil(t, pkg.Spec.Enabled) // Should be nil (unset)
}

func TestToPackage(t *testing.T) {
	py := &PackageYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "NvimPackage",
		Metadata: PackageMetadata{
			Name:        "go-dev",
			Description: "Go development essentials",
			Category:    "language",
			Tags:        []string{"golang", "lsp"},
		},
		Spec: PackageSpec{
			Extends: "core",
			Plugins: []string{"nvim-lspconfig", "nvim-treesitter", "telescope"},
		},
	}

	pkg := py.ToPackage()

	assert.Equal(t, "go-dev", pkg.Name)
	assert.Equal(t, "Go development essentials", pkg.Description)
	assert.Equal(t, "language", pkg.Category)
	assert.Equal(t, []string{"golang", "lsp"}, pkg.Tags)
	assert.Equal(t, "core", pkg.Extends)
	assert.Equal(t, []string{"nvim-lspconfig", "nvim-treesitter", "telescope"}, pkg.Plugins)
	assert.True(t, pkg.Enabled) // Should default to true
}

func TestToPackageWithDisabled(t *testing.T) {
	enabled := false
	py := &PackageYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "NvimPackage",
		Metadata: PackageMetadata{
			Name: "disabled-package",
		},
		Spec: PackageSpec{
			Enabled: &enabled,
		},
	}

	pkg := py.ToPackage()
	assert.False(t, pkg.Enabled)
}

func TestToYAML(t *testing.T) {
	pkg := &Package{
		Name:        "go-dev",
		Description: "Go development essentials",
		Category:    "language",
		Tags:        []string{"golang", "lsp"},
		Extends:     "core",
		Plugins:     []string{"nvim-lspconfig", "nvim-treesitter", "telescope"},
		Enabled:     true,
	}

	py := pkg.ToYAML()

	assert.Equal(t, "devopsmaestro.io/v1", py.APIVersion)
	assert.Equal(t, "NvimPackage", py.Kind)
	assert.Equal(t, "go-dev", py.Metadata.Name)
	assert.Equal(t, "Go development essentials", py.Metadata.Description)
	assert.Equal(t, "language", py.Metadata.Category)
	assert.Equal(t, []string{"golang", "lsp"}, py.Metadata.Tags)
	assert.Equal(t, "core", py.Spec.Extends)
	assert.Equal(t, []string{"nvim-lspconfig", "nvim-treesitter", "telescope"}, []string(py.Spec.Plugins))
	assert.Nil(t, py.Spec.Enabled) // Should be nil since package is enabled (default)
}

func TestToYAMLWithDisabled(t *testing.T) {
	pkg := &Package{
		Name:    "disabled-package",
		Enabled: false,
	}

	py := pkg.ToYAML()
	require.NotNil(t, py.Spec.Enabled)
	assert.False(t, *py.Spec.Enabled)
}

func TestParseYAML(t *testing.T) {
	yamlData := `apiVersion: devopsmaestro.io/v1
kind: NvimPackage
metadata:
  name: go-dev
  description: "Go development essentials"
  category: language
  tags:
    - golang
    - lsp
spec:
  extends: core
  plugins:
    - nvim-lspconfig
    - nvim-treesitter
    - telescope`

	pkg, err := ParseYAML([]byte(yamlData))
	require.NoError(t, err)

	assert.Equal(t, "go-dev", pkg.Name)
	assert.Equal(t, "Go development essentials", pkg.Description)
	assert.Equal(t, "language", pkg.Category)
	assert.Equal(t, []string{"golang", "lsp"}, pkg.Tags)
	assert.Equal(t, "core", pkg.Extends)
	assert.Equal(t, []string{"nvim-lspconfig", "nvim-treesitter", "telescope"}, pkg.Plugins)
	assert.True(t, pkg.Enabled)
}

func TestParseYAMLWithStringOrSlice(t *testing.T) {
	// Test single plugin as string
	yamlData := `apiVersion: devopsmaestro.io/v1
kind: NvimPackage
metadata:
  name: simple-package
spec:
  plugins: telescope`

	pkg, err := ParseYAML([]byte(yamlData))
	require.NoError(t, err)
	assert.Equal(t, []string{"telescope"}, pkg.Plugins)
}

func TestToYAMLBytes(t *testing.T) {
	pkg := &Package{
		Name:    "test-package",
		Plugins: []string{"telescope", "treesitter"},
		Enabled: true,
	}

	data, err := pkg.ToYAMLBytes()
	require.NoError(t, err)
	assert.Contains(t, string(data), "name: test-package")
	assert.Contains(t, string(data), "telescope")
	assert.Contains(t, string(data), "treesitter")
}

func TestValidatePackageYAML(t *testing.T) {
	tests := []struct {
		name      string
		pkg       PackageYAML
		wantError string
	}{
		{
			name: "valid package",
			pkg: PackageYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "NvimPackage",
				Metadata:   PackageMetadata{Name: "test"},
				Spec:       PackageSpec{Plugins: []string{"telescope"}},
			},
			wantError: "",
		},
		{
			name: "invalid api version",
			pkg: PackageYAML{
				APIVersion: "invalid/v1",
				Kind:       "NvimPackage",
				Metadata:   PackageMetadata{Name: "test"},
			},
			wantError: "unsupported apiVersion",
		},
		{
			name: "invalid kind",
			pkg: PackageYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "InvalidKind",
				Metadata:   PackageMetadata{Name: "test"},
			},
			wantError: "invalid kind",
		},
		{
			name: "missing name",
			pkg: PackageYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "NvimPackage",
				Metadata:   PackageMetadata{},
			},
			wantError: "metadata.name is required",
		},
		{
			name: "empty plugin name",
			pkg: PackageYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "NvimPackage",
				Metadata:   PackageMetadata{Name: "test"},
				Spec:       PackageSpec{Plugins: []string{"telescope", ""}},
			},
			wantError: "plugin name at index 1 cannot be empty",
		},
		{
			name: "self-extending package",
			pkg: PackageYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "NvimPackage",
				Metadata:   PackageMetadata{Name: "test"},
				Spec:       PackageSpec{Extends: "test"},
			},
			wantError: "package cannot extend itself",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePackageYAML(&tt.pkg)
			if tt.wantError == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantError)
			}
		})
	}
}
