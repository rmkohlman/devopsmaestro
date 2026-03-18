package handlers

import (
	"encoding/json"
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/crd"
	"github.com/rmkohlman/MaestroSDK/resource"

	"gopkg.in/yaml.v3"
)

const KindCRD = "CustomResourceDefinition"

// Built-in kinds that cannot be overridden by CRDs
var builtInKinds = []string{
	"Workspace", "App", "Domain", "Ecosystem",
	"NvimPlugin", "NvimTheme", "NvimPackage",
	"TerminalPrompt", "TerminalPackage", "TerminalPlugin", "TerminalEmulator",
	"Registry", "GitRepo", "Credential", "CustomResourceDefinition",
}

// CRDHandler handles CustomResourceDefinition resources
type CRDHandler struct{}

// NewCRDHandler creates a new CRD handler
func NewCRDHandler() *CRDHandler {
	return &CRDHandler{}
}

// Kind returns "CustomResourceDefinition"
func (h *CRDHandler) Kind() string {
	return KindCRD
}

// Apply creates or updates a CRD from YAML data
func (h *CRDHandler) Apply(ctx resource.Context, data []byte) (resource.Resource, error) {
	// Parse YAML
	var crdYAML models.CRDYAML
	if err := yaml.Unmarshal(data, &crdYAML); err != nil {
		return nil, fmt.Errorf("failed to parse CRD YAML: %w", err)
	}

	// Validate required fields
	if crdYAML.Spec.Names.Kind == "" {
		return nil, &crd.InvalidSchemaError{
			Kind:    "CustomResourceDefinition",
			Message: "spec.names.kind is required",
		}
	}
	if crdYAML.Spec.Names.Singular == "" {
		return nil, &crd.InvalidSchemaError{
			Kind:    "CustomResourceDefinition",
			Message: "spec.names.singular is required",
		}
	}
	if crdYAML.Spec.Names.Plural == "" {
		return nil, &crd.InvalidSchemaError{
			Kind:    "CustomResourceDefinition",
			Message: "spec.names.plural is required",
		}
	}
	if crdYAML.Spec.Scope == "" {
		return nil, &crd.InvalidSchemaError{
			Kind:    "CustomResourceDefinition",
			Message: "spec.scope is required",
		}
	}
	if len(crdYAML.Spec.Versions) == 0 {
		return nil, &crd.InvalidSchemaError{
			Kind:    "CustomResourceDefinition",
			Message: "spec.versions is required and must not be empty",
		}
	}

	// Check for built-in kind collision
	if isBuiltInKind(crdYAML.Spec.Names.Kind) {
		return nil, &crd.InvalidSchemaError{
			Kind:    crdYAML.Spec.Names.Kind,
			Message: fmt.Sprintf("cannot override built-in kind: %s", crdYAML.Spec.Names.Kind),
		}
	}

	// Validate schema for each version
	validator := crd.NewSchemaValidator()
	for _, version := range crdYAML.Spec.Versions {
		if version.Schema != nil {
			if err := validator.Compile(version.Schema); err != nil {
				return nil, err
			}
		}
	}

	// Get DataStore
	ds, err := resource.DataStoreAs[db.CustomResourceStore](ctx)
	if err != nil {
		return nil, err
	}

	// Serialize versions to JSON
	versionsJSON, err := json.Marshal(crdYAML.Spec.Versions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal versions: %w", err)
	}

	// Serialize short names to JSON
	var shortNamesJSON []byte
	if len(crdYAML.Spec.Names.ShortNames) > 0 {
		shortNamesJSON, err = json.Marshal(crdYAML.Spec.Names.ShortNames)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal short names: %w", err)
		}
	}

	// Convert to model
	crdModel := &models.CustomResourceDefinition{
		Kind:     crdYAML.Spec.Names.Kind,
		Group:    crdYAML.Spec.Group,
		Singular: crdYAML.Spec.Names.Singular,
		Plural:   crdYAML.Spec.Names.Plural,
		Scope:    crdYAML.Spec.Scope,
	}

	if len(versionsJSON) > 0 {
		crdModel.Versions.String = string(versionsJSON)
		crdModel.Versions.Valid = true
	}

	if len(shortNamesJSON) > 0 {
		crdModel.ShortNames.String = string(shortNamesJSON)
		crdModel.ShortNames.Valid = true
	}

	// Check if CRD exists
	existing, err := ds.GetCRDByKind(crdModel.Kind)
	if err == nil && existing != nil {
		// Update existing
		crdModel.ID = existing.ID
		crdModel.CreatedAt = existing.CreatedAt
		if err := ds.UpdateCRD(crdModel); err != nil {
			return nil, fmt.Errorf("failed to update CRD: %w", err)
		}
	} else {
		// Create new
		if err := ds.CreateCRD(crdModel); err != nil {
			return nil, fmt.Errorf("failed to create CRD: %w", err)
		}
		// Fetch to get the ID
		crdModel, err = ds.GetCRDByKind(crdModel.Kind)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve created CRD: %w", err)
		}
	}

	return crdModel, nil
}

// Get retrieves a CRD by kind
func (h *CRDHandler) Get(ctx resource.Context, name string) (resource.Resource, error) {
	ds, err := resource.DataStoreAs[db.CustomResourceStore](ctx)
	if err != nil {
		return nil, err
	}

	crdModel, err := ds.GetCRDByKind(name)
	if err != nil {
		return nil, err
	}

	return crdModel, nil
}

// List returns all CRDs
func (h *CRDHandler) List(ctx resource.Context) ([]resource.Resource, error) {
	ds, err := resource.DataStoreAs[db.CustomResourceStore](ctx)
	if err != nil {
		return nil, err
	}

	crds, err := ds.ListCRDs()
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, len(crds))
	for i, crdModel := range crds {
		resources[i] = crdModel
	}

	return resources, nil
}

// Delete removes a CRD
func (h *CRDHandler) Delete(ctx resource.Context, name string) error {
	ds, err := resource.DataStoreAs[db.CustomResourceStore](ctx)
	if err != nil {
		return err
	}

	// Check if any instances of this CRD exist
	instances, err := ds.ListCustomResources(name)
	if err != nil {
		return fmt.Errorf("failed to check for CRD instances: %w", err)
	}

	if len(instances) > 0 {
		return &crd.CRDAlreadyExistsError{
			Kind: fmt.Sprintf("%s (has %d instances)", name, len(instances)),
		}
	}

	return ds.DeleteCRD(name)
}

// ToYAML serializes a CRD to YAML
func (h *CRDHandler) ToYAML(res resource.Resource) ([]byte, error) {
	crdModel, ok := res.(*models.CustomResourceDefinition)
	if !ok {
		return nil, fmt.Errorf("expected *models.CustomResourceDefinition, got %T", res)
	}

	// Parse versions from JSON
	var versions []models.CRDVersion
	if crdModel.Versions.Valid && crdModel.Versions.String != "" {
		if err := json.Unmarshal([]byte(crdModel.Versions.String), &versions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal versions: %w", err)
		}
	}

	// Parse short names from JSON
	var shortNames []string
	if crdModel.ShortNames.Valid && crdModel.ShortNames.String != "" {
		if err := json.Unmarshal([]byte(crdModel.ShortNames.String), &shortNames); err != nil {
			return nil, fmt.Errorf("failed to unmarshal short names: %w", err)
		}
	}

	// Convert to YAML structure
	crdYAML := models.CRDYAML{
		APIVersion: "devopsmaestro.io/v1alpha1",
		Kind:       "CustomResourceDefinition",
		Metadata: models.CRDMetadata{
			Name: fmt.Sprintf("%s.%s", crdModel.Plural, crdModel.Group),
		},
		Spec: models.CRDSpec{
			Group: crdModel.Group,
			Names: models.CRDNames{
				Kind:       crdModel.Kind,
				Singular:   crdModel.Singular,
				Plural:     crdModel.Plural,
				ShortNames: shortNames,
			},
			Scope:    crdModel.Scope,
			Versions: versions,
		},
	}

	return yaml.Marshal(crdYAML)
}

// isBuiltInKind checks if a kind is a built-in type
func isBuiltInKind(kind string) bool {
	for _, builtin := range builtInKinds {
		if builtin == kind {
			return true
		}
	}
	return false
}
