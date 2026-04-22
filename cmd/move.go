package cmd

import (
	"fmt"

	"devopsmaestro/pkg/resource/handlers"
	"github.com/rmkohlman/MaestroSDK/render"

	"github.com/spf13/cobra"
)

// Flags for `dvm move` subcommands.
var (
	moveToDomain     string
	moveToSystem     string
	moveEcosystem    string
	moveSystemDryRun bool
	moveAppDryRun    bool
)

// moveCmd is the parent for `dvm move <kind> <name>`.
//
// Subcommands:
//   - dvm move system <name> --to-domain <d> [-e <eco>]
//   - dvm move app    <name> --to-system <s> [--to-domain <d>] [-e <eco>]
//
// See `dvm app detach` for fully detaching an App from its System.
var moveCmd = &cobra.Command{
	Use:   "move",
	Short: "Move a resource to a new parent in the hierarchy",
	Long: `Move a resource to a new parent in the App → System → Domain → Ecosystem hierarchy.

Move operations are atomic: parent FKs are rewritten in a single transaction
and child resources cascade automatically (e.g. moving a System rewrites all
of its child Apps so the hierarchy stays consistent).

Examples:
  # Move a System to a different Domain (cascades all child Apps)
  dvm move system payments --to-domain backend

  # Move an App to a different System
  dvm move app checkout --to-system orders

  # Disambiguate when the name is not unique across ecosystems
  dvm move system payments --to-domain backend -e prod`,
}

var moveSystemCmd = &cobra.Command{
	Use:   "system <name>",
	Short: "Move a System to a new Domain",
	Long: `Move a System to a new Domain. All child Apps are reparented atomically.

Examples:
  dvm move system payments --to-domain backend
  dvm move system payments --to-domain backend -e prod`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if moveToDomain == "" {
			return fmt.Errorf("--to-domain is required for `move system`")
		}

		ctx, err := buildResourceContext(cmd)
		if err != nil {
			return err
		}

		h := handlers.NewSystemHandler()

		if moveSystemDryRun {
			render.Plain(fmt.Sprintf("Would move system/%s to domain/%s (ecosystem hint: %q)",
				name, moveToDomain, moveEcosystem))
			return nil
		}

		result, err := h.Move(ctx, name, handlers.MoveTarget{
			EcosystemName: moveEcosystem,
			DomainName:    moveToDomain,
		})
		if err != nil {
			return err
		}
		return renderMoveResult(cmd, result)
	},
}

var moveAppCmd = &cobra.Command{
	Use:   "app <name>",
	Short: "Move an App to a new System (or Domain)",
	Long: `Move an App to a new parent System (default) or directly to a Domain
(when --to-system is omitted and --to-domain is provided).

Examples:
  dvm move app checkout --to-system orders
  dvm move app checkout --to-system orders --to-domain backend
  dvm move app checkout --to-domain backend          # no system, attached to domain
  dvm move app checkout --to-system orders -e prod`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if moveToSystem == "" && moveToDomain == "" {
			return fmt.Errorf("at least one of --to-system or --to-domain is required for `move app`")
		}

		ctx, err := buildResourceContext(cmd)
		if err != nil {
			return err
		}

		h := handlers.NewAppHandler()

		if moveAppDryRun {
			render.Plain(fmt.Sprintf("Would move app/%s to system=%q domain=%q (ecosystem hint: %q)",
				name, moveToSystem, moveToDomain, moveEcosystem))
			return nil
		}

		result, err := h.Move(ctx, name, handlers.MoveTarget{
			EcosystemName: moveEcosystem,
			DomainName:    moveToDomain,
			SystemName:    moveToSystem,
		})
		if err != nil {
			return err
		}
		return renderMoveResult(cmd, result)
	},
}

// renderMoveResult prints a kubectl-style success line and honors -o yaml/json
// when set.
func renderMoveResult(cmd *cobra.Command, result *handlers.MoveResult) error {
	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "yaml", "json":
		// Structured output: emit a small summary object via render.OutputWith.
		payload := map[string]any{
			"kind":         result.Kind,
			"name":         result.Name,
			"fromParent":   result.FromParent,
			"toParent":     result.ToParent,
			"cascadedApps": result.CascadedApps,
			"noOp":         result.NoOp,
		}
		return render.OutputWith(output, payload, render.Options{})
	default:
		render.Success(result.String())
		return nil
	}
}

func init() {
	rootCmd.AddCommand(moveCmd)
	moveCmd.AddCommand(moveSystemCmd)
	moveCmd.AddCommand(moveAppCmd)

	// move system flags
	moveSystemCmd.Flags().StringVar(&moveToDomain, "to-domain", "", "Target Domain name (required)")
	moveSystemCmd.Flags().StringVarP(&moveEcosystem, "ecosystem", "e", "", "Ecosystem hint when names are ambiguous")
	AddDryRunFlag(moveSystemCmd, &moveSystemDryRun)
	AddOutputFlag(moveSystemCmd, "")

	// move app flags
	moveAppCmd.Flags().StringVar(&moveToSystem, "to-system", "", "Target System name")
	moveAppCmd.Flags().StringVar(&moveToDomain, "to-domain", "", "Target Domain name (required if --to-system omitted)")
	moveAppCmd.Flags().StringVarP(&moveEcosystem, "ecosystem", "e", "", "Ecosystem hint when names are ambiguous")
	AddDryRunFlag(moveAppCmd, &moveAppDryRun)
	AddOutputFlag(moveAppCmd, "")
}
