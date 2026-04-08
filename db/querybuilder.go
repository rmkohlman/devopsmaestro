package db

import (
	"fmt"
	"strings"
)

// SQLiteQueryBuilder implements QueryBuilder for SQLite dialect.
type SQLiteQueryBuilder struct{}

// NewSQLiteQueryBuilder creates a new SQLite query builder.
func NewSQLiteQueryBuilder() *SQLiteQueryBuilder {
	return &SQLiteQueryBuilder{}
}

// Placeholder returns ? for SQLite (positional placeholders).
func (b *SQLiteQueryBuilder) Placeholder(index int) string {
	return "?"
}

// Now returns SQLite's current timestamp function.
func (b *SQLiteQueryBuilder) Now() string {
	return "datetime('now')"
}

// Boolean returns SQLite boolean representation (0/1).
func (b *SQLiteQueryBuilder) Boolean(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

// UpsertSuffix returns SQLite's ON CONFLICT clause.
func (b *SQLiteQueryBuilder) UpsertSuffix(conflictColumns []string, updateColumns []string) string {
	if len(conflictColumns) == 0 || len(updateColumns) == 0 {
		return ""
	}

	var updates []string
	for _, col := range updateColumns {
		updates = append(updates, fmt.Sprintf("%s = excluded.%s", col, col))
	}

	return fmt.Sprintf("ON CONFLICT(%s) DO UPDATE SET %s",
		strings.Join(conflictColumns, ", "),
		strings.Join(updates, ", "))
}

// LimitOffset returns SQLite's LIMIT/OFFSET clause.
func (b *SQLiteQueryBuilder) LimitOffset(limit, offset int) string {
	if limit <= 0 {
		return ""
	}
	if offset <= 0 {
		return fmt.Sprintf("LIMIT %d", limit)
	}
	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}

// JSONExtractEquals returns a SQLite condition for matching a JSON field value.
// Requires 4 bind args: key, value, key, value.
func (b *SQLiteQueryBuilder) JSONExtractEquals(column string) string {
	return fmt.Sprintf(
		`(json_extract(%s, '$.' || ?) = ? OR %s LIKE '%%"' || ? || '":"' || ? || '"%%')`,
		column, column,
	)
}

// JSONExtractEqualsArgs returns 4 for SQLite (key, value, key, value).
func (b *SQLiteQueryBuilder) JSONExtractEqualsArgs() int {
	return 4
}

// Dialect returns "sqlite".
func (b *SQLiteQueryBuilder) Dialect() string {
	return "sqlite"
}

// Ensure SQLiteQueryBuilder implements QueryBuilder
var _ QueryBuilder = (*SQLiteQueryBuilder)(nil)

// PostgresQueryBuilder implements QueryBuilder for PostgreSQL dialect.
type PostgresQueryBuilder struct{}

// NewPostgresQueryBuilder creates a new PostgreSQL query builder.
func NewPostgresQueryBuilder() *PostgresQueryBuilder {
	return &PostgresQueryBuilder{}
}

// Placeholder returns $n for PostgreSQL (numbered placeholders).
func (b *PostgresQueryBuilder) Placeholder(index int) string {
	return fmt.Sprintf("$%d", index)
}

// Now returns PostgreSQL's current timestamp function.
func (b *PostgresQueryBuilder) Now() string {
	return "NOW()"
}

// Boolean returns PostgreSQL boolean representation (TRUE/FALSE).
func (b *PostgresQueryBuilder) Boolean(value bool) string {
	if value {
		return "TRUE"
	}
	return "FALSE"
}

// UpsertSuffix returns PostgreSQL's ON CONFLICT clause.
func (b *PostgresQueryBuilder) UpsertSuffix(conflictColumns []string, updateColumns []string) string {
	if len(conflictColumns) == 0 || len(updateColumns) == 0 {
		return ""
	}

	var updates []string
	for _, col := range updateColumns {
		updates = append(updates, fmt.Sprintf("%s = EXCLUDED.%s", col, col))
	}

	return fmt.Sprintf("ON CONFLICT (%s) DO UPDATE SET %s",
		strings.Join(conflictColumns, ", "),
		strings.Join(updates, ", "))
}

// LimitOffset returns PostgreSQL's LIMIT/OFFSET clause.
func (b *PostgresQueryBuilder) LimitOffset(limit, offset int) string {
	if limit <= 0 {
		return ""
	}
	if offset <= 0 {
		return fmt.Sprintf("LIMIT %d", limit)
	}
	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}

// JSONExtractEquals returns a PostgreSQL condition for matching a JSON field value.
// Requires 2 bind args: key, value.
func (b *PostgresQueryBuilder) JSONExtractEquals(column string) string {
	return fmt.Sprintf(`%s->>? = ?`, column)
}

// JSONExtractEqualsArgs returns 2 for PostgreSQL (key, value).
func (b *PostgresQueryBuilder) JSONExtractEqualsArgs() int {
	return 2
}

// Dialect returns "postgres".
func (b *PostgresQueryBuilder) Dialect() string {
	return "postgres"
}

// Ensure PostgresQueryBuilder implements QueryBuilder
var _ QueryBuilder = (*PostgresQueryBuilder)(nil)

// QueryBuilderFor returns the appropriate QueryBuilder for the given driver type.
func QueryBuilderFor(driverType DriverType) QueryBuilder {
	switch driverType {
	case DriverPostgres:
		return NewPostgresQueryBuilder()
	case DriverSQLite, DriverMemory, DriverDuckDB:
		fallthrough
	default:
		return NewSQLiteQueryBuilder()
	}
}
