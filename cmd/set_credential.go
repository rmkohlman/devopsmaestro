package cmd

import (
	"fmt"
	"time"

	"devopsmaestro/utils"
	"github.com/rmkohlman/MaestroSDK/render"

	"github.com/spf13/cobra"
)

// setCredentialCmd updates properties on an existing credential
var setCredentialCmd = &cobra.Command{
	Use:   "credential <name>",
	Short: "Set credential properties",
	Long: `Update properties on an existing credential.

Currently supports setting expiration for credential rotation reminders.
The --expires flag accepts duration strings: Go durations (24h, 720h) or
day-based durations (90d, 365d).

Examples:
  dvm set credential github-token --expires 90d --app my-api
  dvm set credential api-key --expires 8760h --ecosystem prod
  dvm set credential db-pass --expires 365d --domain backend
  dvm set credential deploy-key --expires 0 --app my-api   # clear expiration`,
	Args: cobra.ExactArgs(1),
	RunE: runSetCredential,
}

func init() {
	setCmd.AddCommand(setCredentialCmd)

	setCredentialCmd.Flags().String("expires", "", "Set expiration duration (e.g., 90d, 24h, 8760h; 0 to clear)")
	addCredentialScopeFlags(setCredentialCmd)
}

func runSetCredential(cmd *cobra.Command, args []string) error {
	credName := args[0]

	ds, err := getDataStore(cmd)
	if err != nil {
		return fmt.Errorf("dataStore not found in context")
	}

	// Resolve scope
	scopeType, scopeID, err := resolveCredentialScopeFromFlags(cmd, ds)
	if err != nil {
		return err
	}

	// Get existing credential
	cred, err := ds.GetCredential(scopeType, scopeID, credName)
	if err != nil {
		return fmt.Errorf("credential '%s' not found in %s scope: %w", credName, scopeType, err)
	}

	// Process --expires flag
	expiresStr, _ := cmd.Flags().GetString("expires")
	if expiresStr == "" {
		return fmt.Errorf("at least one property flag is required (e.g., --expires)")
	}

	if expiresStr == "0" {
		// Clear expiration
		cred.ExpiresAt = nil
		render.Info("Clearing expiration")
	} else {
		dur, err := utils.ParseDuration(expiresStr)
		if err != nil {
			return fmt.Errorf("invalid --expires value: %w", err)
		}
		expiresAt := time.Now().Add(dur)
		cred.ExpiresAt = &expiresAt
		render.Info(fmt.Sprintf("Setting expiration to %s", expiresAt.Format("2006-01-02 15:04:05")))
	}

	// Update credential
	if err := ds.UpdateCredential(cred); err != nil {
		return fmt.Errorf("failed to update credential: %w", err)
	}

	render.Success(fmt.Sprintf("Credential '%s' updated", credName))
	return nil
}
