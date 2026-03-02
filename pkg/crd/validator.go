package crd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// SchemaValidator validates resource specs against OpenAPI v3 schemas
type SchemaValidator interface {
	// Compile prepares a schema for validation
	// Returns error if schema is invalid
	Compile(schema map[string]interface{}) error

	// Validate checks if data conforms to the compiled schema
	// Returns error with validation details if invalid
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
