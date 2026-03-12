# TerminalPrompt YAML Reference

**Kind:** `TerminalPrompt`  
**APIVersion:** `devopsmaestro.io/v1`

A TerminalPrompt defines a shell prompt configuration. Prompts are stored in the database and applied to workspaces. They support multiple prompt engines (Starship, Powerlevel10k, Oh-My-Posh) with structured module configuration and theme color integration.

## Full Example

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPrompt
metadata:
  name: my-starship-prompt
  description: "Custom Starship prompt with theme integration"
  category: development
  tags: ["starship", "git", "go"]
  labels:
    prompt.type: starship
spec:
  type: starship
  addNewline: true
  palette: theme
  format: |
    $directory$git_branch$git_status$golang$python$nodejs
    [❯](bold ${theme.blue})
  modules:
    directory:
      style: "bold ${theme.cyan}"
      options:
        truncation_length: 3
    git_branch:
      format: "[$branch]($style) "
      style: "${theme.purple}"
    git_status:
      format: "[$all_status$ahead_behind]($style) "
      style: "${theme.red}"
    golang:
      format: "[🐹 $version]($style) "
      style: "${theme.sky}"
  character:
    success_symbol: "[❯](bold ${theme.green})"
    error_symbol: "[❯](bold ${theme.red})"
  paletteRef: coolnight-synthwave
```

## Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | ✅ | Must be `devopsmaestro.io/v1` |
| `kind` | string | ✅ | Must be `TerminalPrompt` |
| `metadata.name` | string | ✅ | Unique prompt name |
| `metadata.description` | string | ❌ | Human-readable description |
| `metadata.category` | string | ❌ | Prompt category for organization |
| `metadata.tags` | array | ❌ | Tags for filtering |
| `metadata.labels` | object | ❌ | Key-value labels |
| `metadata.annotations` | object | ❌ | Key-value annotations |
| `spec.type` | string | ✅ | Prompt engine: `starship`, `powerlevel10k`, `oh-my-posh` |
| `spec.addNewline` | boolean | ❌ | Add blank line before prompt (Starship) |
| `spec.palette` | string | ❌ | Palette name to use (`theme` = active theme) |
| `spec.format` | string | ❌ | Prompt format string |
| `spec.modules` | object | ❌ | Module-specific configuration (keyed by module name) |
| `spec.character` | object | ❌ | Prompt character/symbol configuration |
| `spec.paletteRef` | string | ❌ | Reference to a named color palette or theme |
| `spec.colors` | object | ❌ | Custom color overrides (map of name → hex) |
| `spec.rawConfig` | string | ❌ | Raw engine config (for advanced use; bypasses structured fields) |
| `spec.enabled` | boolean | ❌ | Enable or disable the prompt (default: `true`) |

## Field Details

### metadata.name (required)
The unique identifier for this prompt. Used when referencing the prompt in workspace configuration.

**Examples:**
- `my-starship-prompt`
- `dev-prompt`
- `minimal-prompt`

### metadata.category (optional)
Category for organization and filtering.

**Common values:**
- `development` - Developer-focused prompts
- `minimal` - Minimal/clean prompts
- `feature-rich` - Information-dense prompts
- `specialty` - Specialized prompts

### spec.type (required)
The prompt engine to use.

| Value | Description |
|-------|-------------|
| `starship` | Cross-shell prompt written in Rust |
| `powerlevel10k` | Zsh-specific high-performance prompt |
| `oh-my-posh` | Cross-platform prompt engine |

### spec.addNewline (optional, Starship)
When `true`, adds a blank line before the prompt. Equivalent to `add_newline = true` in `starship.toml`. Defaults to `false`.

```yaml
spec:
  type: starship
  addNewline: true
```

### spec.palette (optional)
Palette name to use for color variables. Set to `theme` to use the active DevOpsMaestro theme's colors.

```yaml
spec:
  palette: theme   # Use the active workspace theme
```

### spec.format (optional)
The prompt format string. Uses the engine's native format syntax with optional `${theme.color}` variable interpolation.

```yaml
spec:
  format: |
    $directory$git_branch
    [❯](bold ${theme.blue})
```

### spec.modules (optional)
Per-module configuration, keyed by module name. Each entry is a `ModuleConfig` object.

```yaml
spec:
  modules:
    directory:
      style: "bold ${theme.cyan}"
      options:
        truncation_length: 3
    git_branch:
      format: "[$branch]($style) "
      style: "${theme.purple}"
      symbol: " "
    git_status:
      disabled: false
      format: "[$all_status$ahead_behind]($style) "
      style: "${theme.red}"
```

**ModuleConfig fields:**

| Field | Type | Description |
|-------|------|-------------|
| `disabled` | boolean | Disable this module entirely |
| `format` | string | Module format string |
| `style` | string | Module color/style |
| `symbol` | string | Symbol to display |
| `options` | object | Engine-specific additional options |

### spec.character (optional)
Configures the prompt character symbol that appears at the end of the prompt line.

```yaml
spec:
  character:
    success_symbol: "[❯](bold ${theme.green})"
    error_symbol: "[❯](bold ${theme.red})"
    vicmd_symbol: "[❮](bold ${theme.purple})"
    viins_symbol: "[❯](bold ${theme.blue})"
```

### spec.paletteRef (optional)
Reference to a named palette or theme to use for color resolution. Takes precedence over `palette`.

```yaml
spec:
  paletteRef: coolnight-synthwave
```

### spec.colors (optional)
Custom color overrides as a map of color name to hex value.

```yaml
spec:
  colors:
    primary: "#bd93f9"
    secondary: "#ff79c6"
    accent: "#50fa7b"
```

### spec.rawConfig (optional)
Raw engine configuration string. For advanced users who want to bypass the structured fields and provide the config directly (e.g., a complete `starship.toml`).

```yaml
spec:
  rawConfig: |
    add_newline = true
    [directory]
    style = "bold cyan"
    truncation_length = 3
```

## Theme Variables

Prompt configurations support `${theme.color}` variable interpolation. Variables are resolved from the active theme's color palette at config generation time.

| Variable | Description |
|----------|-------------|
| `${theme.red}` | Primary red |
| `${theme.green}` | Primary green |
| `${theme.blue}` | Primary blue |
| `${theme.purple}` | Primary purple |
| `${theme.cyan}` | Primary cyan |
| `${theme.yellow}` | Primary yellow |
| `${theme.orange}` | Primary orange |
| `${theme.pink}` | Primary pink |
| `${theme.gray}` | Primary gray |
| `${theme.sky}` | Sky blue |
| `${theme.dim}` | Dimmed text |
| `${theme.reset}` | ANSI reset |

## CLI Commands

### List Prompts

```bash
# List all prompts
dvt get prompts

# Filter by type
dvt get prompts --type starship

# Output as YAML
dvt get prompts -o yaml
```

### Get Prompt Details

```bash
# Get prompt details
dvt get prompt my-starship-prompt

# Export as YAML
dvt get prompt my-starship-prompt -o yaml
```

### Apply Prompts

```bash
# Apply from file
dvt prompt apply -f prompt.yaml

# Apply from URL
dvt prompt apply -f https://configs.example.com/prompt.yaml
```

### Generate Configuration

```bash
# Generate starship.toml from prompt
dvt prompt generate my-starship-prompt

# Set as active prompt
dvt prompt set my-starship-prompt
```

### Delete Prompts

```bash
dvt prompt delete my-starship-prompt
```

## Examples

### Minimal Starship Prompt

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPrompt
metadata:
  name: minimal
  description: "Clean single-line prompt"
spec:
  type: starship
  format: "$directory$git_branch [❯](bold ${theme.blue}) "
  modules:
    directory:
      style: "bold ${theme.cyan}"
    git_branch:
      format: "[$branch]($style) "
      style: "${theme.purple}"
```

### Multi-line Development Prompt

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPrompt
metadata:
  name: dev-starship
  description: "Development-focused Starship prompt with language context"
  category: development
  tags: ["starship", "multiline", "git"]
spec:
  type: starship
  addNewline: true
  palette: theme
  format: |
    [${theme.green}╭─${theme.reset}] $directory$git_branch$git_status$golang$python$nodejs
    [${theme.green}╰─${theme.blue}❯${theme.reset}] 
  modules:
    directory:
      style: "bold ${theme.cyan}"
      options:
        truncation_length: 4
    git_branch:
      format: "[$branch]($style) "
      style: "${theme.purple}"
    git_status:
      format: "[$all_status$ahead_behind]($style) "
      style: "${theme.red}"
    golang:
      format: "[🐹 $version]($style) "
      style: "${theme.sky}"
    python:
      format: "[🐍 $version]($style) "
      style: "${theme.yellow}"
    nodejs:
      format: "[⬢ $version]($style) "
      style: "${theme.green}"
  character:
    success_symbol: "[❯](bold ${theme.green})"
    error_symbol: "[❯](bold ${theme.red})"
```

### Workspace Context Prompt

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPrompt
metadata:
  name: workspace-context
  description: "Prompt showing workspace context"
spec:
  type: starship
  format: |
    [${theme.dim}($env_var.DVM_WORKSPACE)${theme.reset}] $directory$git_branch
    [${theme.blue}❯${theme.reset}] 
  modules:
    directory:
      style: "bold ${theme.cyan}"
      format: "[$path]($style) "
    git_branch:
      format: "[$branch]($style) "
      style: "${theme.purple}"
    env_var.DVM_WORKSPACE:
      format: "$env_value"
      style: "${theme.orange}"
```

### Raw Config (Advanced)

For prompts that need full control over the generated config:

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPrompt
metadata:
  name: advanced-starship
  description: "Advanced prompt using raw config"
spec:
  type: starship
  rawConfig: |
    add_newline = true
    format = "$directory$git_branch\n[❯](bold blue) "

    [directory]
    style = "bold cyan"
    truncation_length = 3

    [git_branch]
    format = "[$branch]($style) "
    style = "purple"
```

## Validation

- `metadata.name` is required and must be non-empty
- `spec.type` is required and must be `starship`, `powerlevel10k`, or `oh-my-posh`
- `spec.rawConfig` and structured fields (`spec.format`, `spec.modules`) may be used together but `rawConfig` takes full control when present

## Storage

TerminalPrompts are stored in the database and linked to workspaces:

- **Database table**: `terminal_prompts`
- **Workspace relationship**: Set via `spec.terminal.prompt` in a Workspace YAML
- **Qualified naming**: Prompts are workspace-qualified (`app/workspace/prompt-name`)

## Related Resources

- [Workspace](workspace.md) - Reference prompts via `spec.terminal.prompt`
- [NvimTheme](nvim-theme.md) - Theme palettes used for `${theme.color}` interpolation
