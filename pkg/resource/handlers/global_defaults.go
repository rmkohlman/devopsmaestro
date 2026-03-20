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
	BuildArgs map[string]string     `yaml:"buildArgs,omitempty"`
	CACerts   []models.CACertConfig `yaml:"caCerts,omitempty"`
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

// Apply restores global defaults (build-args and CA-certs) from YAML data.
func (h *GlobalDefaultsHandler) Apply(ctx resource.Context, data []byte) (resource.Resource, error) {
	var doc globalDefaultsYAML
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse GlobalDefaults YAML: %w", err)
	}

	ds, err := resource.DataStoreAs[db.DefaultsStore](ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DefaultStore: %w", err)
	}

	// Restore build args
	if len(doc.Spec.BuildArgs) > 0 {
		b, err := json.Marshal(doc.Spec.BuildArgs)
		if err != nil {
			return nil, fmt.Errorf("failed to encode build args: %w", err)
		}
		if err := ds.SetDefault("build-args", string(b)); err != nil {
			return nil, fmt.Errorf("failed to set global build args: %w", err)
		}
	}

	// Restore CA certs
	if len(doc.Spec.CACerts) > 0 {
		b, err := json.Marshal(doc.Spec.CACerts)
		if err != nil {
			return nil, fmt.Errorf("failed to encode CA certs: %w", err)
		}
		if err := ds.SetDefault("ca-certs", string(b)); err != nil {
			return nil, fmt.Errorf("failed to set global CA certs: %w", err)
		}
	}

	return &GlobalDefaultsResource{
		buildArgs: doc.Spec.BuildArgs,
		caCerts:   doc.Spec.CACerts,
	}, nil
}

// Get retrieves global defaults. Name is ignored (there's only one).
func (h *GlobalDefaultsHandler) Get(ctx resource.Context, name string) (resource.Resource, error) {
	ds, err := resource.DataStoreAs[db.DefaultsStore](ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DefaultStore: %w", err)
	}

	buildArgs, caCerts, err := loadGlobalDefaults(ds)
	if err != nil {
		return nil, err
	}

	return &GlobalDefaultsResource{
		buildArgs: buildArgs,
		caCerts:   caCerts,
	}, nil
}

// List returns a single GlobalDefaults resource if any defaults exist.
func (h *GlobalDefaultsHandler) List(ctx resource.Context) ([]resource.Resource, error) {
	ds, err := resource.DataStoreAs[db.DefaultsStore](ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DefaultStore: %w", err)
	}

	buildArgs, caCerts, err := loadGlobalDefaults(ds)
	if err != nil {
		return nil, err
	}

	// Only include if there's actually something to export
	if len(buildArgs) == 0 && len(caCerts) == 0 {
		return nil, nil
	}

	return []resource.Resource{
		&GlobalDefaultsResource{
			buildArgs: buildArgs,
			caCerts:   caCerts,
		},
	}, nil
}

// Delete clears all global defaults. Name is ignored.
func (h *GlobalDefaultsHandler) Delete(ctx resource.Context, name string) error {
	ds, err := resource.DataStoreAs[db.DefaultsStore](ctx)
	if err != nil {
		return fmt.Errorf("failed to get DefaultStore: %w", err)
	}

	if err := ds.SetDefault("build-args", ""); err != nil {
		return fmt.Errorf("failed to clear global build args: %w", err)
	}
	if err := ds.SetDefault("ca-certs", ""); err != nil {
		return fmt.Errorf("failed to clear global CA certs: %w", err)
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
			BuildArgs: gdr.buildArgs,
			CACerts:   gdr.caCerts,
		},
	}

	return yaml.Marshal(doc)
}

// loadGlobalDefaults reads build-args and ca-certs from the defaults table.
func loadGlobalDefaults(ds db.DefaultsStore) (map[string]string, []models.CACertConfig, error) {
	var buildArgs map[string]string
	var caCerts []models.CACertConfig

	raw, err := ds.GetDefault("build-args")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get global build args: %w", err)
	}
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &buildArgs); err != nil {
			return nil, nil, fmt.Errorf("failed to parse global build args: %w", err)
		}
	}

	raw, err = ds.GetDefault("ca-certs")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get global CA certs: %w", err)
	}
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &caCerts); err != nil {
			return nil, nil, fmt.Errorf("failed to parse global CA certs: %w", err)
		}
	}

	return buildArgs, caCerts, nil
}

// ---------------------------------------------------------------------------
// GlobalDefaultsResource implements resource.Resource
// ---------------------------------------------------------------------------

// GlobalDefaultsResource wraps global defaults to implement resource.Resource.
type GlobalDefaultsResource struct {
	buildArgs map[string]string
	caCerts   []models.CACertConfig
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

// BuildArgs returns the underlying build args map.
func (r *GlobalDefaultsResource) BuildArgs() map[string]string {
	return r.buildArgs
}

// CACerts returns the underlying CA certs slice.
func (r *GlobalDefaultsResource) CACerts() []models.CACertConfig {
	return r.caCerts
}

// NewGlobalDefaultsResource creates a new GlobalDefaultsResource.
func NewGlobalDefaultsResource(buildArgs map[string]string, caCerts []models.CACertConfig) *GlobalDefaultsResource {
	return &GlobalDefaultsResource{
		buildArgs: buildArgs,
		caCerts:   caCerts,
	}
}
