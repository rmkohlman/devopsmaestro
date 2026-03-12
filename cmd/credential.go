package cmd

import (
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/render"

	"github.com/spf13/cobra"
)

// CredentialScopeFlags holds the values for credential scope resolution flags.
// Exactly one of these fields must be set when creating/getting/deleting a credential.
type CredentialScopeFlags struct {
	Ecosystem string
	Domain    string
	App       string
	Workspace string
}

// addCredentialScopeFlags adds the standard credential scope flags to a command.
func addCredentialScopeFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("ecosystem", "e", "", "Ecosystem scope")
	cmd.Flags().StringP("domain", "d", "", "Domain scope")
	cmd.Flags().StringP("app", "a", "", "App scope")
	cmd.Flags().StringP("workspace", "w", "", "Workspace scope")
}

// resolveCredentialScopeFromFlags reads the four scope flags from the command,
// validates that exactly one is set, and resolves the scope name to a
// (CredentialScopeType, scopeID) pair via the DataStore.
func resolveCredentialScopeFromFlags(cmd *cobra.Command, ds db.DataStore) (models.CredentialScopeType, int64, error) {
	eco, _ := cmd.Flags().GetString("ecosystem")
	dom, _ := cmd.Flags().GetString("domain")
	app, _ := cmd.Flags().GetString("app")
	ws, _ := cmd.Flags().GetString("workspace")

	// Count how many scope flags are set
	count := 0
	if eco != "" {
		count++
	}
	if dom != "" {
		count++
	}
	if app != "" {
		count++
	}
	if ws != "" {
		count++
	}

	if count != 1 {
		return "", 0, fmt.Errorf("exactly one scope (--ecosystem, --domain, --app, or --workspace) is required, got %d", count)
	}

	// Resolve the scope name to an ID
	switch {
	case eco != "":
		e, err := ds.GetEcosystemByName(eco)
		if err != nil {
			return "", 0, fmt.Errorf("ecosystem '%s' not found: %w", eco, err)
		}
		return models.CredentialScopeEcosystem, int64(e.ID), nil

	case dom != "":
		// Try active context first, fall back to global search
		dbCtx, err := ds.GetContext()
		if err == nil && dbCtx.ActiveEcosystemID != nil {
			d, err := ds.GetDomainByName(*dbCtx.ActiveEcosystemID, dom)
			if err == nil {
				return models.CredentialScopeDomain, int64(d.ID), nil
			}
		}
		// Fall back to global search
		domains, err := ds.ListAllDomains()
		if err != nil {
			return "", 0, fmt.Errorf("failed to list domains: %w", err)
		}
		for _, d := range domains {
			if d.Name == dom {
				return models.CredentialScopeDomain, int64(d.ID), nil
			}
		}
		return "", 0, fmt.Errorf("domain '%s' not found", dom)

	case app != "":
		a, err := ds.GetAppByNameGlobal(app)
		if err != nil {
			return "", 0, fmt.Errorf("app '%s' not found: %w", app, err)
		}
		return models.CredentialScopeApp, int64(a.ID), nil

	case ws != "":
		// Try active context first, fall back to global search
		dbCtx, err := ds.GetContext()
		if err == nil && dbCtx.ActiveAppID != nil {
			w, err := ds.GetWorkspaceByName(*dbCtx.ActiveAppID, ws)
			if err == nil {
				return models.CredentialScopeWorkspace, int64(w.ID), nil
			}
		}
		// Fall back to global search
		workspaces, err := ds.ListAllWorkspaces()
		if err != nil {
			return "", 0, fmt.Errorf("failed to list workspaces: %w", err)
		}
		for _, w := range workspaces {
			if w.Name == ws {
				return models.CredentialScopeWorkspace, int64(w.ID), nil
			}
		}
		return "", 0, fmt.Errorf("workspace '%s' not found", ws)
	}

	return "", 0, fmt.Errorf("exactly one scope (--ecosystem, --domain, --app, or --workspace) is required, got 0")
}

// createCredentialCmd creates a new credential
var createCredentialCmd = &cobra.Command{
	Use:     "credential <name>",
	Aliases: []string{"cred"},
	Short:   "Create a credential",
	Long: `Create a new credential configuration.

Credentials reference secrets stored in the macOS Keychain or environment variables.
They are scoped to exactly one resource (ecosystem, domain, app, or workspace).

Sources:
  keychain - Reference a macOS Keychain item (requires --service)
  env      - Reference an environment variable (requires --env-var)

Examples:
  dvm create credential github-token --source keychain --service com.github.token --app my-api
  dvm create credential api-key --source env --env-var MY_API_KEY --ecosystem prod
  dvm create cred db-pass --source keychain --service com.db.password --domain backend`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		credName := args[0]

		// Get DataStore from context
		ds, err := getDataStore(cmd)
		if err != nil {
			return fmt.Errorf("dataStore not found in context")
		}

		// Read flags
		source, _ := cmd.Flags().GetString("source")
		service, _ := cmd.Flags().GetString("service")
		envVar, _ := cmd.Flags().GetString("env-var")
		description, _ := cmd.Flags().GetString("description")

		// Validate source
		if source == "" {
			return fmt.Errorf("--source is required (keychain or env)")
		}
		if source != "keychain" && source != "env" {
			return fmt.Errorf("--source must be 'keychain' or 'env', got '%s'", source)
		}

		// Validate conditional flags
		if source == "keychain" && service == "" {
			return fmt.Errorf("--service is required when --source=keychain")
		}
		if source == "env" && envVar == "" {
			return fmt.Errorf("--env-var is required when --source=env")
		}

		// Resolve scope
		scopeType, scopeID, err := resolveCredentialScopeFromFlags(cmd, ds)
		if err != nil {
			return err
		}

		// Build credential
		cred := &models.CredentialDB{
			Name:      credName,
			ScopeType: scopeType,
			ScopeID:   scopeID,
			Source:    source,
		}

		if service != "" {
			cred.Service = &service
		}
		if envVar != "" {
			cred.EnvVar = &envVar
		}
		if description != "" {
			cred.Description = &description
		}

		// Create credential
		if err := ds.CreateCredential(cred); err != nil {
			return fmt.Errorf("failed to create credential: %w", err)
		}

		render.Success(fmt.Sprintf("Credential '%s' created (scope: %s)", credName, scopeType))
		return nil
	},
}

func init() {
	createCmd.AddCommand(createCredentialCmd)

	// Source flags
	createCredentialCmd.Flags().String("source", "", "Credential source: keychain or env (required)")
	createCredentialCmd.Flags().String("service", "", "Keychain service name (required when --source=keychain)")
	createCredentialCmd.Flags().String("env-var", "", "Environment variable name (required when --source=env)")
	createCredentialCmd.Flags().String("description", "", "Credential description")

	// Scope flags
	addCredentialScopeFlags(createCredentialCmd)
}
