package registry

// Type aliases for registry types
const (
	AliasOCI  = "oci"
	AliasPyPI = "pypi"
	AliasNPM  = "npm"
	AliasGo   = "go"
	AliasHTTP = "http"

	TypeZot       = "zot"
	TypeDevpi     = "devpi"
	TypeVerdaccio = "verdaccio"
	TypeAthens    = "athens"
	TypeSquid     = "squid"
)

// aliasMap defines the mapping from type aliases to concrete registry types
var aliasMap = map[string]string{
	AliasOCI:  TypeZot,
	AliasPyPI: TypeDevpi,
	AliasNPM:  TypeVerdaccio,
	AliasGo:   TypeAthens,
	AliasHTTP: TypeSquid,
}

// reverseAliasMap defines the reverse mapping from registry types to aliases
var reverseAliasMap = map[string]string{
	TypeZot:       AliasOCI,
	TypeDevpi:     AliasPyPI,
	TypeVerdaccio: AliasNPM,
	TypeAthens:    AliasGo,
	TypeSquid:     AliasHTTP,
}

// ResolveAlias resolves a type alias to its concrete registry type.
// If the input is not a known alias, it returns the input unchanged.
func ResolveAlias(alias string) string {
	if resolved, ok := aliasMap[alias]; ok {
		return resolved
	}
	return alias
}

// GetAliasForType returns the alias for a given registry type.
// Returns empty string and false if no alias exists.
func GetAliasForType(registryType string) (string, bool) {
	alias, ok := reverseAliasMap[registryType]
	return alias, ok
}

// GetAllAliases returns a map of all type aliases to their registry types.
func GetAllAliases() map[string]string {
	// Return a copy to prevent external modification
	result := make(map[string]string, len(aliasMap))
	for k, v := range aliasMap {
		result[k] = v
	}
	return result
}
