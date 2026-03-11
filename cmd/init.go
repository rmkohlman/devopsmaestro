package cmd

import (
	"devopsmaestro/db"
	"devopsmaestro/pkg/paths"
	"fmt"
	"io/fs"
	"log/slog"
	"os"

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
		pc := paths.New(homeDir)
		dvmDir := pc.Root()
		dirs := []string{
			dvmDir,
			pc.TemplatesDir(),
			pc.NvimTemplatesDir(),
			pc.ShellTemplatesDir(),
			pc.BackupsDir(),
			pc.LogsDir(),
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
		configPath := pc.ConfigFile()
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			slog.Debug("creating default config", "path", configPath)
			defaultConfig := fmt.Sprintf(`# DevOpsMaestro Configuration
database:
  type: sqlite
  path: ~/%s/%s

store: sql

runtime:
  type: auto  # Options: auto, docker, containerd, kubernetes
  # auto: Automatically detects runtime (checks for docker.sock or containerd.sock)

templates:
  nvim: ~/%s/templates/nvim
  shell: ~/%s/templates/shell
`, paths.DVMDirName, paths.DatabaseFile, paths.DVMDirName, paths.DVMDirName)
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
		ds, dsErr := getDataStore(cmd)
		if dsErr != nil {
			slog.Error("dataStore not initialized in context", "error", dsErr)
			fmt.Println("Error: DataStore not initialized")
			return
		}

		driver := ds.Driver()
		if driver == nil {
			slog.Error("driver not available from dataStore")
			fmt.Println("Error: Database driver not available")
			return
		}

		// Get migrations filesystem from context
		ctx := cmd.Context()
		migrationsFS := ctx.Value("migrationsFS").(fs.FS)
		if migrationsFS == nil {
			slog.Error("migrations filesystem not available in context")
			fmt.Println("Error: Migrations filesystem not available")
			return
		}

		fmt.Println("Running database migrations...")
		slog.Debug("running database migrations")
		if err := db.RunMigrations(driver, migrationsFS); err != nil {
			slog.Error("failed to run database migrations", "error", err)
			fmt.Printf("Error: Failed to run database migrations: %v\n", err)
			return
		}
		slog.Info("database migrations completed")

		fmt.Println("\n✓ DevOpsMaestro initialized successfully!")
		slog.Info("initialization completed successfully", "config_dir", dvmDir)
		fmt.Println("\nNext steps:")
		fmt.Println("  1. Copy your dev environment templates:")
		fmt.Println("     (We'll create a script for this)")
		fmt.Println("  2. Create your first app:")
		fmt.Println("     dvm create app <name> --from-cwd")
		fmt.Println("  3. Start coding:")
		fmt.Println("     dvm use workspace main && dvm attach")
	},
}

func init() {
	adminCmd.AddCommand(initCmd)
}
