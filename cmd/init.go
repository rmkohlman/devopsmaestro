package cmd

import (
	"devopsmaestro/db"
	"devopsmaestro/pkg/paths"
	"devopsmaestro/render"
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
		render.Progress("Initializing DevOpsMaestro...")
		slog.Info("starting initialization")

		// Get home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			slog.Error("failed to get home directory", "error", err)
			render.Errorf("Failed to get home directory: %v", err)
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

		render.Progress("Creating directory structure...")
		for _, dir := range dirs {
			slog.Debug("creating directory", "path", dir)
			if err := os.MkdirAll(dir, 0755); err != nil {
				slog.Error("failed to create directory", "path", dir, "error", err)
				render.Errorf("Failed to create directory %s: %v", dir, err)
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
				render.Errorf("Failed to create config.yaml: %v", err)
				return
			}
			render.Success("Created config.yaml")
			slog.Info("created config file", "path", configPath)
		} else {
			slog.Debug("config file already exists", "path", configPath)
		}

		// Initialize database
		ds, dsErr := getDataStore(cmd)
		if dsErr != nil {
			slog.Error("dataStore not initialized in context", "error", dsErr)
			render.Error("DataStore not initialized")
			return
		}

		driver := ds.Driver()
		if driver == nil {
			slog.Error("driver not available from dataStore")
			render.Error("Database driver not available")
			return
		}

		// Get migrations filesystem from context
		ctx := cmd.Context()
		migrationsFS := ctx.Value("migrationsFS").(fs.FS)
		if migrationsFS == nil {
			slog.Error("migrations filesystem not available in context")
			render.Error("Migrations filesystem not available")
			return
		}

		render.Progress("Running database migrations...")
		slog.Debug("running database migrations")
		if err := db.RunMigrations(driver, migrationsFS); err != nil {
			slog.Error("failed to run database migrations", "error", err)
			render.Errorf("Failed to run database migrations: %v", err)
			return
		}
		slog.Info("database migrations completed")

		render.Blank()
		render.Success("DevOpsMaestro initialized successfully!")
		slog.Info("initialization completed successfully", "config_dir", dvmDir)
		render.Blank()
		render.Info("Next steps:")
		render.Info("  1. Copy your dev environment templates:")
		render.Info("     (We'll create a script for this)")
		render.Info("  2. Create your first app:")
		render.Info("     dvm create app <name> --from-cwd")
		render.Info("  3. Start coding:")
		render.Info("     dvm use workspace main && dvm attach")
	},
}

func init() {
	adminCmd.AddCommand(initCmd)
}
