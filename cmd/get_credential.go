package cmd

import (
	"fmt"
	"sort"
	"strings"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/render"

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

			rows := make([][]string, 0, len(creds))
			for _, c := range creds {
				scope := resolveScopeName(ds, c.ScopeType, c.ScopeID)
				target := formatTargetVars(c)
				desc := ""
				if c.Description != nil {
					desc = *c.Description
				}
				rows = append(rows, []string{c.Name, scope, c.Source, target, desc, formatExpirationStatus(c)})
			}
			render.OutputWith(getOutputFormat, render.TableData{
				Headers: []string{"NAME", "SCOPE", "SOURCE", "TARGET", "DESCRIPTION", "EXPIRES"},
				Rows:    rows,
			}, render.Options{Type: render.TypeTable})
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

		rows := make([][]string, 0, len(creds))
		for _, c := range creds {
			scope := resolveScopeName(ds, c.ScopeType, c.ScopeID)
			target := formatTargetVars(c)
			desc := ""
			if c.Description != nil {
				desc = *c.Description
			}
			rows = append(rows, []string{c.Name, scope, c.Source, target, desc, formatExpirationStatus(c)})
		}
		render.OutputWith(getOutputFormat, render.TableData{
			Headers: []string{"NAME", "SCOPE", "SOURCE", "TARGET", "DESCRIPTION", "EXPIRES"},
			Rows:    rows,
		}, render.Options{Type: render.TypeTable})
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

		// For JSON/YAML, output the model data via ToYAML (issue #183)
		if getOutputFormat == "json" || getOutputFormat == "yaml" {
			scopeName := resolveCredentialScopeTargetName(ds, cred.ScopeType, cred.ScopeID)
			yamlDoc := cred.ToYAML(scopeName)
			return render.OutputWith(getOutputFormat, yamlDoc, render.Options{})
		}

		// Display credential details (plain text)
		render.Plainf("Name:      %s", cred.Name)
		render.Plainf("Scope:     %s", resolveScopeName(ds, cred.ScopeType, cred.ScopeID))
		render.Plainf("Source:    %s", cred.Source)
		if cred.VaultSecret != nil {
			render.Plainf("Secret:    %s", *cred.VaultSecret)
		}
		if cred.VaultEnv != nil {
			render.Plainf("Vault Env: %s", *cred.VaultEnv)
		}
		if cred.VaultUsernameSecret != nil {
			render.Plainf("Username Secret: %s", *cred.VaultUsernameSecret)
		}
		if cred.EnvVar != nil {
			render.Plainf("EnvVar:    %s", *cred.EnvVar)
		}
		if cred.Description != nil {
			render.Plainf("Desc:      %s", *cred.Description)
		}
		if cred.UsernameVar != nil {
			render.Plainf("Username:  %s", *cred.UsernameVar)
		}
		if cred.PasswordVar != nil {
			render.Plainf("Password:  %s", *cred.PasswordVar)
		}
		// Show expiration status
		if cred.ExpiresAt != nil {
			render.Plainf("Expires:   %s", cred.ExpiresAt.Format("2006-01-02 15:04:05"))
			status := cred.ExpirationStatus()
			switch status {
			case "expired":
				render.Warning(fmt.Sprintf("Status:    %s", status))
			case "expiring soon":
				render.Warning(fmt.Sprintf("Status:    %s", status))
			default:
				render.Plainf("Status:    %s", status)
			}
		}
		if cred.HasVaultFields() {
			fields, err := cred.GetVaultFieldsMap()
			if err == nil && len(fields) > 0 {
				render.Plainf("Fields:")
				// Sort keys for deterministic output
				keys := make([]string, 0, len(fields))
				for k := range fields {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, envVar := range keys {
					fieldName := fields[envVar]
					if envVar == fieldName {
						render.Plainf("  %s", envVar)
					} else {
						render.Plainf("  %s <- %s", envVar, fieldName)
					}
				}
			}
		}

		return nil
	},
}

func init() {
	getCmd.AddCommand(getCredentialsCmd)
	getCmd.AddCommand(getCredentialCmd)

	// --all flag for list command
	AddAllFlag(getCredentialsCmd, "List all credentials across all scopes")

	// Scope flags for both commands
	addCredentialScopeFlags(getCredentialsCmd)
	addCredentialScopeFlags(getCredentialCmd)
}

// resolveScopeName resolves a credential scope to a human-readable "type: name" string.
func resolveScopeName(ds db.DataStore, scopeType models.CredentialScopeType, scopeID int64) string {
	var name string
	switch scopeType {
	case models.CredentialScopeEcosystem:
		if e, err := ds.GetEcosystemByID(int(scopeID)); err == nil {
			name = e.Name
		}
	case models.CredentialScopeDomain:
		if d, err := ds.GetDomainByID(int(scopeID)); err == nil {
			name = d.Name
		}
	case models.CredentialScopeApp:
		if a, err := ds.GetAppByID(int(scopeID)); err == nil {
			name = a.Name
		}
	case models.CredentialScopeWorkspace:
		if w, err := ds.GetWorkspaceByID(int(scopeID)); err == nil {
			name = w.Name
		}
	}
	if name != "" {
		return fmt.Sprintf("%s: %s", scopeType, name)
	}
	return fmt.Sprintf("%s (ID: %d)", scopeType, scopeID)
}

// resolveCredentialScopeTargetName resolves a credential scope to just the
// target entity name (e.g. "prod" for an ecosystem). This is the raw name
// required by CredentialDB.ToYAML, as opposed to resolveScopeName which
// returns a "type: name" string for human display.
func resolveCredentialScopeTargetName(ds db.DataStore, scopeType models.CredentialScopeType, scopeID int64) string {
	switch scopeType {
	case models.CredentialScopeEcosystem:
		if e, err := ds.GetEcosystemByID(int(scopeID)); err == nil {
			return e.Name
		}
	case models.CredentialScopeDomain:
		if d, err := ds.GetDomainByID(int(scopeID)); err == nil {
			return d.Name
		}
	case models.CredentialScopeApp:
		if a, err := ds.GetAppByID(int(scopeID)); err == nil {
			return a.Name
		}
	case models.CredentialScopeWorkspace:
		if w, err := ds.GetWorkspaceByID(int(scopeID)); err == nil {
			return w.Name
		}
	}
	return ""
}

// formatTargetVars returns a comma-separated string of the env var names
// that this credential will inject at build/attach time.
func formatTargetVars(c *models.CredentialDB) string {
	// Priority: vault fields > dual-field > env var > credential name
	if c.HasVaultFields() {
		fields, err := c.GetVaultFieldsMap()
		if err == nil && len(fields) > 0 {
			keys := make([]string, 0, len(fields))
			for envVar := range fields {
				keys = append(keys, envVar)
			}
			sort.Strings(keys)
			return strings.Join(keys, ", ")
		}
	}
	var parts []string
	if c.UsernameVar != nil {
		parts = append(parts, *c.UsernameVar)
	}
	if c.PasswordVar != nil {
		parts = append(parts, *c.PasswordVar)
	}
	if len(parts) > 0 {
		return strings.Join(parts, ", ")
	}
	if c.EnvVar != nil {
		return *c.EnvVar
	}
	// Fallback: credential name is used as the env var name
	return c.Name
}

// formatExpirationStatus returns a human-readable expiration string for table display.
// Returns "-" for credentials with no expiration set.
func formatExpirationStatus(c *models.CredentialDB) string {
	status := c.ExpirationStatus()
	if status == "" {
		return "-"
	}
	return status
}
