package main

import (
	"devopsmaestro/cmd"
	"devopsmaestro/db"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func run(dbInstance db.Database, dataStoreInstance db.DataStore, executor cmd.Executor) int {
	// Execute the root command of the CLI tool
	cmd.Execute(&dbInstance, &dataStoreInstance, &executor)

	return 0
}
func main() {
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
