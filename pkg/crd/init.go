package crd

import (
	"fmt"
	"log/slog"
	"sync"

	"devopsmaestro/db"
	"devopsmaestro/pkg/resource"
)

var (
	initOnce sync.Once
	initErr  error
)

// InitializeFallbackHandler sets up the DynamicHandler as the fallback
// for handling custom resource types defined by CRDs.
// This should be called once per CLI invocation, after the DataStore is available.
//
// The initialization process:
//  1. Creates a DataStoreAdapter to bridge db.CustomResourceStore and CRD interfaces
//  2. Creates a CRDResolver and loads existing CRDs from the database
//  3. Creates schema and scope validators
//  4. Creates the DynamicHandler with all dependencies
//  5. Registers it as the fallback handler in pkg/resource
//
// If no CRDs are found during initialization, this is not considered an error.
// CRDs can be created later via `dvm apply -f crd.yaml`.
func InitializeFallbackHandler(ds db.CustomResourceStore) error {
	initOnce.Do(func() {
		if ds == nil {
			initErr = fmt.Errorf("datastore cannot be nil")
			return
		}

		// Create adapter that implements CRDStore and CustomResourceStore
		adapter := NewDataStoreAdapter(ds)

		// Create resolver and load existing CRDs
		resolver := NewCRDResolver(adapter)
		if err := resolver.Refresh(); err != nil {
			// Log but don't fail - there may not be any CRDs yet
			// This is OK - CRDs can be created later
			slog.Debug("no CRDs loaded during initialization", "error", err)
		}

		// Create validators
		schemaValidator := NewSchemaValidator()
		scopeValidator := NewScopeValidator()

		// Create handler
		handler := NewDynamicHandler(resolver, schemaValidator, scopeValidator, adapter)

		// Register as fallback
		resource.SetFallbackHandler(handler)

		slog.Debug("CRD fallback handler initialized successfully")
	})

	return initErr
}

// ResetInitialization resets the initialization state.
// This is primarily for testing purposes.
func ResetInitialization() {
	initOnce = sync.Once{}
	initErr = nil
}
