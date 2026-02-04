# Installation Guide

This repository contains two tools:
- **DevOpsMaestro (dvm)** - Workspace/container management
- **NvimOps (nvp)** - Standalone Neovim plugin & theme manager

---

## DevOpsMaestro (dvm) Installation

### Homebrew (Recommended)

```bash
# Add the tap
brew tap rmkohlman/tap

# Install DevOpsMaestro
brew install rmkohlman/tap/devopsmaestro

# Verify installation
dvm version   # Should show v0.7.0
```

### Download Pre-Built Binary

**macOS (Apple Silicon):**
```bash
curl -L https://github.com/rmkohlman/devopsmaestro/releases/download/v0.7.0/devopsmaestro_0.7.0_darwin_arm64.tar.gz | tar xz
chmod +x dvm
sudo mv dvm /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/rmkohlman/devopsmaestro.git
cd devopsmaestro
go build -o dvm .
sudo mv dvm /usr/local/bin/
```

### DevOpsMaestro Quick Start

```bash
# Initialize
dvm init

# Create a project
dvm create project myproject --from-cwd

# Build container image
dvm build

# Attach to workspace (full terminal support!)
dvm attach
```

---

## NvimOps (nvp) Installation

### Homebrew (Recommended)

```bash
# Add the tap
brew tap rmkohlman/tap

# Install nvp
brew install rmkohlman/tap/nvimops

# Verify installation
nvp version
```

### Download Pre-Built Binary

**macOS (Apple Silicon):**
```bash
curl -L https://github.com/rmkohlman/devopsmaestro/releases/download/v0.7.0/nvp_0.7.0_darwin_arm64.tar.gz | tar xz
chmod +x nvp
sudo mv nvp /usr/local/bin/
```

**macOS (Intel):**
```bash
curl -L https://github.com/rmkohlman/devopsmaestro/releases/download/v0.7.0/nvp_0.7.0_darwin_amd64.tar.gz | tar xz
chmod +x nvp
sudo mv nvp /usr/local/bin/
```

**Linux (x86_64):**
```bash
curl -L https://github.com/rmkohlman/devopsmaestro/releases/download/v0.7.0/nvp_0.7.0_linux_amd64.tar.gz | tar xz
chmod +x nvp
sudo mv nvp /usr/local/bin/
```

**Linux (ARM64):**
```bash
curl -L https://github.com/rmkohlman/devopsmaestro/releases/download/v0.7.0/nvp_0.7.0_linux_arm64.tar.gz | tar xz
chmod +x nvp
sudo mv nvp /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/rmkohlman/devopsmaestro.git
cd devopsmaestro
go build -o nvp ./cmd/nvp/
sudo mv nvp /usr/local/bin/
```

### NvimOps Quick Start

```bash
# Initialize
nvp init

# Install plugins from library
nvp library list
nvp library install telescope treesitter lspconfig

# Install a theme
nvp theme library list
nvp theme library install tokyonight-custom --use

# Generate Lua files for Neovim
nvp generate

# Files created in ~/.config/nvim/lua/plugins/nvp/ and ~/.config/nvim/lua/theme/
```

---

## Development Installation

For contributors and developers working on DevOpsMaestro:

### Quick Install (Development)

```bash
cd /path/to/devopsmaestro
make install-dev
```

This installs `dvm` to `~/.local/bin` without requiring sudo.

Add to your PATH in `~/.zshrc` or `~/.bashrc`:
```bash
export PATH="$HOME/.local/bin:$PATH"
```

Then reload your shell:
```bash
source ~/.zshrc
```

### System-Wide Install

```bash
cd /path/to/devopsmaestro
sudo make install
```

This installs to `/usr/local/bin` (requires sudo).

### Uninstall

```bash
# For Homebrew install
brew uninstall devopsmaestro
brew uninstall nvimops

# For dev install
make uninstall PREFIX=$HOME/.local

# For system install
sudo make uninstall
```

---

## Verify Installation

After installation, verify it works:

```bash
dvm version   # Should show v0.7.0
dvm --help
dvm status
```

---

## Development Workflow

For active development:

```bash
# Build and run without installing
make dev

# Or just build
make build
./dvm <command>

# Run tests
make test
go test ./... -race

# Format code
make fmt

# Lint (requires golangci-lint)
make lint
```

### Makefile Targets

```bash
make help          # Show all available targets
make build         # Build the binary
make install       # Install to /usr/local/bin (requires sudo)
make install-dev   # Install to ~/.local/bin (no sudo)
make uninstall     # Remove binary
make test          # Run tests
make clean         # Remove build artifacts
make deps          # Download dependencies
make fmt           # Format code
make lint          # Run linters
make version       # Show version info
```

---

## Troubleshooting

### Command not found after install

Make sure the install directory is in your PATH:

```bash
# For install-dev
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# For system install, /usr/local/bin is usually already in PATH
echo $PATH | grep -q /usr/local/bin && echo "✓ Already in PATH" || echo "✗ Not in PATH"
```

### Permission denied during install

For system install, use sudo:
```bash
sudo make install
```

Or use dev install instead (no sudo needed):
```bash
make install-dev
```

### Database errors

Initialize the database:
```bash
dvm init
```

### macOS Gatekeeper warning

If you see "cannot be opened because the developer cannot be verified":
```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine /usr/local/bin/dvm
```

Or right-click the binary in Finder and select "Open" to allow it.

---

## Next Steps

After installation:

1. Initialize your environment: `dvm init`
2. Check status: `dvm status`
3. Create a project: `dvm create project myproject --from-cwd`
4. Build container: `dvm build`
5. Attach to workspace: `dvm attach`

See the [README](README.md) for full command reference.
