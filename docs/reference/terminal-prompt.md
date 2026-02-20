# TerminalPrompt

TerminalPrompt resources define shell prompt configurations for development environments. They support multiple prompt engines (Starship, Powerlevel10k, Oh-My-Posh) with theme integration for consistent colors.

## Resource Definition

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPrompt
metadata:
  name: my-starship-prompt
  labels:
    prompt.type: starship
spec:
  type: starship
  description: "Custom Starship prompt with theme integration"
  config:
    format: |
      [${theme.green}$directory${theme.reset}](bold) $git_branch$git_status
      [${theme.blue}❯${theme.reset}] 
    right_format: |
      $time
    git_branch:
      format: "[$branch]($style) "
      style: "${theme.purple}"
    git_status:
      format: "[$all_status$ahead_behind]($style) "
      style: "${theme.red}"
    directory:
      truncation_length: 3
      style: "${theme.cyan}"
    time:
      disabled: false
      format: "[${theme.dim}$time${theme.reset}]"
```

## Fields

### metadata

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique prompt name |
| `labels` | map | No | Resource labels for organization |
| `annotations` | map | No | Additional metadata |

### spec

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | Yes | Prompt engine: `starship`, `powerlevel10k`, `oh-my-posh` |
| `description` | string | No | Human-readable prompt description |
| `config` | object | Yes | Engine-specific configuration |

## Prompt Engines

### Starship

Starship uses TOML configuration with theme variable interpolation:

```yaml
spec:
  type: starship
  config:
    format: |
      $directory$git_branch$git_status
      [${theme.blue}❯${theme.reset}] 
    directory:
      style: "${theme.cyan}"
    git_branch:
      style: "${theme.purple}"
```

### Powerlevel10k

Powerlevel10k configuration for Zsh:

```yaml
spec:
  type: powerlevel10k
  config:
    elements:
      - dir
      - vcs
      - status
    colors:
      dir: "${theme.cyan}"
      vcs: "${theme.purple}"
```

### Oh-My-Posh

Oh-My-Posh JSON configuration with theme support:

```yaml
spec:
  type: oh-my-posh
  config:
    blocks:
      - type: prompt
        segments:
          - type: path
            style: folder
            foreground: "${theme.cyan}"
```

## Theme Variables

TerminalPrompt configurations support theme variable interpolation:

| Variable | Description | Example |
|----------|-------------|---------|
| `${theme.red}` | Primary red color | `#ff5555` |
| `${theme.green}` | Primary green color | `#50fa7b` |
| `${theme.blue}` | Primary blue color | `#8be9fd` |
| `${theme.purple}` | Primary purple color | `#bd93f9` |
| `${theme.cyan}` | Primary cyan color | `#8be9fd` |
| `${theme.yellow}` | Primary yellow color | `#f1fa8c` |
| `${theme.orange}` | Primary orange color | `#ffb86c` |
| `${theme.pink}` | Primary pink color | `#ff79c6` |
| `${theme.gray}` | Primary gray color | `#6272a4` |
| `${theme.sky}` | Sky blue color | `#87ceeb` |
| `${theme.dim}` | Dimmed text color | `#44475a` |
| `${theme.reset}` | Reset formatting | ANSI reset code |

Theme variables are resolved from the active theme's color palette when generating configuration files.

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

# Apply from GitHub
dvt prompt apply -f github:rmkohlman/dvm-config/prompts/starship.yaml
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

### Basic Starship Prompt

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPrompt
metadata:
  name: simple-starship
spec:
  type: starship
  description: "Simple Starship prompt"
  config:
    format: |
      $directory $git_branch
      [❯](bold ${theme.blue}) 
    directory:
      style: "bold ${theme.cyan}"
    git_branch:
      format: "[$branch]($style) "
      style: "${theme.purple}"
```

### Multi-line Starship with Status

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPrompt
metadata:
  name: dev-starship
  labels:
    environment: development
spec:
  type: starship
  description: "Development-focused Starship prompt"
  config:
    format: |
      [${theme.green}╭─${theme.reset}] $directory$git_branch$git_status$golang$python$nodejs
      [${theme.green}╰─${theme.blue}❯${theme.reset}] 
    right_format: |
      $cmd_duration $time
    directory:
      truncation_length: 4
      style: "bold ${theme.cyan}"
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
    time:
      disabled: false
      format: "[${theme.dim}$time${theme.reset}]"
    cmd_duration:
      min_time: 2000
      format: "[took ${theme.yellow}$duration${theme.reset}]"
```

### Workspace Context Prompt

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPrompt
metadata:
  name: workspace-context
spec:
  type: starship
  description: "Prompt showing workspace context"
  config:
    format: |
      [${theme.dim}($env_var.DVM_WORKSPACE)${theme.reset}] $directory $git_branch
      [${theme.blue}❯${theme.reset}] 
    directory:
      style: "bold ${theme.cyan}"
      format: "[$path]($style) "
    git_branch:
      format: "[$branch]($style) "
      style: "${theme.purple}"
    env_var:
      DVM_WORKSPACE:
        format: "$env_value"
        style: "${theme.orange}"
```

## Validation

TerminalPrompt resources are validated on creation:

- **Required fields**: `metadata.name`, `spec.type`, `spec.config`
- **Valid types**: Must be one of `starship`, `powerlevel10k`, `oh-my-posh`
- **Theme variables**: Theme variable syntax must be valid (`${theme.color}`)
- **Configuration**: Engine-specific config validation

## Storage

TerminalPrompts are stored in the database and linked to workspaces:

- **Database table**: `terminal_prompts`
- **Workspace relationship**: Prompts can be set as active per workspace
- **Qualified naming**: Prompts are workspace-qualified (`app/workspace/prompt-name`)

## Integration

TerminalPrompts integrate with the broader DevOpsMaestro ecosystem:

- **Theme resolution**: Automatic theme variable interpolation
- **Workspace context**: Environment variables provide workspace context
- **Build process**: Prompts can be applied during container builds
- **Shell integration**: Generated configs work with existing shell setups