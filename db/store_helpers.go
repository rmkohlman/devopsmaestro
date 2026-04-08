package db

import (
	"fmt"
	"regexp"
)

// validLabelKeyPattern matches label keys that are safe for use in JSON extract paths.
// Only allows alphanumeric characters, hyphens, underscores, and dots.
var validLabelKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// validateLabelKey checks that a label key contains only safe characters for
// use in SQL JSON extraction queries. This prevents SQL injection via crafted keys.
func validateLabelKey(key string) error {
	if key == "" {
		return fmt.Errorf("label key must not be empty")
	}
	if !validLabelKeyPattern.MatchString(key) {
		return fmt.Errorf("label key %q contains invalid characters: only alphanumeric, hyphens, underscores, and dots are allowed", key)
	}
	return nil
}

// jsonExtractEqualsArgs builds the bind arguments for QueryBuilder.JSONExtractEquals.
// SQLite needs (key, value, key, value); PostgreSQL needs (key, value).
func jsonExtractEqualsArgs(qb QueryBuilder, key, value string) []interface{} {
	n := qb.JSONExtractEqualsArgs()
	if n == 2 {
		return []interface{}{key, value}
	}
	// SQLite fallback: 4 args (key, value for json_extract; key, value for LIKE)
	return []interface{}{key, value, key, value}
}

// deleteByName removes a row from the given table where name matches.
// It returns ErrNotFound if no row was affected.
func (ds *SQLDataStore) deleteByName(table, resourceLabel, name string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE name = ?`, table)
	result, err := ds.driver.Execute(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete %s: %w", resourceLabel, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return NewErrNotFound(resourceLabel, name)
	}
	return nil
}
