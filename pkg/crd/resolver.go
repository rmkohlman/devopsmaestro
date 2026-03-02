package crd

import (
	"strings"
	"sync"
)

// CRDResolver resolves resource kinds to their CRD definitions.
// It supports resolution by kind, singular, plural, or short names.
type CRDResolver interface {
	// Resolve finds a CRD by any of its names (kind, singular, plural, or short name)
	// Returns nil if not found. Case-insensitive.
	Resolve(name string) *CRDDefinition

	// Register adds a CRD to the resolver
	// Returns error if a CRD with the same kind already exists
	Register(crd *CRDDefinition) error

	// Unregister removes a CRD from the resolver
	Unregister(kind string) error

	// List returns all registered CRDs
	List() []*CRDDefinition

	// Refresh reloads CRDs from the store
	Refresh() error
}

// CRDStore is the persistence layer for CRDs
type CRDStore interface {
	// CreateCRD creates a new CRD
	CreateCRD(crd *CRDDefinition) error

	// GetCRD retrieves a CRD by kind
	GetCRD(kind string) (*CRDDefinition, error)

	// ListCRDs returns all CRDs
	ListCRDs() ([]*CRDDefinition, error)

	// UpdateCRD updates an existing CRD
	UpdateCRD(crd *CRDDefinition) error

	// DeleteCRD removes a CRD
	DeleteCRD(kind string) error
}

// DefaultCRDResolver is the default implementation of CRDResolver
type DefaultCRDResolver struct {
	store   CRDStore
	crds    map[string]*CRDDefinition // kind (lowercase) -> CRD
	aliases map[string]string         // alias (lowercase) -> kind (lowercase)
	mu      sync.RWMutex
}

// NewCRDResolver creates a new DefaultCRDResolver
func NewCRDResolver(store CRDStore) *DefaultCRDResolver {
	return &DefaultCRDResolver{
		store:   store,
		crds:    make(map[string]*CRDDefinition),
		aliases: make(map[string]string),
	}
}

// Resolve finds a CRD by any of its names (case-insensitive)
func (r *DefaultCRDResolver) Resolve(name string) *CRDDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	lowerName := strings.ToLower(name)

	// Check if it's a direct kind match
	if crd, exists := r.crds[lowerName]; exists {
		return crd
	}

	// Check if it's an alias (singular, plural, or short name)
	if kind, exists := r.aliases[lowerName]; exists {
		return r.crds[kind]
	}

	return nil
}

// Register adds a CRD to the resolver
func (r *DefaultCRDResolver) Register(crd *CRDDefinition) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	lowerKind := strings.ToLower(crd.Names.Kind)

	// Check if already exists
	if _, exists := r.crds[lowerKind]; exists {
		return &CRDAlreadyExistsError{Kind: crd.Names.Kind}
	}

	// Store the CRD
	r.crds[lowerKind] = crd

	// Register aliases
	if crd.Names.Singular != "" {
		r.aliases[strings.ToLower(crd.Names.Singular)] = lowerKind
	}
	if crd.Names.Plural != "" {
		r.aliases[strings.ToLower(crd.Names.Plural)] = lowerKind
	}
	for _, short := range crd.Names.ShortNames {
		r.aliases[strings.ToLower(short)] = lowerKind
	}

	return nil
}

// Unregister removes a CRD from the resolver
func (r *DefaultCRDResolver) Unregister(kind string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	lowerKind := strings.ToLower(kind)

	// Check if exists
	crd, exists := r.crds[lowerKind]
	if !exists {
		return &CRDNotFoundError{Kind: kind}
	}

	// Remove aliases
	if crd.Names.Singular != "" {
		delete(r.aliases, strings.ToLower(crd.Names.Singular))
	}
	if crd.Names.Plural != "" {
		delete(r.aliases, strings.ToLower(crd.Names.Plural))
	}
	for _, short := range crd.Names.ShortNames {
		delete(r.aliases, strings.ToLower(short))
	}

	// Remove CRD
	delete(r.crds, lowerKind)

	return nil
}

// List returns all registered CRDs
func (r *DefaultCRDResolver) List() []*CRDDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	crds := make([]*CRDDefinition, 0, len(r.crds))
	for _, crd := range r.crds {
		crds = append(crds, crd)
	}
	return crds
}

// Refresh reloads CRDs from the store
func (r *DefaultCRDResolver) Refresh() error {
	// Load all CRDs from store
	crds, err := r.store.ListCRDs()
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Clear existing data
	r.crds = make(map[string]*CRDDefinition)
	r.aliases = make(map[string]string)

	// Re-register all CRDs
	for _, crd := range crds {
		lowerKind := strings.ToLower(crd.Names.Kind)
		r.crds[lowerKind] = crd

		// Register aliases
		if crd.Names.Singular != "" {
			r.aliases[strings.ToLower(crd.Names.Singular)] = lowerKind
		}
		if crd.Names.Plural != "" {
			r.aliases[strings.ToLower(crd.Names.Plural)] = lowerKind
		}
		for _, short := range crd.Names.ShortNames {
			r.aliases[strings.ToLower(short)] = lowerKind
		}
	}

	return nil
}
