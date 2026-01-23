return {
	"obsidian-nvim/obsidian.nvim",
	version = "*", -- use latest release
	dependencies = {
		"nvim-lua/plenary.nvim",
		"nvim-telescope/telescope.nvim",
	},
	opts = {
		workspaces = {
			{
				name = "BeansBrain",
				path = "~/Library/CloudStorage/OneDrive-RogersCommunicationsInc/obsidian-notes/BeansBrain",
			},
		},
		ui = {
			enable = false,
		},
		legacy_commands = false,
	},
	keys = {
		-- Daily notes
		{ "<leader>od", "<cmd>Obsidian today<cr>", desc = "Obsidian: Today's daily note" },
		{ "<leader>ot", "<cmd>Obsidian tomorrow<cr>", desc = "Obsidian: Tomorrow's daily note" },
		{ "<leader>oy", "<cmd>Obsidian yesterday<cr>", desc = "Obsidian: Yesterday's daily note" },
		{ "<leader>oa", "<cmd>Obsidian dailies<cr>", desc = "Obsidian: All daily notes" },

		-- Note management
		{ "<leader>on", "<cmd>Obsidian new<cr>", desc = "Obsidian: Create new note" },
		{ "<leader>of", "<cmd>Obsidian new_from_template<cr>", desc = "Obsidian: New from template" },
		{ "<leader>oo", "<cmd>Obsidian open<cr>", desc = "Obsidian: Open note in app" },
		{ "<leader>oq", "<cmd>Obsidian quick_switch<cr>", desc = "Obsidian: Quick switch" },
		{ "<leader>ow", "<cmd>Obsidian workspace<cr>", desc = "Obsidian: Switch workspace" },

		-- Search
		{ "<leader>os", "<cmd>Obsidian search<cr>", desc = "Obsidian: Search vault content" },
		{ "<leader>ob", "<cmd>Obsidian backlinks<cr>", desc = "Obsidian: Show backlinks" },
		{ "<leader>ol", "<cmd>Obsidian tags<cr>", desc = "Obsidian: Search by tags" },
		{ "<leader>og", "<cmd>Obsidian follow_link<cr>", desc = "Obsidian: Follow link/smart action" },

		-- Navigation
		{
			"]o",
			function()
				require("obsidian").nav_link("next")
			end,
			desc = "Obsidian: Go to next link",
		},
		{
			"[o",
			function()
				require("obsidian").nav_link("prev")
			end,
			desc = "Obsidian: Go to previous link",
		},

		-- Visual mode actions
		{
			"<leader>oe",
			"<cmd>Obsidian extract_note<cr>",
			mode = "v",
			desc = "Obsidian: Extract to new note",
		},
		{
			"<leader>oll",
			"<cmd>Obsidian link<cr>",
			mode = "v",
			desc = "Obsidian: Link visual selection",
		},
		{
			"<leader>onn",
			"<cmd>Obsidian link_new<cr>",
			mode = "v",
			desc = "Obsidian: Link new note to selection",
		},

		-- Other commands
		{ "<leader>otc", "<cmd>Obsidian toc<cr>", desc = "Obsidian: Show table of contents" },
		{ "<leader>oi", "<cmd>Obsidian paste_img<cr>", desc = "Obsidian: Paste image" },
		{ "<leader>orr", "<cmd>Obsidian rename<cr>", desc = "Obsidian: Rename note" },
		{ "<leader>otgl", "<cmd>Obsidian toggle_checkbox<cr>", desc = "Obsidian: Toggle checkbox" },
	},
}
