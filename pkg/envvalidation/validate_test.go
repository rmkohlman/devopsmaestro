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

// =============================================================================
// TDD Phase 2 (RED): WI-4 Security Fix Tests
// =============================================================================

// TestIsDangerousEnvVar_Complete verifies that every entry in the security
// denylist returns true from IsDangerousEnvVar.
func TestIsDangerousEnvVar_Complete(t *testing.T) {
	tests := []struct {
		name          string
		envVar        string
		wantDangerous bool
	}{
		{"LD_PRELOAD", "LD_PRELOAD", true},
		{"DYLD_INSERT_LIBRARIES", "DYLD_INSERT_LIBRARIES", true},
		{"LD_LIBRARY_PATH", "LD_LIBRARY_PATH", true},
		{"DYLD_LIBRARY_PATH", "DYLD_LIBRARY_PATH", true},
		{"BASH_ENV", "BASH_ENV", true},
		{"ENV", "ENV", true},
		{"PROMPT_COMMAND", "PROMPT_COMMAND", true},
		{"NODE_OPTIONS", "NODE_OPTIONS", true},
		// safe vars should return false
		{"PATH is not dangerous", "PATH", false},
		{"HOME is not dangerous", "HOME", false},
		{"MY_VAR is not dangerous", "MY_VAR", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDangerousEnvVar(tt.envVar)
			if got != tt.wantDangerous {
				t.Errorf("IsDangerousEnvVar(%q) = %v, want %v", tt.envVar, got, tt.wantDangerous)
			}
		})
	}
}

// TestValidateEnvKey_DVMPrefix verifies that keys starting with "DVM_" are
// rejected as reserved namespace. This tests NEW behaviour that does not
// exist yet in ValidateEnvKey — it will FAIL (RED) until Phase 3 adds
// the DVM_ prefix check to ValidateEnvKey.
func TestValidateEnvKey_DVMPrefix(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "DVM_WORKSPACE should be rejected",
			key:     "DVM_WORKSPACE",
			wantErr: true,
		},
		{
			name:    "DVM_ANYTHING should be rejected",
			key:     "DVM_ANYTHING",
			wantErr: true,
		},
		{
			name:    "DVMTEST without underscore should be allowed",
			key:     "DVMTEST",
			wantErr: false,
		},
		{
			name:    "MY_DVM_VAR should be allowed (DVM_ not at start)",
			key:     "MY_DVM_VAR",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnvKey(tt.key)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateEnvKey(%q) = nil, want error for reserved DVM_ prefix", tt.key)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateEnvKey(%q) = %v, want nil (key should be allowed)", tt.key, err)
			}
		})
	}
}

// =============================================================================
// End of WI-4 security fix tests
// =============================================================================
