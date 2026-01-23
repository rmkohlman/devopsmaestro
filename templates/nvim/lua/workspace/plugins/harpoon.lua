return {
	"ThePrimeagen/harpoon",
	branch = "harpoon2",
	dependencies = { "nvim-lua/plenary.nvim" },
	config = function()
		local harpoon = require("harpoon")
		local keymap = vim.keymap -- for conciseness

		harpoon:setup({})

		-- Harpoon mappings with <leader>m prefix for "marks"
		keymap.set("n", "<leader>ma", function()
			harpoon:list():add()
		end, { desc = "Harpoon: Mark file" })
		keymap.set("n", "<leader>mm", function()
			harpoon.ui:toggle_quick_menu(harpoon:list())
		end, { desc = "Harpoon: Toggle quick menu" })

		-- Telescope configuration for Harpoon
		local conf = require("telescope.config").values
		local function toggle_telescope(harpoon_files)
			local file_paths = {}
			for _, item in ipairs(harpoon_files.items) do
				table.insert(file_paths, item.value)
			end

			require("telescope.pickers")
				.new({}, {
					prompt_title = "Harpoon",
					finder = require("telescope.finders").new_table({
						results = file_paths,
					}),
					previewer = conf.file_previewer({}),
					sorter = conf.generic_sorter({}),
				})
				:find()
		end

		keymap.set("n", "<leader>ms", function()
			toggle_telescope(harpoon:list())
		end, { desc = "Harpoon: Open with Telescope" })
	end,
}
