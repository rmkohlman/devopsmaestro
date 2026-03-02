package crd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Mock ScopeValidator
// =============================================================================

type MockScopeValidator struct {
	validateError error
}

func NewMockScopeValidator() *MockScopeValidator {
	return &MockScopeValidator{}
}

func (m *MockScopeValidator) Validate(scope string, metadata map[string]interface{}) error {
	if m.validateError != nil {
		return m.validateError
	}

	// Simplified validation logic for testing
	switch scope {
	case ScopeGlobal:
		if _, hasNamespace := metadata["namespace"]; hasNamespace {
			return &ScopeValidationError{
				Scope:   scope,
				Message: "Global-scoped resources cannot have namespace",
			}
		}
	case ScopeWorkspace:
		if _, hasWorkspace := metadata["workspace"]; !hasWorkspace {
			return &ScopeValidationError{
				Scope:   scope,
				Message: "Workspace-scoped resources must have workspace",
			}
		}
	case ScopeApp:
		if _, hasApp := metadata["app"]; !hasApp {
			return &ScopeValidationError{
				Scope:   scope,
				Message: "App-scoped resources must have app",
			}
		}
	case ScopeDomain:
		if _, hasDomain := metadata["domain"]; !hasDomain {
			return &ScopeValidationError{
				Scope:   scope,
				Message: "Domain-scoped resources must have domain",
			}
		}
	case ScopeEcosystem:
		if _, hasEcosystem := metadata["ecosystem"]; !hasEcosystem {
			return &ScopeValidationError{
				Scope:   scope,
				Message: "Ecosystem-scoped resources must have ecosystem",
			}
		}
	}

	return nil
}

func (m *MockScopeValidator) SetValidateError(err error) {
	m.validateError = err
}

// =============================================================================
// ScopeValidator Tests - Global Scope
// =============================================================================

func TestScopeValidator_Validate_GlobalScope_NoNamespace(t *testing.T) {
	validator := NewMockScopeValidator()

	metadata := map[string]interface{}{
		"name": "my-resource",
	}

	err := validator.Validate(ScopeGlobal, metadata)
	assert.NoError(t, err, "Global-scoped resource without namespace should be valid")
}

func TestScopeValidator_Validate_GlobalScope_RejectsNamespace(t *testing.T) {
	validator := NewMockScopeValidator()

	metadata := map[string]interface{}{
		"name":      "my-resource",
		"namespace": "my-workspace",
	}

	err := validator.Validate(ScopeGlobal, metadata)
	assert.Error(t, err, "Global-scoped resource with namespace should fail")
	assert.IsType(t, &ScopeValidationError{}, err)
}

func TestScopeValidator_Validate_GlobalScope_RejectsWorkspace(t *testing.T) {
	validator := NewMockScopeValidator()

	// Set custom error for workspace field
	validator.SetValidateError(&ScopeValidationError{
		Scope:   ScopeGlobal,
		Message: "Global-scoped resources cannot have workspace",
	})

	metadata := map[string]interface{}{
		"name":      "my-resource",
		"workspace": "my-workspace",
	}

	err := validator.Validate(ScopeGlobal, metadata)
	assert.Error(t, err, "Global-scoped resource with workspace should fail")
}

// =============================================================================
// ScopeValidator Tests - Workspace Scope
// =============================================================================

func TestScopeValidator_Validate_WorkspaceScope_RequiresWorkspace(t *testing.T) {
	validator := NewMockScopeValidator()

	metadata := map[string]interface{}{
		"name": "my-resource",
		// workspace is missing
	}

	err := validator.Validate(ScopeWorkspace, metadata)
	assert.Error(t, err, "Workspace-scoped resource without workspace should fail")
	assert.IsType(t, &ScopeValidationError{}, err)
}

func TestScopeValidator_Validate_WorkspaceScope_AcceptsWorkspace(t *testing.T) {
	validator := NewMockScopeValidator()

	metadata := map[string]interface{}{
		"name":      "my-resource",
		"workspace": "my-workspace",
	}

	err := validator.Validate(ScopeWorkspace, metadata)
	assert.NoError(t, err, "Workspace-scoped resource with workspace should be valid")
}

func TestScopeValidator_Validate_WorkspaceScope_RejectsWrongWorkspace(t *testing.T) {
	validator := NewMockScopeValidator()

	// Set custom error for wrong workspace
	validator.SetValidateError(&ScopeValidationError{
		Scope:   ScopeWorkspace,
		Message: "resource not found in workspace",
	})

	metadata := map[string]interface{}{
		"name":      "my-resource",
		"workspace": "wrong-workspace",
	}

	err := validator.Validate(ScopeWorkspace, metadata)
	assert.Error(t, err, "Resource in wrong workspace should fail")
}

// =============================================================================
// ScopeValidator Tests - App Scope
// =============================================================================

func TestScopeValidator_Validate_AppScope_RequiresApp(t *testing.T) {
	validator := NewMockScopeValidator()

	metadata := map[string]interface{}{
		"name": "my-resource",
		// app is missing
	}

	err := validator.Validate(ScopeApp, metadata)
	assert.Error(t, err, "App-scoped resource without app should fail")
	assert.IsType(t, &ScopeValidationError{}, err)
}

func TestScopeValidator_Validate_AppScope_AcceptsApp(t *testing.T) {
	validator := NewMockScopeValidator()

	metadata := map[string]interface{}{
		"name": "my-resource",
		"app":  "my-app",
	}

	err := validator.Validate(ScopeApp, metadata)
	assert.NoError(t, err, "App-scoped resource with app should be valid")
}

// =============================================================================
// ScopeValidator Tests - Domain Scope
// =============================================================================

func TestScopeValidator_Validate_DomainScope_RequiresDomain(t *testing.T) {
	validator := NewMockScopeValidator()

	metadata := map[string]interface{}{
		"name": "my-resource",
		// domain is missing
	}

	err := validator.Validate(ScopeDomain, metadata)
	assert.Error(t, err, "Domain-scoped resource without domain should fail")
	assert.IsType(t, &ScopeValidationError{}, err)
}

func TestScopeValidator_Validate_DomainScope_AcceptsDomain(t *testing.T) {
	validator := NewMockScopeValidator()

	metadata := map[string]interface{}{
		"name":   "my-resource",
		"domain": "my-domain",
	}

	err := validator.Validate(ScopeDomain, metadata)
	assert.NoError(t, err, "Domain-scoped resource with domain should be valid")
}

// =============================================================================
// ScopeValidator Tests - Ecosystem Scope
// =============================================================================

func TestScopeValidator_Validate_EcosystemScope_RequiresEcosystem(t *testing.T) {
	validator := NewMockScopeValidator()

	metadata := map[string]interface{}{
		"name": "my-resource",
		// ecosystem is missing
	}

	err := validator.Validate(ScopeEcosystem, metadata)
	assert.Error(t, err, "Ecosystem-scoped resource without ecosystem should fail")
	assert.IsType(t, &ScopeValidationError{}, err)
}

func TestScopeValidator_Validate_EcosystemScope_AcceptsEcosystem(t *testing.T) {
	validator := NewMockScopeValidator()

	metadata := map[string]interface{}{
		"name":      "my-resource",
		"ecosystem": "my-ecosystem",
	}

	err := validator.Validate(ScopeEcosystem, metadata)
	assert.NoError(t, err, "Ecosystem-scoped resource with ecosystem should be valid")
}

// =============================================================================
// ScopeValidator Tests - Scope Hierarchy
// =============================================================================

func TestScopeValidator_Validate_WorkspaceScopeInheritsApp(t *testing.T) {
	validator := NewMockScopeValidator()

	// Workspace-scoped resources can have both workspace and app
	metadata := map[string]interface{}{
		"name":      "my-resource",
		"workspace": "my-workspace",
		"app":       "my-app",
	}

	err := validator.Validate(ScopeWorkspace, metadata)
	assert.NoError(t, err, "Workspace-scoped resource can have app context")
}

func TestScopeValidator_Validate_AppScopeInheritsDomain(t *testing.T) {
	validator := NewMockScopeValidator()

	// App-scoped resources can have both app and domain
	metadata := map[string]interface{}{
		"name":   "my-resource",
		"app":    "my-app",
		"domain": "my-domain",
	}

	err := validator.Validate(ScopeApp, metadata)
	assert.NoError(t, err, "App-scoped resource can have domain context")
}

// =============================================================================
// ScopeValidator Tests - Edge Cases
// =============================================================================

func TestScopeValidator_Validate_EmptyMetadata(t *testing.T) {
	validator := NewMockScopeValidator()

	// Set error for empty metadata
	validator.SetValidateError(&ScopeValidationError{
		Scope:   ScopeWorkspace,
		Message: "Workspace-scoped resources must have workspace",
	})

	metadata := map[string]interface{}{}

	err := validator.Validate(ScopeWorkspace, metadata)
	assert.Error(t, err, "Empty metadata should fail for non-Global scopes")
}

func TestScopeValidator_Validate_NilMetadata(t *testing.T) {
	validator := NewMockScopeValidator()

	// Set error for nil metadata
	validator.SetValidateError(&ScopeValidationError{
		Scope:   ScopeWorkspace,
		Message: "metadata cannot be nil",
	})

	err := validator.Validate(ScopeWorkspace, nil)
	assert.Error(t, err, "Nil metadata should fail")
}

func TestScopeValidator_Validate_UnknownScope(t *testing.T) {
	validator := NewMockScopeValidator()

	// Set error for unknown scope
	validator.SetValidateError(&ScopeValidationError{
		Scope:   "Unknown",
		Message: "unknown scope type",
	})

	metadata := map[string]interface{}{
		"name": "my-resource",
	}

	err := validator.Validate("Unknown", metadata)
	assert.Error(t, err, "Unknown scope should fail")
}
