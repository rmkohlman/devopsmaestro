package main

import (
	"devopsmaestro/cmd"
	"devopsmaestro/db"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

// Version information (set via ldflags during build)
var (
	Version   = "dev"
	BuildTime = "unknown"
	Commit    = "unknown"
)

func run(dbInstance db.Database, dataStoreInstance db.DataStore, executor cmd.Executor, migrationsFS fs.FS) int {
	// Execute the root command of the CLI tool
	cmd.Execute(&dbInstance, &dataStoreInstance, &executor, migrationsFS)

	return 0
}

func loadConfig() error {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Set config path to ~/.devopsmaestro
	configPath := filepath.Join(homeDir, ".devopsmaestro")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	viper.AutomaticEnv()

	// Try to read config, but don't fail if it doesn't exist (init command will create it)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config: %w", err)
		}
		// Config not found is OK - init command will create it
	}

	return nil
}

func main() {
	// Set version information for the CLI
	cmd.Version = Version
	cmd.BuildTime = BuildTime
	cmd.Commit = Commit

	// Check if this is a command that doesn't need database
	// (completion, version, help)
	skipDB := false
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "completion", "version", "--version", "-v", "help", "--help", "-h":
			skipDB = true
		}
	}

	// Load configuration
	if err := loadConfig(); err != nil {
		// Don't fail for commands that don't need config
		if !skipDB {
			fmt.Printf("Failed to load configuration: %v\n", err)
			os.Exit(1)
		}
	}

	// Set default values if config is not found (for init command)
	if viper.GetString("database.type") == "" {
		viper.Set("database.type", "sqlite")
		viper.Set("database.path", "~/.devopsmaestro/devopsmaestro.db")
		viper.Set("store", "sql")
	}

	var dbInstance db.Database
	var dataStoreInstance db.DataStore
	var executor cmd.Executor

	// Only initialize database for commands that need it
	if !skipDB {
		// Initialize the database connection
		var err error
		dbInstance, err = db.InitializeDBConnection()
		if err != nil {
			fmt.Printf("Failed to initialize database: %v\n", err)
			os.Exit(1)
		}

		// Ensure the database connection is closed when the program exits
		defer func() {
			if dbInstance != nil {
				if err := dbInstance.Close(); err != nil {
					fmt.Printf("Failed to close database connection: %v\n", err)
				}
			}
		}()

		// Create an instance of DataStore
		dataStoreInstance, err = db.StoreFactory(dbInstance)
		if err != nil {
			fmt.Printf("Failed to create DataStore instance: %v\n", err)
			os.Exit(1)
		}

		executor = cmd.NewExecutor(dataStoreInstance, dbInstance)
	}

	// Get migrations subdirectory from embedded filesystem
	migrationsSubFS, err := fs.Sub(MigrationsFS, "db/migrations")
	if err != nil {
		fmt.Printf("Failed to access embedded migrations: %v\n", err)
		os.Exit(1)
	}

	os.Exit(run(dbInstance, dataStoreInstance, executor, migrationsSubFS))
}
