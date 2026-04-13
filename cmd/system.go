package cmd

import (
	"github.com/spf13/cobra"
)

var (
	systemDescription string
	systemDomain      string
	systemEcosystem   string
)

// Dry-run flags for system commands
var (
	createSystemDryRun bool
	useSystemDryRun    bool
	deleteSystemDryRun bool
)

// createSystemCmd creates a new system
var createSystemCmd = &cobra.Command{
	Use:     "system <name>",
	Aliases: []string{"sys"},
	Short:   "Create a new system",
	Long: `Create a new system within a domain.

A system represents an organizational grouping within a domain:
Ecosystem -> Domain -> System -> App -> Workspace.
Systems group related applications together at a finer level than domains.

Examples:
  # Create a system in the active domain
  dvm create system auth-service
  dvm create sys auth-service             # Short form

  # Create a system in a specific domain
  dvm create system auth-service --domain backend

  # Create with description and ecosystem
  dvm create system auth-service --description "Authentication services" --ecosystem prod`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return createSystem(cmd, args[0])
	},
}

// getSystemsCmd lists all systems
var getSystemsCmd = &cobra.Command{
	Use:     "systems",
	Aliases: []string{"sys"},
	Short:   "List all systems",
	Long: `List all systems, optionally filtered by domain.

Examples:
  dvm get systems                       # List systems in active domain
  dvm get sys                           # Short form
  dvm get systems --domain backend
  dvm get systems -A                    # List all systems across all domains
  dvm get systems --all                 # Same as -A
  dvm get systems -o yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getSystems(cmd)
	},
}

// getSystemCmd gets a specific system
var getSystemCmd = &cobra.Command{
	Use:     "system <name>",
	Aliases: []string{"sys"},
	Short:   "Get a specific system",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getSystem(cmd, args[0])
	},
}

// useSystemCmd switches the active system
var useSystemCmd = &cobra.Command{
	Use:     "system <name>",
	Aliases: []string{"sys"},
	Short:   "Switch to a system",
	Long: `Set the specified system as the active context.

Use 'none' as the name to clear the system context (also clears app and workspace).

Examples:
  dvm use system auth-service    # Set active system
  dvm use sys auth-service       # Short form
  dvm use system none            # Clear system context
  dvm use system auth --export   # Print 'export DVM_SYSTEM=auth' for shell eval`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return useSystem(cmd, args[0])
	},
}

// deleteSystemCmd deletes a system
var deleteSystemCmd = &cobra.Command{
	Use:     "system <name>",
	Aliases: []string{"sys"},
	Short:   "Delete a system",
	Long: `Delete a system by name.

By default, you will be prompted for confirmation. Use --force to skip.

Examples:
  dvm delete system auth-service
  dvm delete sys auth-service                  # Short form
  dvm delete system auth-service --domain backend
  dvm delete system auth-service --force       # Skip confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteSystem(cmd, args[0])
	},
}

func init() {
	// Add system commands to parent commands
	createCmd.AddCommand(createSystemCmd)
	getCmd.AddCommand(getSystemsCmd)
	getCmd.AddCommand(getSystemCmd)
	useCmd.AddCommand(useSystemCmd)

	if deleteCmd != nil {
		deleteCmd.AddCommand(deleteSystemCmd)
	}

	// System creation flags
	createSystemCmd.Flags().StringVar(&systemDescription, "description", "", "System description")
	createSystemCmd.Flags().StringVar(&systemDomain, "domain", "", "Domain name (defaults to active domain)")
	createSystemCmd.Flags().StringVar(&systemEcosystem, "ecosystem", "", "Ecosystem name (defaults to active ecosystem)")
	AddDryRunFlag(createSystemCmd, &createSystemDryRun)

	// Use system flags
	AddDryRunFlag(useSystemCmd, &useSystemDryRun)
	useSystemCmd.Flags().Bool("export", false, "Print 'export DVM_SYSTEM=<name>' for shell eval")

	// System get/delete flags
	getSystemsCmd.Flags().StringP("domain", "d", "", "Domain name (defaults to active domain)")
	AddAllFlag(getSystemsCmd, "List systems from all domains")
	getSystemsCmd.Flags().BoolVar(&showTheme, "show-theme", false, "Show theme resolution information")
	getSystemCmd.Flags().StringP("domain", "d", "", "Domain name (defaults to active domain)")
	getSystemCmd.Flags().BoolVar(&showTheme, "show-theme", false, "Show theme resolution information")
	deleteSystemCmd.Flags().StringP("domain", "d", "", "Domain name (defaults to active domain)")
	AddForceConfirmFlag(deleteSystemCmd)
	AddDryRunFlag(deleteSystemCmd, &deleteSystemDryRun)
}
