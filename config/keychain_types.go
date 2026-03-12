package config

// KeychainType represents the macOS keychain entry class.
type KeychainType string

const (
	// KeychainTypeGeneric identifies generic password entries (class genp).
	// These are created by applications like dvm, gh CLI, or Keychain Access "New Password Item".
	KeychainTypeGeneric KeychainType = "generic"

	// KeychainTypeInternet identifies internet password entries (class inet).
	// These are created by the macOS Passwords app, Safari, and web browsers.
	KeychainTypeInternet KeychainType = "internet"
)

// KeychainLookup encapsulates the parameters for a keychain search.
type KeychainLookup struct {
	Label        string       // The keychain entry's display name (matched with -l flag)
	KeychainType KeychainType // "generic" or "internet"
}
