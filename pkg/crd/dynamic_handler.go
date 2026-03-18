package crd

import (
	"fmt"

	"github.com/rmkohlman/MaestroSDK/resource"

	"gopkg.in/yaml.v3"
)

// DynamicHandler handles custom resources defined by CRDs
type DynamicHandler struct {
	resolver        CRDResolver
	schemaValidator SchemaValidator
	scopeValidator  ScopeValidator
	store           CustomResourceStore
}

// CustomResourceStore persists custom resource instances
type CustomResourceStore interface {
	// CreateResource creates a new custom resource instance
	CreateResource(kind string, data map[string]interface{}) error

	// GetResource retrieves a custom resource by kind and name
	GetResource(kind, name string) (map[string]interface{}, error)

	// ListResources returns all resources of a given kind
	ListResources(kind string) ([]map[string]interface{}, error)

	// UpdateResource updates an existing custom resource
	UpdateResource(kind string, data map[string]interface{}) error

	// DeleteResource removes a custom resource
	DeleteResource(kind, name string) error
}

// NewDynamicHandler creates a new DynamicHandler
func NewDynamicHandler(resolver CRDResolver, schemaValidator SchemaValidator, scopeValidator ScopeValidator, store CustomResourceStore) *DynamicHandler {
	return &DynamicHandler{
		resolver:        resolver,
		schemaValidator: schemaValidator,
		scopeValidator:  scopeValidator,
		store:           store,
	}
}

// Kind returns "CustomResource" to indicate this is the dynamic handler
func (h *DynamicHandler) Kind() string {
	return "CustomResource"
}

// customResource is the internal representation of a custom resource
type customResource struct {
	data map[string]interface{}
}

func (r *customResource) GetKind() string {
	if kind, ok := r.data["kind"].(string); ok {
		return kind
	}
	return "CustomResource"
}

func (r *customResource) GetName() string {
	if metadata, ok := r.data["metadata"].(map[string]interface{}); ok {
		if name, ok := metadata["name"].(string); ok {
			return name
		}
	}
	return ""
}

func (r *customResource) Validate() error {
	// Basic validation
	if r.GetKind() == "" {
		return fmt.Errorf("kind is required")
	}
	if r.GetName() == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

// Apply creates or updates a custom resource
func (h *DynamicHandler) Apply(ctx resource.Context, data []byte) (resource.Resource, error) {
	// 1. Parse YAML to extract kind and metadata
	var resourceData map[string]interface{}
	if err := yaml.Unmarshal(data, &resourceData); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Extract kind
	kind, ok := resourceData["kind"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'kind' field")
	}

	// Extract metadata
	metadata, ok := resourceData["metadata"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'metadata' field")
	}

	// Extract name from metadata
	name, ok := metadata["name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'metadata.name' field")
	}

	// 2. Resolve kind using resolver
	crd := h.resolver.Resolve(kind)
	if crd == nil {
		return nil, &UnknownKindError{Kind: kind}
	}

	// 3. Validate scope using scope validator
	if err := h.scopeValidator.Validate(crd.Scope, metadata); err != nil {
		return nil, err
	}

	// 4. Validate spec against schema using schema validator
	spec, ok := resourceData["spec"].(map[string]interface{})
	if !ok {
		// spec is optional, but if present must be an object
		if _, hasSpec := resourceData["spec"]; hasSpec {
			return nil, fmt.Errorf("invalid 'spec' field: must be an object")
		}
		spec = make(map[string]interface{})
	}

	// Get the storage version schema
	var schema map[string]interface{}
	for _, version := range crd.Versions {
		if version.Storage {
			schema = version.Schema.OpenAPIV3Schema
			break
		}
	}

	if schema != nil {
		// Compile and validate
		validator := NewSchemaValidator()
		if err := validator.Compile(schema); err != nil {
			return nil, err
		}
		if err := validator.Validate(spec); err != nil {
			return nil, err
		}
	}

	// 5. Create or update resource in store
	// Check if resource exists
	existing, err := h.store.GetResource(crd.Names.Kind, name)
	if err != nil {
		// Resource doesn't exist - create it
		if err := h.store.CreateResource(crd.Names.Kind, resourceData); err != nil {
			return nil, fmt.Errorf("failed to create resource: %w", err)
		}
	} else {
		// Resource exists - update it
		_ = existing // ignore for now
		if err := h.store.UpdateResource(crd.Names.Kind, resourceData); err != nil {
			return nil, fmt.Errorf("failed to update resource: %w", err)
		}
	}

	return &customResource{data: resourceData}, nil
}

// Get retrieves a custom resource by name
func (h *DynamicHandler) Get(ctx resource.Context, name string) (resource.Resource, error) {
	// Extract kind from context (stored in DataStore field as a map)
	kind, err := h.extractKindFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Resolve the CRD
	crd := h.resolver.Resolve(kind)
	if crd == nil {
		return nil, &UnknownKindError{Kind: kind}
	}

	// Retrieve from store
	data, err := h.store.GetResource(crd.Names.Kind, name)
	if err != nil {
		return nil, err
	}

	return &customResource{data: data}, nil
}

// List returns all custom resources of the resolved kind
func (h *DynamicHandler) List(ctx resource.Context) ([]resource.Resource, error) {
	// Extract kind from context
	kind, err := h.extractKindFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Resolve the CRD
	crd := h.resolver.Resolve(kind)
	if crd == nil {
		return nil, &UnknownKindError{Kind: kind}
	}

	// Get all resources of that kind from store
	resourceDataList, err := h.store.ListResources(crd.Names.Kind)
	if err != nil {
		return nil, err
	}

	// Filter by namespace if provided
	namespace, _ := h.extractNamespaceFromContext(ctx)
	var filtered []resource.Resource
	for _, data := range resourceDataList {
		if namespace != "" {
			// Check if resource has matching namespace
			if metadata, ok := data["metadata"].(map[string]interface{}); ok {
				if ns, ok := metadata["namespace"].(string); ok && ns != namespace {
					continue // Skip non-matching namespace
				}
			}
		}
		filtered = append(filtered, &customResource{data: data})
	}

	return filtered, nil
}

// Delete removes a custom resource
func (h *DynamicHandler) Delete(ctx resource.Context, name string) error {
	// Extract kind from context
	kind, err := h.extractKindFromContext(ctx)
	if err != nil {
		return err
	}

	// Resolve the CRD
	crd := h.resolver.Resolve(kind)
	if crd == nil {
		return &UnknownKindError{Kind: kind}
	}

	// Delete from store
	if err := h.store.DeleteResource(crd.Names.Kind, name); err != nil {
		return err
	}

	return nil
}

// extractKindFromContext extracts the kind from the context
// The kind can be stored in the DataStore field as a map with a "kind" key
func (h *DynamicHandler) extractKindFromContext(ctx resource.Context) (string, error) {
	// Try to extract from DataStore field (which is any type)
	if ctxMap, ok := ctx.DataStore.(map[string]string); ok {
		if kind, exists := ctxMap["kind"]; exists {
			return kind, nil
		}
	}

	// Fallback: return error
	return "", fmt.Errorf("kind not specified in context")
}

// extractNamespaceFromContext extracts the namespace from the context
func (h *DynamicHandler) extractNamespaceFromContext(ctx resource.Context) (string, error) {
	// Try to extract from DataStore field
	if ctxMap, ok := ctx.DataStore.(map[string]string); ok {
		if namespace, exists := ctxMap["namespace"]; exists {
			return namespace, nil
		}
	}

	return "", nil // No namespace filter
}

// ToYAML serializes a custom resource to YAML
func (h *DynamicHandler) ToYAML(res resource.Resource) ([]byte, error) {
	if res == nil {
		return nil, fmt.Errorf("resource cannot be nil")
	}

	// Cast to customResource
	cr, ok := res.(*customResource)
	if !ok {
		return nil, fmt.Errorf("invalid resource type: expected *customResource")
	}

	// Serialize to YAML
	data, err := yaml.Marshal(cr.data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return data, nil
}
