package db

import "fmt"

// MigrationVersion returns the current database migration version by querying
// the schema_migrations table (managed by golang-migrate).
// Returns 0 if no migrations have been applied yet (table missing or empty).
func (ds *SQLDataStore) MigrationVersion() (int, error) {
	var version int
	var dirty bool

	err := ds.driver.QueryRow(
		"SELECT version, dirty FROM schema_migrations LIMIT 1",
	).Scan(&version, &dirty)

	if err != nil {
		// If the table doesn't exist or has no rows, return 0
		// This handles fresh databases before any migrations
		return 0, nil
	}

	if dirty {
		return version, fmt.Errorf("database migration is in dirty state at version %d", version)
	}

	return version, nil
}
