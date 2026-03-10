package db

// DataStore is the high-level interface for application data operations.
// It composes all domain-specific sub-interfaces for backward compatibility.
//
// New consumers should depend on the narrowest sub-interface they need
// (e.g., AppStore, WorkspaceStore) rather than the full DataStore interface.
// The sub-interfaces are defined in datastore_interfaces.go.
//
// Implementations can use any underlying Driver (SQLite, PostgreSQL, DuckDB, etc.).
type DataStore interface {
	EcosystemStore
	DomainStore
	AppStore
	WorkspaceStore
	ContextStore
	PluginStore
	ThemeStore
	TerminalPromptStore
	TerminalProfileStore
	TerminalPluginStore
	TerminalEmulatorStore
	CredentialStore
	GitRepoStore
	DefaultsStore
	NvimPackageStore
	TerminalPackageStore
	RegistryStore
	RegistryHistoryStore
	CustomResourceStore

	// Driver Access

	// Driver returns the underlying database driver.
	// Useful for advanced operations or transactions.
	Driver() Driver

	// Health and Maintenance

	// Close releases any resources held by the DataStore.
	Close() error

	// Ping verifies the database connection is alive.
	Ping() error
}

// DataStoreConfig provides configuration for creating a DataStore.
type DataStoreConfig struct {
	// Driver is the database driver to use.
	Driver Driver

	// QueryBuilder is the SQL dialect-specific query builder.
	// If nil, a default builder will be selected based on the driver type.
	QueryBuilder QueryBuilder
}
