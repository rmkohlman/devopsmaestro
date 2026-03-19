# opencode.nvim Plugin

[opencode.nvim](https://github.com/nickjvandyke/opencode.nvim) brings the opencode AI assistant into Neovim. Use it to ask questions, generate code, and send context from your editor — all without leaving nvim.

---

## How it works

opencode.nvim communicates with opencode over an HTTP API on port 4096. When you trigger a keybinding, the plugin either:

- **Connects to an existing opencode instance** running in your terminal, or
- **Starts an embedded opencode terminal** inside Neovim if none is running

This means the plugin works standalone — you do not need the [opencode workspace tool](../dvm/opencode.md) to use it. You do need an API key in your environment.

---

## Installing the plugin

### From the library

```bash
# Install snacks.nvim (recommended dependency)
nvp library install snacks

# Install opencode.nvim
nvp library install opencode

# Regenerate Lua configuration
nvp generate
```

### From the rmkohlman package

The `rmkohlman` package includes both `snacks` and `opencode` pre-configured:

```bash
nvp library install-package rmkohlman
nvp generate
```

### From a YAML file or URL

```bash
# From GitHub (always latest library version)
nvp apply -f github:rmkohlman/nvim-yaml-plugins/plugins/snacks.yaml
nvp apply -f github:rmkohlman/nvim-yaml-plugins/plugins/opencode.yaml
nvp generate
```

---

## Workspace YAML setup

To add opencode.nvim via workspace declaration, include `opencode` and `snacks` in your nvim plugin list:

```yaml
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: dev
  app: my-api
spec:
  nvim:
    structure: lazyvim
    pluginPackage: rmkohlman   # includes snacks + opencode already
```

Or add them to a custom plugin list:

```yaml
spec:
  nvim:
    structure: lazyvim
    plugins:
      - folke/snacks.nvim
      - nickjvandyke/opencode.nvim
    mergeMode: append
```

---

## Default keybindings

| Key | Mode | Action |
|-----|------|--------|
| `<C-a>` | Normal, Visual | Ask opencode (`@this` context) |
| `<C-x>` | Normal, Visual | Select opencode action |
| `<C-.>` | Normal, Terminal | Toggle opencode window |
| `go` | Normal, Visual | Add range to opencode (operator) |
| `goo` | Normal | Add current line to opencode |
| `<S-C-u>` | Normal | Scroll opencode up |
| `<S-C-d>` | Normal | Scroll opencode down |

> **Note:** `<C-a>` and `<C-x>` replace Vim's built-in increment/decrement. Those operations are remapped to `+` (increment) and `-` (decrement).

### snacks.nvim picker integration

When using snacks.nvim, you can send picker results directly to opencode:

| Key | Context | Action |
|-----|---------|--------|
| `<A-a>` | Any snacks picker | Send selected item(s) to opencode |

---

## Plugin YAML definitions

### opencode.nvim

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: opencode
  category: ai
spec:
  repo: nickjvandyke/opencode.nvim
  version: "v0.5.2"
  dependencies:
    - repo: folke/snacks.nvim
  cmd: ["Opencode"]
  config: |
    ---@type opencode.Opts
    vim.g.opencode_opts = {}
    vim.o.autoread = true

    vim.keymap.set({ "n", "x" }, "<C-a>", function()
      require("opencode").ask("@this: ", { submit = true })
    end, { desc = "Ask opencode" })

    vim.keymap.set({ "n", "t" }, "<C-.>", function()
      require("opencode").toggle()
    end, { desc = "Toggle opencode" })
    -- ... (see library YAML for full keybindings)
```

### snacks.nvim

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: snacks
  category: utility
spec:
  repo: folke/snacks.nvim
  version: "v2.30.0"
  event: ["VeryLazy"]
  config: |
    require("snacks").setup({
      input = {},
      picker = {
        actions = {
          opencode_send = function(...)
            return require("opencode").snacks_picker_send(...)
          end,
        },
        win = {
          input = {
            keys = { ["<a-a>"] = { "opencode_send", mode = { "n", "i" } } },
          },
        },
      },
    })
```

---

## API key configuration

opencode.nvim starts opencode using the API key from your environment. Set the appropriate variable before launching Neovim (or before `dvm attach`):

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
nvim     # or: dvm attach
```

See [API key configuration](../dvm/opencode.md#api-key-configuration) in the workspace tool guide for how to manage keys with `dvm create credential`.

---

## Managing the plugin

```bash
# Check installation status
nvp get opencode
nvp get snacks

# View full YAML definition
nvp get opencode -o yaml

# Disable without removing
nvp disable opencode

# Re-enable
nvp enable opencode

# Remove entirely
nvp delete opencode
nvp delete snacks
```

After any change, regenerate:

```bash
nvp generate
```

---

## Combined setup with the workspace CLI tool

Using opencode.nvim with the workspace tool gives you the full experience:

1. The workspace tool installs the opencode TUI binary in your container
2. The nvim plugin connects to that running instance over port 4096
3. Context from your editor flows directly into the opencode session

Enable both in your workspace YAML:

```yaml
spec:
  tools:
    opencode: true          # installs the CLI binary in the container
  nvim:
    pluginPackage: rmkohlman  # includes snacks + opencode plugin
```

See [opencode CLI Tool](../dvm/opencode.md) for workspace tool setup details.

---

## Troubleshooting

### Plugin not loading

```bash
# Verify installation
nvp get opencode
nvp get snacks

# Regenerate Lua files
nvp generate

# Check Neovim health inside nvim
:checkhealth
```

### opencode can't connect to the running instance

The plugin connects to opencode on `localhost:4096`. If the CLI is running but the plugin can't connect:

1. Confirm opencode is running (`ps aux | grep opencode`)
2. Confirm it was started without a conflicting `--port` flag
3. Restart opencode from the terminal, then re-trigger the keybinding

### No API key error

```bash
# Confirm the key is set in the environment
echo $ANTHROPIC_API_KEY

# Re-export and re-attach
export ANTHROPIC_API_KEY="sk-ant-..."
dvm attach
```

---

## Related

- [opencode CLI Tool](../dvm/opencode.md) — Workspace tool setup
- [Plugins Reference](plugins.md) — Managing nvp plugins
- [Plugin Packages](packages.md) — rmkohlman package
- [NvimPlugin YAML Reference](../reference/nvim-plugin.md) — Plugin YAML schema
