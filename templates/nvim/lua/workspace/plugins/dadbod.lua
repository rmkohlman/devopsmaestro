return {
	-- Core dadbod plugin
	{ "tpope/vim-dadbod" },

	-- Database UI
	{
		"kristijanhusak/vim-dadbod-ui",
		dependencies = { "tpope/vim-dadbod" },
		cmd = { "DBUI", "DBUIToggle" },
		-- Use the 'init' function for global settings that should run early
		init = function()
			-- Use nerd fonts for a nicer visual display
			vim.g.db_ui_use_nerd_fonts = 1
		end,
		-- Use the 'config' function for settings that require the plugin to be loaded
		config = function()
			-- Optional: Set a specific location for saved queries
			vim.g.db_ui_save_location = "~/.config/nvim/db_ui"
		end,
		-- Define the keymaps directly in the plugin spec
		keys = {
			{ "<leader>bu", "<Cmd>DBUIToggle<Cr>", desc = "Toggle UI" },
			{ "<leader>bf", "<Cmd>DBUIFindBuffer<Cr>", desc = "Find buffer" },
			{ "<leader>br", "<Cmd>DBUIRenameBuffer<Cr>", desc = "Rename buffer" },
		},
	},

	-- Autocompletion
	{
		"kristijanhusak/vim-dadbod-completion",
		dependencies = { "tpope/vim-dadbod" },
		ft = { "sql", "mysql", "plsql" },
		lazy = true,
	},
}
