package cmd

import (
	"fmt"

	"devopsmaestro/pkg/resource/handlers"
	"github.com/rmkohlman/MaestroSDK/render"

	"github.com/spf13/cobra"
)

// `dvm app detach <name> --from-system [-e <eco>]` (issue #397).
//
// We expose the detach as a subcommand under a new top-level `app` parent
// (rather than reusing the existing top-level `detach` command, which is
// already taken for container detach).
//
// Detach semantics: sets BOTH SystemID=NULL and DomainID=NULL on the App,
// fully detaching it from System+Domain so it lives at Ecosystem level.

var (
	appDetachFromSystem bool
	appDetachEcosystem  string
	appDetachDryRun     bool
)

// appCmd is a parent for noun-scoped app subcommands like `dvm app detach`.
// Existing `dvm create app`, `dvm get app`, `dvm delete app`, `dvm move app`
// remain on their own kubectl-style verbs — this parent is only for verbs
// that don't have a clean top-level home (collisions, semantics).
var appCmd = &cobra.Command{
	Use:   "app",
	Short: "App-scoped operations (detach, etc.)",
	Long: `App-scoped operations that don't have a top-level kubectl-style verb.

Most app operations live under the standard verbs:
  dvm create app   dvm get app   dvm delete app   dvm move app

This parent hosts verbs that would collide with other commands (e.g. detach,
which is reserved at top level for container detach).`,
}

var appDetachCmd = &cobra.Command{
	Use:   "detach <name>",
	Short: "Detach an App from its System (and Domain), leaving it at ecosystem level",
	Long: `Detach an App from its parent System. Both SystemID and DomainID are
cleared so the App ends up at the Ecosystem level (issue #397, use case 3).

Examples:
  # Detach an App that is uniquely named in the database
  dvm app detach checkout --from-system

  # Disambiguate when the App name exists in multiple ecosystems
  dvm app detach checkout --from-system -e prod

  # Preview without writing
  dvm app detach checkout --from-system --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !appDetachFromSystem {
			return fmt.Errorf("--from-system is required (the only supported detach target today)")
		}
		name := args[0]

		ctx, err := buildResourceContext(cmd)
		if err != nil {
			return err
		}
		h := handlers.NewAppHandler()

		if appDetachDryRun {
			render.Plain(fmt.Sprintf("Would detach app/%s from its system (and domain) — ecosystem hint: %q",
				name, appDetachEcosystem))
			return nil
		}

		force, _ := cmd.Flags().GetBool("force")
		confirmed, err := confirmDelete(
			fmt.Sprintf("Detach app '%s' from its System (sets SystemID=NULL and DomainID=NULL)?", name),
			force,
		)
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}

		result, err := h.Detach(ctx, name, appDetachEcosystem)
		if err != nil {
			return err
		}
		return renderMoveResult(cmd, result)
	},
}

func init() {
	rootCmd.AddCommand(appCmd)
	appCmd.AddCommand(appDetachCmd)

	appDetachCmd.Flags().BoolVar(&appDetachFromSystem, "from-system", false, "Detach the App from its parent System (required)")
	appDetachCmd.Flags().StringVarP(&appDetachEcosystem, "ecosystem", "e", "", "Ecosystem hint when the App name is ambiguous")
	AddDryRunFlag(appDetachCmd, &appDetachDryRun)
	AddForceConfirmFlag(appDetachCmd)
	AddOutputFlag(appDetachCmd, "")
}
