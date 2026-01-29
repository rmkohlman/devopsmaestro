package db

import (
	"testing"
)

func TestSQLiteQueryBuilder_Placeholder(t *testing.T) {
	builder := NewSQLiteQueryBuilder()

	tests := []struct {
		index    int
		expected string
	}{
		{1, "?"},
		{2, "?"},
		{10, "?"},
	}

	for _, tt := range tests {
		result := builder.Placeholder(tt.index)
		if result != tt.expected {
			t.Errorf("Placeholder(%d) = %q, want %q", tt.index, result, tt.expected)
		}
	}
}

func TestSQLiteQueryBuilder_Now(t *testing.T) {
	builder := NewSQLiteQueryBuilder()
	result := builder.Now()
	expected := "datetime('now')"

	if result != expected {
		t.Errorf("Now() = %q, want %q", result, expected)
	}
}

func TestSQLiteQueryBuilder_Boolean(t *testing.T) {
	builder := NewSQLiteQueryBuilder()

	tests := []struct {
		value    bool
		expected string
	}{
		{true, "1"},
		{false, "0"},
	}

	for _, tt := range tests {
		result := builder.Boolean(tt.value)
		if result != tt.expected {
			t.Errorf("Boolean(%v) = %q, want %q", tt.value, result, tt.expected)
		}
	}
}

func TestSQLiteQueryBuilder_UpsertSuffix(t *testing.T) {
	builder := NewSQLiteQueryBuilder()

	tests := []struct {
		name            string
		conflictColumns []string
		updateColumns   []string
		expected        string
	}{
		{
			name:            "single column conflict",
			conflictColumns: []string{"id"},
			updateColumns:   []string{"name", "value"},
			expected:        "ON CONFLICT(id) DO UPDATE SET name = excluded.name, value = excluded.value",
		},
		{
			name:            "multiple column conflict",
			conflictColumns: []string{"workspace_id", "plugin_id"},
			updateColumns:   []string{"enabled"},
			expected:        "ON CONFLICT(workspace_id, plugin_id) DO UPDATE SET enabled = excluded.enabled",
		},
		{
			name:            "empty conflict columns",
			conflictColumns: []string{},
			updateColumns:   []string{"name"},
			expected:        "",
		},
		{
			name:            "empty update columns",
			conflictColumns: []string{"id"},
			updateColumns:   []string{},
			expected:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.UpsertSuffix(tt.conflictColumns, tt.updateColumns)
			if result != tt.expected {
				t.Errorf("UpsertSuffix() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSQLiteQueryBuilder_LimitOffset(t *testing.T) {
	builder := NewSQLiteQueryBuilder()

	tests := []struct {
		name     string
		limit    int
		offset   int
		expected string
	}{
		{"limit only", 10, 0, "LIMIT 10"},
		{"limit and offset", 10, 20, "LIMIT 10 OFFSET 20"},
		{"zero limit", 0, 10, ""},
		{"negative limit", -1, 10, ""},
		{"large values", 1000, 5000, "LIMIT 1000 OFFSET 5000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.LimitOffset(tt.limit, tt.offset)
			if result != tt.expected {
				t.Errorf("LimitOffset(%d, %d) = %q, want %q", tt.limit, tt.offset, result, tt.expected)
			}
		})
	}
}

func TestSQLiteQueryBuilder_Dialect(t *testing.T) {
	builder := NewSQLiteQueryBuilder()
	result := builder.Dialect()
	expected := "sqlite"

	if result != expected {
		t.Errorf("Dialect() = %q, want %q", result, expected)
	}
}

func TestPostgresQueryBuilder_Placeholder(t *testing.T) {
	builder := NewPostgresQueryBuilder()

	tests := []struct {
		index    int
		expected string
	}{
		{1, "$1"},
		{2, "$2"},
		{10, "$10"},
	}

	for _, tt := range tests {
		result := builder.Placeholder(tt.index)
		if result != tt.expected {
			t.Errorf("Placeholder(%d) = %q, want %q", tt.index, result, tt.expected)
		}
	}
}

func TestPostgresQueryBuilder_Now(t *testing.T) {
	builder := NewPostgresQueryBuilder()
	result := builder.Now()
	expected := "NOW()"

	if result != expected {
		t.Errorf("Now() = %q, want %q", result, expected)
	}
}

func TestPostgresQueryBuilder_Boolean(t *testing.T) {
	builder := NewPostgresQueryBuilder()

	tests := []struct {
		value    bool
		expected string
	}{
		{true, "TRUE"},
		{false, "FALSE"},
	}

	for _, tt := range tests {
		result := builder.Boolean(tt.value)
		if result != tt.expected {
			t.Errorf("Boolean(%v) = %q, want %q", tt.value, result, tt.expected)
		}
	}
}

func TestPostgresQueryBuilder_Dialect(t *testing.T) {
	builder := NewPostgresQueryBuilder()
	result := builder.Dialect()
	expected := "postgres"

	if result != expected {
		t.Errorf("Dialect() = %q, want %q", result, expected)
	}
}

func TestQueryBuilderFor(t *testing.T) {
	tests := []struct {
		driverType      DriverType
		expectedDialect string
	}{
		{DriverSQLite, "sqlite"},
		{DriverMemory, "sqlite"},
		{DriverPostgres, "postgres"},
		{DriverDuckDB, "sqlite"}, // DuckDB uses SQLite-compatible syntax
		{"unknown", "sqlite"},    // Default to SQLite
	}

	for _, tt := range tests {
		t.Run(string(tt.driverType), func(t *testing.T) {
			builder := QueryBuilderFor(tt.driverType)
			if builder.Dialect() != tt.expectedDialect {
				t.Errorf("QueryBuilderFor(%s).Dialect() = %q, want %q",
					tt.driverType, builder.Dialect(), tt.expectedDialect)
			}
		})
	}
}

// Test that query builders implement the interface
func TestQueryBuilderInterface(t *testing.T) {
	var _ QueryBuilder = (*SQLiteQueryBuilder)(nil)
	var _ QueryBuilder = (*PostgresQueryBuilder)(nil)
}
