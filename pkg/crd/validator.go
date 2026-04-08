package crd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// SchemaValidator validates resource spec maps against an OpenAPI v3 JSON Schema.
// The typical workflow is to call Compile once with the CRD's schema definition,
// then call Validate for each resource instance that needs to be checked.
//
// The default implementation is DefaultSchemaValidator, backed by the
// santhosh-tekuri/jsonschema library using JSON Schema Draft 2020-12.
// Create one via NewSchemaValidator.
//
// Implementations are NOT safe for concurrent use with Compile; call Compile once
// during setup and then call Validate concurrently if needed.
type SchemaValidator interface {
	// Compile parses and compiles the provided JSON Schema map, preparing it for
	// subsequent calls to Validate. Must be called before Validate.
	//
	// schema is a map representation of an OpenAPI v3 / JSON Schema object.
	// Returns an InvalidSchemaError if the schema cannot be parsed or compiled.
	Compile(schema map[string]interface{}) error

	// Validate checks whether data conforms to the previously compiled schema.
	// Must be called after a successful Compile.
	//
	// Returns nil if data is valid.
	// Returns a SchemaValidationError with field and message details if invalid.
	// Returns an InvalidSchemaError if Compile has not been called yet.
	Validate(data map[string]interface{}) error
}

// DefaultSchemaValidator is the default implementation using jsonschema library
type DefaultSchemaValidator struct {
	schema *jsonschema.Schema
}

// NewSchemaValidator creates a new DefaultSchemaValidator
func NewSchemaValidator() *DefaultSchemaValidator {
	return &DefaultSchemaValidator{}
}

// Compile prepares a schema for validation
func (v *DefaultSchemaValidator) Compile(schema map[string]interface{}) error {
	// Convert schema to JSON string
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return &InvalidSchemaError{
			Kind:    "Unknown",
			Message: fmt.Sprintf("failed to marshal schema: %v", err),
		}
	}

	// Create a new compiler
	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020

	// Add schema to compiler
	if err := compiler.AddResource("schema.json", strings.NewReader(string(schemaBytes))); err != nil {
		return &InvalidSchemaError{
			Kind:    "Unknown",
			Message: fmt.Sprintf("failed to add schema resource: %v", err),
		}
	}

	// Compile the schema
	compiledSchema, err := compiler.Compile("schema.json")
	if err != nil {
		return &InvalidSchemaError{
			Kind:    "Unknown",
			Message: fmt.Sprintf("failed to compile schema: %v", err),
		}
	}

	v.schema = compiledSchema
	return nil
}

// Validate checks if data conforms to the compiled schema
func (v *DefaultSchemaValidator) Validate(data map[string]interface{}) error {
	if v.schema == nil {
		return &InvalidSchemaError{
			Kind:    "Unknown",
			Message: "schema not compiled",
		}
	}

	// Validate the data
	if err := v.schema.Validate(data); err != nil {
		// Extract validation error details
		if valErr, ok := err.(*jsonschema.ValidationError); ok {
			return &SchemaValidationError{
				Kind:    "Unknown",
				Field:   valErr.InstanceLocation,
				Message: valErr.Message,
			}
		}
		return &SchemaValidationError{
			Kind:    "Unknown",
			Field:   "",
			Message: err.Error(),
		}
	}

	return nil
}
