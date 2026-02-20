# Terminal Prompt Configuration

DevOpsMaestro supports managing shell prompt configurations through the TerminalPrompt resource system. This allows you to define, share, and apply consistent prompt configurations across your development environments.

## Overview

The TerminalPrompt system supports multiple prompt engines:

- **Starship** - Cross-shell prompt written in Rust
- **Powerlevel10k** - Zsh theme with extensive customization
- **Oh-My-Posh** - Cross-platform, cross-shell prompt engine

Prompts are defined as YAML resources and can include theme variables that automatically resolve to your active theme's colors.

## Quick Start

### 1. Create a Prompt

Create a simple Starship prompt configuration:

```yaml
# starship-prompt.yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPrompt
metadata:
  name: my-starship
  labels:
    type: starship
spec:
  type: starship
  description: "Custom Starship prompt with theme integration"
  config:
    format: |
      $directory $git_branch $git_status
      [${theme.blue}❯${theme.reset}] 
    directory:
      style: "bold ${theme.cyan}"
    git_branch:
      format: "[$branch]($style) "
      style: "${theme.purple}"
    git_status:
      format: "[$all_status$ahead_behind]($style) "
      style: "${theme.red}"
```

### 2. Apply the Prompt

```bash
dvt prompt apply -f starship-prompt.yaml
```

### 3. Generate Configuration

```bash
# Generate starship.toml configuration file
dvt prompt generate my-starship

# Set as active prompt
dvt prompt set my-starship
```

## Theme Integration

TerminalPrompt configurations support theme variable interpolation, allowing your prompts to automatically match your active theme colors.

### Available Theme Variables

| Variable | Description | Example Value |
|----------|-------------|---------------|
| `${theme.red}` | Primary red color | `#ff5555` |
| `${theme.green}` | Primary green color | `#50fa7b` |
| `${theme.blue}` | Primary blue color | `#8be9fd` |
| `${theme.purple}` | Primary purple color | `#bd93f9` |
| `${theme.cyan}` | Primary cyan color | `8be9fd` |
| `${theme.yellow}` | Primary yellow color | `#f1fa8c` |
| `${theme.orange}` | Primary orange color | `#ffb86c` |
| `${theme.pink}` | Primary pink color | `#ff79c6` |
| `${theme.gray}` | Primary gray color | `#6272a4` |
| `${theme.sky}` | Sky blue color | `#87ceeb` |
| `${theme.dim}` | Dimmed text color | `#44475a` |
| `${theme.reset}` | Reset formatting | ANSI reset code |

### Example with Theme Variables

```yaml
spec:
  config:
    format: |
      [${theme.green}╭─${theme.reset}] $directory$git_branch$git_status
      [${theme.green}╰─${theme.blue}❯${theme.reset}] 
    directory:
      style: "bold ${theme.cyan}"
      truncation_length: 3
    git_branch:
      format: "[$branch]($style) "
      style: "${theme.purple}"
    git_status:
      format: "[$all_status$ahead_behind]($style) "
      style: "${theme.red}"
    time:
      disabled: false
      format: "[${theme.dim}$time${theme.reset}]"
```

When you generate the configuration, theme variables are automatically resolved to actual color values from your active theme.

## Prompt Engines

### Starship

Starship is a minimal, blazing-fast, and infinitely customizable prompt for any shell. It's cross-platform and highly performant.

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPrompt
metadata:
  name: dev-starship
spec:
  type: starship
  config:
    format: |
      $directory$git_branch$git_status$golang$python$nodejs
      [${theme.blue}❯${theme.reset}] 
    right_format: |
      $cmd_duration $time
    directory:
      truncation_length: 4
      style: "bold ${theme.cyan}"
    git_branch:
      format: "[$branch]($style) "
      style: "${theme.purple}"
    golang:
      format: "[🐹 $version]($style) "
      style: "${theme.sky}"
    python:
      format: "[🐍 $version]($style) "
      style: "${theme.yellow}"
    nodejs:
      format: "[⬢ $version]($style) "
      style: "${theme.green}"
```

### Powerlevel10k

Powerlevel10k is a theme for Zsh that emphasizes speed, flexibility, and out-of-the-box experience.

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPrompt
metadata:
  name: p10k-custom
spec:
  type: powerlevel10k
  config:
    elements:
      - dir
      - vcs
      - status
      - root_indicator
    colors:
      dir: "${theme.cyan}"
      vcs: "${theme.purple}"
      status_ok: "${theme.green}"
      status_error: "${theme.red}"
    options:
      instant_prompt: true
      multiline: true
```

### Oh-My-Posh

Oh-My-Posh is a custom prompt engine for any shell that has the ability to adjust the prompt string with a function or variable.

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPrompt
metadata:
  name: omp-custom
spec:
  type: oh-my-posh
  config:
    blocks:
      - type: prompt
        alignment: left
        segments:
          - type: path
            style: folder
            background: transparent
            foreground: "${theme.cyan}"
          - type: git
            style: plain
            foreground: "${theme.purple}"
            properties:
              branch_icon: " "
              fetch_status: true
```

## Workspace Integration

TerminalPrompts can include workspace context in the prompt display:

```yaml
spec:
  config:
    format: |
      [${theme.dim}($env_var.DVM_WORKSPACE)${theme.reset}] $directory $git_branch
      [${theme.blue}❯${theme.reset}] 
    env_var:
      DVM_WORKSPACE:
        format: "$env_value"
        style: "${theme.orange}"
```

When attached to a workspace container, the `DVM_WORKSPACE` environment variable contains the current workspace name.

## Command Reference

### List Prompts

```bash
# List all terminal prompts
dvt get prompts

# Filter by type
dvt get prompts --type starship

# Output as YAML for sharing
dvt get prompts -o yaml
```

### Get Prompt Details

```bash
# Show specific prompt details
dvt get prompt my-starship

# Export as YAML
dvt get prompt my-starship -o yaml > my-prompt.yaml
```

### Apply Prompts

```bash
# Apply from file
dvt prompt apply -f prompt.yaml

# Apply from URL
dvt prompt apply -f https://configs.example.com/prompt.yaml

# Apply from GitHub repository
dvt prompt apply -f github:rmkohlman/dvm-config/prompts/starship.yaml
```

### Generate and Set Prompts

```bash
# Generate configuration file (starship.toml)
dvt prompt generate my-starship

# Set as active prompt for current workspace
dvt prompt set my-starship
```

### Delete Prompts

```bash
dvt prompt delete my-starship
```

## Personal Configuration Repository

You can create a personal configuration repository to store and share your prompt configurations:

```bash
# Create a config repo structure
mkdir -p dvm-config/prompts
cd dvm-config

# Add your prompts
cat > prompts/starship.yaml << 'EOF'
apiVersion: devopsmaestro.io/v1
kind: TerminalPrompt
metadata:
  name: personal-starship
spec:
  type: starship
  config:
    format: |
      $directory$git_branch$git_status
      [${theme.blue}❯${theme.reset}] 
EOF

# Push to GitHub
git init && git add . && git commit -m "Add personal prompts"
gh repo create dvm-config --public
git push -u origin main

# Apply from anywhere
dvt prompt apply -f github:username/dvm-config/prompts/starship.yaml
```

## Best Practices

### 1. Use Theme Variables

Always use theme variables instead of hardcoded colors:

```yaml
# Good
style: "${theme.cyan}"

# Avoid
style: "#8be9fd"
```

### 2. Include Workspace Context

Include workspace information for better context:

```yaml
format: |
  [${theme.dim}($env_var.DVM_WORKSPACE)${theme.reset}] $directory
  [${theme.blue}❯${theme.reset}] 
```

### 3. Organize by Environment

Use descriptive names and labels:

```yaml
metadata:
  name: dev-starship
  labels:
    environment: development
    type: starship
```

### 4. Version Control Your Prompts

Store prompts in version control for sharing and backup:

```bash
# Export current prompt
dvt get prompt my-starship -o yaml > prompts/my-starship.yaml
```

### 5. Test Theme Changes

When switching themes, regenerate prompts to pick up new colors:

```bash
dvm set theme coolnight-ocean
dvt prompt generate my-starship  # Pick up new theme colors
```

## Troubleshooting

### Prompt Not Updating

If your prompt isn't reflecting changes:

1. Regenerate the configuration:
   ```bash
   dvt prompt generate my-starship
   ```

2. Restart your shell or source the config:
   ```bash
   source ~/.zshrc  # or ~/.bashrc
   ```

### Theme Variables Not Resolving

Ensure you have an active theme set:

```bash
dvm get context  # Check active theme
dvm set theme coolnight-ocean  # Set a theme if none
```

### Configuration File Location

Generated configurations are typically saved to:

- **Starship**: `~/.config/starship.toml`
- **Powerlevel10k**: `~/.p10k.zsh`
- **Oh-My-Posh**: `~/.poshthemes/<name>.json`

## See Also

- [TerminalPrompt YAML Reference](../reference/terminal-prompt.md)
- [Theme System Documentation](../advanced/theme-system.md)
- [WezTerm Configuration](wezterm.md)
- [Shell Completion](shell-completion.md)