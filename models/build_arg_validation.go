package models

import (
	"devopsmaestro/pkg/envvalidation"
)

// ValidateBuildArgKey validates a single build argument key.
// Keys must match [A-Z_][A-Z0-9_]*, must not be empty, must not use the
// reserved DVM_ prefix, and must not be in the security denylist.
//
// This is a thin convenience wrapper over envvalidation.ValidateEnvKey so
// callers in the models package don't need to import envvalidation directly.
func ValidateBuildArgKey(key string) error {
	return envvalidation.ValidateEnvKey(key)
}

// IsDangerousEnvVar returns true if the given key is in the security
// denylist of environment variable names that could be exploited for
// privilege escalation or code injection when used as build args.
//
// This is a thin convenience wrapper over envvalidation.IsDangerousEnvVar.
func IsDangerousEnvVar(key string) bool {
	return envvalidation.IsDangerousEnvVar(key)
}
