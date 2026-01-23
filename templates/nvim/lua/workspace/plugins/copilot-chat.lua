return {
	"CopilotC-Nvim/CopilotChat.nvim",
	branch = "canary", -- Use the canary branch for the latest updates
	dependencies = {
		-- The core Copilot plugin
		{ "zbirenbaum/copilot.lua" },
		-- Plenary for async functions
		{ "nvim-lua/plenary.nvim" },
		-- Telescope for interactive prompts
		{ "nvim-telescope/telescope.nvim" },
	},

	-- Configure the plugin with easy-to-use key mappings
	config = function()
		local chat = require("CopilotChat")
		local keymap = vim.keymap.set

		-- --- START OF ADDED CODE FOR GLOB SUPPORT ---

		-- Define a helper function to select files via glob pattern
		local glob_select = function(opts)
			opts = opts or {}
			-- Use a default pattern if none is provided, or prompt the user
			local pattern = opts.pattern or vim.fn.input("Glob pattern: ")
			if pattern == "" then
				return {}
			end

			-- Use vim.fn.glob to find files. 'true' enables globstar (**) and list format
			local files = vim.split(vim.fn.glob(pattern, true, true) or "", "\n")
			local resources = {}
			for _, file in ipairs(files) do
				if file ~= "" then
					table.insert(resources, {
						uri = vim.uri.from_local(file),
						content = vim.fn.readfile(file),
					})
				end
			end
			return resources
		end

		-- Register the new selection function in the CopilotChat module
		require("CopilotChat.select").glob = glob_select

		-- Configure the plugin with easy-to-use key mappings
		-- config = function()
		-- 	local chat = require("CopilotChat")
		-- 	local keymap = vim.keymap.set

		chat.setup({
			-- Customize window behavior
			window = {
				layout = "float", -- Use a floating window for chat
				width = 0.5,
			},
			-- Key mappings for normal and visual modes
			prompts = {
				-- Use built-in prompts
				Explain = {
					prompt = "Explain this code",
					selection = require("CopilotChat.select").visual,
				},
				Fix = {
					prompt = "Fix this code",
					selection = require("CopilotChat.select").visual,
				},
				Tests = {
					prompt = "Create tests for this code",
					selection = require("CopilotChat.select").visual,
				},
				Review = {
					prompt = "Review this code for security issues and refactor suggestions",
					selection = require("CopilotChat.select").visual,
				},
			},
		})

		-- General chat commands
		keymap("n", "<leader>cc", "<cmd>CopilotChat<cr>", { desc = "Toggle Copilot Chat" })
		keymap("n", "<leader>cca", "<cmd>CopilotChatAdd<cr>", { desc = "Add code to chat context" })
		keymap("n", "<leader>ccq", function()
			local input = vim.fn.input("Quick Chat: ")
			if input ~= "" then
				require("CopilotChat").ask(input)
			end
		end, { desc = "Quick chat" })
		keymap("n", "<leader>ccp", "<cmd>CopilotChatPrompts<cr>", { desc = "View prompt templates" })
		keymap("n", "<leader>cch", "<cmd>CopilotChatHistory<cr>", { desc = "View chat history" })
		keymap("n", "<leader>cce", "<cmd>CopilotChatReset<cr>", { desc = "Reset chat session" })

		-- Visual mode prompts (use built-in prompts)
		keymap("v", "<leader>ccx", "<cmd>CopilotChat Explain<cr>", { desc = "Explain visual selection" })
		keymap("v", "<leader>ccf", "<cmd>CopilotChat Fix<cr>", { desc = "Fix visual selection" })
		keymap("v", "<leader>cct", "<cmd>CopilotChat Tests<cr>", { desc = "Test visual selection" })
		keymap("v", "<leader>ccr", "<cmd>CopilotChat Review<cr>", { desc = "Review visual selection" })

		-- Chat window key mappings (for when the chat window is open)
		vim.api.nvim_create_autocmd("FileType", {
			pattern = "copilot-chat",
			callback = function()
				-- Close the chat window with q
				keymap("n", "q", "<cmd>CopilotChatClose<cr>", { buffer = true, desc = "Close chat" })
				-- Move up and down history with Ctrl-p/n
				keymap(
					"n",
					"<c-p>",
					"<cmd>CopilotChatInputHistoryPrevious<cr>",
					{ buffer = true, desc = "Previous prompt" }
				)
				keymap("n", "<c-n>", "<cmd>CopilotChatInputHistoryNext<cr>", { buffer = true, desc = "Next prompt" })

				-- Define a command to make the previous buffer modifiable
				local make_prev_buf_modifiable = function()
					-- Store current window info, switch to previous buffer's window, set modifiable
					local current_win = vim.api.nvim_get_current_win()
					vim.api.nvim_win_set_buf(0, vim.fn.bufnr("#"))
					vim.opt_local.modifiable = true
					vim.api.nvim_set_current_win(current_win) -- Switch back to chat window
				end

				-- Wrap the yank/accept command with the pre-command
				keymap("n", "<C-y>", function()
					make_prev_buf_modifiable()
					vim.cmd("CopilotChatYankDiff") -- Use the explicit command to apply diff
				end, { buffer = true, desc = "Accept nearest diff and make modifiable" })

				keymap("n", "gd", function()
					make_prev_buf_modifiable()
					vim.cmd("CopilotChatDiff") -- Use the explicit command to show diff
				end, { buffer = true, desc = "Show diff and make modifiable" })
			end,
		})
	end,
}
