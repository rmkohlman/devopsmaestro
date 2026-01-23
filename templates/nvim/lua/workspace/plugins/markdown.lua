return {
	"MeanderingProgrammer/render-markdown.nvim",
	ft = { "markdown", "quarto" },
	dependencies = {
		"nvim-treesitter/nvim-treesitter",
		"nvim-mini/mini.nvim",
	},
	config = function()
		require("render-markdown").setup({
			-- Whether markdown should be rendered by default.
			enabled = true,

			-- Vim modes that will show a rendered view of the markdown file.
			render_modes = { "n", "c", "t" },

			-- Maximum file size (in MB) that the plugin will attempt to render.
			max_file_size = 10.0,

			-- Any other defaults or overrides can be added here.
			-- The plugin renders common components like headings,
			-- code blocks, and lists with sensible defaults.
		})
	end,
}
