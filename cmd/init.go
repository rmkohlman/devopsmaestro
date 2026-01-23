package cmd

import (
	"devopsmaestro/db"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize DevOpsMaestro",
	Long:  `Initialize DevOpsMaestro by setting up the database, configuration, and template directories.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Initializing DevOpsMaestro...")

		// Get home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Error: Failed to get home directory: %v\n", err)
			return
		}

		// Create ~/.devopsmaestro directory structure
		dvmDir := filepath.Join(homeDir, ".devopsmaestro")
		dirs := []string{
			dvmDir,
			filepath.Join(dvmDir, "templates"),
			filepath.Join(dvmDir, "templates", "nvim"),
			filepath.Join(dvmDir, "templates", "shell"),
			filepath.Join(dvmDir, "backups"),
		}

		fmt.Println("Creating directory structure...")
		for _, dir := range dirs {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Printf("Error: Failed to create directory %s: %v\n", dir, err)
				return
			}
		}

		// Create default config.yaml
		configPath := filepath.Join(dvmDir, "config.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			defaultConfig := `# DevOpsMaestro Configuration
database:
  type: sqlite
  path: ~/.devopsmaestro/devopsmaestro.db

store: sql

runtime:
  type: auto  # Options: auto, docker, containerd, kubernetes
  # auto: Automatically detects runtime (checks for docker.sock or containerd.sock)

templates:
  nvim: ~/.devopsmaestro/templates/nvim
  shell: ~/.devopsmaestro/templates/shell
`
			if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
				fmt.Printf("Error: Failed to create config.yaml: %v\n", err)
				return
			}
			fmt.Println("✓ Created config.yaml")
		}

		// Initialize database
		ctx := cmd.Context()
		database := ctx.Value("database").(*db.Database)
		if database == nil {
			fmt.Println("Error: Database not initialized")
			return
		}

		fmt.Println("Running database migrations...")
		if err := db.InitializeDatabase(*database); err != nil {
			fmt.Printf("Error: Failed to initialize database: %v\n", err)
			return
		}

		fmt.Println("\n✓ DevOpsMaestro initialized successfully!")
		fmt.Println("\nNext steps:")
		fmt.Println("  1. Copy your dev environment templates:")
		fmt.Println("     (We'll create a script for this)")
		fmt.Println("  2. Create your first project:")
		fmt.Println("     dvm create project <name> --from-cwd")
		fmt.Println("  3. Start coding:")
		fmt.Println("     dvm use workspace main && dvm attach")
	},
}

func init() {
	adminCmd.AddCommand(initCmd)
}
