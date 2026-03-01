package registry_test

import (
	"context"
	"database/sql"
	"fmt"

	"devopsmaestro/models"
	"devopsmaestro/pkg/registry"
)

// Example: Starting a registry by name (CLI integration)
func ExampleServiceFactory_CreateManager_fromDatabase() {
	// Simulate database lookup
	reg := &models.Registry{
		ID:        1,
		Name:      "my-cache",
		Type:      "zot",
		Port:      5001,
		Lifecycle: "on-demand",
		Status:    "stopped",
	}

	// Create factory
	factory := registry.NewServiceFactory()

	// Create manager from registry resource
	manager, err := factory.CreateManager(reg)
	if err != nil {
		fmt.Printf("Failed to create manager: %v\n", err)
		return
	}

	// Start the service
	ctx := context.Background()
	if err := manager.Start(ctx); err != nil {
		fmt.Printf("Failed to start: %v\n", err)
		return
	}

	// Get endpoint
	endpoint := manager.GetEndpoint()
	fmt.Printf("Registry started: %s\n", endpoint)

	// Check status
	if manager.IsRunning(ctx) {
		fmt.Println("Status: running")
	}

	// Stop when done
	manager.Stop(ctx)

	// Output format (actual output varies):
	// Registry started: localhost:5001
	// Status: running
}

// Example: Creating different registry types
func ExampleServiceFactory_multipleTypes() {
	// NOTE: This is a conceptual example showing multiple registry types.
	// Real Start() calls require actual binaries to be available.

	// Zot container registry (mock for example)
	zotManager := &registry.MockGoModuleProxy{
		GetEndpointFunc: func() string { return "localhost:5000" },
	}
	fmt.Printf("Zot: %s\n", zotManager.GetEndpoint())

	// Athens Go module proxy (mock for example)
	athensManager := &registry.MockGoModuleProxy{
		GetEndpointFunc: func() string { return "http://localhost:3000" },
	}
	fmt.Printf("Athens: %s\n", athensManager.GetEndpoint())

	// Output:
	// Zot: localhost:5000
	// Athens: http://localhost:3000
}

// Example: Using strategies directly
func ExampleZotStrategy_CreateManager() {
	strategy := registry.NewZotStrategy()

	// Get defaults
	fmt.Printf("Default port: %d\n", strategy.GetDefaultPort())
	fmt.Printf("Default storage: %s\n", strategy.GetDefaultStorage())

	// Create manager
	reg := &models.Registry{
		Name: "zot-registry",
		Type: "zot",
		Port: 5001,
	}

	manager, err := strategy.CreateManager(reg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Manager endpoint: %s\n", manager.GetEndpoint())

	// Output:
	// Default port: 5000
	// Default storage: /var/lib/zot
	// Manager endpoint: localhost:5001
}

// Example: Configuration validation
func ExampleRegistryStrategy_ValidateConfig() {
	factory := registry.NewServiceFactory()

	// Valid config
	reg := &models.Registry{
		Name: "configured-zot",
		Type: "zot",
		Port: 5001,
		Config: sql.NullString{
			Valid:  true,
			String: `{"storage": "/custom/path"}`,
		},
	}

	// Validation happens in CreateManager
	_, err := factory.CreateManager(reg)
	if err != nil {
		fmt.Printf("Validation failed: %v\n", err)
	} else {
		fmt.Println("Valid configuration")
	}

	// Invalid config (malformed JSON)
	regInvalid := &models.Registry{
		Name: "invalid",
		Type: "zot",
		Port: 5001,
		Config: sql.NullString{
			Valid:  true,
			String: `{invalid json}`,
		},
	}

	_, err = factory.CreateManager(regInvalid)
	if err != nil {
		fmt.Printf("Error: config validation failed\n")
	}

	// Output:
	// Valid configuration
	// Error: config validation failed
}

// Example: Lifecycle management
func ExampleServiceManager_lifecycle() {
	// NOTE: This is a conceptual example showing the lifecycle pattern.
	// In practice, Start() requires a real binary to be available.
	// See integration tests for actual lifecycle testing.

	ctx := context.Background()

	// Use mock for demonstration
	manager := &registry.MockGoModuleProxy{
		IsRunningFunc:   func(ctx context.Context) bool { return false },
		GetEndpointFunc: func() string { return "localhost:5099" },
	}

	// Check initial state
	if !manager.IsRunning(ctx) {
		fmt.Println("Initially stopped")
	}

	// Simulate start
	manager.IsRunningFunc = func(ctx context.Context) bool { return true }
	fmt.Println("Started successfully")

	// Check running state
	if manager.IsRunning(ctx) {
		fmt.Printf("Running at: %s\n", manager.GetEndpoint())
	}

	// Simulate stop
	manager.IsRunningFunc = func(ctx context.Context) bool { return false }
	fmt.Println("Stopped successfully")

	// Verify stopped
	if !manager.IsRunning(ctx) {
		fmt.Println("Confirmed stopped")
	}

	// Output:
	// Initially stopped
	// Started successfully
	// Running at: localhost:5099
	// Stopped successfully
	// Confirmed stopped
}

// Example: Error handling
func ExampleServiceFactory_errorHandling() {
	factory := registry.NewServiceFactory()

	// Unsupported type
	_, err := factory.GetStrategy("unknown")
	if err != nil {
		fmt.Println("Error: unsupported registry type")
	}

	// Invalid registry (missing name)
	invalidReg := &models.Registry{
		Type: "zot",
		Port: 5000,
	}
	_, err = factory.CreateManager(invalidReg)
	if err != nil {
		fmt.Println("Error: invalid registry")
	}

	// Output:
	// Error: unsupported registry type
	// Error: invalid registry
}

// Example: CLI integration pattern
func Example_cliRegistryStart() {
	// Simulates: dvm registry start my-cache
	// NOTE: This is a conceptual example. Real start requires the binary.

	// 1. Get registry from database
	registryName := "my-cache"
	// reg, err := datastore.GetRegistryByName(registryName)

	// For example, simulate the lookup result:
	reg := &models.Registry{
		ID:        1,
		Name:      registryName,
		Type:      "zot",
		Port:      5001,
		Lifecycle: "on-demand",
		Status:    "stopped",
	}

	// 2. Create service manager (use mock for example)
	manager := &registry.MockGoModuleProxy{
		GetEndpointFunc: func() string { return "localhost:5001" },
	}

	// 3. Simulate successful start
	// In real code: if err := manager.Start(ctx); err != nil { ... }

	// 4. Update database status
	reg.Status = "running"
	// datastore.UpdateRegistry(reg)

	// 5. Show success message
	fmt.Printf("✓ Registry '%s' started\n", registryName)
	fmt.Printf("  Endpoint: %s\n", manager.GetEndpoint())
	fmt.Printf("  Type: %s\n", reg.Type)
	fmt.Printf("  Lifecycle: %s\n", reg.Lifecycle)

	// Output:
	// ✓ Registry 'my-cache' started
	//   Endpoint: localhost:5001
	//   Type: zot
	//   Lifecycle: on-demand
}

// Example: Getting supported types
func ExampleServiceFactory_SupportedTypes() {
	factory := registry.NewServiceFactory()

	types := factory.SupportedTypes()
	fmt.Println("Supported registry types:")
	for _, t := range types {
		port, _ := factory.GetDefaultPort(t)
		storage, _ := factory.GetDefaultStorage(t)
		fmt.Printf("  - %s (port: %d, storage: %s)\n", t, port, storage)
	}

	// Output (order may vary):
	// Supported registry types:
	//   - zot (port: 5000, storage: /var/lib/zot)
	//   - athens (port: 3000, storage: /var/lib/athens)
	//   - devpi (port: 3141, storage: /var/lib/devpi)
	//   - verdaccio (port: 4873, storage: /var/lib/verdaccio)
	//   - squid (port: 3128, storage: /var/cache/squid)
}
