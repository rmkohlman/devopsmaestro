package db

import (
	"fmt"
	"os"
	"path/filepath"
)

// DriverConfig provides configuration for creating a Driver.
type DriverConfig struct {
	// Type is the database driver type (sqlite, postgres, duckdb, memory).
	Type DriverType

	// Path is the file path for file-based databases (SQLite, DuckDB).
	// Ignored for server-based databases.
	Path string

	// Host is the server hostname for server-based databases.
	Host string

	// Port is the server port for server-based databases.
	Port string

	// Database is the database name for server-based databases.
	Database string

	// Username is the authentication username.
	Username string

	// Password is the authentication password.
	Password string

	// SSLMode is the SSL connection mode (disable, require, verify-ca, verify-full).
	SSLMode string

	// SSLCert is the path to the client certificate file.
	SSLCert string

	// SSLKey is the path to the client key file.
	SSLKey string

	// SSLRootCert is the path to the root certificate file.
	SSLRootCert string

	// Schema is the database schema (PostgreSQL search_path).
	Schema string

	// MaxOpenConns is the maximum number of open connections (0 = unlimited).
	MaxOpenConns int

	// MaxIdleConns is the maximum number of idle connections.
	MaxIdleConns int

	// ConnMaxLifetimeSeconds is the maximum connection lifetime in seconds.
	ConnMaxLifetimeSeconds int
}

// DriverCreator is a function that creates a Driver from configuration.
type DriverCreator func(cfg DriverConfig) (Driver, error)

// driverRegistry holds registered driver creators.
var driverRegistry = make(map[DriverType]DriverCreator)

// RegisterDriver registers a driver creator for the given type.
func RegisterDriver(driverType DriverType, creator DriverCreator) {
	driverRegistry[driverType] = creator
}

// NewDriver creates a Driver based on the provided configuration.
// This is the main factory function for creating database drivers.
func NewDriver(cfg DriverConfig) (Driver, error) {
	creator, exists := driverRegistry[cfg.Type]
	if !exists {
		return nil, fmt.Errorf("unsupported driver type: %s", cfg.Type)
	}

	return creator(cfg)
}

// NewDefaultDriver creates a Driver with default SQLite configuration.
func NewDefaultDriver() (Driver, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	return NewDriver(DriverConfig{
		Type: DriverSQLite,
		Path: filepath.Join(homeDir, ".devopsmaestro", "devopsmaestro.db"),
	})
}

// NewMemoryDriver creates an in-memory SQLite driver for testing.
func NewMemoryDriver() (Driver, error) {
	return NewDriver(DriverConfig{
		Type: DriverMemory,
	})
}

// ExpandPath expands ~ in file paths and ensures parent directory exists.
func ExpandPath(path string) (string, error) {
	if path == "" {
		return path, nil
	}

	// Expand tilde
	if len(path) >= 2 && path[:2] == "~/" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to expand ~: %w", err)
		}
		path = filepath.Join(homeDir, path[2:])
	}

	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Ensure parent directory exists
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	return absPath, nil
}

// ListRegisteredDrivers returns all registered driver types.
func ListRegisteredDrivers() []DriverType {
	types := make([]DriverType, 0, len(driverRegistry))
	for t := range driverRegistry {
		types = append(types, t)
	}
	return types
}

// IsDriverRegistered checks if a driver type is registered.
func IsDriverRegistered(driverType DriverType) bool {
	_, exists := driverRegistry[driverType]
	return exists
}
