package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildInsertQuery(t *testing.T) {
	type TestModel struct {
		ID    int    `db:"id"`
		Name  string `db:"name"`
		Email string `db:"email"`
		Age   int    `db:"age"`
	}

	model := &TestModel{
		ID:    1,
		Name:  "John Doe",
		Email: "john.doe@example.com",
		Age:   30,
	}

	queryBuilder := NewSQLQueryBuilder()
	query, values := queryBuilder.BuildInsertQuery(model)

	expectedQuery := "INSERT INTO testmodel (id, name, email, age) VALUES ($1, $2, $3, $4)"
	expectedValues := []interface{}{1, "John Doe", "john.doe@example.com", 30}

	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, expectedValues, values)
}

func TestBuildUpdateQuery(t *testing.T) {
	type TestModel struct {
		ID    int    `db:"id"`
		Name  string `db:"name"`
		Email string `db:"email"`
		Age   int    `db:"age"`
	}

	model := &TestModel{
		Name:  "John Doe",
		Email: "john.doe@example.com",
		Age:   30,
	}

	queryBuilder := NewSQLQueryBuilder()
	query, values := queryBuilder.BuildUpdateQuery(model, "id = $1", 1)

	expectedQuery := "UPDATE testmodel SET name = $1, email = $2, age = $3 WHERE id = $4"
	expectedValues := []interface{}{"John Doe", "john.doe@example.com", 30, 1}

	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, expectedValues, values)
}

func TestBuildSelectQuery(t *testing.T) {
	type TestModel struct {
		ID    int    `db:"id"`
		Name  string `db:"name"`
		Email string `db:"email"`
		Age   int    `db:"age"`
	}

	model := &TestModel{}

	queryBuilder := NewSQLQueryBuilder()
	query, values := queryBuilder.BuildSelectQuery(model, "id = $1", 1)

	expectedQuery := "SELECT id, name, email, age FROM testmodel WHERE id = $1"
	expectedValues := []interface{}{1}

	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, expectedValues, values)
}

func TestIsZeroValue(t *testing.T) {
	assert.True(t, isZeroValue(0))
	assert.True(t, isZeroValue(""))
	assert.True(t, isZeroValue(nil))
	assert.False(t, isZeroValue(1))
	assert.False(t, isZeroValue("non-zero"))
}
