// Package handlers provides resource handlers for different resource types.
//
// # Registration
//
// Handlers must be registered before use. Call RegisterAll() at application startup:
//
//	import "devopsmaestro/pkg/resource/handlers"
//
//	func main() {
//	    handlers.RegisterAll()
//	    // ... rest of application
//	}
//
// Or register individual handlers:
//
//	resource.Register(handlers.NewNvimPluginHandler())
//	resource.Register(handlers.NewNvimThemeHandler())
package handlers

import (
	"sync"

	"devopsmaestro/pkg/resource"
)

var registerOnce sync.Once

// RegisterAll registers all available resource handlers.
// Call this at application startup before using the resource package.
// This function is idempotent and safe to call multiple times.
func RegisterAll() {
	registerOnce.Do(func() {
		resource.Register(NewNvimPluginHandler())
		resource.Register(NewNvimThemeHandler())
	})
}
