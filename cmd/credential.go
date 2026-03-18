package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/envvalidation"
	"github.com/rmkohlman/MaestroSDK/render"

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

Credentials reference secrets stored in MaestroVault or environment variables.
They are scoped to exactly one resource (ecosystem, domain, app, or workspace).

Sources:
  vault - Reference a MaestroVault secret (requires --vault-secret)
  env   - Reference an environment variable (requires --env-var)

Vault Fields:
  Use --vault-field to map individual fields from a multi-field vault secret
  to environment variables. Each --vault-field maps one field.

  Format: --vault-field ENV_VAR=field_name   (explicit mapping)
          --vault-field FIELD_NAME           (auto-map: env var = field name)

Examples:
  # GitHub PAT stored in MaestroVault
  dvm create credential github-creds \
    --source vault \
    --vault-secret "github-pat" \
    --vault-env production \
    --username-var GITHUB_USERNAME \
    --password-var GITHUB_PAT \
    --ecosystem myorg

  # Multi-field vault secret with explicit field mapping
  dvm create credential db-creds \
    --source vault \
    --vault-secret "database/prod" \
    --vault-field DB_HOST=host \
    --vault-field DB_PORT=port \
    --vault-field DB_PASSWORD=password \
    --ecosystem prod

  # API key from environment variable
  dvm create credential api-key \
    --source env \
    --env-var MY_API_KEY \
    --ecosystem prod`,
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
		vaultSecret, _ := cmd.Flags().GetString("vault-secret")
		vaultEnv, _ := cmd.Flags().GetString("vault-env")
		vaultUsernameSecret, _ := cmd.Flags().GetString("vault-username-secret")
		envVar, _ := cmd.Flags().GetString("env-var")
		description, _ := cmd.Flags().GetString("description")
		usernameVar, _ := cmd.Flags().GetString("username-var")
		passwordVar, _ := cmd.Flags().GetString("password-var")

		// Validate source
		if source == "" {
			return fmt.Errorf("--source is required (vault or env)")
		}
		if source != "vault" && source != "env" {
			return fmt.Errorf("--source must be 'vault' or 'env', got '%s'", source)
		}

		// Validate conditional flags
		if source == "vault" && vaultSecret == "" {
			return fmt.Errorf("--vault-secret is required when --source=vault")
		}
		if source == "env" && envVar == "" {
			return fmt.Errorf("--env-var is required when --source=env")
		}

		// Validate vault-only flags
		if source != "vault" && (vaultSecret != "" || vaultEnv != "" || vaultUsernameSecret != "") {
			return fmt.Errorf("--vault-secret, --vault-env, and --vault-username-secret are only valid with --source=vault")
		}

		// Validate dual-field flags are only used with vault source
		if source != "vault" && (usernameVar != "" || passwordVar != "") {
			return fmt.Errorf("--username-var and --password-var are only valid with --source=vault")
		}

		// Validate env var names if provided
		if usernameVar != "" {
			if err := envvalidation.ValidateEnvKey(usernameVar); err != nil {
				return fmt.Errorf("invalid --username-var: %w", err)
			}
		}
		if passwordVar != "" {
			if err := envvalidation.ValidateEnvKey(passwordVar); err != nil {
				return fmt.Errorf("invalid --password-var: %w", err)
			}
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

		if vaultSecret != "" {
			cred.VaultSecret = &vaultSecret
		}
		if vaultEnv != "" {
			cred.VaultEnv = &vaultEnv
		}
		if vaultUsernameSecret != "" {
			cred.VaultUsernameSecret = &vaultUsernameSecret
		}
		if envVar != "" {
			cred.EnvVar = &envVar
		}
		if description != "" {
			cred.Description = &description
		}
		if usernameVar != "" {
			cred.UsernameVar = &usernameVar
		}
		if passwordVar != "" {
			cred.PasswordVar = &passwordVar
		}

		// Parse --vault-field flags
		vaultFieldFlags, _ := cmd.Flags().GetStringArray("vault-field")
		if len(vaultFieldFlags) > 0 {
			// Requires vault source
			if source != "vault" {
				return fmt.Errorf("--vault-field requires --source=vault")
			}
			// Requires vault secret
			if vaultSecret == "" {
				return fmt.Errorf("--vault-field requires --vault-secret")
			}
			// Mutual exclusivity
			if usernameVar != "" || passwordVar != "" {
				return fmt.Errorf("--vault-field cannot be used with --username-var or --password-var")
			}
			if vaultUsernameSecret != "" {
				return fmt.Errorf("--vault-field cannot be used with --vault-username-secret")
			}
			// Max 50 fields
			if len(vaultFieldFlags) > 50 {
				return fmt.Errorf("too many vault fields (%d): maximum is 50", len(vaultFieldFlags))
			}
			// Parse fields
			vaultFields := make(map[string]string)
			for _, vf := range vaultFieldFlags {
				if strings.Contains(vf, "=") {
					parts := strings.SplitN(vf, "=", 2)
					envVarName := parts[0]
					fieldName := parts[1]
					if err := envvalidation.ValidateEnvKey(envVarName); err != nil {
						return fmt.Errorf("invalid env var in --vault-field %q: %w", vf, err)
					}
					if fieldName == "" {
						return fmt.Errorf("field name cannot be empty in --vault-field %q", vf)
					}
					vaultFields[envVarName] = fieldName
				} else {
					if err := envvalidation.ValidateEnvKey(vf); err != nil {
						return fmt.Errorf("invalid --vault-field name %q: %w", vf, err)
					}
					vaultFields[vf] = vf
				}
			}
			vaultFieldsJSON, err := json.Marshal(vaultFields)
			if err != nil {
				return fmt.Errorf("failed to serialize vault fields: %w", err)
			}
			vfStr := string(vaultFieldsJSON)
			cred.VaultFields = &vfStr
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
	createCredentialCmd.Flags().String("source", "", "Credential source: vault or env (required)")
	createCredentialCmd.Flags().String("vault-secret", "", "MaestroVault secret name (required when --source=vault)")
	createCredentialCmd.Flags().String("vault-env", "", "MaestroVault environment (optional, e.g. production)")
	createCredentialCmd.Flags().String("vault-username-secret", "", "MaestroVault secret name for username (vault source only)")
	createCredentialCmd.Flags().String("env-var", "", "Environment variable name (required when --source=env)")
	createCredentialCmd.Flags().String("description", "", "Credential description")
	createCredentialCmd.Flags().String("username-var", "", "Environment variable name for the username (vault source only)")
	createCredentialCmd.Flags().String("password-var", "", "Environment variable name for the password (vault source only)")
	createCredentialCmd.Flags().StringArray("vault-field", nil,
		"Map vault field to env var (repeatable): ENV_VAR=field_name or FIELD_NAME for auto-map")

	// Scope flags
	addCredentialScopeFlags(createCredentialCmd)
}
