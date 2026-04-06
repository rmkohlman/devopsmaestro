// Package handlers provides resource handlers for different resource types.
// GlobalDefaults handles the global-level CA certs and build args stored in the
// defaults table. This enables round-trip export/import via dvm get all -o yaml
// and dvm apply -f backup.yaml.
package handlers

import (
	"encoding/json"
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/resource"

	"gopkg.in/yaml.v3"
)

const KindGlobalDefaults = "GlobalDefaults"

// globalDefaultsYAML is the YAML document structure for global defaults.
type globalDefaultsYAML struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Metadata   globalDefaultsMetadata `yaml:"metadata"`
	Spec       globalDefaultsSpec     `yaml:"spec"`
}

type globalDefaultsMetadata struct {
	Name string `yaml:"name"`
}

type globalDefaultsSpec struct {
	Theme               string                `yaml:"theme,omitempty"`
	BuildArgs           map[string]string     `yaml:"buildArgs,omitempty"`
	CACerts             []models.CACertConfig `yaml:"caCerts,omitempty"`
	NvimPackage         string                `yaml:"nvimPackage,omitempty"`
	TerminalPackage     string                `yaml:"terminalPackage,omitempty"`
	Plugins             []string              `yaml:"plugins,omitempty"`
	RegistryOCI         string                `yaml:"registryOci,omitempty"`
	RegistryPyPI        string                `yaml:"registryPypi,omitempty"`
	RegistryNPM         string                `yaml:"registryNpm,omitempty"`
	RegistryGo          string                `yaml:"registryGo,omitempty"`
	RegistryHTTP        string                `yaml:"registryHttp,omitempty"`
	RegistryIdleTimeout string                `yaml:"registryIdleTimeout,omitempty"`
}

// GlobalDefaultsHandler handles GlobalDefaults resources.
type GlobalDefaultsHandler struct{}

// NewGlobalDefaultsHandler creates a new GlobalDefaults handler.
func NewGlobalDefaultsHandler() *GlobalDefaultsHandler {
	return &GlobalDefaultsHandler{}
}

func (h *GlobalDefaultsHandler) Kind() string {
	return KindGlobalDefaults
}

// Apply restores all global defaults from YAML data.
func (h *GlobalDefaultsHandler) Apply(ctx resource.Context, data []byte) (resource.Resource, error) {
	var doc globalDefaultsYAML
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse GlobalDefaults YAML: %w", err)
	}

	ds, err := resource.DataStoreAs[db.DefaultsStore](ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DefaultStore: %w", err)
	}

	spec := &doc.Spec

	// Restore theme
	if spec.Theme != "" {
		if err := ds.SetDefault("theme", spec.Theme); err != nil {
			return nil, fmt.Errorf("failed to set global theme: %w", err)
		}
	}

	// Restore build args (JSON object)
	if len(spec.BuildArgs) > 0 {
		b, err := json.Marshal(spec.BuildArgs)
		if err != nil {
			return nil, fmt.Errorf("failed to encode build args: %w", err)
		}
		if err := ds.SetDefault("build-args", string(b)); err != nil {
			return nil, fmt.Errorf("failed to set global build args: %w", err)
		}
	}

	// Restore CA certs (JSON array)
	if len(spec.CACerts) > 0 {
		b, err := json.Marshal(spec.CACerts)
		if err != nil {
			return nil, fmt.Errorf("failed to encode CA certs: %w", err)
		}
		if err := ds.SetDefault("ca-certs", string(b)); err != nil {
			return nil, fmt.Errorf("failed to set global CA certs: %w", err)
		}
	}

	// Restore nvim package
	if spec.NvimPackage != "" {
		if err := ds.SetDefault("nvim-package", spec.NvimPackage); err != nil {
			return nil, fmt.Errorf("failed to set global nvim-package: %w", err)
		}
	}

	// Restore terminal package
	if spec.TerminalPackage != "" {
		if err := ds.SetDefault("terminal-package", spec.TerminalPackage); err != nil {
			return nil, fmt.Errorf("failed to set global terminal-package: %w", err)
		}
	}

	// Restore plugins (JSON array)
	if len(spec.Plugins) > 0 {
		b, err := json.Marshal(spec.Plugins)
		if err != nil {
			return nil, fmt.Errorf("failed to encode plugins: %w", err)
		}
		if err := ds.SetDefault("plugins", string(b)); err != nil {
			return nil, fmt.Errorf("failed to set global plugins: %w", err)
		}
	}

	// Restore registry type defaults
	registryDefaults := map[string]string{
		"registry-oci":  spec.RegistryOCI,
		"registry-pypi": spec.RegistryPyPI,
		"registry-npm":  spec.RegistryNPM,
		"registry-go":   spec.RegistryGo,
		"registry-http": spec.RegistryHTTP,
	}
	for key, val := range registryDefaults {
		if val != "" {
			if err := ds.SetDefault(key, val); err != nil {
				return nil, fmt.Errorf("failed to set global %s: %w", key, err)
			}
		}
	}

	// Restore registry idle timeout
	if spec.RegistryIdleTimeout != "" {
		if err := ds.SetDefault("registry-idle-timeout", spec.RegistryIdleTimeout); err != nil {
			return nil, fmt.Errorf("failed to set global registry-idle-timeout: %w", err)
		}
	}

	return newGlobalDefaultsResourceFromSpec(spec), nil
}

// Get retrieves global defaults. Name is ignored (there's only one).
func (h *GlobalDefaultsHandler) Get(ctx resource.Context, name string) (resource.Resource, error) {
	ds, err := resource.DataStoreAs[db.DefaultsStore](ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DefaultStore: %w", err)
	}

	spec, err := loadGlobalDefaults(ds)
	if err != nil {
		return nil, err
	}

	return newGlobalDefaultsResourceFromSpec(spec), nil
}

// List returns a single GlobalDefaults resource if any defaults exist.
func (h *GlobalDefaultsHandler) List(ctx resource.Context) ([]resource.Resource, error) {
	ds, err := resource.DataStoreAs[db.DefaultsStore](ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DefaultStore: %w", err)
	}

	spec, err := loadGlobalDefaults(ds)
	if err != nil {
		return nil, err
	}

	// Only include if there's actually something to export
	if spec.isEmpty() {
		return nil, nil
	}

	return []resource.Resource{newGlobalDefaultsResourceFromSpec(spec)}, nil
}

// Delete clears all global defaults. Name is ignored.
func (h *GlobalDefaultsHandler) Delete(ctx resource.Context, name string) error {
	ds, err := resource.DataStoreAs[db.DefaultsStore](ctx)
	if err != nil {
		return fmt.Errorf("failed to get DefaultStore: %w", err)
	}

	allKeys := []string{
		"theme",
		"build-args",
		"ca-certs",
		"nvim-package",
		"terminal-package",
		"plugins",
		"registry-oci",
		"registry-pypi",
		"registry-npm",
		"registry-go",
		"registry-http",
		"registry-idle-timeout",
	}

	for _, key := range allKeys {
		if err := ds.SetDefault(key, ""); err != nil {
			return fmt.Errorf("failed to clear global %s: %w", key, err)
		}
	}

	return nil
}

// ToYAML serializes a GlobalDefaults resource to YAML.
func (h *GlobalDefaultsHandler) ToYAML(res resource.Resource) ([]byte, error) {
	gdr, ok := res.(*GlobalDefaultsResource)
	if !ok {
		return nil, fmt.Errorf("expected GlobalDefaultsResource, got %T", res)
	}

	doc := globalDefaultsYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       KindGlobalDefaults,
		Metadata:   globalDefaultsMetadata{Name: "global-defaults"},
		Spec: globalDefaultsSpec{
			Theme:               gdr.theme,
			BuildArgs:           gdr.buildArgs,
			CACerts:             gdr.caCerts,
			NvimPackage:         gdr.nvimPackage,
			TerminalPackage:     gdr.terminalPackage,
			Plugins:             gdr.plugins,
			RegistryOCI:         gdr.registryOCI,
			RegistryPyPI:        gdr.registryPyPI,
			RegistryNPM:         gdr.registryNPM,
			RegistryGo:          gdr.registryGo,
			RegistryHTTP:        gdr.registryHTTP,
			RegistryIdleTimeout: gdr.registryIdleTimeout,
		},
	}

	return yaml.Marshal(doc)
}

// loadGlobalDefaults reads all keys from the defaults table into a globalDefaultsSpec.
func loadGlobalDefaults(ds db.DefaultsStore) (*globalDefaultsSpec, error) {
	spec := &globalDefaultsSpec{}

	// Build args (JSON object)
	raw, err := ds.GetDefault("build-args")
	if err != nil {
		return nil, fmt.Errorf("failed to get global build args: %w", err)
	}
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &spec.BuildArgs); err != nil {
			return nil, fmt.Errorf("failed to parse global build args: %w", err)
		}
	}

	// CA certs (JSON array)
	raw, err = ds.GetDefault("ca-certs")
	if err != nil {
		return nil, fmt.Errorf("failed to get global CA certs: %w", err)
	}
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &spec.CACerts); err != nil {
			return nil, fmt.Errorf("failed to parse global CA certs: %w", err)
		}
	}

	// Theme (plain string)
	spec.Theme, err = ds.GetDefault("theme")
	if err != nil {
		return nil, fmt.Errorf("failed to get global theme: %w", err)
	}

	// Nvim package (plain string)
	spec.NvimPackage, err = ds.GetDefault("nvim-package")
	if err != nil {
		return nil, fmt.Errorf("failed to get global nvim-package: %w", err)
	}

	// Terminal package (plain string)
	spec.TerminalPackage, err = ds.GetDefault("terminal-package")
	if err != nil {
		return nil, fmt.Errorf("failed to get global terminal-package: %w", err)
	}

	// Plugins (JSON array)
	raw, err = ds.GetDefault("plugins")
	if err != nil {
		return nil, fmt.Errorf("failed to get global plugins: %w", err)
	}
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &spec.Plugins); err != nil {
			return nil, fmt.Errorf("failed to parse global plugins: %w", err)
		}
	}

	// Registry type defaults (plain strings)
	spec.RegistryOCI, err = ds.GetDefault("registry-oci")
	if err != nil {
		return nil, fmt.Errorf("failed to get global registry-oci: %w", err)
	}
	spec.RegistryPyPI, err = ds.GetDefault("registry-pypi")
	if err != nil {
		return nil, fmt.Errorf("failed to get global registry-pypi: %w", err)
	}
	spec.RegistryNPM, err = ds.GetDefault("registry-npm")
	if err != nil {
		return nil, fmt.Errorf("failed to get global registry-npm: %w", err)
	}
	spec.RegistryGo, err = ds.GetDefault("registry-go")
	if err != nil {
		return nil, fmt.Errorf("failed to get global registry-go: %w", err)
	}
	spec.RegistryHTTP, err = ds.GetDefault("registry-http")
	if err != nil {
		return nil, fmt.Errorf("failed to get global registry-http: %w", err)
	}

	// Registry idle timeout (plain string)
	spec.RegistryIdleTimeout, err = ds.GetDefault("registry-idle-timeout")
	if err != nil {
		return nil, fmt.Errorf("failed to get global registry-idle-timeout: %w", err)
	}

	return spec, nil
}

// ---------------------------------------------------------------------------
// globalDefaultsSpec helpers
// ---------------------------------------------------------------------------

// isEmpty returns true when no defaults are set.
func (s *globalDefaultsSpec) isEmpty() bool {
	return s.Theme == "" &&
		len(s.BuildArgs) == 0 &&
		len(s.CACerts) == 0 &&
		s.NvimPackage == "" &&
		s.TerminalPackage == "" &&
		len(s.Plugins) == 0 &&
		s.RegistryOCI == "" &&
		s.RegistryPyPI == "" &&
		s.RegistryNPM == "" &&
		s.RegistryGo == "" &&
		s.RegistryHTTP == "" &&
		s.RegistryIdleTimeout == ""
}

// ---------------------------------------------------------------------------
// GlobalDefaultsResource implements resource.Resource
// ---------------------------------------------------------------------------

// GlobalDefaultsResource wraps global defaults to implement resource.Resource.
type GlobalDefaultsResource struct {
	theme               string
	buildArgs           map[string]string
	caCerts             []models.CACertConfig
	nvimPackage         string
	terminalPackage     string
	plugins             []string
	registryOCI         string
	registryPyPI        string
	registryNPM         string
	registryGo          string
	registryHTTP        string
	registryIdleTimeout string
}

func (r *GlobalDefaultsResource) GetKind() string {
	return KindGlobalDefaults
}

func (r *GlobalDefaultsResource) GetName() string {
	return "global-defaults"
}

func (r *GlobalDefaultsResource) Validate() error {
	return nil
}

// Theme returns the underlying theme name.
func (r *GlobalDefaultsResource) Theme() string {
	return r.theme
}

// BuildArgs returns the underlying build args map.
func (r *GlobalDefaultsResource) BuildArgs() map[string]string {
	return r.buildArgs
}

// CACerts returns the underlying CA certs slice.
func (r *GlobalDefaultsResource) CACerts() []models.CACertConfig {
	return r.caCerts
}

// newGlobalDefaultsResourceFromSpec creates a GlobalDefaultsResource from a spec.
func newGlobalDefaultsResourceFromSpec(spec *globalDefaultsSpec) *GlobalDefaultsResource {
	return &GlobalDefaultsResource{
		theme:               spec.Theme,
		buildArgs:           spec.BuildArgs,
		caCerts:             spec.CACerts,
		nvimPackage:         spec.NvimPackage,
		terminalPackage:     spec.TerminalPackage,
		plugins:             spec.Plugins,
		registryOCI:         spec.RegistryOCI,
		registryPyPI:        spec.RegistryPyPI,
		registryNPM:         spec.RegistryNPM,
		registryGo:          spec.RegistryGo,
		registryHTTP:        spec.RegistryHTTP,
		registryIdleTimeout: spec.RegistryIdleTimeout,
	}
}

// NewGlobalDefaultsResource creates a new GlobalDefaultsResource.
func NewGlobalDefaultsResource(theme string, buildArgs map[string]string, caCerts []models.CACertConfig) *GlobalDefaultsResource {
	return &GlobalDefaultsResource{
		theme:     theme,
		buildArgs: buildArgs,
		caCerts:   caCerts,
	}
}
