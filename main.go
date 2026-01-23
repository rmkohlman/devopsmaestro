package main

import (
	"devopsmaestro/cmd"
	"devopsmaestro/db"
	"fmt"
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

func run(dbInstance db.Database, dataStoreInstance db.DataStore, executor cmd.Executor) int {
	// Execute the root command of the CLI tool
	cmd.Execute(&dbInstance, &dataStoreInstance, &executor)

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

	// Load configuration
	if err := loadConfig(); err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Set default values if config is not found (for init command)
	if viper.GetString("database.type") == "" {
		viper.Set("database.type", "sqlite")
		viper.Set("database.path", "~/.devopsmaestro/devopsmaestro.db")
		viper.Set("store", "sql")
	}

	// Initialize the database connection
	dbInstance, err := db.InitializeDBConnection()
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
	dataStoreInstance, err := db.StoreFactory(dbInstance)
	if err != nil {
		fmt.Printf("Failed to create DataStore instance: %v\n", err)
		os.Exit(1)
	}

	executor := cmd.NewExecutor(dataStoreInstance, dbInstance)

	os.Exit(run(dbInstance, dataStoreInstance, executor))
}
