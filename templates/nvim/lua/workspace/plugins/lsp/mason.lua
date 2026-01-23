return {
	"williamboman/mason.nvim",
	dependencies = {
		"williamboman/mason-lspconfig.nvim",
		"WhoIsSethDaniel/mason-tool-installer.nvim",
	},
	config = function()
		-- import mason
		local mason = require("mason")

		-- import mason-lspconfig
		local mason_lspconfig = require("mason-lspconfig")

		local mason_tool_installer = require("mason-tool-installer")

		-- enable mason and configure icons
		mason.setup({
			ui = {
				icons = {
					package_installed = "✓",
					package_pending = "➜",
					package_uninstalled = "✗",
				},
			},
		})

		mason_lspconfig.setup({
			-- list of LSP servers for mason to install
			ensure_installed = {
				"gopls",
				"rust_analyzer",
				"pyright",
				"lua_ls",
				-- "ocaml_lsp", -- Correct LSP server name
				"bashls",
				"clangd",
				-- "tsserver", -- Correct LSP server name
				"html", -- Correct name for HTML server
				"cssls", -- Correct name for CSS server
				"tailwindcss", -- Correct name for TailwindCSS server
				"svelte", -- Correct name for Svelte server
				"yamlls", -- Correct name for YAML server
				"graphql",
				"emmet_ls",
				"prismals",
				"dockerls", -- Correct name for Dockerfile server
				"helm_ls",
				"jsonls", -- Correct name for JSON server
				"lemminx", -- Correct name for XML server
				"sqlls",
				-- "promql", -- Correct name for PromQL server
				-- "aws_lsp",
				-- "ruby_lsp" might need special configuration.
				-- The name is correct but has known issues in Neovim.
				-- "postgresql_language_server" is a valid name,
				-- but `sqlls` might cover it. You can try adding it
				-- if needed.
			},
		})

		mason_tool_installer.setup({
			ensure_installed = {
				"prettier", -- prettier formatter
				"stylua", -- lua formatter
				"isort", -- python formatter
				"black", -- python formatter
				"pylint",
				"eslint_d", -- Correct linter name
				"ocamlformat",
				"goimports",
				"shellcheck", -- Correct linter name
			},
		})
	end,
}
