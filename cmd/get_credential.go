package cmd

import (
	"fmt"

	"devopsmaestro/render"

	"github.com/spf13/cobra"
)

// getCredentialsCmd lists credentials (plural form)
var getCredentialsCmd = &cobra.Command{
	Use:     "credentials",
	Aliases: []string{"cred", "creds"},
	Short:   "List credentials",
	Long: `List credential configurations.

By default, lists credentials filtered by scope.
Use --all/-A to list credentials across all scopes.

Examples:
  dvm get credentials --all                    # List all credentials
  dvm get credentials -A                       # Same (short form)
  dvm get credentials --app my-api             # List credentials for an app
  dvm get credentials --ecosystem prod         # List credentials for an ecosystem
  dvm get creds -A                             # Alias`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get DataStore from context
		ds, err := getDataStore(cmd)
		if err != nil {
			return fmt.Errorf("dataStore not found in context")
		}

		allFlag, _ := cmd.Flags().GetBool("all")

		if allFlag {
			// List all credentials across all scopes
			creds, err := ds.ListAllCredentials()
			if err != nil {
				return fmt.Errorf("failed to list credentials: %w", err)
			}

			if len(creds) == 0 {
				render.Info("No credentials found")
				return nil
			}

			for _, c := range creds {
				render.Plainf("  %s  (scope: %s, source: %s)", c.Name, c.ScopeType, c.Source)
			}
			return nil
		}

		// Filter by scope
		scopeType, scopeID, err := resolveCredentialScopeFromFlags(cmd, ds)
		if err != nil {
			return err
		}

		creds, err := ds.ListCredentialsByScope(scopeType, scopeID)
		if err != nil {
			return fmt.Errorf("failed to list credentials: %w", err)
		}

		if len(creds) == 0 {
			render.Info(fmt.Sprintf("No credentials found for %s scope", scopeType))
			return nil
		}

		for _, c := range creds {
			render.Plainf("  %s  (source: %s)", c.Name, c.Source)
		}
		return nil
	},
}

// getCredentialCmd gets a single credential by name (singular form)
var getCredentialCmd = &cobra.Command{
	Use:     "credential <name>",
	Aliases: []string{"cred"},
	Short:   "Get a specific credential",
	Long: `Get a specific credential by name within a scope.

Requires exactly one scope flag to identify which credential to retrieve.

Examples:
  dvm get credential github-token --app my-api
  dvm get credential api-key --ecosystem prod
  dvm get cred db-pass --domain backend`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		credName := args[0]

		// Get DataStore from context
		ds, err := getDataStore(cmd)
		if err != nil {
			return fmt.Errorf("dataStore not found in context")
		}

		// Resolve scope
		scopeType, scopeID, err := resolveCredentialScopeFromFlags(cmd, ds)
		if err != nil {
			return err
		}

		cred, err := ds.GetCredential(scopeType, scopeID, credName)
		if err != nil {
			return fmt.Errorf("credential '%s' not found in %s scope: %w", credName, scopeType, err)
		}

		// Display credential details
		render.Plainf("Name:      %s", cred.Name)
		render.Plainf("Scope:     %s (ID: %d)", cred.ScopeType, cred.ScopeID)
		render.Plainf("Source:    %s", cred.Source)
		if cred.Service != nil {
			render.Plainf("Service:   %s", *cred.Service)
		}
		if cred.EnvVar != nil {
			render.Plainf("EnvVar:    %s", *cred.EnvVar)
		}
		if cred.Description != nil {
			render.Plainf("Desc:      %s", *cred.Description)
		}

		return nil
	},
}

func init() {
	getCmd.AddCommand(getCredentialsCmd)
	getCmd.AddCommand(getCredentialCmd)

	// --all flag for list command
	getCredentialsCmd.Flags().BoolP("all", "A", false, "List all credentials across all scopes")

	// Scope flags for both commands
	addCredentialScopeFlags(getCredentialsCmd)
	addCredentialScopeFlags(getCredentialCmd)
}
