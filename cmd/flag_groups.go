package cmd

import "github.com/spf13/cobra"

// AddOutputFlag registers the standard -o/--output flag on a command.
// Use this for commands that support table, yaml, json, etc. output formats.
// The default value is the format used when --output is not specified.
func AddOutputFlag(cmd *cobra.Command, defaultVal string) {
	cmd.Flags().StringP("output", "o", defaultVal, "Output format (table|yaml|json)")
}

// AddForceConfirmFlag registers the --force flag for skipping confirmation prompts.
// The flag is a simple bool retrieved via cmd.Flags().GetBool("force").
func AddForceConfirmFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("force", false, "Skip confirmation prompt")
}

// AddDryRunFlag registers the --dry-run flag for preview-only operations.
// The dest pointer receives the flag value directly via BoolVar binding.
func AddDryRunFlag(cmd *cobra.Command, dest *bool) {
	cmd.Flags().BoolVar(dest, "dry-run", false, "Preview changes without applying")
}

// AddAllFlag registers the -A/--all flag for listing across all scopes.
// The description should explain what "all" means in context (e.g. "List all apps from all domains").
func AddAllFlag(cmd *cobra.Command, description string) {
	cmd.Flags().BoolP("all", "A", false, description)
}
