package main

import (
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/spf13/cobra"
)

// =============================================================================
// ENABLE/DISABLE COMMANDS
// =============================================================================

var enableCmd = &cobra.Command{
	Use:   "enable <name>...",
	Short: "Enable plugins for Lua generation",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setPluginsEnabled(args, true)
	},
}

var disableCmd = &cobra.Command{
	Use:   "disable <name>...",
	Short: "Disable plugins (exclude from Lua generation)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setPluginsEnabled(args, false)
	},
}

func setPluginsEnabled(names []string, enabled bool) error {
	mgr, err := getManager()
	if err != nil {
		return err
	}
	defer mgr.Close()

	action := "enabled"
	if !enabled {
		action = "disabled"
	}

	for _, name := range names {
		p, err := mgr.Get(name)
		if err != nil {
			render.WarningfToStderr("plugin not found: %s", name)
			continue
		}

		p.Enabled = enabled
		if err := mgr.Apply(p); err != nil {
			render.WarningfToStderr("failed to update %s: %v", name, err)
			continue
		}

		render.Successf("Plugin '%s' %s", name, action)
	}

	return nil
}
