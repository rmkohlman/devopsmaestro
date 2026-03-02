package crd

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// UnknownKindError Tests
// =============================================================================

func TestUnknownKindError_Error(t *testing.T) {
	err := &UnknownKindError{Kind: "Database"}
	assert.Equal(t, "unknown kind: Database", err.Error())
}

func TestUnknownKindError_IsCRDError(t *testing.T) {
	err := &UnknownKindError{Kind: "Database"}
	assert.True(t, err.IsCRDError())
}

func TestUnknownKindError_IsError(t *testing.T) {
	var err error = &UnknownKindError{Kind: "Database"}
	assert.NotNil(t, err)
}

// =============================================================================
// SchemaValidationError Tests
// =============================================================================

func TestSchemaValidationError_Error(t *testing.T) {
	err := &SchemaValidationError{
		Kind:    "Database",
		Field:   "spec.engine",
		Message: "must be one of: postgres, mysql, sqlite",
	}
	expected := "schema validation failed for Database.spec.engine: must be one of: postgres, mysql, sqlite"
	assert.Equal(t, expected, err.Error())
}

func TestSchemaValidationError_IsCRDError(t *testing.T) {
	err := &SchemaValidationError{Kind: "Database", Field: "spec.engine", Message: "invalid"}
	assert.True(t, err.IsCRDError())
}

// =============================================================================
// ScopeValidationError Tests
// =============================================================================

func TestScopeValidationError_Error(t *testing.T) {
	err := &ScopeValidationError{
		Kind:     "Database",
		Scope:    "Workspace",
		Resource: "my-db",
		Message:  "workspace is required for Workspace-scoped resources",
	}
	expected := "scope validation failed for my-db (Workspace scope): workspace is required for Workspace-scoped resources"
	assert.Equal(t, expected, err.Error())
}

func TestScopeValidationError_IsCRDError(t *testing.T) {
	err := &ScopeValidationError{Kind: "Database", Scope: "Workspace", Resource: "my-db", Message: "invalid"}
	assert.True(t, err.IsCRDError())
}

// =============================================================================
// CRDNotFoundError Tests
// =============================================================================

func TestCRDNotFoundError_Error(t *testing.T) {
	err := &CRDNotFoundError{Kind: "Database"}
	assert.Equal(t, "CRD not found: Database", err.Error())
}

func TestCRDNotFoundError_IsCRDError(t *testing.T) {
	err := &CRDNotFoundError{Kind: "Database"}
	assert.True(t, err.IsCRDError())
}

// =============================================================================
// CRDAlreadyExistsError Tests
// =============================================================================

func TestCRDAlreadyExistsError_Error(t *testing.T) {
	err := &CRDAlreadyExistsError{Kind: "Database"}
	assert.Equal(t, "CRD already exists: Database", err.Error())
}

func TestCRDAlreadyExistsError_IsCRDError(t *testing.T) {
	err := &CRDAlreadyExistsError{Kind: "Database"}
	assert.True(t, err.IsCRDError())
}

// =============================================================================
// InvalidSchemaError Tests
// =============================================================================

func TestInvalidSchemaError_Error(t *testing.T) {
	err := &InvalidSchemaError{
		Kind:    "Database",
		Message: "schema must have 'type' field",
	}
	expected := "invalid schema for Database: schema must have 'type' field"
	assert.Equal(t, expected, err.Error())
}

func TestInvalidSchemaError_IsCRDError(t *testing.T) {
	err := &InvalidSchemaError{Kind: "Database", Message: "invalid"}
	assert.True(t, err.IsCRDError())
}

// =============================================================================
// Error Type Assertions
// =============================================================================

func TestCRDErrors_AreErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"UnknownKindError", &UnknownKindError{Kind: "Test"}},
		{"SchemaValidationError", &SchemaValidationError{Kind: "Test", Field: "test", Message: "test"}},
		{"ScopeValidationError", &ScopeValidationError{Kind: "Test", Scope: "Global", Resource: "test", Message: "test"}},
		{"CRDNotFoundError", &CRDNotFoundError{Kind: "Test"}},
		{"CRDAlreadyExistsError", &CRDAlreadyExistsError{Kind: "Test"}},
		{"InvalidSchemaError", &InvalidSchemaError{Kind: "Test", Message: "test"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.err)
			assert.NotEmpty(t, tt.err.Error())
		})
	}
}

func TestCRDErrors_ImplementCRDError(t *testing.T) {
	tests := []struct {
		name string
		err  CRDError
	}{
		{"UnknownKindError", &UnknownKindError{Kind: "Test"}},
		{"SchemaValidationError", &SchemaValidationError{Kind: "Test", Field: "test", Message: "test"}},
		{"ScopeValidationError", &ScopeValidationError{Kind: "Test", Scope: "Global", Resource: "test", Message: "test"}},
		{"CRDNotFoundError", &CRDNotFoundError{Kind: "Test"}},
		{"CRDAlreadyExistsError", &CRDAlreadyExistsError{Kind: "Test"}},
		{"InvalidSchemaError", &InvalidSchemaError{Kind: "Test", Message: "test"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, tt.err.IsCRDError())
		})
	}
}

// =============================================================================
// Error Wrapping Tests
// =============================================================================

func TestCRDErrors_Unwrap(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		target   error
		expected bool
	}{
		{
			name:     "UnknownKindError with errors.Is",
			err:      &UnknownKindError{Kind: "Database"},
			target:   &UnknownKindError{Kind: "Database"},
			expected: false, // Different instances
		},
		{
			name:     "SchemaValidationError with errors.Is",
			err:      &SchemaValidationError{Kind: "Database", Field: "test", Message: "test"},
			target:   &SchemaValidationError{},
			expected: false, // Different instances
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errors.Is(tt.err, tt.target)
			assert.Equal(t, tt.expected, result)
		})
	}
}
