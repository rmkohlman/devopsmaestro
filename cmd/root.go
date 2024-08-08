package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "DevOpsMaestro",
	Short: "DevOpsMaestro is a CLI tool for managing development environments and workflows",
	Long:  `DevOpsMaestro is a sophisticated CLI tool designed to streamline the management of development environments, testing, deployments, and maintenance of code and software.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default action when no subcommands are provided
		fmt.Println("Welcome to DevOpsMaestro! Use --help to see available commands.")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Add all subcommands here
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(describeCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(attachCmd)
	rootCmd.AddCommand(detachCmd)
	rootCmd.AddCommand(useCmd)
	rootCmd.AddCommand(releaseCmd)
	rootCmd.AddCommand(projectCmd)
	rootCmd.AddCommand(workspaceCmd)
	rootCmd.AddCommand(dependencyCmd)
	rootCmd.AddCommand(taskCmd)
	rootCmd.AddCommand(workflowCmd)
	rootCmd.AddCommand(pipelineCmd)
	rootCmd.AddCommand(orchestrationCmd)
	rootCmd.AddCommand(prototypeCmd)
	rootCmd.AddCommand(dataLakeCmd)
	rootCmd.AddCommand(dataStoreCmd)
	rootCmd.AddCommand(dataRecordCmd)
	rootCmd.AddCommand(contextCmd)
	rootCmd.AddCommand(storageCmd)
}
