package main

import (
	"devopsmaestro/cmd"
	"devopsmaestro/db"
	"devopsmaestro/pkg/paths"
	"devopsmaestro/render"
	"fmt"
	"io/fs"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

// Version information (set via ldflags during build)
var (
	Version   = "dev"
	BuildTime = "unknown"
	Commit    = "unknown"
)

func run(dataStoreInstance db.DataStore, executor cmd.Executor, migrationsFS fs.FS) int {
	// Execute the root command of the CLI tool
	cmd.Execute(&dataStoreInstance, &executor, migrationsFS)

	return 0
}

func loadConfig() error {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Set config path to ~/.devopsmaestro
	configPath := paths.New(homeDir).Root()
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
			render.Errorf("Failed to load configuration: %v", err)
			os.Exit(1)
		}
	}

	// Set default values if config is not found (for init command)
	if viper.GetString("database.type") == "" {
		viper.Set("database.type", "sqlite")
		viper.Set("database.path", "~/"+paths.DVMDirName+"/"+paths.DatabaseFile)
		viper.Set("store", "sql")
	}

	var dataStoreInstance db.DataStore
	var executor cmd.Executor

	// Only initialize database for commands that need it
	if !skipDB {
		// Initialize the database connection and DataStore
		var err error
		dataStoreInstance, err = db.CreateDataStore()
		if err != nil {
			render.Errorf("Failed to initialize database: %v", err)
			os.Exit(1)
		}

		// Ensure the database connection is closed when the program exits
		defer func() {
			if dataStoreInstance != nil {
				if err := dataStoreInstance.Close(); err != nil {
					render.Warningf("Failed to close database connection: %v", err)
				}
			}
		}()

		executor = cmd.NewExecutor(dataStoreInstance)
	}

	// Get migrations subdirectory from embedded filesystem
	migrationsSubFS, err := fs.Sub(MigrationsFS, "db/migrations")
	if err != nil {
		render.Errorf("Failed to access embedded migrations: %v", err)
		os.Exit(1)
	}

	os.Exit(run(dataStoreInstance, executor, migrationsSubFS))
}
