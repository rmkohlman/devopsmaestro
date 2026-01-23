return {
	-- Installs the main Copilot plugin
	{
		"zbirenbaum/copilot.lua",
		cmd = "Copilot",
		event = "InsertEnter",
		config = function()
			require("copilot").setup({
				suggestion = { enabled = false }, -- Don't use inline suggestions, use cmp instead
				panel = { enabled = false }, -- Don't use the panel, use cmp instead
			})
		end,
	},

	-- Installs the nvim-cmp source for Copilot
	{
		"zbirenbaum/copilot-cmp",
		dependencies = "zbirenbaum/copilot.lua",
		config = function()
			require("copilot_cmp").setup()
		end,
	},
}
