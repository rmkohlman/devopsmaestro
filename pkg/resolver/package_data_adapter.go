// Package resolver provides an adapter that wraps db.DataStore to implement
// PackageDataAccessor. This allows the PackageHierarchyResolver to use the
// real database without depending on the db package directly.
package resolver

import (
	"devopsmaestro/db"
)

// DataStorePackageAdapter adapts db.DataStore to the PackageDataAccessor interface.
type DataStorePackageAdapter struct {
	ds db.DataStore
}

// NewDataStorePackageAdapter creates a new adapter wrapping the given DataStore.
func NewDataStorePackageAdapter(ds db.DataStore) *DataStorePackageAdapter {
	return &DataStorePackageAdapter{ds: ds}
}

// GetPackageEcosystemByID retrieves ecosystem package data by ID.
func (a *DataStorePackageAdapter) GetPackageEcosystemByID(id int) (PackageEcosystemData, error) {
	eco, err := a.ds.GetEcosystemByID(id)
	if err != nil {
		return PackageEcosystemData{}, err
	}
	return PackageEcosystemData{
		ID:              eco.ID,
		Name:            eco.Name,
		NvimPackage:     eco.NvimPackage,
		TerminalPackage: eco.TerminalPackage,
	}, nil
}

// GetPackageDomainByID retrieves domain package data by ID.
func (a *DataStorePackageAdapter) GetPackageDomainByID(id int) (PackageDomainData, error) {
	dom, err := a.ds.GetDomainByID(id)
	if err != nil {
		return PackageDomainData{}, err
	}
	return PackageDomainData{
		ID:              dom.ID,
		EcosystemID:     dom.EcosystemID,
		Name:            dom.Name,
		NvimPackage:     dom.NvimPackage,
		TerminalPackage: dom.TerminalPackage,
	}, nil
}

// GetPackageAppByID retrieves app package data by ID.
func (a *DataStorePackageAdapter) GetPackageAppByID(id int) (PackageAppData, error) {
	app, err := a.ds.GetAppByID(id)
	if err != nil {
		return PackageAppData{}, err
	}
	return PackageAppData{
		ID:              app.ID,
		DomainID:        app.DomainID,
		Name:            app.Name,
		NvimPackage:     app.NvimPackage,
		TerminalPackage: app.TerminalPackage,
	}, nil
}

// GetPackageWorkspaceByID retrieves workspace package data by ID.
func (a *DataStorePackageAdapter) GetPackageWorkspaceByID(id int) (PackageWorkspaceData, error) {
	ws, err := a.ds.GetWorkspaceByID(id)
	if err != nil {
		return PackageWorkspaceData{}, err
	}
	return PackageWorkspaceData{
		ID:              ws.ID,
		AppID:           ws.AppID,
		Name:            ws.Name,
		NvimPackage:     ws.NvimPackage,
		TerminalPackage: ws.TerminalPackage,
	}, nil
}

// GetDefault retrieves a default value by key.
func (a *DataStorePackageAdapter) GetDefault(key string) (string, error) {
	return a.ds.GetDefault(key)
}

// Ensure DataStorePackageAdapter implements PackageDataAccessor.
var _ PackageDataAccessor = (*DataStorePackageAdapter)(nil)
