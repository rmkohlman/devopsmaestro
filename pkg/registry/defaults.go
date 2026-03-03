package registry

import (
	"context"

	"devopsmaestro/db"
)

// Registry defaults keys
const (
	DefaultKeyOCI           = "registry-oci"
	DefaultKeyPyPI          = "registry-pypi"
	DefaultKeyNPM           = "registry-npm"
	DefaultKeyGo            = "registry-go"
	DefaultKeyHTTP          = "registry-http"
	DefaultKeyIdleTimeout   = "registry-idle-timeout"
	DefaultIdleTimeoutValue = "30m"
)

// RegistryDefaults provides type-safe access to registry default settings.
type RegistryDefaults struct {
	store db.DataStore
}

// NewRegistryDefaults creates a new RegistryDefaults instance.
func NewRegistryDefaults(store db.DataStore) *RegistryDefaults {
	return &RegistryDefaults{store: store}
}

// GetOCIRegistry returns the default OCI registry name.
func (rd *RegistryDefaults) GetOCIRegistry(ctx context.Context) (string, error) {
	return rd.store.GetDefault(DefaultKeyOCI)
}

// SetOCIRegistry sets the default OCI registry name.
func (rd *RegistryDefaults) SetOCIRegistry(ctx context.Context, registryName string) error {
	return rd.store.SetDefault(DefaultKeyOCI, registryName)
}

// GetPyPIRegistry returns the default PyPI registry name.
func (rd *RegistryDefaults) GetPyPIRegistry(ctx context.Context) (string, error) {
	return rd.store.GetDefault(DefaultKeyPyPI)
}

// SetPyPIRegistry sets the default PyPI registry name.
func (rd *RegistryDefaults) SetPyPIRegistry(ctx context.Context, registryName string) error {
	return rd.store.SetDefault(DefaultKeyPyPI, registryName)
}

// GetNPMRegistry returns the default NPM registry name.
func (rd *RegistryDefaults) GetNPMRegistry(ctx context.Context) (string, error) {
	return rd.store.GetDefault(DefaultKeyNPM)
}

// SetNPMRegistry sets the default NPM registry name.
func (rd *RegistryDefaults) SetNPMRegistry(ctx context.Context, registryName string) error {
	return rd.store.SetDefault(DefaultKeyNPM, registryName)
}

// GetGoRegistry returns the default Go registry name.
func (rd *RegistryDefaults) GetGoRegistry(ctx context.Context) (string, error) {
	return rd.store.GetDefault(DefaultKeyGo)
}

// SetGoRegistry sets the default Go registry name.
func (rd *RegistryDefaults) SetGoRegistry(ctx context.Context, registryName string) error {
	return rd.store.SetDefault(DefaultKeyGo, registryName)
}

// GetHTTPRegistry returns the default HTTP proxy registry name.
func (rd *RegistryDefaults) GetHTTPRegistry(ctx context.Context) (string, error) {
	return rd.store.GetDefault(DefaultKeyHTTP)
}

// SetHTTPRegistry sets the default HTTP proxy registry name.
func (rd *RegistryDefaults) SetHTTPRegistry(ctx context.Context, registryName string) error {
	return rd.store.SetDefault(DefaultKeyHTTP, registryName)
}

// GetIdleTimeout returns the global idle timeout setting.
// Returns "30m" as default if not set.
func (rd *RegistryDefaults) GetIdleTimeout(ctx context.Context) (string, error) {
	value, err := rd.store.GetDefault(DefaultKeyIdleTimeout)
	if err != nil {
		return "", err
	}
	if value == "" {
		return DefaultIdleTimeoutValue, nil
	}
	return value, nil
}

// SetIdleTimeout sets the global idle timeout setting.
func (rd *RegistryDefaults) SetIdleTimeout(ctx context.Context, timeout string) error {
	return rd.store.SetDefault(DefaultKeyIdleTimeout, timeout)
}

// GetAllDefaults returns all registry defaults as a map.
func (rd *RegistryDefaults) GetAllDefaults(ctx context.Context) (map[string]string, error) {
	result := make(map[string]string)

	// Get all registry type defaults
	oci, err := rd.GetOCIRegistry(ctx)
	if err != nil {
		return nil, err
	}
	result[DefaultKeyOCI] = oci

	pypi, err := rd.GetPyPIRegistry(ctx)
	if err != nil {
		return nil, err
	}
	result[DefaultKeyPyPI] = pypi

	npm, err := rd.GetNPMRegistry(ctx)
	if err != nil {
		return nil, err
	}
	result[DefaultKeyNPM] = npm

	goReg, err := rd.GetGoRegistry(ctx)
	if err != nil {
		return nil, err
	}
	result[DefaultKeyGo] = goReg

	http, err := rd.GetHTTPRegistry(ctx)
	if err != nil {
		return nil, err
	}
	result[DefaultKeyHTTP] = http

	// Get idle timeout (will return default if not set)
	timeout, err := rd.GetIdleTimeout(ctx)
	if err != nil {
		return nil, err
	}
	result[DefaultKeyIdleTimeout] = timeout

	return result, nil
}

// ClearDefault clears a specific default setting.
func (rd *RegistryDefaults) ClearDefault(ctx context.Context, key string) error {
	return rd.store.DeleteDefault(key)
}

// GetByAlias gets a default registry by type alias.
func (rd *RegistryDefaults) GetByAlias(ctx context.Context, alias string) (string, error) {
	key := getKeyForAlias(alias)
	return rd.store.GetDefault(key)
}

// SetByAlias sets a default registry by type alias.
func (rd *RegistryDefaults) SetByAlias(ctx context.Context, alias string, registryName string) error {
	key := getKeyForAlias(alias)
	return rd.store.SetDefault(key, registryName)
}

// getKeyForAlias converts an alias to its default key.
func getKeyForAlias(alias string) string {
	switch alias {
	case AliasOCI:
		return DefaultKeyOCI
	case AliasPyPI:
		return DefaultKeyPyPI
	case AliasNPM:
		return DefaultKeyNPM
	case AliasGo:
		return DefaultKeyGo
	case AliasHTTP:
		return DefaultKeyHTTP
	default:
		// If not a known alias, just prefix with "registry-"
		return "registry-" + alias
	}
}
