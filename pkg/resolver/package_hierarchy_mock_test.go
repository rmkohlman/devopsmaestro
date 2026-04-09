// Package resolver — test helpers for package hierarchy tests.
// Provides PackageMockDataStore implementing PackageDataAccessor.
// File is .pending — CI skips it until the implementation exists.
package resolver

import (
	"database/sql"
	"errors"
)

// ---------------------------------------------------------------------------
// PackageMockDataStore — implements PackageDataAccessor for testing
// ---------------------------------------------------------------------------

// PackageMockDataStore is the test mock for package hierarchy resolution.
type PackageMockDataStore struct {
	ecosystems map[int]*packageMockEcosystem
	domains    map[int]*packageMockDomain
	apps       map[int]*packageMockApp
	workspaces map[int]*packageMockWorkspace
	defaults   map[string]string
}

type packageMockEcosystem struct {
	ID              int
	Name            string
	NvimPackage     sql.NullString
	TerminalPackage sql.NullString
}

type packageMockDomain struct {
	ID              int
	EcosystemID     int
	Name            string
	NvimPackage     sql.NullString
	TerminalPackage sql.NullString
}

type packageMockApp struct {
	ID              int
	DomainID        int
	Name            string
	NvimPackage     sql.NullString
	TerminalPackage sql.NullString
}

type packageMockWorkspace struct {
	ID              int
	AppID           int
	Name            string
	NvimPackage     sql.NullString
	TerminalPackage sql.NullString
}

func NewPackageMockDataStore() *PackageMockDataStore {
	return &PackageMockDataStore{
		ecosystems: make(map[int]*packageMockEcosystem),
		domains:    make(map[int]*packageMockDomain),
		apps:       make(map[int]*packageMockApp),
		workspaces: make(map[int]*packageMockWorkspace),
		defaults:   make(map[string]string),
	}
}

func newPackageMockEcosystem(id int, name string, nvimPkg, termPkg *string) *packageMockEcosystem {
	e := &packageMockEcosystem{ID: id, Name: name}
	if nvimPkg != nil {
		e.NvimPackage = sql.NullString{String: *nvimPkg, Valid: true}
	}
	if termPkg != nil {
		e.TerminalPackage = sql.NullString{String: *termPkg, Valid: true}
	}
	return e
}

func newPackageMockDomain(id, ecoID int, name string, nvimPkg, termPkg *string) *packageMockDomain {
	d := &packageMockDomain{ID: id, EcosystemID: ecoID, Name: name}
	if nvimPkg != nil {
		d.NvimPackage = sql.NullString{String: *nvimPkg, Valid: true}
	}
	if termPkg != nil {
		d.TerminalPackage = sql.NullString{String: *termPkg, Valid: true}
	}
	return d
}

func newPackageMockApp(id, domainID int, name string, nvimPkg, termPkg *string) *packageMockApp {
	a := &packageMockApp{ID: id, DomainID: domainID, Name: name}
	if nvimPkg != nil {
		a.NvimPackage = sql.NullString{String: *nvimPkg, Valid: true}
	}
	if termPkg != nil {
		a.TerminalPackage = sql.NullString{String: *termPkg, Valid: true}
	}
	return a
}

func newPackageMockWorkspace(id, appID int, name string, nvimPkg, termPkg *string) *packageMockWorkspace {
	w := &packageMockWorkspace{ID: id, AppID: appID, Name: name}
	if nvimPkg != nil {
		w.NvimPackage = sql.NullString{String: *nvimPkg, Valid: true}
	}
	if termPkg != nil {
		w.TerminalPackage = sql.NullString{String: *termPkg, Valid: true}
	}
	return w
}

func newPackageMockWorkspaceFromNullable(id, appID int, name string, nvimPkg, termPkg sql.NullString) *packageMockWorkspace {
	return &packageMockWorkspace{
		ID: id, AppID: appID, Name: name,
		NvimPackage: nvimPkg, TerminalPackage: termPkg,
	}
}

// SetDefault stores a default in the mock.
func (m *PackageMockDataStore) SetDefault(key, value string) {
	m.defaults[key] = value
}

// GetDefault retrieves a default from the mock.
// Implements the GetDefault method required by PackageDataAccessor.
func (m *PackageMockDataStore) GetDefault(key string) (string, error) {
	if v, ok := m.defaults[key]; ok {
		return v, nil
	}
	return "", errors.New("default not found: " + key)
}

// GetPackageEcosystemByID implements PackageDataAccessor.
func (m *PackageMockDataStore) GetPackageEcosystemByID(id int) (PackageEcosystemData, error) {
	e, ok := m.ecosystems[id]
	if !ok {
		return PackageEcosystemData{}, errors.New("ecosystem not found")
	}
	return PackageEcosystemData{
		ID: e.ID, Name: e.Name,
		NvimPackage: e.NvimPackage, TerminalPackage: e.TerminalPackage,
	}, nil
}

// GetPackageDomainByID implements PackageDataAccessor.
func (m *PackageMockDataStore) GetPackageDomainByID(id int) (PackageDomainData, error) {
	d, ok := m.domains[id]
	if !ok {
		return PackageDomainData{}, errors.New("domain not found")
	}
	return PackageDomainData{
		ID: d.ID, EcosystemID: d.EcosystemID, Name: d.Name,
		NvimPackage: d.NvimPackage, TerminalPackage: d.TerminalPackage,
	}, nil
}

// GetPackageAppByID implements PackageDataAccessor.
func (m *PackageMockDataStore) GetPackageAppByID(id int) (PackageAppData, error) {
	a, ok := m.apps[id]
	if !ok {
		return PackageAppData{}, errors.New("app not found")
	}
	return PackageAppData{
		ID: a.ID, DomainID: a.DomainID, Name: a.Name,
		NvimPackage: a.NvimPackage, TerminalPackage: a.TerminalPackage,
	}, nil
}

// GetPackageWorkspaceByID implements PackageDataAccessor.
func (m *PackageMockDataStore) GetPackageWorkspaceByID(id int) (PackageWorkspaceData, error) {
	w, ok := m.workspaces[id]
	if !ok {
		return PackageWorkspaceData{}, errors.New("workspace not found")
	}
	return PackageWorkspaceData{
		ID: w.ID, AppID: w.AppID, Name: w.Name,
		NvimPackage: w.NvimPackage, TerminalPackage: w.TerminalPackage,
	}, nil
}
