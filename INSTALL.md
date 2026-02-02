# Installation Guide

This repository contains two tools:
- **NvimOps (nvp)** - Standalone Neovim plugin & theme manager
- **DevOpsMaestro (dvm)** - Workspace/container management (requires CGO)

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
curl -L https://github.com/rmkohlman/devopsmaestro/releases/latest/download/nvp_0.5.0_darwin_arm64.tar.gz | tar xz
chmod +x nvp
sudo mv nvp /usr/local/bin/
```

**macOS (Intel):**
```bash
curl -L https://github.com/rmkohlman/devopsmaestro/releases/latest/download/nvp_0.5.0_darwin_amd64.tar.gz | tar xz
chmod +x nvp
sudo mv nvp /usr/local/bin/
```

**Linux (x86_64):**
```bash
curl -L https://github.com/rmkohlman/devopsmaestro/releases/latest/download/nvp_0.5.0_linux_amd64.tar.gz | tar xz
chmod +x nvp
sudo mv nvp /usr/local/bin/
```

**Linux (ARM64):**
```bash
curl -L https://github.com/rmkohlman/devopsmaestro/releases/latest/download/nvp_0.5.0_linux_arm64.tar.gz | tar xz
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

## DevOpsMaestro (dvm) Installation

> **Note:** dvm requires CGO for SQLite and must be built locally on macOS.

### Build from Source (Required)

```bash
git clone https://github.com/rmkohlman/devopsmaestro.git
cd devopsmaestro
go build -o dvm .
sudo mv dvm /usr/local/bin/
```

### DevOpsMaestro Quick Start

```bash
# Initialize
dvm admin init

# Create a project
dvm create project myproject --from-cwd

# Build container
dvm build

# Attach to workspace
dvm attach
```

---

## Quick Install (Development)

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

## System-Wide Install

```bash
cd /path/to/devopsmaestro
sudo make install
```

This installs to `/usr/local/bin` (requires sudo).

## Uninstall

```bash
# For dev install
make uninstall PREFIX=$HOME/.local

# For system install
sudo make uninstall
```

## Homebrew (Future)

Once we publish to Homebrew, you'll be able to install with:

```bash
brew install dvm
```

### How Homebrew Installation Will Work

1. **Create a GitHub Release**
   - Tag a version: `git tag v0.1.0 && git push --tags`
   - GitHub Actions builds binaries for multiple platforms
   - Creates a release with checksums

2. **Publish Homebrew Formula**
   
   Two options:
   
   **Option A: Personal Tap (Easier to start)**
   ```bash
   # Create your tap repository
   brew tap yourusername/tap https://github.com/yourusername/homebrew-tap
   
   # Users install with:
   brew install yourusername/tap/dvm
   ```
   
   **Option B: Official Homebrew (More visibility, stricter requirements)**
   ```bash
   # Submit PR to homebrew-core
   # Users install with:
   brew install dvm
   ```

3. **Users Get Full Homebrew Experience**
   ```bash
   brew install dvm         # Install
   brew upgrade dvm         # Upgrade to latest version
   brew uninstall dvm       # Uninstall
   brew info dvm            # Show info
   brew list --versions dvm # Show installed versions
   ```

## Current Status

✅ **Available Now:**
- Local development install (`make install-dev`)
- System install (`make install`)
- Version info (`dvm version`)
- Works from any directory

🚧 **Coming Soon:**
- GitHub releases with pre-built binaries
- Official Homebrew tap
- Shell completions (bash/zsh/fish)
- Auto-update feature

## Verify Installation

After installation, verify it works:

```bash
dvm version
dvm --help
dvm plugin list
```

## Building from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/devopsmaestro.git
cd devopsmaestro

# Install dependencies
make deps

# Build
make build

# Run tests
make test

# Install
make install-dev
```

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

# Format code
make fmt

# Lint (requires golangci-lint)
make lint

# Build release binaries for all platforms
make release
```

## Makefile Targets

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
make release       # Build for multiple platforms
```

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

## Next Steps

After installation:

1. Initialize your environment: `dvm init`
2. View available plugins: `dvm plugin list`
3. Create a project: `dvm project create my-project`
4. Create a workspace: `dvm workspace create`
5. Build and attach: `dvm build && dvm attach`
