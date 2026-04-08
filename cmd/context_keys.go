package cmd

import (
	"fmt"
	"io/fs"
)

// contextKey is a private type for context keys defined in this package.
// Using a typed key prevents collisions with keys from other packages.
type contextKey string

const (
	// CtxKeyDataStore is the context key for the DataStore (db.DataStore).
	// Exported so cmd/dvt and cmd/nvp can share the same key.
	CtxKeyDataStore contextKey = "dataStore"

	// ctxKeyExecutor is the context key for the Executor interface.
	ctxKeyExecutor contextKey = "executor"

	// ctxKeyMigrationsFS is the context key for the migrations filesystem.
	ctxKeyMigrationsFS contextKey = "migrationsFS"
)

// getMigrationsFS extracts the migrations filesystem from the cobra command context.
func getMigrationsFSFromContext(ctx interface{ Value(any) any }) (fs.FS, error) {
	val := ctx.Value(ctxKeyMigrationsFS)
	if val == nil {
		return nil, errMissingContext("migrationsFS")
	}
	if migrationsFS, ok := val.(fs.FS); ok {
		return migrationsFS, nil
	}
	return nil, errInvalidType("migrationsFS", val)
}

// errMissingContext returns a standardised "not found in context" error.
func errMissingContext(name string) error {
	return &contextError{key: name, msg: "not found in context"}
}

// errInvalidType returns a standardised "invalid type" error.
func errInvalidType(name string, val any) error {
	return &contextError{key: name, msg: "invalid type in context", val: val}
}

// contextError is a structured error for context lookup failures.
type contextError struct {
	key string
	msg string
	val any // non-nil only for type errors
}

func (e *contextError) Error() string {
	if e.val != nil {
		return e.key + " " + e.msg + ": " + typeString(e.val)
	}
	return e.key + " " + e.msg
}

// typeString returns the Go type of a value as a string (e.g. "*db.SQLiteDataStore").
func typeString(v any) string {
	if v == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%T", v)
}
