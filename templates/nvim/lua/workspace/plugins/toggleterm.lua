-- return {
--   -- amongst your other plugins
--   {'akinsho/toggleterm.nvim', version = "*", config = true},
--
--   require("toggleterm").setup{
-- 	direction = "horizontal",
-- 	size = 40,
-- 	open_mapping = [[<M-j>]]
--   },
-- }

return {
	"akinsho/toggleterm.nvim",
	version = "*", -- Specify a version or tag to avoid breaking changes
	opts = {
		-- Change the direction to 'float', 'vertical', or 'horizontal'
		direction = "float",
		-- Set the size of the terminal. Can be a number or a function
		size = 20,
		-- Configure keymaps for opening and closing the terminal
		open_mapping = [[<M-j>]],
		-- Make the terminal floating window look curved
		float_opts = {
			border = "curved",
		},
		-- Set to false to see the buffer number, otherwise true
		hide_numbers = true,
		-- Make the background of terminals darker
		shade_terminals = true,
	},
	config = function(_, opts)
		require("toggleterm").setup(opts)
	end,
}
