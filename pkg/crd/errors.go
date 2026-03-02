package crd

import "fmt"

// CRDError is the base interface for all CRD errors
type CRDError interface {
	error
	IsCRDError() bool
}

// UnknownKindError is returned when a resource kind cannot be resolved
type UnknownKindError struct {
	Kind string
}

func (e *UnknownKindError) Error() string {
	return fmt.Sprintf("unknown kind: %s", e.Kind)
}

func (e *UnknownKindError) IsCRDError() bool {
	return true
}

// SchemaValidationError is returned when a resource fails schema validation
type SchemaValidationError struct {
	Kind    string
	Field   string
	Message string
}

func (e *SchemaValidationError) Error() string {
	return fmt.Sprintf("schema validation failed for %s.%s: %s", e.Kind, e.Field, e.Message)
}

func (e *SchemaValidationError) IsCRDError() bool {
	return true
}

// ScopeValidationError is returned when a resource violates scope constraints
type ScopeValidationError struct {
	Kind     string
	Scope    string
	Resource string
	Message  string
}

func (e *ScopeValidationError) Error() string {
	return fmt.Sprintf("scope validation failed for %s (%s scope): %s", e.Resource, e.Scope, e.Message)
}

func (e *ScopeValidationError) IsCRDError() bool {
	return true
}

// CRDNotFoundError is returned when a CRD cannot be found
type CRDNotFoundError struct {
	Kind string
}

func (e *CRDNotFoundError) Error() string {
	return fmt.Sprintf("CRD not found: %s", e.Kind)
}

func (e *CRDNotFoundError) IsCRDError() bool {
	return true
}

// CRDAlreadyExistsError is returned when trying to create a duplicate CRD
type CRDAlreadyExistsError struct {
	Kind string
}

func (e *CRDAlreadyExistsError) Error() string {
	return fmt.Sprintf("CRD already exists: %s", e.Kind)
}

func (e *CRDAlreadyExistsError) IsCRDError() bool {
	return true
}

// InvalidSchemaError is returned when a CRD schema is invalid
type InvalidSchemaError struct {
	Kind    string
	Message string
}

func (e *InvalidSchemaError) Error() string {
	return fmt.Sprintf("invalid schema for %s: %s", e.Kind, e.Message)
}

func (e *InvalidSchemaError) IsCRDError() bool {
	return true
}
