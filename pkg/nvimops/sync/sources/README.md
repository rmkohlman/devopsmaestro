# LazyVim Source Handler

The LazyVim source handler allows nvp to import plugin configurations from the [LazyVim](https://github.com/LazyVim/LazyVim) Neovim configuration framework.

## Features

- **Real-time fetching**: Downloads plugin definitions directly from the LazyVim GitHub repository
- **Smart parsing**: Extracts plugin specifications from Lua files using regex patterns
- **Category mapping**: Maps LazyVim plugin files to appropriate categories (editor, syntax, lsp, etc.)
- **Dependency tracking**: Extracts plugin dependencies from Lua configurations
- **Filtering support**: Filter plugins by category, tags, and other labels
- **Version tracking**: Tracks LazyVim versions for reproducible syncs
- **Caching support**: Tracks last sync SHA to avoid unnecessary re-fetches

## Usage

### List Available Plugins
```bash
# Show all available LazyVim plugins
nvp source sync lazyvim --dry-run

# Filter by category
nvp source sync lazyvim --dry-run -l category=lsp
nvp source sync lazyvim --dry-run -l category=editor
```

### Sync Plugins
```bash
# Sync all LazyVim plugins
nvp source sync lazyvim

# Sync specific categories
nvp source sync lazyvim -l category=coding
nvp source sync lazyvim -l category=ui

# Sync with overwrite (replace existing plugins)
nvp source sync lazyvim --force
```

### Plugin Categories

The handler maps LazyVim files to categories:

| LazyVim File | nvp Category |
|--------------|--------------|
| `coding.lua` | `coding` |
| `colorscheme.lua` | `theme` |
| `editor.lua` | `editor` |
| `formatting.lua` | `formatting` |
| `linting.lua` | `linting` |
| `treesitter.lua` | `syntax` |
| `ui.lua` | `ui` |
| `util.lua` | `utility` |
| `lsp/*.lua` | `lsp` |
| Other files | `misc` |

### Generated Plugin Names

Plugins are prefixed with `lazyvim-` to avoid conflicts:
- `nvim-telescope/telescope.nvim` → `lazyvim-telescope`
- `hrsh7th/nvim-cmp` → `lazyvim-cmp`
- `nvim-treesitter/nvim-treesitter` → `lazyvim-treesitter`

### Labels Added

Each synced plugin includes these labels:
- `source=lazyvim`
- `category=<mapped-category>`
- `lazyvim-file=<original-filename>`
- `lazyvim-version=<version>` (if available)

## Implementation Details

### GitHub API Usage

The handler uses the GitHub Contents API to:
1. List plugin files in `lua/lazyvim/plugins/`
2. Fetch individual file contents via raw URLs
3. Track versions via releases API or commit SHAs

### Lua Parsing Strategy

Uses pragmatic regex-based parsing to extract:
- Plugin repository URLs (`"owner/repo"`)
- Dependencies (`dependencies = { ... }`)
- Configuration blocks (`config = function() ... end`)
- Plugin options (`opts = { ... }`)

**Limitations:**
- Complex Lua functions are not fully parsed
- Conditional logic in plugin specs is ignored
- Only handles common LazyVim patterns

### Error Handling

- Gracefully handles network failures
- Skips unparseable files with warnings
- Continues processing other files if one fails
- Logs errors for debugging

## Testing

Run the test suite:
```bash
go test ./pkg/nvimops/sync/sources/... -v
```

Test real API access:
```bash
nvp source sync lazyvim --dry-run -v
```

## Future Enhancements

- [ ] More sophisticated Lua parsing
- [ ] Support for LazyVim extras/ directory
- [ ] Plugin option extraction improvements
- [ ] Better keymap and event extraction
- [ ] Support for conditional plugin loading
- [ ] Integration with LazyVim update notifications