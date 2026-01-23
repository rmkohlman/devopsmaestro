package db

import (
	"fmt"
	"reflect"
	"strings"
)

// SQLQueryBuilder is responsible for building SQL queries dynamically using reflection.
type SQLQueryBuilder struct{}

// NewSQLQueryBuilder creates a new instance of SQLQueryBuilder.
func NewSQLQueryBuilder() *SQLQueryBuilder {
	return &SQLQueryBuilder{}
}

// BuildInsertQuery builds an INSERT SQL query from a model object using reflection.
// Fields with zero values (e.g., nil, empty) are omitted from the query, allowing the database to use defaults.
func (sq *SQLQueryBuilder) BuildInsertQuery(model interface{}) (string, []interface{}) {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem() // Get the value that the pointer points to
	}
	typ := val.Type()

	var fieldNames []string
	var placeholders []string
	var values []interface{}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i).Interface()

		// Skip unexported fields
		if !val.Field(i).CanInterface() {
			continue
		}

		// Handle zero values: Omit from the query if value is zero
		if isZeroValue(value) {
			continue // Skip this field
		}

		// Use field tag to get the actual DB column name, or default to field name
		columnName := field.Tag.Get("db")
		if columnName == "" {
			columnName = field.Name
		}

		fieldNames = append(fieldNames, columnName)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(values)+1))
		values = append(values, value)
	}

	tableName := strings.ToLower(typ.Name()) + "s" // Use the struct name as the table name (pluralized)

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(fieldNames, ", "), strings.Join(placeholders, ", "))
	return query, values
}

// BuildUpdateQuery builds an UPDATE SQL query from a model object using reflection.
func (sq *SQLQueryBuilder) BuildUpdateQuery(model interface{}, condition string, conditionArgs ...interface{}) (string, []interface{}) {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem() // Get the value that the pointer points to
	}
	typ := val.Type()

	var setClauses []string
	var values []interface{}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i).Interface()

		if !val.Field(i).CanInterface() {
			continue
		}

		columnName := field.Tag.Get("db")
		if columnName == "" {
			columnName = field.Name
		}

		// Exclude the ID field from the SET clause and values array
		if columnName == "id" {
			continue
		}

		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", columnName, len(values)+1))
		values = append(values, value)
	}

	// Add condition arguments at the end
	for i, arg := range conditionArgs {
		condition = strings.Replace(condition, fmt.Sprintf("$%d", i+1), fmt.Sprintf("$%d", len(values)+i+1), 1)
		values = append(values, arg)
	}

	tableName := strings.ToLower(typ.Name()) + "s"
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", tableName, strings.Join(setClauses, ", "), condition)
	return query, values
}

// BuildSelectQuery builds a SELECT SQL query from a model object using reflection.
func (sq *SQLQueryBuilder) BuildSelectQuery(model interface{}, condition string, conditionArgs ...interface{}) (string, []interface{}) {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem() // Get the value that the pointer points to
	}
	typ := val.Type()

	var fieldNames []string
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		if !val.Field(i).CanInterface() {
			continue
		}

		columnName := field.Tag.Get("db")
		if columnName == "" {
			columnName = field.Name
		}

		fieldNames = append(fieldNames, columnName)
	}

	tableName := strings.ToLower(typ.Name()) + "s"
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", strings.Join(fieldNames, ", "), tableName, condition)
	return query, conditionArgs
}

// Helper function to check if a value is zero
func isZeroValue(value interface{}) bool {
	if value == nil {
		return true
	}
	return reflect.DeepEqual(value, reflect.Zero(reflect.TypeOf(value)).Interface())
}
