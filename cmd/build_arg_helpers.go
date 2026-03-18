package cmd

import (
	"encoding/json"
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/pkg/envvalidation"
)

// defaultsBuildArgsKey is the key used in the defaults table to store global build args.
const defaultsBuildArgsKey = "build-args"

// GetGlobalBuildArgs retrieves the global build args from the defaults table.
// Returns an empty (non-nil) map when no args have been set.
func GetGlobalBuildArgs(ds db.DataStore) (map[string]string, error) {
	raw, err := ds.GetDefault(defaultsBuildArgsKey)
	if err != nil {
		return nil, fmt.Errorf("getting global build args: %w", err)
	}
	if raw == "" {
		return map[string]string{}, nil
	}
	var args map[string]string
	if err := json.Unmarshal([]byte(raw), &args); err != nil {
		return nil, fmt.Errorf("parsing global build args JSON: %w", err)
	}
	if args == nil {
		args = map[string]string{}
	}
	return args, nil
}

// SetGlobalBuildArg sets a single global build arg key/value pair.
// It validates the key, then performs a read-modify-write on the defaults table.
// Returns an error if the key is invalid, dangerous, or uses the reserved DVM_ prefix.
func SetGlobalBuildArg(ds db.DataStore, key, value string) error {
	if err := envvalidation.ValidateEnvKey(key); err != nil {
		return err
	}

	args, err := GetGlobalBuildArgs(ds)
	if err != nil {
		return err
	}

	args[key] = value

	b, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("encoding global build args: %w", err)
	}

	return ds.SetDefault(defaultsBuildArgsKey, string(b))
}

// DeleteGlobalBuildArg removes a single key from the global build args store.
// If the key does not exist, the operation is a no-op (no error returned).
// The key is validated before any DB write.
func DeleteGlobalBuildArg(ds db.DataStore, key string) error {
	args, err := GetGlobalBuildArgs(ds)
	if err != nil {
		return err
	}

	// No-op if key doesn't exist
	if _, exists := args[key]; !exists {
		return nil
	}

	delete(args, key)

	b, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("encoding global build args: %w", err)
	}

	return ds.SetDefault(defaultsBuildArgsKey, string(b))
}
