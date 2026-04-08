package crd

// ScopeValidator validates that a resource's metadata satisfies the constraints
// imposed by its declared scope. Supported scopes are defined as constants in this
// package: ScopeGlobal, ScopeEcosystem, ScopeDomain, ScopeApp, ScopeWorkspace.
//
// The default implementation is DefaultScopeValidator, created via NewScopeValidator.
// Inject this interface when writing unit tests that need to override scope validation.
type ScopeValidator interface {
	// Validate checks whether the given metadata map satisfies the scope's required
	// and prohibited fields.
	//
	// scope must be one of the ScopeGlobal, ScopeEcosystem, ScopeDomain, ScopeApp,
	// or ScopeWorkspace constants. An unknown scope returns a ScopeValidationError.
	//
	// metadata must not be nil. Missing required keys or presence of prohibited keys
	// both return a ScopeValidationError with a descriptive message.
	Validate(scope string, metadata map[string]interface{}) error
}

// Supported scope types
const (
	ScopeGlobal    = "Global"
	ScopeEcosystem = "Ecosystem"
	ScopeDomain    = "Domain"
	ScopeApp       = "App"
	ScopeWorkspace = "Workspace"
)

// DefaultScopeValidator is the default implementation of ScopeValidator
type DefaultScopeValidator struct{}

// NewScopeValidator creates a new DefaultScopeValidator
func NewScopeValidator() *DefaultScopeValidator {
	return &DefaultScopeValidator{}
}

// Validate checks if a resource is valid for the given scope
func (v *DefaultScopeValidator) Validate(scope string, metadata map[string]interface{}) error {
	if metadata == nil {
		return &ScopeValidationError{
			Scope:   scope,
			Message: "metadata cannot be nil",
		}
	}

	switch scope {
	case ScopeGlobal:
		// Global scope should not have namespace or scoping fields
		if _, hasNamespace := metadata["namespace"]; hasNamespace {
			return &ScopeValidationError{
				Scope:   scope,
				Message: "Global-scoped resources cannot have namespace",
			}
		}
		if _, hasWorkspace := metadata["workspace"]; hasWorkspace {
			return &ScopeValidationError{
				Scope:   scope,
				Message: "Global-scoped resources cannot have workspace",
			}
		}

	case ScopeWorkspace:
		// Workspace scope requires workspace field
		if _, hasWorkspace := metadata["workspace"]; !hasWorkspace {
			return &ScopeValidationError{
				Scope:   scope,
				Message: "Workspace-scoped resources must have workspace",
			}
		}

	case ScopeApp:
		// App scope requires app field
		if _, hasApp := metadata["app"]; !hasApp {
			return &ScopeValidationError{
				Scope:   scope,
				Message: "App-scoped resources must have app",
			}
		}

	case ScopeDomain:
		// Domain scope requires domain field
		if _, hasDomain := metadata["domain"]; !hasDomain {
			return &ScopeValidationError{
				Scope:   scope,
				Message: "Domain-scoped resources must have domain",
			}
		}

	case ScopeEcosystem:
		// Ecosystem scope requires ecosystem field
		if _, hasEcosystem := metadata["ecosystem"]; !hasEcosystem {
			return &ScopeValidationError{
				Scope:   scope,
				Message: "Ecosystem-scoped resources must have ecosystem",
			}
		}

	default:
		return &ScopeValidationError{
			Scope:   scope,
			Message: "unknown scope type",
		}
	}

	return nil
}
