package cmd

import (
	"fmt"

	"devopsmaestro/pkg/nvimbridge"
	"devopsmaestro/pkg/themebridge"
	"github.com/rmkohlman/MaestroNvim/nvimops"
	theme "github.com/rmkohlman/MaestroTheme"

	"github.com/spf13/cobra"
)

// getNvimManager creates a nvimops.Manager using the DataStore from the command context.
// This uses the PluginDBStoreAdapter to bridge between the PluginStore interface and the
// DataStore interface, providing a unified storage location for both nvp and dvm.
//
// The returned manager should be closed when done:
//
//	mgr, err := getNvimManager(cmd)
//	if err != nil {
//	    return err
//	}
//	defer mgr.Close()
func getNvimManager(cmd *cobra.Command) (nvimops.Manager, error) {
	datastore, err := getDataStore(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get datastore: %w", err)
	}

	// Create PluginDBStoreAdapter that implements store.PluginStore using the DataStore
	// Note: We don't own the connection (datastore lifecycle is managed by root.go)
	adapter := nvimbridge.NewPluginDBStoreAdapter(datastore)

	// Create manager with the adapter
	mgr, err := nvimops.NewWithOptions(nvimops.Options{
		Store: adapter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create nvim manager: %w", err)
	}

	return mgr, nil
}

// getThemeStore creates a theme.Store using the DataStore from the command context.
// This uses the DBStoreAdapter to bridge between the theme.Store interface and the
// DataStore interface, providing a unified storage location for both nvp and dvm.
func getThemeStore(cmd *cobra.Command) (theme.Store, error) {
	datastore, err := getDataStore(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get datastore: %w", err)
	}

	// Create DBStoreAdapter that implements theme.Store using the DataStore
	// Note: We don't own the connection (datastore lifecycle is managed by root.go)
	adapter := themebridge.NewDBStoreAdapter(datastore)

	return adapter, nil
}
