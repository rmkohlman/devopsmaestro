# TerminalPackage YAML Reference

**Kind:** `TerminalPackage`  
**APIVersion:** `devopsmaestro.io/v1`

A TerminalPackage is a named, reusable collection of terminal configuration. It bundles together shell plugins, prompt configurations, and terminal profiles into a single unit that can be referenced by a Workspace. Packages support single inheritance via `spec.extends` so a base package can be shared and specialized.

## Full Example

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPackage
metadata:
  name: dev-essentials
  description: "Full developer terminal setup with productivity plugins"
  category: development
  tags:
    - zsh
    - starship
    - productivity
  labels:
    team: platform
    env: dev
  annotations:
    maintainer: platform-team
spec:
  extends: core
  plugins:
    - zsh-autosuggestions
    - zsh-syntax-highlighting
    - fzf-tab
  prompts:
    - starship-minimal
  profiles:
    - developer
  promptStyle: powerline-segments
  promptExtensions:
    - git-status
    - node-version
  wezterm:
    fontSize: 14
    colorScheme: "coolnight-ocean"
    fontFamily: "JetBrains Mono"
  enabled: true
```

## Minimal Example

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPackage
metadata:
  name: core
spec:
  plugins:
    - zsh-autosuggestions
```

## Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | ✅ | Must be `devopsmaestro.io/v1` |
| `kind` | string | ✅ | Must be `TerminalPackage` |
| `metadata.name` | string | ✅ | Unique package name |
| `metadata.description` | string | ❌ | Human-readable description |
| `metadata.category` | string | ❌ | Package category for organization |
| `metadata.tags` | array | ❌ | List of string tags |
| `metadata.labels` | object | ❌ | Key-value labels |
| `metadata.annotations` | object | ❌ | Key-value annotations |
| `spec.extends` | string | ❌ | Name of a parent TerminalPackage to inherit from |
| `spec.plugins` | array or string | ❌ | List of TerminalPlugin names to include |
| `spec.prompts` | array or string | ❌ | List of TerminalPrompt names to include |
| `spec.profiles` | array or string | ❌ | List of terminal profile preset names to include |
| `spec.promptStyle` | string | ❌ | Modular prompt style name (e.g., `powerline-segments`) |
| `spec.promptExtensions` | array or string | ❌ | Modular prompt extension names |
| `spec.wezterm` | object | ❌ | Embedded WezTerm configuration |
| `spec.wezterm.fontSize` | integer | ❌ | Font size in points |
| `spec.wezterm.colorScheme` | string | ❌ | WezTerm color scheme name |
| `spec.wezterm.fontFamily` | string | ❌ | Font family name |
| `spec.enabled` | boolean | ❌ | Whether the package is active (default: `true`) |

## Field Details

### metadata.name (required)

The unique identifier for this package. Used when referencing the package from a Workspace or another TerminalPackage's `spec.extends`.

**Examples:**
- `core`
- `dev-essentials`
- `minimal-shell`
- `platform-defaults`

### metadata.category (optional)

Category for organization and filtering.

**Common values:**
- `development` — Full developer setup
- `minimal` — Lightweight base configurations
- `ops` — Operations and infrastructure tooling
- `personal` — User-specific customizations

### metadata.tags (optional)

A list of string tags for filtering and discovery. Tags are stored as a comma-separated value internally.

```yaml
metadata:
  tags:
    - zsh
    - starship
    - productivity
```

### spec.extends (optional)

The name of a parent TerminalPackage. The current package inherits all plugins, prompts, and profiles from the parent. A package cannot extend itself.

```yaml
spec:
  extends: core
```

!!! note "Single Inheritance"
    Only one level of inheritance is supported. `spec.extends` takes a single package name string.

### spec.plugins (optional)

A list of TerminalPlugin names to include in this package. Each entry must match the `metadata.name` of an existing `TerminalPlugin` resource. Accepts a single string or a YAML list.

```yaml
spec:
  plugins:
    - zsh-autosuggestions
    - zsh-syntax-highlighting
    - fzf-tab
```

Single-plugin shorthand:

```yaml
spec:
  plugins: zsh-autosuggestions
```

### spec.prompts (optional)

A list of TerminalPrompt names to include. Each entry must match the `metadata.name` of an existing `TerminalPrompt` resource. Accepts a single string or a YAML list.

```yaml
spec:
  prompts:
    - starship-minimal
    - p10k-lean
```

### spec.profiles (optional)

A list of terminal profile preset names to include. These represent terminal emulator profile configurations (e.g., WezTerm tabs/domains). Accepts a single string or a YAML list.

```yaml
spec:
  profiles:
    - developer
    - dark-mode
```

### spec.promptStyle (optional)

The name of a modular prompt style. Used with the modular prompt system (v0.19.0+) to configure prompt segment layout independently of a named TerminalPrompt resource.

```yaml
spec:
  promptStyle: powerline-segments
```

### spec.promptExtensions (optional)

A list of modular prompt extension names. Extensions add segments (e.g., git status, node version) to the prompt defined by `spec.promptStyle`. Accepts a single string or a YAML list.

```yaml
spec:
  promptStyle: powerline-segments
  promptExtensions:
    - git-status
    - node-version
    - python-env
```

!!! info "Modular Prompt System"
    `spec.promptStyle` and `spec.promptExtensions` work together. A package uses the modular prompt system when both fields are non-empty. See [Terminal Prompt](terminal-prompt.md) for named prompt resources.

### spec.wezterm (optional)

Inline WezTerm configuration embedded directly in the package. Use this for package-specific terminal appearance overrides without a separate WezTermConfig resource.

```yaml
spec:
  wezterm:
    fontSize: 14
    colorScheme: "coolnight-ocean"
    fontFamily: "JetBrains Mono"
```

| Sub-field | Type | Description |
|-----------|------|-------------|
| `fontSize` | integer | Font size in points |
| `colorScheme` | string | WezTerm color scheme name |
| `fontFamily` | string | Font family name string |

### spec.enabled (optional)

Controls whether this package is active. Defaults to `true`. Set to `false` to deactivate the package without deleting it. Only written to YAML when `false` to keep output clean.

```yaml
spec:
  enabled: false
```

## Examples

### Base Package (no inheritance)

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPackage
metadata:
  name: core
  description: "Minimal base configuration shared by all profiles"
  category: minimal
spec:
  plugins:
    - zsh-autosuggestions
    - zsh-syntax-highlighting
  prompts:
    - starship-minimal
```

### Inherited Package

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPackage
metadata:
  name: dev-essentials
  description: "Developer setup extending core"
  category: development
  tags:
    - zsh
    - fzf
spec:
  extends: core
  plugins:
    - fzf-tab
    - zoxide
  profiles:
    - developer
```

### Package with Modular Prompt

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPackage
metadata:
  name: powerline-dev
  description: "Dev setup with powerline-style segments"
spec:
  extends: core
  promptStyle: powerline-segments
  promptExtensions:
    - git-status
    - node-version
    - python-env
```

### Package with WezTerm Config

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPackage
metadata:
  name: dark-theme-setup
  description: "Full dark theme terminal experience"
  category: personal
spec:
  extends: dev-essentials
  wezterm:
    fontSize: 13
    colorScheme: "coolnight-ocean"
    fontFamily: "JetBrains Mono"
```

### Disabled Package

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPackage
metadata:
  name: experimental
  description: "Work-in-progress, not active"
spec:
  plugins:
    - some-new-plugin
  enabled: false
```

## CLI Commands

### Apply a Package

```bash
# Apply from file
dvm apply -f package.yaml

# Apply from URL
dvm apply -f https://configs.example.com/terminal-package.yaml
```

### List Packages

```bash
# List all terminal packages
dvm get terminal packages

# Output as YAML
dvm get terminal packages -o yaml
```

### Get Package Details

```bash
# Get a specific package
dvm get terminal package dev-essentials

# Export as YAML
dvm get terminal package dev-essentials -o yaml
```

### Delete a Package

```bash
dvm delete terminal package dev-essentials
```

## Validation Rules

- `metadata.name` is required and must be non-empty
- `spec.extends` must not reference the package itself (no self-inheritance)
- All plugin names in `spec.plugins` must be non-empty strings
- All prompt names in `spec.prompts` must be non-empty strings
- All profile names in `spec.profiles` must be non-empty strings
- `spec.enabled` defaults to `true` when not specified
- `apiVersion`, if provided, must be `devopsmaestro.io/v1`
- `kind`, if provided, must be `TerminalPackage`

## Storage

TerminalPackages are stored in the database and referenced by Workspaces:

- **Database table**: `terminal_packages`
- **Plugin relationship**: Plugin names reference `terminal_plugins.name`
- **Prompt relationship**: Prompt names reference `terminal_prompts.name`
- **Workspace relationship**: Referenced by name in `Workspace.spec.terminal.package`

## Related Resources

- [TerminalPlugin](terminal-plugin.md) — Shell plugins referenced via `spec.plugins[]`
- [TerminalPrompt](terminal-prompt.md) — Prompt configurations referenced via `spec.prompts[]`
- [WeztermConfig](wezterm-config.md) — Standalone WezTerm configuration resource
- [Workspace](workspace.md) — References terminal packages via `spec.terminal.package`
