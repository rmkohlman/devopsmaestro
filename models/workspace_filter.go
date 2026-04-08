package models

// WorkspaceFilter defines criteria for finding workspaces.
// All fields are optional - only non-empty values are used for filtering.
type WorkspaceFilter struct {
	// EcosystemName filters by ecosystem name.
	EcosystemName string

	// DomainName filters by domain name.
	DomainName string

	// AppName filters by app name.
	AppName string

	// WorkspaceName filters by workspace name.
	WorkspaceName string
}

// IsEmpty returns true if no filter criteria are set.
func (f WorkspaceFilter) IsEmpty() bool {
	return f.EcosystemName == "" &&
		f.DomainName == "" &&
		f.AppName == "" &&
		f.WorkspaceName == ""
}

// AppWithHierarchy contains an app along with its full hierarchy information.
// This is used for ambiguity detection when resolving apps by name.
type AppWithHierarchy struct {
	// App is the resolved app.
	App *App

	// Domain is the parent domain.
	Domain *Domain

	// Ecosystem is the parent ecosystem.
	Ecosystem *Ecosystem
}

// DomainWithHierarchy contains a domain along with its parent ecosystem.
// This is used for ambiguity detection when resolving domains by name.
type DomainWithHierarchy struct {
	// Domain is the resolved domain.
	Domain *Domain

	// Ecosystem is the parent ecosystem.
	Ecosystem *Ecosystem
}

// WorkspaceWithHierarchy contains a workspace along with its full hierarchy information.
// This is used for workspace resolution and display purposes.
type WorkspaceWithHierarchy struct {
	// Workspace is the resolved workspace.
	Workspace *Workspace

	// App is the parent app.
	App *App

	// Domain is the parent domain.
	Domain *Domain

	// Ecosystem is the parent ecosystem.
	Ecosystem *Ecosystem
}

// FullPath returns the full hierarchical path of the workspace.
// Format: ecosystem/domain/app/workspace
func (w *WorkspaceWithHierarchy) FullPath() string {
	return w.Ecosystem.Name + "/" + w.Domain.Name + "/" + w.App.Name + "/" + w.Workspace.Name
}

// ShortPath returns a shorter path when some hierarchy is implied.
// Format: app/workspace
func (w *WorkspaceWithHierarchy) ShortPath() string {
	return w.App.Name + "/" + w.Workspace.Name
}
