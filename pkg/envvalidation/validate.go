// Package envvalidation provides validation for environment variable maps
// used in workspace runtime injection.
//
// This package enforces security policies on environment variables to prevent
// accidental exposure of sensitive data and ensure key naming conventions.
package envvalidation

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ErrInvalidKey is returned when an env var key fails validation.
var ErrInvalidKey = errors.New("invalid env var key")

// ErrForbiddenKey is returned when an env var key is in the forbidden list.
var ErrForbiddenKey = errors.New("forbidden env var key")

// ErrReservedKey is returned when an env var key uses the reserved DVM_ prefix.
var ErrReservedKey = errors.New("key uses reserved DVM_ prefix")

// validEnvKeyPattern matches strict environment variable names.
// Must start with uppercase letter or underscore, followed by uppercase letters,
// digits, or underscores only. This is stricter than POSIX to prevent
// accidental injection via mixed-case or special characters.
var validEnvKeyPattern = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

// dangerousEnvVars is a denylist of environment variable names that could be
// exploited for code injection or library preloading attacks.
var dangerousEnvVars = map[string]bool{
	"LD_PRELOAD":            true,
	"LD_LIBRARY_PATH":       true,
	"DYLD_INSERT_LIBRARIES": true,
	"DYLD_LIBRARY_PATH":     true,
	"NODE_OPTIONS":          true,
	"BASH_ENV":              true,
	"ENV":                   true,
	"PROMPT_COMMAND":        true,
}

// ValidateEnvKey validates a single environment variable key.
// Keys must be non-empty, contain only [A-Z0-9_], and not start with a digit.
// Returns ErrInvalidKey (wrapped) if the key is malformed.
// Returns ErrForbiddenKey (wrapped) if the key is in the security denylist.
// Returns ErrReservedKey (wrapped) if the key uses the reserved DVM_ prefix.
func ValidateEnvKey(key string) error {
	if key == "" {
		return fmt.Errorf("environment variable name cannot be empty: %w", ErrInvalidKey)
	}
	if !validEnvKeyPattern.MatchString(key) {
		return fmt.Errorf("invalid environment variable name %q: must match [A-Z_][A-Z0-9_]*: %w", key, ErrInvalidKey)
	}
	if dangerousEnvVars[key] {
		return fmt.Errorf("environment variable %q is in the security denylist (potentially dangerous): %w", key, ErrForbiddenKey)
	}
	if IsDVMReservedKey(key) {
		return fmt.Errorf("environment variable %q uses the reserved DVM_ prefix: %w", key, ErrReservedKey)
	}
	return nil
}

// ValidateEnvMap validates all keys in an env map.
// Returns the first validation error encountered, or nil if all keys are valid.
func ValidateEnvMap(env map[string]string) error {
	for k := range env {
		if err := ValidateEnvKey(k); err != nil {
			return err
		}
	}
	return nil
}

// SanitizeEnvMap filters out any env entries whose keys fail validation.
// Returns a new map with only valid entries, plus a slice of skipped keys.
func SanitizeEnvMap(env map[string]string) (sanitized map[string]string, skipped []string) {
	sanitized = make(map[string]string)
	for k, v := range env {
		if err := ValidateEnvKey(k); err != nil {
			skipped = append(skipped, k)
			continue
		}
		sanitized[k] = v
	}
	return sanitized, skipped
}

// IsDangerousEnvVar returns true if the env var name is in the security denylist.
func IsDangerousEnvVar(name string) bool {
	return dangerousEnvVars[name]
}

// IsDVMReservedKey returns true if the key uses the reserved DVM_ prefix.
func IsDVMReservedKey(key string) bool {
	return strings.HasPrefix(key, "DVM_")
}
