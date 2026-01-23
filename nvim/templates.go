package nvim

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// initFromTemplate initializes config from a template
func (m *manager) initFromTemplate(opts InitOptions) error {
	switch opts.Template {
	case "kickstart":
		return m.initKickstart(opts)
	case "lazyvim":
		return m.initLazyVim(opts)
	case "astronvim":
		return m.initAstroNvim(opts)
	case "minimal":
		return m.initMinimal(opts)
	case "custom":
		if opts.GitURL == "" {
			return fmt.Errorf("custom template requires --git-url")
		}
		return m.initCustom(opts)
	default:
		return fmt.Errorf("unknown template: %s (available: kickstart, lazyvim, astronvim, minimal, custom)", opts.Template)
	}
}

// initKickstart initializes with kickstart.nvim template
func (m *manager) initKickstart(opts InitOptions) error {
	if !opts.GitClone {
		return m.initMinimal(opts)
	}

	gitURL := "https://github.com/nvim-lua/kickstart.nvim.git"
	return m.cloneTemplate(gitURL, opts.ConfigPath, opts.Subdir)
}

// initLazyVim initializes with LazyVim template
func (m *manager) initLazyVim(opts InitOptions) error {
	if !opts.GitClone {
		return m.initMinimal(opts)
	}

	gitURL := "https://github.com/LazyVim/starter.git"
	return m.cloneTemplate(gitURL, opts.ConfigPath, opts.Subdir)
}

// initAstroNvim initializes with AstroNvim template
func (m *manager) initAstroNvim(opts InitOptions) error {
	if !opts.GitClone {
		return m.initMinimal(opts)
	}

	gitURL := "https://github.com/AstroNvim/template.git"
	return m.cloneTemplate(gitURL, opts.ConfigPath, opts.Subdir)
}

// initCustom initializes from a custom Git URL
func (m *manager) initCustom(opts InitOptions) error {
	if opts.GitURL == "" {
		return fmt.Errorf("custom template requires GitURL to be set")
	}
	return m.cloneTemplate(opts.GitURL, opts.ConfigPath, opts.Subdir)
}

// initMinimal creates a minimal init.lua
func (m *manager) initMinimal(opts InitOptions) error {
	// Get the template from embedded files or from templates/ directory
	var initLua []byte
	var err error

	// Try to read from templates directory (for development)
	templatePath := "templates/minimal/init.lua"
	initLua, err = os.ReadFile(templatePath)

	// If not found, use embedded template (fallback)
	if err != nil {
		initLua = []byte(minimalTemplate)
	}

	initPath := filepath.Join(opts.ConfigPath, "init.lua")
	return os.WriteFile(initPath, initLua, 0644)
}

// Embedded minimal template (fallback if templates/ not available)
const minimalTemplate = `-- DevOpsMaestro Minimal Neovim Configuration
-- Perfect for beginners and those who want to build their config from scratch

-- Set leader keys
vim.g.mapleader = " "
vim.g.maplocalleader = "\\"

-- Basic options
vim.opt.number = true
vim.opt.relativenumber = true
vim.opt.mouse = 'a'
vim.opt.clipboard = 'unnamedplus'
vim.opt.ignorecase = true
vim.opt.smartcase = true
vim.opt.hlsearch = true
vim.opt.signcolumn = 'yes'
vim.opt.updatetime = 250

-- Basic keymaps
vim.keymap.set('n', '<Esc>', '<cmd>nohlsearch<CR>')
vim.keymap.set('n', '<leader>w', '<cmd>w<CR>')
vim.keymap.set('n', '<leader>q', '<cmd>q<CR>')

-- Plugin manager (lazy.nvim)
local lazypath = vim.fn.stdpath 'data' .. '/lazy/lazy.nvim'
if not vim.loop.fs_stat(lazypath) then
  vim.fn.system {
    'git', 'clone', '--filter=blob:none',
    'https://github.com/folke/lazy.nvim.git',
    '--branch=stable', lazypath,
  }
end
vim.opt.rtp:prepend(lazypath)

-- Plugins
require('lazy').setup({
  { 'catppuccin/nvim', name = 'catppuccin', priority = 1000,
    config = function() vim.cmd.colorscheme 'catppuccin-mocha' end },
  'tpope/vim-sleuth',
}, {})
`

// cloneTemplate clones a git repository as the config
func (m *manager) cloneTemplate(gitURL, configPath, subdir string) error {
	// Normalize URL (handle github:user/repo format)
	gitURL = NormalizeGitURL(gitURL)

	// Remove .git-keep or any existing files if directory exists
	os.RemoveAll(configPath)

	// Clone to a temporary directory if we need a subdirectory
	targetPath := configPath
	if subdir != "" {
		// Clone to temp directory
		tmpDir, err := os.MkdirTemp("", "dvm-nvim-clone-*")
		if err != nil {
			return fmt.Errorf("failed to create temp directory: %w", err)
		}
		defer os.RemoveAll(tmpDir)
		targetPath = tmpDir
	}

	// Clone the repository
	cmd := exec.Command("git", "clone", "--depth=1", gitURL, targetPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %w\nOutput: %s", err, output)
	}

	// If subdirectory specified, move it to the config path
	if subdir != "" {
		subdirPath := filepath.Join(targetPath, subdir)

		// Check if subdirectory exists
		if _, err := os.Stat(subdirPath); os.IsNotExist(err) {
			return fmt.Errorf("subdirectory %s not found in repository", subdir)
		}

		// Move subdirectory contents to config path
		if err := os.MkdirAll(configPath, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		// Copy subdirectory contents
		copyCmd := exec.Command("cp", "-r", subdirPath+"/.", configPath)
		if output, err := copyCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to copy subdirectory: %w\nOutput: %s", err, output)
		}
	}

	// Remove .git directory to make it independent
	gitDir := filepath.Join(configPath, ".git")
	if err := os.RemoveAll(gitDir); err != nil {
		// Not fatal, just warn
		fmt.Printf("Warning: failed to remove .git directory: %v\n", err)
	}

	return nil
}

// saveStatus saves the current status to a file
func (m *manager) saveStatus(status *Status) error {
	// Create directory if it doesn't exist
	statusDir := filepath.Dir(m.statusFile)
	if err := os.MkdirAll(statusDir, 0755); err != nil {
		return fmt.Errorf("failed to create status directory: %w", err)
	}

	// Marshal status to JSON
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	// Write to file
	if err := os.WriteFile(m.statusFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write status file: %w", err)
	}

	return nil
}

// loadStatus loads the status from a file
func (m *manager) loadStatus() (*Status, error) {
	// Read file
	data, err := os.ReadFile(m.statusFile)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON
	var status Status
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("failed to unmarshal status: %w", err)
	}

	return &status, nil
}
