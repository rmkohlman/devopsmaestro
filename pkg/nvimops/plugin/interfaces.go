// Package plugin provides types and utilities for Neovim plugin management.
package plugin

// LuaGenerator defines the interface for converting plugins to Lua code.
// This allows different Lua generation strategies to be swapped in.
type LuaGenerator interface {
	// GenerateLua converts a Plugin to lazy.nvim compatible Lua code.
	// Returns the raw Lua code without file headers.
	GenerateLua(p *Plugin) (string, error)

	// GenerateLuaFile generates Lua with a header comment for file output.
	GenerateLuaFile(p *Plugin) (string, error)
}

// Verify Generator implements LuaGenerator interface
var _ LuaGenerator = (*Generator)(nil)
