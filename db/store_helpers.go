package db

import "fmt"

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
