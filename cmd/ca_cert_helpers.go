// Package cmd provides helper functions for managing global CA certificate defaults.
// Global CA certs are stored in the defaults table under the key "ca-certs"
// as a JSON-encoded []models.CACertConfig array.
//
// This is the CA cert analogue of build_arg_helpers.go.
package cmd

import (
	"encoding/json"
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
)

// defaultsCACertsKey is the key used in the defaults table to store global CA certs.
const defaultsCACertsKey = "ca-certs"

// GetGlobalCACerts retrieves the global CA certs from the defaults table.
// Returns an empty (non-nil) slice when no certs have been set.
func GetGlobalCACerts(ds db.DataStore) ([]models.CACertConfig, error) {
	raw, err := ds.GetDefault(defaultsCACertsKey)
	if err != nil {
		return nil, fmt.Errorf("getting global CA certs: %w", err)
	}
	if raw == "" {
		return []models.CACertConfig{}, nil
	}
	var certs []models.CACertConfig
	if err := json.Unmarshal([]byte(raw), &certs); err != nil {
		return nil, fmt.Errorf("parsing global CA certs JSON: %w", err)
	}
	if certs == nil {
		certs = []models.CACertConfig{}
	}
	return certs, nil
}

// SetGlobalCACert adds or updates a single global CA cert by name.
// It performs a read-modify-write on the defaults table, matching by cert Name.
// If a cert with the same name already exists, it is replaced (upsert semantics).
func SetGlobalCACert(ds db.DataStore, cert models.CACertConfig) error {
	certs, err := GetGlobalCACerts(ds)
	if err != nil {
		return err
	}

	// Upsert: replace existing cert with same name, or append
	found := false
	for i, c := range certs {
		if c.Name == cert.Name {
			certs[i] = cert
			found = true
			break
		}
	}
	if !found {
		certs = append(certs, cert)
	}

	b, err := json.Marshal(certs)
	if err != nil {
		return fmt.Errorf("encoding global CA certs: %w", err)
	}

	return ds.SetDefault(defaultsCACertsKey, string(b))
}

// DeleteGlobalCACert removes a single CA cert by name from the global defaults.
// If the cert does not exist, the operation is a no-op (no error returned).
func DeleteGlobalCACert(ds db.DataStore, name string) error {
	certs, err := GetGlobalCACerts(ds)
	if err != nil {
		return err
	}

	// Find and remove
	idx := -1
	for i, c := range certs {
		if c.Name == name {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil // no-op
	}

	certs = append(certs[:idx], certs[idx+1:]...)

	if len(certs) == 0 {
		// Clear the key entirely when empty
		return ds.SetDefault(defaultsCACertsKey, "")
	}

	b, err := json.Marshal(certs)
	if err != nil {
		return fmt.Errorf("encoding global CA certs: %w", err)
	}

	return ds.SetDefault(defaultsCACertsKey, string(b))
}

// removeCACertFromSlice removes a cert by name from a CACertConfig slice.
// Returns the updated slice and true if the cert was found/removed, false otherwise.
func removeCACertFromSlice(certs []models.CACertConfig, name string) ([]models.CACertConfig, bool) {
	for i, c := range certs {
		if c.Name == name {
			return append(certs[:i], certs[i+1:]...), true
		}
	}
	return certs, false
}

// upsertCACertInSlice adds or replaces a cert by name in a CACertConfig slice.
// Returns the updated slice.
func upsertCACertInSlice(certs []models.CACertConfig, cert models.CACertConfig) []models.CACertConfig {
	for i, c := range certs {
		if c.Name == cert.Name {
			certs[i] = cert
			return certs
		}
	}
	return append(certs, cert)
}

// deleteCACertFromWrappedJSON removes a CA cert by name from a JSON blob of the form
// {"caCerts": [...], ...}. Used for app.build_config and workspace.build_config.
func deleteCACertFromWrappedJSON(raw, certName string) (string, error) {
	if raw == "" {
		return "", nil
	}
	var wrapper map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &wrapper); err != nil {
		return "", fmt.Errorf("parsing build config JSON: %w", err)
	}

	certsRaw, ok := wrapper["caCerts"]
	if !ok {
		return raw, nil // no "caCerts" key — nothing to delete
	}

	var certs []models.CACertConfig
	if err := json.Unmarshal(certsRaw, &certs); err != nil {
		return raw, nil
	}

	certs, found := removeCACertFromSlice(certs, certName)
	if !found {
		return raw, nil
	}

	if len(certs) == 0 {
		delete(wrapper, "caCerts")
	} else {
		b, err := json.Marshal(certs)
		if err != nil {
			return "", fmt.Errorf("encoding CA certs: %w", err)
		}
		wrapper["caCerts"] = b
	}

	result, err := json.Marshal(wrapper)
	if err != nil {
		return "", fmt.Errorf("encoding build config JSON: %w", err)
	}
	return string(result), nil
}

// parseCACertsFromDirectJSON parses a direct JSON array of CACertConfig (used for eco/domain ca_certs column).
func parseCACertsFromDirectJSON(raw string) []models.CACertConfig {
	if raw == "" {
		return nil
	}
	var certs []models.CACertConfig
	if err := json.Unmarshal([]byte(raw), &certs); err != nil {
		return nil
	}
	return certs
}

// parseCACertsFromWrappedJSON extracts the "caCerts" field from a wrapped build config JSON.
func parseCACertsFromWrappedJSON(raw string) []models.CACertConfig {
	if raw == "" {
		return nil
	}
	var wrapper struct {
		CACerts []models.CACertConfig `json:"caCerts"`
	}
	if err := json.Unmarshal([]byte(raw), &wrapper); err != nil {
		return nil
	}
	return wrapper.CACerts
}
