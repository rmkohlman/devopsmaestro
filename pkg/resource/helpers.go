package resource

import "fmt"

// DataStoreAs extracts the DataStore from a Context and asserts it to the given type T.
// This provides compile-time-safe extraction without coupling the resource package
// to any specific DataStore implementation.
//
// Usage:
//
//	ds, err := resource.DataStoreAs[db.EcosystemStore](ctx)
//	ds, err := resource.DataStoreAs[db.DataStore](ctx)
func DataStoreAs[T any](ctx Context) (T, error) {
	if ctx.DataStore == nil {
		var zero T
		return zero, fmt.Errorf("DataStore not provided in context")
	}
	ds, ok := ctx.DataStore.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("DataStore does not implement %T (got %T)", zero, ctx.DataStore)
	}
	return ds, nil
}

// PluginStoreAs extracts the PluginStore from a Context and asserts it to the given type T.
// This provides compile-time-safe extraction without coupling the resource package
// to any specific PluginStore implementation.
//
// Usage:
//
//	ps, err := resource.PluginStoreAs[store.PluginStore](ctx)
func PluginStoreAs[T any](ctx Context) (T, error) {
	if ctx.PluginStore == nil {
		var zero T
		return zero, fmt.Errorf("PluginStore not provided in context")
	}
	ps, ok := ctx.PluginStore.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("PluginStore does not implement %T (got %T)", zero, ctx.PluginStore)
	}
	return ps, nil
}

// ThemeStoreAs extracts the ThemeStore from a Context and asserts it to the given type T.
// This provides compile-time-safe extraction without coupling the resource package
// to any specific ThemeStore implementation.
//
// Usage:
//
//	ts, err := resource.ThemeStoreAs[theme.Store](ctx)
func ThemeStoreAs[T any](ctx Context) (T, error) {
	if ctx.ThemeStore == nil {
		var zero T
		return zero, fmt.Errorf("ThemeStore not provided in context")
	}
	ts, ok := ctx.ThemeStore.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("ThemeStore does not implement %T (got %T)", zero, ctx.ThemeStore)
	}
	return ts, nil
}
