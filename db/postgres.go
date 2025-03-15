package db

import (
	"database/sql"
	"fmt"

	"github.com/spf13/viper"

	_ "github.com/lib/pq" // PostgreSQL driver
)

type PostgresDB struct {
	conn *sql.DB
	_dsn string
}

// Register the PostgreSQL implementation with the factory
func init() {
	RegisterDatabase("POSTGRES", NewPostgresDB)
}

// NewPostgresDB creates a new PostgreSQL database connection and returns it as a Database interface
func NewPostgresDB() (Database, error) {
	return connectPostgres(postgresDSN())
}

// connectPostgres establishes the database connection using the DSN and returns a PostgresDB instance
func connectPostgres(dataSourceName string) (*PostgresDB, error) {
	conn, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	return &PostgresDB{conn: conn}, nil
}

// buildDSN constructs the DSN string for PostgreSQL connection
func buildPostgresDSN(host, port, dbname, user, password, sslmode, sslcert, sslkey, sslrootcert, searchPath string) string {
	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		host, port, dbname, user, password, sslmode)

	if sslcert != "" {
		dsn += fmt.Sprintf(" sslcert=%s", sslcert)
	}
	if sslkey != "" {
		dsn += fmt.Sprintf(" sslkey=%s", sslkey)
	}
	if sslrootcert != "" {
		dsn += fmt.Sprintf(" sslrootcert=%s", sslrootcert)
	}
	if searchPath != "" {
		dsn += fmt.Sprintf(" search_path=%s", searchPath)
	}

	return dsn
}

// dsn retrieves the configuration values and constructs the DSN string
func postgresDSN() string {
	host := viper.GetString("database_host")
	port := viper.GetString("database_port")
	dbname := viper.GetString("database_name")
	user := viper.GetString("database_username")
	password := viper.GetString("database_password")
	sslmode := viper.GetString("database_sslmode")
	sslcert := viper.GetString("database_sslcert")
	sslkey := viper.GetString("database_sslkey")
	sslrootcert := viper.GetString("database_sslrootcert")
	searchPath := viper.GetString("database_search_path")

	return buildPostgresDSN(host, port, dbname, user, password, sslmode, sslcert, sslkey, sslrootcert, searchPath)
}

// DSN returns the DSN string
func (p *PostgresDB) DSN() string {
	if p._dsn == "" {
		p._dsn = postgresDSN()
	}
	return p._dsn
}

func (p *PostgresDB) MigrationDSN() string {
	return fmt.Sprintf("postgres://%s", p.DSN())
}

// Connect is part of the Database interface, ensuring the connection to the database
func (p *PostgresDB) Connect() error {
	return p.conn.Ping()
}

// Close closes the database connection as part of the Database interface
func (p *PostgresDB) Close() error {
	return p.conn.Close()
}

// Execute runs a command that doesn't return rows (e.g., INSERT, UPDATE, DELETE)
func (p *PostgresDB) Execute(query string, args ...interface{}) (interface{}, error) {
	return p.conn.Exec(query, args...)
}

// FetchOne retrieves a single record from the database (e.g., SELECT WHERE id = ?)
func (p *PostgresDB) FetchOne(query string, args ...interface{}) (interface{}, error) {
	return p.conn.QueryRow(query, args...), nil
}

// FetchMany retrieves multiple records from the database (e.g., SELECT * FROM table)
func (p *PostgresDB) FetchMany(query string, args ...interface{}) (interface{}, error) {
	return p.conn.Query(query, args...)
}

// Ensure PostgresDB implements the Database interface
var _ Database = (*PostgresDB)(nil)
