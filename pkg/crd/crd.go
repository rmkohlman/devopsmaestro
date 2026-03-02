// Package crd provides Custom Resource Definition (CRD) support for DevOpsMaestro.
// This allows users to define and manage custom resource types beyond the built-in kinds.
package crd

import "time"

// CRDDefinition represents a Custom Resource Definition.
// This defines a new resource type that users can create instances of.
type CRDDefinition struct {
	// Group is the API group (e.g., "devopsmaestro.io")
	Group string

	// Names defines the different ways to reference this CRD
	Names CRDNames

	// Scope defines where instances can be created (Global, Workspace, App, Domain, Ecosystem)
	Scope string

	// Versions is the list of API versions for this CRD
	Versions []CRDVersion

	// CreatedAt is when this CRD was registered
	CreatedAt time.Time

	// UpdatedAt is when this CRD was last modified
	UpdatedAt time.Time
}

// CRDNames defines the naming options for a CRD
type CRDNames struct {
	// Kind is the primary type name (e.g., "Database")
	Kind string

	// Singular is the singular name (e.g., "database")
	Singular string

	// Plural is the plural name (e.g., "databases")
	Plural string

	// ShortNames are abbreviated forms (e.g., ["db"])
	ShortNames []string
}

// CRDVersion represents a version of a CRD schema
type CRDVersion struct {
	// Name is the version identifier (e.g., "v1alpha1")
	Name string

	// Served indicates if this version is actively served
	Served bool

	// Storage indicates if this is the storage version
	Storage bool

	// Schema is the OpenAPI v3 schema for validation
	Schema CRDSchema
}

// CRDSchema contains the OpenAPI v3 schema definition
type CRDSchema struct {
	// OpenAPIV3Schema is the validation schema
	OpenAPIV3Schema map[string]interface{}
}
