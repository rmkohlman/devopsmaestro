package cmd

import (
	"context"
	"devopsmaestro/db"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dvm",
	Short: "DevOpsMaestro CLI",
	Long: `DevOpsMaestro (dvm) is a CLI tool designed to manage development environments,
testing, deployments, and maintenance of code and software projects. It allows you to
create, manage, and deploy workspaces, projects, dependencies, and more.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(database *db.Database, dataStore *db.DataStore, executor *Executor) {
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// Set the database, dataStore, and executor for all commands
		ctx := context.WithValue(cmd.Context(), "database", database)
		ctx = context.WithValue(ctx, "dataStore", dataStore)
		ctx = context.WithValue(ctx, "executor", executor)
		cmd.SetContext(ctx)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Initialize any flags or configuration settings here
}
