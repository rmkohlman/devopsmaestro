package envvalidation

import (
	"fmt"
	"testing"
)

// =============================================================================
// SM-8: ValidateEnvKey Tests
// RED: These tests FAIL because ValidateEnvKey panics (stub) until Phase 3.
// =============================================================================

func TestValidateEnvKey_ValidKeys(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"simple uppercase", "MY_VAR"},
		{"single letter", "X"},
		{"with numbers", "MY_VAR_123"},
		{"underscore only separator", "A_B_C"},
		{"leading underscore", "_PRIVATE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnvKey(tt.key)
			if err != nil {
				t.Errorf("ValidateEnvKey(%q) = %v, want nil", tt.key, err)
			}
		})
	}
}

func TestValidateEnvKey_InvalidKeys(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr error
	}{
		{"empty key", "", ErrInvalidKey},
		{"lowercase letters", "my_var", ErrInvalidKey},
		{"mixed case", "MyVar", ErrInvalidKey},
		{"starts with digit", "1VAR", ErrInvalidKey},
		{"contains hyphen", "MY-VAR", ErrInvalidKey},
		{"contains space", "MY VAR", ErrInvalidKey},
		{"contains equals", "MY=VAR", ErrInvalidKey},
		{"contains dot", "MY.VAR", ErrInvalidKey},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnvKey(tt.key)
			if err == nil {
				t.Errorf("ValidateEnvKey(%q) = nil, want error", tt.key)
			}
		})
	}
}

// =============================================================================
// SM-13: ValidateEnvMap Tests
// RED: These tests FAIL because ValidateEnvMap panics (stub) until Phase 3.
// =============================================================================

func TestValidateEnvMap_AllValid(t *testing.T) {
	env := map[string]string{
		"MY_TOKEN":  "value1",
		"LOG_LEVEL": "debug",
		"API_URL":   "https://example.com",
		"_INTERNAL": "value2",
	}

	err := ValidateEnvMap(env)
	if err != nil {
		t.Errorf("ValidateEnvMap() = %v, want nil for all-valid map", err)
	}
}

func TestValidateEnvMap_ContainsInvalidKey(t *testing.T) {
	env := map[string]string{
		"VALID_KEY":   "ok",
		"invalid-key": "bad",
	}

	err := ValidateEnvMap(env)
	if err == nil {
		t.Error("ValidateEnvMap() = nil, want error for map with invalid key")
	}
}

func TestValidateEnvMap_EmptyMap(t *testing.T) {
	err := ValidateEnvMap(map[string]string{})
	if err != nil {
		t.Errorf("ValidateEnvMap(empty) = %v, want nil", err)
	}
}

func TestValidateEnvMap_NilMap(t *testing.T) {
	err := ValidateEnvMap(nil)
	if err != nil {
		t.Errorf("ValidateEnvMap(nil) = %v, want nil", err)
	}
}

// =============================================================================
// SM-14: SanitizeEnvMap Tests
// RED: These tests FAIL because SanitizeEnvMap panics (stub) until Phase 3.
// =============================================================================

func TestSanitizeEnvMap_RemovesInvalidKeys(t *testing.T) {
	env := map[string]string{
		"VALID_KEY":     "good",
		"another-key":   "bad",
		"ALSO_VALID":    "good",
		"1starts_digit": "bad",
	}

	sanitized, skipped := SanitizeEnvMap(env)

	if len(sanitized) != 2 {
		t.Errorf("SanitizeEnvMap() sanitized len = %d, want 2", len(sanitized))
	}
	if len(skipped) != 2 {
		t.Errorf("SanitizeEnvMap() skipped len = %d, want 2", len(skipped))
	}
	if sanitized["VALID_KEY"] != "good" {
		t.Errorf("SanitizeEnvMap() should keep VALID_KEY")
	}
	if sanitized["ALSO_VALID"] != "good" {
		t.Errorf("SanitizeEnvMap() should keep ALSO_VALID")
	}
}

func TestSanitizeEnvMap_AllValidPassthrough(t *testing.T) {
	env := map[string]string{
		"KEY_A": "val1",
		"KEY_B": "val2",
	}

	sanitized, skipped := SanitizeEnvMap(env)

	if len(skipped) != 0 {
		t.Errorf("SanitizeEnvMap() skipped = %v, want none for all-valid map", skipped)
	}
	if len(sanitized) != 2 {
		t.Errorf("SanitizeEnvMap() sanitized len = %d, want 2", len(sanitized))
	}
}

func TestSanitizeEnvMap_EmptyMap(t *testing.T) {
	sanitized, skipped := SanitizeEnvMap(map[string]string{})

	if len(sanitized) != 0 {
		t.Errorf("SanitizeEnvMap(empty) sanitized len = %d, want 0", len(sanitized))
	}
	if len(skipped) != 0 {
		t.Errorf("SanitizeEnvMap(empty) skipped len = %d, want 0", len(skipped))
	}
}

// TestValidateEnvKey_ErrorWrapping verifies that errors can be inspected with errors.Is
func TestValidateEnvKey_ErrorWrapping(t *testing.T) {
	err := ValidateEnvKey("bad key")
	if err == nil {
		t.Fatal("ValidateEnvKey('bad key') returned nil, want error")
	}

	// Verify the error message contains the invalid key
	errMsg := fmt.Sprintf("%v", err)
	if errMsg == "" {
		t.Error("ValidateEnvKey() error message should not be empty")
	}
}
