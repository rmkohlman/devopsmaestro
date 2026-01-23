-- Create a single autocommand group for all yank-related autocommands
local highlight_yank_group = vim.api.nvim_create_augroup("HighlightYank", { clear = true })

-- Define the custom highlight group for yanking whenever the colorscheme changes
vim.api.nvim_create_autocmd("ColorScheme", {
	group = highlight_yank_group,
	callback = function()
		-- vim.api.nvim_set_hl(0, "YankHighlight", { bg = "#FFFF00", fg = "#000000" })
		-- vim.api.nvim_set_hl(0, "YankHighlight", { bg = "#143652", fg = "#CBE0F0" })
		-- vim.api.nvim_set_hl(0, "YankHighlight", { bg = "#f9e2af" })
		-- vim.api.nvim_set_hl(0, "YankHighlight", { bg = "#a6e3a1" })
		vim.api.nvim_set_hl(0, "YankHighlight", { bg = "#45475a" })
	end,
})

-- The autocommand to trigger the highlight effect on yank
vim.api.nvim_create_autocmd("TextYankPost", {
	group = highlight_yank_group,
	callback = function()
		vim.highlight.on_yank({
			timeout = 300,
			higroup = "YankHighlight", -- Use your custom highlight group
		})
	end,
})

-- The 'Y' keymap is redundant because 'Y' defaults to 'yy' in normal Vim/Neovim.
-- If you want to keep it explicit, you can, but it is not necessary.
vim.keymap.set("n", "Y", "yy")
