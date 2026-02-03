package cmd

import (
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/ui"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// printDeprecationWarning prints a deprecation warning to stderr
func printDeprecationWarning(oldCmd, newCmd string) {
	warning := fmt.Sprintf("⚠️  '%s' is deprecated. Use '%s' instead.\n\n", oldCmd, newCmd)
	fmt.Fprint(os.Stderr, ui.WarningStyle.Render(warning))
}

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage nvim plugins",
	Long: `Create, update, list, and manage reusable nvim plugin definitions.

NOTE: This command is deprecated. Use the new namespaced commands instead:
  dvm get nvim plugins          # List plugins
  dvm get nvim plugin <name>    # Get specific plugin
  dvm apply nvim plugin -f      # Apply from file
  dvm delete nvim plugin <name> # Delete plugin`,
}

var pluginApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a plugin definition from file or stdin",
	Long: `Apply a plugin definition from a YAML file or stdin to the database.
If the plugin already exists, it will be updated.

DEPRECATED: Use 'dvm apply nvim plugin -f <file>' instead.

Examples:
  # Apply from file
  dvm plugin apply -f telescope.yaml
  
  # Apply from stdin
  cat telescope.yaml | dvm plugin apply -f -
  echo "apiVersion: devopsmaestro.io/v1..." | dvm plugin apply -f -
  
  # Apply multiple files
  dvm plugin apply -f plugin1.yaml -f plugin2.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		printDeprecationWarning("dvm plugin apply", "dvm apply nvim plugin -f <file>")

		files, _ := cmd.Flags().GetStringSlice("filename")

		if len(files) == 0 {
			return fmt.Errorf("must specify at least one file with -f flag")
		}

		for _, filePath := range files {
			if err := applyPluginFromFile(cmd, filePath); err != nil {
				return err
			}
		}

		return nil
	},
}

func applyPluginFromFile(cmd *cobra.Command, filePath string) error {
	var data []byte
	var err error

	// Read from stdin if filePath is "-"
	if filePath == "-" {
		data, err = os.ReadFile("/dev/stdin")
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %v", err)
		}
	} else {
		// Read YAML file
		data, err = os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}
	}

	// Parse YAML
	var pluginYAML models.NvimPluginYAML
	if err := yaml.Unmarshal(data, &pluginYAML); err != nil {
		return fmt.Errorf("failed to parse YAML: %v", err)
	}

	// Validate kind
	if pluginYAML.Kind != "NvimPlugin" {
		return fmt.Errorf("invalid kind: expected 'NvimPlugin', got '%s'", pluginYAML.Kind)
	}

	// Convert to database model
	plugin := &models.NvimPluginDB{}
	if err := plugin.FromYAML(pluginYAML); err != nil {
		return fmt.Errorf("failed to convert plugin: %v", err)
	}

	// Get database connection
	database, err := db.InitializeDBConnection()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer database.Close()

	datastore, err := db.StoreFactory(database)
	if err != nil {
		return fmt.Errorf("failed to create datastore: %v", err)
	}

	// Check if plugin already exists
	existing, err := datastore.GetPluginByName(plugin.Name)
	if err == nil && existing != nil {
		// Update existing plugin
		plugin.ID = existing.ID
		plugin.CreatedAt = existing.CreatedAt
		if err := datastore.UpdatePlugin(plugin); err != nil {
			return fmt.Errorf("failed to update plugin: %v", err)
		}
		source := filePath
		if filePath == "-" {
			source = "stdin"
		}
		fmt.Printf("✓ Plugin '%s' configured (from %s)\n", plugin.Name, source)
	} else {
		// Create new plugin
		if err := datastore.CreatePlugin(plugin); err != nil {
			return fmt.Errorf("failed to create plugin: %v", err)
		}
		source := filePath
		if filePath == "-" {
			source = "stdin"
		}
		fmt.Printf("✓ Plugin '%s' created (from %s)\n", plugin.Name, source)
	}

	return nil
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available plugins",
	Long: `List all nvim plugins stored in the database.

DEPRECATED: Use 'dvm get nvim plugins' instead.

Examples:
  # List as table (default)
  dvm plugin list
  
  # List as YAML
  dvm plugin list -o yaml
  
  # List as JSON
  dvm plugin list -o json
  
  # Filter by category
  dvm plugin list --category lsp
  dvm plugin list --category lsp -o yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		printDeprecationWarning("dvm plugin list", "dvm get nvim plugins")

		// Get database connection
		database, err := db.InitializeDBConnection()
		if err != nil {
			return fmt.Errorf("failed to connect to database: %v", err)
		}
		defer database.Close()

		datastore, err := db.StoreFactory(database)
		if err != nil {
			return fmt.Errorf("failed to create datastore: %v", err)
		}

		// Get category filter if provided
		category, _ := cmd.Flags().GetString("category")

		var plugins []*models.NvimPluginDB
		if category != "" {
			plugins, err = datastore.ListPluginsByCategory(category)
		} else {
			plugins, err = datastore.ListPlugins()
		}

		if err != nil {
			return fmt.Errorf("failed to list plugins: %v", err)
		}

		if len(plugins) == 0 {
			fmt.Println("No plugins found")
			return nil
		}

		// Get output format
		outputFormat, _ := cmd.Flags().GetString("output")

		// Output in requested format
		switch outputFormat {
		case "yaml":
			return outputPluginsYAML(plugins)
		case "json":
			return outputPluginsJSON(plugins)
		case "table", "":
			return outputPluginsTable(plugins)
		default:
			return fmt.Errorf("unsupported output format: %s (supported: table, yaml, json)", outputFormat)
		}
	},
}

// NOTE: outputPluginsTable, outputPluginsYAML, and outputPluginsJSON functions
// are defined in get.go with colorized output support via the ui package.

var pluginGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get a plugin definition",
	Long: `Retrieve a plugin definition from the database and output in specified format.

DEPRECATED: Use 'dvm get nvim plugin <name>' instead.

Examples:
  # Get as YAML (default)
  dvm plugin get telescope
  
  # Get as JSON
  dvm plugin get telescope -o json
  
  # Save to file
  dvm plugin get telescope -o yaml > telescope.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		printDeprecationWarning("dvm plugin get", "dvm get nvim plugin <name>")

		name := args[0]
		outputFormat, _ := cmd.Flags().GetString("output")

		// Get database connection
		database, err := db.InitializeDBConnection()
		if err != nil {
			return fmt.Errorf("failed to connect to database: %v", err)
		}
		defer database.Close()

		datastore, err := db.StoreFactory(database)
		if err != nil {
			return fmt.Errorf("failed to create datastore: %v", err)
		}

		// Get plugin
		plugin, err := datastore.GetPluginByName(name)
		if err != nil {
			return fmt.Errorf("failed to get plugin: %v", err)
		}

		// Convert to YAML structure
		pluginYAML, err := plugin.ToYAML()
		if err != nil {
			return fmt.Errorf("failed to convert plugin to YAML: %v", err)
		}

		// Output in requested format
		var data []byte
		switch outputFormat {
		case "json":
			data, err = json.MarshalIndent(pluginYAML, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %v", err)
			}
		case "yaml", "":
			data, err = yaml.Marshal(pluginYAML)
			if err != nil {
				return fmt.Errorf("failed to marshal YAML: %v", err)
			}
		default:
			return fmt.Errorf("unsupported output format: %s (supported: yaml, json)", outputFormat)
		}

		fmt.Println(string(data))
		return nil
	},
}

var pluginDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a plugin definition from DVM's database",
	Long: `Delete a plugin definition from DVM's database.

DEPRECATED: Use 'dvm delete nvim plugin <name>' instead.

This removes the plugin YAML definition that DVM stores for generating
nvim configurations in workspace containers. It does NOT affect:
- Your local nvim installation
- Any existing container images
- Plugins already installed in running containers

The plugin definition can be re-added later with 'dvm apply nvim plugin -f'.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		printDeprecationWarning("dvm plugin delete", "dvm delete nvim plugin <name>")

		name := args[0]

		// Get database connection
		database, err := db.InitializeDBConnection()
		if err != nil {
			return fmt.Errorf("failed to connect to database: %v", err)
		}
		defer database.Close()

		datastore, err := db.StoreFactory(database)
		if err != nil {
			return fmt.Errorf("failed to create datastore: %v", err)
		}

		// Confirm deletion
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Delete plugin definition '%s' from DVM database? (y/N): ", name)
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Aborted")
				return nil
			}
		}

		// Delete plugin
		if err := datastore.DeletePlugin(name); err != nil {
			return fmt.Errorf("failed to delete plugin: %v", err)
		}

		fmt.Printf("✓ Plugin definition '%s' removed from DVM database\n", name)
		return nil
	},
}

var pluginEditCmd = &cobra.Command{
	Use:   "edit [name]",
	Short: "Edit a plugin definition in your editor",
	Long: `Open a plugin definition in your default editor (EDITOR env var).
After editing, the plugin will be automatically applied to the database.

Examples:
  dvm plugin edit telescope
  EDITOR=vim dvm plugin edit mason`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Get database connection
		database, err := db.InitializeDBConnection()
		if err != nil {
			return fmt.Errorf("failed to connect to database: %v", err)
		}
		defer database.Close()

		datastore, err := db.StoreFactory(database)
		if err != nil {
			return fmt.Errorf("failed to create datastore: %v", err)
		}

		// Get plugin
		plugin, err := datastore.GetPluginByName(name)
		if err != nil {
			return fmt.Errorf("failed to get plugin: %v", err)
		}

		// Convert to YAML
		pluginYAML, err := plugin.ToYAML()
		if err != nil {
			return fmt.Errorf("failed to convert plugin to YAML: %v", err)
		}

		data, err := yaml.Marshal(pluginYAML)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %v", err)
		}

		// Create temp file
		tmpfile, err := os.CreateTemp("", fmt.Sprintf("dvm-plugin-%s-*.yaml", name))
		if err != nil {
			return fmt.Errorf("failed to create temp file: %v", err)
		}
		defer os.Remove(tmpfile.Name())

		if _, err := tmpfile.Write(data); err != nil {
			return fmt.Errorf("failed to write temp file: %v", err)
		}
		tmpfile.Close()

		// Get editor from env
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi" // Default to vi
		}

		// Open editor
		editCmd := exec.Command(editor, tmpfile.Name())
		editCmd.Stdin = os.Stdin
		editCmd.Stdout = os.Stdout
		editCmd.Stderr = os.Stderr

		if err := editCmd.Run(); err != nil {
			return fmt.Errorf("editor exited with error: %v", err)
		}

		// Read edited content
		editedData, err := os.ReadFile(tmpfile.Name())
		if err != nil {
			return fmt.Errorf("failed to read edited file: %v", err)
		}

		// Parse and apply
		var editedYAML models.NvimPluginYAML
		if err := yaml.Unmarshal(editedData, &editedYAML); err != nil {
			return fmt.Errorf("failed to parse edited YAML: %v", err)
		}

		// Validate kind
		if editedYAML.Kind != "NvimPlugin" {
			return fmt.Errorf("invalid kind: expected 'NvimPlugin', got '%s'", editedYAML.Kind)
		}

		// Convert and update
		updatedPlugin := &models.NvimPluginDB{}
		if err := updatedPlugin.FromYAML(editedYAML); err != nil {
			return fmt.Errorf("failed to convert plugin: %v", err)
		}

		updatedPlugin.ID = plugin.ID
		updatedPlugin.CreatedAt = plugin.CreatedAt

		if err := datastore.UpdatePlugin(updatedPlugin); err != nil {
			return fmt.Errorf("failed to update plugin: %v", err)
		}

		fmt.Printf("✓ Plugin '%s' updated successfully\n", updatedPlugin.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pluginCmd)
	pluginCmd.AddCommand(pluginApplyCmd)
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginGetCmd)
	pluginCmd.AddCommand(pluginEditCmd)
	pluginCmd.AddCommand(pluginDeleteCmd)

	// Add flags
	pluginApplyCmd.Flags().StringSliceP("filename", "f", []string{}, "Plugin YAML file(s) to apply (use '-' for stdin)")
	pluginApplyCmd.MarkFlagRequired("filename")
	pluginGetCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml or json")
	pluginListCmd.Flags().StringP("output", "o", "table", "Output format: table, yaml, or json")
	pluginListCmd.Flags().StringP("category", "c", "", "Filter by category")
	pluginDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}
