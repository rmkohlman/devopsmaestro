package cmd

import (
	"devopsmaestro/models"

	"github.com/spf13/cobra"
)

// HierarchyFlags holds the values for workspace hierarchy resolution flags.
// These flags allow users to specify partial criteria for workspace resolution
// without needing multiple `dvm use` commands.
type HierarchyFlags struct {
	Ecosystem string
	Domain    string
	App       string
	Workspace string
}

// AddHierarchyFlags adds the standard hierarchy resolution flags to a command.
// Flags:
//   - `-e, --ecosystem` - Filter by ecosystem name
//   - `-d, --domain`    - Filter by domain name
//   - `-a, --app`       - Filter by app name
//   - `-w, --workspace` - Filter by workspace name
func AddHierarchyFlags(cmd *cobra.Command, flags *HierarchyFlags) {
	cmd.Flags().StringVarP(&flags.Ecosystem, "ecosystem", "e", "", "Filter by ecosystem name")
	cmd.Flags().StringVarP(&flags.Domain, "domain", "d", "", "Filter by domain name")
	cmd.Flags().StringVarP(&flags.App, "app", "a", "", "Filter by app name")
	cmd.Flags().StringVarP(&flags.Workspace, "workspace", "w", "", "Filter by workspace name")
}

// ToFilter converts HierarchyFlags to a WorkspaceFilter for use with the resolver.
func (f *HierarchyFlags) ToFilter() models.WorkspaceFilter {
	return models.WorkspaceFilter{
		EcosystemName: f.Ecosystem,
		DomainName:    f.Domain,
		AppName:       f.App,
		WorkspaceName: f.Workspace,
	}
}

// HasAnyFlag returns true if any hierarchy flag was provided.
func (f *HierarchyFlags) HasAnyFlag() bool {
	return f.Ecosystem != "" || f.Domain != "" || f.App != "" || f.Workspace != ""
}

// IsEmpty returns true if no flags were provided.
func (f *HierarchyFlags) IsEmpty() bool {
	return !f.HasAnyFlag()
}
