package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestBuildSqliteDSN(t *testing.T) {
	filePath := "/tmp/test.db"
	expectedDSN := "file:/tmp/test.db?cache=shared&mode=rwc"
	dsn := buildSqliteDSN(filePath)
	assert.Equal(t, expectedDSN, dsn)
}

func TestSqliteDSN(t *testing.T) {
	homeDir := os.Getenv("HOME")
	expectedFilePath := filepath.Join(homeDir, ".config/dvm/db/dvm.db")
	expectedDSN := "file:" + expectedFilePath + "?cache=shared&mode=rwc"
	viper.Set("database_file_path", "")
	dsn := sqliteDSN()
	assert.Equal(t, expectedDSN, dsn)

	viper.Set("database_file_path", "/tmp/test.db")
	expectedDSN = "file:/tmp/test.db?cache=shared&mode=rwc"
	dsn = sqliteDSN()
	assert.Equal(t, expectedDSN, dsn)
}

func TestNewSQLiteDB(t *testing.T) {
	viper.Set("database_file_path", ":memory:")
	db, err := NewSQLiteDB()
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()
}

func TestSQLiteDB_Connect(t *testing.T) {
	viper.Set("database_file_path", ":memory:")
	db, err := NewSQLiteDB()
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()

	err = db.Connect()
	assert.NoError(t, err)
}

func TestSQLiteDB_Close(t *testing.T) {
	viper.Set("database_file_path", ":memory:")
	db, err := NewSQLiteDB()
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()

	err = db.Close()
	assert.NoError(t, err)
}

func TestSQLiteDB_Execute(t *testing.T) {
	viper.Set("database_file_path", ":memory:")
	db, err := NewSQLiteDB()
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()

	_, err = db.Execute("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	assert.NoError(t, err)

	_, err = db.Execute("INSERT INTO test (name) VALUES (?)", "test_name")
	assert.NoError(t, err)
}

func TestSQLiteDB_FetchOne(t *testing.T) {
	viper.Set("database_file_path", ":memory:")
	db, err := NewSQLiteDB()
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()

	_, err = db.Execute("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	assert.NoError(t, err)

	_, err = db.Execute("INSERT INTO test (name) VALUES (?)", "test_name")
	assert.NoError(t, err)

	row, err := db.FetchOne("SELECT name FROM test WHERE id = ?", 1)
	assert.NoError(t, err)

	var name string
	err = row.(*sql.Row).Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, "test_name", name)
}

func TestSQLiteDB_FetchMany(t *testing.T) {
	viper.Set("database_file_path", ":memory:")
	db, err := NewSQLiteDB()
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()

	_, err = db.Execute("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	assert.NoError(t, err)

	_, err = db.Execute("INSERT INTO test (name) VALUES (?)", "test_name1")
	assert.NoError(t, err)
	_, err = db.Execute("INSERT INTO test (name) VALUES (?)", "test_name2")
	assert.NoError(t, err)

	rows, err := db.FetchMany("SELECT name FROM test")
	assert.NoError(t, err)
	defer rows.(*sql.Rows).Close()

	var names []string
	for rows.(*sql.Rows).Next() {
		var name string
		err = rows.(*sql.Rows).Scan(&name)
		assert.NoError(t, err)
		names = append(names, name)
	}

	assert.ElementsMatch(t, []string{"test_name1", "test_name2"}, names)
}
