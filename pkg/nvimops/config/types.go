// Package config provides types and generators for Neovim core configuration.
// It generates the complete lua/workspace/ directory structure from YAML definitions.
package config

// CoreConfig represents the complete Neovim core configuration.
// This maps to the lua/workspace/core/ directory structure.
type CoreConfig struct {
	// APIVersion for forward compatibility (e.g., "nvp/v1")
	APIVersion string `yaml:"apiVersion,omitempty"`

	// Kind identifies this as a CoreConfig
	Kind string `yaml:"kind,omitempty"`

	// Namespace is the lua module namespace (default: "workspace")
	// This determines the directory name: lua/{namespace}/
	Namespace string `yaml:"namespace,omitempty"`

	// Leader key (default: " " for space)
	Leader string `yaml:"leader,omitempty"`

	// Options maps to vim.opt.* settings
	Options map[string]interface{} `yaml:"options,omitempty"`

	// Globals maps to vim.g.* settings
	Globals map[string]interface{} `yaml:"globals,omitempty"`

	// Keymaps defines global key mappings
	Keymaps []Keymap `yaml:"keymaps,omitempty"`

	// Autocmds defines autocommands
	Autocmds []Autocmd `yaml:"autocmds,omitempty"`

	// BasePlugins are simple plugins loaded in plugins/init.lua
	// These are plugins without complex config (just repo strings)
	BasePlugins []string `yaml:"basePlugins,omitempty"`
}

// Keymap represents a single key mapping.
type Keymap struct {
	// Mode(s): "n", "i", "v", "x", or combinations like "nv"
	Mode string `yaml:"mode"`

	// Key is the key sequence (e.g., "<leader>ff", "kj")
	Key string `yaml:"key"`

	// Action is the command or mapping target
	Action string `yaml:"action"`

	// Desc is the description shown in which-key
	Desc string `yaml:"desc,omitempty"`

	// Silent suppresses command output
	Silent bool `yaml:"silent,omitempty"`

	// Noremap prevents recursive mapping (default: true via vim.keymap.set)
	Noremap *bool `yaml:"noremap,omitempty"`
}

// Autocmd represents an autocommand.
type Autocmd struct {
	// Group name for the autocommand group
	Group string `yaml:"group"`

	// Events that trigger this autocmd (e.g., ["BufEnter", "BufNew"])
	Events []string `yaml:"events"`

	// Pattern to match (e.g., "*.go", "*")
	Pattern string `yaml:"pattern,omitempty"`

	// Callback is the Lua code to execute
	Callback string `yaml:"callback,omitempty"`

	// Command is a vim command to execute (alternative to Callback)
	Command string `yaml:"command,omitempty"`

	// Desc describes the autocommand
	Desc string `yaml:"desc,omitempty"`
}

// DefaultCoreConfig returns a sensible default configuration
// that matches the nvim-config repository structure.
func DefaultCoreConfig() *CoreConfig {
	return &CoreConfig{
		APIVersion: "nvp/v1",
		Kind:       "CoreConfig",
		Namespace:  "workspace",
		Leader:     " ",
		Options: map[string]interface{}{
			// Numbers
			"relativenumber": true,
			"number":         true,

			// Tabs & indentation
			"tabstop":    2,
			"shiftwidth": 2,
			"expandtab":  true,
			"autoindent": true,

			// Search
			"ignorecase": true,
			"smartcase":  true,
			"cursorline": true,

			// Appearance
			"termguicolors": true,
			"background":    "dark",
			"signcolumn":    "yes",
			"wrap":          false,

			// Behavior
			"backspace":  "indent,eol,start",
			"splitright": true,
			"splitbelow": true,
			"timeoutlen": 500,
			"clipboard":  "append:unnamedplus",
		},
		Globals: map[string]interface{}{
			"netrw_liststyle": 3,
		},
		Keymaps: []Keymap{
			// Exit insert mode
			{Mode: "i", Key: "kj", Action: "<ESC>", Desc: "Exit insert mode with kj"},

			// Clear search highlights
			{Mode: "n", Key: "<leader>nh", Action: ":nohl<CR>", Desc: "Clear search highlights"},

			// Increment/decrement
			{Mode: "n", Key: "<leader>+", Action: "<C-a>", Desc: "Increment number"},
			{Mode: "n", Key: "<leader>-", Action: "<C-x>", Desc: "Decrement number"},

			// Window management
			{Mode: "n", Key: "<leader>sv", Action: "<C-w>v", Desc: "Split window vertically"},
			{Mode: "n", Key: "<leader>sh", Action: "<C-w>s", Desc: "Split window horizontally"},
			{Mode: "n", Key: "<leader>se", Action: "<C-w>=", Desc: "Make splits equal size"},
			{Mode: "n", Key: "<leader>sx", Action: "<cmd>close<CR>", Desc: "Close current split"},

			// Tab management
			{Mode: "n", Key: "<leader>to", Action: "<cmd>tabnew<CR>", Desc: "Open new tab"},
			{Mode: "n", Key: "<leader>tx", Action: "<cmd>tabclose<CR>", Desc: "Close current tab"},
			{Mode: "n", Key: "<leader>tn", Action: "<cmd>tabn<CR>", Desc: "Go to next tab"},
			{Mode: "n", Key: "<leader>tp", Action: "<cmd>tabp<CR>", Desc: "Go to previous tab"},
			{Mode: "n", Key: "<leader>tf", Action: "<cmd>tabnew %<CR>", Desc: "Open current buffer in new tab"},

			// Move highlighted blocks
			{Mode: "v", Key: "J", Action: ":m '>+1<CR>gv=gv", Silent: true},
			{Mode: "v", Key: "K", Action: ":m '<-2<CR>gv=gv", Silent: true},

			// Yank whole line with Y
			{Mode: "n", Key: "Y", Action: "yy"},
		},
		Autocmds: []Autocmd{
			{
				Group:    "HighlightYank",
				Events:   []string{"ColorScheme"},
				Callback: `vim.api.nvim_set_hl(0, "YankHighlight", { bg = "#45475a" })`,
				Desc:     "Define yank highlight color",
			},
			{
				Group:  "HighlightYank",
				Events: []string{"TextYankPost"},
				Callback: `vim.highlight.on_yank({
	timeout = 300,
	higroup = "YankHighlight",
})`,
				Desc: "Highlight on yank",
			},
		},
		BasePlugins: []string{
			"nvim-lua/plenary.nvim",
			"christoomey/vim-tmux-navigator",
		},
	}
}
