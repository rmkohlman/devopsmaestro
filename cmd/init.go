package cmd

import (
	"devopsmaestro/db"
	"fmt"
	"io/fs"
	"log/slog"
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
		slog.Info("starting initialization")

		// Get home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			slog.Error("failed to get home directory", "error", err)
			fmt.Printf("Error: Failed to get home directory: %v\n", err)
			return
		}
		slog.Debug("resolved home directory", "path", homeDir)

		// Create ~/.devopsmaestro directory structure
		dvmDir := filepath.Join(homeDir, ".devopsmaestro")
		dirs := []string{
			dvmDir,
			filepath.Join(dvmDir, "templates"),
			filepath.Join(dvmDir, "templates", "nvim"),
			filepath.Join(dvmDir, "templates", "shell"),
			filepath.Join(dvmDir, "backups"),
			filepath.Join(dvmDir, "logs"),
		}

		fmt.Println("Creating directory structure...")
		for _, dir := range dirs {
			slog.Debug("creating directory", "path", dir)
			if err := os.MkdirAll(dir, 0755); err != nil {
				slog.Error("failed to create directory", "path", dir, "error", err)
				fmt.Printf("Error: Failed to create directory %s: %v\n", dir, err)
				return
			}
		}
		slog.Debug("directory structure created", "root", dvmDir)

		// Create default config.yaml
		configPath := filepath.Join(dvmDir, "config.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			slog.Debug("creating default config", "path", configPath)
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
				slog.Error("failed to create config.yaml", "path", configPath, "error", err)
				fmt.Printf("Error: Failed to create config.yaml: %v\n", err)
				return
			}
			fmt.Println("✓ Created config.yaml")
			slog.Info("created config file", "path", configPath)
		} else {
			slog.Debug("config file already exists", "path", configPath)
		}

		// Initialize database
		ctx := cmd.Context()
		database := ctx.Value("database").(*db.Database)
		if database == nil {
			slog.Error("database not initialized in context")
			fmt.Println("Error: Database not initialized")
			return
		}

		// Get migrations filesystem from context
		migrationsFS := ctx.Value("migrationsFS").(fs.FS)
		if migrationsFS == nil {
			slog.Error("migrations filesystem not available in context")
			fmt.Println("Error: Migrations filesystem not available")
			return
		}

		fmt.Println("Running database migrations...")
		slog.Debug("running database migrations")
		if err := db.InitializeDatabase(*database, migrationsFS); err != nil {
			slog.Error("failed to initialize database", "error", err)
			fmt.Printf("Error: Failed to initialize database: %v\n", err)
			return
		}
		slog.Info("database migrations completed")

		fmt.Println("\n✓ DevOpsMaestro initialized successfully!")
		slog.Info("initialization completed successfully", "config_dir", dvmDir)
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
