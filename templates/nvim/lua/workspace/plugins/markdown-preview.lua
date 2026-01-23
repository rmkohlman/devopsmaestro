return {
	"iamcco/markdown-preview.nvim",
	ft = "markdown",
	build = "cd app && npm install",
	keys = {
		{ "<leader>mp", "<cmd>MarkdownPreviewToggle<cr>", desc = "Markdown: Toggle Preview" },
	},
	config = function()
		-- Global plugin options
		vim.g.mkdp_auto_close = 1 -- Auto close the preview window when leaving the buffer
		vim.g.mkdp_open_to_the_world = 0 -- Only available locally
		vim.g.mkdp_command_for_global = 0 -- Only enabled for markdown files
		vim.g.mkdp_refresh_slow = 0 -- Auto-refresh as you type
		vim.g.mkdp_page_title = "「${name}」" -- Set the title of the preview page
		vim.g.mkdp_filetypes = { "markdown", "quarto" } -- Enable the plugin for both filetypes
		-- vim.g.mkdp_browser = "google-chrome"
		vim.g.mkdp_echo_preview_url = 1
	end,
}
