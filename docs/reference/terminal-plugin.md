# TerminalPlugin YAML Reference

**Kind:** `TerminalPlugin`  
**APIVersion:** `devopsmaestro.io/v1`

A TerminalPlugin defines a shell plugin configuration that can be registered in DevOpsMaestro and referenced by TerminalPackage resources. Plugins represent shell extensions (zsh autosuggestions, syntax highlighting, etc.) managed by a plugin manager such as oh-my-zsh, prezto, or loaded manually.

## Full Example

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPlugin
metadata:
  name: zsh-autosuggestions
  description: "Fish-like autosuggestions for zsh"
  category: productivity
  labels:
    shell: zsh
    manager: oh-my-zsh
  annotations:
    maintainer: zsh-users
spec:
  repo: "zsh-users/zsh-autosuggestions"
  shell: zsh
  manager: oh-my-zsh
  loadCommand: "zsh-autosuggestions"
  sourceFile: ""
  dependencies:
    - zsh
  envVars:
    ZSH_AUTOSUGGEST_HIGHLIGHT_STYLE: "fg=#888888"
    ZSH_AUTOSUGGEST_STRATEGY: "history completion"
```

## Minimal Example

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPlugin
metadata:
  name: zsh-syntax-highlighting
spec:
  repo: "zsh-users/zsh-syntax-highlighting"
```

## Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | ✅ | Must be `devopsmaestro.io/v1` |
| `kind` | string | ✅ | Must be `TerminalPlugin` |
| `metadata.name` | string | ✅ | Unique plugin name |
| `metadata.description` | string | ❌ | Human-readable description |
| `metadata.category` | string | ❌ | Plugin category for organization |
| `metadata.labels` | object | ❌ | Key-value labels |
| `metadata.annotations` | object | ❌ | Key-value annotations |
| `spec.repo` | string | ✅ | Plugin repository (GitHub path or URL) |
| `spec.shell` | string | ❌ | Target shell: `zsh`, `bash`, `fish` (default: `zsh`) |
| `spec.manager` | string | ❌ | Plugin manager: `oh-my-zsh`, `prezto`, `manual` (default: `manual`) |
| `spec.loadCommand` | string | ❌ | Command to activate the plugin |
| `spec.sourceFile` | string | ❌ | File path to source for manual plugins |
| `spec.dependencies` | array | ❌ | List of dependency names required by this plugin |
| `spec.envVars` | object | ❌ | Environment variables to set when the plugin loads |

## Field Details

### metadata.name (required)

The unique identifier for this plugin. Used when referencing the plugin in a TerminalPackage.

**Examples:**
- `zsh-autosuggestions`
- `zsh-syntax-highlighting`
- `fzf-tab`
- `powerlevel10k`

### metadata.category (optional)

Category for organization and filtering.

**Common values:**
- `productivity` — Autosuggestions, history search, completions
- `syntax` — Syntax highlighting
- `navigation` — Directory jumping (zoxide, z, autojump)
- `git` — Git integration
- `prompt` — Prompt frameworks

### spec.repo (required)

The plugin's GitHub repository in `owner/repo` format. Used as a unique identifier and to locate the plugin source.

```yaml
spec:
  repo: "zsh-users/zsh-autosuggestions"
```

### spec.shell (optional)

The target shell for this plugin. Defaults to `zsh` if not specified.

| Value | Description |
|-------|-------------|
| `zsh` | Z shell (default) |
| `bash` | Bourne Again Shell |
| `fish` | Friendly Interactive Shell |

### spec.manager (optional)

The plugin manager used to load this plugin. Defaults to `manual` if not specified.

| Value | Description |
|-------|-------------|
| `oh-my-zsh` | Oh My Zsh plugin manager |
| `prezto` | Prezto Zsh framework |
| `manual` | Manually sourced (no manager) |

### spec.loadCommand (optional)

The command string that activates the plugin within the shell framework. For oh-my-zsh plugins, this is the plugin name as it appears in the `plugins=(...)` array.

```yaml
spec:
  manager: oh-my-zsh
  loadCommand: "zsh-autosuggestions"
```

### spec.sourceFile (optional)

Path to the file to source for manually-loaded plugins. Used when `manager: manual`.

```yaml
spec:
  manager: manual
  sourceFile: "${ZSH_CUSTOM}/plugins/my-plugin/my-plugin.zsh"
```

### spec.dependencies (optional)

List of dependency names that must be present for this plugin to function. These are informational — DevOpsMaestro does not automatically install them.

```yaml
spec:
  dependencies:
    - zsh
    - git
    - fzf
```

### spec.envVars (optional)

Environment variables set when the plugin is loaded. These are injected into the shell rc file before the plugin is sourced.

```yaml
spec:
  envVars:
    ZSH_AUTOSUGGEST_HIGHLIGHT_STYLE: "fg=#888888"
    ZSH_AUTOSUGGEST_STRATEGY: "history completion"
    ZSH_AUTOSUGGEST_BUFFER_MAX_SIZE: "20"
```

## Examples

### Oh My Zsh Plugin

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPlugin
metadata:
  name: zsh-autosuggestions
  description: "Fish-like autosuggestions for zsh"
  category: productivity
spec:
  repo: "zsh-users/zsh-autosuggestions"
  shell: zsh
  manager: oh-my-zsh
  loadCommand: "zsh-autosuggestions"
  envVars:
    ZSH_AUTOSUGGEST_HIGHLIGHT_STYLE: "fg=#888888"
```

### Manually Sourced Plugin

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPlugin
metadata:
  name: fzf-tab
  description: "Replace zsh default completion with fzf"
  category: navigation
spec:
  repo: "Aloxaf/fzf-tab"
  shell: zsh
  manager: manual
  sourceFile: "${ZSH_CUSTOM}/plugins/fzf-tab/fzf-tab.zsh"
  dependencies:
    - fzf
```

### Syntax Highlighting

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPlugin
metadata:
  name: zsh-syntax-highlighting
  description: "Fish-style syntax highlighting for zsh"
  category: syntax
spec:
  repo: "zsh-users/zsh-syntax-highlighting"
  shell: zsh
  manager: oh-my-zsh
  loadCommand: "zsh-syntax-highlighting"
```

### Directory Navigation

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPlugin
metadata:
  name: zoxide
  description: "Smarter cd command with frecency tracking"
  category: navigation
spec:
  repo: "ajeetdsouza/zoxide"
  shell: zsh
  manager: manual
  loadCommand: "eval \"$(zoxide init zsh)\""
  dependencies:
    - zoxide
```

## CLI Commands

### Apply a Plugin

```bash
# Apply from file
dvm apply -f plugin.yaml

# Apply from URL
dvm apply -f https://configs.example.com/plugin.yaml
```

### List Plugins

```bash
# List all terminal plugins
dvm get terminal plugins

# Output as YAML
dvm get terminal plugins -o yaml
```

### Get Plugin Details

```bash
# Get a specific plugin
dvm get terminal plugin zsh-autosuggestions

# Export as YAML
dvm get terminal plugin zsh-autosuggestions -o yaml
```

### Delete a Plugin

```bash
dvm delete terminal plugin zsh-autosuggestions
```

## Validation

- `metadata.name` is required and must be non-empty
- `spec.repo` is required and must be non-empty
- `spec.shell` defaults to `zsh` when not specified
- `spec.manager` defaults to `manual` when not specified

## Storage

TerminalPlugins are stored in the database and referenced by TerminalPackages:

- **Database table**: `terminal_plugins`
- **Package relationship**: Referenced by name in `TerminalPackage.spec.plugins[]`

## Related Resources

- [TerminalPackage](terminal-package.md) — References plugins via `spec.plugins[]`
- [Workspace](workspace.md) — References terminal plugins via `spec.terminal.plugins[]`
- [TerminalPrompt](terminal-prompt.md) — Companion resource for shell prompt configuration
