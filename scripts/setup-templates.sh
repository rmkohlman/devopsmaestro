#!/bin/bash
set -e

# DevOpsMaestro Setup Script
# Copies your dev environment templates to ~/.devopsmaestro

echo "=== DevOpsMaestro Template Setup ==="
echo
echo "This script will copy your personal dev environment templates"
echo "to ~/.devopsmaestro/templates/ for use in workspace containers."
echo

# Get home directory
HOME_DIR="$HOME"
DVM_DIR="$HOME_DIR/.devopsmaestro"
TEMPLATES_DIR="$DVM_DIR/templates"

# Ensure dvm is initialized
if [ ! -d "$DVM_DIR" ]; then
    echo "Error: DevOpsMaestro not initialized"
    echo "Please run: dvm admin init"
    exit 1
fi

echo "Target directory: $TEMPLATES_DIR"
echo

# Copy Neovim configuration
if [ -d "$HOME_DIR/.config/nvim" ]; then
    echo "Copying Neovim configuration..."
    mkdir -p "$TEMPLATES_DIR/nvim"
    cp -r "$HOME_DIR/.config/nvim/"* "$TEMPLATES_DIR/nvim/" 2>/dev/null || true
    echo "✓ Neovim config copied"
else
    echo "⚠ Neovim config not found at ~/.config/nvim (skipping)"
fi

# Copy shell configuration
echo
echo "Copying shell configuration..."
mkdir -p "$TEMPLATES_DIR/shell"

# Copy .zshrc (filter out macOS-specific parts)
if [ -f "$HOME_DIR/.zshrc" ]; then
    echo "Processing .zshrc..."
    # Remove homebrew and macOS-specific lines
    grep -v -E '(homebrew|/opt/homebrew|HOMEBREW|brew |Darwin specific)' "$HOME_DIR/.zshrc" > "$TEMPLATES_DIR/shell/.zshrc" || true
    echo "✓ .zshrc copied (filtered for container compatibility)"
else
    echo "⚠ .zshrc not found (skipping)"
fi

# Copy Powerlevel10k config
if [ -f "$HOME_DIR/.p10k.zsh" ]; then
    echo "Copying Powerlevel10k theme config..."
    cp "$HOME_DIR/.p10k.zsh" "$TEMPLATES_DIR/shell/.p10k.zsh"
    echo "✓ .p10k.zsh copied"
else
    echo "⚠ .p10k.zsh not found (skipping)"
fi

# Copy oh-my-zsh custom plugins (if any)
if [ -d "$HOME_DIR/.oh-my-zsh/custom" ]; then
    echo "Copying oh-my-zsh custom plugins..."
    mkdir -p "$TEMPLATES_DIR/shell/oh-my-zsh-custom"
    cp -r "$HOME_DIR/.oh-my-zsh/custom/"* "$TEMPLATES_DIR/shell/oh-my-zsh-custom/" 2>/dev/null || true
    echo "✓ oh-my-zsh custom plugins copied"
fi

# Copy git config (filter out macOS-specific credential helpers)
if [ -f "$HOME_DIR/.gitconfig" ]; then
    echo "Copying git configuration..."
    grep -v -E '(osxkeychain|credential-osxkeychain)' "$HOME_DIR/.gitconfig" > "$TEMPLATES_DIR/shell/.gitconfig" || true
    echo "✓ .gitconfig copied (filtered for container compatibility)"
fi

# Create a README in the templates directory
cat > "$TEMPLATES_DIR/README.md" << 'EOF'
# DevOpsMaestro Templates

This directory contains your personal dev environment templates
that will be copied into workspace containers.

## Structure

- `nvim/` - Neovim configuration from ~/.config/nvim
- `shell/` - Shell configuration files
  - `.zshrc` - zsh configuration (filtered for containers)
  - `.p10k.zsh` - Powerlevel10k theme config
  - `oh-my-zsh-custom/` - Custom oh-my-zsh plugins
  - `.gitconfig` - Git configuration

## Updating Templates

Re-run this setup script to update templates:
```bash
./scripts/setup-templates.sh
```

## Using in Dockerfiles

These templates are automatically copied into workspace containers
during image build. See the multi-stage Dockerfile pattern in the
DevOpsMaestro documentation.
EOF

echo
echo "=== Setup Complete ==="
echo
echo "Templates directory: $TEMPLATES_DIR"
echo
ls -la "$TEMPLATES_DIR"
echo
echo "Next steps:"
echo "  1. Create a project: dvm create project <name> --from-cwd"
echo "  2. Add a Dockerfile to your project that uses these templates"
echo "  3. Start coding: dvm use workspace main && dvm attach"
