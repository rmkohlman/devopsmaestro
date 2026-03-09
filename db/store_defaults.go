package db

import (
	"database/sql"
	"errors"
	"fmt"
)

// =============================================================================
// Default Operations
// =============================================================================

// GetDefault retrieves a default value by key.
// Returns empty string if key is not found (not an error).
func (ds *SQLDataStore) GetDefault(key string) (string, error) {
	query := `SELECT value FROM defaults WHERE key = ?`

	var value string
	err := ds.driver.QueryRow(query, key).Scan(&value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil // Not found, return empty string (not an error)
		}
		return "", fmt.Errorf("failed to get default for key %s: %w", key, err)
	}

	return value, nil
}

// SetDefault sets a default value for the given key.
// Uses upsert behavior (INSERT OR REPLACE).
func (ds *SQLDataStore) SetDefault(key, value string) error {
	query := `
		INSERT OR REPLACE INTO defaults (key, value, updated_at) 
		VALUES (?, ?, CURRENT_TIMESTAMP)
	`

	_, err := ds.driver.Execute(query, key, value)
	if err != nil {
		return fmt.Errorf("failed to set default for key %s: %w", key, err)
	}

	return nil
}

// DeleteDefault removes a default value by key.
// No error if key doesn't exist.
func (ds *SQLDataStore) DeleteDefault(key string) error {
	query := `DELETE FROM defaults WHERE key = ?`

	_, err := ds.driver.Execute(query, key)
	if err != nil {
		return fmt.Errorf("failed to delete default for key %s: %w", key, err)
	}

	// No error if key doesn't exist (DELETE just affects 0 rows)
	return nil
}

// ListDefaults retrieves all default values as a key-value map.
func (ds *SQLDataStore) ListDefaults() (map[string]string, error) {
	query := `SELECT key, value FROM defaults ORDER BY key`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list defaults: %w", err)
	}
	defer rows.Close()

	defaults := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("failed to scan default: %w", err)
		}
		defaults[key] = value
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating defaults: %w", err)
	}

	return defaults, nil
}
