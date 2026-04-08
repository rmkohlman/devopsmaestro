package cmd

import (
	"context"
	"devopsmaestro/db"
	"devopsmaestro/pkg/registry"
	"fmt"
	"github.com/rmkohlman/MaestroSDK/paths"
	"github.com/rmkohlman/MaestroSDK/render"
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

		// Get path configuration
		pc, err := paths.Default()
		if err != nil {
			slog.Error("failed to get home directory", "error", err)
			render.Errorf("Failed to get home directory: %v", err)
			return
		}

		// Create ~/.devopsmaestro directory structure
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
			if err := os.MkdirAll(dir, 0700); err != nil {
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
			if err := os.WriteFile(configPath, []byte(defaultConfig), 0600); err != nil {
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
		migrationsFS, fsErr := getMigrationsFSFromContext(ctx)
		if fsErr != nil {
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

		// Bootstrap default registries (non-fatal)
		bootstrapAllDefaultRegistries(ctx, ds, ds, "on-demand")

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

// bootstrapAllDefaultRegistries creates default registries for all supported
// package manager types. Each alias resolves to a concrete registry type:
// oci→zot, pypi→devpi, npm→verdaccio, go→athens, http→squid.
// Failures are collected and returned (non-fatal), one per alias that failed.
func bootstrapAllDefaultRegistries(ctx context.Context, registryStore db.RegistryStore, defaultsStore db.DefaultsStore, lifecycle string) []error {
	aliases := []string{"oci", "pypi", "npm", "go", "http"}
	render.Progress("Bootstrapping default registries...")

	errs := make([]error, 0)
	var created, existing int
	for _, alias := range aliases {
		wasCreated, err := registry.EnsureDefaultRegistry(ctx, registryStore, defaultsStore, alias, lifecycle)
		if err != nil {
			slog.Warn("failed to create default registry", "alias", alias, "error", err)
			render.Warning(fmt.Sprintf("Could not create default %s registry: %s", alias, err.Error()))
			errs = append(errs, err)
			continue
		}
		if wasCreated {
			created++
			slog.Info("created default registry", "alias", alias)
		} else {
			existing++
		}
	}

	if created > 0 {
		render.Success(fmt.Sprintf("Created %d default registries (oci, pypi, npm, go, http)", created))
	}
	if existing > 0 {
		render.Info(fmt.Sprintf("%d registries already configured", existing))
	}

	return errs
}
