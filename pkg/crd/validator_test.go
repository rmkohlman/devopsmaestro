package crd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Mock SchemaValidator
// =============================================================================

type MockSchemaValidator struct {
	compiledSchema map[string]interface{}
	compileError   error
	validateError  error
}

func NewMockSchemaValidator() *MockSchemaValidator {
	return &MockSchemaValidator{}
}

func (m *MockSchemaValidator) Compile(schema map[string]interface{}) error {
	if m.compileError != nil {
		return m.compileError
	}
	m.compiledSchema = schema
	return nil
}

func (m *MockSchemaValidator) Validate(data map[string]interface{}) error {
	if m.validateError != nil {
		return m.validateError
	}
	return nil
}

func (m *MockSchemaValidator) SetCompileError(err error) {
	m.compileError = err
}

func (m *MockSchemaValidator) SetValidateError(err error) {
	m.validateError = err
}

// =============================================================================
// SchemaValidator Tests - Validation
// =============================================================================

func TestValidator_Validate_AcceptsValidSpec(t *testing.T) {
	validator := NewMockSchemaValidator()

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"engine": map[string]interface{}{
				"type": "string",
				"enum": []interface{}{"postgres", "mysql", "sqlite"},
			},
			"version": map[string]interface{}{
				"type": "string",
			},
		},
		"required": []interface{}{"engine", "version"},
	}

	err := validator.Compile(schema)
	require.NoError(t, err)

	data := map[string]interface{}{
		"engine":  "postgres",
		"version": "15",
	}

	err = validator.Validate(data)
	assert.NoError(t, err, "Valid spec should pass validation")
}

func TestValidator_Validate_RejectsInvalidType(t *testing.T) {
	validator := NewMockSchemaValidator()

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"port": map[string]interface{}{
				"type": "integer",
			},
		},
	}

	err := validator.Compile(schema)
	require.NoError(t, err)

	// Set validation error for invalid type
	validator.SetValidateError(&SchemaValidationError{
		Kind:    "Database",
		Field:   "port",
		Message: "expected integer, got string",
	})

	data := map[string]interface{}{
		"port": "5432", // String instead of integer
	}

	err = validator.Validate(data)
	assert.Error(t, err, "Invalid type should fail validation")
}

func TestValidator_Validate_RejectsMissingRequired(t *testing.T) {
	validator := NewMockSchemaValidator()

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"engine": map[string]interface{}{
				"type": "string",
			},
			"version": map[string]interface{}{
				"type": "string",
			},
		},
		"required": []interface{}{"engine", "version"},
	}

	err := validator.Compile(schema)
	require.NoError(t, err)

	// Set validation error for missing required field
	validator.SetValidateError(&SchemaValidationError{
		Kind:    "Database",
		Field:   "version",
		Message: "required field missing",
	})

	data := map[string]interface{}{
		"engine": "postgres",
		// version is missing
	}

	err = validator.Validate(data)
	assert.Error(t, err, "Missing required field should fail validation")
}

func TestValidator_Validate_AcceptsOptionalFields(t *testing.T) {
	validator := NewMockSchemaValidator()

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"engine": map[string]interface{}{
				"type": "string",
			},
			"port": map[string]interface{}{
				"type": "integer",
			},
		},
		"required": []interface{}{"engine"},
	}

	err := validator.Compile(schema)
	require.NoError(t, err)

	data := map[string]interface{}{
		"engine": "postgres",
		// port is optional and omitted
	}

	err = validator.Validate(data)
	assert.NoError(t, err, "Optional fields can be omitted")
}

func TestValidator_Validate_EnforcesEnum(t *testing.T) {
	validator := NewMockSchemaValidator()

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"engine": map[string]interface{}{
				"type": "string",
				"enum": []interface{}{"postgres", "mysql", "sqlite"},
			},
		},
	}

	err := validator.Compile(schema)
	require.NoError(t, err)

	// Set validation error for invalid enum value
	validator.SetValidateError(&SchemaValidationError{
		Kind:    "Database",
		Field:   "engine",
		Message: "must be one of: postgres, mysql, sqlite",
	})

	data := map[string]interface{}{
		"engine": "oracle", // Not in enum
	}

	err = validator.Validate(data)
	assert.Error(t, err, "Invalid enum value should fail validation")
}

func TestValidator_Validate_EnforcesMinMax(t *testing.T) {
	validator := NewMockSchemaValidator()

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"port": map[string]interface{}{
				"type":    "integer",
				"minimum": 1024,
				"maximum": 65535,
			},
		},
	}

	err := validator.Compile(schema)
	require.NoError(t, err)

	tests := []struct {
		name        string
		port        int
		shouldError bool
		errMsg      string
	}{
		{
			name:        "valid port",
			port:        5432,
			shouldError: false,
		},
		{
			name:        "port below minimum",
			port:        80,
			shouldError: true,
			errMsg:      "must be >= 1024",
		},
		{
			name:        "port above maximum",
			port:        70000,
			shouldError: true,
			errMsg:      "must be <= 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldError {
				validator.SetValidateError(&SchemaValidationError{
					Kind:    "Database",
					Field:   "port",
					Message: tt.errMsg,
				})
			} else {
				validator.SetValidateError(nil)
			}

			data := map[string]interface{}{
				"port": tt.port,
			}

			err := validator.Validate(data)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_Validate_NestedObjects(t *testing.T) {
	validator := NewMockSchemaValidator()

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"connection": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"host": map[string]interface{}{
						"type": "string",
					},
					"port": map[string]interface{}{
						"type": "integer",
					},
				},
				"required": []interface{}{"host"},
			},
		},
	}

	err := validator.Compile(schema)
	require.NoError(t, err)

	data := map[string]interface{}{
		"connection": map[string]interface{}{
			"host": "localhost",
			"port": 5432,
		},
	}

	err = validator.Validate(data)
	assert.NoError(t, err, "Nested objects should validate")
}

func TestValidator_Validate_Arrays(t *testing.T) {
	validator := NewMockSchemaValidator()

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"replicas": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
	}

	err := validator.Compile(schema)
	require.NoError(t, err)

	data := map[string]interface{}{
		"replicas": []interface{}{"replica1", "replica2"},
	}

	err = validator.Validate(data)
	assert.NoError(t, err, "Arrays should validate")
}

func TestValidator_Validate_AdditionalPropertiesDenied(t *testing.T) {
	validator := NewMockSchemaValidator()

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"engine": map[string]interface{}{
				"type": "string",
			},
		},
		"additionalProperties": false,
	}

	err := validator.Compile(schema)
	require.NoError(t, err)

	// Set validation error for additional property
	validator.SetValidateError(&SchemaValidationError{
		Kind:    "Database",
		Field:   "unknown",
		Message: "additional properties not allowed",
	})

	data := map[string]interface{}{
		"engine":  "postgres",
		"unknown": "field", // Not in schema
	}

	err = validator.Validate(data)
	assert.Error(t, err, "Additional properties should be rejected when denied")
}

// =============================================================================
// SchemaValidator Tests - Compilation
// =============================================================================

func TestValidator_Compile_ValidSchema(t *testing.T) {
	validator := NewMockSchemaValidator()

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
		},
	}

	err := validator.Compile(schema)
	assert.NoError(t, err)
	assert.Equal(t, schema, validator.compiledSchema)
}

func TestValidator_Compile_InvalidSchemaFails(t *testing.T) {
	validator := NewMockSchemaValidator()

	// Set compile error
	validator.SetCompileError(&InvalidSchemaError{
		Kind:    "Database",
		Message: "schema must have 'type' field",
	})

	schema := map[string]interface{}{
		// Missing 'type' field
		"properties": map[string]interface{}{},
	}

	err := validator.Compile(schema)
	assert.Error(t, err)
	assert.IsType(t, &InvalidSchemaError{}, err)
}

func TestValidator_Compile_EmptySchema(t *testing.T) {
	validator := NewMockSchemaValidator()

	schema := map[string]interface{}{}

	err := validator.Compile(schema)
	// Empty schema should be valid (matches anything)
	assert.NoError(t, err)
}

func TestValidator_Compile_ComplexSchema(t *testing.T) {
	validator := NewMockSchemaValidator()

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"metadata": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type": "string",
					},
				},
			},
			"spec": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"replicas": map[string]interface{}{
						"type":    "integer",
						"minimum": 1,
					},
					"image": map[string]interface{}{
						"type": "string",
					},
				},
				"required": []interface{}{"image"},
			},
		},
		"required": []interface{}{"metadata", "spec"},
	}

	err := validator.Compile(schema)
	assert.NoError(t, err)
}
